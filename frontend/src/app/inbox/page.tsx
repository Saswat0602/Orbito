'use client';

import * as React from 'react';
import { Inbox, Mail, RefreshCw, CheckCircle2, ArrowRight } from 'lucide-react';
import { toast } from 'sonner';

import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import { useAppSelector } from '@/lib/store/hooks';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

export default function InboxPage() {
  const { can } = useAuthRbac();
  const activeProjectId = useAppSelector((state) => state.auth.activeProjectId);
  const [isSyncing, setIsSyncing] = React.useState(false);

  const handleSync = () => {
    setIsSyncing(true);
    setTimeout(() => {
      setIsSyncing(false);
      toast.success('Inbox sync completed. No new messages.');
    }, 1200);
  };

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Inbox className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Project Inbox
            </h1>
          </div>
          <p className="text-xs text-zinc-400">
            Automatically convert inbound customer and stakeholder emails into structured tasks.
          </p>
        </div>

        <Button
          variant="outline"
          size="sm"
          className="gap-2 text-xs"
          isLoading={isSyncing}
          onClick={handleSync}
        >
          <RefreshCw className="h-3.5 w-3.5" />
          <span>Sync Now</span>
        </Button>
      </div>

      {/* Inbox State */}
      <Card className="border-zinc-800 bg-zinc-900/60 p-12 text-center">
        <Mail className="h-10 w-10 text-zinc-500 mx-auto mb-3" />
        <h3 className="text-base font-semibold text-zinc-200">
          Inbox is up to date
        </h3>
        <p className="text-xs text-zinc-400 mt-1 max-w-sm mx-auto">
          Inbound emails sent to your configured project IMAP address will appear here for review and one-click task conversion.
        </p>
      </Card>
    </div>
  );
}

