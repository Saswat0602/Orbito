'use client';

import * as React from 'react';
import {
  ShieldAlert,
  Users,
  Building2,
  Server,
  ToggleLeft,
  ToggleRight,
  Activity,
  CheckCircle2,
} from 'lucide-react';
import { toast } from 'sonner';

import { Role } from '@/lib/rbac';
import { RoleGuard } from '@/components/auth/RoleGuard';
import { useGetRegistrationStatusQuery } from '@/lib/store/api/authApi';
import { useGetOrganizationsQuery } from '@/lib/store/api/organizationsApi';
import { useGetUsersQuery } from '@/lib/store/api/usersApi';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

export default function AdminConsolePage() {
  const { data: registrationStatus, isLoading: isRegLoading } =
    useGetRegistrationStatusQuery();
  const { data: organizations = [] } = useGetOrganizationsQuery();
  const { data: users = [] } = useGetUsersQuery();

  return (
    <RoleGuard requiredRole={Role.SUPER_ADMIN}>
      <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
                System Administration
              </h1>
              <Badge variant="destructive" className="text-xs">
                Root Access
              </Badge>
            </div>
            <p className="text-xs text-zinc-400">
              Manage multi-tenant organizations, system registrations, and cluster topology.
            </p>
          </div>
        </div>

        {/* Overview Stats */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Total Organizations
              </span>
              <Building2 className="h-4 w-4 text-zinc-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">
              {organizations.length}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">Active tenant containers</p>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Registered Users
              </span>
              <Users className="h-4 w-4 text-zinc-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">{users.length}</div>
            <p className="text-[11px] text-zinc-500 mt-1">Platform-wide accounts</p>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                System Health
              </span>
              <Activity className="h-4 w-4 text-emerald-400" />
            </div>
            <div className="text-2xl font-bold text-emerald-400">Operational</div>
            <p className="text-[11px] text-zinc-500 mt-1">All services responding</p>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Public Sign-Ups
              </span>
              <Server className="h-4 w-4 text-zinc-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">
              {registrationStatus?.enabled ? 'Enabled' : 'Restricted'}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">Registration gate</p>
          </Card>
        </div>

        {/* Organizations Registry Table */}
        <Card className="border-zinc-800 bg-zinc-900/80">
          <CardHeader>
            <CardTitle className="text-base">Registered Tenants</CardTitle>
            <CardDescription className="text-xs">
              Direct access and administrative audit for all organizations.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {organizations.length === 0 ? (
              <div className="text-center py-8 text-xs text-zinc-500">
                No organizations created yet.
              </div>
            ) : (
              <div className="divide-y divide-zinc-800/80">
                {organizations.map((org) => (
                  <div
                    key={org.id}
                    className="flex items-center justify-between py-3 px-2 hover:bg-zinc-850/50 rounded-lg transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <div className="h-8 w-8 rounded-md bg-zinc-800 flex items-center justify-center text-xs font-bold text-zinc-200">
                        {org.name[0]?.toUpperCase()}
                      </div>
                      <div>
                        <div className="text-sm font-semibold text-zinc-200">
                          {org.name}
                        </div>
                        <div className="text-xs text-zinc-500 font-mono">
                          slug: {org.slug}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="text-[10px]">
                        ID: {org.id.slice(0, 8)}...
                      </Badge>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </RoleGuard>
  );
}

