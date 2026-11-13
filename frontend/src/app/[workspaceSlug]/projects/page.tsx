'use client';

import * as React from 'react';
import { useState, useMemo, useEffect } from 'react';
import { useParams } from 'next/navigation';
import {
  FolderGit2,
  Plus,
  Search,
  X,
  Filter,
  Layers,
  AlertCircle,
  SlidersHorizontal,
  RotateCcw,
} from 'lucide-react';
import { ProjectStatus, ProjectPriority, type Project } from '@orbito/types';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppDispatch } from '@/lib/store/hooks';
import {
  setActiveProject,
  setActiveProjectSlug,
} from '@/lib/store/slices/authSlice';
import { useGetWorkspaceBySlugQuery } from '@/lib/store/api/workspacesApi';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import {
  CreateProjectModal,
  ProjectCard,
} from '@/features/projects';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';

const STATUS_FILTER_OPTIONS: Array<{
  value: 'ALL' | ProjectStatus;
  label: string;
}> = [
  { value: 'ALL', label: 'All Statuses' },
  { value: ProjectStatus.ACTIVE, label: 'Active' },
  { value: ProjectStatus.PLANNING, label: 'Planning' },
  { value: ProjectStatus.ON_HOLD, label: 'On Hold' },
  { value: ProjectStatus.COMPLETED, label: 'Completed' },
  { value: ProjectStatus.ARCHIVED, label: 'Archived' },
];

const PRIORITY_FILTER_OPTIONS: Array<{
  value: 'ALL' | ProjectPriority;
  label: string;
  dotColor?: string;
}> = [
  { value: 'ALL', label: 'All Priorities' },
  { value: ProjectPriority.LOW, label: 'Low', dotColor: 'bg-blue-400' },
  { value: ProjectPriority.MEDIUM, label: 'Medium', dotColor: 'bg-amber-400' },
  { value: ProjectPriority.HIGH, label: 'High', dotColor: 'bg-orange-400' },
  { value: ProjectPriority.CRITICAL, label: 'Critical', dotColor: 'bg-rose-500' },
];

export default function WorkspaceProjectsPage() {
  const params = useParams<{ workspaceSlug: string }>();
  const dispatch = useAppDispatch();
  const workspaceSlug = params?.workspaceSlug || '';
  const { isManager, isOwner, isSuperAdmin } = useAuthRbac();

  const canManage = isManager || isOwner || isSuperAdmin;

  useEffect(() => {
    dispatch(setActiveProject(null));
    dispatch(setActiveProjectSlug(null));
  }, [dispatch]);

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<'ALL' | ProjectStatus>('ALL');
  const [priorityFilter, setPriorityFilter] = useState<'ALL' | ProjectPriority>('ALL');

  // Debounce search query
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery.trim().toLowerCase());
    }, 250);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const { data: workspace } = useGetWorkspaceBySlugQuery(workspaceSlug, {
    skip: !workspaceSlug,
  });

  const { data: projects = [], isLoading } = useGetProjectsQuery(
    workspace?.id ? { workspaceId: workspace.id } : undefined,
    { skip: !workspace?.id }
  );

  // Status Counts
  const statusCounts = useMemo(() => {
    const counts: Record<string, number> = { ALL: projects.length };
    projects.forEach((p) => {
      const s = p.status || 'ACTIVE';
      counts[s] = (counts[s] || 0) + 1;
    });
    return counts;
  }, [projects]);

  // Priority Counts
  const priorityCounts = useMemo(() => {
    const counts: Record<string, number> = { ALL: projects.length };
    projects.forEach((p) => {
      const pr = p.priority || 'MEDIUM';
      counts[pr] = (counts[pr] || 0) + 1;
    });
    return counts;
  }, [projects]);

  // Filtered Projects
  const filteredProjects = useMemo(() => {
    return projects.filter((project) => {
      // Status Filter
      if (statusFilter !== 'ALL' && project.status !== statusFilter) {
        return false;
      }
      // Priority Filter
      if (priorityFilter !== 'ALL' && project.priority !== priorityFilter) {
        return false;
      }
      // Search Filter
      if (debouncedSearch) {
        const nameMatch = project.name.toLowerCase().includes(debouncedSearch);
        const slugMatch = project.slug.toLowerCase().includes(debouncedSearch);
        const descMatch = (project.description || '')
          .toLowerCase()
          .includes(debouncedSearch);
        if (!nameMatch && !slugMatch && !descMatch) {
          return false;
        }
      }
      return true;
    });
  }, [projects, statusFilter, priorityFilter, debouncedSearch]);

  const hasActiveFilters =
    searchQuery.trim() !== '' ||
    statusFilter !== 'ALL' ||
    priorityFilter !== 'ALL';

  const handleClearFilters = () => {
    setSearchQuery('');
    setDebouncedSearch('');
    setStatusFilter('ALL');
    setPriorityFilter('ALL');
  };

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b]">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5">
            <div className="h-8 w-8 rounded-md bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-100">
              <FolderGit2 className="h-4.5 w-4.5" />
            </div>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-zinc-100">
              Workspace Projects
            </h1>
            <Badge
              variant="outline"
              className="text-xs font-mono font-bold bg-zinc-950 text-zinc-300 border-zinc-800"
            >
              {projects.length} {projects.length === 1 ? 'Project' : 'Projects'}
            </Badge>
          </div>
          <p className="text-xs text-zinc-400">
            Portfolio projects organized under{' '}
            <strong className="text-zinc-200">
              {workspace?.name || workspaceSlug}
            </strong>
            .
          </p>
        </div>

        {canManage && (
          <Button
            variant="primary"
            size="sm"
            onClick={() => setIsCreateOpen(true)}
            className="gap-2 text-xs shrink-0"
          >
            <Plus className="h-4 w-4" />
            <span>New Project</span>
          </Button>
        )}
      </div>

      {/* Filter Toolbar: Search, Status Pills & Priority Pills */}
      <div className="space-y-3 bg-zinc-950/60 p-4 rounded-md border border-zinc-850">
        {/* Search & Meta Summary */}
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
          {/* Search Box */}
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search projects by name, key, or description..."
              className="pl-8.5 pr-8 py-1.5 text-xs bg-zinc-900 border-zinc-800 text-zinc-200 placeholder:text-zinc-500 focus:border-zinc-700"
            />
            {searchQuery && (
              <button
                type="button"
                onClick={() => setSearchQuery('')}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 transition-colors p-0.5"
                title="Clear search"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            )}
          </div>

          {/* Active Filter Clear Trigger */}
          {hasActiveFilters && (
            <div className="flex items-center gap-2">
              <span className="text-xs text-zinc-400">
                Showing{' '}
                <strong className="text-zinc-200">
                  {filteredProjects.length}
                </strong>{' '}
                of {projects.length}
              </span>
              <button
                type="button"
                onClick={handleClearFilters}
                className="flex items-center gap-1 text-xs text-zinc-300 hover:text-white transition-colors cursor-pointer px-2 py-1 rounded-md bg-zinc-900 border border-zinc-800 hover:border-zinc-750"
              >
                <RotateCcw className="w-3 h-3" />
                <span>Reset Filters</span>
              </button>
            </div>
          )}
        </div>

        {/* Status Filter Pills */}
        <div className="flex items-center gap-1.5 flex-wrap pt-1">
          <span className="text-[11px] font-medium text-zinc-500 flex items-center gap-1 mr-1">
            <Layers className="w-3 h-3" />
            <span>Status:</span>
          </span>
          {STATUS_FILTER_OPTIONS.map((opt) => {
            const isSelected = statusFilter === opt.value;
            const count = statusCounts[opt.value] || 0;

            return (
              <button
                key={opt.value}
                type="button"
                onClick={() => setStatusFilter(opt.value)}
                className={`flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors cursor-pointer ${
                  isSelected
                    ? 'bg-zinc-100 text-zinc-950 font-bold shadow-xs'
                    : 'bg-zinc-900 text-zinc-400 border border-zinc-800 hover:text-zinc-200 hover:border-zinc-700'
                }`}
              >
                <span>{opt.label}</span>
                <span
                  className={`text-[10px] font-mono px-1 rounded-full ${
                    isSelected
                      ? 'bg-zinc-950 text-zinc-200'
                      : 'bg-zinc-800 text-zinc-400'
                  }`}
                >
                  {count}
                </span>
              </button>
            );
          })}
        </div>

        {/* Priority Filter Pills */}
        <div className="flex items-center gap-1.5 flex-wrap pt-1 border-t border-zinc-850/60">
          <span className="text-[11px] font-medium text-zinc-500 flex items-center gap-1 mr-1">
            <AlertCircle className="w-3 h-3" />
            <span>Priority:</span>
          </span>
          {PRIORITY_FILTER_OPTIONS.map((opt) => {
            const isSelected = priorityFilter === opt.value;
            const count = priorityCounts[opt.value] || 0;

            return (
              <button
                key={opt.value}
                type="button"
                onClick={() => setPriorityFilter(opt.value)}
                className={`flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors cursor-pointer ${
                  isSelected
                    ? 'bg-zinc-100 text-zinc-950 font-bold shadow-xs'
                    : 'bg-zinc-900 text-zinc-400 border border-zinc-800 hover:text-zinc-200 hover:border-zinc-700'
                }`}
              >
                {opt.dotColor && (
                  <span className={`w-1.5 h-1.5 rounded-full ${opt.dotColor}`} />
                )}
                <span>{opt.label}</span>
                <span
                  className={`text-[10px] font-mono px-1 rounded-full ${
                    isSelected
                      ? 'bg-zinc-950 text-zinc-200'
                      : 'bg-zinc-800 text-zinc-400'
                  }`}
                >
                  {count}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Projects Grid Content */}
      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3, 4, 5, 6].map((idx) => (
            <div
              key={idx}
              className="rounded-lg border border-zinc-800 bg-zinc-900/60 p-5 space-y-4"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <Skeleton className="h-9 w-9 rounded-md bg-zinc-800" />
                  <div className="space-y-1">
                    <Skeleton className="h-4 w-32 bg-zinc-800" />
                    <Skeleton className="h-3 w-16 bg-zinc-850 font-mono" />
                  </div>
                </div>
                <Skeleton className="h-4 w-14 rounded bg-zinc-850" />
              </div>
              <div className="space-y-1.5">
                <Skeleton className="h-3 w-full bg-zinc-850/80" />
                <Skeleton className="h-3 w-3/4 bg-zinc-850/60" />
              </div>
              <div className="flex items-center justify-between pt-3 border-t border-zinc-850/80">
                <Skeleton className="h-3 w-20 bg-zinc-850" />
                <Skeleton className="h-5 w-5 rounded-full bg-zinc-800" />
              </div>
            </div>
          ))}
        </div>
      ) : projects.length === 0 ? (
        /* Workspace Empty State */
        <div className="border border-dashed border-zinc-800 rounded-md bg-zinc-950/40 p-12 text-center max-w-lg mx-auto space-y-4">
          <div className="w-12 h-12 rounded-full bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-500 mx-auto">
            <FolderGit2 className="h-6 w-6" />
          </div>
          <div>
            <h3 className="text-base font-semibold text-zinc-200">
              No projects created yet
            </h3>
            <p className="text-xs text-zinc-400 mt-1 leading-relaxed">
              Create your first project to start tracking backlogs, agile sprints,
              milestones, and interactive task boards.
            </p>
          </div>
          {canManage && (
            <Button
              variant="primary"
              size="sm"
              onClick={() => setIsCreateOpen(true)}
              className="gap-2 text-xs"
            >
              <Plus className="h-4 w-4" />
              <span>Create Project</span>
            </Button>
          )}
        </div>
      ) : filteredProjects.length === 0 ? (
        /* Filter Empty State */
        <div className="border border-zinc-850 rounded-md bg-zinc-950/40 p-12 text-center max-w-md mx-auto space-y-4">
          <div className="w-10 h-10 rounded-full bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-500 mx-auto">
            <Filter className="h-5 w-5" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-zinc-200">
              No matching projects
            </h3>
            <p className="text-xs text-zinc-400 mt-1">
              No projects match your current search and filter combination.
            </p>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={handleClearFilters}
            className="text-xs gap-1.5"
          >
            <RotateCcw className="w-3.5 h-3.5" />
            <span>Reset Filters</span>
          </Button>
        </div>
      ) : (
        /* Project Cards Grid */
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredProjects.map((project) => (
            <ProjectCard
              key={project.id}
              project={project as Project & { _count?: { tasks?: number; members?: number; sprints?: number } }}
              workspaceSlug={workspaceSlug}
              canManage={canManage}
            />
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
