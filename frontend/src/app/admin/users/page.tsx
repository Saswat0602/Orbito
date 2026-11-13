'use client';

import * as React from 'react';
import { useState } from 'react';
import Link from 'next/link';
import {
  Users,
  Search,
  Shield,
  UserCheck,
  UserX,
  KeyRound,
  Trash2,
  ChevronLeft,
  ChevronRight,
  Filter,
  Eye,
} from 'lucide-react';
import { toast } from 'sonner';

import { Role, UserStatus } from '@orbito/types';
import { RoleGuard } from '@/components/auth/RoleGuard';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import {
  useGetAdminUsersQuery,
  useUpdateAdminUserRoleMutation,
  useUpdateAdminUserStatusMutation,
  useDeleteAdminUserMutation,
  useResetAdminUserPasswordMutation,
} from '@/lib/store/api/adminApi';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Input } from '@/components/ui/input';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { ConfirmationModal } from '@/components/modals/ConfirmationModal';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';

export default function AdminUsersPage() {
  const { user: currentUser } = useAuthRbac();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [selectedRole, setSelectedRole] = useState<Role | undefined>(undefined);
  const [selectedStatus, setSelectedStatus] = useState<UserStatus | undefined>(undefined);
  const [userToDelete, setUserToDelete] = useState<{ id: string; email: string } | null>(null);

  const { data: usersResponse, isLoading } = useGetAdminUsersQuery({
    page,
    limit: 15,
    search: search.trim() || undefined,
    role: selectedRole,
    status: selectedStatus,
  });

  const [updateUserRole] = useUpdateAdminUserRoleMutation();
  const [updateUserStatus] = useUpdateAdminUserStatusMutation();
  const [deleteUser, { isLoading: isDeletingUser }] = useDeleteAdminUserMutation();
  const [resetPassword] = useResetAdminUserPasswordMutation();

  const handleRoleChange = async (userId: string, newRole: Role) => {
    try {
      await updateUserRole({ id: userId, role: newRole }).unwrap();
      toast.success(`User role updated to ${newRole}`);
    } catch (err: unknown) {
      const msg = err && typeof err === 'object' && 'data' in err
        ? (err as { data: { message?: string } }).data?.message
        : 'Failed to update user role';
      toast.error(msg || 'Failed to update role');
    }
  };

  const handleStatusToggle = async (userId: string, currentStatus: UserStatus) => {
    const newStatus =
      currentStatus === UserStatus.ACTIVE ? UserStatus.SUSPENDED : UserStatus.ACTIVE;
    try {
      await updateUserStatus({ id: userId, status: newStatus }).unwrap();
      toast.success(`User status changed to ${newStatus}`);
    } catch (err: unknown) {
      const msg = err && typeof err === 'object' && 'data' in err
        ? (err as { data: { message?: string } }).data?.message
        : 'Failed to update user status';
      toast.error(msg || 'Failed to update status');
    }
  };

  const handleResetPassword = async (userId: string, email: string) => {
    try {
      const res = await resetPassword(userId).unwrap();
      toast.success(res.message || `Password reset link sent to ${email}`);
    } catch (err: unknown) {
      toast.error('Failed to generate reset link');
    }
  };

  const handleConfirmDeleteUser = async () => {
    if (!userToDelete) return;
    try {
      await deleteUser(userToDelete.id).unwrap();
      toast.success(`User ${userToDelete.email} deleted.`);
      setUserToDelete(null);
    } catch {
      toast.error('Failed to delete user');
    }
  };

  const users = usersResponse?.data || [];
  const totalPages = usersResponse?.totalPages || 1;

  return (
    <RoleGuard requiredRole={Role.SUPER_ADMIN}>
      <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b]">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-zinc-800/80">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Users className="h-6 w-6 text-zinc-100" />
              <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
                User Directory
              </h1>
              <Badge variant="outline" className="text-[10px] border-zinc-800 text-zinc-400">
                {usersResponse?.total ?? 0} Accounts
              </Badge>
            </div>
            <p className="text-xs text-zinc-400">
              Manage platform-wide accounts, promote roles, toggle account statuses, and reset credentials.
            </p>
          </div>
        </div>

        {/* Filter & Search Bar */}
        <div className="flex flex-col sm:flex-row items-center gap-3">
          <div className="relative flex-1 w-full">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
            <Input
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
              placeholder="Search by name or email..."
              className="pl-9 h-9 text-xs bg-zinc-900 border-zinc-800 text-zinc-100"
            />
          </div>

          <div className="flex items-center gap-2 w-full sm:w-auto">
            {/* Role Filter */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm" className="h-9 text-xs border-zinc-800 text-zinc-300">
                  <Shield className="h-3.5 w-3.5 mr-1.5 text-zinc-400" />
                  {selectedRole ? `Role: ${selectedRole}` : 'All Roles'}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="bg-zinc-900 border-zinc-800 text-xs">
                <DropdownMenuItem onClick={() => setSelectedRole(undefined)}>All Roles</DropdownMenuItem>
                <DropdownMenuSeparator className="bg-zinc-800" />
                <DropdownMenuItem onClick={() => setSelectedRole(Role.SUPER_ADMIN)}>SUPER_ADMIN</DropdownMenuItem>
                <DropdownMenuItem onClick={() => setSelectedRole(Role.OWNER)}>OWNER</DropdownMenuItem>
                <DropdownMenuItem onClick={() => setSelectedRole(Role.MANAGER)}>MANAGER</DropdownMenuItem>
                <DropdownMenuItem onClick={() => setSelectedRole(Role.MEMBER)}>MEMBER</DropdownMenuItem>
                <DropdownMenuItem onClick={() => setSelectedRole(Role.VIEWER)}>VIEWER</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

            {/* Status Filter */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm" className="h-9 text-xs border-zinc-800 text-zinc-300">
                  <Filter className="h-3.5 w-3.5 mr-1.5 text-zinc-400" />
                  {selectedStatus ? `Status: ${selectedStatus}` : 'All Statuses'}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="bg-zinc-900 border-zinc-800 text-xs">
                <DropdownMenuItem onClick={() => setSelectedStatus(undefined)}>All Statuses</DropdownMenuItem>
                <DropdownMenuSeparator className="bg-zinc-800" />
                <DropdownMenuItem onClick={() => setSelectedStatus(UserStatus.ACTIVE)}>Active</DropdownMenuItem>
                <DropdownMenuItem onClick={() => setSelectedStatus(UserStatus.SUSPENDED)}>Suspended</DropdownMenuItem>
                <DropdownMenuItem onClick={() => setSelectedStatus(UserStatus.INACTIVE)}>Inactive</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {/* Users Table */}
        <Card className="border-zinc-800 bg-zinc-900/60">
          <CardContent className="p-0">
            {isLoading ? (
              <div className="divide-y divide-zinc-800/80">
                {[1, 2, 3, 4, 5].map((idx) => (
                  <div
                    key={idx}
                    className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-zinc-950/20"
                  >
                    <div className="flex items-center gap-3">
                      <Skeleton className="h-9 w-9 rounded-md bg-zinc-800 shrink-0" />
                      <div className="space-y-1.5">
                        <div className="flex items-center gap-2">
                          <Skeleton className="h-4 w-36 bg-zinc-800" />
                          <Skeleton className="h-4 w-12 rounded bg-zinc-850" />
                        </div>
                        <Skeleton className="h-3 w-48 bg-zinc-850" />
                      </div>
                    </div>
                    <div className="flex items-center gap-3 shrink-0">
                      <Skeleton className="h-5 w-20 rounded bg-zinc-850" />
                      <Skeleton className="h-6 w-6 rounded bg-zinc-850" />
                    </div>
                  </div>
                ))}
              </div>
            ) : users.length === 0 ? (
              <div className="py-16 text-center text-xs text-zinc-500">No users found matching criteria.</div>
            ) : (
              <div className="divide-y divide-zinc-800/80">
                {users.map((u) => {
                  const isCurrent = u.id === currentUser?.id;
                  const initials =
                    (u.firstName?.[0] || '') + (u.lastName?.[0] || '') ||
                    u.email.slice(0, 2).toUpperCase();

                  return (
                    <div
                      key={u.id}
                      className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-zinc-850/40 transition-colors"
                    >
                      <Link
                        href={`/admin/users/${u.id}`}
                        className="flex items-center gap-3 min-w-0 group"
                      >
                        <Avatar className="h-9 w-9 rounded-md border border-zinc-800 shrink-0 group-hover:border-zinc-500 transition-colors">
                          <AvatarFallback className="bg-zinc-800 text-zinc-200 text-xs font-bold">
                            {initials}
                          </AvatarFallback>
                        </Avatar>
                        <div className="flex flex-col min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-semibold text-zinc-100 group-hover:text-white transition-colors truncate">
                              {u.firstName || u.lastName ? `${u.firstName || ''} ${u.lastName || ''}`.trim() : 'Unnamed User'}
                            </span>
                            {isCurrent && (
                              <Badge variant="outline" className="text-[9px] border-zinc-700 text-zinc-400 py-0 px-1">
                                You
                              </Badge>
                            )}
                          </div>
                          <span className="text-[11px] text-zinc-400 font-mono truncate">{u.email}</span>
                        </div>
                      </Link>

                      <div className="flex items-center gap-3 shrink-0 self-end sm:self-center">
                        {/* Status Badge */}
                        <Badge
                          variant="outline"
                          className="text-[10px] uppercase font-medium border-zinc-800 bg-zinc-900/80 text-zinc-300 flex items-center gap-1.5"
                        >
                          <span
                            className={`h-1.5 w-1.5 rounded-full shrink-0 ${
                              u.status === UserStatus.ACTIVE ? 'bg-emerald-400' : 'bg-rose-400'
                            }`}
                          />
                          {u.status}
                        </Badge>

                        {/* Role Switcher Dropdown */}
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild disabled={isCurrent}>
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-7 text-[11px] border-zinc-800 bg-zinc-900 text-zinc-300 hover:bg-zinc-800"
                            >
                              <Shield className="h-3 w-3 mr-1 text-zinc-500" />
                              {u.role}
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end" className="bg-zinc-900 border-zinc-800 text-xs">
                            <DropdownMenuLabel className="text-[10px] text-zinc-500">Change Role</DropdownMenuLabel>
                            <DropdownMenuSeparator className="bg-zinc-800" />
                            <DropdownMenuItem onClick={() => handleRoleChange(u.id, Role.SUPER_ADMIN)}>SUPER_ADMIN</DropdownMenuItem>
                            <DropdownMenuItem onClick={() => handleRoleChange(u.id, Role.OWNER)}>OWNER</DropdownMenuItem>
                            <DropdownMenuItem onClick={() => handleRoleChange(u.id, Role.MANAGER)}>MANAGER</DropdownMenuItem>
                            <DropdownMenuItem onClick={() => handleRoleChange(u.id, Role.MEMBER)}>MEMBER</DropdownMenuItem>
                            <DropdownMenuItem onClick={() => handleRoleChange(u.id, Role.VIEWER)}>VIEWER</DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>

                        {/* Actions Dropdown */}
                        {!isCurrent && (
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="sm" className="h-7 w-7 p-0 text-zinc-400 hover:text-zinc-100">
                                •••
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end" className="bg-zinc-900 border-zinc-800 text-xs">
                              <DropdownMenuItem asChild>
                                <Link href={`/admin/users/${u.id}`} className="cursor-pointer flex items-center">
                                  <Eye className="h-3.5 w-3.5 mr-2 text-zinc-400" />
                                  Inspect User
                                </Link>
                              </DropdownMenuItem>
                              <DropdownMenuSeparator className="bg-zinc-800" />
                              <DropdownMenuItem
                                onClick={() => handleStatusToggle(u.id, u.status)}
                                className="cursor-pointer"
                              >
                                {u.status === UserStatus.ACTIVE ? (
                                  <>
                                    <UserX className="h-3.5 w-3.5 mr-2 text-amber-400" />
                                    Suspend Account
                                  </>
                                ) : (
                                  <>
                                    <UserCheck className="h-3.5 w-3.5 mr-2 text-emerald-400" />
                                    Activate Account
                                  </>
                                )}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleResetPassword(u.id, u.email)}
                                className="cursor-pointer"
                              >
                                <KeyRound className="h-3.5 w-3.5 mr-2 text-blue-400" />
                                Send Reset Password
                              </DropdownMenuItem>
                              <DropdownMenuSeparator className="bg-zinc-800" />
                              <DropdownMenuItem
                                onClick={() => setUserToDelete({ id: u.id, email: u.email })}
                                className="text-rose-400 hover:text-rose-300 cursor-pointer"
                              >
                                <Trash2 className="h-3.5 w-3.5 mr-2" />
                                Delete Account
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Pagination Bar */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between pt-2">
            <span className="text-xs text-zinc-500">
              Page {page} of {totalPages}
            </span>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className="h-8 text-xs border-zinc-800 text-zinc-300"
              >
                <ChevronLeft className="h-3.5 w-3.5 mr-1" /> Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                className="h-8 text-xs border-zinc-800 text-zinc-300"
              >
                Next <ChevronRight className="h-3.5 w-3.5 ml-1" />
              </Button>
            </div>
          </div>
        )}

        {/* Delete User Confirmation Modal */}
        <ConfirmationModal
          open={!!userToDelete}
          onOpenChange={(open) => !open && setUserToDelete(null)}
          title="Delete User Account"
          description={
            <span>
              Are you sure you want to delete user{' '}
              <strong className="text-zinc-200">{userToDelete?.email}</strong>?
              This action cannot be undone and will permanently remove their account, memberships, and session records.
            </span>
          }
          confirmLabel="Delete User"
          variant="destructive"
          isLoading={isDeletingUser}
          onConfirm={handleConfirmDeleteUser}
        />
      </div>
    </RoleGuard>
  );
}
