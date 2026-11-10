'use client';

import * as React from 'react';
import { useState } from 'react';
import { Plus, Sparkles, Loader2 } from 'lucide-react';
import { toast } from 'sonner';
import type { TaskPriority, TaskRecurrence } from '@orbito/types';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { DualModeEditor } from '@/components/common/DualModeEditor';
import { TaskRecurrenceSelector } from './TaskRecurrenceSelector';
import { useCreateTaskMutation } from '@/lib/store/api/tasksApi';
import { useGenerateDescriptionMutation } from '@/lib/store/api/aiChatApi';

interface CreateTaskModalProps {
  isOpen: boolean;
  onClose: () => void;
  projectId: string;
  defaultCategory?: string;
}

export function CreateTaskModal({
  isOpen,
  onClose,
  projectId,
  defaultCategory = 'TODO',
}: CreateTaskModalProps) {
  const [createTask, { isLoading }] = useCreateTaskMutation();
  const [generateDescription, { isLoading: isGeneratingDesc }] = useGenerateDescriptionMutation();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [priority, setPriority] = useState<TaskPriority>('MEDIUM' as TaskPriority);
  const [recurrenceConfig, setRecurrenceConfig] = useState<Partial<TaskRecurrence> | null>(null);

  const handleGenerateAiDescription = async () => {
    if (!title.trim()) {
      toast.error('Please enter a task title first');
      return;
    }
    try {
      const result = await generateDescription({ title: title.trim() }).unwrap();
      if (result.description) {
        setDescription(result.description);
        toast.success('Description generated with AI');
      }
    } catch {
      toast.error('Failed to generate description. Ensure AI is enabled in settings.');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !projectId) return;

    try {
      await createTask({
        title: title.trim(),
        description: description.trim() || undefined,
        priority: priority as any,
        projectId,
        isRecurring: !!recurrenceConfig,
        recurringConfig: recurrenceConfig || undefined,
      } as any).unwrap();

      toast.success('Task created successfully');
      setTitle('');
      setDescription('');
      setPriority('MEDIUM' as TaskPriority);
      setRecurrenceConfig(null);
      onClose();
    } catch {
      toast.error('Failed to create task');
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create New Task"
      description="Add an issue or backlog item to the project board."
    >
      <form onSubmit={handleSubmit} className="space-y-4 pt-2">
        <div className="space-y-1.5">
          <Label htmlFor="task-title" className="text-xs text-zinc-300">
            Title
          </Label>
          <Input
            id="task-title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. Implement refresh token rotation"
            required
            className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
          />
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="task-priority" className="text-xs text-zinc-300">
            Priority
          </Label>
          <select
            id="task-priority"
            value={priority}
            onChange={(e) => setPriority(e.target.value as TaskPriority)}
            className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 rounded-md px-3 py-2 text-xs outline-none focus:border-zinc-700"
          >
            <option value="LOW">Low</option>
            <option value="MEDIUM">Medium</option>
            <option value="HIGH">High</option>
            <option value="URGENT">Urgent</option>
          </select>
        </div>

        <div className="space-y-1.5">
          <Label className="text-xs text-zinc-300">Repeat Task</Label>
          <TaskRecurrenceSelector
            initialConfig={recurrenceConfig}
            onChange={setRecurrenceConfig}
          />
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label htmlFor="task-desc" className="text-xs text-zinc-300">
              Description
            </Label>
            <button
              type="button"
              onClick={handleGenerateAiDescription}
              disabled={isGeneratingDesc || !title.trim()}
              className="flex items-center gap-1 text-[11px] font-medium text-violet-400 hover:text-violet-300 disabled:opacity-40 disabled:hover:text-violet-400 cursor-pointer transition-colors"
              title="Auto-draft description using AI"
            >
              {isGeneratingDesc ? (
                <Loader2 className="w-3 h-3 animate-spin text-violet-400" />
              ) : (
                <Sparkles className="w-3 h-3 text-violet-400" />
              )}
              <span>{isGeneratingDesc ? 'Generating...' : 'Generate with AI'}</span>
            </button>
          </div>
          <DualModeEditor
            value={description}
            onChange={setDescription}
            placeholder="Technical specifications, requirements, or links... (Type @ to mention)"
            minRows={3}
          />
        </div>

        <div className="flex items-center justify-end gap-2 pt-2 border-t border-zinc-800">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onClose}
            className="text-xs"
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            size="sm"
            isLoading={isLoading}
            className="text-xs gap-1.5"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>Create Task</span>
          </Button>
        </div>
      </form>
    </Modal>
  );
}

