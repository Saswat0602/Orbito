export interface AiConversation {
  id: string;
  title: string;
  sessionId: string;
  createdAt: string;
  updatedAt: string;
  messages?: AiChatMessage[];
}

export interface AiChatMessage {
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

export interface AiChatRequest {
  message: string;
  history?: Array<{
    role: 'user' | 'assistant' | 'system';
    content: string;
  }>;
  workspaceId?: string;
  projectId?: string;
  sessionId?: string;
  currentOrganizationId?: string;
}

export interface AiChatResponse {
  message: string;
  success: boolean;
  error?: string;
  conversationId?: string;
  actions?: Array<{
    type: string;
    description: string;
    payload?: Record<string, unknown>;
  }>;
}

export interface GenerateDescriptionRequest {
  title: string;
  taskType?: string;
}

export interface GenerateDescriptionResponse {
  description: string;
  success: boolean;
  error?: string;
}

export interface TestAiConnectionRequest {
  apiKey: string;
  model: string;
  apiUrl: string;
}

export interface TestAiConnectionResponse {
  success: boolean;
  message?: string;
  error?: string;
}
