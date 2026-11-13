import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export function AdminSettingsSkeleton() {
  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b] select-none">
      {/* 1. Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-5 border-b border-zinc-800/80">
        <div className="space-y-1.5">
          <div className="flex items-center gap-2.5">
            <Skeleton className="h-7 w-7 rounded-md bg-zinc-900 border border-zinc-800" />
            <Skeleton className="h-6 w-60 bg-zinc-800 rounded-md" />
            <Skeleton className="h-4 w-20 bg-zinc-850 rounded-full" />
          </div>
          <Skeleton className="h-3.5 w-80 bg-zinc-850" />
        </div>
      </div>

      {/* 2. Flat Horizontal Tabs Bar Skeleton */}
      <div className="flex flex-wrap gap-1.5 p-1 rounded-lg bg-zinc-900/60 border border-zinc-800/80 max-w-4xl">
        <Skeleton className="h-8 w-32 rounded-md bg-zinc-800" />
        <Skeleton className="h-8 w-28 rounded-md bg-zinc-850" />
        <Skeleton className="h-8 w-32 rounded-md bg-zinc-850" />
        <Skeleton className="h-8 w-28 rounded-md bg-zinc-850" />
        <Skeleton className="h-8 w-32 rounded-md bg-zinc-850" />
      </div>

      {/* 3. Settings Panel Skeleton */}
      <div className="border border-zinc-800/80 rounded-lg bg-zinc-900/30 divide-y divide-zinc-800/80 max-w-4xl">
        {[1, 2, 3, 4, 5].map((idx) => (
          <div
            key={idx}
            className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4"
          >
            <div className="space-y-1.5 max-w-lg flex-1">
              <div className="flex items-center gap-2">
                <Skeleton className="h-4 w-40 bg-zinc-800" />
                <Skeleton className="h-4 w-16 rounded bg-zinc-850" />
              </div>
              <Skeleton className="h-3 w-72 bg-zinc-850" />
            </div>
            <div className="shrink-0">
              {idx <= 2 ? (
                <Skeleton className="h-6 w-11 rounded-full bg-zinc-800" />
              ) : (
                <Skeleton className="h-8 w-56 rounded-md bg-zinc-950 border border-zinc-800" />
              )}
            </div>
          </div>
        ))}
      </div>

      {/* 4. Save Button Placeholder */}
      <div className="flex justify-end max-w-4xl">
        <Skeleton className="h-8 w-36 rounded-md bg-zinc-800" />
      </div>
    </div>
  );
}

