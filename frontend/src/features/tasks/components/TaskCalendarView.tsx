'use client';

import React, { useState, useMemo } from 'react';
import {
  ChevronLeft,
  ChevronRight,
  Calendar as CalendarIcon,
  Clock,
  AlertCircle,
  Plus,
} from 'lucide-react';
import type { Task } from '@orbito/types';
import { useAppDispatch } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';

interface TaskCalendarViewProps {
  tasks: Task[];
  canEdit?: boolean;
  onOpenCreateTask?: (date?: string) => void;
}

export const TaskCalendarView: React.FC<TaskCalendarViewProps> = ({
  tasks = [],
  canEdit = false,
  onOpenCreateTask,
}) => {
  const dispatch = useAppDispatch();
  const [currentDate, setCurrentDate] = useState(() => new Date());
  const [viewMode, setViewMode] = useState<'month' | 'week'>('month');

  // Month navigation helpers
  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();

  const handlePrev = () => {
    if (viewMode === 'month') {
      setCurrentDate(new Date(year, month - 1, 1));
    } else {
      const prevWeek = new Date(currentDate);
      prevWeek.setDate(prevWeek.getDate() - 7);
      setCurrentDate(prevWeek);
    }
  };

  const handleNext = () => {
    if (viewMode === 'month') {
      setCurrentDate(new Date(year, month + 1, 1));
    } else {
      const nextWeek = new Date(currentDate);
      nextWeek.setDate(nextWeek.getDate() + 7);
      setCurrentDate(nextWeek);
    }
  };

  const handleToday = () => {
    setCurrentDate(new Date());
  };

  // Format date helper: YYYY-MM-DD
  const toDateKey = (date: Date): string => {
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, '0');
    const d = String(date.getDate()).padStart(2, '0');
    return `${y}-${m}-${d}`;
  };

  const todayKey = toDateKey(new Date());

  // Group tasks by dueDate (or fallback createdAt if needed)
  const tasksByDate = useMemo(() => {
    const map = new Map<string, Task[]>();

    for (const task of tasks) {
      const rawDate = task.dueDate || task.createdAt;
      if (!rawDate) continue;

      try {
        const d = new Date(rawDate);
        if (isNaN(d.getTime())) continue;
        const key = toDateKey(d);
        const existing = map.get(key) || [];
        existing.push(task);
        map.set(key, existing);
      } catch {
        // Skip invalid date
      }
    }

    return map;
  }, [tasks]);

  // Compute Calendar Month Grid (42 cells: 6 weeks x 7 days)
  const monthGridDays = useMemo(() => {
    const firstDayOfMonth = new Date(year, month, 1);
    const dayOfWeek = firstDayOfMonth.getDay(); // 0 is Sunday

    const startDate = new Date(firstDayOfMonth);
    startDate.setDate(startDate.getDate() - dayOfWeek);

    const days = [];
    for (let i = 0; i < 42; i++) {
      const cellDate = new Date(startDate);
      cellDate.setDate(cellDate.getDate() + i);

      const dateKey = toDateKey(cellDate);
      const isCurrentMonth = cellDate.getMonth() === month;
      const isToday = dateKey === todayKey;
      const dayTasks = tasksByDate.get(dateKey) || [];

      days.push({
        date: cellDate,
        dateKey,
        isCurrentMonth,
        isToday,
        tasks: dayTasks,
      });
    }

    return days;
  }, [year, month, tasksByDate, todayKey]);

  // Compute Week Grid (7 days)
  const weekGridDays = useMemo(() => {
    const curr = new Date(currentDate);
    const day = curr.getDay(); // Sunday - Saturday : 0 - 6
    const sunday = new Date(curr);
    sunday.setDate(sunday.getDate() - day);

    const days = [];
    for (let i = 0; i < 7; i++) {
      const cellDate = new Date(sunday);
      cellDate.setDate(cellDate.getDate() + i);

      const dateKey = toDateKey(cellDate);
      const isToday = dateKey === todayKey;
      const dayTasks = tasksByDate.get(dateKey) || [];

      days.push({
        date: cellDate,
        dateKey,
        isToday,
        tasks: dayTasks,
      });
    }

    return days;
  }, [currentDate, tasksByDate, todayKey]);

  const monthTitle = useMemo(() => {
    return currentDate.toLocaleDateString('en-US', {
      month: 'long',
      year: 'numeric',
    });
  }, [currentDate]);

  const weekDaysHeader = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

  // Status border helper
  const getStatusBorderColor = (category?: string) => {
    switch (category) {
      case 'DONE':
        return 'border-l-emerald-500';
      case 'IN_PROGRESS':
        return 'border-l-blue-500';
      case 'IN_REVIEW':
        return 'border-l-amber-500';
      default:
        return 'border-l-zinc-500';
    }
  };

  return (
    <div className="flex flex-col h-[calc(100vh-170px)] rounded-md border border-zinc-800 bg-zinc-950 overflow-hidden select-none">
      {/* Calendar Header Controls */}
      <div className="flex flex-wrap items-center justify-between border-b border-zinc-800 bg-zinc-900/60 px-4 py-2.5">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={handlePrev}
              className="rounded p-1 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200 transition"
              title="Previous"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <button
              type="button"
              onClick={handleNext}
              className="rounded p-1 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200 transition"
              title="Next"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>

          <h3 className="text-sm font-semibold tracking-wide text-zinc-100 min-w-36">
            {monthTitle}
          </h3>

          <button
            type="button"
            onClick={handleToday}
            className="rounded border border-zinc-700 bg-zinc-900 px-2 py-0.5 text-xs font-medium text-zinc-300 hover:bg-zinc-800 transition"
          >
            Today
          </button>
        </div>

        {/* View Toggle */}
        <div className="flex items-center gap-2">
          <div className="flex rounded-md border border-zinc-800 bg-zinc-900/80 p-0.5">
            <button
              type="button"
              onClick={() => setViewMode('month')}
              className={`rounded px-2.5 py-1 text-xs font-medium transition ${
                viewMode === 'month'
                  ? 'bg-zinc-800 text-zinc-100 shadow-sm'
                  : 'text-zinc-400 hover:text-zinc-200'
              }`}
            >
              Month
            </button>
            <button
              type="button"
              onClick={() => setViewMode('week')}
              className={`rounded px-2.5 py-1 text-xs font-medium transition ${
                viewMode === 'week'
                  ? 'bg-zinc-800 text-zinc-100 shadow-sm'
                  : 'text-zinc-400 hover:text-zinc-200'
              }`}
            >
              Week
            </button>
          </div>
        </div>
      </div>

      {/* Weekday Labels Header */}
      <div className="grid grid-cols-7 border-b border-zinc-800 bg-zinc-900/40 text-center text-[11px] font-medium text-zinc-400 py-1.5">
        {weekDaysHeader.map((d) => (
          <div key={d} className="uppercase tracking-wider">
            {d}
          </div>
        ))}
      </div>

      {/* Calendar Grid View */}
      {viewMode === 'month' ? (
        <div className="grid grid-cols-7 grid-rows-6 flex-1 divide-x divide-y divide-zinc-800/60 overflow-hidden bg-zinc-950">
          {monthGridDays.map((day) => {
            const isOverdue =
              day.dateKey < todayKey && day.tasks.some((t) => t.status?.category !== 'DONE');

            return (
              <div
                key={day.dateKey}
                className={`group relative flex flex-col p-1.5 transition-colors overflow-hidden ${
                  !day.isCurrentMonth
                    ? 'bg-zinc-950/40 opacity-40'
                    : 'bg-zinc-950/80 hover:bg-zinc-900/20'
                } ${day.isToday ? 'bg-zinc-900/40' : ''}`}
              >
                {/* Date Header */}
                <div className="flex items-center justify-between mb-1">
                  <span
                    className={`flex h-5 w-5 items-center justify-center rounded-full text-xs font-medium ${
                      day.isToday
                        ? 'bg-zinc-100 text-zinc-950 font-bold'
                        : 'text-zinc-400 group-hover:text-zinc-200'
                    }`}
                  >
                    {day.date.getDate()}
                  </span>

                  {canEdit && (
                    <button
                      type="button"
                      onClick={() => onOpenCreateTask?.(day.dateKey)}
                      className="opacity-0 group-hover:opacity-100 p-0.5 text-zinc-500 hover:text-zinc-200 transition"
                      title="Add task on this date"
                    >
                      <Plus className="h-3.5 w-3.5" />
                    </button>
                  )}
                </div>

                {/* Tasks List in Cell */}
                <div className="flex-1 space-y-1 overflow-y-auto no-scrollbar pr-0.5">
                  {day.tasks.slice(0, 3).map((task) => {
                    const isTaskDone = task.status?.category === 'DONE';
                    const isTaskOverdue =
                      !isTaskDone && task.dueDate && toDateKey(new Date(task.dueDate)) < todayKey;

                    return (
                      <div
                        key={task.id}
                        onClick={() => dispatch(setTaskDrawerId(task.id))}
                        className={`group/task flex cursor-pointer items-center justify-between rounded border border-zinc-800 bg-zinc-900/90 px-1.5 py-1 text-[11px] transition hover:border-zinc-700 hover:bg-zinc-850 border-l-2 ${getStatusBorderColor(
                          task.status?.category
                        )}`}
                      >
                        <div className="flex items-center gap-1.5 truncate">
                          {isTaskOverdue && (
                            <span title="Overdue" className="shrink-0 flex items-center">
                              <AlertCircle className="h-3 w-3 text-red-400" />
                            </span>
                          )}
                          <span
                            className={`truncate ${
                              isTaskDone ? 'text-zinc-500 line-through' : 'text-zinc-200'
                            }`}
                            title={task.title}
                          >
                            {task.slug ? `[${task.slug}] ` : ''}
                            {task.title}
                          </span>
                        </div>

                        {task.priority === 'URGENT' || task.priority === 'HIGH' ? (
                          <span className="h-1.5 w-1.5 rounded-full bg-red-400 shrink-0" />
                        ) : null}
                      </div>
                    );
                  })}

                  {day.tasks.length > 3 && (
                    <div
                      onClick={() => dispatch(setTaskDrawerId(day.tasks[3].id))}
                      className="cursor-pointer text-[10px] text-zinc-500 hover:text-zinc-300 font-medium px-1"
                    >
                      +{day.tasks.length - 3} more
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        /* Week View Grid */
        <div className="grid grid-cols-7 flex-1 divide-x divide-zinc-800/80 overflow-y-auto bg-zinc-950">
          {weekGridDays.map((day) => (
            <div
              key={day.dateKey}
              className={`flex flex-col p-2 space-y-2 ${
                day.isToday ? 'bg-zinc-900/30' : 'bg-transparent'
              }`}
            >
              <div className="flex items-center justify-between border-b border-zinc-800/60 pb-2">
                <div className="flex items-center gap-2">
                  <span
                    className={`flex h-6 w-6 items-center justify-center rounded-full text-xs font-semibold ${
                      day.isToday ? 'bg-zinc-100 text-zinc-950' : 'text-zinc-300'
                    }`}
                  >
                    {day.date.getDate()}
                  </span>
                  <span className="text-xs text-zinc-400 font-medium">
                    {day.date.toLocaleDateString('en-US', { weekday: 'short' })}
                  </span>
                </div>

                {canEdit && (
                  <button
                    type="button"
                    onClick={() => onOpenCreateTask?.(day.dateKey)}
                    className="p-1 text-zinc-500 hover:text-zinc-200 transition"
                  >
                    <Plus className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>

              <div className="flex-1 space-y-1.5 overflow-y-auto pr-0.5">
                {day.tasks.length === 0 ? (
                  <span className="text-[11px] text-zinc-600 italic block py-4 text-center">
                    No scheduled tasks
                  </span>
                ) : (
                  day.tasks.map((task) => {
                    const isTaskDone = task.status?.category === 'DONE';
                    const isTaskOverdue =
                      !isTaskDone && task.dueDate && toDateKey(new Date(task.dueDate)) < todayKey;

                    return (
                      <div
                        key={task.id}
                        onClick={() => dispatch(setTaskDrawerId(task.id))}
                        className={`group/task flex cursor-pointer flex-col gap-1 rounded border border-zinc-800 bg-zinc-900/90 p-2 text-xs transition hover:border-zinc-700 hover:bg-zinc-850 border-l-3 ${getStatusBorderColor(
                          task.status?.category
                        )}`}
                      >
                        <div className="flex items-center justify-between gap-1">
                          <span className="font-mono text-[10px] text-zinc-500">
                            {task.slug || 'TASK'}
                          </span>
                          {isTaskOverdue && (
                            <span className="flex items-center gap-0.5 text-[10px] text-red-400 font-medium">
                              <AlertCircle className="h-2.5 w-2.5" /> Overdue
                            </span>
                          )}
                        </div>

                        <span
                          className={`font-medium ${
                            isTaskDone ? 'text-zinc-500 line-through' : 'text-zinc-200'
                          }`}
                        >
                          {task.title}
                        </span>

                        <div className="flex items-center justify-between text-[10px] text-zinc-400 pt-1">
                          <span>{task.status?.name || 'To Do'}</span>
                          {task.assignees && task.assignees.length > 0 && (
                            <span className="truncate max-w-20 text-zinc-400">
                              {task.assignees[0].user?.firstName || 'Assigned'}
                            </span>
                          )}
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

