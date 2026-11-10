'use client';

import * as React from 'react';
import { useState } from 'react';
import { CheckSquare, Square, Plus, Trash2, CheckCircle2 } from 'lucide-react';
import { toast } from 'sonner';
import {
  useGetSubtasksQuery,
  useCreateSubtaskMutation,
  useUpdateTaskMutation,
  useDeleteTaskMutation,
} from '@/lib/store/api/tasksApi';
import type { TaskSubtasksProps } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export const TaskSubtasks = React.memo(function TaskSubtasks({ taskId, projectId, canEdit }: TaskSubtasksProps) {
  const { data: subtasks = [], isLoading } = useGetSubtasksQuery(taskId, {
    skip: !taskId,
  });

  const [createSubtask, { isLoading: isCreating }] = useCreateSubtaskMutation();
  const [updateTask] = useUpdateTaskMutation();
  const [deleteTask] = useDeleteTaskMutation();

  const [newTitle, setNewTitle] = useState('');
  const [isAdding, setIsAdding] = useState(false);

  const completedCount = subtasks.filter((s) => s.status?.category === 'DONE').length;
  const progressPercent = subtasks.length > 0 ? Math.round((completedCount / subtasks.length) * 100) : 0;

  const handleToggleSubtask = async (subtaskId: string, currentCategory: string) => {
    if (!canEdit) return;
    const nextCategory = currentCategory === 'DONE' ? 'TODO' : 'DONE';
    try {
      await updateTask({
        id: subtaskId,
        updates: {
          status: { category: nextCategory as any } as any,
        },
      }).unwrap();
      toast.success(nextCategory === 'DONE' ? 'Subtask completed' : 'Subtask reopened');
    } catch {
      toast.error('Failed to update subtask');
    }
  };

  const handleAddSubtask = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTitle.trim() || !projectId || !taskId) return;

    try {
      await createSubtask({
        parentTaskId: taskId,
        projectId,
        title: newTitle.trim(),
      }).unwrap();
      setNewTitle('');
      setIsAdding(false);
      toast.success('Subtask added');
    } catch {
      toast.error('Failed to add subtask');
    }
  };

  const handleDeleteSubtask = async (subtaskId: string) => {
    if (!canEdit) return;
    try {
      await deleteTask(subtaskId).unwrap();
      toast.success('Subtask deleted');
    } catch {
      toast.error('Failed to delete subtask');
    }
  };

  return (
    <div className="space-y-3 pt-4 border-t border-zinc-800">
      {/* Header & Progress */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <CheckSquare className="h-4 w-4 text-zinc-400" />
          <span className="text-xs font-bold text-zinc-300">Subtasks</span>
        </div>
        {subtasks.length > 0 && (
          <span className="text-[11px] font-mono text-zinc-400">
            {completedCount}/{subtasks.length} ({progressPercent}%)
          </span>
        )}
      </div>

      {/* Progress Bar */}
      {subtasks.length > 0 && (
        <div className="w-full bg-zinc-850 h-1.5 rounded-full overflow-hidden">
          <div
            className="bg-emerald-500 h-full transition-all duration-300"
            style={{ width: `${progressPercent}%` }}
          />
        </div>
      )}

      {/* Subtasks List */}
      <div className="space-y-1.5">
        {isLoading ? (
          <div className="py-2 text-xs text-zinc-500 text-center">Loading subtasks...</div>
        ) : subtasks.length === 0 && !isAdding ? (
          <div className="text-[11px] text-zinc-600 italic">No subtasks created.</div>
        ) : (
          subtasks.map((subtask) => {
            const isDone = subtask.status?.category === 'DONE';

            return (
              <div
                key={subtask.id}
                className="flex items-center justify-between p-2 rounded-md bg-zinc-950/40 hover:bg-zinc-950 border border-zinc-850 group transition-colors"
              >
                <div className="flex items-center gap-2.5 min-w-0">
                  <button
                    type="button"
                    disabled={!canEdit}
                    onClick={() => handleToggleSubtask(subtask.id, subtask.status?.category || 'TODO')}
                    className="text-zinc-500 hover:text-zinc-200 transition-colors disabled:cursor-not-allowed"
                  >
                    {isDone ? (
                      <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                    ) : (
                      <Square className="h-4 w-4 text-zinc-500 hover:text-zinc-300" />
                    )}
                  </button>
                  <span
                    className={`text-xs truncate ${
                      isDone ? 'line-through text-zinc-500' : 'text-zinc-200'
                    }`}
                  >
                    {subtask.title}
                  </span>
                </div>

                {canEdit && (
                  <button
                    type="button"
                    onClick={() => handleDeleteSubtask(subtask.id)}
                    className="opacity-0 group-hover:opacity-100 p-1 text-zinc-500 hover:text-rose-400 transition-all cursor-pointer"
                    title="Delete subtask"
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                )}
              </div>
            );
          })
        )}
      </div>

      {/* Add Subtask Form / Trigger */}
      {canEdit && (
        <div>
          {isAdding ? (
            <form onSubmit={handleAddSubtask} className="flex gap-2 pt-1">
              <Input
                value={newTitle}
                onChange={(e) => setNewTitle(e.target.value)}
                placeholder="Subtask title..."
                autoFocus
                className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100 h-8"
              />
              <Button
                type="submit"
                size="sm"
                variant="primary"
                isLoading={isCreating}
                disabled={!newTitle.trim()}
                className="text-xs h-8"
              >
                Add
              </Button>
              <Button
                type="button"
                size="sm"
                variant="secondary"
                onClick={() => {
                  setIsAdding(false);
                  setNewTitle('');
                }}
                className="text-xs h-8"
              >
                Cancel
              </Button>
            </form>
          ) : (
            <button
              type="button"
              onClick={() => setIsAdding(true)}
              className="flex items-center gap-1.5 text-xs text-zinc-400 hover:text-zinc-200 cursor-pointer pt-1 transition-colors"
            >
              <Plus className="h-3.5 w-3.5 text-zinc-500" />
              <span>Add Subtask</span>
            </button>
          )}
        </div>
      )}
    </div>
  );
});

