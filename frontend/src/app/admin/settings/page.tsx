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
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

export default function AdminSettingsPage() {
  const { data: configData, isLoading: isConfigLoading } = useGetAdminConfigQuery();
  const { data: regStatus, refetch: refetchRegStatus } = useGetRegistrationStatusQuery();

  const [saveConfig, { isLoading: isSaving }] = useSaveAdminConfigMutation();
  const [testSmtp, { isLoading: isTestingSmtp }] = useTestAdminSmtpMutation();
  const [testAiConnection, { isLoading: isTestingAi }] = useTestAiConnectionMutation();

  // AI Configuration State
  const [aiApiUrl, setAiApiUrl] = useState('');
  const [aiApiKey, setAiApiKey] = useState('');
  const [aiModel, setAiModel] = useState('');
  const [hasInitializedAi, setHasInitializedAi] = useState(false);

  useEffect(() => {
    if (configData?.settings && !hasInitializedAi) {
      const url = configData.settings.find((s) => s.key === 'ai_api_url')?.value || '';
      const key = configData.settings.find((s) => s.key === 'ai_api_key')?.value || '';
      const model = configData.settings.find((s) => s.key === 'ai_model')?.value || '';
      if (url || key || model) {
        setAiApiUrl(url);
        setAiApiKey(key);
        setAiModel(model);
        setHasInitializedAi(true);
      }
    }
  }, [configData, hasInitializedAi]);

  // Find registration setting
  const regSetting = configData?.settings.find((s) => s.key === 'enablePublicRegistration');
  const isPublicRegEnabled = regSetting ? regSetting.value === 'true' : (regStatus?.enabled ?? false);

  // Find AI status
  const aiEnabledSetting = configData?.settings.find((s) => s.key === 'ai_enabled');
  const isAiEnabled = aiEnabledSetting ? aiEnabledSetting.value === 'true' : false;

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

  const handleSaveAiConfig = async () => {
    try {
      await saveConfig({
        settings: [
          {
            key: 'ai_api_url',
            value: aiApiUrl.trim(),
            category: 'ai',
            description: 'AI Provider Base URL',
          },
          {
            key: 'ai_api_key',
            value: aiApiKey.trim(),
            category: 'ai',
            description: 'AI Provider API Key',
          },
          {
            key: 'ai_model',
            value: aiModel.trim(),
            category: 'ai',
            description: 'AI Model Identifier',
          },
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
              <Badge variant="outline" className="text-[10px] uppercase font-mono tracking-wider border-zinc-700 bg-zinc-850/80 text-zinc-300">
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
                  <div className="h-8 w-8 rounded-md bg-zinc-800/80 border border-zinc-700/80 flex items-center justify-center text-zinc-200">
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

                <div className="flex items-center gap-3">
                  <span className="text-xs font-mono font-medium text-zinc-400">
                    {isPublicRegEnabled ? 'Enabled' : 'Restricted'}
                  </span>
                  <Switch
                    checked={isPublicRegEnabled}
                    onCheckedChange={handleToggleRegistration}
                    disabled={isSaving}
                    aria-label="Toggle Public Registration"
                  />
                </div>
              </div>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="p-3 rounded-md bg-zinc-950/60 border border-zinc-800/80 flex items-start gap-2.5 text-xs text-zinc-400">
                <span className={`h-1.5 w-1.5 rounded-full mt-1.5 shrink-0 ${isPublicRegEnabled ? 'bg-emerald-400' : 'bg-zinc-600'}`} />
                <p className="leading-relaxed">
                  {isPublicRegEnabled ? (
                    <>Self-service registration is <strong className="text-zinc-200 font-medium">active</strong> at <code className="text-zinc-300 px-1 py-0.5 rounded bg-zinc-900 border border-zinc-800 font-mono">/register</code>. Any user can register their own tenant organization.</>
                  ) : (
                    <>Public registration is <strong className="text-zinc-200 font-medium">restricted</strong>. Only users with a direct invite token from an existing workspace or organization can onboard.</>
                  )}
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Orbito AI Copilot & LLM Service */}
          <Card className="border-zinc-800 bg-zinc-900/60">
            <CardHeader>
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="flex items-center gap-2.5">
                  <div className="h-8 w-8 rounded-md bg-zinc-800/80 border border-zinc-700/80 flex items-center justify-center text-zinc-200">
                    <Sparkles className="h-4 w-4" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <CardTitle className="text-sm font-semibold text-zinc-100">
                        Orbito AI Copilot & LLM Provider
                      </CardTitle>
                      <Badge variant="outline" className="text-[10px] font-mono tracking-wider border-zinc-800 bg-zinc-850/60 text-zinc-400">
                        Generative AI
                      </Badge>
                    </div>
                    <CardDescription className="text-xs text-zinc-400">
                      Configure autonomous task reasoning, sprint analysis, and description drafting models.
                    </CardDescription>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <span className="text-xs font-mono font-medium text-zinc-400">
                    {isAiEnabled ? 'Enabled' : 'Disabled'}
                  </span>
                  <Switch
                    checked={isAiEnabled}
                    onCheckedChange={handleToggleAi}
                    disabled={isSaving}
                    aria-label="Toggle AI Copilot"
                  />
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4 pt-0">
              {/* Provider Quick Presets */}
              <div className="space-y-1.5">
                <span className="text-[11px] font-medium text-zinc-400">Quick Provider Presets</span>
                <div className="flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      setAiApiUrl('https://openrouter.ai/api/v1');
                      setAiModel('anthropic/claude-3.5-sonnet');
                    }}
                    className="px-2.5 py-1 text-xs rounded-md bg-zinc-950 border border-zinc-800 text-zinc-300 hover:border-zinc-600 hover:text-white hover:bg-zinc-900 transition-colors cursor-pointer"
                  >
                    OpenRouter (Claude 3.5 Sonnet)
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setAiApiUrl('https://api.openai.com/v1');
                      setAiModel('gpt-4o-mini');
                    }}
                    className="px-2.5 py-1 text-xs rounded-md bg-zinc-950 border border-zinc-800 text-zinc-300 hover:border-zinc-600 hover:text-white hover:bg-zinc-900 transition-colors cursor-pointer"
                  >
                    OpenAI (GPT-4o Mini)
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setAiApiUrl('http://localhost:11434');
                      setAiModel('llama3.2');
                    }}
                    className="px-2.5 py-1 text-xs rounded-md bg-zinc-950 border border-zinc-800 text-zinc-300 hover:border-zinc-600 hover:text-white hover:bg-zinc-900 transition-colors cursor-pointer"
                  >
                    Ollama Local (llama3.2)
                  </button>
                </div>
              </div>

              {/* Input Form Fields */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
                <div className="space-y-1.5 md:col-span-2">
                  <label className="text-[11px] font-medium text-zinc-400">
                    Provider API Endpoint URL
                  </label>
                  <Input
                    value={aiApiUrl}
                    onChange={(e) => setAiApiUrl(e.target.value)}
                    placeholder="https://openrouter.ai/api/v1"
                    className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-[11px] font-medium text-zinc-400">
                    Model Identifier
                  </label>
                  <Input
                    value={aiModel}
                    onChange={(e) => setAiModel(e.target.value)}
                    placeholder="anthropic/claude-3.5-sonnet or gpt-4o"
                    className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-[11px] font-medium text-zinc-400">
                    Provider API Key <span className="text-zinc-500">(Optional for Ollama)</span>
                  </label>
                  <Input
                    type="password"
                    value={aiApiKey}
                    onChange={(e) => setAiApiKey(e.target.value)}
                    placeholder="sk-or-... or sk-..."
                    className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
                  />
                </div>
              </div>

              {/* Action Buttons */}
              <div className="flex flex-wrap items-center justify-between gap-3 pt-3 border-t border-zinc-800">
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
                  className="h-8 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold shadow-xs cursor-pointer transition-colors"
                >
                  Save AI Configuration
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* OpenID Connect (OIDC) / Enterprise SSO */}
          <Card className="border-zinc-800 bg-zinc-900/60">
            <CardHeader>
              <div className="flex items-center gap-2.5">
                <div className="h-8 w-8 rounded-md bg-zinc-800/80 border border-zinc-700/80 flex items-center justify-center text-zinc-200">
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
                  <div className="flex items-center gap-2 font-medium text-zinc-200">
                    <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
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
                  <div className="h-8 w-8 rounded-md bg-zinc-800/80 border border-zinc-700/80 flex items-center justify-center text-zinc-200">
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
                  className="h-8 text-xs border-zinc-800 text-zinc-300 hover:bg-zinc-850 cursor-pointer"
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

