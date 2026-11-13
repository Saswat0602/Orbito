'use client';

import React, { useState } from 'react';
import { useParams } from 'next/navigation';
import {
  Calendar,
  Clock,
  User,
  Paperclip,
  Download,
  AlertTriangle,
  ShieldCheck,
  FileText,
  File,
  Image as ImageIcon,
  Film,
  Eye,
  CheckCircle2,
} from 'lucide-react';
import { useGetPublicTaskQuery } from '@/lib/store/api/tasksApi';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { TaskDetailSkeleton } from '@/components/skeletons';
import {
  AttachmentLightboxModal,
  LightboxAttachment,
} from '@/features/tasks/components/AttachmentLightboxModal';

function formatBytes(bytes: number, decimals = 1): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

export default function PublicSharedTaskPage() {
  const params = useParams<{ token: string }>();
  const token = params?.token || '';

  const {
    data: task,
    isLoading,
    isError,
    error,
  } = useGetPublicTaskQuery(token, {
    skip: !token,
  });

  const [activeLightboxAttachment, setActiveLightboxAttachment] =
    useState<LightboxAttachment | null>(null);

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'URGENT':
        return 'text-rose-400 border-rose-800 bg-rose-950/40';
      case 'HIGH':
        return 'text-amber-400 border-amber-800 bg-amber-950/40';
      case 'MEDIUM':
        return 'text-sky-400 border-sky-800 bg-sky-950/40';
      default:
        return 'text-zinc-400 border-zinc-800 bg-zinc-950/40';
    }
  };

  const getFileIcon = (mimeType?: string) => {
    if (mimeType?.startsWith('image/'))
      return <ImageIcon className="h-4 w-4 text-sky-400" />;
    if (mimeType?.startsWith('video/'))
      return <Film className="h-4 w-4 text-purple-400" />;
    if (mimeType?.includes('pdf') || mimeType?.includes('text'))
      return <FileText className="h-4 w-4 text-amber-400" />;
    return <File className="h-4 w-4 text-zinc-400" />;
  };

  const handleDownloadAttachment = async (attachmentId: string, fileName: string) => {
    try {
      const res = await fetch(`/api/public/tasks/${token}/attachments/${attachmentId}`);
      if (!res.ok) throw new Error('Failed to get attachment URL');
      const data = await res.json();
      if (data?.url) {
        window.open(data.url, '_blank');
      }
    } catch {
      window.open(`/api/public/tasks/${token}/attachments/${attachmentId}`, '_blank');
    }
  };

  return (
    <div className="min-h-screen bg-[#09090b] text-zinc-100 flex flex-col selection:bg-zinc-800">
      {/* Top Banner Navigation */}
      <header className="h-14 border-b border-zinc-800 bg-zinc-950/70 backdrop-blur-md px-6 flex items-center justify-between sticky top-0 z-20">
        <div className="flex items-center gap-3">
          <div className="h-6 w-6 rounded bg-zinc-100 text-zinc-950 flex items-center justify-center font-bold text-xs tracking-tight">
            O
          </div>
          <span className="font-semibold text-sm tracking-tight text-zinc-200">
            Orbito
          </span>
          <span className="text-zinc-600 text-xs">/</span>
          <span className="text-xs text-zinc-400">Public Shared Task</span>
        </div>

        <div className="flex items-center gap-2">
          <Badge
            variant="outline"
            className="gap-1.5 text-[11px] font-normal border-zinc-800 text-zinc-400 bg-zinc-900/60"
          >
            <ShieldCheck className="h-3.5 w-3.5 text-emerald-400" />
            <span>Read-Only View</span>
          </Badge>
        </div>
      </header>

      {/* Main Container */}
      <main className="flex-1 flex justify-center p-4 sm:p-8">
        <div className="w-full max-w-3xl space-y-6">
          {isLoading ? (
            <TaskDetailSkeleton />
          ) : isError || !task ? (
            <div className="p-12 border border-zinc-800 rounded-md bg-zinc-950/60 text-center space-y-3">
              <div className="mx-auto h-12 w-12 rounded-full bg-red-950/50 border border-red-800/80 flex items-center justify-center text-red-400">
                <AlertTriangle className="h-6 w-6" />
              </div>
              <h2 className="text-sm font-semibold text-zinc-200">
                Link Expired or Invalid
              </h2>
              <p className="text-xs text-zinc-500 max-w-md mx-auto">
                This public task link is either expired, revoked, or incorrect. Please contact the project team member who shared this link with you.
              </p>
            </div>
          ) : (
            <div className="border border-zinc-800 rounded-md bg-zinc-950/80 shadow-2xl overflow-hidden divide-y divide-zinc-850">
              {/* Task Header Details */}
              <div className="p-6 space-y-4">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge
                    variant="outline"
                    className="text-[11px] font-medium border-zinc-700 bg-zinc-900 text-zinc-200 uppercase px-2 py-0.5"
                  >
                    {task.status?.name || 'Status'}
                  </Badge>

                  <Badge
                    variant="outline"
                    className={`text-[11px] font-medium uppercase px-2 py-0.5 ${getPriorityColor(
                      task.priority
                    )}`}
                  >
                    {task.priority || 'MEDIUM'}
                  </Badge>

                  {task.dueDate && (
                    <div className="flex items-center gap-1.5 text-xs text-zinc-400 border border-zinc-800 rounded px-2 py-0.5 bg-zinc-900/40">
                      <Calendar className="h-3.5 w-3.5 text-zinc-500" />
                      <span>
                        Due{' '}
                        {new Date(task.dueDate).toLocaleDateString('en-US', {
                          month: 'short',
                          day: 'numeric',
                          year: 'numeric',
                        })}
                      </span>
                    </div>
                  )}
                </div>

                <h1 className="text-xl font-bold tracking-tight text-white leading-snug">
                  {task.title}
                </h1>

                {/* Assignees and Creator */}
                <div className="flex flex-wrap items-center gap-4 text-xs text-zinc-400 pt-1">
                  {task.assignees && task.assignees.length > 0 && (
                    <div className="flex items-center gap-2">
                      <span className="text-zinc-500">Assigned:</span>
                      <div className="flex items-center gap-1.5">
                        {task.assignees.map((assignee, idx) => (
                          <div
                            key={idx}
                            className="flex items-center gap-1.5 bg-zinc-900 border border-zinc-800 rounded px-2 py-0.5"
                          >
                            <Avatar className="h-4 w-4 rounded-sm border border-zinc-800">
                              <AvatarFallback className="text-[9px] bg-zinc-800 text-zinc-300">
                                {((assignee.firstName?.[0] || '') +
                                  (assignee.lastName?.[0] || '')).toUpperCase() || 'U'}
                              </AvatarFallback>
                            </Avatar>
                            <span className="text-zinc-300">
                              {assignee.firstName} {assignee.lastName}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {task.createdBy && (
                    <div className="flex items-center gap-1.5 text-zinc-500">
                      <span>Created by:</span>
                      <span className="text-zinc-400 font-medium">
                        {task.createdBy.firstName} {task.createdBy.lastName}
                      </span>
                    </div>
                  )}
                </div>
              </div>

              {/* Task Description */}
              <div className="p-6 space-y-2">
                <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-400">
                  Description
                </h3>
                {task.description ? (
                  <div className="text-xs text-zinc-300 leading-relaxed whitespace-pre-wrap font-sans bg-zinc-900/30 p-4 rounded border border-zinc-850">
                    {task.description}
                  </div>
                ) : (
                  <p className="text-xs text-zinc-600 italic">
                    No description provided for this task.
                  </p>
                )}
              </div>

              {/* Attachments Section */}
              {task.attachments && task.attachments.length > 0 && (
                <div className="p-6 space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Paperclip className="h-4 w-4 text-zinc-400" />
                      <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-400">
                        Attachments ({task.attachments.length})
                      </h3>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
                    {task.attachments.map((att) => (
                      <div
                        key={att.id}
                        onClick={() =>
                          setActiveLightboxAttachment({
                            id: att.id,
                            fileName: att.fileName,
                            fileSize: att.fileSize,
                            mimeType: att.mimeType,
                            previewUrl: `/api/public/tasks/${token}/attachments/${att.id}`,
                            downloadUrl: `/api/public/tasks/${token}/attachments/${att.id}`,
                          })
                        }
                        className="flex items-center justify-between p-3 rounded-md bg-zinc-900/50 hover:bg-zinc-900 border border-zinc-800 cursor-pointer transition group"
                      >
                        <div className="flex items-center gap-2.5 truncate pr-2">
                          {getFileIcon(att.mimeType)}
                          <div className="truncate">
                            <span className="text-xs font-medium text-zinc-200 block truncate group-hover:text-white">
                              {att.fileName}
                            </span>
                            <span className="text-[10px] text-zinc-500">
                              {formatBytes(att.fileSize)}
                            </span>
                          </div>
                        </div>

                        <div
                          className="flex items-center gap-1 shrink-0"
                          onClick={(e) => e.stopPropagation()}
                        >
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={() =>
                              handleDownloadAttachment(att.id, att.fileName)
                            }
                            className="h-7 w-7 p-0 text-zinc-400 hover:text-zinc-100"
                            title="Download"
                          >
                            <Download className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </main>

      {/* Lightbox Modal */}
      <AttachmentLightboxModal
        isOpen={!!activeLightboxAttachment}
        onClose={() => setActiveLightboxAttachment(null)}
        attachment={activeLightboxAttachment}
      />
    </div>
  );
}