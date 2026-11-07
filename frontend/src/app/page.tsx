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

  
    </div>
  );
}
