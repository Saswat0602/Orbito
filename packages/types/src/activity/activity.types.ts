import type { ActivityType } from '../enums';
import type { UserSummary } from '../users/user.types';

export interface ActivityLog {
  id: string;
  userId: string;
  organizationId?: string | null;
  workspaceId?: string | null;
  projectId?: string | null;
  taskId?: string | null;
  type: ActivityType | string;
  action?: string;
  entityType?: string;
  entityId?: string;
  description: string;
  metadata?: Record<string, unknown>;
  oldValue?: unknown;
  newValue?: unknown;
  user?: UserSummary;
  createdAt: string;
}

export interface TaskActivityType {
  id: string;
  type: ActivityType | string;
  description: string;
  entityType: string;
  entityId: string;
  oldValue?: unknown;
  newValue?: unknown;
  createdAt: string;
  user?: UserSummary;
  task?: {
    id: string;
    title: string;
    slug?: string;
  };
}

export interface ActivityApiResponse {
  activities: ActivityLog[];
  pagination: {
    currentPage: number;
    totalPages: number;
    totalCount: number;
    hasNextPage: boolean;
    hasPrevPage?: boolean;
    hasPrevious?: boolean;
  };
}
