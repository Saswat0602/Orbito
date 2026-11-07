import type {
  TaskPriority,
  TaskType,
  DependencyType,
  RecurrenceType,
  RecurrenceEndType,
} from '../enums';
import type { UserSummary } from '../users/user.types';
import type { ProjectSummary } from '../projects/project.types';
import type { SprintSummary } from '../sprints/sprint.types';
import type { TaskStatus } from '../workflows/workflow.types';
import type { Attachment } from '../attachments/attachment.types';

export interface TaskSummary {
  id: string;
  taskNumber: number;
  slug: string;
  title: string;
  priority: TaskPriority;
  taskType?: TaskType;
  statusId: string;
  projectId: string;
  rank: string;
  dueDate?: string | null;
  startDate?: string | null;
  estimatedHours?: number | null;
  loggedHours?: number;
  parentTaskId?: string | null;
  sprintId?: string | null;
  isArchived?: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface TaskAssignee {
  id: string;
  taskId: string;
  userId: string;
  user: UserSummary;
}

export interface TaskReporter {
  id: string;
  taskId: string;
  userId: string;
  user: UserSummary;
}

export interface TaskWatcher {
  id: string;
  taskId: string;
  userId: string;
  user: UserSummary;
}

export interface Label {
  id: string;
  name: string;
  color: string;
  description?: string | null;
  organizationId?: string | null;
}

export interface TaskLabel {
  id: string;
  taskId: string;
  labelId: string;
  label: Label;
}

export interface TaskAttachment extends Attachment {
  taskId?: string;
}

export interface TaskComment {
  id: string;
  taskId: string;
  authorId: string;
  content: string;
  parentCommentId?: string | null;
  author: UserSummary;
  createdAt: string;
  updatedAt: string;
  replies?: TaskComment[];
  // Email integration fields
  emailMessageId?: string;
  sentAsEmail?: boolean;
  emailRecipients?: string[];
  emailSentAt?: string;
}

export interface TaskDependency {
  id: string;
  dependentTaskId: string;
  blockingTaskId: string;
  type: DependencyType;
  dependentTask?: TaskSummary;
  blockingTask?: TaskSummary;
}

export interface TimeEntry {
  id: string;
  taskId: string;
  userId: string;
  duration: number; // in seconds
  description?: string | null;
  date: string;
  user?: UserSummary;
  createdAt: string;
  updatedAt: string;
}

export interface TaskRecurrence {
  id: string;
  taskId: string;
  type: RecurrenceType;
  interval: number;
  daysOfWeek?: number[];
  dayOfMonth?: number | null;
  monthOfYear?: number | null;
  endType: RecurrenceEndType;
  endDate?: string | null;
  occurrencesCount?: number | null;
  currentOccurrence: number;
  nextOccurrenceDate?: string | null;
}

export interface Task extends TaskSummary {
  description?: string | null;
  status: TaskStatus;
  project: ProjectSummary;
  sprint?: SprintSummary | null;
  parentTask?: TaskSummary | null;
  subtasks?: TaskSummary[];
  assignees: TaskAssignee[];
  reporters: TaskReporter[];
  watchers?: TaskWatcher[];
  labels: TaskLabel[];
  attachments?: TaskAttachment[];
  comments?: TaskComment[];
  dependencies?: TaskDependency[];
  timeEntries?: TimeEntry[];
  recurrence?: TaskRecurrence | null;
}
