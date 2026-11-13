'use client';

import * as React from 'react';
import { useState, useEffect, useRef } from 'react';
import {
  Settings,
  User,
  Shield,
  Palette,
  Save,
  KeyRound,
  CheckCircle2,
  AlertTriangle,
  ExternalLink,
  Bot,
  Upload,
  Trash2,
  Camera,
  Briefcase,
  Building,
} from 'lucide-react';
import { toast } from 'sonner';
import Link from 'next/link';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppSelector } from '@/lib/store/hooks';
import { useUpdateProfileMutation, useChangePasswordMutation } from '@/lib/store/api/usersApi';
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
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { ThemeToggle } from '@/components/layout/ThemeToggle';
import { AISettings } from '@/features/settings';
import { AdminSettingsSkeleton } from '@/components/skeletons';

export default function SettingsPage() {
  const { user, isManager, isOwner, roleBadge } = useAuthRbac();
  const activeWorkspaceSlug = useAppSelector((state) => state.auth.activeWorkspaceSlug);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Profile Form State
  const userPrefs = (user?.preferences as Record<string, unknown> | undefined) || {};
  const [firstName, setFirstName] = useState(user?.firstName || '');
  const [lastName, setLastName] = useState(user?.lastName || '');
  const [bio, setBio] = useState(user?.bio || '');
  const [mobileNumber, setMobileNumber] = useState(user?.mobileNumber || '');
  const [avatar, setAvatar] = useState(user?.avatar || '');
  const [jobTitle, setJobTitle] = useState(typeof userPrefs.jobTitle === 'string' ? userPrefs.jobTitle : '');
  const [department, setDepartment] = useState(typeof userPrefs.department === 'string' ? userPrefs.department : '');

  // Password Change Form State
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  // Sync state whenever auth user loads
  useEffect(() => {
    if (user) {
      const prefs = (user.preferences as Record<string, unknown> | undefined) || {};
      setFirstName(user.firstName || '');
      setLastName(user.lastName || '');
      setBio(user.bio || '');
      setMobileNumber(user.mobileNumber || '');
      setAvatar(user.avatar || '');
      setJobTitle(typeof prefs.jobTitle === 'string' ? prefs.jobTitle : '');
      setDepartment(typeof prefs.department === 'string' ? prefs.department : '');
    }
  }, [user]);

  const [updateProfile, { isLoading: isUpdatingProfile }] = useUpdateProfileMutation();
  const [changePassword, { isLoading: isChangingPassword }] = useChangePasswordMutation();

  const handleAvatarFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!['image/jpeg', 'image/png', 'image/webp', 'image/gif'].includes(file.type)) {
      toast.error('Please upload an image file (PNG, JPG, WEBP, or GIF)');
      return;
    }

    if (file.size > 2 * 1024 * 1024) {
      toast.error('Image size must be less than 2MB');
      return;
    }

    const reader = new FileReader();
    reader.onload = (event) => {
      const result = event.target?.result;
      if (typeof result === 'string') {
        setAvatar(result);
        toast.success('Avatar image ready. Click "Save Changes" to apply.');
      }
    };
    reader.readAsDataURL(file);
  };

  const handleRemoveAvatar = () => {
    setAvatar('');
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleProfileSave = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const existingPrefs = (user?.preferences as Record<string, unknown> | undefined) || {};
      await updateProfile({
        firstName: firstName.trim(),
        lastName: lastName.trim(),
        bio: bio.trim(),
        mobileNumber: mobileNumber.trim() || undefined,
        avatar: avatar || undefined,
        preferences: {
          ...existingPrefs,
          jobTitle: jobTitle.trim() || undefined,
          department: department.trim() || undefined,
        },
      }).unwrap();
      toast.success('Profile details updated successfully');
    } catch {
      toast.error('Failed to update profile details');
    }
  };

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentPassword) {
      toast.error('Current password is required');
      return;
    }
    if (newPassword.length < 8) {
      toast.error('New password must be at least 8 characters long');
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error('New passwords do not match');
      return;
    }

    try {
      await changePassword({
        currentPassword,
        newPassword,
      }).unwrap();
      toast.success('Password changed successfully');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: unknown) {
      let message = 'Failed to change password. Please check your current password.';
      if (err && typeof err === 'object' && 'data' in err) {
        const data = (err as { data?: { message?: string } }).data;
        if (data?.message && typeof data.message === 'string') {
          message = data.message;
        }
      }
      toast.error(message);
    }
  };

  if (!user) {
    return <AdminSettingsSkeleton />;
  }

  const initials = ((firstName?.[0] || '') + (lastName?.[0] || '')).toUpperCase() || 'U';

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5">
            <Settings className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Account Settings
            </h1>
          </div>
          <p className="text-xs text-zinc-400">
            Manage your personal profile, credentials, and global appearance preferences.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge className={`text-xs uppercase ${roleBadge.className}`}>
            {roleBadge.label}
          </Badge>
          {activeWorkspaceSlug && (
            <Link href={`/${activeWorkspaceSlug}/settings`}>
              <Button variant="outline" size="sm" className="gap-1.5 text-xs border-zinc-800">
                <span>Workspace Settings</span>
                <ExternalLink className="h-3 w-3 text-zinc-500" />
              </Button>
            </Link>
          )}
        </div>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="profile" className="space-y-6">
        <TabsList className="bg-zinc-900 border border-zinc-800 p-1 rounded-md flex-wrap h-auto gap-1">
          <TabsTrigger value="profile" className="gap-2 text-xs">
            <User className="h-3.5 w-3.5" />
            <span>Profile</span>
          </TabsTrigger>
          <TabsTrigger value="security" className="gap-2 text-xs">
            <Shield className="h-3.5 w-3.5" />
            <span>Security & Password</span>
          </TabsTrigger>
          <TabsTrigger value="ai" className="gap-2 text-xs">
            <Bot className="h-3.5 w-3.5 text-indigo-400" />
            <span>AI Assistant</span>
          </TabsTrigger>
          <TabsTrigger value="preferences" className="gap-2 text-xs">
            <Palette className="h-3.5 w-3.5" />
            <span>Appearance</span>
          </TabsTrigger>
        </TabsList>

        {/* Profile Tab */}
        <TabsContent value="profile">
          <Card className="border-zinc-800 bg-zinc-900/80 max-w-2xl">
            <CardHeader>
              <CardTitle className="text-base text-zinc-100">Personal Details</CardTitle>
              <CardDescription className="text-xs text-zinc-400">
                Update how your name, avatar, and contact information appear across workspaces and projects.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleProfileSave} className="space-y-5">
                {/* Avatar Section */}
                <div className="flex flex-col sm:flex-row sm:items-center gap-4 p-4 rounded-md bg-zinc-950/40 border border-zinc-850">
                  <div className="relative group shrink-0">
                    <Avatar className="h-16 w-16 rounded-md border-2 border-zinc-700 bg-zinc-900 overflow-hidden shadow-sm">
                      {avatar ? (
                        <AvatarImage src={avatar} alt={firstName || 'User'} className="object-cover" />
                      ) : null}
                      <AvatarFallback className="text-base font-bold bg-zinc-800 text-zinc-200">
                        {initials}
                      </AvatarFallback>
                    </Avatar>
                  </div>

                  <div className="space-y-1.5 flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => fileInputRef.current?.click()}
                        className="gap-1.5 text-xs h-8 border-zinc-800 bg-zinc-900 hover:bg-zinc-800 text-zinc-200 cursor-pointer"
                      >
                        <Upload className="h-3.5 w-3.5" />
                        <span>Change Photo</span>
                      </Button>

                      {avatar && (
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={handleRemoveAvatar}
                          className="gap-1.5 text-xs h-8 text-rose-400 hover:text-rose-300 hover:bg-rose-950/20 cursor-pointer"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                          <span>Remove</span>
                        </Button>
                      )}
                    </div>
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept="image/png, image/jpeg, image/webp, image/gif"
                      onChange={handleAvatarFile}
                      className="hidden"
                    />
                    <p className="text-[11px] text-zinc-500">
                      Recommended: Square JPG, PNG, or WEBP under 2MB.
                    </p>
                  </div>
                </div>

                {/* Name Grid */}
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="space-y-1.5">
                    <Label htmlFor="firstName" className="text-xs text-zinc-300">First Name</Label>
                    <Input
                      id="firstName"
                      value={firstName}
                      onChange={(e) => setFirstName(e.target.value)}
                      className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="lastName" className="text-xs text-zinc-300">Last Name</Label>
                    <Input
                      id="lastName"
                      value={lastName}
                      onChange={(e) => setLastName(e.target.value)}
                      className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                    />
                  </div>
                </div>

                {/* Role / Job Title & Department */}
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="space-y-1.5">
                    <Label htmlFor="jobTitle" className="text-xs text-zinc-300">Job Title</Label>
                    <Input
                      id="jobTitle"
                      placeholder="e.g. Lead Software Architect"
                      value={jobTitle}
                      onChange={(e) => setJobTitle(e.target.value)}
                      className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="department" className="text-xs text-zinc-300">Department / Squad</Label>
                    <Input
                      id="department"
                      placeholder="e.g. Platform Engineering"
                      value={department}
                      onChange={(e) => setDepartment(e.target.value)}
                      className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                    />
                  </div>
                </div>

                {/* Email Address (Read-only) */}
                <div className="space-y-1.5">
                  <Label htmlFor="email" className="text-xs text-zinc-300">Email Address</Label>
                  <Input
                    id="email"
                    value={user?.email || ''}
                    disabled
                    className="bg-zinc-950/60 border-zinc-850 text-xs text-zinc-500 cursor-not-allowed"
                  />
                  <p className="text-[11px] text-zinc-500">
                    Email address is tied to your account identity and login credentials.
                  </p>
                </div>

                {/* Bio */}
                <div className="space-y-1.5">
                  <Label htmlFor="bio" className="text-xs text-zinc-300">Professional Bio</Label>
                  <Input
                    id="bio"
                    placeholder="Short bio or status summary..."
                    value={bio}
                    onChange={(e) => setBio(e.target.value)}
                    className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                  />
                </div>

                {/* Mobile Number */}
                <div className="space-y-1.5">
                  <Label htmlFor="mobileNumber" className="text-xs text-zinc-300">Contact Number</Label>
                  <Input
                    id="mobileNumber"
                    placeholder="+1 (555) 000-0000"
                    value={mobileNumber}
                    onChange={(e) => setMobileNumber(e.target.value)}
                    className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                  />
                </div>

                <Button
                  type="submit"
                  variant="primary"
                  size="sm"
                  isLoading={isUpdatingProfile}
                  className="gap-2 mt-2 text-xs"
                >
                  <Save className="h-3.5 w-3.5" />
                  <span>Save Changes</span>
                </Button>
              </form>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Security & Password Tab */}
        <TabsContent value="security">
          <Card className="border-zinc-800 bg-zinc-900/80 max-w-2xl">
            <CardHeader>
              <CardTitle className="text-base text-zinc-100">Change Password</CardTitle>
              <CardDescription className="text-xs text-zinc-400">
                Ensure your account uses a strong, unique password with at least 8 characters.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handlePasswordChange} className="space-y-4">
                <div className="space-y-1.5">
                  <Label htmlFor="currentPassword" className="text-xs text-zinc-300">Current Password</Label>
                  <Input
                    id="currentPassword"
                    type="password"
                    required
                    value={currentPassword}
                    onChange={(e) => setCurrentPassword(e.target.value)}
                    className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="newPassword" className="text-xs text-zinc-300">New Password</Label>
                  <Input
                    id="newPassword"
                    type="password"
                    required
                    minLength={8}
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="confirmPassword" className="text-xs text-zinc-300">Confirm New Password</Label>
                  <Input
                    id="confirmPassword"
                    type="password"
                    required
                    minLength={8}
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                  />
                </div>

                <Button
                  type="submit"
                  variant="primary"
                  size="sm"
                  isLoading={isChangingPassword}
                  className="gap-2 mt-2 text-xs"
                >
                  <KeyRound className="h-3.5 w-3.5" />
                  <span>Update Password</span>
                </Button>
              </form>
            </CardContent>
          </Card>
        </TabsContent>

        {/* AI Assistant Tab */}
        <TabsContent value="ai">
          <AISettings />
        </TabsContent>

        {/* Appearance Tab */}
        <TabsContent value="preferences">
          <Card className="border-zinc-800 bg-zinc-900/80 max-w-2xl">
            <CardHeader>
              <CardTitle className="text-base text-zinc-100">Appearance & Interface</CardTitle>
              <CardDescription className="text-xs text-zinc-400">
                Customize your visual theme and dashboard display preferences.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="flex items-center justify-between p-4 rounded-md bg-zinc-950/60 border border-zinc-850">
                <div className="space-y-1">
                  <h4 className="text-xs font-semibold text-zinc-200">Obsidian Monochrome Theme</h4>
                  <p className="text-[11px] text-zinc-400">
                    Switch between Deep Obsidian Dark and High-Contrast Light mode.
                  </p>
                </div>
                <ThemeToggle />
              </div>

              <div className="p-4 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-2">
                <h4 className="text-xs font-semibold text-zinc-200">Typography & Geometry</h4>
                <p className="text-[11px] text-zinc-400">
                  Orbito enforces crisp 4px border-radii (`rounded-md`), subtle 1px zinc borders, and dense typography optimized for engineering team productivity.
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
