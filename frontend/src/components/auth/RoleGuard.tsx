'use client';

import * as React from 'react';
import Link from 'next/link';
import { ShieldAlert, ArrowLeft } from 'lucide-react';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { Role, type AppAction, type AppResource } from '@/lib/rbac';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';

interface RoleGuardProps {
  children: React.ReactNode;
  requiredRole?: Role;
  permission?: { action: AppAction; resource: AppResource };
  fallback?: React.ReactNode;
}

export function RoleGuard({
  children,
  requiredRole,
  permission,
  fallback,
}: RoleGuardProps) {
  const { user, role, hasRole, can, isLoading } = useAuthRbac();

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
      </div>
    );
  }

  // Check role requirement
  const isRoleAllowed = requiredRole ? hasRole(requiredRole) : true;

  // Check permission requirement
  const isPermissionAllowed = permission
    ? can(permission.action, permission.resource)
    : true;

  const isAllowed = isRoleAllowed && isPermissionAllowed;

  if (!isAllowed) {
    if (fallback) return <>{fallback}</>;

    return (
      <div className="flex min-h-[60vh] items-center justify-center p-6">
        <Card className="max-w-md w-full border border-zinc-800 bg-zinc-900/90 shadow-2xl text-center">
          <CardContent className="pt-8 pb-6 px-6 space-y-4">
            <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-rose-950/60 border border-rose-800/50 text-rose-400">
              <ShieldAlert className="h-7 w-7" />
            </div>

            <div className="space-y-2">
              <h3 className="text-xl font-bold tracking-tight text-zinc-100">
                Access Restricted
              </h3>
              <p className="text-sm text-zinc-400 leading-relaxed">
                Your account role (
                <strong className="text-zinc-200 uppercase">{role}</strong>)
                does not have permission to view or manage this resource.
              </p>
            </div>

            <div className="pt-2">
              <Link href="/">
                <Button variant="outline" className="gap-2">
                  <ArrowLeft className="h-4 w-4" />
                  <span>Return to Workspace</span>
                </Button>
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return <>{children}</>;
}

