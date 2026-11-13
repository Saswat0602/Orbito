import { Skeleton } from '@/components/ui/skeleton';

export default function ForgotPasswordLoading() {
  return (
    <div className="flex min-h-screen items-center justify-center p-4 py-8 bg-[#09090b]">
      <div className="w-full max-w-md border border-zinc-800 rounded-lg bg-zinc-900/90 p-6 space-y-6 shadow-2xl">
        <div className="text-center space-y-2">
          <Skeleton className="h-6 w-44 mx-auto rounded" />
          <Skeleton className="h-4 w-60 mx-auto" />
        </div>
        <div className="space-y-3 pt-2">
          <div className="space-y-1.5">
            <Skeleton className="h-3 w-16" />
            <Skeleton className="h-9 w-full rounded-md" />
          </div>
          <Skeleton className="h-9 w-full rounded-md" />
        </div>
      </div>
    </div>
  );
}

