import type { StatusCategory } from '../enums';

export interface TaskStatus {
  id: string;
  name: string;
  color?: string | null;
  category: StatusCategory | string;
  position: number;
  isDefault: boolean;
  order?: number;
  workflowId?: string;
  createdAt?: string;
  updatedAt?: string;
  _count?: {
    tasks: number;
  };
}

export interface StatusTransition {
  id: string;
  workflowId: string;
  fromStatusId: string;
  toStatusId: string;
}

export interface WorkflowStage {
  id: string;
  name: string;
  description?: string;
  color?: string;
  order: number;
  workflowId: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface WorkflowStats {
  totalStages: number;
  totalProjects: number;
  totalTasks: number;
  activeProjects: number;
  completedTasks: number;
  utilizationRate: number;
}

export interface Workflow {
  id: string;
  name: string;
  description?: string | null;
  organizationId: string;
  isDefault: boolean;
  statuses?: TaskStatus[];
  transitions?: StatusTransition[];
  createdAt: string;
  updatedAt: string;
}

export interface CreateTaskStatusDto {
  name: string;
  color: string;
  category: StatusCategory | 'TODO' | 'IN_PROGRESS' | 'DONE';
  position?: number;
  workflowId: string;
}

export interface UpdateTaskStatusDto {
  name?: string;
  color?: string;
  category?: StatusCategory | 'TODO' | 'IN_PROGRESS' | 'DONE';
  position?: number;
  isDefault?: boolean;
  workflowId?: string;
}
