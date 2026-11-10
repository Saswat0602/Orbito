import { apiSlice } from './apiSlice';
import type {
  Task,
  TaskQueryParams,
  TaskGroupedResponse,
  TaskReorderPayload,
  TaskComment,
} from '@orbito/types';
import { getSocket } from '../../socket';
import { ENDPOINTS } from '../../api/endpoint';

export const tasksApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getTasks: builder.query<Task[], TaskQueryParams | void>({
      query: (params) => ({
        url: ENDPOINTS.tasks.list,
        params: params || undefined,
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Task' as const, id })),
              { type: 'Task', id: 'LIST' },
            ]
          : [{ type: 'Task', id: 'LIST' }],
      async onCacheEntryAdded(arg, { updateCachedData, cacheDataLoaded, cacheEntryRemoved }) {
        try {
          await cacheDataLoaded;
          const socket = getSocket();

          if (arg?.projectId) {
            socket.emit('join:project', { projectId: arg.projectId });
          }

          const onTaskCreated = (payload: { data: Task }) => {
            updateCachedData((draft) => {
              if (payload?.data && !draft.some((t) => t.id === payload.data.id)) {
                draft.unshift(payload.data);
              }
            });
          };

          const onTaskUpdated = (payload: {
            data: { taskId: string; updates: Partial<Task> };
          }) => {
            updateCachedData((draft) => {
              const task = draft.find((t) => t.id === payload?.data?.taskId);
              if (task && payload?.data?.updates) {
                Object.assign(task, payload.data.updates);
              }
            });
          };

          const onTaskDeleted = (payload: { data: { taskId: string } }) => {
            updateCachedData((draft) => {
              return draft.filter((t) => t.id !== payload?.data?.taskId);
            });
          };

          socket.on('task:created', onTaskCreated);
          socket.on('task:updated', onTaskUpdated);
          socket.on('task:deleted', onTaskDeleted);

          await cacheEntryRemoved;
          if (arg?.projectId) {
            socket.emit('leave:project', { projectId: arg.projectId });
          }
          socket.off('task:created', onTaskCreated);
          socket.off('task:updated', onTaskUpdated);
          socket.off('task:deleted', onTaskDeleted);
        } catch {
          // Socket gracefully handled
        }
      },
    }),

    getGroupedTasks: builder.query<TaskGroupedResponse, TaskQueryParams>({
      query: (params) => ({
        url: ENDPOINTS.tasks.grouped,
        params,
      }),
      providesTags: [{ type: 'Task', id: 'LIST' }],
    }),

    getTask: builder.query<Task, string>({
      query: (id) => ENDPOINTS.tasks.byId(id),
      providesTags: (_result, _err, id) => [{ type: 'Task', id }],
      async onCacheEntryAdded(id, { updateCachedData, cacheDataLoaded, cacheEntryRemoved }) {
        try {
          await cacheDataLoaded;
          const socket = getSocket();
          socket.emit('join:task', { taskId: id });

          const onTaskUpdated = (payload: {
            data: { taskId: string; updates: Partial<Task> };
          }) => {
            if (payload?.data?.taskId === id) {
              updateCachedData((draft) => {
                Object.assign(draft, payload.data.updates);
              });
            }
          };

          const onCommentAdded = (payload: {
            data: { taskId: string; comment: TaskComment };
          }) => {
            if (payload?.data?.taskId === id) {
              updateCachedData((draft) => {
                if (!draft.comments) draft.comments = [];
                if (!draft.comments.some((c) => c.id === payload.data.comment.id)) {
                  draft.comments.push(payload.data.comment);
                }
              });
            }
          };

          socket.on('task:updated', onTaskUpdated);
          socket.on('comment:added', onCommentAdded);

          await cacheEntryRemoved;
          socket.emit('leave:task', { taskId: id });
          socket.off('task:updated', onTaskUpdated);
          socket.off('comment:added', onCommentAdded);
        } catch {
          // Socket gracefully handled
        }
      },
    }),

    createTask: builder.mutation<Task, Partial<Task>>({
      query: (task) => ({
        url: ENDPOINTS.tasks.create,
        method: 'POST',
        body: task,
      }),
      invalidatesTags: [{ type: 'Task', id: 'LIST' }],
    }),

    updateTask: builder.mutation<Task, { id: string; updates: Partial<Task> }>({
      query: ({ id, updates }) => ({
        url: ENDPOINTS.tasks.byId(id),
        method: 'PATCH',
        body: updates,
      }),
      invalidatesTags: (_result, _err, { id }) => [
        { type: 'Task', id },
        { type: 'Task', id: 'LIST' },
      ],
    }),

    reorderTask: builder.mutation<Task, { taskId: string; payload: TaskReorderPayload }>({
      query: ({ taskId, payload }) => ({
        url: ENDPOINTS.taskRanks.reorder(taskId),
        method: 'PATCH',
        body: payload,
      }),
      invalidatesTags: [{ type: 'Task', id: 'LIST' }],
    }),

    deleteTask: builder.mutation<void, string>({
      query: (id) => ({
        url: ENDPOINTS.tasks.byId(id),
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _err, id) => [
        { type: 'Task', id },
        { type: 'Task', id: 'LIST' },
      ],
    }),

    getTaskComments: builder.query<TaskComment[], string>({
      query: (taskId) => ({
        url: ENDPOINTS.taskComments.list,
        params: { taskId },
      }),
      providesTags: (_res, _err, taskId) => [{ type: 'Comment', id: taskId }],
    }),

    createTaskComment: builder.mutation<TaskComment, { taskId: string; content: string }>({
      query: (body) => ({
        url: ENDPOINTS.taskComments.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: (_res, _err, { taskId }) => [
        { type: 'Comment', id: taskId },
        { type: 'Task', id: taskId },
      ],
    }),

    deleteTaskComment: builder.mutation<void, { id: string; taskId: string }>({
      query: ({ id }) => ({
        url: ENDPOINTS.taskComments.byId(id),
        method: 'DELETE',
      }),
      invalidatesTags: (_res, _err, { taskId }) => [{ type: 'Comment', id: taskId }],
    }),

    updateTaskAssignees: builder.mutation<Task, { taskId: string; assigneeIds: string[] }>({
      query: ({ taskId, assigneeIds }) => ({
        url: `/tasks/${taskId}/assignees`,
        method: 'PATCH',
        body: { assigneeIds },
      }),
      invalidatesTags: (_res, _err, { taskId }) => [
        { type: 'Task', id: taskId },
        { type: 'Task', id: 'LIST' },
      ],
    }),
  }),
});

export const {
  useGetTasksQuery,
  useGetGroupedTasksQuery,
  useGetTaskQuery,
  useCreateTaskMutation,
  useUpdateTaskMutation,
  useReorderTaskMutation,
  useDeleteTaskMutation,
  useGetTaskCommentsQuery,
  useCreateTaskCommentMutation,
  useDeleteTaskCommentMutation,
  useUpdateTaskAssigneesMutation,
} = tasksApi;
