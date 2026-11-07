'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { Lock, Eye, EyeOff, CheckCircle2, AlertTriangle } from 'lucide-react';
import { toast } from 'sonner';

import {
  useResetPasswordMutation,
  useVerifyResetTokenQuery,
} from '@/lib/store/api/authApi';
import {
  Button,
  Input,
  Label,
  Card,
  CardContent,
  CardFooter,
  Alert,
  AlertDescription,
} from '@/components/ui';
import { AuthHeader } from '@/components/auth/AuthHeader';

export default function ResetPasswordPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get('token') || '';

  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Check token validity
  const { data: tokenCheck, isLoading: isCheckingToken } = useVerifyResetTokenQuery(
    token,
    { skip: !token }
  );

  const [resetPassword, { isLoading: isResetting }] = useResetPasswordMutation();

  const isTokenValid = useMemo(() => {
    if (!token) return false;
    return tokenCheck?.valid ?? true;
  }, [token, tokenCheck]);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setErrorMessage(null);

    if (!password || !confirmPassword) {
      setErrorMessage('Please enter and confirm your new password.');
      return;
    }

    if (password.length < 8) {
      setErrorMessage('Password must be at least 8 characters long.');
      return;
    }

    if (password !== confirmPassword) {
      setErrorMessage('Passwords do not match.');
      return;
    }

    try {
      await resetPassword({
        token,
        password,
        confirmPassword,
      }).unwrap();

      setIsSuccess(true);
      toast.success('Password updated successfully');
      setTimeout(() => {
        router.push('/login');
      }, 2000);
    } catch (err: unknown) {
      const error = err as { data?: { message?: string | string[] }; status?: number };
      const rawMessage = error.data?.message;
      const message = Array.isArray(rawMessage)
        ? rawMessage.join(', ')
        : rawMessage || 'Failed to reset password. The reset link may have expired.';
      setErrorMessage(message);
    }
  };

  return (
    <div className="min-h-screen flex flex-col justify-center items-center px-4 py-12 sm:px-6 lg:px-8 bg-[#09090b]">
      {/* Subtle background glow */}
      <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[350px] bg-zinc-800/20 blur-[120px] pointer-events-none rounded-full" />

      <div className="w-full max-w-md relative z-10">
        <AuthHeader
          title="Create new password"
          subtitle="Set a secure password to restore access to your account."
        />

        <Card className="border border-zinc-800/90 bg-zinc-900/90 shadow-2xl backdrop-blur-md">
          <CardContent className="pt-6">
            {!token ? (
              <div className="text-center py-4 space-y-4">
                <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-rose-950/60 border border-rose-800/50 text-rose-400">
                  <AlertTriangle className="h-6 w-6" />
                </div>
                <div className="space-y-1">
                  <h3 className="text-lg font-semibold text-zinc-100">
                    Missing Reset Token
                  </h3>
                  <p className="text-xs text-zinc-400 leading-relaxed">
                    This password reset link is invalid or incomplete.
                  </p>
                </div>
                <div className="pt-2">
                  <Link href="/forgot-password">
                    <Button variant="primary" className="w-full">
                      Request a new reset link
                    </Button>
                  </Link>
                </div>
              </div>
            ) : isCheckingToken ? (
              <div className="text-center py-8 text-xs text-zinc-400">
                Verifying reset link authenticity...
              </div>
            ) : !isTokenValid ? (
              <div className="text-center py-4 space-y-4">
                <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-rose-950/60 border border-rose-800/50 text-rose-400">
                  <AlertTriangle className="h-6 w-6" />
                </div>
                <div className="space-y-1">
                  <h3 className="text-lg font-semibold text-zinc-100">
                    Link Expired or Invalid
                  </h3>
                  <p className="text-xs text-zinc-400 leading-relaxed">
                    {tokenCheck?.message || 'This password reset link has expired.'}
                  </p>
                </div>
                <div className="pt-2">
                  <Link href="/forgot-password">
                    <Button variant="primary" className="w-full">
                      Request a new link
                    </Button>
                  </Link>
                </div>
              </div>
            ) : isSuccess ? (
              <div className="text-center py-4 space-y-4">
                <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-950/60 border border-emerald-800/50 text-emerald-400">
                  <CheckCircle2 className="h-6 w-6" />
                </div>
                <div className="space-y-1">
                  <h3 className="text-lg font-semibold text-zinc-100">
                    Password Reset Complete
                  </h3>
                  <p className="text-xs text-zinc-400 leading-relaxed">
                    Your password has been securely updated. Redirecting to sign in...
                  </p>
                </div>
                <div className="pt-2">
                  <Link href="/login">
                    <Button variant="primary" className="w-full">
                      Sign in now
                    </Button>
                  </Link>
                </div>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-4" noValidate>
                {errorMessage && (
                  <Alert variant="destructive">
                    <AlertDescription>{errorMessage}</AlertDescription>
                  </Alert>
                )}

                <div className="space-y-1.5">
                  <Label htmlFor="password" required>
                    New Password
                  </Label>
                  <Input
                    id="password"
                    type={showPassword ? 'text' : 'password'}
                    placeholder="Minimum 8 characters"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    autoComplete="new-password"
                    required
                    disabled={isResetting}
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

                <div className="space-y-1.5">
                  <Label htmlFor="confirmPassword" required>
                    Confirm Password
                  </Label>
                  <Input
                    id="confirmPassword"
                    type={showPassword ? 'text' : 'password'}
                    placeholder="Re-enter your password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    autoComplete="new-password"
                    required
                    disabled={isResetting}
                    leftIcon={<Lock className="h-4 w-4" />}
                  />
                </div>

                <Button
                  type="submit"
                  variant="primary"
                  size="lg"
                  className="w-full font-semibold tracking-tight"
                  isLoading={isResetting}
                >
                  Save New Password
                </Button>
              </form>
            )}
          </CardContent>

          <CardFooter className="flex justify-center border-t border-zinc-800/80 py-4 text-xs">
            <Link
              href="/login"
              className="text-zinc-400 hover:text-zinc-100 transition-colors"
            >
              Back to Sign In
            </Link>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
