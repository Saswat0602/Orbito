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

export interface TaskActivitiesResponse {
  activities: ActivityLogItem[];
  pagination?: {
    currentPage: number;
    totalPages: number;
    totalCount: number;
    hasNextPage: boolean;
    hasPrevPage: boolean;
  };
}

export const activityLogApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getActivityLogs: builder.query<ActivityLogItem[], { limit?: number; offset?: number; entityType?: string } | void>({
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
    getTaskActivityLogs: builder.query<
      TaskActivitiesResponse,
      { taskId: string; limit?: number; page?: number } | string
    >({
      query: (arg) => {
        const taskId = typeof arg === 'string' ? arg : arg.taskId;
        const params = typeof arg === 'string' ? {} : { limit: arg.limit, page: arg.page };
        return {
          url: ENDPOINTS.activityLog.taskActivities(taskId),
          params,
        };
      },
      providesTags: (_res, _err, arg) => [
        { type: 'ActivityLog', id: typeof arg === 'string' ? arg : arg.taskId },
      ],
    }),
  }),
});

export const {
  useGetActivityLogsQuery,
  useGetProjectActivityLogsQuery,
  useGetTaskActivityLogsQuery,
} = activityLogApi;

