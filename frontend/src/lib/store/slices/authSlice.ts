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

