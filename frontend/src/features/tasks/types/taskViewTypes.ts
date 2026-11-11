export type GroupByField = 'none' | 'status' | 'priority' | 'assignee' | 'type' | 'dueDate';

export type SortField = 'title' | 'priority' | 'status' | 'dueDate' | 'createdAt' | 'storyPoints' | 'slug';

export type SortOrder = 'asc' | 'desc';

export interface ColumnConfig {
  id: string;
  label: string;
  visible: boolean;
  minWidth?: number;
  fixed?: boolean;
}

export const DEFAULT_COLUMNS: ColumnConfig[] = [
  { id: 'slug', label: 'Key', visible: true, minWidth: 100, fixed: true },
  { id: 'type', label: 'Type', visible: true, minWidth: 90 },
  { id: 'title', label: 'Task Title', visible: true, minWidth: 260, fixed: true },
  { id: 'status', label: 'Status', visible: true, minWidth: 120 },
  { id: 'priority', label: 'Priority', visible: true, minWidth: 100 },
  { id: 'assignees', label: 'Assignees', visible: true, minWidth: 140 },
  { id: 'sprint', label: 'Sprint', visible: true, minWidth: 120 },
  { id: 'storyPoints', label: 'Points', visible: true, minWidth: 80 },
  { id: 'dueDate', label: 'Due Date', visible: true, minWidth: 130 },
  { id: 'timeTracking', label: 'Logged Hours', visible: false, minWidth: 110 },
  { id: 'recurrence', label: 'Recurring', visible: false, minWidth: 100 },
  { id: 'createdAt', label: 'Created', visible: false, minWidth: 110 },
];

