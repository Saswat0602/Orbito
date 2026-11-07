import { Role, UserStatus } from '../enums';

export interface AdminDashboardStats {
  totalUsers: number;
  totalOrganizations: number;
  totalWorkspaces: number;
  totalProjects: number;
  totalTasks: number;
  newUsersThisWeek: number;
  newOrgsThisWeek: number;
  newProjectsThisWeek: number;
  newTasksThisWeek: number;
  activeUsers: number;
}

export interface AdminUserItem {
  id: string;
  email: string;
  firstName: string | null;
  lastName: string | null;
  avatar: string | null;
  role: Role;
  status: UserStatus;
  lastLoginAt: string | null;
  createdAt: string;
  _count?: {
    organizationMembers: number;
  };
}

export interface AdminUsersResponse {
  data: AdminUserItem[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

export interface AdminOrganizationItem {
  id: string;
  name: string;
  slug: string;
  description: string | null;
  archive: boolean;
  createdAt: string;
  owner: {
    id: string;
    firstName: string | null;
    lastName: string | null;
    email: string;
  } | null;
  _count: {
    members: number;
    workspaces: number;
  };
}

export interface AdminOrganizationsResponse {
  data: AdminOrganizationItem[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

export interface QueueHealthStats {
  backend: string;
  requestedBackend: string;
  fallbackOccurred: boolean;
  redisAvailable: boolean;
  queues: string[];
  stats: {
    totalJobs: number;
    completedJobs: number;
    failedJobs: number;
    activeJobs: number;
    successRate: number;
  };
  timestamp: string;
}

