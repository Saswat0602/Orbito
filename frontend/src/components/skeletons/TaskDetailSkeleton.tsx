import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export function TaskDetailSkeleton() {
  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#09090b] text-zinc-100 select-none overflow-hidden">
      {/* 1. Task Top Bar Skeleton */}
      <div className="h-12 px-6 border-b border-zinc-800 flex items-center justify-between shrink-0 bg-zinc-900/40">
        <div className="flex items-center gap-2">
          <Skeleton className="h-4 w-20 bg-zinc-800 font-mono" />
          <span className="text-zinc-600">/</span>
          <Skeleton className="h-4 w-24 bg-zinc-850" />
        </div>
        <div className="flex items-center gap-2">
          <Skeleton className="h-7 w-7 rounded bg-zinc-850" />
          <Skeleton className="h-7 w-7 rounded bg-zinc-850" />
          <Skeleton className="h-7 w-7 rounded bg-zinc-850" />
        </div>
      </div>

      {/* 2. Main Two-Column Layout */}
      <div className="flex-1 flex flex-col lg:flex-row min-h-0 overflow-y-auto">
        {/* Left Section: Title, Description, Subtasks, Comments */}
        <div className="flex-1 p-6 lg:p-8 space-y-6 lg:border-r border-zinc-800/80 overflow-y-auto">
          {/* Title Placeholder */}
          <div className="space-y-2">
            <Skeleton className="h-7 w-4/5 bg-zinc-800 rounded-md" />
            <Skeleton className="h-4 w-1/3 bg-zinc-850" />
          </div>

          {/* Description Section Placeholder */}
          <div className="space-y-3 pt-2">
            <div className="flex items-center justify-between pb-1">
              <Skeleton className="h-4 w-24 bg-zinc-800 uppercase text-[11px]" />
              <Skeleton className="h-6 w-20 rounded bg-zinc-850" />
            </div>
            <div className="p-4 rounded-md bg-zinc-900/40 border border-zinc-850 space-y-2.5">
              <Skeleton className="h-3.5 w-full bg-zinc-850/80" />
              <Skeleton className="h-3.5 w-11/12 bg-zinc-850/80" />
              <Skeleton className="h-3.5 w-4/5 bg-zinc-850/70" />
              <Skeleton className="h-3.5 w-2/3 bg-zinc-850/60" />
            </div>
          </div>

          {/* Subtasks Block Placeholder */}
          <div className="space-y-2.5 pt-2">
            <div className="flex items-center justify-between pb-1">
              <Skeleton className="h-4 w-20 bg-zinc-800 uppercase text-[11px]" />
              <Skeleton className="h-6 w-20 rounded bg-zinc-850" />
            </div>
            <div className="space-y-2">
              {[1, 2, 3].map((s) => (
                <div
                  key={s}
                  className="p-2.5 rounded-md bg-zinc-950/40 border border-zinc-850 flex items-center gap-3"
                >
                  <Skeleton className="h-4 w-4 rounded bg-zinc-800" />
                  <Skeleton className="h-3.5 w-2/3 bg-zinc-850" />
                </div>
              ))}
            </div>
          </div>

          {/* Comments Feed Placeholder */}
          <div className="space-y-4 pt-4 border-t border-zinc-800/80">
            <div className="flex items-center gap-2">
              <Skeleton className="h-6 w-24 rounded bg-zinc-800" />
              <Skeleton className="h-6 w-24 rounded bg-zinc-850" />
            </div>

            {/* Comment Box */}
            <div className="p-3 rounded-md bg-zinc-900/50 border border-zinc-800 space-y-2">
              <Skeleton className="h-12 w-full bg-zinc-950 rounded" />
              <div className="flex justify-between items-center pt-1">
                <div className="flex gap-1.5">
                  <Skeleton className="h-5 w-5 rounded bg-zinc-850" />
                  <Skeleton className="h-5 w-5 rounded bg-zinc-850" />
                </div>
                <Skeleton className="h-7 w-20 rounded bg-zinc-800" />
              </div>
            </div>

            {/* Past Comments */}
            {[1, 2].map((c) => (
              <div key={c} className="flex items-start gap-3 p-3 rounded-md bg-zinc-950/30">
                <Skeleton className="h-7 w-7 rounded-full bg-zinc-800 shrink-0" />
                <div className="space-y-1.5 flex-1">
                  <div className="flex items-center gap-2">
                    <Skeleton className="h-3.5 w-28 bg-zinc-800" />
                    <Skeleton className="h-3 w-16 bg-zinc-850/60" />
                  </div>
                  <Skeleton className="h-3 w-5/6 bg-zinc-850" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Right Section: Metadata Sidebar (35%) */}
        <div className="w-full lg:w-80 p-6 space-y-5 bg-zinc-900/20 shrink-0 overflow-y-auto">
          <Skeleton className="h-4 w-28 bg-zinc-800 uppercase text-[11px] pb-1 border-b border-zinc-800/60" />

          {/* Properties Grid */}
          <div className="space-y-4">
            {[
              { label: 'STATUS' },
              { label: 'PRIORITY' },
              { label: 'ASSIGNEES' },
              { label: 'SPRINT' },
              { label: 'DUE DATE' },
              { label: 'STORY POINTS' },
            ].map((prop, pIdx) => (
              <div key={pIdx} className="space-y-1.5">
                <Skeleton className="h-2.5 w-16 bg-zinc-850 uppercase text-[9px]" />
                <Skeleton className="h-8 w-full rounded-md bg-zinc-950 border border-zinc-850" />
              </div>
            ))}
          </div>

          {/* Time Tracking Widget Placeholder */}
          <div className="p-3.5 rounded-md bg-zinc-950 border border-zinc-850 space-y-2.5">
            <div className="flex items-center justify-between">
              <Skeleton className="h-3 w-24 bg-zinc-800" />
              <Skeleton className="h-6 w-14 rounded bg-zinc-850" />
            </div>
            <Skeleton className="h-2 w-full rounded-full bg-zinc-850" />
            <Skeleton className="h-3 w-32 bg-zinc-850" />
          </div>
        </div>
      </div>
    </div>
  );
}

