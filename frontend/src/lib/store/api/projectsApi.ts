import { apiSlice } from './apiSlice';
import type { Project } from '@orbito/types';
import { ENDPOINTS } from '../../api/endpoint';

export interface GanttTask {
  id: string;
  name: string;
  start: string;
  end: string;
  progress: number;
  dependencies: string[];
  type: string;
  assignees: { id: string; name: string; avatar?: string }[];
}

export interface ResourceAllocation {
  userId: string;
  userName: string;
  avatar?: string;
  tasksCount: number;
  estimatedHours: number;
  loggedHours: number;
}

export const projectsApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getProjects: builder.query<Project[], { workspaceId?: string } | void>({
      query: (params) => ({
        url: ENDPOINTS.projects.list,
        params: params || {},
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Project' as const, id })),
              { type: 'Project', id: 'LIST' },
            ]
          : [{ type: 'Project', id: 'LIST' }],
    }),

    getProject: builder.query<Project, string>({
      query: (idOrSlug) => ENDPOINTS.projects.byId(idOrSlug),
      providesTags: (_result, _err, idOrSlug) => [{ type: 'Project', id: idOrSlug }],
    }),

    getProjectBySlug: builder.query<Project, string>({
      query: (slug) => ENDPOINTS.projects.bySlug(slug),
      providesTags: (result) => (result ? [{ type: 'Project', id: result.id }] : ['Project']),
    }),

    getProjectGantt: builder.query<GanttTask[], string>({
      query: (projectId) => ENDPOINTS.gantt.project(projectId),
      providesTags: (_result, _err, projectId) => [{ type: 'Project', id: projectId }],
    }),

    getProjectResources: builder.query<ResourceAllocation[], string>({
      query: (projectId) => ENDPOINTS.gantt.resources(projectId),
      providesTags: (_result, _err, projectId) => [{ type: 'Project', id: projectId }],
    }),

    createProject: builder.mutation<Project, Partial<Project>>({
      query: (project) => ({
        url: ENDPOINTS.projects.create,
        method: 'POST',
        body: project,
      }),
      invalidatesTags: [{ type: 'Project', id: 'LIST' }],
    }),

    updateProject: builder.mutation<Project, { id: string; updates: Partial<Project> }>({
      query: ({ id, updates }) => ({
        url: ENDPOINTS.projects.byId(id),
        method: 'PATCH',
        body: updates,
      }),
      invalidatesTags: (_result, _err, { id }) => [
        { type: 'Project', id },
        { type: 'Project', id: 'LIST' },
      ],
    }),
  }),
});

export const {
  useGetProjectsQuery,
  useGetProjectQuery,
  useGetProjectBySlugQuery,
  useGetProjectGanttQuery,
  useGetProjectResourcesQuery,
  useCreateProjectMutation,
  useUpdateProjectMutation,
} = projectsApi;
