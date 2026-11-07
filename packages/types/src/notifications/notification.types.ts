import type { NotificationType, NotificationPriority } from '../enums';

export interface Notification {
  id: string;
  userId: string;
  type: NotificationType;
  priority: NotificationPriority;
  title: string;
  message: string;
  isRead: boolean;
  linkUrl?: string | null;
  metadata?: Record<string, unknown>;
  createdAt: string;
  readAt?: string | null;
}
