'use client';

import React, { useState, useMemo } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import {
  Calendar,
  CheckCircle2,
  AlertCircle,
  TrendingDown,
  Target,
} from 'lucide-react';
import type { Sprint } from '@orbito/types';
import type { SprintBurndownTask } from '@/types';
import { useGetSprintBurndownQuery } from '@/lib/store/api/sprintsApi';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface SprintBurndownChartProps {
  projectSlug: string;
  sprints: Sprint[];
  initialSprintId?: string;
}

export const SprintBurndownChart: React.FC<SprintBurndownChartProps> = ({
  projectSlug,
  sprints,
  initialSprintId,
}) => {
  // Find active sprint as default, or initial, or first
  const defaultSprintId = useMemo(() => {
    if (initialSprintId) return initialSprintId;
    const active = sprints.find((s) => s.status === 'ACTIVE');
    return active?.id || sprints[0]?.id || '';
  }, [sprints, initialSprintId]);

  const [selectedSprintId, setSelectedSprintId] = useState<string>(defaultSprintId);

  const selectedSprint = useMemo(() => {
    return sprints.find((s) => s.id === selectedSprintId) || sprints[0];
  }, [sprints, selectedSprintId]);

  // Fetch burndown data for selected sprint
  const { data: rawData = [], isLoading } = useGetSprintBurndownQuery(
    { projectSlug, sprintId: selectedSprint?.id || '' },
    { skip: !projectSlug || !selectedSprint?.id }
  );

  // Compute daily burndown points
  const { chartData, totalPoints, completedPoints, remainingPoints, isAhead } =
    useMemo(() => {
      const tasks = (rawData as SprintBurndownTask[]) || [];
      const total = tasks.reduce(
        (sum, t) => sum + (t.storyPoints && t.storyPoints > 0 ? t.storyPoints : 1),
        0
      );

      // Determine sprint duration window
      const start = selectedSprint?.startDate
        ? new Date(selectedSprint.startDate)
        : selectedSprint?.createdAt
        ? new Date(selectedSprint.createdAt)
        : new Date();
      start.setHours(0, 0, 0, 0);

      const end = selectedSprint?.endDate
        ? new Date(selectedSprint.endDate)
        : new Date(start.getTime() + 14 * 86400000);
      end.setHours(23, 59, 59, 999);

      const today = new Date();
      today.setHours(23, 59, 59, 999);

      const totalDays = Math.max(
        1,
        Math.round((end.getTime() - start.getTime()) / 86400000)
      );

      let currentRemaining = total;
      const completedCount = tasks.filter((t) => t.completedAt !== null).length;
      const completedPts = tasks
        .filter((t) => t.completedAt !== null)
        .reduce(
          (sum, t) => sum + (t.storyPoints && t.storyPoints > 0 ? t.storyPoints : 1),
          0
        );

      const points = [];

      for (let i = 0; i <= totalDays; i++) {
        const currentDate = new Date(start.getTime() + i * 86400000);
        const dateKey = currentDate.toLocaleDateString('en-US', {
          month: 'short',
          day: 'numeric',
        });

        // Ideal burndown line
        const ideal = Math.max(0, Math.round(total - (total / totalDays) * i));

        // Actual burndown up to today
        let actual: number | null = null;
        if (currentDate <= today) {
          const completedByThisDate = tasks
            .filter((t) => t.completedAt && new Date(t.completedAt) <= currentDate)
            .reduce(
              (sum, t) =>
                sum + (t.storyPoints && t.storyPoints > 0 ? t.storyPoints : 1),
              0
            );
          actual = Math.max(0, total - completedByThisDate);
          currentRemaining = actual;
        }

        points.push({
          date: dateKey,
          ideal,
          actual,
        });
      }

      const lastDayIdeal = points.find((p) => p.actual !== null)?.ideal ?? total;
      const ahead = currentRemaining <= lastDayIdeal;

      return {
        chartData: points,
        totalPoints: total,
        completedPoints: completedPts,
        remainingPoints: Math.max(0, total - completedPts),
        isAhead: ahead,
      };
    }, [rawData, selectedSprint]);

  if (sprints.length === 0) {
    return (
      <Card className="border-zinc-800 bg-zinc-950/40 p-12 text-center text-xs text-zinc-500 italic">
        No sprints available to generate burndown chart.
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Top Header & Sprint Selector */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-zinc-800 pb-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <TrendingDown className="h-4 w-4 text-sky-400" />
            <h2 className="text-sm font-semibold text-zinc-100">
              Sprint Burndown Trajectory
            </h2>
          </div>
          <p className="text-xs text-zinc-400">
            Compare remaining work against the ideal linear sprint completion curve.
          </p>
        </div>

        {/* Sprint Picker */}
        <div className="flex items-center gap-2">
          <span className="text-xs text-zinc-400">Sprint:</span>
          <Select
            value={selectedSprint?.id || (sprints[0]?.id ?? 'none')}
            onValueChange={(val) => setSelectedSprintId(val)}
          >
            <SelectTrigger className="w-52 h-8 bg-zinc-900 border-zinc-700 text-zinc-200 text-xs">
              <SelectValue placeholder="Choose a sprint" />
            </SelectTrigger>
            <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
              {sprints.map((s) => (
                <SelectItem key={s.id} value={s.id} className="text-xs">
                  {s.name} ({s.status})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Metric Cards Banner */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <Card className="border-zinc-800 bg-zinc-950/80 p-4 space-y-1">
          <span className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider block">
            Committed Points
          </span>
          <span className="text-xl font-bold text-zinc-100">{totalPoints}</span>
        </Card>

        <Card className="border-zinc-800 bg-zinc-950/80 p-4 space-y-1">
          <span className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider block">
            Completed Points
          </span>
          <span className="text-xl font-bold text-emerald-400">
            {completedPoints}
          </span>
        </Card>

        <Card className="border-zinc-800 bg-zinc-950/80 p-4 space-y-1">
          <span className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider block">
            Remaining Points
          </span>
          <span className="text-xl font-bold text-sky-400">
            {remainingPoints}
          </span>
        </Card>

        <Card className="border-zinc-800 bg-zinc-950/80 p-4 space-y-1">
          <span className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider block">
            Schedule Pace
          </span>
          <div className="pt-1">
            {isAhead ? (
              <Badge variant="outline" className="text-xs border-emerald-800 text-emerald-400 bg-emerald-950/40">
                On Track / Ahead
              </Badge>
            ) : (
              <Badge variant="outline" className="text-xs border-amber-800 text-amber-400 bg-amber-950/40">
                Behind Schedule
              </Badge>
            )}
          </div>
        </Card>
      </div>

      {/* Main Burndown Line Chart */}
      <Card className="border-zinc-800 bg-zinc-950/80 p-6 space-y-4">
        <div className="h-80 w-full">
          {isLoading ? (
            <div className="h-full flex flex-col justify-between py-4 space-y-3">
              <div className="flex justify-between items-center">
                <Skeleton className="h-3 w-28" />
                <Skeleton className="h-3 w-20" />
              </div>
              <Skeleton className="h-52 w-full rounded-md" />
              <div className="flex justify-between items-center">
                <Skeleton className="h-2.5 w-16" />
                <Skeleton className="h-2.5 w-16" />
                <Skeleton className="h-2.5 w-16" />
                <Skeleton className="h-2.5 w-16" />
              </div>
            </div>
          ) : chartData.length === 0 ? (
            <div className="h-full flex items-center justify-center text-xs text-zinc-500 italic">
              No task burndown data available for this sprint.
            </div>
          ) : (
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData} margin={{ top: 10, right: 20, left: 0, bottom: 10 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                <XAxis
                  dataKey="date"
                  stroke="#71717a"
                  fontSize={11}
                  tickLine={false}
                />
                <YAxis
                  stroke="#71717a"
                  fontSize={11}
                  tickLine={false}
                  domain={[0, 'auto']}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#09090b',
                    borderColor: '#27272a',
                    borderRadius: '6px',
                    fontSize: '11px',
                    color: '#f4f4f5',
                  }}
                />
                <Legend
                  wrapperStyle={{ fontSize: '11px', paddingTop: '10px' }}
                />
                <Line
                  type="monotone"
                  dataKey="ideal"
                  name="Ideal Burndown"
                  stroke="#71717a"
                  strokeDasharray="5 5"
                  strokeWidth={2}
                  dot={false}
                />
                <Line
                  type="monotone"
                  dataKey="actual"
                  name="Actual Remaining"
                  stroke="#38bdf8"
                  strokeWidth={2.5}
                  connectNulls={false}
                  dot={{ r: 3, fill: '#38bdf8' }}
                  activeDot={{ r: 5 }}
                />
              </LineChart>
            </ResponsiveContainer>
          )}
        </div>
      </Card>
    </div>
  );
};