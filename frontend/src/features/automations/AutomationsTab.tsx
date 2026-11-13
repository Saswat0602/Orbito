'use client';

import * as React from 'react';
import { useState } from 'react';
import { Zap, Plus, Trash2, ArrowRight, Play, Pause, Activity } from 'lucide-react';
import { toast } from 'sonner';

import {
  useGetAutomationRulesQuery,
  useToggleAutomationRuleMutation,
  useDeleteAutomationRuleMutation,
} from '@/lib/store/api/automationApi';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { CreateRuleModal } from '@/components/modals/CreateRuleModal';
import { ConfirmationModal } from '@/components/modals/ConfirmationModal';
import { RuleStatus } from '@/types';

interface AutomationsTabProps {
  workspaceId: string;
  organizationId?: string;
  isManagerOrOwner: boolean;
}

export function AutomationsTab({
  workspaceId,
  organizationId,
  isManagerOrOwner,
}: AutomationsTabProps) {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [ruleToDelete, setRuleToDelete] = useState<{ id: string; name: string } | null>(null);

  const { data: rules = [], isLoading } = useGetAutomationRulesQuery(
    workspaceId ? { workspaceId } : undefined,
    { skip: !workspaceId }
  );

  const [toggleRule, { isLoading: isToggling }] = useToggleAutomationRuleMutation();
  const [deleteRule, { isLoading: isDeleting }] = useDeleteAutomationRuleMutation();

  const handleToggle = async (id: string, currentStatus: string) => {
    try {
      await toggleRule(id).unwrap();
      toast.success(
        currentStatus === 'ACTIVE'
          ? 'Automation rule paused'
          : 'Automation rule activated'
      );
    } catch {
      toast.error('Failed to toggle automation rule');
    }
  };

  const handleConfirmDelete = async () => {
    if (!ruleToDelete) return;
    try {
      await deleteRule(ruleToDelete.id).unwrap();
      toast.success('Automation rule deleted');
      setRuleToDelete(null);
    } catch {
      toast.error('Failed to delete automation rule');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-2">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <h3 className="text-base font-semibold text-zinc-100">
              Workspace Automations
            </h3>
            <Badge variant="outline" className="text-[10px] font-mono">
              {rules.length} {rules.length === 1 ? 'Rule' : 'Rules'}
            </Badge>
          </div>
          <p className="text-xs text-zinc-400">
            Automate task transitions, assignments, and notifications when triggers occur.
          </p>
        </div>

        {isManagerOrOwner && (
          <Button
            variant="primary"
            size="sm"
            onClick={() => setIsCreateOpen(true)}
            className="gap-1.5 text-xs"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>New Automation Rule</span>
          </Button>
        )}
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, idx) => (
            <Card key={idx} className="border-zinc-800 bg-zinc-900/60 p-4 space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2.5">
                  <Skeleton className="h-4 w-4 rounded" />
                  <Skeleton className="h-4 w-40" />
                  <Skeleton className="h-4 w-14 rounded" />
                </div>
                <Skeleton className="h-7 w-20 rounded" />
              </div>
              <Skeleton className="h-3 w-3/4" />
            </Card>
          ))}
        </div>
      ) : rules.length === 0 ? (
        <Card className="border-zinc-800 bg-zinc-900/60 p-10 text-center space-y-3">
          <div className="h-10 w-10 rounded-full bg-zinc-850 flex items-center justify-center text-zinc-400 mx-auto">
            <Zap className="h-5 w-5" />
          </div>
          <div className="space-y-1">
            <h4 className="text-sm font-semibold text-zinc-200">
              No automation rules yet
            </h4>
            <p className="text-xs text-zinc-500 max-w-sm mx-auto">
              Create rules to automatically move tasks, assign owners, or post alerts on events.
            </p>
          </div>
          {isManagerOrOwner && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setIsCreateOpen(true)}
              className="text-xs gap-1.5 mt-2"
            >
              <Plus className="h-3 w-3" />
              <span>Create Your First Rule</span>
            </Button>
          )}
        </Card>
      ) : (
        <div className="space-y-3 max-w-3xl">
          {rules.map((rule) => {
            const isActive =
              rule.status === RuleStatus.ACTIVE || (rule.status as string) === 'ACTIVE';
            return (
              <Card
                key={rule.id}
                className={`border-zinc-800 bg-zinc-900/80 p-4 transition-all ${
                  isActive ? 'border-zinc-750' : 'opacity-70'
                }`}
              >
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                  <div className="space-y-1.5 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-semibold text-zinc-100 truncate">
                        {rule.name}
                      </span>
                      <Badge
                        variant={isActive ? 'default' : 'secondary'}
                        className="text-[9px] uppercase tracking-wider"
                      >
                        {rule.status}
                      </Badge>
                      {rule.executionCount !== undefined && rule.executionCount > 0 && (
                        <span className="text-[10px] text-zinc-500 flex items-center gap-1">
                          <Activity className="h-2.5 w-2.5" />
                          <span>{rule.executionCount} runs</span>
                        </span>
                      )}
                    </div>

                    {rule.description && (
                      <p className="text-xs text-zinc-400 line-clamp-1">
                        {rule.description}
                      </p>
                    )}

                    {/* Trigger & Action Pipeline Flow */}
                    <div className="flex items-center gap-2 pt-1 text-[11px] text-zinc-400">
                      <span className="px-2 py-0.5 rounded bg-zinc-800 border border-zinc-700/60 font-mono text-zinc-300">
                        {String(rule.triggerType || rule.trigger || '').replace(/_/g, ' ')}
                      </span>
                      <ArrowRight className="h-3 w-3 text-zinc-600" />
                      <span className="px-2 py-0.5 rounded bg-zinc-800 border border-zinc-700/60 font-mono text-zinc-300">
                        {String(rule.actionType || rule.action || '').replace(/_/g, ' ')}
                      </span>
                    </div>
                  </div>

                  {/* Actions */}
                  {isManagerOrOwner && (
                    <div className="flex items-center gap-2 shrink-0 self-end sm:self-center">
                      <Button
                        type="button"
                        variant="secondary"
                        size="sm"
                        onClick={() => handleToggle(rule.id, rule.status)}
                        disabled={isToggling}
                        className="h-7 text-xs gap-1"
                      >
                        {isActive ? (
                          <>
                            <Pause className="h-3 w-3" />
                            <span>Pause</span>
                          </>
                        ) : (
                          <>
                            <Play className="h-3 w-3" />
                            <span>Activate</span>
                          </>
                        )}
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => setRuleToDelete({ id: rule.id, name: rule.name })}
                        disabled={isDeleting}
                        aria-label="Delete rule"
                        className="h-7 w-7 p-0 text-zinc-400 hover:text-rose-400 hover:bg-rose-950/40"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  )}
                </div>
              </Card>
            );
          })}
        </div>
      )}

      {/* Create Rule Modal */}
      <CreateRuleModal
        open={isCreateOpen}
        onOpenChange={setIsCreateOpen}
        workspaceId={workspaceId}
        organizationId={organizationId}
      />

      {/* Delete Rule Confirmation Modal */}
      <ConfirmationModal
        open={!!ruleToDelete}
        onOpenChange={(open) => !open && setRuleToDelete(null)}
        title="Delete Automation Rule"
        description={
          <span>
            Are you sure you want to delete automation rule{' '}
            <strong className="text-zinc-200">&ldquo;{ruleToDelete?.name}&rdquo;</strong>?
            This rule will be removed permanently and won&apos;t trigger again.
          </span>
        }
        confirmLabel="Delete Rule"
        variant="destructive"
        isLoading={isDeleting}
        onConfirm={handleConfirmDelete}
      />
    </div>
  );
}
