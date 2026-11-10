/**
 * Task Power Tools and Time Tracking UI Contracts
 */
export interface TaskSubtasksProps {
  taskId: string;
  projectId: string;
  canEdit: boolean;
}

export interface TaskLabelsProps {
  taskId: string;
  canEdit: boolean;
}

export interface TaskTimeTrackingProps {
  taskId: string;
  estimatedHours?: number | null;
  canEdit: boolean;
}

export interface TaskAttachmentsProps {
  taskId: string;
  canEdit: boolean;
}

export interface CreateTimeEntryRequest {
  taskId: string;
  duration: number; // duration in seconds
  description?: string;
  date?: string;
}

export interface ActiveTimerResponse {
  active: boolean;
  taskId?: string;
  startedAt?: string;
  elapsedSeconds?: number;
}

export interface TaskDependencyItem {
  id: string;
  dependentTaskId: string;
  blockingTaskId: string;
  type: 'BLOCKS' | 'IS_BLOCKED_BY' | 'RELATES_TO';
  createdAt: string;
  blockingTask?: {
    id: string;
    title: string;
    slug?: string;
    status?: { name: string };
  };
  dependentTask?: {
    id: string;
    title: string;
    slug?: string;
    status?: { name: string };
  };
}

export interface TaskDependenciesResponse {
  dependsOn: TaskDependencyItem[];
  blocks: TaskDependencyItem[];
}

export interface CreateTaskDependencyRequest {
  dependentTaskId: string;
  blockingTaskId: string;
  type?: 'BLOCKS' | 'IS_BLOCKED_BY' | 'RELATES_TO';
}

export interface TaskDependenciesProps {
  taskId: string;
  projectId: string;
  canEdit: boolean;
}
