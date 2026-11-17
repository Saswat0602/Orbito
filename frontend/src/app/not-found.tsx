import Link from 'next/link';
import { ArrowLeft, Home, FileQuestion } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-[#09090b] text-zinc-100 p-6 selection:bg-zinc-800">
      <div className="max-w-md w-full border border-zinc-800 bg-zinc-900/80 p-8 rounded-lg shadow-2xl text-center space-y-6">
        <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-zinc-800/80 border border-zinc-700/60 text-zinc-300">
          <FileQuestion className="h-7 w-7" />
        </div>

        <div className="space-y-2">
          <span className="text-[11px] font-mono font-semibold uppercase tracking-wider text-zinc-500">
            Error 404
          </span>
          <h1 className="text-xl font-bold tracking-tight text-zinc-100">
            Page Not Found
          </h1>
          <p className="text-xs text-zinc-400 leading-relaxed">
            The page, project, or resource you requested could not be located or may have been archived.
          </p>
        </div>

        <div className="flex flex-col sm:flex-row items-center gap-3 pt-2">
          <Button
            asChild
            variant="outline"
            size="sm"
            className="w-full text-xs border-zinc-800 bg-zinc-900 hover:bg-zinc-800"
          >
            <Link href="/" className="inline-flex items-center gap-1.5">
              <ArrowLeft className="h-3.5 w-3.5" />
              <span>Go Back</span>
            </Link>
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
