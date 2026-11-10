'use client';

import * as React from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';
import {
  FolderGit2,
  ListTodo,
  CheckCircle2,
  Clock,
  AlertCircle,
  Plus,
  Users,
  Target,
  ArrowRight,
  Building2,
} from 'lucide-react';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppDispatch } from '@/lib/store/hooks';
import { setActiveProject, setActiveProjectSlug } from '@/lib/store/slices/authSlice';
import { useGetWorkspaceBySlugQuery } from '@/lib/store/api/workspacesApi';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import { useGetTasksQuery } from '@/lib/store/api/tasksApi';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

export default function WorkspaceOverviewPage() {
  const params = useParams<{ workspaceSlug: string }>();
  const router = useRouter();
  const dispatch = useAppDispatch();
  const workspaceSlug = params.workspaceSlug;
  const { user, isManager, isOwner, roleBadge } = useAuthRbac();

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

  // Fetch tasks for first project or workspace
  const firstProjectId = projects[0]?.id;
  const { data: tasks = [] } = useGetTasksQuery(
    firstProjectId ? { projectId: firstProjectId } : undefined,
    { skip: !firstProjectId }
  );

  const totalTasks = tasks.length;
  const completedTasks = tasks.filter((t) => t.status?.category === 'DONE').length;
  const inProgressTasks = tasks.filter((t) => t.status?.category === 'IN_PROGRESS').length;
  const urgentTasks = tasks.filter((t) => t.priority === 'URGENT' || t.priority === 'HIGH').length;

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
      {/* Workspace Header */}
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

        <div className="flex items-center gap-2">
          {projects[0] && (
            <Link href={`/${workspaceSlug}/${projects[0].slug}/tasks`}>
              <Button variant="primary" size="sm" className="gap-2">
                <span>Open Tasks Board</span>
                <ArrowRight className="h-3.5 w-3.5" />
              </Button>
            </Link>
          )}
        </div>
      </div>

      {/* Metrics Row */}
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
          <p className="text-[11px] text-zinc-500 mt-1">Currently active</p>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/60 p-5">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
              Completed
            </span>
            <CheckCircle2 className="h-4 w-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-emerald-400">{completedTasks}</div>
          <p className="text-[11px] text-zinc-500 mt-1">Closed items</p>
        </Card>

        <Card className="border-zinc-800 bg-zinc-900/60 p-5">
          <div className="flex items-center justify-between pb-2">
            <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
              High / Urgent
            </span>
            <AlertCircle className="h-4 w-4 text-rose-400" />
          </div>
          <div className="text-2xl font-bold text-rose-400">{urgentTasks}</div>
          <p className="text-[11px] text-zinc-500 mt-1">Priority focus</p>
        </Card>
      </div>

      {/* Projects Section */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <h3 className="text-base font-semibold text-zinc-100">Projects</h3>
            <p className="text-xs text-zinc-400">
              Active projects scoped within this workspace.
            </p>
          </div>
          {(isManager || isOwner) && (
            <Button variant="outline" size="sm" className="gap-1.5 text-xs">
              <Plus className="h-3.5 w-3.5" />
              <span>New Project</span>
            </Button>
          )}
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
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {projects.map((project) => (
              <Link
                key={project.id}
                href={`/${workspaceSlug}/${project.slug}/tasks`}
                className="group block"
              >
                <Card className="border-zinc-800 bg-zinc-900/80 group-hover:border-zinc-700 transition-all p-5 space-y-3 h-full">
                  <div className="flex items-center justify-between">
                    <Badge variant="outline" className="text-[10px] font-mono uppercase">
                      {project.slug}
                    </Badge>
                  </div>
                  <div>
                    <h4 className="text-sm font-semibold text-zinc-100 group-hover:text-white transition-colors">
                      {project.name}
                    </h4>
                    {project.description && (
                      <p className="text-xs text-zinc-400 mt-1 line-clamp-2">
                        {project.description}
                      </p>
                    )}
                  </div>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

