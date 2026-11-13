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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
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
          <Select value={role} onValueChange={(val) => setRole(val as Role)}>
            <SelectTrigger
              id="invite-role"
              className="w-full bg-zinc-950 border-zinc-800 text-zinc-200 text-xs h-9"
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
              <SelectItem value={Role.MEMBER} className="text-xs">
                Member - Read & edit tasks
              </SelectItem>
              <SelectItem value={Role.MANAGER} className="text-xs">
                Manager - Manage sprints & projects
              </SelectItem>
              <SelectItem value={Role.VIEWER} className="text-xs">
                Viewer - Read-only access
              </SelectItem>
              <SelectItem value={Role.OWNER} className="text-xs">
                Owner - Full administrative control
              </SelectItem>
            </SelectContent>
          </Select>
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

