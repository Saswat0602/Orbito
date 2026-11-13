import { apiSlice } from './apiSlice';
import type { UserSummary } from '@orbito/types';
import { ENDPOINTS } from '../../api/endpoint';

export interface ActivityLogItem {
  id: string;
  type: string;
  description: string;
  userId?: string;
  projectId?: string | null;
  taskId?: string | null;
  taskSlug?: string | null;
  projectSlug?: string | null;
  workspaceSlug?: string | null;
  sprintSlug?: string | null;
  metadata?: Record<string, unknown>;
  createdAt: string;
  user?: UserSummary | { id?: string; name?: string; firstName?: string; lastName?: string; email?: string; avatar?: string | null };
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
    getActivityLogs: builder.query<
      ActivityLogItem[],
      {
        limit?: number;
        offset?: number;
        page?: number;
        entityType?: string;
        workspaceId?: string;
        organizationId?: string;
        projectId?: string;
      } | void
    >({
      query: (params) => ({
        url: ENDPOINTS.activityLog.list,
        params: params || {},
      }),
      transformResponse: (response: unknown): ActivityLogItem[] => {
        if (Array.isArray(response)) return response as ActivityLogItem[];
        if (response && typeof response === 'object') {
          const obj = response as Record<string, unknown>;
          if (Array.isArray(obj.activities)) return obj.activities as ActivityLogItem[];
          if (Array.isArray(obj.data)) return obj.data as ActivityLogItem[];
        }
        return [];
      },
      providesTags: ['ActivityLog'],
    }),
    getProjectActivityLogs: builder.query<ActivityLogItem[], string>({
      query: (projectId) => ENDPOINTS.activityLog.byProject(projectId),
      transformResponse: (response: unknown): ActivityLogItem[] => {
        if (Array.isArray(response)) return response as ActivityLogItem[];
        if (response && typeof response === 'object') {
          const obj = response as Record<string, unknown>;
          if (Array.isArray(obj.activities)) return obj.activities as ActivityLogItem[];
          if (Array.isArray(obj.data)) return obj.data as ActivityLogItem[];
        }
        return [];
      },
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

