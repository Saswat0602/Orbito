import type { TaskPriority, TaskType, ScopeType, ViewType } from '../enums';
import type { PaginationParams } from '../common/common.types';
import type { Task } from './task.types';

export type GroupByField =
  | 'none'
  | 'status'
  | 'priority'
  | 'project'
  | 'assignee'
  | 'type'
  | 'dueDate'
  | 'createdAt';

export interface TaskGroup {
  key: string;
  label: string;
  tasks: Task[];
}

export interface TaskGroupResponse {
  key: string;
  label: string;
  totalCount: number;
  tasks: Task[];
  page?: number;
}

export interface TaskReorderPayload {
  taskId?: string;
  statusId?: string;
  prevRank?: string | null;
  nextRank?: string | null;
  sprintId?: string | null;
  scopeType?: ScopeType | string;
  scopeId?: string;
  viewType?: ViewType | string;
  afterTaskId?: string | null;
  beforeTaskId?: string | null;
}

export interface TaskReorderResult {
  taskId: string;
  statusId: string;
  rank: string;
  sprintId?: string | null;
}

export interface TaskStatusPositionUpdate {
  taskId: string;
  statusId: string;
  rank: string;
}

export interface TaskQueryParams extends PaginationParams {
  projectId?: string;
  workspaceId?: string;
  organizationId?: string;
  sprintId?: string;
  statusId?: string;
  assigneeId?: string;
  priority?: TaskPriority;
  taskType?: TaskType;
  search?: string;
  isArchived?: boolean;
}

export interface TaskGroupedResponse {
  groups: {
    key: string;
    label: string;
    tasks: Task[];
  }[];
  totalCount: number;
}

export interface CreateTaskDto {
  title: string;
  description?: string;
  type?: TaskType;
  priority?: TaskPriority;
  projectId: string;
  assigneeIds?: string[];
  reporterIds?: string[];
  statusId?: string;
  sprintId?: string;
  parentTaskId?: string;
  startDate?: string;
  dueDate?: string;
  storyPoints?: number;
  originalEstimate?: number;
  customFields?: Record<string, unknown>;
  labels?: string[];
}

export interface UpdateTaskDto extends Partial<CreateTaskDto> {}

export interface TaskFilter {
  search?: string;
  assigneeIds?: string[];
  reporterIds?: string[];
  statusIds?: string[];
  labelIds?: string[];
  priorities?: TaskPriority[];
  types?: TaskType[];
  sprintIds?: string[];
  dueDateFrom?: string;
  dueDateTo?: string;
  createdFrom?: string;
  createdTo?: string;
  hasAttachments?: boolean;
  hasComments?: boolean;
}
