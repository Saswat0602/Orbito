import { apiSlice } from './apiSlice';
import { ENDPOINTS } from '../../api/endpoint';

export interface IntegrationStatus {
  connected: boolean;
  lastSyncAt?: string | null;
  accountEmail?: string | null;
}

export const integrationsApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getJiraStatus: builder.query<IntegrationStatus, string>({
      query: (workspaceId) => ENDPOINTS.jiraSync.workspace.status(workspaceId),
      providesTags: ['Integration'],
    }),
    syncJira: builder.mutation<{ message: string }, string>({
      query: (workspaceId) => ({
        url: ENDPOINTS.jiraSync.workspace.sync(workspaceId),
        method: 'POST',
      }),
      invalidatesTags: ['Integration', 'Task', 'Project'],
    }),
    getTrelloStatus: builder.query<IntegrationStatus, string>({
      query: (workspaceId) => ENDPOINTS.trelloSync.workspace.status(workspaceId),
      providesTags: ['Integration'],
    }),
    syncTrello: builder.mutation<{ message: string }, string>({
      query: (workspaceId) => ({
        url: ENDPOINTS.trelloSync.workspace.sync(workspaceId),
        method: 'POST',
      }),
      invalidatesTags: ['Integration', 'Task', 'Project'],
    }),
  }),
});

export const {
  useGetJiraStatusQuery,
  useSyncJiraMutation,
  useGetTrelloStatusQuery,
  useSyncTrelloMutation,
} = integrationsApi;

