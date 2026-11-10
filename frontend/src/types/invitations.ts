/**
 * Workspace and Entity Invitation Contracts
 */
export interface CreateInvitationRequest {
  inviteeEmail: string;
  role: string;
  organizationId?: string;
  workspaceId?: string;
  projectId?: string;
}

export interface InvitationResponse {
  id: string;
  inviteeEmail: string;
  role: string;
  token: string;
  status: string;
  expiresAt: string;
  organizationId?: string | null;
  workspaceId?: string | null;
  projectId?: string | null;
}

export interface VerifiedInvitation {
  id: string;
  inviteeEmail: string;
  role: string;
  token: string;
  status: string;
  expiresAt: string;
  inviter?: {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
    username?: string;
  };
  organization?: {
    id: string;
    name: string;
    slug: string;
  };
  workspace?: {
    id: string;
    name: string;
    slug: string;
  };
  project?: {
    id: string;
    name: string;
  };
}

