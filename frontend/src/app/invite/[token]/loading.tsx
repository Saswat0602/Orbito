import { Skeleton } from '@/components/ui/skeleton';

export default function InviteVerificationLoading() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-[#09090b] px-4">
      <div className="max-w-md w-full border border-zinc-800 rounded-lg bg-zinc-900/90 p-6 space-y-5 shadow-2xl">
        <div className="flex flex-col items-center text-center space-y-2">
          <Skeleton className="h-12 w-12 rounded-full" />
          <Skeleton className="h-5 w-48" />
          <Skeleton className="h-3.5 w-64" />
        </div>
        <div className="space-y-3 pt-2">
          <div className="p-4 rounded-md border border-zinc-800 bg-zinc-950/60 space-y-2.5">
            <div className="flex items-center gap-3">
              <Skeleton className="h-8 w-8 rounded-full" />
              <div className="space-y-1">
                <Skeleton className="h-3.5 w-28" />
                <Skeleton className="h-3 w-36" />
              </div>
            </div>
            <div className="flex items-center justify-between pt-2 border-t border-zinc-850">
              <Skeleton className="h-4 w-16" />
              <Skeleton className="h-4 w-20 rounded" />
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2 pt-2">
          <Skeleton className="h-9 flex-1 rounded-md" />
          <Skeleton className="h-9 w-24 rounded-md" />
        </div>
      </div>
    </div>
  );
}

