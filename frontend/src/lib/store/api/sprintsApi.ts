import { apiSlice } from './apiSlice';
import type { Sprint, SprintBurndownPoint } from '@orbito/types';
import { getSocket } from '../../socket';
import { ENDPOINTS } from '../../api/endpoint';

export const sprintsApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getSprints: builder.query<Sprint[], { projectId?: string } | void>({
      query: (arg) => ({
        url: ENDPOINTS.sprints.list,
        params: arg?.projectId ? { projectId: arg.projectId } : undefined,
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Sprint' as const, id })),
              { type: 'Sprint', id: 'LIST' },
            ]
          : [{ type: 'Sprint', id: 'LIST' }],
      async onCacheEntryAdded(arg, { updateCachedData, cacheDataLoaded, cacheEntryRemoved }) {
        try {
          await cacheDataLoaded;
          const socket = getSocket();
          if (arg?.projectId) {
            socket.emit('join:project', { projectId: arg.projectId });
          }

          const onSprintStarted = (payload: { data: Sprint }) => {
            updateCachedData((draft) => {
              const sprint = draft.find((s) => s.id === payload?.data?.id);
              if (sprint) {
                Object.assign(sprint, payload.data);
              }
            });
          };

          const onSprintCompleted = (payload: { data: Sprint }) => {
            updateCachedData((draft) => {
              const sprint = draft.find((s) => s.id === payload?.data?.id);
              if (sprint) {
                Object.assign(sprint, payload.data);
              }
            });
          };

          socket.on('sprint:started', onSprintStarted);
          socket.on('sprint:completed', onSprintCompleted);

          await cacheEntryRemoved;
          if (arg?.projectId) {
            socket.emit('leave:project', { projectId: arg.projectId });
          }
          socket.off('sprint:started', onSprintStarted);
          socket.off('sprint:completed', onSprintCompleted);
        } catch {
          // Socket gracefully handled
        }
      },
    }),

    getSprint: builder.query<Sprint, string>({
      query: (id) => ENDPOINTS.sprints.byId(id),
      providesTags: (_result, _err, id) => [{ type: 'Sprint', id }],
    }),

    createSprint: builder.mutation<Sprint, Partial<Sprint>>({
      query: (body) => ({
        url: ENDPOINTS.sprints.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'Sprint', id: 'LIST' }],
    }),

    startSprint: builder.mutation<
      Sprint,
      string | { id: string; payload?: { startDate?: string; endDate?: string; goal?: string } }
    >({
      query: (arg) => {
        const id = typeof arg === 'string' ? arg : arg.id;
        const body = typeof arg === 'string' ? undefined : arg.payload;
        return {
          url: ENDPOINTS.sprints.start(id),
          method: 'PATCH',
          body,
        };
      },
      invalidatesTags: (_result, _err, arg) => {
        const id = typeof arg === 'string' ? arg : arg.id;
        return [
          { type: 'Sprint', id },
          { type: 'Sprint', id: 'LIST' },
        ];
      },
    }),

    completeSprint: builder.mutation<
      Sprint,
      string | { id: string; payload?: { moveToSprintId?: string } }
    >({
      query: (arg) => {
        const id = typeof arg === 'string' ? arg : arg.id;
        const body = typeof arg === 'string' ? undefined : arg.payload;
        return {
          url: ENDPOINTS.sprints.complete(id),
          method: 'PATCH',
          body,
        };
      },
      invalidatesTags: (_result, _err, arg) => {
        const id = typeof arg === 'string' ? arg : arg.id;
        return [
          { type: 'Sprint', id },
          { type: 'Sprint', id: 'LIST' },
        ];
      },
    }),

    getSprintBurndown: builder.query<
      SprintBurndownPoint[],
      { projectSlug: string; sprintId: string }
    >({
      query: ({ projectSlug, sprintId }) =>
        ENDPOINTS.projects.sprintBurndown(projectSlug, sprintId),
      providesTags: (_result, _err, { sprintId }) => [{ type: 'Sprint', id: sprintId }],
    }),
  }),
});

export const {
  useGetSprintsQuery,
  useGetSprintQuery,
  useCreateSprintMutation,
  useStartSprintMutation,
  useCompleteSprintMutation,
  useGetSprintBurndownQuery,
} = sprintsApi;
