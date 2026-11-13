'use client';

import * as React from 'react';
import { Eye, EyeOff, Copy, Check, Save } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface SecuritySsoTabProps {
  isSsoEnabled: boolean;
  onToggleSso: () => void;
  ssoProvider: string;
  setSsoProvider: (val: string) => void;
  ssoIssuerUrl: string;
  setSsoIssuerUrl: (val: string) => void;
  ssoClientId: string;
  setSsoClientId: (val: string) => void;
  ssoClientSecret: string;
  setSsoClientSecret: (val: string) => void;
  showSsoSecret: boolean;
  setShowSsoSecret: (val: boolean) => void;
  ssoDefaultRole: string;
  setSsoDefaultRole: (val: string) => void;
  sessionTimeout: string;
  setSessionTimeout: (val: string) => void;
  copiedCallback: boolean;
  onCopyCallback: () => void;
  onSave: () => void;
  isSaving: boolean;
}

export function SecuritySsoTab({
  isSsoEnabled,
  onToggleSso,
  ssoProvider,
  setSsoProvider,
  ssoIssuerUrl,
  setSsoIssuerUrl,
  ssoClientId,
  setSsoClientId,
  ssoClientSecret,
  setSsoClientSecret,
  showSsoSecret,
  setShowSsoSecret,
  ssoDefaultRole,
  setSsoDefaultRole,
  sessionTimeout,
  setSessionTimeout,
  copiedCallback,
  onCopyCallback,
  onSave,
  isSaving,
}: SecuritySsoTabProps) {
  return (
    <div className="space-y-6">
      <div className="border border-zinc-800/80 rounded-lg bg-zinc-900/30 divide-y divide-zinc-800/80">
        {/* Row 1: OIDC Master Switch */}
        <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-0.5 max-w-lg">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-zinc-200">OpenID Connect (OIDC) SSO</span>
              <span
                className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono ${
                  isSsoEnabled
                    ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800/60'
                    : 'bg-zinc-800 text-zinc-400 border border-zinc-700'
                }`}
              >
                {isSsoEnabled ? 'SSO Active' : 'Disabled'}
              </span>
            </div>
            <p className="text-xs text-zinc-400 leading-relaxed">
              Allow enterprise employees to log in using centralized Identity Providers (Google Workspace, Okta, Microsoft Azure AD, Keycloak).
            </p>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            <Switch
              checked={isSsoEnabled}
              onCheckedChange={onToggleSso}
              disabled={isSaving}
              aria-label="Toggle SSO"
            />
          </div>
        </div>

        {/* Row 2: Provider Presets */}
        <div className="p-4 space-y-3">
          <span className="text-xs font-medium text-zinc-300">Identity Provider Preset</span>
          <div className="flex flex-wrap gap-2">
            {[
              { name: 'Google Workspace', issuer: 'https://accounts.google.com' },
              { name: 'Okta', issuer: 'https://your-domain.okta.com' },
              { name: 'Microsoft Entra ID', issuer: 'https://login.microsoftonline.com/{tenant}/v2.0' },
              { name: 'Keycloak', issuer: 'https://keycloak.yourdomain.com/realms/orbito' },
            ].map((p) => (
              <button
                key={p.name}
                type="button"
                onClick={() => {
                  setSsoProvider(p.name);
                  setSsoIssuerUrl(p.issuer);
                }}
                className={`px-2.5 py-1 text-xs rounded-md border transition-colors cursor-pointer ${
                  ssoProvider === p.name
                    ? 'bg-zinc-800 border-zinc-600 text-white font-medium'
                    : 'bg-zinc-950 border-zinc-800 text-zinc-400 hover:text-zinc-200 hover:border-zinc-700'
                }`}
              >
                {p.name}
              </button>
            ))}
          </div>
        </div>

        {/* Row 3: OIDC Form Fields */}
        <div className="p-4 space-y-3">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
            <div className="space-y-1.5 sm:col-span-2">
              <label className="text-[11px] font-medium text-zinc-400">OIDC Issuer Discovery URL</label>
              <Input
                value={ssoIssuerUrl}
                onChange={(e) => setSsoIssuerUrl(e.target.value)}
                placeholder="https://accounts.google.com"
                className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400">Client Identifier (Client ID)</label>
              <Input
                value={ssoClientId}
                onChange={(e) => setSsoClientId(e.target.value)}
                placeholder="client-id-xyz.apps.googleusercontent.com"
                className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400">Client Secret</label>
              <div className="relative">
                <Input
                  type={showSsoSecret ? 'text' : 'password'}
                  value={ssoClientSecret}
                  onChange={(e) => setSsoClientSecret(e.target.value)}
                  placeholder="••••••••••••••••••••••••"
                  className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200 pr-8"
                />
                <button
                  type="button"
                  onClick={() => setShowSsoSecret(!showSsoSecret)}
                  className="absolute right-2 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-200 cursor-pointer"
                >
                  {showSsoSecret ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Row 4: Callback URL Display */}
        <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-0.5">
            <span className="text-sm font-medium text-zinc-200">OIDC Redirect Callback URI</span>
            <p className="text-xs text-zinc-400">
              Register this exact callback endpoint in your Identity Provider OAuth console.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <code className="px-2.5 py-1 rounded bg-zinc-950 border border-zinc-800 text-xs font-mono text-zinc-300">
              /api/auth/oidc/callback
            </code>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onCopyCallback}
              className="h-7 px-2 border-zinc-800 bg-zinc-950 hover:bg-zinc-850 text-zinc-300 cursor-pointer"
            >
              {copiedCallback ? <Check className="h-3 w-3 text-emerald-400" /> : <Copy className="h-3 w-3" />}
            </Button>
          </div>
        </div>

        {/* Row 5: Default SSO Role & Session TTL */}
        <div className="p-4 grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-zinc-300">Default Role for New SSO Users</label>
            <Select
              value={ssoDefaultRole}
              onValueChange={(val) => setSsoDefaultRole(val)}
            >
              <SelectTrigger className="w-full h-8 bg-zinc-950 border-zinc-800 text-xs text-zinc-200">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                <SelectItem value="MEMBER" className="text-xs">
                  MEMBER (Standard Workspace Contributor)
                </SelectItem>
                <SelectItem value="VIEWER" className="text-xs">
                  VIEWER (Read-Only Access)
                </SelectItem>
                <SelectItem value="MANAGER" className="text-xs">
                  MANAGER (Workspace Manager)
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-medium text-zinc-300">Session Inactivity Timeout</label>
            <Select
              value={sessionTimeout}
              onValueChange={(val) => setSessionTimeout(val)}
            >
              <SelectTrigger className="w-full h-8 bg-zinc-950 border-zinc-800 text-xs text-zinc-200">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                <SelectItem value="15" className="text-xs">15 Minutes (High Security)</SelectItem>
                <SelectItem value="60" className="text-xs">1 Hour (Standard)</SelectItem>
                <SelectItem value="240" className="text-xs">4 Hours</SelectItem>
                <SelectItem value="1440" className="text-xs">24 Hours</SelectItem>
                <SelectItem value="10080" className="text-xs">7 Days (Extended Remember Me)</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      {/* Save Button for SSO Tab */}
      <div className="flex justify-end">
        <Button
          type="button"
          size="sm"
          onClick={onSave}
          disabled={isSaving}
          className="h-8 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold cursor-pointer transition-colors shadow-xs"
        >
          <Save className="h-3.5 w-3.5 mr-1.5 text-zinc-950" />
          Save SSO & Security Settings
        </Button>
      </div>
    </div>
  );
}
