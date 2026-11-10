'use client';

import * as React from 'react';
import { useState, useEffect } from 'react';
import {
  Mail,
  RefreshCw,
  Plus,
  Trash2,
  CheckCircle2,
  AlertCircle,
  Clock,
  Sparkles,
  ArrowUpRight,
  ShieldCheck,
  Send,
  Filter,
  Check,
} from 'lucide-react';
import { toast } from 'sonner';
import { TaskPriority, TaskType } from '@orbito/types';
import type { WorkspaceMember } from '@orbito/types';
import type {
  ProjectInbox,
  InboxMessage,
  InboxRule,
  MessageStatus,
  SetupEmailDto,
} from '@/types';
import {
  useGetProjectInboxQuery,
  useUpdateProjectInboxMutation,
  useSetupEmailAccountMutation,
  useTestEmailConnectionMutation,
  useSyncInboxMutation,
  useGetInboxMessagesQuery,
  useConvertMessageToTaskMutation,
  useIgnoreMessageMutation,
  useGetInboxRulesQuery,
  useCreateInboxRuleMutation,
  useDeleteInboxRuleMutation,
} from '@/lib/store/api/inboxApi';
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
import { Badge } from '@/components/ui/badge';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';

interface ProjectInboxTabProps {
  projectId: string;
  workspaceMembers?: WorkspaceMember[];
  canEdit: boolean;
}

const PROVIDER_PRESETS = [
  {
    name: 'Google Workspace / Gmail',
    host: 'imap.gmail.com',
    port: 993,
    ssl: true,
  },
  {
    name: 'Microsoft 365 / Outlook',
    host: 'outlook.office365.com',
    port: 993,
    ssl: true,
  },
  {
    name: 'Custom IMAP',
    host: '',
    port: 993,
    ssl: true,
  },
];

export function ProjectInboxTab({
  projectId,
  workspaceMembers = [],
  canEdit,
}: ProjectInboxTabProps) {
  const [activeSection, setActiveSection] = useState<'config' | 'rules' | 'messages'>('config');

  // Queries
  const { data: inbox, isLoading: isInboxLoading } = useGetProjectInboxQuery(projectId, {
    skip: !projectId,
  });
  const { data: rules = [], isLoading: isRulesLoading } = useGetInboxRulesQuery(projectId, {
    skip: !projectId,
  });
  const { data: messages = [], isLoading: isMessagesLoading } = useGetInboxMessagesQuery(
    { projectId },
    { skip: !projectId }
  );

  // Mutations
  const [setupEmailAccount, { isLoading: isSavingEmail }] = useSetupEmailAccountMutation();
  const [updateInbox, { isLoading: isSavingInbox }] = useUpdateProjectInboxMutation();
  const [testConnection, { isLoading: isTesting }] = useTestEmailConnectionMutation();
  const [syncInbox, { isLoading: isSyncing }] = useSyncInboxMutation();
  const [convertMessage, { isLoading: isConverting }] = useConvertMessageToTaskMutation();
  const [ignoreMessage] = useIgnoreMessageMutation();
  const [createRule, { isLoading: isCreatingRule }] = useCreateInboxRuleMutation();
  const [deleteRule, { isLoading: isDeletingRule }] = useDeleteInboxRuleMutation();

  // IMAP Form State
  const [emailAddress, setEmailAddress] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [imapHost, setImapHost] = useState('imap.gmail.com');
  const [imapPort, setImapPort] = useState(993);
  const [imapUsername, setImapUsername] = useState('');
  const [imapPassword, setImapPassword] = useState('');
  const [imapUseSsl, setImapUseSsl] = useState(true);
  const [imapFolder, setImapFolder] = useState('INBOX');

  // Ingestion Preferences
  const [autoCreateTask, setAutoCreateTask] = useState(true);
  const [defaultTaskType, setDefaultTaskType] = useState<TaskType>(TaskType.TASK);
  const [defaultPriority, setDefaultPriority] = useState<TaskPriority>(TaskPriority.MEDIUM);
  const [defaultAssigneeId, setDefaultAssigneeId] = useState('');

  // Rule Creation Dialog State
  const [isRuleModalOpen, setIsRuleModalOpen] = useState(false);
  const [ruleName, setRuleName] = useState('');
  const [senderContains, setSenderContains] = useState('');
  const [subjectContains, setSubjectContains] = useState('');
  const [rulePriority, setRulePriority] = useState<TaskPriority>(TaskPriority.HIGH);
  const [ruleAssigneeId, setRuleAssigneeId] = useState('');
  const [stopOnMatch, setStopOnMatch] = useState(false);

  // Populate form from existing inbox configuration
  useEffect(() => {
    if (inbox) {
      setAutoCreateTask(inbox.autoCreateTask ?? true);
      setDefaultTaskType(inbox.defaultTaskType || TaskType.TASK);
      setDefaultPriority(inbox.defaultPriority || TaskPriority.MEDIUM);
      setDefaultAssigneeId(inbox.defaultAssigneeId || '');

      if (inbox.emailAccount) {
        setEmailAddress(inbox.emailAccount.emailAddress || '');
        setDisplayName(inbox.emailAccount.displayName || '');
        setImapHost(inbox.emailAccount.imapHost || 'imap.gmail.com');
        setImapPort(inbox.emailAccount.imapPort || 993);
        setImapUsername(inbox.emailAccount.imapUsername || '');
        setImapUseSsl(inbox.emailAccount.imapUseSsl ?? true);
        setImapFolder(inbox.emailAccount.imapFolder || 'INBOX');
      }
    }
  }, [inbox]);

  const handleApplyPreset = (preset: (typeof PROVIDER_PRESETS)[number]) => {
    setImapHost(preset.host);
    setImapPort(preset.port);
    setImapUseSsl(preset.ssl);
  };

  const handleSaveEmailAccount = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!emailAddress || !imapHost || !imapUsername) {
      toast.error('Please enter email address, IMAP host, and username');
      return;
    }

    try {
      const emailDto: SetupEmailDto = {
        emailAddress: emailAddress.trim(),
        displayName: displayName.trim() || undefined,
        imapHost: imapHost.trim(),
        imapPort: Number(imapPort),
        imapUsername: imapUsername.trim(),
        imapPassword: imapPassword || undefined,
        imapUseSsl,
        imapFolder: imapFolder.trim() || 'INBOX',
        smtpHost: imapHost.includes('gmail')
          ? 'smtp.gmail.com'
          : imapHost.includes('office365')
          ? 'smtp.office365.com'
          : imapHost.replace('imap.', 'smtp.'),
        smtpPort: 587,
        smtpUsername: imapUsername.trim(),
        smtpPassword: imapPassword || undefined,
        smtpRequireTls: true,
      };

      await setupEmailAccount({ projectId, dto: emailDto }).unwrap();

      await updateInbox({
        projectId,
        dto: {
          autoCreateTask,
          defaultTaskType,
          defaultPriority,
          defaultAssigneeId: defaultAssigneeId || undefined,
        },
      }).unwrap();

      toast.success('Inbound email account & preferences saved');
    } catch {
      toast.error('Failed to save email configuration');
    }
  };

  const handleTestConnection = async () => {
    if (!inbox?.emailAccount?.id) {
      toast.info('Save account configuration first before testing');
      return;
    }
    try {
      const res = await testConnection({
        projectId,
        accountId: inbox.emailAccount.id,
      }).unwrap();
      if (res.success) {
        toast.success(res.message || 'IMAP connection successful');
      } else {
        toast.error(res.message || 'Connection test failed');
      }
    } catch {
      toast.error('Connection test failed. Check host credentials and port.');
    }
  };

  const handleSyncNow = async () => {
    try {
      await syncInbox(projectId).unwrap();
      toast.success('Email sync job queued successfully');
    } catch {
      toast.error('Failed to queue email sync job');
    }
  };

  const handleCreateRule = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!ruleName.trim()) {
      toast.error('Rule name is required');
      return;
    }

    const conditions: Record<string, unknown> = {};
    if (senderContains.trim()) {
      conditions.from = { contains: senderContains.trim() };
    }
    if (subjectContains.trim()) {
      conditions.subject = { contains: subjectContains.trim() };
    }

    const actions: Record<string, unknown> = {
      setPriority: rulePriority,
    };
    if (ruleAssigneeId) {
      actions.assignTo = ruleAssigneeId;
    }

    try {
      await createRule({
        projectId,
        dto: {
          name: ruleName.trim(),
          conditions,
          actions,
          stopOnMatch,
          enabled: true,
        },
      }).unwrap();

      toast.success(`Rule "${ruleName.trim()}" created`);
      setIsRuleModalOpen(false);
      setRuleName('');
      setSenderContains('');
      setSubjectContains('');
    } catch {
      toast.error('Failed to create inbound rule');
    }
  };

  const handleDeleteRule = async (ruleId: string, name: string) => {
    try {
      await deleteRule({ projectId, ruleId }).unwrap();
      toast.success(`Removed rule "${name}"`);
    } catch {
      toast.error('Failed to delete rule');
    }
  };

  const handleConvertMessage = async (messageId: string) => {
    try {
      await convertMessage({ projectId, messageId }).unwrap();
      toast.success('Converted inbound email to task successfully');
    } catch {
      toast.error('Failed to convert message to task');
    }
  };

  const handleIgnoreMessage = async (messageId: string) => {
    try {
      await ignoreMessage({ projectId, messageId }).unwrap();
      toast.success('Marked message as ignored');
    } catch {
      toast.error('Failed to ignore message');
    }
  };

  const isConfigured = Boolean(inbox?.emailAccount?.emailAddress);
  const pendingCount = messages.filter((m) => m.status === 'PENDING').length;

  return (
    <div className="space-y-6">
      {/* Top Overview & Action Banner */}
      <Card className="border-zinc-800 bg-zinc-900/80">
        <CardHeader className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-zinc-800/80">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Mail className="h-4 w-4 text-zinc-300" />
              <CardTitle className="text-base text-zinc-100">
                Inbound Email & Auto-Creation
              </CardTitle>
              {isConfigured ? (
                <Badge
                  variant="outline"
                  className="text-[10px] bg-emerald-950/60 text-emerald-300 border-emerald-800/60 font-mono"
                >
                  <ShieldCheck className="h-3 w-3 mr-1 inline" />
                  Connected
                </Badge>
              ) : (
                <Badge
                  variant="outline"
                  className="text-[10px] bg-zinc-800 text-zinc-400 border-zinc-750 font-mono"
                >
                  Not Configured
                </Badge>
              )}
            </div>
            <CardDescription className="text-xs text-zinc-400">
              Ingest support emails and customer tickets directly into tasks
              using IMAP and routing automation rules.
            </CardDescription>
          </div>

          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={!isConfigured || isSyncing}
              onClick={handleSyncNow}
              className="gap-1.5 text-xs border-zinc-700 bg-zinc-950/80 hover:bg-zinc-800 text-zinc-200"
            >
              <RefreshCw className={`h-3.5 w-3.5 ${isSyncing ? 'animate-spin' : ''}`} />
              <span>{isSyncing ? 'Syncing...' : 'Sync Now'}</span>
            </Button>
          </div>
        </CardHeader>

        {/* Section Navigation Tabs */}
        <div className="px-6 pt-4 flex items-center gap-2 border-b border-zinc-800/60 overflow-x-auto">
          <button
            type="button"
            onClick={() => setActiveSection('config')}
            className={`pb-3 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 ${
              activeSection === 'config'
                ? 'border-zinc-100 text-zinc-100'
                : 'border-transparent text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <span>Mailbox Connection</span>
          </button>

          <button
            type="button"
            onClick={() => setActiveSection('rules')}
            className={`pb-3 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 ${
              activeSection === 'rules'
                ? 'border-zinc-100 text-zinc-100'
                : 'border-transparent text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <span>Automation Rules</span>
            <span className="text-[10px] font-mono px-1.5 py-0.2 rounded-full bg-zinc-800 text-zinc-300">
              {rules.length}
            </span>
          </button>

          <button
            type="button"
            onClick={() => setActiveSection('messages')}
            className={`pb-3 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 ${
              activeSection === 'messages'
                ? 'border-zinc-100 text-zinc-100'
                : 'border-transparent text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <span>Incoming Feed</span>
            {pendingCount > 0 && (
              <span className="text-[10px] font-mono px-1.5 py-0.2 rounded-full bg-amber-950 text-amber-300 border border-amber-800">
                {pendingCount} new
              </span>
            )}
          </button>
        </div>

        <CardContent className="pt-6">
          {/* SECTION 1: MAILBOX CONNECTION & PREFERENCES */}
          {activeSection === 'config' && (
            <form onSubmit={handleSaveEmailAccount} className="space-y-6 max-w-3xl">
              {/* Presets */}
              <div className="space-y-2">
                <Label className="text-xs text-zinc-300">Quick Provider Presets</Label>
                <div className="flex items-center gap-2 flex-wrap">
                  {PROVIDER_PRESETS.map((preset) => (
                    <Button
                      key={preset.name}
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleApplyPreset(preset)}
                      className={`text-xs ${
                        imapHost === preset.host
                          ? 'border-zinc-500 bg-zinc-800 text-zinc-100'
                          : 'border-zinc-800 bg-zinc-950/60 text-zinc-400 hover:text-zinc-200'
                      }`}
                    >
                      {preset.name}
                    </Button>
                  ))}
                </div>
              </div>

              {/* IMAP Credentials Grid */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <Label htmlFor="email" className="text-xs text-zinc-300">
                    Mailbox Email Address
                  </Label>
                  <Input
                    id="email"
                    type="email"
                    value={emailAddress}
                    onChange={(e) => setEmailAddress(e.target.value)}
                    placeholder="support@acme.com"
                    disabled={!canEdit}
                    className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="dispName" className="text-xs text-zinc-300">
                    Sender Display Name (Optional)
                  </Label>
                  <Input
                    id="dispName"
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                    placeholder="Acme Support Desk"
                    disabled={!canEdit}
                    className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="host" className="text-xs text-zinc-300">
                    IMAP Host
                  </Label>
                  <Input
                    id="host"
                    value={imapHost}
                    onChange={(e) => setImapHost(e.target.value)}
                    placeholder="imap.gmail.com"
                    disabled={!canEdit}
                    className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="port" className="text-xs text-zinc-300">
                    IMAP Port
                  </Label>
                  <Input
                    id="port"
                    type="number"
                    value={imapPort}
                    onChange={(e) => setImapPort(Number(e.target.value))}
                    disabled={!canEdit}
                    className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="username" className="text-xs text-zinc-300">
                    IMAP Username
                  </Label>
                  <Input
                    id="username"
                    value={imapUsername}
                    onChange={(e) => setImapUsername(e.target.value)}
                    placeholder="support@acme.com"
                    disabled={!canEdit}
                    className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="password" className="text-xs text-zinc-300">
                    IMAP App Password
                  </Label>
                  <Input
                    id="password"
                    type="password"
                    value={imapPassword}
                    onChange={(e) => setImapPassword(e.target.value)}
                    placeholder={isConfigured ? '••••••••••••' : 'App-specific password'}
                    disabled={!canEdit}
                    className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                  />
                </div>
              </div>

              {/* Ingestion Behavior */}
              <div className="pt-4 border-t border-zinc-800/80 space-y-4">
                <h4 className="text-xs font-semibold text-zinc-200">
                  Task Auto-Creation Preferences
                </h4>

                <div className="flex items-center gap-3">
                  <input
                    id="autoTask"
                    type="checkbox"
                    checked={autoCreateTask}
                    onChange={(e) => setAutoCreateTask(e.target.checked)}
                    disabled={!canEdit}
                    className="h-4 w-4 rounded border-zinc-700 bg-zinc-950 text-indigo-600 focus:ring-0"
                  />
                  <Label htmlFor="autoTask" className="text-xs text-zinc-300 cursor-pointer">
                    Automatically create tasks from incoming emails
                  </Label>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
                  <div className="space-y-1.5">
                    <Label className="text-xs text-zinc-300">Default Task Type</Label>
                    <select
                      value={defaultTaskType}
                      onChange={(e) => setDefaultTaskType(e.target.value as TaskType)}
                      disabled={!canEdit}
                      className="w-full h-9 rounded-md bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-200"
                    >
                      <option value="TASK">Task</option>
                      <option value="BUG">Bug</option>
                      <option value="STORY">Story</option>
                      <option value="EPIC">Epic</option>
                    </select>
                  </div>

                  <div className="space-y-1.5">
                    <Label className="text-xs text-zinc-300">Default Priority</Label>
                    <select
                      value={defaultPriority}
                      onChange={(e) => setDefaultPriority(e.target.value as TaskPriority)}
                      disabled={!canEdit}
                      className="w-full h-9 rounded-md bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-200"
                    >
                      <option value="LOW">Low</option>
                      <option value="MEDIUM">Medium</option>
                      <option value="HIGH">High</option>
                      <option value="URGENT">Urgent</option>
                    </select>
                  </div>

                  <div className="space-y-1.5">
                    <Label className="text-xs text-zinc-300">Default Assignee</Label>
                    <select
                      value={defaultAssigneeId}
                      onChange={(e) => setDefaultAssigneeId(e.target.value)}
                      disabled={!canEdit}
                      className="w-full h-9 rounded-md bg-zinc-950 border border-zinc-800 px-3 text-xs text-zinc-200"
                    >
                      <option value="">Unassigned</option>
                      {workspaceMembers.map((m) => {
                        const name =
                          m.user?.firstName && m.user?.lastName
                            ? `${m.user.firstName} ${m.user.lastName}`
                            : m.user?.username || m.user?.email || 'User';
                        return (
                          <option key={m.id} value={m.user?.id || m.id}>
                            {name}
                          </option>
                        );
                      })}
                    </select>
                  </div>
                </div>
              </div>

              {canEdit && (
                <div className="pt-2 flex items-center gap-3">
                  <Button
                    type="submit"
                    variant="primary"
                    size="sm"
                    isLoading={isSavingEmail || isSavingInbox}
                    className="gap-1.5 text-xs"
                  >
                    <Check className="h-3.5 w-3.5" />
                    <span>Save Configuration</span>
                  </Button>

                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    disabled={!isConfigured || isTesting}
                    onClick={handleTestConnection}
                    className="gap-1.5 text-xs border-zinc-800 bg-zinc-950/80 text-zinc-300"
                  >
                    <ShieldCheck className="h-3.5 w-3.5" />
                    <span>{isTesting ? 'Testing...' : 'Test IMAP Connection'}</span>
                  </Button>
                </div>
              )}
            </form>
          )}

          {/* SECTION 2: AUTOMATION RULES */}
          {activeSection === 'rules' && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h4 className="text-xs font-semibold text-zinc-100">
                    Inbound Routing Rules
                  </h4>
                  <p className="text-[11px] text-zinc-500">
                    Automate task priority, assignment, and categorization based on sender or keywords.
                  </p>
                </div>

                {canEdit && (
                  <Button
                    type="button"
                    variant="primary"
                    size="sm"
                    onClick={() => setIsRuleModalOpen(true)}
                    className="gap-1.5 text-xs"
                  >
                    <Plus className="h-3.5 w-3.5" />
                    <span>Add Rule</span>
                  </Button>
                )}
              </div>

              {isRulesLoading ? (
                <div className="py-8 text-center text-xs text-zinc-500">
                  Loading automation rules...
                </div>
              ) : rules.length === 0 ? (
                <div className="py-12 text-center text-xs text-zinc-500 border border-dashed border-zinc-800 rounded-md">
                  No inbound rules created yet. Add a rule to triage incoming messages.
                </div>
              ) : (
                <div className="divide-y divide-zinc-800 border border-zinc-850 rounded-md overflow-hidden">
                  {rules.map((rule) => {
                    const conds = (rule.conditions || {}) as Record<string, unknown>;
                    const actions = (rule.actions || {}) as Record<string, unknown>;
                    const fromObj = conds.from as Record<string, unknown> | undefined;
                    const fromCond = fromObj?.contains as string | undefined;
                    const subjObj = conds.subject as Record<string, unknown> | undefined;
                    const subjCond = subjObj?.contains as string | undefined;
                    const actionPriority = actions.setPriority as string | undefined;

                    return (
                      <div
                        key={rule.id}
                        className="flex items-center justify-between p-3.5 bg-zinc-950/40 hover:bg-zinc-900/40 transition-colors"
                      >
                        <div className="space-y-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-semibold text-zinc-200">
                              {rule.name}
                            </span>
                            {rule.enabled ? (
                              <Badge
                                variant="outline"
                                className="text-[9px] bg-emerald-950/60 text-emerald-300 border-emerald-800/60"
                              >
                                Active
                              </Badge>
                            ) : (
                              <Badge
                                variant="outline"
                                className="text-[9px] text-zinc-500 border-zinc-800"
                              >
                                Disabled
                              </Badge>
                            )}
                          </div>

                          <div className="flex items-center gap-2 text-[11px] text-zinc-400 flex-wrap">
                            {fromCond && (
                              <span className="font-mono bg-zinc-900 px-1.5 py-0.5 rounded border border-zinc-800">
                                Sender: {fromCond}
                              </span>
                            )}
                            {subjCond && (
                              <span className="font-mono bg-zinc-900 px-1.5 py-0.5 rounded border border-zinc-800">
                                Subject: {subjCond}
                              </span>
                            )}
                            {actionPriority && (
                              <Badge
                                variant="outline"
                                className="text-[9px] uppercase font-mono text-zinc-300"
                              >
                                Priority: {actionPriority}
                              </Badge>
                            )}
                          </div>
                        </div>

                        {canEdit && (
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            disabled={isDeletingRule}
                            onClick={() => handleDeleteRule(rule.id, rule.name)}
                            className="h-7 w-7 p-0 text-zinc-500 hover:text-rose-400"
                            title="Delete rule"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          )}

          {/* SECTION 3: INCOMING MESSAGE FEED */}
          {activeSection === 'messages' && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h4 className="text-xs font-semibold text-zinc-100">
                    Incoming Email Queue
                  </h4>
                  <p className="text-[11px] text-zinc-500">
                    Review incoming support emails and convert them into project tasks.
                  </p>
                </div>
              </div>

              {isMessagesLoading ? (
                <div className="py-8 text-center text-xs text-zinc-500">
                  Loading messages...
                </div>
              ) : messages.length === 0 ? (
                <div className="py-12 text-center text-xs text-zinc-500 border border-dashed border-zinc-800 rounded-md">
                  No incoming emails received yet. Click "Sync Now" to fetch from IMAP.
                </div>
              ) : (
                <div className="divide-y divide-zinc-800 border border-zinc-850 rounded-md overflow-hidden">
                  {messages.map((msg) => {
                    const isPending = msg.status === 'PENDING';
                    const isConverted = msg.status === 'CONVERTED';

                    return (
                      <div
                        key={msg.id}
                        className="flex flex-col sm:flex-row sm:items-center justify-between p-3.5 bg-zinc-950/40 hover:bg-zinc-900/40 transition-colors gap-3"
                      >
                        <div className="space-y-1 min-w-0">
                          <div className="flex items-center gap-2 flex-wrap">
                            <span className="text-xs font-semibold text-zinc-200 truncate">
                              {msg.subject || '(No Subject)'}
                            </span>
                            <Badge
                              variant="outline"
                              className={`text-[9px] font-mono uppercase ${
                                isConverted
                                  ? 'bg-emerald-950/60 text-emerald-300 border-emerald-800/60'
                                  : isPending
                                  ? 'bg-amber-950/60 text-amber-300 border-amber-800/60'
                                  : 'text-zinc-500 border-zinc-800'
                              }`}
                            >
                              {msg.status}
                            </Badge>
                          </div>

                          <div className="text-[11px] text-zinc-400 flex items-center gap-2 flex-wrap">
                            <span>From: <strong>{msg.fromName || msg.fromEmail}</strong></span>
                            <span className="text-zinc-600">•</span>
                            <span className="text-zinc-500">
                              {new Date(msg.createdAt).toLocaleDateString()}
                            </span>
                          </div>

                          {msg.snippet && (
                            <p className="text-[11px] text-zinc-500 line-clamp-1">
                              {msg.snippet}
                            </p>
                          )}
                        </div>

                        <div className="flex items-center gap-2 self-end sm:self-auto shrink-0">
                          {isPending && canEdit && (
                            <>
                              <Button
                                type="button"
                                variant="primary"
                                size="sm"
                                disabled={isConverting}
                                onClick={() => handleConvertMessage(msg.id)}
                                className="gap-1 text-xs h-7 px-2.5"
                              >
                                <ArrowUpRight className="h-3 w-3" />
                                <span>Convert to Task</span>
                              </Button>

                              <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                onClick={() => handleIgnoreMessage(msg.id)}
                                className="text-xs h-7 px-2 border-zinc-800 text-zinc-400 hover:text-zinc-200"
                              >
                                Ignore
                              </Button>
                            </>
                          )}

                          {isConverted && msg.taskId && (
                            <span className="text-[11px] text-emerald-400 flex items-center gap-1 font-mono">
                              <CheckCircle2 className="h-3 w-3" />
                              Task Created
                            </span>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Inbound Rule Modal */}
      <Dialog open={isRuleModalOpen} onOpenChange={setIsRuleModalOpen}>
        <DialogContent className="bg-zinc-900 border-zinc-800 max-w-md">
          <DialogHeader>
            <DialogTitle className="text-base text-zinc-100">
              Create Inbound Automation Rule
            </DialogTitle>
            <DialogDescription className="text-xs text-zinc-400">
              Define matching conditions and actions for emails received in this inbox.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleCreateRule} className="space-y-4 pt-2">
            <div className="space-y-1.5">
              <Label htmlFor="rule-name" className="text-xs text-zinc-300">
                Rule Name
              </Label>
              <Input
                id="rule-name"
                value={ruleName}
                onChange={(e) => setRuleName(e.target.value)}
                placeholder="e.g. VIP Customer Urgent Routing"
                className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                autoFocus
              />
            </div>

            <div className="space-y-3 pt-2 border-t border-zinc-800/80">
              <span className="text-[11px] uppercase font-mono font-semibold text-zinc-400">
                Conditions (When...)
              </span>

              <div className="space-y-1.5">
                <Label htmlFor="sender-cond" className="text-xs text-zinc-300">
                  Sender Email Contains
                </Label>
                <Input
                  id="sender-cond"
                  value={senderContains}
                  onChange={(e) => setSenderContains(e.target.value)}
                  placeholder="e.g. @acme-client.com"
                  className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                />
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="subj-cond" className="text-xs text-zinc-300">
                  Subject Keyword Contains
                </Label>
                <Input
                  id="subj-cond"
                  value={subjectContains}
                  onChange={(e) => setSubjectContains(e.target.value)}
                  placeholder="e.g. [Bug], Critical, Outage"
                  className="bg-zinc-950 border-zinc-800 text-xs text-zinc-100"
                />
              </div>
            </div>

            <div className="space-y-3 pt-2 border-t border-zinc-800/80">
              <span className="text-[11px] uppercase font-mono font-semibold text-zinc-400">
                Actions (Then...)
              </span>

              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label className="text-xs text-zinc-300">Set Priority</Label>
                  <select
                    value={rulePriority}
                    onChange={(e) => setRulePriority(e.target.value as TaskPriority)}
                    className="w-full h-8 rounded bg-zinc-950 border border-zinc-800 px-2.5 text-xs text-zinc-200"
                  >
                    <option value="LOW">Low</option>
                    <option value="MEDIUM">Medium</option>
                    <option value="HIGH">High</option>
                    <option value="URGENT">Urgent</option>
                  </select>
                </div>

                <div className="space-y-1.5">
                  <Label className="text-xs text-zinc-300">Assign Collaborator</Label>
                  <select
                    value={ruleAssigneeId}
                    onChange={(e) => setRuleAssigneeId(e.target.value)}
                    className="w-full h-8 rounded bg-zinc-950 border border-zinc-800 px-2.5 text-xs text-zinc-200"
                  >
                    <option value="">(None)</option>
                    {workspaceMembers.map((m) => {
                      const name =
                        m.user?.firstName && m.user?.lastName
                          ? `${m.user.firstName} ${m.user.lastName}`
                          : m.user?.username || m.user?.email || 'User';
                      return (
                        <option key={m.id} value={m.user?.id || m.id}>
                          {name}
                        </option>
                      );
                    })}
                  </select>
                </div>
              </div>

              <div className="flex items-center gap-2 pt-1">
                <input
                  id="stopMatch"
                  type="checkbox"
                  checked={stopOnMatch}
                  onChange={(e) => setStopOnMatch(e.target.checked)}
                  className="h-3.5 w-3.5 rounded border-zinc-700 bg-zinc-950 text-indigo-600"
                />
                <Label htmlFor="stopMatch" className="text-xs text-zinc-400 cursor-pointer">
                  Stop evaluating subsequent rules on match
                </Label>
              </div>
            </div>

            <DialogFooter className="pt-3 gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setIsRuleModalOpen(false)}
                className="text-xs"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="primary"
                size="sm"
                isLoading={isCreatingRule}
                className="text-xs"
              >
                Create Rule
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default ProjectInboxTab;

