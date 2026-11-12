'use client';

import * as React from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ArrowLeft, Shield, Lock, Eye, Server, Database, CheckCircle2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';

export default function PrivacyPolicyPage() {
  const router = useRouter();

  return (
    <div className="min-h-screen bg-[#09090b] text-zinc-100 selection:bg-zinc-800">
      {/* Top Header */}
      <header className="sticky top-0 z-20 bg-[#09090b]/80 backdrop-blur-md border-b border-zinc-800/80 px-4 sm:px-8 py-3">
        <div className="max-w-4xl mx-auto flex items-center justify-between">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => router.back()}
            className="text-zinc-400 hover:text-zinc-100 hover:bg-zinc-850 h-8 gap-1.5 text-xs"
          >
            <ArrowLeft className="h-3.5 w-3.5" /> Back
          </Button>

          <div className="flex items-center gap-2">
            <div className="h-5 w-5 rounded-sm bg-zinc-100 text-zinc-900 flex items-center justify-center font-bold text-xs">
              O
            </div>
            <span className="text-xs font-semibold tracking-tight text-zinc-200">Orbito Privacy</span>
          </div>

          <Button
            asChild
            variant="outline"
            size="sm"
            className="h-8 text-xs border-zinc-800 text-zinc-300 hover:text-white"
          >
            <Link href="/login">Sign In</Link>
          </Button>
        </div>
      </header>

      {/* Main Privacy Content Container */}
      <main className="max-w-4xl mx-auto px-4 sm:px-8 py-10 space-y-8">
        {/* Title */}
        <div className="space-y-2 border-b border-zinc-800/80 pb-6">
          <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded-md border border-zinc-850 bg-zinc-900/60 text-[11px] font-mono text-zinc-400">
            <Shield className="h-3.5 w-3.5 text-emerald-400" />
            <span>PRIVACY-POLICY-V2</span>
          </div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-100">Privacy Policy</h1>
          <p className="text-xs text-zinc-400">
            Effective Date: January 1, 2026 &bull; Last Updated: November 12, 2026
          </p>
        </div>

        {/* Privacy Sections */}
        <div className="space-y-6 text-xs text-zinc-300 leading-relaxed">
          <Card className="border-zinc-800 bg-zinc-900/50">
            <CardContent className="p-6 space-y-4">
              <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
                <span className="text-zinc-500 font-mono">01.</span> Overview & Scope
              </h2>
              <p>
                Orbito is committed to safeguarding the privacy and security of enterprise data. This Privacy Policy details how we collect, process, isolate, and protect information when you utilize our platform services, APIs, and web clients.
              </p>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/50">
            <CardContent className="p-6 space-y-4">
              <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
                <span className="text-zinc-500 font-mono">02.</span> Information We Collect
              </h2>
              <p>We collect and process the following categories of information:</p>
              <ul className="list-disc pl-5 space-y-1.5 text-zinc-400">
                <li>
                  <strong className="text-zinc-200">Account Credentials:</strong> Name, work email address, username, hashed password, and optional avatar image.
                </li>
                <li>
                  <strong className="text-zinc-200">Work Management Data:</strong> Projects, tasks, sprint plans, subtasks, task comments, attachments, time tracking entries, and activity logs.
                </li>
                <li>
                  <strong className="text-zinc-200">System Logs & Telemetry:</strong> Anonymized server logs, audit timestamps, and session security tokens required for authentication integrity.
                </li>
              </ul>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/50">
            <CardContent className="p-6 space-y-4">
              <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
                <span className="text-zinc-500 font-mono">03.</span> Multi-Tenant Data Isolation
              </h2>
              <p>
                Orbito uses relational scoping and row-level authorization barriers. All organizations, workspaces, and projects are strictly isolated by foreign key constraints and NestJS middleware guards. No tenant can view, query, or mutate another tenant&rsquo;s entities without explicit permissions granted by a Super Administrator or Organization Owner.
              </p>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/50">
            <CardContent className="p-6 space-y-4">
              <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
                <span className="text-zinc-500 font-mono">04.</span> Authentication & Cookie Security
              </h2>
              <p>
                Orbito issues cryptographically signed JWT tokens stored strictly in <code className="px-1 py-0.5 rounded bg-zinc-950 border border-zinc-800 text-zinc-300">httpOnly</code>, <code className="px-1 py-0.5 rounded bg-zinc-950 border border-zinc-800 text-zinc-300">Secure</code>, <code className="px-1 py-0.5 rounded bg-zinc-950 border border-zinc-800 text-zinc-300">SameSite=Lax</code> cookies. This eliminates client-side XSS token theft vectors. Sessions are periodically re-validated against PostgreSQL and Redis stores.
              </p>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/50">
            <CardContent className="p-6 space-y-4">
              <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
                <span className="text-zinc-500 font-mono">05.</span> Data Retention & Right to Be Forgotten
              </h2>
              <p>
                Organizations retain full authority to request data deletion. Super Administrators can suspend or permanently purge user records and tenant databases. Upon verified deletion, all user tokens, personal identifiers, and associated assignments are sanitized from the active persistence cluster.
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Footer Navigation */}
        <div className="pt-6 border-t border-zinc-800/80 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-zinc-500">
          <div>&copy; {new Date().getFullYear()} Orbito Platform. All rights reserved.</div>
          <div className="flex items-center gap-4">
            <Link href="/terms-of-service" className="hover:text-zinc-300 transition-colors">
              Terms of Service
            </Link>
            <span>&bull;</span>
            <Link href="/login" className="hover:text-zinc-300 transition-colors">
              Sign In
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}
