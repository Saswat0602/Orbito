import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

interface TasksViewSkeletonProps {
  view?: 'kanban' | 'list' | 'board' | 'table';
}

export function TasksViewSkeleton({ view = 'kanban' }: TasksViewSkeletonProps) {
  const isKanban = view === 'kanban' || view === 'board';

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#09090b] p-6 space-y-4 overflow-hidden select-none">
      {/* 1. Control Action Bar Skeleton */}
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-3 shrink-0 pb-1">
        <div className="flex flex-wrap items-center gap-2 flex-1">
          {/* Search Bar Placeholder */}
          <Skeleton className="h-8 w-56 rounded-md bg-zinc-900 border border-zinc-800" />
          {/* Filter Pills */}
          <Skeleton className="h-8 w-24 rounded-md bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-8 w-24 rounded-md bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-8 w-28 rounded-md bg-zinc-900 border border-zinc-800 hidden sm:block" />
        </div>

        <div className="flex items-center gap-2 shrink-0">
          {/* View Mode Switcher Skeleton */}
          <div className="flex items-center rounded-md bg-zinc-900 p-0.5 border border-zinc-800 gap-1">
            <Skeleton className="h-7 w-14 rounded-sm bg-zinc-800" />
            <Skeleton className="h-7 w-14 rounded-sm bg-zinc-850" />
            <Skeleton className="h-7 w-14 rounded-sm bg-zinc-850 hidden sm:block" />
          </div>

          {/* New Task Button Skeleton */}
          <Skeleton className="h-8 w-28 rounded-md bg-zinc-800" />
        </div>
      </div>

      {/* 2. Main Content Canvas */}
      {isKanban ? (
        /* Kanban Board Columns */
        <div className="flex-1 flex gap-4 overflow-hidden pt-1">
          {[
            { label: 'BACKLOG', color: 'bg-zinc-800' },
            { label: 'TO DO', color: 'bg-zinc-800' },
            { label: 'IN PROGRESS', color: 'bg-zinc-800' },
            { label: 'DONE', color: 'bg-zinc-800' },
          ].map((col, cIdx) => (
            <div
              key={cIdx}
              className="flex-1 min-w-[270px] max-w-xs flex flex-col rounded-lg bg-zinc-900/30 border border-zinc-850/80 p-3 overflow-hidden"
            >
              {/* Column Header */}
              <div className="flex items-center justify-between pb-3 mb-2 border-b border-zinc-850/70">
                <div className="flex items-center gap-2">
                  <div className={`h-2.5 w-2.5 rounded-full ${col.color}`} />
                  <Skeleton className="h-4 w-20 bg-zinc-800" />
                  <Skeleton className="h-4 w-6 rounded-full bg-zinc-850" />
                </div>
                <Skeleton className="h-5 w-5 rounded bg-zinc-850" />
              </div>

              {/* Task Cards Stack */}
              <div className="flex-1 space-y-3 overflow-y-hidden pt-1">
                {[1, 2, 3].map((cardIdx) => (
                  <div
                    key={cardIdx}
                    className="p-3.5 rounded-md bg-zinc-900/80 border border-zinc-800/80 space-y-3 shadow-xs"
                  >
                    {/* Key & Priority */}
                    <div className="flex items-center justify-between">
                      <Skeleton className="h-3 w-16 bg-zinc-800" />
                      <Skeleton className="h-4 w-12 rounded bg-zinc-850" />
                    </div>

                    {/* Title */}
                    <div className="space-y-1.5">
                      <Skeleton className="h-3.5 w-full bg-zinc-800" />
                      {cardIdx % 2 === 0 && <Skeleton className="h-3.5 w-3/4 bg-zinc-800/70" />}
                    </div>

                    {/* Footer: Assignee & Date */}
                    <div className="flex items-center justify-between pt-1 border-t border-zinc-850/60">
                      <div className="flex items-center gap-1.5">
                        <Skeleton className="h-5 w-5 rounded-full bg-zinc-800" />
                        <Skeleton className="h-3 w-14 bg-zinc-850" />
                      </div>
                      <Skeleton className="h-3 w-16 bg-zinc-850" />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      ) : (
        /* Table / List View Skeleton */
        <div className="flex-1 rounded-lg border border-zinc-800 bg-zinc-900/40 overflow-hidden flex flex-col">
          {/* Table Header Row */}
          <div className="h-10 px-4 bg-zinc-900/80 border-b border-zinc-800 flex items-center justify-between gap-4">
            <Skeleton className="h-4 w-4 rounded bg-zinc-800" />
            <Skeleton className="h-3.5 w-48 bg-zinc-800 flex-1 ml-3" />
            <Skeleton className="h-3.5 w-24 bg-zinc-800 hidden sm:block" />
            <Skeleton className="h-3.5 w-20 bg-zinc-800 hidden md:block" />
            <Skeleton className="h-3.5 w-28 bg-zinc-800 hidden lg:block" />
            <Skeleton className="h-3.5 w-20 bg-zinc-800" />
          </div>

          {/* Table Body Rows */}
          <div className="flex-1 divide-y divide-zinc-850/60 overflow-hidden">
            {[1, 2, 3, 4, 5, 6, 7, 8].map((rowIdx) => (
              <div
                key={rowIdx}
                className="h-12 px-4 flex items-center justify-between gap-4 bg-zinc-950/20"
              >
                <Skeleton className="h-4 w-4 rounded bg-zinc-850" />
                <div className="flex items-center gap-3 flex-1 ml-3">
                  <Skeleton className="h-3.5 w-14 bg-zinc-850 font-mono" />
                  <Skeleton
                    className={`h-3.5 bg-zinc-800 ${
                      rowIdx % 3 === 0 ? 'w-3/5' : rowIdx % 2 === 0 ? 'w-2/5' : 'w-4/5'
                    }`}
                  />
                </div>
                <Skeleton className="h-5 w-20 rounded bg-zinc-850 hidden sm:block" />
                <Skeleton className="h-4 w-16 rounded bg-zinc-850 hidden md:block" />
                <div className="items-center gap-2 hidden lg:flex">
                  <Skeleton className="h-5 w-5 rounded-full bg-zinc-850" />
                  <Skeleton className="h-3 w-16 bg-zinc-850" />
                </div>
                <Skeleton className="h-3.5 w-16 bg-zinc-850" />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

