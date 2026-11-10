'use client';

import React, { useState, useMemo } from 'react';
import {
  Target,
  Play,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Plus,
  Search,
  MoveRight,
  User,
  AlertCircle,
  Calendar,
  LineChart,
} from 'lucide-react';
import { toast } from 'sonner';
import type { Sprint, Task, Project } from '@orbito/types';
import { useAppDispatch } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';
import { useGetTasksQuery, useUpdateTaskMutation } from '@/lib/store/api/tasksApi';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu';

interface SprintPlanningViewProps {
  project: Project;
  sprints: Sprint[];
  canManage: boolean;
  onStartSprint: (sprint: Sprint) => void;
  onCompleteSprint: (sprint: Sprint) => void;
  onSwitchToBurndown?: (sprintId: string) => void;
  onOpenCreateTask?: (sprintId?: string) => void;
}

export const SprintPlanningView: React.FC<SprintPlanningViewProps> = ({
  project,
  sprints,
  canManage,
  onStartSprint,
  onCompleteSprint,
  onSwitchToBurndown,
  onOpenCreateTask,
}) => {
  const dispatch = useAppDispatch();
  const [backlogSearch, setBacklogSearch] = useState('');
  const [collapsedSprints, setCollapsedSprints] = useState<Record<string, boolean>>({});

  // Fetch all tasks for this project
  const { data: allTasks = [], isLoading: isTasksLoading } = useGetTasksQuery(
    project?.id ? { projectId: project.id } : undefined,
    { skip: !project?.id }
  );

  const [updateTask, { isLoading: isMovingTask }] = useUpdateTaskMutation();

  // Partition tasks by sprintId
  const { sprintTasksMap, backlogTasks } = useMemo(() => {
    const map: Record<string, Task[]> = {};
    const backlog: Task[] = [];

    allTasks.forEach((task) => {
      if (task.sprintId) {
        if (!map[task.sprintId]) map[task.sprintId] = [];
        map[task.sprintId].push(task);
      } else {
        backlog.push(task);
      }
    });

    return { sprintTasksMap: map, backlogTasks: backlog };
  }, [allTasks]);

  // Filtered backlog
  const filteredBacklog = useMemo(() => {
    if (!backlogSearch.trim()) return backlogTasks;
    const q = backlogSearch.toLowerCase();
    return backlogTasks.filter(
      (t) =>
        t.title.toLowerCase().includes(q) ||
        (t.slug && t.slug.toLowerCase().includes(q))
    );
  }, [backlogTasks, backlogSearch]);

  const toggleSprintCollapse = (sprintId: string) => {
    setCollapsedSprints((prev) => ({
      ...prev,
      [sprintId]: !prev[sprintId],
    }));
  };

  const handleMoveTask = async (taskId: string, targetSprintId: string | null) => {
    try {
      await updateTask({
        id: taskId,
        updates: { sprintId: targetSprintId as any },
      }).unwrap();

      toast.success(
        targetSprintId
          ? 'Task assigned to sprint'
          : 'Task moved to backlog'
      );
    } catch {
      toast.error('Failed to move task');
    }
  };

  const getPriorityBadgeClass = (priority?: string) => {
    switch (priority) {
      case 'URGENT':
        return 'border-rose-900 bg-rose-950/40 text-rose-400';
      case 'HIGH':
        return 'border-amber-900 bg-amber-950/40 text-amber-400';
      case 'MEDIUM':
        return 'border-sky-900 bg-sky-950/40 text-sky-400';
      default:
        return 'border-zinc-800 bg-zinc-900 text-zinc-400';
    }
  };

  const getSprintStatusBadge = (status: string) => {
    switch (status) {
      case 'ACTIVE':
        return (
          <Badge variant="outline" className="text-[10px] border-emerald-800 text-emerald-400 bg-emerald-950/40">
            Active
          </Badge>
        );
      case 'COMPLETED':
        return (
          <Badge variant="outline" className="text-[10px] border-sky-800 text-sky-400 bg-sky-950/40">
            Completed
          </Badge>
        );
      default:
        return (
          <Badge variant="outline" className="text-[10px] border-zinc-700 text-zinc-300 bg-zinc-900">
            Planning
          </Badge>
        );
    }
  };

  // Group sprints by status
  const sortedSprints = useMemo(() => {
    const active = sprints.filter((s) => s.status === 'ACTIVE');
    const planning = sprints.filter((s) => s.status === 'PLANNING' || s.status === ('PLANNED' as any));
    const completed = sprints.filter((s) => s.status === 'COMPLETED');
    return [...active, ...planning, ...completed];
  }, [sprints]);

  return (
    <div className="flex flex-col lg:flex-row gap-6 items-start">
      {/* Left / Main Column: Sprints List */}
      <div className="flex-1 w-full space-y-4">
        <div className="flex items-center justify-between pb-1">
          <div className="flex items-center gap-2">
            <Target className="h-4 w-4 text-zinc-400" />
            <h2 className="text-sm font-semibold text-zinc-100">Sprint Iterations</h2>
            <span className="text-[11px] font-mono text-zinc-500">
              ({sprints.length} sprints)
            </span>
          </div>
        </div>

        {sortedSprints.length === 0 ? (
          <Card className="border-zinc-800 bg-zinc-950/40 p-8 text-center text-xs text-zinc-500 italic">
            No sprints created yet. Create a sprint to start planning your iteration.
          </Card>
        ) : (
          sortedSprints.map((sprint) => {
            const isCollapsed = !!collapsedSprints[sprint.id];
            const tasksInSprint = sprintTasksMap[sprint.id] || [];
            const totalPoints = tasksInSprint.reduce(
              (sum, t) => sum + (t.storyPoints || 0),
              0
            );
            const completedTasks = tasksInSprint.filter(
              (t) => t.status?.category === 'DONE'
            ).length;

            return (
              <div
                key={sprint.id}
                className="rounded-md border border-zinc-800 bg-zinc-950/80 shadow-sm overflow-hidden"
              >
                {/* Sprint Header */}
                <div
                  onClick={() => toggleSprintCollapse(sprint.id)}
                  className="px-4 py-3 bg-zinc-900/40 hover:bg-zinc-900/60 border-b border-zinc-800/80 flex items-center justify-between gap-3 cursor-pointer select-none transition"
                >
                  <div className="flex items-center gap-2.5 truncate">
                    <button
                      type="button"
                      className="text-zinc-500 hover:text-zinc-200"
                    >
                      {isCollapsed ? (
                        <ChevronRight className="h-4 w-4" />
                      ) : (
                        <ChevronDown className="h-4 w-4" />
                      )}
                    </button>

                    <span className="text-xs font-semibold text-zinc-100 truncate">
                      {sprint.name}
                    </span>

                    {getSprintStatusBadge(sprint.status)}

                    {sprint.goal && (
                      <span className="hidden sm:inline text-[11px] text-zinc-400 truncate max-w-xs" title={sprint.goal}>
                        &mdash; {sprint.goal}
                      </span>
                    )}
                  </div>

                  {/* Summary Badges & Action Buttons */}
                  <div
                    className="flex items-center gap-2 shrink-0"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <Badge variant="outline" className="text-[10px] font-mono border-zinc-800 text-zinc-400">
                      {tasksInSprint.length} tasks {totalPoints > 0 ? `• ${totalPoints} pts` : ''}
                    </Badge>

                    {sprint.status === 'ACTIVE' && (
                      <Badge variant="outline" className="text-[10px] font-mono border-emerald-900 text-emerald-400">
                        {completedTasks}/{tasksInSprint.length} Done
                      </Badge>
                    )}

                    {sprint.status === 'ACTIVE' && onSwitchToBurndown && (
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => onSwitchToBurndown(sprint.id)}
                        className="h-7 px-2 text-[11px] gap-1 border-zinc-700 bg-zinc-800 text-zinc-300 hover:text-white"
                        title="View Burndown Chart"
                      >
                        <LineChart className="h-3 w-3" />
                        <span className="hidden sm:inline">Burndown</span>
                      </Button>
                    )}

                    {canManage && sprint.status === 'PLANNING' && (
                      <Button
                        type="button"
                        variant="primary"
                        size="sm"
                        onClick={() => onStartSprint(sprint)}
                        className="h-7 px-2 text-[11px] gap-1"
                      >
                        <Play className="h-3 w-3" />
                        <span>Start</span>
                      </Button>
                    )}

                    {canManage && sprint.status === 'ACTIVE' && (
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => onCompleteSprint(sprint)}
                        className="h-7 px-2 text-[11px] gap-1 border-emerald-800 text-emerald-400 hover:bg-emerald-950/60"
                      >
                        <CheckCircle2 className="h-3 w-3" />
                        <span>Complete</span>
                      </Button>
                    )}
                  </div>
                </div>

                {/* Sprint Task Rows */}
                {!isCollapsed && (
                  <div className="divide-y divide-zinc-850/60 p-1">
                    {tasksInSprint.length === 0 ? (
                      <div className="p-4 text-center text-xs text-zinc-500 italic">
                        No tasks in this sprint. Assign tasks from the backlog on the right.
                      </div>
                    ) : (
                      tasksInSprint.map((task) => {
                        const assignee = task.assignees?.[0]?.user;
                        const initials = assignee
                          ? ((assignee.firstName?.[0] || '') + (assignee.lastName?.[0] || '')).toUpperCase() || 'U'
                          : null;

                        return (
                          <div
                            key={task.id}
                            onClick={() => dispatch(setTaskDrawerId(task.id))}
                            className="px-3 py-2 flex items-center justify-between gap-3 text-xs hover:bg-zinc-900/60 transition cursor-pointer group rounded"
                          >
                            <div className="flex items-center gap-2 truncate pr-2">
                              <span className="font-mono text-[10px] text-zinc-500 shrink-0">
                                {task.slug || 'TASK'}
                              </span>
                              <span className="text-zinc-200 truncate group-hover:text-white font-medium">
                                {task.title}
                              </span>
                            </div>

                            <div
                              className="flex items-center gap-2 shrink-0"
                              onClick={(e) => e.stopPropagation()}
                            >
                              {task.storyPoints !== null && task.storyPoints !== undefined && (
                                <Badge variant="outline" className="text-[10px] font-mono border-zinc-800 text-zinc-400">
                                  {task.storyPoints} pts
                                </Badge>
                              )}

                              <Badge
                                variant="outline"
                                className={`text-[9px] uppercase px-1 py-0 ${getPriorityBadgeClass(
                                  task.priority
                                )}`}
                              >
                                {task.priority}
                              </Badge>

                              {assignee ? (
                                <Avatar className="h-4 w-4 rounded-sm border border-zinc-800">
                                  <AvatarFallback className="text-[8px] bg-zinc-800 text-zinc-300">
                                    {initials}
                                  </AvatarFallback>
                                </Avatar>
                              ) : null}

                              {canManage && (
                                <button
                                  type="button"
                                  onClick={() => handleMoveTask(task.id, null)}
                                  disabled={isMovingTask}
                                  className="text-[10px] text-zinc-500 hover:text-amber-400 px-1 py-0.5 rounded transition"
                                  title="Send back to Backlog"
                                >
                                  Backlog
                                </button>
                              )}
                            </div>
                          </div>
                        );
                      })
                    )}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      {/* Right Column: Product Backlog Pane (Width: 380px) */}
      <div className="w-full lg:w-96 shrink-0 rounded-md border border-zinc-800 bg-zinc-950/80 p-4 space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <h3 className="text-sm font-semibold text-zinc-100">Product Backlog</h3>
            <Badge variant="outline" className="text-[10px] font-mono border-zinc-800 text-zinc-400">
              {backlogTasks.length}
            </Badge>
          </div>

          {onOpenCreateTask && canManage && (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => onOpenCreateTask(undefined)}
              className="h-7 px-2 text-[11px] gap-1 border-zinc-800 text-zinc-300 hover:text-white"
            >
              <Plus className="h-3 w-3" />
              <span>Add Task</span>
            </Button>
          )}
        </div>

        {/* Backlog Search */}
        <div className="relative">
          <Search className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-zinc-500" />
          <Input
            value={backlogSearch}
            onChange={(e) => setBacklogSearch(e.target.value)}
            placeholder="Search backlog tasks..."
            className="pl-8 bg-zinc-900 border-zinc-800 text-xs text-zinc-200 h-8"
          />
        </div>

        {/* Backlog Items List */}
        <div className="space-y-2 max-h-[calc(100vh-260px)] overflow-y-auto pr-1">
          {isTasksLoading ? (
            <div className="py-8 text-center text-xs text-zinc-500">
              Loading backlog items...
            </div>
          ) : filteredBacklog.length === 0 ? (
            <div className="p-6 border border-zinc-850 rounded bg-zinc-900/30 text-center text-xs text-zinc-500 italic">
              {backlogTasks.length === 0
                ? 'Backlog is empty. All tasks are committed to sprints!'
                : 'No backlog tasks match search.'}
            </div>
          ) : (
            filteredBacklog.map((task) => (
              <div
                key={task.id}
                onClick={() => dispatch(setTaskDrawerId(task.id))}
                className="p-3 rounded-md border border-zinc-800/80 bg-zinc-900/50 hover:bg-zinc-900 hover:border-zinc-700 transition cursor-pointer group space-y-2"
              >
                <div className="flex items-start justify-between gap-2">
                  <span className="font-mono text-[10px] text-zinc-500">
                    {task.slug || 'TASK'}
                  </span>
                  <Badge
                    variant="outline"
                    className={`text-[9px] uppercase px-1 py-0 ${getPriorityBadgeClass(
                      task.priority
                    )}`}
                  >
                    {task.priority}
                  </Badge>
                </div>

                <p className="text-xs text-zinc-200 font-medium group-hover:text-white line-clamp-2">
                  {task.title}
                </p>

                <div
                  className="flex items-center justify-between pt-1 border-t border-zinc-850/60"
                  onClick={(e) => e.stopPropagation()}
                >
                  <span className="text-[10px] font-mono text-zinc-500">
                    {task.storyPoints ? `${task.storyPoints} pts` : 'No pts'}
                  </span>

                  {canManage && (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 px-1.5 text-[10px] text-zinc-400 hover:text-white gap-1"
                        >
                          <span>Move to Sprint</span>
                          <MoveRight className="h-2.5 w-2.5" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" className="bg-zinc-900 border-zinc-800 text-xs">
                        {sprints
                          .filter((s) => s.status !== 'COMPLETED')
                          .map((s) => (
                            <DropdownMenuItem
                              key={s.id}
                              onClick={() => handleMoveTask(task.id, s.id)}
                              className="text-zinc-200 hover:bg-zinc-800 cursor-pointer text-xs"
                            >
                              {s.name} ({s.status})
                            </DropdownMenuItem>
                          ))}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};