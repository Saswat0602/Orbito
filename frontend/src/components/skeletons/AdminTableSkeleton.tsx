import * as React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

interface AdminTableSkeletonProps {
  title?: string;
  description?: string;
  rowCount?: number;
}

export function AdminTableSkeleton({
  title = 'Directory',
  description = 'Manage platform resources and permissions.',
  rowCount = 6,
}: AdminTableSkeletonProps) {
  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b] select-none">
      {/* 1. Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-5 border-b border-zinc-800/80">
        <div className="space-y-1.5">
          <div className="flex items-center gap-2">
            <Skeleton className="h-6 w-48 bg-zinc-800 rounded-md" />
            <Skeleton className="h-4 w-16 bg-zinc-850 rounded-full" />
          </div>
          <Skeleton className="h-3.5 w-72 bg-zinc-850" />
        </div>

        <div className="flex items-center gap-2">
          <Skeleton className="h-8 w-28 rounded-md bg-zinc-800" />
        </div>
      </div>

      {/* 2. Filters & Search Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-2 flex-1">
          <Skeleton className="h-8 w-60 rounded-md bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-8 w-28 rounded-md bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-8 w-28 rounded-md bg-zinc-900 border border-zinc-800 hidden sm:block" />
        </div>
        <Skeleton className="h-4 w-32 bg-zinc-850" />
      </div>

      {/* 3. Main Table / List Container */}
      <div className="rounded-lg border border-zinc-800 bg-zinc-900/60 divide-y divide-zinc-850/80 overflow-hidden">
        {/* Mock Header Row */}
        <div className="h-10 px-4 bg-zinc-900/80 flex items-center justify-between gap-4 border-b border-zinc-800">
          <Skeleton className="h-3.5 w-32 bg-zinc-800" />
          <Skeleton className="h-3.5 w-20 bg-zinc-800 hidden sm:block" />
          <Skeleton className="h-3.5 w-16 bg-zinc-850 hidden md:block" />
          <Skeleton className="h-3.5 w-24 bg-zinc-850 hidden lg:block" />
          <Skeleton className="h-3.5 w-12 bg-zinc-850" />
        </div>

        {/* Rows */}
        {Array.from({ length: rowCount }).map((_, idx) => (
          <div
            key={idx}
            className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-zinc-950/20"
          >
            {/* Entity / User Details */}
            <div className="flex items-center gap-3">
              <Skeleton className="h-9 w-9 rounded-md bg-zinc-800 shrink-0" />
              <div className="space-y-1.5">
                <div className="flex items-center gap-2">
                  <Skeleton className="h-4 w-36 bg-zinc-800" />
                  <Skeleton className="h-3.5 w-12 rounded bg-zinc-850" />
                </div>
                <Skeleton className="h-3 w-48 bg-zinc-850" />
              </div>
            </div>

            {/* Badges, Timestamps & Kebab Menu */}
            <div className="flex items-center gap-3 shrink-0">
              <Skeleton className="h-5 w-20 rounded bg-zinc-850 hidden sm:block" />
              <Skeleton className="h-4 w-20 bg-zinc-850 hidden md:block" />
              <Skeleton className="h-6 w-6 rounded bg-zinc-850" />
            </div>
          </div>
        ))}
      </div>

      {/* 4. Pagination Footer Skeleton */}
      <div className="flex items-center justify-between pt-2">
        <Skeleton className="h-3.5 w-36 bg-zinc-850" />
        <div className="flex items-center gap-1.5">
          <Skeleton className="h-7 w-16 rounded bg-zinc-900 border border-zinc-800" />
          <Skeleton className="h-7 w-16 rounded bg-zinc-900 border border-zinc-800" />
        </div>
      </div>
    </div>
  );
}

