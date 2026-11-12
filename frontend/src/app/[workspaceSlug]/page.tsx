'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import {
  FolderGit2,
  CheckCircle2,
  Clock,
  Plus,
  Users,
  ArrowRight,
  Building2,
  CalendarCheck,
  AlertCircle,
  TrendingUp,
  Activity,
  UserPlus,
  Sparkles,
  Layers,
  Flame,
} from 'lucide-react';

dayjs.extend(relativeTime);

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppDispatch } from '@/lib/store/hooks';
import { setActiveProject, setActiveProjectSlug } from '@/lib/store/slices/authSlice';
import {
  useGetWorkspaceBySlugQuery,
  useGetWorkspaceMembersQuery,
} from '@/lib/store/api/workspacesApi';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import { useGetTasksQuery } from '@/lib/store/api/tasksApi';
import { useGetActivityLogsQuery } from '@/lib/store/api/activityLogApi';
import { CreateProjectModal } from '@/features/projects';
import { InviteMemberModal } from '@/components/modals/InviteMemberModal';
import { TodayAgendaDialog } from '@/features/tasks';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

export default function WorkspaceOverviewPage() {
  const params = useParams<{ workspaceSlug: string }>();
  const router = useRouter();
  const dispatch = useAppDispatch();
  const workspaceSlug = params.workspaceSlug;
  const { user, isManager, isOwner, roleBadge } = useAuthRbac();

  const [isCreateProjectOpen, setIsCreateProjectOpen] = useState(false);
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const [isTodayAgendaOpen, setIsTodayAgendaOpen] = useState(false);

  // Reset active project when viewing workspace overview
  React.useEffect(() => {
    dispatch(setActiveProject(null));
    dispatch(setActiveProjectSlug(null));
  }, [dispatch]);

  // Fetch workspace by slug
  const {
    data: workspace,
    isLoading: isWsLoading,
    isError: isWsError,
  } = useGetWorkspaceBySlugQuery(workspaceSlug || '', {
    skip: !workspaceSlug,
  });

  // Fetch projects in this workspace
  const { data: projects = [], isLoading: isProjectsLoading } =
    useGetProjectsQuery(workspace?.id ? { workspaceId: workspace.id } : undefined, {
      skip: !workspace?.id,
    });

  // Fetch workspace members for real count
  const { data: members = [] } = useGetWorkspaceMembersQuery(workspace?.id || '', {
    skip: !workspace?.id,
  });

  // Fetch ALL tasks in this workspace across all projects
  const { data: tasks = [], isLoading: isTasksLoading } = useGetTasksQuery(
    workspace?.id ? { workspaceId: workspace.id } : undefined,
    { skip: !workspace?.id }
  );

  // Fetch recent workspace activity
  const { data: recentLogs = [] } = useGetActivityLogsQuery({ limit: 4 });

  // Workspace-wide metric calculations
  const metrics = useMemo(() => {
    const total = tasks.length;
    const completed = tasks.filter((t) => t.status?.category === 'DONE').length;
    const inProgress = tasks.filter((t) => t.status?.category === 'IN_PROGRESS').length;
    const inReview = tasks.filter((t) => t.status?.category === 'IN_REVIEW').length;
    const todo = tasks.filter((t) => t.status?.category === 'TODO' || !t.status).length;
    const overdue = tasks.filter(
      (t) => t.dueDate && dayjs(t.dueDate).isBefore(dayjs(), 'day') && t.status?.category !== 'DONE'
    ).length;
    const rate = total > 0 ? Math.round((completed / total) * 100) : 0;

    const urgent = tasks.filter((t) => t.priority === 'URGENT').length;
    const high = tasks.filter((t) => t.priority === 'HIGH').length;
    const medium = tasks.filter((t) => t.priority === 'MEDIUM').length;
    const low = tasks.filter((t) => t.priority === 'LOW').length;

    return {
      total,
      completed,
      inProgress,
      inReview,
      todo,
      overdue,
      rate,
      urgent,
      high,
      medium,
      low,
    };
  }, [tasks]);

  if (isWsLoading) {
    return (
      <div className="flex h-screen items-center justify-center bg-[#09090b]">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
      </div>
    );
  }

  if (isWsError || !workspace) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center p-6 bg-[#09090b]">
        <Card className="max-w-md w-full border-zinc-800 bg-zinc-900 text-center p-6 space-y-4">
          <Building2 className="h-10 w-10 text-zinc-500 mx-auto" />
          <h3 className="text-lg font-bold text-zinc-100">Workspace Not Found</h3>
          <p className="text-xs text-zinc-400">
            No workspace matching slug <strong className="text-zinc-200">{workspaceSlug}</strong> was found.
          </p>
          <Link href="/">
            <Button variant="primary" size="sm" className="w-full">
              Go to Default Workspace
            </Button>
          </Link>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Workspace Header & Action Hub */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5">
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              {workspace.name}
            </h1>
            <Badge className={`text-[10px] uppercase ${roleBadge.className}`}>
              {roleBadge.label}
            </Badge>
          </div>
          <p className="text-xs text-zinc-400">
            {workspace.description || 'Workspace overview, project deliverables, and team iterations.'}
          </p>
        </div>

        {/* Header Action Buttons */}
        <div className="flex flex-wrap items-center gap-2">
          {/* Today's Agenda Button */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => setIsTodayAgendaOpen(true)}
            className="gap-1.5 text-xs border-zinc-800 bg-zinc-900/60 hover:bg-zinc-800 text-zinc-200 cursor-pointer"
          >
            <CalendarCheck className="h-3.5 w-3.5 text-indigo-400" />
            <span>Today&apos;s Agenda</span>
          </Button>

          {/* Invite Member Button */}
          {(isManager || isOwner) && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setIsInviteOpen(true)}
              className="gap-1.5 text-xs border-zinc-800 bg-zinc-900/60 hover:bg-zinc-800 text-zinc-200 cursor-pointer"
            >
              <UserPlus className="h-3.5 w-3.5 text-zinc-400" />
              <span>Invite</span>
            </Button>
          )}

          {/* New Project Button */}
          {(isManager || isOwner) && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setIsCreateProjectOpen(true)}
              className="gap-1.5 text-xs border-zinc-800 bg-zinc-900/60 hover:bg-zinc-800 text-zinc-200 cursor-pointer"
            >
              <Plus className="h-3.5 w-3.5" />
              <span>New Project</span>
            </Button>
          )}

          {/* Jump to first project or projects catalog */}
          {projects[0] ? (
            <Link href={`/${workspaceSlug}/${projects[0].slug}/tasks`}>
              <Button variant="primary" size="sm" className="gap-2 text-xs">
                <span>Open Tasks Board</span>
                <ArrowRight className="h-3.5 w-3.5" />
              </Button>
            </Link>
          ) : (
            <Link href={`/${workspaceSlug}/projects`}>
              <Button variant="primary" size="sm" className="gap-2 text-xs">
                <span>View Projects</span>
                <ArrowRight className="h-3.5 w-3.5" />
              </Button>
            </Link>
          )}
        </div>
      </div>

      {/* Primary KPI Metrics Row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Total Projects */}
        <Link href={`/${workspaceSlug}/projects`} className="block group">
          <Card className="border-zinc-800 bg-zinc-900/60 p-5 group-hover:border-zinc-700 transition-colors h-full">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Total Projects
              </span>
              <FolderGit2 className="h-4 w-4 text-zinc-400 group-hover:text-zinc-200 transition-colors" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">{projects.length}</div>
            <p className="text-[11px] text-zinc-500 mt-1 flex items-center gap-1">
              <span>View catalog</span>
              <ArrowRight className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity" />
            </p>
          </Card>
        </Link>

        {/* Team Members */}
        <Link href={`/${workspaceSlug}/members`} className="block group">
          <Card className="border-zinc-800 bg-zinc-900/60 p-5 group-hover:border-zinc-700 transition-colors h-full">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Team Members
              </span>
              <Users className="h-4 w-4 text-zinc-400 group-hover:text-zinc-200 transition-colors" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">{members.length}</div>
            <p className="text-[11px] text-zinc-500 mt-1 flex items-center gap-1">
              <span>Manage directory</span>
              <ArrowRight className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity" />
            </p>
          </Card>
        </Link>

        {/* In Progress Tasks */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-5">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
              In Progress
            </span>
            <Clock className="h-4 w-4 text-amber-400" />
          </div>
          <div className="text-2xl font-bold text-amber-400">{metrics.inProgress}</div>
          <p className="text-[11px] text-zinc-500 mt-1">Across all workspace projects</p>
        </Card>

        {/* Completion Rate / Delivered */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-5">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
              Completion Rate
            </span>
            <TrendingUp className="h-4 w-4 text-emerald-400" />
          </div>
          <div className="flex items-baseline gap-2">
            <div className="text-2xl font-bold text-emerald-400">{metrics.rate}%</div>
            <span className="text-xs text-zinc-500 font-mono">
              ({metrics.completed}/{metrics.total})
            </span>
          </div>
          {/* Mini Progress Bar */}
          <div className="w-full bg-zinc-800 h-1.5 rounded-full mt-2.5 overflow-hidden">
            <div
              className="bg-emerald-500 h-full rounded-full transition-all duration-500"
              style={{ width: `${metrics.rate}%` }}
            />
          </div>
        </Card>
      </div>

      {/* Secondary Analytics: Workload Distribution & Overdue Alert */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Status Distribution */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Layers className="h-4 w-4 text-zinc-400" />
              <h3 className="text-sm font-semibold text-zinc-100">Workflow Stages</h3>
            </div>
            <span className="text-xs text-zinc-500 font-mono">{metrics.total} Tasks</span>
          </div>

          <div className="space-y-3">
            {/* Done */}
            <div className="space-y-1">
              <div className="flex justify-between text-xs">
                <span className="text-zinc-300">Done</span>
                <span className="text-emerald-400 font-mono">{metrics.completed}</span>
              </div>
              <div className="w-full bg-zinc-800 h-1.5 rounded-full overflow-hidden">
                <div
                  className="bg-emerald-500 h-full rounded-full transition-all"
                  style={{ width: metrics.total > 0 ? `${(metrics.completed / metrics.total) * 100}%` : '0%' }}
                />
              </div>
            </div>

            {/* In Review */}
            <div className="space-y-1">
              <div className="flex justify-between text-xs">
                <span className="text-zinc-300">In Review</span>
                <span className="text-blue-400 font-mono">{metrics.inReview}</span>
              </div>
              <div className="w-full bg-zinc-800 h-1.5 rounded-full overflow-hidden">
                <div
                  className="bg-blue-500 h-full rounded-full transition-all"
                  style={{ width: metrics.total > 0 ? `${(metrics.inReview / metrics.total) * 100}%` : '0%' }}
                />
              </div>
            </div>

            {/* In Progress */}
            <div className="space-y-1">
              <div className="flex justify-between text-xs">
                <span className="text-zinc-300">In Progress</span>
                <span className="text-amber-400 font-mono">{metrics.inProgress}</span>
              </div>
              <div className="w-full bg-zinc-800 h-1.5 rounded-full overflow-hidden">
                <div
                  className="bg-amber-500 h-full rounded-full transition-all"
                  style={{ width: metrics.total > 0 ? `${(metrics.inProgress / metrics.total) * 100}%` : '0%' }}
                />
              </div>
            </div>

            {/* To Do */}
            <div className="space-y-1">
              <div className="flex justify-between text-xs">
                <span className="text-zinc-300">To Do / Backlog</span>
                <span className="text-zinc-400 font-mono">{metrics.todo}</span>
              </div>
              <div className="w-full bg-zinc-800 h-1.5 rounded-full overflow-hidden">
                <div
                  className="bg-zinc-600 h-full rounded-full transition-all"
                  style={{ width: metrics.total > 0 ? `${(metrics.todo / metrics.total) * 100}%` : '0%' }}
                />
              </div>
            </div>
          </div>
        </Card>

        {/* Priority Breakdown */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Flame className="h-4 w-4 text-orange-400" />
              <h3 className="text-sm font-semibold text-zinc-100">Priority Levels</h3>
            </div>
            {metrics.overdue > 0 && (
              <Badge variant="outline" className="text-[10px] text-rose-400 border-rose-800/40 bg-rose-950/20 font-mono">
                {metrics.overdue} Overdue
              </Badge>
            )}
          </div>

          <div className="space-y-3">
            {/* Urgent */}
            <div className="flex items-center justify-between p-2 rounded bg-zinc-950/40 border border-zinc-850">
              <div className="flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-rose-500" />
                <span className="text-xs text-zinc-300">Urgent</span>
              </div>
              <span className="text-xs font-mono font-bold text-rose-400">{metrics.urgent}</span>
            </div>

            {/* High */}
            <div className="flex items-center justify-between p-2 rounded bg-zinc-950/40 border border-zinc-850">
              <div className="flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-amber-500" />
                <span className="text-xs text-zinc-300">High</span>
              </div>
              <span className="text-xs font-mono font-bold text-amber-400">{metrics.high}</span>
            </div>

            {/* Medium */}
            <div className="flex items-center justify-between p-2 rounded bg-zinc-950/40 border border-zinc-850">
              <div className="flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-blue-500" />
                <span className="text-xs text-zinc-300">Medium</span>
              </div>
              <span className="text-xs font-mono font-bold text-blue-400">{metrics.medium}</span>
            </div>

            {/* Low */}
            <div className="flex items-center justify-between p-2 rounded bg-zinc-950/40 border border-zinc-850">
              <div className="flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-emerald-500" />
                <span className="text-xs text-zinc-300">Low</span>
              </div>
              <span className="text-xs font-mono font-bold text-emerald-400">{metrics.low}</span>
            </div>
          </div>
        </Card>

        {/* Recent Activity Snapshot */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-indigo-400" />
              <h3 className="text-sm font-semibold text-zinc-100">Live Activity</h3>
            </div>
            <Link
              href={`/${workspaceSlug}/activities`}
              className="text-[11px] text-zinc-400 hover:text-zinc-200 transition-colors flex items-center gap-1"
            >
              <span>View all</span>
              <ArrowRight className="h-3 w-3" />
            </Link>
          </div>

          {recentLogs.length === 0 ? (
            <div className="py-6 text-center text-xs text-zinc-500">
              No recent activity recorded
            </div>
          ) : (
            <div className="space-y-3">
              {recentLogs.slice(0, 4).map((log) => (
                <div key={log.id} className="flex items-start gap-2.5 text-xs">
                  <Avatar className="h-6 w-6 mt-0.5 rounded-sm border border-zinc-800 shrink-0">
                    <AvatarFallback className="text-[10px] bg-zinc-800 text-zinc-300">
                      {log.user?.firstName?.[0] || 'U'}
                    </AvatarFallback>
                  </Avatar>
                  <div className="min-w-0 flex-1">
                    <p className="text-zinc-200 truncate font-medium">
                      {log.description || log.type || 'Activity recorded'}
                    </p>
                    <span className="text-[10px] text-zinc-500 font-mono">
                      {dayjs(log.createdAt).fromNow()}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>

      {/* Projects Catalog Grid Section */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <h3 className="text-base font-semibold text-zinc-100">Projects</h3>
            <Badge variant="outline" className="text-[10px] font-mono">
              {projects.length} {projects.length === 1 ? 'Project' : 'Projects'}
            </Badge>
          </div>
          <div className="flex items-center gap-2">
            <Link href={`/${workspaceSlug}/projects`}>
              <Button
                variant="outline"
                size="sm"
                className="gap-1.5 text-xs border-zinc-800 bg-zinc-900/60 hover:bg-zinc-800 text-zinc-300"
              >
                <span>Projects Catalog</span>
                <ArrowRight className="h-3 w-3" />
              </Button>
            </Link>
            {(isManager || isOwner) && (
              <Button
                variant="primary"
                size="sm"
                onClick={() => setIsCreateProjectOpen(true)}
                className="gap-1.5 text-xs"
              >
                <Plus className="h-3.5 w-3.5" />
                <span>New Project</span>
              </Button>
            )}
          </div>
        </div>

        {isProjectsLoading ? (
          <div className="flex h-32 items-center justify-center">
            <div className="h-5 w-5 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
          </div>
        ) : projects.length === 0 ? (
          <Card className="border-zinc-800 bg-zinc-900/60 p-8 text-center">
            <FolderGit2 className="h-8 w-8 text-zinc-500 mx-auto mb-2" />
            <h4 className="text-sm font-semibold text-zinc-200">No projects yet</h4>
            <p className="text-xs text-zinc-500 mt-1">
              Create your first project to start organizing sprints and backlog items.
            </p>
            {(isManager || isOwner) && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsCreateProjectOpen(true)}
                className="gap-1.5 text-xs mt-3"
              >
                <Plus className="h-3.5 w-3.5" />
                <span>Create Project</span>
              </Button>
            )}
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {projects.map((project) => {
              const projectTasks = tasks.filter((t) => t.projectId === project.id);
              const pTotal = projectTasks.length;
              const pCompleted = projectTasks.filter((t) => t.status?.category === 'DONE').length;
              const pRate = pTotal > 0 ? Math.round((pCompleted / pTotal) * 100) : 0;

              return (
                <Link
                  key={project.id}
                  href={`/${workspaceSlug}/${project.slug}/tasks`}
                  className="group block"
                >
                  <Card className="border-zinc-800 bg-zinc-900/80 group-hover:border-zinc-700 transition-all p-5 space-y-3 h-full flex flex-col justify-between">
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <Badge variant="outline" className="text-[10px] font-mono uppercase">
                          {project.slug}
                        </Badge>
                        <span className="text-[11px] text-zinc-500 group-hover:text-zinc-300 flex items-center gap-1 transition-colors">
                          <span>Board</span>
                          <ArrowRight className="h-3 w-3" />
                        </span>
                      </div>
                      <div>
                        <h4 className="text-sm font-semibold text-zinc-100 group-hover:text-white transition-colors truncate">
                          {project.name}
                        </h4>
                        {project.description && (
                          <p className="text-xs text-zinc-400 mt-1 line-clamp-2">
                            {project.description}
                          </p>
                        )}
                      </div>
                    </div>

                    {/* Project Task Stats */}
                    <div className="pt-2 border-t border-zinc-850/60 space-y-1.5">
                      <div className="flex items-center justify-between text-[11px] text-zinc-400 font-mono">
                        <span>{pCompleted} of {pTotal} tasks</span>
                        <span className="text-emerald-400">{pRate}%</span>
                      </div>
                      <div className="w-full bg-zinc-800 h-1 rounded-full overflow-hidden">
                        <div
                          className="bg-emerald-500 h-full rounded-full transition-all"
                          style={{ width: `${pRate}%` }}
                        />
                      </div>
                    </div>
                  </Card>
                </Link>
              );
            })}
          </div>
        )}
      </div>

      {/* Modals */}
      {workspace?.id && (
        <>
          <CreateProjectModal
            isOpen={isCreateProjectOpen}
            onClose={() => setIsCreateProjectOpen(false)}
            workspaceId={workspace.id}
          />
          <InviteMemberModal
            open={isInviteOpen}
            onOpenChange={setIsInviteOpen}
            workspaceId={workspace.id}
            workspaceName={workspace.name}
          />
        </>
      )}

      {/* Today's Agenda Briefing Modal */}
      <TodayAgendaDialog
        isOpen={isTodayAgendaOpen}
        onClose={() => setIsTodayAgendaOpen(false)}
        workspaceSlug={workspaceSlug}
      />
    </div>
  );
}
