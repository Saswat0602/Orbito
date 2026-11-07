'use client';

import * as React from 'react';
import { useState, useEffect, Suspense } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { Mail, Lock, Eye, EyeOff, ShieldCheck, Loader2 } from 'lucide-react';
import { toast } from 'sonner';

import { useLoginMutation, useExchangeSsoMutation } from '@/lib/store/api/authApi';
import { useAppDispatch } from '@/lib/store/hooks';
import { setCredentials } from '@/lib/store/slices/authSlice';
import {
  Button,
  Input,
  Label,
  Card,
  CardContent,
  CardFooter,
  Alert,
  AlertDescription,
  GoogleButton,
} from '@/components/ui';
import {
  AuthHeader,
  DevSeededAccounts,
  type SeededAccount,
} from '@/components/auth';

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const returnUrl = searchParams.get('returnUrl') || '/';
  const dispatch = useAppDispatch();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const [login, { isLoading }] = useLoginMutation();
  const [exchangeSso, { isLoading: isExchangingSso }] = useExchangeSsoMutation();

  // Handle SSO / OIDC callback exchange
  const ssoParam = searchParams.get('sso');
  const errorParam = searchParams.get('error');
  const messageParam = searchParams.get('message');

  useEffect(() => {
    if (errorParam) {
      setErrorMessage(
        messageParam || 'SSO authentication failed. Please try again or use email.'
      );
    } else if (ssoParam === 'callback') {
      const handleSsoExchange = async () => {
        try {
          const response = await exchangeSso().unwrap();
          dispatch(
            setCredentials({
              user: response.user,
            })
          );
          toast.success(
            `Welcome to Orbito${response.user.firstName ? `, ${response.user.firstName}` : ''}!`
          );
          router.replace(returnUrl);
        } catch {
          setErrorMessage('Failed to complete SSO authentication. Please try again.');
          toast.error('SSO exchange failed');
        }
      };
      void handleSsoExchange();
    }
  }, [ssoParam, errorParam, messageParam, exchangeSso, dispatch, router, returnUrl]);

  const executeLogin = async (targetEmail: string, targetPassword: string) => {
    setErrorMessage(null);

    try {
      const response = await login({
        email: targetEmail,
        password: targetPassword,
      }).unwrap();

      dispatch(
        setCredentials({
          user: response.user,
        })
      );
      toast.success(
        `Welcome back to Orbito${response.user.firstName ? `, ${response.user.firstName}` : ''}!`
      );
      router.push(returnUrl);
    } catch (err: unknown) {
      const error = err as { data?: { message?: string } };
      const message =
        error?.data?.message || 'Invalid email or password. Please try again.';
      setErrorMessage(message);
      toast.error('Authentication failed');
    }
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setErrorMessage(null);

    if (!email || !password) {
      setErrorMessage('Please enter both email and password.');
      return;
    }

    await executeLogin(email, password);
  };

  const handleSelectAccount = (account: SeededAccount, autoSubmit = false) => {
    setEmail(account.email);
    setPassword(account.password);
    setErrorMessage(null);

    if (autoSubmit) {
      void executeLogin(account.email, account.password);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center p-4 py-8 bg-[#09090b]">
      <Card className="w-full max-w-md border-zinc-800/80 bg-zinc-900/90 shadow-2xl backdrop-blur-md">
        <AuthHeader
          title="Sign in to Orbito"
          description="Enter your credentials to access your workspace."
        />

        <CardContent className="space-y-4 pt-4">
          {isExchangingSso ? (
            <div className="flex flex-col items-center justify-center p-6 space-y-3 bg-zinc-950/60 rounded-md border border-zinc-800">
              <Loader2 className="h-6 w-6 animate-spin text-zinc-300" />
              <p className="text-xs text-zinc-400">Completing SSO authentication...</p>
            </div>
          ) : (
            <GoogleButton text="Continue with Google / SSO" />
          )}

          <div className="relative flex items-center justify-center py-2">
            <div className="w-full border-t border-zinc-800" />
            <span className="absolute bg-zinc-900 px-3 text-[11px] uppercase tracking-wider text-zinc-500 font-semibold">
              Or with email
            </span>
          </div>

          {errorMessage && (
            <Alert variant="destructive" className="py-2.5">
              <AlertDescription className="text-xs">
                {errorMessage}
              </AlertDescription>
            </Alert>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="email">Work Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="name@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
                leftIcon={<Mail className="h-4 w-4 text-zinc-500" />}
                disabled={isLoading}
              />
            </div>

            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <Label htmlFor="password">Password</Label>
                <Link
                  href="/forgot-password"
                  className="text-xs text-zinc-400 hover:text-zinc-200 transition-colors"
                >
                  Forgot password?
                </Link>
              </div>
              <Input
                id="password"
                type={showPassword ? 'text' : 'password'}
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete="current-password"
                leftIcon={<Lock className="h-4 w-4 text-zinc-500" />}
                rightIcon={
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="cursor-pointer text-zinc-500 hover:text-zinc-300"
                    tabIndex={-1}
                    aria-label={
                      showPassword ? 'Hide password' : 'Show password'
                    }
                  >
                    {showPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </button>
                }
                disabled={isLoading}
              />
            </div>

            {/* High-contrast, bold Sign In Button */}
            <Button
              type="submit"
              variant="primary"
              className="w-full h-10 font-semibold text-zinc-950 bg-zinc-100 hover:bg-zinc-200 transition-colors shadow-sm cursor-pointer"
              isLoading={isLoading}
            >
              Sign In
            </Button>
          </form>

          {/* Seeded Accounts directly below the Sign In button */}
          <DevSeededAccounts
            onSelectAccount={handleSelectAccount}
            selectedEmail={email}
            isLoggingIn={isLoading}
          />

          {/* 7-Day Session Notice */}
          <div className="flex items-center gap-2 pt-2 px-1 text-[11px] text-zinc-500">
            <ShieldCheck className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
            <span>
              Encrypted cookie session valid for up to 7 days with RTR.
            </span>
          </div>
        </CardContent>

        <CardFooter className="justify-center border-t border-zinc-800/60 py-4">
          <p className="text-xs text-zinc-400">
            Don&apos;t have an account?{' '}
            <Link
              href="/register"
              className="font-semibold text-zinc-200 hover:underline hover:text-white transition-colors"
            >
              Create an account
            </Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-screen items-center justify-center bg-[#09090b]">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
        </div>
      }
    >
      <LoginForm />
    </Suspense>
  );
}
