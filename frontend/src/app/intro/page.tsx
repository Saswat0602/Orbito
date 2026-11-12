'use client';

import * as React from 'react';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  Compass,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  ArrowRight,
  Sparkles,
  Layers,
  Code2,
  Briefcase,
  Users,
  Kanban,
  Calendar,
  Zap,
  Target,
  SkipForward,
} from 'lucide-react';
import { toast } from 'sonner';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useUpdateProfileMutation } from '@/lib/store/api/usersApi';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

interface Question {
  id: string;
  title: string;
  description: string;
  options: Array<{
    value: string;
    label: string;
    description: string;
    icon: React.ElementType;
  }>;
}

const QUESTIONS: Question[] = [
  {
    id: 'primaryGoal',
    title: 'What brings you to Orbito today?',
    description: 'We will personalize your initial navigation and metrics based on your main goal.',
    options: [
      {
        value: 'sprint_delivery',
        label: 'Engineering & Sprint Delivery',
        description: 'Run agile sprints, manage backlogs, track velocity, and execute releases.',
        icon: Zap,
      },
      {
        value: 'product_roadmapping',
        label: 'Product Roadmapping & Epics',
        description: 'Organize features, prioritize requirements, and coordinate cross-team deliverables.',
        icon: Target,
      },
      {
        value: 'executive_portfolio',
        label: 'Executive Portfolio Oversight',
        description: 'Monitor multi-workspace health, resource utilization, and overall project completion.',
        icon: Layers,
      },
      {
        value: 'personal_focus',
        label: 'Personal Task & Daily Flow',
        description: 'Time tracking, daily agenda briefings, and distraction-free task completion.',
        icon: Compass,
      },
    ],
  },
  {
    id: 'role',
    title: 'What best describes your role?',
    description: 'This helps customize your default workspace permissions and views.',
    options: [
      {
        value: 'architect',
        label: 'Staff / Software Architect',
        description: 'Focusing on clean systems architecture, dependencies, and code standards.',
        icon: Code2,
      },
      {
        value: 'engineering_lead',
        label: 'Engineering Lead / Manager',
        description: 'Leading squad sprints, unblocking team members, and reviewing velocity.',
        icon: Users,
      },
      {
        value: 'product_manager',
        label: 'Product Manager / Owner',
        description: 'Defining backlog scopes, milestone criteria, and client requirements.',
        icon: Briefcase,
      },
      {
        value: 'fullstack_engineer',
        label: 'Fullstack / Frontend Engineer',
        description: 'Building tasks, logging hours, and collaborating through code reviews.',
        icon: Sparkles,
      },
    ],
  },
  {
    id: 'teamSize',
    title: 'How large is your active squad or team?',
    description: 'We will tune workspace notifications and member directories accordingly.',
    options: [
      {
        value: 'solo',
        label: 'Just Me (Solo)',
        description: 'Independent engineering or personal projects.',
        icon: Users,
      },
      {
        value: '2-10',
        label: 'Small Squad (2–10 members)',
        description: 'Fast-paced, close-knit collaborative delivery.',
        icon: Users,
      },
      {
        value: '11-50',
        label: 'Scaleup Team (11–50 members)',
        description: 'Multiple projects, sprint cadence, and cross-functional teams.',
        icon: Users,
      },
      {
        value: '50+',
        label: 'Enterprise (50+ members)',
        description: 'Multi-workspace governance, SSO, and executive audit logging.',
        icon: Layers,
      },
    ],
  },
  {
    id: 'methodology',
    title: 'What delivery methodology do you prefer?',
    description: 'You can change this or mix views at any time across all projects.',
    options: [
      {
        value: 'scrum',
        label: 'Agile Scrum Sprints',
        description: 'Timeboxed sprints, story points, burndown metrics, and velocity charts.',
        icon: Zap,
      },
      {
        value: 'kanban',
        label: 'Kanban Board Flow',
        description: 'Continuous delivery with visual column WIP limits and status stages.',
        icon: Kanban,
      },
      {
        value: 'gantt',
        label: 'Gantt Timeline & Dependencies',
        description: 'Critical path planning, blocker trees, and milestone scheduling.',
        icon: Calendar,
      },
    ],
  },
];

export default function OnboardingIntroPage() {
  const router = useRouter();
  const { user } = useAuthRbac();
  const [updateProfile, { isLoading: isSaving }] = useUpdateProfileMutation();

  const [currentStep, setCurrentStep] = useState(0);
  const [answers, setAnswers] = useState<Record<string, string>>({});

  const question = QUESTIONS[currentStep];
  const totalSteps = QUESTIONS.length;
  const progressPercent = Math.round(((currentStep + 1) / totalSteps) * 100);

  const selectedValue = answers[question.id] || '';

  const handleSelect = (val: string) => {
    setAnswers((prev) => ({ ...prev, [question.id]: val }));
  };

  const handleNext = async () => {
    if (currentStep < totalSteps - 1) {
      setCurrentStep((prev) => prev + 1);
      return;
    }

    // Final step complete: save answers to preferences
    try {
      const existingPrefs = (user?.preferences as Record<string, unknown> | undefined) || {};
      await updateProfile({
        preferences: {
          ...existingPrefs,
          onboarding: answers,
          onboardingCompleted: true,
          onboardedAt: new Date().toISOString(),
        },
      }).unwrap();

      toast.success('Welcome to Orbito! Your workspace experience is configured.');
      router.push('/');
    } catch {
      // Even if saving preferences fails, continue to workspace
      router.push('/');
    }
  };

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep((prev) => prev - 1);
    }
  };

  const handleSkip = () => {
    router.push('/');
  };

  return (
    <div className="min-h-screen bg-[#09090b] text-zinc-100 flex flex-col justify-between p-4 sm:p-8">
      {/* Top Bar */}
      <header className="max-w-3xl w-full mx-auto flex items-center justify-between py-2">
        <div className="flex items-center gap-2">
          <div className="h-6 w-6 rounded-md bg-zinc-100 text-zinc-900 flex items-center justify-center font-bold text-xs">
            O
          </div>
          <span className="text-xs font-bold tracking-tight text-zinc-200">Orbito Experience</span>
        </div>

        <Button
          variant="ghost"
          size="sm"
          onClick={handleSkip}
          className="text-xs text-zinc-500 hover:text-zinc-300 hover:bg-zinc-850 gap-1"
        >
          <span>Skip to Workspace</span>
          <SkipForward className="h-3 w-3" />
        </Button>
      </header>

      {/* Main Wizard Form */}
      <main className="max-w-2xl w-full mx-auto my-auto space-y-6">
        {/* Progress Header */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-xs text-zinc-400">
            <span className="font-mono">
              Step {currentStep + 1} of {totalSteps}
            </span>
            <span className="text-zinc-500">{progressPercent}% Completed</span>
          </div>
          <div className="h-1.5 w-full bg-zinc-800 rounded-full overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-violet-500 to-indigo-500 transition-all duration-300 ease-out"
              style={{ width: `${progressPercent}%` }}
            />
          </div>
        </div>

        {/* Question Card */}
        <Card className="border-zinc-800 bg-zinc-900/80 shadow-2xl">
          <CardHeader className="pb-4">
            <Badge
              variant="outline"
              className="w-fit text-[10px] font-mono border-zinc-800 text-zinc-400 mb-1"
            >
              Question {currentStep + 1}
            </Badge>
            <CardTitle className="text-lg font-bold text-zinc-100">
              {question.title}
            </CardTitle>
            <CardDescription className="text-xs text-zinc-400">
              {question.description}
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-3 pt-0">
            {question.options.map((opt) => {
              const Icon = opt.icon;
              const isSelected = selectedValue === opt.value;

              return (
                <button
                  key={opt.value}
                  type="button"
                  onClick={() => handleSelect(opt.value)}
                  className={`w-full flex items-start gap-3.5 p-3.5 rounded-md border text-left transition-all cursor-pointer ${
                    isSelected
                      ? 'border-violet-500 bg-violet-950/20 text-zinc-100 ring-1 ring-violet-500/50'
                      : 'border-zinc-800/80 bg-zinc-950/40 text-zinc-300 hover:border-zinc-700 hover:bg-zinc-850/40'
                  }`}
                >
                  <div
                    className={`p-2 rounded-md shrink-0 mt-0.5 ${
                      isSelected
                        ? 'bg-violet-500/20 text-violet-300'
                        : 'bg-zinc-850 text-zinc-400'
                    }`}
                  >
                    <Icon className="h-4 w-4" />
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-semibold text-zinc-100">
                        {opt.label}
                      </span>
                      {isSelected && (
                        <CheckCircle2 className="h-4 w-4 text-violet-400 shrink-0" />
                      )}
                    </div>
                    <p className="text-[11px] text-zinc-400 mt-0.5 leading-relaxed">
                      {opt.description}
                    </p>
                  </div>
                </button>
              );
            })}
          </CardContent>

          <CardFooter className="pt-2 pb-5 border-t border-zinc-800/60 flex items-center justify-between">
            <Button
              variant="outline"
              size="sm"
              onClick={handleBack}
              disabled={currentStep === 0}
              className="h-8 text-xs border-zinc-800 text-zinc-400 hover:text-zinc-200 disabled:opacity-30"
            >
              <ChevronLeft className="h-3.5 w-3.5 mr-1" /> Back
            </Button>

            <Button
              size="sm"
              disabled={!selectedValue || isSaving}
              onClick={handleNext}
              className="h-8 text-xs bg-zinc-100 text-zinc-900 hover:bg-zinc-200 font-medium"
            >
              {currentStep === totalSteps - 1 ? (
                <>
                  Complete & Go to Workspace <ArrowRight className="h-3.5 w-3.5 ml-1.5" />
                </>
              ) : (
                <>
                  Next <ChevronRight className="h-3.5 w-3.5 ml-1" />
                </>
              )}
            </Button>
          </CardFooter>
        </Card>
      </main>

      {/* Footer */}
      <footer className="text-center text-[11px] text-zinc-600 py-2">
        Orbito Workspace Personalization &bull; You can update preferences anytime in Profile Settings
      </footer>
    </div>
  );
}
