'use client';

import * as React from 'react';
import { Target, Calendar, Play, CheckCircle2 } from 'lucide-react';
import type { Sprint } from '@orbito/types';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';

interface SprintCardProps {
  sprint: Sprint;
  onStartSprint?: (sprint: Sprint) => void;
  onCompleteSprint?: (sprint: Sprint) => void;
}

export function SprintCard({
  sprint,
  onStartSprint,
  onCompleteSprint,
}: SprintCardProps) {
  const { isManager, isOwner } = useAuthRbac();
  const canManage = isManager || isOwner;

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'ACTIVE':
        return (
          <Badge className="bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 text-[10px] uppercase font-mono">
            Active
          </Badge>
        );
      case 'COMPLETED':
        return (
          <Badge className="bg-zinc-800 text-zinc-400 border border-zinc-700 text-[10px] uppercase font-mono">
            Closed
          </Badge>
        );
      case 'PLANNING':
      default:
        return (
          <Badge className="bg-sky-500/10 text-sky-400 border border-sky-500/30 text-[10px] uppercase font-mono">
            Planning
          </Badge>
        );
    }
  };

  return (
    <Card className="border-zinc-800 bg-zinc-900/80 p-5 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2.5">
          <div className="h-7 w-7 rounded-md bg-zinc-800 flex items-center justify-center text-zinc-300">
            <Target className="h-4 w-4" />
          </div>
          <div>
            <h4 className="text-sm font-bold text-zinc-100">{sprint.name}</h4>
            {sprint.goal && (
              <p className="text-xs text-zinc-400 line-clamp-1">{sprint.goal}</p>
            )}
          </div>
        </div>

        <div className="flex items-center gap-2">
          {getStatusBadge(sprint.status)}

          {canManage && sprint.status === 'PLANNING' && onStartSprint && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => onStartSprint(sprint)}
              className="gap-1.5 text-xs h-7 px-2.5"
            >
              <Play className="h-3 w-3 text-emerald-400" />
              <span>Start</span>
            </Button>
          )}

          {canManage && sprint.status === 'ACTIVE' && onCompleteSprint && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => onCompleteSprint(sprint)}
              className="gap-1.5 text-xs h-7 px-2.5 border-emerald-500/40 text-emerald-300 hover:bg-emerald-950/40"
            >
              <CheckCircle2 className="h-3 w-3" />
              <span>Complete</span>
            </Button>
          )}
        </div>
      </div>

      <div className="flex items-center gap-4 text-xs text-zinc-500 pt-2 border-t border-zinc-850">
        <div className="flex items-center gap-1.5">
          <Calendar className="h-3.5 w-3.5 text-zinc-500" />
          <span>
            {sprint.startDate
              ? new Date(sprint.startDate).toLocaleDateString()
              : 'Not started'}{' '}
            -{' '}
            {sprint.endDate
              ? new Date(sprint.endDate).toLocaleDateString()
              : 'No end date'}
          </span>
        </div>
      </div>
    </Card>
  );
}

