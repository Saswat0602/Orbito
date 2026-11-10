'use client';

import * as React from 'react';
import { useState } from 'react';
import { UserPlus, Copy, Check, Mail, Shield } from 'lucide-react';
import { toast } from 'sonner';

import { useCreateInvitationMutation } from '@/lib/store/api/invitationsApi';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

interface InviteMemberModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  workspaceId?: string | null;
  workspaceName?: string;
  organizationId?: string | null;
}

export function InviteMemberModal({
  open,
  onOpenChange,
  workspaceId,
  workspaceName = 'Current Workspace',
  organizationId,
}: InviteMemberModalProps) {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('MEMBER');
  const [inviteUrl, setInviteUrl] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const [createInvitation, { isLoading }] = useCreateInvitationMutation();

  const handleSendInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim()) {
      toast.error('Please enter a valid email address');
      return;
    }

    try {
      const res = await createInvitation({
        inviteeEmail: email.trim().toLowerCase(),
        role,
        workspaceId: workspaceId || undefined,
        organizationId: organizationId || undefined,
      }).unwrap();

      const origin = typeof window !== 'undefined' ? window.location.origin : '';
      const generatedLink = `${origin}/invite/${res.token}`;
      setInviteUrl(generatedLink);
      toast.success(`Invitation generated for ${email}`);
    } catch (err: unknown) {
      const msg =
        err && typeof err === 'object' && 'data' in err
          ? (err as { data: { message?: string } }).data?.message
          : 'Failed to send invitation';
      toast.error(msg || 'Failed to send invitation');
    }
  };

  const handleCopyLink = () => {
    if (!inviteUrl) return;
    navigator.clipboard.writeText(inviteUrl);
    setCopied(true);
    toast.success('Invitation link copied to clipboard!');
    setTimeout(() => setCopied(false), 2000);
  };

  const handleClose = () => {
    setEmail('');
    setRole('MEMBER');
    setInviteUrl(null);
    setCopied(false);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md bg-zinc-900 border-zinc-800 text-zinc-100 p-6 rounded-md">
        <DialogHeader className="space-y-1.5 pb-2">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-md bg-zinc-800 border border-zinc-700/80 flex items-center justify-center text-zinc-200">
              <UserPlus className="h-4 w-4" />
            </div>
            <DialogTitle className="text-base font-semibold text-zinc-100">
              Invite to {workspaceName}
            </DialogTitle>
          </div>
          <DialogDescription className="text-xs text-zinc-400">
            Invite colleagues to collaborate on projects and tasks in this workspace.
          </DialogDescription>
        </DialogHeader>

        {!inviteUrl ? (
          <form onSubmit={handleSendInvite} className="space-y-4 pt-2">
            <div className="space-y-1.5">
              <Label htmlFor="inviteEmail" className="text-xs text-zinc-300">
                Email Address
              </Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
                <Input
                  id="inviteEmail"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="colleague@company.com"
                  required
                  className="pl-9 h-9 text-xs bg-zinc-950 border-zinc-800 text-zinc-100 placeholder:text-zinc-500"
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="inviteRole" className="text-xs text-zinc-300">
                Workspace Role
              </Label>
              <select
                id="inviteRole"
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="w-full h-9 px-3 rounded-md bg-zinc-950 border border-zinc-800 text-xs text-zinc-200 focus:outline-none focus:ring-1 focus:ring-zinc-600"
              >
                <option value="MEMBER">Member (Can create & edit tasks)</option>
                <option value="MANAGER">Manager (Can manage projects & settings)</option>
                <option value="VIEWER">Viewer (Read-only access)</option>
              </select>
            </div>

            <DialogFooter className="pt-2 flex sm:justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleClose}
                className="h-8 text-xs border-zinc-800 text-zinc-300"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                size="sm"
                disabled={isLoading}
                className="h-8 text-xs bg-zinc-100 text-zinc-950 font-semibold hover:bg-white"
              >
                {isLoading ? 'Generating Invite...' : 'Generate Invite'}
              </Button>
            </DialogFooter>
          </form>
        ) : (
          <div className="space-y-4 pt-2">
            <div className="p-3 rounded-md bg-emerald-950/30 border border-emerald-800/50 text-xs text-emerald-400 space-y-1">
              <span className="font-semibold">✓ Invitation Ready</span>
              <p className="text-[11px] text-zinc-400">
                Share this secure link with <strong>{email}</strong> to onboard them as a <strong>{role}</strong>.
              </p>
            </div>

            <div className="flex items-center gap-2">
              <Input
                readOnly
                value={inviteUrl}
                className="h-9 text-xs bg-zinc-950 border-zinc-800 text-zinc-300 font-mono select-all"
              />
              <Button
                type="button"
                size="sm"
                onClick={handleCopyLink}
                className="h-9 px-3 text-xs bg-zinc-800 hover:bg-zinc-700 text-zinc-100 shrink-0"
              >
                {copied ? <Check className="h-3.5 w-3.5 text-emerald-400 mr-1" /> : <Copy className="h-3.5 w-3.5 mr-1" />}
                {copied ? 'Copied' : 'Copy'}
              </Button>
            </div>

            <DialogFooter className="pt-2">
              <Button
                type="button"
                size="sm"
                onClick={handleClose}
                className="w-full h-8 text-xs bg-zinc-100 text-zinc-950 font-semibold hover:bg-white"
              >
                Done
              </Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

