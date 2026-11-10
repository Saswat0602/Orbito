'use client';

import * as React from 'react';
import { useRef } from 'react';
import { Paperclip, Upload, Download, Trash2, File, FileText, Image as ImageIcon } from 'lucide-react';
import { toast } from 'sonner';
import {
  useGetTaskAttachmentsQuery,
  useUploadTaskAttachmentMutation,
  useDeleteTaskAttachmentMutation,
} from '@/lib/store/api/tasksApi';
import { Button } from '@/components/ui/button';

interface TaskAttachmentsProps {
  taskId: string;
  canEdit: boolean;
}

function formatBytes(bytes: number, decimals = 1) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

export const TaskAttachments = React.memo(function TaskAttachments({ taskId, canEdit }: TaskAttachmentsProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  const { data: attachments = [], isLoading } = useGetTaskAttachmentsQuery(taskId, {
    skip: !taskId,
  });

  const [uploadAttachment, { isLoading: isUploading }] = useUploadTaskAttachmentMutation();
  const [deleteAttachment] = useDeleteTaskAttachmentMutation();

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !taskId) return;

    const formData = new FormData();
    formData.append('file', file);

    try {
      await uploadAttachment({ taskId, formData }).unwrap();
      toast.success('File uploaded successfully');
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    } catch {
      toast.error('Failed to upload file');
    }
  };

  const handleDelete = async (attachmentId: string) => {
    if (!canEdit) return;
    try {
      await deleteAttachment({ id: attachmentId, taskId }).unwrap();
      toast.success('Attachment deleted');
    } catch {
      toast.error('Failed to delete attachment');
    }
  };

  const getFileIcon = (mimeType?: string) => {
    if (mimeType?.startsWith('image/')) return <ImageIcon className="h-4 w-4 text-sky-400" />;
    if (mimeType?.includes('pdf') || mimeType?.includes('text')) return <FileText className="h-4 w-4 text-amber-400" />;
    return <File className="h-4 w-4 text-zinc-400" />;
  };

  return (
    <div className="space-y-3 pt-4 border-t border-zinc-800">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Paperclip className="h-4 w-4 text-zinc-400" />
          <span className="text-xs font-bold text-zinc-300">Attachments</span>
          {attachments.length > 0 && (
            <span className="text-[10px] font-mono text-zinc-500">({attachments.length})</span>
          )}
        </div>

        {canEdit && (
          <div>
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileChange}
              className="hidden"
            />
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => fileInputRef.current?.click()}
              isLoading={isUploading}
              className="gap-1.5 text-xs h-7 border-zinc-800"
            >
              <Upload className="h-3 w-3 text-zinc-400" />
              <span>Upload</span>
            </Button>
          </div>
        )}
      </div>

      {/* Attachments List */}
      <div className="space-y-1.5">
        {isLoading ? (
          <div className="py-2 text-xs text-zinc-500 text-center">Loading attachments...</div>
        ) : attachments.length === 0 ? (
          <div className="text-[11px] text-zinc-600 italic">No files attached to this task.</div>
        ) : (
          attachments.map((att) => (
            <div
              key={att.id}
              className="flex items-center justify-between p-2.5 rounded-md bg-zinc-950/40 hover:bg-zinc-950 border border-zinc-850 group transition-colors text-xs"
            >
              <div className="flex items-center gap-2.5 min-w-0">
                {getFileIcon(att.mimeType)}
                <div className="flex flex-col min-w-0">
                  <span className="text-xs text-zinc-200 truncate font-medium">{att.fileName}</span>
                  <span className="text-[10px] text-zinc-500">{formatBytes(att.fileSize)}</span>
                </div>
              </div>

              <div className="flex items-center gap-1 shrink-0">
                <a
                  href={`/api/task-attachments/${att.id}/download`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="p-1 rounded text-zinc-500 hover:text-zinc-200 transition-colors"
                  title="Download attachment"
                >
                  <Download className="h-3.5 w-3.5" />
                </a>

                {canEdit && (
                  <button
                    type="button"
                    onClick={() => handleDelete(att.id)}
                    className="p-1 rounded text-zinc-500 hover:text-rose-400 transition-colors cursor-pointer"
                    title="Delete attachment"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
});

