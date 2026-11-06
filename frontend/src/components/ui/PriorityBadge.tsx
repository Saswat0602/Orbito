'use client';

import React from 'react';

interface PriorityBadgeProps {
    priority: 'low' | 'medium' | 'high';
    size?: 'sm' | 'md' | 'lg';
}

const priorityConfig = {
    high: {
        label: 'High',
        color: 'bg-red-500',
        lightBg: 'bg-red-50 dark:bg-red-950',
        textColor: 'text-red-700 dark:text-red-300',
        borderColor: 'border-red-200 dark:border-red-800',
    },
    medium: {
        label: 'Medium',
        color: 'bg-yellow-500',
        lightBg: 'bg-yellow-50 dark:bg-yellow-950',
        textColor: 'text-yellow-700 dark:text-yellow-300',
        borderColor: 'border-yellow-200 dark:border-yellow-800',
    },
    low: {
        label: 'Low',
        color: 'bg-blue-500',
        lightBg: 'bg-blue-50 dark:bg-blue-950',
        textColor: 'text-blue-700 dark:text-blue-300',
        borderColor: 'border-blue-200 dark:border-blue-800',
    },
};

export const PriorityBadge: React.FC<PriorityBadgeProps> = ({ priority, size = 'md' }) => {
    const config = priorityConfig[priority];

    const sizeClasses = {
        sm: 'px-2 py-0.5 text-[10px]',
        md: 'px-2.5 py-1 text-xs',
        lg: 'px-3 py-1.5 text-sm',
    };

    return (
        <span
            className={`inline-flex items-center gap-1.5 ${sizeClasses[size]} ${config.lightBg} ${config.textColor} border ${config.borderColor} rounded-full font-bold uppercase tracking-wide`}
        >
            <span className={`w-1.5 h-1.5 rounded-full ${config.color}`} />
            {config.label}
        </span>
    );
};

export const PriorityDot: React.FC<{ priority: 'low' | 'medium' | 'high' }> = ({ priority }) => {
    const config = priorityConfig[priority];

    return (
        <div className={`w-3 h-3 rounded-full ${config.color} shadow-sm`} title={`${config.label} Priority`} />
    );
};
