import { apiSlice } from './apiSlice';
import type {
  Workspace,
  WorkspaceMember,
  CreateWorkspaceRequest,
  UpdateWorkspaceRequest,
} from '@orbito/types';
import { ENDPOINTS } from '../../api/endpoint';

export type {
  CreateWorkspaceRequest,
  UpdateWorkspaceRequest,
};

export const workspacesApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getWorkspaces: builder.query<Workspace[], { organizationId?: string } | void>({
      query: (params) => ({
        url: ENDPOINTS.workspaces.list,
        params: params || {},
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Workspace' as const, id })),
              { type: 'Workspace', id: 'LIST' },
            ]
          : [{ type: 'Workspace', id: 'LIST' }],
    }),
    getWorkspace: builder.query<Workspace, string>({
      query: (id) => ENDPOINTS.workspaces.byId(id),
      providesTags: (_result, _error, id) => [{ type: 'Workspace', id }],
    }),
    getWorkspaceBySlug: builder.query<Workspace, string>({
      query: (slug) => ENDPOINTS.workspaces.bySlug(slug),
      providesTags: (result) => (result ? [{ type: 'Workspace', id: result.id }] : ['Workspace']),
    }),
    createWorkspace: builder.mutation<Workspace, CreateWorkspaceRequest>({
      query: (body) => ({
        url: ENDPOINTS.workspaces.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'Workspace', id: 'LIST' }],
    }),
    updateWorkspace: builder.mutation<Workspace, UpdateWorkspaceRequest>({
      query: ({ id, ...body }) => ({
        url: ENDPOINTS.workspaces.byId(id),
        method: 'PATCH',
        body,
      }),
      invalidatesTags: (_result, _error, { id }) => [
        { type: 'Workspace', id },
        { type: 'Workspace', id: 'LIST' },
      ],
    }),
    deleteWorkspace: builder.mutation<void, string>({
      query: (id) => ({
        url: ENDPOINTS.workspaces.byId(id),
        method: 'DELETE',
      }),
      invalidatesTags: [{ type: 'Workspace', id: 'LIST' }],
    }),
    getWorkspaceMembers: builder.query<WorkspaceMember[], string>({
      query: (workspaceId) => ENDPOINTS.workspaces.members(workspaceId),
      transformResponse: (response: any) => (Array.isArray(response) ? response : response?.data || []),
      providesTags: (_result, _error, workspaceId) => [
        { type: 'Workspace', id: `MEMBERS_${workspaceId}` },
      ],
    }),
    addWorkspaceMember: builder.mutation<
      WorkspaceMember,
      { workspaceId: string; email: string; role: string }
    >({
      query: (body) => ({
        url: ENDPOINTS.workspaceMembers.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: (_result, _error, { workspaceId }) => [
        { type: 'Workspace', id: `MEMBERS_${workspaceId}` },
      ],
    }),
    removeWorkspaceMember: builder.mutation<
      void,
      { workspaceId: string; memberId: string }
    >({
      query: ({ memberId }) => ({
        url: ENDPOINTS.workspaceMembers.byId(memberId),
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, { workspaceId }) => [
        { type: 'Workspace', id: `MEMBERS_${workspaceId}` },
      ],
    }),
  }),
});

export const {
  useGetWorkspacesQuery,
  useGetWorkspaceQuery,
  useGetWorkspaceBySlugQuery,
  useCreateWorkspaceMutation,
  useUpdateWorkspaceMutation,
  useDeleteWorkspaceMutation,
  useGetWorkspaceMembersQuery,
  useAddWorkspaceMemberMutation,
  useRemoveWorkspaceMemberMutation,
} = workspacesApi;

