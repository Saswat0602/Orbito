'use client';

import * as React from 'react';
import { useState } from 'react';
import {
  Mail,
  Copy,
  Check,
  RotateCw,
  Trash2,
  Clock,
  Shield,
  AlertCircle,
  UserPlus,
} from 'lucide-react';
import { toast } from 'sonner';
import {
  useGetEntityInvitationsQuery,
  useResendInvitationMutation,
  useDeleteInvitationMutation,
} from '@/lib/store/api/invitationsApi';
import { getRoleBadgeStyle } from '@/lib/rbac';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';

interface PendingInvitationsListProps {
  workspaceId: string;
  canManage: boolean;
  onOpenInviteModal?: () => void;
}

export function PendingInvitationsList({
  workspaceId,
  canManage,
  onOpenInviteModal,
}: PendingInvitationsListProps) {
  const [copiedTokenId, setCopiedTokenId] = useState<string | null>(null);

  const {
    data: invitations = [],
    isLoading,
    isError,
    refetch,
  } = useGetEntityInvitationsQuery(
    { entityType: 'workspace', entityId: workspaceId },
    { skip: !workspaceId }
  );

  const [resendInvitation, { isLoading: isResending }] =
    useResendInvitationMutation();
  const [deleteInvitation, { isLoading: isDeleting }] =
    useDeleteInvitationMutation();

  const handleCopyLink = async (token: string, id: string) => {
    const origin =
      typeof window !== 'undefined'
        ? window.location.origin
        : 'http://localhost:3001';
    const link = `${origin}/invite/${token}`;

    try {
      await navigator.clipboard.writeText(link);
      setCopiedTokenId(id);
      toast.success('Invitation link copied to clipboard');
      setTimeout(() => setCopiedTokenId(null), 2500);
    } catch {
      toast.error('Failed to copy link');
    }
  };

  const handleResend = async (id: string, email: string) => {
    try {
      await resendInvitation(id).unwrap();
      toast.success(`Invitation resent to ${email}`);
      refetch();
    } catch {
      toast.error('Failed to resend invitation');
    }
  };

  const handleRevoke = async (id: string, email: string) => {
    if (confirm(`Revoke invitation for ${email}? The invite link will become invalid immediately.`)) {
      try {
        await deleteInvitation(id).unwrap();
        toast.success(`Invitation for ${email} revoked`);
        refetch();
      } catch {
        toast.error('Failed to revoke invitation');
      }
    }
  };

  if (isLoading) {
    return (
      <div className="flex h-40 items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
          <span className="text-xs text-zinc-500">Checking pending invitations...</span>
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-8 text-center text-xs text-rose-400 bg-rose-950/20 border border-rose-900/40 rounded-md">
        Failed to load pending invitations. Please try again.
      </div>
    );
  }

  // Filter pending/active invitations
  const pendingInvites = invitations.filter(
    (inv) => inv.status === 'PENDING' || !inv.status
  );

  if (pendingInvites.length === 0) {
    return (
      <Card className="border-dashed border-zinc-800 bg-zinc-950/40 p-12 text-center max-w-lg mx-auto space-y-4">
        <div className="w-12 h-12 rounded-full bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-500 mx-auto">
          <Mail className="h-6 w-6" />
        </div>
        <div>
          <h3 className="text-sm font-semibold text-zinc-200">
            No pending invitations
          </h3>
          <p className="text-xs text-zinc-400 mt-1">
            All invited teammates have accepted or no active invitations are pending.
          </p>
        </div>
        {canManage && onOpenInviteModal && (
          <Button
            variant="outline"
            size="sm"
            onClick={onOpenInviteModal}
            className="text-xs gap-1.5"
          >
            <UserPlus className="w-3.5 h-3.5" />
            <span>Invite Team Member</span>
          </Button>
        )}
      </Card>
    );
  }

  return (
    <Card className="border-zinc-800 bg-zinc-900/80 divide-y divide-zinc-850 overflow-hidden">
      {pendingInvites.map((inv) => {
        const roleStyle = getRoleBadgeStyle(inv.role);
        const expiresDate = inv.expiresAt ? new Date(inv.expiresAt) : null;
        const isExpired = expiresDate ? expiresDate.getTime() < Date.now() : false;

        return (
          <div
            key={inv.id}
            className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 hover:bg-zinc-850/30 transition-colors text-xs"
          >
            {/* Invitee Info */}
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 rounded-md bg-zinc-950 border border-zinc-800 flex items-center justify-center text-zinc-400 shrink-0">
                <Mail className="w-4 h-4" />
              </div>
              <div className="space-y-0.5">
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-zinc-200 text-xs">
                    {inv.inviteeEmail}
                  </span>
                  <Badge
                    variant="outline"
                    className={`text-[9px] uppercase font-mono px-1.5 py-0 ${roleStyle.className}`}
                  >
                    {roleStyle.label}
                  </Badge>
                  {isExpired && (
                    <Badge
                      variant="outline"
                      className="text-[9px] uppercase font-mono px-1.5 py-0 bg-rose-950/40 text-rose-300 border-rose-800/60"
                    >
                      Expired
                    </Badge>
                  )}
                </div>
                <div className="flex items-center gap-2 text-[11px] text-zinc-500">
                  <Clock className="w-3 h-3 text-zinc-500" />
                  <span>
                    {expiresDate
                      ? isExpired
                        ? `Expired on ${expiresDate.toLocaleDateString()}`
                        : `Expires ${expiresDate.toLocaleDateString()}`
                      : 'Never expires'}
                  </span>
                </div>
              </div>
            </div>

            {/* Actions: Copy Link, Resend, Revoke */}
            {canManage && (
              <div className="flex items-center gap-2 self-end sm:self-center">
                {/* 1-Click Copy Link */}
                <button
                  type="button"
                  onClick={() => handleCopyLink(inv.token, inv.id)}
                  className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md bg-zinc-950 border border-zinc-800 text-zinc-300 hover:text-white hover:border-zinc-700 transition-colors text-xs cursor-pointer"
                  title="Copy direct invite link"
                >
                  {copiedTokenId === inv.id ? (
                    <>
                      <Check className="w-3.5 h-3.5 text-emerald-400" />
                      <span className="text-emerald-400 font-medium">Copied</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5 text-zinc-400" />
                      <span>Copy Link</span>
                    </>
                  )}
                </button>

                {/* Resend Invite */}
                <button
                  type="button"
                  disabled={isResending}
                  onClick={() => handleResend(inv.id, inv.inviteeEmail)}
                  className="p-1.5 rounded-md bg-zinc-950 border border-zinc-800 text-zinc-400 hover:text-zinc-200 hover:border-zinc-700 transition-colors cursor-pointer disabled:opacity-50"
                  title="Resend invitation email"
                >
                  <RotateCw className="w-3.5 h-3.5" />
                </button>

                {/* Revoke Invite */}
                <button
                  type="button"
                  disabled={isDeleting}
                  onClick={() => handleRevoke(inv.id, inv.inviteeEmail)}
                  className="p-1.5 rounded-md bg-zinc-950 border border-zinc-800 text-zinc-500 hover:text-rose-400 hover:border-rose-900 transition-colors cursor-pointer disabled:opacity-50"
                  title="Revoke invitation"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
            )}
          </div>
        );
      })}
    </Card>
  );
}

