'use client';

import * as React from 'react';
import type { ProjectHealthStats } from '@orbito/types';

export interface ExtendedProjectHealthStats extends Partial<ProjectHealthStats> {
  totalTasks?: number;
  completedTasks?: number;
  overdueTasks?: number;
  upcomingTasks?: number;
  completionPredictor?: number;
  completionRate?: number;
  heatmapData?: number[];
}

interface ProjectHealthIndicatorProps {
  stats?: ExtendedProjectHealthStats;
  compact?: boolean;
}

export function ProjectHealthIndicator({
  stats,
  compact = false,
}: ProjectHealthIndicatorProps) {
  const totalTasks = stats?.totalTasks ?? 0;
  const completionRate =
    stats?.completionPredictor ??
    stats?.completionRate ??
    (totalTasks > 0 && stats?.completedTasks !== undefined
      ? Math.round((stats.completedTasks / totalTasks) * 100)
      : 0);

  // Default 14-day heatmap data if not explicitly provided
  const heatmapData = stats?.heatmapData || [
    0, 1, 2, 0, 3, 1, 0, 2, 4, 1, 0, 2, 1, 0,
  ];
  const maxTasksInDay = Math.max(...heatmapData, 1);

  if (!stats || totalTasks === 0) {
    return (
      <div className="flex items-center justify-between w-full pt-2.5 border-t border-zinc-800/80 text-[11px] text-zinc-500">
        <span>No tasks recorded</span>
        <span className="font-mono text-[10px] text-zinc-600">0%</span>
      </div>
    );
  }

  // Determine health color according to performance threshold
  const getHealthColorClass = (rate: number) => {
    if (rate >= 75) return 'text-emerald-400 stroke-emerald-400';
    if (rate >= 40) return 'text-amber-400 stroke-amber-400';
    return 'text-rose-400 stroke-rose-400';
  };

  const getHealthBgClass = (rate: number) => {
    if (rate >= 75) return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
    if (rate >= 40) return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
    return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
  };

  const healthColor = getHealthColorClass(completionRate);
  const healthBadge = getHealthBgClass(completionRate);

  return (
    <div className="flex items-center justify-between w-full pt-2.5 border-t border-zinc-800/80">
      {/* Circular Progress Gauge */}
      <div
        className="flex items-center gap-2 group/health cursor-help"
        title={`Health Score: ${completionRate}% (${stats.completedTasks || 0} of ${totalTasks} completed)`}
      >
        <div className="relative w-5 h-5 flex items-center justify-center">
          <svg viewBox="0 0 36 36" className="w-5 h-5 -rotate-90">
            {/* Background ring */}
            <path
              className="text-zinc-800"
              d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
              fill="none"
              stroke="currentColor"
              strokeWidth="4"
            />
            {/* Progress animated ring */}
            <path
              className={`transition-all duration-500 ${healthColor}`}
              strokeDasharray={`${Math.min(100, Math.max(0, completionRate))}, 100`}
              d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
              fill="none"
              stroke="currentColor"
              strokeWidth="4"
              strokeLinecap="round"
            />
          </svg>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="text-xs font-mono font-bold text-zinc-200">
            {completionRate}%
          </span>
          <span
            className={`text-[9px] font-mono px-1 py-0.2 rounded border ${healthBadge}`}
          >
            {completionRate >= 75
              ? 'On Track'
              : completionRate >= 40
              ? 'Needs Focus'
              : 'At Risk'}
          </span>
        </div>
      </div>

      {/* 14-Day Workload Heatmap */}
      {!compact && (
        <div
          className="flex items-center gap-[2.5px] cursor-help"
          title="Workload Activity (Upcoming 14-day task distribution)"
        >
          {heatmapData.map((count, i) => {
            const intensity = count / maxTasksInDay;
            let bgClass = 'bg-zinc-800/80';
            if (intensity > 0) {
              if (intensity <= 0.3) bgClass = 'bg-violet-950 border border-violet-800/40';
              else if (intensity <= 0.7) bgClass = 'bg-violet-700/80';
              else bgClass = 'bg-violet-500 shadow-xs shadow-violet-500/50';
            }

            return (
              <div
                key={i}
                className={`w-1.5 h-3.5 rounded-[1px] transition-colors ${bgClass}`}
                title={
                  count > 0
                    ? `Day +${i}: ${count} task${count > 1 ? 's' : ''}`
                    : `Day +${i}: No tasks scheduled`
                }
              />
            );
          })}
        </div>
      )}
    </div>
  );
}

