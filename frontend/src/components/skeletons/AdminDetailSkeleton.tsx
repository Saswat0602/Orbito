import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export function AdminDetailSkeleton() {
  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b] select-none">
      {/* 1. Header with Back Button */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-5 border-b border-zinc-800/80">
        <div className="flex items-center gap-3">
          <Skeleton className="h-8 w-8 rounded-md bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-10 w-10 rounded-md bg-zinc-800" />
          <div className="space-y-1.5">
            <div className="flex items-center gap-2">
              <Skeleton className="h-6 w-44 bg-zinc-800 rounded-md" />
              <Skeleton className="h-4 w-16 bg-zinc-850 rounded-full" />
            </div>
            <Skeleton className="h-3 w-56 bg-zinc-850" />
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-24 rounded-md bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-8 w-28 rounded-md bg-zinc-850" />
        </div>
      </div>

      {/* 2. Main Two-Column Details Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column (2 cols) */}
        <div className="lg:col-span-2 space-y-6">
          {/* Card 1: Main Details */}
          <div className="rounded-lg border border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
            <div className="space-y-1 pb-2 border-b border-zinc-800/60">
              <Skeleton className="h-4 w-32 bg-zinc-800" />
              <Skeleton className="h-3 w-48 bg-zinc-850" />
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-1">
              {[1, 2, 3, 4, 5, 6].map((field) => (
                <div key={field} className="space-y-1.5">
                  <Skeleton className="h-3 w-20 bg-zinc-850 uppercase text-[10px]" />
                  <Skeleton className="h-8 w-full rounded-md bg-zinc-950 border border-zinc-850" />
                </div>
              ))}
            </div>
          </div>

          {/* Card 2: Nested Items / Workspaces Table */}
          <div className="rounded-lg border border-zinc-800 bg-zinc-900/60 divide-y divide-zinc-850/80 overflow-hidden">
            <div className="p-4 bg-zinc-900/80 flex items-center justify-between">
              <Skeleton className="h-4 w-32 bg-zinc-800" />
              <Skeleton className="h-7 w-20 rounded bg-zinc-850" />
            </div>
            {[1, 2, 3].map((row) => (
              <div key={row} className="p-3.5 flex items-center justify-between bg-zinc-950/20">
                <div className="flex items-center gap-2.5">
                  <Skeleton className="h-7 w-7 rounded bg-zinc-800" />
                  <div className="space-y-1">
                    <Skeleton className="h-3.5 w-32 bg-zinc-800" />
                    <Skeleton className="h-2.5 w-20 bg-zinc-850" />
                  </div>
                </div>
                <Skeleton className="h-5 w-16 rounded bg-zinc-850" />
              </div>
            ))}
          </div>
        </div>

        {/* Right Column (1 col): Metadata & Danger Zone */}
        <div className="space-y-6">
          <div className="rounded-lg border border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
            <Skeleton className="h-4 w-28 bg-zinc-800 pb-2 border-b border-zinc-800/60" />
            <div className="space-y-3 pt-1">
              <div className="flex justify-between">
                <Skeleton className="h-3 w-16 bg-zinc-850" />
                <Skeleton className="h-3 w-24 bg-zinc-800" />
              </div>
              <div className="flex justify-between">
                <Skeleton className="h-3 w-20 bg-zinc-850" />
                <Skeleton className="h-3 w-28 bg-zinc-800" />
              </div>
              <div className="flex justify-between">
                <Skeleton className="h-3 w-20 bg-zinc-850" />
                <Skeleton className="h-3 w-20 bg-zinc-800" />
              </div>
            </div>
          </div>

          <div className="rounded-lg border border-rose-950/40 bg-zinc-900/40 p-5 space-y-3">
            <Skeleton className="h-4 w-24 bg-rose-950/60" />
            <Skeleton className="h-3 w-full bg-zinc-850" />
            <Skeleton className="h-8 w-full rounded-md bg-rose-950/40 border border-rose-900/30" />
          </div>
        </div>
      </div>
    </div>
  );
}

