'use client';

import * as React from 'react';
import { useState } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { Mail, Lock, Eye, EyeOff, ShieldCheck } from 'lucide-react';
import { toast } from 'sonner';

import { useLoginMutation } from '@/lib/store/api/authApi';
import { useAppDispatch } from '@/lib/store/hooks';
import { setCredentials } from '@/lib/store/slices/authSlice';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import { Card, CardContent, CardFooter } from '@/components/ui/Card';
import { Alert, AlertDescription } from '@/components/ui/Alert';
import { GoogleButton } from '@/components/ui/GoogleButton';
import { AuthHeader } from '@/components/auth/AuthHeader';

export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const returnUrl = searchParams.get('returnUrl') || '/tasks';
  const dispatch = useAppDispatch();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const [login, { isLoading }] = useLoginMutation();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setErrorMessage(null);

    if (!email || !password) {
      setErrorMessage('Please enter both email and password.');
      return;
    }

    try {
      const response = await login({ email, password }).unwrap();
      dispatch(
        setCredentials({
          user: response.user,
          accessToken: response.access_token,
          refreshToken: response.refresh_token,
        })
      );
      toast.success('Signed in successfully');
      router.push(returnUrl);
    } catch (err: unknown) {
      const error = err as { data?: { message?: string | string[] }; status?: number };
      const rawMessage = error.data?.message;
      const message = Array.isArray(rawMessage)
        ? rawMessage.join(', ')
        : rawMessage || 'Invalid email or password. Please verify your credentials.';
      setErrorMessage(message);
    }
  };

  return (
    <div className="min-h-screen flex flex-col justify-center items-center px-4 py-12 sm:px-6 lg:px-8 bg-[#09090b]">
      {/* Subtle radial glow */}
      <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[350px] bg-zinc-800/20 blur-[120px] pointer-events-none rounded-full" />

      <div className="w-full max-w-md relative z-10">
        <AuthHeader
          title="Sign in to your account"
          subtitle="Welcome back. Enter your credentials to access your workspaces."
        />

        <Card className="border border-zinc-800/90 bg-zinc-900/90 shadow-2xl backdrop-blur-md">
          <CardContent className="pt-6 space-y-5">
            {/* Google OAuth Login Button */}
            <GoogleButton label="Continue with Google" disabled={isLoading} />

            {/* Visual Divider */}
            <div className="relative flex items-center py-1">
              <div className="flex-grow border-t border-zinc-800" />
              <span className="flex-shrink mx-3 text-[11px] font-semibold uppercase tracking-wider text-zinc-500">
                Or continue with email
              </span>
              <div className="flex-grow border-t border-zinc-800" />
            </div>

            {/* Error Notification */}
            {errorMessage && (
              <Alert variant="error">
                <AlertDescription>{errorMessage}</AlertDescription>
              </Alert>
            )}

            {/* Login Form */}
            <form onSubmit={handleSubmit} className="space-y-4" noValidate>
              <div className="space-y-1.5">
                <Label htmlFor="email" required>
                  Email address
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="name@work-email.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  autoComplete="email"
                  autoFocus
                  required
                  disabled={isLoading}
                  leftIcon={<Mail className="h-4 w-4" />}
                />
              </div>

              <div className="space-y-1.5">
                <div className="flex items-center justify-between">
                  <Label htmlFor="password" required>
                    Password
                  </Label>
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
                  placeholder="••••••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="current-password"
                  required
                  disabled={isLoading}
                  leftIcon={<Lock className="h-4 w-4" />}
                  rightIcon={
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="text-zinc-400 hover:text-zinc-200 focus:outline-none transition-colors"
                      aria-label={showPassword ? 'Hide password' : 'Show password'}
                    >
                      {showPassword ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </button>
                  }
                />
              </div>

              {/* Security & 7-Day Session Notice */}
              <div className="flex items-center gap-2 py-1 text-xs text-zinc-400">
                <ShieldCheck className="h-4 w-4 text-emerald-400 shrink-0" />
                <span>Encrypted 7-day session with automatic refresh rotation</span>
              </div>

              <Button
                type="submit"
                variant="primary"
                size="lg"
                className="w-full font-semibold tracking-tight"
                isLoading={isLoading}
              >
                Sign In
              </Button>
            </form>
          </CardContent>

          <CardFooter className="flex justify-center border-t border-zinc-800/80 py-4 text-xs text-zinc-400">
            <span>Don&apos;t have an account?</span>
            <Link
              href="/register"
              className="ml-1.5 font-medium text-zinc-100 hover:underline hover:text-white transition-colors"
            >
              Create an account
            </Link>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
