import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export function NotificationsSkeleton() {
  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b] select-none">
      {/* 1. Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-5 border-b border-zinc-800/80">
        <div className="space-y-1.5">
          <div className="flex items-center gap-2.5">
            <Skeleton className="h-6 w-48 bg-zinc-800 rounded-md" />
            <Skeleton className="h-4 w-12 bg-zinc-850 rounded-full" />
          </div>
          <Skeleton className="h-3.5 w-72 bg-zinc-850" />
        </div>

        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-28 rounded-md bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-8 w-32 rounded-md bg-zinc-800" />
        </div>
      </div>

      {/* 2. Category Tabs */}
      <div className="flex items-center gap-2 pb-1 border-b border-zinc-850">
        <Skeleton className="h-7 w-14 rounded-md bg-zinc-800" />
        <Skeleton className="h-7 w-16 rounded-md bg-zinc-850" />
        <Skeleton className="h-7 w-16 rounded-md bg-zinc-850" />
        <Skeleton className="h-7 w-20 rounded-md bg-zinc-850 hidden sm:block" />
      </div>

      {/* 3. Notifications List */}
      <div className="rounded-lg border border-zinc-800 bg-zinc-900/60 divide-y divide-zinc-850/80 overflow-hidden">
        {[1, 2, 3, 4, 5, 6].map((idx) => (
          <div
            key={idx}
            className="p-4 flex items-start justify-between gap-3 bg-zinc-950/20"
          >
            <div className="flex items-start gap-3 flex-1 min-w-0">
              <Skeleton className="h-8 w-8 rounded-full bg-zinc-800 shrink-0 mt-0.5" />
              <div className="space-y-1.5 flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <Skeleton className="h-3.5 w-36 bg-zinc-800" />
                  <Skeleton className="h-3 w-16 rounded bg-zinc-850" />
                </div>
                <Skeleton
                  className={`h-3 bg-zinc-850 ${
                    idx % 2 === 0 ? 'w-4/5' : 'w-3/5'
                  }`}
                />
                <Skeleton className="h-2.5 w-20 bg-zinc-850/60" />
              </div>
            </div>

            <div className="flex items-center gap-2 shrink-0">
              {idx <= 2 && <div className="h-2 w-2 rounded-full bg-zinc-700" />}
              <Skeleton className="h-6 w-6 rounded bg-zinc-850" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

