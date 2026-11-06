'use client';

import React from 'react';
import { useDroppable } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { Task, StatusConfig } from '@/types';
import { TaskCard } from './TaskCard';
import { Plus } from 'lucide-react';

interface StatusColumnProps {
    status: StatusConfig;
    tasks: Task[];
    onTaskClick: (task: Task) => void;
    onAddTask: (status: string) => void;
}

export const StatusColumn: React.FC<StatusColumnProps> = ({
    status,
    tasks,
    onTaskClick,
    onAddTask,
}) => {
    const { setNodeRef } = useDroppable({
        id: status.id,
    });

    const taskIds = tasks.map((task) => task.id);

    return (
        <div className="flex flex-col min-w-[280px] md:min-w-[320px] h-full">
            {/* Column Header */}
            <div className="flex items-center justify-between mb-4 md:mb-5 px-2 md:px-3">
                <div className="flex items-center gap-2 md:gap-3">
                    <div className={`p-1.5 md:p-2 rounded-lg ${status.bg} ${status.color}`}>
                        <div className="w-3 h-3 md:w-4 md:h-4" />
                    </div>
                    <h4 className="font-black uppercase text-[10px] md:text-[11px] tracking-widest">
                        {status.label}
                    </h4>
                </div>
                <span className="bg-[rgb(var(--color-bg-primary))] border border-[rgb(var(--color-border-primary))] text-[rgb(var(--color-text-secondary))] text-[10px] font-bold px-2 py-0.5 rounded-full shadow-sm">
                    {tasks.length}
                </span>
            </div>

            {/* Droppable Area */}
            <div
                ref={setNodeRef}
                className={`flex-1 space-y-3 md:space-y-4 ${status.bg} p-3 md:p-4 rounded-3xl border border-[rgb(var(--color-border-primary))] overflow-y-auto scrollbar-thin`}
            >
                <SortableContext items={taskIds} strategy={verticalListSortingStrategy}>
                    {tasks.map((task) => (
                        <TaskCard
                            key={task.id}
                            task={task}
                            onClick={() => onTaskClick(task)}
                        />
                    ))}
                </SortableContext>

                {/* Add Task Button */}
                <button
                    onClick={() => onAddTask(status.id)}
                    className="w-full py-3 md:py-4 border-2 border-dashed border-[rgb(var(--color-border-primary))] rounded-2xl text-[rgb(var(--color-text-tertiary))] hover:text-[rgb(var(--color-primary))] hover:border-[rgb(var(--color-primary))] hover:bg-[rgb(var(--color-bg-primary))] transition-all flex items-center justify-center gap-2 text-xs font-black uppercase tracking-widest"
                >
                    <Plus size={16} /> New Entry
                </button>
            </div>
        </div>
    );
};
