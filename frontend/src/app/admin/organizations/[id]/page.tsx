'use client';

import * as React from 'react';
import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  ArrowLeft,
  Building2,
  Users,
  Briefcase,
  Archive,
  ArchiveRestore,
  UserCheck,
  Calendar,
  Clock,
  Trash2,
  ExternalLink,
  Shield,
  Layers,
  AlertTriangle,
  Loader2,
  UserCog,
} from 'lucide-react';
import { toast } from 'sonner';

import { Role } from '@orbito/types';
import { RoleGuard } from '@/components/auth/RoleGuard';
import {
  useGetAdminOrganizationDetailQuery,
  useToggleAdminOrgArchiveMutation,
  useTransferAdminOrgOwnershipMutation,
  useDeleteAdminOrganizationMutation,
} from '@/lib/store/api/adminApi';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { AdminDetailSkeleton } from '@/components/skeletons';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export default function AdminOrganizationDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const orgId = params.id;

  const { data: org, isLoading, refetch } = useGetAdminOrganizationDetailQuery(orgId, {
    skip: !orgId,
  });

  const [toggleArchive, { isLoading: isTogglingArchive }] = useToggleAdminOrgArchiveMutation();
  const [transferOwnership, { isLoading: isTransferring }] = useTransferAdminOrgOwnershipMutation();
  const [deleteOrg, { isLoading: isDeleting }] = useDeleteAdminOrganizationMutation();

  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [isTransferModalOpen, setIsTransferModalOpen] = useState(false);
  const [selectedNewOwnerId, setSelectedNewOwnerId] = useState<string>('');

  const handleToggleArchive = async () => {
    if (!orgId || !org) return;
    try {
      const res = await toggleArchive(orgId).unwrap();
      toast.success(
        res.archive
          ? `Tenant "${res.name}" has been suspended.`
          : `Tenant "${res.name}" has been activated.`
      );
      refetch();
    } catch {
      toast.error('Failed to change organization archive status');
    }
  };

  const handleTransferOwnership = async () => {
    if (!orgId || !selectedNewOwnerId) return;
    try {
      await transferOwnership({ id: orgId, newOwnerId: selectedNewOwnerId }).unwrap();
      toast.success('Organization ownership successfully transferred');
      setIsTransferModalOpen(false);
      refetch();
    } catch {
      toast.error('Failed to transfer organization ownership');
    }
  };

  const handleDeleteOrganization = async () => {
    if (!orgId || !org) return;
    try {
      await deleteOrg(orgId).unwrap();
      toast.success(`Organization "${org.name}" permanently deleted`);
      setIsDeleteModalOpen(false);
      router.push('/admin/organizations');
    } catch {
      toast.error('Failed to delete organization');
    }
  };

  if (isLoading) {
    return (
      <RoleGuard requiredRole={Role.SUPER_ADMIN}>
        <AdminDetailSkeleton />
      </RoleGuard>
    );
  }

  if (!org) {
    return (
      <RoleGuard requiredRole={Role.SUPER_ADMIN}>
        <div className="flex-1 p-6 md:p-8 space-y-4">
          <Link
            href="/admin/organizations"
            className="inline-flex items-center gap-2 text-xs text-zinc-400 hover:text-zinc-200 transition-colors"
          >
            <ArrowLeft className="h-4 w-4" /> Back to Organizations
          </Link>
          <Card className="border-zinc-800 bg-zinc-900/60 p-8 text-center">
            <Building2 className="h-10 w-10 text-zinc-600 mx-auto mb-3" />
            <h3 className="text-sm font-semibold text-zinc-200">Organization Not Found</h3>
            <p className="text-xs text-zinc-500 mt-1">
              The requested tenant container with ID {orgId} does not exist or has been deleted.
            </p>
          </Card>
        </div>
      </RoleGuard>
    );
  }

  const totalProjects = (org.workspaces || []).reduce(
    (acc, ws) => acc + (ws._count?.projects || 0),
    0
  );

  return (
    <RoleGuard requiredRole={Role.SUPER_ADMIN}>
      <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b]">
        {/* Navigation Breadcrumb */}
        <div className="flex items-center justify-between pb-2 border-b border-zinc-800/80">
          <div className="flex items-center gap-2 text-xs text-zinc-400">
            <Link
              href="/admin/organizations"
              className="hover:text-zinc-200 transition-colors flex items-center gap-1"
            >
              <ArrowLeft className="h-3.5 w-3.5" />
              Organizations
            </Link>
            <span className="text-zinc-600">/</span>
            <span className="text-zinc-200 font-medium truncate max-w-xs">{org.name}</span>
          </div>
          <Badge
            variant="outline"
            className="text-[10px] uppercase font-medium border-zinc-800 bg-zinc-900/80 text-zinc-300 flex items-center gap-1.5"
          >
            <span
              className={`h-1.5 w-1.5 rounded-full shrink-0 ${
                org.archive ? 'bg-amber-400' : 'bg-emerald-400'
              }`}
            />
            {org.archive ? 'Suspended / Archived' : 'Active Tenant'}
          </Badge>
        </div>

        {/* Organization Header Profile Card */}
        <Card className="border-zinc-800 bg-zinc-900/60">
          <CardContent className="p-6">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
              <div className="flex items-start gap-4">
                <div className="h-14 w-14 rounded-md bg-zinc-800 border border-zinc-700/80 flex items-center justify-center text-xl font-bold text-zinc-200 shrink-0">
                  {org.name[0]?.toUpperCase()}
                </div>
                <div className="space-y-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <h1 className="text-xl font-bold text-zinc-100">{org.name}</h1>
                    <span className="text-xs text-zinc-500 font-mono">({org.slug})</span>
                  </div>
                  {org.description && (
                    <p className="text-xs text-zinc-400 max-w-xl">{org.description}</p>
                  )}
                  <div className="flex flex-wrap items-center gap-4 text-[11px] text-zinc-500 pt-1 font-mono">
                    <span>ID: {org.id}</span>
                    <span>•</span>
                    <span className="flex items-center gap-1">
                      <Calendar className="h-3 w-3" />
                      Created {new Date(org.createdAt).toLocaleDateString()}
                    </span>
                    <span>•</span>
                    <span className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      Updated {new Date(org.updatedAt).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="flex flex-wrap items-center gap-2 shrink-0">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleToggleArchive}
                  disabled={isTogglingArchive}
                  className="h-8 text-xs border-zinc-800 text-zinc-300 hover:text-zinc-100 hover:bg-zinc-800 cursor-pointer"
                >
                  {org.archive ? (
                    <>
                      <ArchiveRestore className="h-3.5 w-3.5 mr-1.5" />
                      Activate Tenant
                    </>
                  ) : (
                    <>
                      <Archive className="h-3.5 w-3.5 mr-1.5" />
                      Suspend Tenant
                    </>
                  )}
                </Button>

                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setIsTransferModalOpen(true)}
                  className="h-8 text-xs border-zinc-800 text-zinc-300 hover:text-zinc-100 hover:bg-zinc-800 cursor-pointer"
                >
                  <UserCog className="h-3.5 w-3.5 mr-1.5 text-zinc-400" />
                  Transfer Ownership
                </Button>

                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setIsDeleteModalOpen(true)}
                  className="h-8 text-xs bg-zinc-900 hover:bg-zinc-850 border-zinc-800 hover:border-zinc-700 text-zinc-400 hover:text-rose-400 cursor-pointer transition-colors"
                >
                  <Trash2 className="h-3.5 w-3.5 mr-1.5" />
                  Delete Tenant
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Quick Stats Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="border-zinc-800 bg-zinc-900/60 p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs text-zinc-400">Total Workspaces</span>
              <Briefcase className="h-4 w-4 text-zinc-500" />
            </div>
            <div className="text-2xl font-bold text-zinc-100 mt-2">
              {org._count?.workspaces ?? org.workspaces?.length ?? 0}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">Tenant working scopes</p>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs text-zinc-400">Total Members</span>
              <Users className="h-4 w-4 text-zinc-500" />
            </div>
            <div className="text-2xl font-bold text-zinc-100 mt-2">
              {org._count?.members ?? org.members?.length ?? 0}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">Assigned users</p>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs text-zinc-400">Total Projects</span>
              <Layers className="h-4 w-4 text-zinc-500" />
            </div>
            <div className="text-2xl font-bold text-zinc-100 mt-2">{totalProjects}</div>
            <p className="text-[11px] text-zinc-500 mt-1">Across all workspaces</p>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs text-zinc-400">Primary Owner</span>
              <UserCheck className="h-4 w-4 text-emerald-400" />
            </div>
            <div className="text-sm font-semibold text-zinc-200 mt-2 truncate">
              {org.owner
                ? `${org.owner.firstName || ''} ${org.owner.lastName || ''}`.trim() || org.owner.email
                : 'No owner assigned'}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1 font-mono truncate">
              {org.owner?.email || 'N/A'}
            </p>
          </Card>
        </div>

        {/* Workspaces Section */}
        <Card className="border-zinc-800 bg-zinc-900/60">
          <CardHeader className="p-4 pb-3 border-b border-zinc-800/80">
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="text-sm font-semibold text-zinc-200 flex items-center gap-2">
                  <Briefcase className="h-4 w-4 text-zinc-400" />
                  Tenant Workspaces ({org.workspaces?.length || 0})
                </CardTitle>
                <CardDescription className="text-xs text-zinc-400 mt-0.5">
                  Workspaces owned and partitioned within this organization container
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            {(!org.workspaces || org.workspaces.length === 0) ? (
              <div className="py-12 text-center text-xs text-zinc-500">
                No workspaces created in this organization yet.
              </div>
            ) : (
              <div className="divide-y divide-zinc-800/80">
                {org.workspaces.map((ws) => (
                  <div
                    key={ws.id}
                    className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-zinc-850/40 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <div className="h-8 w-8 rounded-md bg-zinc-800 border border-zinc-700/60 flex items-center justify-center text-xs font-bold text-zinc-300">
                        {ws.name[0]?.toUpperCase()}
                      </div>
                      <div>
                        <div className="text-xs font-semibold text-zinc-100">{ws.name}</div>
                        <div className="text-[11px] text-zinc-500 font-mono">
                          slug: {ws.slug} • id: {ws.id.slice(0, 8)}...
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-4 text-xs text-zinc-400">
                      <span>{ws._count?.projects ?? 0} projects</span>
                      <span>•</span>
                      <span>{ws._count?.members ?? 0} members</span>
                      <Link
                        href={`/${ws.slug}`}
                        target="_blank"
                        className="inline-flex items-center gap-1 text-[11px] text-zinc-400 hover:text-zinc-200 transition-colors ml-2"
                      >
                        Open Workspace <ExternalLink className="h-3 w-3" />
                      </Link>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Organization Members Section */}
        <Card className="border-zinc-800 bg-zinc-900/60">
          <CardHeader className="p-4 pb-3 border-b border-zinc-800/80">
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="text-sm font-semibold text-zinc-200 flex items-center gap-2">
                  <Users className="h-4 w-4 text-zinc-400" />
                  Organization Members ({org.members?.length || 0})
                </CardTitle>
                <CardDescription className="text-xs text-zinc-400 mt-0.5">
                  Users with tenant-level membership and assigned permission roles
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            {(!org.members || org.members.length === 0) ? (
              <div className="py-12 text-center text-xs text-zinc-500">
                No members recorded in this organization.
              </div>
            ) : (
              <div className="divide-y divide-zinc-800/80">
                {org.members.map((m) => {
                  const initials =
                    (m.user.firstName?.[0] || '') + (m.user.lastName?.[0] || '') ||
                    m.user.email.slice(0, 2).toUpperCase();

                  const isOwner = org.owner?.id === m.user.id;

                  return (
                    <div
                      key={m.user.id}
                      className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-zinc-850/40 transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        <Avatar className="h-8 w-8 rounded-md border border-zinc-800 shrink-0">
                          {m.user.avatar && <AvatarImage src={m.user.avatar} alt={m.user.email} />}
                          <AvatarFallback className="bg-zinc-800 text-zinc-200 text-xs font-bold">
                            {initials}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-semibold text-zinc-100">
                              {m.user.firstName || m.user.lastName
                                ? `${m.user.firstName || ''} ${m.user.lastName || ''}`.trim()
                                : 'Unnamed User'}
                            </span>
                            {isOwner && (
                              <Badge
                                variant="outline"
                                className="text-[9px] border-emerald-800 text-emerald-400 py-0 px-1"
                              >
                                Tenant Owner
                              </Badge>
                            )}
                          </div>
                          <span className="text-[11px] text-zinc-400 font-mono">{m.user.email}</span>
                        </div>
                      </div>

                      <div className="flex items-center gap-3">
                        <Badge
                          variant="outline"
                          className="text-[10px] border-zinc-700 text-zinc-300 font-mono"
                        >
                          {m.role}
                        </Badge>
                        <span className="text-[11px] text-zinc-500">
                          Joined {new Date(m.createdAt).toLocaleDateString()}
                        </span>
                        <Link
                          href={`/admin/users/${m.user.id}`}
                          className="inline-flex items-center gap-1 text-[11px] text-zinc-400 hover:text-zinc-200 ml-2"
                        >
                          Inspect Account <ExternalLink className="h-3 w-3" />
                        </Link>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Transfer Ownership Dialog */}
        <Dialog open={isTransferModalOpen} onOpenChange={setIsTransferModalOpen}>
          <DialogContent className="bg-zinc-900 border-zinc-800 text-zinc-100 max-w-md">
            <DialogHeader>
              <DialogTitle className="text-sm font-semibold flex items-center gap-2">
                <UserCog className="h-4 w-4 text-zinc-300" />
                Transfer Tenant Ownership
              </DialogTitle>
              <DialogDescription className="text-xs text-zinc-400">
                Select an organization member to promote as the new primary owner of{' '}
                <span className="text-zinc-200 font-semibold">{org.name}</span>.
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-3 py-2">
              <label className="text-xs text-zinc-300 font-medium">Select New Owner</label>
              <Select value={selectedNewOwnerId} onValueChange={setSelectedNewOwnerId}>
                <SelectTrigger className="bg-zinc-950 border-zinc-800 text-xs text-zinc-200">
                  <SelectValue placeholder="Choose a member..." />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-zinc-800 text-xs text-zinc-200">
                  {(org.members || [])
                    .filter((m) => m.user.id !== org.owner?.id)
                    .map((m) => (
                      <SelectItem key={m.user.id} value={m.user.id}>
                        {m.user.firstName || m.user.lastName
                          ? `${m.user.firstName || ''} ${m.user.lastName || ''} (${m.user.email})`
                          : m.user.email}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>

            <DialogFooter className="gap-2 sm:gap-0">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsTransferModalOpen(false)}
                className="h-8 text-xs border-zinc-800 text-zinc-300"
              >
                Cancel
              </Button>
              <Button
                size="sm"
                disabled={!selectedNewOwnerId || isTransferring}
                onClick={handleTransferOwnership}
                className="h-8 text-xs bg-zinc-100 text-zinc-900 hover:bg-zinc-200"
              >
                {isTransferring ? 'Transferring...' : 'Confirm Transfer'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete Organization Confirmation Dialog */}
        <Dialog open={isDeleteModalOpen} onOpenChange={setIsDeleteModalOpen}>
          <DialogContent className="bg-zinc-900 border-zinc-800 text-zinc-100 max-w-md">
            <DialogHeader>
              <DialogTitle className="text-sm font-semibold flex items-center gap-2 text-rose-400">
                <AlertTriangle className="h-4 w-4 text-rose-400" />
                Delete Organization Permanently
              </DialogTitle>
              <DialogDescription className="text-xs text-zinc-400">
                Are you sure you want to permanently delete tenant{' '}
                <span className="text-zinc-200 font-semibold">{org.name}</span>? This action is
                irreversible and will cascade-delete all workspaces, projects, tasks, and member
                associations.
              </DialogDescription>
            </DialogHeader>

            <DialogFooter className="gap-2 sm:gap-0 pt-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsDeleteModalOpen(false)}
                className="h-8 text-xs border-zinc-800 text-zinc-300"
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                size="sm"
                disabled={isDeleting}
                onClick={handleDeleteOrganization}
                className="h-8 text-xs bg-rose-600 hover:bg-rose-700 text-white"
              >
                {isDeleting ? 'Deleting...' : 'Permanently Delete'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </RoleGuard>
  );
}

