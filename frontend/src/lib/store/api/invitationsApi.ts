import { apiSlice } from './apiSlice';
import { ENDPOINTS } from '../../api/endpoint';

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

export const invitationsApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    verifyInvitation: builder.query<VerifiedInvitation, string>({
      query: (token) => ENDPOINTS.invitations.verify(token),
    }),

    acceptInvitation: builder.mutation<{ message: string; workspace?: { slug: string } }, string>({
      query: (token) => ({
        url: ENDPOINTS.invitations.accept(token),
        method: 'PATCH',
      }),
      invalidatesTags: ['Invitation', 'Workspace', 'Organization', 'Project'],
    }),

    declineInvitation: builder.mutation<{ message: string }, string>({
      query: (token) => ({
        url: ENDPOINTS.invitations.decline(token),
        method: 'PATCH',
      }),
      invalidatesTags: ['Invitation'],
    }),

    createInvitation: builder.mutation<InvitationResponse, CreateInvitationRequest>({
      query: (body) => ({
        url: ENDPOINTS.invitations.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Invitation'],
    }),

    getUserInvitations: builder.query<InvitationResponse[], void>({
      query: () => '/invitations/user',
      providesTags: ['Invitation'],
    }),

    getEntityInvitations: builder.query<
      InvitationResponse[],
      { entityType: 'organization' | 'workspace' | 'project'; entityId: string }
    >({
      query: ({ entityType, entityId }) => `/invitations/entity/${entityType}/${entityId}`,
      providesTags: ['Invitation'],
    }),

    resendInvitation: builder.mutation<{ message: string }, string>({
      query: (id) => ({
        url: `/invitations/${id}/resend`,
        method: 'POST',
      }),
      invalidatesTags: ['Invitation'],
    }),

    deleteInvitation: builder.mutation<{ message: string }, string>({
      query: (id) => ({
        url: ENDPOINTS.invitations.cancel(id),
        method: 'DELETE',
      }),
      invalidatesTags: ['Invitation'],
    }),
  }),
});

export const {
  useCreateInvitationMutation,
  useGetUserInvitationsQuery,
  useGetEntityInvitationsQuery,
  useResendInvitationMutation,
  useDeleteInvitationMutation,
  useVerifyInvitationQuery,
  useAcceptInvitationMutation,
  useDeclineInvitationMutation,
} = invitationsApi;

