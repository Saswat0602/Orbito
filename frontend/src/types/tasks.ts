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
