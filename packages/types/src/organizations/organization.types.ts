import type { Role } from '../enums';
import type { UserSummary } from '../users/user.types';

export interface OrganizationSummary {
  id: string;
  name: string;
  slug: string;
  logoUrl?: string | null;
}

export interface Organization extends OrganizationSummary {
  description?: string | null;
  ownerId: string;
  isPersonal: boolean;
  maxWorkspaces: number;
  maxMembers: number;
  plan: string;
  createdAt: string;
  updatedAt: string;
  owner?: UserSummary;
}

export interface OrganizationMember {
  id: string;
  organizationId: string;
  userId: string;
  role: Role;
  joinedAt: string;
  user?: UserSummary;
}

export interface CreateOrganizationRequest {
  name: string;
  slug?: string;
  description?: string;
}

export interface UpdateOrganizationRequest {
  name?: string;
  description?: string;
  logoUrl?: string;
}
