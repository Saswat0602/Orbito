'use client';

import * as React from 'react';
import { useState } from 'react';
import { useParams } from 'next/navigation';
import { CalendarDays, Plus } from 'lucide-react';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useGetProjectBySlugQuery } from '@/lib/store/api/projectsApi';
import { useGetTasksQuery } from '@/lib/store/api/tasksApi';
import { TaskCalendarView, CreateTaskModal } from '@/features/tasks';
import { CalendarTimelineSkeleton } from '@/components/skeletons';
import { Button } from '@/components/ui/button';

export default function ProjectCalendarPage() {
  const params = useParams<{ workspaceSlug: string; projectSlug: string }>();
  const { workspaceSlug, projectSlug } = params;
  const { isMember } = useAuthRbac();

  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [createCategory, setCreateCategory] = useState('TODO');

  const { data: project, isLoading: isProjectLoading } = useGetProjectBySlugQuery(
    projectSlug || '',
    { skip: !projectSlug }
  );

  const { data: tasks = [], isLoading: isTasksLoading } = useGetTasksQuery(
    project?.id ? { projectId: project.id } : undefined,
    { skip: !project?.id }
  );

  const handleOpenCreateModal = (category = 'TODO') => {
    setCreateCategory(category);
    setIsCreateModalOpen(true);
  };

  if (isProjectLoading || isTasksLoading) {
    return <CalendarTimelineSkeleton view="calendar" />;
  }

  if (!project) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b] text-xs text-zinc-500">
        Project not found.
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#09090b] p-6 space-y-4 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <CalendarDays className="h-4 w-4 text-zinc-400" />
          <h2 className="text-sm font-semibold tracking-wide text-zinc-100">
            {project.name} &mdash; Calendar Schedule
          </h2>
          <span className="rounded bg-zinc-800 px-2 py-0.5 font-mono text-[11px] text-zinc-400">
            {tasks.length} tasks
          </span>
        </div>

        {isMember && (
          <Button
            variant="primary"
            size="sm"
            onClick={() => handleOpenCreateModal('TODO')}
            className="gap-1.5 text-xs h-8"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>New Task</span>
          </Button>
        )}
      </div>

      {/* Main Calendar Grid */}
      <div className="flex-1 min-h-0 overflow-hidden flex flex-col">
        <TaskCalendarView
          tasks={tasks}
          canEdit={isMember}
          onOpenCreateTask={() => handleOpenCreateModal('TODO')}
        />
      </div>

      {/* Task Creation Modal */}
      <CreateTaskModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        projectId={project.id}
        defaultCategory={createCategory}
      />
    </div>
  );
}

