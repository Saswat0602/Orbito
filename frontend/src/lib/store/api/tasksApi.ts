import { apiSlice } from './apiSlice';
import type {
  Task,
  TaskQueryParams,
  TaskGroupedResponse,
  TaskReorderPayload,
  TaskComment,
  Label,
  Attachment,
  TimeEntry,
  TaskRecurrence,
  RecurringTaskConfig,
  TaskStatus,
  TaskType,
} from '@orbito/types';
import type {
  TaskDependenciesResponse,
  TaskDependencyItem,
  CreateTaskDependencyRequest,
  BulkCreateTasksRequest,
  BulkCreateTasksResponse,
  CreateTaskShareRequest,
  TaskShareResponse,
  PublicSharedTask,
} from '@/types';
import { getSocket } from '../../socket';
import { ENDPOINTS } from '../../api/endpoint';

export const tasksApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getTasks: builder.query<Task[], TaskQueryParams | void>({
      query: (params) => ({
        url: ENDPOINTS.tasks.list,
        params: params || undefined,
      }),
      transformResponse: (response: unknown): Task[] => {
        if (Array.isArray(response)) return response as Task[];
        if (response && typeof response === 'object') {
          const obj = response as Record<string, unknown>;
          if (Array.isArray(obj.data)) return obj.data as Task[];
          if (Array.isArray(obj.tasks)) return obj.tasks as Task[];
        }
        return [];
      },
      providesTags: (result) => {
        const items = Array.isArray(result) ? result : [];
        return [
          ...items.map(({ id }) => ({ type: 'Task' as const, id })),
          { type: 'Task', id: 'LIST' },
        ];
      },
      async onCacheEntryAdded(arg, { updateCachedData, cacheDataLoaded, cacheEntryRemoved }) {
        try {
          await cacheDataLoaded;
          const socket = getSocket();

          if (arg?.projectId) {
            socket.emit('join:project', { projectId: arg.projectId });
          }

          const onTaskCreated = (payload: { data: Task }) => {
            updateCachedData((draft) => {
              if (payload?.data && Array.isArray(draft) && !draft.some((t) => t.id === payload.data.id)) {
                draft.unshift(payload.data);
              }
            });
          };

          const onTaskUpdated = (payload: {
            data: { taskId: string; updates: Partial<Task> };
          }) => {
            updateCachedData((draft) => {
              if (Array.isArray(draft)) {
                const task = draft.find((t) => t.id === payload?.data?.taskId);
                if (task && payload?.data?.updates) {
                  Object.assign(task, payload.data.updates);
                }
              }
            });
          };

          const onTaskDeleted = (payload: { data: { taskId: string } }) => {
            updateCachedData((draft) => {
              if (Array.isArray(draft)) {
                return draft.filter((t) => t.id !== payload?.data?.taskId);
              }
              return draft;
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

    updateTask: builder.mutation<
      Task,
      {
        id: string;
        updates: Partial<Task> & {
          reporterIds?: string[];
          assigneeIds?: string[];
          type?: TaskType;
          statusId?: string;
          parentTaskId?: string | null;
          stopRecurrence?: boolean;
        };
      }
    >({
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

    // Subtasks
    getSubtasks: builder.query<Task[], string>({
      query: (parentTaskId) => ({
        url: ENDPOINTS.tasks.list,
        params: { parentTaskId },
      }),
      transformResponse: (response: unknown): Task[] => {
        if (Array.isArray(response)) return response as Task[];
        if (response && typeof response === 'object') {
          const obj = response as Record<string, unknown>;
          if (Array.isArray(obj.data)) return obj.data as Task[];
          if (Array.isArray(obj.tasks)) return obj.tasks as Task[];
        }
        return [];
      },
      providesTags: (_res, _err, parentTaskId) => [{ type: 'Task', id: `SUBTASKS_${parentTaskId}` }],
    }),

    createSubtask: builder.mutation<Task, { parentTaskId: string; title: string; projectId: string }>({
      query: (body) => ({
        url: ENDPOINTS.tasks.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: (_res, _err, { parentTaskId }) => [
        { type: 'Task', id: `SUBTASKS_${parentTaskId}` },
        { type: 'Task', id: parentTaskId },
        { type: 'Task', id: 'LIST' },
      ],
    }),

    // Labels
    getTaskLabels: builder.query<Label[], string>({
      query: (taskId) => `/labels/task/${taskId}`,
      providesTags: (_res, _err, taskId) => [{ type: 'Label', id: taskId }],
    }),

    getAllLabels: builder.query<Label[], { organizationId?: string } | void>({
      query: (params) => ({
        url: ENDPOINTS.labels.list,
        params: params || {},
      }),
      providesTags: ['Label'],
    }),

    createLabel: builder.mutation<Label, { name: string; color: string; organizationId?: string }>({
      query: (body) => ({
        url: ENDPOINTS.labels.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Label'],
    }),

    assignTaskLabel: builder.mutation<void, { taskId: string; labelId: string }>({
      query: (body) => ({
        url: '/labels/assign',
        method: 'POST',
        body,
      }),
      invalidatesTags: (_res, _err, { taskId }) => [
        { type: 'Label', id: taskId },
        { type: 'Task', id: taskId },
        { type: 'Task', id: 'LIST' },
      ],
    }),

    removeTaskLabel: builder.mutation<void, { taskId: string; labelId: string }>({
      query: ({ taskId, labelId }) => ({
        url: `/labels/task/${taskId}/label/${labelId}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_res, _err, { taskId }) => [
        { type: 'Label', id: taskId },
        { type: 'Task', id: taskId },
        { type: 'Task', id: 'LIST' },
      ],
    }),

    // Attachments
    getTaskAttachments: builder.query<Attachment[], string>({
      query: (taskId) => `/task-attachments/task/${taskId}`,
      providesTags: (_res, _err, taskId) => [{ type: 'Attachment', id: taskId }],
    }),

    uploadTaskAttachment: builder.mutation<Attachment, { taskId: string; formData: FormData }>({
      query: ({ taskId, formData }) => ({
        url: `/task-attachments/upload/${taskId}`,
        method: 'POST',
        body: formData,
      }),
      invalidatesTags: (_res, _err, { taskId }) => [{ type: 'Attachment', id: taskId }],
    }),

    deleteTaskAttachment: builder.mutation<void, { id: string; taskId: string }>({
      query: ({ id }) => ({
        url: ENDPOINTS.taskAttachments.byId(id),
        method: 'DELETE',
      }),
      invalidatesTags: (_res, _err, { taskId }) => [{ type: 'Attachment', id: taskId }],
    }),

    // Dependencies
    getTaskDependencies: builder.query<TaskDependenciesResponse, string>({
      query: (taskId) => ENDPOINTS.taskDependencies.byTask(taskId),
      providesTags: (_res, _err, taskId) => [{ type: 'TaskDependency' as const, id: taskId }],
    }),

    createTaskDependency: builder.mutation<TaskDependencyItem, CreateTaskDependencyRequest>({
      query: (body) => ({
        url: ENDPOINTS.taskDependencies.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: (_res, _err, { dependentTaskId, blockingTaskId }) => [
        { type: 'TaskDependency', id: dependentTaskId },
        { type: 'TaskDependency', id: blockingTaskId },
        { type: 'Task', id: dependentTaskId },
        { type: 'Task', id: blockingTaskId },
      ],
    }),

    deleteTaskDependency: builder.mutation<void, { id: string; taskId: string }>({
      query: ({ id }) => ({
        url: ENDPOINTS.taskDependencies.byId(id),
        method: 'DELETE',
      }),
      invalidatesTags: (_res, _err, { taskId }) => [
        { type: 'TaskDependency', id: taskId },
        { type: 'Task', id: taskId },
      ],
    }),

    // Bulk Tasks Import
    bulkCreateTasks: builder.mutation<BulkCreateTasksResponse, BulkCreateTasksRequest>({
      query: (body) => ({
        url: ENDPOINTS.tasks.bulkCreate,
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'Task', id: 'LIST' }],
    }),

    getTaskStatusesByProject: builder.query<TaskStatus[], string>({
      query: (projectId) => ({
        url: ENDPOINTS.taskStatuses.byProject,
        params: { projectId },
      }),
      providesTags: (_result, _err, projectId) => [
        { type: 'TaskStatus', id: `PROJECT_${projectId}` },
        { type: 'TaskStatus', id: 'LIST' },
      ],
    }),

    createTaskStatus: builder.mutation<
      TaskStatus,
      {
        name: string;
        color: string;
        category: string;
        position?: number;
        workflowId: string;
        projectId?: string;
      }
    >({
      query: (body) => ({
        url: ENDPOINTS.taskStatuses.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: (_res, _err, arg) => [
        { type: 'TaskStatus', id: 'LIST' },
        ...(arg.projectId ? [{ type: 'TaskStatus' as const, id: `PROJECT_${arg.projectId}` }] : []),
        { type: 'Task', id: 'LIST' },
      ],
    }),

    updateTaskStatus: builder.mutation<
      TaskStatus,
      {
        id: string;
        updates: Partial<{ name: string; color: string; category: string; position: number }>;
        projectId?: string;
      }
    >({
      query: ({ id, updates }) => ({
        url: ENDPOINTS.taskStatuses.byId(id),
        method: 'PATCH',
        body: updates,
      }),
      invalidatesTags: (_res, _err, arg) => [
        { type: 'TaskStatus', id: 'LIST' },
        { type: 'TaskStatus', id: arg.id },
        ...(arg.projectId ? [{ type: 'TaskStatus' as const, id: `PROJECT_${arg.projectId}` }] : []),
        { type: 'Task', id: 'LIST' },
      ],
    }),

    updateTaskStatusPositions: builder.mutation<
      TaskStatus[],
      {
        statusUpdates: Array<{ id: string; position: number }>;
        projectId?: string;
      }
    >({
      query: ({ statusUpdates }) => ({
        url: ENDPOINTS.taskStatuses.positions,
        method: 'PATCH',
        body: { statusUpdates },
      }),
      invalidatesTags: (_res, _err, arg) => [
        { type: 'TaskStatus', id: 'LIST' },
        ...(arg.projectId ? [{ type: 'TaskStatus' as const, id: `PROJECT_${arg.projectId}` }] : []),
        { type: 'Task', id: 'LIST' },
      ],
    }),

    deleteTaskStatus: builder.mutation<
      { success?: boolean; message?: string },
      { id: string; projectId?: string }
    >({
      query: ({ id }) => ({
        url: ENDPOINTS.taskStatuses.byId(id),
        method: 'DELETE',
      }),
      invalidatesTags: (_res, _err, arg) => [
        { type: 'TaskStatus', id: 'LIST' },
        { type: 'TaskStatus', id: arg.id },
        ...(arg.projectId ? [{ type: 'TaskStatus' as const, id: `PROJECT_${arg.projectId}` }] : []),
        { type: 'Task', id: 'LIST' },
      ],
    }),

    // Public Task Sharing
    createTaskShare: builder.mutation<TaskShareResponse, CreateTaskShareRequest>({
      query: (body) => ({
        url: ENDPOINTS.taskShares.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: (_res, _err, { taskId }) => [{ type: 'TaskShare', id: taskId }],
    }),

    getTaskShares: builder.query<TaskShareResponse[], string>({
      query: (taskId) => ENDPOINTS.taskShares.byTask(taskId),
      providesTags: (_res, _err, taskId) => [{ type: 'TaskShare', id: taskId }],
    }),

    revokeTaskShare: builder.mutation<{ message: string }, { shareId: string; taskId: string }>({
      query: ({ shareId }) => ({
        url: ENDPOINTS.taskShares.revoke(shareId),
        method: 'DELETE',
      }),
      invalidatesTags: (_res, _err, { taskId }) => [{ type: 'TaskShare', id: taskId }],
    }),

    getPublicTask: builder.query<PublicSharedTask, string>({
      query: (token) => ENDPOINTS.taskShares.publicTask(token),
    }),

    // Recurrence Engine
    getTaskRecurrence: builder.query<TaskRecurrence | null, string>({
      query: (taskId) => ENDPOINTS.tasks.recurrence(taskId),
      providesTags: (_res, _err, taskId) => [{ type: 'Task', id: `RECUR_${taskId}` }],
    }),

    addTaskRecurrence: builder.mutation<
      TaskRecurrence,
      { taskId: string; config: Partial<TaskRecurrence> }
    >({
      query: ({ taskId, config }) => ({
        url: ENDPOINTS.tasks.recurrence(taskId),
        method: 'POST',
        body: config,
      }),
      invalidatesTags: (_res, _err, { taskId }) => [
        { type: 'Task', id: taskId },
        { type: 'Task', id: `RECUR_${taskId}` },
        { type: 'Task', id: 'LIST' },
      ],
    }),

    updateTaskRecurrence: builder.mutation<
      TaskRecurrence,
      { taskId: string; config: Partial<TaskRecurrence> }
    >({
      query: ({ taskId, config }) => ({
        url: ENDPOINTS.tasks.recurrence(taskId),
        method: 'PUT',
        body: config,
      }),
      invalidatesTags: (_res, _err, { taskId }) => [
        { type: 'Task', id: taskId },
        { type: 'Task', id: `RECUR_${taskId}` },
        { type: 'Task', id: 'LIST' },
      ],
    }),

    deleteTaskRecurrence: builder.mutation<{ success: boolean }, string>({
      query: (taskId) => ({
        url: ENDPOINTS.tasks.recurrence(taskId),
        method: 'DELETE',
      }),
      invalidatesTags: (_res, _err, taskId) => [
        { type: 'Task', id: taskId },
        { type: 'Task', id: `RECUR_${taskId}` },
        { type: 'Task', id: 'LIST' },
      ],
    }),

    completeTaskOccurrence: builder.mutation<{ success: boolean; nextTask?: Task }, string>({
      query: (taskId) => ({
        url: ENDPOINTS.tasks.completeOccurrence(taskId),
        method: 'POST',
      }),
      invalidatesTags: (_res, _err, taskId) => [
        { type: 'Task', id: taskId },
        { type: 'Task', id: 'LIST' },
      ],
    }),

    // Bulk Task Operations
    bulkUpdateTasksStatus: builder.mutation<
      { updatedCount: number; updatedTasks: Task[]; failedTasks: Array<{ id: string; reason: string }> },
      { taskIds?: string[]; statusId?: string; projectId?: string; all?: boolean; excludedIds?: string[] }
    >({
      query: (body) => ({
        url: ENDPOINTS.tasks.bulkStatus,
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'Task', id: 'LIST' }],
    }),

    bulkAssignTasks: builder.mutation<
      { assignedCount: number; updatedTasks: Task[]; failedTasks: Array<{ id: string; reason: string }> },
      { taskIds?: string[]; assigneeIds?: string[]; projectId?: string; all?: boolean; excludedIds?: string[] }
    >({
      query: (body) => ({
        url: ENDPOINTS.tasks.bulkAssign,
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'Task', id: 'LIST' }],
    }),

    bulkDeleteTasks: builder.mutation<
      { deletedCount: number; failedTasks: Array<{ id: string; reason: string }> },
      { taskIds?: string[]; projectId?: string; all?: boolean; excludedIds?: string[] }
    >({
      query: (body) => ({
        url: ENDPOINTS.tasks.bulkDelete,
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'Task', id: 'LIST' }],
    }),

    // Today's Agenda
    getTodaysTasks: builder.query<
      {
        tasks: Task[];
        pagination: {
          currentPage: number;
          totalPages: number;
          totalCount: number;
          hasNextPage: boolean;
          hasPrevPage: boolean;
        };
      },
      { organizationId?: string; page?: number; limit?: number } | void
    >({
      query: (params) => ({
        url: ENDPOINTS.tasks.today,
        params: params || undefined,
      }),
      providesTags: [{ type: 'Task', id: 'LIST' }],
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
  useGetSubtasksQuery,
  useCreateSubtaskMutation,
  useGetTaskLabelsQuery,
  useGetAllLabelsQuery,
  useCreateLabelMutation,
  useAssignTaskLabelMutation,
  useRemoveTaskLabelMutation,
  useGetTaskAttachmentsQuery,
  useUploadTaskAttachmentMutation,
  useDeleteTaskAttachmentMutation,
  useGetTaskDependenciesQuery,
  useCreateTaskDependencyMutation,
  useDeleteTaskDependencyMutation,
  useBulkCreateTasksMutation,
  useBulkUpdateTasksStatusMutation,
  useBulkAssignTasksMutation,
  useBulkDeleteTasksMutation,
  useGetTodaysTasksQuery,
  useGetTaskStatusesByProjectQuery,
  useCreateTaskStatusMutation,
  useUpdateTaskStatusMutation,
  useUpdateTaskStatusPositionsMutation,
  useDeleteTaskStatusMutation,
  useCreateTaskShareMutation,
  useGetTaskSharesQuery,
  useRevokeTaskShareMutation,
  useGetPublicTaskQuery,
  useGetTaskRecurrenceQuery,
  useAddTaskRecurrenceMutation,
  useUpdateTaskRecurrenceMutation,
  useDeleteTaskRecurrenceMutation,
  useCompleteTaskOccurrenceMutation,
} = tasksApi;
