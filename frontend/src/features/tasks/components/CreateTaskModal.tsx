'use client';

import * as React from 'react';
import { useState } from 'react';
import { Plus } from 'lucide-react';
import { toast } from 'sonner';
import type { TaskPriority } from '@orbito/types';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { useCreateTaskMutation } from '@/lib/store/api/tasksApi';

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

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [priority, setPriority] = useState<TaskPriority>('MEDIUM' as TaskPriority);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !projectId) return;

    try {
      await createTask({
        title: title.trim(),
        description: description.trim() || undefined,
        priority: priority as any,
        projectId,
      }).unwrap();

      toast.success('Task created successfully');
      setTitle('');
      setDescription('');
      setPriority('MEDIUM' as TaskPriority);
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
          <Label htmlFor="task-desc" className="text-xs text-zinc-300">
            Description
          </Label>
          <Textarea
            id="task-desc"
            rows={4}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Technical specifications, requirements, or links..."
            className="bg-zinc-950 border-zinc-800 text-xs text-zinc-200 resize-none"
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

