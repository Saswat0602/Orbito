import { apiSlice } from './apiSlice';
import type {
  Project,
  GanttTask,
  ResourceAllocation,
} from '@orbito/types';
import type { ProjectChartDataResponse } from '@/types';
import { ENDPOINTS } from '../../api/endpoint';

export type {
  GanttTask,
  ResourceAllocation,
};

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

    deleteProject: builder.mutation<void, string>({
      query: (id) => ({
        url: ENDPOINTS.projects.byId(id),
        method: 'DELETE',
      }),
      invalidatesTags: [{ type: 'Project', id: 'LIST' }],
    }),

    archiveProject: builder.mutation<Project, string>({
      query: (id) => ({
        url: `/projects/archive/${id}`,
        method: 'PATCH',
      }),
      invalidatesTags: (_result, _err, id) => [
        { type: 'Project', id },
        { type: 'Project', id: 'LIST' },
      ],
    }),

    getProjectCharts: builder.query<
      ProjectChartDataResponse,
      { slug: string; types: string[] }
    >({
      query: ({ slug, types }) => {
        const params = new URLSearchParams();
        types.forEach((t) => params.append('types', t));
        return {
          url: `${ENDPOINTS.projects.charts(slug)}?${params.toString()}`,
        };
      },
      providesTags: (_res, _err, { slug }) => [{ type: 'Project', id: slug }],
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
  useDeleteProjectMutation,
  useArchiveProjectMutation,
  useGetProjectChartsQuery,
} = projectsApi;
