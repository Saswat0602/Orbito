'use client';

import React from 'react';
import { Plus } from 'lucide-react';
import { ThemeToggle } from './ThemeToggle';

interface HeaderProps {
    title: string;
    subtitle?: string;
    onAddTask?: () => void;
    showAddButton?: boolean;
}

export const Header: React.FC<HeaderProps> = ({
    title,
    subtitle,
    onAddTask,
    showAddButton = true,
}) => {
    return (
        <header className="px-4 md:px-8 py-4 md:py-6 flex justify-between items-center bg-[rgb(var(--color-bg-primary))] border-b border-[rgb(var(--color-border-primary))]">
            <div className="flex-1 min-w-0">
                <h2 className="text-lg md:text-2xl font-bold truncate">
                    {title}
                </h2>
                {subtitle && (
                    <p className="text-[rgb(var(--color-text-tertiary))] text-xs md:text-sm mt-0.5 truncate">
                        {subtitle}
                    </p>
                )}
            </div>

            <div className="flex items-center gap-2 md:gap-3 ml-4">
                <ThemeToggle />

                {showAddButton && onAddTask && (
                    <button
                        onClick={onAddTask}
                        className="flex items-center gap-2 bg-[rgb(var(--color-primary))] text-white px-3 md:px-5 py-2 md:py-2.5 rounded-xl hover:bg-[rgb(var(--color-primary-hover))] transition-all shadow-lg text-sm md:text-base font-semibold whitespace-nowrap"
                    >
                        <Plus size={18} className="md:w-5 md:h-5" />
                        <span className="hidden sm:inline">Add Task</span>
                        <span className="sm:hidden">Add</span>
                    </button>
                )}
            </div>
        </header>
    );
};
