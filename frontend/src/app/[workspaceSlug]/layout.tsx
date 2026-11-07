'use client';

import * as React from 'react';
import { useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAppDispatch } from '@/lib/store/hooks';
import {
  setActiveWorkspace,
  setActiveWorkspaceSlug,
} from '@/lib/store/slices/authSlice';
import { useGetWorkspaceBySlugQuery } from '@/lib/store/api/workspacesApi';

export default function WorkspaceLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const params = useParams<{ workspaceSlug: string }>();
  const router = useRouter();
  const dispatch = useAppDispatch();
  const workspaceSlug = params?.workspaceSlug;

  const {
    data: workspace,
    isLoading,
    isError,
  } = useGetWorkspaceBySlugQuery(workspaceSlug || '', {
    skip: !workspaceSlug,
  });

  useEffect(() => {
    if (workspace) {
      dispatch(setActiveWorkspace(workspace.id));
      dispatch(setActiveWorkspaceSlug(workspace.slug));
    }
  }, [workspace, dispatch]);

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b] min-h-[50vh]">
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
          <span className="text-xs text-zinc-400">Loading workspace...</span>
        </div>
      </div>
    );
  }

  if (isError || (!isLoading && !workspace)) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b] p-6">
        <div className="max-w-md w-full text-center space-y-4 p-8 rounded-md border border-zinc-800 bg-zinc-900/60">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-rose-500/10 text-rose-400">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              className="h-6 w-6"
            >
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
          </div>
          <h2 className="text-lg font-bold text-zinc-100">Workspace Not Found</h2>
          <p className="text-xs text-zinc-400">
            The workspace <strong className="text-zinc-200">{workspaceSlug}</strong> does not
            exist or you do not have permission to view it.
          </p>
          <button
            onClick={() => router.push('/')}
            className="inline-flex items-center justify-center px-4 py-2 rounded-md text-xs font-semibold bg-zinc-100 text-zinc-950 hover:bg-zinc-200 transition-colors"
          >
            Return to Overview
          </button>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}

