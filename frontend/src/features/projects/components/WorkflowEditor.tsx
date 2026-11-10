'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import {
  Layers,
  Plus,
  ChevronLeft,
  ChevronRight,
  Pencil,
  Trash2,
  Check,
  AlertCircle,
  HelpCircle,
} from 'lucide-react';
import { toast } from 'sonner';
import type { TaskStatus } from '@orbito/types';
import {
  useGetTaskStatusesByProjectQuery,
  useCreateTaskStatusMutation,
  useUpdateTaskStatusMutation,
  useUpdateTaskStatusPositionsMutation,
  useDeleteTaskStatusMutation,
} from '@/lib/store/api/tasksApi';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';

interface WorkflowEditorProps {
  projectId: string;
  workflowId?: string;
  canEdit: boolean;
}

const CATEGORY_CONFIG: Record<
  string,
  { label: string; badgeClass: string; desc: string }
> = {
  TODO: {
    label: 'To Do',
    badgeClass: 'bg-zinc-800 text-zinc-300 border-zinc-700',
    desc: 'Unstarted backlog & incoming sprint tasks',
  },
  IN_PROGRESS: {
    label: 'In Progress',
    badgeClass: 'bg-amber-950/60 text-amber-300 border-amber-800/60',
    desc: 'Active engineering and development in flight',
  },
  IN_REVIEW: {
    label: 'In Review',
    badgeClass: 'bg-sky-950/60 text-sky-300 border-sky-800/60',
    desc: 'Code review, design sign-off, or QA verification',
  },
  DONE: {
    label: 'Done',
    badgeClass: 'bg-emerald-950/60 text-emerald-300 border-emerald-800/60',
    desc: 'Completed and verified deliverables',
  },
};

const COLOR_PRESETS = [
  { name: 'Zinc', hex: '#71717a' },
  { name: 'Sky', hex: '#0284c7' },
  { name: 'Indigo', hex: '#6366f1' },
  { name: 'Violet', hex: '#8b5cf6' },
  { name: 'Emerald', hex: '#10b981' },
  { name: 'Amber', hex: '#f59e0b' },
  { name: 'Rose', hex: '#f43f5e' },
  { name: 'Cyan', hex: '#06b6d4' },
];

export function WorkflowEditor({
  projectId,
  workflowId: initialWorkflowId,
  canEdit,
}: WorkflowEditorProps) {
  const { data: rawStatuses = [], isLoading, error } =
    useGetTaskStatusesByProjectQuery(projectId, { skip: !projectId });

  const [createTaskStatus, { isLoading: isCreating }] =
    useCreateTaskStatusMutation();
  const [updateTaskStatus, { isLoading: isUpdating }] =
    useUpdateTaskStatusMutation();
  const [updatePositions, { isLoading: isReordering }] =
    useUpdateTaskStatusPositionsMutation();
  const [deleteTaskStatus, { isLoading: isDeleting }] =
    useDeleteTaskStatusMutation();

  // Local sorted statuses
  const statuses = useMemo(() => {
    return [...rawStatuses].sort((a, b) => a.position - b.position);
  }, [rawStatuses]);

  const resolvedWorkflowId =
    initialWorkflowId || statuses[0]?.workflowId || '';

  // Modal states
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingStatus, setEditingStatus] = useState<TaskStatus | null>(null);
  const [deletingStatus, setDeletingStatus] = useState<TaskStatus | null>(null);

  // Form states
  const [name, setName] = useState('');
  const [category, setCategory] = useState<string>('IN_PROGRESS');
  const [color, setColor] = useState('#0284c7');

  const openCreateModal = () => {
    setName('');
    setCategory('IN_PROGRESS');
    setColor('#0284c7');
    setIsCreateOpen(true);
  };

  const openEditModal = (status: TaskStatus) => {
    setEditingStatus(status);
    setName(status.name);
    setCategory(status.category);
    setColor(status.color || '#0284c7');
  };

  const handleCreateStatus = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error('Status name is required');
      return;
    }
    if (!resolvedWorkflowId) {
      toast.error('Unable to determine workflow for this project');
      return;
    }

    try {
      await createTaskStatus({
        name: name.trim(),
        category,
        color,
        workflowId: resolvedWorkflowId,
        projectId,
        position: statuses.length + 1,
      }).unwrap();

      toast.success(`Created status "${name.trim()}"`);
      setIsCreateOpen(false);
    } catch {
      toast.error('Failed to create task status');
    }
  };

  const handleUpdateStatus = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingStatus) return;
    if (!name.trim()) {
      toast.error('Status name is required');
      return;
    }

    try {
      await updateTaskStatus({
        id: editingStatus.id,
        projectId,
        updates: {
          name: name.trim(),
          category,
          color,
        },
      }).unwrap();

      toast.success('Status updated successfully');
      setEditingStatus(null);
    } catch {
      toast.error('Failed to update status');
    }
  };

  const handleMove = async (index: number, direction: 'left' | 'right') => {
    if (!canEdit || isReordering) return;
    const targetIndex = direction === 'left' ? index - 1 : index + 1;
    if (targetIndex < 0 || targetIndex >= statuses.length) return;

    const newStatuses = [...statuses];
    const temp = newStatuses[index];
    newStatuses[index] = newStatuses[targetIndex];
    newStatuses[targetIndex] = temp;

    const statusUpdates = newStatuses.map((s, idx) => ({
      id: s.id,
      position: idx + 1,
    }));

    try {
      await updatePositions({ statusUpdates, projectId }).unwrap();
      toast.success('Workflow order updated');
    } catch {
      toast.error('Failed to reorder statuses');
    }
  };

  const handleDeleteConfirm = async () => {
    if (!deletingStatus) return;
    try {
      await deleteTaskStatus({ id: deletingStatus.id, projectId }).unwrap();
      toast.success(`Removed status "${deletingStatus.name}"`);
      setDeletingStatus(null);
    } catch {
      toast.error('Failed to delete status. Ensure no tasks are in this stage.');
    }
  };

  return (
    <div className="space-y-6">
      <Card className="border-zinc-800 bg-zinc-900/80">
        <CardHeader className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-zinc-800/80">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Layers className="h-4 w-4 text-zinc-300" />
              <CardTitle className="text-base text-zinc-100">
                Kanban Workflow Pipeline
              </CardTitle>
              <Badge variant="outline" className="text-[10px] font-mono">
                {statuses.length} Stages
              </Badge>
            </div>
            <CardDescription className="text-xs text-zinc-400">
              Customize column stages, colors, and the left-to-right progression
              for tasks and sprints.
            </CardDescription>
          </div>

          {canEdit && (
            <Button
              type="button"
              variant="primary"
              size="sm"
              onClick={openCreateModal}
              className="gap-1.5 text-xs self-start sm:self-auto"
            >
              <Plus className="h-3.5 w-3.5" />
              <span>Add Stage Column</span>
            </Button>
          )}
        </CardHeader>

        <CardContent className="pt-6 space-y-6">
          {isLoading ? (
            <div className="py-12 text-center text-xs text-zinc-500">
              Loading workflow pipeline...
            </div>
          ) : error ? (
            <div className="py-8 text-center text-xs text-rose-400 flex items-center justify-center gap-2">
              <AlertCircle className="h-4 w-4" />
              <span>Failed to load statuses for this project.</span>
            </div>
          ) : statuses.length === 0 ? (
            <div className="py-12 text-center text-xs text-zinc-500">
              No workflow statuses configured yet.
            </div>
          ) : (
            <div className="space-y-3">
              {/* Pipeline summary timeline */}
              <div className="hidden md:flex items-center gap-2 p-3 rounded-md bg-zinc-950/70 border border-zinc-800/60 overflow-x-auto">
                {statuses.map((status, idx) => {
                  return (
                    <React.Fragment key={status.id}>
                      <div className="flex items-center gap-2 shrink-0">
                        <span
                          className="h-2 w-2 rounded-full ring-2 ring-zinc-800"
                          style={{ backgroundColor: status.color || '#71717a' }}
                        />
                        <span className="text-xs font-medium text-zinc-200">
                          {status.name}
                        </span>
                        <span className="text-[10px] font-mono text-zinc-500">
                          ({status._count?.tasks ?? 0})
                        </span>
                      </div>
                      {idx < statuses.length - 1 && (
                        <ChevronRight className="h-3.5 w-3.5 text-zinc-600 shrink-0" />
                      )}
                    </React.Fragment>
                  );
                })}
              </div>

              {/* Status cards list */}
              <div className="grid grid-cols-1 gap-2.5">
                {statuses.map((status, index) => {
                  const cat =
                    CATEGORY_CONFIG[status.category] || CATEGORY_CONFIG.TODO;
                  const taskCount = status._count?.tasks ?? 0;

                  return (
                    <div
                      key={status.id}
                      className="group flex flex-col sm:flex-row sm:items-center justify-between p-3.5 rounded-md bg-zinc-950/60 border border-zinc-850 hover:border-zinc-750 transition-all gap-3"
                    >
                      {/* Left info */}
                      <div className="flex items-center gap-3 min-w-0">
                        {/* Order pill */}
                        <span className="h-6 w-6 rounded bg-zinc-900 border border-zinc-800 flex items-center justify-center text-[10px] font-mono font-semibold text-zinc-400 shrink-0">
                          {index + 1}
                        </span>

                        {/* Status color indicator */}
                        <span
                          className="h-3.5 w-3.5 rounded-full ring-2 ring-zinc-800 shrink-0"
                          style={{
                            backgroundColor: status.color || '#71717a',
                          }}
                        />

                        {/* Title and metadata */}
                        <div className="flex flex-col min-w-0">
                          <div className="flex items-center gap-2 flex-wrap">
                            <span className="text-xs font-semibold text-zinc-100 truncate">
                              {status.name}
                            </span>
                            <Badge
                              variant="outline"
                              className={`text-[9px] uppercase px-1.5 py-0 font-mono ${cat.badgeClass}`}
                            >
                              {cat.label}
                            </Badge>
                            {status.isDefault && (
                              <Badge
                                variant="outline"
                                className="text-[9px] text-zinc-400 border-zinc-800"
                              >
                                Default
                              </Badge>
                            )}
                          </div>
                          <span className="text-[11px] text-zinc-500 truncate">
                            {cat.desc}
                          </span>
                        </div>
                      </div>

                      {/* Right controls */}
                      <div className="flex items-center gap-2 self-end sm:self-auto shrink-0">
                        {/* Task count pill */}
                        <span
                          className="text-[11px] font-mono px-2 py-0.5 rounded bg-zinc-900 text-zinc-400 border border-zinc-800"
                          title="Tasks currently in this stage"
                        >
                          {taskCount} {taskCount === 1 ? 'task' : 'tasks'}
                        </span>

                        {canEdit && (
                          <div className="flex items-center gap-1 border-l border-zinc-800 pl-2">
                            {/* Move left */}
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              disabled={index === 0 || isReordering}
                              onClick={() => handleMove(index, 'left')}
                              className="h-7 w-7 p-0 text-zinc-400 hover:text-zinc-100 disabled:opacity-30"
                              title="Move left in pipeline"
                            >
                              <ChevronLeft className="h-3.5 w-3.5" />
                            </Button>

                            {/* Move right */}
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              disabled={
                                index === statuses.length - 1 || isReordering
                              }
                              onClick={() => handleMove(index, 'right')}
                              className="h-7 w-7 p-0 text-zinc-400 hover:text-zinc-100 disabled:opacity-30"
                              title="Move right in pipeline"
                            >
                              <ChevronRight className="h-3.5 w-3.5" />
                            </Button>

                            {/* Edit */}
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              onClick={() => openEditModal(status)}
                              className="h-7 w-7 p-0 text-zinc-400 hover:text-zinc-100"
                              title="Edit stage settings"
                            >
                              <Pencil className="h-3.5 w-3.5" />
                            </Button>

                            {/* Delete */}
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              disabled={taskCount > 0 || isDeleting}
                              onClick={() => setDeletingStatus(status)}
                              className="h-7 w-7 p-0 text-zinc-500 hover:text-rose-400 disabled:opacity-30"
                              title={
                                taskCount > 0
                                  ? 'Cannot delete: Move existing tasks first'
                                  : 'Delete stage'
                              }
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Status Dialog */}
      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
        <DialogContent className="bg-zinc-900 border-zinc-800 max-w-md">
          <DialogHeader>
            <DialogTitle className="text-base text-zinc-100">
              Add Workflow Stage Column
            </DialogTitle>
            <DialogDescription className="text-xs text-zinc-400">
              Create a new status column for tasks in this project.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleCreateStatus} className="space-y-4 pt-2">
            <div className="space-y-1.5">
              <Label htmlFor="status-name" className="text-xs text-zinc-300">
                Stage Name
              </Label>
              <Input
                id="status-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. QA Verification, Blocked, Ready for Release"
                className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                autoFocus
              />
            </div>

            <div className="space-y-1.5">
              <Label className="text-xs text-zinc-300">Category</Label>
              <div className="grid grid-cols-2 gap-2">
                {Object.entries(CATEGORY_CONFIG).map(([key, config]) => {
                  const isSelected = category === key;
                  return (
                    <button
                      key={key}
                      type="button"
                      onClick={() => setCategory(key)}
                      className={`p-2.5 rounded-md border text-left transition-all ${
                        isSelected
                          ? 'bg-zinc-800/80 border-zinc-600 text-zinc-100'
                          : 'bg-zinc-950/60 border-zinc-850 text-zinc-400 hover:text-zinc-200'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-medium">{config.label}</span>
                        {isSelected && <Check className="h-3 w-3 text-emerald-400" />}
                      </div>
                      <p className="text-[10px] text-zinc-500 mt-0.5 truncate">
                        {config.desc}
                      </p>
                    </button>
                  );
                })}
              </div>
            </div>

            <div className="space-y-1.5">
              <Label className="text-xs text-zinc-300">Accent Color</Label>
              <div className="flex items-center gap-2 flex-wrap">
                {COLOR_PRESETS.map((preset) => (
                  <button
                    key={preset.hex}
                    type="button"
                    onClick={() => setColor(preset.hex)}
                    className="h-6 w-6 rounded-full border border-zinc-700 flex items-center justify-center transition-transform hover:scale-110"
                    style={{ backgroundColor: preset.hex }}
                    title={preset.name}
                  >
                    {color.toLowerCase() === preset.hex.toLowerCase() && (
                      <Check className="h-3 w-3 text-white drop-shadow" />
                    )}
                  </button>
                ))}
                <Input
                  value={color}
                  onChange={(e) => setColor(e.target.value)}
                  className="w-24 h-7 text-[11px] font-mono bg-zinc-950 border-zinc-800 text-zinc-200 px-2"
                  placeholder="#0284c7"
                />
              </div>
            </div>

            <DialogFooter className="pt-3 gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setIsCreateOpen(false)}
                className="text-xs"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="primary"
                size="sm"
                isLoading={isCreating}
                className="text-xs"
              >
                Create Stage
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Edit Status Dialog */}
      <Dialog
        open={Boolean(editingStatus)}
        onOpenChange={(open) => !open && setEditingStatus(null)}
      >
        <DialogContent className="bg-zinc-900 border-zinc-800 max-w-md">
          <DialogHeader>
            <DialogTitle className="text-base text-zinc-100">
              Edit Stage Column
            </DialogTitle>
            <DialogDescription className="text-xs text-zinc-400">
              Modify the title, category, or color for this workflow stage.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleUpdateStatus} className="space-y-4 pt-2">
            <div className="space-y-1.5">
              <Label htmlFor="edit-status-name" className="text-xs text-zinc-300">
                Stage Name
              </Label>
              <Input
                id="edit-status-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
              />
            </div>

            <div className="space-y-1.5">
              <Label className="text-xs text-zinc-300">Category</Label>
              <div className="grid grid-cols-2 gap-2">
                {Object.entries(CATEGORY_CONFIG).map(([key, config]) => {
                  const isSelected = category === key;
                  return (
                    <button
                      key={key}
                      type="button"
                      onClick={() => setCategory(key)}
                      className={`p-2.5 rounded-md border text-left transition-all ${
                        isSelected
                          ? 'bg-zinc-800/80 border-zinc-600 text-zinc-100'
                          : 'bg-zinc-950/60 border-zinc-850 text-zinc-400 hover:text-zinc-200'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-medium">{config.label}</span>
                        {isSelected && <Check className="h-3 w-3 text-emerald-400" />}
                      </div>
                      <p className="text-[10px] text-zinc-500 mt-0.5 truncate">
                        {config.desc}
                      </p>
                    </button>
                  );
                })}
              </div>
            </div>

            <div className="space-y-1.5">
              <Label className="text-xs text-zinc-300">Accent Color</Label>
              <div className="flex items-center gap-2 flex-wrap">
                {COLOR_PRESETS.map((preset) => (
                  <button
                    key={preset.hex}
                    type="button"
                    onClick={() => setColor(preset.hex)}
                    className="h-6 w-6 rounded-full border border-zinc-700 flex items-center justify-center transition-transform hover:scale-110"
                    style={{ backgroundColor: preset.hex }}
                    title={preset.name}
                  >
                    {color.toLowerCase() === preset.hex.toLowerCase() && (
                      <Check className="h-3 w-3 text-white drop-shadow" />
                    )}
                  </button>
                ))}
                <Input
                  value={color}
                  onChange={(e) => setColor(e.target.value)}
                  className="w-24 h-7 text-[11px] font-mono bg-zinc-950 border-zinc-800 text-zinc-200 px-2"
                />
              </div>
            </div>

            <DialogFooter className="pt-3 gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setEditingStatus(null)}
                className="text-xs"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="primary"
                size="sm"
                isLoading={isUpdating}
                className="text-xs"
              >
                Save Changes
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Status Confirmation Dialog */}
      <Dialog
        open={Boolean(deletingStatus)}
        onOpenChange={(open) => !open && setDeletingStatus(null)}
      >
        <DialogContent className="bg-zinc-900 border-zinc-800 max-w-sm">
          <DialogHeader>
            <DialogTitle className="text-base text-zinc-100 flex items-center gap-2">
              <AlertCircle className="h-4 w-4 text-rose-400" />
              <span>Delete Stage</span>
            </DialogTitle>
            <DialogDescription className="text-xs text-zinc-400">
              Are you sure you want to remove the stage column{' '}
              <strong className="text-zinc-200">
                "{deletingStatus?.name}"
              </strong>
              ?
            </DialogDescription>
          </DialogHeader>

          <DialogFooter className="pt-3 gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setDeletingStatus(null)}
              className="text-xs"
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="destructive"
              size="sm"
              isLoading={isDeleting}
              onClick={handleDeleteConfirm}
              className="text-xs"
            >
              Delete Stage
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default WorkflowEditor;

