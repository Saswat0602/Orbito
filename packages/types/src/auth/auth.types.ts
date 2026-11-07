import type { Role } from '../enums';
import type { User, UserSummary } from '../users/user.types';

export interface AuthTokens {
  accessToken: string;
  refreshToken?: string;
  expiresIn?: number;
}

export interface AuthResponse extends Partial<AuthTokens> {
  user: User;
  access_token?: string;
  refresh_token?: string;
  message?: string;
}

export interface LoginRequest {
  email: string;
  password?: string;
}

export interface RegisterRequest {
  email: string;
  password?: string;
  username?: string;
  firstName?: string;
  lastName?: string;
  invitationToken?: string;
}

export interface ResetPasswordRequest {
  token: string;
  password?: string;
  confirmPassword?: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface SetupAdminData {
  email: string;
  password?: string;
  firstName: string;
  lastName: string;
  username?: string;
}

export interface SetupStatusResponse {
  required: boolean;
  canSetup: boolean;
  message?: string;
}

export interface SetupResponse {
  success: boolean;
  message: string;
  user: User;
}

export interface AccessControlResult {
  allowed: boolean;
  reason?: string;
  requiredRole?: Role;
  actualRole?: Role;
}
