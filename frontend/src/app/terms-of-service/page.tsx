'use client';

import * as React from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ArrowLeft, Shield, FileText, CheckCircle2, Building2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';

export default function TermsOfServicePage() {
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
            <span className="text-xs font-semibold tracking-tight text-zinc-200">Orbito Legal</span>
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

      {/* Main Legal Content Container */}
      <main className="max-w-4xl mx-auto px-4 sm:px-8 py-10 space-y-8">
        {/* Title */}
        <div className="space-y-2 border-b border-zinc-800/80 pb-6">
          <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded-md border border-zinc-850 bg-zinc-900/60 text-[11px] font-mono text-zinc-400">
            <FileText className="h-3.5 w-3.5 text-zinc-400" />
            <span>TERMS-OF-SERVICE-V2</span>
          </div>
          <h1 className="text-3xl font-bold tracking-tight text-zinc-100">Terms of Service</h1>
          <p className="text-xs text-zinc-400">
            Effective Date: January 1, 2026 &bull; Last Updated: November 12, 2026
          </p>
        </div>

        {/* Legal Sections */}
        <div className="space-y-6 text-xs text-zinc-300 leading-relaxed">
          <Card className="border-zinc-800 bg-zinc-900/50">
            <CardContent className="p-6 space-y-4">
              <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
                <span className="text-zinc-500 font-mono">01.</span> Acceptance of Terms
              </h2>
              <p>
                By creating an account, accessing, or using the Orbito platform (&ldquo;Orbito&rdquo;, &ldquo;we&rdquo;, &ldquo;us&rdquo;, or &ldquo;our&rdquo;), you agree to be bound by these Terms of Service and all incorporated policies. If you are entering into these terms on behalf of an enterprise or organization, you represent that you possess the legal authority to bind such entity.
              </p>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/50">
            <CardContent className="p-6 space-y-4">
              <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
                <span className="text-zinc-500 font-mono">02.</span> Multi-Tenant Workspaces & Platform Architecture
              </h2>
              <p>
                Orbito provides project management, task tracking, sprint planning, analytics, and collaboration tooling. Data is segmented across strict multi-tenant boundaries (Organizations and Workspaces). Tenant administrators are responsible for managing user roles, permissions, invitations, and access controls within their respective scopes.
              </p>
              <ul className="list-disc pl-5 space-y-1 text-zinc-400">
                <li>Super Administrators manage platform infrastructure and tenant provisioning.</li>
                <li>Organization Owners manage billing, organizational boundaries, and workspace assignments.</li>
                <li>Workspace Managers manage sprint lifecycles, project boards, and member participation.</li>
              </ul>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/50">
            <CardContent className="p-6 space-y-4">
              <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
                <span className="text-zinc-500 font-mono">03.</span> User Accounts & Security
              </h2>
              <p>
                You are responsible for maintaining the confidentiality of your credentials and session cookies. You must promptly notify system administrators if you suspect unauthorized access. Orbito employs encrypted JWT dual-token sessions and enterprise role-based access control (RBAC).
              </p>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/50">
            <CardContent className="p-6 space-y-4">
              <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
                <span className="text-zinc-500 font-mono">04.</span> Acceptable Use Policy
              </h2>
              <p>
                You agree not to engage in any activity that interferes with or disrupts the services, servers, or networks connected to Orbito. You must not attempt to circumvent RBAC boundaries, probe internal endpoints, reverse engineer proprietary algorithms, or store unlawful content.
              </p>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/50">
            <CardContent className="p-6 space-y-4">
              <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
                <span className="text-zinc-500 font-mono">05.</span> Intellectual Property & Data Ownership
              </h2>
              <p>
                You retain full ownership of all tasks, descriptions, attachments, and proprietary project materials uploaded to Orbito. Orbito retains all rights, title, and interest in and to the platform codebase, user interfaces, design system, and documentation.
              </p>
            </CardContent>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/50">
            <CardContent className="p-6 space-y-4">
              <h2 className="text-sm font-bold text-zinc-100 flex items-center gap-2">
                <span className="text-zinc-500 font-mono">06.</span> Limitation of Liability & Termination
              </h2>
              <p>
                In no event shall Orbito or its contributors be liable for indirect, incidental, special, consequential, or punitive damages resulting from your use or inability to use the service. System administrators reserve the right to suspend or terminate accounts violating these terms.
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Footer Navigation */}
        <div className="pt-6 border-t border-zinc-800/80 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-zinc-500">
          <div>&copy; {new Date().getFullYear()} Orbito Platform. All rights reserved.</div>
          <div className="flex items-center gap-4">
            <Link href="/privacy-policy" className="hover:text-zinc-300 transition-colors">
              Privacy Policy
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
