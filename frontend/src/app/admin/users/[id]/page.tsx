'use client';

import * as React from 'react';
import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  ArrowLeft,
  Users,
  Shield,
  UserCheck,
  UserX,
  KeyRound,
  Trash2,
  Calendar,
  Mail,
  Phone,
  Briefcase,
  Clock,
  CheckCircle2,
  AlertTriangle,
  Loader2,
} from 'lucide-react';
import { toast } from 'sonner';

import { Role, UserStatus } from '@orbito/types';
import { RoleGuard } from '@/components/auth/RoleGuard';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useGetUserByIdQuery } from '@/lib/store/api/usersApi';
import {
  useUpdateAdminUserRoleMutation,
  useUpdateAdminUserStatusMutation,
  useResetAdminUserPasswordMutation,
  useDeleteAdminUserMutation,
} from '@/lib/store/api/adminApi';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { AdminDetailSkeleton } from '@/components/skeletons';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
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

export default function AdminUserDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const userId = params.id;
  const { user: currentUser } = useAuthRbac();

  const { data: user, isLoading, refetch } = useGetUserByIdQuery(userId, {
    skip: !userId,
  });

  const [updateUserRole, { isLoading: isUpdatingRole }] = useUpdateAdminUserRoleMutation();
  const [updateUserStatus, { isLoading: isUpdatingStatus }] = useUpdateAdminUserStatusMutation();
  const [resetPassword, { isLoading: isResettingPassword }] = useResetAdminUserPasswordMutation();
  const [deleteUser, { isLoading: isDeleting }] = useDeleteAdminUserMutation();

  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);

  const isSelf = currentUser?.id === userId;

  const handleRoleChange = async (newRole: Role) => {
    if (!userId) return;
    try {
      await updateUserRole({ id: userId, role: newRole }).unwrap();
      toast.success(`User role updated to ${newRole}`);
      refetch();
    } catch {
      toast.error('Failed to update user role');
    }
  };

  const handleStatusToggle = async () => {
    if (!userId || !user) return;
    const newStatus = user.status === UserStatus.ACTIVE ? UserStatus.INACTIVE : UserStatus.ACTIVE;
    try {
      await updateUserStatus({ id: userId, status: newStatus }).unwrap();
      toast.success(`User status updated to ${newStatus}`);
      refetch();
    } catch {
      toast.error('Failed to update user status');
    }
  };

  const handleResetPassword = async () => {
    if (!userId) return;
    try {
      const res = await resetPassword(userId).unwrap();
      toast.success(res.message || 'Password reset link sent to user email');
    } catch {
      toast.error('Failed to trigger password reset');
    }
  };

  const handleDeleteUser = async () => {
    if (!userId) return;
    try {
      await deleteUser(userId).unwrap();
      toast.success('User deleted successfully');
      router.push('/admin/users');
    } catch {
      toast.error('Failed to delete user');
    }
  };

  if (isLoading) {
    return (
      <RoleGuard requiredRole={Role.SUPER_ADMIN}>
        <AdminDetailSkeleton />
      </RoleGuard>
    );
  }

  if (!user) {
    return (
      <RoleGuard requiredRole={Role.SUPER_ADMIN}>
        <div className="flex-1 flex flex-col items-center justify-center bg-[#09090b] p-6 text-center space-y-3">
          <AlertTriangle className="h-8 w-8 text-amber-500" />
          <h2 className="text-sm font-semibold text-zinc-300">User Not Found</h2>
          <p className="text-xs text-zinc-500">The user record does not exist or has been removed.</p>
          <Link href="/admin/users">
            <Button variant="outline" size="sm" className="text-xs border-zinc-800">
              Back to User Directory
            </Button>
          </Link>
        </div>
      </RoleGuard>
    );
  }

  const displayName = `${user.firstName || ''} ${user.lastName || ''}`.trim() || user.username || 'User';
  const initials = displayName.slice(0, 2).toUpperCase() || 'U';
  const userPrefs = (user.preferences as Record<string, unknown> | undefined) || {};

  return (
    <RoleGuard requiredRole={Role.SUPER_ADMIN}>
      <div className="flex-1 flex flex-col min-h-0 bg-[#09090b] p-6 space-y-6 overflow-y-auto">
        {/* Breadcrumb & Navigation */}
        <div className="flex items-center gap-2 text-xs text-zinc-400">
          <Link href="/admin/users" className="flex items-center gap-1.5 hover:text-zinc-200 transition">
            <ArrowLeft className="h-3.5 w-3.5" />
            <span>User Directory</span>
          </Link>
          <span className="text-zinc-600">/</span>
          <span className="text-zinc-300 font-medium truncate">{displayName}</span>
        </div>

        {/* User Profile Header Card */}
        <Card className="border-zinc-800 bg-zinc-900/60 p-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
            <div className="flex items-center gap-4">
              <Avatar className="h-16 w-16 rounded-md border-2 border-zinc-700 bg-zinc-850">
                {user.avatar ? <AvatarImage src={user.avatar} alt={displayName} /> : null}
                <AvatarFallback className="text-base font-bold bg-zinc-800 text-zinc-200">
                  {initials}
                </AvatarFallback>
              </Avatar>

              <div className="space-y-1">
                <div className="flex items-center gap-2.5 flex-wrap">
                  <h1 className="text-xl font-bold tracking-tight text-zinc-100">{displayName}</h1>
                  <Badge
                    variant="outline"
                    className="text-[10px] font-mono border-zinc-750 text-zinc-300 bg-zinc-850/80"
                  >
                    {user.role}
                  </Badge>
                  <Badge
                    variant="outline"
                    className="text-[10px] font-mono border-zinc-800 text-zinc-300 bg-zinc-900/80 flex items-center gap-1.5"
                  >
                    <span
                      className={`h-1.5 w-1.5 rounded-full shrink-0 ${
                        user.status === UserStatus.ACTIVE ? 'bg-emerald-400' : 'bg-rose-400'
                      }`}
                    />
                    {user.status}
                  </Badge>
                </div>
                <p className="text-xs text-zinc-400 font-mono">{user.email}</p>
                {user.username && (
                  <span className="text-[11px] text-zinc-500 font-mono block">@{user.username}</span>
                )}
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-center gap-2 flex-wrap">
              {!isSelf && (
                <>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleStatusToggle}
                    disabled={isUpdatingStatus}
                    className="gap-1.5 text-xs h-8 border-zinc-800 bg-zinc-900 text-zinc-300 hover:bg-zinc-800"
                  >
                    {user.status === UserStatus.ACTIVE ? (
                      <>
                        <UserX className="h-3.5 w-3.5 text-rose-400" />
                        <span>Deactivate User</span>
                      </>
                    ) : (
                      <>
                        <UserCheck className="h-3.5 w-3.5 text-emerald-400" />
                        <span>Activate User</span>
                      </>
                    )}
                  </Button>

                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleResetPassword}
                    disabled={isResettingPassword}
                    className="gap-1.5 text-xs h-8 border-zinc-800 bg-zinc-900 text-zinc-300 hover:bg-zinc-800"
                  >
                    <KeyRound className="h-3.5 w-3.5 text-amber-400" />
                    <span>Send Reset Email</span>
                  </Button>

                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => setIsDeleteModalOpen(true)}
                    className="gap-1.5 text-xs h-8 bg-rose-600 hover:bg-rose-500 text-white"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                    <span>Delete</span>
                  </Button>
                </>
              )}
            </div>
          </div>
        </Card>

        {/* Details Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Identity & Account Settings */}
          <Card className="border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
            <h3 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
              <Shield className="h-4 w-4 text-purple-400" />
              <span>Security & Privileges</span>
            </h3>

            <div className="space-y-3 divide-y divide-zinc-850 text-xs">
              <div className="flex items-center justify-between pt-2">
                <span className="text-zinc-400">Global System Role</span>
                {!isSelf ? (
                  <Select
                    value={user.role}
                    onValueChange={(val: Role) => handleRoleChange(val)}
                    disabled={isUpdatingRole}
                  >
                    <SelectTrigger className="w-36 h-7 text-xs bg-zinc-900 border-zinc-800 text-zinc-200">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent className="bg-zinc-900 border-zinc-800 text-xs text-zinc-200">
                      <SelectItem value={Role.MEMBER}>MEMBER</SelectItem>
                      <SelectItem value={Role.SUPER_ADMIN}>SUPER_ADMIN</SelectItem>
                    </SelectContent>
                  </Select>
                ) : (
                  <span className="font-mono text-zinc-300">{user.role} (Locked)</span>
                )}
              </div>

              <div className="flex items-center justify-between pt-3">
                <span className="text-zinc-400">Account Status</span>
                <span className="font-mono text-zinc-300">{user.status}</span>
              </div>

              <div className="flex items-center justify-between pt-3">
                <span className="text-zinc-400">Account ID</span>
                <span className="font-mono text-[11px] text-zinc-500">{user.id}</span>
              </div>
            </div>
          </Card>

          {/* Profile & Metadata */}
          <Card className="border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
            <h3 className="text-sm font-semibold text-zinc-100 flex items-center gap-2">
              <Briefcase className="h-4 w-4 text-sky-400" />
              <span>Profile & Department</span>
            </h3>

            <div className="space-y-3 divide-y divide-zinc-850 text-xs">
              <div className="flex items-center justify-between pt-2">
                <span className="text-zinc-400">Job Title</span>
                <span className="text-zinc-200 font-medium">
                  {typeof userPrefs.jobTitle === 'string' ? userPrefs.jobTitle : 'Not specified'}
                </span>
              </div>

              <div className="flex items-center justify-between pt-3">
                <span className="text-zinc-400">Department / Squad</span>
                <span className="text-zinc-200 font-medium">
                  {typeof userPrefs.department === 'string' ? userPrefs.department : 'Not specified'}
                </span>
              </div>

              <div className="flex items-center justify-between pt-3">
                <span className="text-zinc-400">Mobile Number</span>
                <span className="font-mono text-zinc-300">{user.mobileNumber || '-'}</span>
              </div>

              <div className="flex items-center justify-between pt-3">
                <span className="text-zinc-400">Registered Date</span>
                <span className="font-mono text-zinc-300">
                  {new Date(user.createdAt).toLocaleDateString()}
                </span>
              </div>
            </div>
          </Card>
        </div>

        {/* Delete Confirmation Modal */}
        <Dialog open={isDeleteModalOpen} onOpenChange={setIsDeleteModalOpen}>
          <DialogContent className="sm:max-w-sm bg-zinc-950 border-zinc-800 text-zinc-100 p-6">
            <DialogHeader className="space-y-2">
              <DialogTitle className="text-sm font-semibold text-zinc-100">
                Confirm User Deletion
              </DialogTitle>
              <DialogDescription className="text-xs text-zinc-400">
                Are you sure you want to permanently delete user &ldquo;{displayName}&rdquo; ({user.email})? This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter className="pt-3 gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsDeleteModalOpen(false)}
                className="text-xs border-zinc-800 bg-zinc-900 text-zinc-300"
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                size="sm"
                onClick={handleDeleteUser}
                disabled={isDeleting}
                className="text-xs bg-rose-600 hover:bg-rose-500 text-white"
              >
                Delete User
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </RoleGuard>
  );
}
