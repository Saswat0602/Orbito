import { TaskStatus, StatusConfig, Week } from '@/types';

export const WEEKS: Week[] = [
    { id: 1, label: 'Week 1', range: 'Jan 01 - Jan 07' },
    { id: 2, label: 'Week 2', range: 'Jan 08 - Jan 14' },
    { id: 3, label: 'Week 3', range: 'Jan 15 - Jan 21' },
    { id: 4, label: 'Week 4', range: 'Jan 22 - Jan 31' },
];

export const STATUSES: StatusConfig[] = [
    {
        id: 'todo',
        label: 'To Do',
        icon: 'Circle',
        color: 'text-slate-400',
        bg: 'bg-slate-50 dark:bg-slate-800'
    },
    {
        id: 'in-progress',
        label: 'In Progress',
        icon: 'Loader2',
        color: 'text-amber-500',
        bg: 'bg-amber-50/30 dark:bg-amber-950/30'
    },
    {
        id: 'needs-improvement',
        label: 'Improve',
        icon: 'AlertCircle',
        color: 'text-rose-500',
        bg: 'bg-rose-50/30 dark:bg-rose-950/30'
    },
    {
        id: 'completed',
        label: 'Completed',
        icon: 'CheckCircle2',
        color: 'text-emerald-500',
        bg: 'bg-emerald-50/30 dark:bg-emerald-950/30'
    }
];

export const STORAGE_KEYS = {
    TASKS: 'orbito_tasks',
    TAGS: 'orbito_tags',
    THEME: 'orbito_theme',
    FILTERS: 'orbito_filters',
} as const;

export const API_CONFIG = {
    BASE_URL: process.env.NEXT_PUBLIC_API_URL || '/api',
    TIMEOUT: 10000,
} as const;
