import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export function AdminOverviewSkeleton() {
  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b] select-none">
      {/* 1. Admin Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1.5">
          <div className="flex items-center gap-2">
            <Skeleton className="h-7 w-56 bg-zinc-800 rounded-md" />
            <Skeleton className="h-4 w-20 bg-zinc-850 rounded-full" />
          </div>
          <Skeleton className="h-3.5 w-80 bg-zinc-850" />
        </div>

        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-32 rounded-md bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-8 w-32 rounded-md bg-zinc-900 border border-zinc-800" />
        </div>
      </div>

      {/* 2. Top Metric Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div
            key={i}
            className="rounded-lg border border-zinc-800 bg-zinc-900/60 p-5 space-y-3"
          >
            <div className="flex items-center justify-between">
              <Skeleton className="h-3 w-28 bg-zinc-800 uppercase" />
              <Skeleton className="h-4 w-4 rounded bg-zinc-850" />
            </div>
            <Skeleton className="h-8 w-16 bg-zinc-800" />
            <Skeleton className="h-2.5 w-32 bg-zinc-850" />
          </div>
        ))}
      </div>

      {/* 3. Telemetry & Tenant Sections */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="rounded-lg border border-zinc-800 bg-zinc-900/40 p-5 space-y-4">
          <div className="space-y-1 pb-2 border-b border-zinc-800/60">
            <Skeleton className="h-4 w-36 bg-zinc-800" />
            <Skeleton className="h-3 w-52 bg-zinc-850" />
          </div>
          <div className="space-y-2.5 pt-1">
            {[1, 2, 3, 4].map((q) => (
              <div
                key={q}
                className="p-3 rounded-md bg-zinc-950/40 border border-zinc-850 flex items-center justify-between"
              >
                <div className="space-y-1">
                  <Skeleton className="h-3.5 w-32 bg-zinc-800" />
                  <Skeleton className="h-2.5 w-20 bg-zinc-850" />
                </div>
                <Skeleton className="h-5 w-16 rounded bg-zinc-850" />
              </div>
            ))}
          </div>
        </div>

        <div className="rounded-lg border border-zinc-800 bg-zinc-900/40 p-5 space-y-4">
          <div className="space-y-1 pb-2 border-b border-zinc-800/60">
            <Skeleton className="h-4 w-40 bg-zinc-800" />
            <Skeleton className="h-3 w-48 bg-zinc-850" />
          </div>
          <div className="space-y-2.5 pt-1">
            {[1, 2, 3, 4].map((o) => (
              <div
                key={o}
                className="p-3 rounded-md bg-zinc-950/40 border border-zinc-850 flex items-center justify-between"
              >
                <div className="flex items-center gap-2.5">
                  <Skeleton className="h-7 w-7 rounded bg-zinc-800" />
                  <div className="space-y-1">
                    <Skeleton className="h-3.5 w-28 bg-zinc-800" />
                    <Skeleton className="h-2.5 w-16 bg-zinc-850" />
                  </div>
                </div>
                <Skeleton className="h-6 w-16 rounded bg-zinc-850" />
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

