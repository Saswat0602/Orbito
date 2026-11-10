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

export interface BulkTaskItem {
  title: string;
  description?: string;
  type?: 'TASK' | 'BUG' | 'STORY' | 'EPIC';
  priority?: 'LOW' | 'MEDIUM' | 'HIGH' | 'URGENT';
  dueDate?: string;
}

export interface BulkCreateTasksRequest {
  projectId: string;
  statusId: string;
  sprintId?: string;
  tasks: BulkTaskItem[];
}

export interface BulkCreateTasksResponse {
  created: number;
  failed: number;
  failures: Array<{
    index: number;
    title: string;
    reason: string;
  }>;
}

export interface CreateTaskShareRequest {
  taskId: string;
  expiresInDays: number;
}

export interface TaskShareResponse {
  id: string;
  token: string;
  shareUrl: string;
  expiresAt: string;
  createdAt: string;
}

export interface PublicSharedTask {
  title: string;
  description?: string;
  status: {
    name: string;
    color: string;
  };
  priority: string;
  dueDate?: string;
  assignees: Array<{
    firstName: string;
    lastName: string;
  }>;
  createdBy?: {
    firstName: string;
    lastName: string;
    avatar?: string;
  };
  attachments: Array<{
    id: string;
    fileName: string;
    fileSize: number;
    mimeType: string;
  }>;
}

