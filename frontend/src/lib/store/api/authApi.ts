import { apiSlice } from './apiSlice';
import type { User, AuthResponse } from '@orbito/types';

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
        url: '/auth/login',
        method: 'POST',
        body: credentials,
      }),
      invalidatesTags: ['Auth', 'User'],
      async onQueryStarted(_arg, { queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;
          if (typeof window !== 'undefined') {
            if (data.access_token) {
              localStorage.setItem('orbito_access_token', data.access_token);
            }
            if (data.refresh_token) {
              localStorage.setItem('orbito_refresh_token', data.refresh_token);
            }
          }
        } catch {
          // Handled by UI
        }
      },
    }),
    register: builder.mutation<User, RegisterRequest>({
      query: (userData) => ({
        url: '/auth/register',
        method: 'POST',
        body: userData,
      }),
    }),
    logout: builder.mutation<void, void>({
      query: () => ({
        url: '/auth/logout',
        method: 'POST',
      }),
      invalidatesTags: ['Auth', 'User'],
      async onQueryStarted(_arg, { queryFulfilled }) {
        try {
          await queryFulfilled;
        } finally {
          if (typeof window !== 'undefined') {
            localStorage.removeItem('orbito_access_token');
            localStorage.removeItem('orbito_refresh_token');
          }
        }
      },
    }),
    getMe: builder.query<User, void>({
      query: () => '/auth/me',
      providesTags: ['User'],
    }),
  }),
});

export const { useLoginMutation, useRegisterMutation, useLogoutMutation, useGetMeQuery } = authApi;

