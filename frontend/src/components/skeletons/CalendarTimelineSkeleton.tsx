import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

interface CalendarTimelineSkeletonProps {
  view?: 'calendar' | 'timeline';
}

export function CalendarTimelineSkeleton({ view = 'calendar' }: CalendarTimelineSkeletonProps) {
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#09090b] p-6 space-y-4 overflow-hidden select-none">
      {/* 1. Header & Navigation Controls */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 pb-1 border-b border-zinc-850">
        <div className="flex items-center gap-3">
          <Skeleton className="h-6 w-36 bg-zinc-800 rounded-md" />
          <div className="flex items-center gap-1">
            <Skeleton className="h-7 w-7 rounded bg-zinc-900 border border-zinc-800" />
            <Skeleton className="h-7 w-14 rounded bg-zinc-900 border border-zinc-800" />
            <Skeleton className="h-7 w-7 rounded bg-zinc-900 border border-zinc-800" />
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-28 rounded-md bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-8 w-28 rounded-md bg-zinc-800" />
        </div>
      </div>

      {/* 2. Main View Grid Skeleton */}
      {view === 'calendar' ? (
        /* Month Grid */
        <div className="flex-1 rounded-lg border border-zinc-800 bg-zinc-900/40 flex flex-col overflow-hidden">
          {/* Day Names Header */}
          <div className="grid grid-cols-7 border-b border-zinc-800 bg-zinc-900/80 text-center py-2">
            {['SUN', 'MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT'].map((d) => (
              <div key={d} className="flex justify-center">
                <Skeleton className="h-3 w-8 bg-zinc-800" />
              </div>
            ))}
          </div>

          {/* 5-Week Grid Cells */}
          <div className="flex-1 grid grid-cols-7 grid-rows-5 divide-x divide-y divide-zinc-850/60 overflow-hidden">
            {Array.from({ length: 35 }).map((_, idx) => (
              <div key={idx} className="p-2 flex flex-col justify-between min-h-[70px] bg-zinc-950/20">
                <div className="flex justify-between items-center">
                  <Skeleton className="h-3 w-4 rounded bg-zinc-850 font-mono text-[10px]" />
                  {idx % 4 === 0 && <Skeleton className="h-2 w-2 rounded-full bg-zinc-800" />}
                </div>
                {idx % 3 === 0 && (
                  <Skeleton className="h-4 w-full rounded bg-zinc-800/80 my-0.5" />
                )}
                {idx % 5 === 0 && (
                  <Skeleton className="h-4 w-4/5 rounded bg-zinc-850 my-0.5" />
                )}
              </div>
            ))}
          </div>
        </div>
      ) : (
        /* Timeline / Gantt Grid */
        <div className="flex-1 rounded-lg border border-zinc-800 bg-zinc-900/40 flex flex-col overflow-hidden">
          {/* Timeline Days Header */}
          <div className="h-10 px-4 bg-zinc-900/80 border-b border-zinc-800 flex items-center justify-between">
            <Skeleton className="h-3.5 w-36 bg-zinc-800" />
            <div className="flex items-center gap-4 flex-1 justify-end">
              <Skeleton className="h-3 w-12 bg-zinc-850" />
              <Skeleton className="h-3 w-12 bg-zinc-850" />
              <Skeleton className="h-3 w-12 bg-zinc-850" />
              <Skeleton className="h-3 w-12 bg-zinc-850" />
            </div>
          </div>

          {/* Timeline Task Rows with Gantt Bars */}
          <div className="flex-1 divide-y divide-zinc-850/60 overflow-hidden">
            {[1, 2, 3, 4, 5, 6, 7].map((row) => (
              <div
                key={row}
                className="h-12 px-4 flex items-center justify-between gap-4 bg-zinc-950/20"
              >
                <div className="flex items-center gap-2.5 w-48 shrink-0">
                  <Skeleton className="h-3.5 w-14 bg-zinc-850 font-mono" />
                  <Skeleton className="h-3.5 w-28 bg-zinc-800" />
                </div>
                <div className="flex-1 relative h-6 flex items-center">
                  <Skeleton
                    className={`h-5 rounded-md bg-zinc-800/80 ${
                      row === 1
                        ? 'w-1/2 ml-8'
                        : row === 2
                        ? 'w-1/3 ml-32'
                        : row === 3
                        ? 'w-2/5 ml-16'
                        : 'w-1/4 ml-48'
                    }`}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

