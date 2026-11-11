'use client';

import * as React from 'react';
import { useState } from 'react';
import {
  GitFork,
  CheckCircle2,
  Circle,
  Plus,
  Trash2,
  ArrowUpRight,
  CornerDownRight,
  ExternalLink,
} from 'lucide-react';
import { toast } from 'sonner';
import type { Task } from '@orbito/types';
import {
  useGetSubtasksQuery,
  useCreateSubtaskMutation,
  useUpdateTaskMutation,
  useDeleteTaskMutation,
  useGetTaskStatusesByProjectQuery,
} from '@/lib/store/api/tasksApi';
import { useAppDispatch } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

interface TaskHierarchyProps {
  task: Task;
  canEdit: boolean;
  onNavigateToTask?: (taskId: string) => void;
}

export function TaskHierarchy({
  task,
  canEdit,
  onNavigateToTask,
}: TaskHierarchyProps) {
  const dispatch = useAppDispatch();
  const [newSubtaskTitle, setNewSubtaskTitle] = useState('');

  const { data: subtasks = [], isLoading: isSubtasksLoading } =
    useGetSubtasksQuery(task.id, { skip: !task.id });

  const { data: projectStatuses = [] } = useGetTaskStatusesByProjectQuery(
    task.projectId,
    { skip: !task.projectId }
  );

  const [createSubtask, { isLoading: isCreating }] =
    useCreateSubtaskMutation();
  const [updateTask] = useUpdateTaskMutation();
  const [deleteTask] = useDeleteTaskMutation();

  const doneStatus = projectStatuses.find((s) => s.category === 'DONE');
  const todoStatus =
    projectStatuses.find((s) => s.category === 'TODO') || projectStatuses[0];

  const totalSubtasks = subtasks.length;
  const completedSubtasks = subtasks.filter(
    (s) => s.status?.category === 'DONE'
  ).length;
  const completionPercentage =
    totalSubtasks > 0
      ? Math.round((completedSubtasks / totalSubtasks) * 100)
      : 0;

  const handleAddSubtask = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newSubtaskTitle.trim() || !canEdit) return;

    try {
      await createSubtask({
        parentTaskId: task.id,
        title: newSubtaskTitle.trim(),
        projectId: task.projectId,
      }).unwrap();

      setNewSubtaskTitle('');
      toast.success('Subtask added');
    } catch {
      toast.error('Failed to add subtask');
    }
  };

  const handleToggleSubtask = async (subtask: Task) => {
    if (!canEdit) return;
    const isCurrentlyDone = subtask.status?.category === 'DONE';
    const targetStatus = isCurrentlyDone ? todoStatus : doneStatus;
    if (!targetStatus) return;

    try {
      await updateTask({
        id: subtask.id,
        updates: { statusId: targetStatus.id },
      }).unwrap();
    } catch {
      toast.error('Failed to update subtask status');
    }
  };

  const handlePromoteToParent = async (subtaskId: string, title: string) => {
    if (!canEdit) return;
    try {
      await updateTask({
        id: subtaskId,
        updates: { parentTaskId: null },
      }).unwrap();
      toast.success(`Promoted "${title}" to top-level task`);
    } catch {
      toast.error('Failed to promote subtask');
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

  const handleOpenSubtask = (subtaskId: string) => {
    if (onNavigateToTask) {
      onNavigateToTask(subtaskId);
    } else {
      dispatch(setTaskDrawerId(subtaskId));
    }
  };

  return (
    <div className="space-y-3 pt-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <GitFork className="h-4 w-4 text-zinc-400" />
          <h4 className="text-xs font-semibold text-zinc-200">
            Subtask Hierarchy
          </h4>
          <span className="text-[10px] font-mono text-zinc-500">
            ({completedSubtasks}/{totalSubtasks})
          </span>
        </div>

        {totalSubtasks > 0 && (
          <span className="text-[11px] font-mono text-zinc-400">
            {completionPercentage}% complete
          </span>
        )}
      </div>

      {/* Progress Bar */}
      {totalSubtasks > 0 && (
        <div className="w-full h-1.5 rounded-full bg-zinc-800 overflow-hidden">
          <div
            className="h-full bg-emerald-500 transition-all duration-300"
            style={{ width: `${completionPercentage}%` }}
          />
        </div>
      )}

      {/* Subtasks List */}
      <div className="space-y-1.5">
        {isSubtasksLoading ? (
          <div className="py-2 text-[11px] text-zinc-500">Loading subtasks...</div>
        ) : subtasks.length === 0 ? (
          <div className="py-2 text-[11px] text-zinc-500 italic">
            No subtasks attached to this item.
          </div>
        ) : (
          <div className="divide-y divide-zinc-800/60 border border-zinc-800/80 rounded-md overflow-hidden bg-zinc-950/40">
            {subtasks.map((st) => {
              const isDone = st.status?.category === 'DONE';
              const firstAssignee = st.assignees?.[0]?.user;
              const initials = firstAssignee
                ? (
                    (firstAssignee.firstName?.[0] || '') +
                    (firstAssignee.lastName?.[0] || '')
                  ).toUpperCase() || 'U'
                : null;

              return (
                <div
                  key={st.id}
                  className="flex items-center justify-between p-2.5 hover:bg-zinc-900/40 transition-colors group"
                >
                  <div className="flex items-center gap-2.5 min-w-0">
                    <button
                      type="button"
                      disabled={!canEdit}
                      onClick={() => handleToggleSubtask(st)}
                      className="text-zinc-500 hover:text-zinc-300 transition-colors shrink-0"
                    >
                      {isDone ? (
                        <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                      ) : (
                        <Circle className="h-4 w-4 text-zinc-600 hover:text-zinc-400" />
                      )}
                    </button>

                    <button
                      type="button"
                      onClick={() => handleOpenSubtask(st.id)}
                      className={`text-xs text-left truncate transition-colors hover:text-zinc-100 ${
                        isDone ? 'line-through text-zinc-500' : 'text-zinc-200'
                      }`}
                      title={st.title}
                    >
                      {st.title}
                    </button>
                  </div>

                  <div className="flex items-center gap-2 shrink-0">
                    {firstAssignee && (
                      <Avatar className="h-5 w-5 rounded-full border border-zinc-800">
                        <AvatarFallback className="text-[9px] bg-zinc-800 text-zinc-300">
                          {initials}
                        </AvatarFallback>
                      </Avatar>
                    )}

                    {st.priority && (
                      <Badge
                        variant="outline"
                        className="text-[9px] uppercase px-1 py-0 font-mono text-zinc-400 border-zinc-800"
                      >
                        {st.priority}
                      </Badge>
                    )}

                    {canEdit && (
                      <div className="opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-1">
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() =>
                            handlePromoteToParent(st.id, st.title)
                          }
                          className="h-6 w-6 p-0 text-zinc-500 hover:text-zinc-200"
                          title="Promote to top-level task"
                        >
                          <ArrowUpRight className="h-3.5 w-3.5" />
                        </Button>

                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDeleteSubtask(st.id)}
                          className="h-6 w-6 p-0 text-zinc-500 hover:text-rose-400"
                          title="Delete subtask"
                        >
                          <Trash2 className="h-3 w-3" />
                        </Button>
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}

        {/* Add Subtask Input */}
        {canEdit && (
          <form onSubmit={handleAddSubtask} className="flex items-center gap-2 pt-1">
            <CornerDownRight className="h-3.5 w-3.5 text-zinc-500 ml-1 shrink-0" />
            <Input
              value={newSubtaskTitle}
              onChange={(e) => setNewSubtaskTitle(e.target.value)}
              placeholder="Add a new subtask..."
              className="h-7 text-xs bg-zinc-950 border-zinc-800 text-zinc-200 placeholder:text-zinc-600"
            />
            <Button
              type="submit"
              variant="outline"
              size="sm"
              disabled={!newSubtaskTitle.trim() || isCreating}
              className="h-7 px-2.5 text-xs border-zinc-800 bg-zinc-900/60 hover:bg-zinc-800 text-zinc-200 shrink-0"
            >
              <Plus className="h-3 w-3 mr-1" />
              <span>Add</span>
            </Button>
          </form>
        )}
      </div>
    </div>
  );
}

export default TaskHierarchy;
