'use client';

import * as React from 'react';
import { useState } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { FolderGit2, Plus, Calendar } from 'lucide-react';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppDispatch } from '@/lib/store/hooks';
import { setActiveProject, setActiveProjectSlug } from '@/lib/store/slices/authSlice';
import { useGetWorkspaceBySlugQuery } from '@/lib/store/api/workspacesApi';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import { CreateProjectModal } from '@/features/projects';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

export default function WorkspaceProjectsPage() {
  const params = useParams<{ workspaceSlug: string }>();
  const dispatch = useAppDispatch();
  const { workspaceSlug } = params;
  const { isManager, isOwner } = useAuthRbac();

  React.useEffect(() => {
    dispatch(setActiveProject(null));
    dispatch(setActiveProjectSlug(null));
  }, [dispatch]);

  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const { data: workspace } = useGetWorkspaceBySlugQuery(workspaceSlug || '', {
    skip: !workspaceSlug,
  });

  const { data: projects = [], isLoading } = useGetProjectsQuery(
    workspace?.id ? { workspaceId: workspace.id } : undefined,
    { skip: !workspace?.id }
  );

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <FolderGit2 className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Workspace Projects
            </h1>
          </div>
          <p className="text-xs text-zinc-400">
            Projects organized under <strong className="text-zinc-200">{workspace?.name || workspaceSlug}</strong>.
          </p>
        </div>

        {(isManager || isOwner) && (
          <Button
            variant="primary"
            size="sm"
            onClick={() => setIsCreateOpen(true)}
            className="gap-2 text-xs"
          >
            <Plus className="h-4 w-4" />
            <span>New Project</span>
          </Button>
        )}
      </div>

      {isLoading ? (
        <div className="flex h-48 items-center justify-center">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
        </div>
      ) : projects.length === 0 ? (
        <Card className="border-zinc-800 bg-zinc-900/60 p-12 text-center">
          <FolderGit2 className="h-10 w-10 text-zinc-500 mx-auto mb-3" />
          <h3 className="text-base font-semibold text-zinc-200">No projects found</h3>
          <p className="text-xs text-zinc-400 mt-1 max-w-sm mx-auto">
            Create a project to start tracking backlogs, milestones, and task boards.
          </p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
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
                <div className="flex items-center gap-4 text-xs text-zinc-500 border-t border-zinc-850 pt-3">
                  <div className="flex items-center gap-1.5">
                    <Calendar className="h-3.5 w-3.5" />
                    <span>{new Date(project.createdAt).toLocaleDateString()}</span>
                  </div>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}

      {/* Create Project Modal */}
      <CreateProjectModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        workspaceId={workspace?.id || ''}
      />
    </div>
  );
}
