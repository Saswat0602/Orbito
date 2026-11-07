import { apiSlice } from './apiSlice';
import type { User, AuthResponse } from '@orbito/types';
import { ENDPOINTS } from '../../api/endpoint';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  username: string;
}

export const authApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    login: builder.mutation<AuthResponse, LoginRequest>({
      query: (credentials) => ({
        url: ENDPOINTS.auth.login,
        method: 'POST',
        body: credentials,
      }),
      invalidatesTags: ['Auth', 'User'],
    }),
    register: builder.mutation<User, RegisterRequest>({
      query: (userData) => ({
        url: ENDPOINTS.auth.register,
        method: 'POST',
        body: userData,
      }),
    }),
    logout: builder.mutation<void, void>({
      query: () => ({
        url: ENDPOINTS.auth.logout,
        method: 'POST',
      }),
      invalidatesTags: ['Auth', 'User'],
    }),
    getMe: builder.query<User, void>({
      query: () => ENDPOINTS.auth.me,
      providesTags: ['User'],
    }),
    getProfile: builder.query<User, void>({
      query: () => ENDPOINTS.auth.profile,
      providesTags: ['User'],
    }),
  }),
});

export const {
  useLoginMutation,
  useRegisterMutation,
  useLogoutMutation,
  useGetMeQuery,
  useGetProfileQuery,
} = authApi;
