'use client';

import * as React from 'react';
import { useState, useEffect, useRef, useMemo } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import {
  X,
  Maximize2,
  ArrowLeft,
  CheckSquare,
  AlertCircle,
  Bookmark,
  Zap,
  CornerDownRight,
  Share2,
  Trash2,
  Check,
  Sparkles,
  Loader2,
  Calendar,
  User,
  UserCheck,
  Timer,
  Hash,
  Layers,
  Copy,
  MessageSquare,
  History,
  Clock,
  Send,
  Pencil,
  AlertTriangle,
} from 'lucide-react';
import { toast } from 'sonner';
import type { Task, TaskPriority, TaskType, TaskStatus } from '@orbito/types';
import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { setTaskDrawerId } from '@/lib/store/slices/uiSlice';
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';
import {
  useGetTaskQuery,
  useUpdateTaskMutation,
  useDeleteTaskMutation,
  useGetTaskCommentsQuery,
  useCreateTaskCommentMutation,
  useDeleteTaskCommentMutation,
  useUpdateTaskAssigneesMutation,
  useGetTaskStatusesByProjectQuery,
} from '@/lib/store/api/tasksApi';
import { useGetTaskActivityLogsQuery } from '@/lib/store/api/activityLogApi';
import { useGenerateDescriptionMutation } from '@/lib/store/api/aiChatApi';
import { useGetWorkspaceMembersQuery } from '@/lib/store/api/workspacesApi';
import { useGetSprintsQuery } from '@/lib/store/api/sprintsApi';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { DualModeEditor } from '@/components/common/DualModeEditor';
import ReactMarkdown from 'react-markdown';
import { TaskHierarchy } from './TaskHierarchy';
import { TaskLabels } from './TaskLabels';
import { TaskTimeTracking } from './TaskTimeTracking';
import { TaskAttachments } from './TaskAttachments';
import { TaskDependencies } from './TaskDependencies';
import { TaskRecurrenceSelector } from './TaskRecurrenceSelector';
import { ShareTaskDialog } from './ShareTaskDialog';
import { MentionAutocomplete, type MentionMember } from './MentionAutocomplete';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { ConfirmationModal } from '@/components/modals/ConfirmationModal';

export interface TaskDetailWorkspaceProps {
  taskId: string;
  workspaceSlug?: string;
  projectSlug?: string;
  mode: 'drawer' | 'fullpage';
  onClose?: () => void;
}

const TASK_TYPE_CONFIG: Record<
  TaskType,
  {
    label: string;
    icon: React.ComponentType<{ className?: string }>;
    badgeClass: string;
    iconClass: string;
  }
> = {
  TASK: {
    label: 'Task',
    icon: CheckSquare,
    badgeClass: 'bg-zinc-800 text-zinc-300 border-zinc-700',
    iconClass: 'text-zinc-400',
  },
  BUG: {
    label: 'Bug',
    icon: AlertCircle,
    badgeClass: 'bg-rose-950/40 text-rose-300 border-rose-800/50',
    iconClass: 'text-rose-400',
  },
  STORY: {
    label: 'Story',
    icon: Bookmark,
    badgeClass: 'bg-emerald-950/40 text-emerald-300 border-emerald-800/50',
    iconClass: 'text-emerald-400',
  },
  EPIC: {
    label: 'Epic',
    icon: Zap,
    badgeClass: 'bg-purple-950/40 text-purple-300 border-purple-800/50',
    iconClass: 'text-purple-400',
  },
  SUBTASK: {
    label: 'Subtask',
    icon: CornerDownRight,
    badgeClass: 'bg-sky-950/40 text-sky-300 border-sky-800/50',
    iconClass: 'text-sky-400',
  },
};

const TASK_PRIORITIES: Array<{
  value: TaskPriority;
  label: string;
  dotColor: string;
}> = [
  { value: 'LOW' as TaskPriority, label: 'Low', dotColor: 'bg-blue-400' },
  { value: 'MEDIUM' as TaskPriority, label: 'Medium', dotColor: 'bg-amber-400' },
  { value: 'HIGH' as TaskPriority, label: 'High', dotColor: 'bg-orange-400' },
  { value: 'URGENT' as TaskPriority, label: 'Urgent', dotColor: 'bg-rose-500' },
];

const STORY_POINTS_OPTIONS: Array<number | null> = [null, 1, 2, 3, 5, 8, 13, 21];

export function TaskDetailWorkspace({
  taskId,
  workspaceSlug: propWorkspaceSlug,
  projectSlug: propProjectSlug,
  mode,
  onClose,
}: TaskDetailWorkspaceProps) {
  const router = useRouter();
  const dispatch = useAppDispatch();

  const authWorkspaceSlug = useAppSelector((state) => state.auth.activeWorkspaceSlug);
  const authProjectSlug = useAppSelector((state) => state.auth.activeProjectSlug);
  const activeWorkspaceId = useAppSelector((state) => state.auth.activeWorkspaceId);
  const currentUser = useAppSelector((state) => state.auth.user);
  const { isViewer, isManager, isOwner, isSuperAdmin } = useAuthRbac();

  const canEdit = !isViewer;
  const canDelete = isManager || isOwner || isSuperAdmin;

  // Task Query
  const {
    data: task,
    isLoading: isTaskLoading,
    isError,
  } = useGetTaskQuery(taskId, { skip: !taskId });

  // Effective Slugs
  const effectiveWorkspaceSlug =
    propWorkspaceSlug ||
    authWorkspaceSlug ||
    '';
  const effectiveProjectSlug =
    propProjectSlug || task?.project?.slug || authProjectSlug || '';

  // Comments & Activities
  const { data: comments = [], isLoading: isCommentsLoading } =
    useGetTaskCommentsQuery(taskId, { skip: !taskId });

  const [bottomTab, setBottomTab] = useState<'comments' | 'activity'>('comments');

  const { data: activityData, isLoading: isActivityLoading } =
    useGetTaskActivityLogsQuery(taskId, {
      skip: !taskId || bottomTab !== 'activity',
    });
  const activities = activityData?.activities || [];

  // Project Members, Statuses & Sprints
  const { data: workspaceMembers = [] } = useGetWorkspaceMembersQuery(
    activeWorkspaceId || '',
    { skip: !activeWorkspaceId }
  );

  const { data: projectStatuses = [] } = useGetTaskStatusesByProjectQuery(
    task?.projectId || '',
    { skip: !task?.projectId }
  );

  const { data: sprints = [] } = useGetSprintsQuery(
    { projectId: task?.projectId },
    { skip: !task?.projectId }
  );

  // Mutations
  const [updateTask, { isLoading: isUpdating }] = useUpdateTaskMutation();
  const [deleteTask, { isLoading: isDeleting }] = useDeleteTaskMutation();
  const [createComment, { isLoading: isPostingComment }] =
    useCreateTaskCommentMutation();
  const [deleteComment] = useDeleteTaskCommentMutation();
  const [updateAssignees] = useUpdateTaskAssigneesMutation();
  const [generateDescription, { isLoading: isGeneratingDesc }] =
    useGenerateDescriptionMutation();

  // Local Component State
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [titleDraft, setTitleDraft] = useState('');
  const [isEditingDescription, setIsEditingDescription] = useState(false);
  const [descriptionDraft, setDescriptionDraft] = useState('');
  const [commentText, setCommentText] = useState('');
  const [mentionQuery, setMentionQuery] = useState('');
  const [isMentionOpen, setIsMentionOpen] = useState(false);
  const commentInputRef = useRef<HTMLInputElement>(null);
  const [isShareDialogOpen, setIsShareDialogOpen] = useState(false);
  const [isCopiedSlug, setIsCopiedSlug] = useState(false);
  const [isDeleteTaskModalOpen, setIsDeleteTaskModalOpen] = useState(false);
  const [commentToDelete, setCommentToDelete] = useState<string | null>(null);

  const mentionMembers: MentionMember[] = useMemo(() => {
    return workspaceMembers.map((m) => {
      const name =
        m.user?.firstName && m.user?.lastName
          ? `${m.user.firstName} ${m.user.lastName}`
          : m.user?.username || m.user?.email || 'Member';
      return {
        id: m.userId || m.user?.id || m.id,
        name,
        email: m.user?.email,
        avatar: m.user?.avatar,
        role: m.role,
      };
    });
  }, [workspaceMembers]);

  const handleCommentChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setCommentText(val);

    const cursorPos = e.target.selectionStart || val.length;
    const textBeforeCursor = val.slice(0, cursorPos);
    const atMatch = textBeforeCursor.match(/(?:^|\s)@([a-zA-Z0-9_]*)$/);

    if (atMatch) {
      setMentionQuery(atMatch[1] || '');
      setIsMentionOpen(true);
    } else {
      setIsMentionOpen(false);
    }
  };

  const handleSelectMention = (member: MentionMember) => {
    const cursorPos = commentInputRef.current?.selectionStart || commentText.length;
    const textBeforeCursor = commentText.slice(0, cursorPos);
    const textAfterCursor = commentText.slice(cursorPos);

    const updatedBefore = textBeforeCursor.replace(/(?:^|\s)@([a-zA-Z0-9_]*)$/, (match) => {
      const prefix = match.startsWith(' ') ? ' ' : '';
      return `${prefix}@[${member.name}] `;
    });

    const newText = updatedBefore + textAfterCursor;
    setCommentText(newText);
    setIsMentionOpen(false);

    setTimeout(() => {
      if (commentInputRef.current) {
        commentInputRef.current.focus();
        const newCursor = updatedBefore.length;
        commentInputRef.current.setSelectionRange(newCursor, newCursor);
      }
    }, 0);
  };

  useEffect(() => {
    if (task) {
      setTitleDraft(task.title || '');
      setDescriptionDraft(task.description || '');
    }
  }, [task]);

  if (!taskId) return null;

  const handleCopyTaskSlug = async () => {
    if (!task) return;
    const taskKey = task.slug || `TASK-${task.taskNumber || task.id.slice(0, 4)}`;
    try {
      await navigator.clipboard.writeText(taskKey);
      setIsCopiedSlug(true);
      toast.success(`Copied ${taskKey} to clipboard`);
      setTimeout(() => setIsCopiedSlug(false), 2000);
    } catch {
      toast.error('Failed to copy to clipboard');
    }
  };

  const handleExpandToFullScreen = () => {
    if (!task) return;
    const taskIdentifier = task.slug || task.id;
    const targetUrl = `/${effectiveWorkspaceSlug}/${effectiveProjectSlug}/tasks/${taskIdentifier}`;
    if (onClose) onClose();
    router.push(targetUrl);
  };

  const handleNavigateToTask = (targetTaskId: string) => {
    if (mode === 'drawer') {
      dispatch(setTaskDrawerId(targetTaskId));
    } else {
      router.push(
        `/${effectiveWorkspaceSlug}/${effectiveProjectSlug}/tasks/${targetTaskId}`
      );
    }
  };

  const handleSaveTitle = async () => {
    if (!task || !canEdit) return;
    const trimmed = titleDraft.trim();
    if (!trimmed) {
      toast.error('Task title cannot be empty');
      return;
    }
    if (trimmed === task.title) {
      setIsEditingTitle(false);
      return;
    }
    try {
      await updateTask({
        id: task.id,
        updates: { title: trimmed },
      }).unwrap();
      setIsEditingTitle(false);
      toast.success('Title updated');
    } catch {
      toast.error('Failed to update title');
    }
  };

  const handleSaveDescription = async () => {
    if (!task || !canEdit) return;
    try {
      await updateTask({
        id: task.id,
        updates: { description: descriptionDraft.trim() },
      }).unwrap();
      setIsEditingDescription(false);
      toast.success('Description updated');
    } catch {
      toast.error('Failed to update description');
    }
  };

  const handleGenerateAiDescription = async () => {
    if (!task || !canEdit) return;
    const sourceTitle = titleDraft.trim() || task.title;
    if (!sourceTitle) {
      toast.error('Task title is required to generate description');
      return;
    }
    try {
      const res = await generateDescription({ title: sourceTitle }).unwrap();
      if (res.description) {
        setDescriptionDraft(res.description);
        await updateTask({
          id: task.id,
          updates: { description: res.description },
        }).unwrap();
        setIsEditingDescription(false);
        toast.success('Description generated & saved with AI');
      }
    } catch {
      toast.error('Failed to generate description. Ensure AI is enabled in settings.');
    }
  };

  const handleStatusChange = async (newStatusId: string) => {
    if (!task || !canEdit) return;
    try {
      await updateTask({
        id: task.id,
        updates: { statusId: newStatusId },
      }).unwrap();
      const matched = projectStatuses.find((s) => s.id === newStatusId);
      toast.success(`Status updated to ${matched?.name || 'new status'}`);
    } catch {
      toast.error('Failed to update status');
    }
  };

  const handlePriorityChange = async (newPriority: TaskPriority) => {
    if (!task || !canEdit) return;
    try {
      await updateTask({
        id: task.id,
        updates: { priority: newPriority },
      }).unwrap();
      toast.success(`Priority updated to ${newPriority}`);
    } catch {
      toast.error('Failed to update priority');
    }
  };

  const handleTaskTypeChange = async (newType: TaskType) => {
    if (!task || !canEdit) return;
    try {
      await updateTask({
        id: task.id,
        updates: { type: newType },
      }).unwrap();
      toast.success(`Task type updated to ${TASK_TYPE_CONFIG[newType]?.label || newType}`);
    } catch {
      toast.error('Failed to update task type');
    }
  };

  const handleSprintChange = async (sprintIdValue: string) => {
    if (!task || !canEdit) return;
    try {
      const sprintId = sprintIdValue === 'none' ? null : sprintIdValue;
      await updateTask({
        id: task.id,
        updates: { sprintId },
      }).unwrap();
      toast.success(
        sprintIdValue === 'none'
          ? 'Moved to Backlog (No sprint)'
          : 'Sprint assignment updated'
      );
    } catch {
      toast.error('Failed to update sprint assignment');
    }
  };

  const handleStoryPointsChange = async (points: number | null) => {
    if (!task || !canEdit) return;
    try {
      await updateTask({
        id: task.id,
        updates: { storyPoints: points },
      }).unwrap();
      toast.success(
        points === null
          ? 'Cleared story points'
          : `Story points set to ${points}`
      );
    } catch {
      toast.error('Failed to update story points');
    }
  };

  const handleStartDateChange = async (val: string) => {
    if (!task || !canEdit) return;
    try {
      const iso = val ? new Date(val).toISOString() : null;
      await updateTask({
        id: task.id,
        updates: { startDate: iso },
      }).unwrap();
      toast.success('Start date updated');
    } catch {
      toast.error('Failed to update start date');
    }
  };

  const handleDueDateChange = async (val: string) => {
    if (!task || !canEdit) return;
    try {
      const iso = val ? new Date(val).toISOString() : null;
      await updateTask({
        id: task.id,
        updates: { dueDate: iso },
      }).unwrap();
      toast.success('Due date updated');
    } catch {
      toast.error('Failed to update due date');
    }
  };

  const handleAssigneeChange = async (userId: string) => {
    if (!task || !canEdit) return;
    try {
      const assigneeIds = userId === 'unassigned' ? [] : [userId];
      await updateAssignees({ taskId: task.id, assigneeIds }).unwrap();
      toast.success(userId === 'unassigned' ? 'Task unassigned' : 'Assignee updated');
    } catch {
      toast.error('Failed to update assignee');
    }
  };

  const handleReporterChange = async (userId: string) => {
    if (!task || !canEdit) return;
    try {
      await updateTask({
        id: task.id,
        updates: { reporterIds: [userId] },
      }).unwrap();
      toast.success('Reporter updated');
    } catch {
      toast.error('Failed to update reporter');
    }
  };

  const handleCommentSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commentText.trim() || !task) return;
    try {
      await createComment({
        taskId: task.id,
        content: commentText.trim(),
      }).unwrap();
      setCommentText('');
      toast.success('Comment posted');
    } catch {
      toast.error('Failed to post comment');
    }
  };

  const handleConfirmDeleteComment = async () => {
    if (!task || !commentToDelete) return;
    try {
      await deleteComment({ id: commentToDelete, taskId: task.id }).unwrap();
      toast.success('Comment deleted');
      setCommentToDelete(null);
    } catch {
      toast.error('Failed to delete comment');
    }
  };

  const handleConfirmDeleteTask = async () => {
    if (!task || !canDelete) return;
    try {
      await deleteTask(task.id).unwrap();
      toast.success('Task deleted');
      setIsDeleteTaskModalOpen(false);
      if (onClose) onClose();
      if (mode === 'fullpage') {
        router.push(`/${effectiveWorkspaceSlug}/${effectiveProjectSlug}/tasks`);
      }
    } catch {
      toast.error('Failed to delete task');
    }
  };

  // Loading State
  if (isTaskLoading) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-12 space-y-4">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-zinc-500 border-t-zinc-100" />
        <span className="text-xs text-zinc-400">Loading task workspace...</span>
      </div>
    );
  }

  // Error State
  if (isError || !task) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-12 text-center space-y-4">
        <div className="h-10 w-10 rounded-full bg-rose-500/10 flex items-center justify-center text-rose-400">
          <AlertTriangle className="h-5 w-5" />
        </div>
        <div>
          <h3 className="text-sm font-bold text-zinc-100">Task Not Found</h3>
          <p className="text-xs text-zinc-400 mt-1">
            The requested task could not be loaded or was removed.
          </p>
        </div>
        {mode === 'drawer' ? (
          <Button variant="outline" size="sm" onClick={onClose}>
            Close Drawer
          </Button>
        ) : (
          <Link
            href={`/${effectiveWorkspaceSlug}/${effectiveProjectSlug}/tasks`}
            className="inline-flex items-center gap-1.5 text-xs font-semibold text-zinc-300 hover:text-zinc-100"
          >
            <ArrowLeft className="w-3.5 h-3.5" /> Back to Tasks Board
          </Link>
        )}
      </div>
    );
  }

  const currentType = (task.taskType || 'TASK') as TaskType;
  const currentTypeMeta = TASK_TYPE_CONFIG[currentType] || TASK_TYPE_CONFIG.TASK;
  const TypeIcon = currentTypeMeta.icon;

  const currentAssignee = task.assignees?.[0]?.user;
  const currentReporter = task.reporters?.[0]?.user;
  const taskSlug = task.slug || `TASK-${task.taskNumber || task.id.slice(0, 4)}`;

  // Relative Due Date Status
  const getDueDateStatus = () => {
    if (!task.dueDate) return null;
    const due = new Date(task.dueDate);
    const now = new Date();
    const isDone = task.status?.category === 'DONE';
    const diffDays = Math.ceil((due.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));

    if (isDone) return null;
    if (diffDays < 0) {
      return {
        label: `Overdue by ${Math.abs(diffDays)}d`,
        className: 'text-rose-400 bg-rose-950/40 border-rose-800/60',
      };
    }
    if (diffDays === 0) {
      return {
        label: 'Due today',
        className: 'text-amber-400 bg-amber-950/40 border-amber-800/60',
      };
    }
    if (diffDays === 1) {
      return {
        label: 'Due tomorrow',
        className: 'text-blue-400 bg-blue-950/40 border-blue-800/60',
      };
    }
    return null;
  };

  const dueStatus = getDueDateStatus();

  return (
    <div
      className={`flex flex-col min-h-full ${
        mode === 'fullpage'
          ? 'max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6'
          : 'h-full bg-zinc-900'
      }`}
    >
      {/* Top Header / Action Bar */}
      <div
        className={`flex items-center justify-between gap-4 border-b border-zinc-800 ${
          mode === 'fullpage' ? 'pb-4' : 'px-6 py-4 bg-zinc-950/40 shrink-0'
        }`}
      >
        {/* Left Header Group */}
        <div className="flex items-center gap-3 flex-wrap">
          {mode === 'fullpage' && (
            <Link
              href={`/${effectiveWorkspaceSlug}/${effectiveProjectSlug}/tasks`}
              className="inline-flex items-center gap-1.5 text-xs font-medium text-zinc-400 hover:text-zinc-200 transition-colors mr-2 cursor-pointer"
            >
              <ArrowLeft className="w-3.5 h-3.5" />
              <span>Back to Board</span>
            </Link>
          )}

          {/* Task Key Slug with 1-Click Copy */}
          <button
            type="button"
            onClick={handleCopyTaskSlug}
            className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-zinc-950 border border-zinc-800 text-xs font-mono font-bold text-zinc-300 hover:border-zinc-700 hover:text-zinc-100 transition-colors cursor-pointer group"
            title="Click to copy task key"
          >
            <span>{taskSlug}</span>
            {isCopiedSlug ? (
              <Check className="w-3 h-3 text-emerald-400" />
            ) : (
              <Copy className="w-3 h-3 text-zinc-500 group-hover:text-zinc-300" />
            )}
          </button>

          {/* Task Type Badge */}
          <Badge
            variant="outline"
            className={`flex items-center gap-1 text-[11px] font-medium px-2 py-0.5 ${currentTypeMeta.badgeClass}`}
          >
            <TypeIcon className={`w-3 h-3 ${currentTypeMeta.iconClass}`} />
            <span>{currentTypeMeta.label}</span>
          </Badge>

          {/* Status Badge */}
          {task.status && (
            <Badge
              variant="outline"
              className="text-[10px] uppercase font-mono tracking-wider px-2 py-0.5 bg-zinc-950 border-zinc-800 text-zinc-400"
            >
              {task.status.name || task.status.category}
            </Badge>
          )}
        </div>

        {/* Right Header Actions */}
        <div className="flex items-center gap-2">
          {mode === 'drawer' && (
            <button
              type="button"
              onClick={handleExpandToFullScreen}
              className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors cursor-pointer"
              title="Expand to Full Page"
            >
              <Maximize2 className="h-4 w-4" />
            </button>
          )}

          <button
            type="button"
            onClick={() => setIsShareDialogOpen(true)}
            className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors cursor-pointer"
            title="Share task publicly"
          >
            <Share2 className="h-4 w-4" />
          </button>

          {canDelete && (
            <button
              type="button"
              onClick={() => setIsDeleteTaskModalOpen(true)}
              disabled={isDeleting}
              className="p-1.5 rounded-md text-zinc-500 hover:text-rose-400 hover:bg-zinc-800 transition-colors cursor-pointer disabled:opacity-40"
              title="Delete task"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          )}

          {mode === 'drawer' && onClose && (
            <button
              type="button"
              onClick={onClose}
              className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors cursor-pointer"
              title="Close drawer"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>

      {/* Main Workspace Layout (2-Column Split) */}
      <div
        className={`grid grid-cols-1 lg:grid-cols-3 gap-6 sm:gap-8 ${
          mode === 'drawer' ? 'p-6 overflow-y-auto flex-1' : ''
        }`}
      >
        {/* Left Column (2/3 Main Canvas) */}
        <div className="lg:col-span-2 space-y-6">
          {/* Parent Task Link (if present) */}
          {task.parentTask && (
            <div className="flex items-center gap-2 p-2.5 rounded-md bg-zinc-950/60 border border-zinc-850 text-xs text-zinc-400">
              <CornerDownRight className="w-3.5 h-3.5 text-zinc-500" />
              <span>Subtask of:</span>
              <button
                type="button"
                onClick={() => handleNavigateToTask(task.parentTask?.id || '')}
                className="font-semibold text-zinc-200 hover:text-violet-400 hover:underline transition-colors cursor-pointer"
              >
                {task.parentTask.slug || 'PARENT'}: {task.parentTask.title}
              </button>
            </div>
          )}

          {/* Task Title Section */}
          <div className="space-y-2">
            {isEditingTitle ? (
              <div className="space-y-2">
                <Input
                  value={titleDraft}
                  onChange={(e) => setTitleDraft(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') handleSaveTitle();
                    if (e.key === 'Escape') {
                      setTitleDraft(task.title);
                      setIsEditingTitle(false);
                    }
                  }}
                  autoFocus
                  placeholder="Task title..."
                  className="text-lg font-bold bg-zinc-950 border-zinc-700 text-zinc-100 px-3 py-2 rounded-md"
                />
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant="primary"
                    onClick={handleSaveTitle}
                    isLoading={isUpdating}
                    className="text-xs gap-1"
                  >
                    <Check className="w-3 h-3" /> Save
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => {
                      setTitleDraft(task.title);
                      setIsEditingTitle(false);
                    }}
                    className="text-xs"
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            ) : (
              <div className="group flex items-start justify-between gap-3">
                <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-zinc-100 leading-snug">
                  {task.title}
                </h1>
                {canEdit && (
                  <button
                    type="button"
                    onClick={() => setIsEditingTitle(true)}
                    className="opacity-0 group-hover:opacity-100 p-1.5 rounded-md text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 transition-all cursor-pointer shrink-0 mt-0.5"
                    title="Edit title"
                  >
                    <Pencil className="w-3.5 h-3.5" />
                  </button>
                )}
              </div>
            )}
          </div>

          {/* Task Description Section */}
          <div className="space-y-3">
            <div className="flex items-center justify-between pb-1 border-b border-zinc-800">
              <span className="text-xs font-semibold text-zinc-300">Description</span>
              <div className="flex items-center gap-2">
                {canEdit && (
                  <button
                    type="button"
                    onClick={handleGenerateAiDescription}
                    disabled={isGeneratingDesc}
                    className="flex items-center gap-1.5 text-[11px] font-medium text-violet-400 hover:text-violet-300 disabled:opacity-40 cursor-pointer transition-colors"
                    title="Draft or enhance technical description with AI"
                  >
                    {isGeneratingDesc ? (
                      <Loader2 className="w-3 h-3 animate-spin text-violet-400" />
                    ) : (
                      <Sparkles className="w-3 h-3 text-violet-400" />
                    )}
                    <span>{isGeneratingDesc ? 'Generating...' : 'AI Assist'}</span>
                  </button>
                )}
                {canEdit && !isEditingDescription && (
                  <button
                    type="button"
                    onClick={() => setIsEditingDescription(true)}
                    className="text-[11px] font-medium text-zinc-400 hover:text-zinc-200 cursor-pointer"
                  >
                    Edit
                  </button>
                )}
              </div>
            </div>

            {isEditingDescription ? (
              <div className="space-y-3">
                <DualModeEditor
                  value={descriptionDraft}
                  disabled={!canEdit}
                  onChange={setDescriptionDraft}
                  placeholder="Add technical requirements, user stories, acceptance criteria, or context... (Type @ to mention teammates)"
                  workspaceId={activeWorkspaceId || undefined}
                  minRows={6}
                />
                <div className="flex items-center justify-end gap-2">
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => {
                      setDescriptionDraft(task.description || '');
                      setIsEditingDescription(false);
                    }}
                    className="text-xs"
                  >
                    Cancel
                  </Button>
                  <Button
                    size="sm"
                    variant="primary"
                    onClick={handleSaveDescription}
                    isLoading={isUpdating}
                    className="text-xs gap-1"
                  >
                    <Check className="w-3 h-3" /> Save Changes
                  </Button>
                </div>
              </div>
            ) : task.description?.trim() ? (
              <div
                onClick={() => canEdit && setIsEditingDescription(true)}
                className="p-4 rounded-md bg-zinc-950/60 border border-zinc-850 prose prose-invert prose-xs sm:prose-sm max-w-none text-zinc-300 leading-relaxed cursor-pointer hover:border-zinc-800 transition-colors"
                title={canEdit ? 'Click to edit description' : undefined}
              >
                <ReactMarkdown>
                  {task.description.replace(/@\[([^\]]+)\]/g, '**`@$1`**')}
                </ReactMarkdown>
              </div>
            ) : (
              <div
                onClick={() => canEdit && setIsEditingDescription(true)}
                className="p-5 rounded-md border border-dashed border-zinc-800 text-center text-xs text-zinc-500 hover:border-zinc-700 hover:text-zinc-400 cursor-pointer transition-colors"
              >
                No description provided. Click here or use AI Assist to add specifications.
              </div>
            )}
          </div>

          {/* Task Hierarchy & Subtasks Section */}
          <div className="p-4 rounded-md bg-zinc-950/40 border border-zinc-800">
            <TaskHierarchy
              task={task}
              canEdit={canEdit}
              onNavigateToTask={handleNavigateToTask}
            />
          </div>

          {/* Task Dependencies Section */}
          <div className="p-4 rounded-md bg-zinc-950/40 border border-zinc-800">
            <TaskDependencies
              taskId={task.id}
              projectId={task.projectId}
              canEdit={canEdit}
            />
          </div>

          {/* Task Attachments Section */}
          <div className="p-4 rounded-md bg-zinc-950/40 border border-zinc-800">
            <TaskAttachments taskId={task.id} canEdit={canEdit} />
          </div>

          {/* Comments & Activity Feed Tabs */}
          <div className="space-y-4 pt-4 border-t border-zinc-800">
            <div className="flex items-center justify-between pb-1 border-b border-zinc-850">
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setBottomTab('comments')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-t-md text-xs font-semibold transition-colors border-b-2 -mb-px cursor-pointer ${
                    bottomTab === 'comments'
                      ? 'border-violet-500 text-zinc-100 bg-zinc-900/60'
                      : 'border-transparent text-zinc-400 hover:text-zinc-200'
                  }`}
                >
                  <MessageSquare className="h-3.5 w-3.5 text-zinc-400" />
                  <span>Comments</span>
                  <span className="text-[10px] font-mono px-1 rounded-full bg-zinc-800 text-zinc-400">
                    {comments.length}
                  </span>
                </button>

                <button
                  type="button"
                  onClick={() => setBottomTab('activity')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-t-md text-xs font-semibold transition-colors border-b-2 -mb-px cursor-pointer ${
                    bottomTab === 'activity'
                      ? 'border-violet-500 text-zinc-100 bg-zinc-900/60'
                      : 'border-transparent text-zinc-400 hover:text-zinc-200'
                  }`}
                >
                  <History className="h-3.5 w-3.5 text-zinc-400" />
                  <span>Activity History</span>
                  {activities.length > 0 && (
                    <span className="text-[10px] font-mono px-1 rounded-full bg-zinc-800 text-zinc-400">
                      {activities.length}
                    </span>
                  )}
                </button>
              </div>
            </div>

            {bottomTab === 'comments' ? (
              <div className="space-y-4">
                {/* Comments Feed */}
                <div className="space-y-3 max-h-96 overflow-y-auto pr-1">
                  {isCommentsLoading ? (
                    <div className="py-6 text-center text-xs text-zinc-500">
                      Loading conversation...
                    </div>
                  ) : comments.length === 0 ? (
                    <div className="py-6 text-center text-xs text-zinc-600">
                      No comments yet. Start the conversation below.
                    </div>
                  ) : (
                    comments.map((c) => {
                      const authorName =
                        c.author?.firstName && c.author?.lastName
                          ? `${c.author.firstName} ${c.author.lastName}`
                          : c.author?.username || 'Team Member';
                      const initials =
                        (
                          (c.author?.firstName?.[0] || '') +
                          (c.author?.lastName?.[0] || '')
                        ).toUpperCase() || 'U';
                      const isOwnComment = currentUser?.id === c.authorId;

                      return (
                        <div
                          key={c.id}
                          className="p-3.5 rounded-md bg-zinc-950/60 border border-zinc-850 space-y-1.5 text-xs"
                        >
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                              <Avatar className="h-5 w-5 rounded-sm border border-zinc-800">
                                <AvatarFallback className="text-[9px] bg-zinc-800 text-zinc-300">
                                  {initials}
                                </AvatarFallback>
                              </Avatar>
                              <span className="font-semibold text-zinc-200 text-xs">
                                {authorName}
                              </span>
                              <span className="text-[10px] text-zinc-500">
                                {c.createdAt
                                  ? new Date(c.createdAt).toLocaleDateString(undefined, {
                                      month: 'short',
                                      day: 'numeric',
                                      hour: '2-digit',
                                      minute: '2-digit',
                                    })
                                  : ''}
                              </span>
                            </div>

                            {(isOwnComment || canDelete) && (
                              <button
                                type="button"
                                onClick={() => setCommentToDelete(c.id)}
                                className="text-zinc-600 hover:text-rose-400 transition-colors p-1"
                                title="Delete comment"
                              >
                                <Trash2 className="h-3 w-3" />
                              </button>
                            )}
                          </div>

                          <div className="text-zinc-300 text-xs pl-7 leading-relaxed prose prose-invert prose-xs max-w-none [&_p]:leading-relaxed [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:bg-violet-950/40 [&_code]:border [&_code]:border-violet-800/40 [&_code]:text-violet-300 [&_code]:rounded-xs">
                            <ReactMarkdown>
                              {c.content.replace(/@\[([^\]]+)\]/g, '**`@$1`**')}
                            </ReactMarkdown>
                          </div>
                        </div>
                      );
                    })
                  )}
                </div>

                {/* Comment Input Box */}
                {canEdit && (
                  <form onSubmit={handleCommentSubmit} className="flex gap-2 pt-2">
                    <div className="relative flex-1">
                      <MentionAutocomplete
                        isOpen={isMentionOpen}
                        searchQuery={mentionQuery}
                        members={mentionMembers}
                        onSelect={handleSelectMention}
                        onClose={() => setIsMentionOpen(false)}
                      />
                      <Input
                        ref={commentInputRef}
                        value={commentText}
                        onChange={handleCommentChange}
                        placeholder="Write a comment or update... (Type @ to mention)"
                        className="bg-zinc-950/80 border-zinc-800 text-xs text-zinc-200"
                      />
                    </div>
                    <Button
                      type="submit"
                      size="sm"
                      variant="primary"
                      isLoading={isPostingComment}
                      disabled={!commentText.trim()}
                      className="gap-1.5 text-xs shrink-0"
                    >
                      <Send className="h-3.5 w-3.5" />
                      <span>Send</span>
                    </Button>
                  </form>
                )}
              </div>
            ) : (
              /* Activity Feed */
              <div className="space-y-3 max-h-96 overflow-y-auto pr-1">
                {isActivityLoading ? (
                  <div className="py-6 text-center text-xs text-zinc-500">
                    Loading activity timeline...
                  </div>
                ) : activities.length === 0 ? (
                  <div className="py-6 text-center text-xs text-zinc-600">
                    No activity recorded for this task yet.
                  </div>
                ) : (
                  activities.map((act) => {
                    const userName = act.user?.firstName
                      ? `${act.user.firstName} ${act.user.lastName || ''}`.trim()
                      : 'Team Member';
                    const initials = act.user?.firstName?.[0]?.toUpperCase() || 'U';

                    return (
                      <div
                        key={act.id}
                        className="p-3 rounded-md bg-zinc-950/60 border border-zinc-850 flex items-start gap-2.5 text-xs"
                      >
                        <Avatar className="h-5 w-5 rounded-sm border border-zinc-800 shrink-0 mt-0.5">
                          <AvatarFallback className="text-[9px] bg-zinc-800 text-zinc-300 font-bold">
                            {initials}
                          </AvatarFallback>
                        </Avatar>

                        <div className="flex-1 min-w-0 space-y-1">
                          <div className="flex items-center justify-between gap-2">
                            <span className="font-semibold text-zinc-200 text-xs truncate">
                              {userName}
                            </span>
                            <span className="text-[10px] text-zinc-500 flex items-center gap-1 shrink-0">
                              <Clock className="w-2.5 h-2.5" />
                              {new Date(act.createdAt).toLocaleDateString(undefined, {
                                month: 'short',
                                day: 'numeric',
                                hour: '2-digit',
                                minute: '2-digit',
                              })}
                            </span>
                          </div>
                          <p className="text-zinc-400 text-xs leading-relaxed">
                            {act.description}
                          </p>
                          <div className="pt-0.5">
                            <Badge
                              variant="outline"
                              className="text-[9px] uppercase font-mono px-1 py-0"
                            >
                              {act.type}
                            </Badge>
                          </div>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            )}
          </div>
        </div>

        {/* Right Column (1/3 Metadata Inspector Panel) */}
        <div className="lg:col-span-1 space-y-5">
          <div className="p-4 sm:p-5 rounded-md bg-zinc-950/60 border border-zinc-800 space-y-4">
            <h3 className="text-xs font-bold uppercase tracking-wider text-zinc-400 border-b border-zinc-850 pb-2">
              Task Details
            </h3>

            {/* Status Picker */}
            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400 flex items-center gap-1.5">
                <Layers className="h-3 w-3 text-zinc-500" />
                <span>Status</span>
              </label>
              <Select
                value={task.statusId || (projectStatuses[0]?.id ?? 'default')}
                disabled={!canEdit}
                onValueChange={(val) => handleStatusChange(val)}
              >
                <SelectTrigger className="w-full bg-zinc-900 border-zinc-800 text-zinc-200 h-8 text-xs">
                  <SelectValue placeholder="Select status" />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                  {projectStatuses.length > 0 ? (
                    projectStatuses.map((s) => (
                      <SelectItem key={s.id} value={s.id} className="text-xs">
                        {s.name} ({s.category})
                      </SelectItem>
                    ))
                  ) : (
                    <SelectItem value={task.statusId || 'default'} className="text-xs">
                      {task.status?.name || 'Default'}
                    </SelectItem>
                  )}
                </SelectContent>
              </Select>
            </div>

            {/* Priority Picker */}
            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400 flex items-center gap-1.5">
                <AlertCircle className="h-3 w-3 text-zinc-500" />
                <span>Priority</span>
              </label>
              <Select
                value={task.priority || 'MEDIUM'}
                disabled={!canEdit}
                onValueChange={(val) => handlePriorityChange(val as TaskPriority)}
              >
                <SelectTrigger className="w-full bg-zinc-900 border-zinc-800 text-zinc-200 h-8 text-xs">
                  <SelectValue placeholder="Select priority" />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                  {TASK_PRIORITIES.map((p) => (
                    <SelectItem key={p.value} value={p.value} className="text-xs">
                      {p.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Task Type Picker */}
            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400 flex items-center gap-1.5">
                <CheckSquare className="h-3 w-3 text-zinc-500" />
                <span>Task Type</span>
              </label>
              <Select
                value={currentType}
                disabled={!canEdit}
                onValueChange={(val) => handleTaskTypeChange(val as TaskType)}
              >
                <SelectTrigger className="w-full bg-zinc-900 border-zinc-800 text-zinc-200 h-8 text-xs">
                  <SelectValue placeholder="Select task type" />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                  {(Object.keys(TASK_TYPE_CONFIG) as TaskType[]).map((typeKey) => (
                    <SelectItem key={typeKey} value={typeKey} className="text-xs">
                      {TASK_TYPE_CONFIG[typeKey].label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Sprint (Agile) */}
            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400 flex items-center gap-1.5">
                <Timer className="h-3 w-3 text-zinc-500" />
                <span>Sprint</span>
              </label>
              <Select
                value={task.sprintId || 'none'}
                disabled={!canEdit}
                onValueChange={(val) => handleSprintChange(val)}
              >
                <SelectTrigger className="w-full bg-zinc-900 border-zinc-800 text-zinc-200 h-8 text-xs">
                  <SelectValue placeholder="Select sprint" />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                  <SelectItem value="none" className="text-xs">
                    Backlog (No Sprint)
                  </SelectItem>
                  {sprints.map((s) => (
                    <SelectItem key={s.id} value={s.id} className="text-xs">
                      {s.name} ({s.status})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Story Points (Fibonacci Estimation) */}
            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <label className="text-[11px] font-medium text-zinc-400 flex items-center gap-1.5">
                  <Hash className="h-3 w-3 text-zinc-500" />
                  <span>Story Points</span>
                </label>
                {task.storyPoints !== null && task.storyPoints !== undefined && (
                  <span className="text-[10px] font-mono text-zinc-400 font-bold">
                    {task.storyPoints} pts
                  </span>
                )}
              </div>
              <div className="flex flex-wrap gap-1">
                {STORY_POINTS_OPTIONS.map((pts) => {
                  const isSelected =
                    pts === null
                      ? task.storyPoints === null || task.storyPoints === undefined
                      : task.storyPoints === pts;

                  return (
                    <button
                      key={pts === null ? 'none' : pts}
                      type="button"
                      disabled={!canEdit}
                      onClick={() => handleStoryPointsChange(pts)}
                      className={`px-2 py-0.5 rounded text-[11px] font-mono font-medium transition-colors cursor-pointer disabled:cursor-not-allowed ${
                        isSelected
                          ? 'bg-violet-600 text-white font-bold'
                          : 'bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-zinc-200 hover:border-zinc-700'
                      }`}
                    >
                      {pts === null ? 'None' : pts}
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Dates (Start & Due Date) */}
            <div className="space-y-3 pt-2 border-t border-zinc-850">
              {/* Start Date */}
              <div className="space-y-1">
                <label className="text-[11px] font-medium text-zinc-400 flex items-center gap-1.5">
                  <Calendar className="h-3 w-3 text-zinc-500" />
                  <span>Start Date</span>
                </label>
                <input
                  type="date"
                  disabled={!canEdit}
                  value={task.startDate ? task.startDate.slice(0, 10) : ''}
                  onChange={(e) => handleStartDateChange(e.target.value)}
                  className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 rounded px-2.5 py-1 text-xs outline-none focus:border-zinc-700 cursor-pointer disabled:cursor-not-allowed"
                />
              </div>

              {/* Due Date */}
              <div className="space-y-1">
                <div className="flex items-center justify-between">
                  <label className="text-[11px] font-medium text-zinc-400 flex items-center gap-1.5">
                    <Calendar className="h-3 w-3 text-zinc-500" />
                    <span>Due Date</span>
                  </label>
                  {dueStatus && (
                    <span
                      className={`text-[9px] font-medium px-1.5 py-0.2 rounded border ${dueStatus.className}`}
                    >
                      {dueStatus.label}
                    </span>
                  )}
                </div>
                <input
                  type="date"
                  disabled={!canEdit}
                  value={task.dueDate ? task.dueDate.slice(0, 10) : ''}
                  onChange={(e) => handleDueDateChange(e.target.value)}
                  className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 rounded px-2.5 py-1 text-xs outline-none focus:border-zinc-700 cursor-pointer disabled:cursor-not-allowed"
                />
              </div>
            </div>

            {/* Assignee */}
            <div className="space-y-1.5 pt-2 border-t border-zinc-850">
              <label className="text-[11px] font-medium text-zinc-400 flex items-center gap-1.5">
                <User className="h-3 w-3 text-zinc-500" />
                <span>Assignee</span>
              </label>
              <Select
                value={currentAssignee?.id || 'unassigned'}
                disabled={!canEdit}
                onValueChange={(val) => handleAssigneeChange(val)}
              >
                <SelectTrigger className="w-full bg-zinc-900 border-zinc-800 text-zinc-200 h-8 text-xs">
                  <SelectValue placeholder="Select assignee" />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                  <SelectItem value="unassigned" className="text-xs">
                    Unassigned
                  </SelectItem>
                  {workspaceMembers.map((m) => {
                    const name =
                      m.user?.firstName && m.user?.lastName
                        ? `${m.user.firstName} ${m.user.lastName}`
                        : m.user?.username || m.user?.email || 'User';
                    return (
                      <SelectItem key={m.userId} value={m.userId} className="text-xs">
                        {name}
                      </SelectItem>
                    );
                  })}
                </SelectContent>
              </Select>
            </div>

            {/* Reporter */}
            <div className="space-y-1.5">
              <label className="text-[11px] font-medium text-zinc-400 flex items-center gap-1.5">
                <UserCheck className="h-3 w-3 text-zinc-500" />
                <span>Reporter</span>
              </label>
              {canDelete ? (
                <Select
                  value={currentReporter?.id || workspaceMembers[0]?.userId || 'none'}
                  onValueChange={(val) => handleReporterChange(val)}
                >
                  <SelectTrigger className="w-full bg-zinc-900 border-zinc-800 text-zinc-200 h-8 text-xs">
                    <SelectValue placeholder="Select reporter" />
                  </SelectTrigger>
                  <SelectContent className="bg-zinc-900 border-zinc-800 text-zinc-200">
                    {workspaceMembers.map((m) => {
                      const name =
                        m.user?.firstName && m.user?.lastName
                          ? `${m.user.firstName} ${m.user.lastName}`
                          : m.user?.username || m.user?.email || 'User';
                      return (
                        <SelectItem key={m.userId} value={m.userId} className="text-xs">
                          {name}
                        </SelectItem>
                      );
                    })}
                  </SelectContent>
                </Select>
              ) : (
                <div className="flex items-center gap-2 p-2 rounded bg-zinc-900/60 border border-zinc-800 text-xs text-zinc-300">
                  <Avatar className="h-4 w-4 rounded-sm border border-zinc-700">
                    <AvatarFallback className="text-[8px] bg-zinc-800 text-zinc-300">
                      {(currentReporter?.firstName?.[0] || 'U').toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <span className="truncate">
                    {currentReporter?.firstName
                      ? `${currentReporter.firstName} ${currentReporter.lastName || ''}`.trim()
                      : currentReporter?.email || 'System'}
                  </span>
                </div>
              )}
            </div>

            {/* Time Tracking */}
            <div className="pt-2 border-t border-zinc-850">
              <TaskTimeTracking
                taskId={task.id}
                estimatedHours={task.estimatedHours}
                canEdit={canEdit}
              />
            </div>

            {/* Task Recurrence */}
            <div className="pt-2 border-t border-zinc-850">
              <div className="space-y-1 mb-1.5">
                <label className="text-[11px] font-medium text-zinc-400">
                  Recurrence
                </label>
              </div>
              <TaskRecurrenceSelector
                taskId={task.id}
                isRecurring={task.isRecurring}
                initialConfig={task.recurringConfig}
                canEdit={canEdit}
              />
            </div>

            {/* Labels */}
            <div className="pt-2 border-t border-zinc-850">
              <TaskLabels taskId={task.id} canEdit={canEdit} />
            </div>
          </div>
        </div>
      </div>

      {/* Share Dialog */}
      {task && (
        <ShareTaskDialog
          isOpen={isShareDialogOpen}
          onClose={() => setIsShareDialogOpen(false)}
          taskId={task.id}
          taskTitle={task.title}
        />
      )}

      {/* Delete Task Confirmation Modal */}
      <ConfirmationModal
        open={isDeleteTaskModalOpen}
        onOpenChange={setIsDeleteTaskModalOpen}
        title="Delete Task"
        description={
          <span>
            Are you sure you want to permanently delete task{' '}
            <strong className="text-zinc-200">
              {task?.slug ? `[${task.slug}] ` : ''}&ldquo;{task?.title}&rdquo;
            </strong>
            ? This action cannot be undone and will delete all attachments, subtasks, and comments.
          </span>
        }
        confirmLabel="Delete Task"
        variant="destructive"
        isLoading={isDeleting}
        onConfirm={handleConfirmDeleteTask}
      />

      {/* Delete Comment Confirmation Modal */}
      <ConfirmationModal
        open={!!commentToDelete}
        onOpenChange={(open) => !open && setCommentToDelete(null)}
        title="Delete Comment"
        description="Are you sure you want to delete this comment? This action cannot be undone."
        confirmLabel="Delete Comment"
        variant="destructive"
        onConfirm={handleConfirmDeleteComment}
      />
    </div>
  );
}
