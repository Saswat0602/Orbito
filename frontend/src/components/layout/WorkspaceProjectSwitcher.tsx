'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import {
  ChevronsUpDown,
  Check,
  Plus,
  Briefcase,
  FolderGit2,
  Building2,
  Layers,
} from 'lucide-react';

import { useAppDispatch } from '@/lib/store/hooks';
import {
  setActiveWorkspace,
  setActiveWorkspaceSlug,
  setActiveProject,
  setActiveProjectSlug,
  setActiveOrganization,
} from '@/lib/store/slices/authSlice';
import type { Organization, Workspace, Project } from '@orbito/types';
import {
  Popover,
  PopoverTrigger,
  PopoverContent,
} from '@/components/ui/popover';
import {
  Command,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
  CommandSeparator,
} from '@/components/ui/command';

interface WorkspaceProjectSwitcherProps {
  organizations: Organization[];
  workspaces: Workspace[];
  projects: Project[];
  activeOrgId: string | null;
  activeWorkspace: Workspace | null;
  activeProject: Project | null;
  activeOrgName: string;
  activeWorkspaceName: string;
  activeProjectName: string;
  onItemSelect?: () => void;
}

export function WorkspaceProjectSwitcher({
  organizations,
  workspaces,
  projects,
  activeOrgId,
  activeWorkspace,
  activeProject,
  activeOrgName,
  activeWorkspaceName,
  activeProjectName,
  onItemSelect,
}: WorkspaceProjectSwitcherProps) {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const [open, setOpen] = useState(false);

  const activeWorkspaceSlug = activeWorkspace?.slug || '';

  // Workspace monogram for the icon
  const workspaceMonogram = useMemo(() => {
    if (!activeWorkspace?.name) return 'W';
    const parts = activeWorkspace.name.trim().split(/\s+/);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return activeWorkspace.name.slice(0, 2).toUpperCase();
  }, [activeWorkspace]);

  const handleSelectProject = (proj: Project | null) => {
    if (proj) {
      dispatch(setActiveProject(proj.id));
      dispatch(setActiveProjectSlug(proj.slug));
      setOpen(false);
      onItemSelect?.();

      if (activeWorkspaceSlug) {
        router.push(`/${activeWorkspaceSlug}/${proj.slug}/tasks`);
      }
    } else {
      dispatch(setActiveProject(null));
      dispatch(setActiveProjectSlug(null));
      setOpen(false);
      onItemSelect?.();

      if (activeWorkspaceSlug) {
        router.push(`/${activeWorkspaceSlug}`);
      }
    }
  };

  const handleSelectWorkspace = (ws: Workspace) => {
    dispatch(setActiveWorkspace(ws.id));
    dispatch(setActiveWorkspaceSlug(ws.slug));
    // Clear project selection when switching to a different workspace
    dispatch(setActiveProject(null));
    dispatch(setActiveProjectSlug(null));
    setOpen(false);
    onItemSelect?.();

    router.push(`/${ws.slug}`);
  };

  const handleSelectOrganization = (org: Organization) => {
    dispatch(setActiveOrganization(org.id));
    dispatch(setActiveWorkspace(null));
    dispatch(setActiveWorkspaceSlug(null));
    dispatch(setActiveProject(null));
    dispatch(setActiveProjectSlug(null));
    setOpen(false);
    onItemSelect?.();

    router.push('/');
  };

  const projectDisplayText = useMemo(() => {
    if (activeProject?.name) return activeProject.name;
    return 'All Projects';
  }, [activeProject]);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-expanded={open}
          aria-label="Select workspace and project"
          className="w-full group flex items-center justify-between p-2 rounded-md bg-zinc-900/80 hover:bg-zinc-850 border border-zinc-800 hover:border-zinc-700 text-left transition-colors cursor-pointer shadow-xs focus:outline-none focus:ring-1 focus:ring-zinc-600"
        >
          <div className="flex items-center gap-2.5 min-w-0">
            {/* Workspace Icon Monogram */}
            <div className="h-7 w-7 rounded-sm bg-gradient-to-br from-zinc-800 to-zinc-900 border border-zinc-700/80 flex items-center justify-center text-zinc-200 font-bold text-[11px] shrink-0 shadow-inner">
              {workspaceMonogram}
            </div>

            {/* Hierarchical Text */}
            <div className="flex flex-col min-w-0">
              <span className="text-[10px] font-semibold text-zinc-400 group-hover:text-zinc-300 uppercase tracking-wider truncate leading-tight">
                {activeWorkspaceName}
              </span>
              <span className="text-xs font-bold text-zinc-100 group-hover:text-white truncate leading-tight">
                {projectDisplayText}
              </span>
            </div>
          </div>

          <ChevronsUpDown className="h-3.5 w-3.5 text-zinc-500 group-hover:text-zinc-300 shrink-0 ml-1.5 transition-colors" />
        </button>
      </PopoverTrigger>

      <PopoverContent
        align="start"
        sideOffset={6}
        className="w-72 p-0 bg-zinc-900 border-zinc-800 text-zinc-100 shadow-xl rounded-md"
      >
        <Command className="bg-transparent">
          <CommandInput
            placeholder="Search workspace or project..."
            className="text-xs h-9 border-b border-zinc-800 text-zinc-200 placeholder:text-zinc-500"
          />

          <CommandList className="max-h-72 p-1 overflow-y-auto">
            <CommandEmpty className="py-4 text-center text-xs text-zinc-500">
              No matching items found.
            </CommandEmpty>

            {/* Current Workspace Projects */}
            <CommandGroup
              heading={
                <span className="text-[10px] font-bold uppercase tracking-wider text-zinc-500">
                  Projects ({activeWorkspaceName})
                </span>
              }
            >
              {/* Option to view all projects / Workspace Overview */}
              <CommandItem
                value="proj-all-workspace-overview"
                onSelect={() => handleSelectProject(null)}
                className="flex items-center justify-between px-2.5 py-1.5 rounded-sm text-xs cursor-pointer text-zinc-200 hover:bg-zinc-800 aria-selected:bg-zinc-800"
              >
                <div className="flex items-center gap-2 truncate">
                  <Layers className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
                  <span className="truncate font-medium">All Projects (Overview)</span>
                </div>
                {!activeProject && <Check className="h-3.5 w-3.5 text-zinc-100 shrink-0" />}
              </CommandItem>

              {projects.map((proj) => {
                const isSelected = proj.id === activeProject?.id;
                return (
                  <CommandItem
                    key={proj.id}
                    value={`proj-${proj.name}-${proj.slug}`}
                    onSelect={() => handleSelectProject(proj)}
                    className="flex items-center justify-between px-2.5 py-1.5 rounded-sm text-xs cursor-pointer text-zinc-200 hover:bg-zinc-800 aria-selected:bg-zinc-800"
                  >
                    <div className="flex items-center gap-2 truncate">
                      <FolderGit2 className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
                      <span className="truncate font-medium">{proj.name}</span>
                    </div>
                    {isSelected && <Check className="h-3.5 w-3.5 text-zinc-100 shrink-0" />}
                  </CommandItem>
                );
              })}
            </CommandGroup>

            <CommandSeparator className="my-1 bg-zinc-800" />

            {/* Workspaces List */}
            <CommandGroup
              heading={
                <span className="text-[10px] font-bold uppercase tracking-wider text-zinc-500">
                  Workspaces ({activeOrgName})
                </span>
              }
            >
              {workspaces.map((ws) => {
                const isSelected = ws.id === activeWorkspace?.id;
                return (
                  <CommandItem
                    key={ws.id}
                    value={`ws-${ws.name}-${ws.slug}`}
                    onSelect={() => handleSelectWorkspace(ws)}
                    className="flex items-center justify-between px-2.5 py-1.5 rounded-sm text-xs cursor-pointer text-zinc-200 hover:bg-zinc-800 aria-selected:bg-zinc-800"
                  >
                    <div className="flex items-center gap-2 truncate">
                      <Briefcase className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
                      <span className="truncate font-medium">{ws.name}</span>
                    </div>
                    {isSelected && <Check className="h-3.5 w-3.5 text-zinc-100 shrink-0" />}
                  </CommandItem>
                );
              })}
            </CommandGroup>

            {/* Multi-Tenant Organization Switcher (if > 1 org) */}
            {organizations.length > 1 && (
              <>
                <CommandSeparator className="my-1 bg-zinc-800" />
                <CommandGroup
                  heading={
                    <span className="text-[10px] font-bold uppercase tracking-wider text-zinc-500">
                      Switch Organization
                    </span>
                  }
                >
                  {organizations.map((org) => {
                    const isSelected = org.id === activeOrgId;
                    return (
                      <CommandItem
                        key={org.id}
                        value={`org-${org.name}-${org.slug}`}
                        onSelect={() => handleSelectOrganization(org)}
                        className="flex items-center justify-between px-2.5 py-1.5 rounded-sm text-xs cursor-pointer text-zinc-200 hover:bg-zinc-800 aria-selected:bg-zinc-800"
                      >
                        <div className="flex items-center gap-2 truncate">
                          <Building2 className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
                          <span className="truncate font-medium">{org.name}</span>
                        </div>
                        {isSelected && <Check className="h-3.5 w-3.5 text-zinc-100 shrink-0" />}
                      </CommandItem>
                    );
                  })}
                </CommandGroup>
              </>
            )}

            {/* Quick Actions */}
            <CommandSeparator className="my-1 bg-zinc-800" />
            <CommandGroup
              heading={
                <span className="text-[10px] font-bold uppercase tracking-wider text-zinc-500">
                  Quick Actions
                </span>
              }
            >
              {activeWorkspaceSlug && (
                <CommandItem
                  value="action-new-project"
                  onSelect={() => {
                    setOpen(false);
                    onItemSelect?.();
                    router.push(`/${activeWorkspaceSlug}/projects`);
                  }}
                  className="flex items-center gap-2 px-2.5 py-1.5 rounded-sm text-xs text-zinc-300 hover:bg-zinc-800 aria-selected:bg-zinc-800 cursor-pointer"
                >
                  <Plus className="h-3.5 w-3.5 text-zinc-400" />
                  <span>New Project...</span>
                </CommandItem>
              )}
              <CommandItem
                value="action-all-workspaces"
                onSelect={() => {
                  setOpen(false);
                  onItemSelect?.();
                  router.push('/');
                }}
                className="flex items-center gap-2 px-2.5 py-1.5 rounded-sm text-xs text-zinc-300 hover:bg-zinc-800 aria-selected:bg-zinc-800 cursor-pointer"
              >
                <Layers className="h-3.5 w-3.5 text-zinc-400" />
                <span>All Workspaces Overview</span>
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
