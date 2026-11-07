'use client';

import * as React from 'react';
import { useState, useMemo } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Mail, Lock, User, AtSign, Eye, EyeOff, Check, X } from 'lucide-react';
import { toast } from 'sonner';

import { useRegisterMutation } from '@/lib/store/api/authApi';
import { useAppDispatch } from '@/lib/store/hooks';
import { setCredentials } from '@/lib/store/slices/authSlice';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import { Card, CardContent, CardFooter } from '@/components/ui/Card';
import { Alert, AlertDescription } from '@/components/ui/Alert';
import { GoogleButton } from '@/components/ui/GoogleButton';
import { AuthHeader } from '@/components/auth/AuthHeader';

export default function RegisterPage() {
  const router = useRouter();
  const dispatch = useAppDispatch();

  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const [registerUser, { isLoading }] = useRegisterMutation();

  // Real-time password strength calculation
  const passwordChecks = useMemo(() => {
    return {
      length: password.length >= 8,
      hasUpper: /[A-Z]/.test(password),
      hasLower: /[a-z]/.test(password),
      hasNumber: /[0-9]/.test(password),
      hasSpecial: /[^A-Za-z0-9]/.test(password),
    };
  }, [password]);

  const passwordScore = useMemo(() => {
    const passed = Object.values(passwordChecks).filter(Boolean).length;
    return passed;
  }, [passwordChecks]);

  const strengthColor = useMemo(() => {
    if (passwordScore <= 2) return 'bg-red-500';
    if (passwordScore <= 4) return 'bg-amber-400';
    return 'bg-emerald-400';
  }, [passwordScore]);

  const strengthLabel = useMemo(() => {
    if (passwordScore <= 2) return 'Weak';
    if (passwordScore <= 4) return 'Medium';
    return 'Strong';
  }, [passwordScore]);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setErrorMessage(null);

    if (!firstName || !lastName || !username || !email || !password) {
      setErrorMessage('Please fill in all required fields.');
      return;
    }

    if (password.length < 8) {
      setErrorMessage('Password must be at least 8 characters.');
      return;
    }

    try {
      const response = await registerUser({
        firstName,
        lastName,
        username: username.toLowerCase().trim(),
        email: email.toLowerCase().trim(),
        password,
      }).unwrap();

      if (response.user) {
        dispatch(
          setCredentials({
            user: response.user,
            accessToken: response.access_token,
            refreshToken: response.refresh_token,
          })
        );
      }

      toast.success('Account created successfully!');
      router.push('/tasks');
    } catch (err: unknown) {
      const error = err as { data?: { message?: string | string[] }; status?: number };
      const rawMessage = error.data?.message;
      const message = Array.isArray(rawMessage)
        ? rawMessage.join(', ')
        : rawMessage || 'Registration failed. A user with this email or username may already exist.';
      setErrorMessage(message);
    }
  };

  return (
    <div className="min-h-screen flex flex-col justify-center items-center px-4 py-12 sm:px-6 lg:px-8 bg-[#09090b]">
      {/* Subtle background glow */}
      <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[350px] bg-zinc-800/20 blur-[120px] pointer-events-none rounded-full" />

      <div className="w-full max-w-lg relative z-10">
        <AuthHeader
          title="Create your account"
          subtitle="Join Orbito to orchestrate tasks, sprints, and teams with precision."
        />

        <Card className="border border-zinc-800/90 bg-zinc-900/90 shadow-2xl backdrop-blur-md">
          <CardContent className="pt-6 space-y-5">
            {/* Google OAuth Login Button */}
            <GoogleButton label="Sign up with Google" disabled={isLoading} />

            {/* Visual Divider */}
            <div className="relative flex items-center py-1">
              <div className="flex-grow border-t border-zinc-800" />
              <span className="flex-shrink mx-3 text-[11px] font-semibold uppercase tracking-wider text-zinc-500">
                Or register with email
              </span>
              <div className="flex-grow border-t border-zinc-800" />
            </div>

            {/* Error Notification */}
            {errorMessage && (
              <Alert variant="error">
                <AlertDescription>{errorMessage}</AlertDescription>
              </Alert>
            )}

            {/* Registration Form */}
            <form onSubmit={handleSubmit} className="space-y-4" noValidate>
              {/* Name fields row */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label htmlFor="firstName" required>
                    First Name
                  </Label>
                  <Input
                    id="firstName"
                    type="text"
                    placeholder="Alex"
                    value={firstName}
                    onChange={(e) => setFirstName(e.target.value)}
                    required
                    disabled={isLoading}
                    leftIcon={<User className="h-4 w-4" />}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="lastName" required>
                    Last Name
                  </Label>
                  <Input
                    id="lastName"
                    type="text"
                    placeholder="Morgan"
                    value={lastName}
                    onChange={(e) => setLastName(e.target.value)}
                    required
                    disabled={isLoading}
                  />
                </div>
              </div>

              {/* Username field */}
              <div className="space-y-1.5">
                <Label htmlFor="username" required>
                  Username
                </Label>
                <Input
                  id="username"
                  type="text"
                  placeholder="alexm"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  autoComplete="username"
                  required
                  disabled={isLoading}
                  leftIcon={<AtSign className="h-4 w-4" />}
                />
              </div>

              {/* Email field */}
              <div className="space-y-1.5">
                <Label htmlFor="email" required>
                  Work Email
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="alex@company.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  autoComplete="email"
                  required
                  disabled={isLoading}
                  leftIcon={<Mail className="h-4 w-4" />}
                />
              </div>

              {/* Password field */}
              <div className="space-y-1.5">
                <Label htmlFor="password" required>
                  Password
                </Label>
                <Input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  placeholder="Minimum 8 characters"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="new-password"
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

                {/* Password Strength Indicator */}
                {password && (
                  <div className="pt-2 space-y-2">
                    <div className="flex items-center justify-between text-xs">
                      <span className="text-zinc-400">Password strength:</span>
                      <span className="font-semibold text-zinc-200">{strengthLabel}</span>
                    </div>
                    <div className="h-1.5 w-full bg-zinc-800 rounded-full overflow-hidden flex gap-1">
                      <div
                        className={`h-full transition-all duration-300 rounded-full ${strengthColor}`}
                        style={{ width: `${(passwordScore / 5) * 100}%` }}
                      />
                    </div>
                    <div className="grid grid-cols-2 gap-1 text-[11px] text-zinc-400 pt-1">
                      <span className={`flex items-center gap-1 ${passwordChecks.length ? 'text-emerald-400' : 'text-zinc-500'}`}>
                        {passwordChecks.length ? <Check className="h-3 w-3" /> : <X className="h-3 w-3" />} 8+ chars
                      </span>
                      <span className={`flex items-center gap-1 ${passwordChecks.hasUpper && passwordChecks.hasLower ? 'text-emerald-400' : 'text-zinc-500'}`}>
                        {passwordChecks.hasUpper && passwordChecks.hasLower ? <Check className="h-3 w-3" /> : <X className="h-3 w-3" />} Upper & lower
                      </span>
                      <span className={`flex items-center gap-1 ${passwordChecks.hasNumber ? 'text-emerald-400' : 'text-zinc-500'}`}>
                        {passwordChecks.hasNumber ? <Check className="h-3 w-3" /> : <X className="h-3 w-3" />} Number
                      </span>
                      <span className={`flex items-center gap-1 ${passwordChecks.hasSpecial ? 'text-emerald-400' : 'text-zinc-500'}`}>
                        {passwordChecks.hasSpecial ? <Check className="h-3 w-3" /> : <X className="h-3 w-3" />} Symbol
                      </span>
                    </div>
                  </div>
                )}
              </div>

              <Button
                type="submit"
                variant="primary"
                size="lg"
                className="w-full font-semibold tracking-tight mt-2"
                isLoading={isLoading}
              >
                Create Account
              </Button>
            </form>
          </CardContent>

          <CardFooter className="flex justify-center border-t border-zinc-800/80 py-4 text-xs text-zinc-400">
            <span>Already have an account?</span>
            <Link
              href="/login"
              className="ml-1.5 font-medium text-zinc-100 hover:underline hover:text-white transition-colors"
            >
              Sign in
            </Link>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
