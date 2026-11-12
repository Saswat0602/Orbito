'use client';

import * as React from 'react';
import { useState } from 'react';
import { Mail, Shield, UserPlus, Loader2 } from 'lucide-react';
import { toast } from 'sonner';

import { useInviteProjectMemberMutation } from '@/lib/store/api/projectsApi';
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface ProjectInviteMemberModalProps {
  isOpen: boolean;
  onClose: () => void;
  projectId: string;
  projectName: string;
}

export function ProjectInviteMemberModal({
  isOpen,
  onClose,
  projectId,
  projectName,
}: ProjectInviteMemberModalProps) {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState<'MEMBER' | 'MANAGER' | 'VIEWER'>('MEMBER');

  const [inviteProjectMember, { isLoading }] = useInviteProjectMemberMutation();

  const handleClose = () => {
    setEmail('');
    setRole('MEMBER');
    onClose();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim() || !projectId) return;

    try {
      await inviteProjectMember({
        projectId,
        email: email.trim().toLowerCase(),
        role,
      }).unwrap();

      toast.success(`Invitation sent to ${email}`);
      handleClose();
    } catch (err: unknown) {
      let message = 'Failed to send project invitation';
      if (err && typeof err === 'object' && 'data' in err) {
        const data = (err as { data?: { message?: string } }).data;
        if (data?.message && typeof data.message === 'string') {
          message = data.message;
        }
      }
      toast.error(message);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => (!open ? handleClose() : null)}>
      <DialogContent className="sm:max-w-md bg-zinc-950 border-zinc-800 text-zinc-100 p-6">
        <DialogHeader className="space-y-1.5">
          <div className="flex items-center gap-2">
            <UserPlus className="h-4 w-4 text-zinc-400" />
            <DialogTitle className="text-base font-semibold text-zinc-100">
              Invite to {projectName}
            </DialogTitle>
          </div>
          <DialogDescription className="text-xs text-zinc-400">
            Send an invitation to assign tasks, sprints, and project boards to a teammate.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          {/* Email input */}
          <div className="space-y-1.5">
            <Label htmlFor="project-invite-email" className="text-xs text-zinc-300">
              Email Address
            </Label>
            <div className="relative">
              <Mail className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
              <Input
                id="project-invite-email"
                type="email"
                required
                placeholder="colleague@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="pl-8 text-xs bg-zinc-900 border-zinc-800 text-zinc-100 focus:border-zinc-600 h-9"
              />
            </div>
          </div>

          {/* Role select */}
          <div className="space-y-1.5">
            <Label htmlFor="project-invite-role" className="text-xs text-zinc-300">
              Project Role
            </Label>
            <Select
              value={role}
              onValueChange={(val: 'MEMBER' | 'MANAGER' | 'VIEWER') => setRole(val)}
            >
              <SelectTrigger
                id="project-invite-role"
                className="text-xs bg-zinc-900 border-zinc-800 text-zinc-100 h-9"
              >
                <div className="flex items-center gap-2">
                  <Shield className="h-3.5 w-3.5 text-zinc-400" />
                  <SelectValue />
                </div>
              </SelectTrigger>
              <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                <SelectItem value="MEMBER" className="text-xs cursor-pointer">
                  <div className="space-y-0.5">
                    <span className="font-medium text-zinc-200">Member</span>
                    <p className="text-[11px] text-zinc-500">Can view, create, and update assigned tasks.</p>
                  </div>
                </SelectItem>
                <SelectItem value="MANAGER" className="text-xs cursor-pointer">
                  <div className="space-y-0.5">
                    <span className="font-medium text-zinc-200">Manager</span>
                    <p className="text-[11px] text-zinc-500">Can manage sprints, invite members, and configure settings.</p>
                  </div>
                </SelectItem>
                <SelectItem value="VIEWER" className="text-xs cursor-pointer">
                  <div className="space-y-0.5">
                    <span className="font-medium text-zinc-200">Viewer</span>
                    <p className="text-[11px] text-zinc-500">Read-only view of boards and sprint progress.</p>
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <DialogFooter className="pt-3 gap-2 sm:gap-0">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleClose}
              disabled={isLoading}
              className="text-xs border-zinc-800 bg-zinc-900 text-zinc-300 hover:bg-zinc-850"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="sm"
              disabled={isLoading || !email.trim()}
              className="text-xs gap-1.5"
            >
              {isLoading && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
              <span>Send Invitation</span>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

