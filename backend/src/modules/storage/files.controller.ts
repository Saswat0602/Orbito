// src/modules/files/files.controller.ts
import {
  Controller,
  Get,
  Param,
  Res,
  UseGuards,
  Post,
  UseInterceptors,
  UploadedFile,
  BadRequestException,
} from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Response } from 'express';
import {
  ApiTags,
  ApiOperation,
  ApiParam,
  ApiResponse,
  ApiBearerAuth,
  ApiConsumes,
  ApiBody,
} from '@nestjs/swagger';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { FileInterceptor } from '@nestjs/platform-express';
import { StorageService } from './storage.service';
import { Public } from '../auth/decorators/public.decorator';

// The only caller of POST /uploads/upload/:folder is the profile-picture
// uploader (frontend authApi.uploadFileToS3), always with folder "avatar".
// Locking the folder param to that keeps an authed user from writing into
// tasks/ or editor-images/ and having it served back by the public routes below.
const ALLOWED_UPLOAD_FOLDERS = new Set(['avatar']);

// 50MB limit. Override with MAX_AVATAR_SIZE (bytes).
const DEFAULT_MAX_AVATAR_SIZE = 50 * 1024 * 1024;

@ApiTags('Files')
@ApiBearerAuth('JWT-auth')
@Controller('uploads')
export class FilesController {
  private readonly maxAvatarSize: number;
  private readonly allowedAvatarTypes: string[];

  constructor(
    private readonly storageService: StorageService,
    private readonly configService: ConfigService,
  ) {
    const rawMaxAvatarSize = this.configService.get<string>(
      'MAX_AVATAR_SIZE',
      String(DEFAULT_MAX_AVATAR_SIZE),
    );
    this.maxAvatarSize = parseInt(rawMaxAvatarSize, 10);
    // Fail fast on a bad value instead of silently allowing unbounded uploads.
    if (!Number.isFinite(this.maxAvatarSize) || this.maxAvatarSize <= 0) {
      throw new Error(
        `MAX_AVATAR_SIZE must be a positive number of bytes, got "${rawMaxAvatarSize}"`,
      );
    }
    this.allowedAvatarTypes = this.configService
      .get<string>('ALLOWED_IMAGE_TYPES', 'image/jpeg,image/png,image/gif,image/webp')
      .split(',')
      .map((t) => t.trim());
  }

  private formatBytes(bytes: number): string {
    if (bytes >= 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)}GB`;
    return `${Math.round(bytes / (1024 * 1024))}MB`;
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

  @Public()
  @Get('tasks/:taskId/:filename')
  @ApiOperation({ summary: 'Serve task attachment file' })
  @ApiParam({ name: 'taskId', description: 'Task ID' })
  @ApiParam({ name: 'filename', description: 'Filename' })
  @ApiResponse({ status: 200, description: 'File served successfully' })
  @ApiResponse({ status: 404, description: 'File not found' })
  async serveTaskFile(
    @Param('taskId') taskId: string,
    @Param('filename') filename: string,
    @Res() res: Response,
  ) {
    // Sanitize path components to prevent path injection
    const safeTaskId = this.sanitizePathComponent(taskId);
    const safeFilename = this.sanitizePathComponent(filename);

    const key = `tasks/${safeTaskId}/${safeFilename}`;

    if (this.storageService.isUsingS3()) {
      await this.storageService.streamFromS3(key, res);
    } else {
      this.storageService.streamFromLocal(key, res);
    }
  }
  @Post('upload/:folder')
  @UseGuards(JwtAuthGuard)
  @ApiOperation({ summary: 'Upload a profile picture (only folder currently supported: avatar)' })
  @ApiParam({ name: 'folder', example: 'avatar', description: 'Must be "avatar"' })
  @ApiConsumes('multipart/form-data')
  @ApiBody({
    description: 'Image file to upload (JPEG, PNG, GIF, WebP, max 50MB by default)',
    schema: {
      type: 'object',
      properties: {
        file: { type: 'string', format: 'binary' },
      },
    },
  })
  @ApiResponse({
    status: 400,
    description: 'Invalid folder, missing file, wrong type, or too large',
  })
  @UseInterceptors(
    FileInterceptor('file', {
      // Decorators run before DI, so this can't read MAX_AVATAR_SIZE - it's just
      // a hard ceiling stopping multer from buffering an unbounded upload into
      // memory. The real, configurable limit is enforced below once DI is available.
      limits: { fileSize: DEFAULT_MAX_AVATAR_SIZE },
    }),
  )
  async uploadFile(@Param('folder') folder: string, @UploadedFile() file: Express.Multer.File) {
    // Sanitize, then restrict to the folders this route actually serves -
    // an arbitrary folder would let an authed user write into tasks/ or
    // editor-images/ and have it served back by the public routes below.
    const safeFolder = this.sanitizePathComponent(folder);
    if (!ALLOWED_UPLOAD_FOLDERS.has(safeFolder)) {
      throw new BadRequestException(`Unsupported upload folder: ${safeFolder}`);
    }

    if (!file) {
      throw new BadRequestException('No file uploaded');
    }
    if (!this.allowedAvatarTypes.includes(file.mimetype)) {
      throw new BadRequestException(
        `Invalid file type. Allowed types: ${this.allowedAvatarTypes.join(', ')}`,
      );
    }
    if (file.size > this.maxAvatarSize) {
      throw new BadRequestException(
        `File size exceeds ${this.formatBytes(this.maxAvatarSize)} limit`,
      );
    }

    const savedFile = await this.storageService.saveFile(file, safeFolder);

    return {
      message: 'File uploaded successfully',
      ...savedFile,
      inCloud: this.storageService.isUsingS3(),
    };
  }
  @Public()
  @Get(':folder/:filename')
  @ApiOperation({ summary: 'Serve file from storage' })
  @ApiResponse({ status: 200, description: 'File served successfully' })
  @ApiResponse({ status: 404, description: 'File not found' })
  async serveFile(
    @Param('folder') folder: string,
    @Param('filename') filename: string,
    @Res() res: Response,
  ) {
    // Sanitize path components to prevent path injection
    const safeFolder = this.sanitizePathComponent(folder);
    const safeFilename = this.sanitizePathComponent(filename);

    const key = `${safeFolder}/${safeFilename}`;
    // Check if using S3 or local storage
    if (this.storageService.isUsingS3()) {
      // For S3, stream the file directly
      await this.storageService.streamFromS3(key, res);
    } else {
      // For local storage, use the existing method
      this.storageService.streamFromLocal(key, res);
    }
  }
  @Public()
  @Get('editor-images/:filename')
  @ApiOperation({ summary: 'Serve editor image file' })
  @ApiParam({ name: 'filename', description: 'Image filename' })
  @ApiResponse({ status: 200, description: 'File served successfully' })
  @ApiResponse({ status: 404, description: 'File not found' })
  async serveEditorImage(@Param('filename') filename: string, @Res() res: Response) {
    // Sanitize path components to prevent path injection
    const safeFilename = this.sanitizePathComponent(filename);

    const key = `editor-images/${safeFilename}`;

    // Check if using S3 or local storage
    if (this.storageService.isUsingS3()) {
      // For S3, stream the file directly
      await this.storageService.streamFromS3(key, res);
    } else {
      // For local storage, use the existing method
      this.storageService.streamFromLocal(key, res);
    }
  }
}
