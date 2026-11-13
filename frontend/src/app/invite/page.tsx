'use client';

import * as React from 'react';
import { Suspense, useEffect } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { Skeleton } from '@/components/ui/skeleton';

function InviteSkeletonFallback() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-[#09090b] px-4">
      <div className="max-w-md w-full border border-zinc-800 rounded-lg bg-zinc-900/90 p-6 space-y-5 shadow-2xl">
        <div className="flex flex-col items-center text-center space-y-2">
          <Skeleton className="h-12 w-12 rounded-full" />
          <Skeleton className="h-5 w-48" />
          <Skeleton className="h-3.5 w-64" />
        </div>
        <div className="space-y-3 pt-2">
          <Skeleton className="h-20 w-full rounded-md" />
        </div>
        <Skeleton className="h-9 w-full rounded-md" />
      </div>
    </div>
  );
}

function InviteRedirectContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const token = searchParams.get('token');

  useEffect(() => {
    if (token) {
      router.replace(`/invite/${token}`);
    } else {
      router.replace('/login');
    }
  }, [token, router]);

  return <InviteSkeletonFallback />;
}

export default function InviteRedirectPage() {
  return (
    <Suspense fallback={<InviteSkeletonFallback />}>
      <InviteRedirectContent />
    </Suspense>
  );
}
