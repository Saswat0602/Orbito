'use client';

import * as React from 'react';
import {
  Cpu,
  Activity,
  CheckCircle2,
  AlertTriangle,
  ExternalLink,
  Layers,
  RefreshCw,
  Server,
  Zap,
} from 'lucide-react';

import { Role } from '@orbito/types';
import { RoleGuard } from '@/components/auth/RoleGuard';
import { useGetQueueHealthQuery } from '@/lib/store/api/adminApi';
import { API_BASE_URL } from '@/lib/api/endpoint';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

export default function AdminQueuesPage() {
  const { data: queueHealth, isLoading, refetch, isFetching } = useGetQueueHealthQuery(undefined, {
    pollingInterval: 10000, // Auto refresh health every 10s
  });

  const bullBoardUrl = `${API_BASE_URL}/queues`;

  const stats = queueHealth?.stats || {
    totalJobs: 0,
    completedJobs: 0,
    failedJobs: 0,
    activeJobs: 0,
    successRate: 100,
  };

  return (
    <RoleGuard requiredRole={Role.SUPER_ADMIN}>
      <div className="flex-1 overflow-y-auto p-6 md:p-8 space-y-6 bg-[#09090b]">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-zinc-800/80">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Cpu className="h-6 w-6 text-zinc-100" />
              <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
                Queue & Background Workers
              </h1>
              <Badge variant="outline" className="text-[10px] border-zinc-800 text-zinc-400">
                BullMQ Pipeline
              </Badge>
            </div>
            <p className="text-xs text-zinc-400">
              Live telemetry on asynchronous task queues, Redis broker health, and worker job throughput.
            </p>
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => refetch()}
              disabled={isFetching}
              className="h-8 text-xs border-zinc-800 text-zinc-300"
            >
              <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${isFetching ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
            <a
              href={bullBoardUrl}
              target="_blank"
              rel="noopener noreferrer"
            >
              <Button size="sm" className="h-8 text-xs bg-zinc-100 text-zinc-950 font-semibold hover:bg-white cursor-pointer">
                <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
                Open BullBoard UI
              </Button>
            </a>
          </div>
        </div>

        {/* Telemetry Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Broker Backend
              </span>
              <Server className="h-4 w-4 text-zinc-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-100 uppercase">
              {queueHealth?.backend || 'BullMQ'}
            </div>
            <div className="flex items-center gap-1.5 text-[11px] mt-1">
              <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
              <span className="text-emerald-400">
                {queueHealth?.redisAvailable ? 'Redis 7.0 Connected' : 'In-Memory Fallback'}
              </span>
            </div>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Total Processed
              </span>
              <Zap className="h-4 w-4 text-amber-400" />
            </div>
            <div className="text-2xl font-bold text-zinc-100">
              {stats.totalJobs}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">
              {stats.activeJobs} jobs active now
            </p>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Completed Jobs
              </span>
              <CheckCircle2 className="h-4 w-4 text-emerald-400" />
            </div>
            <div className="text-2xl font-bold text-emerald-400">
              {stats.completedJobs}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">
              Success rate: {Math.round(stats.successRate || 100)}%
            </p>
          </Card>

          <Card className="border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-2">
              <span className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                Failed Jobs
              </span>
              <AlertTriangle className="h-4 w-4 text-rose-400" />
            </div>
            <div className="text-2xl font-bold text-rose-400">
              {stats.failedJobs}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">
              Dead letter queue monitoring
            </p>
          </Card>
        </div>

        {/* Registered Queues Table */}
        <Card className="border-zinc-800 bg-zinc-900/60">
          <CardHeader>
            <CardTitle className="text-base text-zinc-100">Active Queue Pipelines</CardTitle>
            <CardDescription className="text-xs text-zinc-400">
              Isolated job channels registered under the BullMQ orchestration service.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {(!queueHealth?.queues || queueHealth.queues.length === 0) ? (
              <div className="text-center py-8 text-xs text-zinc-500">
                No active queues detected. Queues will initialize automatically on first dispatch.
              </div>
            ) : (
              <div className="divide-y divide-zinc-800/80">
                {queueHealth.queues.map((qName) => (
                  <div
                    key={qName}
                    className="py-3 px-2 flex items-center justify-between hover:bg-zinc-850/40 rounded-md transition-colors"
                  >
                    <div className="flex items-center gap-2.5">
                      <div className="h-2 w-2 rounded-full bg-emerald-400" />
                      <span className="text-xs font-mono font-medium text-zinc-200">{qName}</span>
                    </div>

                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="text-[10px] border-zinc-800 text-zinc-400">
                        BullMQ Worker: Running
                      </Badge>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Direct BullBoard Dashboard Banner */}
        <Card className="border-zinc-800 bg-gradient-to-r from-zinc-900 to-zinc-950 p-6 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="space-y-1">
            <h3 className="text-sm font-semibold text-zinc-100">
              Deep Queue Inspection & Job Retries
            </h3>
            <p className="text-xs text-zinc-400">
              Use the embedded BullBoard dashboard to retry failed jobs, view payloads, purge queues, and inspect worker concurrency.
            </p>
          </div>
          <a
            href={bullBoardUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="shrink-0"
          >
            <Button variant="outline" size="sm" className="h-9 text-xs border-zinc-700 bg-zinc-800 text-zinc-200 hover:bg-zinc-700">
              Launch BullBoard <ExternalLink className="h-3.5 w-3.5 ml-1.5" />
            </Button>
          </a>
        </Card>
      </div>
    </RoleGuard>
  );
}

