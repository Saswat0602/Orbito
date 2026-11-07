import { apiSlice } from './apiSlice';
import { ENDPOINTS } from '../../api/endpoint';

export interface ChatMessage {
  id?: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp?: string;
  actions?: Array<{
    type: string;
    description: string;
    payload?: Record<string, unknown>;
  }>;
}

export interface SendChatMessageRequest {
  message: string;
  conversationId?: string;
  projectId?: string;
  context?: Record<string, unknown>;
}

export interface ChatResponse {
  reply: string;
  conversationId: string;
  actions?: Array<{
    type: string;
    description: string;
    payload?: Record<string, unknown>;
  }>;
}

export const aiChatApi = apiSlice.injectEndpoints({
  endpoints: (builder) => ({
    sendChatMessage: builder.mutation<ChatResponse, SendChatMessageRequest>({
      query: (body) => ({
        url: ENDPOINTS.aiChat.chat,
        method: 'POST',
        body,
      }),
    }),
    getConversations: builder.query<Array<{ id: string; title: string; createdAt: string }>, void>({
      query: () => ENDPOINTS.aiChat.conversations,
    }),
    getMessages: builder.query<ChatMessage[], string>({
      query: (conversationId) => ENDPOINTS.aiChat.conversationMessages(conversationId),
    }),
  }),
});

export const {
  useSendChatMessageMutation,
  useGetConversationsQuery,
  useGetMessagesQuery,
} = aiChatApi;

