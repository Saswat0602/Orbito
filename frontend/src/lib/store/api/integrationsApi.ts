import { apiSlice } from './apiSlice';
import { ENDPOINTS } from '../../api/endpoint';
import type {
  IntegrationStatus,
  JiraWorkspaceStatus,
  TrelloWorkspaceStatus,
} from '@/types';

export type { IntegrationStatus, JiraWorkspaceStatus, TrelloWorkspaceStatus };

export interface ConnectJiraWorkspaceDto {
  jiraSiteUrl: string;
  jiraEmail: string;
  jiraApiToken: string;
}

export interface JiraProjectItem {
  id: string;
  key: string;
  name: string;
  projectTypeKey?: string;
  avatarUrls?: Record<string, string>;
}

export interface JiraStatusItem {
  id: string;
  name: string;
  statusCategory?: {
    id: number;
    key: string;
    name: string;
    colorName?: string;
  };
}

export interface ConnectTrelloWorkspaceDto {
  trelloApiKey: string;
  trelloToken: string;
  trelloWorkspaceId?: string;
}

export interface TrelloBoardItem {
  id: string;
  name: string;
  desc?: string;
  closed?: boolean;
  url?: string;
}

export const integrationsApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getJiraStatus: builder.query<JiraWorkspaceStatus | null, string>({
      query: (workspaceId) => ENDPOINTS.jiraSync.workspace.status(workspaceId),
      providesTags: ['Integration'],
      transformResponse: (response: JiraWorkspaceStatus) => response,
      transformErrorResponse: (response: { status: number }) => {
        if (response.status === 404) return null;
        return response;
      },
    }),

    connectJiraWorkspace: builder.mutation<
      JiraWorkspaceStatus,
      { workspaceId: string; data: ConnectJiraWorkspaceDto }
    >({
      query: ({ workspaceId, data }) => ({
        url: ENDPOINTS.jiraSync.workspace.connect(workspaceId),
        method: 'POST',
        body: data,
      }),
      invalidatesTags: ['Integration'],
    }),

    getJiraProjects: builder.query<JiraProjectItem[], string>({
      query: (workspaceId) => ENDPOINTS.jiraSync.workspace.projects(workspaceId),
      transformResponse: (response: JiraProjectItem[] | unknown) => {
        if (Array.isArray(response)) return response;
        return [];
      },
      providesTags: ['Integration'],
    }),

    getJiraProjectStatuses: builder.query<
      JiraStatusItem[],
      { workspaceId: string; projectKey: string }
    >({
      query: ({ workspaceId, projectKey }) =>
        ENDPOINTS.jiraSync.workspace.projectStatuses(workspaceId, projectKey),
      transformResponse: (response: JiraStatusItem[] | unknown) => {
        if (Array.isArray(response)) return response;
        return [];
      },
    }),

    importJiraProjects: builder.mutation<
      { imported: number },
      {
        workspaceId: string;
        projects: (string | { key: string; statusMappings?: Record<string, string> })[];
      }
    >({
      query: ({ workspaceId, projects }) => ({
        url: ENDPOINTS.jiraSync.workspace.import(workspaceId),
        method: 'POST',
        body: { projects },
      }),
      invalidatesTags: ['Integration', 'Project', 'Task'],
    }),

    syncAllJira: builder.mutation<{ message: string }, string>({
      query: (workspaceId) => ({
        url: ENDPOINTS.jiraSync.workspace.syncAll(workspaceId),
        method: 'POST',
      }),
      invalidatesTags: ['Integration', 'Task', 'Project'],
    }),

    disconnectJiraWorkspace: builder.mutation<void, string>({
      query: (workspaceId) => ({
        url: ENDPOINTS.jiraSync.workspace.disconnect(workspaceId),
        method: 'DELETE',
      }),
      invalidatesTags: ['Integration'],
    }),

    getTrelloStatus: builder.query<TrelloWorkspaceStatus | null, string>({
      query: (workspaceId) => ENDPOINTS.trelloSync.workspace.status(workspaceId),
      providesTags: ['Integration'],
      transformResponse: (response: TrelloWorkspaceStatus) => response,
      transformErrorResponse: (response: { status: number }) => {
        if (response.status === 404) return null;
        return response;
      },
    }),

    connectTrelloWorkspace: builder.mutation<
      TrelloWorkspaceStatus,
      { workspaceId: string; data: ConnectTrelloWorkspaceDto }
    >({
      query: ({ workspaceId, data }) => ({
        url: ENDPOINTS.trelloSync.workspace.connect(workspaceId),
        method: 'POST',
        body: data,
      }),
      invalidatesTags: ['Integration'],
    }),

    getTrelloBoards: builder.query<TrelloBoardItem[], string>({
      query: (workspaceId) => ENDPOINTS.trelloSync.workspace.boards(workspaceId),
      transformResponse: (response: TrelloBoardItem[] | unknown) => {
        if (Array.isArray(response)) return response;
        return [];
      },
      providesTags: ['Integration'],
    }),

    importTrelloBoards: builder.mutation<
      { imported: number },
      { workspaceId: string; boardIds: string[] }
    >({
      query: ({ workspaceId, boardIds }) => ({
        url: ENDPOINTS.trelloSync.workspace.import(workspaceId),
        method: 'POST',
        body: { boardIds },
      }),
      invalidatesTags: ['Integration', 'Project', 'Task'],
    }),

    syncAllTrello: builder.mutation<{ message: string }, string>({
      query: (workspaceId) => ({
        url: ENDPOINTS.trelloSync.workspace.syncAll(workspaceId),
        method: 'POST',
      }),
      invalidatesTags: ['Integration', 'Task', 'Project'],
    }),

    disconnectTrelloWorkspace: builder.mutation<void, string>({
      query: (workspaceId) => ({
        url: ENDPOINTS.trelloSync.workspace.disconnect(workspaceId),
        method: 'DELETE',
      }),
      invalidatesTags: ['Integration'],
    }),
  }),
});

export const {
  useGetJiraStatusQuery,
  useConnectJiraWorkspaceMutation,
  useGetJiraProjectsQuery,
  useGetJiraProjectStatusesQuery,
  useImportJiraProjectsMutation,
  useSyncAllJiraMutation,
  useDisconnectJiraWorkspaceMutation,
  useGetTrelloStatusQuery,
  useConnectTrelloWorkspaceMutation,
  useGetTrelloBoardsQuery,
  useImportTrelloBoardsMutation,
  useSyncAllTrelloMutation,
  useDisconnectTrelloWorkspaceMutation,
} = integrationsApi;
