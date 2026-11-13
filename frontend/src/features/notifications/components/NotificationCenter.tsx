'use client';

import * as React from 'react';
import { useState, useMemo, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  Bell,
  CheckCheck,
  Trash2,
  Search,
  X,
  Filter,
  RefreshCw,
  MessageSquare,
  CheckSquare,
  CheckCircle2,
  Clock,
  UserPlus,
  AlertCircle,
  ExternalLink,
  RotateCcw,
  Inbox,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { toast } from 'sonner';
import {
  useGetNotificationsQuery,
  useGetUnreadCountQuery,
  useMarkAsReadMutation,
  useMarkAllAsReadMutation,
  useDeleteNotificationMutation,
  useDeleteMultipleNotificationsMutation,
  type NotificationItem,
} from '@/lib/store/api/notificationsApi';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { ConfirmationModal } from '@/components/modals/ConfirmationModal';

type NotificationCategory =
  | 'ALL'
  | 'TASKS'
  | 'COMMENTS'
  | 'STATUS'
  | 'INVITATIONS'
  | 'SYSTEM';

const CATEGORY_TABS: Array<{ value: NotificationCategory; label: string }> = [
  { value: 'ALL', label: 'All' },
  { value: 'TASKS', label: 'Tasks' },
  { value: 'COMMENTS', label: 'Comments' },
  { value: 'STATUS', label: 'Status Updates' },
  { value: 'INVITATIONS', label: 'Invitations' },
  { value: 'SYSTEM', label: 'System' },
];

export function NotificationCenter() {
  const router = useRouter();

  const [category, setCategory] = useState<NotificationCategory>('ALL');
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [page, setPage] = useState(1);
  const [isClearReadModalOpen, setIsClearReadModalOpen] = useState(false);
  const pageSize = 15;

  // Debounce search
  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedSearch(searchQuery.trim().toLowerCase());
    }, 250);
    return () => clearTimeout(handler);
  }, [searchQuery]);

  const {
    data: notifications = [],
    isLoading,
    isFetching,
    refetch,
  } = useGetNotificationsQuery(
    { isRead: unreadOnly ? false : undefined },
    { pollingInterval: 30000 }
  );

  const { data: unreadCount = 0 } = useGetUnreadCountQuery();

  const [markAsRead] = useMarkAsReadMutation();
  const [markAllAsRead, { isLoading: isMarkingAll }] = useMarkAllAsReadMutation();
  const [deleteNotification] = useDeleteNotificationMutation();
  const [deleteMultiple] = useDeleteMultipleNotificationsMutation();

  // Category matcher
  const matchesCategory = (type: string, cat: NotificationCategory): boolean => {
    const upper = (type || '').toUpperCase();
    switch (cat) {
      case 'ALL':
        return true;
      case 'TASKS':
        return (
          upper.includes('TASK_ASSIGNED') ||
          upper.includes('TASK_CREATED') ||
          upper.includes('TASK_DUE')
        );
      case 'COMMENTS':
        return upper.includes('COMMENT') || upper.includes('MENTION');
      case 'STATUS':
        return upper.includes('STATUS') || upper.includes('COMPLETED');
      case 'INVITATIONS':
        return upper.includes('INVIT') || upper.includes('MEMBER');
      case 'SYSTEM':
        return upper.includes('SYSTEM') || upper.includes('PROJECT');
      default:
        return true;
    }
  };

  // Category counts
  const categoryCounts = useMemo(() => {
    const counts: Record<NotificationCategory, number> = {
      ALL: notifications.length,
      TASKS: 0,
      COMMENTS: 0,
      STATUS: 0,
      INVITATIONS: 0,
      SYSTEM: 0,
    };
    notifications.forEach((item) => {
      (Object.keys(counts) as NotificationCategory[]).forEach((cat) => {
        if (cat !== 'ALL' && matchesCategory(item.type, cat)) {
          counts[cat]++;
        }
      });
    });
    return counts;
  }, [notifications]);

  // Filtered Notifications
  const filteredNotifications = useMemo(() => {
    return notifications.filter((item) => {
      // Category filter
      if (!matchesCategory(item.type, category)) return false;

      // Search filter
      if (debouncedSearch) {
        const titleMatch = item.title.toLowerCase().includes(debouncedSearch);
        const msgMatch = item.message.toLowerCase().includes(debouncedSearch);
        if (!titleMatch && !msgMatch) return false;
      }
      return true;
    });
  }, [notifications, category, debouncedSearch]);

  // Pagination slice
  const totalPages = Math.max(1, Math.ceil(filteredNotifications.length / pageSize));
  const currentPage = Math.min(page, totalPages);
  const paginatedItems = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return filteredNotifications.slice(start, start + pageSize);
  }, [filteredNotifications, currentPage, pageSize]);

  const handleMarkAllRead = async () => {
    try {
      await markAllAsRead().unwrap();
      toast.success('All notifications marked as read');
    } catch {
      toast.error('Failed to mark all as read');
    }
  };

  const readNotificationIds = useMemo(
    () => notifications.filter((n) => n.isRead).map((n) => n.id),
    [notifications]
  );

  const handleDeleteRead = () => {
    if (readNotificationIds.length === 0) {
      toast.info('No read notifications to clear');
      return;
    }
    setIsClearReadModalOpen(true);
  };

  const handleConfirmDeleteRead = async () => {
    if (readNotificationIds.length === 0) return;
    try {
      await deleteMultiple(readNotificationIds).unwrap();
      toast.success(`Cleared ${readNotificationIds.length} read notification(s)`);
      setIsClearReadModalOpen(false);
    } catch {
      toast.error('Failed to clear read notifications');
    }
  };

  const handleItemClick = async (item: NotificationItem) => {
    if (!item.isRead) {
      try {
        await markAsRead(item.id).unwrap();
      } catch {
        // Silently catch
      }
    }
    if (item.actionUrl) {
      router.push(item.actionUrl);
    }
  };

  const handleDeleteSingle = async (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    try {
      await deleteNotification(id).unwrap();
      toast.success('Notification removed');
    } catch {
      toast.error('Failed to remove notification');
    }
  };

  // Icon mapping
  const getNotificationIcon = (type: string) => {
    const upper = (type || '').toUpperCase();
    if (upper.includes('COMMENT') || upper.includes('MENTION')) {
      return { icon: MessageSquare, color: 'text-zinc-200 bg-zinc-850 border border-zinc-750' };
    }
    if (upper.includes('STATUS')) {
      return { icon: CheckCircle2, color: 'text-zinc-200 bg-zinc-850 border border-zinc-750' };
    }
    if (upper.includes('DUE')) {
      return { icon: Clock, color: 'text-zinc-200 bg-zinc-850 border border-zinc-750' };
    }
    if (upper.includes('INVIT')) {
      return { icon: UserPlus, color: 'text-zinc-200 bg-zinc-850 border border-zinc-750' };
    }
    if (upper.includes('TASK')) {
      return { icon: CheckSquare, color: 'text-zinc-200 bg-zinc-850 border border-zinc-750' };
    }
    return { icon: Bell, color: 'text-zinc-400 bg-zinc-900 border border-zinc-800' };
  };

  return (
    <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2.5">
            <div className="h-8 w-8 rounded-md bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-100">
              <Bell className="h-4.5 w-4.5" />
            </div>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-zinc-100">
              Notifications Center
            </h1>
            {unreadCount > 0 ? (
              <Badge className="text-xs font-mono font-bold bg-rose-500/20 text-rose-300 border-rose-500/30">
                {unreadCount} Unread
              </Badge>
            ) : (
              <Badge variant="outline" className="text-xs font-mono text-zinc-400">
                All Read
              </Badge>
            )}
          </div>
          <p className="text-xs text-zinc-400">
            Real-time feed of task assignments, mentions, status updates, and workspace activities.
          </p>
        </div>

        {/* Global Header Actions */}
        <div className="flex items-center gap-2 flex-wrap">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            disabled={isFetching}
            className="text-xs gap-1.5"
            title="Refresh notifications"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isFetching ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </Button>

          {unreadCount > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleMarkAllRead}
              disabled={isMarkingAll}
              className="text-xs gap-1.5"
            >
              <CheckCheck className="w-3.5 h-3.5 text-zinc-300" />
              <span>Mark All Read</span>
            </Button>
          )}

          <Button
            variant="outline"
            size="sm"
            onClick={handleDeleteRead}
            className="text-xs gap-1.5 text-zinc-400 hover:text-rose-400"
            title="Delete all read notifications"
          >
            <Trash2 className="w-3.5 h-3.5" />
            <span>Clear Read</span>
          </Button>
        </div>
      </div>

      {/* Toolbar: Search, Unread Toggle & Categories */}
      <div className="space-y-3 bg-zinc-950/60 p-4 rounded-md border border-zinc-850">
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
          {/* Search Box */}
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search notifications by title or message..."
              className="pl-8.5 pr-8 py-1.5 text-xs bg-zinc-900 border-zinc-800 text-zinc-200 placeholder:text-zinc-500 focus:border-zinc-700"
            />
            {searchQuery && (
              <button
                type="button"
                onClick={() => setSearchQuery('')}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 transition-colors p-0.5"
                title="Clear search"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            )}
          </div>

          {/* Unread Only Toggle */}
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setUnreadOnly(false)}
              className={`px-3 py-1 rounded-md text-xs font-medium transition-colors cursor-pointer ${
                !unreadOnly
                  ? 'bg-zinc-100 text-zinc-950 font-bold'
                  : 'bg-zinc-900 text-zinc-400 hover:text-zinc-200'
              }`}
            >
              All Notifications
            </button>
            <button
              type="button"
              onClick={() => setUnreadOnly(true)}
              className={`px-3 py-1 rounded-md text-xs font-medium transition-colors cursor-pointer flex items-center gap-1.5 ${
                unreadOnly
                  ? 'bg-zinc-100 text-zinc-950 font-bold'
                  : 'bg-zinc-900 text-zinc-400 hover:text-zinc-200'
              }`}
            >
              <span>Unread Only</span>
              {unreadCount > 0 && (
                <span className="text-[10px] font-mono px-1 rounded-full bg-rose-500 text-white font-bold">
                  {unreadCount}
                </span>
              )}
            </button>
          </div>
        </div>

        {/* Category Pills */}
        <div className="flex items-center gap-1.5 flex-wrap pt-1 border-t border-zinc-850/60">
          <span className="text-[11px] font-medium text-zinc-500 flex items-center gap-1 mr-1">
            <Filter className="w-3 h-3" />
            <span>Category:</span>
          </span>
          {CATEGORY_TABS.map((tab) => {
            const isSelected = category === tab.value;
            const count = categoryCounts[tab.value] || 0;

            return (
              <button
                key={tab.value}
                type="button"
                onClick={() => {
                  setCategory(tab.value);
                  setPage(1);
                }}
                className={`flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-colors cursor-pointer ${
                  isSelected
                    ? 'bg-zinc-100 text-zinc-950 font-semibold shadow-xs'
                    : 'bg-zinc-900 text-zinc-400 border border-zinc-800 hover:text-zinc-200 hover:border-zinc-700'
                }`}
              >
                <span>{tab.label}</span>
                <span
                  className={`text-[10px] font-mono px-1 rounded-full ${
                    isSelected ? 'bg-zinc-950/60 text-zinc-100' : 'bg-zinc-800 text-zinc-400'
                  }`}
                >
                  {count}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Notifications List Container */}
      <Card className="border-zinc-800 bg-zinc-900/80 divide-y divide-zinc-850 overflow-hidden shadow-sm">
        {isLoading ? (
          <div className="divide-y divide-zinc-850/80">
            {[1, 2, 3, 4, 5].map((idx) => (
              <div
                key={idx}
                className="p-4 flex items-start justify-between gap-3 bg-zinc-950/20"
              >
                <div className="flex items-start gap-3 flex-1 min-w-0">
                  <Skeleton className="h-8 w-8 rounded-full bg-zinc-800 shrink-0 mt-0.5" />
                  <div className="space-y-1.5 flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <Skeleton className="h-3.5 w-36 bg-zinc-800" />
                      <Skeleton className="h-3 w-16 rounded bg-zinc-850" />
                    </div>
                    <Skeleton
                      className={`h-3 bg-zinc-850 ${
                        idx % 2 === 0 ? 'w-4/5' : 'w-3/5'
                      }`}
                    />
                    <Skeleton className="h-2.5 w-20 bg-zinc-850/60" />
                  </div>
                </div>
                <Skeleton className="h-6 w-6 rounded bg-zinc-850 shrink-0" />
              </div>
            ))}
          </div>
        ) : filteredNotifications.length === 0 ? (
          <div className="py-16 text-center space-y-3 px-4">
            <div className="w-12 h-12 rounded-full bg-zinc-950 border border-zinc-800 flex items-center justify-center text-zinc-500 mx-auto">
              <Inbox className="w-6 h-6" />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-zinc-200">
                {searchQuery || category !== 'ALL' || unreadOnly
                  ? 'No notifications match your filters'
                  : 'All caught up!'}
              </h3>
              <p className="text-xs text-zinc-500 mt-1 max-w-sm mx-auto">
                {searchQuery || category !== 'ALL' || unreadOnly
                  ? 'Try clearing filters or search terms to see all notifications.'
                  : 'You have zero unread notifications. New alerts will show up here.'}
              </p>
            </div>
            {(searchQuery || category !== 'ALL' || unreadOnly) && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setSearchQuery('');
                  setCategory('ALL');
                  setUnreadOnly(false);
                }}
                className="text-xs gap-1.5"
              >
                <RotateCcw className="w-3.5 h-3.5" />
                <span>Reset Filters</span>
              </Button>
            )}
          </div>
        ) : (
          paginatedItems.map((item) => {
            const { icon: TypeIcon, color } = getNotificationIcon(item.type);
            const isRead = item.isRead;

            return (
              <div
                key={item.id}
                onClick={() => handleItemClick(item)}
                className={`p-4 flex items-start justify-between gap-4 hover:bg-zinc-850/40 transition-colors cursor-pointer group ${
                  !isRead ? 'bg-zinc-950/40' : 'opacity-85'
                }`}
              >
                {/* Left: Icon & Notification Info */}
                <div className="flex items-start gap-3.5 min-w-0">
                  <div
                    className={`w-9 h-9 rounded-md flex items-center justify-center shrink-0 border border-zinc-800 ${color}`}
                  >
                    <TypeIcon className="w-4 h-4" />
                  </div>

                  <div className="space-y-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="text-xs font-semibold text-zinc-100 group-hover:text-white">
                        {item.title}
                      </span>
                      {!isRead && (
                        <span className="w-2 h-2 rounded-full bg-rose-500 shadow-xs shadow-rose-500/50 shrink-0" />
                      )}
                      <Badge
                        variant="outline"
                        className="text-[9px] font-mono uppercase px-1.5 py-0 bg-zinc-950 text-zinc-500 border-zinc-800"
                      >
                        {item.type.replace(/_/g, ' ')}
                      </Badge>
                    </div>

                    <p className="text-xs text-zinc-300 leading-relaxed break-words">
                      {item.message}
                    </p>

                    <div className="flex items-center gap-3 text-[11px] text-zinc-500 pt-0.5">
                      <span className="flex items-center gap-1">
                        <Clock className="w-3 h-3 text-zinc-500" />
                        {new Date(item.createdAt).toLocaleDateString(undefined, {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </span>

                      {item.actionUrl && (
                        <span className="text-violet-400 group-hover:text-violet-300 flex items-center gap-1 font-medium">
                          <span>View target</span>
                          <ExternalLink className="w-2.5 h-2.5" />
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                {/* Right: Quick Delete Button */}
                <button
                  type="button"
                  onClick={(e) => handleDeleteSingle(e, item.id)}
                  className="opacity-0 group-hover:opacity-100 p-1.5 rounded-md text-zinc-500 hover:text-rose-400 hover:bg-zinc-800 transition-all cursor-pointer shrink-0 mt-0.5"
                  title="Delete notification"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
            );
          })
        )}
      </Card>

      {/* Pagination Bar */}
      {filteredNotifications.length > pageSize && (
        <div className="flex items-center justify-between px-2 pt-2 text-xs text-zinc-400">
          <span>
            Showing {(currentPage - 1) * pageSize + 1}–
            {Math.min(currentPage * pageSize, filteredNotifications.length)} of{' '}
            {filteredNotifications.length} notifications
          </span>

          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={currentPage <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              className="text-xs gap-1"
            >
              <ChevronLeft className="w-3.5 h-3.5" /> Previous
            </Button>
            <span className="font-mono text-xs px-2 text-zinc-300 font-semibold">
              {currentPage} / {totalPages}
            </span>
            <Button
              variant="outline"
              size="sm"
              disabled={currentPage >= totalPages}
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              className="text-xs gap-1"
            >
              Next <ChevronRight className="w-3.5 h-3.5" />
            </Button>
          </div>
        </div>
      )}

      {/* Clear Read Confirmation Modal */}
      <ConfirmationModal
        open={isClearReadModalOpen}
        onOpenChange={setIsClearReadModalOpen}
        title="Clear Read Notifications"
        description={`Are you sure you want to delete ${readNotificationIds.length} read notification${
          readNotificationIds.length === 1 ? '' : 's'
        }? This action cannot be undone.`}
        confirmLabel="Clear Notifications"
        variant="destructive"
        onConfirm={handleConfirmDeleteRead}
      />
    </div>
  );
}

