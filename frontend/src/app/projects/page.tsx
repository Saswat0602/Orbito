'use client';

import * as React from 'react';
import Link from 'next/link';
import { FolderGit2, Plus, Users, Hash, Calendar } from 'lucide-react';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { setActiveProject } from '@/lib/store/slices/authSlice';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

export default function ProjectsDirectoryPage() {
  const dispatch = useAppDispatch();
  const { can, isManager, isOwner } = useAuthRbac();
  const activeWorkspaceId = useAppSelector((state) => state.auth.activeWorkspaceId);

  const { data: projects = [], isLoading } = useGetProjectsQuery(
    activeWorkspaceId ? { workspaceId: activeWorkspaceId } : undefined
  );

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <FolderGit2 className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Projects
            </h1>
          </div>
          <p className="text-xs text-zinc-400">
            Workspaces contain projects with independent backlogs, keys, and statuses.
          </p>
        </div>

        {(isManager || isOwner) && (
          <Button variant="primary" size="sm" className="gap-2">
            <Plus className="h-4 w-4" />
            <span>New Project</span>
          </Button>
        )}
      </div>

      {/* Projects Grid */}
      {isLoading ? (
        <div className="flex h-48 items-center justify-center">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
        </div>
      ) : projects.length === 0 ? (
        <Card className="border-zinc-800 bg-zinc-900/60 p-12 text-center">
          <FolderGit2 className="h-10 w-10 text-zinc-500 mx-auto mb-3" />
          <h3 className="text-base font-semibold text-zinc-200">No projects yet</h3>
          <p className="text-xs text-zinc-400 mt-1 max-w-sm mx-auto">
            Create your first project to start tracking issues, sprint milestones, and tasks.
          </p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {projects.map((project) => (
            <Link
              key={project.id}
              href="/tasks"
              onClick={() => dispatch(setActiveProject(project.id))}
              className="group block"
            >
              <Card className="border-zinc-800 bg-zinc-900/80 group-hover:border-zinc-700 transition-all shadow-md h-full">
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <Badge variant="outline" className="text-[10px] uppercase font-mono">
                      {project.slug || 'PRJ'}
                    </Badge>
                  </div>
                  <CardTitle className="text-base mt-2 group-hover:text-white transition-colors">
                    {project.name}
                  </CardTitle>
                  {project.description && (
                    <CardDescription className="text-xs line-clamp-2">
                      {project.description}
                    </CardDescription>
                  )}
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="flex items-center gap-4 text-xs text-zinc-500 border-t border-zinc-850 pt-3">
                    <div className="flex items-center gap-1.5">
                      <Calendar className="h-3.5 w-3.5" />
                      <span>{new Date(project.createdAt).toLocaleDateString()}</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

