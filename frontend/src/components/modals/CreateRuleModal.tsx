'use client';

import * as React from 'react';
import { useState } from 'react';
import { Zap, Loader2 } from 'lucide-react';
import { toast } from 'sonner';

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useCreateAutomationRuleMutation } from '@/lib/store/api/automationApi';
import { TriggerType, ActionType } from '@/types';

interface CreateRuleModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  workspaceId: string;
  organizationId?: string;
  projectId?: string;
}

const TRIGGER_OPTIONS: { label: string; value: TriggerType; desc: string }[] = [
  {
    label: 'Status Changed',
    value: TriggerType.STATUS_CHANGED,
    desc: 'When a task moves to another workflow status',
  },
  {
    label: 'Task Created',
    value: TriggerType.TASK_CREATED,
    desc: 'When a new task is created',
  },
  {
    label: 'Assignee Changed',
    value: TriggerType.ASSIGNEE_CHANGED,
    desc: 'When a task is assigned or reassigned',
  },
  {
    label: 'Due Date Approaching',
    value: TriggerType.DUE_DATE_APPROACHING,
    desc: 'When a task deadline is approaching',
  },
];

const ACTION_OPTIONS: { label: string; value: ActionType; desc: string }[] = [
  {
    label: 'Update Status',
    value: ActionType.UPDATE_STATUS,
    desc: 'Move task to a specific status category',
  },
  {
    label: 'Assign User',
    value: ActionType.ASSIGN_USER,
    desc: 'Assign to a designated team member',
  },
  {
    label: 'Send Notification',
    value: ActionType.SEND_NOTIFICATION,
    desc: 'Broadcast an alert to task watchers',
  },
  {
    label: 'Send Email',
    value: ActionType.SEND_EMAIL,
    desc: 'Send an email update to stakeholders',
  },
];

export function CreateRuleModal({
  open,
  onOpenChange,
  workspaceId,
  organizationId,
  projectId,
}: CreateRuleModalProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [triggerType, setTriggerType] = useState<TriggerType>(TriggerType.STATUS_CHANGED);
  const [actionType, setActionType] = useState<ActionType>(ActionType.UPDATE_STATUS);

  const [createRule, { isLoading }] = useCreateAutomationRuleMutation();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error('Please enter a rule name');
      return;
    }

    try {
      await createRule({
        name: name.trim(),
        description: description.trim() || undefined,
        workspaceId,
        organizationId,
        projectId,
        triggerType,
        actionType,
        triggerConfig: {},
        actionConfig: {},
      }).unwrap();

      toast.success('Automation rule created');
      onOpenChange(false);
      setName('');
      setDescription('');
    } catch (err: unknown) {
      const errMsg =
        (err as { data?: { message?: string } })?.data?.message ||
        'Failed to create automation rule';
      toast.error(errMsg);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-zinc-900 border-zinc-800 text-zinc-100 max-w-lg rounded-md p-6">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <div className="h-7 w-7 rounded-md bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400">
              <Zap className="h-4 w-4" />
            </div>
            <DialogTitle className="text-sm font-semibold text-zinc-100">
              Create Automation Rule
            </DialogTitle>
          </div>
          <DialogDescription className="text-xs text-zinc-400">
            Define automated triggers and actions across your workspace tasks.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <Label htmlFor="ruleName" className="text-xs text-zinc-300">
              Rule Name
            </Label>
            <Input
              id="ruleName"
              placeholder="e.g. Auto-notify on task status changed"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="bg-zinc-950/80 border-zinc-800 text-xs"
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="ruleDesc" className="text-xs text-zinc-300">
              Description (Optional)
            </Label>
            <Input
              id="ruleDesc"
              placeholder="Brief explanation of when this rule triggers"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="bg-zinc-950/80 border-zinc-800 text-xs"
            />
          </div>

          {/* Trigger Selection */}
          <div className="space-y-1.5">
            <Label className="text-xs text-zinc-300">When this happens (Trigger):</Label>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {TRIGGER_OPTIONS.map((opt) => {
                const isSelected = triggerType === opt.value;
                return (
                  <button
                    key={opt.value}
                    type="button"
                    onClick={() => setTriggerType(opt.value)}
                    className={`p-2.5 rounded-md border text-left transition-all cursor-pointer ${
                      isSelected
                        ? 'bg-zinc-850 border-zinc-600 text-zinc-100 shadow-xs'
                        : 'bg-zinc-950/60 border-zinc-800 text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900'
                    }`}
                  >
                    <p className="text-xs font-semibold">{opt.label}</p>
                    <p className="text-[10px] text-zinc-500 mt-0.5 line-clamp-1">{opt.desc}</p>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Action Selection */}
          <div className="space-y-1.5">
            <Label className="text-xs text-zinc-300">Then do this (Action):</Label>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {ACTION_OPTIONS.map((opt) => {
                const isSelected = actionType === opt.value;
                return (
                  <button
                    key={opt.value}
                    type="button"
                    onClick={() => setActionType(opt.value)}
                    className={`p-2.5 rounded-md border text-left transition-all cursor-pointer ${
                      isSelected
                        ? 'bg-zinc-850 border-zinc-600 text-zinc-100 shadow-xs'
                        : 'bg-zinc-950/60 border-zinc-800 text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900'
                    }`}
                  >
                    <p className="text-xs font-semibold">{opt.label}</p>
                    <p className="text-[10px] text-zinc-500 mt-0.5 line-clamp-1">{opt.desc}</p>
                  </button>
                );
              })}
            </div>
          </div>

          <DialogFooter className="flex items-center justify-end gap-2 pt-4 border-t border-zinc-800/80">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
              className="text-xs"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="sm"
              disabled={isLoading || !name.trim()}
              className="text-xs gap-1.5"
            >
              {isLoading ? (
                <>
                  <Loader2 className="h-3 w-3 animate-spin" />
                  <span>Creating Rule...</span>
                </>
              ) : (
                <span>Create Rule</span>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
