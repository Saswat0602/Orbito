/**
 * Orbito Application Constants
 */

export const APP_CONFIG = {
  NAME: 'Orbito',
  TAGLINE: 'Enterprise Engineering Workflow & Task Management',
  VERSION: '1.0.0',
  DEFAULT_THEME: 'dark',
} as const;

export const STORAGE_KEYS = {
  THEME: 'orbito_theme',
  SIDEBAR_COLLAPSED: 'orbito_sidebar_collapsed',
  LAST_WORKSPACE: 'orbito_last_workspace',
  LAST_PROJECT: 'orbito_last_project',
} as const;

export const TASK_PRIORITIES = {
  URGENT: {
    key: 'URGENT',
    label: 'Urgent',
    color: 'text-rose-400',
    bg: 'bg-rose-500/10',
    border: 'border-rose-500/20',
  },
  HIGH: {
    key: 'HIGH',
    label: 'High',
    color: 'text-amber-400',
    bg: 'bg-amber-500/10',
    border: 'border-amber-500/20',
  },
  MEDIUM: {
    key: 'MEDIUM',
    label: 'Medium',
    color: 'text-sky-400',
    bg: 'bg-sky-500/10',
    border: 'border-sky-500/20',
  },
  LOW: {
    key: 'LOW',
    label: 'Low',
    color: 'text-zinc-400',
    bg: 'bg-zinc-800',
    border: 'border-zinc-700',
  },
} as const;

export const DEFAULT_KANBAN_COLUMNS = [
  { key: 'TODO', label: 'To Do', color: 'border-zinc-700 text-zinc-300' },
  { key: 'IN_PROGRESS', label: 'In Progress', color: 'border-amber-500/60 text-amber-400' },
  { key: 'IN_REVIEW', label: 'In Review', color: 'border-sky-500/60 text-sky-400' },
  { key: 'DONE', label: 'Done', color: 'border-emerald-500/60 text-emerald-400' },
] as const;

export const PAGINATION_DEFAULTS = {
  DEFAULT_PAGE: 1,
  DEFAULT_LIMIT: 20,
  MAX_LIMIT: 100,
} as const;

