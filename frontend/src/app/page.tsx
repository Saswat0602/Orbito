'use client';

import * as React from 'react';
import Link from 'next/link';
import {
  ListTodo,
  CheckCircle2,
  Clock,
  AlertCircle,
  FolderGit2,
  Target,
  ArrowRight,
  TrendingUp,
} from 'lucide-react';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppSelector } from '@/lib/store/hooks';
import { useGetTasksQuery } from '@/lib/store/api/tasksApi';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import { useGetSprintsQuery } from '@/lib/store/api/sprintsApi';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

export default function DashboardOverviewPage() {
  const { user, role, roleBadge } = useAuthRbac();
  const authState = useAppSelector((state) => state.auth);
  const activeProjectId = authState.activeProjectId;
  const activeWorkspaceId = authState.activeWorkspaceId;

  // Real data queries
  const { data: tasks = [], isLoading: isTasksLoading } = useGetTasksQuery(
    activeProjectId ? { projectId: activeProjectId } : undefined
  );

  const { data: projects = [] } = useGetProjectsQuery(
    activeWorkspaceId ? { workspaceId: activeWorkspaceId } : undefined
  );

  const { data: sprints = [] } = useGetSprintsQuery(
    activeProjectId ? { projectId: activeProjectId } : undefined,
    { skip: !activeProjectId }
  );

  // Metrics computation
  const totalTasks = tasks.length;
  const completedTasks = tasks.filter((t) => t.status?.category === 'DONE').length;
  const inProgressTasks = tasks.filter((t) => t.status?.category === 'IN_PROGRESS').length;
  const urgentTasks = tasks.filter((t) => t.priority === 'URGENT' || t.priority === 'HIGH').length;

  const completionRate = totalTasks ? Math.round((completedTasks / totalTasks) * 100) : 0;

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Welcome Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5">
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Welcome back, {user?.firstName || user?.username || 'Team Member'}
            </h1>
            <Badge className={`text-[10px] uppercase ${roleBadge.className}`}>
              {roleBadge.label}
            </Badge>
          </div>
          <p className="text-xs text-zinc-400">
            Overview of project health, sprint velocity, and priority deliverables.
          </p>
        </div>

        <Link href="/tasks">
          <Button variant="primary" size="sm" className="gap-2">
            <span>View Board</span>
            <ArrowRight className="h-3.5 w-3.5" />
          </Button>
        </Link>
      </div>

      {/* Overview Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="border-zinc-800 bg-zinc-900/60 p-5">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
              Total Backlog
            </span>
            <ListTodo className="h-4 w-4 text-zinc-400" />
          </div>
          <div className="text-2xl font-bold text-zinc-100">{totalTasks}</div>
          <p className="text-[11px] text-zinc-500 mt-1">Managed tasks</p>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/60 p-5">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
              In Progress
            </span>
            <Clock className="h-4 w-4 text-amber-400" />
          </div>
          <div className="text-2xl font-bold text-amber-400">{inProgressTasks}</div>
          <p className="text-[11px] text-zinc-500 mt-1">Currently in flight</p>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/60 p-5">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
              Completed
            </span>
            <CheckCircle2 className="h-4 w-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-emerald-400">{completedTasks}</div>
          <p className="text-[11px] text-zinc-500 mt-1">{completionRate}% total completion</p>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/60 p-5">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
              High / Urgent
            </span>
            <AlertCircle className="h-4 w-4 text-rose-400" />
          </div>
          <div className="text-2xl font-bold text-rose-400">{urgentTasks}</div>
          <p className="text-[11px] text-zinc-500 mt-1">Priority items requiring focus</p>
        </Card>
      </div>

      {/* Two-Column Detail Section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Active Tasks Feed */}
        <div className="lg:col-span-2 space-y-4">
          <Card className="border-zinc-800 bg-zinc-900/80">
            <CardHeader className="flex flex-row items-center justify-between pb-3">
              <div>
                <CardTitle className="text-base">Active Deliverables</CardTitle>
                <CardDescription className="text-xs">
                  Recently updated tasks across your project.
                </CardDescription>
              </div>
              <Link href="/tasks">
                <Button variant="ghost" size="sm" className="text-xs gap-1">
                  <span>See all</span>
                  <ArrowRight className="h-3 w-3" />
                </Button>
              </Link>
            </CardHeader>
            <CardContent>
              {isTasksLoading ? (
                <div className="flex h-32 items-center justify-center">
                  <div className="h-5 w-5 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
                </div>
              ) : tasks.length === 0 ? (
                <div className="text-center py-8 text-xs text-zinc-500">
                  No tasks found. Create a task in the Tasks board.
                </div>
              ) : (
                <div className="divide-y divide-zinc-800/80">
                  {tasks.slice(0, 5).map((task) => (
                    <div
                      key={task.id}
                      className="py-3 flex items-center justify-between gap-3 hover:bg-zinc-850/40 px-2 rounded-lg transition-colors"
                    >
                      <div className="min-w-0 flex items-center gap-2.5">
                        <Badge variant="outline" className="text-[9px] font-mono shrink-0 uppercase">
                          {task.slug || 'ORB-1'}
                        </Badge>
                        <span className="truncate text-xs font-medium text-zinc-200">
                          {task.title}
                        </span>
                      </div>

                      <div className="flex items-center gap-2 shrink-0">
                        <Badge
                          variant={
                            task.priority === 'URGENT'
                              ? 'destructive'
                              : task.priority === 'HIGH'
                              ? 'default'
                              : 'secondary'
                          }
                          className="text-[9px]"
                        >
                          {task.priority}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Projects & Sprints Quick Panel */}
        <div className="space-y-4">
          <Card className="border-zinc-800 bg-zinc-900/80">
            <CardHeader className="pb-3">
              <CardTitle className="text-base flex items-center gap-2">
                <FolderGit2 className="h-4 w-4" />
                <span>Workspace Projects</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {projects.length === 0 ? (
                <div className="text-center py-4 text-xs text-zinc-500">
                  No projects configured.
                </div>
              ) : (
                projects.slice(0, 4).map((project) => (
                  <Link
                    key={project.id}
                    href="/tasks"
                    className="flex items-center justify-between p-2.5 rounded-lg bg-zinc-900 hover:bg-zinc-850 border border-zinc-850 transition-colors"
                  >
                    <div className="truncate text-xs font-medium text-zinc-200">
                      {project.name}
                    </div>
                    <Badge variant="outline" className="text-[9px] uppercase font-mono">
                      {project.slug}
                    </Badge>
                  </Link>
                ))
              )}
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/80">
            <CardHeader className="pb-3">
              <CardTitle className="text-base flex items-center gap-2">
                <Target className="h-4 w-4" />
                <span>Active Sprints</span>
              </CardTitle>
            </CardHeader>
            <CardContent>
              {sprints.length === 0 ? (
                <div className="text-center py-4 text-xs text-zinc-500">
                  No sprint in progress.
                </div>
              ) : (
                sprints.slice(0, 3).map((sprint) => (
                  <div
                    key={sprint.id}
                    className="flex items-center justify-between p-2 rounded-lg bg-zinc-900 border border-zinc-850 mb-2 last:mb-0"
                  >
                    <span className="text-xs text-zinc-200 truncate">
                      {sprint.name}
                    </span>
                    <Badge
                      variant={sprint.status === 'ACTIVE' ? 'default' : 'secondary'}
                      className="text-[9px]"
                    >
                      {sprint.status}
                    </Badge>
                  </div>
                ))
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
