'use client';

import * as React from 'react';
import { useState } from 'react';
import {
  Trash2,
  X,
  CheckCircle,
  ChevronDown,
  UserPlus,
  Check,
  Loader2,
  AlertTriangle,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { toast } from 'sonner';
import {
  useBulkUpdateTasksStatusMutation,
  useBulkAssignTasksMutation,
  useBulkDeleteTasksMutation,
} from '@/lib/store/api/tasksApi';

interface BulkMember {
  id: string;
  userId?: string;
  user?: {
    id: string;
    firstName?: string | null;
    lastName?: string | null;
    email: string;
    username?: string | null;
    avatar?: string | null;
  };
}

interface BulkStatus {
  id: string;
  name: string;
  color?: string | null;
  category?: string;
}

interface BulkActionsBarProps {
  selectedCount: number;
  totalCount: number;
  selectedTaskIds: string[];
  projectId: string;
  availableStatuses?: BulkStatus[];
  availableMembers?: BulkMember[];
  onSelectAll?: () => void;
  onClearSelection: () => void;
  canManage?: boolean;
}

export const BulkActionsBar: React.FC<BulkActionsBarProps> = ({
  selectedCount,
  totalCount,
  selectedTaskIds,
  projectId,
  availableStatuses = [],
  availableMembers = [],
  onSelectAll,
  onClearSelection,
  canManage = true,
}) => {
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const [bulkUpdateStatus, { isLoading: isUpdatingStatus }] = useBulkUpdateTasksStatusMutation();
  const [bulkAssign, { isLoading: isAssigning }] = useBulkAssignTasksMutation();
  const [bulkDelete, { isLoading: isDeleting }] = useBulkDeleteTasksMutation();

  if (selectedCount === 0) return null;

  const isAllSelected = selectedCount >= totalCount;

  const handleStatusSelect = async (statusId: string) => {
    try {
      const res = await bulkUpdateStatus({
        taskIds: selectedTaskIds,
        statusId,
        projectId,
      }).unwrap();

      toast.success(`Updated status for ${res.updatedCount || selectedCount} tasks`);
      onClearSelection();
    } catch (err: unknown) {
      let message = 'Failed to update tasks status';
      if (err && typeof err === 'object' && 'data' in err) {
        const data = (err as { data?: { message?: string } }).data;
        if (data?.message && typeof data.message === 'string') {
          message = data.message;
        }
      }
      toast.error(message);
    }
  };

  const handleAssignSelect = async (userId: string) => {
    try {
      const res = await bulkAssign({
        taskIds: selectedTaskIds,
        assigneeIds: [userId],
        projectId,
      }).unwrap();

      toast.success(`Assigned ${res.assignedCount || selectedCount} tasks`);
      onClearSelection();
    } catch (err: unknown) {
      let message = 'Failed to assign tasks';
      if (err && typeof err === 'object' && 'data' in err) {
        const data = (err as { data?: { message?: string } }).data;
        if (data?.message && typeof data.message === 'string') {
          message = data.message;
        }
      }
      toast.error(message);
    }
  };

  const handleConfirmDelete = async () => {
    try {
      const res = await bulkDelete({
        taskIds: selectedTaskIds,
        projectId,
      }).unwrap();

      toast.success(`Deleted ${res.deletedCount || selectedCount} tasks`);
      setShowDeleteConfirm(false);
      onClearSelection();
    } catch (err: unknown) {
      let message = 'Failed to delete tasks';
      if (err && typeof err === 'object' && 'data' in err) {
        const data = (err as { data?: { message?: string } }).data;
        if (data?.message && typeof data.message === 'string') {
          message = data.message;
        }
      }
      toast.error(message);
    }
  };

  return (
    <>
      <div className="fixed bottom-8 left-1/2 -translate-x-1/2 z-50 animate-in slide-in-from-bottom-6 duration-200">
        <div className="bg-zinc-900/95 border border-zinc-800 rounded-xl shadow-2xl backdrop-blur-md px-3 py-2 flex items-center gap-3 text-xs text-zinc-200">
          {/* Count Badge & Select All */}
          <div className="flex items-center gap-2 pr-3 border-r border-zinc-800">
            <span className="flex items-center justify-center bg-indigo-600 text-white text-[11px] font-bold h-5 px-1.5 rounded-sm">
              {selectedCount}
            </span>
            <span className="font-medium text-zinc-300">
              {selectedCount === 1 ? 'task selected' : 'tasks selected'}
            </span>

            {onSelectAll && totalCount > selectedCount && (
              <button
                type="button"
                onClick={onSelectAll}
                className="text-[11px] text-indigo-400 hover:text-indigo-300 font-medium ml-1 transition-colors cursor-pointer"
              >
                Select all {totalCount}
              </button>
            )}
          </div>

          {/* Action Buttons */}
          <div className="flex items-center gap-1.5">
            {/* Update Status Dropdown */}
            {canManage && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={isUpdatingStatus}
                    className="h-8 px-2.5 gap-1.5 text-xs text-zinc-300 hover:text-zinc-100 hover:bg-zinc-800 cursor-pointer"
                  >
                    {isUpdatingStatus ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin text-indigo-400" />
                    ) : (
                      <CheckCircle className="h-3.5 w-3.5 text-zinc-400" />
                    )}
                    <span>Status</span>
                    <ChevronDown className="h-3 w-3 opacity-60" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="center"
                  className="w-48 bg-zinc-900 border-zinc-800 text-zinc-200 p-1 rounded-md shadow-xl"
                >
                  <DropdownMenuLabel className="text-[10px] uppercase font-bold text-zinc-500 tracking-wider px-2 py-1">
                    Set Status To
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator className="bg-zinc-800" />
                  <div className="max-h-56 overflow-y-auto py-1 space-y-0.5">
                    {availableStatuses.map((status) => (
                      <DropdownMenuItem
                        key={status.id}
                        onClick={() => handleStatusSelect(status.id)}
                        className="flex items-center gap-2 px-2 py-1.5 rounded text-xs text-zinc-300 hover:bg-zinc-800 hover:text-zinc-100 cursor-pointer"
                      >
                        <span
                          className="h-2 w-2 rounded-full shrink-0"
                          style={{ backgroundColor: status.color || '#a1a1aa' }}
                        />
                        <span className="truncate">{status.name}</span>
                      </DropdownMenuItem>
                    ))}
                    {availableStatuses.length === 0 && (
                      <div className="p-2 text-center text-xs text-zinc-500">
                        No statuses available
                      </div>
                    )}
                  </div>
                </DropdownMenuContent>
              </DropdownMenu>
            )}

            {/* Assign Member Dropdown */}
            {canManage && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={isAssigning}
                    className="h-8 px-2.5 gap-1.5 text-xs text-zinc-300 hover:text-zinc-100 hover:bg-zinc-800 cursor-pointer"
                  >
                    {isAssigning ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin text-indigo-400" />
                    ) : (
                      <UserPlus className="h-3.5 w-3.5 text-zinc-400" />
                    )}
                    <span>Assign</span>
                    <ChevronDown className="h-3 w-3 opacity-60" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="center"
                  className="w-56 bg-zinc-900 border-zinc-800 text-zinc-200 p-1 rounded-md shadow-xl"
                >
                  <DropdownMenuLabel className="text-[10px] uppercase font-bold text-zinc-500 tracking-wider px-2 py-1">
                    Assign To Member
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator className="bg-zinc-800" />
                  <div className="max-h-56 overflow-y-auto py-1 space-y-0.5">
                    {availableMembers.map((member) => {
                      const u = member.user;
                      const uid = u?.id || member.userId || member.id;
                      const displayName = u
                        ? `${u.firstName || ''} ${u.lastName || ''}`.trim() || u.username || u.email
                        : 'Member';
                      const initials = displayName.slice(0, 2).toUpperCase();

                      return (
                        <DropdownMenuItem
                          key={uid}
                          onClick={() => handleAssignSelect(uid)}
                          className="flex items-center gap-2.5 px-2 py-1.5 rounded text-xs text-zinc-300 hover:bg-zinc-800 hover:text-zinc-100 cursor-pointer"
                        >
                          <Avatar className="h-5 w-5 rounded-sm border border-zinc-800">
                            <AvatarFallback className="text-[9px] bg-zinc-800 text-zinc-300">
                              {initials}
                            </AvatarFallback>
                          </Avatar>
                          <span className="truncate">{displayName}</span>
                        </DropdownMenuItem>
                      );
                    })}
                    {availableMembers.length === 0 && (
                      <div className="p-2 text-center text-xs text-zinc-500">
                        No members available
                      </div>
                    )}
                  </div>
                </DropdownMenuContent>
              </DropdownMenu>
            )}

            {/* Bulk Delete Button */}
            {canManage && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowDeleteConfirm(true)}
                disabled={isDeleting}
                className="h-8 px-2.5 gap-1.5 text-xs text-rose-400 hover:text-rose-300 hover:bg-rose-950/30 cursor-pointer"
                title="Delete selected tasks"
              >
                <Trash2 className="h-3.5 w-3.5" />
                <span>Delete</span>
              </Button>
            )}

            {/* Deselect / Close Button */}
            <button
              type="button"
              onClick={onClearSelection}
              className="p-1 rounded hover:bg-zinc-800 text-zinc-400 hover:text-zinc-200 cursor-pointer transition-colors ml-1"
              title="Clear selection"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      <Dialog open={showDeleteConfirm} onOpenChange={setShowDeleteConfirm}>
        <DialogContent className="sm:max-w-md bg-zinc-950 border-zinc-800 text-zinc-200">
          <DialogHeader>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-md bg-rose-950/40 border border-rose-800/40 flex items-center justify-center text-rose-400 shrink-0">
                <AlertTriangle className="h-5 w-5" />
              </div>
              <div>
                <DialogTitle className="text-sm font-semibold text-zinc-100">
                  Delete {selectedCount} {selectedCount === 1 ? 'Task' : 'Tasks'}?
                </DialogTitle>
                <DialogDescription className="text-xs text-zinc-400 mt-1">
                  Are you sure you want to delete the selected {selectedCount === 1 ? 'task' : 'tasks'}? This action cannot be undone.
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>

          <DialogFooter className="mt-4 gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowDeleteConfirm(false)}
              className="border-zinc-800 text-zinc-300 hover:bg-zinc-900"
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={handleConfirmDelete}
              disabled={isDeleting}
              className="bg-rose-600 hover:bg-rose-700 text-white gap-1.5"
            >
              {isDeleting && <Loader2 className="h-3 w-3 animate-spin" />}
              <span>Delete Permanently</span>
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
};

