'use client';

import * as React from 'react';
import { useState } from 'react';
import { useParams } from 'next/navigation';
import { Users, UserPlus, Mail, Shield, Trash2 } from 'lucide-react';
import { toast } from 'sonner';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import {
  useGetWorkspaceBySlugQuery,
  useGetWorkspaceMembersQuery,
  useRemoveWorkspaceMemberMutation,
} from '@/lib/store/api/workspacesApi';
import { Role, getRoleBadgeStyle } from '@/lib/rbac';
import { InviteMemberModal } from '@/features/workspaces';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

export default function WorkspaceMembersPage() {
  const params = useParams<{ workspaceSlug: string }>();
  const { workspaceSlug } = params;
  const { isManager, isOwner, user } = useAuthRbac();

  const [isInviteOpen, setIsInviteOpen] = useState(false);

  const { data: workspace } = useGetWorkspaceBySlugQuery(workspaceSlug || '', {
    skip: !workspaceSlug,
  });

  const { data: members = [], isLoading } = useGetWorkspaceMembersQuery(
    workspace?.id || '',
    { skip: !workspace?.id }
  );

  const [removeMember] = useRemoveWorkspaceMemberMutation();

  const handleRemoveMember = async (memberId: string, memberName: string) => {
    if (!workspace?.id) return;
    if (confirm(`Are you sure you want to remove ${memberName} from this workspace?`)) {
      try {
        await removeMember({ workspaceId: workspace.id, memberId }).unwrap();
        toast.success(`${memberName} removed from workspace`);
      } catch {
        toast.error('Failed to remove member');
      }
    }
  };

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5">
            <Users className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Workspace Members
            </h1>
            <Badge variant="outline" className="text-xs font-mono">
              {members.length} {members.length === 1 ? 'Member' : 'Members'}
            </Badge>
          </div>
          <p className="text-xs text-zinc-400">
            Collaborators with access to projects inside{' '}
            <strong className="text-zinc-200">{workspace?.name || workspaceSlug}</strong>.
          </p>
        </div>

        {(isManager || isOwner) && (
          <Button
            variant="primary"
            size="sm"
            className="gap-1.5 text-xs"
            onClick={() => setIsInviteOpen(true)}
          >
            <UserPlus className="h-3.5 w-3.5" />
            <span>Invite Member</span>
          </Button>
        )}
      </div>

      {/* Members List */}
      <Card className="border-zinc-800 bg-zinc-900/80 max-w-4xl divide-y divide-zinc-850">
        {isLoading ? (
          <div className="flex h-32 items-center justify-center">
            <div className="h-5 w-5 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
          </div>
        ) : members.length === 0 ? (
          <div className="py-8 text-center text-xs text-zinc-500">
            No members found.
          </div>
        ) : (
          members.map((member) => {
            const roleStyle = getRoleBadgeStyle(member.role);
            const initials = (
              (member.user?.firstName?.[0] || '') + (member.user?.lastName?.[0] || '')
            ).toUpperCase() || member.user?.username?.[0]?.toUpperCase() || 'U';

            const isCurrentUser = member.user?.id === user?.id;

            return (
              <div
                key={member.id}
                className="flex items-center justify-between p-4 hover:bg-zinc-850/30 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <Avatar className="h-9 w-9 rounded-md border border-zinc-700 bg-zinc-800">
                    <AvatarFallback className="text-xs font-bold text-zinc-200">
                      {initials}
                    </AvatarFallback>
                  </Avatar>
                  <div className="space-y-0.5">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-semibold text-zinc-200">
                        {member.user?.firstName && member.user?.lastName
                          ? `${member.user.firstName} ${member.user.lastName}`
                          : member.user?.username || 'Member'}
                      </span>
                      {isCurrentUser && (
                        <span className="text-[10px] text-zinc-500">(You)</span>
                      )}
                    </div>
                    <div className="flex items-center gap-1.5 text-[11px] text-zinc-500">
                      <Mail className="h-3 w-3" />
                      <span>{member.user?.email}</span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <Badge className={`text-[10px] uppercase font-mono ${roleStyle.className}`}>
                    {roleStyle.label}
                  </Badge>

                  {(isOwner || isManager) && !isCurrentUser && (
                    <button
                      type="button"
                      onClick={() =>
                        handleRemoveMember(
                          member.id,
                          member.user?.username || 'member'
                        )
                      }
                      className="p-1.5 rounded text-zinc-500 hover:text-rose-400 hover:bg-zinc-800 transition-colors cursor-pointer"
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

      {/* Invite Member Modal */}
      <InviteMemberModal
        isOpen={isInviteOpen}
        onClose={() => setIsInviteOpen(false)}
        workspaceId={workspace?.id || ''}
      />
    </div>
  );
}
