/**
 * @orbito/types
 * Core domain contracts shared between NestJS backend and Next.js frontend
 */

export enum Role {
  SUPER_ADMIN = 'SUPER_ADMIN',
  OWNER = 'OWNER',
  MANAGER = 'MANAGER',
  MEMBER = 'MEMBER',
  VIEWER = 'VIEWER',
}

export enum TaskPriority {
  URGENT = 'URGENT',
  HIGH = 'HIGH',
  MEDIUM = 'MEDIUM',
  LOW = 'LOW',
}

export enum ProjectStatus {
  PLANNING = 'PLANNING',
  ACTIVE = 'ACTIVE',
  ON_HOLD = 'ON_HOLD',
  COMPLETED = 'COMPLETED',
  CANCELLED = 'CANCELLED',
}

export enum ProjectVisibility {
  PRIVATE = 'PRIVATE',
  INTERNAL = 'INTERNAL',
  PUBLIC = 'PUBLIC',
}

export interface UserSummary {
  id: string;
  name: string;
  email: string;
  avatarUrl?: string | null;
  status: string;
}

export interface TaskSummary {
  id: string;
  taskNumber: number;
  slug: string;
  title: string;
  description?: string | null;
  priority: TaskPriority;
  statusId: string;
  projectId: string;
  workspaceId: string;
  rank: string;
  dueDate?: string | null;
  estimatedHours?: number | null;
  loggedHours?: number;
  assignees: UserSummary[];
  createdAt: string;
  updatedAt: string;
}

export interface WebSocketEventPayload<T = unknown> {
  event: string;
  timestamp: string;
  organizationId?: string;
  workspaceId?: string;
  projectId?: string;
  actorId: string;
  data: T;
}

export interface ApiResponse<T = unknown> {
  success: boolean;
  data: T;
  message?: string;
}

export interface PaginatedResponse<T = unknown> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}
