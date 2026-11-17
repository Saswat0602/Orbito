'use client';

import * as React from 'react';
import { useEffect } from 'react';
import Link from 'next/link';
import { AlertTriangle, RotateCcw, Home } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function ErrorBoundary({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log unexpected runtime client error for observability
    console.error('[Orbito App Router Error Boundary]', error);
  }, [error]);

  return (
    <div className="flex min-h-[70vh] flex-col items-center justify-center bg-[#09090b] text-zinc-100 p-6 selection:bg-zinc-800">
      <div className="max-w-md w-full border border-rose-900/40 bg-zinc-900/90 p-8 rounded-lg shadow-2xl text-center space-y-6">
        <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-rose-950/40 border border-rose-800/60 text-rose-400">
          <AlertTriangle className="h-7 w-7" />
        </div>

        <div className="space-y-2">
          <span className="text-[11px] font-mono font-semibold uppercase tracking-wider text-rose-400/90">
            Application Error
          </span>
          <h1 className="text-xl font-bold tracking-tight text-zinc-100">
            Something went wrong
          </h1>
          <p className="text-xs text-zinc-400 leading-relaxed">
            An unexpected error occurred while rendering this view. Your data is safe.
          </p>
          {error.digest && (
            <p className="text-[10px] font-mono text-zinc-500 pt-1">
              Error Digest: {error.digest}
            </p>
          )}
        </div>

        <div className="flex flex-col sm:flex-row items-center gap-3 pt-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => reset()}
            className="w-full text-xs border-zinc-800 bg-zinc-900 hover:bg-zinc-800 gap-1.5"
          >
            <RotateCcw className="h-3.5 w-3.5" />
            <span>Try Again</span>
          </Button>

          <Button
            asChild
            variant="primary"
            size="sm"
            className="w-full text-xs"
          >
            <Link href="/" className="inline-flex items-center gap-1.5">
              <Home className="h-3.5 w-3.5" />
              <span>Dashboard</span>
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
