import * as React from "react";
import { cn } from "@/lib/utils";
import { ReactNode } from "react";

interface EmptyStateProps {
  icon: ReactNode;
  title: string;
  description: string;
  action?: ReactNode;
  className?: string;
}

export function EmptyState({ icon, title, description, action, className }: EmptyStateProps) {
  return (
    <div className={cn("flex flex-col items-center justify-center text-center py-12 px-4", className)}>
      <div className="w-14 h-14 bg-zinc-900 border border-zinc-800 rounded-xl flex items-center justify-center mx-auto mb-4 text-zinc-400">
        {icon}
      </div>
      <h3 className="text-base font-semibold text-zinc-100 mb-1.5">{title}</h3>
      <p className="text-xs text-zinc-400 mb-5 max-w-sm leading-relaxed">{description}</p>
      {action}
    </div>
  );
}
