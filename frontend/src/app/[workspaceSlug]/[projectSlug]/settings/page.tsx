'use client';

import * as React from 'react';
import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import {
  Settings,
  FolderGit2,
  Users,
  Layers,
  Mail,
  Save,
  Trash2,
  Copy,
  Check,
  AlertTriangle,
  ArrowLeft,
} from 'lucide-react';
import { toast } from 'sonner';
import Link from 'next/link';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import {
  useGetProjectBySlugQuery,
  useUpdateProjectMutation,
  useDeleteProjectMutation,
} from '@/lib/store/api/projectsApi';
import { useGetWorkspaceMembersQuery } from '@/lib/store/api/workspacesApi';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { getRoleBadgeStyle } from '@/lib/rbac';
import { WorkflowEditor } from '@/features/projects/components/WorkflowEditor';
import { ProjectInboxTab } from '@/features/projects/components/ProjectInboxTab';

export default function ProjectSettingsPage() {
  const params = useParams<{ workspaceSlug: string; projectSlug: string }>();
  const router = useRouter();
  const { workspaceSlug, projectSlug } = params;
  const { isManager, isOwner, isSuperAdmin, roleBadge } = useAuthRbac();

  const { data: project, isLoading: isProjectLoading } = useGetProjectBySlugQuery(
    projectSlug || '',
    { skip: !projectSlug }
  );

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const [deleteConfirmation, setDeleteConfirmation] = useState('');

  useEffect(() => {
    if (project) {
      setName(project.name || '');
      setDescription(project.description || '');
    }
  }, [project]);

  const [updateProject, { isLoading: isUpdating }] = useUpdateProjectMutation();
  const [deleteProject, { isLoading: isDeleting }] = useDeleteProjectMutation();

  const { data: workspaceMembers = [] } = useGetWorkspaceMembersQuery(
    project?.workspaceId || '',
    { skip: !project?.workspaceId }
  );

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!project?.id) return;
    try {
      await updateProject({
        id: project.id,
        updates: {
          name: name.trim() || project.name,
          description: description.trim(),
        },
      }).unwrap();
      toast.success('Project updated successfully');
    } catch {
      toast.error('Failed to update project');
    }
  };

  const handleCopy = (text: string, key: string) => {
    navigator.clipboard.writeText(text);
    setCopiedKey(key);
    toast.success('Copied to clipboard');
    setTimeout(() => setCopiedKey(null), 2000);
  };

  const handleDeleteProject = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!project?.id) return;
    if (deleteConfirmation !== project.name) {
      toast.error('Project name does not match');
      return;
    }

    try {
      await deleteProject(project.id).unwrap();
      toast.success('Project deleted successfully');
      router.push(`/${workspaceSlug}`);
    } catch {
      toast.error('Failed to delete project');
    }
  };

  if (isProjectLoading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b]">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
      </div>
    );
  }

  if (!project) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#09090b] text-xs text-zinc-500">
        Project not found.
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Link
              href={`/${workspaceSlug}/${projectSlug}/tasks`}
              className="text-zinc-500 hover:text-zinc-200 transition-colors p-1 -ml-1 rounded"
              title="Back to Tasks"
            >
              <ArrowLeft className="h-4 w-4" />
            </Link>
            <FolderGit2 className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Project Settings
            </h1>
            <Badge variant="outline" className="text-[10px] font-mono uppercase">
              {project.slug}
            </Badge>
          </div>
          <p className="text-xs text-zinc-400">
            Configure parameters, workflows, and lifecycle settings for{' '}
            <strong className="text-zinc-200">{project.name}</strong>.
          </p>
        </div>

        <Badge className={`text-xs uppercase ${roleBadge.className}`}>
          {roleBadge.label}
        </Badge>
      </div>

      <Tabs defaultValue="general" className="space-y-6">
        <TabsList className="bg-zinc-900 border border-zinc-800 p-1 rounded-md">
          <TabsTrigger value="general" className="gap-2 text-xs">
            <Settings className="h-3.5 w-3.5" />
            <span>General</span>
          </TabsTrigger>
          <TabsTrigger value="members" className="gap-2 text-xs">
            <Users className="h-3.5 w-3.5" />
            <span>Project Collaborators</span>
          </TabsTrigger>
          <TabsTrigger value="workflow" className="gap-2 text-xs">
            <Layers className="h-3.5 w-3.5" />
            <span>Workflow & Board</span>
          </TabsTrigger>
          <TabsTrigger value="inbox" className="gap-2 text-xs">
            <Mail className="h-3.5 w-3.5" />
            <span>Inbound Email & Rules</span>
          </TabsTrigger>
          {(isManager || isOwner || isSuperAdmin) && (
            <TabsTrigger value="danger" className="gap-2 text-xs text-rose-400 data-[state=active]:text-rose-300">
              <AlertTriangle className="h-3.5 w-3.5" />
              <span>Danger Zone</span>
            </TabsTrigger>
          )}
        </TabsList>

        {/* General Settings */}
        <TabsContent value="general">
          <div className="space-y-6 max-w-3xl">
            <Card className="border-zinc-800 bg-zinc-900/80">
              <CardHeader>
                <CardTitle className="text-base">Project Identity</CardTitle>
                <CardDescription className="text-xs">
                  Project name and purpose visible on boards, issues, and sprint backlogs.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleSave} className="space-y-4">
                  <div className="space-y-1.5">
                    <Label htmlFor="projName" className="text-xs text-zinc-300">Project Name</Label>
                    <Input
                      id="projName"
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      disabled={!isManager && !isOwner}
                      className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                    />
                  </div>

                  <div className="space-y-1.5">
                    <Label htmlFor="projDesc" className="text-xs text-zinc-300">Description</Label>
                    <Input
                      id="projDesc"
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                      disabled={!isManager && !isOwner}
                      placeholder="e.g. Next-generation user authentication service"
                      className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                    />
                  </div>

                  {(isManager || isOwner) && (
                    <Button
                      type="submit"
                      variant="primary"
                      size="sm"
                      className="gap-2 mt-2 text-xs"
                      isLoading={isUpdating}
                    >
                      <Save className="h-3.5 w-3.5" />
                      <span>Save Changes</span>
                    </Button>
                  )}
                </form>
              </CardContent>
            </Card>

            {/* Read-only IDs */}
            <Card className="border-zinc-800 bg-zinc-900/80 p-5 space-y-4">
              <h4 className="text-xs font-semibold text-zinc-200">System Identifiers</h4>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1">
                  <span className="text-[10px] uppercase font-mono text-zinc-500">Project Slug</span>
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-zinc-300">{project.slug}</span>
                    <button
                      type="button"
                      onClick={() => handleCopy(project.slug, 'slug')}
                      className="text-zinc-500 hover:text-zinc-200 transition-colors"
                      title="Copy Slug"
                    >
                      {copiedKey === 'slug' ? <Check className="h-3 w-3 text-emerald-400" /> : <Copy className="h-3 w-3" />}
                    </button>
                  </div>
                </div>

                <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1">
                  <span className="text-[10px] uppercase font-mono text-zinc-500">Project ID</span>
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-zinc-400 truncate max-w-[180px]">{project.id}</span>
                    <button
                      type="button"
                      onClick={() => handleCopy(project.id, 'id')}
                      className="text-zinc-500 hover:text-zinc-200 transition-colors"
                      title="Copy ID"
                    >
                      {copiedKey === 'id' ? <Check className="h-3 w-3 text-emerald-400" /> : <Copy className="h-3 w-3" />}
                    </button>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        </TabsContent>

        {/* Project Collaborators */}
        <TabsContent value="members">
          <Card className="border-zinc-800 bg-zinc-900/80 max-w-3xl">
            <CardHeader>
              <CardTitle className="text-base">Assigned Team</CardTitle>
              <CardDescription className="text-xs">
                Collaborators eligible to be assigned issues, sprints, and reviews in this project.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {workspaceMembers.length === 0 ? (
                <div className="py-8 text-center text-xs text-zinc-500">No collaborators found.</div>
              ) : (
                <div className="divide-y divide-zinc-800 border border-zinc-850 rounded-md overflow-hidden">
                  {workspaceMembers.map((member) => {
                    const roleStyle = getRoleBadgeStyle(member.role);
                    const memberName =
                      member.user?.firstName && member.user?.lastName
                        ? `${member.user.firstName} ${member.user.lastName}`
                        : member.user?.username || 'User';
                    const initials =
                      ((member.user?.firstName?.[0] || '') + (member.user?.lastName?.[0] || '')).toUpperCase() || 'U';

                    return (
                      <div
                        key={member.id}
                        className="flex items-center justify-between p-3.5 bg-zinc-950/40 hover:bg-zinc-900/40 transition-colors"
                      >
                        <div className="flex items-center gap-3 min-w-0">
                          <Avatar className="h-8 w-8 rounded-md border border-zinc-800">
                            <AvatarFallback className="bg-zinc-800 text-zinc-200 text-xs font-bold">
                              {initials}
                            </AvatarFallback>
                          </Avatar>
                          <div className="flex flex-col min-w-0">
                            <span className="text-xs font-semibold text-zinc-200 truncate">{memberName}</span>
                            <span className="text-[11px] text-zinc-500 truncate">{member.user?.email}</span>
                          </div>
                        </div>

                        <Badge className={`text-[10px] uppercase ${roleStyle.className}`}>
                          {roleStyle.label}
                        </Badge>
                      </div>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Workflow & Board */}
        <TabsContent value="workflow" className="max-w-3xl">
          <WorkflowEditor
            projectId={project.id}
            workflowId={project.workflowId || undefined}
            canEdit={isManager || isOwner || isSuperAdmin}
          />
        </TabsContent>

        {/* Inbound Email & Rules */}
        <TabsContent value="inbox" className="max-w-3xl">
          <ProjectInboxTab
            projectId={project.id}
            workspaceMembers={workspaceMembers}
            canEdit={isManager || isOwner || isSuperAdmin}
          />
        </TabsContent>

        {/* Danger Zone */}
        {(isManager || isOwner || isSuperAdmin) && (
          <TabsContent value="danger">
            <Card className="border-rose-950/60 bg-zinc-900/80 max-w-2xl border">
              <CardHeader>
                <div className="flex items-center gap-2 text-rose-400">
                  <AlertTriangle className="h-4 w-4" />
                  <CardTitle className="text-base text-rose-300">Delete Project</CardTitle>
                </div>
                <CardDescription className="text-xs text-zinc-400">
                  Permanently deletes this project and all tasks, sprints, ranks, and attachments. This action cannot be undone.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleDeleteProject} className="space-y-4">
                  <div className="space-y-1.5">
                    <Label htmlFor="confirmProjName" className="text-xs text-zinc-300">
                      Please type <strong className="text-rose-400">{project.name}</strong> to confirm:
                    </Label>
                    <Input
                      id="confirmProjName"
                      value={deleteConfirmation}
                      onChange={(e) => setDeleteConfirmation(e.target.value)}
                      placeholder={project.name}
                      className="bg-zinc-950 border-rose-900/60 text-xs text-zinc-100"
                    />
                  </div>

                  <Button
                    type="submit"
                    variant="destructive"
                    size="sm"
                    disabled={deleteConfirmation !== project.name || isDeleting}
                    isLoading={isDeleting}
                    className="gap-2 text-xs"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                    <span>Permanently Delete Project</span>
                  </Button>
                </form>
              </CardContent>
            </Card>
          </TabsContent>
        )}
      </Tabs>
    </div>
  );
}

