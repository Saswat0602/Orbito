import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { BaseQueryFn, FetchArgs, FetchBaseQueryError } from '@reduxjs/toolkit/query';
import { Mutex } from 'async-mutex';
import { API_BASE_URL, ENDPOINTS } from '../../api/endpoint';

const mutex = new Mutex();

const baseQuery = fetchBaseQuery({
  baseUrl: API_BASE_URL,
  credentials: 'include', // Automatically send and receive httpOnly & auth cookies
  prepareHeaders: (headers) => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('access_token');
      if (token && !headers.has('authorization')) {
        headers.set('authorization', `Bearer ${token}`);
      }
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

  if (result.error) {
    // 1. Connection refused or server offline: abort re-auth and do not spam
    if (result.error.status === 'FETCH_ERROR') {
      return result;
    }

    // 2. Unauthorized (401): attempt synchronized cookie refresh
    if (result.error.status === 401) {
      const url = typeof args === 'string' ? args : args.url;
      // Skip reauth on login/refresh calls to prevent loops
      if (url.includes(ENDPOINTS.auth.login) || url.includes(ENDPOINTS.auth.refresh)) {
        return result;
      }

      if (!mutex.isLocked()) {
        const release = await mutex.acquire();
        try {
          // Cookie-based token refresh: browser sends refresh_token cookie automatically
          const refreshResult = await baseQuery(
            {
              url: ENDPOINTS.auth.refresh,
              method: 'POST',
              body: {},
            },
            api,
            extraOptions
          );

          if (refreshResult.data) {
            const data = refreshResult.data as { access_token?: string; refresh_token?: string };
            if (typeof window !== 'undefined') {
              if (data?.access_token) {
                localStorage.setItem('access_token', data.access_token);
                document.cookie = `access_token=${data.access_token}; Path=/; SameSite=Lax; Max-Age=3600`;
              }
              if (data?.refresh_token) {
                localStorage.setItem('refresh_token', data.refresh_token);
                document.cookie = `refresh_token=${data.refresh_token}; Path=/; SameSite=Lax; Max-Age=604800`;
              }
            }
            // Retry the original query now that cookies are refreshed
            result = await baseQuery(args, api, extraOptions);
          } else {
            // Refresh cookie expired or invalid: redirect to login with error parameter to clear cookies
            if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/login')) {
              window.location.href = '/login?error=session_expired';
            }
          }
        } finally {
          release();
        }
      } else {
        await mutex.waitForUnlock();
        result = await baseQuery(args, api, extraOptions);
      }
    }
  }

  return result;
};

export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: baseQueryWithReauth,
  tagTypes: [
    'Auth',
    'User',
    'Organization',
    'Workspace',
    'Project',
    'Task',
    'Sprint',
    'Comment',
    'Automation',
    'Notification',
    'ActivityLog',
    'Inbox',
    'Integration',
    'Label',
    'TimeEntry',
    'Attachment',
    'TaskDependency',
    'TaskShare',
    'AiChat',
    'Admin',
    'Queue',
    'Invitation',
    'TaskStatus',
    'Workflow',
  ],
  endpoints: () => ({}),
});
