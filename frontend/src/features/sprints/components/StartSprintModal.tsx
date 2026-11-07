'use client';

import * as React from 'react';
import { useState } from 'react';
import { Play } from 'lucide-react';
import { toast } from 'sonner';
import type { Sprint } from '@orbito/types';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { useStartSprintMutation } from '@/lib/store/api/sprintsApi';

interface StartSprintModalProps {
  isOpen: boolean;
  onClose: () => void;
  sprint: Sprint | null;
}

export function StartSprintModal({
  isOpen,
  onClose,
  sprint,
}: StartSprintModalProps) {
  const [startSprint, { isLoading }] = useStartSprintMutation();

  const [startDate, setStartDate] = useState(
    new Date().toISOString().split('T')[0]
  );
  const [endDate, setEndDate] = useState(
    new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  );
  const [goal, setGoal] = useState(sprint?.goal || '');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!sprint) return;

    try {
      await startSprint({
        id: sprint.id,
        payload: {
          startDate,
          endDate,
          goal: goal.trim() || undefined,
        },
      }).unwrap();

      toast.success(`Sprint "${sprint.name}" started!`);
      onClose();
    } catch {
      toast.error('Failed to start sprint');
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`Start Sprint: ${sprint?.name || ''}`}
      description="Commence this sprint iteration and activate velocity tracking."
    >
      <form onSubmit={handleSubmit} className="space-y-4 pt-2">
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1.5">
            <Label htmlFor="start-date" className="text-xs text-zinc-300">
              Start Date
            </Label>
            <Input
              id="start-date"
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              required
              className="bg-zinc-950 border-zinc-800 text-xs text-zinc-200"
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="end-date" className="text-xs text-zinc-300">
              End Date
            </Label>
            <Input
              id="end-date"
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              required
              className="bg-zinc-950 border-zinc-800 text-xs text-zinc-200"
            />
          </div>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="sprint-goal" className="text-xs text-zinc-300">
            Sprint Goal
          </Label>
          <Textarea
            id="sprint-goal"
            rows={3}
            value={goal}
            onChange={(e) => setGoal(e.target.value)}
            placeholder="What does the team aim to achieve during this iteration?"
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
            <Play className="h-3 w-3" />
            <span>Start Sprint</span>
          </Button>
        </div>
      </form>
    </Modal>
  );
}

