---
name: orbito-rtk-query-data-flow
description: >-
  Guides state management and asynchronous data fetching in the Orbito frontend using Redux Toolkit Query (RTK Query), tag-based caching, optimistic updates, and resilient error recovery.
---

# Orbito RTK Query & Data Flow Guide

This skill governs API integration, caching strategies, and data synchronization in `frontend/src/lib/store/api/`.

---

## 1. RTK Query Architecture in Orbito

All API queries in Orbito branch from the central `apiSlice` (`src/lib/store/api/apiSlice.ts`):
1. **Base Query with Re-auth**: Uses `baseQueryWithReauth` to handle automatic cookie refresh with `async-mutex` on `401 Unauthorized` responses.
2. **Endpoint Injection**: Modules inject endpoints into `apiSlice.injectEndpoints` rather than defining disconnected API slices.
3. **Tagged Cache System**: RTK Query tags ensure fine-grained cache invalidation across mutations without full-page reloads.

---

## 2. Best Practices for Authoring Endpoints

### 2.1 Safe Response Unwrapping (`transformResponse`)
Backend services in Orbito often return container objects with pagination (e.g. `{ tasks: Task[], pagination: { ... } }`). Always use `transformResponse` to guard against `undefined` or object crashes in UI consumers:

```typescript
getTasks: builder.query<TaskItem[], { projectId?: string; statusId?: string } | void>({
  query: (params) => ({
    url: ENDPOINTS.tasks.list,
    params: params || {},
  }),
  transformResponse: (response: { tasks?: TaskItem[] } | TaskItem[] | unknown) => {
    if (Array.isArray(response)) return response;
    if (response && typeof response === 'object' && Array.isArray((response as any).tasks)) {
      return (response as any).tasks;
    }
    return [];
  },
  providesTags: (result) =>
    result
      ? [
          ...result.map(({ id }) => ({ type: 'Task' as const, id })),
          { type: 'Task', id: 'LIST' },
        ]
      : [{ type: 'Task', id: 'LIST' }],
}),
```

### 2.2 Cache Invalidation on Mutations
Mutations should declare explicit `invalidatesTags` to notify all subscribed views to re-fetch:
```typescript
createTask: builder.mutation<TaskItem, CreateTaskDto>({
  query: (body) => ({
    url: ENDPOINTS.tasks.create,
    method: 'POST',
    body,
  }),
  invalidatesTags: [{ type: 'Task', id: 'LIST' }, 'Sprint', 'ActivityLog'],
}),
```

### 2.3 Optimistic Updates (`onQueryStarted`)
For fast interactions like task status dragging or checkbox completion:
```typescript
updateTaskStatus: builder.mutation<TaskItem, { id: string; statusId: string; projectId: string }>({
  query: ({ id, statusId }) => ({
    url: ENDPOINTS.tasks.update(id),
    method: 'PATCH',
    body: { statusId },
  }),
  async onQueryStarted({ id, statusId, projectId }, { dispatch, queryFulfilled }) {
    // Optimistically update project tasks cache
    const patchResult = dispatch(
      tasksApi.util.updateQueryData('getTasks', { projectId }, (draft) => {
        const task = draft.find((t) => t.id === id);
        if (task) {
          task.statusId = statusId;
        }
      })
    );
    try {
      await queryFulfilled;
    } catch {
      patchResult.undo(); // Roll back on network error
    }
  },
  invalidatesTags: (result, error, { id }) => [{ type: 'Task', id }],
}),
```

---

## 3. Polling and Refetch Patterns

- **Unread Badges & Notifications**:
  ```tsx
  const { data: count = 0 } = useGetUnreadCountQuery(undefined, {
    pollingInterval: 30000, // 30-second interval
    refetchOnFocus: true,
  });
  ```
- **Real-Time Queues**:
  ```tsx
  const { data: queues, refetch } = useGetQueueStatsQuery(undefined, {
    pollingInterval: 5000, // 5-second interval for background worker monitoring
  });
  ```

---

## 4. Troubleshooting Checklist

1. **`TypeError: data.map is not a function`**: The endpoint returned `{ items: [...] }` but the component expected an array. Add `transformResponse` in the API definition and defensive array checking in the component.
2. **Missing Auth Cookie**: Verify `credentials: 'include'` is declared in `baseQuery` in `apiSlice.ts`.
3. **Infinite Reauth Loop**: Verify that calls to `/auth/login` and `/auth/refresh` explicitly skip the 401 mutex trigger.
