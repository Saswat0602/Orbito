'use client';

import * as React from 'react';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Briefcase, Plus, ArrowRight } from 'lucide-react';
import { toast } from 'sonner';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import {
  useGetWorkspacesQuery,
  useCreateWorkspaceMutation,
} from '@/lib/store/api/workspacesApi';
import { useGetOrganizationsQuery } from '@/lib/store/api/organizationsApi';
import {
  setActiveWorkspace,
  setActiveWorkspaceSlug,
} from '@/lib/store/slices/authSlice';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

export default function RootRedirectPage() {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const { user, isAuthenticated } = useAuthRbac();
  const authState = useAppSelector((state) => state.auth);

  const { data: organizations = [] } = useGetOrganizationsQuery(undefined, {
    skip: !authState.isAuthenticated || !authState.user,
  });
  const activeOrgId = authState.activeOrganizationId || organizations[0]?.id;

  const {
    data: workspaces = [],
    isLoading: isWorkspacesLoading,
  } = useGetWorkspacesQuery(
    activeOrgId ? { organizationId: activeOrgId } : undefined,
    { skip: !activeOrgId }
  );

  const [createWorkspace, { isLoading: isCreating }] = useCreateWorkspaceMutation();
  const [newWorkspaceName, setNewWorkspaceName] = useState('');

  // If user has workspaces, automatically redirect to their primary workspace URL
  useEffect(() => {
    if (!isWorkspacesLoading && workspaces.length > 0) {
      const primary = workspaces[0];
      dispatch(setActiveWorkspace(primary.id));
      dispatch(setActiveWorkspaceSlug(primary.slug));
      router.replace(`/${primary.slug}`);
    }
  }, [workspaces, isWorkspacesLoading, dispatch, router]);

  const handleCreateInitialWorkspace = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newWorkspaceName.trim() || !activeOrgId) return;

    try {
      const created = await createWorkspace({
        name: newWorkspaceName.trim(),
        organizationId: activeOrgId,
      }).unwrap();

      toast.success('Workspace created successfully');
      dispatch(setActiveWorkspace(created.id));
      dispatch(setActiveWorkspaceSlug(created.slug));
      router.replace(`/${created.slug}`);
    } catch {
      toast.error('Failed to create workspace');
    }
  };

  if (isWorkspacesLoading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b] min-h-screen">
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
          <span className="text-xs text-zinc-400">Resolving workspace...</span>
        </div>
      </div>
    );
  }

  // If authenticated but no workspaces exist yet:
  return (
    <div className="flex-1 flex items-center justify-center bg-[#09090b] p-6 min-h-screen">
      <Card className="max-w-md w-full border-zinc-800 bg-zinc-900/80 shadow-2xl">
        <CardHeader className="text-center pb-2">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-md bg-zinc-800 text-zinc-100 mb-2">
            <Briefcase className="h-6 w-6" />
          </div>
          <CardTitle className="text-xl font-bold text-zinc-100">
            Create Your Workspace
          </CardTitle>
          <CardDescription className="text-xs text-zinc-400">
            Every project, task, sprint, and collaborator belongs to a workspace.
          </CardDescription>
        </CardHeader>

        <CardContent className="pt-4">
          <form onSubmit={handleCreateInitialWorkspace} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="ws-name" className="text-xs text-zinc-300">
                Workspace Name
              </Label>
              <Input
                id="ws-name"
                placeholder="e.g. Engineering HQ"
                value={newWorkspaceName}
                onChange={(e) => setNewWorkspaceName(e.target.value)}
                required
                className="bg-zinc-950/80 border-zinc-800 text-zinc-100 text-xs"
              />
            </div>

            <Button
              type="submit"
              variant="primary"
              size="sm"
              isLoading={isCreating}
              className="w-full gap-2 text-xs"
            >
              <span>Get Started</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
