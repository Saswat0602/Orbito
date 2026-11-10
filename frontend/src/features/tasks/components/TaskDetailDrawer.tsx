'use client';

import * as React from 'react';
import { useState, useEffect } from 'react';
import {
  X,
  User,
  Trash2,
  Send,
  MessageSquare,
  Layers,
  Clock,
} from 'lucide-react';
import { toast } from 'sonner';
import type { Task, TaskPriority } from '@orbito/types';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';
import {
  useGetTaskQuery,
  useUpdateTaskMutation,
  useDeleteTaskMutation,
  useGetTaskCommentsQuery,
  useCreateTaskCommentMutation,
  useDeleteTaskCommentMutation,
  useUpdateTaskAssigneesMutation,
} from '@/lib/store/api/tasksApi';
import { useGetWorkspaceMembersQuery } from '@/lib/store/api/workspacesApi';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

const STATUS_OPTIONS = [
  { value: 'TODO', label: 'To Do' },
  { value: 'IN_PROGRESS', label: 'In Progress' },
  { value: 'IN_REVIEW', label: 'In Review' },
  { value: 'DONE', label: 'Done' },
];

export function TaskDetailDrawer() {
  const dispatch = useAppDispatch();
  const taskDrawerId = useAppSelector((state) => state.ui.taskDrawerId);
  const activeWorkspaceId = useAppSelector((state) => state.auth.activeWorkspaceId);
  const currentUser = useAppSelector((state) => state.auth.user);

  const {
    data: task,
    isLoading: isTaskLoading,
    isError,
  } = useGetTaskQuery(taskDrawerId || '', {
    skip: !taskDrawerId,
  });

  const { data: comments = [], isLoading: isCommentsLoading } = useGetTaskCommentsQuery(
    taskDrawerId || '',
    { skip: !taskDrawerId }
  );

  const { data: workspaceMembers = [] } = useGetWorkspaceMembersQuery(
    activeWorkspaceId || '',
    { skip: !activeWorkspaceId }
  );

  const [updateTask, { isLoading: isUpdating }] = useUpdateTaskMutation();
  const [deleteTask, { isLoading: isDeleting }] = useDeleteTaskMutation();
  const [createComment, { isLoading: isPostingComment }] = useCreateTaskCommentMutation();
  const [deleteComment] = useDeleteTaskCommentMutation();
  const [updateAssignees] = useUpdateTaskAssigneesMutation();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [priority, setPriority] = useState<TaskPriority>('MEDIUM' as TaskPriority);
  const [statusCategory, setStatusCategory] = useState('TODO');
  const [commentText, setCommentText] = useState('');

  useEffect(() => {
    if (task) {
      setTitle(task.title || '');
      setDescription(task.description || '');
      setPriority((task.priority as TaskPriority) || ('MEDIUM' as TaskPriority));
      setStatusCategory(task.status?.category || 'TODO');
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
          title: title.trim() || task.title,
          description: description.trim(),
          priority,
        },
      }).unwrap();
    } catch {
      toast.error('Failed to update task');
    }
  };

  const handleStatusChange = async (newCategory: string) => {
    if (!task) return;
    setStatusCategory(newCategory);
    try {
      await updateTask({
        id: task.id,
        updates: {
          status: { ...task.status, category: newCategory as any } as any,
        },
      }).unwrap();
      toast.success(`Status updated to ${newCategory.replace('_', ' ')}`);
    } catch {
      toast.error('Failed to update status');
    }
  };

  const handleAssigneeChange = async (userId: string) => {
    if (!task) return;
    try {
      const assigneeIds = userId === 'unassigned' ? [] : [userId];
      await updateAssignees({ taskId: task.id, assigneeIds }).unwrap();
      toast.success(userId === 'unassigned' ? 'Task unassigned' : 'Assignee updated');
    } catch {
      toast.error('Failed to update assignee');
    }
  };

  const handleCommentSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commentText.trim() || !task) return;

    try {
      await createComment({
        taskId: task.id,
        content: commentText.trim(),
      }).unwrap();
      setCommentText('');
      toast.success('Comment added');
    } catch {
      toast.error('Failed to post comment');
    }
  };

  const handleDeleteComment = async (commentId: string) => {
    if (!task) return;
    try {
      await deleteComment({ id: commentId, taskId: task.id }).unwrap();
      toast.success('Comment deleted');
    } catch {
      toast.error('Failed to delete comment');
    }
  };

  const handleDeleteTask = async () => {
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

  const currentAssignee = task?.assignees?.[0]?.user;

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
                {task.status.name || statusCategory}
              </Badge>
            )}
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleDeleteTask}
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
        {isTaskLoading ? (
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

            {/* Properties Grid: Status, Priority, Assignee */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 p-3.5 rounded-md bg-zinc-950/60 border border-zinc-850 text-xs">
              {/* Status */}
              <div className="space-y-1">
                <span className="text-zinc-500 text-[11px] font-medium flex items-center gap-1">
                  <Layers className="h-3 w-3" />
                  <span>Status</span>
                </span>
                <select
                  value={statusCategory}
                  onChange={(e) => handleStatusChange(e.target.value)}
                  className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 rounded px-2 py-1.5 text-xs outline-none focus:border-zinc-700 cursor-pointer"
                >
                  {STATUS_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Priority */}
              <div className="space-y-1">
                <span className="text-zinc-500 text-[11px] font-medium">Priority</span>
                <select
                  value={priority}
                  onChange={(e) => {
                    const next = e.target.value as TaskPriority;
                    setPriority(next);
                    updateTask({ id: task.id, updates: { priority: next } });
                  }}
                  className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 rounded px-2 py-1.5 text-xs outline-none focus:border-zinc-700 cursor-pointer"
                >
                  <option value="LOW">Low</option>
                  <option value="MEDIUM">Medium</option>
                  <option value="HIGH">High</option>
                  <option value="URGENT">Urgent</option>
                </select>
              </div>

              {/* Assignee */}
              <div className="space-y-1">
                <span className="text-zinc-500 text-[11px] font-medium flex items-center gap-1">
                  <User className="h-3 w-3" />
                  <span>Assignee</span>
                </span>
                <select
                  value={currentAssignee?.id || 'unassigned'}
                  onChange={(e) => handleAssigneeChange(e.target.value)}
                  className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 rounded px-2 py-1.5 text-xs outline-none focus:border-zinc-700 cursor-pointer"
                >
                  <option value="unassigned">Unassigned</option>
                  {workspaceMembers.map((m) => {
                    const name =
                      m.user?.firstName && m.user?.lastName
                        ? `${m.user.firstName} ${m.user.lastName}`
                        : m.user?.username || m.user?.email || 'User';
                    return (
                      <option key={m.userId} value={m.userId}>
                        {name}
                      </option>
                    );
                  })}
                </select>
              </div>
            </div>

            {/* Description */}
            <div className="space-y-2">
              <span className="text-xs font-semibold text-zinc-400">Description</span>
              <Textarea
                rows={5}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                onBlur={handleSaveTitleAndDesc}
                placeholder="Add more details, acceptance criteria, or context..."
                className="bg-zinc-950/80 border-zinc-800 text-xs text-zinc-200 resize-none"
              />
            </div>

            {/* Comments Timeline */}
            <div className="space-y-3 pt-4 border-t border-zinc-800">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <MessageSquare className="h-4 w-4 text-zinc-400" />
                  <span className="text-xs font-bold text-zinc-300">Activity & Comments</span>
                </div>
                <span className="text-[10px] font-mono text-zinc-500">{comments.length} comments</span>
              </div>

              {/* Comments Feed */}
              <div className="space-y-3 max-h-60 overflow-y-auto pr-1">
                {isCommentsLoading ? (
                  <div className="py-4 text-center text-xs text-zinc-500">Loading discussion...</div>
                ) : comments.length === 0 ? (
                  <div className="py-4 text-center text-xs text-zinc-600">
                    No comments yet. Start the conversation below.
                  </div>
                ) : (
                  comments.map((c) => {
                    const authorName =
                      c.author?.firstName && c.author?.lastName
                        ? `${c.author.firstName} ${c.author.lastName}`
                        : c.author?.username || 'Team Member';
                    const initials =
                      ((c.author?.firstName?.[0] || '') + (c.author?.lastName?.[0] || '')).toUpperCase() || 'U';
                    const isOwnComment = currentUser?.id === c.authorId;

                    return (
                      <div
                        key={c.id}
                        className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1.5 text-xs"
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <Avatar className="h-5 w-5 rounded-sm border border-zinc-800">
                              <AvatarFallback className="text-[9px] bg-zinc-800 text-zinc-300">
                                {initials}
                              </AvatarFallback>
                            </Avatar>
                            <span className="font-semibold text-zinc-200 text-xs">{authorName}</span>
                            <span className="text-[10px] text-zinc-500">
                              {c.createdAt
                                ? new Date(c.createdAt).toLocaleDateString(undefined, {
                                    month: 'short',
                                    day: 'numeric',
                                    hour: '2-digit',
                                    minute: '2-digit',
                                  })
                                : ''}
                            </span>
                          </div>

                          {isOwnComment && (
                            <button
                              type="button"
                              onClick={() => handleDeleteComment(c.id)}
                              className="text-zinc-600 hover:text-rose-400 transition-colors p-1"
                              title="Delete comment"
                            >
                              <Trash2 className="h-3 w-3" />
                            </button>
                          )}
                        </div>

                        <p className="text-zinc-300 text-xs pl-7 whitespace-pre-wrap">{c.content}</p>
                      </div>
                    );
                  })
                )}
              </div>

              {/* Comment Input Box */}
              <form onSubmit={handleCommentSubmit} className="flex gap-2 pt-2">
                <Input
                  value={commentText}
                  onChange={(e) => setCommentText(e.target.value)}
                  placeholder="Write a comment or update..."
                  className="bg-zinc-950/80 border-zinc-800 text-xs text-zinc-200"
                />
                <Button
                  type="submit"
                  size="sm"
                  variant="primary"
                  isLoading={isPostingComment}
                  disabled={!commentText.trim()}
                  className="gap-1.5 text-xs shrink-0"
                >
                  <Send className="h-3.5 w-3.5" />
                  <span>Send</span>
                </Button>
              </form>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
