'use client';

import * as React from 'react';
import { useState, useMemo, useEffect } from 'react';
import Link from 'next/link';
import {
  Briefcase,
  Search,
  Plus,
  Archive,
  ArchiveRestore,
  ExternalLink,
  Users,
  FolderGit2,
  Calendar,
  AlertCircle,
  MoreVertical,
  Building2,
  CheckCircle2,
  Loader2,
} from 'lucide-react';
import { toast } from 'sonner';

import { useAppSelector } from '@/lib/store/hooks';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import {
  useGetWorkspacesQuery,
  useGetArchivedWorkspacesQuery,
  useCreateWorkspaceMutation,
  useArchiveWorkspaceMutation,
  useUnarchiveWorkspaceMutation,
} from '@/lib/store/api/workspacesApi';
import { useGetOrganizationsQuery } from '@/lib/store/api/organizationsApi';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';

export default function WorkspacesDirectoryPage() {
  const authState = useAppSelector((state) => state.auth);
  const activeOrgId = authState.activeOrganizationId;
  const { user, isOwner, isManager, isSuperAdmin } = useAuthRbac();
  const canManage = isOwner || isManager || isSuperAdmin;

  const { data: organizations = [] } = useGetOrganizationsQuery();
  const currentOrg = useMemo(() => {
    return organizations.find((o) => o.id === activeOrgId) || organizations[0] || null;
  }, [organizations, activeOrgId]);

  const orgId = currentOrg?.id || '';

  const [activeTab, setActiveTab] = useState<'active' | 'archived'>('active');
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');

  // Modals state
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [workspaceToArchive, setWorkspaceToArchive] = useState<{ id: string; name: string } | null>(null);
  const [newWorkspaceName, setNewWorkspaceName] = useState('');
  const [newWorkspaceDesc, setNewWorkspaceDesc] = useState('');

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(searchQuery), 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Queries
  const { data: activeWorkspaces = [], isLoading: isActiveLoading } = useGetWorkspacesQuery(
    orgId ? { organizationId: orgId } : undefined,
    { skip: !orgId }
  );

  const { data: archivedWorkspaces = [], isLoading: isArchivedLoading } = useGetArchivedWorkspacesQuery(
    orgId,
    { skip: !orgId || activeTab !== 'archived' }
  );

  // Mutations
  const [createWorkspace, { isLoading: isCreating }] = useCreateWorkspaceMutation();
  const [archiveWorkspace, { isLoading: isArchiving }] = useArchiveWorkspaceMutation();
  const [unarchiveWorkspace, { isLoading: isUnarchiving }] = useUnarchiveWorkspaceMutation();

  // Filtered lists
  const filteredActive = useMemo(() => {
    if (!debouncedSearch.trim()) return activeWorkspaces;
    const q = debouncedSearch.toLowerCase();
    return activeWorkspaces.filter(
      (ws) =>
        ws.name.toLowerCase().includes(q) ||
        ws.slug.toLowerCase().includes(q) ||
        (ws.description && ws.description.toLowerCase().includes(q))
    );
  }, [activeWorkspaces, debouncedSearch]);

  const filteredArchived = useMemo(() => {
    if (!debouncedSearch.trim()) return archivedWorkspaces;
    const q = debouncedSearch.toLowerCase();
    return archivedWorkspaces.filter(
      (ws) =>
        ws.name.toLowerCase().includes(q) ||
        ws.slug.toLowerCase().includes(q) ||
        (ws.description && ws.description.toLowerCase().includes(q))
    );
  }, [archivedWorkspaces, debouncedSearch]);

  const handleCreateWorkspace = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newWorkspaceName.trim() || !orgId) return;

    try {
      await createWorkspace({
        name: newWorkspaceName.trim(),
        description: newWorkspaceDesc.trim() || undefined,
        organizationId: orgId,
      }).unwrap();

      toast.success('Workspace created successfully');
      setNewWorkspaceName('');
      setNewWorkspaceDesc('');
      setIsCreateModalOpen(false);
    } catch {
      toast.error('Failed to create workspace');
    }
  };

  const handleConfirmArchive = async () => {
    if (!workspaceToArchive) return;
    try {
      await archiveWorkspace(workspaceToArchive.id).unwrap();
      toast.success(`Workspace "${workspaceToArchive.name}" archived`);
      setWorkspaceToArchive(null);
    } catch {
      toast.error('Failed to archive workspace');
    }
  };

  const handleUnarchive = async (id: string, name: string) => {
    try {
      await unarchiveWorkspace(id).unwrap();
      toast.success(`Workspace "${name}" restored to active status`);
    } catch {
      toast.error('Failed to unarchive workspace');
    }
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#09090b] p-6 space-y-6 overflow-y-auto">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5">
            <Briefcase className="h-5 w-5 text-zinc-200" />
            <h1 className="text-xl font-bold tracking-tight text-zinc-100">
              Workspaces Directory
            </h1>
            {currentOrg && (
              <Badge variant="outline" className="text-[11px] font-mono border-zinc-800 text-zinc-400">
                {currentOrg.name}
              </Badge>
            )}
          </div>
          <p className="text-xs text-zinc-400">
            Browse, manage, and archive workspace scopes across your organization.
          </p>
        </div>

        {canManage && (
          <Button
            variant="primary"
            size="sm"
            onClick={() => setIsCreateModalOpen(true)}
            className="gap-1.5 text-xs h-8 cursor-pointer shrink-0"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>New Workspace</span>
          </Button>
        )}
      </div>

      {/* Toolbar: Search & Tab Switcher */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <Tabs
          value={activeTab}
          onValueChange={(val: string) => setActiveTab(val as 'active' | 'archived')}
          className="w-auto"
        >
          <TabsList className="bg-zinc-900 border border-zinc-800 p-1 rounded-md h-9 gap-1">
            <TabsTrigger value="active" className="text-xs gap-1.5 px-3">
              <Briefcase className="h-3.5 w-3.5 text-zinc-400" />
              <span>Active Workspaces</span>
              <span className="text-[10px] font-mono text-zinc-500">
                ({activeWorkspaces.length})
              </span>
            </TabsTrigger>
            <TabsTrigger value="archived" className="text-xs gap-1.5 px-3">
              <Archive className="h-3.5 w-3.5 text-zinc-500" />
              <span>Archived</span>
              <span className="text-[10px] font-mono text-zinc-500">
                ({archivedWorkspaces.length})
              </span>
            </TabsTrigger>
          </TabsList>
        </Tabs>

        <div className="relative w-full sm:w-72">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
          <Input
            placeholder="Search workspaces by name or slug..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-8 h-8 text-xs bg-zinc-900 border-zinc-800 text-zinc-200"
          />
        </div>
      </div>

      {/* Workspaces Grid */}
      {activeTab === 'active' ? (
        isActiveLoading ? (
          <div className="flex items-center justify-center p-16 text-xs text-zinc-500">
            <div className="h-5 w-5 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100 mr-2" />
            Loading active workspaces...
          </div>
        ) : filteredActive.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-16 text-center border border-zinc-800/80 rounded-md bg-zinc-950/40 space-y-3">
            <Briefcase className="h-8 w-8 text-zinc-600" />
            <h3 className="text-xs font-semibold text-zinc-300">No Workspaces Found</h3>
            <p className="text-[11px] text-zinc-500 max-w-sm">
              {debouncedSearch
                ? 'No active workspaces matched your search term.'
                : 'Create your first workspace to begin organizing projects and team members.'}
            </p>
            {canManage && !debouncedSearch && (
              <Button
                variant="primary"
                size="sm"
                onClick={() => setIsCreateModalOpen(true)}
                className="text-xs h-8 gap-1.5"
              >
                <Plus className="h-3.5 w-3.5" />
                <span>Create Workspace</span>
              </Button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredActive.map((ws) => (
              <Card
                key={ws.id}
                className="border-zinc-800 bg-zinc-900/60 hover:border-zinc-700 transition flex flex-col justify-between p-5 space-y-4"
              >
                <div className="space-y-2">
                  <div className="flex items-start justify-between gap-2">
                    <div className="space-y-0.5">
                      <Link
                        href={`/${ws.slug}`}
                        className="text-base font-semibold text-zinc-100 hover:text-white transition flex items-center gap-1.5"
                      >
                        <span>{ws.name}</span>
                        <ExternalLink className="h-3.5 w-3.5 text-zinc-500" />
                      </Link>
                      <span className="text-[11px] font-mono text-zinc-500 block">
                        /{ws.slug}
                      </span>
                    </div>

                    {canManage && (
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-7 w-7 p-0 text-zinc-400 hover:text-zinc-200"
                          >
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-40 border-zinc-800 bg-zinc-900 text-xs text-zinc-200">
                          <Link href={`/${ws.slug}/settings`}>
                            <DropdownMenuItem className="cursor-pointer">
                              Settings
                            </DropdownMenuItem>
                          </Link>
                          <DropdownMenuItem
                            onClick={() => setWorkspaceToArchive({ id: ws.id, name: ws.name })}
                            className="cursor-pointer text-amber-400 focus:text-amber-300"
                          >
                            Archive Workspace
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    )}
                  </div>

                  {ws.description && (
                    <p className="text-xs text-zinc-400 line-clamp-2">
                      {ws.description}
                    </p>
                  )}
                </div>

                <div className="pt-3 border-t border-zinc-850 flex items-center justify-between text-xs">
                  <span className="text-[11px] text-zinc-500 font-mono">
                    Created {new Date(ws.createdAt).toLocaleDateString()}
                  </span>
                  <Link href={`/${ws.slug}`}>
                    <Button variant="outline" size="sm" className="h-7 text-xs border-zinc-800 text-zinc-300 hover:text-white">
                      <span>Open &rarr;</span>
                    </Button>
                  </Link>
                </div>
              </Card>
            ))}
          </div>
        )
      ) : isArchivedLoading ? (
        <div className="flex items-center justify-center p-16 text-xs text-zinc-500">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100 mr-2" />
          Loading archived workspaces...
        </div>
      ) : filteredArchived.length === 0 ? (
        <div className="flex flex-col items-center justify-center p-16 text-center border border-zinc-800/80 rounded-md bg-zinc-950/40 space-y-2">
          <Archive className="h-8 w-8 text-zinc-600" />
          <h3 className="text-xs font-semibold text-zinc-300">No Archived Workspaces</h3>
          <p className="text-[11px] text-zinc-500">
            Workspaces you archive will appear here, safely preserved with their project history.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredArchived.map((ws) => (
            <Card
              key={ws.id}
              className="border-zinc-850 bg-zinc-950/60 p-5 space-y-4 flex flex-col justify-between"
            >
              <div className="space-y-1.5">
                <div className="flex items-start justify-between gap-2">
                  <div className="space-y-0.5">
                    <span className="text-base font-semibold text-zinc-400">
                      {ws.name}
                    </span>
                    <span className="text-[11px] font-mono text-zinc-600 block">
                      /{ws.slug}
                    </span>
                  </div>
                  <Badge variant="outline" className="border-amber-900/60 text-amber-500 bg-amber-950/20 text-[10px]">
                    Archived
                  </Badge>
                </div>
                {ws.description && (
                  <p className="text-xs text-zinc-500 line-clamp-2">
                    {ws.description}
                  </p>
                )}
              </div>

              <div className="pt-3 border-t border-zinc-850/60 flex items-center justify-between">
                <span className="text-[11px] text-zinc-600 font-mono">
                  Read-only archive
                </span>
                {canManage && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleUnarchive(ws.id, ws.name)}
                    disabled={isUnarchiving}
                    className="gap-1.5 h-7 text-xs border-zinc-700 bg-zinc-900 text-zinc-200 hover:bg-zinc-800 cursor-pointer"
                  >
                    <ArchiveRestore className="h-3.5 w-3.5 text-emerald-400" />
                    <span>Unarchive</span>
                  </Button>
                )}
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Create Workspace Modal */}
      <Dialog open={isCreateModalOpen} onOpenChange={setIsCreateModalOpen}>
        <DialogContent className="sm:max-w-md bg-zinc-950 border-zinc-800 text-zinc-100 p-6">
          <DialogHeader className="space-y-1.5">
            <DialogTitle className="text-base font-semibold text-zinc-100">
              Create New Workspace
            </DialogTitle>
            <DialogDescription className="text-xs text-zinc-400">
              A workspace isolates projects, boards, sprints, and membership boundaries.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleCreateWorkspace} className="space-y-4 pt-2">
            <div className="space-y-1.5">
              <Label htmlFor="ws-name" className="text-xs text-zinc-300">
                Workspace Name
              </Label>
              <Input
                id="ws-name"
                required
                placeholder="e.g. Core Engineering"
                value={newWorkspaceName}
                onChange={(e) => setNewWorkspaceName(e.target.value)}
                className="bg-zinc-900 border-zinc-800 text-xs text-zinc-100 h-9"
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="ws-desc" className="text-xs text-zinc-300">
                Description (Optional)
              </Label>
              <Input
                id="ws-desc"
                placeholder="e.g. Platform microservices, backend APIs, and infrastructure..."
                value={newWorkspaceDesc}
                onChange={(e) => setNewWorkspaceDesc(e.target.value)}
                className="bg-zinc-900 border-zinc-800 text-xs text-zinc-100 h-9"
              />
            </div>

            <DialogFooter className="pt-3 gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setIsCreateModalOpen(false)}
                className="text-xs border-zinc-800 bg-zinc-900 text-zinc-300"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="primary"
                size="sm"
                disabled={isCreating || !newWorkspaceName.trim()}
                className="text-xs gap-1.5"
              >
                {isCreating && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                <span>Create Workspace</span>
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Archive Confirmation Dialog */}
      <Dialog open={!!workspaceToArchive} onOpenChange={(open) => (!open ? setWorkspaceToArchive(null) : null)}>
        <DialogContent className="sm:max-w-sm bg-zinc-950 border-zinc-800 text-zinc-100 p-6">
          <DialogHeader className="space-y-2">
            <DialogTitle className="text-sm font-semibold text-zinc-100">
              Archive Workspace
            </DialogTitle>
            <DialogDescription className="text-xs text-zinc-400">
              Are you sure you want to archive &ldquo;{workspaceToArchive?.name}&rdquo;? It will be hidden from everyday views but can be unarchived at any time.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="pt-3 gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setWorkspaceToArchive(null)}
              className="text-xs border-zinc-800 bg-zinc-900 text-zinc-300"
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={handleConfirmArchive}
              disabled={isArchiving}
              className="text-xs bg-amber-600 hover:bg-amber-500 text-white"
            >
              Archive
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

