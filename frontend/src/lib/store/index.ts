import { configureStore } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import { apiSlice } from './api/apiSlice';
import authReducer from './slices/authSlice';
import uiReducer from './slices/uiSlice';

export const store = configureStore({
  reducer: {
    [apiSlice.reducerPath]: apiSlice.reducer,
    auth: authReducer,
    ui: uiReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: false,
    }).concat(apiSlice.middleware),
});

setupListeners(store.dispatch);

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

// Export all RTK Query slices & hooks
export * from './api/apiSlice';
export * from './api/authApi';
export * from './api/organizationsApi';
export * from './api/workspacesApi';
export * from './api/projectsApi';
export * from './api/tasksApi';
export * from './api/sprintsApi';
export * from './api/usersApi';
export * from './api/notificationsApi';
export * from './api/activityLogApi';
export * from './api/automationApi';
export * from './api/aiChatApi';
export * from './api/integrationsApi';
export * from './api/timeEntriesApi';

// Export Slice actions
export * from './slices/authSlice';
export * from './slices/uiSlice';

// Export typed Redux hooks
export * from './hooks';
