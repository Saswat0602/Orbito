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
  ChevronDown,
  Menu,
  X,
  Plus,
  Building2,
  Briefcase,
  Check,
  Activity,
} from 'lucide-react';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import {
  setActiveOrganization,
  setActiveWorkspace,
  setActiveWorkspaceSlug,
  setActiveProject,
  setActiveProjectSlug,
} from '@/lib/store/slices/authSlice';
import { useGetOrganizationsQuery } from '@/lib/store/api/organizationsApi';
import { useGetWorkspacesQuery } from '@/lib/store/api/workspacesApi';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import { Role } from '@/lib/rbac';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';

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

  // Fetch Tenancy hierarchy
  const { data: organizations = [] } = useGetOrganizationsQuery();

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

  const activeProject = useMemo(() => {
    return (
      projects.find((p) => p.id === authState.activeProjectId) ||
      projects.find((p) => p.slug === authState.activeProjectSlug) ||
      projects[0] ||
      null
    );
  }, [projects, authState.activeProjectId, authState.activeProjectSlug]);

  const activeProjectSlug = activeProject?.slug || authState.activeProjectSlug || '';

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

  // Main Navigation Items scoped by Workspace / Project
  const navigation = useMemo(() => {
    if (!activeWorkspaceSlug) {
      return [{ name: 'Overview', href: '/', icon: LayoutDashboard, exact: true }];
    }

    const baseWorkspaceUrl = `/${activeWorkspaceSlug}`;
    const tasksUrl = activeProjectSlug
      ? `${baseWorkspaceUrl}/${activeProjectSlug}/tasks`
      : `${baseWorkspaceUrl}/projects`;
    const sprintsUrl = activeProjectSlug
      ? `${baseWorkspaceUrl}/${activeProjectSlug}/sprints`
      : `${baseWorkspaceUrl}/projects`;

    return [
      { name: 'Dashboard', href: baseWorkspaceUrl, icon: LayoutDashboard, exact: true },
      { name: 'Projects', href: `${baseWorkspaceUrl}/projects`, icon: FolderGit2, exact: false },
      { name: 'Tasks', href: tasksUrl, icon: ListTodo, exact: false },
      { name: 'Sprints', href: sprintsUrl, icon: Target, exact: false },
      { name: 'Members', href: `${baseWorkspaceUrl}/members`, icon: Users, exact: false },
      { name: 'Activity Log', href: `${baseWorkspaceUrl}/activities`, icon: Activity, exact: false },
    ];
  }, [activeWorkspaceSlug, activeProjectSlug]);

  // RBAC Management Links (Manager and above)
  const managementLinks = useMemo(() => {
    if (!activeWorkspaceSlug) return [];
    const baseWorkspaceUrl = `/${activeWorkspaceSlug}`;
    return [
      { name: 'Workspace Settings', href: `${baseWorkspaceUrl}/settings`, icon: Settings },
    ];
  }, [activeWorkspaceSlug]);

  const handleSelectWorkspace = (ws: typeof workspaces[0]) => {
    dispatch(setActiveWorkspace(ws.id));
    dispatch(setActiveWorkspaceSlug(ws.slug));
    router.push(`/${ws.slug}`);
  };

  const handleSelectProject = (proj: typeof projects[0]) => {
    dispatch(setActiveProject(proj.id));
    dispatch(setActiveProjectSlug(proj.slug));
    if (activeWorkspaceSlug) {
      router.push(`/${activeWorkspaceSlug}/${proj.slug}/tasks`);
    }
  };

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
        {/* Brand & Organization */}
        <div className="p-4 border-b border-zinc-800/80 space-y-3">
          <div className="flex items-center gap-2.5">
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

          {/* Interactive Workspace Switcher Dropdown */}
          <div className="space-y-1.5">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type="button"
                  className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-md bg-zinc-900 border border-zinc-800 text-xs text-left hover:border-zinc-700 transition-colors cursor-pointer"
                >
                  <div className="flex items-center gap-2 truncate">
                    <Briefcase className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
                    <span className="truncate text-zinc-200 font-medium">
                      {activeWorkspaceName}
                    </span>
                  </div>
                  <ChevronDown className="h-3 w-3 text-zinc-500 shrink-0" />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="w-56 bg-zinc-900 border-zinc-800">
                <DropdownMenuLabel className="text-[10px] uppercase text-zinc-500 font-semibold">
                  Workspaces ({activeOrgName})
                </DropdownMenuLabel>
                <DropdownMenuSeparator className="bg-zinc-800" />
                {workspaces.map((ws) => (
                  <DropdownMenuItem
                    key={ws.id}
                    onClick={() => handleSelectWorkspace(ws)}
                    className="flex items-center justify-between text-xs text-zinc-200 hover:bg-zinc-800 cursor-pointer"
                  >
                    <span className="truncate">{ws.name}</span>
                    {ws.id === activeWorkspaceId && (
                      <Check className="h-3.5 w-3.5 text-zinc-400" />
                    )}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>

            {/* Interactive Project Switcher Dropdown */}
            {projects.length > 0 && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button
                    type="button"
                    className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-md bg-zinc-900/60 border border-zinc-850 text-xs text-left hover:border-zinc-700 transition-colors cursor-pointer"
                  >
                    <div className="flex items-center gap-2 truncate">
                      <FolderGit2 className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
                      <span className="truncate text-zinc-300">
                        {activeProjectName}
                      </span>
                    </div>
                    <ChevronDown className="h-3 w-3 text-zinc-500 shrink-0" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="w-56 bg-zinc-900 border-zinc-800">
                  <DropdownMenuLabel className="text-[10px] uppercase text-zinc-500 font-semibold">
                    Projects
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator className="bg-zinc-800" />
                  {projects.map((proj) => (
                    <DropdownMenuItem
                      key={proj.id}
                      onClick={() => handleSelectProject(proj)}
                      className="flex items-center justify-between text-xs text-zinc-200 hover:bg-zinc-800 cursor-pointer"
                    >
                      <span className="truncate">{proj.name}</span>
                      {proj.id === activeProject?.id && (
                        <Check className="h-3.5 w-3.5 text-zinc-400" />
                      )}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            )}
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

            {/* Actions: Sign Out */}
            <div className="flex items-center gap-1 shrink-0">
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
