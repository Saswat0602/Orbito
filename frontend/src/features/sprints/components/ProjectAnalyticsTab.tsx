'use client';

import React, { useMemo } from 'react';
import {
  PieChart,
  Pie,
  Cell,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import {
  Activity,
  CheckCircle2,
  Bug,
  Target,
  Layers,
  TrendingUp,
} from 'lucide-react';
import type {
  ProjectKPIMetrics,
  TaskStatusFlow,
  TaskTypeCount,
  TaskPriorityCount,
} from '@/types';
import { useGetProjectChartsQuery } from '@/lib/store/api/projectsApi';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

interface ProjectAnalyticsTabProps {
  projectSlug: string;
}

const STATUS_COLORS: Record<string, string> = {
  TODO: '#71717a',
  IN_PROGRESS: '#38bdf8',
  IN_REVIEW: '#f59e0b',
  DONE: '#10b981',
};

const PRIORITY_COLORS: Record<string, string> = {
  URGENT: '#f43f5e',
  HIGH: '#f59e0b',
  MEDIUM: '#38bdf8',
  LOW: '#71717a',
};

const TYPE_COLORS: Record<string, string> = {
  TASK: '#38bdf8',
  BUG: '#f43f5e',
  STORY: '#10b981',
  EPIC: '#a855f7',
};

export const ProjectAnalyticsTab: React.FC<ProjectAnalyticsTabProps> = ({
  projectSlug,
}) => {
  const { data: chartResponse, isLoading } = useGetProjectChartsQuery(
    {
      slug: projectSlug,
      types: ['kpi-metrics', 'task-status', 'task-priority', 'task-type'],
    },
    { skip: !projectSlug }
  );

  const kpis: ProjectKPIMetrics | undefined = chartResponse?.['kpi-metrics'];
  const statusFlow: TaskStatusFlow[] = (chartResponse?.['task-status'] as TaskStatusFlow[]) || [];
  const priorityDistribution: TaskPriorityCount[] =
    (chartResponse?.['task-priority'] as TaskPriorityCount[]) || [];
  const typeDistribution: TaskTypeCount[] =
    (chartResponse?.['task-type'] as TaskTypeCount[]) || [];

  // Formatted data for Status Donut
  const statusChartData = useMemo(() => {
    return statusFlow.map((item) => ({
      name: item.status.name,
      value: item.count,
      color: item.status.color || STATUS_COLORS[item.status.category] || '#71717a',
    }));
  }, [statusFlow]);

  // Formatted data for Priority Bar
  const priorityChartData = useMemo(() => {
    return priorityDistribution.map((item) => ({
      name: item.priority,
      count: item._count.priority,
      color: PRIORITY_COLORS[item.priority] || '#71717a',
    }));
  }, [priorityDistribution]);

  // Formatted data for Type Donut
  const typeChartData = useMemo(() => {
    return typeDistribution.map((item) => ({
      name: item.type,
      value: item._count.type,
      color: TYPE_COLORS[item.type] || '#71717a',
    }));
  }, [typeDistribution]);

  if (isLoading) {
    return (
      <div className="py-16 flex flex-col items-center justify-center gap-3 text-zinc-500">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
        <span className="text-xs">Loading project KPI metrics and charts...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 4 KPI Metric Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="border-zinc-800 bg-zinc-950/80 p-5 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-semibold text-zinc-400 uppercase tracking-wider">
              Completion Rate
            </span>
            <CheckCircle2 className="h-4 w-4 text-emerald-400" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-zinc-100">
              {kpis?.completionRate ?? 0}%
            </span>
            <span className="text-xs text-zinc-400">
              ({kpis?.completedTasks ?? 0}/{kpis?.totalTasks ?? 0} tasks)
            </span>
          </div>
          {/* Progress bar */}
          <div className="w-full bg-zinc-900 rounded-full h-1 overflow-hidden mt-1">
            <div
              className="bg-emerald-500 h-full rounded-full"
              style={{ width: `${kpis?.completionRate ?? 0}%` }}
            />
          </div>
        </Card>

        <Card className="border-zinc-800 bg-zinc-950/80 p-5 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-semibold text-zinc-400 uppercase tracking-wider">
              Active Sprints
            </span>
            <Target className="h-4 w-4 text-sky-400" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-sky-400">
              {kpis?.activeSprints ?? 0}
            </span>
            <span className="text-xs text-zinc-400">Active iteration(s)</span>
          </div>
          <p className="text-[11px] text-zinc-500">
            Sprint cycle currently running live
          </p>
        </Card>

        <Card className="border-zinc-800 bg-zinc-950/80 p-5 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-semibold text-zinc-400 uppercase tracking-wider">
              Bug Resolution Rate
            </span>
            <Bug className="h-4 w-4 text-rose-400" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-zinc-100">
              {kpis?.bugResolutionRate ?? 0}%
            </span>
            <span className="text-xs text-zinc-400">
              ({kpis?.resolvedBugs ?? 0}/{kpis?.totalBugs ?? 0} resolved)
            </span>
          </div>
          {/* Progress bar */}
          <div className="w-full bg-zinc-900 rounded-full h-1 overflow-hidden mt-1">
            <div
              className="bg-rose-500 h-full rounded-full"
              style={{ width: `${kpis?.bugResolutionRate ?? 0}%` }}
            />
          </div>
        </Card>

        <Card className="border-zinc-800 bg-zinc-950/80 p-5 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-semibold text-zinc-400 uppercase tracking-wider">
              Work Item Volume
            </span>
            <Layers className="h-4 w-4 text-zinc-400" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-bold text-zinc-100">
              {kpis?.totalTasks ?? 0}
            </span>
            <span className="text-xs text-zinc-400">Total Project Tasks</span>
          </div>
          <p className="text-[11px] text-zinc-500">
            Across backlog and all iterations
          </p>
        </Card>
      </div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Status Distribution Donut */}
        <Card className="border-zinc-800 bg-zinc-950/80 p-5 space-y-3">
          <div className="flex items-center justify-between border-b border-zinc-850 pb-2">
            <h4 className="text-xs font-semibold text-zinc-200">
              Task Status Distribution
            </h4>
            <span className="text-[10px] text-zinc-500 font-mono">
              {statusChartData.reduce((s, i) => s + i.value, 0)} tasks
            </span>
          </div>

          <div className="h-56 w-full">
            {statusChartData.length === 0 ? (
              <div className="h-full flex items-center justify-center text-xs text-zinc-500 italic">
                No status data available
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={statusChartData}
                    cx="50%"
                    cy="50%"
                    innerRadius={50}
                    outerRadius={75}
                    paddingAngle={3}
                    dataKey="value"
                  >
                    {statusChartData.map((entry, index) => (
                      <Cell key={`status-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#09090b',
                      borderColor: '#27272a',
                      borderRadius: '6px',
                      fontSize: '11px',
                    }}
                  />
                  <Legend
                    wrapperStyle={{ fontSize: '10px', paddingTop: '8px' }}
                  />
                </PieChart>
              </ResponsiveContainer>
            )}
          </div>
        </Card>

        {/* Priority Breakdown Bar Chart */}
        <Card className="border-zinc-800 bg-zinc-950/80 p-5 space-y-3">
          <div className="flex items-center justify-between border-b border-zinc-850 pb-2">
            <h4 className="text-xs font-semibold text-zinc-200">
              Priority Breakdown
            </h4>
            <span className="text-[10px] text-zinc-500 font-mono">Severity</span>
          </div>

          <div className="h-56 w-full">
            {priorityChartData.length === 0 ? (
              <div className="h-full flex items-center justify-center text-xs text-zinc-500 italic">
                No priority data available
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={priorityChartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                  <XAxis dataKey="name" stroke="#71717a" fontSize={10} tickLine={false} />
                  <YAxis stroke="#71717a" fontSize={10} tickLine={false} />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#09090b',
                      borderColor: '#27272a',
                      borderRadius: '6px',
                      fontSize: '11px',
                    }}
                  />
                  <Bar dataKey="count" name="Tasks" radius={[4, 4, 0, 0]}>
                    {priorityChartData.map((entry, index) => (
                      <Cell key={`prio-${index}`} fill={entry.color} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>
        </Card>

        {/* Task Type Distribution Donut */}
        <Card className="border-zinc-800 bg-zinc-950/80 p-5 space-y-3">
          <div className="flex items-center justify-between border-b border-zinc-850 pb-2">
            <h4 className="text-xs font-semibold text-zinc-200">
              Work Item Types
            </h4>
            <span className="text-[10px] text-zinc-500 font-mono">Composition</span>
          </div>

          <div className="h-56 w-full">
            {typeChartData.length === 0 ? (
              <div className="h-full flex items-center justify-center text-xs text-zinc-500 italic">
                No type data available
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={typeChartData}
                    cx="50%"
                    cy="50%"
                    innerRadius={50}
                    outerRadius={75}
                    paddingAngle={3}
                    dataKey="value"
                  >
                    {typeChartData.map((entry, index) => (
                      <Cell key={`type-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#09090b',
                      borderColor: '#27272a',
                      borderRadius: '6px',
                      fontSize: '11px',
                    }}
                  />
                  <Legend
                    wrapperStyle={{ fontSize: '10px', paddingTop: '8px' }}
                  />
                </PieChart>
              </ResponsiveContainer>
            )}
          </div>
        </Card>
      </div>
    </div>
  );
};