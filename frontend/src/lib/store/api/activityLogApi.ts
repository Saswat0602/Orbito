import { apiSlice } from './apiSlice';
import type { UserSummary } from '@orbito/types';
import { ENDPOINTS } from '../../api/endpoint';

export interface ActivityLogItem {
  id: string;
  type: string;
  description: string;
  userId: string;
  projectId?: string | null;
  taskId?: string | null;
  metadata?: Record<string, unknown>;
  createdAt: string;
  user?: UserSummary;
}

export const activityLogApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getActivityLogs: builder.query<ActivityLogItem[], { limit?: number; offset?: number } | void>({
      query: (params) => ({
        url: ENDPOINTS.activityLog.list,
        params: params || {},
      }),
      providesTags: ['ActivityLog'],
    }),
    getProjectActivityLogs: builder.query<ActivityLogItem[], string>({
      query: (projectId) => ENDPOINTS.activityLog.byProject(projectId),
      providesTags: ['ActivityLog'],
    }),
  }),
});

export const {
  useGetActivityLogsQuery,
  useGetProjectActivityLogsQuery,
} = activityLogApi;

