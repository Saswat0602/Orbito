'use client';

import * as React from 'react';
import { useState, useMemo, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import {
  Users,
  UserPlus,
  Mail,
  Shield,
  Trash2,
  Search,
  X,
  Clock,
  UserCheck,
  RotateCcw,
  ExternalLink,
} from 'lucide-react';
import { toast } from 'sonner';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import {
  useGetWorkspaceBySlugQuery,
  useGetWorkspaceMembersQuery,
  useRemoveWorkspaceMemberMutation,
  useUpdateWorkspaceMemberRoleMutation,
} from '@/lib/store/api/workspacesApi';
import { useGetEntityInvitationsQuery } from '@/lib/store/api/invitationsApi';
import { Role, getRoleBadgeStyle } from '@/lib/rbac';
import {
  InviteMemberModal,
  PendingInvitationsList,
} from '@/features/workspaces';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { ConfirmationModal } from '@/components/modals/ConfirmationModal';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export default function WorkspaceMembersPage() {
  const params = useParams<{ workspaceSlug: string }>();
  const workspaceSlug = params?.workspaceSlug || '';
  const { isManager, isOwner, isSuperAdmin, user } = useAuthRbac();

  const canManage = isManager || isOwner || isSuperAdmin;

  const [activeTab, setActiveTab] = useState<'members' | 'invitations'>('members');
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const [memberToRemove, setMemberToRemove] = useState<{ id: string; name: string } | null>(null);

  // Debounce search
  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedSearch(searchQuery.trim().toLowerCase());
    }, 250);
    return () => clearTimeout(handler);
  }, [searchQuery]);

  const { data: workspace } = useGetWorkspaceBySlugQuery(workspaceSlug, {
    skip: !workspaceSlug,
  });

  const { data: members = [], isLoading: isMembersLoading } =
    useGetWorkspaceMembersQuery(workspace?.id || '', {
      skip: !workspace?.id,
    });

  const { data: invitations = [] } = useGetEntityInvitationsQuery(
    { entityType: 'workspace', entityId: workspace?.id || '' },
    { skip: !workspace?.id }
  );

  const pendingInvitations = invitations.filter(
    (inv) => inv.status === 'PENDING' || !inv.status
  );

  const [removeMember, { isLoading: isRemoving }] = useRemoveWorkspaceMemberMutation();
  const [updateMemberRole, { isLoading: isUpdatingRole }] =
    useUpdateWorkspaceMemberRoleMutation();

  // Role Distribution
  const roleCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    members.forEach((m) => {
      const r = m.role || 'MEMBER';
      counts[r] = (counts[r] || 0) + 1;
    });
    return counts;
  }, [members]);

  // Filtered Members
  const filteredMembers = useMemo(() => {
    if (!debouncedSearch) return members;
    return members.filter((m) => {
      const fullName = `${m.user?.firstName || ''} ${m.user?.lastName || ''}`.toLowerCase();
      const username = (m.user?.username || '').toLowerCase();
      const email = (m.user?.email || '').toLowerCase();
      return (
        fullName.includes(debouncedSearch) ||
        username.includes(debouncedSearch) ||
        email.includes(debouncedSearch)
      );
    });
  }, [members, debouncedSearch]);

  const handleRoleChange = async (memberId: string, newRole: string, memberName: string) => {
    if (!workspace?.id || !canManage) return;
    try {
      await updateMemberRole({
        workspaceId: workspace.id,
        memberId,
        role: newRole,
      }).unwrap();
      toast.success(`${memberName}'s role updated to ${newRole}`);
    } catch {
      toast.error('Failed to update member role');
    }
  };

  const handleConfirmRemoveMember = async () => {
    if (!workspace?.id || !canManage || !memberToRemove) return;
    try {
      await removeMember({ workspaceId: workspace.id, memberId: memberToRemove.id }).unwrap();
      toast.success(`${memberToRemove.name} removed from workspace`);
      setMemberToRemove(null);
    } catch {
      toast.error('Failed to remove member');
    }
  };

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b]">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5">
            <div className="h-8 w-8 rounded-md bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-100">
              <Users className="h-4.5 w-4.5" />
            </div>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-zinc-100">
              Workspace Team & Members
            </h1>
            <Badge
              variant="outline"
              className="text-xs font-mono font-bold bg-zinc-950 text-zinc-300 border-zinc-800"
            >
              {members.length} {members.length === 1 ? 'Member' : 'Members'}
            </Badge>
          </div>
          <p className="text-xs text-zinc-400">
            Manage collaborators, assign role permissions, and track pending invitations for{' '}
            <strong className="text-zinc-200">
              {workspace?.name || workspaceSlug}
            </strong>
            .
          </p>
        </div>

        {canManage && (
          <Button
            variant="primary"
            size="sm"
            className="gap-1.5 text-xs shrink-0"
            onClick={() => setIsInviteOpen(true)}
          >
            <UserPlus className="h-3.5 w-3.5" />
            <span>Invite Member</span>
          </Button>
        )}
      </div>

      {/* Tabs Switcher: Active Members vs Pending Invitations */}
      <div className="flex items-center justify-between pb-1 border-b border-zinc-850">
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setActiveTab('members')}
            className={`flex items-center gap-2 px-3.5 py-2 rounded-t-md text-xs font-semibold transition-colors border-b-2 -mb-px cursor-pointer ${
              activeTab === 'members'
                ? 'border-violet-500 text-zinc-100 bg-zinc-900/60'
                : 'border-transparent text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <Users className="h-3.5 w-3.5 text-zinc-400" />
            <span>Active Members</span>
            <span className="text-[10px] font-mono px-1.5 py-0.2 rounded-full bg-zinc-800 text-zinc-300 font-bold">
              {members.length}
            </span>
          </button>

          <button
            type="button"
            onClick={() => setActiveTab('invitations')}
            className={`flex items-center gap-2 px-3.5 py-2 rounded-t-md text-xs font-semibold transition-colors border-b-2 -mb-px cursor-pointer ${
              activeTab === 'invitations'
                ? 'border-violet-500 text-zinc-100 bg-zinc-900/60'
                : 'border-transparent text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <Mail className="h-3.5 w-3.5 text-zinc-400" />
            <span>Pending Invitations</span>
            {pendingInvitations.length > 0 && (
              <span className="text-[10px] font-mono px-1.5 py-0.2 rounded-full bg-amber-500/20 text-amber-300 border border-amber-500/30 font-bold">
                {pendingInvitations.length}
              </span>
            )}
          </button>
        </div>

        {/* Role Distribution Chips (on Members Tab) */}
        {activeTab === 'members' && members.length > 0 && (
          <div className="hidden md:flex items-center gap-1.5 text-[11px] text-zinc-400">
            {Object.keys(roleCounts).map((roleKey) => (
              <span
                key={roleKey}
                className="px-2 py-0.5 rounded bg-zinc-900 border border-zinc-850 font-mono text-[10px]"
              >
                {roleKey}: <strong className="text-zinc-200">{roleCounts[roleKey]}</strong>
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Tab 1: Active Members */}
      {activeTab === 'members' && (
        <div className="space-y-4">
          {/* Member Search Bar */}
          <div className="flex items-center justify-between gap-3">
            <div className="relative flex-1 max-w-sm">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search members by name or email..."
                className="pl-8.5 pr-8 py-1.5 text-xs bg-zinc-950 border-zinc-800 text-zinc-200 placeholder:text-zinc-500 focus:border-zinc-700"
              />
              {searchQuery && (
                <button
                  type="button"
                  onClick={() => setSearchQuery('')}
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 transition-colors p-0.5"
                  title="Clear search"
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              )}
            </div>

            {searchQuery && (
              <span className="text-xs text-zinc-400">
                Found <strong className="text-zinc-200">{filteredMembers.length}</strong> matching members
              </span>
            )}
          </div>

          {/* Members Table / List */}
          <Card className="border-zinc-800 bg-zinc-900/80 divide-y divide-zinc-850 overflow-hidden">
            {isMembersLoading ? (
              <div className="flex h-40 items-center justify-center">
                <div className="flex flex-col items-center gap-3">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
                  <span className="text-xs text-zinc-500">Loading workspace directory...</span>
                </div>
              </div>
            ) : filteredMembers.length === 0 ? (
              <div className="py-12 text-center text-xs text-zinc-500 space-y-2">
                <Users className="w-8 h-8 text-zinc-600 mx-auto mb-2" />
                <p className="font-semibold text-zinc-300">No members match your search</p>
                <p className="text-zinc-500">Try adjusting your search terms</p>
              </div>
            ) : (
              filteredMembers.map((member) => {
                const roleStyle = getRoleBadgeStyle(member.role);
                const initials =
                  (
                    (member.user?.firstName?.[0] || '') +
                    (member.user?.lastName?.[0] || '')
                  ).toUpperCase() ||
                  member.user?.username?.[0]?.toUpperCase() ||
                  'U';

                const isCurrentUser = member.user?.id === user?.id;
                const memberDisplayName =
                  member.user?.firstName && member.user?.lastName
                    ? `${member.user.firstName} ${member.user.lastName}`
                    : member.user?.username || member.user?.email || 'Member';

                return (
                  <div
                    key={member.id}
                    className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 hover:bg-zinc-850/30 transition-colors text-xs"
                  >
                    {/* User Identity */}
                    <Link
                      href={`/${workspaceSlug}/members/${member.userId || member.user?.id}`}
                      className="flex items-center gap-3 group"
                    >
                      <Avatar className="h-9 w-9 rounded-md border border-zinc-700 bg-zinc-800 shrink-0 group-hover:border-zinc-500 transition-colors">
                        <AvatarFallback className="text-xs font-bold text-zinc-200">
                          {initials}
                        </AvatarFallback>
                      </Avatar>
                      <div className="space-y-0.5">
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-zinc-200 group-hover:text-white transition-colors text-xs">
                            {memberDisplayName}
                          </span>
                          {isCurrentUser && (
                            <Badge
                              variant="outline"
                              className="text-[9px] uppercase font-mono px-1.5 py-0 bg-zinc-850 text-zinc-300 border-zinc-750"
                            >
                              You
                            </Badge>
                          )}
                        </div>
                        <div className="flex items-center gap-1.5 text-[11px] text-zinc-500">
                          <Mail className="h-3 w-3" />
                          <span>{member.user?.email}</span>
                        </div>
                      </div>
                    </Link>

                    {/* Role Selector & Actions */}
                    <div className="flex items-center gap-3 self-end sm:self-center">
                      {canManage && !isCurrentUser ? (
                        <Select
                          value={member.role}
                          disabled={isUpdatingRole}
                          onValueChange={(val) =>
                            handleRoleChange(member.id, val, memberDisplayName)
                          }
                        >
                          <SelectTrigger className="w-28 h-7 bg-zinc-950 border-zinc-800 text-zinc-200 text-xs">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                            <SelectItem value={Role.MEMBER} className="text-xs">
                              Member
                            </SelectItem>
                            <SelectItem value={Role.MANAGER} className="text-xs">
                              Manager
                            </SelectItem>
                            <SelectItem value={Role.VIEWER} className="text-xs">
                              Viewer
                            </SelectItem>
                            <SelectItem value={Role.OWNER} className="text-xs">
                              Owner
                            </SelectItem>
                          </SelectContent>
                        </Select>
                      ) : (
                        <Badge
                          className={`text-[10px] uppercase font-mono px-2 py-0.5 ${roleStyle.className}`}
                        >
                          {roleStyle.label}
                        </Badge>
                      )}

                      {canManage && !isCurrentUser && (
                        <button
                          type="button"
                          onClick={() => setMemberToRemove({ id: member.id, name: memberDisplayName })}
                          className="p-1.5 rounded-md text-zinc-500 hover:text-rose-400 hover:bg-zinc-800 transition-colors cursor-pointer"
                          title="Remove member"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      )}
                    </div>
                  </div>
                );
              })
            )}
          </Card>
        </div>
      )}

      {/* Tab 2: Pending Invitations */}
      {activeTab === 'invitations' && workspace?.id && (
        <PendingInvitationsList
          workspaceId={workspace.id}
          canManage={canManage}
          onOpenInviteModal={() => setIsInviteOpen(true)}
        />
      )}

      {/* Invite Member Modal */}
      {workspace?.id && (
        <InviteMemberModal
          isOpen={isInviteOpen}
          onClose={() => setIsInviteOpen(false)}
          workspaceId={workspace.id}
        />
      )}

      {/* Remove Member Confirmation Modal */}
      <ConfirmationModal
        open={!!memberToRemove}
        onOpenChange={(open) => !open && setMemberToRemove(null)}
        title="Remove Workspace Member"
        description={
          <span>
            Are you sure you want to remove{' '}
            <strong className="text-zinc-200">{memberToRemove?.name}</strong> from this workspace?
            They will lose access to all projects, tasks, and sprints immediately.
          </span>
        }
        confirmLabel="Remove Member"
        variant="destructive"
        isLoading={isRemoving}
        onConfirm={handleConfirmRemoveMember}
      />
    </div>
  );
}
