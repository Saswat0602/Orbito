'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import {
  LayoutDashboard,
  ListTodo,
  Target,
  FolderGit2,
  Users,
  Shield,
  Settings,
  LogOut,
  Menu,
  X,
  Building2,
  Activity,
  Sliders,
  Cpu,
  Layers,
  ArrowLeft,
} from 'lucide-react';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { useGetOrganizationsQuery } from '@/lib/store/api/organizationsApi';
import {
  useGetWorkspacesQuery,
  useGetWorkspaceMembersQuery,
} from '@/lib/store/api/workspacesApi';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { WorkspaceProjectSwitcher } from './WorkspaceProjectSwitcher';

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
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

  // Fetch Tenancy hierarchy only when user session is active
  const { data: organizations = [] } = useGetOrganizationsQuery(undefined, {
    skip: !authState.isAuthenticated || !authState.user,
  });

  const activeOrgId = authState.activeOrganizationId || organizations[0]?.id || null;

  const { data: workspaces = [] } = useGetWorkspacesQuery(
    activeOrgId ? { organizationId: activeOrgId } : undefined,
    { skip: !activeOrgId }
  );

  const activeWorkspace = useMemo(() => {
    return (
      workspaces.find((w) => w.id === authState.activeWorkspaceId) ||
      workspaces.find((w) => w.slug === authState.activeWorkspaceSlug) ||
      workspaces[0] ||
      null
    );
  }, [workspaces, authState.activeWorkspaceId, authState.activeWorkspaceSlug]);

  const activeWorkspaceId = activeWorkspace?.id || null;
  const activeWorkspaceSlug = activeWorkspace?.slug || authState.activeWorkspaceSlug || '';

  const { data: projects = [] } = useGetProjectsQuery(
    activeWorkspaceId ? { workspaceId: activeWorkspaceId } : undefined,
    { skip: !activeWorkspaceId }
  );

  const { data: workspaceMembers = [] } = useGetWorkspaceMembersQuery(
    activeWorkspaceId || '',
    { skip: !activeWorkspaceId }
  );

  // Active project is only considered active if explicitly set in state
  const activeProject = useMemo(() => {
    if (!authState.activeProjectId && !authState.activeProjectSlug) return null;
    return (
      projects.find((p) => p.id === authState.activeProjectId) ||
      projects.find((p) => p.slug === authState.activeProjectSlug) ||
      null
    );
  }, [projects, authState.activeProjectId, authState.activeProjectSlug]);

  const activeProjectSlug = activeProject?.slug || '';

  // Active names
  const activeOrgName = useMemo(() => {
    return organizations.find((o) => o.id === activeOrgId)?.name || 'Default Organization';
  }, [organizations, activeOrgId]);

  const activeWorkspaceName = useMemo(() => {
    return activeWorkspace?.name || 'Select Workspace';
  }, [activeWorkspace]);

  const activeProjectName = useMemo(() => {
    return activeProject?.name || 'All Projects';
  }, [activeProject]);

  // Context-Aware Navigation Items
  const projectNavigation = useMemo(() => {
    if (!activeWorkspaceSlug || !activeProjectSlug) return [];
    const baseProjectUrl = `/${activeWorkspaceSlug}/${activeProjectSlug}`;

    const items = [
      { name: 'Tasks & Board', href: `${baseProjectUrl}/tasks`, icon: ListTodo, exact: false },
      { name: 'Sprints & Backlog', href: `${baseProjectUrl}/sprints`, icon: Target, exact: false },
    ];

    if (isManager || isOwner) {
      items.push({
        name: 'Project Settings',
        href: `${baseProjectUrl}/settings`,
        icon: Settings,
        exact: false,
      });
    }

    return items;
  }, [activeWorkspaceSlug, activeProjectSlug, isManager, isOwner]);

  const workspaceNavigation = useMemo(() => {
    if (!activeWorkspaceSlug) {
      return [{ name: 'Overview', href: '/', icon: LayoutDashboard, exact: true }];
    }

    const baseWorkspaceUrl = `/${activeWorkspaceSlug}`;

    return [
      { name: 'Dashboard', href: baseWorkspaceUrl, icon: LayoutDashboard, exact: true },
      {
        name: 'Projects Directory',
        href: `${baseWorkspaceUrl}/projects`,
        icon: FolderGit2,
        exact: false,
        badge: projects.length > 0 ? projects.length : undefined,
      },
      {
        name: 'Members',
        href: `${baseWorkspaceUrl}/members`,
        icon: Users,
        exact: false,
        badge: workspaceMembers.length > 0 ? workspaceMembers.length : undefined,
      },
      { name: 'Activity Log', href: `${baseWorkspaceUrl}/activities`, icon: Activity, exact: false },
    ];
  }, [activeWorkspaceSlug, projects.length, workspaceMembers.length]);

  // RBAC Management Links (Manager and above)
  const managementLinks = useMemo(() => {
    if (!activeWorkspaceSlug) return [];
    const baseWorkspaceUrl = `/${activeWorkspaceSlug}`;
    return [
      { name: 'Workspace Settings', href: `${baseWorkspaceUrl}/settings`, icon: Settings },
    ];
  }, [activeWorkspaceSlug]);

  // RBAC Administration Links (Super Admin only)
  const adminLinks = useMemo(() => {
    return [
      { name: 'System Overview', href: '/admin', icon: Shield, exact: true, badge: 'Root' },
      { name: 'User Directory', href: '/admin/users', icon: Users, exact: false },
      { name: 'Organizations', href: '/admin/organizations', icon: Building2, exact: false },
      { name: 'System & SSO', href: '/admin/settings', icon: Sliders, exact: false },
      { name: 'Queue & Workers', href: '/admin/queues', icon: Cpu, exact: false, badge: 'BullMQ' },
    ];
  }, []);

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
        className="md:hidden fixed top-3 left-3 z-50 p-2 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-300 shadow-md cursor-pointer"
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
        {/* Brand & Unified Hierarchy Switcher Header */}
        <div className="p-4 border-b border-zinc-800/80 space-y-3">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-md bg-zinc-100 text-zinc-950 font-bold shadow-sm">
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

          {/* Unified High-Craft Workspace & Project Switcher */}
          <WorkspaceProjectSwitcher
            organizations={organizations}
            workspaces={workspaces}
            projects={projects}
            activeOrgId={activeOrgId}
            activeWorkspace={activeWorkspace}
            activeProject={activeProject}
            activeOrgName={activeOrgName}
            activeWorkspaceName={activeWorkspaceName}
            activeProjectName={activeProjectName}
            onItemSelect={() => setIsMobileOpen(false)}
          />
        </div>

        {/* Navigation Links Scrollable Body */}
        <div className="flex-1 overflow-y-auto px-3 py-4 space-y-6">
          {/* Case 1: Specific Project Selected -> Show Focused Project Tools */}
          {activeProject && projectNavigation.length > 0 && (
            <div className="space-y-1">
              <div className="flex items-center justify-between px-3 pb-1">
                <span className="text-[10px] font-bold uppercase tracking-wider text-zinc-400 truncate">
                  {activeProject.name}
                </span>
                <span className="text-[9px] font-mono text-zinc-500 uppercase">Project</span>
              </div>
              {projectNavigation.map((item) => {
                const isActive = pathname?.startsWith(item.href);
                const Icon = item.icon;

                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    onClick={() => setIsMobileOpen(false)}
                    className={`
                      flex items-center gap-2.5 px-3 py-2 rounded-md text-xs font-medium transition-colors
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

          {/* Workspace Links Group */}
          <div className="space-y-1">
            <span className="px-3 text-[10px] font-bold uppercase tracking-wider text-zinc-500">
              {activeProject ? 'Workspace Context' : 'Workspace'}
            </span>
            {workspaceNavigation.map((item) => {
              const isActive = item.exact
                ? pathname === item.href
                : pathname === item.href || (item.href !== '/' && pathname?.startsWith(item.href));
              const Icon = item.icon;

              return (
                <Link
                  key={item.name}
                  href={item.href}
                  onClick={() => setIsMobileOpen(false)}
                  className={`
                    flex items-center justify-between px-3 py-2 rounded-md text-xs font-medium transition-colors
                    ${
                      isActive
                        ? 'bg-zinc-100 text-zinc-950 font-semibold shadow-xs'
                        : 'text-zinc-400 hover:text-zinc-100 hover:bg-zinc-900'
                    }
                  `}
                >
                  <div className="flex items-center gap-2.5 truncate">
                    <Icon className="h-4 w-4 shrink-0" />
                    <span className="truncate">{item.name}</span>
                  </div>
                  {item.badge !== undefined && (
                    <span
                      className={`text-[10px] px-1.5 py-0.2 rounded-sm font-mono font-medium ${
                        isActive
                          ? 'bg-zinc-200 text-zinc-900 font-bold'
                          : 'bg-zinc-800 text-zinc-400'
                      }`}
                    >
                      {item.badge}
                    </span>
                  )}
                </Link>
              );
            })}
          </div>

          {/* Management Links (RBAC: Manager & Owner only) */}
          {(isManager || isOwner) && managementLinks.length > 0 && (
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
                      flex items-center gap-2.5 px-3 py-2 rounded-md text-xs font-medium transition-colors
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
              <div className="flex items-center justify-between px-3">
                <span className="text-[10px] font-bold uppercase tracking-wider text-zinc-500">
                  Administration
                </span>
                <span className="text-[9px] font-semibold tracking-wider text-rose-500 bg-rose-950/60 border border-rose-900/60 px-1 py-0.2 rounded-sm">
                  ROOT
                </span>
              </div>
              {adminLinks.map((item) => {
                const isActive = item.exact
                  ? pathname === item.href
                  : pathname?.startsWith(item.href);
                const Icon = item.icon;

                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    onClick={() => setIsMobileOpen(false)}
                    className={`
                      flex items-center justify-between px-3 py-2 rounded-md text-xs font-medium transition-colors
                      ${
                        isActive
                          ? 'bg-rose-950/80 text-rose-200 border border-rose-800 shadow-xs'
                          : 'text-zinc-400 hover:text-zinc-100 hover:bg-zinc-900'
                      }
                    `}
                  >
                    <div className="flex items-center gap-2.5 truncate">
                      <Icon className="h-4 w-4 shrink-0" />
                      <span className="truncate">{item.name}</span>
                    </div>
                    {item.badge && (
                      <span
                        className={`text-[9px] px-1 py-0.2 rounded-sm font-semibold uppercase tracking-wider ${
                          isActive
                            ? 'bg-rose-900/90 text-rose-100'
                            : 'bg-zinc-800 text-zinc-400'
                        }`}
                      >
                        {item.badge}
                      </span>
                    )}
                  </Link>
                );
              })}
            </div>
          )}
        </div>

        {/* User Profile & Session Footer */}
        <div className="p-3 border-t border-zinc-800/80 bg-zinc-950/50">
          <div className="flex items-center justify-between p-2 rounded-md bg-zinc-900/60 border border-zinc-850">
            <Link href="/settings" className="flex items-center gap-2.5 min-w-0 group cursor-pointer">
              <Avatar className="h-8 w-8 rounded-md border border-zinc-800 group-hover:border-zinc-700 transition-colors">
                <AvatarFallback className="bg-zinc-800 text-zinc-200 text-xs font-bold">
                  {userInitials}
                </AvatarFallback>
              </Avatar>
              <div className="flex flex-col min-w-0">
                <span className="truncate text-xs font-semibold text-zinc-200 group-hover:text-white transition-colors leading-tight">
                  {user?.firstName && user?.lastName
                    ? `${user.firstName} ${user.lastName}`
                    : user?.username || 'User'}
                </span>
                <span className={`text-[10px] font-medium tracking-wide uppercase ${roleBadge.className}`}>
                  {roleBadge.label}
                </span>
              </div>
            </Link>

            {/* Actions: Settings & Sign Out */}
            <div className="flex items-center gap-1 shrink-0">
              <Link href="/settings">
                <button
                  type="button"
                  className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors cursor-pointer"
                  title="Account Settings"
                >
                  <Settings className="h-3.5 w-3.5" />
                </button>
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
