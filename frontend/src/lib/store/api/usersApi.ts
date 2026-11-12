import { apiSlice } from './apiSlice';
import type { User, UserSummary } from '@orbito/types';
import { ENDPOINTS } from '../../api/endpoint';

export interface UpdateProfileRequest {
  firstName?: string;
  lastName?: string;
  bio?: string;
  mobileNumber?: string;
  language?: string;
  timezone?: string;
  avatar?: string;
  jobTitle?: string;
  department?: string;
  preferences?: Record<string, unknown>;
}

export interface ChangePasswordRequest {
  currentPassword?: string;
  newPassword?: string;
}

export const usersApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getUsers: builder.query<User[], { search?: string; limit?: number } | void>({
      query: (params) => ({
        url: ENDPOINTS.users.list,
        params: params || {},
      }),
      providesTags: ['User'],
    }),
    searchUsers: builder.query<UserSummary[], string>({
      query: (q) => ({
        url: ENDPOINTS.users.search,
        params: { q },
      }),
    }),
    getUserById: builder.query<User, string>({
      query: (id) => ENDPOINTS.users.byId(id),
      providesTags: (_result, _error, id) => [{ type: 'User', id }],
    }),
    updateProfile: builder.mutation<User, UpdateProfileRequest>({
      query: (body) => ({
        url: ENDPOINTS.users.profile,
        method: 'PATCH',
        body,
      }),
      invalidatesTags: ['User'],
    }),
    changePassword: builder.mutation<{ message: string }, ChangePasswordRequest>({
      query: (body) => ({
        url: ENDPOINTS.users.changePassword,
        method: 'POST',
        body,
      }),
    }),
  }),
});

export const {
  useGetUsersQuery,
  useSearchUsersQuery,
  useGetUserByIdQuery,
  useUpdateProfileMutation,
  useChangePasswordMutation,
} = usersApi;

