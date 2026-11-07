import type { Role, ProjectStatus, ProjectPriority, ProjectVisibility } from '../enums';
import type { UserSummary } from '../users/user.types';
import type { WorkspaceSummary } from '../workspaces/workspace.types';

export interface ProjectSummary {
  id: string;
  name: string;
  slug: string;
  key?: string;
  workspaceId: string;
}

export interface Project extends ProjectSummary {
  description?: string | null;
  status: ProjectStatus;
  priority: ProjectPriority;
  visibility: ProjectVisibility;
  startDate?: string | null;
  targetDate?: string | null;
  createdAt: string;
  updatedAt: string;
  workspace?: WorkspaceSummary;
  members?: ProjectMember[];
}

export interface ProjectMember {
  id: string;
  projectId: string;
  userId: string;
  role: Role;
  joinedAt: string;
  user?: UserSummary;
}

export interface GanttTask {
  id: string;
  name: string;
  start: string;
  end: string;
  progress: number;
  dependencies: string[];
  type: string;
  assignees: { id: string; name: string; avatar?: string }[];
}

export interface ResourceAllocation {
  userId: string;
  userName: string;
  avatar?: string;
  tasksCount: number;
  estimatedHours: number;
  loggedHours: number;
}

export interface ProjectHealthStats {
  totalTasks: number;
  completedTasks: number;
  overdueTasks: number;
  completionRate: number;
  velocityScore?: number;
}

export interface TimeRange {
  start: Date;
  end: Date;
  days: Date[];
}
