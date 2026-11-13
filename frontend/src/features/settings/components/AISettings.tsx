'use client';

import * as React from 'react';
import { useState, useEffect } from 'react';
import {
  Bot,
  Sparkles,
  Sliders,
  Mic,
  Code2,
  Kanban,
  Briefcase,
  Save,
  CheckCircle2,
  HelpCircle,
} from 'lucide-react';
import { toast } from 'sonner';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useUpdateProfileMutation } from '@/lib/store/api/usersApi';

interface AiPersona {
  id: string;
  name: string;
  description: string;
  icon: React.ComponentType<{ className?: string }>;
  promptSnippet: string;
}

const PERSONAS: AiPersona[] = [
  {
    id: 'agile_coach',
    name: 'Agile Scrum Master',
    description: 'Specializes in sprint planning, story point estimation, burndown rates, and blocker resolution.',
    icon: Kanban,
    promptSnippet: 'You are an expert Agile Scrum Master guiding a high-velocity software engineering squad.',
  },
  {
    id: 'tech_architect',
    name: 'Staff Software Architect',
    description: 'Enforces clean code, strict architectural layer separation, anti-hallucination, and edge case resilience.',
    icon: Code2,
    promptSnippet: 'You are a Principal Software Architect enforcing enterprise clean code, strict typing, and system safety.',
  },
  {
    id: 'product_manager',
    name: 'Product Manager',
    description: 'Focuses on user stories, acceptance criteria, product goals, and measurable business deliverables.',
    icon: Briefcase,
    promptSnippet: 'You are a Senior Product Manager focused on business impact, clear specifications, and user value.',
  },
  {
    id: 'custom',
    name: 'Custom System Persona',
    description: 'Define your own specialized persona instructions for the Orbito Copilot.',
    icon: Bot,
    promptSnippet: '',
  },
];

export function AISettings() {
  const { user } = useAuthRbac();
  const [updateProfile, { isLoading }] = useUpdateProfileMutation();

  const userAiPrefs = (user?.preferences as { ai?: Record<string, unknown> } | undefined)?.ai || {};

  const [selectedPersona, setSelectedPersona] = useState<string>(
    typeof userAiPrefs.persona === 'string' ? userAiPrefs.persona : 'tech_architect'
  );
  const [temperature, setTemperature] = useState<number>(
    typeof userAiPrefs.temperature === 'number' ? userAiPrefs.temperature : 0.7
  );
  const [autoSendVoice, setAutoSendVoice] = useState<boolean>(
    Boolean(userAiPrefs.autoSendVoice)
  );
  const [customPrompt, setCustomPrompt] = useState<string>(
    typeof userAiPrefs.customPrompt === 'string' ? userAiPrefs.customPrompt : ''
  );

  useEffect(() => {
    if (user?.preferences) {
      const prefs = (user.preferences as { ai?: Record<string, unknown> })?.ai || {};
      if (typeof prefs.persona === 'string') setSelectedPersona(prefs.persona);
      if (typeof prefs.temperature === 'number') setTemperature(prefs.temperature);
      if (typeof prefs.autoSendVoice === 'boolean') setAutoSendVoice(prefs.autoSendVoice);
      if (typeof prefs.customPrompt === 'string') setCustomPrompt(prefs.customPrompt);
    }
  }, [user]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const existingPrefs = (user?.preferences as Record<string, unknown>) || {};
      await updateProfile({
        preferences: {
          ...existingPrefs,
          ai: {
            persona: selectedPersona,
            temperature,
            autoSendVoice,
            customPrompt: selectedPersona === 'custom' ? customPrompt.trim() : '',
          },
        },
      }).unwrap();

      toast.success('AI Assistant preferences saved successfully');
    } catch {
      toast.error('Failed to save AI preferences');
    }
  };

  return (
    <Card className="border-zinc-800 bg-zinc-900/80 max-w-2xl">
      <CardHeader>
        <div className="flex items-center gap-2">
          <Bot className="h-5 w-5 text-zinc-200" />
          <CardTitle className="text-base text-zinc-100">Orbito AI Copilot Preferences</CardTitle>
        </div>
        <CardDescription className="text-xs text-zinc-400">
          Personalize how the AI Assistant (`⌘J`) responds to your queries, drafts task descriptions, and analyzes sprints.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <form onSubmit={handleSave} className="space-y-6">
          {/* Persona Selection */}
          <div className="space-y-3">
            <Label className="text-xs font-semibold text-zinc-300 flex items-center justify-between">
              <span>Assistant Persona</span>
              <span className="text-[11px] font-normal text-zinc-500">
                Select default communication style
              </span>
            </Label>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
              {PERSONAS.map((persona) => {
                const Icon = persona.icon;
                const isSelected = selectedPersona === persona.id;

                return (
                  <div
                    key={persona.id}
                    onClick={() => setSelectedPersona(persona.id)}
                    className={`p-3.5 rounded-md border text-left cursor-pointer transition-all flex flex-col justify-between space-y-2 select-none ${
                      isSelected
                        ? 'bg-zinc-800/90 border-zinc-500 text-zinc-100 ring-1 ring-zinc-500/30'
                        : 'bg-zinc-950/60 border-zinc-800 hover:bg-zinc-900 hover:border-zinc-700'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Icon className={`h-4 w-4 ${isSelected ? 'text-zinc-100' : 'text-zinc-400'}`} />
                        <span className="text-xs font-semibold text-zinc-200">
                          {persona.name}
                        </span>
                      </div>
                      {isSelected && (
                        <CheckCircle2 className="h-3.5 w-3.5 text-zinc-100 shrink-0" />
                      )}
                    </div>
                    <p className="text-[11px] text-zinc-400 line-clamp-2 leading-relaxed">
                      {persona.description}
                    </p>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Custom Persona Textarea */}
          {selectedPersona === 'custom' && (
            <div className="space-y-1.5 animate-in fade-in duration-200">
              <Label htmlFor="customPrompt" className="text-xs text-zinc-300">
                Custom System Persona Instructions
              </Label>
              <textarea
                id="customPrompt"
                rows={3}
                value={customPrompt}
                onChange={(e) => setCustomPrompt(e.target.value)}
                placeholder="e.g. You are a senior QA engineer. Always verify testing coverage, boundary checks, and automated mock specifications..."
                className="w-full rounded-md bg-zinc-950 border border-zinc-800 p-2.5 text-xs text-zinc-200 placeholder:text-zinc-600 focus:outline-none focus:border-zinc-600 focus:ring-1 focus:ring-zinc-600 font-sans"
              />
            </div>
          )}

          {/* Creativity / Temperature Slider */}
          <div className="space-y-2 pt-2 border-t border-zinc-850">
            <div className="flex items-center justify-between">
              <Label htmlFor="temperature" className="text-xs font-semibold text-zinc-300 flex items-center gap-1.5">
                <Sliders className="h-3.5 w-3.5 text-zinc-400" />
                <span>Creativity & Temperature</span>
              </Label>
              <span className="text-xs font-mono font-bold text-zinc-200">
                {temperature.toFixed(2)}
              </span>
            </div>

            <input
              id="temperature"
              type="range"
              min="0.1"
              max="1.0"
              step="0.05"
              value={temperature}
              onChange={(e) => setTemperature(parseFloat(e.target.value))}
              className="w-full h-1.5 bg-zinc-800 rounded-lg appearance-none cursor-pointer accent-zinc-200"
            />

            <div className="flex justify-between text-[10px] text-zinc-500 font-mono">
              <span>0.1 (Precise & Deterministic)</span>
              <span>0.7 (Balanced)</span>
              <span>1.0 (Creative Brainstorming)</span>
            </div>
          </div>

          {/* Voice Dictation Auto-Send Toggle */}
          <div className="pt-2 border-t border-zinc-850">
            <div className="flex items-center justify-between p-3 rounded-md bg-zinc-950/40 border border-zinc-850">
              <div className="flex items-center gap-3">
                <div className="h-8 w-8 rounded-md bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-300 shrink-0">
                  <Mic className="h-4 w-4 text-zinc-300" />
                </div>
                <div>
                  <h5 className="text-xs font-medium text-zinc-200">
                    Voice Dictation Auto-Submit
                  </h5>
                  <p className="text-[11px] text-zinc-500">
                    Automatically dispatch prompts to Copilot when speech recognition detects silence.
                  </p>
                </div>
              </div>

              <Switch
                checked={autoSendVoice}
                onCheckedChange={setAutoSendVoice}
                aria-label="Toggle voice dictation auto-submit"
              />
            </div>
          </div>

          {/* Save Button */}
          <Button
            type="submit"
            size="sm"
            disabled={isLoading}
            className="gap-2 text-xs bg-zinc-100 hover:bg-white text-zinc-950 font-semibold shadow-xs cursor-pointer transition-colors"
          >
            <Save className="h-3.5 w-3.5" />
            <span>{isLoading ? 'Saving...' : 'Save AI Preferences'}</span>
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

