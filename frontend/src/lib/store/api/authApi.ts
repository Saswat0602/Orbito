import { apiSlice } from './apiSlice';
import type {
  User,
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
} from '@orbito/types';
import { ENDPOINTS } from '../../api/endpoint';
import type {
  MessageResponse,
  VerifyResetTokenResponse,
  RegistrationStatusResponse,
} from '@/types';

export interface SetupRequiredResponse {
  required: boolean;
  canSetup: boolean;
  message?: string;
}

export interface SetupAdminRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  username?: string;
}

export type {
  LoginRequest,
  RegisterRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  MessageResponse,
  VerifyResetTokenResponse,
  RegistrationStatusResponse,
};

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
    register: builder.mutation<AuthResponse, RegisterRequest>({
      query: (userData) => ({
        url: ENDPOINTS.auth.register,
        method: 'POST',
        body: userData,
      }),
      invalidatesTags: ['Auth', 'User'],
    }),
    logout: builder.mutation<{ message: string }, void>({
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
    forgotPassword: builder.mutation<MessageResponse, ForgotPasswordRequest>({
      query: (body) => ({
        url: ENDPOINTS.auth.forgotPassword,
        method: 'POST',
        body,
      }),
    }),
    resetPassword: builder.mutation<MessageResponse, ResetPasswordRequest>({
      query: (body) => ({
        url: ENDPOINTS.auth.resetPassword,
        method: 'POST',
        body,
      }),
    }),
    verifyResetToken: builder.query<VerifyResetTokenResponse, string>({
      query: (token) => `${ENDPOINTS.auth.verifyResetToken}/${token}`,
    }),
    getRegistrationStatus: builder.query<RegistrationStatusResponse, void>({
      query: () => ENDPOINTS.auth.registrationStatus,
    }),
    getOidcConfig: builder.query<boolean, void>({
      query: () => '/auth/oidc/config',
    }),
    exchangeSso: builder.mutation<AuthResponse, void>({
      query: () => ({
        url: '/auth/oidc/exchange',
        method: 'POST',
      }),
      invalidatesTags: ['Auth', 'User'],
    }),
    getSetupRequired: builder.query<SetupRequiredResponse, void>({
      query: () => ENDPOINTS.auth.setupRequired,
    }),
    setupSuperAdmin: builder.mutation<AuthResponse, SetupAdminRequest>({
      query: (body) => ({
        url: ENDPOINTS.auth.setup,
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Auth', 'User'],
    }),
  }),
});

export const {
  useLoginMutation,
  useRegisterMutation,
  useLogoutMutation,
  useGetMeQuery,
  useGetProfileQuery,
  useForgotPasswordMutation,
  useResetPasswordMutation,
  useVerifyResetTokenQuery,
  useGetRegistrationStatusQuery,
  useGetOidcConfigQuery,
  useExchangeSsoMutation,
  useGetSetupRequiredQuery,
  useSetupSuperAdminMutation,
} = authApi;
