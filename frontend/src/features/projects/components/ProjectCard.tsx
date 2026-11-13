'use client';

import * as React from 'react';
import Link from 'next/link';
import {
  FolderGit2,
  CheckSquare,
  Target,
  Calendar,
  Settings,
  ArrowRight,
  Flame,
  Shield,
  Layers,
} from 'lucide-react';
import { ProjectStatus, ProjectPriority, type Project } from '@orbito/types';
import { useGetProjectChartsQuery } from '@/lib/store/api/projectsApi';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { ProjectHealthIndicator } from './ProjectHealthIndicator';

interface ProjectCardProps {
  project: Project & {
    _count?: {
      tasks?: number;
      members?: number;
      sprints?: number;
    };
  };
  workspaceSlug: string;
  canManage?: boolean;
}

const STATUS_BADGE_CONFIG: Record<
  ProjectStatus,
  { label: string; className: string }
> = {
  [ProjectStatus.ACTIVE]: {
    label: 'Active',
    className: 'bg-emerald-950/40 text-emerald-300 border-emerald-800/60',
  },
  [ProjectStatus.PLANNING]: {
    label: 'Planning',
    className: 'bg-cyan-950/40 text-cyan-300 border-cyan-800/60',
  },
  [ProjectStatus.ON_HOLD]: {
    label: 'On Hold',
    className: 'bg-amber-950/40 text-amber-300 border-amber-800/60',
  },
  [ProjectStatus.COMPLETED]: {
    label: 'Completed',
    className: 'bg-purple-950/40 text-purple-300 border-purple-800/60',
  },
  [ProjectStatus.ARCHIVED]: {
    label: 'Archived',
    className: 'bg-zinc-800 text-zinc-400 border-zinc-700',
  },
};

const PRIORITY_BADGE_CONFIG: Record<
  ProjectPriority,
  { label: string; dotColor: string }
> = {
  [ProjectPriority.LOW]: { label: 'Low', dotColor: 'bg-blue-400' },
  [ProjectPriority.MEDIUM]: { label: 'Medium', dotColor: 'bg-amber-400' },
  [ProjectPriority.HIGH]: { label: 'High', dotColor: 'bg-orange-400' },
  [ProjectPriority.CRITICAL]: { label: 'Critical', dotColor: 'bg-rose-500' },
};

export function ProjectCard({
  project,
  workspaceSlug,
  canManage = false,
}: ProjectCardProps) {
  // Query KPI metrics for health indicator and accurate progress
  const { data: chartData } = useGetProjectChartsQuery(
    { slug: project.slug, types: ['kpi-metrics'] },
    { skip: !project.slug }
  );

  const kpi = chartData?.['kpi-metrics'];
  const totalTasks = kpi?.totalTasks ?? project._count?.tasks ?? 0;
  const completedTasks = kpi?.completedTasks ?? 0;
  const completionRate = kpi
    ? Math.round(kpi.completionRate)
    : totalTasks > 0
    ? Math.round((completedTasks / totalTasks) * 100)
    : 0;

  const statusConfig =
    STATUS_BADGE_CONFIG[project.status] || STATUS_BADGE_CONFIG.ACTIVE;
  const priorityConfig =
    PRIORITY_BADGE_CONFIG[project.priority] || PRIORITY_BADGE_CONFIG.MEDIUM;

  const projectUrl = `/${workspaceSlug}/${project.slug}/tasks`;
  const sprintsUrl = `/${workspaceSlug}/${project.slug}/sprints`;
  const settingsUrl = `/${workspaceSlug}/${project.slug}/settings`;

  const members = project.members || [];
  const displayMembers = members.slice(0, 4);
  const remainingMembersCount = Math.max(0, members.length - 4);

  return (
    <Card className="group relative border-zinc-800 bg-zinc-900/80 hover:border-zinc-700 transition-all p-5 flex flex-col justify-between h-full space-y-4 shadow-sm hover:shadow-md">
      <div className="space-y-3">
        {/* Card Header: Avatar, Name, Slug & Status */}
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-start gap-3 min-w-0">
            {/* Project Leading Icon */}
            <div className="w-10 h-10 rounded-md bg-zinc-950 border border-zinc-800 flex items-center justify-center text-zinc-200 font-bold shrink-0 text-sm group-hover:border-violet-500/50 group-hover:text-violet-300 transition-colors">
              {project.name.charAt(0).toUpperCase()}
            </div>

            <div className="min-w-0">
              <div className="flex items-center gap-2 flex-wrap">
                <Link
                  href={projectUrl}
                  className="text-sm font-bold text-zinc-100 group-hover:text-white truncate hover:underline"
                >
                  {project.name}
                </Link>
                <Badge
                  variant="outline"
                  className="text-[10px] font-mono uppercase tracking-wider px-1.5 py-0 bg-zinc-950 text-zinc-400 border-zinc-800"
                >
                  {project.slug}
                </Badge>
              </div>

              {project.description ? (
                <p className="text-xs text-zinc-400 mt-1 line-clamp-2 leading-relaxed">
                  {project.description}
                </p>
              ) : (
                <p className="text-xs text-zinc-600 mt-1 italic">
                  No description provided
                </p>
              )}
            </div>
          </div>

          {/* Status Badge */}
          <Badge
            variant="outline"
            className={`text-[10px] font-semibold px-2 py-0.5 shrink-0 ${statusConfig.className}`}
          >
            {statusConfig.label}
          </Badge>
        </div>

        {/* Priority & Sprints Pills */}
        <div className="flex items-center gap-2 pt-1">
          <div className="flex items-center gap-1.5 px-2 py-0.5 rounded bg-zinc-950 border border-zinc-850 text-[11px] text-zinc-300">
            <span
              className={`w-1.5 h-1.5 rounded-full ${priorityConfig.dotColor}`}
            />
            <span>{priorityConfig.label}</span>
          </div>

          {(kpi?.activeSprints || 0) > 0 && (
            <div className="flex items-center gap-1 px-2 py-0.5 rounded bg-zinc-950 border border-zinc-850 text-[11px] text-zinc-300">
              <Flame className="w-3 h-3 text-orange-400" />
              <span>
                {kpi?.activeSprints} Active Sprint
                {kpi?.activeSprints !== 1 ? 's' : ''}
              </span>
            </div>
          )}
        </div>

        {/* Task Progress Bar */}
        <div className="space-y-1.5 pt-1">
          <div className="flex items-center justify-between text-xs">
            <span className="text-zinc-400 flex items-center gap-1.5">
              <CheckSquare className="w-3 h-3 text-zinc-500" />
              <span>
                {completedTasks} of {totalTasks} tasks completed
              </span>
            </span>
            <span className="font-mono text-zinc-300 font-semibold">
              {completionRate}%
            </span>
          </div>
          <div className="h-1.5 w-full rounded-full bg-zinc-950 overflow-hidden border border-zinc-850">
            <div
              className={`h-full transition-all duration-500 ${
                completionRate === 100 ? 'bg-zinc-100' : 'bg-zinc-300'
              }`}
              style={{ width: `${Math.min(100, Math.max(0, completionRate))}%` }}
            />
          </div>
        </div>
      </div>

      {/* Card Footer: Members, Health & Actions */}
      <div className="space-y-3 pt-3 border-t border-zinc-850">
        <div className="flex items-center justify-between">
          {/* Members Avatar Stack */}
          <div className="flex items-center -space-x-1.5">
            {displayMembers.length > 0 ? (
              displayMembers.map((m) => {
                const initials =
                  (
                    (m.user?.firstName?.[0] || '') +
                    (m.user?.lastName?.[0] || '')
                  ).toUpperCase() || 'U';
                return (
                  <Avatar
                    key={m.id}
                    className="h-6 w-6 rounded-full border border-zinc-800 bg-zinc-900"
                    title={
                      m.user?.firstName
                        ? `${m.user.firstName} ${m.user.lastName || ''}`.trim()
                        : m.user?.email || 'Member'
                    }
                  >
                    <AvatarFallback className="text-[9px] bg-zinc-800 text-zinc-300 font-semibold">
                      {initials}
                    </AvatarFallback>
                  </Avatar>
                );
              })
            ) : (
              <span className="text-[11px] text-zinc-500 flex items-center gap-1">
                <Shield className="w-3 h-3 text-zinc-600" /> Workspace Team
              </span>
            )}
            {remainingMembersCount > 0 && (
              <span className="flex items-center justify-center h-6 w-6 rounded-full border border-zinc-800 bg-zinc-950 text-[10px] font-mono text-zinc-400">
                +{remainingMembersCount}
              </span>
            )}
          </div>

          {/* Quick Action Navigation Buttons */}
          <div className="flex items-center gap-1">
            <Link
              href={projectUrl}
              className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
              title="Open Tasks Board"
            >
              <CheckSquare className="w-3.5 h-3.5" />
            </Link>
            <Link
              href={sprintsUrl}
              className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
              title="Open Sprints"
            >
              <Target className="w-3.5 h-3.5" />
            </Link>
            {canManage && (
              <Link
                href={settingsUrl}
                className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
                title="Project Settings"
              >
                <Settings className="w-3.5 h-3.5" />
              </Link>
            )}
          </div>
        </div>

        {/* Circular SVG Gauge & 14-day Heatmap */}
        <ProjectHealthIndicator
          stats={{
            totalTasks,
            completedTasks,
            completionRate,
            completionPredictor: completionRate,
          }}
        />
      </div>
    </Card>
  );
}
