'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import { useParams } from 'next/navigation';
import {
  Kanban,
  List,
  Plus,
  Search,
  Filter,
} from 'lucide-react';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useGetProjectBySlugQuery } from '@/lib/store/api/projectsApi';
import { useGetTasksQuery } from '@/lib/store/api/tasksApi';
import { KanbanBoard, CreateTaskModal } from '@/features/tasks';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';

export default function ProjectTasksPage() {
  const params = useParams<{ workspaceSlug: string; projectSlug: string }>();
  const { workspaceSlug, projectSlug } = params;

  const { can, isMember } = useAuthRbac();
  const [viewMode, setViewMode] = useState<'board' | 'list'>('board');
  const [searchQuery, setSearchQuery] = useState('');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [createCategory, setCreateCategory] = useState('TODO');

  // Fetch Project
  const { data: project, isLoading: isProjectLoading } = useGetProjectBySlugQuery(
    projectSlug || '',
    { skip: !projectSlug }
  );

  // Fetch Tasks for Project
  const { data: tasks = [], isLoading: isTasksLoading } = useGetTasksQuery(
    project?.id ? { projectId: project.id } : undefined,
    { skip: !project?.id }
  );

  // Filter tasks by title or slug
  const filteredTasks = useMemo(() => {
    if (!searchQuery.trim()) return tasks;
    const query = searchQuery.toLowerCase();
    return tasks.filter(
      (t) =>
        t.title.toLowerCase().includes(query) ||
        (t.slug && t.slug.toLowerCase().includes(query))
    );
  }, [tasks, searchQuery]);

  const handleOpenCreateModal = (category = 'TODO') => {
    setCreateCategory(category);
    setIsCreateModalOpen(true);
  };

  if (isProjectLoading || isTasksLoading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b]">
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
          <span className="text-xs text-zinc-400">Loading tasks...</span>
        </div>
      </div>
    );
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
      {/* Control Bar: Search, Filters, View Modes, New Task */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 shrink-0">
        <div className="flex items-center gap-3 flex-1 max-w-md">
          <div className="relative flex-1">
            <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-zinc-500" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Filter tasks by name or key..."
              className="pl-8 bg-zinc-950/80 border-zinc-800 text-xs text-zinc-200 h-8"
            />
          </div>

          <div className="flex items-center gap-1 bg-zinc-900 border border-zinc-850 p-0.5 rounded-md">
            <button
              type="button"
              onClick={() => setViewMode('board')}
              className={`p-1.5 rounded-md transition-colors ${
                viewMode === 'board'
                  ? 'bg-zinc-800 text-zinc-100 shadow-xs'
                  : 'text-zinc-500 hover:text-zinc-300'
              }`}
              title="Board View"
            >
              <Kanban className="h-3.5 w-3.5" />
            </button>
            <button
              type="button"
              onClick={() => setViewMode('list')}
              className={`p-1.5 rounded-md transition-colors ${
                viewMode === 'list'
                  ? 'bg-zinc-800 text-zinc-100 shadow-xs'
                  : 'text-zinc-500 hover:text-zinc-300'
              }`}
              title="List View"
            >
              <List className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>

        {isMember && (
          <Button
            variant="primary"
            size="sm"
            onClick={() => handleOpenCreateModal('TODO')}
            className="gap-1.5 text-xs shrink-0"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>New Task</span>
          </Button>
        )}
      </div>

      {/* Main View Area */}
      <div className="flex-1 min-h-0 overflow-hidden flex flex-col">
        {viewMode === 'board' ? (
          <KanbanBoard
            tasks={filteredTasks}
            projectId={project.id}
            onAddTask={(category) => handleOpenCreateModal(category)}
          />
        ) : (
          /* List View */
          <div className="flex-1 overflow-y-auto border border-zinc-800 rounded-md bg-zinc-950/40 divide-y divide-zinc-850">
            {filteredTasks.length === 0 ? (
              <div className="p-8 text-center text-xs text-zinc-500">
                No tasks match the filter criteria.
              </div>
            ) : (
              filteredTasks.map((task) => (
                <div
                  key={task.id}
                  className="flex items-center justify-between p-3.5 hover:bg-zinc-900/60 transition-colors cursor-pointer"
                >
                  <div className="flex items-center gap-3">
                    <span className="font-mono text-xs text-zinc-500 uppercase">
                      {task.slug || `TASK-${task.id.slice(0, 4)}`}
                    </span>
                    <span className="text-xs font-medium text-zinc-200">
                      {task.title}
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="text-[10px] uppercase font-mono">
                      {task.status?.name || 'TODO'}
                    </Badge>
                    <Badge variant="outline" className="text-[10px] uppercase font-mono">
                      {task.priority}
                    </Badge>
                  </div>
                </div>
              ))
            )}
          </div>
        )}
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
