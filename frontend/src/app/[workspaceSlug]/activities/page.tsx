'use client';

import * as React from 'react';
import { useParams } from 'next/navigation';
import { Activity, Clock } from 'lucide-react';
import { useGetActivityLogsQuery } from '@/lib/store/api/activityLogApi';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

export default function WorkspaceActivitiesPage() {
  const params = useParams<{ workspaceSlug: string }>();
  const { workspaceSlug } = params;

  const { data: logs = [], isLoading } = useGetActivityLogsQuery({ limit: 50 });

  return (
    <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-8 bg-[#09090b]">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5">
            <Activity className="h-6 w-6 text-zinc-100" />
            <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
              Workspace Activity
            </h1>
            <Badge variant="outline" className="text-xs font-mono">
              {logs.length} {logs.length === 1 ? 'Event' : 'Events'}
            </Badge>
          </div>
          <p className="text-xs text-zinc-400">
            Real-time audit log of team actions in <span className="font-mono text-zinc-300">{workspaceSlug}</span>.
          </p>
        </div>
      </div>

      {isLoading ? (
        <div className="flex h-48 items-center justify-center">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
        </div>
      ) : logs.length === 0 ? (
        <Card className="border-zinc-800 bg-zinc-900/60 p-12 text-center">
          <Activity className="h-10 w-10 text-zinc-500 mx-auto mb-3" />
          <h3 className="text-base font-semibold text-zinc-200">
            No activity logged yet
          </h3>
          <p className="text-xs text-zinc-400 mt-1 max-w-sm mx-auto">
            Team actions will automatically generate activity entries here.
          </p>
        </Card>
      ) : (
        <Card className="border-zinc-800 bg-zinc-900/80 divide-y divide-zinc-850">
          {logs.map((log) => (
            <div key={log.id} className="p-4 flex items-start gap-3 hover:bg-zinc-850/40 transition-colors">
              <Avatar className="h-8 w-8 mt-0.5 rounded-md border border-zinc-800">
                <AvatarFallback className="bg-zinc-800 text-zinc-300 text-xs font-bold">
                  {log.user?.firstName?.[0] || 'U'}
                </AvatarFallback>
              </Avatar>

              <div className="flex-1 min-w-0 space-y-1">
                <div className="flex items-center justify-between gap-2">
                  <span className="text-xs font-semibold text-zinc-200">
                    {log.user?.firstName
                      ? `${log.user.firstName} ${log.user.lastName || ''}`
                      : 'Team Member'}
                  </span>
                  <span className="text-[11px] text-zinc-500 flex items-center gap-1 shrink-0">
                    <Clock className="h-3 w-3" />
                    {new Date(log.createdAt).toLocaleTimeString([], {
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </span>
                </div>
                <p className="text-xs text-zinc-400 leading-relaxed">
                  {log.description}
                </p>
                <div className="pt-1">
                  <Badge variant="outline" className="text-[9px] uppercase">
                    {log.type}
                  </Badge>
                </div>
              </div>
            </div>
          ))}
        </Card>
      )}
    </div>
  );
}

