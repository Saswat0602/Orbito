'use client';

import React, { useEffect, useState, useCallback } from 'react';
import {
  X,
  Download,
  ZoomIn,
  ZoomOut,
  RotateCcw,
  File,
  FileText,
  Image as ImageIcon,
  Film,
  ExternalLink,
} from 'lucide-react';
import { Button } from '@/components/ui/button';

export interface LightboxAttachment {
  id: string;
  fileName: string;
  fileSize: number;
  mimeType?: string;
  previewUrl?: string;
  downloadUrl?: string;
}

interface AttachmentLightboxModalProps {
  isOpen: boolean;
  onClose: () => void;
  attachment: LightboxAttachment | null;
}

function formatBytes(bytes: number, decimals = 1): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

export const AttachmentLightboxModal: React.FC<AttachmentLightboxModalProps> = ({
  isOpen,
  onClose,
  attachment,
}) => {
  const [zoomScale, setZoomScale] = useState(1);

  // Reset zoom whenever attachment changes
  useEffect(() => {
    setZoomScale(1);
  }, [attachment]);

  // Handle ESC key to close
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    },
    [onClose]
  );

  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
      window.addEventListener('keydown', handleKeyDown);
    } else {
      document.body.style.overflow = '';
    }

    return () => {
      document.body.style.overflow = '';
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, handleKeyDown]);

  if (!isOpen || !attachment) return null;

  const previewUrl =
    attachment.previewUrl || `/api/task-attachments/${attachment.id}/preview`;
  const downloadUrl =
    attachment.downloadUrl || `/api/task-attachments/${attachment.id}/download`;

  const mime = attachment.mimeType?.toLowerCase() || '';
  const ext = attachment.fileName.split('.').pop()?.toLowerCase() || '';

  const isImage =
    mime.startsWith('image/') ||
    ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'].includes(ext);

  const isVideo =
    mime.startsWith('video/') || ['mp4', 'webm', 'ogg', 'mov'].includes(ext);

  const isPdf = mime.includes('pdf') || ext === 'pdf';

  const handleZoomIn = () => setZoomScale((prev) => Math.min(prev + 0.25, 3));
  const handleZoomOut = () => setZoomScale((prev) => Math.max(prev - 0.25, 0.5));
  const handleZoomReset = () => setZoomScale(1);

  return (
    <div className="fixed inset-0 z-50 flex flex-col bg-black/90 backdrop-blur-sm animate-in fade-in-0 duration-150 select-none">
      {/* Top Header Bar */}
      <div className="flex h-14 items-center justify-between border-b border-zinc-800 bg-zinc-950/80 px-6 shrink-0 z-10">
        <div className="flex items-center gap-3 truncate pr-4">
          {isImage ? (
            <ImageIcon className="h-4 w-4 text-sky-400 shrink-0" />
          ) : isVideo ? (
            <Film className="h-4 w-4 text-purple-400 shrink-0" />
          ) : isPdf ? (
            <FileText className="h-4 w-4 text-amber-400 shrink-0" />
          ) : (
            <File className="h-4 w-4 text-zinc-400 shrink-0" />
          )}

          <div className="truncate">
            <span className="text-xs font-semibold text-zinc-100 block truncate">
              {attachment.fileName}
            </span>
            <span className="text-[10px] text-zinc-400">
              {formatBytes(attachment.fileSize)}
            </span>
          </div>
        </div>

        {/* Action Controls */}
        <div className="flex items-center gap-2">
          {isImage && (
            <div className="flex items-center gap-1 border border-zinc-800 rounded bg-zinc-900/60 p-0.5 mr-2">
              <button
                type="button"
                onClick={handleZoomOut}
                disabled={zoomScale <= 0.5}
                className="p-1 rounded text-zinc-400 hover:text-zinc-200 disabled:opacity-40"
                title="Zoom Out"
              >
                <ZoomOut className="h-3.5 w-3.5" />
              </button>
              <button
                type="button"
                onClick={handleZoomReset}
                className="px-1.5 py-0.5 text-[10px] font-mono text-zinc-300 hover:text-white"
                title="Reset Zoom"
              >
                {Math.round(zoomScale * 100)}%
              </button>
              <button
                type="button"
                onClick={handleZoomIn}
                disabled={zoomScale >= 3}
                className="p-1 rounded text-zinc-400 hover:text-zinc-200 disabled:opacity-40"
                title="Zoom In"
              >
                <ZoomIn className="h-3.5 w-3.5" />
              </button>
              <button
                type="button"
                onClick={handleZoomReset}
                className="p-1 rounded text-zinc-400 hover:text-zinc-200"
                title="Reset"
              >
                <RotateCcw className="h-3 w-3" />
              </button>
            </div>
          )}

          <a
            href={downloadUrl}
            target="_blank"
            rel="noopener noreferrer"
            download={attachment.fileName}
            className="flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium text-zinc-200 bg-zinc-900 hover:bg-zinc-800 border border-zinc-800 rounded transition"
            title="Download file"
          >
            <Download className="h-3.5 w-3.5" />
            <span>Download</span>
          </a>

          <button
            type="button"
            onClick={onClose}
            className="p-1.5 rounded-md text-zinc-400 hover:text-white hover:bg-zinc-800 transition"
            title="Close (Esc)"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
      </div>

      {/* Main Lightbox Content Area */}
      <div
        className="flex-1 flex items-center justify-center p-6 overflow-hidden relative"
        onClick={onClose}
      >
        <div
          className="relative max-h-full max-w-full flex items-center justify-center"
          onClick={(e) => e.stopPropagation()}
        >
          {isImage ? (
            <div className="overflow-auto max-h-[82vh] max-w-[90vw] flex items-center justify-center rounded">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={previewUrl}
                alt={attachment.fileName}
                style={{
                  transform: `scale(${zoomScale})`,
                  transition: 'transform 0.15s ease-out',
                }}
                className="max-h-[80vh] max-w-[85vw] object-contain rounded shadow-2xl"
              />
            </div>
          ) : isVideo ? (
            <video
              controls
              autoPlay
              src={previewUrl}
              className="max-h-[80vh] max-w-[85vw] rounded border border-zinc-800 shadow-2xl bg-black"
            >
              Your browser does not support HTML5 video preview.
            </video>
          ) : isPdf ? (
            <div className="w-[85vw] h-[80vh] max-w-5xl rounded border border-zinc-800 bg-zinc-950 shadow-2xl overflow-hidden flex flex-col">
              <iframe
                src={previewUrl}
                className="w-full h-full border-none"
                title={attachment.fileName}
              />
            </div>
          ) : (
            /* Fallback file preview card */
            <div className="p-8 border border-zinc-800 rounded-lg bg-zinc-950/90 text-center flex flex-col items-center gap-3 shadow-2xl max-w-md">
              <div className="h-16 w-16 rounded-full bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-400">
                <File className="h-8 w-8" />
              </div>
              <div>
                <h4 className="text-sm font-semibold text-zinc-100">
                  {attachment.fileName}
                </h4>
                <p className="text-xs text-zinc-400 mt-1">
                  Preview is not supported for this file type ({ext.toUpperCase() || 'Binary'})
                </p>
                <p className="text-xs text-zinc-500 mt-0.5">
                  Size: {formatBytes(attachment.fileSize)}
                </p>
              </div>

              <a
                href={downloadUrl}
                target="_blank"
                rel="noopener noreferrer"
                download={attachment.fileName}
                className="mt-2 flex items-center gap-2 px-4 py-2 text-xs font-semibold text-zinc-950 bg-zinc-100 hover:bg-white rounded transition shadow-md"
              >
                <Download className="h-4 w-4" />
                <span>Download File</span>
              </a>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};