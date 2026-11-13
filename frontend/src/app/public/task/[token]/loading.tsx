import { TaskDetailSkeleton } from '@/components/skeletons';

export default function PublicTaskLoading() {
  return (
    <div className="min-h-screen bg-[#09090b] text-zinc-100 flex flex-col">
      <header className="h-14 border-b border-zinc-800 bg-zinc-950/70 px-6 flex items-center justify-between sticky top-0 z-20">
        <div className="flex items-center gap-3">
          <div className="h-6 w-6 rounded bg-zinc-800 animate-pulse" />
          <div className="h-4 w-20 rounded bg-zinc-800 animate-pulse" />
        </div>
      </header>
      <main className="flex-1 flex justify-center p-4 sm:p-8">
        <div className="w-full max-w-3xl">
          <TaskDetailSkeleton />
        </div>
      </main>
    </div>
  );
}

