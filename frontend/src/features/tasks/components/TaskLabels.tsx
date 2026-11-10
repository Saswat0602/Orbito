'use client';

import * as React from 'react';
import { useState } from 'react';
import { Tag, Plus, X, Check } from 'lucide-react';
import { toast } from 'sonner';
import {
  useGetTaskLabelsQuery,
  useGetAllLabelsQuery,
  useAssignTaskLabelMutation,
  useRemoveTaskLabelMutation,
  useCreateLabelMutation,
} from '@/lib/store/api/tasksApi';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

interface TaskLabelsProps {
  taskId: string;
  canEdit: boolean;
}

const PRESET_COLORS = [
  '#ef4444', // Red
  '#f97316', // Orange
  '#eab308', // Yellow
  '#10b981', // Emerald
  '#06b6d4', // Cyan
  '#3b82f6', // Blue
  '#8b5cf6', // Violet
  '#ec4899', // Pink
  '#71717a', // Zinc
];

export const TaskLabels = React.memo(function TaskLabels({ taskId, canEdit }: TaskLabelsProps) {
  const { data: taskLabels = [], isLoading: isTaskLabelsLoading } = useGetTaskLabelsQuery(taskId, {
    skip: !taskId,
  });

  const { data: allLabels = [] } = useGetAllLabelsQuery();

  const [assignLabel] = useAssignTaskLabelMutation();
  const [removeLabel] = useRemoveTaskLabelMutation();
  const [createLabel, { isLoading: isCreating }] = useCreateLabelMutation();

  const [isPopoverOpen, setIsPopoverOpen] = useState(false);
  const [newLabelName, setNewLabelName] = useState('');
  const [newLabelColor, setNewLabelColor] = useState(PRESET_COLORS[0]);

  const assignedLabelIds = new Set(taskLabels.map((l) => l.id));
  const unassignedLabels = allLabels.filter((l) => !assignedLabelIds.has(l.id));

  const handleAssign = async (labelId: string) => {
    try {
      await assignLabel({ taskId, labelId }).unwrap();
      toast.success('Label added');
    } catch {
      toast.error('Failed to add label');
    }
  };

  const handleRemove = async (labelId: string) => {
    if (!canEdit) return;
    try {
      await removeLabel({ taskId, labelId }).unwrap();
      toast.success('Label removed');
    } catch {
      toast.error('Failed to remove label');
    }
  };

  const handleCreateAndAssign = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newLabelName.trim()) return;

    try {
      const created = await createLabel({
        name: newLabelName.trim(),
        color: newLabelColor,
      }).unwrap();

      await assignLabel({ taskId, labelId: created.id }).unwrap();
      setNewLabelName('');
      toast.success('Label created and added');
      setIsPopoverOpen(false);
    } catch {
      toast.error('Failed to create label');
    }
  };

  return (
    <div className="space-y-2 pt-4 border-t border-zinc-800">
      <div className="flex items-center gap-2">
        <Tag className="h-4 w-4 text-zinc-400" />
        <span className="text-xs font-bold text-zinc-300">Labels</span>
      </div>

      <div className="flex flex-wrap items-center gap-1.5">
        {taskLabels.map((label) => (
          <span
            key={label.id}
            className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[11px] font-medium bg-zinc-950 border border-zinc-800 text-zinc-200"
          >
            <span
              className="h-2 w-2 rounded-full shrink-0"
              style={{ backgroundColor: label.color || '#71717a' }}
            />
            <span>{label.name}</span>
            {canEdit && (
              <button
                type="button"
                onClick={() => handleRemove(label.id)}
                className="text-zinc-500 hover:text-zinc-300 transition-colors cursor-pointer"
                title="Remove label"
              >
                <X className="h-3 w-3" />
              </button>
            )}
          </span>
        ))}

        {canEdit && (
          <Popover open={isPopoverOpen} onOpenChange={setIsPopoverOpen}>
            <PopoverTrigger asChild>
              <button
                type="button"
                className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-medium bg-zinc-900 hover:bg-zinc-850 border border-zinc-800 text-zinc-400 hover:text-zinc-200 transition-colors cursor-pointer"
              >
                <Plus className="h-3 w-3" />
                <span>Add Label</span>
              </button>
            </PopoverTrigger>
            <PopoverContent
              align="start"
              className="w-64 p-3 bg-zinc-900 border-zinc-800 text-zinc-100 shadow-2xl rounded-md space-y-3"
            >
              <div className="text-xs font-semibold text-zinc-200">Assign or Create Label</div>

              {/* Available Labels */}
              {unassignedLabels.length > 0 && (
                <div className="space-y-1 max-h-36 overflow-y-auto">
                  <span className="text-[10px] uppercase font-mono text-zinc-500">Available</span>
                  {unassignedLabels.map((l) => (
                    <button
                      key={l.id}
                      type="button"
                      onClick={() => handleAssign(l.id)}
                      className="w-full flex items-center justify-between p-1.5 rounded hover:bg-zinc-800 text-left text-xs transition-colors cursor-pointer"
                    >
                      <div className="flex items-center gap-2">
                        <span
                          className="h-2.5 w-2.5 rounded-full shrink-0"
                          style={{ backgroundColor: l.color }}
                        />
                        <span className="text-zinc-200">{l.name}</span>
                      </div>
                      <Plus className="h-3 w-3 text-zinc-500" />
                    </button>
                  ))}
                </div>
              )}

              {/* Create New Label Form */}
              <form onSubmit={handleCreateAndAssign} className="pt-2 border-t border-zinc-800 space-y-2">
                <span className="text-[10px] uppercase font-mono text-zinc-500">New Label</span>
                <Input
                  value={newLabelName}
                  onChange={(e) => setNewLabelName(e.target.value)}
                  placeholder="Label name..."
                  className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100 h-7"
                />

                <div className="flex items-center gap-1.5">
                  {PRESET_COLORS.map((c) => (
                    <button
                      key={c}
                      type="button"
                      onClick={() => setNewLabelColor(c)}
                      className="h-4 w-4 rounded-full border border-zinc-800 transition-transform flex items-center justify-center cursor-pointer"
                      style={{
                        backgroundColor: c,
                        transform: newLabelColor === c ? 'scale(1.2)' : 'none',
                      }}
                    >
                      {newLabelColor === c && <Check className="h-2.5 w-2.5 text-white stroke-[3]" />}
                    </button>
                  ))}
                </div>

                <Button
                  type="submit"
                  size="sm"
                  variant="primary"
                  isLoading={isCreating}
                  disabled={!newLabelName.trim()}
                  className="w-full text-xs h-7 mt-1"
                >
                  Create & Assign
                </Button>
              </form>
            </PopoverContent>
          </Popover>
        )}
      </div>
    </div>
  );
});

