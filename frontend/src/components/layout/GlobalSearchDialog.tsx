'use client';

import * as React from 'react';
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  FolderGit2,
  ListTodo,
  Briefcase,
  Users,
  Settings,
  Shield,
  Layers,
  ArrowRight,
} from 'lucide-react';

import { useAppSelector } from '@/lib/store/hooks';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useGetWorkspacesQuery } from '@/lib/store/api/workspacesApi';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import {
  CommandDialog,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
  CommandSeparator,
} from '@/components/ui/command';

interface GlobalSearchDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function GlobalSearchDialog({ open, onOpenChange }: GlobalSearchDialogProps) {
  const router = useRouter();
  const { isSuperAdmin, isManager, isOwner } = useAuthRbac();
  const authState = useAppSelector((state) => state.auth);

  const activeOrgId = authState.activeOrganizationId;
  const activeWorkspaceId = authState.activeWorkspaceId;
  const activeWorkspaceSlug = authState.activeWorkspaceSlug || '';

  const { data: workspaces = [] } = useGetWorkspacesQuery(
    activeOrgId ? { organizationId: activeOrgId } : undefined,
    { skip: !activeOrgId }
  );

  const { data: projects = [] } = useGetProjectsQuery(
    activeWorkspaceId ? { workspaceId: activeWorkspaceId } : undefined,
    { skip: !activeWorkspaceId }
  );

  // Global Keyboard Shortcut: ⌘K or Ctrl+K
  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        onOpenChange(!open);
      }
    };

    document.addEventListener('keydown', down);
    return () => document.removeEventListener('keydown', down);
  }, [open, onOpenChange]);

  const handleSelect = (url: string) => {
    onOpenChange(false);
    router.push(url);
  };

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Global Search"
      description="Search across projects, workspaces, and system settings"
      className="bg-zinc-900 border-zinc-800 text-zinc-100 max-w-xl rounded-md"
    >
      <CommandInput
        placeholder="Type a command or search projects & workspaces... (⌘K)"
        className="text-xs h-11 border-b border-zinc-800 text-zinc-100 placeholder:text-zinc-500"
      />

      <CommandList className="max-h-80 p-1 overflow-y-auto">
        <CommandEmpty className="py-6 text-center text-xs text-zinc-500">
          No matching items or commands found.
        </CommandEmpty>

        {/* Projects in Active Workspace */}
        {projects.length > 0 && activeWorkspaceSlug && (
          <CommandGroup
            heading={
              <span className="text-[10px] font-bold uppercase tracking-wider text-zinc-500">
                Projects
              </span>
            }
          >
            {projects.map((proj) => (
              <CommandItem
                key={proj.id}
                value={`project-${proj.name}-${proj.slug}`}
                onSelect={() => handleSelect(`/${activeWorkspaceSlug}/${proj.slug}/tasks`)}
                className="flex items-center justify-between px-3 py-2 text-xs text-zinc-200 hover:bg-zinc-800 aria-selected:bg-zinc-800 rounded-sm cursor-pointer"
              >
                <div className="flex items-center gap-2.5 truncate">
                  <FolderGit2 className="h-4 w-4 text-zinc-400 shrink-0" />
                  <span className="font-medium truncate">{proj.name}</span>
                </div>
                <span className="text-[10px] text-zinc-500 font-mono">Tasks &rarr;</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {/* Workspaces */}
        {workspaces.length > 0 && (
          <>
            <CommandSeparator className="my-1 bg-zinc-800" />
            <CommandGroup
              heading={
                <span className="text-[10px] font-bold uppercase tracking-wider text-zinc-500">
                  Workspaces
                </span>
              }
            >
              {workspaces.map((ws) => (
                <CommandItem
                  key={ws.id}
                  value={`workspace-${ws.name}-${ws.slug}`}
                  onSelect={() => handleSelect(`/${ws.slug}`)}
                  className="flex items-center justify-between px-3 py-2 text-xs text-zinc-200 hover:bg-zinc-800 aria-selected:bg-zinc-800 rounded-sm cursor-pointer"
                >
                  <div className="flex items-center gap-2.5 truncate">
                    <Briefcase className="h-4 w-4 text-zinc-400 shrink-0" />
                    <span className="font-medium truncate">{ws.name}</span>
                  </div>
                  <span className="text-[10px] text-zinc-500 font-mono">Overview &rarr;</span>
                </CommandItem>
              ))}
            </CommandGroup>
          </>
        )}

        {/* Navigation & Tools */}
        <CommandSeparator className="my-1 bg-zinc-800" />
        <CommandGroup
          heading={
            <span className="text-[10px] font-bold uppercase tracking-wider text-zinc-500">
              Quick Navigation
            </span>
          }
        >
          {activeWorkspaceSlug && (
            <>
              <CommandItem
                value="nav-workspace-overview"
                onSelect={() => handleSelect(`/${activeWorkspaceSlug}`)}
                className="flex items-center gap-2 px-3 py-2 text-xs text-zinc-300 hover:bg-zinc-800 aria-selected:bg-zinc-800 rounded-sm cursor-pointer"
              >
                <Layers className="h-4 w-4 text-zinc-400" />
                <span>Workspace Dashboard</span>
              </CommandItem>
              <CommandItem
                value="nav-workspace-projects"
                onSelect={() => handleSelect(`/${activeWorkspaceSlug}/projects`)}
                className="flex items-center gap-2 px-3 py-2 text-xs text-zinc-300 hover:bg-zinc-800 aria-selected:bg-zinc-800 rounded-sm cursor-pointer"
              >
                <FolderGit2 className="h-4 w-4 text-zinc-400" />
                <span>All Workspace Projects</span>
              </CommandItem>
              <CommandItem
                value="nav-workspace-members"
                onSelect={() => handleSelect(`/${activeWorkspaceSlug}/members`)}
                className="flex items-center gap-2 px-3 py-2 text-xs text-zinc-300 hover:bg-zinc-800 aria-selected:bg-zinc-800 rounded-sm cursor-pointer"
              >
                <Users className="h-4 w-4 text-zinc-400" />
                <span>Team Members</span>
              </CommandItem>
            </>
          )}

          {(isManager || isOwner) && activeWorkspaceSlug && (
            <CommandItem
              value="nav-workspace-settings"
              onSelect={() => handleSelect(`/${activeWorkspaceSlug}/settings`)}
              className="flex items-center gap-2 px-3 py-2 text-xs text-zinc-300 hover:bg-zinc-800 aria-selected:bg-zinc-800 rounded-sm cursor-pointer"
            >
              <Settings className="h-4 w-4 text-zinc-400" />
              <span>Workspace Settings</span>
            </CommandItem>
          )}

          {isSuperAdmin && (
            <>
              <CommandItem
                value="nav-admin-console"
                onSelect={() => handleSelect('/admin')}
                className="flex items-center gap-2 px-3 py-2 text-xs text-rose-300 hover:bg-rose-950/40 aria-selected:bg-rose-950/40 rounded-sm cursor-pointer"
              >
                <Shield className="h-4 w-4 text-rose-400" />
                <span>Administration Console</span>
              </CommandItem>
              <CommandItem
                value="nav-admin-users"
                onSelect={() => handleSelect('/admin/users')}
                className="flex items-center gap-2 px-3 py-2 text-xs text-rose-300 hover:bg-rose-950/40 aria-selected:bg-rose-950/40 rounded-sm cursor-pointer"
              >
                <Users className="h-4 w-4 text-rose-400" />
                <span>Global User Directory</span>
              </CommandItem>
            </>
          )}
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  );
}

