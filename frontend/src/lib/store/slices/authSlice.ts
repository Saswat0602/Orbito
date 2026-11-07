import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { User } from '@orbito/types';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  activeOrganizationId: string | null;
  activeWorkspaceId: string | null;
  activeProjectId: string | null;
}

const initialState: AuthState = {
  user: null,
  isAuthenticated: false,
  activeOrganizationId: null,
  activeWorkspaceId: null,
  activeProjectId: null,
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
    setActiveProject: (state, action: PayloadAction<string | null>) => {
      state.activeProjectId = action.payload;
    },
    logOut: (state) => {
      state.user = null;
      state.isAuthenticated = false;
      state.activeOrganizationId = null;
      state.activeWorkspaceId = null;
      state.activeProjectId = null;
    },
  },
});

export const {
  setCredentials,
  setActiveOrganization,
  setActiveWorkspace,
  setActiveProject,
  logOut,
} = authSlice.actions;

export default authSlice.reducer;

