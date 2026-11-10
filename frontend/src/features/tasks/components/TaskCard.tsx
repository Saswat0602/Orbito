'use client';

import * as React from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import {
  Clock,
  AlertCircle,
  GripVertical,
  Calendar,
  Repeat,
} from 'lucide-react';
import type { Task, TaskPriority } from '@orbito/types';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { useAppDispatch } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';

interface TaskCardProps {
  task: Task;
  isOverlay?: boolean;
}

export const TaskCard = React.memo(function TaskCard({ task, isOverlay = false }: TaskCardProps) {
  const dispatch = useAppDispatch();

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: task.id,
    data: {
      type: 'Task',
      task,
    },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const getPriorityStyle = (priority: TaskPriority) => {
    switch (priority) {
      case 'URGENT':
        return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
      case 'HIGH':
        return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
      case 'MEDIUM':
        return 'bg-sky-500/10 text-sky-400 border-sky-500/20';
      case 'LOW':
      default:
        return 'bg-zinc-800 text-zinc-400 border-zinc-700';
    }
  };

  const handleClick = (e: React.MouseEvent) => {
    // Only open drawer if not dragging
    if (!isDragging) {
      dispatch(setTaskDrawerId(task.id));
    }
  };

  const primaryAssignee = task.assignees?.[0]?.user;

  const assigneeInitials = React.useMemo(() => {
    if (!primaryAssignee) return null;
    const first = primaryAssignee.firstName?.[0] || '';
    const last = primaryAssignee.lastName?.[0] || '';
    return (first + last).toUpperCase() || primaryAssignee.username?.[0]?.toUpperCase() || 'U';
  }, [primaryAssignee]);

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`
        group relative rounded-md border bg-zinc-900/90 p-3.5 space-y-3 transition-all
        cursor-pointer select-none
        ${
          isDragging
            ? 'opacity-40 border-zinc-500 ring-2 ring-zinc-400/20 shadow-xl'
            : 'border-zinc-800 hover:border-zinc-700 hover:bg-zinc-900 shadow-sm'
        }
        ${isOverlay ? 'rotate-1 shadow-2xl border-zinc-600 ring-1 ring-zinc-500' : ''}
      `}
      onClick={handleClick}
    >
      {/* Top Meta: Identifier + Drag Handle */}
      <div className="flex items-center justify-between gap-2">
        <span className="font-mono text-[10px] font-semibold text-zinc-500 uppercase tracking-wider">
          {task.slug || `TASK-${task.id.slice(0, 4)}`}
        </span>

        <div className="flex items-center gap-1">
          {task.isRecurring && (
            <span
              title="Recurring task"
              className="inline-flex items-center gap-0.5 text-[9px] text-cyan-400 bg-cyan-950/40 border border-cyan-800/40 px-1 py-0.5 rounded-xs font-medium"
            >
              <Repeat className="w-2.5 h-2.5" />
              <span>Repeat</span>
            </span>
          )}

          <Badge
            variant="outline"
            className={`text-[9px] uppercase font-semibold px-1.5 py-0 border ${getPriorityStyle(
              task.priority as TaskPriority
            )}`}
          >
            {task.priority}
          </Badge>

          {/* Drag Handle */}
          <button
            {...attributes}
            {...listeners}
            type="button"
            className="p-1 -mr-1 text-zinc-600 hover:text-zinc-300 opacity-0 group-hover:opacity-100 transition-opacity cursor-grab active:cursor-grabbing"
            aria-label="Drag task"
            onClick={(e) => e.stopPropagation()}
          >
            <GripVertical className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>

      {/* Title */}
      <h4 className="text-xs font-medium text-zinc-200 line-clamp-2 leading-relaxed">
        {task.title}
      </h4>

      {/* Footer: Due date + Assignee */}
      <div className="flex items-center justify-between pt-1 border-t border-zinc-850/60 text-[11px] text-zinc-500">
        <div className="flex items-center gap-1.5">
          {task.dueDate ? (
            <div className="flex items-center gap-1 text-zinc-400">
              <Calendar className="h-3 w-3" />
              <span>{new Date(task.dueDate).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}</span>
            </div>
          ) : (
            <span className="text-[10px] text-zinc-600">No due date</span>
          )}
        </div>

        {primaryAssignee ? (
          <Avatar
            className="h-5 w-5 rounded-full border border-zinc-800"
            title={primaryAssignee.username || undefined}
          >
            <AvatarFallback className="bg-zinc-800 text-zinc-300 text-[9px] font-bold">
              {assigneeInitials}
            </AvatarFallback>
          </Avatar>
        ) : (
          <div className="h-5 w-5 rounded-full border border-dashed border-zinc-700 flex items-center justify-center text-[9px] text-zinc-600">
            -
          </div>
        )}
      </div>
    </div>
  );
});
