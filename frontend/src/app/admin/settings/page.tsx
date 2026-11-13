'use client';

import * as React from 'react';
import { useState, useEffect } from 'react';
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
  Sparkles,
  Bot,
  Loader2,
  Cpu,
  HardDrive,
  Eye,
  EyeOff,
  Copy,
  Check,
  Server,
  Save,
  Zap,
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
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';

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
  const [enforceMfa, setEnforceMfa] = useState(false);
  const [copiedCallback, setCopiedCallback] = useState(false);

  // 3. AI Copilot State
  const [aiApiUrl, setAiApiUrl] = useState('https://openrouter.ai/api/v1');
  const [aiApiKey, setAiApiKey] = useState('');
  const [showAiKey, setShowAiKey] = useState(false);
  const [aiModel, setAiModel] = useState('anthropic/claude-3.5-sonnet');
  const [aiMaxTokens, setAiMaxTokens] = useState('4096');
  const [aiTemperature, setAiTemperature] = useState('0.2');

  // 4. Mail / SMTP State
  const [smtpHost, setSmtpHost] = useState('');
  const [smtpPort, setSmtpPort] = useState('587');
  const [smtpUser, setSmtpUser] = useState('');
  const [smtpPass, setSmtpPass] = useState('');
  const [showSmtpPass, setShowSmtpPass] = useState(false);
  const [smtpFrom, setSmtpFrom] = useState('noreply@orbito.io');
  const [smtpFromName, setSmtpFromName] = useState('Orbito Platform');

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
      setEnforceMfa(getVal('enforce_mfa') === 'true');

      setAiApiUrl(getVal('ai_api_url', 'https://openrouter.ai/api/v1'));
      setAiApiKey(getVal('ai_api_key', ''));
      setAiModel(getVal('ai_model', 'anthropic/claude-3.5-sonnet'));
      setAiMaxTokens(getVal('ai_max_tokens', '4096'));
      setAiTemperature(getVal('ai_temperature', '0.2'));

      setSmtpHost(getVal('smtp_host', ''));
      setSmtpPort(getVal('smtp_port', '587'));
      setSmtpUser(getVal('smtp_user', ''));
      setSmtpPass(getVal('smtp_pass', ''));
      setSmtpFrom(getVal('smtp_from', 'noreply@orbito.io'));
      setSmtpFromName(getVal('smtp_from_name', 'Orbito Platform'));

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
          { key: 'enforce_mfa', value: enforceMfa.toString(), category: 'security', description: 'Enforce MFA policy' },
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
        ],
      }).unwrap();
      toast.success('AI Copilot configuration saved successfully');
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

          {/* ========================================================================= */}
          {/* TAB 1: GENERAL & ACCESS                                                  */}
          {/* ========================================================================= */}
          <TabsContent value="general" className="space-y-6 outline-none">
            <div className="border border-zinc-800/80 rounded-lg bg-zinc-900/30 divide-y divide-zinc-800/80">
              {/* Row 1: Public Registration Gate */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="space-y-0.5 max-w-lg">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-zinc-200">Public User Registration</span>
                    <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono ${isPublicRegEnabled ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800/60' : 'bg-zinc-800 text-zinc-400 border border-zinc-700'}`}>
                      {isPublicRegEnabled ? 'Open (/register)' : 'Restricted (Invite Only)'}
                    </span>
                  </div>
                  <p className="text-xs text-zinc-400 leading-relaxed">
                    Allow external visitors to self-register their own organizations. When restricted, onboarding requires an explicit workspace or tenant invite token.
                  </p>
                </div>
                <div className="flex items-center gap-3 shrink-0">
                  <Switch
                    checked={isPublicRegEnabled}
                    onCheckedChange={handleToggleRegistration}
                    disabled={isSaving}
                    aria-label="Toggle Public Registration"
                  />
                </div>
              </div>

              {/* Row 2: Maintenance Mode */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="space-y-0.5 max-w-lg">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-zinc-200">Maintenance Mode Gate</span>
                    <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono ${isMaintenanceMode ? 'bg-amber-950/80 text-amber-400 border border-amber-800/60' : 'bg-zinc-800 text-zinc-400 border border-zinc-700'}`}>
                      {isMaintenanceMode ? 'Maintenance Active' : 'Operational'}
                    </span>
                  </div>
                  <p className="text-xs text-zinc-400 leading-relaxed">
                    Temporarily pause public and standard member requests during infrastructure maintenance. Platform Super Admins retain full access.
                  </p>
                </div>
                <div className="flex items-center gap-3 shrink-0">
                  <Switch
                    checked={isMaintenanceMode}
                    onCheckedChange={handleToggleMaintenance}
                    disabled={isSaving}
                    aria-label="Toggle Maintenance Mode"
                  />
                </div>
              </div>

              {/* Row 3: Platform Instance Name */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="space-y-0.5 max-w-lg">
                  <span className="text-sm font-medium text-zinc-200">Platform Instance Name</span>
                  <p className="text-xs text-zinc-400 leading-relaxed">
                    The name displayed on browser tabs, notification emails, and top navigation shells.
                  </p>
                </div>
                <div className="w-full sm:w-72 shrink-0">
                  <Input
                    value={platformName}
                    onChange={(e) => setPlatformName(e.target.value)}
                    placeholder="Orbito Enterprise"
                    className="h-8 bg-zinc-950 border-zinc-800 text-xs text-zinc-200"
                  />
                </div>
              </div>

              {/* Row 4: Support Contact Email */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="space-y-0.5 max-w-lg">
                  <span className="text-sm font-medium text-zinc-200">System Support Email</span>
                  <p className="text-xs text-zinc-400 leading-relaxed">
                    Primary administrative contact displayed to users encountering permission or billing gates.
                  </p>
                </div>
                <div className="w-full sm:w-72 shrink-0">
                  <Input
                    value={supportEmail}
                    onChange={(e) => setSupportEmail(e.target.value)}
                    placeholder="support@orbito.io"
                    className="h-8 bg-zinc-950 border-zinc-800 text-xs text-zinc-200"
                  />
                </div>
              </div>

              {/* Row 5: Reference Timezone */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="space-y-0.5 max-w-lg">
                  <span className="text-sm font-medium text-zinc-200">Default Reference Timezone</span>
                  <p className="text-xs text-zinc-400 leading-relaxed">
                    Default time reference for scheduled automated jobs, sprint boundaries, and daily activity digests.
                  </p>
                </div>
                <div className="w-full sm:w-72 shrink-0">
                  <select
                    value={defaultTimezone}
                    onChange={(e) => setDefaultTimezone(e.target.value)}
                    className="w-full h-8 px-2.5 rounded-md bg-zinc-950 border border-zinc-800 text-xs text-zinc-200 outline-none focus:border-zinc-600"
                  >
                    <option value="UTC">UTC (Universal Coordinated)</option>
                    <option value="America/New_York">Eastern Time (US / New York)</option>
                    <option value="America/Los_Angeles">Pacific Time (US / Los Angeles)</option>
                    <option value="Europe/London">London (GMT / BST)</option>
                    <option value="Asia/Tokyo">Tokyo (JST)</option>
                    <option value="Asia/Kolkata">India Standard Time (IST)</option>
                  </select>
                </div>
              </div>
            </div>

            {/* Save Button for General Tab */}
            <div className="flex justify-end">
              <Button
                type="button"
                size="sm"
                onClick={handleSaveGeneral}
                disabled={isSaving}
                className="h-8 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold cursor-pointer transition-colors shadow-xs"
              >
                <Save className="h-3.5 w-3.5 mr-1.5 text-zinc-950" />
                Save General Settings
              </Button>
            </div>
          </TabsContent>

          {/* ========================================================================= */}
          {/* TAB 2: SECURITY & SSO                                                    */}
          {/* ========================================================================= */}
          <TabsContent value="sso" className="space-y-6 outline-none">
            <div className="border border-zinc-800/80 rounded-lg bg-zinc-900/30 divide-y divide-zinc-800/80">
              {/* Row 1: OIDC Master Switch */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="space-y-0.5 max-w-lg">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-zinc-200">OpenID Connect (OIDC) SSO</span>
                    <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono ${isSsoEnabled ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800/60' : 'bg-zinc-800 text-zinc-400 border border-zinc-700'}`}>
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
                    onCheckedChange={handleToggleSso}
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
                      className={`px-2.5 py-1 text-xs rounded-md border transition-colors cursor-pointer ${ssoProvider === p.name ? 'bg-zinc-800 border-zinc-600 text-white font-medium' : 'bg-zinc-950 border-zinc-800 text-zinc-400 hover:text-zinc-200 hover:border-zinc-700'}`}
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
                    onClick={copyCallbackUrl}
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
                  <select
                    value={ssoDefaultRole}
                    onChange={(e) => setSsoDefaultRole(e.target.value)}
                    className="w-full h-8 px-2.5 rounded-md bg-zinc-950 border border-zinc-800 text-xs text-zinc-200 outline-none focus:border-zinc-600"
                  >
                    <option value="MEMBER">MEMBER (Standard Workspace Contributor)</option>
                    <option value="VIEWER">VIEWER (Read-Only Access)</option>
                    <option value="MANAGER">MANAGER (Workspace Manager)</option>
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-zinc-300">Session Inactivity Timeout</label>
                  <select
                    value={sessionTimeout}
                    onChange={(e) => setSessionTimeout(e.target.value)}
                    className="w-full h-8 px-2.5 rounded-md bg-zinc-950 border border-zinc-800 text-xs text-zinc-200 outline-none focus:border-zinc-600"
                  >
                    <option value="15">15 Minutes (High Security)</option>
                    <option value="60">1 Hour (Standard)</option>
                    <option value="240">4 Hours</option>
                    <option value="1440">24 Hours</option>
                    <option value="10080">7 Days (Extended Remember Me)</option>
                  </select>
                </div>
              </div>
            </div>

            {/* Save Button for SSO Tab */}
            <div className="flex justify-end">
              <Button
                type="button"
                size="sm"
                onClick={handleSaveSsoConfig}
                disabled={isSaving}
                className="h-8 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold cursor-pointer transition-colors shadow-xs"
              >
                <Save className="h-3.5 w-3.5 mr-1.5 text-zinc-950" />
                Save SSO & Security Settings
              </Button>
            </div>
          </TabsContent>

          {/* ========================================================================= */}
          {/* TAB 3: AI COPILOT & LLM                                                  */}
          {/* ========================================================================= */}
          <TabsContent value="ai" className="space-y-6 outline-none">
            <div className="border border-zinc-800/80 rounded-lg bg-zinc-900/30 divide-y divide-zinc-800/80">
              {/* Row 1: AI Master Switch */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="space-y-0.5 max-w-lg">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-zinc-200">Orbito AI Copilot & Generative Reasoning</span>
                    <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono ${isAiEnabled ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800/60' : 'bg-zinc-800 text-zinc-400 border border-zinc-700'}`}>
                      {isAiEnabled ? 'AI Active' : 'Disabled'}
                    </span>
                  </div>
                  <p className="text-xs text-zinc-400 leading-relaxed">
                    Enable autonomous task execution, sprint estimation, description drafting, and conversational queries across the platform.
                  </p>
                </div>
                <div className="flex items-center gap-3 shrink-0">
                  <Switch
                    checked={isAiEnabled}
                    onCheckedChange={handleToggleAi}
                    disabled={isSaving}
                    aria-label="Toggle AI Copilot"
                  />
                </div>
              </div>

              {/* Row 2: Provider Quick Presets */}
              <div className="p-4 space-y-3">
                <span className="text-xs font-medium text-zinc-300">Quick Model & Provider Presets</span>
                <div className="flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      setAiApiUrl('https://openrouter.ai/api/v1');
                      setAiModel('anthropic/claude-3.5-sonnet');
                    }}
                    className="px-2.5 py-1 text-xs rounded-md bg-zinc-950 border border-zinc-800 text-zinc-300 hover:border-zinc-600 hover:text-white transition-colors cursor-pointer"
                  >
                    OpenRouter (Claude 3.5 Sonnet)
                  </button>

                  <button
                    type="button"
                    onClick={() => {
                      setAiApiUrl('https://api.openai.com/v1');
                      setAiModel('gpt-4o-mini');
                    }}
                    className="px-2.5 py-1 text-xs rounded-md bg-zinc-950 border border-zinc-800 text-zinc-300 hover:border-zinc-600 hover:text-white transition-colors cursor-pointer"
                  >
                    OpenAI (GPT-4o Mini)
                  </button>

                  <button
                    type="button"
                    onClick={() => {
                      setAiApiUrl('https://api.groq.com/openai/v1');
                      setAiModel('llama-3.3-70b-versatile');
                    }}
                    className="px-2.5 py-1 text-xs rounded-md bg-zinc-950 border border-zinc-800 text-zinc-300 hover:border-zinc-600 hover:text-white transition-colors cursor-pointer"
                  >
                    Groq (Llama 3.3 70B)
                  </button>

                  <button
                    type="button"
                    onClick={() => {
                      setAiApiUrl('http://localhost:11434');
                      setAiModel('llama3.2');
                    }}
                    className="px-2.5 py-1 text-xs rounded-md bg-zinc-950 border border-zinc-800 text-zinc-300 hover:border-zinc-600 hover:text-white transition-colors cursor-pointer"
                  >
                    Ollama Local (llama3.2)
                  </button>
                </div>
              </div>

              {/* Row 3: AI Configuration Inputs */}
              <div className="p-4 space-y-3">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                  <div className="space-y-1.5 sm:col-span-2">
                    <label className="text-[11px] font-medium text-zinc-400">Provider API Base URL</label>
                    <Input
                      value={aiApiUrl}
                      onChange={(e) => setAiApiUrl(e.target.value)}
                      placeholder="https://openrouter.ai/api/v1"
                      className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                    />
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-[11px] font-medium text-zinc-400">Model Identifier</label>
                    <Input
                      value={aiModel}
                      onChange={(e) => setAiModel(e.target.value)}
                      placeholder="anthropic/claude-3.5-sonnet"
                      className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                    />
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-[11px] font-medium text-zinc-400">
                      Provider API Key <span className="text-zinc-500">(Optional for Local Ollama)</span>
                    </label>
                    <div className="relative">
                      <Input
                        type={showAiKey ? 'text' : 'password'}
                        value={aiApiKey}
                        onChange={(e) => setAiApiKey(e.target.value)}
                        placeholder="sk-or-... or sk-..."
                        className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200 pr-8"
                      />
                      <button
                        type="button"
                        onClick={() => setShowAiKey(!showAiKey)}
                        className="absolute right-2 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-200 cursor-pointer"
                      >
                        {showAiKey ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                      </button>
                    </div>
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-[11px] font-medium text-zinc-400">Max Generation Tokens</label>
                    <Input
                      value={aiMaxTokens}
                      onChange={(e) => setAiMaxTokens(e.target.value)}
                      placeholder="4096"
                      className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                    />
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-[11px] font-medium text-zinc-400">Reasoning Temperature (0.0 - 1.0)</label>
                    <Input
                      value={aiTemperature}
                      onChange={(e) => setAiTemperature(e.target.value)}
                      placeholder="0.2"
                      className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                    />
                  </div>
                </div>
              </div>
            </div>

            {/* Action Bar for AI Tab */}
            <div className="flex flex-wrap items-center justify-between gap-3">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleTestAiConnection}
                disabled={isTestingAi}
                className="h-8 text-xs border-zinc-800 bg-zinc-950/60 hover:bg-zinc-850 text-zinc-300 cursor-pointer"
              >
                {isTestingAi ? (
                  <>
                    <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin text-zinc-300" />
                    Testing Endpoint...
                  </>
                ) : (
                  <>
                    <Cpu className="h-3.5 w-3.5 mr-1.5 text-zinc-400" />
                    Test AI Connection
                  </>
                )}
              </Button>

              <Button
                type="button"
                size="sm"
                onClick={handleSaveAiConfig}
                disabled={isSaving}
                className="h-8 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold cursor-pointer transition-colors shadow-xs"
              >
                <Save className="h-3.5 w-3.5 mr-1.5 text-zinc-950" />
                Save AI Configuration
              </Button>
            </div>
          </TabsContent>

          {/* ========================================================================= */}
          {/* TAB 4: MAIL & SMTP                                                       */}
          {/* ========================================================================= */}
          <TabsContent value="email" className="space-y-6 outline-none">
            <div className="border border-zinc-800/80 rounded-lg bg-zinc-900/30 divide-y divide-zinc-800/80">
              {/* Row 1: Header / Status */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="space-y-0.5">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-zinc-200">Outbound Mail Pipeline</span>
                    <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono bg-zinc-800 text-zinc-300 border border-zinc-700">
                      Source: {configData?.smtpSources?.hostSource || 'Environment (.env)'}
                    </span>
                  </div>
                  <p className="text-xs text-zinc-400">
                    Transactional email delivery for workspace invitations, password recovery, and assignment pings.
                  </p>
                </div>
              </div>

              {/* Row 2: Host & Port */}
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
            </div>

            {/* Action Bar for SMTP */}
            <div className="flex flex-wrap items-center justify-between gap-3">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleTestSmtp}
                disabled={isTestingSmtp}
                className="h-8 text-xs border-zinc-800 bg-zinc-950/60 hover:bg-zinc-850 text-zinc-300 cursor-pointer"
              >
                <Send className="h-3.5 w-3.5 mr-1.5 text-zinc-400" />
                {isTestingSmtp ? 'Testing Handshake...' : 'Test SMTP Connection'}
              </Button>

              <Button
                type="button"
                size="sm"
                onClick={handleSaveSmtpConfig}
                disabled={isSaving}
                className="h-8 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold cursor-pointer transition-colors shadow-xs"
              >
                <Save className="h-3.5 w-3.5 mr-1.5 text-zinc-950" />
                Save Mail Configuration
              </Button>
            </div>
          </TabsContent>

          {/* ========================================================================= */}
          {/* TAB 5: STORAGE & CLOUD MEDIA                                             */}
          {/* ========================================================================= */}
          <TabsContent value="storage" className="space-y-6 outline-none">
            <div className="border border-zinc-800/80 rounded-lg bg-zinc-900/30 divide-y divide-zinc-800/80">
              {/* Row 1: Storage Mode Selector */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="space-y-0.5 max-w-lg">
                  <span className="text-sm font-medium text-zinc-200">Asset Storage Engine</span>
                  <p className="text-xs text-zinc-400">
                    Where task attachments, rich-editor images, and user avatars are persisted.
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => setStorageProvider('local')}
                    className={`px-3 py-1.5 text-xs rounded-md border transition-colors cursor-pointer ${storageProvider === 'local' ? 'bg-zinc-800 border-zinc-600 text-white font-medium' : 'bg-zinc-950 border-zinc-800 text-zinc-400 hover:text-zinc-200'}`}
                  >
                    Local Filesystem (/uploads)
                  </button>
                  <button
                    type="button"
                    onClick={() => setStorageProvider('s3')}
                    className={`px-3 py-1.5 text-xs rounded-md border transition-colors cursor-pointer ${storageProvider === 's3' ? 'bg-zinc-800 border-zinc-600 text-white font-medium' : 'bg-zinc-950 border-zinc-800 text-zinc-400 hover:text-zinc-200'}`}
                  >
                    Amazon S3 / S3-Compatible
                  </button>
                </div>
              </div>

              {/* Row 2: S3 Parameters (shown if S3 is active) */}
              {storageProvider === 's3' && (
                <div className="p-4 space-y-3">
                  <span className="text-xs font-medium text-zinc-300">S3 Cloud Credentials & Bucket</span>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                    <div className="space-y-1.5">
                      <label className="text-[11px] font-medium text-zinc-400">S3 Bucket Name</label>
                      <Input
                        value={s3Bucket}
                        onChange={(e) => setS3Bucket(e.target.value)}
                        placeholder="orbito-production-assets"
                        className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                      />
                    </div>

                    <div className="space-y-1.5">
                      <label className="text-[11px] font-medium text-zinc-400">AWS Region</label>
                      <Input
                        value={s3Region}
                        onChange={(e) => setS3Region(e.target.value)}
                        placeholder="us-east-1"
                        className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                      />
                    </div>

                    <div className="space-y-1.5 sm:col-span-2">
                      <label className="text-[11px] font-medium text-zinc-400">
                        Custom S3 Endpoint <span className="text-zinc-500">(Optional for MinIO, Cloudflare R2, Wasabi)</span>
                      </label>
                      <Input
                        value={s3Endpoint}
                        onChange={(e) => setS3Endpoint(e.target.value)}
                        placeholder="https://<account-id>.r2.cloudflarestorage.com"
                        className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                      />
                    </div>

                    <div className="space-y-1.5">
                      <label className="text-[11px] font-medium text-zinc-400">Access Key ID</label>
                      <Input
                        value={s3AccessKey}
                        onChange={(e) => setS3AccessKey(e.target.value)}
                        placeholder="AKIAIOSFODNN7EXAMPLE"
                        className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                      />
                    </div>

                    <div className="space-y-1.5">
                      <label className="text-[11px] font-medium text-zinc-400">Secret Access Key</label>
                      <div className="relative">
                        <Input
                          type={showS3Secret ? 'text' : 'password'}
                          value={s3SecretKey}
                          onChange={(e) => setS3SecretKey(e.target.value)}
                          placeholder="••••••••••••••••••••••••"
                          className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200 pr-8"
                        />
                        <button
                          type="button"
                          onClick={() => setShowS3Secret(!showS3Secret)}
                          className="absolute right-2 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-200 cursor-pointer"
                        >
                          {showS3Secret ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Row 3: Max File Size Limit */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="space-y-0.5 max-w-lg">
                  <span className="text-sm font-medium text-zinc-200">Max Upload File Size</span>
                  <p className="text-xs text-zinc-400">
                    Maximum allowed upload size per file across task attachments and editor images.
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <Input
                    type="number"
                    value={maxUploadMb}
                    onChange={(e) => setMaxUploadMb(e.target.value)}
                    className="h-8 w-24 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                  />
                  <span className="text-xs text-zinc-400 font-mono">MB</span>
                </div>
              </div>
            </div>

            {/* Save Button for Storage Tab */}
            <div className="flex justify-end">
              <Button
                type="button"
                size="sm"
                onClick={handleSaveStorageConfig}
                disabled={isSaving}
                className="h-8 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold cursor-pointer transition-colors shadow-xs"
              >
                <Save className="h-3.5 w-3.5 mr-1.5 text-zinc-950" />
                Save Storage Settings
              </Button>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </RoleGuard>
  );
}
