'use client';

import * as React from 'react';
import { useState } from 'react';
import { useParams } from 'next/navigation';
import {
  Target,
  Plus,
  TrendingDown,
  Zap,
  BarChart2,
  Layers,
  Users,
} from 'lucide-react';
import { toast } from 'sonner';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useGetProjectBySlugQuery } from '@/lib/store/api/projectsApi';
import {
  useGetSprintsQuery,
  useCreateSprintMutation,
} from '@/lib/store/api/sprintsApi';
import { useGetTasksQuery } from '@/lib/store/api/tasksApi';
import {
  SprintPlanningView,
  SprintBurndownChart,
  SprintVelocityChart,
  SprintCapacityBar,
  ProjectAnalyticsTab,
  StartSprintModal,
  CompleteSprintModal,
} from '@/features/sprints';
import { CreateTaskModal } from '@/features/tasks';
import { SprintsSkeleton } from '@/components/skeletons';
import type { Sprint } from '@orbito/types';
import { Button } from '@/components/ui/button';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

export default function ScopedSprintsPage() {
  const params = useParams<{ workspaceSlug: string; projectSlug: string }>();
  const { workspaceSlug, projectSlug } = params;
  const { isManager, isOwner } = useAuthRbac();
  const canManage = isManager || isOwner;

  const [activeTab, setActiveTab] = useState<
    'planning' | 'burndown' | 'velocity' | 'analytics'
  >('planning');

  const { data: project, isLoading: isProjectLoading } = useGetProjectBySlugQuery(
    projectSlug || '',
    { skip: !projectSlug }
  );

  const { data: sprints = [], isLoading: isSprintsLoading } = useGetSprintsQuery(
    project?.id ? { projectId: project.id } : undefined,
    { skip: !project?.id }
  );

  const { data: tasks = [] } = useGetTasksQuery(
    project?.id ? { projectId: project.id } : undefined,
    { skip: !project?.id }
  );

  const [createSprint, { isLoading: isCreating }] = useCreateSprintMutation();

  const [selectedSprint, setSelectedSprint] = useState<Sprint | null>(null);
  const [burndownSprintId, setBurndownSprintId] = useState<string | undefined>(undefined);
  const [isStartOpen, setIsStartOpen] = useState(false);
  const [isCompleteOpen, setIsCompleteOpen] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isCreateTaskOpen, setIsCreateTaskOpen] = useState(false);
  const [newSprintName, setNewSprintName] = useState('');
  const [newSprintGoal, setNewSprintGoal] = useState('');

  const activeSprint = sprints.find((s) => s.status === 'ACTIVE') || sprints[0];

  const handleCreateSprint = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newSprintName.trim() || !project?.id) return;

    try {
      await createSprint({
        name: newSprintName.trim(),
        goal: newSprintGoal.trim() || undefined,
        projectId: project.id,
        projectSlug: project.slug,
      } as any).unwrap();

      toast.success('Sprint created successfully');
      setNewSprintName('');
      setNewSprintGoal('');
      setIsCreateOpen(false);
    } catch {
      toast.error('Failed to create sprint');
    }
  };

  const handleStartSprint = (sprint: Sprint) => {
    setSelectedSprint(sprint);
    setIsStartOpen(true);
  };

  const handleCompleteSprint = (sprint: Sprint) => {
    setSelectedSprint(sprint);
    setIsCompleteOpen(true);
  };

  const handleSwitchToBurndown = (sprintId: string) => {
    setBurndownSprintId(sprintId);
    setActiveTab('burndown');
  };

  if (isProjectLoading || isSprintsLoading) {
    return <SprintsSkeleton />;
  }

  if (!project) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b] text-xs text-zinc-500">
        Project not found.
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Target className="h-5 w-5 text-zinc-100" />
            <h1 className="text-xl font-bold tracking-tight text-zinc-100">
              {project.name} &mdash; Agile Sprints & Analytics
            </h1>
          </div>
          <p className="text-xs text-zinc-400">
            Sprint planning, backlog management, and real-time velocity metrics.
          </p>
        </div>

        {canManage && (
          <Button
            variant="primary"
            size="sm"
            onClick={() => setIsCreateOpen(true)}
            className="gap-1.5 text-xs h-8"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>Create Sprint</span>
          </Button>
        )}
      </div>

      {/* Agile Mode Navigation Tabs */}
      <div className="flex items-center gap-1.5 bg-zinc-950 border border-zinc-850 p-1 rounded-md w-fit overflow-x-auto max-w-full">
        <button
          type="button"
          onClick={() => setActiveTab('planning')}
          className={`flex items-center gap-2 px-3 py-1.5 rounded text-xs font-medium transition ${
            activeTab === 'planning'
              ? 'bg-zinc-800 text-zinc-100 shadow-xs'
              : 'text-zinc-400 hover:text-zinc-200'
          }`}
        >
          <Layers className="h-3.5 w-3.5" />
          <span>Planning & Backlog</span>
        </button>

        <button
          type="button"
          onClick={() => setActiveTab('burndown')}
          className={`flex items-center gap-2 px-3 py-1.5 rounded text-xs font-medium transition ${
            activeTab === 'burndown'
              ? 'bg-zinc-800 text-zinc-100 shadow-xs'
              : 'text-zinc-400 hover:text-zinc-200'
          }`}
        >
          <TrendingDown className="h-3.5 w-3.5" />
          <span>Burndown Trajectory</span>
        </button>

        <button
          type="button"
          onClick={() => setActiveTab('velocity')}
          className={`flex items-center gap-2 px-3 py-1.5 rounded text-xs font-medium transition ${
            activeTab === 'velocity'
              ? 'bg-zinc-800 text-zinc-100 shadow-xs'
              : 'text-zinc-400 hover:text-zinc-200'
          }`}
        >
          <Zap className="h-3.5 w-3.5" />
          <span>Velocity & Capacity</span>
        </button>

        <button
          type="button"
          onClick={() => setActiveTab('analytics')}
          className={`flex items-center gap-2 px-3 py-1.5 rounded text-xs font-medium transition ${
            activeTab === 'analytics'
              ? 'bg-zinc-800 text-zinc-100 shadow-xs'
              : 'text-zinc-400 hover:text-zinc-200'
          }`}
        >
          <BarChart2 className="h-3.5 w-3.5" />
          <span>Project Analytics</span>
        </button>
      </div>

      {/* Tab Panels */}
      {activeTab === 'planning' ? (
        <SprintPlanningView
          project={project}
          sprints={sprints}
          canManage={canManage}
          onStartSprint={handleStartSprint}
          onCompleteSprint={handleCompleteSprint}
          onSwitchToBurndown={handleSwitchToBurndown}
          onOpenCreateTask={() => setIsCreateTaskOpen(true)}
        />
      ) : activeTab === 'burndown' ? (
        <SprintBurndownChart
          projectSlug={projectSlug}
          sprints={sprints}
          initialSprintId={burndownSprintId}
        />
      ) : activeTab === 'velocity' ? (
        <div className="space-y-8">
          <SprintVelocityChart projectSlug={projectSlug} />
          {activeSprint ? (
            <SprintCapacityBar sprint={activeSprint} tasks={tasks} />
          ) : (
            <div className="p-8 border border-zinc-850 rounded bg-zinc-950/40 text-center text-xs text-zinc-500 italic">
              No active sprint to calculate team capacity.
            </div>
          )}
        </div>
      ) : (
        <ProjectAnalyticsTab projectSlug={projectSlug} />
      )}

      {/* Create Sprint Modal */}
      <Modal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        title="Create New Sprint"
        description="Add a new planning sprint iteration to this project."
      >
        <form onSubmit={handleCreateSprint} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <Label htmlFor="sprint-name" className="text-xs text-zinc-300">
              Sprint Name
            </Label>
            <Input
              id="sprint-name"
              value={newSprintName}
              onChange={(e) => setNewSprintName(e.target.value)}
              placeholder="e.g. Sprint 24 - Core Workflow"
              required
              className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="sprint-goal" className="text-xs text-zinc-300">
              Sprint Goal (Optional)
            </Label>
            <Input
              id="sprint-goal"
              value={newSprintGoal}
              onChange={(e) => setNewSprintGoal(e.target.value)}
              placeholder="e.g. Deliver interactive burndown charts and capacity bars"
              className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
            />
          </div>

          <div className="flex items-center justify-end gap-2 pt-2 border-t border-zinc-800">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setIsCreateOpen(false)}
              className="text-xs h-8"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="sm"
              isLoading={isCreating}
              className="text-xs gap-1.5 h-8"
            >
              <Plus className="h-3.5 w-3.5" />
              <span>Create Sprint</span>
            </Button>
          </div>
        </form>
      </Modal>

      {/* Start Sprint Modal */}
      <StartSprintModal
        isOpen={isStartOpen}
        onClose={() => setIsStartOpen(false)}
        sprint={selectedSprint}
      />

      {/* Complete Sprint Modal */}
      <CompleteSprintModal
        isOpen={isCompleteOpen}
        onClose={() => setIsCompleteOpen(false)}
        sprint={selectedSprint}
      />

      {/* Create Backlog Task Modal */}
      <CreateTaskModal
        isOpen={isCreateTaskOpen}
        onClose={() => setIsCreateTaskOpen(false)}
        projectId={project.id}
        defaultCategory="TODO"
      />
    </div>
  );
}
