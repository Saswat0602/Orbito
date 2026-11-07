'use client';

import * as React from 'react';
import { useState } from 'react';
import { Mail, UserPlus } from 'lucide-react';
import { toast } from 'sonner';
import { Role } from '@/lib/rbac';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useAddWorkspaceMemberMutation } from '@/lib/store/api/workspacesApi';

interface InviteMemberModalProps {
  isOpen: boolean;
  onClose: () => void;
  workspaceId: string;
}

export function InviteMemberModal({
  isOpen,
  onClose,
  workspaceId,
}: InviteMemberModalProps) {
  const [addMember, { isLoading }] = useAddWorkspaceMemberMutation();

  const [email, setEmail] = useState('');
  const [role, setRole] = useState<Role>(Role.MEMBER);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim() || !workspaceId) return;

    try {
      await addMember({
        workspaceId,
        email: email.trim(),
        role,
      }).unwrap();

      toast.success(`Invitation sent to ${email}`);
      setEmail('');
      setRole(Role.MEMBER);
      onClose();
    } catch {
      toast.error('Failed to send invitation');
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Invite Workspace Member"
      description="Collaborate with team members across projects and sprints."
    >
      <form onSubmit={handleSubmit} className="space-y-4 pt-2">
        <div className="space-y-1.5">
          <Label htmlFor="invite-email" className="text-xs text-zinc-300">
            Email Address
          </Label>
          <div className="relative">
            <Mail className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-zinc-500" />
            <Input
              id="invite-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="colleague@company.com"
              required
              className="pl-8 bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
            />
          </div>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="invite-role" className="text-xs text-zinc-300">
            Access Role
          </Label>
          <select
            id="invite-role"
            value={role}
            onChange={(e) => setRole(e.target.value as Role)}
            className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 rounded-md px-3 py-2 text-xs outline-none focus:border-zinc-700"
          >
            <option value={Role.MEMBER}>Member - Read & edit tasks</option>
            <option value={Role.MANAGER}>Manager - Manage sprints & projects</option>
            <option value={Role.VIEWER}>Viewer - Read-only access</option>
            <option value={Role.OWNER}>Owner - Full administrative control</option>
          </select>
        </div>

        <div className="flex items-center justify-end gap-2 pt-2 border-t border-zinc-800">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onClose}
            className="text-xs"
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            size="sm"
            isLoading={isLoading}
            className="text-xs gap-1.5"
          >
            <UserPlus className="h-3.5 w-3.5" />
            <span>Send Invitation</span>
          </Button>
        </div>
      </form>
    </Modal>
  );
}

