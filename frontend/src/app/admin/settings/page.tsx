'use client';

import * as React from 'react';
import { useState, useEffect } from 'react';
import {
  Sliders,
  Mail,
  KeyRound,
  Globe,
  Sparkles,
  HardDrive,
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
import { useTestAiConnectionMutation } from '@/lib/store/api/aiChatApi';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';

import { GeneralSettingsTab } from './components/GeneralSettingsTab';
import { SecuritySsoTab } from './components/SecuritySsoTab';
import { AiCopilotTab } from './components/AiCopilotTab';
import { MailSmtpTab } from './components/MailSmtpTab';
import { StorageSettingsTab } from './components/StorageSettingsTab';

export default function AdminSettingsPage() {
  const { data: configData, isLoading: isConfigLoading, refetch: refetchConfig } = useGetAdminConfigQuery();
  const { data: regStatus, refetch: refetchRegStatus } = useGetRegistrationStatusQuery();

  const [saveConfig, { isLoading: isSaving }] = useSaveAdminConfigMutation();
  const [testSmtp, { isLoading: isTestingSmtp }] = useTestAdminSmtpMutation();
  const [testAiConnection, { isLoading: isTestingAi }] = useTestAiConnectionMutation();

  // Tab State
  const [activeTab, setActiveTab] = useState('general');

  // 1. General & Platform State
  const [platformName, setPlatformName] = useState('Orbito Enterprise');
  const [supportEmail, setSupportEmail] = useState('support@orbito.io');
  const [defaultTimezone, setDefaultTimezone] = useState('UTC');
  const [isMaintenanceMode, setIsMaintenanceMode] = useState(false);

  // 2. Security & SSO State
  const [isSsoEnabled, setIsSsoEnabled] = useState(false);
  const [ssoProvider, setSsoProvider] = useState('Google');
  const [ssoIssuerUrl, setSsoIssuerUrl] = useState('https://accounts.google.com');
  const [ssoClientId, setSsoClientId] = useState('');
  const [ssoClientSecret, setSsoClientSecret] = useState('');
  const [showSsoSecret, setShowSsoSecret] = useState(false);
  const [ssoDefaultRole, setSsoDefaultRole] = useState('MEMBER');
  const [sessionTimeout, setSessionTimeout] = useState('60');
  const [copiedCallback, setCopiedCallback] = useState(false);

  // 3. AI Copilot State
  const [aiApiUrl, setAiApiUrl] = useState('https://openrouter.ai/api/v1');
  const [aiApiKey, setAiApiKey] = useState('');
  const [showAiKey, setShowAiKey] = useState(false);
  const [aiModel, setAiModel] = useState('anthropic/claude-3.5-sonnet');
  const [aiMaxTokens, setAiMaxTokens] = useState('4096');
  const [aiTemperature, setAiTemperature] = useState('0.2');
  const [aiTone, setAiTone] = useState('Professional');
  const [aiGreetingMessage, setAiGreetingMessage] = useState('How can I help you today?');
  const [aiSystemPrompt, setAiSystemPrompt] = useState('');
  const [featTaskBreakdown, setFeatTaskBreakdown] = useState(true);
  const [featDescriptionGen, setFeatDescriptionGen] = useState(true);
  const [featCodeReview, setFeatCodeReview] = useState(true);
  const [featSprintSummary, setFeatSprintSummary] = useState(true);

  // 4. Mail / SMTP State
  const [smtpHost, setSmtpHost] = useState('');
  const [smtpPort, setSmtpPort] = useState('587');
  const [smtpUser, setSmtpUser] = useState('');
  const [smtpPass, setSmtpPass] = useState('');
  const [showSmtpPass, setShowSmtpPass] = useState(false);
  const [smtpFrom, setSmtpFrom] = useState('noreply@orbito.io');
  const [smtpFromName, setSmtpFromName] = useState('Orbito Platform');
  const [testRecipientEmail, setTestRecipientEmail] = useState('');
  const [inboundEmailEnabled, setInboundEmailEnabled] = useState(false);
  const [inboundEmailDomain, setInboundEmailDomain] = useState('tasks.orbito.io');
  const [emailNotifyInvitations, setEmailNotifyInvitations] = useState(true);
  const [emailNotifyAssignments, setEmailNotifyAssignments] = useState(true);
  const [emailNotifyMentions, setEmailNotifyMentions] = useState(true);
  const [emailNotifyOverdue, setEmailNotifyOverdue] = useState(true);

  // 5. Storage State
  const [storageProvider, setStorageProvider] = useState<'local' | 's3'>('local');
  const [s3Bucket, setS3Bucket] = useState('');
  const [s3Region, setS3Region] = useState('us-east-1');
  const [s3Endpoint, setS3Endpoint] = useState('');
  const [s3AccessKey, setS3AccessKey] = useState('');
  const [s3SecretKey, setS3SecretKey] = useState('');
  const [showS3Secret, setShowS3Secret] = useState(false);
  const [maxUploadMb, setMaxUploadMb] = useState('50');

  const [hasInitialized, setHasInitialized] = useState(false);

  // Initialize form state from database settings
  useEffect(() => {
    if (configData?.settings && !hasInitialized) {
      const getVal = (k: string, def = '') =>
        configData.settings.find((s) => s.key === k)?.value || def;

      setPlatformName(getVal('platform_name', 'Orbito Enterprise'));
      setSupportEmail(getVal('support_email', 'support@orbito.io'));
      setDefaultTimezone(getVal('default_timezone', 'UTC'));
      setIsMaintenanceMode(getVal('maintenance_mode') === 'true');

      setIsSsoEnabled(getVal('sso_enabled') === 'true');
      setSsoProvider(getVal('sso_provider_name', 'Google'));
      setSsoIssuerUrl(getVal('sso_issuer_url', 'https://accounts.google.com'));
      setSsoClientId(getVal('sso_client_id', ''));
      setSsoClientSecret(getVal('sso_client_secret', ''));
      setSsoDefaultRole(getVal('sso_default_role', 'MEMBER'));
      setSessionTimeout(getVal('session_ttl_minutes', '60'));

      setAiApiUrl(getVal('ai_api_url', 'https://openrouter.ai/api/v1'));
      setAiApiKey(getVal('ai_api_key', ''));
      setAiModel(getVal('ai_model', 'anthropic/claude-3.5-sonnet'));
      setAiMaxTokens(getVal('ai_max_tokens', '4096'));
      setAiTemperature(getVal('ai_temperature', '0.2'));
      setAiTone(getVal('ai_tone', 'Professional'));
      setAiGreetingMessage(getVal('ai_greeting_message', 'How can I help you today?'));
      setAiSystemPrompt(getVal('ai_system_prompt', ''));
      setFeatTaskBreakdown(getVal('ai_feat_task_breakdown', 'true') === 'true');
      setFeatDescriptionGen(getVal('ai_feat_description_gen', 'true') === 'true');
      setFeatCodeReview(getVal('ai_feat_code_review', 'true') === 'true');
      setFeatSprintSummary(getVal('ai_feat_sprint_summary', 'true') === 'true');

      setSmtpHost(getVal('smtp_host', ''));
      setSmtpPort(getVal('smtp_port', '587'));
      setSmtpUser(getVal('smtp_user', ''));
      setSmtpPass(getVal('smtp_pass', ''));
      setSmtpFrom(getVal('smtp_from', 'noreply@orbito.io'));
      setSmtpFromName(getVal('smtp_from_name', 'Orbito Platform'));
      setInboundEmailEnabled(getVal('inbound_email_enabled') === 'true');
      setInboundEmailDomain(getVal('inbound_email_domain', 'tasks.orbito.io'));
      setEmailNotifyInvitations(getVal('email_notify_invitations', 'true') === 'true');
      setEmailNotifyAssignments(getVal('email_notify_assignments', 'true') === 'true');
      setEmailNotifyMentions(getVal('email_notify_mentions', 'true') === 'true');
      setEmailNotifyOverdue(getVal('email_notify_overdue', 'true') === 'true');

      setStorageProvider(getVal('storage_provider', 'local') === 's3' ? 's3' : 'local');
      setS3Bucket(getVal('s3_bucket', ''));
      setS3Region(getVal('s3_region', 'us-east-1'));
      setS3Endpoint(getVal('s3_endpoint', ''));
      setS3AccessKey(getVal('s3_access_key', ''));
      setS3SecretKey(getVal('s3_secret_key', ''));
      setMaxUploadMb(getVal('max_upload_size_mb', '50'));

      setHasInitialized(true);
    }
  }, [configData, hasInitialized]);

  // Read registration and AI status from active config
  const regSetting = configData?.settings.find((s) => s.key === 'enablePublicRegistration');
  const isPublicRegEnabled = regSetting ? regSetting.value === 'true' : (regStatus?.enabled ?? false);

  const aiEnabledSetting = configData?.settings.find((s) => s.key === 'ai_enabled');
  const isAiEnabled = aiEnabledSetting ? aiEnabledSetting.value === 'true' : false;

  // --- Handlers ---

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
          : 'Public user registration is now RESTRICTED'
      );
    } catch {
      toast.error('Failed to update registration setting');
    }
  };

  const handleToggleAi = async () => {
    const nextVal = (!isAiEnabled).toString();
    try {
      await saveConfig({
        settings: [
          {
            key: 'ai_enabled',
            value: nextVal,
            category: 'ai',
            description: 'Enable or disable platform-wide AI Copilot',
          },
        ],
      }).unwrap();
      toast.success(nextVal === 'true' ? 'AI Copilot is now ENABLED' : 'AI Copilot is now DISABLED');
    } catch {
      toast.error('Failed to update AI status');
    }
  };

  const handleToggleMaintenance = async () => {
    const nextVal = (!isMaintenanceMode).toString();
    try {
      await saveConfig({
        settings: [
          {
            key: 'maintenance_mode',
            value: nextVal,
            category: 'general',
            description: 'Platform maintenance mode gate',
          },
        ],
      }).unwrap();
      setIsMaintenanceMode(!isMaintenanceMode);
      toast.success(
        nextVal === 'true'
          ? 'Maintenance mode ACTIVATED. Standard logins paused.'
          : 'Maintenance mode DEACTIVATED. System operational.'
      );
    } catch {
      toast.error('Failed to update maintenance mode');
    }
  };

  const handleToggleSso = async () => {
    const nextVal = (!isSsoEnabled).toString();
    try {
      await saveConfig({
        settings: [
          {
            key: 'sso_enabled',
            value: nextVal,
            category: 'security',
            description: 'Enable or disable OpenID Connect Single Sign-On',
          },
        ],
      }).unwrap();
      setIsSsoEnabled(!isSsoEnabled);
      toast.success(nextVal === 'true' ? 'Enterprise SSO is now ENABLED' : 'Enterprise SSO is now DISABLED');
    } catch {
      toast.error('Failed to update SSO status');
    }
  };

  const handleSaveGeneral = async () => {
    try {
      await saveConfig({
        settings: [
          { key: 'platform_name', value: platformName.trim(), category: 'general', description: 'Platform instance name' },
          { key: 'support_email', value: supportEmail.trim(), category: 'general', description: 'System support email' },
          { key: 'default_timezone', value: defaultTimezone, category: 'general', description: 'System reference timezone' },
        ],
      }).unwrap();
      toast.success('General platform preferences saved');
    } catch {
      toast.error('Failed to save general preferences');
    }
  };

  const handleSaveSsoConfig = async () => {
    try {
      await saveConfig({
        settings: [
          { key: 'sso_provider_name', value: ssoProvider, category: 'security', description: 'IdP Provider Name' },
          { key: 'sso_issuer_url', value: ssoIssuerUrl.trim(), category: 'security', description: 'OIDC Issuer Discovery URL' },
          { key: 'sso_client_id', value: ssoClientId.trim(), category: 'security', description: 'OIDC Client Identifier' },
          { key: 'sso_client_secret', value: ssoClientSecret.trim(), category: 'security', description: 'OIDC Client Secret', isEncrypted: true },
          { key: 'sso_default_role', value: ssoDefaultRole, category: 'security', description: 'Default role on SSO onboarding' },
          { key: 'session_ttl_minutes', value: sessionTimeout, category: 'security', description: 'Inactivity session timeout' },
        ],
      }).unwrap();
      toast.success('Enterprise SSO & Security parameters saved');
    } catch {
      toast.error('Failed to save SSO configuration');
    }
  };

  const handleSaveAiConfig = async () => {
    try {
      await saveConfig({
        settings: [
          { key: 'ai_api_url', value: aiApiUrl.trim(), category: 'ai', description: 'AI Provider Base URL' },
          { key: 'ai_api_key', value: aiApiKey.trim(), category: 'ai', description: 'AI Provider API Key', isEncrypted: true },
          { key: 'ai_model', value: aiModel.trim(), category: 'ai', description: 'AI Model Identifier' },
          { key: 'ai_max_tokens', value: aiMaxTokens.trim(), category: 'ai', description: 'Maximum AI token generation' },
          { key: 'ai_temperature', value: aiTemperature.trim(), category: 'ai', description: 'LLM reasoning temperature' },
          { key: 'ai_tone', value: aiTone, category: 'ai', description: 'AI assistant personality tone' },
          { key: 'ai_greeting_message', value: aiGreetingMessage.trim(), category: 'ai', description: 'Copilot welcome message' },
          { key: 'ai_system_prompt', value: aiSystemPrompt.trim(), category: 'ai', description: 'Custom organization AI instructions' },
          { key: 'ai_feat_task_breakdown', value: featTaskBreakdown.toString(), category: 'ai', description: 'Enable task breakdown skill' },
          { key: 'ai_feat_description_gen', value: featDescriptionGen.toString(), category: 'ai', description: 'Enable description generator skill' },
          { key: 'ai_feat_code_review', value: featCodeReview.toString(), category: 'ai', description: 'Enable code review skill' },
          { key: 'ai_feat_sprint_summary', value: featSprintSummary.toString(), category: 'ai', description: 'Enable sprint summary skill' },
        ],
      }).unwrap();
      toast.success('AI Copilot & Multi-Provider LLM configuration saved successfully');
    } catch {
      toast.error('Failed to save AI settings');
    }
  };

  const handleTestAiConnection = async () => {
    if (!aiApiUrl.trim() || !aiModel.trim()) {
      toast.error('Please specify both API URL and Model name before testing');
      return;
    }
    try {
      const res = await testAiConnection({
        apiUrl: aiApiUrl.trim(),
        apiKey: aiApiKey.trim(),
        model: aiModel.trim(),
      }).unwrap();

      if (res.success) {
        toast.success(res.message || 'AI Provider connection verified successfully!');
      } else {
        toast.error(res.error || res.message || 'AI connection test failed');
      }
    } catch (err: unknown) {
      const msg =
        (err as { data?: { message?: string } })?.data?.message ||
        'Failed to test AI connection. Check endpoint URL and network access.';
      toast.error(msg);
    }
  };

  const handleSaveSmtpConfig = async () => {
    try {
      await saveConfig({
        settings: [
          { key: 'smtp_host', value: smtpHost.trim(), category: 'email', description: 'SMTP server hostname' },
          { key: 'smtp_port', value: smtpPort.trim(), category: 'email', description: 'SMTP port' },
          { key: 'smtp_user', value: smtpUser.trim(), category: 'email', description: 'SMTP username' },
          { key: 'smtp_pass', value: smtpPass.trim(), category: 'email', description: 'SMTP password', isEncrypted: true },
          { key: 'smtp_from', value: smtpFrom.trim(), category: 'email', description: 'Default outbound sender email' },
          { key: 'smtp_from_name', value: smtpFromName.trim(), category: 'email', description: 'Display name for outbound emails' },
          { key: 'inbound_email_enabled', value: inboundEmailEnabled.toString(), category: 'email', description: 'Inbound email task creation' },
          { key: 'inbound_email_domain', value: inboundEmailDomain.trim(), category: 'email', description: 'Inbound email domain' },
          { key: 'email_notify_invitations', value: emailNotifyInvitations.toString(), category: 'email', description: 'Invitation emails' },
          { key: 'email_notify_assignments', value: emailNotifyAssignments.toString(), category: 'email', description: 'Assignment emails' },
          { key: 'email_notify_mentions', value: emailNotifyMentions.toString(), category: 'email', description: 'Mention emails' },
          { key: 'email_notify_overdue', value: emailNotifyOverdue.toString(), category: 'email', description: 'Overdue task reminder emails' },
        ],
      }).unwrap();
      toast.success('Mail delivery topology saved successfully');
    } catch {
      toast.error('Failed to save mail configuration');
    }
  };

  const handleTestSmtp = async () => {
    try {
      const res = await testSmtp().unwrap();
      toast.success(res.message || 'SMTP connection and authentication verified successfully!');
    } catch (err: unknown) {
      const msg =
        err && typeof err === 'object' && 'data' in err
          ? (err as { data: { message?: string } }).data?.message
          : 'SMTP connection failed';
      toast.error(msg || 'SMTP test failed. Verify host, port, and credentials.');
    }
  };

  const handleSendTestEmail = async () => {
    if (!testRecipientEmail || !testRecipientEmail.trim()) {
      toast.error('Please enter a recipient email address to send a live test message');
      return;
    }
    try {
      const res = await testSmtp({ email: testRecipientEmail.trim() }).unwrap();
      toast.success(res.message || `Test email successfully dispatched to ${testRecipientEmail.trim()}`);
    } catch (err: unknown) {
      const msg =
        err && typeof err === 'object' && 'data' in err
          ? (err as { data: { message?: string } }).data?.message
          : 'Failed to send test email';
      toast.error(msg || 'Email delivery failed. Verify outbound SMTP credentials.');
    }
  };

  const handleSaveStorageConfig = async () => {
    try {
      await saveConfig({
        settings: [
          { key: 'storage_provider', value: storageProvider, category: 'storage', description: 'Primary asset storage backend' },
          { key: 's3_bucket', value: s3Bucket.trim(), category: 'storage', description: 'S3 bucket name' },
          { key: 's3_region', value: s3Region.trim(), category: 'storage', description: 'S3 cloud region' },
          { key: 's3_endpoint', value: s3Endpoint.trim(), category: 'storage', description: 'S3 custom endpoint URL' },
          { key: 's3_access_key', value: s3AccessKey.trim(), category: 'storage', description: 'S3 access key ID' },
          { key: 's3_secret_key', value: s3SecretKey.trim(), category: 'storage', description: 'S3 secret key', isEncrypted: true },
          { key: 'max_upload_size_mb', value: maxUploadMb.trim(), category: 'storage', description: 'Max upload size in MB' },
        ],
      }).unwrap();
      toast.success('Storage and asset delivery settings saved');
    } catch {
      toast.error('Failed to save storage settings');
    }
  };

  const copyCallbackUrl = () => {
    const origin = typeof window !== 'undefined' ? window.location.origin : 'https://orbito.yourdomain.com';
    navigator.clipboard.writeText(`${origin}/api/auth/oidc/callback`);
    setCopiedCallback(true);
    toast.success('OIDC Callback URL copied to clipboard');
    setTimeout(() => setCopiedCallback(false), 2000);
  };

  return (
    <RoleGuard requiredRole={Role.SUPER_ADMIN}>
      <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b] text-zinc-100">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-5 border-b border-zinc-800/80">
          <div className="space-y-1">
            <div className="flex items-center gap-2.5">
              <div className="h-7 w-7 rounded-md bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-200">
                <Sliders className="h-4 w-4" />
              </div>
              <h1 className="text-xl font-bold tracking-tight text-zinc-100">
                System & Security Settings
              </h1>
              <Badge variant="outline" className="text-[10px] uppercase font-mono tracking-wider border-zinc-700 bg-zinc-850/80 text-zinc-300">
                Root Config
              </Badge>
            </div>
            <p className="text-xs text-zinc-400">
              Configure instance preferences, enterprise federated SSO, AI providers, transactional mail, and cloud storage.
            </p>
          </div>
        </div>

        {/* Unified Tabs Layout */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6 max-w-4xl">
          {/* Flat Horizontal Tabs Bar */}
          <TabsList className="bg-zinc-900/60 p-1 rounded-lg border border-zinc-800/80 flex flex-wrap h-auto gap-1">
            <TabsTrigger
              value="general"
              className="data-[state=active]:bg-zinc-800 data-[state=active]:text-zinc-100 text-zinc-400 rounded-md text-xs font-medium px-3.5 py-1.5 transition-all flex items-center gap-2 cursor-pointer"
            >
              <Globe className="h-3.5 w-3.5" />
              <span>General & Access</span>
            </TabsTrigger>

            <TabsTrigger
              value="sso"
              className="data-[state=active]:bg-zinc-800 data-[state=active]:text-zinc-100 text-zinc-400 rounded-md text-xs font-medium px-3.5 py-1.5 transition-all flex items-center gap-2 cursor-pointer"
            >
              <KeyRound className="h-3.5 w-3.5" />
              <span>Security & SSO</span>
            </TabsTrigger>

            <TabsTrigger
              value="ai"
              className="data-[state=active]:bg-zinc-800 data-[state=active]:text-zinc-100 text-zinc-400 rounded-md text-xs font-medium px-3.5 py-1.5 transition-all flex items-center gap-2 cursor-pointer"
            >
              <Sparkles className="h-3.5 w-3.5" />
              <span>AI Copilot & LLM</span>
            </TabsTrigger>

            <TabsTrigger
              value="email"
              className="data-[state=active]:bg-zinc-800 data-[state=active]:text-zinc-100 text-zinc-400 rounded-md text-xs font-medium px-3.5 py-1.5 transition-all flex items-center gap-2 cursor-pointer"
            >
              <Mail className="h-3.5 w-3.5" />
              <span>Mail & SMTP</span>
            </TabsTrigger>

            <TabsTrigger
              value="storage"
              className="data-[state=active]:bg-zinc-800 data-[state=active]:text-zinc-100 text-zinc-400 rounded-md text-xs font-medium px-3.5 py-1.5 transition-all flex items-center gap-2 cursor-pointer"
            >
              <HardDrive className="h-3.5 w-3.5" />
              <span>Storage & Assets</span>
            </TabsTrigger>
          </TabsList>

          {/* TAB 1: GENERAL & ACCESS */}
          <TabsContent value="general" className="outline-none">
            <GeneralSettingsTab
              platformName={platformName}
              setPlatformName={setPlatformName}
              supportEmail={supportEmail}
              setSupportEmail={setSupportEmail}
              defaultTimezone={defaultTimezone}
              setDefaultTimezone={setDefaultTimezone}
              isMaintenanceMode={isMaintenanceMode}
              onToggleMaintenance={handleToggleMaintenance}
              isPublicRegEnabled={isPublicRegEnabled}
              onToggleRegistration={handleToggleRegistration}
              onSave={handleSaveGeneral}
              isSaving={isSaving}
            />
          </TabsContent>

          {/* TAB 2: SECURITY & SSO */}
          <TabsContent value="sso" className="outline-none">
            <SecuritySsoTab
              isSsoEnabled={isSsoEnabled}
              onToggleSso={handleToggleSso}
              ssoProvider={ssoProvider}
              setSsoProvider={setSsoProvider}
              ssoIssuerUrl={ssoIssuerUrl}
              setSsoIssuerUrl={setSsoIssuerUrl}
              ssoClientId={ssoClientId}
              setSsoClientId={setSsoClientId}
              ssoClientSecret={ssoClientSecret}
              setSsoClientSecret={setSsoClientSecret}
              showSsoSecret={showSsoSecret}
              setShowSsoSecret={setShowSsoSecret}
              ssoDefaultRole={ssoDefaultRole}
              setSsoDefaultRole={setSsoDefaultRole}
              sessionTimeout={sessionTimeout}
              setSessionTimeout={setSessionTimeout}
              copiedCallback={copiedCallback}
              onCopyCallback={copyCallbackUrl}
              onSave={handleSaveSsoConfig}
              isSaving={isSaving}
            />
          </TabsContent>

          {/* TAB 3: AI COPILOT & LLM */}
          <TabsContent value="ai" className="outline-none">
            <AiCopilotTab
              isAiEnabled={isAiEnabled}
              onToggleAi={handleToggleAi}
              aiApiUrl={aiApiUrl}
              setAiApiUrl={setAiApiUrl}
              aiApiKey={aiApiKey}
              setAiApiKey={setAiApiKey}
              showAiKey={showAiKey}
              setShowAiKey={setShowAiKey}
              aiModel={aiModel}
              setAiModel={setAiModel}
              aiMaxTokens={aiMaxTokens}
              setAiMaxTokens={setAiMaxTokens}
              aiTemperature={aiTemperature}
              setAiTemperature={setAiTemperature}
              aiTone={aiTone}
              setAiTone={setAiTone}
              aiGreetingMessage={aiGreetingMessage}
              setAiGreetingMessage={setAiGreetingMessage}
              aiSystemPrompt={aiSystemPrompt}
              setAiSystemPrompt={setAiSystemPrompt}
              featTaskBreakdown={featTaskBreakdown}
              setFeatTaskBreakdown={setFeatTaskBreakdown}
              featDescriptionGen={featDescriptionGen}
              setFeatDescriptionGen={setFeatDescriptionGen}
              featCodeReview={featCodeReview}
              setFeatCodeReview={setFeatCodeReview}
              featSprintSummary={featSprintSummary}
              setFeatSprintSummary={setFeatSprintSummary}
              onTestConnection={handleTestAiConnection}
              isTesting={isTestingAi}
              onSave={handleSaveAiConfig}
              isSaving={isSaving}
            />
          </TabsContent>

          {/* TAB 4: MAIL & SMTP */}
          <TabsContent value="email" className="outline-none">
            <MailSmtpTab
              smtpHost={smtpHost}
              setSmtpHost={setSmtpHost}
              smtpPort={smtpPort}
              setSmtpPort={setSmtpPort}
              smtpUser={smtpUser}
              setSmtpUser={setSmtpUser}
              smtpPass={smtpPass}
              setSmtpPass={setSmtpPass}
              showSmtpPass={showSmtpPass}
              setShowSmtpPass={setShowSmtpPass}
              smtpFrom={smtpFrom}
              setSmtpFrom={setSmtpFrom}
              smtpFromName={smtpFromName}
              setSmtpFromName={setSmtpFromName}
              hostSource={configData?.smtpSources?.hostSource}
              testRecipientEmail={testRecipientEmail}
              setTestRecipientEmail={setTestRecipientEmail}
              inboundEmailEnabled={inboundEmailEnabled}
              setInboundEmailEnabled={setInboundEmailEnabled}
              inboundEmailDomain={inboundEmailDomain}
              setInboundEmailDomain={setInboundEmailDomain}
              emailNotifyInvitations={emailNotifyInvitations}
              setEmailNotifyInvitations={setEmailNotifyInvitations}
              emailNotifyAssignments={emailNotifyAssignments}
              setEmailNotifyAssignments={setEmailNotifyAssignments}
              emailNotifyMentions={emailNotifyMentions}
              setEmailNotifyMentions={setEmailNotifyMentions}
              emailNotifyOverdue={emailNotifyOverdue}
              setEmailNotifyOverdue={setEmailNotifyOverdue}
              onTestSmtp={handleTestSmtp}
              onSendTestEmail={handleSendTestEmail}
              isTestingSmtp={isTestingSmtp}
              onSave={handleSaveSmtpConfig}
              isSaving={isSaving}
            />
          </TabsContent>

          {/* TAB 5: STORAGE & CLOUD MEDIA */}
          <TabsContent value="storage" className="outline-none">
            <StorageSettingsTab
              storageProvider={storageProvider}
              setStorageProvider={setStorageProvider}
              s3Bucket={s3Bucket}
              setS3Bucket={setS3Bucket}
              s3Region={s3Region}
              setS3Region={setS3Region}
              s3Endpoint={s3Endpoint}
              setS3Endpoint={setS3Endpoint}
              s3AccessKey={s3AccessKey}
              setS3AccessKey={setS3AccessKey}
              s3SecretKey={s3SecretKey}
              setS3SecretKey={setS3SecretKey}
              showS3Secret={showS3Secret}
              setShowS3Secret={setShowS3Secret}
              maxUploadMb={maxUploadMb}
              setMaxUploadMb={setMaxUploadMb}
              onSave={handleSaveStorageConfig}
              isSaving={isSaving}
            />
          </TabsContent>
        </Tabs>
      </div>
    </RoleGuard>
  );
}
