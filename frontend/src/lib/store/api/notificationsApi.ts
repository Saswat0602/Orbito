import { apiSlice } from './apiSlice';
import { ENDPOINTS } from '../../api/endpoint';
import type { NotificationItem, NotificationsResponse } from '@/types';

export type { NotificationItem, NotificationsResponse };

export const notificationsApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getNotifications: builder.query<
      NotificationItem[],
      { isRead?: boolean; page?: number; limit?: number; organizationId?: string } | void
    >({
      query: (params) => ({
        url: ENDPOINTS.notifications.list,
        params: params || {},
      }),
      transformResponse: (response: NotificationsResponse | NotificationItem[] | unknown) => {
        if (Array.isArray(response)) return response;
        if (
          response &&
          typeof response === 'object' &&
          Array.isArray((response as NotificationsResponse).notifications)
        ) {
          return (response as NotificationsResponse).notifications;
        }
        return [];
      },
      providesTags: ['Notification'],
    }),

    getUnreadCount: builder.query<number, { organizationId?: string } | void>({
      query: (params) => ({
        url: ENDPOINTS.notifications.unreadCount,
        params: params || {},
      }),
      transformResponse: (response: number | { count: number } | unknown) => {
        if (typeof response === 'number') return response;
        if (
          response &&
          typeof response === 'object' &&
          typeof (response as { count?: number }).count === 'number'
        ) {
          return (response as { count: number }).count;
        }
        return 0;
      },
      providesTags: ['Notification'],
    }),

    markAsRead: builder.mutation<{ message: string }, string>({
      query: (id) => ({
        url: ENDPOINTS.notifications.markAsRead(id),
        method: 'PATCH',
      }),
      invalidatesTags: ['Notification'],
    }),

    markAllAsRead: builder.mutation<{ message: string }, { organizationId?: string } | void>({
      query: (params) => ({
        url: ENDPOINTS.notifications.markAllAsRead,
        method: 'PATCH',
        params: params || {},
      }),
      invalidatesTags: ['Notification'],
    }),
  }),
});

export const {
  useGetNotificationsQuery,
  useGetUnreadCountQuery,
  useMarkAsReadMutation,
  useMarkAllAsReadMutation,
} = notificationsApi;
