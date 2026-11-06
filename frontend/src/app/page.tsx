'use client';

import React, { useMemo } from 'react';
import { useTasks } from '@/lib/hooks/useTasks';
import { WEEKS } from '@/lib/utils/constants';
import { Target, TrendingUp, Calendar, ChevronRight, ListTodo } from 'lucide-react';
import Link from 'next/link';

export default function DashboardPage() {
  const { data: tasks = [], isLoading } = useTasks();

  const stats = useMemo(() => {
    const total = tasks.length;
    const completedCount = tasks.filter((t) => t.status === 'completed').length;
    const progressCount = tasks.filter((t) => t.status === 'in-progress').length;
    const improveCount = tasks.filter((t) => t.status === 'needs-improvement').length;

    const monthlyProgress = total ? Math.round((completedCount / total) * 100) : 0;

    const weeklyBreakdown = WEEKS.map((w) => {
      const weekTasks = tasks.filter((t) => t.week === w.id);
      const done = weekTasks.filter((t) => t.status === 'completed').length;
      return {
        ...w,
        count: weekTasks.length,
        progress: weekTasks.length ? Math.round((done / weekTasks.length) * 100) : 0,
      };
    });

    return { monthlyProgress, total, completedCount, progressCount, improveCount, weeklyBreakdown };
  }, [tasks]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-[rgb(var(--color-primary))] border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-[rgb(var(--color-text-secondary))] font-semibold">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* Header */}
      <header className="px-8 py-6 bg-[rgb(var(--color-bg-primary))] border-b border-[rgb(var(--color-border-primary))]">
        <h2 className="text-2xl font-bold flex items-center gap-3">
          January Roadmap <span className="text-[rgb(var(--color-text-tertiary))] font-light">/</span> 2026
        </h2>
        <p className="text-[rgb(var(--color-text-tertiary))] text-sm mt-0.5">
          Review your month-long strategy and total progress.
        </p>
      </header>

      {/* Content */}
      <div className="flex-1 overflow-y-auto bg-[rgb(var(--color-bg-secondary))] scrollbar-thin">
        <div className="p-8 max-w-6xl mx-auto space-y-8 animate-slide-in-bottom">
          {/* Monthly Health Dashboard */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Big Progress Card */}
            <div className="lg:col-span-2 bg-[rgb(var(--color-bg-primary))] p-8 rounded-[2rem] border border-[rgb(var(--color-border-primary))] flex items-center gap-10 shadow-custom-sm">
              <div className="relative shrink-0">
                <svg className="w-32 h-32 transform -rotate-90">
                  <circle
                    cx="64"
                    cy="64"
                    r="58"
                    stroke="currentColor"
                    strokeWidth="10"
                    fill="transparent"
                    className="text-[rgb(var(--color-bg-tertiary))]"
                  />
                  <circle
                    cx="64"
                    cy="64"
                    r="58"
                    stroke="currentColor"
                    strokeWidth="10"
                    fill="transparent"
                    strokeDasharray={364.4}
                    strokeDashoffset={364.4 - (364.4 * stats.monthlyProgress) / 100}
                    className="text-[rgb(var(--color-primary))] transition-all duration-1000"
                  />
                </svg>
                <div className="absolute inset-0 flex flex-col items-center justify-center">
                  <span className="text-3xl font-black">{stats.monthlyProgress}%</span>
                </div>
              </div>
              <div>
                <h3 className="text-2xl font-bold mb-2">Monthly Completion</h3>
                <p className="text-[rgb(var(--color-text-secondary))] leading-relaxed text-sm max-w-md">
                  You have finished{' '}
                  <span className="font-bold text-[rgb(var(--color-primary))]">{stats.completedCount}</span> tasks
                  out of <span className="font-bold">{stats.total}</span> planned for January. Keep moving to reach
                  your goals.
                </p>
                <div className="mt-6 flex gap-4">
                  <div className="flex items-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-amber-400" />
                    <span className="text-xs font-bold text-[rgb(var(--color-text-secondary))]">
                      {stats.progressCount} In Progress
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-2 h-2 rounded-full bg-rose-500" />
                    <span className="text-xs font-bold text-[rgb(var(--color-text-secondary))]">
                      {stats.improveCount} Needs Work
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* Status Breakdown Mini Cards */}
            <div className="space-y-4">
              <div className="bg-[rgb(var(--color-bg-primary))] p-6 rounded-3xl border border-[rgb(var(--color-border-primary))] flex justify-between items-center shadow-custom-sm">
                <div>
                  <p className="text-xs font-bold text-[rgb(var(--color-text-tertiary))] uppercase mb-1">
                    Total Goals
                  </p>
                  <p className="text-2xl font-bold">{stats.total}</p>
                </div>
                <div className="p-3 bg-[rgb(var(--color-bg-tertiary))] rounded-2xl">
                  <Target size={20} className="text-[rgb(var(--color-text-tertiary))]" />
                </div>
              </div>
              <div className="bg-[rgb(var(--color-bg-primary))] p-6 rounded-3xl border border-[rgb(var(--color-border-primary))] flex justify-between items-center shadow-custom-sm">
                <div>
                  <p className="text-xs font-bold text-emerald-500 uppercase mb-1">Success Rate</p>
                  <p className="text-2xl font-bold text-emerald-600">
                    {stats.monthlyProgress > 70 ? 'High' : stats.monthlyProgress > 40 ? 'Medium' : 'Low'}
                  </p>
                </div>
                <div className="p-3 bg-emerald-50 dark:bg-emerald-950 rounded-2xl">
                  <TrendingUp size={20} className="text-emerald-500" />
                </div>
              </div>
            </div>
          </div>

          {/* Week Grid Navigator */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {stats.weeklyBreakdown.map((wd) => (
              <Link
                key={wd.id}
                href={`/tasks?week=${wd.id}`}
                className="group bg-[rgb(var(--color-bg-primary))] p-6 rounded-[2rem] border border-[rgb(var(--color-border-primary))] hover:border-[rgb(var(--color-primary))] hover:shadow-xl transition-all text-left"
              >
                <div className="flex justify-between items-start mb-6">
                  <div className="p-3 bg-[rgb(var(--color-primary-light))] text-[rgb(var(--color-primary))] rounded-2xl group-hover:bg-[rgb(var(--color-primary))] group-hover:text-white transition-all">
                    <Calendar size={20} />
                  </div>
                  <span className="text-[10px] font-black text-[rgb(var(--color-text-tertiary))] uppercase tracking-tighter">
                    Week 0{wd.id}
                  </span>
                </div>
                <h4 className="font-bold text-lg">{wd.label}</h4>
                <p className="text-xs text-[rgb(var(--color-text-tertiary))] mb-6 font-medium">{wd.range}</p>

                <div className="space-y-2">
                  <div className="flex justify-between text-[10px] font-bold text-[rgb(var(--color-text-tertiary))] uppercase">
                    <span>Progress</span>
                    <span className="text-[rgb(var(--color-primary))]">{wd.progress}%</span>
                  </div>
                  <div className="w-full bg-[rgb(var(--color-bg-tertiary))] h-2 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-[rgb(var(--color-primary))] transition-all duration-700"
                      style={{ width: `${wd.progress}%` }}
                    />
                  </div>
                </div>
              </Link>
            ))}
          </div>

          {/* Combined Monthly List */}
          <div className="bg-[rgb(var(--color-bg-primary))] rounded-[2rem] border border-[rgb(var(--color-border-primary))] overflow-hidden shadow-custom-sm">
            <div className="p-6 border-b border-[rgb(var(--color-border-secondary))] flex justify-between items-center bg-[rgb(var(--color-bg-tertiary))]/50">
              <h3 className="font-bold flex items-center gap-2">
                <ListTodo size={18} className="text-[rgb(var(--color-primary))]" />
                Comprehensive Monthly Log
              </h3>
              <span className="text-[10px] font-black bg-[rgb(var(--color-primary-light))] text-[rgb(var(--color-primary))] px-3 py-1 rounded-full uppercase">
                All Weeks
              </span>
            </div>
            <div className="p-6">
              <div className="space-y-10">
                {WEEKS.map((w) => {
                  const weekTasks = tasks.filter((t) => t.week === w.id);
                  return (
                    <div key={w.id}>
                      <div className="flex items-center gap-3 mb-4">
                        <span className="text-xs font-black text-[rgb(var(--color-text-tertiary))] uppercase tracking-widest">
                          {w.label}
                        </span>
                        <div className="h-[1px] flex-1 bg-[rgb(var(--color-border-secondary))]" />
                        <span className="text-[10px] text-[rgb(var(--color-text-tertiary))] font-bold">{w.range}</span>
                      </div>
                      <div className="grid gap-3">
                        {weekTasks.map((task) => (
                          <Link
                            key={task.id}
                            href={`/tasks?week=${task.week}`}
                            className="flex items-center justify-between p-4 bg-[rgb(var(--color-bg-tertiary))] rounded-2xl hover:bg-[rgb(var(--color-bg-primary))] border border-transparent hover:border-[rgb(var(--color-border-primary))] transition-all cursor-pointer group"
                          >
                            <div className="flex items-center gap-4">
                              <div
                                className={`w-2 h-2 rounded-full ${task.status === 'completed'
                                    ? 'bg-emerald-500'
                                    : task.status === 'in-progress'
                                      ? 'bg-amber-500'
                                      : task.status === 'needs-improvement'
                                        ? 'bg-rose-500'
                                        : 'bg-slate-400'
                                  }`}
                              />
                              <span className="font-bold text-sm">{task.text}</span>
                            </div>
                            <div className="flex items-center gap-6">
                              <span className="text-[10px] font-bold text-[rgb(var(--color-text-tertiary))] uppercase tracking-wider">
                                {task.status.replace('-', ' ')}
                              </span>
                              <ChevronRight
                                size={16}
                                className="text-[rgb(var(--color-text-tertiary))] group-hover:text-[rgb(var(--color-primary))] transition-all"
                              />
                            </div>
                          </Link>
                        ))}
                        {weekTasks.length === 0 && (
                          <p className="text-xs text-[rgb(var(--color-text-tertiary))] italic text-center py-4 border-2 border-dashed border-[rgb(var(--color-border-secondary))] rounded-2xl">
                            No tasks scheduled for this week.
                          </p>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
