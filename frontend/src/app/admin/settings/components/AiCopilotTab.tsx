'use client';

import * as React from 'react';
import {
  Sparkles,
  Bot,
  Cpu,
  Save,
  Loader2,
  Eye,
  EyeOff,
  Sliders,
  CheckCircle2,
  Zap,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

interface AiCopilotTabProps {
  isAiEnabled: boolean;
  onToggleAi: () => void;
  aiApiUrl: string;
  setAiApiUrl: (val: string) => void;
  aiApiKey: string;
  setAiApiKey: (val: string) => void;
  showAiKey: boolean;
  setShowAiKey: (val: boolean) => void;
  aiModel: string;
  setAiModel: (val: string) => void;
  aiMaxTokens: string;
  setAiMaxTokens: (val: string) => void;
  aiTemperature: string;
  setAiTemperature: (val: string) => void;
  aiTone: string;
  setAiTone: (val: string) => void;
  aiGreetingMessage: string;
  setAiGreetingMessage: (val: string) => void;
  aiSystemPrompt: string;
  setAiSystemPrompt: (val: string) => void;
  featTaskBreakdown: boolean;
  setFeatTaskBreakdown: (val: boolean) => void;
  featDescriptionGen: boolean;
  setFeatDescriptionGen: (val: boolean) => void;
  featCodeReview: boolean;
  setFeatCodeReview: (val: boolean) => void;
  featSprintSummary: boolean;
  setFeatSprintSummary: (val: boolean) => void;
  onTestConnection: () => void;
  isTesting: boolean;
  onSave: () => void;
  isSaving: boolean;
}

const PROVIDER_PRESETS = [
  {
    name: 'AWS Bedrock',
    tag: 'Cloud Gateway',
    url: 'https://bedrock-runtime.us-east-1.amazonaws.com',
    model: 'anthropic.claude-3-5-sonnet-20241022-v2:0',
  },
  {
    name: 'Groq',
    tag: 'Ultra-Fast',
    url: 'https://api.groq.com/openai/v1',
    model: 'llama-3.3-70b-versatile',
  },
  {
    name: 'OpenRouter',
    tag: 'Multi-Model',
    url: 'https://openrouter.ai/api/v1',
    model: 'anthropic/claude-3.5-sonnet',
  },
  {
    name: 'OpenAI',
    tag: 'Direct API',
    url: 'https://api.openai.com/v1',
    model: 'gpt-4o-mini',
  },
  {
    name: 'Anthropic',
    tag: 'Direct API',
    url: 'https://api.anthropic.com/v1',
    model: 'claude-3-5-sonnet-20241022',
  },
  {
    name: 'Google Gemini',
    tag: 'Direct API',
    url: 'https://generativelanguage.googleapis.com/v1beta',
    model: 'gemini-1.5-pro-latest',
  },
  {
    name: 'Ollama Local',
    tag: 'Self-Hosted',
    url: 'http://localhost:11434',
    model: 'llama3.2',
  },
];

const TONE_OPTIONS = [
  {
    id: 'Professional',
    label: 'Professional',
    description: 'Clean, objective enterprise voice suitable for corporate reporting and task tracking.',
  },
  {
    id: 'Senior Architect',
    label: 'Senior Architect',
    description: 'Deep technical rigor, design patterns, microservices architecture, and clean code hygiene.',
  },
  {
    id: 'Scrum Master',
    label: 'Scrum Master',
    description: 'Agile ceremonies, user story refinement, sprint cadence, and blocker elimination.',
  },
  {
    id: 'Casual & Direct',
    label: 'Casual & Direct',
    description: 'Fast, concise answers without corporate fluff or excessive boilerplate.',
  },
];

export function AiCopilotTab({
  isAiEnabled,
  onToggleAi,
  aiApiUrl,
  setAiApiUrl,
  aiApiKey,
  setAiApiKey,
  showAiKey,
  setShowAiKey,
  aiModel,
  setAiModel,
  aiMaxTokens,
  setAiMaxTokens,
  aiTemperature,
  setAiTemperature,
  aiTone,
  setAiTone,
  aiGreetingMessage,
  setAiGreetingMessage,
  aiSystemPrompt,
  setAiSystemPrompt,
  featTaskBreakdown,
  setFeatTaskBreakdown,
  featDescriptionGen,
  setFeatDescriptionGen,
  featCodeReview,
  setFeatCodeReview,
  featSprintSummary,
  setFeatSprintSummary,
  onTestConnection,
  isTesting,
  onSave,
  isSaving,
}: AiCopilotTabProps) {
  return (
    <div className="space-y-6">
      <div className="border border-zinc-800/80 rounded-lg bg-zinc-900/30 divide-y divide-zinc-800/80">
        {/* Row 1: AI Master Switch */}
        <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="space-y-0.5 max-w-lg">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-zinc-200">Orbito AI Copilot & Generative Reasoning</span>
              <span
                className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono ${
                  isAiEnabled
                    ? 'bg-emerald-950/80 text-emerald-400 border border-emerald-800/60'
                    : 'bg-zinc-800 text-zinc-400 border border-zinc-700'
                }`}
              >
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
              onCheckedChange={onToggleAi}
              disabled={isSaving}
              aria-label="Toggle AI Copilot"
            />
          </div>
        </div>

        {/* Row 2: Provider Quick Presets */}
        <div className="p-4 space-y-2.5">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-zinc-300">Supported LLM Provider Presets</span>
            <span className="text-[11px] text-zinc-500">Clicking populates URL & recommended model</span>
          </div>
          <div className="flex flex-wrap gap-2">
            {PROVIDER_PRESETS.map((p) => {
              const isActive = aiApiUrl.includes(p.url) || aiModel === p.model;
              return (
                <button
                  key={p.name}
                  type="button"
                  onClick={() => {
                    setAiApiUrl(p.url);
                    setAiModel(p.model);
                  }}
                  className={`px-2.5 py-1.5 text-xs rounded-md border transition-all cursor-pointer flex items-center gap-1.5 ${
                    isActive
                      ? 'bg-zinc-800 border-zinc-600 text-white font-medium shadow-xs'
                      : 'bg-zinc-950 border-zinc-800 text-zinc-400 hover:text-zinc-200 hover:border-zinc-700'
                  }`}
                >
                  <span>{p.name}</span>
                  <span className="text-[9px] font-mono px-1 py-0.2 rounded bg-zinc-900 border border-zinc-750 text-zinc-400">
                    {p.tag}
                  </span>
                </button>
              );
            })}
          </div>
        </div>

        {/* Row 3: Primary Connection Inputs */}
        <div className="p-4 space-y-3">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
            <div className="space-y-1.5 sm:col-span-2">
              <label className="text-[11px] font-medium text-zinc-400">Provider API Base URL</label>
              <Input
                value={aiApiUrl}
                onChange={(e) => setAiApiUrl(e.target.value)}
                placeholder="https://openrouter.ai/api/v1 or https://api.groq.com/openai/v1"
                className="h-8 bg-zinc-950 border-zinc-800 text-xs font-mono text-zinc-200"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400">Model Identifier</label>
              <Input
                value={aiModel}
                onChange={(e) => setAiModel(e.target.value)}
                placeholder="anthropic/claude-3.5-sonnet or llama-3.3-70b-versatile"
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

        {/* Row 4: AI Personality Tone & Communication Style */}
        <div className="p-4 space-y-3">
          <div className="space-y-0.5">
            <span className="text-sm font-medium text-zinc-200">AI Personality Tone & Style</span>
            <p className="text-xs text-zinc-400">
              Select how Orbito Copilot structures its responses and recommendations across the platform.
            </p>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5 pt-1">
            {TONE_OPTIONS.map((t) => {
              const isSelected = aiTone === t.id;
              return (
                <div
                  key={t.id}
                  onClick={() => setAiTone(t.id)}
                  className={`p-3 rounded-md border cursor-pointer transition-all text-xs ${
                    isSelected
                      ? 'bg-zinc-850 border-zinc-600 text-zinc-100 shadow-xs'
                      : 'bg-zinc-950 border-zinc-800 text-zinc-400 hover:border-zinc-700 hover:text-zinc-300'
                  }`}
                >
                  <div className="flex items-center justify-between font-medium text-zinc-200 mb-1">
                    <span>{t.label}</span>
                    {isSelected && <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />}
                  </div>
                  <p className="text-[11px] text-zinc-400 leading-normal">{t.description}</p>
                </div>
              );
            })}
          </div>
        </div>

        {/* Row 5: Initial Greeting / Welcome Message */}
        <div className="p-4 flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
          <div className="space-y-0.5 max-w-lg">
            <span className="text-sm font-medium text-zinc-200">Copilot Initial Greeting Message</span>
            <p className="text-xs text-zinc-400 leading-relaxed">
              The welcome greeting displayed to teammates when opening the AI assistant drawer.
            </p>
          </div>
          <div className="w-full sm:w-80 shrink-0">
            <Input
              value={aiGreetingMessage}
              onChange={(e) => setAiGreetingMessage(e.target.value)}
              placeholder="How can I help you today?"
              className="h-8 bg-zinc-950 border-zinc-800 text-xs text-zinc-200"
            />
          </div>
        </div>

        {/* Row 6: Custom System Prompt Instructions */}
        <div className="p-4 space-y-2">
          <div className="space-y-0.5">
            <span className="text-sm font-medium text-zinc-200">Custom Organization System Prompt</span>
            <p className="text-xs text-zinc-400">
              Provide additional organization-specific rules, security constraints, or task conventions injected directly into the LLM context.
            </p>
          </div>
          <textarea
            value={aiSystemPrompt}
            onChange={(e) => setAiSystemPrompt(e.target.value)}
            rows={3}
            placeholder="e.g. Always format dates as YYYY-MM-DD. Enforce strict acceptance criteria formatting with Gherkin Given/When/Then syntax. Highlight security and RBAC implications for all administrative tasks."
            className="w-full p-2.5 rounded-md bg-zinc-950 border border-zinc-800 text-xs text-zinc-200 placeholder:text-zinc-600 focus:outline-hidden focus:border-zinc-600 resize-y font-mono leading-relaxed"
          />
        </div>

        {/* Row 7: AI Autonomous Capability Switches */}
        <div className="p-4 space-y-3">
          <div className="space-y-0.5">
            <span className="text-sm font-medium text-zinc-200">Autonomous Capabilities & Tool Access</span>
            <p className="text-xs text-zinc-400">
              Select which specialized AI skill modules are active for organization members.
            </p>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
            <div className="flex items-center justify-between p-2.5 rounded-md bg-zinc-950 border border-zinc-800/80">
              <div className="space-y-0.5 pr-2">
                <span className="text-xs font-medium text-zinc-200">Task Decomposition & Estimation</span>
                <p className="text-[11px] text-zinc-400">Break large user stories into deliverable subtasks.</p>
              </div>
              <Switch
                checked={featTaskBreakdown}
                onCheckedChange={setFeatTaskBreakdown}
                aria-label="Toggle Task Decomposition"
              />
            </div>

            <div className="flex items-center justify-between p-2.5 rounded-md bg-zinc-950 border border-zinc-800/80">
              <div className="space-y-0.5 pr-2">
                <span className="text-xs font-medium text-zinc-200">Generative Task Descriptions</span>
                <p className="text-[11px] text-zinc-400">Draft rich acceptance criteria from brief task titles.</p>
              </div>
              <Switch
                checked={featDescriptionGen}
                onCheckedChange={setFeatDescriptionGen}
                aria-label="Toggle Description Generation"
              />
            </div>

            <div className="flex items-center justify-between p-2.5 rounded-md bg-zinc-950 border border-zinc-800/80">
              <div className="space-y-0.5 pr-2">
                <span className="text-xs font-medium text-zinc-200">Technical Code & PR Review</span>
                <p className="text-[11px] text-zinc-400">Surface architectural flaws and security vulnerabilities.</p>
              </div>
              <Switch
                checked={featCodeReview}
                onCheckedChange={setFeatCodeReview}
                aria-label="Toggle Code Review"
              />
            </div>

            <div className="flex items-center justify-between p-2.5 rounded-md bg-zinc-950 border border-zinc-800/80">
              <div className="space-y-0.5 pr-2">
                <span className="text-xs font-medium text-zinc-200">Sprint Summaries & Blocker Detection</span>
                <p className="text-[11px] text-zinc-400">Synthesize sprint burndown risks and team velocity.</p>
              </div>
              <Switch
                checked={featSprintSummary}
                onCheckedChange={setFeatSprintSummary}
                aria-label="Toggle Sprint Summary"
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
          onClick={onTestConnection}
          disabled={isTesting}
          className="h-8 text-xs border-zinc-800 bg-zinc-950/60 hover:bg-zinc-850 text-zinc-300 cursor-pointer"
        >
          {isTesting ? (
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
          onClick={onSave}
          disabled={isSaving}
          className="h-8 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold cursor-pointer transition-colors shadow-xs"
        >
          <Save className="h-3.5 w-3.5 mr-1.5 text-zinc-950" />
          Save AI Configuration
        </Button>
      </div>
    </div>
  );
}
