import type { Task, TaskComment } from '../tasks/task.types';
import type { TaskReorderResult } from '../tasks/task-dto.types';

export interface WebSocketEventPayload<T = unknown> {
  event: string;
  data: T;
  timestamp: string;
}

export type TaskCreatedEvent = WebSocketEventPayload<Task>;
export type TaskUpdatedEvent = WebSocketEventPayload<Partial<Task> & { id: string }>;
export type TaskMovedEvent = WebSocketEventPayload<TaskReorderResult>;
export type TaskDeletedEvent = WebSocketEventPayload<{ id: string; projectId: string }>;
export type CommentAddedEvent = WebSocketEventPayload<TaskComment>;
export type UserPresenceEvent = WebSocketEventPayload<{
  userId: string;
  status: 'ONLINE' | 'OFFLINE' | 'AWAY';
}>;
export type UserTypingEvent = WebSocketEventPayload<{
  userId: string;
  taskId: string;
  isTyping: boolean;
}>;
