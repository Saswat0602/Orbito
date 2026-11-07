import * as React from "react";
import { cn } from "@/lib/utils";

export interface InputProps extends React.ComponentProps<"input"> {
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  error?: string;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, leftIcon, rightIcon, error, ...props }, ref) => {
    const inputElement = (
      <input
        ref={ref}
        type={type}
        data-slot="input"
        className={cn(
          "flex h-9 w-full min-w-0 rounded-md border border-zinc-800 bg-zinc-900/80 px-3 py-1 text-sm text-zinc-100 shadow-xs transition-colors outline-none",
          "file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium",
          "placeholder:text-zinc-500",
          "focus-visible:border-zinc-400 focus-visible:ring-1 focus-visible:ring-zinc-400",
          "disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50",
          leftIcon && "pl-9",
          rightIcon && "pr-9",
          error && "border-rose-500 focus-visible:ring-rose-500/50",
          className
        )}
        {...props}
      />
    );

    if (leftIcon || rightIcon || error) {
      return (
        <div className="relative w-full">
          {leftIcon && (
            <div className="absolute left-3 top-1/2 -translate-y-1/2 flex items-center pointer-events-none text-zinc-400">
              {leftIcon}
            </div>
          )}
          {inputElement}
          {rightIcon && (
            <div className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center text-zinc-400">
              {rightIcon}
            </div>
          )}
          {error && (
            <p className="mt-1 text-xs text-rose-400 font-medium">{error}</p>
          )}
        </div>
      );
    }

    return inputElement;
  }
);

Input.displayName = "Input";

export { Input };
