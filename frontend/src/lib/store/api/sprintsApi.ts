import { apiSlice } from './apiSlice';
import type { Sprint, SprintBurndownPoint } from '@orbito/types';
import { getSocket } from '../../socket';

export const sprintsApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getSprints: builder.query<Sprint[], { projectId: string }>({
      query: ({ projectId }) => ({
        url: `/sprints`,
        params: { projectId },
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
          socket.emit('join:project', { projectId: arg.projectId });

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
          socket.emit('leave:project', { projectId: arg.projectId });
          socket.off('sprint:started', onSprintStarted);
          socket.off('sprint:completed', onSprintCompleted);
        } catch {
          // Socket gracefully handled
        }
      },
    }),

    getSprint: builder.query<Sprint, string>({
      query: (id) => `/sprints/${id}`,
      providesTags: (_result, _err, id) => [{ type: 'Sprint', id }],
    }),

    startSprint: builder.mutation<Sprint, string>({
      query: (id) => ({
        url: `/sprints/${id}/start`,
        method: 'PATCH',
      }),
      invalidatesTags: (_result, _err, id) => [
        { type: 'Sprint', id },
        { type: 'Sprint', id: 'LIST' },
      ],
    }),

    completeSprint: builder.mutation<Sprint, string>({
      query: (id) => ({
        url: `/sprints/${id}/complete`,
        method: 'PATCH',
      }),
      invalidatesTags: (_result, _err, id) => [
        { type: 'Sprint', id },
        { type: 'Sprint', id: 'LIST' },
      ],
    }),

    getSprintBurndown: builder.query<
      SprintBurndownPoint[],
      { projectSlug: string; sprintId: string }
    >({
      query: ({ projectSlug, sprintId }) =>
        `/projects/${projectSlug}/charts/sprint-burndown/${sprintId}`,
      providesTags: (_result, _err, { sprintId }) => [{ type: 'Sprint', id: sprintId }],
    }),
  }),
});

export const {
  useGetSprintsQuery,
  useGetSprintQuery,
  useStartSprintMutation,
  useCompleteSprintMutation,
  useGetSprintBurndownQuery,
} = sprintsApi;

