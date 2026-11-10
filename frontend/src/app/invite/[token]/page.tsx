'use client';

import * as React from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  Mail,
  Building2,
  Shield,
  CheckCircle2,
  XCircle,
  ArrowRight,
  UserPlus,
  LogIn,
  AlertCircle,
} from 'lucide-react';
import { toast } from 'sonner';

import {
  useVerifyInvitationQuery,
  useAcceptInvitationMutation,
  useDeclineInvitationMutation,
} from '@/lib/store/api/invitationsApi';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { Card, CardContent, CardFooter, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { getRoleBadgeStyle } from '@/lib/rbac';

export default function InviteAcceptancePage() {
  const params = useParams<{ token: string }>();
  const router = useRouter();
  const token = params?.token || '';

  const { isAuthenticated, user } = useAuthRbac();

  const {
    data: invitation,
    isLoading: isVerifying,
    isError: isVerifyError,
  } = useVerifyInvitationQuery(token, {
    skip: !token,
  });

  const [acceptInvitation, { isLoading: isAccepting }] = useAcceptInvitationMutation();
  const [declineInvitation, { isLoading: isDeclining }] = useDeclineInvitationMutation();

  const handleAccept = async () => {
    if (!token) return;
    try {
      const res = await acceptInvitation(token).unwrap();
      toast.success(res.message || 'Invitation accepted successfully!');
      const targetSlug = res.workspace?.slug || invitation?.workspace?.slug;
      if (targetSlug) {
        router.replace(`/${targetSlug}`);
      } else {
        router.replace('/');
      }
    } catch (err: unknown) {
      const msg =
        err && typeof err === 'object' && 'data' in err
          ? (err as { data: { message?: string } }).data?.message
          : 'Failed to accept invitation';
      toast.error(msg || 'Failed to accept invitation');
    }
  };

  const handleDecline = async () => {
    if (!token) return;
    try {
      await declineInvitation(token).unwrap();
      toast.info('Invitation declined');
      router.replace('/');
    } catch {
      toast.error('Failed to decline invitation');
    }
  };

  if (isVerifying) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#09090b] px-4">
        <div className="flex flex-col items-center gap-3">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
          <span className="text-xs text-zinc-400 font-medium">Verifying invitation security token...</span>
        </div>
      </div>
    );
  }

  if (isVerifyError || !invitation || invitation.status !== 'PENDING') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#09090b] px-4">
        <Card className="max-w-md w-full border-zinc-800 bg-zinc-900/90 text-zinc-100 shadow-2xl">
          <CardHeader className="text-center pb-2">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-rose-500/10 text-rose-400 mb-2">
              <XCircle className="h-6 w-6" />
            </div>
            <CardTitle className="text-lg font-bold text-zinc-100">Invalid or Expired Invitation</CardTitle>
            <CardDescription className="text-xs text-zinc-400 pt-1">
              This invitation token is invalid, has already been accepted, or has expired.
            </CardDescription>
          </CardHeader>
          <CardFooter className="pt-4 flex flex-col gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => router.replace('/login')}
              className="w-full text-xs border-zinc-800"
            >
              Go to Sign In
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  const inviterName =
    invitation.inviter?.firstName && invitation.inviter?.lastName
      ? `${invitation.inviter.firstName} ${invitation.inviter.lastName}`
      : invitation.inviter?.username || invitation.inviter?.email || 'A team member';

  const roleStyle = getRoleBadgeStyle(invitation.role);
  const targetName = invitation.workspace?.name || invitation.organization?.name || 'Workspace';

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#09090b] px-4 py-12 relative">
      {/* Background radial gradient */}
      <div className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[300px] bg-zinc-800/20 blur-[100px] pointer-events-none rounded-full" />

      <Card className="max-w-md w-full border-zinc-800 bg-zinc-900/95 text-zinc-100 shadow-2xl relative z-10 backdrop-blur-md">
        <CardHeader className="text-center pb-3">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-400 mb-2">
            <Mail className="h-6 w-6" />
          </div>
          <CardTitle className="text-xl font-bold tracking-tight text-zinc-100">
            You&apos;ve Been Invited!
          </CardTitle>
          <CardDescription className="text-xs text-zinc-400 pt-1">
            <strong className="text-zinc-200">{inviterName}</strong> invited you to collaborate in Orbito.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* Target Entity Box */}
          <div className="p-3.5 rounded-md bg-zinc-950/70 border border-zinc-850 space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded bg-zinc-900 border border-zinc-800 text-zinc-300">
                  <Building2 className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-zinc-200">{targetName}</div>
                  <div className="text-[10px] text-zinc-500 uppercase tracking-wider font-mono">
                    {invitation.workspace ? 'Workspace Scope' : 'Organization Scope'}
                  </div>
                </div>
              </div>

              <span
                className={`inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium border ${roleStyle.className}`}
              >
                {roleStyle.label}
              </span>
            </div>

            <div className="pt-2 border-t border-zinc-900 text-[11px] text-zinc-400 flex items-center justify-between">
              <span>Invited Email:</span>
              <span className="font-mono text-zinc-300">{invitation.inviteeEmail}</span>
            </div>
          </div>

          {/* User state context */}
          {isAuthenticated ? (
            <div className="flex items-center gap-2.5 p-2.5 rounded bg-zinc-950/40 border border-zinc-850 text-xs text-zinc-300">
              <Avatar className="h-6 w-6 rounded-sm border border-zinc-800">
                <AvatarFallback className="bg-zinc-800 text-zinc-200 text-[10px]">
                  {user?.firstName?.[0] || 'U'}
                </AvatarFallback>
              </Avatar>
              <div className="truncate flex-1">
                Signed in as <span className="font-semibold text-zinc-100">{user?.email}</span>
              </div>
            </div>
          ) : (
            <div className="p-3 rounded bg-zinc-950/40 border border-zinc-850 text-xs text-zinc-400 space-y-1">
              <div className="flex items-center gap-1.5 text-zinc-300 font-medium">
                <AlertCircle className="h-3.5 w-3.5 text-amber-400" />
                <span>Account Required</span>
              </div>
              <p className="text-[11px] text-zinc-500 leading-relaxed">
                Sign in with your existing Orbito credentials or register a new account to join this workspace.
              </p>
            </div>
          )}
        </CardContent>

        <CardFooter className="flex flex-col gap-2.5 pt-2">
          {isAuthenticated ? (
            <div className="flex gap-2 w-full">
              <Button
                variant="primary"
                onClick={handleAccept}
                isLoading={isAccepting}
                disabled={isDeclining}
                className="flex-1 text-xs gap-1.5"
              >
                <CheckCircle2 className="h-4 w-4" />
                <span>Accept Invitation</span>
              </Button>
              <Button
                variant="outline"
                onClick={handleDecline}
                isLoading={isDeclining}
                disabled={isAccepting}
                className="text-xs border-zinc-800 text-zinc-400 hover:text-rose-400 hover:border-rose-900"
              >
                Decline
              </Button>
            </div>
          ) : (
            <div className="flex flex-col gap-2 w-full">
              <Link href={`/login?returnUrl=/invite/${token}`} className="w-full">
                <Button variant="primary" className="w-full text-xs gap-1.5">
                  <LogIn className="h-4 w-4" />
                  <span>Sign In to Accept</span>
                </Button>
              </Link>
              <Link href={`/register?inviteToken=${token}&returnUrl=/invite/${token}`} className="w-full">
                <Button variant="outline" className="w-full text-xs border-zinc-800 gap-1.5 text-zinc-300">
                  <UserPlus className="h-4 w-4" />
                  <span>Create Account to Join</span>
                </Button>
              </Link>
            </div>
          )}
        </CardFooter>
      </Card>
    </div>
  );
}
