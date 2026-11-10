'use client';

import React, { useMemo } from 'react';
import { Users, AlertCircle, CheckCircle2, User, Clock } from 'lucide-react';
import type { Sprint, Task } from '@orbito/types';
import { Card } from '@/components/ui/card';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';

interface SprintCapacityBarProps {
  sprint: Sprint;
  tasks: Task[];
}

interface MemberCapacity {
  userId: string;
  name: string;
  initials: string;
  taskCount: number;
  totalPoints: number;
  completedPoints: number;
  percentageOfSprint: number;
}

export const SprintCapacityBar: React.FC<SprintCapacityBarProps> = ({
  sprint,
  tasks,
}) => {
  const { memberStats, unassignedTasks, totalSprintPoints } = useMemo(() => {
    // Filter tasks for this sprint
    const sprintTasks = tasks.filter((t) => t.sprintId === sprint.id);

    const totalPts = sprintTasks.reduce(
      (sum, t) => sum + (t.storyPoints && t.storyPoints > 0 ? t.storyPoints : 1),
      0
    );

    const map: Record<string, MemberCapacity> = {};
    const unassigned: Task[] = [];

    sprintTasks.forEach((task) => {
      const primaryAssignee = task.assignees?.[0]?.user;
      const pts = task.storyPoints && task.storyPoints > 0 ? task.storyPoints : 1;
      const isDone = task.status?.category === 'DONE';

      if (!primaryAssignee) {
        unassigned.push(task);
      } else {
        const id = primaryAssignee.id;
        if (!map[id]) {
          const initials = (
            (primaryAssignee.firstName?.[0] || '') +
            (primaryAssignee.lastName?.[0] || '')
          ).toUpperCase() || 'U';

          map[id] = {
            userId: id,
            name: `${primaryAssignee.firstName} ${primaryAssignee.lastName}`,
            initials,
            taskCount: 0,
            totalPoints: 0,
            completedPoints: 0,
            percentageOfSprint: 0,
          };
        }

        map[id].taskCount += 1;
        map[id].totalPoints += pts;
        if (isDone) {
          map[id].completedPoints += pts;
        }
      }
    });

    const members = Object.values(map).map((m) => ({
      ...m,
      percentageOfSprint:
        totalPts > 0 ? Math.round((m.totalPoints / totalPts) * 100) : 0,
    }));

    // Sort by points descending
    members.sort((a, b) => b.totalPoints - a.totalPoints);

    return {
      memberStats: members,
      unassignedTasks: unassigned,
      totalSprintPoints: totalPts,
    };
  }, [sprint, tasks]);

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
        <div className="flex items-center gap-2">
          <Users className="h-4 w-4 text-sky-400" />
          <h3 className="text-sm font-semibold text-zinc-100">
            Workload & Capacity Balance &mdash; {sprint.name}
          </h3>
        </div>

        <Badge variant="outline" className="text-xs font-mono border-zinc-800 text-zinc-400">
          {totalSprintPoints} Total Story Points
        </Badge>
      </div>

      {/* Unassigned Work Warning */}
      {unassignedTasks.length > 0 && (
        <div className="p-3 rounded-md border border-amber-800/60 bg-amber-950/20 flex items-center justify-between gap-3 text-xs text-amber-300">
          <div className="flex items-center gap-2">
            <AlertCircle className="h-4 w-4 text-amber-400 shrink-0" />
            <span>
              <strong>{unassignedTasks.length} tasks</strong> in this sprint currently have no assignee ({unassignedTasks.reduce((s, t) => s + (t.storyPoints || 1), 0)} pts unassigned).
            </span>
          </div>
        </div>
      )}

      {/* Members Workload Cards */}
      {memberStats.length === 0 ? (
        <Card className="border-zinc-800 bg-zinc-950/40 p-6 text-center text-xs text-zinc-500 italic">
          No team members assigned to tasks in this sprint yet.
        </Card>
      ) : (
        <div className="space-y-3">
          {memberStats.map((member) => {
            const completionPercent =
              member.totalPoints > 0
                ? Math.round((member.completedPoints / member.totalPoints) * 100)
                : 0;

            return (
              <Card
                key={member.userId}
                className="border-zinc-800 bg-zinc-950/80 p-4 space-y-3"
              >
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-2.5">
                    <Avatar className="h-7 w-7 rounded-sm border border-zinc-800">
                      <AvatarFallback className="text-[10px] bg-zinc-800 text-zinc-200">
                        {member.initials}
                      </AvatarFallback>
                    </Avatar>

                    <div>
                      <span className="text-xs font-semibold text-zinc-100 block">
                        {member.name}
                      </span>
                      <span className="text-[10px] text-zinc-400">
                        {member.taskCount} tasks &bull; {member.totalPoints} points ({member.percentageOfSprint}% of sprint)
                      </span>
                    </div>
                  </div>

                  <div className="text-right">
                    <span className="text-xs font-mono font-medium text-emerald-400">
                      {member.completedPoints} / {member.totalPoints} pts done
                    </span>
                    <span className="text-[10px] text-zinc-500 block">
                      {completionPercent}% completed
                    </span>
                  </div>
                </div>

                {/* Progress bar */}
                <div className="w-full bg-zinc-900 rounded-full h-1.5 overflow-hidden">
                  <div
                    className="bg-emerald-500 h-full rounded-full transition-all duration-300"
                    style={{ width: `${completionPercent}%` }}
                  />
                </div>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
};