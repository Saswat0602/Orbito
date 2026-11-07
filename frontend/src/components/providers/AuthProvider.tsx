'use client';

import * as React from 'react';
import { useEffect } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { setCredentials, logOut } from '@/lib/store/slices/authSlice';
import { useGetMeQuery } from '@/lib/store/api/authApi';

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

  const isPublicRoute = PUBLIC_ROUTES.some((route) =>
    pathname?.startsWith(route)
  );

  // If already have user in store or route is public, skip getMe query
  const shouldSkipGetMe = isPublicRoute || (authState.isAuthenticated && !!authState.user);

  const {
    data: meUser,
    isLoading: isMeLoading,
    isError: isMeError,
  } = useGetMeQuery(undefined, {
    skip: shouldSkipGetMe,
  });

  useEffect(() => {
    if (meUser) {
      dispatch(setCredentials({ user: meUser }));
    }
  }, [meUser, dispatch]);

  useEffect(() => {
    if (!isPublicRoute && isMeError) {
      dispatch(logOut());
      router.replace('/login');
    }
  }, [isPublicRoute, isMeError, dispatch, router]);

  // While checking authentication on protected route
  if (!isPublicRoute && !authState.isAuthenticated && isMeLoading) {
    return (
      <div className="flex h-screen w-full items-center justify-center bg-[#09090b]">
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
          <span className="text-xs text-zinc-400 font-medium">Authenticating session...</span>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}

