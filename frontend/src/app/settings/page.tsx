'use client';

import * as React from 'react';
import { useState } from 'react';
import {
  Settings,
  User,
  Building2,
  Puzzle,
  Zap,
  Save,
  CheckCircle2,
} from 'lucide-react';
import { toast } from 'sonner';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useUpdateProfileMutation } from '@/lib/store/api/usersApi';
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

export default function SettingsPage() {
  const { user, role, isManager, isOwner, roleBadge } = useAuthRbac();

  const [firstName, setFirstName] = useState(user?.firstName || '');
  const [lastName, setLastName] = useState(user?.lastName || '');
  const [bio, setBio] = useState(user?.bio || '');

  const [updateProfile, { isLoading: isUpdating }] = useUpdateProfileMutation();

  const handleProfileSave = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await updateProfile({ firstName, lastName, bio }).unwrap();
      toast.success('Profile details updated');
    } catch {
      toast.error('Failed to update profile');
    }
  };

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Settings className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Settings
            </h1>
          </div>
          <p className="text-xs text-zinc-400">
            Manage your personal profile, workspace configurations, and integrations.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge className={`text-xs uppercase ${roleBadge.className}`}>
            {roleBadge.label}
          </Badge>
        </div>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="profile" className="space-y-6">
        <TabsList className="bg-zinc-900 border border-zinc-800 p-1 rounded-lg">
          <TabsTrigger value="profile" className="gap-2 text-xs">
            <User className="h-3.5 w-3.5" />
            <span>Profile</span>
          </TabsTrigger>
          {(isManager || isOwner) && (
            <>
              <TabsTrigger value="workspace" className="gap-2 text-xs">
                <Building2 className="h-3.5 w-3.5" />
                <span>Workspace</span>
              </TabsTrigger>
              <TabsTrigger value="integrations" className="gap-2 text-xs">
                <Puzzle className="h-3.5 w-3.5" />
                <span>Integrations</span>
              </TabsTrigger>
              <TabsTrigger value="automations" className="gap-2 text-xs">
                <Zap className="h-3.5 w-3.5" />
                <span>Automations</span>
              </TabsTrigger>
            </>
          )}
        </TabsList>

        {/* Profile Tab */}
        <TabsContent value="profile">
          <Card className="border-zinc-800 bg-zinc-900/80 max-w-2xl">
            <CardHeader>
              <CardTitle className="text-base">Personal Details</CardTitle>
              <CardDescription className="text-xs">
                Update how your name and profile appear across projects.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleProfileSave} className="space-y-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="space-y-1.5">
                    <Label htmlFor="firstName">First Name</Label>
                    <Input
                      id="firstName"
                      value={firstName}
                      onChange={(e) => setFirstName(e.target.value)}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="lastName">Last Name</Label>
                    <Input
                      id="lastName"
                      value={lastName}
                      onChange={(e) => setLastName(e.target.value)}
                    />
                  </div>
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="email">Email Address</Label>
                  <Input
                    id="email"
                    value={user?.email || ''}
                    disabled
                    className="bg-zinc-950 text-zinc-500 cursor-not-allowed"
                  />
                  <p className="text-[11px] text-zinc-500">
                    Email address is tied to your account identity.
                  </p>
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="bio">Bio</Label>
                  <Input
                    id="bio"
                    placeholder="Staff Software Engineer..."
                    value={bio}
                    onChange={(e) => setBio(e.target.value)}
                  />
                </div>

                <Button
                  type="submit"
                  variant="primary"
                  size="sm"
                  isLoading={isUpdating}
                  className="gap-2 mt-2"
                >
                  <Save className="h-3.5 w-3.5" />
                  <span>Save Changes</span>
                </Button>
              </form>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Workspace Tab (Manager/Owner only) */}
        {(isManager || isOwner) && (
          <TabsContent value="workspace">
            <Card className="border-zinc-800 bg-zinc-900/80 max-w-2xl">
              <CardHeader>
                <CardTitle className="text-base">Workspace Configuration</CardTitle>
                <CardDescription className="text-xs">
                  Global parameters, default workflows, and membership boundaries.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="p-4 rounded-lg bg-zinc-850/50 border border-zinc-800 text-xs text-zinc-300">
                  Workspace preferences are inherited by all child projects and agile boards.
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        )}

        {/* Integrations Tab */}
        {(isManager || isOwner) && (
          <TabsContent value="integrations">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-3xl">
              <Card className="border-zinc-800 bg-zinc-900/80 p-5 space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-semibold text-zinc-200">Jira Sync</h4>
                  <Badge variant="outline" className="text-[10px]">Ready</Badge>
                </div>
                <p className="text-xs text-zinc-400">
                  Bidirectional two-way sync for Jira Cloud and Server issues.
                </p>
                <Button variant="secondary" size="sm" className="w-full text-xs">
                  Configure Jira
                </Button>
              </Card>

              <Card className="border-zinc-800 bg-zinc-900/80 p-5 space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-semibold text-zinc-200">Trello Boards</h4>
                  <Badge variant="outline" className="text-[10px]">Ready</Badge>
                </div>
                <p className="text-xs text-zinc-400">
                  Import Trello cards and sync lists with task status columns.
                </p>
                <Button variant="secondary" size="sm" className="w-full text-xs">
                  Configure Trello
                </Button>
              </Card>
            </div>
          </TabsContent>
        )}

        {/* Automations Tab */}
        {(isManager || isOwner) && (
          <TabsContent value="automations">
            <Card className="border-zinc-800 bg-zinc-900/80 max-w-2xl p-6 text-center">
              <Zap className="h-8 w-8 text-zinc-500 mx-auto mb-2" />
              <h4 className="text-sm font-semibold text-zinc-200">
                Trigger-Condition-Action Rules
              </h4>
              <p className="text-xs text-zinc-400 mt-1 max-w-md mx-auto">
                Automate task transitions, auto-assignments, and alert dispatches with the backend BullMQ engine.
              </p>
            </Card>
          </TabsContent>
        )}
      </Tabs>
    </div>
  );
}

