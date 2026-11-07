'use client';

import * as React from 'react';
import { useState, useMemo, Suspense } from 'react';
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

function ResetPasswordForm() {
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
      <div className="w-full max-w-md relative z-10">
        <AuthHeader
          title="Create new password"
          description="Set a secure password to restore access to your account."
        />

        <Card className="mt-8 border-zinc-800/80 bg-zinc-900/90 shadow-2xl backdrop-blur-md">
          <CardContent className="pt-6 space-y-4">
            {isCheckingToken ? (
              <div className="py-8 flex flex-col items-center justify-center space-y-3">
                <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-zinc-100" />
                <p className="text-xs text-zinc-400">Verifying reset token...</p>
              </div>
            ) : !token || !isTokenValid ? (
              <div className="space-y-4 py-2">
                <Alert variant="destructive">
                  <div className="flex items-center gap-2">
                    <AlertTriangle className="h-4 w-4 shrink-0" />
                    <AlertDescription className="text-xs">
                      This reset link is invalid or has expired. Please request a new one.
                    </AlertDescription>
                  </div>
                </Alert>
                <Link href="/forgot-password" className="block">
                  <Button variant="outline" className="w-full text-xs">
                    Request New Link
                  </Button>
                </Link>
              </div>
            ) : isSuccess ? (
              <div className="py-6 flex flex-col items-center text-center space-y-3">
                <div className="h-12 w-12 rounded-full bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
                  <CheckCircle2 className="h-6 w-6" />
                </div>
                <h3 className="text-sm font-semibold text-zinc-100">Password Changed</h3>
                <p className="text-xs text-zinc-400">
                  Redirecting to sign-in in a moment...
                </p>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-4">
                {errorMessage && (
                  <Alert variant="destructive" className="py-2.5">
                    <AlertDescription className="text-xs">{errorMessage}</AlertDescription>
                  </Alert>
                )}

                <div className="space-y-1.5">
                  <Label htmlFor="password" required>
                    New Password
                  </Label>
                  <Input
                    id="password"
                    type={showPassword ? 'text' : 'password'}
                    placeholder="At least 8 characters"
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
                        className="cursor-pointer text-zinc-500 hover:text-zinc-300"
                        tabIndex={-1}
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

export default function ResetPasswordPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-screen items-center justify-center bg-[#09090b]">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
        </div>
      }
    >
      <ResetPasswordForm />
    </Suspense>
  );
}
