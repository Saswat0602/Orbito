'use client';

import React, { useState, useMemo } from 'react';
import {
  Link2,
  ShieldAlert,
  ArrowRight,
  Trash2,
  Plus,
  Loader2,
  CheckCircle2,
  AlertCircle,
} from 'lucide-react';
import {
  useGetTaskDependenciesQuery,
  useCreateTaskDependencyMutation,
  useDeleteTaskDependencyMutation,
  useGetTasksQuery,
} from '@/lib/store/api/tasksApi';
import type { TaskDependenciesProps } from '@/types';
import { toast } from 'sonner';

export const TaskDependencies: React.FC<TaskDependenciesProps> = ({
  taskId,
  projectId,
  canEdit,
}) => {
  const [isAdding, setIsAdding] = useState(false);
  const [relationType, setRelationType] = useState<'IS_BLOCKED_BY' | 'BLOCKS'>('IS_BLOCKED_BY');
  const [selectedTargetTaskId, setSelectedTargetTaskId] = useState('');

  // Fetch dependencies for this task
  const { data: dependencies, isLoading: isDepsLoading } = useGetTaskDependenciesQuery(taskId);

  // Fetch all tasks in project to populate selector
  const { data: projectTasks = [] } = useGetTasksQuery(
    projectId ? { projectId } : undefined,
    { skip: !projectId || !isAdding }
  );

  const [createDependency, { isLoading: isCreating }] = useCreateTaskDependencyMutation();
  const [deleteDependency, { isLoading: isDeleting }] = useDeleteTaskDependencyMutation();

  const dependsOn = dependencies?.dependsOn || [];
  const blocks = dependencies?.blocks || [];

  // Exclude current task and tasks already in dependencies
  const availableTargetTasks = useMemo(() => {
    const linkedIds = new Set<string>([
      taskId,
      ...dependsOn.map((d) => d.blockingTaskId),
      ...blocks.map((b) => b.dependentTaskId),
    ]);
    return projectTasks.filter((t) => !linkedIds.has(t.id));
  }, [projectTasks, taskId, dependsOn, blocks]);

  const handleAddDependency = async () => {
    if (!selectedTargetTaskId) {
      toast.error('Please select a task to link');
      return;
    }

    try {
      if (relationType === 'IS_BLOCKED_BY') {
        // Current task is blocked by selected task
        await createDependency({
          dependentTaskId: taskId,
          blockingTaskId: selectedTargetTaskId,
          type: 'BLOCKS',
        }).unwrap();
      } else {
        // Current task blocks selected task
        await createDependency({
          dependentTaskId: selectedTargetTaskId,
          blockingTaskId: taskId,
          type: 'BLOCKS',
        }).unwrap();
      }

      toast.success('Task dependency linked');
      setSelectedTargetTaskId('');
      setIsAdding(false);
    } catch (err: unknown) {
      const errorMsg =
        (err as { data?: { message?: string } })?.data?.message ||
        'Failed to link dependency. Check for circular relationships.';
      toast.error(errorMsg);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteDependency({ id, taskId }).unwrap();
      toast.success('Dependency removed');
    } catch {
      toast.error('Failed to remove dependency');
    }
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Link2 className="h-4 w-4 text-zinc-400" />
          <h4 className="text-xs font-semibold uppercase tracking-wider text-zinc-400">
            Dependencies
          </h4>
          {(dependsOn.length > 0 || blocks.length > 0) && (
            <span className="rounded-full bg-zinc-800 px-1.5 py-0.2 text-[10px] font-medium text-zinc-300">
              {dependsOn.length + blocks.length}
            </span>
          )}
        </div>

        {canEdit && !isAdding && (
          <button
            type="button"
            onClick={() => setIsAdding(true)}
            className="flex items-center gap-1 text-xs text-zinc-400 transition-colors hover:text-zinc-200"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>Add</span>
          </button>
        )}
      </div>

      {/* Add Dependency Panel */}
      {isAdding && (
        <div className="rounded-md border border-zinc-800 bg-zinc-900/80 p-3 space-y-2.5 text-xs">
          <div className="flex items-center justify-between">
            <span className="font-medium text-zinc-200">Link Task Dependency</span>
            <button
              type="button"
              onClick={() => {
                setIsAdding(false);
                setSelectedTargetTaskId('');
              }}
              className="text-zinc-500 hover:text-zinc-300"
            >
              Cancel
            </button>
          </div>

          <div className="space-y-1.5">
            <label className="text-[11px] text-zinc-400">Relationship Type</label>
            <select
              value={relationType}
              onChange={(e) => setRelationType(e.target.value as 'IS_BLOCKED_BY' | 'BLOCKS')}
              className="w-full rounded border border-zinc-700 bg-zinc-950 px-2.5 py-1.5 text-xs text-zinc-200 focus:border-zinc-500 focus:outline-none"
            >
              <option value="IS_BLOCKED_BY">This task is blocked by...</option>
              <option value="BLOCKS">This task blocks...</option>
            </select>
          </div>

          <div className="space-y-1.5">
            <label className="text-[11px] text-zinc-400">Select Task</label>
            <select
              value={selectedTargetTaskId}
              onChange={(e) => setSelectedTargetTaskId(e.target.value)}
              className="w-full rounded border border-zinc-700 bg-zinc-950 px-2.5 py-1.5 text-xs text-zinc-200 focus:border-zinc-500 focus:outline-none"
            >
              <option value="">-- Choose task from project --</option>
              {availableTargetTasks.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.slug ? `[${t.slug}] ` : ''}{t.title}
                </option>
              ))}
            </select>
          </div>

          <button
            type="button"
            disabled={!selectedTargetTaskId || isCreating}
            onClick={handleAddDependency}
            className="flex w-full items-center justify-center gap-1.5 rounded bg-zinc-100 py-1.5 font-medium text-zinc-950 transition hover:bg-white disabled:opacity-50"
          >
            {isCreating ? (
              <>
                <Loader2 className="h-3 w-3 animate-spin" />
                <span>Linking...</span>
              </>
            ) : (
              <span>Link Dependency</span>
            )}
          </button>
        </div>
      )}

      {/* Dependency Lists */}
      {isDepsLoading ? (
        <div className="flex items-center gap-2 py-2 text-xs text-zinc-500">
          <Loader2 className="h-3.5 w-3.5 animate-spin" />
          <span>Loading dependencies...</span>
        </div>
      ) : dependsOn.length === 0 && blocks.length === 0 ? (
        <p className="text-xs text-zinc-500 italic">No dependencies linked.</p>
      ) : (
        <div className="space-y-3">
          {/* Blocked By List */}
          {dependsOn.length > 0 && (
            <div className="space-y-1.5">
              <span className="flex items-center gap-1 text-[11px] font-medium text-amber-400">
                <AlertCircle className="h-3 w-3" />
                Blocked By (Prerequisites)
              </span>
              <div className="space-y-1">
                {dependsOn.map((dep) => {
                  const target = dep.blockingTask;
                  const isDone = target?.status?.name?.toLowerCase().includes('done');
                  return (
                    <div
                      key={dep.id}
                      className="flex items-center justify-between rounded border border-zinc-800 bg-zinc-900/50 px-2.5 py-1.5 text-xs"
                    >
                      <div className="flex items-center gap-2 truncate">
                        {isDone ? (
                          <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400 shrink-0" />
                        ) : (
                          <ShieldAlert className="h-3.5 w-3.5 text-amber-400 shrink-0" />
                        )}
                        <span className="font-mono text-[11px] text-zinc-400 shrink-0">
                          {target?.slug || 'TASK'}
                        </span>
                        <span className="truncate text-zinc-200" title={target?.title}>
                          {target?.title || 'Unknown Task'}
                        </span>
                        {target?.status && (
                          <span className="rounded bg-zinc-800 px-1.5 py-0.5 text-[10px] text-zinc-400 shrink-0">
                            {target.status.name}
                          </span>
                        )}
                      </div>

                      {canEdit && (
                        <button
                          type="button"
                          onClick={() => handleDelete(dep.id)}
                          disabled={isDeleting}
                          className="ml-2 text-zinc-500 transition-colors hover:text-red-400"
                          title="Remove dependency"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Blocks List */}
          {blocks.length > 0 && (
            <div className="space-y-1.5">
              <span className="flex items-center gap-1 text-[11px] font-medium text-blue-400">
                <ArrowRight className="h-3 w-3" />
                Blocks (Dependent Tasks)
              </span>
              <div className="space-y-1">
                {blocks.map((dep) => {
                  const target = dep.dependentTask;
                  return (
                    <div
                      key={dep.id}
                      className="flex items-center justify-between rounded border border-zinc-800 bg-zinc-900/50 px-2.5 py-1.5 text-xs"
                    >
                      <div className="flex items-center gap-2 truncate">
                        <span className="font-mono text-[11px] text-zinc-400 shrink-0">
                          {target?.slug || 'TASK'}
                        </span>
                        <span className="truncate text-zinc-200" title={target?.title}>
                          {target?.title || 'Unknown Task'}
                        </span>
                        {target?.status && (
                          <span className="rounded bg-zinc-800 px-1.5 py-0.5 text-[10px] text-zinc-400 shrink-0">
                            {target.status.name}
                          </span>
                        )}
                      </div>

                      {canEdit && (
                        <button
                          type="button"
                          onClick={() => handleDelete(dep.id)}
                          disabled={isDeleting}
                          className="ml-2 text-zinc-500 transition-colors hover:text-red-400"
                          title="Remove dependency"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

