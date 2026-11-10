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
  User,
  Calendar,
} from 'lucide-react';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppDispatch } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';
import { useGetProjectBySlugQuery } from '@/lib/store/api/projectsApi';
import { useGetTasksQuery } from '@/lib/store/api/tasksApi';
import { KanbanBoard, CreateTaskModal } from '@/features/tasks';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

export default function ProjectTasksPage() {
  const params = useParams<{ workspaceSlug: string; projectSlug: string }>();
  const dispatch = useAppDispatch();
  const { workspaceSlug, projectSlug } = params;

  const { can, isMember } = useAuthRbac();
  const [viewMode, setViewMode] = useState<'board' | 'list'>('board');
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [priorityFilter, setPriorityFilter] = useState('ALL');
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

  // Fast memoized task filtration
  const filteredTasks = useMemo(() => {
    return tasks.filter((t) => {
      if (searchQuery.trim()) {
        const query = searchQuery.toLowerCase();
        const matchQuery =
          t.title.toLowerCase().includes(query) ||
          (t.slug && t.slug.toLowerCase().includes(query));
        if (!matchQuery) return false;
      }
      if (statusFilter !== 'ALL') {
        const cat = t.status?.category || 'TODO';
        if (cat !== statusFilter) return false;
      }
      if (priorityFilter !== 'ALL') {
        if (t.priority !== priorityFilter) return false;
      }
      return true;
    });
  }, [tasks, searchQuery, statusFilter, priorityFilter]);

  const handleOpenCreateModal = (category = 'TODO') => {
    setCreateCategory(category);
    setIsCreateModalOpen(true);
  };

  const handleTaskClick = (taskId: string) => {
    dispatch(setTaskDrawerId(taskId));
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
      {/* Control Bar: Search, Status Filter, Priority Filter, View Modes, New Task */}
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-3 shrink-0">
        <div className="flex flex-wrap items-center gap-2.5 flex-1">
          {/* Search Input */}
          <div className="relative min-w-[200px] max-w-xs flex-1">
            <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-zinc-500" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Filter tasks by name or key..."
              className="pl-8 bg-zinc-950/80 border-zinc-800 text-xs text-zinc-200 h-8"
            />
          </div>

          {/* Status Filter */}
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="bg-zinc-950/80 border border-zinc-800 text-zinc-300 rounded px-2.5 py-1.5 text-xs outline-none focus:border-zinc-700 cursor-pointer h-8"
          >
            <option value="ALL">All Statuses</option>
            <option value="TODO">To Do</option>
            <option value="IN_PROGRESS">In Progress</option>
            <option value="IN_REVIEW">In Review</option>
            <option value="DONE">Done</option>
          </select>

          {/* Priority Filter */}
          <select
            value={priorityFilter}
            onChange={(e) => setPriorityFilter(e.target.value)}
            className="bg-zinc-950/80 border border-zinc-800 text-zinc-300 rounded px-2.5 py-1.5 text-xs outline-none focus:border-zinc-700 cursor-pointer h-8"
          >
            <option value="ALL">All Priorities</option>
            <option value="URGENT">Urgent</option>
            <option value="HIGH">High</option>
            <option value="MEDIUM">Medium</option>
            <option value="LOW">Low</option>
          </select>

          {/* View Mode Toggle: Board vs List */}
          <div className="flex items-center gap-1 bg-zinc-900 border border-zinc-850 p-0.5 rounded-md h-8">
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

        {/* New Task Button */}
        {isMember && (
          <Button
            variant="primary"
            size="sm"
            onClick={() => handleOpenCreateModal('TODO')}
            className="gap-1.5 text-xs shrink-0 h-8"
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
          /* List View with Full Click-Through */
          <div className="flex-1 overflow-y-auto border border-zinc-800 rounded-md bg-zinc-950/40 divide-y divide-zinc-850">
            {filteredTasks.length === 0 ? (
              <div className="p-8 text-center text-xs text-zinc-500">
                No tasks match the filter criteria.
              </div>
            ) : (
              filteredTasks.map((task) => {
                const primaryAssignee = task.assignees?.[0]?.user;
                const initials = primaryAssignee
                  ? ((primaryAssignee.firstName?.[0] || '') + (primaryAssignee.lastName?.[0] || '')).toUpperCase() || 'U'
                  : null;

                return (
                  <div
                    key={task.id}
                    onClick={() => handleTaskClick(task.id)}
                    className="flex items-center justify-between p-3.5 hover:bg-zinc-900/60 transition-colors cursor-pointer group select-none"
                  >
                    <div className="flex items-center gap-3 min-w-0">
                      <span className="font-mono text-xs text-zinc-500 uppercase tracking-wider shrink-0">
                        {task.slug || `TASK-${task.id.slice(0, 4)}`}
                      </span>
                      <span className="text-xs font-medium text-zinc-200 group-hover:text-white transition-colors truncate">
                        {task.title}
                      </span>
                    </div>

                    <div className="flex items-center gap-3 shrink-0">
                      {primaryAssignee && (
                        <div className="flex items-center gap-1.5 text-zinc-400 text-[11px]">
                          <Avatar className="h-5 w-5 rounded-sm border border-zinc-800">
                            <AvatarFallback className="text-[9px] bg-zinc-800 text-zinc-300">
                              {initials}
                            </AvatarFallback>
                          </Avatar>
                          <span className="hidden sm:inline truncate max-w-[100px]">
                            {primaryAssignee.firstName || primaryAssignee.username}
                          </span>
                        </div>
                      )}

                      <Badge variant="outline" className="text-[10px] uppercase font-mono">
                        {task.status?.name || 'TODO'}
                      </Badge>

                      <Badge
                        variant="outline"
                        className={`text-[10px] uppercase font-mono ${
                          task.priority === 'URGENT'
                            ? 'text-rose-400 border-rose-500/30'
                            : task.priority === 'HIGH'
                            ? 'text-amber-400 border-amber-500/30'
                            : 'text-zinc-400'
                        }`}
                      >
                        {task.priority}
                      </Badge>
                    </div>
                  </div>
                );
              })
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
