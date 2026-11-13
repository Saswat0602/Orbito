import { apiSlice } from './apiSlice';
import { ENDPOINTS } from '../../api/endpoint';
import type {
  AiConversation,
  AiChatMessage,
  AiChatRequest,
  AiChatResponse,
  GenerateDescriptionRequest,
  GenerateDescriptionResponse,
  TestAiConnectionRequest,
  TestAiConnectionResponse,
} from '@/types';

export const aiChatApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    getConversations: builder.query<AiConversation[], void>({
      query: () => ENDPOINTS.aiChat.conversations,
      transformResponse: (response: unknown): AiConversation[] => {
        if (Array.isArray(response)) return response as AiConversation[];
        if (response && typeof response === 'object') {
          const obj = response as Record<string, unknown>;
          if (Array.isArray(obj.data)) return obj.data as AiConversation[];
          if (Array.isArray(obj.conversations)) return obj.conversations as AiConversation[];
        }
        return [];
      },
      providesTags: (result) => {
        const items = Array.isArray(result) ? result : [];
        return [
          ...items.map(({ id }) => ({ type: 'AiChat' as const, id })),
          { type: 'AiChat', id: 'LIST' },
        ];
      },
    }),

    createConversation: builder.mutation<
      AiConversation,
      { title?: string; sessionId?: string }
    >({
      query: (body) => ({
        url: ENDPOINTS.aiChat.conversations,
        method: 'POST',
        body,
      }),
      invalidatesTags: [{ type: 'AiChat', id: 'LIST' }],
    }),

    renameConversation: builder.mutation<
      AiConversation,
      { id: string; title: string }
    >({
      query: ({ id, title }) => ({
        url: ENDPOINTS.aiChat.conversationById(id),
        method: 'PATCH',
        body: { title },
      }),
      invalidatesTags: (_res, _err, { id }) => [
        { type: 'AiChat', id },
        { type: 'AiChat', id: 'LIST' },
      ],
    }),

    deleteConversation: builder.mutation<{ success: boolean }, string>({
      query: (id) => ({
        url: ENDPOINTS.aiChat.conversationById(id),
        method: 'DELETE',
      }),
      invalidatesTags: [{ type: 'AiChat', id: 'LIST' }],
    }),

    getMessages: builder.query<AiChatMessage[], string>({
      query: (conversationId) =>
        ENDPOINTS.aiChat.conversationMessages(conversationId),
      providesTags: (_res, _err, id) => [{ type: 'AiChat', id }],
    }),

    sendChatMessage: builder.mutation<AiChatResponse, AiChatRequest>({
      query: (body) => ({
        url: ENDPOINTS.aiChat.chat,
        method: 'POST',
        body,
      }),
    }),

    generateDescription: builder.mutation<
      GenerateDescriptionResponse,
      GenerateDescriptionRequest
    >({
      query: (body) => ({
        url: ENDPOINTS.aiChat.generateDescription,
        method: 'POST',
        body,
      }),
    }),

    testAiConnection: builder.mutation<
      TestAiConnectionResponse,
      TestAiConnectionRequest
    >({
      query: (body) => ({
        url: ENDPOINTS.aiChat.testConnection,
        method: 'POST',
        body,
      }),
    }),

    clearAiContext: builder.mutation<{ success: boolean }, string>({
      query: (sessionId) => ({
        url: ENDPOINTS.aiChat.clearContext(sessionId),
        method: 'DELETE',
      }),
    }),
  }),
});

export const {
  useGetConversationsQuery,
  useCreateConversationMutation,
  useRenameConversationMutation,
  useDeleteConversationMutation,
  useGetMessagesQuery,
  useSendChatMessageMutation,
  useGenerateDescriptionMutation,
  useTestAiConnectionMutation,
  useClearAiContextMutation,
} = aiChatApi;
