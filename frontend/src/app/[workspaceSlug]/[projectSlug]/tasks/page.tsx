'use client';

import * as React from 'react';
import { useState, useMemo, useCallback } from 'react';
import { useParams } from 'next/navigation';
import dayjs from 'dayjs';
import {
  Kanban,
  List,
  Plus,
  Search,
  Calendar,
  GanttChartSquare,
  Upload,
  Repeat,
  Download,
  CalendarCheck,
  ChevronDown,
  ChevronRight,
  FileSpreadsheet,
  FileText,
  Shapes,
  Clock,
  Hash,
} from 'lucide-react';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppDispatch } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';
import { useGetProjectBySlugQuery } from '@/lib/store/api/projectsApi';
import {
  useGetTasksQuery,
  useGetTaskStatusesByProjectQuery,
  useGetTodaysTasksQuery,
} from '@/lib/store/api/tasksApi';
import { useGetWorkspaceMembersQuery } from '@/lib/store/api/workspacesApi';
import {
  KanbanBoard,
  CreateTaskModal,
  TaskCalendarView,
  TaskGanttView,
  CsvImportModal,
  GroupByManager,
  SortingManager,
  ColumnManager,
  BulkActionsBar,
  TodayAgendaDialog,
  exportTasksToCsv,
  exportTasksToExcel,
  DEFAULT_COLUMNS,
  type ColumnConfig,
  type GroupByField,
  type SortField,
  type SortOrder,
} from '@/features/tasks';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import type { Task, TaskType } from '@orbito/types';

interface TaskGroup {
  id: string;
  title: string;
  count: number;
  tasks: Task[];
  color?: string;
}

export default function ProjectTasksPage() {
  const params = useParams<{ workspaceSlug: string; projectSlug: string }>();
  const dispatch = useAppDispatch();
  const { workspaceSlug, projectSlug } = params;

  const { isMember } = useAuthRbac();
  const [viewMode, setViewMode] = useState<'board' | 'list' | 'calendar' | 'timeline'>('board');
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [priorityFilter, setPriorityFilter] = useState('ALL');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isCsvImportOpen, setIsCsvImportOpen] = useState(false);
  const [isTodayAgendaOpen, setIsTodayAgendaOpen] = useState(false);
  const [createCategory, setCreateCategory] = useState('TODO');

  // Advanced Tasks Organization State
  const [groupBy, setGroupBy] = useState<GroupByField>('none');
  const [sortField, setSortField] = useState<SortField>('createdAt');
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc');
  const [columns, setColumns] = useState<ColumnConfig[]>(DEFAULT_COLUMNS);
  const [selectedTaskIds, setSelectedTaskIds] = useState<string[]>([]);
  const [collapsedGroups, setCollapsedGroups] = useState<Record<string, boolean>>({});

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

  // Fetch Project Statuses
  const { data: projectStatuses = [] } = useGetTaskStatusesByProjectQuery(
    project?.id || '',
    { skip: !project?.id }
  );

  // Fetch Workspace Members for Assignment
  const { data: workspaceMembers = [] } = useGetWorkspaceMembersQuery(
    project?.workspaceId || '',
    { skip: !project?.workspaceId }
  );

  // Fetch Today's Tasks for the Agenda badge
  const { data: todayAgendaData } = useGetTodaysTasksQuery();
  const todayTasksCount = todayAgendaData?.tasks?.length || 0;

  // Filter Tasks
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

  // Sort Tasks
  const sortedTasks = useMemo(() => {
    const list = [...filteredTasks];
    list.sort((a, b) => {
      let cmp = 0;
      if (sortField === 'title') {
        cmp = (a.title || '').localeCompare(b.title || '');
      } else if (sortField === 'priority') {
        const pOrder: Record<string, number> = { URGENT: 4, HIGH: 3, MEDIUM: 2, LOW: 1 };
        cmp = (pOrder[a.priority] || 0) - (pOrder[b.priority] || 0);
      } else if (sortField === 'status') {
        const sA = a.status?.name || '';
        const sB = b.status?.name || '';
        cmp = sA.localeCompare(sB);
      } else if (sortField === 'dueDate') {
        if (!a.dueDate) return sortOrder === 'asc' ? 1 : -1;
        if (!b.dueDate) return sortOrder === 'asc' ? -1 : 1;
        cmp = new Date(a.dueDate).getTime() - new Date(b.dueDate).getTime();
      } else if (sortField === 'storyPoints') {
        cmp = (a.storyPoints || 0) - (b.storyPoints || 0);
      } else if (sortField === 'slug') {
        cmp = (a.slug || '').localeCompare(b.slug || '');
      } else {
        cmp = new Date(a.createdAt || 0).getTime() - new Date(b.createdAt || 0).getTime();
      }
      return sortOrder === 'asc' ? cmp : -cmp;
    });
    return list;
  }, [filteredTasks, sortField, sortOrder]);

  // Group Tasks
  const taskGroups = useMemo((): TaskGroup[] => {
    if (groupBy === 'none') {
      return [{ id: 'all', title: 'All Tasks', count: sortedTasks.length, tasks: sortedTasks }];
    }

    if (groupBy === 'status') {
      const map = new Map<string, { title: string; color?: string; tasks: Task[] }>();
      projectStatuses.forEach((st) => {
        map.set(st.name, { title: st.name, color: st.color || undefined, tasks: [] });
      });
      sortedTasks.forEach((t) => {
        const name = t.status?.name || 'Unassigned Status';
        if (!map.has(name)) {
          map.set(name, { title: name, color: t.status?.color || undefined, tasks: [] });
        }
        map.get(name)!.tasks.push(t);
      });
      return Array.from(map.entries()).map(([id, grp]) => ({
        id,
        title: grp.title,
        color: grp.color,
        count: grp.tasks.length,
        tasks: grp.tasks,
      }));
    }

    if (groupBy === 'priority') {
      const priorities = [
        { id: 'URGENT', title: 'Urgent', color: '#f43f5e' },
        { id: 'HIGH', title: 'High', color: '#f59e0b' },
        { id: 'MEDIUM', title: 'Medium', color: '#3b82f6' },
        { id: 'LOW', title: 'Low', color: '#10b981' },
      ];
      return priorities.map((p) => {
        const grpTasks = sortedTasks.filter((t) => t.priority === p.id);
        return {
          id: p.id,
          title: p.title,
          color: p.color,
          count: grpTasks.length,
          tasks: grpTasks,
        };
      });
    }

    if (groupBy === 'assignee') {
      const map = new Map<string, { title: string; tasks: Task[] }>();
      map.set('unassigned', { title: 'Unassigned', tasks: [] });

      sortedTasks.forEach((t) => {
        if (!t.assignees || t.assignees.length === 0) {
          map.get('unassigned')!.tasks.push(t);
        } else {
          t.assignees.forEach((a) => {
            const u = a.user;
            const name = u ? `${u.firstName || ''} ${u.lastName || ''}`.trim() || u.username || 'User' : 'Member';
            const uid = u?.id || a.userId;
            if (!map.has(uid)) {
              map.set(uid, { title: name, tasks: [] });
            }
            if (!map.get(uid)!.tasks.some((item) => item.id === t.id)) {
              map.get(uid)!.tasks.push(t);
            }
          });
        }
      });

      return Array.from(map.entries()).map(([id, grp]) => ({
        id,
        title: grp.title,
        count: grp.tasks.length,
        tasks: grp.tasks,
      }));
    }

    if (groupBy === 'type') {
      const types = ['TASK', 'BUG', 'STORY', 'EPIC', 'SUBTASK'];
      return types.map((tp) => {
        const grpTasks = sortedTasks.filter((t) => (t.taskType || 'TASK') === tp);
        return {
          id: tp,
          title: tp.charAt(0) + tp.slice(1).toLowerCase(),
          count: grpTasks.length,
          tasks: grpTasks,
        };
      });
    }

    if (groupBy === 'dueDate') {
      const overdue: Task[] = [];
      const today: Task[] = [];
      const thisWeek: Task[] = [];
      const later: Task[] = [];
      const noDate: Task[] = [];
      const now = dayjs();

      sortedTasks.forEach((t) => {
        if (!t.dueDate) {
          noDate.push(t);
          return;
        }
        const d = dayjs(t.dueDate);
        if (d.isBefore(now, 'day')) {
          overdue.push(t);
        } else if (d.isSame(now, 'day')) {
          today.push(t);
        } else if (d.isBefore(now.add(7, 'day'), 'day')) {
          thisWeek.push(t);
        } else {
          later.push(t);
        }
      });

      return [
        { id: 'overdue', title: 'Overdue', color: '#f43f5e', count: overdue.length, tasks: overdue },
        { id: 'today', title: 'Due Today', color: '#10b981', count: today.length, tasks: today },
        { id: 'thisWeek', title: 'Due This Week', color: '#3b82f6', count: thisWeek.length, tasks: thisWeek },
        { id: 'later', title: 'Due Later', color: '#a1a1aa', count: later.length, tasks: later },
        { id: 'noDate', title: 'No Due Date', color: '#71717a', count: noDate.length, tasks: noDate },
      ].filter((g) => g.count > 0 || g.id === 'today');
    }

    return [{ id: 'all', title: 'All Tasks', count: sortedTasks.length, tasks: sortedTasks }];
  }, [groupBy, sortedTasks, projectStatuses]);

  // Column management
  const handleToggleColumn = useCallback((columnId: string) => {
    setColumns((prev) =>
      prev.map((col) => (col.id === columnId ? { ...col, visible: !col.visible } : col))
    );
  }, []);

  const handleResetColumns = useCallback(() => {
    setColumns(DEFAULT_COLUMNS);
  }, []);

  // Selection handlers for Bulk Actions
  const handleToggleTaskSelect = (taskId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setSelectedTaskIds((prev) =>
      prev.includes(taskId) ? prev.filter((id) => id !== taskId) : [...prev, taskId]
    );
  };

  const handleSelectAllVisible = () => {
    if (selectedTaskIds.length >= filteredTasks.length) {
      setSelectedTaskIds([]);
    } else {
      setSelectedTaskIds(filteredTasks.map((t) => t.id));
    }
  };

  const toggleGroupCollapse = (groupId: string) => {
    setCollapsedGroups((prev) => ({ ...prev, [groupId]: !prev[groupId] }));
  };

  const handleOpenCreateModal = (category = 'TODO') => {
    setCreateCategory(category);
    setIsCreateModalOpen(true);
  };

  const handleTaskClick = (taskId: string) => {
    dispatch(setTaskDrawerId(taskId));
  };

  const visibleColumns = useMemo(() => columns.filter((c) => c.visible), [columns]);

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
    <div className="flex-1 flex flex-col min-h-0 bg-[#09090b] p-6 space-y-4 overflow-hidden relative">
      {/* Control Bar: Search, Status Filter, Priority Filter, GroupBy, Sorting, Columns, View Modes */}
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-3 shrink-0">
        <div className="flex flex-wrap items-center gap-2 flex-1">
          {/* Search Input */}
          <div className="relative min-w-[180px] max-w-xs flex-1">
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

          {/* GroupBy Manager */}
          <GroupByManager groupBy={groupBy} onGroupByChange={setGroupBy} />

          {/* Sorting Manager */}
          <SortingManager
            sortField={sortField}
            sortOrder={sortOrder}
            onSortFieldChange={setSortField}
            onSortOrderChange={setSortOrder}
            onResetSort={() => {
              setSortField('createdAt');
              setSortOrder('desc');
            }}
          />

          {/* Column Manager (for List view) */}
          {viewMode === 'list' && (
            <ColumnManager
              columns={columns}
              onToggleColumn={handleToggleColumn}
              onResetColumns={handleResetColumns}
            />
          )}

          {/* View Mode Toggle: Board vs List vs Calendar vs Timeline */}
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
            <button
              type="button"
              onClick={() => setViewMode('calendar')}
              className={`p-1.5 rounded-md transition-colors ${
                viewMode === 'calendar'
                  ? 'bg-zinc-800 text-zinc-100 shadow-xs'
                  : 'text-zinc-500 hover:text-zinc-300'
              }`}
              title="Calendar View"
            >
              <Calendar className="h-3.5 w-3.5" />
            </button>
            <button
              type="button"
              onClick={() => setViewMode('timeline')}
              className={`p-1.5 rounded-md transition-colors ${
                viewMode === 'timeline'
                  ? 'bg-zinc-800 text-zinc-100 shadow-xs'
                  : 'text-zinc-500 hover:text-zinc-300'
              }`}
              title="Gantt / Timeline View"
            >
              <GanttChartSquare className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>

        {/* Action Buttons: Agenda, Export, Import CSV & New Task */}
        <div className="flex items-center gap-2 shrink-0">
          {/* Today Agenda Button */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => setIsTodayAgendaOpen(true)}
            className="gap-1.5 text-xs h-8 border-zinc-800 bg-zinc-950/60 text-zinc-300 hover:bg-zinc-900 hover:text-zinc-100 cursor-pointer relative"
            title="Today's Agenda"
          >
            <CalendarCheck className="h-3.5 w-3.5 text-zinc-400" />
            <span className="hidden sm:inline">Agenda</span>
            {todayTasksCount > 0 && (
              <span className="flex items-center justify-center bg-zinc-100 text-zinc-950 text-[10px] font-bold h-4 w-4 rounded-full">
                {todayTasksCount}
              </span>
            )}
          </Button>

          {/* Export Dropdown */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className="gap-1.5 text-xs h-8 border-zinc-800 bg-zinc-950/60 text-zinc-300 hover:bg-zinc-900 hover:text-zinc-100 cursor-pointer"
                title="Export Tasks"
              >
                <Download className="h-3.5 w-3.5" />
                <span className="hidden sm:inline">Export</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              align="end"
              className="w-48 bg-zinc-900 border-zinc-800 text-zinc-200 p-1 rounded-md shadow-xl"
            >
              <DropdownMenuLabel className="text-[10px] uppercase font-bold text-zinc-500 tracking-wider px-2 py-1">
                Export Options
              </DropdownMenuLabel>
              <DropdownMenuSeparator className="bg-zinc-800" />
              <DropdownMenuItem
                onClick={() => exportTasksToCsv(sortedTasks, columns, `${project.slug}-tasks.csv`)}
                className="flex items-center gap-2 px-2 py-1.5 rounded text-xs text-zinc-300 hover:bg-zinc-800 hover:text-zinc-100 cursor-pointer"
              >
                <FileText className="h-3.5 w-3.5 text-blue-400" />
                <span>Export as CSV</span>
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => exportTasksToExcel(sortedTasks, columns, `${project.slug}-tasks.xls`)}
                className="flex items-center gap-2 px-2 py-1.5 rounded text-xs text-zinc-300 hover:bg-zinc-800 hover:text-zinc-100 cursor-pointer"
              >
                <FileSpreadsheet className="h-3.5 w-3.5 text-emerald-400" />
                <span>Export as Excel (.xls)</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          {isMember && (
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsCsvImportOpen(true)}
                className="gap-1.5 text-xs h-8 border-zinc-800 bg-zinc-950/60 text-zinc-300 hover:bg-zinc-900 hover:text-zinc-100"
              >
                <Upload className="h-3.5 w-3.5" />
                <span className="hidden sm:inline">Import CSV</span>
              </Button>

              <Button
                variant="primary"
                size="sm"
                onClick={() => handleOpenCreateModal('TODO')}
                className="gap-1.5 text-xs h-8"
              >
                <Plus className="h-3.5 w-3.5" />
                <span>New Task</span>
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Main View Area */}
      <div className="flex-1 min-h-0 overflow-hidden flex flex-col">
        {viewMode === 'board' ? (
          <KanbanBoard
            tasks={sortedTasks}
            projectId={project.id}
            onAddTask={(category) => handleOpenCreateModal(category)}
          />
        ) : viewMode === 'list' ? (
          /* High-Craft Table / List View with Grouping & Dynamic Columns */
          <div className="flex-1 overflow-x-auto overflow-y-auto border border-zinc-800 rounded-md bg-zinc-950/40 divide-y divide-zinc-850">
            {/* Table Header (only shown if not grouped or as global header) */}
            <div className="sticky top-0 z-20 flex items-center bg-zinc-900/95 backdrop-blur-xs border-b border-zinc-800 px-3 py-2 text-[11px] font-semibold text-zinc-400 uppercase tracking-wider select-none min-w-[750px]">
              {/* Checkbox column */}
              <div className="w-8 shrink-0 flex items-center justify-center">
                <input
                  type="checkbox"
                  checked={filteredTasks.length > 0 && selectedTaskIds.length === filteredTasks.length}
                  onChange={handleSelectAllVisible}
                  className="rounded border-zinc-700 bg-zinc-800 text-indigo-600 focus:ring-0 cursor-pointer h-3.5 w-3.5"
                  title="Select all tasks"
                />
              </div>

              {/* Dynamic Header Columns */}
              {visibleColumns.map((col) => (
                <div
                  key={col.id}
                  style={{ width: col.minWidth ? `${col.minWidth}px` : undefined }}
                  className={`px-2 shrink-0 ${col.id === 'title' ? 'flex-1' : ''}`}
                >
                  {col.label}
                </div>
              ))}
            </div>

            {/* Empty State */}
            {sortedTasks.length === 0 ? (
              <div className="p-12 text-center text-xs text-zinc-500">
                No tasks match the filter criteria.
              </div>
            ) : (
              /* Groups */
              taskGroups.map((group) => {
                const isCollapsed = collapsedGroups[group.id];

                return (
                  <div key={group.id} className="min-w-[750px]">
                    {/* Group Banner (only when grouped) */}
                    {groupBy !== 'none' && (
                      <div
                        onClick={() => toggleGroupCollapse(group.id)}
                        className="flex items-center gap-2 px-3 py-2 bg-zinc-900/60 border-b border-zinc-850 text-xs font-semibold text-zinc-300 cursor-pointer hover:bg-zinc-900 transition-colors select-none"
                      >
                        {isCollapsed ? (
                          <ChevronRight className="h-3.5 w-3.5 text-zinc-500" />
                        ) : (
                          <ChevronDown className="h-3.5 w-3.5 text-zinc-500" />
                        )}
                        {group.color && (
                          <span
                            className="h-2 w-2 rounded-full shrink-0"
                            style={{ backgroundColor: group.color }}
                          />
                        )}
                        <span className="truncate">{group.title}</span>
                        <Badge variant="outline" className="text-[10px] text-zinc-400 border-zinc-800 font-mono px-1.5 py-0 ml-1">
                          {group.count}
                        </Badge>
                      </div>
                    )}

                    {/* Group Tasks Rows */}
                    {!isCollapsed && (
                      <div className="divide-y divide-zinc-850">
                        {group.tasks.map((task) => {
                          const isSelected = selectedTaskIds.includes(task.id);
                          const primaryAssignee = task.assignees?.[0]?.user;
                          const initials = primaryAssignee
                            ? ((primaryAssignee.firstName?.[0] || '') + (primaryAssignee.lastName?.[0] || '')).toUpperCase() || 'U'
                            : null;

                          return (
                            <div
                              key={task.id}
                              onClick={() => handleTaskClick(task.id)}
                              className={`flex items-center px-3 py-2.5 transition-colors cursor-pointer group select-none min-w-[750px] ${
                                isSelected
                                  ? 'bg-zinc-850/70 hover:bg-zinc-850'
                                  : 'hover:bg-zinc-900/60'
                              }`}
                            >
                              {/* Row Checkbox */}
                              <div
                                className="w-8 shrink-0 flex items-center justify-center"
                                onClick={(e) => handleToggleTaskSelect(task.id, e)}
                              >
                                <input
                                  type="checkbox"
                                  checked={isSelected}
                                  onChange={() => {}}
                                  className="rounded border-zinc-700 bg-zinc-800 text-zinc-100 accent-zinc-200 focus:ring-0 cursor-pointer h-3.5 w-3.5"
                                />
                              </div>

                              {/* Dynamic Columns */}
                              {visibleColumns.map((col) => {
                                switch (col.id) {
                                  case 'slug':
                                    return (
                                      <div
                                        key={col.id}
                                        style={{ width: `${col.minWidth || 100}px` }}
                                        className="px-2 font-mono text-xs text-zinc-500 uppercase tracking-wider shrink-0 truncate"
                                      >
                                        {task.slug || `TASK-${task.taskNumber || task.id.slice(0, 4)}`}
                                      </div>
                                    );

                                  case 'type':
                                    return (
                                      <div
                                        key={col.id}
                                        style={{ width: `${col.minWidth || 90}px` }}
                                        className="px-2 shrink-0 truncate flex items-center gap-1 text-[11px] text-zinc-400 capitalize"
                                      >
                                        <Shapes className="h-3 w-3 text-zinc-500" />
                                        <span>{(task.taskType || 'TASK').toLowerCase()}</span>
                                      </div>
                                    );

                                  case 'title':
                                    return (
                                      <div key={col.id} className="px-2 flex-1 min-w-0 flex items-center gap-2">
                                        <span className="text-xs font-medium text-zinc-200 group-hover:text-white transition-colors truncate">
                                          {task.title}
                                        </span>
                                        {task.isRecurring && (
                                          <span
                                            title="Recurring task"
                                            className="inline-flex items-center gap-1 text-[9px] text-cyan-400 bg-cyan-950/40 border border-cyan-800/40 px-1 py-0.5 rounded-xs font-mono shrink-0"
                                          >
                                            <Repeat className="w-2.5 h-2.5" />
                                          </span>
                                        )}
                                      </div>
                                    );

                                  case 'status':
                                    return (
                                      <div
                                        key={col.id}
                                        style={{ width: `${col.minWidth || 120}px` }}
                                        className="px-2 shrink-0"
                                      >
                                        <Badge
                                          variant="outline"
                                          className="text-[10px] uppercase font-mono max-w-[110px] truncate"
                                          style={
                                            task.status?.color
                                              ? { borderColor: `${task.status.color}40`, color: task.status.color }
                                              : undefined
                                          }
                                        >
                                          {task.status?.name || 'TODO'}
                                        </Badge>
                                      </div>
                                    );

                                  case 'priority':
                                    return (
                                      <div
                                        key={col.id}
                                        style={{ width: `${col.minWidth || 100}px` }}
                                        className="px-2 shrink-0"
                                      >
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
                                    );

                                  case 'assignees':
                                    return (
                                      <div
                                        key={col.id}
                                        style={{ width: `${col.minWidth || 140}px` }}
                                        className="px-2 shrink-0 truncate"
                                      >
                                        {primaryAssignee ? (
                                          <div className="flex items-center gap-1.5 text-zinc-400 text-[11px] truncate">
                                            <Avatar className="h-5 w-5 rounded-sm border border-zinc-800 shrink-0">
                                              <AvatarFallback className="text-[9px] bg-zinc-800 text-zinc-300">
                                                {initials}
                                              </AvatarFallback>
                                            </Avatar>
                                            <span className="truncate max-w-[100px]">
                                              {primaryAssignee.firstName || primaryAssignee.username}
                                            </span>
                                          </div>
                                        ) : (
                                          <span className="text-[11px] text-zinc-600 italic">Unassigned</span>
                                        )}
                                      </div>
                                    );

                                  case 'sprint':
                                    return (
                                      <div
                                        key={col.id}
                                        style={{ width: `${col.minWidth || 120}px` }}
                                        className="px-2 shrink-0 text-xs text-zinc-400 truncate"
                                      >
                                        {task.sprint?.name || <span className="text-zinc-600 italic">Backlog</span>}
                                      </div>
                                    );

                                  case 'storyPoints':
                                    return (
                                      <div
                                        key={col.id}
                                        style={{ width: `${col.minWidth || 80}px` }}
                                        className="px-2 shrink-0 font-mono text-xs text-zinc-400 flex items-center gap-1"
                                      >
                                        {task.storyPoints != null ? (
                                          <>
                                            <Hash className="h-3 w-3 text-zinc-500" />
                                            <span>{task.storyPoints}</span>
                                          </>
                                        ) : (
                                          <span className="text-zinc-600">-</span>
                                        )}
                                      </div>
                                    );

                                  case 'dueDate':
                                    return (
                                      <div
                                        key={col.id}
                                        style={{ width: `${col.minWidth || 130}px` }}
                                        className="px-2 shrink-0 font-mono text-[11px] text-zinc-400 truncate flex items-center gap-1"
                                      >
                                        {task.dueDate ? (
                                          <>
                                            <Clock className="h-3 w-3 text-zinc-500 shrink-0" />
                                            <span>{dayjs(task.dueDate).format('MMM D, YYYY')}</span>
                                          </>
                                        ) : (
                                          <span className="text-zinc-600">-</span>
                                        )}
                                      </div>
                                    );

                                  case 'timeTracking':
                                    return (
                                      <div
                                        key={col.id}
                                        style={{ width: `${col.minWidth || 110}px` }}
                                        className="px-2 shrink-0 font-mono text-[11px] text-zinc-400 truncate"
                                      >
                                        {task.loggedHours || 0}h / {task.estimatedHours || 0}h
                                      </div>
                                    );

                                  case 'recurrence':
                                    return (
                                      <div
                                        key={col.id}
                                        style={{ width: `${col.minWidth || 100}px` }}
                                        className="px-2 shrink-0 text-[11px] text-zinc-400 truncate"
                                      >
                                        {task.isRecurring ? 'Recurring' : '-'}
                                      </div>
                                    );

                                  case 'createdAt':
                                    return (
                                      <div
                                        key={col.id}
                                        style={{ width: `${col.minWidth || 110}px` }}
                                        className="px-2 shrink-0 font-mono text-[11px] text-zinc-400 truncate"
                                      >
                                        {dayjs(task.createdAt).format('MMM D, YYYY')}
                                      </div>
                                    );

                                  default:
                                    return null;
                                }
                              })}
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </div>
        ) : viewMode === 'calendar' ? (
          <TaskCalendarView
            tasks={sortedTasks}
            canEdit={isMember}
            onOpenCreateTask={() => handleOpenCreateModal('TODO')}
          />
        ) : (
          <TaskGanttView
            tasks={sortedTasks}
            canEdit={isMember}
            onAddTask={() => handleOpenCreateModal('TODO')}
          />
        )}
      </div>

      {/* Bulk Actions Floating Bar */}
      <BulkActionsBar
        selectedCount={selectedTaskIds.length}
        totalCount={filteredTasks.length}
        selectedTaskIds={selectedTaskIds}
        projectId={project.id}
        availableStatuses={projectStatuses}
        availableMembers={workspaceMembers}
        onSelectAll={handleSelectAllVisible}
        onClearSelection={() => setSelectedTaskIds([])}
        canManage={isMember}
      />

      {/* Today's Agenda Modal */}
      <TodayAgendaDialog
        isOpen={isTodayAgendaOpen}
        onClose={() => setIsTodayAgendaOpen(false)}
        workspaceSlug={workspaceSlug}
        projectSlug={projectSlug}
        onOpenTask={handleTaskClick}
      />

      {/* Task Creation Modal */}
      <CreateTaskModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        projectId={project.id}
        defaultCategory={createCategory}
      />

      {/* CSV Task Importer Modal */}
      <CsvImportModal
        isOpen={isCsvImportOpen}
        onClose={() => setIsCsvImportOpen(false)}
        projectId={project.id}
      />
    </div>
  );
}
