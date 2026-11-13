'use client';

import * as React from 'react';
import { useState, useMemo, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import {
  Users,
  Search,
  UserPlus,
  Shield,
  Trash2,
  AlertCircle,
  FolderGit2,
  Calendar,
  Check,
  ArrowLeft,
  Mail,
} from 'lucide-react';
import { toast } from 'sonner';

import type { ProjectMember, Role } from '@orbito/types';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import {
  useGetProjectBySlugQuery,
  useGetProjectMembersQuery,
  useUpdateProjectMemberRoleMutation,
  useRemoveProjectMemberMutation,
} from '@/lib/store/api/projectsApi';
import { ProjectInviteMemberModal } from '@/features/projects';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { MembersSkeleton } from '@/components/skeletons';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';

export default function DedicatedProjectMembersPage() {
  const params = useParams<{ workspaceSlug: string; projectSlug: string }>();
  const { workspaceSlug, projectSlug } = params;
  const { user: currentUser, isManager, isOwner } = useAuthRbac();
  const canManage = isManager || isOwner;

  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);
  const [memberToRemove, setMemberToRemove] = useState<ProjectMember | null>(null);

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(searchQuery), 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Query Project
  const { data: project, isLoading: isProjectLoading } = useGetProjectBySlugQuery(
    projectSlug || '',
    { skip: !projectSlug }
  );

  // Query Project Members
  const { data: members = [], isLoading: isMembersLoading } = useGetProjectMembersQuery(
    project?.id ? { projectId: project.id, search: debouncedSearch || undefined } : { projectId: '' },
    { skip: !project?.id }
  );

  const [updateRole, { isLoading: isUpdatingRole }] = useUpdateProjectMemberRoleMutation();
  const [removeMember, { isLoading: isRemovingMember }] = useRemoveProjectMemberMutation();

  const handleRoleChange = async (memberId: string, newRole: string) => {
    if (!project?.id) return;
    try {
      await updateRole({
        id: memberId,
        role: newRole,
        projectId: project.id,
      }).unwrap();
      toast.success('Member role updated successfully');
    } catch {
      toast.error('Failed to update member role');
    }
  };

  const handleConfirmRemove = async () => {
    if (!memberToRemove || !project?.id) return;
    try {
      await removeMember({
        id: memberToRemove.id,
        projectId: project.id,
      }).unwrap();
      toast.success('Member removed from project');
      setMemberToRemove(null);
    } catch {
      toast.error('Failed to remove member');
    }
  };

  const getRoleBadgeStyle = (role: string) => {
    switch (role) {
      case 'OWNER':
      case 'MANAGER':
        return 'border-purple-800 text-purple-400 bg-purple-950/40';
      case 'MEMBER':
        return 'border-zinc-700 text-zinc-300 bg-zinc-900';
      case 'VIEWER':
        return 'border-sky-900 text-sky-400 bg-sky-950/40';
      default:
        return 'border-zinc-800 text-zinc-400 bg-zinc-900';
    }
  };

  if (isProjectLoading) {
    return <MembersSkeleton />;
  }

  if (!project) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center bg-[#09090b] p-6 text-center space-y-3">
        <AlertCircle className="h-8 w-8 text-zinc-600" />
        <h2 className="text-sm font-semibold text-zinc-300">Project Not Found</h2>
        <Link href={`/${workspaceSlug}/projects`}>
          <Button variant="outline" size="sm" className="text-xs border-zinc-800">
            Back to Projects
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#09090b] p-6 space-y-5 overflow-hidden">
      {/* Header & Breadcrumb */}
      <div className="space-y-3 pb-4 border-b border-zinc-800/80">
        <div className="flex items-center gap-2 text-xs text-zinc-400">
          <Link
            href={`/${workspaceSlug}/${projectSlug}/tasks`}
            className="flex items-center gap-1.5 hover:text-zinc-200 transition"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            <span>{project.name} Tasks</span>
          </Link>
          <span className="text-zinc-600">/</span>
          <span className="text-zinc-300 font-medium">Project Members</span>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2.5">
              <Users className="h-5 w-5 text-zinc-300" />
              <h1 className="text-xl font-bold tracking-tight text-zinc-100">
                Project Members & Permissions
              </h1>
              <Badge variant="outline" className="text-[11px] font-mono border-zinc-800 text-zinc-400">
                {members.length} {members.length === 1 ? 'member' : 'members'}
              </Badge>
            </div>
            <p className="text-xs text-zinc-400">
              Manage member assignments, access roles, and permissions for {project.name}.
            </p>
          </div>

          {canManage && (
            <Button
              variant="primary"
              size="sm"
              onClick={() => setIsInviteModalOpen(true)}
              className="gap-1.5 text-xs h-8 cursor-pointer shrink-0"
            >
              <UserPlus className="h-3.5 w-3.5" />
              <span>Invite Member</span>
            </Button>
          )}
        </div>
      </div>

      {/* Toolbar: Search input */}
      <div className="flex items-center justify-between gap-4 shrink-0">
        <div className="relative w-full sm:w-72">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
          <Input
            placeholder="Search by name or email..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-8 h-8 text-xs bg-zinc-900 border-zinc-800 text-zinc-200"
          />
        </div>
      </div>

      {/* Members Table */}
      <div className="flex-1 min-h-0 overflow-auto rounded-md border border-zinc-800 bg-zinc-950/60">
        {isMembersLoading ? (
          <div className="p-4 space-y-3">
            {Array.from({ length: 6 }).map((_, idx) => (
              <div key={idx} className="flex items-center justify-between py-2.5 px-3 border-b border-zinc-850/60 last:border-b-0">
                <div className="flex items-center gap-3">
                  <Skeleton className="h-7 w-7 rounded-full" />
                  <div className="space-y-1">
                    <Skeleton className="h-3.5 w-32" />
                    <Skeleton className="h-3 w-20" />
                  </div>
                </div>
                <div className="flex items-center gap-6">
                  <Skeleton className="h-3.5 w-36 hidden sm:block" />
                  <Skeleton className="h-5 w-20 rounded" />
                  <Skeleton className="h-3.5 w-24 hidden md:block" />
                  <Skeleton className="h-7 w-16 rounded" />
                </div>
              </div>
            ))}
          </div>
        ) : members.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-12 text-center space-y-2">
            <Users className="h-8 w-8 text-zinc-600" />
            <p className="text-xs font-medium text-zinc-400">No project members found</p>
            <p className="text-[11px] text-zinc-500">
              {searchQuery ? 'Try adjusting your search terms.' : 'Invite team members to collaborate on this project.'}
            </p>
          </div>
        ) : (
          <table className="w-full text-left text-xs">
            <thead className="sticky top-0 bg-zinc-900/90 backdrop-blur-sm border-b border-zinc-800 text-zinc-400 font-mono text-[11px]">
              <tr>
                <th className="py-2.5 px-4 font-normal">Member</th>
                <th className="py-2.5 px-4 font-normal">Email</th>
                <th className="py-2.5 px-4 font-normal">Project Role</th>
                <th className="py-2.5 px-4 font-normal">Joined</th>
                {canManage && <th className="py-2.5 px-4 font-normal text-right">Actions</th>}
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-850">
              {members.map((member) => {
                const displayName =
                  member.user
                    ? `${member.user.firstName || ''} ${member.user.lastName || ''}`.trim() || member.user.username || 'User'
                    : 'Teammate';
                const initials = displayName.slice(0, 2).toUpperCase() || 'U';
                const isSelf = currentUser?.id === member.userId;

                return (
                  <tr key={member.id} className="hover:bg-zinc-900/40 transition">
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-3">
                        <Avatar className="h-8 w-8 rounded-md border border-zinc-700 bg-zinc-850">
                          {member.user?.avatar ? (
                            <AvatarImage src={member.user.avatar} alt={displayName} />
                          ) : null}
                          <AvatarFallback className="text-[11px] font-bold bg-zinc-800 text-zinc-300">
                            {initials}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="flex items-center gap-1.5">
                            <span className="font-medium text-zinc-200">{displayName}</span>
                            {isSelf && (
                              <span className="text-[10px] text-zinc-500 font-mono">(You)</span>
                            )}
                          </div>
                          {member.user?.username && (
                            <span className="text-[11px] text-zinc-500 font-mono block">
                              @{member.user.username}
                            </span>
                          )}
                        </div>
                      </div>
                    </td>

                    <td className="py-3 px-4 text-zinc-400 text-xs">
                      {member.user?.email || '-'}
                    </td>

                    <td className="py-3 px-4">
                      {canManage && !isSelf ? (
                        <Select
                          value={member.role}
                          onValueChange={(newRole) => handleRoleChange(member.id, newRole)}
                          disabled={isUpdatingRole}
                        >
                          <SelectTrigger className="w-32 h-7 text-xs bg-zinc-900 border-zinc-800 text-zinc-200">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent className="bg-zinc-900 border-zinc-800 text-xs text-zinc-200">
                            <SelectItem value="MEMBER" className="cursor-pointer">
                              Member
                            </SelectItem>
                            <SelectItem value="MANAGER" className="cursor-pointer">
                              Manager
                            </SelectItem>
                            <SelectItem value="VIEWER" className="cursor-pointer">
                              Viewer
                            </SelectItem>
                          </SelectContent>
                        </Select>
                      ) : (
                        <Badge
                          variant="outline"
                          className={`text-[10px] font-medium uppercase ${getRoleBadgeStyle(member.role)}`}
                        >
                          {member.role}
                        </Badge>
                      )}
                    </td>

                    <td className="py-3 px-4 text-zinc-500 text-[11px] whitespace-nowrap">
                      {member.joinedAt ? new Date(member.joinedAt).toLocaleDateString() : '-'}
                    </td>

                    {canManage && (
                      <td className="py-3 px-4 text-right">
                        {!isSelf && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setMemberToRemove(member)}
                            className="h-7 w-7 p-0 text-zinc-500 hover:text-rose-400 hover:bg-rose-950/20"
                            title="Remove from project"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        )}
                      </td>
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* Invite Member Modal */}
      <ProjectInviteMemberModal
        isOpen={isInviteModalOpen}
        onClose={() => setIsInviteModalOpen(false)}
        projectId={project.id}
        projectName={project.name}
      />

      {/* Remove Confirmation Dialog */}
      <Dialog open={!!memberToRemove} onOpenChange={(open) => (!open ? setMemberToRemove(null) : null)}>
        <DialogContent className="sm:max-w-sm bg-zinc-950 border-zinc-800 text-zinc-100 p-6">
          <DialogHeader className="space-y-2">
            <DialogTitle className="text-sm font-semibold text-zinc-100">
              Remove Member from Project
            </DialogTitle>
            <DialogDescription className="text-xs text-zinc-400">
              Are you sure you want to remove this member from {project.name}? They will lose access to project boards, tasks, and sprints.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="pt-3 gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setMemberToRemove(null)}
              className="text-xs border-zinc-800 bg-zinc-900 text-zinc-300"
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={handleConfirmRemove}
              disabled={isRemovingMember}
              className="text-xs bg-rose-600 hover:bg-rose-500 text-white"
            >
              Remove
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

