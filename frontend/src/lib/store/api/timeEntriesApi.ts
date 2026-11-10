import { apiSlice } from './apiSlice';
import type { TimeEntry } from '@orbito/types';
import { ENDPOINTS } from '../../api/endpoint';
import type {
  CreateTimeEntryRequest,
  ActiveTimerResponse,
} from '@/types';

export type {
  CreateTimeEntryRequest,
  ActiveTimerResponse,
};

export const timeEntriesApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getTimeEntries: builder.query<TimeEntry[], { taskId?: string } | void>({
      query: (params) => {
        if (params && params.taskId) {
          return ENDPOINTS.timeEntries.byTask(params.taskId);
        }
        return ENDPOINTS.timeEntries.list;
      },
      providesTags: ['TimeEntry'],
    }),
    getActiveTimer: builder.query<ActiveTimerResponse, void>({
      query: () => ENDPOINTS.timeEntries.activeTimer,
      providesTags: ['TimeEntry'],
    }),
    startTimer: builder.mutation<{ message: string; startedAt: string }, { taskId: string }>({
      query: (body) => ({
        url: ENDPOINTS.timeEntries.startTimer,
        method: 'POST',
        body,
      }),
      invalidatesTags: ['TimeEntry'],
    }),
    stopTimer: builder.mutation<TimeEntry, void>({
      query: () => ({
        url: ENDPOINTS.timeEntries.stopTimer,
        method: 'POST',
      }),
      invalidatesTags: ['TimeEntry', 'Task'],
    }),
    createTimeEntry: builder.mutation<TimeEntry, CreateTimeEntryRequest>({
      query: (body) => ({
        url: ENDPOINTS.timeEntries.create,
        method: 'POST',
        body,
      }),
      invalidatesTags: ['TimeEntry', 'Task'],
    }),
  }),
});

export const {
  useGetTimeEntriesQuery,
  useGetActiveTimerQuery,
  useStartTimerMutation,
  useStopTimerMutation,
  useCreateTimeEntryMutation,
} = timeEntriesApi;

