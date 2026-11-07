// src/modules/storage/storage.service.ts
import {
  Injectable,
  NotFoundException,
  BadRequestException,
  Logger,
  OnModuleInit,
} from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { S3Service } from './s3.service';
import * as fs from 'fs';
import * as path from 'path';
import { Response } from 'express';

@Injectable()
export class StorageService implements OnModuleInit {
  private readonly logger = new Logger(StorageService.name);
  private useS3: boolean;
  private readonly uploadDir: string;

  constructor(
    private readonly configService: ConfigService,
    private readonly s3Service: S3Service,
  ) {
    this.uploadDir = this.configService.get('UPLOAD_DEST', './uploads');
    this.useS3 = false;
  }

  async onModuleInit() {
    const awsAccessKey = this.configService.get('AWS_ACCESS_KEY_ID');
    const awsSecretKey = this.configService.get('AWS_SECRET_ACCESS_KEY');
    const awsBucket = this.configService.get('AWS_BUCKET_NAME');

    if (awsAccessKey && awsSecretKey && awsBucket) {
      try {
        const connected = await this.checkS3Connection(awsBucket as string);
        this.useS3 = connected;
        if (connected) {
        } else {
          this.logger.warn('❌ S3 connection check failed, falling back to local storage');
          if (!fs.existsSync(this.uploadDir)) {
            fs.mkdirSync(this.uploadDir, { recursive: true });
          }
        }
      } catch (error) {
        this.logger.error(`S3 connection check error: ${error.message}`);
        this.useS3 = false;
        if (!fs.existsSync(this.uploadDir)) {
          fs.mkdirSync(this.uploadDir, { recursive: true });
        }
      }
    } else {
      if (!fs.existsSync(this.uploadDir)) {
        fs.mkdirSync(this.uploadDir, { recursive: true });
      }
    }
  }

  isUsingS3(): boolean {
    return this.useS3;
  }
  private async checkS3Connection(bucketName: string): Promise<boolean> {
    try {
      // Try a simple operation to test credentials
      await this.s3Service.headBucket(bucketName);
      return true;
    } catch {
      // Only log warning in non-test environments to prevent Jest errors
      if (process.env.NODE_ENV !== 'test') {
        console.warn('S3 connection failed, falling back to local storage.');
      }
      return false;
    }
  }

  /**
   * Sanitize path component to prevent path traversal attacks
   */
  private sanitizePathComponent(component: string): string {
    if (!component || typeof component !== 'string') {
      throw new BadRequestException('Invalid path component');
    }
    // Remove any path traversal attempts and directory separators
    const sanitized = component.replace(/\.\./g, '').replace(/\//g, '').replace(/\\/g, '');
    if (sanitized !== component) {
      throw new BadRequestException('Invalid characters in path');
    }
    return sanitized;
  }

  private sanitizeFolderPath(folder: string): string {
    if (!folder || typeof folder !== 'string') {
      throw new BadRequestException('Invalid folder path');
    }

    // Safety check for path traversal (defensive programming)
    if (folder.includes('..')) {
      throw new Error('Path traversal detected in internally generated path - this is a bug!');
    }

    // Normalize: remove empty parts and extra slashes
    const normalized = folder
      .split('/')
      .filter((part) => part.trim().length > 0)
      .join('/');

    if (normalized.length === 0) {
      throw new BadRequestException('Empty folder path');
    }

    return normalized;
  }

  async saveFile(
    file: Express.Multer.File,
    folder: string,
  ): Promise<{ url: string | null; key: string; size: number }> {
    // Sanitize folder and fileName to prevent path injection
    const safeFolder = this.sanitizeFolderPath(folder);
    const fileName = file.originalname;
    const safeFileName = this.sanitizePathComponent(fileName);

    const key = `${safeFolder}/${safeFileName}`;


    if (this.useS3) {
      await this.s3Service.uploadFile(file, key);
      return {
        url: null,
        key,
        size: file.size,
      };
    } else {
      const localPath = path.join(this.uploadDir, safeFolder);

      if (!fs.existsSync(localPath)) {
        fs.mkdirSync(localPath, { recursive: true });
      }

      const filePath = path.join(localPath, safeFileName);
      fs.writeFileSync(filePath, file.buffer);

      return {
        url: `/${safeFolder}/${safeFileName}`,
        key: `${safeFolder}/${safeFileName}`,
        size: file.size,
      };
    }
  }

  async deleteFile(key: string, inCloud: boolean): Promise<void> {
    if (inCloud) {
      await this.s3Service.deleteFile(key);
    } else {
      const filePath = path.join(this.uploadDir, key);
      if (fs.existsSync(filePath)) {
        fs.unlinkSync(filePath);
      }
    }
  }

  async getFileUrl(key: string): Promise<string> {
    return await this.s3Service.getGetPresignedUrl(key);
  }
  // The only extensions we will render in the browser, and the type each is
  // pinned to. Anything absent here downloads as application/octet-stream, so
  // svg/html/xhtml/js/xml cannot execute script on our own origin - which is
  // what an uploaded file rendered inline would otherwise get to do.
  private static readonly INLINE_SAFE_TYPES: Record<string, string> = {
    '.png': 'image/png',
    '.jpg': 'image/jpeg',
    '.jpeg': 'image/jpeg',
    '.gif': 'image/gif',
    '.webp': 'image/webp',
    '.bmp': 'image/bmp',
    '.ico': 'image/x-icon',
    '.pdf': 'application/pdf',
    '.txt': 'text/plain',
    '.csv': 'text/csv',
    '.mp4': 'video/mp4',
    '.webm': 'video/webm',
    '.ogv': 'video/ogg',
    '.mov': 'video/quicktime',
    '.mp3': 'audio/mpeg',
    '.oga': 'audio/ogg',
    '.wav': 'audio/wav',
  };

  /**
   * Pin the content type instead of leaving the response untyped for the browser
   * to sniff, and only allow inline rendering for types that cannot run script.
   * Called for every streamed file, so an upload cannot become stored XSS.
   */
  private setContentHeaders(key: string, res: Response): void {
    const fileName = path.basename(key);
    const inlineType = StorageService.INLINE_SAFE_TYPES[path.extname(fileName).toLowerCase()];
    // Strip characters that would break out of the quoted-string header value.
    const safeName = fileName.replace(/["\\\r\n]/g, '');

    res.setHeader('X-Content-Type-Options', 'nosniff');
    res.setHeader('Content-Type', inlineType ?? 'application/octet-stream');
    res.setHeader(
      'Content-Disposition',
      `${inlineType ? 'inline' : 'attachment'}; filename="${safeName}"`,
    );
  }

  async streamFromS3(storageKey: string, res: Response): Promise<void> {
    try {
      const stream = await this.s3Service.getFileStream(storageKey);

      this.setContentHeaders(storageKey, res);

      // Pipe S3 stream to response
      stream.pipe(res);

      // Handle stream errors
      stream.on('error', (error) => {
        console.error('S3 stream error:', error);
        if (!res.headersSent) {
          res.status(500).json({ message: 'Failed to stream file from S3' });
        }
      });
    } catch (error) {
      console.error('Failed to get S3 stream:', error);
      throw new NotFoundException('File not found in S3');
    }
  }
  streamFromLocal(filePath: string, res: Response): void {
    // Check if file exists
    const relativePath = filePath.startsWith('/') ? filePath.substring(1) : filePath;
    const normalizedPath = path.normalize(relativePath);
    const localPath = path.resolve(this.uploadDir, normalizedPath);

    // Prevent path traversal attacks by ensuring the resolved path is within uploadDir
    const uploadDirResolved = path.resolve(this.uploadDir);
    if (!localPath.startsWith(uploadDirResolved + path.sep) && localPath !== uploadDirResolved) {
      throw new NotFoundException('Invalid file path');
    }

    if (!fs.existsSync(localPath)) {
      throw new NotFoundException('File not found on server');
    }

    try {
      this.setContentHeaders(localPath, res);

      const fileStream = fs.createReadStream(localPath);

      // Pipe local file stream to response
      fileStream.pipe(res);

      // Handle stream errors
      fileStream.on('error', (error) => {
        console.error('Local file stream error:', error);
        if (!res.headersSent) {
          res.status(500).json({ message: 'Failed to stream file from local storage' });
        }
      });
    } catch (error) {
      console.error('Failed to create local file stream:', error);
      throw new NotFoundException('Failed to read file');
    }
  }
}
