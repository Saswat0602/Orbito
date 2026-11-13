'use client';

import * as React from 'react';
import {
  Mail,
  Send,
  Save,
  Loader2,
  Eye,
  EyeOff,
  CheckCircle2,
  Inbox,
  BellRing,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

interface MailSmtpTabProps {
  smtpHost: string;
  setSmtpHost: (val: string) => void;
  smtpPort: string;
  setSmtpPort: (val: string) => void;
  smtpUser: string;
  setSmtpUser: (val: string) => void;
  smtpPass: string;
  setSmtpPass: (val: string) => void;
  showSmtpPass: boolean;
  setShowSmtpPass: (val: boolean) => void;
  smtpFrom: string;
  setSmtpFrom: (val: string) => void;
  smtpFromName: string;
  setSmtpFromName: (val: string) => void;
  hostSource?: string;
  testRecipientEmail: string;
  setTestRecipientEmail: (val: string) => void;
  inboundEmailEnabled: boolean;
  setInboundEmailEnabled: (val: boolean) => void;
  inboundEmailDomain: string;
  setInboundEmailDomain: (val: string) => void;
  emailNotifyInvitations: boolean;
  setEmailNotifyInvitations: (val: boolean) => void;
  emailNotifyAssignments: boolean;
  setEmailNotifyAssignments: (val: boolean) => void;
  emailNotifyMentions: boolean;
  setEmailNotifyMentions: (val: boolean) => void;
  emailNotifyOverdue: boolean;
  setEmailNotifyOverdue: (val: boolean) => void;
  onTestSmtp: () => void;
  onSendTestEmail: () => void;
  isTestingSmtp: boolean;
  onSave: () => void;
  isSaving: boolean;
}

export function MailSmtpTab({
  smtpHost,
  setSmtpHost,
  smtpPort,
  setSmtpPort,
  smtpUser,
  setSmtpUser,
  smtpPass,
  setSmtpPass,
  showSmtpPass,
  setShowSmtpPass,
  smtpFrom,
  setSmtpFrom,
  smtpFromName,
  setSmtpFromName,
  hostSource,
  testRecipientEmail,
  setTestRecipientEmail,
  inboundEmailEnabled,
  setInboundEmailEnabled,
  inboundEmailDomain,
  setInboundEmailDomain,
  emailNotifyInvitations,
  setEmailNotifyInvitations,
  emailNotifyAssignments,
  setEmailNotifyAssignments,
  emailNotifyMentions,
  setEmailNotifyMentions,
  emailNotifyOverdue,
  setEmailNotifyOverdue,
  onTestSmtp,
  onSendTestEmail,
  isTestingSmtp,
  onSave,
  isSaving,
}: MailSmtpTabProps) {
  return (
    <div className="space-y-6">
      <div className="border border-zinc-800/80 rounded-lg bg-zinc-900/30 divide-y divide-zinc-800/80">
        {/* Row 1: Header / Status */}
        <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-0.5">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-zinc-200">Outbound Mail Pipeline</span>
              <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono bg-zinc-800 text-zinc-300 border border-zinc-700">
                Source: {hostSource || 'Environment (.env)'}
              </span>
            </div>
            <p className="text-xs text-zinc-400">
              Transactional email delivery for workspace invitations, password recovery, and assignment pings.
            </p>
          </div>
        </div>

        {/* Row 2: Host & Port & Credentials */}
        <div className="p-4 space-y-3">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs">
            <div className="space-y-1.5 sm:col-span-2">
              <label className="text-[11px] font-medium text-zinc-400">SMTP Host / Server</label>
              <Input
                value={smtpHost}
                onChange={(e) => setSmtpHost(e.target.value)}
                placeholder="smtp.sendgrid.net or smtp.mailgun.org"
                className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400">SMTP Port</label>
              <Input
                value={smtpPort}
                onChange={(e) => setSmtpPort(e.target.value)}
                placeholder="587"
                className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400">SMTP Username</label>
              <Input
                value={smtpUser}
                onChange={(e) => setSmtpUser(e.target.value)}
                placeholder="apikey or user@orbito.io"
                className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
              />
            </div>

            <div className="space-y-1.5 sm:col-span-2">
              <label className="text-[11px] font-medium text-zinc-400">SMTP Password / API Key</label>
              <div className="relative">
                <Input
                  type={showSmtpPass ? 'text' : 'password'}
                  value={smtpPass}
                  onChange={(e) => setSmtpPass(e.target.value)}
                  placeholder="••••••••••••••••••••••••"
                  className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200 pr-8"
                />
                <button
                  type="button"
                  onClick={() => setShowSmtpPass(!showSmtpPass)}
                  className="absolute right-2 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-200 cursor-pointer"
                >
                  {showSmtpPass ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                </button>
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400">Outbound From Address</label>
              <Input
                value={smtpFrom}
                onChange={(e) => setSmtpFrom(e.target.value)}
                placeholder="noreply@orbito.io"
                className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
              />
            </div>

            <div className="space-y-1.5 sm:col-span-2">
              <label className="text-[11px] font-medium text-zinc-400">Outbound Sender Name</label>
              <Input
                value={smtpFromName}
                onChange={(e) => setSmtpFromName(e.target.value)}
                placeholder="Orbito Platform"
                className="h-8 bg-zinc-950 border-zinc-800 text-xs text-zinc-200"
              />
            </div>
          </div>
        </div>

        {/* Row 3: Live Test Email Dispatch */}
        <div className="p-4 space-y-3">
          <div className="space-y-0.5">
            <span className="text-sm font-medium text-zinc-200">Live Test Email Dispatch</span>
            <p className="text-xs text-zinc-400">
              Deliver a physical test message to an inbox to verify TLS handshakes, reverse DNS, and SPF/DKIM delivery.
            </p>
          </div>
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2.5 max-w-xl">
            <Input
              type="email"
              value={testRecipientEmail}
              onChange={(e) => setTestRecipientEmail(e.target.value)}
              placeholder="admin@yourcompany.com"
              className="h-8 bg-zinc-950 border-zinc-800 text-xs text-zinc-200 flex-1 font-mono"
            />
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onSendTestEmail}
              disabled={isTestingSmtp}
              className="h-8 text-xs border-zinc-700 bg-zinc-900 hover:bg-zinc-800 text-zinc-100 font-medium cursor-pointer shrink-0"
            >
              {isTestingSmtp ? (
                <>
                  <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin text-zinc-300" />
                  Sending...
                </>
              ) : (
                <>
                  <Send className="h-3.5 w-3.5 mr-1.5 text-zinc-300" />
                  Send Test Email
                </>
              )}
            </Button>
          </div>
        </div>

        {/* Row 4: Inbound Email Task Creation */}
        <div className="p-4 space-y-3">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="space-y-0.5 max-w-lg">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-zinc-200">Inbound Email Task Creation</span>
                <span
                  className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono ${
                    inboundEmailEnabled
                      ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800/60'
                      : 'bg-zinc-800 text-zinc-400 border border-zinc-700'
                  }`}
                >
                  {inboundEmailEnabled ? 'Inbound Active' : 'Disabled'}
                </span>
              </div>
              <p className="text-xs text-zinc-400 leading-relaxed">
                Allow members to create tasks automatically by forwarding emails to project-specific addresses.
              </p>
            </div>
            <div className="flex items-center gap-3 shrink-0">
              <Switch
                checked={inboundEmailEnabled}
                onCheckedChange={setInboundEmailEnabled}
                aria-label="Toggle Inbound Email"
              />
            </div>
          </div>

          {inboundEmailEnabled && (
            <div className="pt-2 max-w-md space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400">Inbound Mail Domain</label>
              <Input
                value={inboundEmailDomain}
                onChange={(e) => setInboundEmailDomain(e.target.value)}
                placeholder="tasks.orbito.io"
                className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
              />
            </div>
          )}
        </div>

        {/* Row 5: Transactional Notification Event Toggles */}
        <div className="p-4 space-y-3">
          <div className="space-y-0.5">
            <span className="text-sm font-medium text-zinc-200">Transactional Email Event Triggers</span>
            <p className="text-xs text-zinc-400">
              Configure which platform events dispatch transactional emails to teammates.
            </p>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
            <div className="flex items-center justify-between p-2.5 rounded-md bg-zinc-950 border border-zinc-800/80">
              <div className="space-y-0.5 pr-2">
                <span className="text-xs font-medium text-zinc-200">Workspace & Tenant Invitations</span>
                <p className="text-[11px] text-zinc-400">Deliver invitation tokens to new team members.</p>
              </div>
              <Switch
                checked={emailNotifyInvitations}
                onCheckedChange={setEmailNotifyInvitations}
                aria-label="Toggle Invitation Emails"
              />
            </div>

            <div className="flex items-center justify-between p-2.5 rounded-md bg-zinc-950 border border-zinc-800/80">
              <div className="space-y-0.5 pr-2">
                <span className="text-xs font-medium text-zinc-200">Task Assignment Alerts</span>
                <p className="text-[11px] text-zinc-400">Notify assignees when tasks are delegated to them.</p>
              </div>
              <Switch
                checked={emailNotifyAssignments}
                onCheckedChange={setEmailNotifyAssignments}
                aria-label="Toggle Assignment Emails"
              />
            </div>

            <div className="flex items-center justify-between p-2.5 rounded-md bg-zinc-950 border border-zinc-800/80">
              <div className="space-y-0.5 pr-2">
                <span className="text-xs font-medium text-zinc-200">Comment @Mention Alerts</span>
                <p className="text-[11px] text-zinc-400">Notify users when explicitly mentioned in discussions.</p>
              </div>
              <Switch
                checked={emailNotifyMentions}
                onCheckedChange={setEmailNotifyMentions}
                aria-label="Toggle Mention Emails"
              />
            </div>

            <div className="flex items-center justify-between p-2.5 rounded-md bg-zinc-950 border border-zinc-800/80">
              <div className="space-y-0.5 pr-2">
                <span className="text-xs font-medium text-zinc-200">Overdue Task Warnings</span>
                <p className="text-[11px] text-zinc-400">Send reminder digests for tasks nearing or past due dates.</p>
              </div>
              <Switch
                checked={emailNotifyOverdue}
                onCheckedChange={setEmailNotifyOverdue}
                aria-label="Toggle Overdue Emails"
              />
            </div>
          </div>
        </div>
      </div>

      {/* Action Bar for SMTP */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={onTestSmtp}
          disabled={isTestingSmtp}
          className="h-8 text-xs border-zinc-800 bg-zinc-950/60 hover:bg-zinc-850 text-zinc-300 cursor-pointer"
        >
          <Mail className="h-3.5 w-3.5 mr-1.5 text-zinc-400" />
          {isTestingSmtp ? 'Verifying Handshake...' : 'Verify Handshake Only'}
        </Button>

        <Button
          type="button"
          size="sm"
          onClick={onSave}
          disabled={isSaving}
          className="h-8 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold cursor-pointer transition-colors shadow-xs"
        >
          <Save className="h-3.5 w-3.5 mr-1.5 text-zinc-950" />
          Save Mail Configuration
        </Button>
      </div>
    </div>
  );
}
