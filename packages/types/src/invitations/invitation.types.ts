import type { InvitationStatus } from '../enums';
import type { UserSummary } from '../users/user.types';
import type { OrganizationSummary } from '../organizations/organization.types';
import type { WorkspaceSummary } from '../workspaces/workspace.types';
import type { ProjectSummary } from '../projects/project.types';

export type EntityType = 'organization' | 'workspace' | 'project';

export interface Invitation {
  id: string;
  email: string;
  entityType: EntityType;
  entityId: string;
  role: string;
  status: InvitationStatus;
  token: string;
  invitedBy: string;
  invitedAt: string;
  expiresAt: string;
  respondedAt?: string | null;
  createdAt: string;
  updatedAt: string;
  inviter?: UserSummary;
  organization?: OrganizationSummary;
  workspace?: WorkspaceSummary;
  project?: ProjectSummary;
  organizationId?: string;
  workspaceId?: string;
  projectId?: string;
}

export interface CreateInvitationData {
  inviteeEmail: string;
  organizationId?: string;
  workspaceId?: string;
  projectId?: string;
  projectIds?: string[];
  role: string;
}

export interface InvitationFilters {
  status?: InvitationStatus;
  entityType?: EntityType;
  entityId?: string;
}

export interface AcceptInvitationResponse {
  success: boolean;
  message: string;
  invitation: Invitation;
}

export interface DeclineInvitationResponse {
  success: boolean;
  message: string;
  invitation: Invitation;
}

export interface VerifyInvitationResponse {
  success: boolean;
  message: string;
  invitation: Invitation;
  isValid: boolean;
  isExpired: boolean;
  canRespond: boolean;
  inviteeExists: boolean;
}
