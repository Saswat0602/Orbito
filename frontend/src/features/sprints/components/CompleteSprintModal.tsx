'use client';

import * as React from 'react';
import { useState } from 'react';
import { CheckCircle2 } from 'lucide-react';
import { toast } from 'sonner';
import type { Sprint } from '@orbito/types';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { useCompleteSprintMutation } from '@/lib/store/api/sprintsApi';

interface CompleteSprintModalProps {
  isOpen: boolean;
  onClose: () => void;
  sprint: Sprint | null;
}

export function CompleteSprintModal({
  isOpen,
  onClose,
  sprint,
}: CompleteSprintModalProps) {
  const [completeSprint, { isLoading }] = useCompleteSprintMutation();

  const handleComplete = async () => {
    if (!sprint) return;

    try {
      await completeSprint({
        id: sprint.id,
        payload: {},
      }).unwrap();

      toast.success(`Sprint "${sprint.name}" completed!`);
      onClose();
    } catch {
      toast.error('Failed to complete sprint');
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`Complete Sprint: ${sprint?.name || ''}`}
      description="Close out this sprint cycle. Unfinished tasks will be moved back to the backlog."
    >
      <div className="space-y-4 pt-2">
        <div className="p-4 rounded-md bg-zinc-950/60 border border-zinc-800 text-xs text-zinc-300 space-y-1">
          <p>
            Completing this sprint will lock its burndown report and calculate team velocity.
          </p>
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
            type="button"
            variant="primary"
            size="sm"
            isLoading={isLoading}
            onClick={handleComplete}
            className="text-xs gap-1.5 bg-emerald-600 hover:bg-emerald-500 text-white"
          >
            <CheckCircle2 className="h-3.5 w-3.5" />
            <span>Confirm Complete</span>
          </Button>
        </div>
      </div>
    </Modal>
  );
}

