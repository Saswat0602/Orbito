# 09. Frontend Architecture: RTK Query, Interceptors & WebSocket Cache Sync

## 1. Overview & Architecture Philosophy

The frontend of **Orbito** does not reuse the dated visual design of Taskosaur. Instead:
- **Functional Parity Kept as Reference**: All business logic, domain rules, and workflow mechanisms from Taskosaur are preserved in code.
- **Fresh High-Velocity Monochrome UI**: Built with a distinct monochrome aesthetic inspired by **Linear**, **Slack**, and **Monday.com**.
- **Modern Data Architecture**: Powered by **Redux Toolkit Query (RTK Query)** for declarative data fetching, automated cache invalidation, and bidirectional real-time WebSocket cache updates.

```
                  +----------------------------------------------+
                  |           Orbito Modern Frontend (Next.js)   |
                  +----------------------------------------------+
                                         │
                   ┌─────────────────────┴────────────────────┐
                   ▼                                          ▼
      ┌─────────────────────────┐               ┌───────────────────────────┐
      │   Presentation Layer    │               │    State & Data Layer     │
      │  - Monochrome UI (@orbito/ui)           │  - RTK Query Store        │
      │  - Linear-style Command (Cmd+K)         │  - Feature API Slices     │
      │  - Slack Channel/Project Nav            │  - Auth & HTTP Interceptor│
      │  - Monday Custom Task Views             │  - Real-Time Socket Sync  │
      └─────────────────────────┘               └───────────────────────────┘
```

---

## 2. Redux Toolkit Query (RTK Query) Architecture

### 2.1 Store Configuration (`src/lib/store/index.ts`)

```typescript
import { configureStore } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import { apiSlice } from './api/apiSlice';
import authReducer from './slices/authSlice';
import uiReducer from './slices/uiSlice';

export const store = configureStore({
  reducer: {
    [apiSlice.reducerPath]: apiSlice.reducer,
    auth: authReducer,
    ui: uiReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(apiSlice.middleware),
});

setupListeners(store.dispatch);

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

---

### 2.2 Base API Slice with Auth Interceptor & Auto-Refresh (`src/lib/store/api/apiSlice.ts`)

The base query includes an automatic token-refresh mutex to intercept 401 Unauthorized responses and refresh tokens transparently:

```typescript
import {
  createApi,
  fetchBaseQuery,
  BaseQueryFn,
  FetchArgs,
  FetchBaseQueryError,
} from '@reduxjs/toolkit/query/react';
import { Mutex } from 'async-mutex';
import { setCredentials, logout } from '../slices/authSlice';

const mutex = new Mutex();

const baseQuery = fetchBaseQuery({
  baseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:3000/api',
  credentials: 'include',
  prepareHeaders: (headers, { getState }) => {
    const token = (getState() as any).auth.token;
    if (token) {
      headers.set('authorization', `Bearer ${token}`);
    }
    const currentWorkspaceId = (getState() as any).ui.currentWorkspaceId;
    if (currentWorkspaceId) {
      headers.set('x-workspace-id', currentWorkspaceId);
    }
    return headers;
  },
});

export const baseQueryWithReauth: BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError
> = async (args, api, extraOptions) => {
  await mutex.waitForUnlock();
  let result = await baseQuery(args, api, extraOptions);

  if (result.error && result.error.status === 401) {
    if (!mutex.isLocked()) {
      const release = await mutex.acquire();
      try {
        const refreshResult = await baseQuery(
          { url: '/auth/refresh', method: 'POST' },
          api,
          extraOptions
        );

        if (refreshResult.data) {
          api.dispatch(setCredentials(refreshResult.data));
          // Retry original request with refreshed token
          result = await baseQuery(args, api, extraOptions);
        } else {
          api.dispatch(logout());
        }
      } finally {
        release();
      }
    } else {
      await mutex.waitForUnlock();
      result = await baseQuery(args, api, extraOptions);
    }
  }

  return result;
};

export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: baseQueryWithReauth,
  tagTypes: [
    'Task',
    'Project',
    'Workspace',
    'Organization',
    'User',
    'Sprint',
    'Workflow',
    'Comment',
    'Attachment',
    'Activity',
  ],
  endpoints: () => ({}),
});
```

---

## 3. Real-Time Socket.io Cache Invalidation & Streaming Updates

Instead of polling or full refetches, RTK Query uses `onCacheEntryAdded` to stream real-time updates directly into the client cache:

```typescript
// src/lib/store/api/tasksApi.ts
import { apiSlice } from './apiSlice';
import { getSocket } from '../../socket/socketClient';
import { TaskSummary } from '@orbito/types';

export const tasksApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getProjectTasks: builder.query<TaskSummary[], string>({
      query: (projectId) => `/projects/${projectId}/tasks`,
      providesTags: (result, error, projectId) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Task' as const, id })),
              { type: 'Task', id: `LIST_${projectId}` },
            ]
          : [{ type: 'Task', id: `LIST_${projectId}` }],

      async onCacheEntryAdded(
        projectId,
        { updateCachedData, cacheDataLoaded, cacheEntryRemoved }
      ) {
        const socket = getSocket();

        try {
          await cacheDataLoaded;

          // Listen for task movement (LexoRank / Status change)
          const handleTaskMoved = (data: { taskId: string; newStatusId: string; newRank: string }) => {
            updateCachedData((draft) => {
              const task = draft.find((t) => t.id === data.taskId);
              if (task) {
                task.statusId = data.newStatusId;
                task.rank = data.newRank;
              }
            });
          };

          // Listen for task updates
          const handleTaskUpdated = (updatedTask: TaskSummary) => {
            updateCachedData((draft) => {
              const index = draft.findIndex((t) => t.id === updatedTask.id);
              if (index !== -1) {
                draft[index] = { ...draft[index], ...updatedTask };
              }
            });
          };

          // Listen for new task creation
          const handleTaskCreated = (newTask: TaskSummary) => {
            if (newTask.projectId === projectId) {
              updateCachedData((draft) => {
                draft.unshift(newTask);
              });
            }
          };

          socket.on('task:moved', handleTaskMoved);
          socket.on('task:updated', handleTaskUpdated);
          socket.on('task:created', handleTaskCreated);

          await cacheEntryRemoved;
          socket.off('task:moved', handleTaskMoved);
          socket.off('task:updated', handleTaskUpdated);
          socket.off('task:created', handleTaskCreated);
        } catch {
          // Cache entry removed before loading completed
        }
      },
    }),

    // Optimistic mutation for dragging tasks
    moveTask: builder.mutation<
      void,
      { taskId: string; projectId: string; statusId: string; rank: string }
    >({
      query: ({ taskId, ...body }) => ({
        url: `/tasks/${taskId}/move`,
        method: 'PATCH',
        body,
      }),
      async onQueryStarted({ taskId, projectId, statusId, rank }, { dispatch, queryFulfilled }) {
        const patchResult = dispatch(
          tasksApi.util.updateQueryData('getProjectTasks', projectId, (draft) => {
            const task = draft.find((t) => t.id === taskId);
            if (task) {
              task.statusId = statusId;
              task.rank = rank;
            }
          })
        );
        try {
          await queryFulfilled;
        } catch {
          patchResult.undo();
        }
      },
    }),
  }),
});

export const { useGetProjectTasksQuery, useMoveTaskMutation } = tasksApi;
```

---

## 4. Design & Component Architecture (The New Orbito Experience)

The old Taskosaur UI is strictly kept as a reference for backend API endpoints and domain requirements. The new UI is engineered from scratch:

| Feature | Legacy Taskosaur Baseline | New Orbito Experience |
| :--- | :--- | :--- |
| **Color System** | Standard colorful SaaS blue/purple | **Obsidian Monochrome**: Jet black (`#09090b`), Dark Zinc (`#18181b`), High-contrast crisp white typography, subtle zinc borders (`#27272a`). |
| **Navigation** | Top navbar + basic sidebar | **Slack-Style Workspace Sidebar**: Collapsible workspace trees, active presence badges, unread notification dots, and starred project channels. |
| **Quick Actions** | Button clicks only | **Linear-Style Command Palette (`Cmd+K`)**: Instant search across all tasks, quick status toggle, assignee filter, and keyboard shortcuts (`C` = create task). |
| **Views System** | Basic hardcoded board | **Monday-Style Dynamic View Bar**: One-click switcher between Kanban Board, List Table, Agile Sprints, Calendar, and Gantt charts with sticky filters. |
| **Data Fetching** | Raw fetch / axios calls in useEffect | **RTK Query with Cache Streaming**: Instant optimistic updates, zero network waterfalls, automatic background revalidation. |
| **Real-Time Presence** | Page reloads on update | **Live WebSocket Synced**: Collaborative live cursors, avatar presence bubbles ("Alex is editing..."), and instant card moves. |

