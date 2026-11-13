import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export function DashboardSkeleton() {
  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b] select-none">
      {/* 1. Header & Actions Skeleton */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-2">
          <div className="flex items-center gap-2.5">
            <Skeleton className="h-7 w-52 bg-zinc-800 rounded-md" />
            <Skeleton className="h-5 w-16 bg-zinc-850 rounded-full" />
          </div>
          <Skeleton className="h-3.5 w-80 bg-zinc-850" />
        </div>

        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-24 rounded-md bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-8 w-28 rounded-md bg-zinc-850" />
          <Skeleton className="h-8 w-32 rounded-md bg-zinc-800" />
        </div>
      </div>

      {/* 2. Top 4 KPI Metrics Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div
            key={i}
            className="rounded-lg border border-zinc-800 bg-zinc-900/60 p-4 space-y-3"
          >
            <div className="flex items-center justify-between">
              <Skeleton className="h-3 w-20 bg-zinc-800" />
              <Skeleton className="h-4 w-4 rounded bg-zinc-850" />
            </div>
            <div className="flex items-baseline gap-2">
              <Skeleton className="h-7 w-12 bg-zinc-800" />
              <Skeleton className="h-3 w-16 bg-zinc-850" />
            </div>
          </div>
        ))}
      </div>

      {/* 3. Analytical Portfolio & Velocity Blocks */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column: Projects Portfolio */}
        <div className="lg:col-span-2 rounded-lg border border-zinc-800 bg-zinc-900/40 p-5 space-y-4">
          <div className="flex items-center justify-between pb-2 border-b border-zinc-800/60">
            <div className="space-y-1">
              <Skeleton className="h-4 w-32 bg-zinc-800" />
              <Skeleton className="h-3 w-48 bg-zinc-850" />
            </div>
            <Skeleton className="h-7 w-20 rounded bg-zinc-850" />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 pt-1">
            {[1, 2, 3, 4].map((p) => (
              <div
                key={p}
                className="p-3.5 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-3"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Skeleton className="h-6 w-6 rounded bg-zinc-800" />
                    <Skeleton className="h-4 w-28 bg-zinc-800" />
                  </div>
                  <Skeleton className="h-4 w-12 rounded bg-zinc-850" />
                </div>
                <Skeleton className="h-1.5 w-full rounded-full bg-zinc-850" />
                <div className="flex items-center justify-between pt-1">
                  <Skeleton className="h-3 w-16 bg-zinc-850" />
                  <Skeleton className="h-3 w-12 bg-zinc-850" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Right Column: Activity Stream & Health */}
        <div className="rounded-lg border border-zinc-800 bg-zinc-900/40 p-5 space-y-4">
          <div className="space-y-1 pb-2 border-b border-zinc-800/60">
            <Skeleton className="h-4 w-28 bg-zinc-800" />
            <Skeleton className="h-3 w-36 bg-zinc-850" />
          </div>

          <div className="space-y-3 pt-1">
            {[1, 2, 3, 4, 5].map((a) => (
              <div key={a} className="flex items-start gap-3 p-2 rounded-md bg-zinc-950/30">
                <Skeleton className="h-6 w-6 rounded-full bg-zinc-800 shrink-0 mt-0.5" />
                <div className="space-y-1.5 flex-1 min-w-0">
                  <Skeleton className="h-3.5 w-full bg-zinc-800" />
                  <Skeleton className="h-2.5 w-20 bg-zinc-850" />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

