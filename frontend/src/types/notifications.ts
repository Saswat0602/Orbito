export interface NotificationItem {
  id: string;
  userId: string;
  type: string;
  title: string;
  message: string;
  isRead: boolean;
  priority?: string;
  entityType?: string | null;
  entityId?: string | null;
  actionUrl?: string | null;
  createdAt: string;
  readAt?: string | null;
  createdByUser?: {
    id: string;
    firstName: string;
    lastName?: string;
    avatar?: string;
  };
}

export interface NotificationsResponse {
  notifications: NotificationItem[];
  pagination?: {
    currentPage: number;
    totalPages: number;
    totalCount: number;
    hasNextPage: boolean;
    hasPrevPage: boolean;
  };
}

export interface UnreadCountResponse {
  count: number;
}

