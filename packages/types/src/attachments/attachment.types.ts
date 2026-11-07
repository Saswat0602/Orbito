import type { UserSummary } from '../users/user.types';

export type FileType =
  | 'image'
  | 'pdf'
  | 'document'
  | 'spreadsheet'
  | 'text'
  | 'video'
  | 'unknown';

export interface Attachment {
  id: string;
  fileName: string;
  fileSize: number;
  fileType?: string;
  mimeType?: string;
  url?: string;
  filePath?: string;
  createdBy?: string;
  uploadedById?: string;
  uploadedBy?: UserSummary;
  taskId?: string;
  createdAt: string;
  updatedAt?: string;
}

export interface PreviewCache {
  [attachmentId: string]: {
    url: string;
    timestamp: number;
  };
}

export interface CreateAttachmentRequest {
  fileName: string;
  filePath: string;
  fileSize: number;
  mimeType: string;
  taskId: string;
}

export interface AttachmentStats {
  totalAttachments: number;
  totalSize: number;
  fileTypes: Record<string, number>;
}
