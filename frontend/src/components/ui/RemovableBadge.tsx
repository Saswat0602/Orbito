import * as React from "react";
import { ReactNode } from "react";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";

interface RemovableBadgeProps {
  children: ReactNode;
  onRemove?: () => void;
  variant?: "default" | "primary" | "success" | "warning" | "info" | "secondary" | "error";
  className?: string;
}

const badgeVariants = {
  default: "bg-zinc-800 text-zinc-300 border-zinc-700",
  primary: "bg-zinc-100 text-zinc-950 border-zinc-200",
  success: "bg-emerald-950/50 text-emerald-300 border-emerald-800/40",
  warning: "bg-amber-950/50 text-amber-300 border-amber-800/40",
  info: "bg-sky-950/50 text-sky-300 border-sky-800/40",
  secondary: "bg-zinc-800 text-zinc-300 border-zinc-700",
  error: "bg-rose-950/50 text-rose-300 border-rose-800/40",
};

export function RemovableBadge({
  children,
  onRemove,
  variant = "default",
  className,
}: RemovableBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-xs font-medium border transition-colors",
        badgeVariants[variant],
        className
      )}
    >
      <span>{children}</span>
      {onRemove && (
        <button
          type="button"
          onClick={onRemove}
          className="text-zinc-400 hover:text-rose-400 transition-colors cursor-pointer"
          aria-label="Remove"
        >
          <X className="h-3 w-3" />
        </button>
      )}
    </span>
  );
}
