import { apiSlice } from './apiSlice';
import type {
  Project,
  ProjectMember,
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
      transformResponse: (response: unknown): Project[] => {
        if (Array.isArray(response)) return response as Project[];
        if (response && typeof response === 'object') {
          const obj = response as Record<string, unknown>;
          if (Array.isArray(obj.data)) return obj.data as Project[];
          if (Array.isArray(obj.projects)) return obj.projects as Project[];
        }
        return [];
      },
      providesTags: (result) => {
        const items = Array.isArray(result) ? result : [];
        return [
          ...items.map(({ id }) => ({ type: 'Project' as const, id })),
          { type: 'Project', id: 'LIST' },
        ];
      },
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

    getProjectMembers: builder.query<
      ProjectMember[],
      { projectId: string; search?: string }
    >({
      query: ({ projectId, search }) => {
        const params = new URLSearchParams();
        params.append('projectId', projectId);
        if (search) params.append('search', search);
        return {
          url: `/project-members?${params.toString()}`,
        };
      },
      transformResponse: (response: { data: ProjectMember[] } | ProjectMember[]) => {
        return Array.isArray(response) ? response : response.data || [];
      },
      providesTags: (_result, _err, { projectId }) => [
        { type: 'Project', id: `MEMBERS_${projectId}` },
      ],
    }),

    inviteProjectMember: builder.mutation<
      void,
      { projectId: string; email: string; role?: string }
    >({
      query: (body) => ({
        url: '/project-members/invite',
        method: 'POST',
        body,
      }),
      invalidatesTags: (_result, _err, { projectId }) => [
        { type: 'Project', id: `MEMBERS_${projectId}` },
      ],
    }),

    updateProjectMemberRole: builder.mutation<
      ProjectMember,
      { id: string; role: string; projectId: string }
    >({
      query: ({ id, role }) => ({
        url: `/project-members/${id}`,
        method: 'PATCH',
        body: { role },
      }),
      invalidatesTags: (_result, _err, { projectId }) => [
        { type: 'Project', id: `MEMBERS_${projectId}` },
      ],
    }),

    removeProjectMember: builder.mutation<
      void,
      { id: string; projectId: string }
    >({
      query: ({ id }) => ({
        url: `/project-members/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _err, { projectId }) => [
        { type: 'Project', id: `MEMBERS_${projectId}` },
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
  useDeleteProjectMutation,
  useArchiveProjectMutation,
  useGetProjectChartsQuery,
  useGetProjectMembersQuery,
  useInviteProjectMemberMutation,
  useUpdateProjectMemberRoleMutation,
  useRemoveProjectMemberMutation,
} = projectsApi;
