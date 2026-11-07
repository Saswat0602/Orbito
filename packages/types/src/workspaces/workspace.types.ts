import type { Role } from '../enums';
import type { UserSummary } from '../users/user.types';
import type { OrganizationSummary } from '../organizations/organization.types';

export interface WorkspaceSummary {
  id: string;
  name: string;
  slug: string;
  organizationId: string;
}

export interface Workspace extends WorkspaceSummary {
  description?: string | null;
  isPersonal: boolean;
  parentId?: string | null;
  depth: number;
  createdAt: string;
  updatedAt: string;
  organization?: OrganizationSummary;
  children?: WorkspaceSummary[];
}

export interface WorkspaceMember {
  id: string;
  workspaceId: string;
  userId: string;
  role: Role;
  joinedAt: string;
  user?: UserSummary;
}

export interface WorkspaceTreeNode extends WorkspaceSummary {
  children: WorkspaceTreeNode[];
}

export interface CreateWorkspaceRequest {
  name: string;
  organizationId: string;
  description?: string;
  slug?: string;
}

export interface UpdateWorkspaceRequest {
  id: string;
  name?: string;
  description?: string;
}
