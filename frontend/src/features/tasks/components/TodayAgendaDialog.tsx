'use client';

import * as React from 'react';
import { useMemo } from 'react';
import { useRouter } from 'next/navigation';
import dayjs from 'dayjs';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import {
  Calendar,
  Clock,
  CheckCircle2,
  Sparkles,
  AlertCircle,
  ExternalLink,
  ChevronRight,
  Flame,
} from 'lucide-react';
import { useGetTodaysTasksQuery } from '@/lib/store/api/tasksApi';
import type { Task } from '@orbito/types';

interface TodayAgendaDialogProps {
  isOpen: boolean;
  onClose: () => void;
  workspaceSlug?: string;
  projectSlug?: string;
  onOpenTask?: (taskId: string) => void;
}

export const TodayAgendaDialog: React.FC<TodayAgendaDialogProps> = ({
  isOpen,
  onClose,
  workspaceSlug,
  projectSlug,
  onOpenTask,
}) => {
  const router = useRouter();

  const { data, isLoading } = useGetTodaysTasksQuery(undefined, {
    skip: !isOpen,
  });

  const tasks = data?.tasks || [];

  // Group tasks into overdue and due today
  const { overdueTasks, dueTodayTasks, upcomingTasks } = useMemo(() => {
    const overdue: Task[] = [];
    const today: Task[] = [];
    const upcoming: Task[] = [];
    const now = dayjs();

    tasks.forEach((t) => {
      if (!t.dueDate) {
        upcoming.push(t);
        return;
      }
      const due = dayjs(t.dueDate);
      if (due.isBefore(now, 'day')) {
        overdue.push(t);
      } else if (due.isSame(now, 'day')) {
        today.push(t);
      } else {
        upcoming.push(t);
      }
    });

    return { overdueTasks: overdue, dueTodayTasks: today, upcomingTasks: upcoming };
  }, [tasks]);

  const handleTaskClick = (task: Task) => {
    onClose();
    if (onOpenTask) {
      onOpenTask(task.id);
    } else if (workspaceSlug && (task.project?.slug || projectSlug)) {
      router.push(`/${workspaceSlug}/${task.project?.slug || projectSlug}/tasks/${task.id}`);
    }
  };

  const todayFormatted = dayjs().format('dddd, MMMM D, YYYY');

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-lg bg-zinc-950 border-zinc-800 text-zinc-200 p-0 overflow-hidden shadow-2xl rounded-xl">
        <DialogHeader className="p-5 pb-3 border-b border-zinc-850 bg-zinc-900/60">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-md bg-indigo-950/50 border border-indigo-800/40 flex items-center justify-center text-indigo-400 shrink-0">
              <Calendar className="h-5 w-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <DialogTitle className="text-sm font-semibold text-zinc-100">
                  Today&apos;s Agenda
                </DialogTitle>
                <Sparkles className="h-3.5 w-3.5 text-indigo-400" />
                {tasks.length > 0 && (
                  <Badge variant="outline" className="text-[10px] bg-indigo-950/40 text-indigo-300 border-indigo-800/40 font-mono">
                    {tasks.length} total
                  </Badge>
                )}
              </div>
              <DialogDescription className="text-xs text-zinc-400 mt-0.5">
                {todayFormatted}
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="max-h-[60vh] overflow-y-auto p-4 space-y-4">
          {isLoading ? (
            <div className="py-12 flex flex-col items-center justify-center gap-3">
              <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-200" />
              <span className="text-xs text-zinc-400">Loading today&apos;s tasks...</span>
            </div>
          ) : tasks.length === 0 ? (
            /* All Clear Empty State */
            <div className="py-10 text-center flex flex-col items-center">
              <div className="h-12 w-12 rounded-xl bg-emerald-950/40 border border-emerald-800/40 flex items-center justify-center text-emerald-400 mb-3 shadow-inner">
                <CheckCircle2 className="h-6 w-6" />
              </div>
              <h4 className="text-sm font-semibold text-zinc-100 mb-1">All Clear!</h4>
              <p className="text-xs text-zinc-400 max-w-xs mb-3">
                You have no tasks scheduled for today. Take a break or plan your next sprint.
              </p>
              <div className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-emerald-950/30 border border-emerald-800/40 text-[11px] text-emerald-300 font-medium">
                <Sparkles className="h-3 w-3" />
                <span>You&apos;re completely up to date</span>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              {/* Overdue Section */}
              {overdueTasks.length > 0 && (
                <div className="space-y-2">
                  <div className="flex items-center gap-1.5 text-xs font-semibold text-rose-400 uppercase tracking-wider px-1">
                    <AlertCircle className="h-3.5 w-3.5" />
                    <span>Overdue ({overdueTasks.length})</span>
                  </div>
                  <div className="space-y-1.5">
                    {overdueTasks.map((task) => (
                      <TaskAgendaCard
                        key={task.id}
                        task={task}
                        isOverdue
                        onClick={() => handleTaskClick(task)}
                      />
                    ))}
                  </div>
                </div>
              )}

              {/* Due Today Section */}
              {dueTodayTasks.length > 0 && (
                <div className="space-y-2">
                  <div className="flex items-center gap-1.5 text-xs font-semibold text-emerald-400 uppercase tracking-wider px-1">
                    <Clock className="h-3.5 w-3.5" />
                    <span>Due Today ({dueTodayTasks.length})</span>
                  </div>
                  <div className="space-y-1.5">
                    {dueTodayTasks.map((task) => (
                      <TaskAgendaCard
                        key={task.id}
                        task={task}
                        onClick={() => handleTaskClick(task)}
                      />
                    ))}
                  </div>
                </div>
              )}

              {/* Upcoming / Other Section */}
              {upcomingTasks.length > 0 && (
                <div className="space-y-2">
                  <div className="flex items-center gap-1.5 text-xs font-semibold text-zinc-400 uppercase tracking-wider px-1">
                    <Calendar className="h-3.5 w-3.5" />
                    <span>Upcoming ({upcomingTasks.length})</span>
                  </div>
                  <div className="space-y-1.5">
                    {upcomingTasks.map((task) => (
                      <TaskAgendaCard
                        key={task.id}
                        task={task}
                        onClick={() => handleTaskClick(task)}
                      />
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};

interface TaskAgendaCardProps {
  task: Task;
  isOverdue?: boolean;
  onClick: () => void;
}

const TaskAgendaCard: React.FC<TaskAgendaCardProps> = ({
  task,
  isOverdue,
  onClick,
}) => {
  const isHighPriority = task.priority === 'URGENT' || task.priority === 'HIGH';

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => e.key === 'Enter' && onClick()}
      className={`group p-3 rounded-md border text-left transition-all cursor-pointer flex items-center justify-between gap-3 ${
        isOverdue
          ? 'bg-rose-950/20 border-rose-800/40 hover:bg-rose-950/30 hover:border-rose-700/60'
          : 'bg-zinc-900/60 border-zinc-800 hover:bg-zinc-850 hover:border-zinc-700'
      }`}
    >
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2 mb-1">
          <span className="font-mono text-[10px] text-zinc-500 uppercase tracking-wider shrink-0">
            {task.slug || `TASK-${task.taskNumber || task.id.slice(0, 4)}`}
          </span>

          <Badge
            variant="outline"
            className={`text-[9px] uppercase font-mono px-1.5 py-0 ${
              isHighPriority
                ? 'text-rose-400 border-rose-500/30 bg-rose-950/30'
                : 'text-zinc-400 border-zinc-700'
            }`}
          >
            {task.priority}
          </Badge>

          {task.status && (
            <Badge variant="outline" className="text-[9px] px-1.5 py-0 text-zinc-400 border-zinc-800">
              {task.status.name}
            </Badge>
          )}
        </div>

        <h5 className="text-xs font-medium text-zinc-200 group-hover:text-white transition-colors truncate">
          {task.title}
        </h5>

        <div className="flex items-center gap-3 mt-1.5 text-[11px] text-zinc-400">
          {task.dueDate && (
            <span className={`flex items-center gap-1 font-mono ${isOverdue ? 'text-rose-400 font-semibold' : 'text-zinc-400'}`}>
              <Clock className="h-3 w-3 shrink-0" />
              <span>{isOverdue ? `Overdue (${dayjs(task.dueDate).format('MMM D')})` : dayjs(task.dueDate).format('MMM D, YYYY')}</span>
            </span>
          )}

          {task.project && (
            <span className="text-zinc-500 truncate max-w-[120px]">
              {task.project.name}
            </span>
          )}
        </div>
      </div>

      <ChevronRight className="h-4 w-4 text-zinc-600 group-hover:text-zinc-300 transition-colors shrink-0" />
    </div>
  );
};

