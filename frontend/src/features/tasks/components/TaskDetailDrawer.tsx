'use client';

import * as React from 'react';
import { useState, useEffect } from 'react';
import {
  X,
  Calendar,
  User,
  AlertCircle,
  Clock,
  CheckCircle2,
  Trash2,
  Send,
  MessageSquare,
} from 'lucide-react';
import { toast } from 'sonner';
import type { Task, TaskPriority } from '@orbito/types';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';
import {
  useGetTaskQuery,
  useUpdateTaskMutation,
  useDeleteTaskMutation,
} from '@/lib/store/api/tasksApi';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

export function TaskDetailDrawer() {
  const dispatch = useAppDispatch();
  const taskDrawerId = useAppSelector((state) => state.ui.taskDrawerId);

  const {
    data: task,
    isLoading,
    isError,
  } = useGetTaskQuery(taskDrawerId || '', {
    skip: !taskDrawerId,
  });

  const [updateTask, { isLoading: isUpdating }] = useUpdateTaskMutation();
  const [deleteTask, { isLoading: isDeleting }] = useDeleteTaskMutation();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [priority, setPriority] = useState<TaskPriority>('MEDIUM' as TaskPriority);
  const [commentText, setCommentText] = useState('');

  useEffect(() => {
    if (task) {
      setTitle(task.title || '');
      setDescription(task.description || '');
      setPriority((task.priority as TaskPriority) || ('MEDIUM' as TaskPriority));
    }
  }, [task]);

  if (!taskDrawerId) return null;

  const handleClose = () => {
    dispatch(setTaskDrawerId(null));
  };

  const handleSaveTitleAndDesc = async () => {
    if (!task) return;
    try {
      await updateTask({
        id: task.id,
        updates: {
          title,
          description,
          priority,
        },
      }).unwrap();
      toast.success('Task updated');
    } catch {
      toast.error('Failed to update task');
    }
  };

  const handleDelete = async () => {
    if (!task) return;
    if (confirm('Are you sure you want to delete this task?')) {
      try {
        await deleteTask(task.id).unwrap();
        toast.success('Task deleted');
        handleClose();
      } catch {
        toast.error('Failed to delete task');
      }
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex justify-end">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/60 backdrop-blur-xs transition-opacity"
        onClick={handleClose}
      />

      {/* Drawer Container */}
      <div className="relative w-full max-w-xl bg-zinc-900 border-l border-zinc-800 h-full flex flex-col shadow-2xl z-10 overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-zinc-800 bg-zinc-950/40">
          <div className="flex items-center gap-2.5">
            <span className="font-mono text-xs font-bold text-zinc-400">
              {task?.slug || `TASK-${taskDrawerId.slice(0, 4)}`}
            </span>
            {task?.status && (
              <Badge variant="outline" className="text-[10px] uppercase font-mono">
                {task.status.name}
              </Badge>
            )}
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleDelete}
              disabled={isDeleting}
              className="p-1.5 rounded-md text-zinc-500 hover:text-rose-400 hover:bg-zinc-800 transition-colors cursor-pointer"
              title="Delete task"
            >
              <Trash2 className="h-4 w-4" />
            </button>
            <button
              type="button"
              onClick={handleClose}
              className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors cursor-pointer"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Content Body */}
        {isLoading ? (
          <div className="flex-1 flex items-center justify-center">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
          </div>
        ) : isError || !task ? (
          <div className="flex-1 flex flex-col items-center justify-center p-6 text-center">
            <span className="text-sm text-zinc-400">Failed to load task details.</span>
            <button
              onClick={handleClose}
              className="mt-3 text-xs text-zinc-200 underline cursor-pointer"
            >
              Close
            </button>
          </div>
        ) : (
          <div className="flex-1 overflow-y-auto p-6 space-y-6">
            {/* Title Input */}
            <div className="space-y-1">
              <Input
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                onBlur={handleSaveTitleAndDesc}
                placeholder="Task title..."
                className="text-base font-bold bg-transparent border-transparent hover:border-zinc-800 focus:border-zinc-700 px-2 py-1 text-zinc-100 rounded-md"
              />
            </div>

            {/* Properties Grid */}
            <div className="grid grid-cols-2 gap-4 p-3.5 rounded-md bg-zinc-950/60 border border-zinc-850 text-xs">
              <div className="space-y-1">
                <span className="text-zinc-500 text-[11px] font-medium">Priority</span>
                <select
                  value={priority}
                  onChange={(e) => {
                    const next = e.target.value as TaskPriority;
                    setPriority(next);
                    updateTask({ id: task.id, updates: { priority: next } });
                  }}
                  className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 rounded px-2 py-1 text-xs outline-none"
                >
                  <option value="LOW">Low</option>
                  <option value="MEDIUM">Medium</option>
                  <option value="HIGH">High</option>
                  <option value="URGENT">Urgent</option>
                </select>
              </div>

              <div className="space-y-1">
                <span className="text-zinc-500 text-[11px] font-medium">Assignee</span>
                <div className="flex items-center gap-2 pt-1 text-zinc-300">
                  <User className="h-3.5 w-3.5 text-zinc-500" />
                  <span>{task.assignees?.[0]?.user?.username || 'Unassigned'}</span>
                </div>
              </div>
            </div>

            {/* Description */}
            <div className="space-y-2">
              <span className="text-xs font-semibold text-zinc-400">Description</span>
              <Textarea
                rows={6}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                onBlur={handleSaveTitleAndDesc}
                placeholder="Add more details, acceptance criteria, or context..."
                className="bg-zinc-950/80 border-zinc-800 text-xs text-zinc-200 resize-none"
              />
            </div>

            {/* Comments Thread */}
            <div className="space-y-3 pt-4 border-t border-zinc-800">
              <div className="flex items-center gap-2">
                <MessageSquare className="h-4 w-4 text-zinc-400" />
                <span className="text-xs font-bold text-zinc-300">Activity & Comments</span>
              </div>

              {/* Comment Input */}
              <div className="flex gap-2">
                <Input
                  value={commentText}
                  onChange={(e) => setCommentText(e.target.value)}
                  placeholder="Write a comment..."
                  className="bg-zinc-950/80 border-zinc-800 text-xs text-zinc-200"
                />
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => {
                    if (commentText.trim()) {
                      toast.info('Comment posted');
                      setCommentText('');
                    }
                  }}
                >
                  <Send className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
