'use client';

import { useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { logOut, setCredentials } from '@/lib/store/slices/authSlice';
import { useGetMeQuery, useLogoutMutation } from '@/lib/store/api/authApi';
import {
  Role,
  hasRole,
  can,
  getRoleBadgeStyle,
  type AppAction,
  type AppResource,
} from '@/lib/rbac';

export function useAuthRbac() {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const authState = useAppSelector((state) => state.auth);

  const {
    data: meUser,
    isLoading: isMeLoading,
    isError: isMeError,
  } = useGetMeQuery(undefined, {
    skip: !!authState.user,
  });

  const [logoutMutation, { isLoading: isLoggingOut }] = useLogoutMutation();

  const user = authState.user || meUser || null;
  const isAuthenticated = authState.isAuthenticated || !!meUser;
  const userRole = (user?.role as Role) || Role.MEMBER;

  const permissions = useMemo(() => {
    return {
      isSuperAdmin: userRole === Role.SUPER_ADMIN,
      isOwner: userRole === Role.OWNER || userRole === Role.SUPER_ADMIN,
      isManager:
        userRole === Role.MANAGER ||
        userRole === Role.OWNER ||
        userRole === Role.SUPER_ADMIN,
      isMember:
        userRole === Role.MEMBER ||
        userRole === Role.MANAGER ||
        userRole === Role.OWNER ||
        userRole === Role.SUPER_ADMIN,
      isViewer: userRole === Role.VIEWER,
      can: (action: AppAction, resource: AppResource) =>
        can(userRole, action, resource),
      hasRole: (requiredRole: Role) => hasRole(userRole, requiredRole),
      roleBadge: getRoleBadgeStyle(userRole),
    };
  }, [userRole]);

  const handleLogout = async () => {
    try {
      await logoutMutation().unwrap();
    } catch {
      // Continue even if network error on logout
    } finally {
      dispatch(logOut());
      toast.info('Signed out of session');
      router.push('/login');
    }
  };

  return {
    user,
    role: userRole,
    isAuthenticated,
    isLoading: isMeLoading || isLoggingOut,
    isError: isMeError,
    ...permissions,
    logout: handleLogout,
  };
}

