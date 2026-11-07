import type { Role, UserStatus, UserSource } from '../enums';

export interface UserSummary {
  id: string;
  email: string;
  username: string;
  firstName?: string | null;
  lastName?: string | null;
  avatarUrl?: string | null;
  avatar?: string | null;
  role: Role;
}

export interface User extends UserSummary {
  status: UserStatus;
  source: UserSource;
  bio?: string | null;
  mobileNumber?: string | null;
  timezone?: string | null;
  language?: string | null;
  emailVerified: boolean;
  preferences?: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface UpdateUserData {
  firstName?: string;
  lastName?: string;
  email?: string;
  mobileNumber?: string;
  bio?: string;
  timezone?: string;
  language?: string;
  status?: UserStatus | string;
  emailVerified?: boolean;
  avatar?: string;
  preferences?: {
    theme?: string;
    notifications?: {
      email?: boolean;
      push?: boolean;
    };
  };
  onboardInfo?: Record<string, string>;
}

export interface UpdateEmailData {
  email: string;
}
