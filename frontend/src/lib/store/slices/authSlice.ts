import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { User } from '@orbito/types';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  activeOrganizationId: string | null;
  activeWorkspaceId: string | null;
  activeWorkspaceSlug: string | null;
  activeProjectId: string | null;
  activeProjectSlug: string | null;
}

const initialState: AuthState = {
  user: null,
  isAuthenticated: false,
  activeOrganizationId: null,
  activeWorkspaceId: null,
  activeWorkspaceSlug: null,
  activeProjectId: null,
  activeProjectSlug: null,
};

export const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setCredentials: (
      state,
      action: PayloadAction<{ user: User; accessToken?: string; refreshToken?: string }>
    ) => {
      state.user = action.payload.user;
      state.isAuthenticated = true;
      if (typeof window !== 'undefined') {
        if (action.payload.accessToken) {
          localStorage.setItem('access_token', action.payload.accessToken);
          document.cookie = `access_token=${action.payload.accessToken}; Path=/; SameSite=Lax; Max-Age=3600`;
        }
        if (action.payload.refreshToken) {
          localStorage.setItem('refresh_token', action.payload.refreshToken);
          document.cookie = `refresh_token=${action.payload.refreshToken}; Path=/; SameSite=Lax; Max-Age=604800`;
        }
        localStorage.setItem('orbito_user', JSON.stringify(action.payload.user));
      }
    },
    setActiveOrganization: (state, action: PayloadAction<string | null>) => {
      state.activeOrganizationId = action.payload;
    },
    setActiveWorkspace: (state, action: PayloadAction<string | null>) => {
      state.activeWorkspaceId = action.payload;
    },
    setActiveWorkspaceSlug: (state, action: PayloadAction<string | null>) => {
      state.activeWorkspaceSlug = action.payload;
    },
    setActiveProject: (state, action: PayloadAction<string | null>) => {
      state.activeProjectId = action.payload;
    },
    setActiveProjectSlug: (state, action: PayloadAction<string | null>) => {
      state.activeProjectSlug = action.payload;
    },
    logOut: (state) => {
      state.user = null;
      state.isAuthenticated = false;
      state.activeOrganizationId = null;
      state.activeWorkspaceId = null;
      state.activeWorkspaceSlug = null;
      state.activeProjectId = null;
      state.activeProjectSlug = null;
      if (typeof window !== 'undefined') {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('orbito_user');
        document.cookie = 'access_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
        document.cookie = 'refresh_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
      }
    },
  },
});

export const {
  setCredentials,
  setActiveOrganization,
  setActiveWorkspace,
  setActiveWorkspaceSlug,
  setActiveProject,
  setActiveProjectSlug,
  logOut,
} = authSlice.actions;

export default authSlice.reducer;

