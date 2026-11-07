'use client';

import * as React from 'react';
import { useState } from 'react';
import Link from 'next/link';
import { Mail, ArrowLeft, CheckCircle2 } from 'lucide-react';
import { toast } from 'sonner';

import { useForgotPasswordMutation } from '@/lib/store/api/authApi';
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

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [submittedEmail, setSubmittedEmail] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const [forgotPassword, { isLoading }] = useForgotPasswordMutation();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setErrorMessage(null);

    if (!email) {
      setErrorMessage('Please enter your email address.');
      return;
    }

    try {
      await forgotPassword({ email: email.toLowerCase().trim() }).unwrap();
      setSubmittedEmail(email);
      toast.success('Reset link dispatched');
    } catch (err: unknown) {
      const error = err as { data?: { message?: string | string[] }; status?: number };
      const rawMessage = error.data?.message;
      const message = Array.isArray(rawMessage)
        ? rawMessage.join(', ')
        : rawMessage || 'Could not send reset instructions. Please check your email or try again later.';
      setErrorMessage(message);
    }
  };

  return (
    <div className="min-h-screen flex flex-col justify-center items-center px-4 py-12 sm:px-6 lg:px-8 bg-[#09090b]">
      {/* Subtle background glow */}
      <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[350px] bg-zinc-800/20 blur-[120px] pointer-events-none rounded-full" />

      <div className="w-full max-w-md relative z-10">
        <AuthHeader
          title="Reset your password"
          subtitle="Enter your work email address and we'll send you recovery instructions."
        />

        <Card className="border border-zinc-800/90 bg-zinc-900/90 shadow-2xl backdrop-blur-md">
          <CardContent className="pt-6">
            {submittedEmail ? (
              <div className="text-center py-4 space-y-4">
                <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-950/60 border border-emerald-800/50 text-emerald-400">
                  <CheckCircle2 className="h-6 w-6" />
                </div>
                <div className="space-y-1.5">
                  <h3 className="text-lg font-semibold text-zinc-100">
                    Instructions Sent
                  </h3>
                  <p className="text-xs text-zinc-400 leading-relaxed max-w-xs mx-auto">
                    If an account exists for{' '}
                    <strong className="text-zinc-200">{submittedEmail}</strong>,
                    you will receive a password reset link shortly.
                  </p>
                </div>
                <div className="pt-3">
                  <Button
                    type="button"
                    variant="outline"
                    className="w-full"
                    onClick={() => {
                      setSubmittedEmail(null);
                      setEmail('');
                    }}
                  >
                    Try another email
                  </Button>
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
                  <Label htmlFor="email" required>
                    Work Email
                  </Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="name@company.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    autoComplete="email"
                    autoFocus
                    required
                    disabled={isLoading}
                    leftIcon={<Mail className="h-4 w-4" />}
                  />
                </div>

                <Button
                  type="submit"
                  variant="primary"
                  size="lg"
                  className="w-full font-semibold tracking-tight"
                  isLoading={isLoading}
                >
                  Send Reset Link
                </Button>
              </form>
            )}
          </CardContent>

          <CardFooter className="flex justify-center border-t border-zinc-800/80 py-4 text-xs">
            <Link
              href="/login"
              className="inline-flex items-center gap-1.5 text-zinc-400 hover:text-zinc-100 transition-colors"
            >
              <ArrowLeft className="h-3.5 w-3.5" />
              <span>Back to Sign In</span>
            </Link>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
