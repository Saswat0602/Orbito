'use client';

import React, { useMemo } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';
import { Zap, TrendingUp, BarChart2 } from 'lucide-react';
import type { SprintVelocity } from '@/types';
import { useGetProjectChartsQuery } from '@/lib/store/api/projectsApi';
import { Card } from '@/components/ui/card';

interface SprintVelocityChartProps {
  projectSlug: string;
}

export const SprintVelocityChart: React.FC<SprintVelocityChartProps> = ({
  projectSlug,
}) => {
  const { data: chartResponse, isLoading } = useGetProjectChartsQuery(
    { slug: projectSlug, types: ['sprint-velocity'] },
    { skip: !projectSlug }
  );

  const rawVelocity: SprintVelocity[] = useMemo(() => {
    return (chartResponse?.['sprint-velocity'] as SprintVelocity[]) || [];
  }, [chartResponse]);

  const { chartData, avgVelocity, maxVelocity, totalVelocity } = useMemo(() => {
    if (!rawVelocity || rawVelocity.length === 0) {
      return {
        chartData: [],
        avgVelocity: 0,
        maxVelocity: 0,
        totalVelocity: 0,
      };
    }

    const data = rawVelocity.map((v) => ({
      name: v.name,
      velocity: v.velocity || 0,
    }));

    const total = data.reduce((sum, item) => sum + item.velocity, 0);
    const avg = data.length > 0 ? Math.round(total / data.length) : 0;
    const max = Math.max(0, ...data.map((item) => item.velocity));

    return {
      chartData: data,
      avgVelocity: avg,
      maxVelocity: max,
      totalVelocity: total,
    };
  }, [rawVelocity]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-zinc-800 pb-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Zap className="h-4 w-4 text-amber-400" />
            <h2 className="text-sm font-semibold text-zinc-100">
              Historical Sprint Velocity
            </h2>
          </div>
          <p className="text-xs text-zinc-400">
            Completed story points per iteration and rolling team throughput average.
          </p>
        </div>
      </div>

      {/* KPI Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card className="border-zinc-800 bg-zinc-950/80 p-4 space-y-1">
          <span className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider block">
            Average Velocity
          </span>
          <span className="text-xl font-bold text-amber-400">
            {avgVelocity} <span className="text-xs font-normal text-zinc-400">pts / sprint</span>
          </span>
        </Card>

        <Card className="border-zinc-800 bg-zinc-950/80 p-4 space-y-1">
          <span className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider block">
            Peak Velocity
          </span>
          <span className="text-xl font-bold text-zinc-100">
            {maxVelocity} <span className="text-xs font-normal text-zinc-400">pts</span>
          </span>
        </Card>

        <Card className="border-zinc-800 bg-zinc-950/80 p-4 space-y-1">
          <span className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider block">
            Completed Sprints Evaluated
          </span>
          <span className="text-xl font-bold text-zinc-100">
            {rawVelocity.length}
          </span>
        </Card>
      </div>

      {/* Chart */}
      <Card className="border-zinc-800 bg-zinc-950/80 p-6 space-y-4">
        <div className="h-80 w-full">
          {isLoading ? (
            <div className="h-full flex items-center justify-center text-xs text-zinc-500">
              <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100 mr-2" />
              Loading velocity metrics...
            </div>
          ) : chartData.length === 0 ? (
            <div className="h-full flex items-center justify-center text-xs text-zinc-500 italic">
              No completed sprints recorded yet to calculate historical velocity.
            </div>
          ) : (
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData} margin={{ top: 10, right: 20, left: 0, bottom: 10 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                <XAxis
                  dataKey="name"
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
                <Legend wrapperStyle={{ fontSize: '11px', paddingTop: '10px' }} />
                <Bar
                  dataKey="velocity"
                  name="Completed Story Points"
                  fill="#f59e0b"
                  radius={[4, 4, 0, 0]}
                  maxBarSize={48}
                />
                {avgVelocity > 0 && (
                  <ReferenceLine
                    y={avgVelocity}
                    stroke="#38bdf8"
                    strokeDasharray="4 4"
                    label={{
                      value: `Avg: ${avgVelocity} pts`,
                      fill: '#38bdf8',
                      fontSize: 11,
                      position: 'top',
                    }}
                  />
                )}
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>
      </Card>
    </div>
  );
};