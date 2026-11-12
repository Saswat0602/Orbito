import { apiSlice } from './apiSlice';
import type {
  AdminDashboardStats,
  AdminUsersResponse,
  AdminOrganizationsResponse,
  QueueHealthStats,
  Role,
  UserStatus,
} from '@orbito/types';
import { ENDPOINTS } from '../../api/endpoint';

export interface GetAdminUsersParams {
  page?: number;
  limit?: number;
  search?: string;
  status?: UserStatus;
  role?: Role;
}

export interface GetAdminOrganizationsParams {
  page?: number;
  limit?: number;
  search?: string;
}

export interface AdminOrganizationDetail {
  id: string;
  name: string;
  slug: string;
  description?: string | null;
  archive?: boolean;
  createdAt: string;
  updatedAt: string;
  owner: {
    id: string;
    firstName: string | null;
    lastName: string | null;
    email: string;
  } | null;
  members: Array<{
    role: string;
    createdAt: string;
    user: {
      id: string;
      firstName: string | null;
      lastName: string | null;
      email: string;
      avatar: string | null;
      status: string;
    };
  }>;
  workspaces: Array<{
    id: string;
    name: string;
    slug: string;
    _count?: {
      projects: number;
      members: number;
    };
  }>;
  _count?: {
    members: number;
    workspaces: number;
  };
}

export interface AdminSystemConfigResponse {
  settings: Array<{
    id: string;
    key: string;
    value: string;
    description?: string;
    category?: string;
    isEncrypted?: boolean;
    updatedAt: string;
  }>;
  smtpSources?: {
    hostSource: string;
    portSource: string;
    userSource: string;
    passSource: string;
  };
}

export interface SaveAdminConfigPayload {
  settings: Array<{
    key: string;
    value: string;
    description?: string;
    category?: string;
    isEncrypted?: boolean;
  }>;
}

export const adminApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getAdminDashboardStats: builder.query<AdminDashboardStats, void>({
      query: () => ENDPOINTS.admin.dashboard,
      providesTags: ['Admin'],
    }),

    getAdminUsers: builder.query<AdminUsersResponse, GetAdminUsersParams | void>({
      query: (params) => ({
        url: ENDPOINTS.admin.users,
        params: params || {},
      }),
      providesTags: ['Admin', 'User'],
    }),

    updateAdminUserRole: builder.mutation<
      { id: string; email: string; role: Role },
      { id: string; role: Role }
    >({
      query: ({ id, role }) => ({
        url: ENDPOINTS.admin.updateUserRole(id),
        method: 'PATCH',
        body: { role },
      }),
      invalidatesTags: ['Admin', 'User'],
    }),

    updateAdminUserStatus: builder.mutation<
      { id: string; email: string; status: UserStatus },
      { id: string; status: UserStatus }
    >({
      query: ({ id, status }) => ({
        url: ENDPOINTS.admin.updateUserStatus(id),
        method: 'PATCH',
        body: { status },
      }),
      invalidatesTags: ['Admin', 'User'],
    }),

    deleteAdminUser: builder.mutation<{ success: boolean; message: string }, string>({
      query: (id) => ({
        url: ENDPOINTS.admin.deleteUser(id),
        method: 'DELETE',
      }),
      invalidatesTags: ['Admin', 'User'],
    }),

    resetAdminUserPassword: builder.mutation<
      { success: boolean; resetLink?: string; message: string },
      string
    >({
      query: (id) => ({
        url: ENDPOINTS.admin.resetUserPassword(id),
        method: 'POST',
      }),
    }),

    getAdminOrganizations: builder.query<
      AdminOrganizationsResponse,
      GetAdminOrganizationsParams | void
    >({
      query: (params) => ({
        url: ENDPOINTS.admin.organizations,
        params: params || {},
      }),
      providesTags: ['Admin', 'Organization'],
    }),

    getAdminOrganizationDetail: builder.query<AdminOrganizationDetail, string>({
      query: (id) => ENDPOINTS.admin.organizationById(id),
      providesTags: (_result, _error, id) => [{ type: 'Organization', id }],
    }),

    deleteAdminOrganization: builder.mutation<{ success: boolean; message: string }, string>({
      query: (id) => ({
        url: ENDPOINTS.admin.organizationById(id),
        method: 'DELETE',
      }),
      invalidatesTags: ['Admin', 'Organization'],
    }),

    toggleAdminOrgArchive: builder.mutation<
      { id: string; name: string; archive: boolean },
      string
    >({
      query: (id) => ({
        url: ENDPOINTS.admin.toggleOrganizationArchive(id),
        method: 'PATCH',
      }),
      invalidatesTags: ['Admin', 'Organization'],
    }),

    transferAdminOrgOwnership: builder.mutation<
      { id: string; name: string; ownerId: string },
      { id: string; newOwnerId: string }
    >({
      query: ({ id, newOwnerId }) => ({
        url: ENDPOINTS.admin.transferOrganizationOwnership(id),
        method: 'PATCH',
        body: { newOwnerId },
      }),
      invalidatesTags: ['Admin', 'Organization'],
    }),

    getAdminConfig: builder.query<AdminSystemConfigResponse, { category?: string } | void>({
      query: (params) => ({
        url: ENDPOINTS.admin.config,
        params: params || {},
      }),
      providesTags: ['Admin'],
    }),

    saveAdminConfig: builder.mutation<{ success: boolean }, SaveAdminConfigPayload>({
      query: (body) => ({
        url: ENDPOINTS.admin.config,
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Admin'],
    }),

    testAdminSmtp: builder.mutation<{ success: boolean; message: string }, void>({
      query: () => ({
        url: ENDPOINTS.admin.testSmtp,
        method: 'POST',
      }),
    }),

    getQueueHealth: builder.query<QueueHealthStats, void>({
      query: () => ENDPOINTS.queues.health,
      providesTags: ['Queue'],
    }),
  }),
});

export const {
  useGetAdminDashboardStatsQuery,
  useGetAdminUsersQuery,
  useUpdateAdminUserRoleMutation,
  useUpdateAdminUserStatusMutation,
  useDeleteAdminUserMutation,
  useResetAdminUserPasswordMutation,
  useGetAdminOrganizationsQuery,
  useGetAdminOrganizationDetailQuery,
  useDeleteAdminOrganizationMutation,
  useToggleAdminOrgArchiveMutation,
  useTransferAdminOrgOwnershipMutation,
  useGetAdminConfigQuery,
  useSaveAdminConfigMutation,
  useTestAdminSmtpMutation,
  useGetQueueHealthQuery,
} = adminApi;

