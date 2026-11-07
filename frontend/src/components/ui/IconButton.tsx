import * as React from "react";
import { cn } from "@/lib/utils";

interface IconButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  icon: React.ReactNode;
  variant?: "primary" | "secondary" | "outline" | "ghost";
  size?: "xs" | "sm" | "md" | "lg";
}

const iconButtonVariants = {
  primary: "bg-zinc-100 text-zinc-950 hover:bg-white shadow-xs",
  secondary: "bg-zinc-800 text-zinc-200 hover:bg-zinc-750",
  outline: "border border-zinc-800 text-zinc-300 hover:bg-zinc-900 hover:text-zinc-100",
  ghost: "bg-transparent text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/80",
};

const iconButtonSizes = {
  xs: "h-6 w-6 p-0.5 text-xs rounded",
  sm: "h-8 w-8 p-1.5 text-sm rounded-md",
  md: "h-9 w-9 p-2 text-base rounded-md",
  lg: "h-11 w-11 p-2.5 text-lg rounded-lg",
};

export function IconButton({
  icon,
  variant = "ghost",
  size = "sm",
  disabled = false,
  className,
  ...props
}: IconButtonProps) {
  return (
    <button
      disabled={disabled}
      className={cn(
        "inline-flex items-center justify-center transition-all duration-150 select-none cursor-pointer focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-400 disabled:opacity-50 disabled:pointer-events-none active:scale-[0.97]",
        iconButtonVariants[variant],
        iconButtonSizes[size],
        className
      )}
      {...props}
    >
      {icon}
    </button>
  );
}
