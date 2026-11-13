import { apiSlice } from './apiSlice';
import type { Organization, OrganizationMember } from '@orbito/types';
import { ENDPOINTS } from '../../api/endpoint';

export interface CreateOrganizationRequest {
  name: string;
  description?: string;
  website?: string;
}

export interface UpdateOrganizationRequest {
  id: string;
  name?: string;
  description?: string;
  website?: string;
}

export interface InviteOrganizationMemberRequest {
  organizationId: string;
  email: string;
  role: string;
}

export const organizationsApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getOrganizations: builder.query<Organization[], void>({
      query: () => ENDPOINTS.organizations.list,
      transformResponse: (response: unknown): Organization[] => {
        if (Array.isArray(response)) return response as Organization[];
        if (response && typeof response === 'object') {
          const obj = response as Record<string, unknown>;
          if (Array.isArray(obj.data)) return obj.data as Organization[];
          if (Array.isArray(obj.organizations)) return obj.organizations as Organization[];
        }
        return [];
      },
      providesTags: (result) => {
        const items = Array.isArray(result) ? result : [];
        return [
          ...items.map(({ id }) => ({ type: 'Organization' as const, id })),
          { type: 'Organization', id: 'LIST' },
        ];
      },
    }),
    getOrganization: builder.query<Organization, string>({
      query: (id) => ENDPOINTS.organizations.byId(id),
      providesTags: (_result, _error, id) => [{ type: 'Organization', id }],
    }),
    createOrganization: builder.mutation<Organization, CreateOrganizationRequest>({
      query: (body) => ({
        url: ENDPOINTS.organizations.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'Organization', id: 'LIST' }],
    }),
    updateOrganization: builder.mutation<Organization, UpdateOrganizationRequest>({
      query: ({ id, ...body }) => ({
        url: ENDPOINTS.organizations.byId(id),
        method: 'PATCH',
        body,
      }),
      invalidatesTags: (_result, _error, { id }) => [
        { type: 'Organization', id },
        { type: 'Organization', id: 'LIST' },
      ],
    }),
    getOrganizationMembers: builder.query<OrganizationMember[], string>({
      query: (orgId) => ENDPOINTS.organizations.members(orgId),
      providesTags: (_result, _error, orgId) => [{ type: 'Organization', id: `MEMBERS_${orgId}` }],
    }),
    inviteOrganizationMember: builder.mutation<OrganizationMember, InviteOrganizationMemberRequest>({
      query: ({ organizationId, ...body }) => ({
        url: ENDPOINTS.organizations.members(organizationId),
        method: 'POST',
        body,
      }),
      invalidatesTags: (_result, _error, { organizationId }) => [
        { type: 'Organization', id: `MEMBERS_${organizationId}` },
      ],
    }),
    removeOrganizationMember: builder.mutation<void, { organizationId: string; memberId: string }>({
      query: ({ organizationId, memberId }) => ({
        url: ENDPOINTS.organizations.memberById(organizationId, memberId),
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { organizationId }) => [
        { type: 'Organization', id: `MEMBERS_${organizationId}` },
      ],
    }),
    getOrganizationCharts: builder.query<
      Record<string, any>,
      { orgId: string; types: string[]; workspaceId?: string; projectId?: string }
    >({
      query: ({ orgId, types, workspaceId, projectId }) => {
        const params = new URLSearchParams();
        types.forEach((t) => params.append('types', t));
        if (workspaceId) params.append('workspaceId', workspaceId);
        if (projectId) params.append('projectId', projectId);
        return {
          url: `/organizations/${orgId}/charts?${params.toString()}`,
        };
      },
      providesTags: (_res, _err, { orgId }) => [{ type: 'Organization', id: `CHARTS_${orgId}` }],
    }),
  }),
});

export const {
  useGetOrganizationsQuery,
  useGetOrganizationQuery,
  useCreateOrganizationMutation,
  useUpdateOrganizationMutation,
  useGetOrganizationMembersQuery,
  useInviteOrganizationMemberMutation,
  useRemoveOrganizationMemberMutation,
  useGetOrganizationChartsQuery,
} = organizationsApi;

