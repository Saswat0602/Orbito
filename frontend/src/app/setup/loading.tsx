import { Skeleton } from '@/components/ui/skeleton';

export default function SetupLoading() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-[#09090b] text-zinc-100 p-4">
      <div className="w-full max-w-lg border border-zinc-800 rounded-lg bg-zinc-900/80 p-6 space-y-6 shadow-2xl">
        <div className="text-center space-y-2">
          <Skeleton className="h-6 w-48 mx-auto rounded" />
          <Skeleton className="h-7 w-64 mx-auto" />
          <Skeleton className="h-4 w-72 mx-auto" />
        </div>
        <div className="space-y-4 pt-2">
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Skeleton className="h-3 w-16" />
              <Skeleton className="h-9 w-full rounded-md" />
            </div>
            <div className="space-y-1.5">
              <Skeleton className="h-3 w-16" />
              <Skeleton className="h-9 w-full rounded-md" />
            </div>
          </div>
          <div className="space-y-1.5">
            <Skeleton className="h-3 w-20" />
            <Skeleton className="h-9 w-full rounded-md" />
          </div>
          <div className="space-y-1.5">
            <Skeleton className="h-3 w-24" />
            <Skeleton className="h-9 w-full rounded-md" />
          </div>
          <Skeleton className="h-9 w-full rounded-md pt-2" />
        </div>
      </div>
    </div>
  );
}

