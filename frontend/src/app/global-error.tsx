'use client';

import * as React from 'react';
import { useEffect } from 'react';
import { AlertTriangle, RotateCcw } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error('[Orbito Global Root Error]', error);
  }, [error]);

  return (
    <html lang="en" className="dark">
      <body className="bg-[#09090b] text-zinc-100 antialiased min-h-screen flex items-center justify-center p-6 selection:bg-zinc-800">
        <div className="max-w-md w-full border border-rose-900/40 bg-zinc-900/90 p-8 rounded-lg shadow-2xl text-center space-y-6">
          <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-rose-950/40 border border-rose-800/60 text-rose-400">
            <AlertTriangle className="h-7 w-7" />
          </div>

          <div className="space-y-2">
            <span className="text-[11px] font-mono font-semibold uppercase tracking-wider text-rose-400/90">
              Critical Error
            </span>
            <h1 className="text-xl font-bold tracking-tight text-zinc-100">
              System Interface Error
            </h1>
            <p className="text-xs text-zinc-400 leading-relaxed">
              A critical failure prevented the root application shell from loading.
            </p>
            {error.digest && (
              <p className="text-[10px] font-mono text-zinc-500 pt-1">
                Digest: {error.digest}
              </p>
            )}
          </div>

          <div className="pt-2">
            <Button
              variant="primary"
              size="sm"
              onClick={() => reset()}
              className="w-full text-xs gap-1.5"
            >
              <RotateCcw className="h-3.5 w-3.5" />
              <span>Reload Application</span>
            </Button>
          </div>
        </div>
      </body>
    </html>
  );
}
