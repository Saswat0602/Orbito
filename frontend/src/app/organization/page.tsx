'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import Link from 'next/link';
import {
  Building2,
  FolderGit2,
  Users,
  CheckCircle2,
  TrendingUp,
  BarChart3,
  Layers,
  ArrowRight,
  Shield,
  Activity,
  AlertTriangle,
  RefreshCw,
  PieChart,
  Briefcase,
  ExternalLink,
} from 'lucide-react';
import { toast } from 'sonner';

import { useAppSelector } from '@/lib/store/hooks';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import {
  useGetOrganizationsQuery,
  useGetOrganizationQuery,
  useGetOrganizationMembersQuery,
  useGetOrganizationChartsQuery,
} from '@/lib/store/api/organizationsApi';
import { useGetWorkspacesQuery } from '@/lib/store/api/workspacesApi';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

const CHART_TYPES = [
  'kpi-metrics',
  'project-portfolio',
  'team-utilization',
  'task-distribution',
  'task-type',
  'sprint-metrics',
  'quality-metrics',
];

export default function OrganizationDashboardPage() {
  const authState = useAppSelector((state) => state.auth);
  const activeOrgId = authState.activeOrganizationId;
  const { user, isOwner, isManager, isSuperAdmin, roleBadge } = useAuthRbac();

  const { data: organizations = [] } = useGetOrganizationsQuery();
  const currentOrg = useMemo(() => {
    return organizations.find((o) => o.id === activeOrgId) || organizations[0] || null;
  }, [organizations, activeOrgId]);

  const orgId = currentOrg?.id || '';

  const {
    data: chartData = {},
    isLoading: isChartsLoading,
    refetch: refetchCharts,
  } = useGetOrganizationChartsQuery(
    { orgId, types: CHART_TYPES },
    { skip: !orgId }
  );

  const { data: workspaces = [], isLoading: isWorkspacesLoading } = useGetWorkspacesQuery(
    orgId ? { organizationId: orgId } : undefined,
    { skip: !orgId }
  );

  const { data: members = [] } = useGetOrganizationMembersQuery(orgId, { skip: !orgId });

  // Derived KPI data with graceful defaults
  const kpi = chartData['kpi-metrics'] || {};
  const totalWorkspaces = kpi.totalWorkspaces ?? workspaces.length;
  const activeWorkspaces = kpi.activeWorkspaces ?? workspaces.length;
  const totalProjects = kpi.totalProjects ?? 0;
  const completionRate = kpi.projectCompletionRate ?? 0;
  const totalMembers = kpi.totalMembers ?? (members.length || 1);
  const totalTasks = kpi.totalTasks ?? 0;

  // Project Portfolio
  const portfolio = Array.isArray(chartData['project-portfolio'])
    ? chartData['project-portfolio']
    : [];

  // Task distribution
  const taskDist = Array.isArray(chartData['task-distribution'])
    ? chartData['task-distribution']
    : [];

  // Task types
  const taskTypes = Array.isArray(chartData['task-type'])
    ? chartData['task-type']
    : [];

  const handleRefresh = async () => {
    await refetchCharts();
    toast.success('Organization analytics updated');
  };

  if (!currentOrg) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center bg-[#09090b] p-6 text-center space-y-3">
        <Building2 className="h-8 w-8 text-zinc-600" />
        <h2 className="text-sm font-semibold text-zinc-300">No Organization Selected</h2>
        <p className="text-xs text-zinc-500">
          Create or join an organization to access high-level executive analytics.
        </p>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#09090b] p-6 space-y-6 overflow-y-auto">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5">
            <Building2 className="h-5 w-5 text-zinc-200" />
            <h1 className="text-xl font-bold tracking-tight text-zinc-100">
              {currentOrg.name} &mdash; Executive Dashboard
            </h1>
            <Badge className={`text-[10px] uppercase font-mono ${roleBadge.className}`}>
              {roleBadge.label}
            </Badge>
          </div>
          <p className="text-xs text-zinc-400">
            Cross-workspace operational analytics, project portfolio health, and team capacity.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={isChartsLoading}
            className="gap-1.5 text-xs h-8 border-zinc-800 bg-zinc-900 text-zinc-300 hover:bg-zinc-800"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isChartsLoading ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </Button>

          <Link href="/workspaces">
            <Button
              variant="outline"
              size="sm"
              className="gap-1.5 text-xs h-8 border-zinc-800 bg-zinc-900 text-zinc-200 hover:bg-zinc-800"
            >
              <Briefcase className="h-3.5 w-3.5 text-zinc-400" />
              <span>All Workspaces</span>
            </Button>
          </Link>
        </div>
      </div>

      {/* Top 4 KPI Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Total Workspaces */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-4 space-y-1.5">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono text-zinc-400 uppercase tracking-wider">
              Workspaces
            </span>
            <Building2 className="h-4 w-4 text-purple-400" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-zinc-100">{totalWorkspaces}</span>
            <span className="text-xs text-zinc-500 font-mono">
              ({activeWorkspaces} active)
            </span>
          </div>
        </Card>

        {/* Total Projects */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-4 space-y-1.5">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono text-zinc-400 uppercase tracking-wider">
              Total Projects
            </span>
            <FolderGit2 className="h-4 w-4 text-sky-400" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-zinc-100">{totalProjects}</span>
            <span className="text-xs text-zinc-500 font-mono">across all scopes</span>
          </div>
        </Card>

        {/* Total Organization Members */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-4 space-y-1.5">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono text-zinc-400 uppercase tracking-wider">
              Team Capacity
            </span>
            <Users className="h-4 w-4 text-emerald-400" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-zinc-100">{totalMembers}</span>
            <span className="text-xs text-zinc-500 font-mono">collaborators</span>
          </div>
        </Card>

        {/* Project Completion Rate */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-4 space-y-1.5">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono text-zinc-400 uppercase tracking-wider">
              Throughput
            </span>
            <TrendingUp className="h-4 w-4 text-amber-400" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-zinc-100">
              {Math.round(completionRate)}%
            </span>
            <span className="text-xs text-zinc-500 font-mono">delivery rate</span>
          </div>
        </Card>
      </div>

      {/* Main Charts & Analytics Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Project Portfolio Distribution */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <h3 className="text-sm font-semibold text-zinc-100">Project Portfolio Status</h3>
              <p className="text-[11px] text-zinc-400">
                Current lifecycle state of initiatives across the organization.
              </p>
            </div>
            <PieChart className="h-4 w-4 text-zinc-500" />
          </div>

          <div className="space-y-3 pt-2">
            {portfolio.length === 0 ? (
              <div className="p-8 text-center text-xs text-zinc-500 italic">
                No project portfolio data available. Create projects to begin tracking status.
              </div>
            ) : (
              portfolio.map((item: any, idx: number) => {
                const statusName = item.status || `Group ${idx + 1}`;
                const count = item._count?.status ?? item.count ?? 1;
                const total = portfolio.reduce(
                  (sum: number, p: any) => sum + (p._count?.status ?? p.count ?? 1),
                  0
                );
                const percent = total > 0 ? Math.round((count / total) * 100) : 0;

                const colorClass =
                  statusName === 'ACTIVE'
                    ? 'bg-emerald-500'
                    : statusName === 'COMPLETED'
                    ? 'bg-sky-500'
                    : statusName === 'ARCHIVED'
                    ? 'bg-zinc-600'
                    : 'bg-amber-500';

                return (
                  <div key={idx} className="space-y-1">
                    <div className="flex items-center justify-between text-xs">
                      <span className="text-zinc-300 font-medium">{statusName}</span>
                      <span className="font-mono text-zinc-400 text-[11px]">
                        {count} ({percent}%)
                      </span>
                    </div>
                    <div className="h-2 w-full rounded-full bg-zinc-800 overflow-hidden">
                      <div className={`h-full ${colorClass}`} style={{ width: `${percent}%` }} />
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </Card>

        {/* Quality Metrics & Task Type Breakdown */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <h3 className="text-sm font-semibold text-zinc-100">Work Category Distribution</h3>
              <p className="text-[11px] text-zinc-400">
                Composition of active engineering work by task type.
              </p>
            </div>
            <BarChart3 className="h-4 w-4 text-zinc-500" />
          </div>

          <div className="space-y-3 pt-2">
            {taskTypes.length === 0 ? (
              <div className="p-8 text-center text-xs text-zinc-500 italic">
                No task type breakdown recorded yet. Tasks will appear here as they are logged.
              </div>
            ) : (
              taskTypes.map((item: any, idx: number) => {
                const typeName = item.taskType || item.type || `Type ${idx + 1}`;
                const count = item._count?.taskType ?? item.count ?? 1;
                const total = taskTypes.reduce(
                  (sum: number, t: any) => sum + (t._count?.taskType ?? t.count ?? 1),
                  0
                );
                const percent = total > 0 ? Math.round((count / total) * 100) : 0;

                const colorClass =
                  typeName === 'BUG'
                    ? 'bg-rose-500'
                    : typeName === 'STORY'
                    ? 'bg-purple-500'
                    : typeName === 'EPIC'
                    ? 'bg-amber-500'
                    : 'bg-sky-500';

                return (
                  <div key={idx} className="space-y-1">
                    <div className="flex items-center justify-between text-xs">
                      <span className="text-zinc-300 font-medium">{typeName}</span>
                      <span className="font-mono text-zinc-400 text-[11px]">
                        {count} tasks ({percent}%)
                      </span>
                    </div>
                    <div className="h-2 w-full rounded-full bg-zinc-800 overflow-hidden">
                      <div className={`h-full ${colorClass}`} style={{ width: `${percent}%` }} />
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </Card>
      </div>

      {/* Workspaces Portfolio Quick-Launch Grid */}
      <div className="space-y-3 pt-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Layers className="h-4 w-4 text-zinc-400" />
            <h2 className="text-sm font-semibold text-zinc-100">Active Workspaces in {currentOrg.name}</h2>
          </div>
          <Link href="/workspaces" className="text-xs text-zinc-400 hover:text-zinc-200 transition flex items-center gap-1">
            <span>Manage All Workspaces</span>
            <ArrowRight className="h-3 w-3" />
          </Link>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {workspaces.map((ws) => (
            <Link
              key={ws.id}
              href={`/${ws.slug}`}
              className="p-4 rounded-md border border-zinc-800 bg-zinc-900/60 hover:bg-zinc-900 transition flex flex-col justify-between group"
            >
              <div className="space-y-1">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-semibold text-zinc-200 group-hover:text-zinc-100">
                    {ws.name}
                  </span>
                  <ExternalLink className="h-3.5 w-3.5 text-zinc-600 group-hover:text-zinc-400 transition" />
                </div>
                {ws.description && (
                  <p className="text-xs text-zinc-400 line-clamp-1">{ws.description}</p>
                )}
              </div>

              <div className="flex items-center justify-between pt-3 mt-3 border-t border-zinc-850 text-[11px] text-zinc-500 font-mono">
                <span>/{ws.slug}</span>
                <span className="text-zinc-400">Open Dashboard &rarr;</span>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
}
