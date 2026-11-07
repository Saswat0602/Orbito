'use client';

import * as React from 'react';
import { useState } from 'react';
import {
  Sliders,
  ShieldCheck,
  Mail,
  KeyRound,
  CheckCircle2,
  AlertTriangle,
  Send,
  Lock,
  Globe,
  ToggleLeft,
  ToggleRight,
} from 'lucide-react';
import { toast } from 'sonner';

import { Role } from '@orbito/types';
import { RoleGuard } from '@/components/auth/RoleGuard';
import {
  useGetAdminConfigQuery,
  useSaveAdminConfigMutation,
  useTestAdminSmtpMutation,
} from '@/lib/store/api/adminApi';
import { useGetRegistrationStatusQuery } from '@/lib/store/api/authApi';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function AdminSettingsPage() {
  const { data: configData, isLoading: isConfigLoading } = useGetAdminConfigQuery();
  const { data: regStatus, refetch: refetchRegStatus } = useGetRegistrationStatusQuery();

  const [saveConfig, { isLoading: isSaving }] = useSaveAdminConfigMutation();
  const [testSmtp, { isLoading: isTestingSmtp }] = useTestAdminSmtpMutation();

  // Find registration setting
  const regSetting = configData?.settings.find((s) => s.key === 'enablePublicRegistration');
  const isPublicRegEnabled = regSetting ? regSetting.value === 'true' : (regStatus?.enabled ?? false);

  const handleToggleRegistration = async () => {
    const nextVal = (!isPublicRegEnabled).toString();
    try {
      await saveConfig({
        settings: [
          {
            key: 'enablePublicRegistration',
            value: nextVal,
            category: 'security',
            description: 'Allow new users to self-register on the platform',
          },
        ],
      }).unwrap();
      await refetchRegStatus();
      toast.success(
        nextVal === 'true'
          ? 'Public user registration is now ENABLED'
          : 'Public user registration is now RESTRICTED (invitation only)'
      );
    } catch (err: unknown) {
      toast.error('Failed to update registration setting');
    }
  };

  const handleTestSmtp = async () => {
    try {
      const res = await testSmtp().unwrap();
      toast.success(res.message || 'SMTP connection and authentication verified successfully!');
    } catch (err: unknown) {
      const msg = err && typeof err === 'object' && 'data' in err
        ? (err as { data: { message?: string } }).data?.message
        : 'SMTP connection failed';
      toast.error(msg || 'SMTP test failed. Verify host, port, and credentials.');
    }
  };

  return (
    <RoleGuard requiredRole={Role.SUPER_ADMIN}>
      <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-zinc-800/80">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Sliders className="h-6 w-6 text-zinc-100" />
              <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
                System & Security Settings
              </h1>
              <Badge variant="destructive" className="text-[10px] uppercase font-bold tracking-wider">
                Root Config
              </Badge>
            </div>
            <p className="text-xs text-zinc-400">
              Manage enterprise SSO/OIDC integration, public user registration access, and mail delivery topology.
            </p>
          </div>
        </div>

        <div className="grid grid-cols-1 gap-6 max-w-4xl">
          {/* Public Registration Gate */}
          <Card className="border-zinc-800 bg-zinc-900/60">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2.5">
                  <div className="h-8 w-8 rounded-md bg-blue-950/60 border border-blue-900/60 flex items-center justify-center text-blue-400">
                    <Globe className="h-4 w-4" />
                  </div>
                  <div>
                    <CardTitle className="text-sm font-semibold text-zinc-100">
                      Public Registration Gate
                    </CardTitle>
                    <CardDescription className="text-xs text-zinc-400">
                      Control whether external visitors can create accounts without an invitation.
                    </CardDescription>
                  </div>
                </div>

                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleToggleRegistration}
                  disabled={isSaving}
                  className={`h-8 text-xs border-zinc-800 cursor-pointer ${
                    isPublicRegEnabled
                      ? 'bg-emerald-950/40 text-emerald-400 border-emerald-800/80 hover:bg-emerald-900/50'
                      : 'bg-zinc-800/60 text-zinc-400 hover:text-zinc-200'
                  }`}
                >
                  {isPublicRegEnabled ? (
                    <>
                      <ToggleRight className="h-4 w-4 mr-1.5 text-emerald-400" />
                      Public Registration: Enabled
                    </>
                  ) : (
                    <>
                      <ToggleLeft className="h-4 w-4 mr-1.5 text-zinc-500" />
                      Public Registration: Restricted
                    </>
                  )}
                </Button>
              </div>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 text-xs text-zinc-400">
                {isPublicRegEnabled ? (
                  <p className="text-emerald-400/90 font-medium">
                    ✓ Self-service signups are active at <code>/register</code>. Any user can register their own tenant organization.
                  </p>
                ) : (
                  <p className="text-amber-400/90 font-medium">
                    ⚠ Public signups are closed. Only users with a direct invite token from an existing workspace or organization can onboard.
                  </p>
                )}
              </div>
            </CardContent>
          </Card>

          {/* OpenID Connect (OIDC) / Enterprise SSO */}
          <Card className="border-zinc-800 bg-zinc-900/60">
            <CardHeader>
              <div className="flex items-center gap-2.5">
                <div className="h-8 w-8 rounded-md bg-purple-950/60 border border-purple-900/60 flex items-center justify-center text-purple-400">
                  <KeyRound className="h-4 w-4" />
                </div>
                <div>
                  <CardTitle className="text-sm font-semibold text-zinc-100">
                    OpenID Connect (OIDC) / Enterprise SSO
                  </CardTitle>
                  <CardDescription className="text-xs text-zinc-400">
                    Single Sign-On provider discovery and authentication pipeline status.
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4 pt-0">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1">
                  <span className="text-[10px] uppercase font-bold text-zinc-500">SSO Provider Status</span>
                  <div className="flex items-center gap-2 font-medium text-emerald-400">
                    <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />
                    <span>OIDC Exchange Active</span>
                  </div>
                </div>

                <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1">
                  <span className="text-[10px] uppercase font-bold text-zinc-500">Callback Endpoint</span>
                  <div className="font-mono text-zinc-300 truncate">
                    /api/auth/oidc/callback
                  </div>
                </div>

                <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1">
                  <span className="text-[10px] uppercase font-bold text-zinc-500">Security Cookie Architecture</span>
                  <div className="font-medium text-zinc-300">
                    HttpOnly, SameSite=Lax, Secure
                  </div>
                </div>

                <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1">
                  <span className="text-[10px] uppercase font-bold text-zinc-500">Default Role on SSO Onboard</span>
                  <div className="font-medium text-zinc-300">
                    MEMBER (Workspace Scoped)
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Mail Delivery & SMTP Configuration */}
          <Card className="border-zinc-800 bg-zinc-900/60">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2.5">
                  <div className="h-8 w-8 rounded-md bg-amber-950/60 border border-amber-900/60 flex items-center justify-center text-amber-400">
                    <Mail className="h-4 w-4" />
                  </div>
                  <div>
                    <CardTitle className="text-sm font-semibold text-zinc-100">
                      Mail Delivery & SMTP Service
                    </CardTitle>
                    <CardDescription className="text-xs text-zinc-400">
                      Outgoing transactional email delivery (invitations, task notifications, password resets).
                    </CardDescription>
                  </div>
                </div>

                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleTestSmtp}
                  disabled={isTestingSmtp}
                  className="h-8 text-xs border-zinc-800 text-zinc-300 hover:bg-zinc-850"
                >
                  <Send className="h-3.5 w-3.5 mr-1.5 text-zinc-400" />
                  {isTestingSmtp ? 'Testing Connection...' : 'Test SMTP Connection'}
                </Button>
              </div>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1">
                  <span className="text-[10px] uppercase font-bold text-zinc-500">SMTP Host Source</span>
                  <div className="font-medium text-zinc-300">
                    {configData?.smtpSources?.hostSource || 'Environment (.env)'}
                  </div>
                </div>

                <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1">
                  <span className="text-[10px] uppercase font-bold text-zinc-500">SMTP Port Source</span>
                  <div className="font-medium text-zinc-300">
                    {configData?.smtpSources?.portSource || 'Environment (.env)'}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </RoleGuard>
  );
}

