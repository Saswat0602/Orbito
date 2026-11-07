'use client';

import * as React from 'react';
import { useEffect } from 'react';
import { useParams, useRouter, usePathname } from 'next/navigation';
import Link from 'next/link';
import { FolderGit2, ListTodo, Target, GanttChartSquare, Settings } from 'lucide-react';
import { useAppDispatch } from '@/lib/store/hooks';
import {
  setActiveProject,
  setActiveProjectSlug,
} from '@/lib/store/slices/authSlice';
import { useGetProjectBySlugQuery } from '@/lib/store/api/projectsApi';
import { Badge } from '@/components/ui/badge';

export default function ProjectLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const params = useParams<{ workspaceSlug: string; projectSlug: string }>();
  const router = useRouter();
  const pathname = usePathname();
  const dispatch = useAppDispatch();

  const workspaceSlug = params?.workspaceSlug;
  const projectSlug = params?.projectSlug;

  const {
    data: project,
    isLoading,
    isError,
  } = useGetProjectBySlugQuery(projectSlug || '', {
    skip: !projectSlug,
  });

  useEffect(() => {
    if (project) {
      dispatch(setActiveProject(project.id));
      dispatch(setActiveProjectSlug(project.slug));
    }
  }, [project, dispatch]);

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b] min-h-[50vh]">
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
          <span className="text-xs text-zinc-400">Loading project context...</span>
        </div>
      </div>
    );
  }

  if (isError || (!isLoading && !project)) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b] p-6">
        <div className="max-w-md w-full text-center space-y-4 p-8 rounded-md border border-zinc-800 bg-zinc-900/60">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-rose-500/10 text-rose-400">
            <FolderGit2 className="h-6 w-6" />
          </div>
          <h2 className="text-lg font-bold text-zinc-100">Project Not Found</h2>
          <p className="text-xs text-zinc-400">
            The project <strong className="text-zinc-200">{projectSlug}</strong> does not
            exist in this workspace.
          </p>
          <button
            onClick={() => router.push(`/${workspaceSlug}/projects`)}
            className="inline-flex items-center justify-center px-4 py-2 rounded-md text-xs font-semibold bg-zinc-100 text-zinc-950 hover:bg-zinc-200 transition-colors"
          >
            View Workspace Projects
          </button>
        </div>
      </div>
    );
  }

  const tabs = [
    {
      name: 'Tasks',
      href: `/${workspaceSlug}/${projectSlug}/tasks`,
      icon: ListTodo,
    },
    {
      name: 'Sprints',
      href: `/${workspaceSlug}/${projectSlug}/sprints`,
      icon: Target,
    },
  ];

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#09090b]">
      {/* Project Sub-Header */}
      <div className="border-b border-zinc-800/80 px-6 py-3 bg-zinc-950/60 shrink-0">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="h-8 w-8 rounded-md bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-300">
              <FolderGit2 className="h-4 w-4" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-sm font-bold text-zinc-100">
                  {project.name}
                </h2>
                <Badge variant="outline" className="text-[10px] uppercase font-mono">
                  {project.slug}
                </Badge>
              </div>
              {project.description && (
                <p className="text-xs text-zinc-400 line-clamp-1">
                  {project.description}
                </p>
              )}
            </div>
          </div>

          {/* Project View Tabs */}
          <div className="flex items-center gap-1 bg-zinc-900/90 p-1 rounded-md border border-zinc-800/80">
            {tabs.map((tab) => {
              const isActive = pathname === tab.href;
              const Icon = tab.icon;

              return (
                <Link
                  key={tab.name}
                  href={tab.href}
                  className={`
                    flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium transition-colors
                    ${
                      isActive
                        ? 'bg-zinc-100 text-zinc-950 font-semibold shadow-xs'
                        : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/60'
                    }
                  `}
                >
                  <Icon className="h-3.5 w-3.5" />
                  <span>{tab.name}</span>
                </Link>
              );
            })}
          </div>
        </div>
      </div>

      {/* Project Content */}
      <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
        {children}
      </div>
    </div>
  );
}

