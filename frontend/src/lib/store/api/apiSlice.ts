import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { BaseQueryFn, FetchArgs, FetchBaseQueryError } from '@reduxjs/toolkit/query';
import { Mutex } from 'async-mutex';

const mutex = new Mutex();
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:3000/api';

const baseQuery = fetchBaseQuery({
  baseUrl: API_BASE_URL,
  prepareHeaders: (headers) => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('orbito_access_token');
      if (token) {
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

  if (result.error && result.error.status === 401) {
    if (!mutex.isLocked()) {
      const release = await mutex.acquire();
      try {
        const storedRefreshToken =
          typeof window !== 'undefined' ? localStorage.getItem('orbito_refresh_token') : null;

        const refreshResult = await baseQuery(
          {
            url: '/auth/refresh',
            method: 'POST',
            body: storedRefreshToken ? { refreshToken: storedRefreshToken } : {},
          },
          api,
          extraOptions
        );

        if (refreshResult.data) {
          const data = refreshResult.data as {
            access_token?: string;
            accessToken?: string;
            refresh_token?: string;
            refreshToken?: string;
          };
          const newAccessToken = data.access_token || data.accessToken;
          const newRefreshToken = data.refresh_token || data.refreshToken;

          if (newAccessToken && typeof window !== 'undefined') {
            localStorage.setItem('orbito_access_token', newAccessToken);
            if (newRefreshToken) {
              localStorage.setItem('orbito_refresh_token', newRefreshToken);
            }
          }
          // Retry initial query
          result = await baseQuery(args, api, extraOptions);
        } else {
          // Refresh failed - log out
          if (typeof window !== 'undefined') {
            localStorage.removeItem('orbito_access_token');
            localStorage.removeItem('orbito_refresh_token');
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
  ],
  endpoints: () => ({}),
});

