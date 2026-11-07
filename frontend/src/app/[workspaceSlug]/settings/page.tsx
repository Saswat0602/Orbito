'use client';

import * as React from 'react';
import { useState } from 'react';
import { useParams } from 'next/navigation';
import { Settings, Building2, Puzzle, Zap, Save, CheckCircle2 } from 'lucide-react';
import { toast } from 'sonner';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import {
  useGetWorkspaceBySlugQuery,
  useUpdateWorkspaceMutation,
} from '@/lib/store/api/workspacesApi';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';

export default function WorkspaceSettingsPage() {
  const params = useParams<{ workspaceSlug: string }>();
  const { workspaceSlug } = params;
  const { can, isManager, isOwner, roleBadge } = useAuthRbac();

  const { data: workspace } = useGetWorkspaceBySlugQuery(workspaceSlug || '', {
    skip: !workspaceSlug,
  });

  const [name, setName] = useState(workspace?.name || '');
  const [description, setDescription] = useState(workspace?.description || '');

  const [updateWorkspace, { isLoading: isUpdating }] = useUpdateWorkspaceMutation();

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!workspace?.id) return;
    try {
      await updateWorkspace({
        id: workspace.id,
        name: name || workspace.name,
        description,
      }).unwrap();
      toast.success('Workspace updated successfully');
    } catch {
      toast.error('Failed to update workspace');
    }
  };

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Settings className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Workspace Settings
            </h1>
          </div>
          <p className="text-xs text-zinc-400">
            Configure parameters and enterprise connections for{' '}
            <strong className="text-zinc-200">{workspace?.name || workspaceSlug}</strong>.
          </p>
        </div>

        <Badge className={`text-xs uppercase ${roleBadge.className}`}>
          {roleBadge.label}
        </Badge>
      </div>

      <Tabs defaultValue="general" className="space-y-6">
        <TabsList className="bg-zinc-900 border border-zinc-800 p-1 rounded-md">
          <TabsTrigger value="general" className="gap-2 text-xs">
            <Building2 className="h-3.5 w-3.5" />
            <span>General</span>
          </TabsTrigger>
          {(isManager || isOwner) && (
            <>
              <TabsTrigger value="integrations" className="gap-2 text-xs">
                <Puzzle className="h-3.5 w-3.5" />
                <span>Integrations</span>
              </TabsTrigger>
              <TabsTrigger value="automations" className="gap-2 text-xs">
                <Zap className="h-3.5 w-3.5" />
                <span>Automations</span>
              </TabsTrigger>
            </>
          )}
        </TabsList>

        {/* General Settings */}
        <TabsContent value="general">
          <Card className="border-zinc-800 bg-zinc-900/80 max-w-2xl">
            <CardHeader>
              <CardTitle className="text-base">Workspace Identity</CardTitle>
              <CardDescription className="text-xs">
                Visible name and identifier for all projects in this workspace.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSave} className="space-y-4">
                <div className="space-y-1.5">
                  <Label htmlFor="wsName">Workspace Name</Label>
                  <Input
                    id="wsName"
                    defaultValue={workspace?.name}
                    onChange={(e) => setName(e.target.value)}
                    disabled={!isManager && !isOwner}
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="wsSlug">Workspace Slug</Label>
                  <Input
                    id="wsSlug"
                    defaultValue={workspace?.slug}
                    disabled
                    className="bg-zinc-950 text-zinc-500 cursor-not-allowed font-mono"
                  />
                  <p className="text-[11px] text-zinc-500">
                    Slugs are permanent identifiers used in URL routes.
                  </p>
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="wsDesc">Description</Label>
                  <Input
                    id="wsDesc"
                    defaultValue={workspace?.description || ''}
                    onChange={(e) => setDescription(e.target.value)}
                    disabled={!isManager && !isOwner}
                  />
                </div>

                {(isManager || isOwner) && (
                  <Button
                    type="submit"
                    variant="primary"
                    size="sm"
                    className="gap-2 mt-2"
                    isLoading={isUpdating}
                  >
                    <Save className="h-3.5 w-3.5" />
                    <span>Save Changes</span>
                  </Button>
                )}
              </form>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Integrations */}
        {(isManager || isOwner) && (
          <TabsContent value="integrations">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-3xl">
              <Card className="border-zinc-800 bg-zinc-900/80 p-5 space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-semibold text-zinc-200">Jira Cloud</h4>
                  <Badge variant="outline" className="text-[10px]">Configured</Badge>
                </div>
                <p className="text-xs text-zinc-400">
                  Bidirectional issue synchronization with SSRF protection.
                </p>
                <Button variant="secondary" size="sm" className="w-full text-xs">
                  Manage Jira Connection
                </Button>
              </Card>

              <Card className="border-zinc-800 bg-zinc-900/80 p-5 space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-semibold text-zinc-200">Trello Boards</h4>
                  <Badge variant="outline" className="text-[10px]">Configured</Badge>
                </div>
                <p className="text-xs text-zinc-400">
                  Import boards and cards into project status columns.
                </p>
                <Button variant="secondary" size="sm" className="w-full text-xs">
                  Manage Trello Connection
                </Button>
              </Card>
            </div>
          </TabsContent>
        )}

        {/* Automations */}
        {(isManager || isOwner) && (
          <TabsContent value="automations">
            <Card className="border-zinc-800 bg-zinc-900/80 max-w-2xl p-8 text-center space-y-2">
              <Zap className="h-8 w-8 text-zinc-500 mx-auto" />
              <h4 className="text-sm font-semibold text-zinc-200">
                Workspace Automation Engine
              </h4>
              <p className="text-xs text-zinc-400 max-w-md mx-auto">
                Trigger background rule jobs whenever tasks move or status transitions occur.
              </p>
            </Card>
          </TabsContent>
        )}
      </Tabs>
    </div>
  );
}

