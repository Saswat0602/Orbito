'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Search,
  UserPlus,
  ChevronRight,
  FolderGit2,
  Briefcase,
  Shield,
  Layers,
  Sparkles,
} from 'lucide-react';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppSelector, useAppDispatch } from '@/lib/store/hooks';
import { toggleAiChat } from '@/lib/store/slices/uiSlice';
import { useGetWorkspacesQuery } from '@/lib/store/api/workspacesApi';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ThemeToggle } from './ThemeToggle';
import { NotificationPopover } from './NotificationPopover';
import { GlobalSearchDialog } from './GlobalSearchDialog';
import { InviteMemberModal } from '@/components/modals/InviteMemberModal';

export function Navbar() {
  const pathname = usePathname();
  const dispatch = useAppDispatch();
  const { user, isManager, isOwner, isSuperAdmin, roleBadge } = useAuthRbac();
  const authState = useAppSelector((state) => state.auth);

  const [searchOpen, setSearchOpen] = useState(false);
  const [inviteOpen, setInviteOpen] = useState(false);

  const activeOrgId = authState.activeOrganizationId;
  const activeWorkspaceId = authState.activeWorkspaceId;
  const activeWorkspaceSlug = authState.activeWorkspaceSlug || '';

  const { data: workspaces = [] } = useGetWorkspacesQuery(
    activeOrgId ? { organizationId: activeOrgId } : undefined,
    { skip: !activeOrgId }
  );

  const activeWorkspace = useMemo(() => {
    return (
      workspaces.find((w) => w.id === activeWorkspaceId) ||
      workspaces.find((w) => w.slug === activeWorkspaceSlug) ||
      workspaces[0] ||
      null
    );
  }, [workspaces, activeWorkspaceId, activeWorkspaceSlug]);

  const { data: projects = [] } = useGetProjectsQuery(
    activeWorkspace?.id ? { workspaceId: activeWorkspace.id } : undefined,
    { skip: !activeWorkspace?.id }
  );

  const activeProject = useMemo(() => {
    if (!authState.activeProjectId && !authState.activeProjectSlug) return null;
    return (
      projects.find((p) => p.id === authState.activeProjectId) ||
      projects.find((p) => p.slug === authState.activeProjectSlug) ||
      null
    );
  }, [projects, authState.activeProjectId, authState.activeProjectSlug]);

  // Is Admin Route?
  const isAdminRoute = pathname?.startsWith('/admin');

  return (
    <>
      <header className="h-13 bg-[#09090b]/85 backdrop-blur-md border-b border-zinc-800/80 px-4 md:px-6 flex items-center justify-between shrink-0 z-30 sticky top-0">
        {/* Left Section: Breadcrumbs */}
        <div className="flex items-center gap-2 min-w-0">
          {isAdminRoute ? (
            <div className="flex items-center gap-2 text-xs">
              <div className="flex items-center gap-1.5 text-rose-400 font-semibold">
                <Shield className="h-3.5 w-3.5" />
                <span>Administration</span>
              </div>
              <ChevronRight className="h-3 w-3 text-zinc-600" />
              <span className="text-zinc-200 font-medium truncate">
                {pathname === '/admin'
                  ? 'System Overview'
                  : pathname === '/admin/users'
                  ? 'User Directory'
                  : pathname === '/admin/organizations'
                  ? 'Organizations'
                  : pathname === '/admin/settings'
                  ? 'System Settings'
                  : pathname === '/admin/queues'
                  ? 'Queues & Workers'
                  : 'Root Console'}
              </span>
              <Badge variant="destructive" className="text-[9px] py-0 px-1 ml-1">
                Root
              </Badge>
            </div>
          ) : (
            <div className="flex items-center gap-1.5 text-xs min-w-0">
              {/* Workspace Link */}
              {activeWorkspace ? (
                <Link
                  href={`/${activeWorkspace.slug}`}
                  className="flex items-center gap-1.5 text-zinc-400 hover:text-zinc-100 font-medium transition-colors truncate max-w-[140px] sm:max-w-[180px]"
                >
                  <Briefcase className="h-3.5 w-3.5 text-zinc-500 shrink-0" />
                  <span className="truncate">{activeWorkspace.name}</span>
                </Link>
              ) : (
                <span className="text-zinc-400 font-medium">Workspace</span>
              )}

              {/* Project or Overview Indicator */}
              <ChevronRight className="h-3 w-3 text-zinc-600 shrink-0" />

              {activeProject ? (
                <div className="flex items-center gap-1.5 text-zinc-100 font-semibold truncate max-w-[140px] sm:max-w-[200px]">
                  <FolderGit2 className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
                  <span className="truncate">{activeProject.name}</span>
                </div>
              ) : (
                <div className="flex items-center gap-1.5 text-zinc-300 font-medium truncate">
                  <Layers className="h-3.5 w-3.5 text-zinc-500 shrink-0" />
                  <span>Overview</span>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Center Section: Global Command Search Trigger */}
        <div className="hidden sm:flex items-center justify-center flex-1 max-w-sm mx-4">
          <button
            type="button"
            onClick={() => setSearchOpen(true)}
            className="w-full flex items-center justify-between px-3 py-1.5 rounded-md bg-zinc-900/60 hover:bg-zinc-850 border border-zinc-800 text-xs text-zinc-400 hover:text-zinc-200 transition-colors cursor-pointer shadow-2xs"
          >
            <div className="flex items-center gap-2 truncate">
              <Search className="h-3.5 w-3.5 text-zinc-500" />
              <span className="truncate">Search tasks, projects, commands...</span>
            </div>
            <kbd className="hidden md:inline-flex items-center gap-0.5 px-1.5 py-0.5 text-[10px] font-mono text-zinc-400 bg-zinc-800 border border-zinc-700/60 rounded">
              ⌘K
            </kbd>
          </button>
        </div>

        {/* Right Section: Actions (Copilot, Invite, Notifications, Theme, User) */}
        <div className="flex items-center gap-2 shrink-0">
          {/* AI Copilot Quick Trigger */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => dispatch(toggleAiChat())}
            className="h-8 text-xs border-violet-900/50 bg-violet-950/20 hover:bg-violet-900/40 text-violet-300 font-medium cursor-pointer flex items-center gap-1.5"
            title="Orbito AI Copilot (⌘J)"
          >
            <Sparkles className="h-3.5 w-3.5 text-violet-400" />
            <span className="hidden md:inline">Copilot</span>
            <kbd className="hidden lg:inline-flex items-center text-[10px] font-mono text-violet-400/80 bg-violet-950/60 border border-violet-800/40 px-1 rounded-xs">
              ⌘J
            </kbd>
          </Button>

          {/* Mobile search trigger */}
          <button
            type="button"
            onClick={() => setSearchOpen(true)}
            className="sm:hidden p-2 rounded-md bg-zinc-900/60 hover:bg-zinc-800 border border-zinc-800 text-zinc-400 hover:text-zinc-200"
            aria-label="Open search"
          >
            <Search className="h-4 w-4" />
          </button>

          {/* Invite Members Action (RBAC: Manager or Owner only) */}
          {(isManager || isOwner) && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setInviteOpen(true)}
              className="h-8 text-xs border-zinc-800 bg-zinc-900/60 hover:bg-zinc-800 text-zinc-200 font-medium cursor-pointer"
            >
              <UserPlus className="h-3.5 w-3.5 mr-1.5 text-zinc-400" />
              <span>Invite</span>
            </Button>
          )}

          {/* Notifications Popover */}
          <NotificationPopover />

          {/* Theme Toggle (Dark / Light) */}
          <ThemeToggle />
        </div>
      </header>

      {/* Global Command Search Dialog (⌘K) */}
      <GlobalSearchDialog open={searchOpen} onOpenChange={setSearchOpen} />

      {/* Invite Members Modal */}
      <InviteMemberModal
        open={inviteOpen}
        onOpenChange={setInviteOpen}
        workspaceId={activeWorkspace?.id}
        workspaceName={activeWorkspace?.name}
        organizationId={activeOrgId}
      />
    </>
  );
}

