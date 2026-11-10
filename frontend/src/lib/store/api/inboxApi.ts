import { apiSlice } from './apiSlice';
import type {
  ProjectInbox,
  EmailAccount,
  InboxMessage,
  InboxRule,
  CreateInboxDto,
  SetupEmailDto,
  CreateRuleDto,
  MessageFiltersDto,
} from '@/types';
import { ENDPOINTS } from '../../api/endpoint';

export const inboxApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getProjectInbox: builder.query<ProjectInbox | null, string>({
      query: (projectId) => ENDPOINTS.inbox.projectInbox(projectId),
      providesTags: (_result, _err, projectId) => [
        { type: 'Inbox', id: projectId },
      ],
    }),

    createProjectInbox: builder.mutation<
      ProjectInbox,
      { projectId: string; dto: CreateInboxDto }
    >({
      query: ({ projectId, dto }) => ({
        url: ENDPOINTS.inbox.projectInbox(projectId),
        method: 'POST',
        body: dto,
      }),
      invalidatesTags: (_res, _err, { projectId }) => [
        { type: 'Inbox', id: projectId },
      ],
    }),

    updateProjectInbox: builder.mutation<
      ProjectInbox,
      { projectId: string; dto: Partial<CreateInboxDto> }
    >({
      query: ({ projectId, dto }) => ({
        url: ENDPOINTS.inbox.projectInbox(projectId),
        method: 'PUT',
        body: dto,
      }),
      invalidatesTags: (_res, _err, { projectId }) => [
        { type: 'Inbox', id: projectId },
      ],
    }),

    setupEmailAccount: builder.mutation<
      EmailAccount,
      { projectId: string; dto: SetupEmailDto }
    >({
      query: ({ projectId, dto }) => ({
        url: ENDPOINTS.inbox.emailAccount(projectId),
        method: 'PUT',
        body: dto,
      }),
      invalidatesTags: (_res, _err, { projectId }) => [
        { type: 'Inbox', id: projectId },
      ],
    }),

    testEmailConnection: builder.mutation<
      { success: boolean; message: string },
      { projectId: string; accountId: string }
    >({
      query: ({ projectId, accountId }) => ({
        url: ENDPOINTS.inbox.testEmail(projectId, accountId),
        method: 'POST',
      }),
    }),

    syncInbox: builder.mutation<
      { message: string; jobId?: string; status: string },
      string
    >({
      query: (projectId) => ({
        url: ENDPOINTS.inbox.sync(projectId),
        method: 'POST',
      }),
      invalidatesTags: (_res, _err, projectId) => [
        { type: 'Inbox', id: projectId },
        { type: 'Inbox', id: 'MESSAGES' },
        { type: 'Task', id: 'LIST' },
      ],
    }),

    getInboxMessages: builder.query<
      InboxMessage[],
      { projectId: string; filters?: MessageFiltersDto }
    >({
      query: ({ projectId, filters }) => ({
        url: ENDPOINTS.inbox.messages(projectId),
        params: filters || undefined,
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Inbox' as const, id: `MSG_${id}` })),
              { type: 'Inbox', id: 'MESSAGES' },
            ]
          : [{ type: 'Inbox', id: 'MESSAGES' }],
    }),

    getInboxMessage: builder.query<
      InboxMessage,
      { projectId: string; messageId: string }
    >({
      query: ({ projectId, messageId }) =>
        ENDPOINTS.inbox.messageById(projectId, messageId),
      providesTags: (_res, _err, { messageId }) => [
        { type: 'Inbox', id: `MSG_${messageId}` },
      ],
    }),

    convertMessageToTask: builder.mutation<
      { message?: string; task?: { id: string; title: string } },
      { projectId: string; messageId: string }
    >({
      query: ({ projectId, messageId }) => ({
        url: ENDPOINTS.inbox.convertMessage(projectId, messageId),
        method: 'POST',
      }),
      invalidatesTags: (_res, _err, { projectId, messageId }) => [
        { type: 'Inbox', id: projectId },
        { type: 'Inbox', id: 'MESSAGES' },
        { type: 'Inbox', id: `MSG_${messageId}` },
        { type: 'Task', id: 'LIST' },
      ],
    }),

    ignoreMessage: builder.mutation<
      { message?: string },
      { projectId: string; messageId: string }
    >({
      query: ({ projectId, messageId }) => ({
        url: ENDPOINTS.inbox.ignoreMessage(projectId, messageId),
        method: 'POST',
      }),
      invalidatesTags: (_res, _err, { messageId }) => [
        { type: 'Inbox', id: 'MESSAGES' },
        { type: 'Inbox', id: `MSG_${messageId}` },
      ],
    }),

    getInboxRules: builder.query<InboxRule[], string>({
      query: (projectId) => ENDPOINTS.inbox.rules(projectId),
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Inbox' as const, id: `RULE_${id}` })),
              { type: 'Inbox', id: 'RULES' },
            ]
          : [{ type: 'Inbox', id: 'RULES' }],
    }),

    createInboxRule: builder.mutation<
      InboxRule,
      { projectId: string; dto: CreateRuleDto }
    >({
      query: ({ projectId, dto }) => ({
        url: ENDPOINTS.inbox.rules(projectId),
        method: 'POST',
        body: dto,
      }),
      invalidatesTags: [{ type: 'Inbox', id: 'RULES' }],
    }),

    updateInboxRule: builder.mutation<
      InboxRule,
      { projectId: string; ruleId: string; dto: CreateRuleDto }
    >({
      query: ({ projectId, ruleId, dto }) => ({
        url: ENDPOINTS.inbox.ruleById(projectId, ruleId),
        method: 'PUT',
        body: dto,
      }),
      invalidatesTags: (_res, _err, { ruleId }) => [
        { type: 'Inbox', id: 'RULES' },
        { type: 'Inbox', id: `RULE_${ruleId}` },
      ],
    }),

    deleteInboxRule: builder.mutation<
      { message?: string },
      { projectId: string; ruleId: string }
    >({
      query: ({ projectId, ruleId }) => ({
        url: ENDPOINTS.inbox.ruleById(projectId, ruleId),
        method: 'DELETE',
      }),
      invalidatesTags: (_res, _err, { ruleId }) => [
        { type: 'Inbox', id: 'RULES' },
        { type: 'Inbox', id: `RULE_${ruleId}` },
      ],
    }),
  }),
});

export const {
  useGetProjectInboxQuery,
  useCreateProjectInboxMutation,
  useUpdateProjectInboxMutation,
  useSetupEmailAccountMutation,
  useTestEmailConnectionMutation,
  useSyncInboxMutation,
  useGetInboxMessagesQuery,
  useGetInboxMessageQuery,
  useConvertMessageToTaskMutation,
  useIgnoreMessageMutation,
  useGetInboxRulesQuery,
  useCreateInboxRuleMutation,
  useUpdateInboxRuleMutation,
  useDeleteInboxRuleMutation,
} = inboxApi;

