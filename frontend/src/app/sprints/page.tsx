'use client';

import * as React from 'react';
import { Target, Play, CheckCircle2, Clock, Calendar, Plus } from 'lucide-react';
import { toast } from 'sonner';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppSelector } from '@/lib/store/hooks';
import {
  useGetSprintsQuery,
  useStartSprintMutation,
  useCompleteSprintMutation,
} from '@/lib/store/api/sprintsApi';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

export default function SprintsPage() {
  const { can, isManager, isOwner } = useAuthRbac();
  const activeProjectId = useAppSelector((state) => state.auth.activeProjectId);

  const { data: sprints = [], isLoading } = useGetSprintsQuery(
    activeProjectId ? { projectId: activeProjectId } : undefined,
    { skip: !activeProjectId }
  );

  const [startSprint, { isLoading: isStarting }] = useStartSprintMutation();
  const [completeSprint, { isLoading: isCompleting }] = useCompleteSprintMutation();

  const handleStart = async (id: string) => {
    try {
      await startSprint(id).unwrap();
      toast.success('Sprint activated successfully');
    } catch {
      toast.error('Failed to start sprint');
    }
  };

  const handleComplete = async (id: string) => {
    try {
      await completeSprint(id).unwrap();
      toast.success('Sprint marked as completed');
    } catch {
      toast.error('Failed to complete sprint');
    }
  };

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Target className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Agile Sprints
            </h1>
          </div>
          <p className="text-xs text-zinc-400">
            Plan, monitor velocity, and track iterations toward release goals.
          </p>
        </div>

        {(isManager || isOwner) && (
          <Button variant="primary" size="sm" className="gap-2">
            <Plus className="h-4 w-4" />
            <span>Create Sprint</span>
          </Button>
        )}
      </div>

      {/* Sprints Listing */}
      {isLoading ? (
        <div className="flex h-48 items-center justify-center">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
        </div>
      ) : sprints.length === 0 ? (
        <Card className="border-zinc-800 bg-zinc-900/60 p-12 text-center">
          <Target className="h-10 w-10 text-zinc-500 mx-auto mb-3" />
          <h3 className="text-base font-semibold text-zinc-200">No active sprints</h3>
          <p className="text-xs text-zinc-400 mt-1 max-w-sm mx-auto">
            {activeProjectId
              ? 'Create a sprint to start planning tasks for your next iteration.'
              : 'Select a project from the sidebar to view its active sprints.'}
          </p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {sprints.map((sprint) => {
            const isActive = sprint.status === 'ACTIVE';
            const isCompleted = sprint.status === 'COMPLETED';

            return (
              <Card
                key={sprint.id}
                className={`border-zinc-800 bg-zinc-900/80 transition-all ${
                  isActive ? 'border-zinc-500 ring-1 ring-zinc-500/20' : ''
                }`}
              >
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <Badge
                      variant={
                        isActive
                          ? 'default'
                          : isCompleted
                          ? 'secondary'
                          : 'outline'
                      }
                      className="text-[10px]"
                    >
                      {sprint.status}
                    </Badge>
                  </div>
                  <CardTitle className="text-base mt-2">{sprint.name}</CardTitle>
                  {sprint.goal && (
                    <CardDescription className="text-xs line-clamp-2">
                      Goal: {sprint.goal}
                    </CardDescription>
                  )}
                </CardHeader>

                <CardContent className="space-y-4 pt-0">
                  <div className="flex items-center gap-4 text-xs text-zinc-400 border-t border-zinc-850 pt-3">
                    <div className="flex items-center gap-1.5">
                      <Calendar className="h-3.5 w-3.5" />
                      <span>
                        {sprint.startDate
                          ? new Date(sprint.startDate).toLocaleDateString()
                          : 'Unscheduled'}
                      </span>
                    </div>
                  </div>

                  {/* Actions (Manager/Owner only) */}
                  {(isManager || isOwner) && (
                    <div className="pt-2">
                      {sprint.status === 'PLANNING' && (
                        <Button
                          variant="primary"
                          size="sm"
                          className="w-full gap-2 text-xs"
                          isLoading={isStarting}
                          onClick={() => handleStart(sprint.id)}
                        >
                          <Play className="h-3.5 w-3.5" />
                          <span>Start Sprint</span>
                        </Button>
                      )}
                      {isActive && (
                        <Button
                          variant="secondary"
                          size="sm"
                          className="w-full gap-2 text-xs"
                          isLoading={isCompleting}
                          onClick={() => handleComplete(sprint.id)}
                        >
                          <CheckCircle2 className="h-3.5 w-3.5" />
                          <span>Complete Sprint</span>
                        </Button>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}

