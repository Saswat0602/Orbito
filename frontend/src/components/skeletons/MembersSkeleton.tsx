import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export function MembersSkeleton() {
  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b] select-none">
      {/* 1. Header Skeleton */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-5 border-b border-zinc-800/80">
        <div className="space-y-1.5">
          <div className="flex items-center gap-2.5">
            <Skeleton className="h-6 w-52 bg-zinc-800 rounded-md" />
            <Skeleton className="h-4 w-16 bg-zinc-850 rounded-full" />
          </div>
          <Skeleton className="h-3.5 w-80 bg-zinc-850" />
        </div>

        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-32 rounded-md bg-zinc-800" />
        </div>
      </div>

      {/* 2. Tabs & Role Distribution Row */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 pb-1 border-b border-zinc-850">
        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-32 rounded-t-md bg-zinc-800" />
          <Skeleton className="h-8 w-36 rounded-t-md bg-zinc-900" />
        </div>
        <div className="hidden md:flex items-center gap-2">
          <Skeleton className="h-5 w-20 rounded bg-zinc-900 border border-zinc-850" />
          <Skeleton className="h-5 w-20 rounded bg-zinc-900 border border-zinc-850" />
          <Skeleton className="h-5 w-20 rounded bg-zinc-900 border border-zinc-850" />
        </div>
      </div>

      {/* 3. Search Bar */}
      <div className="flex items-center justify-between gap-3">
        <Skeleton className="h-8 w-64 rounded-md bg-zinc-900 border border-zinc-800" />
      </div>

      {/* 4. Members Directory List */}
      <div className="rounded-lg border border-zinc-800 bg-zinc-900/60 divide-y divide-zinc-850/80 overflow-hidden">
        {[1, 2, 3, 4, 5, 6].map((m) => (
          <div
            key={m}
            className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-zinc-950/20"
          >
            {/* User Info */}
            <div className="flex items-center gap-3">
              <Skeleton className="h-9 w-9 rounded-md bg-zinc-800 shrink-0" />
              <div className="space-y-1.5">
                <div className="flex items-center gap-2">
                  <Skeleton className="h-4 w-28 bg-zinc-800" />
                  {m === 1 && (
                    <Skeleton className="h-3.5 w-10 rounded bg-zinc-850" />
                  )}
                </div>
                <Skeleton className="h-3 w-40 bg-zinc-850" />
              </div>
            </div>

            {/* Role, Status & Actions */}
            <div className="flex items-center gap-3 shrink-0">
              <Skeleton className="h-5 w-20 rounded-md bg-zinc-850" />
              <Skeleton className="h-3 w-20 bg-zinc-850 hidden sm:block" />
              <Skeleton className="h-6 w-6 rounded bg-zinc-850" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

