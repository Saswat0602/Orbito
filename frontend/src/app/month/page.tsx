'use client';

import React from 'react';
import { useTasks } from '@/lib/hooks/useTasks';
import { Header } from '@/components/layout/Header';
import { TaskCard } from '@/components/tasks/TaskCard';
import { Task } from '@/types';
import { useRouter } from 'next/navigation';
import { Calendar } from 'lucide-react';

export default function MonthViewPage() {
    const router = useRouter();
    const { data: allTasks = [], isLoading } = useTasks();

    const handleTaskClick = (task: Task) => {
        router.push(`/tasks/${task.id}`);
    };

    // Group tasks by week
    const tasksByWeek = allTasks.reduce((acc, task) => {
        const week = task.week || 1;
        if (!acc[week]) {
            acc[week] = [];
        }
        acc[week].push(task);
        return acc;
    }, {} as Record<number, Task[]>);

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-full">
                <div className="text-center">
                    <div className="w-16 h-16 border-4 border-[rgb(var(--color-primary))] border-t-transparent rounded-full animate-spin mx-auto mb-4" />
                    <p className="text-[rgb(var(--color-text-secondary))] font-semibold">Loading tasks...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full overflow-hidden">
            <Header
                title="Month View"
                subtitle="January 2026 - All Tasks"
                showAddButton={false}
            />

            <div className="flex-1 overflow-y-auto bg-[rgb(var(--color-bg-secondary))] p-4 md:p-8 scrollbar-thin">
                <div className="max-w-7xl mx-auto space-y-8">
                    {/* Month Stats */}
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        <div className="bg-[rgb(var(--color-bg-primary))] p-6 rounded-2xl border border-[rgb(var(--color-border-primary))]">
                            <div className="flex items-center gap-3 mb-2">
                                <Calendar size={20} className="text-[rgb(var(--color-primary))]" />
                                <p className="text-xs font-black text-[rgb(var(--color-text-tertiary))] uppercase tracking-wider">
                                    Total Tasks
                                </p>
                            </div>
                            <p className="text-3xl font-bold">{allTasks.length}</p>
                        </div>

                        <div className="bg-[rgb(var(--color-bg-primary))] p-6 rounded-2xl border border-[rgb(var(--color-border-primary))]">
                            <p className="text-xs font-black text-[rgb(var(--color-text-tertiary))] uppercase tracking-wider mb-2">
                                Completed
                            </p>
                            <p className="text-3xl font-bold text-green-500">
                                {allTasks.filter((t) => t.status === 'completed').length}
                            </p>
                        </div>

                        <div className="bg-[rgb(var(--color-bg-primary))] p-6 rounded-2xl border border-[rgb(var(--color-border-primary))]">
                            <p className="text-xs font-black text-[rgb(var(--color-text-tertiary))] uppercase tracking-wider mb-2">
                                In Progress
                            </p>
                            <p className="text-3xl font-bold text-blue-500">
                                {allTasks.filter((t) => t.status === 'in-progress').length}
                            </p>
                        </div>

                        <div className="bg-[rgb(var(--color-bg-primary))] p-6 rounded-2xl border border-[rgb(var(--color-border-primary))]">
                            <p className="text-xs font-black text-[rgb(var(--color-text-tertiary))] uppercase tracking-wider mb-2">
                                To Do
                            </p>
                            <p className="text-3xl font-bold text-yellow-500">
                                {allTasks.filter((t) => t.status === 'todo').length}
                            </p>
                        </div>
                    </div>

                    {/* Tasks by Week */}
                    {Object.keys(tasksByWeek)
                        .sort((a, b) => parseInt(a) - parseInt(b))
                        .map((weekNum) => {
                            const week = parseInt(weekNum);
                            const tasks = tasksByWeek[week];

                            return (
                                <div key={week} className="space-y-4">
                                    <div className="flex items-center justify-between">
                                        <h2 className="text-xl md:text-2xl font-bold">Week {week}</h2>
                                        <button
                                            onClick={() => router.push(`/tasks?week=${week}`)}
                                            className="text-sm font-semibold text-[rgb(var(--color-primary))] hover:underline"
                                        >
                                            View Week →
                                        </button>
                                    </div>

                                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                        {tasks.map((task) => (
                                            <TaskCard
                                                key={task.id}
                                                task={task}
                                                onClick={() => handleTaskClick(task)}
                                            />
                                        ))}
                                    </div>

                                    {tasks.length === 0 && (
                                        <div className="text-center py-12 bg-[rgb(var(--color-bg-primary))] rounded-2xl border border-[rgb(var(--color-border-primary))]">
                                            <p className="text-[rgb(var(--color-text-tertiary))]">No tasks for this week</p>
                                        </div>
                                    )}
                                </div>
                            );
                        })}

                    {allTasks.length === 0 && (
                        <div className="text-center py-20">
                            <Calendar size={64} className="mx-auto mb-4 text-[rgb(var(--color-text-tertiary))]" />
                            <p className="text-xl font-bold text-[rgb(var(--color-text-secondary))] mb-2">
                                No tasks yet
                            </p>
                            <p className="text-[rgb(var(--color-text-tertiary))]">
                                Start by creating your first task
                            </p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
