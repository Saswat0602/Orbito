'use client';

import * as React from 'react';
import { useState } from 'react';
import {
  Building2,
  Search,
  Users,
  Briefcase,
  Archive,
  ArchiveRestore,
  UserCheck,
  ChevronLeft,
  ChevronRight,
  ShieldAlert,
} from 'lucide-react';
import { toast } from 'sonner';

import { Role } from '@orbito/types';
import { RoleGuard } from '@/components/auth/RoleGuard';
import {
  useGetAdminOrganizationsQuery,
  useToggleAdminOrgArchiveMutation,
  useTransferAdminOrgOwnershipMutation,
} from '@/lib/store/api/adminApi';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function AdminOrganizationsPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');

  const { data: orgsResponse, isLoading } = useGetAdminOrganizationsQuery({
    page,
    limit: 15,
    search: search.trim() || undefined,
  });

  const [toggleArchive] = useToggleAdminOrgArchiveMutation();
  const [transferOwnership] = useTransferAdminOrgOwnershipMutation();

  const handleToggleArchive = async (orgId: string, currentArchive: boolean) => {
    try {
      const res = await toggleArchive(orgId).unwrap();
      toast.success(
        res.archive
          ? `Tenant "${res.name}" has been suspended/archived.`
          : `Tenant "${res.name}" has been activated.`
      );
    } catch (err: unknown) {
      toast.error('Failed to change organization archive status');
    }
  };

  const organizations = orgsResponse?.data || [];
  const totalPages = orgsResponse?.totalPages || 1;

  return (
    <RoleGuard requiredRole={Role.SUPER_ADMIN}>
      <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b]">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-zinc-800/80">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Building2 className="h-6 w-6 text-zinc-100" />
              <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
                Tenant Organizations
              </h1>
              <Badge variant="outline" className="text-[10px] border-zinc-800 text-zinc-400">
                {orgsResponse?.total ?? 0} Tenants
              </Badge>
            </div>
            <p className="text-xs text-zinc-400">
              Audit multi-tenant containers, view workspace and member counts, and suspend or restore tenants.
            </p>
          </div>
        </div>

        {/* Search */}
        <div className="relative max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
          <Input
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            placeholder="Search organizations by name or slug..."
            className="pl-9 h-9 text-xs bg-zinc-900 border-zinc-800 text-zinc-100"
          />
        </div>

        {/* Organizations Table */}
        <Card className="border-zinc-800 bg-zinc-900/60">
          <CardContent className="p-0">
            {isLoading ? (
              <div className="py-16 text-center text-xs text-zinc-500">Loading organizations...</div>
            ) : organizations.length === 0 ? (
              <div className="py-16 text-center text-xs text-zinc-500">No organizations found.</div>
            ) : (
              <div className="divide-y divide-zinc-800/80">
                {organizations.map((org) => {
                  return (
                    <div
                      key={org.id}
                      className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:bg-zinc-850/40 transition-colors"
                    >
                      <div className="flex items-center gap-3 min-w-0">
                        <div className="h-9 w-9 rounded-md bg-zinc-800 border border-zinc-700/80 flex items-center justify-center text-sm font-bold text-zinc-200 shrink-0">
                          {org.name[0]?.toUpperCase()}
                        </div>
                        <div className="flex flex-col min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-semibold text-zinc-100 truncate">
                              {org.name}
                            </span>
                            {org.archive && (
                              <Badge variant="destructive" className="text-[9px] py-0 px-1">
                                Suspended
                              </Badge>
                            )}
                          </div>
                          <span className="text-[11px] text-zinc-500 font-mono truncate">
                            slug: {org.slug} • id: {org.id.slice(0, 8)}...
                          </span>
                        </div>
                      </div>

                      <div className="flex flex-wrap items-center gap-4 text-xs text-zinc-400">
                        {/* Owner Info */}
                        <div className="flex items-center gap-1.5 min-w-[120px]">
                          <UserCheck className="h-3.5 w-3.5 text-zinc-500" />
                          <span className="truncate">
                            {org.owner ? `${org.owner.firstName || ''} ${org.owner.lastName || ''}`.trim() || org.owner.email : 'No owner'}
                          </span>
                        </div>

                        {/* Workspaces Count */}
                        <div className="flex items-center gap-1 min-w-[70px]">
                          <Briefcase className="h-3.5 w-3.5 text-zinc-500" />
                          <span>{org._count?.workspaces ?? 0} ws</span>
                        </div>

                        {/* Members Count */}
                        <div className="flex items-center gap-1 min-w-[70px]">
                          <Users className="h-3.5 w-3.5 text-zinc-500" />
                          <span>{org._count?.members ?? 0} users</span>
                        </div>

                        {/* Archive Toggle Button */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleToggleArchive(org.id, org.archive)}
                          className={`h-7 text-xs border-zinc-800 ${
                            org.archive
                              ? 'text-emerald-400 hover:bg-emerald-950/40'
                              : 'text-zinc-400 hover:text-amber-400 hover:bg-zinc-800'
                          }`}
                        >
                          {org.archive ? (
                            <>
                              <ArchiveRestore className="h-3.5 w-3.5 mr-1 text-emerald-400" />
                              Restore
                            </>
                          ) : (
                            <>
                              <Archive className="h-3.5 w-3.5 mr-1" />
                              Suspend
                            </>
                          )}
                        </Button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Pagination */}
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
      </div>
    </RoleGuard>
  );
}

