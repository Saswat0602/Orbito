'use client';

import * as React from 'react';
import { useEffect, useRef } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import { AlertCircle } from 'lucide-react';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { setCredentials, logOut } from '@/lib/store/slices/authSlice';
import { useGetMeQuery } from '@/lib/store/api/authApi';
import { Button } from '@/components/ui/button';

const PUBLIC_ROUTES = [
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
  '/public',
];

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const dispatch = useAppDispatch();
  const authState = useAppSelector((state) => state.auth);
  const hasRedirectedRef = useRef(false);

  const isPublicRoute = PUBLIC_ROUTES.some((route) =>
    pathname?.startsWith(route)
  );

  // Skip getMe query on public routes or when user is already hydrated
  const shouldSkipGetMe = isPublicRoute || (authState.isAuthenticated && !!authState.user);

  const {
    data: meUser,
    isLoading: isMeLoading,
    isError: isMeError,
    error: meError,
  } = useGetMeQuery(undefined, {
    skip: shouldSkipGetMe,
  });

  useEffect(() => {
    if (meUser) {
      dispatch(setCredentials({ user: meUser }));
      hasRedirectedRef.current = false;
    }
  }, [meUser, dispatch]);

  useEffect(() => {
    if (!isPublicRoute && isMeError && !hasRedirectedRef.current) {
      hasRedirectedRef.current = true;
      dispatch(logOut());
      router.replace('/login?error=session_expired');
    }
  }, [isPublicRoute, isMeError, dispatch, router]);

  // While checking authentication on protected route
  if (!isPublicRoute && !authState.isAuthenticated) {
    if (isMeError) {
      return (
        <div className="flex h-screen w-full items-center justify-center bg-[#09090b] p-6">
          <div className="max-w-sm w-full text-center space-y-4 p-6 rounded-md border border-zinc-800 bg-zinc-900/90 shadow-xl">
            <div className="mx-auto flex h-10 w-10 items-center justify-center rounded-full bg-rose-500/10 text-rose-400">
              <AlertCircle className="h-5 w-5" />
            </div>
            <h3 className="text-sm font-semibold text-zinc-100">Authentication Required</h3>
            <p className="text-xs text-zinc-400 leading-relaxed">
              Unable to authenticate your session or the API server is unreachable.
            </p>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                hasRedirectedRef.current = false;
                router.replace('/login?error=session_expired');
              }}
              className="w-full text-xs"
            >
              Sign In Again
            </Button>
          </div>
        </div>
      );
    }

    if (isMeLoading) {
      return (
        <div className="flex h-screen w-full items-center justify-center bg-[#09090b]">
          <div className="flex flex-col items-center gap-3">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
            <span className="text-xs text-zinc-400 font-medium">Authenticating session...</span>
          </div>
        </div>
      );
    }
  }

  return <>{children}</>;
}

