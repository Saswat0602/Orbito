import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export function SprintsSkeleton() {
  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b] select-none">
      {/* 1. Sprints Header Skeleton */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-5 border-b border-zinc-800/80">
        <div className="space-y-1.5">
          <div className="flex items-center gap-2.5">
            <Skeleton className="h-6 w-44 bg-zinc-800 rounded-md" />
            <Skeleton className="h-4 w-16 bg-zinc-850 rounded-full" />
          </div>
          <Skeleton className="h-3.5 w-72 bg-zinc-850" />
        </div>

        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-28 rounded-md bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-8 w-32 rounded-md bg-zinc-800" />
        </div>
      </div>

      {/* 2. Sprint Velocity & Burndown Summary Bar */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {[
          { label: 'ACTIVE SPRINT' },
          { label: 'PLANNED POINTS' },
          { label: 'BURNDOWN VELOCITY' },
        ].map((m, idx) => (
          <div
            key={idx}
            className="rounded-lg border border-zinc-800 bg-zinc-900/60 p-4 space-y-2.5"
          >
            <div className="flex items-center justify-between">
              <Skeleton className="h-3 w-24 bg-zinc-800" />
              <Skeleton className="h-4 w-4 rounded bg-zinc-850" />
            </div>
            <div className="flex items-baseline gap-2">
              <Skeleton className="h-6 w-16 bg-zinc-800" />
              <Skeleton className="h-3 w-20 bg-zinc-850" />
            </div>
          </div>
        ))}
      </div>

      {/* 3. Active Sprint Panel */}
      <div className="rounded-lg border border-zinc-800 bg-zinc-900/40 divide-y divide-zinc-800/80 overflow-hidden">
        {/* Sprint Header */}
        <div className="p-4 bg-zinc-900/60 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <div className="flex items-center gap-3">
            <Skeleton className="h-4 w-4 rounded bg-zinc-800" />
            <Skeleton className="h-4 w-36 bg-zinc-800" />
            <Skeleton className="h-4 w-14 rounded-full bg-emerald-950/60 border border-emerald-800/40" />
            <Skeleton className="h-3 w-28 bg-zinc-850 hidden sm:block" />
          </div>

          <div className="flex items-center gap-2">
            <Skeleton className="h-7 w-20 rounded bg-zinc-850" />
            <Skeleton className="h-7 w-28 rounded bg-zinc-800" />
          </div>
        </div>

        {/* Sprint Tasks List */}
        <div className="divide-y divide-zinc-850/60 p-1">
          {[1, 2, 3, 4].map((t) => (
            <div
              key={t}
              className="p-3 flex items-center justify-between gap-3 bg-zinc-950/20"
            >
              <div className="flex items-center gap-3 flex-1 min-w-0">
                <Skeleton className="h-4 w-4 rounded bg-zinc-850" />
                <Skeleton className="h-3.5 w-16 bg-zinc-850 font-mono" />
                <Skeleton
                  className={`h-3.5 bg-zinc-800 ${
                    t % 2 === 0 ? 'w-2/3' : 'w-1/2'
                  }`}
                />
              </div>

              <div className="flex items-center gap-3 shrink-0">
                <Skeleton className="h-5 w-16 rounded bg-zinc-850 hidden sm:block" />
                <Skeleton className="h-4 w-12 rounded bg-zinc-850" />
                <Skeleton className="h-5 w-5 rounded-full bg-zinc-850" />
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 4. Backlog Accordion Card */}
      <div className="rounded-lg border border-zinc-800 bg-zinc-900/30 overflow-hidden">
        <div className="p-4 flex items-center justify-between bg-zinc-900/60">
          <div className="flex items-center gap-3">
            <Skeleton className="h-4 w-4 rounded bg-zinc-800" />
            <Skeleton className="h-4 w-24 bg-zinc-800" />
            <Skeleton className="h-4 w-8 rounded-full bg-zinc-850" />
          </div>
          <Skeleton className="h-7 w-24 rounded bg-zinc-850" />
        </div>

        <div className="divide-y divide-zinc-850/60 p-1">
          {[1, 2, 3].map((b) => (
            <div
              key={b}
              className="p-3 flex items-center justify-between gap-3 bg-zinc-950/20"
            >
              <div className="flex items-center gap-3 flex-1 min-w-0">
                <Skeleton className="h-4 w-4 rounded bg-zinc-850" />
                <Skeleton className="h-3.5 w-16 bg-zinc-850 font-mono" />
                <Skeleton className="h-3.5 w-3/5 bg-zinc-800" />
              </div>
              <div className="flex items-center gap-2 shrink-0">
                <Skeleton className="h-4 w-14 rounded bg-zinc-850" />
                <Skeleton className="h-5 w-5 rounded-full bg-zinc-850" />
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

