'use client';

import React, { useState, useMemo, useRef, useEffect } from 'react';
import {
  Calendar,
  ChevronLeft,
  ChevronRight,
  Clock,
  AlertCircle,
  Plus,
  Link2,
  Repeat,
} from 'lucide-react';
import type { Task } from '@orbito/types';
import { useAppDispatch } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';

interface TaskGanttViewProps {
  tasks: Task[];
  canEdit?: boolean;
  onAddTask?: () => void;
}

export const TaskGanttView: React.FC<TaskGanttViewProps> = ({
  tasks = [],
  canEdit = false,
  onAddTask,
}) => {
  const dispatch = useAppDispatch();
  const [zoomMode, setZoomMode] = useState<'days' | 'weeks'>('days');
  const [timelineStart, setTimelineStart] = useState(() => {
    const d = new Date();
    d.setDate(d.getDate() - 7);
    d.setHours(0, 0, 0, 0);
    return d;
  });

  const timelineContainerRef = useRef<HTMLDivElement>(null);

  // Column width config in pixels
  const colWidth = zoomMode === 'days' ? 44 : 90;
  const numCols = zoomMode === 'days' ? 35 : 16; // 35 days or 16 weeks

  // Helper date conversions
  const toDateKey = (date: Date): string => {
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, '0');
    const d = String(date.getDate()).padStart(2, '0');
    return `${y}-${m}-${d}`;
  };

  const todayKey = useMemo(() => {
    const d = new Date();
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
  }, []);

  // Generate column dates
  const columns = useMemo(() => {
    const cols = [];
    for (let i = 0; i < numCols; i++) {
      const colDate = new Date(timelineStart);
      if (zoomMode === 'days') {
        colDate.setDate(colDate.getDate() + i);
      } else {
        colDate.setDate(colDate.getDate() + i * 7);
      }

      cols.push({
        date: colDate,
        dateKey: toDateKey(colDate),
        isToday: toDateKey(colDate) === todayKey,
        dayNum: colDate.getDate(),
        dayName: colDate.toLocaleDateString('en-US', { weekday: 'narrow' }),
        monthName: colDate.toLocaleDateString('en-US', { month: 'short' }),
      });
    }
    return cols;
  }, [timelineStart, zoomMode, numCols, todayKey]);

  // Navigate timeline
  const handlePrev = () => {
    const prev = new Date(timelineStart);
    prev.setDate(prev.getDate() - (zoomMode === 'days' ? 14 : 28));
    setTimelineStart(prev);
  };

  const handleNext = () => {
    const next = new Date(timelineStart);
    next.setDate(next.getDate() + (zoomMode === 'days' ? 14 : 28));
    setTimelineStart(next);
  };

  const handleToday = () => {
    const d = new Date();
    d.setDate(d.getDate() - (zoomMode === 'days' ? 7 : 14));
    d.setHours(0, 0, 0, 0);
    setTimelineStart(d);
  };

  // Timeline boundaries in ms
  const startMs = timelineStart.getTime();
  const endMs =
    columns.length > 0
      ? columns[columns.length - 1].date.getTime() + (zoomMode === 'days' ? 86400000 : 7 * 86400000)
      : startMs + 30 * 86400000;
  const totalDurationMs = endMs - startMs;
  const totalTimelineWidth = numCols * colWidth;

  // Task bar positioning calculation
  const getTaskBarStyle = (task: Task) => {
    const taskStart = task.startDate ? new Date(task.startDate) : task.createdAt ? new Date(task.createdAt) : new Date();
    const taskEnd = task.dueDate ? new Date(task.dueDate) : new Date(taskStart.getTime() + 86400000);

    const taskStartMs = Math.max(startMs, taskStart.getTime());
    const taskEndMs = Math.min(endMs, Math.max(taskEnd.getTime(), taskStartMs + 86400000));

    // Percentage across visible timeline
    const leftPercent = Math.max(0, Math.min(100, ((taskStartMs - startMs) / totalDurationMs) * 100));
    const widthPercent = Math.max(
      1.5,
      Math.min(100 - leftPercent, ((taskEndMs - taskStartMs) / totalDurationMs) * 100)
    );

    const isVisible = taskEnd.getTime() >= startMs && taskStart.getTime() <= endMs;

    return {
      left: `${leftPercent}%`,
      width: `${widthPercent}%`,
      isVisible,
    };
  };

  const getStatusColor = (category?: string) => {
    switch (category) {
      case 'DONE':
        return 'bg-emerald-600/80 border-emerald-500 text-emerald-100';
      case 'IN_PROGRESS':
        return 'bg-blue-600/80 border-blue-500 text-blue-100';
      case 'IN_REVIEW':
        return 'bg-amber-600/80 border-amber-500 text-amber-100';
      default:
        return 'bg-zinc-700/80 border-zinc-600 text-zinc-100';
    }
  };

  return (
    <div className="flex flex-col h-[calc(100vh-170px)] rounded-md border border-zinc-800 bg-zinc-950 overflow-hidden select-none">
      {/* Top Controls Toolbar */}
      <div className="flex items-center justify-between border-b border-zinc-800 bg-zinc-900/60 px-4 py-2 shrink-0">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={handlePrev}
              className="rounded p-1 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200 transition"
              title="Previous Range"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <button
              type="button"
              onClick={handleNext}
              className="rounded p-1 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200 transition"
              title="Next Range"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>

          <span className="text-xs font-semibold text-zinc-200">
            {timelineStart.toLocaleDateString('en-US', { month: 'short', year: 'numeric' })}
            {' — '}
            {new Date(endMs).toLocaleDateString('en-US', { month: 'short', year: 'numeric' })}
          </span>

          <button
            type="button"
            onClick={handleToday}
            className="rounded border border-zinc-700 bg-zinc-900 px-2 py-0.5 text-xs font-medium text-zinc-300 hover:bg-zinc-800 transition"
          >
            Today
          </button>
        </div>

        {/* Zoom Controls */}
        <div className="flex items-center gap-2">
          <div className="flex rounded-md border border-zinc-800 bg-zinc-900/80 p-0.5">
            <button
              type="button"
              onClick={() => setZoomMode('days')}
              className={`rounded px-2.5 py-1 text-xs font-medium transition ${
                zoomMode === 'days'
                  ? 'bg-zinc-800 text-zinc-100 shadow-sm'
                  : 'text-zinc-400 hover:text-zinc-200'
              }`}
            >
              Days
            </button>
            <button
              type="button"
              onClick={() => setZoomMode('weeks')}
              className={`rounded px-2.5 py-1 text-xs font-medium transition ${
                zoomMode === 'weeks'
                  ? 'bg-zinc-800 text-zinc-100 shadow-sm'
                  : 'text-zinc-400 hover:text-zinc-200'
              }`}
            >
              Weeks
            </button>
          </div>
        </div>
      </div>

      {/* Main Dual-Pane Body */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Pane: Task Details Table (Width: 320px) */}
        <div className="w-80 shrink-0 border-r border-zinc-800 bg-zinc-950/90 flex flex-col z-10 shadow-md">
          {/* Table Header */}
          <div className="h-10 border-b border-zinc-800 bg-zinc-900/40 px-3 flex items-center justify-between text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
            <span>Task Title</span>
            <span>Status</span>
          </div>

          {/* Table Rows */}
          <div className="flex-1 overflow-y-auto divide-y divide-zinc-800/60 no-scrollbar">
            {tasks.length === 0 ? (
              <div className="p-6 text-center text-xs text-zinc-500 italic">No tasks found</div>
            ) : (
              tasks.map((task) => {
                const assignee = task.assignees?.[0]?.user;
                const initials = assignee
                  ? ((assignee.firstName?.[0] || '') + (assignee.lastName?.[0] || '')).toUpperCase() || 'U'
                  : null;

                return (
                  <div
                    key={task.id}
                    onClick={() => dispatch(setTaskDrawerId(task.id))}
                    className="h-11 px-3 flex items-center justify-between hover:bg-zinc-900/60 transition cursor-pointer group"
                  >
                    <div className="flex items-center gap-2 truncate pr-2">
                      <span className="font-mono text-[10px] text-zinc-500 shrink-0">
                        {task.slug || 'TASK'}
                      </span>
                      {task.isRecurring && (
                        <span title="Recurring task">
                          <Repeat className="w-2.5 h-2.5 text-cyan-400 shrink-0" />
                        </span>
                      )}
                      <span className="text-xs text-zinc-200 truncate group-hover:text-white" title={task.title}>
                        {task.title}
                      </span>
                    </div>

                    <div className="flex items-center gap-1.5 shrink-0">
                      {assignee && (
                        <Avatar className="h-4 w-4 rounded-sm border border-zinc-800">
                          <AvatarFallback className="text-[8px] bg-zinc-800 text-zinc-300">
                            {initials}
                          </AvatarFallback>
                        </Avatar>
                      )}
                      <Badge variant="outline" className="text-[9px] px-1 py-0 uppercase">
                        {task.status?.name || 'TODO'}
                      </Badge>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>

        {/* Right Pane: Horizontal Timeline Grid */}
        <div ref={timelineContainerRef} className="flex-1 overflow-x-auto overflow-y-auto flex flex-col bg-zinc-950">
          {/* Timeline Header Row */}
          <div className="h-10 border-b border-zinc-800 bg-zinc-900/40 flex sticky top-0 z-10 shrink-0">
            {columns.map((col, idx) => (
              <div
                key={idx}
                style={{ width: colWidth }}
                className={`shrink-0 border-r border-zinc-800/60 flex flex-col items-center justify-center text-[10px] ${
                  col.isToday ? 'bg-zinc-900/80 font-bold text-zinc-100' : 'text-zinc-400'
                }`}
              >
                <span className="text-[9px] uppercase tracking-tighter text-zinc-500">
                  {zoomMode === 'days' ? col.dayName : col.monthName}
                </span>
                <span>{col.dayNum}</span>
              </div>
            ))}
          </div>

          {/* Timeline Task Bars Container */}
          <div className="relative flex-1" style={{ width: totalTimelineWidth }}>
            {/* Vertical Today Reference Marker */}
            {columns.map((col, idx) => {
              if (!col.isToday) return null;
              return (
                <div
                  key={`today-${idx}`}
                  style={{ left: idx * colWidth }}
                  className="absolute top-0 bottom-0 w-[1px] bg-red-500/50 z-0 pointer-events-none"
                />
              );
            })}

            {/* Background Grid Lines */}
            <div className="absolute inset-0 flex pointer-events-none">
              {columns.map((col, idx) => (
                <div
                  key={`grid-${idx}`}
                  style={{ width: colWidth }}
                  className={`h-full shrink-0 border-r border-zinc-900/80 ${
                    col.isToday ? 'bg-zinc-900/10' : ''
                  }`}
                />
              ))}
            </div>

            {/* Task Bar Rows */}
            <div className="divide-y divide-zinc-850/40 relative z-5">
              {tasks.map((task) => {
                const bar = getTaskBarStyle(task);
                const isTaskDone = task.status?.category === 'DONE';
                const isOverdue =
                  !isTaskDone && task.dueDate && toDateKey(new Date(task.dueDate)) < todayKey;

                return (
                  <div key={task.id} className="h-11 relative flex items-center px-1">
                    {bar.isVisible && (
                      <div
                        onClick={() => dispatch(setTaskDrawerId(task.id))}
                        style={{ left: bar.left, width: bar.width }}
                        className={`absolute h-7 rounded border shadow-xs px-2 flex items-center justify-between text-xs cursor-pointer transition hover:scale-[1.01] hover:brightness-110 ${getStatusColor(
                          task.status?.category
                        )}`}
                        title={`${task.slug ? `[${task.slug}] ` : ''}${task.title}`}
                      >
                        <div className="flex items-center gap-1.5 truncate">
                          {isOverdue && (
                            <span title="Overdue">
                              <AlertCircle className="h-3 w-3 text-red-300 shrink-0" />
                            </span>
                          )}
                          <span className="truncate font-medium text-[11px]">
                            {task.slug ? `[${task.slug}] ` : ''}
                            {task.title}
                          </span>
                        </div>

                        {task.priority === 'URGENT' || task.priority === 'HIGH' ? (
                          <span className="h-1.5 w-1.5 rounded-full bg-red-300 shrink-0" />
                        ) : null}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
