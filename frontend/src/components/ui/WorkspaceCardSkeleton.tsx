import React from "react";
import { Card } from "./card";

export const WorkspaceCardSkeleton: React.FC = () => (
  <Card className="p-5 animate-pulse border border-zinc-800 bg-zinc-900/80 rounded-xl space-y-3">
    <div className="flex items-center gap-3">
      <div className="w-8 h-8 rounded-lg bg-zinc-800 shrink-0" />
      <div className="flex-1 space-y-1.5">
        <div className="h-4 bg-zinc-800 rounded w-3/4" />
        <div className="h-3 bg-zinc-800 rounded w-1/2" />
      </div>
    </div>
    <div className="h-3 bg-zinc-800 rounded w-full" />
    <div className="h-3 bg-zinc-800 rounded w-2/3" />
    <div className="flex items-center gap-3 pt-2">
      <div className="h-3 bg-zinc-800 rounded w-16" />
      <div className="h-3 bg-zinc-800 rounded w-16" />
    </div>
  </Card>
);
