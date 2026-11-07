import * as React from "react";
import { cn } from "@/lib/utils";

interface LoadingSkeletonProps {
  className?: string;
}

export function LoadingSkeleton({ className }: LoadingSkeletonProps) {
  return (
    <div className={cn("max-w-7xl mx-auto p-6", className)}>
      <div className="animate-pulse">
        <div className="h-6 bg-zinc-850 rounded w-1/4 mb-3" />
        <div className="h-4 bg-zinc-850 rounded w-1/2 mb-6" />
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-4">
              <div className="h-5 bg-zinc-850 rounded w-1/4 mb-4" />
              <div className="space-y-3">
                <div className="h-4 bg-zinc-850 rounded" />
                <div className="h-4 bg-zinc-850 rounded" />
                <div className="h-4 bg-zinc-850 rounded" />
              </div>
            </div>
          </div>
          <div className="lg:col-span-1">
            <div className="bg-zinc-900 rounded-xl border border-zinc-800 p-4">
              <div className="h-5 bg-zinc-850 rounded w-1/3 mb-4" />
              <div className="space-y-3">
                <div className="h-4 bg-zinc-850 rounded" />
                <div className="h-4 bg-zinc-850 rounded" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
