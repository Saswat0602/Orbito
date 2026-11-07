'use client';

import * as React from 'react';
import { useState } from 'react';
import { useParams } from 'next/navigation';
import { Target, Plus } from 'lucide-react';
import { toast } from 'sonner';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useGetProjectBySlugQuery } from '@/lib/store/api/projectsApi';
import {
  useGetSprintsQuery,
  useCreateSprintMutation,
} from '@/lib/store/api/sprintsApi';
import {
  SprintCard,
  StartSprintModal,
  CompleteSprintModal,
} from '@/features/sprints';
import type { Sprint } from '@orbito/types';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

export default function ScopedSprintsPage() {
  const params = useParams<{ workspaceSlug: string; projectSlug: string }>();
  const { workspaceSlug, projectSlug } = params;
  const { isManager, isOwner } = useAuthRbac();

  const { data: project } = useGetProjectBySlugQuery(projectSlug || '', {
    skip: !projectSlug,
  });

  const { data: sprints = [], isLoading } = useGetSprintsQuery(
    project?.id ? { projectId: project.id } : undefined,
    { skip: !project?.id }
  );

  const [createSprint, { isLoading: isCreating }] = useCreateSprintMutation();

  const [selectedSprint, setSelectedSprint] = useState<Sprint | null>(null);
  const [isStartOpen, setIsStartOpen] = useState(false);
  const [isCompleteOpen, setIsCompleteOpen] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newSprintName, setNewSprintName] = useState('');

  const handleCreateSprint = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newSprintName.trim() || !project?.id) return;

    try {
      await createSprint({
        name: newSprintName.trim(),
        projectId: project.id,
      }).unwrap();

      toast.success('Sprint created successfully');
      setNewSprintName('');
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

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Target className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              {project?.name || 'Project'} Sprints
            </h1>
          </div>
          <p className="text-xs text-zinc-400">
            Sprint cycles and iterations for project{' '}
            <span className="font-mono text-zinc-300">{projectSlug}</span>.
          </p>
        </div>

        {(isManager || isOwner) && (
          <Button
            variant="primary"
            size="sm"
            onClick={() => setIsCreateOpen(true)}
            className="gap-2 text-xs"
          >
            <Plus className="h-4 w-4" />
            <span>Create Sprint</span>
          </Button>
        )}
      </div>

      {/* Sprints List */}
      {isLoading ? (
        <div className="flex h-48 items-center justify-center">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
        </div>
      ) : sprints.length === 0 ? (
        <Card className="border-zinc-800 bg-zinc-900/60 p-12 text-center">
          <Target className="h-10 w-10 text-zinc-500 mx-auto mb-3" />
          <h3 className="text-base font-semibold text-zinc-200">No active sprints</h3>
          <p className="text-xs text-zinc-400 mt-1 max-w-sm mx-auto">
            Create an agile sprint to plan your team&apos;s upcoming commitments.
          </p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {sprints.map((sprint) => (
            <SprintCard
              key={sprint.id}
              sprint={sprint}
              onStartSprint={handleStartSprint}
              onCompleteSprint={handleCompleteSprint}
            />
          ))}
        </div>
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

          <div className="flex items-center justify-end gap-2 pt-2 border-t border-zinc-800">
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
              className="text-xs gap-1.5"
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
    </div>
  );
}
