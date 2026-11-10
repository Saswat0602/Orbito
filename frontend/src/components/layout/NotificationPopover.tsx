'use client';

import * as React from 'react';
import { useRouter } from 'next/navigation';
import {
  Bell,
  CheckCheck,
  Inbox,
  MessageSquare,
  AlertCircle,
  RefreshCw,
} from 'lucide-react';
import { toast } from 'sonner';

import {
  useGetNotificationsQuery,
  useGetUnreadCountQuery,
  useMarkAsReadMutation,
  useMarkAllAsReadMutation,
  type NotificationItem,
} from '@/lib/store/api/notificationsApi';
import {
  Popover,
  PopoverTrigger,
  PopoverContent,
} from '@/components/ui/popover';

export function NotificationPopover() {
  const router = useRouter();

  const {
    data: unreadData,
    refetch: refetchUnread,
  } = useGetUnreadCountQuery(undefined, {
    pollingInterval: 30000, // Poll count every 30s
  });

  const {
    data: notificationsData,
    isLoading,
    isError,
    refetch: refetchNotifications,
  } = useGetNotificationsQuery();

  const [markAsRead] = useMarkAsReadMutation();
  const [markAllAsRead, { isLoading: isMarkingAll }] = useMarkAllAsReadMutation();

  // Safely extract unreadCount regardless of whether backend returns number or { count: number }
  const unreadCount = React.useMemo(() => {
    if (typeof unreadData === 'number') return Math.max(0, unreadData);
    if (unreadData && typeof (unreadData as { count?: number }).count === 'number') {
      return Math.max(0, (unreadData as { count?: number }).count!);
    }
    return 0;
  }, [unreadData]);

  // Safely extract notifications array with zero chance of undefined/object crash
  const notifications: NotificationItem[] = React.useMemo(() => {
    if (!notificationsData) return [];
    if (Array.isArray(notificationsData)) return notificationsData;
    if (
      typeof notificationsData === 'object' &&
      notificationsData !== null &&
      'notifications' in notificationsData &&
      Array.isArray((notificationsData as { notifications?: NotificationItem[] }).notifications)
    ) {
      return (notificationsData as { notifications: NotificationItem[] }).notifications;
    }
    return [];
  }, [notificationsData]);

  const safeNotifications = Array.isArray(notifications) ? notifications : [];

  const handleMarkAll = async () => {
    try {
      await markAllAsRead().unwrap();
      toast.success('All notifications marked as read');
    } catch {
      toast.error('Failed to mark notifications as read');
    }
  };

  const handleItemClick = async (item: NotificationItem) => {
    const isItemRead = Boolean(item.isRead ?? (item as unknown as { read?: boolean }).read);
    if (!isItemRead) {
      try {
        await markAsRead(item.id).unwrap();
      } catch {
        // Silently catch mutation error
      }
    }
    if (item.actionUrl) {
      router.push(item.actionUrl);
    }
  };

  const handleRefresh = () => {
    refetchNotifications();
    refetchUnread();
  };

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label="View notifications"
          className="relative p-2 rounded-md bg-zinc-900/60 hover:bg-zinc-800 border border-zinc-800 text-zinc-300 hover:text-zinc-100 transition-colors cursor-pointer"
        >
          <Bell className="h-4 w-4" />
          {unreadCount > 0 && (
            <span className="absolute top-1 right-1 flex h-4 w-4 items-center justify-center rounded-full bg-rose-600 text-[9px] font-bold text-white shadow-sm ring-1 ring-zinc-950">
              {unreadCount > 9 ? '9+' : unreadCount}
            </span>
          )}
        </button>
      </PopoverTrigger>

      <PopoverContent
        align="end"
        sideOffset={8}
        className="w-80 p-0 bg-zinc-900 border-zinc-800 text-zinc-100 shadow-2xl rounded-md"
      >
        {/* Header */}
        <div className="flex items-center justify-between p-3 border-b border-zinc-800">
          <div className="flex items-center gap-2">
            <span className="text-xs font-semibold text-zinc-100">Notifications</span>
            {unreadCount > 0 && (
              <span className="px-1.5 py-0.2 rounded-sm bg-rose-950/80 border border-rose-900/60 text-rose-300 text-[10px] font-semibold">
                {unreadCount} unread
              </span>
            )}
          </div>
          <div className="flex items-center gap-2">
            {unreadCount > 0 && (
              <button
                type="button"
                onClick={handleMarkAll}
                disabled={isMarkingAll}
                className="text-[11px] text-zinc-400 hover:text-zinc-200 flex items-center gap-1 cursor-pointer transition-colors disabled:opacity-50"
              >
                <CheckCheck className="h-3 w-3" />
                <span>Mark all read</span>
              </button>
            )}
            <button
              type="button"
              onClick={handleRefresh}
              aria-label="Refresh notifications"
              className="p-1 rounded text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 transition-colors"
            >
              <RefreshCw className="h-3 w-3" />
            </button>
          </div>
        </div>

        {/* Notification List */}
        <div className="max-h-80 overflow-y-auto divide-y divide-zinc-800/60">
          {isLoading ? (
            <div className="py-8 text-center text-xs text-zinc-500">
              Loading notifications...
            </div>
          ) : isError ? (
            <div className="py-8 text-center space-y-2 px-4">
              <AlertCircle className="h-6 w-6 text-amber-500/80 mx-auto" />
              <p className="text-xs text-zinc-300 font-medium">Unable to load notifications</p>
              <button
                type="button"
                onClick={handleRefresh}
                className="text-[11px] text-zinc-300 hover:text-white px-2.5 py-1 rounded bg-zinc-800 hover:bg-zinc-700 transition-colors cursor-pointer"
              >
                Retry
              </button>
            </div>
          ) : safeNotifications.length === 0 ? (
            <div className="py-8 text-center space-y-1.5">
              <Inbox className="h-6 w-6 text-zinc-600 mx-auto" />
              <p className="text-xs text-zinc-400 font-medium">All caught up!</p>
              <p className="text-[11px] text-zinc-500">No new alerts or activity.</p>
            </div>
          ) : (
            safeNotifications.map((item) => {
              const isItemRead = Boolean(
                item.isRead ?? (item as unknown as { read?: boolean }).read
              );
              return (
                <div
                  key={item.id}
                  onClick={() => handleItemClick(item)}
                  className={`p-3 text-left hover:bg-zinc-850/60 transition-colors cursor-pointer flex gap-2.5 ${
                    !isItemRead ? 'bg-zinc-950/40' : 'opacity-75'
                  }`}
                >
                  <div className="h-6 w-6 rounded-sm bg-zinc-800 flex items-center justify-center text-zinc-400 shrink-0 mt-0.5">
                    <MessageSquare className="h-3.5 w-3.5" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-1">
                      <span className="text-xs font-semibold text-zinc-200 truncate">
                        {item.title}
                      </span>
                      {!isItemRead && (
                        <span className="h-1.5 w-1.5 rounded-full bg-rose-500 shrink-0" />
                      )}
                    </div>
                    <p className="text-[11px] text-zinc-400 mt-0.5 line-clamp-2">
                      {item.message}
                    </p>
                    <span className="text-[9px] text-zinc-500 mt-1 block">
                      {item.createdAt
                        ? new Date(item.createdAt).toLocaleDateString(undefined, {
                            month: 'short',
                            day: 'numeric',
                            hour: '2-digit',
                            minute: '2-digit',
                          })
                        : ''}
                    </span>
                  </div>
                </div>
              );
            })
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}
