'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  ArrowLeft,
  Target,
  Calendar,
  Layers,
  Plus,
  Search,
  Filter,
  Download,
  Upload,
  KanbanSquare,
  ListFilter,
  GanttChartSquare,
  CheckCircle2,
  AlertCircle,
  FileSpreadsheet,
  FileText,
  Clock,
} from 'lucide-react';
import { toast } from 'sonner';

import type { Task, TaskPriority } from '@orbito/types';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useGetProjectBySlugQuery } from '@/lib/store/api/projectsApi';
import { useGetSprintQuery } from '@/lib/store/api/sprintsApi';
import { useGetTasksQuery } from '@/lib/store/api/tasksApi';
import { useAppDispatch } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';
import {
  KanbanBoard,
  TaskGanttView,
  CreateTaskModal,
  CsvImportModal,
  TaskDetailDrawer,
  exportTasksToCsv,
  exportTasksToExcel,
  DEFAULT_COLUMNS,
} from '@/features/tasks';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu';

export default function DedicatedSprintBoardPage() {
  const params = useParams<{
    workspaceSlug: string;
    projectSlug: string;
    sprintId: string;
  }>();
  const router = useRouter();
  const dispatch = useAppDispatch();
  const { workspaceSlug, projectSlug, sprintId } = params;
  const { isMember, isManager, isOwner } = useAuthRbac();
  const canManage = isManager || isOwner;

  // View state
  const [viewMode, setViewMode] = useState<'board' | 'list' | 'gantt'>('board');
  const [searchQuery, setSearchQuery] = useState('');
  const [priorityFilter, setPriorityFilter] = useState<string>('ALL');

  // Modals state
  const [isCreateTaskOpen, setIsCreateTaskOpen] = useState(false);
  const [isCsvImportOpen, setIsCsvImportOpen] = useState(false);
  const [createCategory, setCreateCategory] = useState('TODO');

  // Queries
  const { data: project, isLoading: isProjectLoading } = useGetProjectBySlugQuery(
    projectSlug || '',
    { skip: !projectSlug }
  );

  const { data: sprint, isLoading: isSprintLoading } = useGetSprintQuery(
    sprintId || '',
    { skip: !sprintId }
  );

  const { data: allProjectTasks = [], isLoading: isTasksLoading } = useGetTasksQuery(
    project?.id ? { projectId: project.id } : undefined,
    { skip: !project?.id }
  );

  // Filter tasks strictly belonging to this sprint
  const sprintTasks = useMemo(() => {
    if (!sprintId) return [];
    return allProjectTasks.filter((t) => t.sprintId === sprintId);
  }, [allProjectTasks, sprintId]);

  // Apply search and priority filters
  const filteredTasks = useMemo(() => {
    return sprintTasks.filter((task) => {
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const matchesTitle = task.title.toLowerCase().includes(q);
        const matchesSlug = task.slug?.toLowerCase().includes(q);
        if (!matchesTitle && !matchesSlug) return false;
      }

      if (priorityFilter !== 'ALL') {
        if (task.priority !== priorityFilter) return false;
      }

      return true;
    });
  }, [sprintTasks, searchQuery, priorityFilter]);

  // Sprint Progress & Metrics
  const totalTasks = sprintTasks.length;
  const completedTasks = sprintTasks.filter(
    (t) => t.status?.category === 'DONE'
  ).length;
  const totalStoryPoints = sprintTasks.reduce(
    (sum, t) => sum + (t.storyPoints || 0),
    0
  );
  const completedStoryPoints = sprintTasks
    .filter((t) => t.status?.category === 'DONE')
    .reduce((sum, t) => sum + (t.storyPoints || 0), 0);

  const completionPercent =
    totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0;

  const handleOpenCreateTask = (category = 'TODO') => {
    setCreateCategory(category);
    setIsCreateTaskOpen(true);
  };

  const handleTaskClick = (taskId: string) => {
    dispatch(setTaskDrawerId(taskId));
  };

  if (isProjectLoading || isSprintLoading || isTasksLoading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b] text-xs text-zinc-500">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100 mr-2" />
        Loading sprint iteration...
      </div>
    );
  }

  if (!project || !sprint) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center bg-[#09090b] p-6 text-center space-y-3">
        <AlertCircle className="h-8 w-8 text-zinc-600" />
        <h2 className="text-sm font-semibold text-zinc-300">Sprint Not Found</h2>
        <p className="text-xs text-zinc-500">
          The requested sprint could not be loaded or may have been deleted.
        </p>
        <Link href={`/${workspaceSlug}/${projectSlug}/sprints`}>
          <Button variant="outline" size="sm" className="text-xs border-zinc-800">
            Back to Sprint Planning
          </Button>
        </Link>
      </div>
    );
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'ACTIVE':
        return (
          <Badge className="bg-emerald-950/60 text-emerald-400 border border-emerald-800 text-[11px] font-medium">
            Active Sprint
          </Badge>
        );
      case 'COMPLETED':
        return (
          <Badge className="bg-sky-950/60 text-sky-400 border border-sky-800 text-[11px] font-medium">
            Completed
          </Badge>
        );
      default:
        return (
          <Badge className="bg-zinc-900 text-zinc-400 border border-zinc-700 text-[11px] font-medium">
            Planning
          </Badge>
        );
    }
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#09090b] p-6 space-y-4 overflow-hidden">
      {/* Top Navigation & Sprint Header */}
      <div className="space-y-3 pb-3 border-b border-zinc-800/80">
        <div className="flex items-center gap-2 text-xs text-zinc-400">
          <Link
            href={`/${workspaceSlug}/${projectSlug}/sprints`}
            className="flex items-center gap-1.5 hover:text-zinc-200 transition"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            <span>Sprint Planning</span>
          </Link>
          <span className="text-zinc-600">/</span>
          <span className="text-zinc-300 font-medium truncate">{sprint.name}</span>
        </div>

        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div className="space-y-1 min-w-0">
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-xl font-bold tracking-tight text-zinc-100 truncate">
                {sprint.name}
              </h1>
              {getStatusBadge(sprint.status)}
              {sprint.startDate && sprint.endDate && (
                <div className="flex items-center gap-1.5 text-xs text-zinc-400">
                  <Calendar className="h-3.5 w-3.5 text-zinc-500" />
                  <span>
                    {new Date(sprint.startDate).toLocaleDateString()} &mdash;{' '}
                    {new Date(sprint.endDate).toLocaleDateString()}
                  </span>
                </div>
              )}
            </div>
            {sprint.goal && (
              <p className="text-xs text-zinc-400 line-clamp-1 italic">
                Goal: &ldquo;{sprint.goal}&rdquo;
              </p>
            )}
          </div>

          {/* Action Buttons */}
          <div className="flex items-center gap-2 shrink-0 flex-wrap">
            {/* Export Dropdown */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="outline"
                  size="sm"
                  className="gap-1.5 text-xs h-8 border-zinc-800 bg-zinc-900 hover:bg-zinc-800 text-zinc-200"
                >
                  <Download className="h-3.5 w-3.5 text-zinc-400" />
                  <span>Export</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-40 border-zinc-800 bg-zinc-900 text-xs text-zinc-200">
                <DropdownMenuItem
                  onClick={() => exportTasksToCsv(sprintTasks, DEFAULT_COLUMNS, `${sprint.name}_tasks.csv`)}
                  className="gap-2 cursor-pointer focus:bg-zinc-800 text-zinc-300"
                >
                  <FileText className="h-3.5 w-3.5 text-zinc-400" />
                  <span>Export as CSV</span>
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => exportTasksToExcel(sprintTasks, DEFAULT_COLUMNS, `${sprint.name}_tasks.xls`)}
                  className="gap-2 cursor-pointer focus:bg-zinc-800 text-zinc-300"
                >
                  <FileSpreadsheet className="h-3.5 w-3.5 text-emerald-400" />
                  <span>Export as Excel</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

            {/* CSV Import */}
            {isMember && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsCsvImportOpen(true)}
                className="gap-1.5 text-xs h-8 border-zinc-800 bg-zinc-900 hover:bg-zinc-800 text-zinc-200 cursor-pointer"
              >
                <Upload className="h-3.5 w-3.5 text-zinc-400" />
                <span>Import CSV</span>
              </Button>
            )}

            {/* Add Task Button */}
            {isMember && (
              <Button
                variant="primary"
                size="sm"
                onClick={() => handleOpenCreateTask('TODO')}
                className="gap-1.5 text-xs h-8 cursor-pointer"
              >
                <Plus className="h-3.5 w-3.5" />
                <span>New Task</span>
              </Button>
            )}
          </div>
        </div>

        {/* Sprint Progress Bar & Quick Stats */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2">
          <div className="p-2.5 rounded-md bg-zinc-900/60 border border-zinc-800/80">
            <span className="text-[10px] text-zinc-500 uppercase tracking-wider block">
              Total Workload
            </span>
            <div className="flex items-baseline gap-1 mt-0.5">
              <span className="text-base font-bold text-zinc-100">{totalTasks}</span>
              <span className="text-[11px] text-zinc-400">tasks</span>
            </div>
          </div>

          <div className="p-2.5 rounded-md bg-zinc-900/60 border border-zinc-800/80">
            <span className="text-[10px] text-zinc-500 uppercase tracking-wider block">
              Completed
            </span>
            <div className="flex items-baseline gap-1 mt-0.5">
              <span className="text-base font-bold text-emerald-400">{completedTasks}</span>
              <span className="text-[11px] text-zinc-400">done</span>
            </div>
          </div>

          <div className="p-2.5 rounded-md bg-zinc-900/60 border border-zinc-800/80">
            <span className="text-[10px] text-zinc-500 uppercase tracking-wider block">
              Story Points Burn
            </span>
            <div className="flex items-baseline gap-1 mt-0.5">
              <span className="text-base font-bold text-sky-400">{completedStoryPoints}</span>
              <span className="text-[11px] text-zinc-400">/ {totalStoryPoints} pts</span>
            </div>
          </div>

          <div className="p-2.5 rounded-md bg-zinc-900/60 border border-zinc-800/80">
            <span className="text-[10px] text-zinc-500 uppercase tracking-wider block">
              Sprint Velocity
            </span>
            <div className="flex items-center gap-2 mt-1">
              <div className="flex-1 h-2 rounded-full bg-zinc-800 overflow-hidden">
                <div
                  className="h-full bg-emerald-500 transition-all duration-300"
                  style={{ width: `${completionPercent}%` }}
                />
              </div>
              <span className="text-[11px] font-mono text-zinc-300">{completionPercent}%</span>
            </div>
          </div>
        </div>
      </div>

      {/* Toolbar: View Switcher, Search & Filters */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 shrink-0">
        <div className="flex items-center gap-1.5 bg-zinc-900 p-1 rounded-md border border-zinc-800">
          <Button
            type="button"
            variant={viewMode === 'board' ? 'primary' : 'ghost'}
            size="sm"
            onClick={() => setViewMode('board')}
            className="gap-1.5 text-xs h-7 px-2.5"
          >
            <KanbanSquare className="h-3.5 w-3.5" />
            <span>Board</span>
          </Button>

          <Button
            type="button"
            variant={viewMode === 'list' ? 'primary' : 'ghost'}
            size="sm"
            onClick={() => setViewMode('list')}
            className="gap-1.5 text-xs h-7 px-2.5"
          >
            <ListFilter className="h-3.5 w-3.5" />
            <span>List</span>
          </Button>

          <Button
            type="button"
            variant={viewMode === 'gantt' ? 'primary' : 'ghost'}
            size="sm"
            onClick={() => setViewMode('gantt')}
            className="gap-1.5 text-xs h-7 px-2.5"
          >
            <GanttChartSquare className="h-3.5 w-3.5" />
            <span>Timeline</span>
          </Button>
        </div>

        <div className="flex items-center gap-2">
          {/* Search */}
          <div className="relative w-48 sm:w-60">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
            <Input
              placeholder="Filter sprint tasks..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-8 h-8 text-xs bg-zinc-900 border-zinc-800 text-zinc-200"
            />
          </div>

          {/* Priority Filter */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className="gap-1.5 text-xs h-8 border-zinc-800 bg-zinc-900 text-zinc-300"
              >
                <Filter className="h-3 w-3 text-zinc-500" />
                <span>
                  {priorityFilter === 'ALL' ? 'Priority: All' : priorityFilter}
                </span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-36 border-zinc-800 bg-zinc-900 text-xs text-zinc-200">
              <DropdownMenuItem onClick={() => setPriorityFilter('ALL')} className="cursor-pointer">
                All Priorities
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setPriorityFilter('URGENT')} className="cursor-pointer text-rose-400">
                Urgent
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setPriorityFilter('HIGH')} className="cursor-pointer text-amber-400">
                High
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setPriorityFilter('MEDIUM')} className="cursor-pointer text-sky-400">
                Medium
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setPriorityFilter('LOW')} className="cursor-pointer text-zinc-400">
                Low
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Main Viewport Content */}
      <div className="flex-1 min-h-0 overflow-hidden flex flex-col">
        {filteredTasks.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center p-8 text-center rounded-md border border-zinc-800/80 bg-zinc-950/40 space-y-3">
            <Layers className="h-8 w-8 text-zinc-600" />
            <h3 className="text-xs font-semibold text-zinc-300">No Tasks in this Sprint View</h3>
            <p className="text-[11px] text-zinc-500 max-w-sm">
              {sprintTasks.length === 0
                ? 'This sprint has no tasks assigned yet. Add tasks from sprint planning or create one below.'
                : 'No tasks match the active filters or search term.'}
            </p>
            {isMember && (
              <Button
                variant="primary"
                size="sm"
                onClick={() => handleOpenCreateTask('TODO')}
                className="text-xs h-8 gap-1.5"
              >
                <Plus className="h-3.5 w-3.5" />
                <span>Add Task to Sprint</span>
              </Button>
            )}
          </div>
        ) : viewMode === 'board' ? (
          <div className="flex-1 min-h-0 overflow-hidden">
            <KanbanBoard
              tasks={filteredTasks}
              projectId={project.id}
              onAddTask={(category) => handleOpenCreateTask(category)}
            />
          </div>
        ) : viewMode === 'list' ? (
          <div className="flex-1 min-h-0 overflow-auto rounded-md border border-zinc-800 bg-zinc-950/60">
            <table className="w-full text-left text-xs">
              <thead className="sticky top-0 bg-zinc-900/90 backdrop-blur-sm border-b border-zinc-800 text-zinc-400 font-mono text-[11px]">
                <tr>
                  <th className="py-2.5 px-4 font-normal">Key</th>
                  <th className="py-2.5 px-4 font-normal">Title</th>
                  <th className="py-2.5 px-4 font-normal">Status</th>
                  <th className="py-2.5 px-4 font-normal">Priority</th>
                  <th className="py-2.5 px-4 font-normal">Assignee</th>
                  <th className="py-2.5 px-4 font-normal">Story Points</th>
                  <th className="py-2.5 px-4 font-normal">Due Date</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-850">
                {filteredTasks.map((task) => (
                  <tr
                    key={task.id}
                    onClick={() => handleTaskClick(task.id)}
                    className="hover:bg-zinc-900/50 cursor-pointer transition select-none"
                  >
                    <td className="py-2.5 px-4 font-mono text-[11px] text-zinc-500 whitespace-nowrap">
                      {task.slug || `#${task.id.slice(0, 6)}`}
                    </td>
                    <td className="py-2.5 px-4 font-medium text-zinc-200">
                      {task.title}
                    </td>
                    <td className="py-2.5 px-4">
                      <Badge variant="outline" className="text-[10px] border-zinc-700 bg-zinc-900 text-zinc-300">
                        {task.status?.name || 'To Do'}
                      </Badge>
                    </td>
                    <td className="py-2.5 px-4">
                      <span
                        className={`text-[10px] font-semibold uppercase px-1.5 py-0.5 rounded border ${
                          task.priority === 'URGENT'
                            ? 'border-rose-900 text-rose-400 bg-rose-950/40'
                            : task.priority === 'HIGH'
                            ? 'border-amber-900 text-amber-400 bg-amber-950/40'
                            : task.priority === 'MEDIUM'
                            ? 'border-sky-900 text-sky-400 bg-sky-950/40'
                            : 'border-zinc-800 text-zinc-400 bg-zinc-900'
                        }`}
                      >
                        {task.priority || 'MEDIUM'}
                      </span>
                    </td>
                    <td className="py-2.5 px-4 text-zinc-400 text-xs">
                      {task.assignees && task.assignees.length > 0 ? (
                        <span>
                          {task.assignees[0].user
                            ? `${task.assignees[0].user.firstName || ''} ${task.assignees[0].user.lastName || ''}`.trim() || task.assignees[0].user.username
                            : 'Assigned'}
                          {task.assignees.length > 1 && ` +${task.assignees.length - 1}`}
                        </span>
                      ) : (
                        <span className="text-zinc-600 italic">Unassigned</span>
                      )}
                    </td>
                    <td className="py-2.5 px-4 font-mono text-zinc-400">
                      {task.storyPoints ?? '-'}
                    </td>
                    <td className="py-2.5 px-4 text-zinc-500 text-[11px]">
                      {task.dueDate
                        ? new Date(task.dueDate).toLocaleDateString()
                        : '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="flex-1 min-h-0 overflow-hidden flex flex-col">
            <TaskGanttView
              tasks={filteredTasks}
              canEdit={isMember}
              onAddTask={() => handleOpenCreateTask('TODO')}
            />
          </div>
        )}
      </div>

      {/* Task Creation Modal with pre-filled sprintId */}
      <CreateTaskModal
        isOpen={isCreateTaskOpen}
        onClose={() => setIsCreateTaskOpen(false)}
        projectId={project.id}
        defaultCategory={createCategory}
      />

      {/* CSV Import Modal with sprintId */}
      <CsvImportModal
        isOpen={isCsvImportOpen}
        onClose={() => setIsCsvImportOpen(false)}
        projectId={project.id}
        sprintId={sprint.id}
      />

      {/* Global Task Detail Drawer */}
      <TaskDetailDrawer />
    </div>
  );
}
