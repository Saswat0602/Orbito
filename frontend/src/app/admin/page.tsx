'use client';

import * as React from 'react';
import Link from 'next/link';
import {
  Users,
  Building2,
  Server,
  Activity,
  Layers,
  FolderGit2,
  ListTodo,
  Cpu,
  Sliders,
  ArrowRight,
  ShieldCheck,
  CheckCircle2,
} from 'lucide-react';

import { Role } from '@/lib/rbac';
import { RoleGuard } from '@/components/auth/RoleGuard';
import {
  useGetAdminDashboardStatsQuery,
  useGetQueueHealthQuery,
} from '@/lib/store/api/adminApi';
import { useGetRegistrationStatusQuery } from '@/lib/store/api/authApi';
import { useGetOrganizationsQuery } from '@/lib/store/api/organizationsApi';
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
  const { data: stats, isLoading: isStatsLoading } = useGetAdminDashboardStatsQuery();
  const { data: queueHealth } = useGetQueueHealthQuery();
  const { data: registrationStatus } = useGetRegistrationStatusQuery();
  const { data: organizations = [] } = useGetOrganizationsQuery();

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
              <Badge variant="outline" className="text-[10px] uppercase font-mono tracking-wider border-zinc-750 bg-zinc-850/80 text-zinc-300">
                Root Access
              </Badge>
            </div>
            <p className="text-xs text-zinc-400">
              Platform-wide telemetry, multi-tenant governance, background workers, and security settings.
            </p>
          </div>

          <div className="flex items-center gap-2">
            <Link href="/admin/settings">
              <Button variant="outline" size="sm" className="h-8 text-xs border-zinc-800 text-zinc-300 hover:bg-zinc-850">
                <Sliders className="h-3.5 w-3.5 mr-1.5" />
                System Settings
              </Button>
            </Link>
            <Link href="/admin/queues">
              <Button variant="outline" size="sm" className="h-8 text-xs border-zinc-800 text-zinc-300 hover:bg-zinc-850">
                <Cpu className="h-3.5 w-3.5 mr-1.5" />
                Queue Monitor
              </Button>
            </Link>
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
              {stats?.totalOrganizations ?? organizations.length}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">
              +{stats?.newOrgsThisWeek ?? 0} provisioned this week
            </p>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Platform Users
              </span>
              <Users className="h-4 w-4 text-zinc-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">
              {stats?.totalUsers ?? '...'}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">
              {stats?.activeUsers ?? 0} active accounts
            </p>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Workspaces & Projects
              </span>
              <FolderGit2 className="h-4 w-4 text-zinc-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">
              {stats ? `${stats.totalWorkspaces} / ${stats.totalProjects}` : '...'}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">
              Active operational clusters
            </p>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Total Tasks Processed
              </span>
              <ListTodo className="h-4 w-4 text-zinc-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">
              {stats?.totalTasks ?? '...'}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">
              +{stats?.newTasksThisWeek ?? 0} created this week
            </p>
          </Card>
        </div>

        {/* Admin Navigation Hub Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Link href="/admin/users" className="group">
            <Card className="border-zinc-800 bg-zinc-900/40 hover:bg-zinc-850/60 hover:border-zinc-700 transition-all p-5 h-full flex flex-col justify-between">
              <div className="space-y-2">
                <div className="h-9 w-9 rounded-md bg-zinc-850/80 border border-zinc-750 flex items-center justify-center text-zinc-300 group-hover:border-zinc-600 group-hover:text-white transition-colors">
                  <Users className="h-5 w-5" />
                </div>
                <h3 className="text-sm font-semibold text-zinc-100 group-hover:text-white transition-colors">
                  User Directory
                </h3>
                <p className="text-xs text-zinc-400">
                  Search, filter, update roles, toggle active/suspended status, and manage password resets.
                </p>
              </div>
              <div className="flex items-center text-xs text-zinc-400 group-hover:text-zinc-200 mt-4 font-medium">
                Manage Users <ArrowRight className="h-3.5 w-3.5 ml-1 transition-transform group-hover:translate-x-0.5" />
              </div>
            </Card>
          </Link>

          <Link href="/admin/organizations" className="group">
            <Card className="border-zinc-800 bg-zinc-900/40 hover:bg-zinc-850/60 hover:border-zinc-700 transition-all p-5 h-full flex flex-col justify-between">
              <div className="space-y-2">
                <div className="h-9 w-9 rounded-md bg-zinc-850/80 border border-zinc-750 flex items-center justify-center text-zinc-300 group-hover:border-zinc-600 group-hover:text-white transition-colors">
                  <Building2 className="h-5 w-5" />
                </div>
                <h3 className="text-sm font-semibold text-zinc-100 group-hover:text-white transition-colors">
                  Tenant Organizations
                </h3>
                <p className="text-xs text-zinc-400">
                  Multi-tenant governance, ownership transfers, audit logs, and archive toggles.
                </p>
              </div>
              <div className="flex items-center text-xs text-zinc-400 group-hover:text-zinc-200 mt-4 font-medium">
                Manage Tenants <ArrowRight className="h-3.5 w-3.5 ml-1 transition-transform group-hover:translate-x-0.5" />
              </div>
            </Card>
          </Link>

          <Link href="/admin/settings" className="group">
            <Card className="border-zinc-800 bg-zinc-900/40 hover:bg-zinc-850/60 hover:border-zinc-700 transition-all p-5 h-full flex flex-col justify-between">
              <div className="space-y-2">
                <div className="h-9 w-9 rounded-md bg-zinc-850/80 border border-zinc-750 flex items-center justify-center text-zinc-300 group-hover:border-zinc-600 group-hover:text-white transition-colors">
                  <Sliders className="h-5 w-5" />
                </div>
                <h3 className="text-sm font-semibold text-zinc-100 group-hover:text-white transition-colors">
                  System & SSO
                </h3>
                <p className="text-xs text-zinc-400">
                  Configure OpenID Connect (OIDC), public sign-up registration gate, and SMTP mailer.
                </p>
              </div>
              <div className="flex items-center text-xs text-zinc-400 group-hover:text-zinc-200 mt-4 font-medium">
                Manage Settings <ArrowRight className="h-3.5 w-3.5 ml-1 transition-transform group-hover:translate-x-0.5" />
              </div>
            </Card>
          </Link>

          <Link href="/admin/queues" className="group">
            <Card className="border-zinc-800 bg-zinc-900/40 hover:bg-zinc-850/60 hover:border-zinc-700 transition-all p-5 h-full flex flex-col justify-between">
              <div className="space-y-2">
                <div className="h-9 w-9 rounded-md bg-zinc-850/80 border border-zinc-750 flex items-center justify-center text-zinc-300 group-hover:border-zinc-600 group-hover:text-white transition-colors">
                  <Cpu className="h-5 w-5" />
                </div>
                <h3 className="text-sm font-semibold text-zinc-100 group-hover:text-white transition-colors">
                  Queue & Workers
                </h3>
                <p className="text-xs text-zinc-400">
                  BullMQ telemetry, Redis cluster health, worker throughput, and BullBoard dashboard.
                </p>
              </div>
              <div className="flex items-center text-xs text-zinc-400 group-hover:text-zinc-200 mt-4 font-medium">
                Inspect Queues <ArrowRight className="h-3.5 w-3.5 ml-1 transition-transform group-hover:translate-x-0.5" />
              </div>
            </Card>
          </Link>
        </div>

        {/* Registered Tenants Preview */}
        <Card className="border-zinc-800 bg-zinc-900/80">
          <CardHeader className="flex flex-row items-center justify-between pb-3">
            <div>
              <CardTitle className="text-base text-zinc-100">Registered Tenants</CardTitle>
              <CardDescription className="text-xs text-zinc-400">
                Active tenant containers running on the Orbito platform.
              </CardDescription>
            </div>
            <Link href="/admin/organizations">
              <Button variant="ghost" size="sm" className="h-7 text-xs text-zinc-400 hover:text-zinc-100">
                View All &rarr;
              </Button>
            </Link>
          </CardHeader>
          <CardContent>
            {organizations.length === 0 ? (
              <div className="text-center py-8 text-xs text-zinc-500">
                No organizations created yet.
              </div>
            ) : (
              <div className="divide-y divide-zinc-800/80">
                {organizations.slice(0, 5).map((org) => (
                  <div
                    key={org.id}
                    className="flex items-center justify-between py-3 px-2 hover:bg-zinc-850/50 rounded-md transition-colors"
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
                      <Badge variant="outline" className="text-[10px] border-zinc-800 text-zinc-400">
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
