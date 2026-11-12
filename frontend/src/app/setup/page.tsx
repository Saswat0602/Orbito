'use client';

import * as React from 'react';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  ShieldCheck,
  CheckCircle2,
  Lock,
  Mail,
  User,
  ArrowRight,
  Loader2,
  AlertCircle,
  Building2,
  Sparkles,
} from 'lucide-react';
import { toast } from 'sonner';

import {
  useGetSetupRequiredQuery,
  useSetupSuperAdminMutation,
} from '@/lib/store/api/authApi';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';

export default function SetupWizardPage() {
  const router = useRouter();
  const { data: setupState, isLoading: isCheckingSetup } = useGetSetupRequiredQuery();
  const [setupSuperAdmin, { isLoading: isSubmitting }] = useSetupSuperAdminMutation();

  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
  });

  const [formError, setFormError] = useState<string | null>(null);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
    if (formError) setFormError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    if (!formData.firstName.trim() || !formData.lastName.trim()) {
      setFormError('First and last name are required');
      return;
    }

    if (!formData.email.trim()) {
      setFormError('Email address is required');
      return;
    }

    if (formData.password.length < 8) {
      setFormError('Password must be at least 8 characters long');
      return;
    }

    if (formData.password !== formData.confirmPassword) {
      setFormError('Passwords do not match');
      return;
    }

    try {
      await setupSuperAdmin({
        email: formData.email.trim().toLowerCase(),
        password: formData.password,
        firstName: formData.firstName.trim(),
        lastName: formData.lastName.trim(),
        username: formData.username.trim() || undefined,
      }).unwrap();

      toast.success('Super Administrator initialized successfully! Please sign in.');
      router.push('/login');
    } catch (err: unknown) {
      const message =
        typeof err === 'object' && err && 'data' in err && typeof (err as { data?: { message?: string } }).data?.message === 'string'
          ? (err as { data: { message: string } }).data.message
          : 'Failed to complete system setup. Please try again.';
      setFormError(message);
      toast.error(message);
    }
  };

  if (isCheckingSetup) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#09090b] text-zinc-100 p-4">
        <div className="flex flex-col items-center gap-3 text-zinc-500">
          <Loader2 className="h-6 w-6 animate-spin text-zinc-400" />
          <span className="text-xs">Checking system initialization state...</span>
        </div>
      </div>
    );
  }

  // If setup is already done
  if (setupState && !setupState.required) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#09090b] text-zinc-100 p-4">
        <Card className="w-full max-w-md border-zinc-800 bg-zinc-900/80 shadow-2xl">
          <CardHeader className="text-center pb-2">
            <div className="mx-auto w-12 h-12 rounded-md bg-emerald-950/40 border border-emerald-800/60 flex items-center justify-center mb-3 text-emerald-400">
              <CheckCircle2 className="h-6 w-6" />
            </div>
            <CardTitle className="text-lg font-bold text-zinc-100">Setup Already Complete</CardTitle>
            <CardDescription className="text-xs text-zinc-400 mt-1">
              Super administrator initialization has already been executed for this Orbito instance.
            </CardDescription>
          </CardHeader>
          <CardContent className="pt-4 text-center">
            <p className="text-xs text-zinc-400 leading-relaxed">
              You can sign in with your credentials to access the platform workspaces and management console.
            </p>
          </CardContent>
          <CardFooter className="pt-2 flex justify-center">
            <Button
              asChild
              className="w-full h-9 text-xs bg-zinc-100 text-zinc-900 hover:bg-zinc-200"
            >
              <Link href="/login">
                Proceed to Sign In <ArrowRight className="h-3.5 w-3.5 ml-1.5" />
              </Link>
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col justify-center items-center bg-[#09090b] text-zinc-100 p-4 sm:p-6">
      <div className="w-full max-w-lg space-y-6">
        {/* Brand Header */}
        <div className="text-center space-y-2">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-md border border-zinc-800 bg-zinc-900/60 text-xs font-medium text-zinc-300 mb-1">
            <ShieldCheck className="h-3.5 w-3.5 text-emerald-400" />
            <span>First-Run Platform Initializer</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
            Orbito Initial Setup
          </h1>
          <p className="text-xs text-zinc-400 max-w-sm mx-auto">
            Create the primary Super Administrator account to initialize security policies and default workspace containers.
          </p>
        </div>

        {/* Setup Form Card */}
        <Card className="border-zinc-800 bg-zinc-900/80 shadow-2xl">
          <form onSubmit={handleSubmit}>
            <CardHeader className="pb-3 border-b border-zinc-800/80">
              <CardTitle className="text-sm font-semibold text-zinc-200">
                Super Administrator Credentials
              </CardTitle>
              <CardDescription className="text-xs text-zinc-400">
                This root account possesses unrestricted system-wide authority across all tenants and settings.
              </CardDescription>
            </CardHeader>

            <CardContent className="space-y-4 pt-5">
              {formError && (
                <div className="p-3 rounded-md bg-rose-950/40 border border-rose-800/60 text-rose-300 text-xs flex items-start gap-2">
                  <AlertCircle className="h-4 w-4 shrink-0 mt-0.5" />
                  <span>{formError}</span>
                </div>
              )}

              {/* Name Row */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-zinc-300">First Name *</label>
                  <Input
                    name="firstName"
                    value={formData.firstName}
                    onChange={handleChange}
                    placeholder="e.g. Alex"
                    required
                    className="h-9 text-xs bg-zinc-950 border-zinc-800 text-zinc-100 focus:border-zinc-600"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-zinc-300">Last Name *</label>
                  <Input
                    name="lastName"
                    value={formData.lastName}
                    onChange={handleChange}
                    placeholder="e.g. Mercer"
                    required
                    className="h-9 text-xs bg-zinc-950 border-zinc-800 text-zinc-100 focus:border-zinc-600"
                  />
                </div>
              </div>

              {/* Username & Email */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-zinc-300">Username (Optional)</label>
                  <Input
                    name="username"
                    value={formData.username}
                    onChange={handleChange}
                    placeholder="admin"
                    className="h-9 text-xs bg-zinc-950 border-zinc-800 text-zinc-100 focus:border-zinc-600 font-mono"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-zinc-300">Work Email *</label>
                  <Input
                    name="email"
                    type="email"
                    value={formData.email}
                    onChange={handleChange}
                    placeholder="admin@orbito.io"
                    required
                    className="h-9 text-xs bg-zinc-950 border-zinc-800 text-zinc-100 focus:border-zinc-600"
                  />
                </div>
              </div>

              {/* Password & Confirm */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-zinc-300">Password *</label>
                  <Input
                    name="password"
                    type="password"
                    value={formData.password}
                    onChange={handleChange}
                    placeholder="••••••••••••"
                    required
                    minLength={8}
                    className="h-9 text-xs bg-zinc-950 border-zinc-800 text-zinc-100 focus:border-zinc-600"
                  />
                  <p className="text-[10px] text-zinc-500">Minimum 8 characters</p>
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-zinc-300">Confirm Password *</label>
                  <Input
                    name="confirmPassword"
                    type="password"
                    value={formData.confirmPassword}
                    onChange={handleChange}
                    placeholder="••••••••••••"
                    required
                    minLength={8}
                    className="h-9 text-xs bg-zinc-950 border-zinc-800 text-zinc-100 focus:border-zinc-600"
                  />
                </div>
              </div>
            </CardContent>

            <CardFooter className="pt-2 pb-5 flex flex-col gap-3">
              <Button
                type="submit"
                disabled={isSubmitting}
                className="w-full h-9 text-xs bg-zinc-100 text-zinc-900 hover:bg-zinc-200 font-medium"
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="h-3.5 w-3.5 animate-spin mr-1.5" />
                    Initializing Platform...
                  </>
                ) : (
                  <>
                    Complete Platform Setup <ArrowRight className="h-3.5 w-3.5 ml-1.5" />
                  </>
                )}
              </Button>
              <div className="text-center text-[11px] text-zinc-500">
                Already configured?{' '}
                <Link href="/login" className="text-zinc-300 hover:text-white underline underline-offset-4">
                  Sign in here
                </Link>
              </div>
            </CardFooter>
          </form>
        </Card>
      </div>
    </div>
  );
}

