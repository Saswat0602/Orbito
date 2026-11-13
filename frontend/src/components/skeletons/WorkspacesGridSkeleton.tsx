import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export function WorkspacesGridSkeleton() {
  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b] select-none">
      {/* 1. Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-5 border-b border-zinc-800/80">
        <div className="space-y-1.5">
          <div className="flex items-center gap-2.5">
            <Skeleton className="h-6 w-44 bg-zinc-800 rounded-md" />
            <Skeleton className="h-4 w-16 bg-zinc-850 rounded-full" />
          </div>
          <Skeleton className="h-3.5 w-72 bg-zinc-850" />
        </div>

        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-36 rounded-md bg-zinc-800" />
        </div>
      </div>

      {/* 2. Search & Filter Bar */}
      <div className="flex items-center justify-between gap-3">
        <Skeleton className="h-8 w-64 rounded-md bg-zinc-900 border border-zinc-800" />
        <Skeleton className="h-3.5 w-24 bg-zinc-850 hidden sm:block" />
      </div>

      {/* 3. Grid of Workspaces */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {[1, 2, 3, 4, 5, 6].map((idx) => (
          <div
            key={idx}
            className="rounded-lg border border-zinc-800 bg-zinc-900/60 p-5 space-y-4 hover:border-zinc-700 transition-colors"
          >
            {/* Header */}
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-3">
                <Skeleton className="h-9 w-9 rounded-md bg-zinc-800" />
                <div className="space-y-1">
                  <Skeleton className="h-4 w-32 bg-zinc-800" />
                  <Skeleton className="h-3 w-20 bg-zinc-850 font-mono" />
                </div>
              </div>
              <Skeleton className="h-4 w-12 rounded bg-zinc-850" />
            </div>

            {/* Description lines */}
            <div className="space-y-1.5">
              <Skeleton className="h-3 w-full bg-zinc-850/80" />
              <Skeleton className="h-3 w-3/4 bg-zinc-850/60" />
            </div>

            {/* Footer */}
            <div className="flex items-center justify-between pt-3 border-t border-zinc-850/80">
              <div className="flex -space-x-1.5">
                <Skeleton className="h-5 w-5 rounded-full bg-zinc-800 border border-zinc-900" />
                <Skeleton className="h-5 w-5 rounded-full bg-zinc-800 border border-zinc-900" />
                <Skeleton className="h-5 w-5 rounded-full bg-zinc-850 border border-zinc-900" />
              </div>
              <Skeleton className="h-3 w-16 bg-zinc-850" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

