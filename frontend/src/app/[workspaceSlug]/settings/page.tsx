'use client';

import * as React from 'react';
import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import {
  Settings,
  Building2,
  Users,
  Zap,
  Save,
  Trash2,
  Copy,
  Check,
  AlertTriangle,
  UserPlus,
} from 'lucide-react';
import { toast } from 'sonner';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import {
  useGetWorkspaceBySlugQuery,
  useUpdateWorkspaceMutation,
  useDeleteWorkspaceMutation,
  useGetWorkspaceMembersQuery,
  useRemoveWorkspaceMemberMutation,
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
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { ConfirmationModal } from '@/components/modals/ConfirmationModal';
import { AutomationsTab } from '@/features/automations';
import { InviteMemberModal } from '@/components/modals/InviteMemberModal';
import { AdminSettingsSkeleton } from '@/components/skeletons';
import { Skeleton } from '@/components/ui/skeleton';
import { getRoleBadgeStyle } from '@/lib/rbac';

export default function WorkspaceSettingsPage() {
  const params = useParams<{ workspaceSlug: string }>();
  const router = useRouter();
  const { workspaceSlug } = params;
  const { isManager, isOwner, isSuperAdmin, roleBadge } = useAuthRbac();

  const { data: workspace, isLoading: isWsLoading } = useGetWorkspaceBySlugQuery(
    workspaceSlug || '',
    { skip: !workspaceSlug }
  );

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const [deleteConfirmation, setDeleteConfirmation] = useState('');
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const [memberToRemove, setMemberToRemove] = useState<{ id: string; name: string } | null>(null);

  useEffect(() => {
    if (workspace) {
      setName(workspace.name || '');
      setDescription(workspace.description || '');
    }
  }, [workspace]);

  const [updateWorkspace, { isLoading: isUpdating }] = useUpdateWorkspaceMutation();
  const [deleteWorkspace, { isLoading: isDeleting }] = useDeleteWorkspaceMutation();

  const { data: members = [], isLoading: isMembersLoading } = useGetWorkspaceMembersQuery(
    workspace?.id || '',
    { skip: !workspace?.id }
  );
  const [removeMember, { isLoading: isRemovingMember }] = useRemoveWorkspaceMemberMutation();

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!workspace?.id) return;
    try {
      await updateWorkspace({
        id: workspace.id,
        name: name.trim() || workspace.name,
        description: description.trim(),
      }).unwrap();
      toast.success('Workspace details updated successfully');
    } catch {
      toast.error('Failed to update workspace');
    }
  };

  const handleCopy = (text: string, key: string) => {
    navigator.clipboard.writeText(text);
    setCopiedKey(key);
    toast.success('Copied to clipboard');
    setTimeout(() => setCopiedKey(null), 2000);
  };

  const handleConfirmRemoveMember = async () => {
    if (!workspace?.id || !memberToRemove) return;
    try {
      await removeMember({ workspaceId: workspace.id, memberId: memberToRemove.id }).unwrap();
      toast.success(`${memberToRemove.name} removed`);
      setMemberToRemove(null);
    } catch {
      toast.error('Failed to remove member');
    }
  };

  const handleDeleteWorkspace = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!workspace?.id) return;
    if (deleteConfirmation !== workspace.name) {
      toast.error('Workspace name does not match');
      return;
    }

    try {
      await deleteWorkspace(workspace.id).unwrap();
      toast.success('Workspace deleted successfully');
      router.push('/');
    } catch {
      toast.error('Failed to delete workspace');
    }
  };

  if (isWsLoading) {
    return <AdminSettingsSkeleton />;
  }

  if (!workspace) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b] text-xs text-zinc-500">
        Workspace not found.
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5">
            <Settings className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Workspace Settings
            </h1>
          </div>
          <p className="text-xs text-zinc-400">
            Configure metadata, member permissions, and automations for{' '}
            <strong className="text-zinc-200">{workspace.name}</strong>.
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
          <TabsTrigger value="members" className="gap-2 text-xs">
            <Users className="h-3.5 w-3.5" />
            <span>Members & Roles</span>
          </TabsTrigger>
          {(isManager || isOwner) && (
            <TabsTrigger value="automations" className="gap-2 text-xs">
              <Zap className="h-3.5 w-3.5" />
              <span>Automations</span>
            </TabsTrigger>
          )}
          {(isOwner || isSuperAdmin) && (
            <TabsTrigger value="danger" className="gap-2 text-xs text-rose-400 data-[state=active]:text-rose-300">
              <AlertTriangle className="h-3.5 w-3.5" />
              <span>Danger Zone</span>
            </TabsTrigger>
          )}
        </TabsList>

        {/* General Settings */}
        <TabsContent value="general">
          <div className="space-y-6 max-w-3xl">
            <Card className="border-zinc-800 bg-zinc-900/80">
              <CardHeader>
                <CardTitle className="text-base">Workspace Identity</CardTitle>
                <CardDescription className="text-xs">
                  Visible name and description for all projects and workstreams.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleSave} className="space-y-4">
                  <div className="space-y-1.5">
                    <Label htmlFor="wsName" className="text-xs text-zinc-300">Workspace Name</Label>
                    <Input
                      id="wsName"
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      disabled={!isManager && !isOwner}
                      className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                    />
                  </div>

                  <div className="space-y-1.5">
                    <Label htmlFor="wsDesc" className="text-xs text-zinc-300">Description</Label>
                    <Input
                      id="wsDesc"
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                      disabled={!isManager && !isOwner}
                      placeholder="e.g. Core product and platform deliverables"
                      className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                    />
                  </div>

                  {(isManager || isOwner) && (
                    <Button
                      type="submit"
                      variant="primary"
                      size="sm"
                      className="gap-2 mt-2 text-xs"
                      isLoading={isUpdating}
                    >
                      <Save className="h-3.5 w-3.5" />
                      <span>Save Changes</span>
                    </Button>
                  )}
                </form>
              </CardContent>
            </Card>

            {/* Read-only IDs & Identifiers */}
            <Card className="border-zinc-800 bg-zinc-900/80 p-5 space-y-4">
              <h4 className="text-xs font-semibold text-zinc-200">System Identifiers</h4>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1">
                  <span className="text-[10px] uppercase font-mono text-zinc-500">Workspace Slug</span>
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-zinc-300">{workspace.slug}</span>
                    <button
                      type="button"
                      onClick={() => handleCopy(workspace.slug, 'slug')}
                      className="text-zinc-500 hover:text-zinc-200 transition-colors"
                      title="Copy Slug"
                    >
                      {copiedKey === 'slug' ? <Check className="h-3 w-3 text-emerald-400" /> : <Copy className="h-3 w-3" />}
                    </button>
                  </div>
                </div>

                <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1">
                  <span className="text-[10px] uppercase font-mono text-zinc-500">Workspace ID</span>
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-zinc-400 truncate max-w-[180px]">{workspace.id}</span>
                    <button
                      type="button"
                      onClick={() => handleCopy(workspace.id, 'id')}
                      className="text-zinc-500 hover:text-zinc-200 transition-colors"
                      title="Copy ID"
                    >
                      {copiedKey === 'id' ? <Check className="h-3 w-3 text-emerald-400" /> : <Copy className="h-3 w-3" />}
                    </button>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        </TabsContent>

        {/* Members & Roles */}
        <TabsContent value="members">
          <Card className="border-zinc-800 bg-zinc-900/80 max-w-4xl">
            <CardHeader className="flex flex-row items-center justify-between pb-4">
              <div>
                <CardTitle className="text-base">Workspace Collaborators</CardTitle>
                <CardDescription className="text-xs">
                  Members with access to projects inside {workspace.name}.
                </CardDescription>
              </div>
              {(isManager || isOwner) && (
                <Button
                  variant="primary"
                  size="sm"
                  onClick={() => setIsInviteOpen(true)}
                  className="gap-1.5 text-xs"
                >
                  <UserPlus className="h-3.5 w-3.5" />
                  <span>Invite Member</span>
                </Button>
              )}
            </CardHeader>
            <CardContent>
              {isMembersLoading ? (
                <div className="space-y-3 py-2">
                  {Array.from({ length: 4 }).map((_, i) => (
                    <div key={i} className="flex items-center justify-between p-3 rounded-md border border-zinc-850 bg-zinc-950/40">
                      <div className="flex items-center gap-3">
                        <Skeleton className="h-8 w-8 rounded-full" />
                        <div className="space-y-1">
                          <Skeleton className="h-3.5 w-28" />
                          <Skeleton className="h-3 w-36" />
                        </div>
                      </div>
                      <Skeleton className="h-5 w-16 rounded" />
                    </div>
                  ))}
                </div>
              ) : members.length === 0 ? (
                <div className="py-8 text-center text-xs text-zinc-500">No members found.</div>
              ) : (
                <div className="divide-y divide-zinc-800 border border-zinc-850 rounded-md overflow-hidden">
                  {members.map((member) => {
                    const roleStyle = getRoleBadgeStyle(member.role);
                    const memberName =
                      member.user?.firstName && member.user?.lastName
                        ? `${member.user.firstName} ${member.user.lastName}`
                        : member.user?.username || 'User';
                    const initials =
                      ((member.user?.firstName?.[0] || '') + (member.user?.lastName?.[0] || '')).toUpperCase() || 'U';

                    return (
                      <div
                        key={member.id}
                        className="flex items-center justify-between p-3.5 bg-zinc-950/40 hover:bg-zinc-900/40 transition-colors"
                      >
                        <div className="flex items-center gap-3 min-w-0">
                          <Avatar className="h-8 w-8 rounded-md border border-zinc-800">
                            <AvatarFallback className="bg-zinc-800 text-zinc-200 text-xs font-bold">
                              {initials}
                            </AvatarFallback>
                          </Avatar>
                          <div className="flex flex-col min-w-0">
                            <span className="text-xs font-semibold text-zinc-200 truncate">{memberName}</span>
                            <span className="text-[11px] text-zinc-500 truncate">{member.user?.email}</span>
                          </div>
                        </div>

                        <div className="flex items-center gap-3 shrink-0">
                          <Badge className={`text-[10px] uppercase ${roleStyle.className}`}>
                            {roleStyle.label}
                          </Badge>

                          {(isManager || isOwner) && member.role !== 'OWNER' && (
                            <button
                              type="button"
                              onClick={() => setMemberToRemove({ id: member.id, name: memberName })}
                              className="p-1 rounded text-zinc-500 hover:text-rose-400 hover:bg-zinc-800 transition-colors cursor-pointer"
                              title="Remove from workspace"
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </button>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Automations */}
        {(isManager || isOwner) && (
          <TabsContent value="automations">
            <AutomationsTab
              workspaceId={workspace.id}
              organizationId={workspace.organizationId}
              isManagerOrOwner={isManager || isOwner}
            />
          </TabsContent>
        )}

        {/* Danger Zone */}
        {(isOwner || isSuperAdmin) && (
          <TabsContent value="danger">
            <Card className="border-rose-950/60 bg-zinc-900/80 max-w-2xl border">
              <CardHeader>
                <div className="flex items-center gap-2 text-rose-400">
                  <AlertTriangle className="h-4 w-4" />
                  <CardTitle className="text-base text-rose-300">Delete Workspace</CardTitle>
                </div>
                <CardDescription className="text-xs text-zinc-400">
                  Permanently deletes this workspace and all associated projects, tasks, sprints, and automations. This action is irreversible.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleDeleteWorkspace} className="space-y-4">
                  <div className="space-y-1.5">
                    <Label htmlFor="confirmName" className="text-xs text-zinc-300">
                      Please type <strong className="text-rose-400">{workspace.name}</strong> to confirm:
                    </Label>
                    <Input
                      id="confirmName"
                      value={deleteConfirmation}
                      onChange={(e) => setDeleteConfirmation(e.target.value)}
                      placeholder={workspace.name}
                      className="bg-zinc-950 border-rose-900/60 text-xs text-zinc-100"
                    />
                  </div>

                  <Button
                    type="submit"
                    variant="destructive"
                    size="sm"
                    disabled={deleteConfirmation !== workspace.name || isDeleting}
                    isLoading={isDeleting}
                    className="gap-2 text-xs"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                    <span>Permanently Delete Workspace</span>
                  </Button>
                </form>
              </CardContent>
            </Card>
          </TabsContent>
        )}
      </Tabs>

      {/* Invite Member Modal */}
      <InviteMemberModal
        open={isInviteOpen}
        onOpenChange={setIsInviteOpen}
        workspaceId={workspace.id}
        workspaceName={workspace.name}
        organizationId={workspace.organizationId}
      />

      {/* Remove Member Confirmation Modal */}
      <ConfirmationModal
        open={!!memberToRemove}
        onOpenChange={(open) => !open && setMemberToRemove(null)}
        title="Remove Member from Workspace"
        description={
          <span>
            Are you sure you want to remove{' '}
            <strong className="text-zinc-200">{memberToRemove?.name}</strong> from{' '}
            <strong className="text-zinc-200">{workspace.name}</strong>?
            They will lose access to all projects and tasks in this workspace.
          </span>
        }
        confirmLabel="Remove Member"
        variant="destructive"
        isLoading={isRemovingMember}
        onConfirm={handleConfirmRemoveMember}
      />
    </div>
  );
}
