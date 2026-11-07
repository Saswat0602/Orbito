'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard,
  ListTodo,
  Target,
  FolderGit2,
  Inbox,
  Activity,
  Zap,
  Puzzle,
  Users,
  Shield,
  Settings,
  LogOut,
  ChevronDown,
  Menu,
  X,
  Plus,
  Building2,
  Briefcase,
} from 'lucide-react';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import {
  setActiveOrganization,
  setActiveWorkspace,
  setActiveProject,
} from '@/lib/store/slices/authSlice';
import { useGetOrganizationsQuery } from '@/lib/store/api/organizationsApi';
import { useGetWorkspacesQuery } from '@/lib/store/api/workspacesApi';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import { Role } from '@/lib/rbac';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

export function Sidebar() {
  const pathname = usePathname();
  const dispatch = useAppDispatch();
  const [isMobileOpen, setIsMobileOpen] = useState(false);

  const {
    user,
    role,
    isSuperAdmin,
    isOwner,
    isManager,
    roleBadge,
    logout,
  } = useAuthRbac();

  const authState = useAppSelector((state) => state.auth);

  // Fetch Tenancy hierarchy
  const { data: organizations = [] } = useGetOrganizationsQuery();

  const activeOrgId = authState.activeOrganizationId || organizations[0]?.id || null;

  const { data: workspaces = [] } = useGetWorkspacesQuery(
    activeOrgId ? { organizationId: activeOrgId } : undefined,
    { skip: !activeOrgId }
  );

  const activeWorkspaceId =
    authState.activeWorkspaceId || workspaces[0]?.id || null;

  const { data: projects = [] } = useGetProjectsQuery(
    activeWorkspaceId ? { workspaceId: activeWorkspaceId } : undefined,
    { skip: !activeWorkspaceId }
  );

  const activeProjectId = authState.activeProjectId || projects[0]?.id || null;

  // Active names
  const activeOrgName = useMemo(() => {
    return organizations.find((o) => o.id === activeOrgId)?.name || 'Select Org';
  }, [organizations, activeOrgId]);

  const activeWorkspaceName = useMemo(() => {
    return workspaces.find((w) => w.id === activeWorkspaceId)?.name || 'Workspace';
  }, [workspaces, activeWorkspaceId]);

  // Main Navigation Items
  const navigation = [
    { name: 'Dashboard', href: '/', icon: LayoutDashboard },
    { name: 'Tasks', href: '/tasks', icon: ListTodo },
    { name: 'Sprints', href: '/sprints', icon: Target },
    { name: 'Projects', href: '/projects', icon: FolderGit2 },
    { name: 'Project Inbox', href: '/inbox', icon: Inbox },
    { name: 'Activity Log', href: '/activity', icon: Activity },
  ];

  // RBAC Management Links (Manager and above)
  const managementLinks = [
    { name: 'Automations', href: '/settings/automation', icon: Zap },
    { name: 'Integrations', href: '/settings/integrations', icon: Puzzle },
    { name: 'Members', href: '/settings/members', icon: Users },
  ];

  const userInitials = useMemo(() => {
    if (!user) return 'OB';
    const first = user.firstName?.[0] || '';
    const last = user.lastName?.[0] || '';
    return (first + last).toUpperCase() || user.username?.[0]?.toUpperCase() || 'U';
  }, [user]);

  return (
    <>
      {/* Mobile Toggle Button */}
      <button
        type="button"
        onClick={() => setIsMobileOpen(!isMobileOpen)}
        className="md:hidden fixed top-3 left-3 z-50 p-2 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-300 shadow-md"
        aria-label="Toggle navigation"
      >
        {isMobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
      </button>

      {/* Mobile Backdrop */}
      {isMobileOpen && (
        <div
          className="md:hidden fixed inset-0 bg-black/70 backdrop-blur-sm z-40"
          onClick={() => setIsMobileOpen(false)}
        />
      )}

      {/* Sidebar Container */}
      <aside
        className={`
          w-64 bg-[#09090b] border-r border-zinc-800/80 flex flex-col shrink-0 h-screen
          fixed md:static inset-y-0 left-0 z-40 transition-transform duration-200 ease-in-out
          ${isMobileOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'}
        `}
      >
        {/* Brand & Organization Switcher */}
        <div className="p-4 border-b border-zinc-800/80">
          <div className="flex items-center gap-2.5 mb-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-zinc-100 text-zinc-950 font-bold shadow-sm">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="h-4 w-4"
              >
                <circle cx="12" cy="12" r="4" />
                <path d="M21.17 8A10.06 10.06 0 0 0 12 2a10 10 0 0 0-8.94 5.5" />
                <path d="M2.83 16A10.06 10.06 0 0 0 12 22a10 10 0 0 0 8.94-5.5" />
              </svg>
            </div>
            <div className="flex flex-col">
              <span className="text-sm font-bold tracking-tight text-zinc-100">
                Orbito
              </span>
              <span className="text-[10px] uppercase font-semibold text-zinc-500 tracking-wider">
                Enterprise
              </span>
            </div>
          </div>

          {/* Organization & Workspace Pills */}
          <div className="space-y-1">
            <div className="flex items-center justify-between px-2.5 py-1.5 rounded-md bg-zinc-900 border border-zinc-850 text-xs">
              <div className="flex items-center gap-2 truncate">
                <Building2 className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
                <span className="truncate text-zinc-200 font-medium">
                  {activeOrgName}
                </span>
              </div>
              <ChevronDown className="h-3 w-3 text-zinc-500 shrink-0" />
            </div>

            <div className="flex items-center justify-between px-2.5 py-1.5 rounded-md bg-zinc-900/60 border border-zinc-850 text-xs">
              <div className="flex items-center gap-2 truncate">
                <Briefcase className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
                <span className="truncate text-zinc-300">
                  {activeWorkspaceName}
                </span>
              </div>
              <ChevronDown className="h-3 w-3 text-zinc-500 shrink-0" />
            </div>
          </div>
        </div>

        {/* Navigation Links Scrollable Body */}
        <div className="flex-1 overflow-y-auto px-3 py-4 space-y-6">
          {/* Main Links */}
          <div className="space-y-1">
            <span className="px-3 text-[10px] font-bold uppercase tracking-wider text-zinc-500">
              Workspace
            </span>
            {navigation.map((item) => {
              const isActive =
                item.href === '/'
                  ? pathname === '/'
                  : pathname?.startsWith(item.href);
              const Icon = item.icon;

              return (
                <Link
                  key={item.name}
                  href={item.href}
                  onClick={() => setIsMobileOpen(false)}
                  className={`
                    flex items-center gap-2.5 px-3 py-2 rounded-lg text-xs font-medium transition-colors
                    ${
                      isActive
                        ? 'bg-zinc-100 text-zinc-950 font-semibold shadow-xs'
                        : 'text-zinc-400 hover:text-zinc-100 hover:bg-zinc-900'
                    }
                  `}
                >
                  <Icon className="h-4 w-4 shrink-0" />
                  <span>{item.name}</span>
                </Link>
              );
            })}
          </div>

          {/* Management Links (RBAC: Manager & Owner only) */}
          {(isManager || isOwner) && (
            <div className="space-y-1">
              <span className="px-3 text-[10px] font-bold uppercase tracking-wider text-zinc-500">
                Management
              </span>
              {managementLinks.map((item) => {
                const isActive = pathname?.startsWith(item.href);
                const Icon = item.icon;

                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    onClick={() => setIsMobileOpen(false)}
                    className={`
                      flex items-center gap-2.5 px-3 py-2 rounded-lg text-xs font-medium transition-colors
                      ${
                        isActive
                          ? 'bg-zinc-100 text-zinc-950 font-semibold shadow-xs'
                          : 'text-zinc-400 hover:text-zinc-100 hover:bg-zinc-900'
                      }
                    `}
                  >
                    <Icon className="h-4 w-4 shrink-0" />
                    <span>{item.name}</span>
                  </Link>
                );
              })}
            </div>
          )}

          {/* Super Admin Console (RBAC: SUPER_ADMIN only) */}
          {isSuperAdmin && (
            <div className="space-y-1">
              <span className="px-3 text-[10px] font-bold uppercase tracking-wider text-zinc-500">
                Administration
              </span>
              <Link
                href="/admin"
                onClick={() => setIsMobileOpen(false)}
                className={`
                  flex items-center justify-between px-3 py-2 rounded-lg text-xs font-medium transition-colors
                  ${
                    pathname?.startsWith('/admin')
                      ? 'bg-rose-950/80 text-rose-200 border border-rose-800'
                      : 'text-rose-400 hover:bg-rose-950/40 hover:text-rose-300'
                  }
                `}
              >
                <div className="flex items-center gap-2.5">
                  <Shield className="h-4 w-4 shrink-0" />
                  <span>Admin Console</span>
                </div>
                <Badge variant="destructive" className="text-[9px] px-1.5 py-0">
                  Root
                </Badge>
              </Link>
            </div>
          )}
        </div>

        {/* User Profile & Session Footer */}
        <div className="p-3 border-t border-zinc-800/80 bg-zinc-950/50">
          <div className="flex items-center justify-between p-2 rounded-lg bg-zinc-900/60 border border-zinc-850">
            <div className="flex items-center gap-2.5 min-w-0">
              <Avatar className="h-8 w-8 rounded-md border border-zinc-800">
                <AvatarFallback className="bg-zinc-800 text-zinc-200 text-xs font-bold">
                  {userInitials}
                </AvatarFallback>
              </Avatar>
              <div className="flex flex-col min-w-0">
                <span className="truncate text-xs font-semibold text-zinc-200 leading-tight">
                  {user?.firstName && user?.lastName
                    ? `${user.firstName} ${user.lastName}`
                    : user?.username || 'User'}
                </span>
                <span className={`text-[10px] font-medium tracking-wide uppercase ${roleBadge.className}`}>
                  {roleBadge.label}
                </span>
              </div>
            </div>

            {/* Actions: Settings & Sign Out */}
            <div className="flex items-center gap-1 shrink-0">
              <Link
                href="/settings"
                className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
                title="Settings"
              >
                <Settings className="h-3.5 w-3.5" />
              </Link>
              <button
                type="button"
                onClick={logout}
                className="p-1.5 rounded-md text-zinc-400 hover:text-rose-400 hover:bg-zinc-800 transition-colors cursor-pointer"
                title="Sign Out"
              >
                <LogOut className="h-3.5 w-3.5" />
              </button>
            </div>
          </div>
        </div>
      </aside>
    </>
  );
}
