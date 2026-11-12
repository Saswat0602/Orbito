'use client';

import * as React from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  ArrowLeft,
  Mail,
  Calendar,
  Briefcase,
  Building,
  Globe,
  Clock,
  CheckCircle2,
  ListTodo,
  ExternalLink,
  Shield,
  Loader2,
  AlertCircle,
} from 'lucide-react';

import { useGetUserByIdQuery } from '@/lib/store/api/usersApi';
import {
  useGetWorkspaceBySlugQuery,
  useGetWorkspaceMembersQuery,
} from '@/lib/store/api/workspacesApi';
import { useGetTasksQuery } from '@/lib/store/api/tasksApi';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { getRoleBadgeStyle } from '@/lib/rbac';

export default function WorkspaceMemberProfilePage() {
  const params = useParams<{ workspaceSlug: string; id: string }>();
  const router = useRouter();
  const workspaceSlug = params?.workspaceSlug || '';
  const memberId = params?.id || '';

  const { data: workspace, isLoading: isWorkspaceLoading } = useGetWorkspaceBySlugQuery(
    workspaceSlug,
    { skip: !workspaceSlug }
  );

  const { data: memberUser, isLoading: isUserLoading } = useGetUserByIdQuery(memberId, {
    skip: !memberId,
  });

  const { data: workspaceMembers = [], isLoading: isMembersLoading } =
    useGetWorkspaceMembersQuery(workspace?.id || '', {
      skip: !workspace?.id,
    });

  const { data: assignedTasks = [], isLoading: isTasksLoading } = useGetTasksQuery(
    workspace?.id ? { workspaceId: workspace.id, assigneeId: memberId } : undefined,
    { skip: !workspace?.id || !memberId }
  );

  const memberRecord = workspaceMembers.find(
    (m) => m.userId === memberId || m.user?.id === memberId
  );

  const isLoading = isWorkspaceLoading || isUserLoading;

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center min-h-[60vh] bg-[#09090b]">
        <div className="flex flex-col items-center gap-3 text-zinc-500">
          <Loader2 className="h-6 w-6 animate-spin text-zinc-400" />
          <span className="text-xs">Loading member profile...</span>
        </div>
      </div>
    );
  }

  if (!memberUser) {
    return (
      <div className="flex-1 p-6 md:p-8 space-y-4 bg-[#09090b]">
        <Link
          href={`/${workspaceSlug}/members`}
          className="inline-flex items-center gap-2 text-xs text-zinc-400 hover:text-zinc-200 transition-colors"
        >
          <ArrowLeft className="h-4 w-4" /> Back to Members
        </Link>
        <Card className="border-zinc-800 bg-zinc-900/60 p-8 text-center max-w-lg mx-auto">
          <AlertCircle className="h-10 w-10 text-zinc-600 mx-auto mb-3" />
          <h3 className="text-sm font-semibold text-zinc-200">Member Not Found</h3>
          <p className="text-xs text-zinc-500 mt-1">
            This user profile does not exist or is not a member of workspace &ldquo;{workspaceSlug}&rdquo;.
          </p>
        </Card>
      </div>
    );
  }

  const displayName =
    `${memberUser.firstName || ''} ${memberUser.lastName || ''}`.trim() ||
    memberUser.username ||
    'Member';
  const initials = displayName.slice(0, 2).toUpperCase() || 'U';
  const preferences = (memberUser.preferences as Record<string, unknown> | undefined) || {};
  const department = (preferences.department as string) || 'Engineering';
  const jobTitle = (preferences.jobTitle as string) || 'Team Contributor';
  const timezone = (preferences.timezone as string) || 'UTC';
  const bio = (preferences.bio as string) || (memberUser as { bio?: string }).bio || null;

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b]">
      {/* Breadcrumb Navigation */}
      <div className="flex items-center justify-between pb-3 border-b border-zinc-800/80">
        <div className="flex items-center gap-2 text-xs text-zinc-400">
          <Link
            href={`/${workspaceSlug}/members`}
            className="hover:text-zinc-200 transition-colors flex items-center gap-1.5"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            <span>Workspace Members</span>
          </Link>
          <span className="text-zinc-600">/</span>
          <span className="text-zinc-200 font-medium truncate max-w-xs">{displayName}</span>
        </div>

        <div className="flex items-center gap-2">
          {memberRecord?.role && (
            <Badge
              variant="outline"
              className={`text-[10px] uppercase font-mono px-2 py-0.5 ${getRoleBadgeStyle(
                memberRecord.role
              )}`}
            >
              {memberRecord.role}
            </Badge>
          )}
          <Badge
            variant="outline"
            className="text-[10px] border-emerald-800/60 text-emerald-400 bg-emerald-950/30 flex items-center gap-1"
          >
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 inline-block animate-pulse" />
            Active
          </Badge>
        </div>
      </div>

      {/* Profile Header Hero Card */}
      <Card className="border-zinc-800 bg-zinc-900/60 shadow-lg">
        <CardContent className="p-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
            <div className="flex items-start sm:items-center gap-4">
              <Avatar className="h-16 w-16 rounded-md border-2 border-zinc-700/80 shrink-0">
                {memberUser.avatar && <AvatarImage src={memberUser.avatar} alt={displayName} />}
                <AvatarFallback className="bg-zinc-800 text-zinc-200 text-lg font-bold">
                  {initials}
                </AvatarFallback>
              </Avatar>

              <div className="space-y-1">
                <div className="flex flex-wrap items-center gap-2">
                  <h1 className="text-xl font-bold text-zinc-100">{displayName}</h1>
                  {memberUser.username && (
                    <span className="text-xs text-zinc-500 font-mono">@{memberUser.username}</span>
                  )}
                </div>
                <div className="flex flex-wrap items-center gap-3 text-xs text-zinc-400">
                  <span className="flex items-center gap-1">
                    <Briefcase className="h-3.5 w-3.5 text-zinc-500" />
                    {jobTitle}
                  </span>
                  <span>&bull;</span>
                  <span className="flex items-center gap-1">
                    <Building className="h-3.5 w-3.5 text-zinc-500" />
                    {department}
                  </span>
                  <span>&bull;</span>
                  <span className="flex items-center gap-1">
                    <Calendar className="h-3.5 w-3.5 text-zinc-500" />
                    Joined {new Date(memberUser.createdAt).toLocaleDateString(undefined, { month: 'short', year: 'numeric' })}
                  </span>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-2 shrink-0">
              <Button
                asChild
                variant="outline"
                size="sm"
                className="h-8 text-xs border-zinc-800 text-zinc-300 hover:text-white"
              >
                <a href={`mailto:${memberUser.email}`}>
                  <Mail className="h-3.5 w-3.5 mr-1.5" />
                  Contact Member
                </a>
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Main Grid: Details & Tasks */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column: Bio & Information */}
        <div className="space-y-6 lg:col-span-1">
          {/* Contact Details Card */}
          <Card className="border-zinc-800 bg-zinc-900/60">
            <CardHeader className="p-4 pb-3 border-b border-zinc-800/80">
              <CardTitle className="text-xs font-semibold text-zinc-200 uppercase tracking-wider">
                Member Information
              </CardTitle>
            </CardHeader>
            <CardContent className="p-4 space-y-3.5 text-xs">
              <div className="space-y-1">
                <span className="text-[11px] text-zinc-500 font-medium">Work Email</span>
                <p className="text-zinc-200 font-mono select-all">{memberUser.email}</p>
              </div>

              <div className="space-y-1">
                <span className="text-[11px] text-zinc-500 font-medium">Workspace Scope</span>
                <p className="text-zinc-300 font-medium">{workspace?.name || workspaceSlug}</p>
              </div>

              <div className="space-y-1">
                <span className="text-[11px] text-zinc-500 font-medium">Timezone</span>
                <p className="text-zinc-300 flex items-center gap-1.5">
                  <Globe className="h-3.5 w-3.5 text-zinc-500" />
                  {timezone}
                </p>
              </div>

              <div className="space-y-1">
                <span className="text-[11px] text-zinc-500 font-medium">Platform Role</span>
                <p className="text-zinc-300 flex items-center gap-1.5">
                  <Shield className="h-3.5 w-3.5 text-zinc-500" />
                  {memberUser.role}
                </p>
              </div>
            </CardContent>
          </Card>

          {/* About / Bio Card */}
          <Card className="border-zinc-800 bg-zinc-900/60">
            <CardHeader className="p-4 pb-3 border-b border-zinc-800/80">
              <CardTitle className="text-xs font-semibold text-zinc-200 uppercase tracking-wider">
                About & Expertise
              </CardTitle>
            </CardHeader>
            <CardContent className="p-4 text-xs text-zinc-400 leading-relaxed">
              {bio ? (
                <p className="whitespace-pre-line text-zinc-300">{bio}</p>
              ) : (
                <p className="italic text-zinc-500">
                  No public bio or profile summary provided yet.
                </p>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Right Column: Assigned Tasks */}
        <div className="space-y-6 lg:col-span-2">
          <Card className="border-zinc-800 bg-zinc-900/60">
            <CardHeader className="p-4 pb-3 border-b border-zinc-800/80">
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-sm font-semibold text-zinc-200 flex items-center gap-2">
                    <ListTodo className="h-4 w-4 text-zinc-400" />
                    Assigned Tasks ({assignedTasks.length})
                  </CardTitle>
                  <CardDescription className="text-xs text-zinc-400 mt-0.5">
                    Active deliverables assigned to {displayName} in this workspace
                  </CardDescription>
                </div>
              </div>
            </CardHeader>

            <CardContent className="p-0">
              {isTasksLoading ? (
                <div className="py-12 text-center text-xs text-zinc-500">
                  Loading assigned tasks...
                </div>
              ) : assignedTasks.length === 0 ? (
                <div className="py-12 text-center text-xs text-zinc-500">
                  No open tasks currently assigned in this workspace.
                </div>
              ) : (
                <div className="divide-y divide-zinc-800/80 max-h-[500px] overflow-y-auto">
                  {assignedTasks.map((t) => {
                    const taskProjectSlug = t.project?.slug || '';
                    const taskUrl = taskProjectSlug
                      ? `/${workspaceSlug}/${taskProjectSlug}/tasks/${t.id}`
                      : '#';

                    return (
                      <div
                        key={t.id}
                        className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-zinc-850/40 transition-colors"
                      >
                        <div className="space-y-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-semibold text-zinc-100 truncate">
                              {t.title}
                            </span>
                            {t.priority && (
                              <Badge
                                variant="outline"
                                className="text-[9px] uppercase font-mono py-0 px-1 border-zinc-800 text-zinc-400"
                              >
                                {t.priority}
                              </Badge>
                            )}
                          </div>
                          <div className="flex items-center gap-3 text-[11px] text-zinc-500">
                            {t.project?.name && <span>Project: {t.project.name}</span>}
                            {t.dueDate && (
                              <span className="flex items-center gap-1 text-zinc-400">
                                <Clock className="h-3 w-3" />
                                Due {new Date(t.dueDate).toLocaleDateString()}
                              </span>
                            )}
                          </div>
                        </div>

                        <div className="flex items-center gap-3 shrink-0">
                          {t.status && (
                            <Badge
                              variant="outline"
                              className="text-[10px] border-zinc-800 text-zinc-300 font-mono"
                            >
                              {t.status.name}
                            </Badge>
                          )}
                          {taskProjectSlug && (
                            <Button
                              asChild
                              variant="ghost"
                              size="sm"
                              className="h-7 text-xs text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800"
                            >
                              <Link href={taskUrl}>
                                View Task <ExternalLink className="h-3 w-3 ml-1" />
                              </Link>
                            </Button>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
