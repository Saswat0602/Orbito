import * as React from 'react';
import { cn } from '@/lib/utils';

interface SkeletonProps extends React.ComponentProps<'div'> {
  variant?: 'pulse' | 'shimmer' | 'subtle';
}

function Skeleton({ className, variant = 'shimmer', ...props }: SkeletonProps) {
  return (
    <div
      data-slot="skeleton"
      role="status"
      aria-label="Loading content..."
      className={cn(
        'rounded-md select-none',
        variant === 'pulse' && 'animate-pulse bg-zinc-850/90',
        variant === 'subtle' && 'bg-zinc-850/50',
        variant === 'shimmer' &&
          'relative overflow-hidden bg-zinc-850/70 before:absolute before:inset-0 before:-translate-x-full before:animate-[shimmer_2s_infinite] before:bg-gradient-to-r before:from-transparent before:via-zinc-750/30 before:to-transparent',
        className
      )}
      {...props}
    />
  );
}

export { Skeleton };
