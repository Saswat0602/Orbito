/**
 * @orbito/types
 * Core domain contracts shared between NestJS backend and Next.js frontend
 */

// ==========================================
// 1. CORE DOMAIN ENUMS
// ==========================================

export enum Role {
  SUPER_ADMIN = 'SUPER_ADMIN',
  OWNER = 'OWNER',
  MANAGER = 'MANAGER',
  MEMBER = 'MEMBER',
  VIEWER = 'VIEWER',
}

export enum UserStatus {
  ACTIVE = 'ACTIVE',
  INACTIVE = 'INACTIVE',
  SUSPENDED = 'SUSPENDED',
}

export enum UserSource {
  LOCAL = 'LOCAL',
  GOOGLE = 'GOOGLE',
  GITHUB = 'GITHUB',
  SAML = 'SAML',
}

export enum TaskPriority {
  LOW = 'LOW',
  MEDIUM = 'MEDIUM',
  HIGH = 'HIGH',
  URGENT = 'URGENT',
}

export enum TaskType {
  TASK = 'TASK',
  BUG = 'BUG',
  STORY = 'STORY',
  EPIC = 'EPIC',
  SUBTASK = 'SUBTASK',
}

export enum StatusCategory {
  TODO = 'TODO',
  IN_PROGRESS = 'IN_PROGRESS',
  DONE = 'DONE',
  CANCELLED = 'CANCELLED',
}

export enum ProjectStatus {
  PLANNING = 'PLANNING',
  ACTIVE = 'ACTIVE',
  ON_HOLD = 'ON_HOLD',
  COMPLETED = 'COMPLETED',
  CANCELLED = 'CANCELLED',
}

export enum ProjectPriority {
  LOW = 'LOW',
  MEDIUM = 'MEDIUM',
  HIGH = 'HIGH',
  URGENT = 'URGENT',
}

export enum ProjectVisibility {
  PRIVATE = 'PRIVATE',
  INTERNAL = 'INTERNAL',
  PUBLIC = 'PUBLIC',
}

export enum SprintStatus {
  PLANNING = 'PLANNING',
  ACTIVE = 'ACTIVE',
  COMPLETED = 'COMPLETED',
  CANCELLED = 'CANCELLED',
}

export enum DependencyType {
  BLOCKS = 'BLOCKS',
  BLOCKED_BY = 'BLOCKED_BY',
  RELATES_TO = 'RELATES_TO',
  DUPLICATES = 'DUPLICATES',
}

export enum ScopeType {
  ORGANIZATION = 'ORGANIZATION',
  WORKSPACE = 'WORKSPACE',
  PROJECT = 'PROJECT',
  TASK = 'TASK',
}

export enum ViewType {
  LIST = 'LIST',
  BOARD = 'BOARD',
  CALENDAR = 'CALENDAR',
  GANTT = 'GANTT',
  TIMELINE = 'TIMELINE',
}

export enum RecurrenceType {
  DAILY = 'DAILY',
  WEEKLY = 'WEEKLY',
  MONTHLY = 'MONTHLY',
  YEARLY = 'YEARLY',
  CUSTOM = 'CUSTOM',
}

export enum RecurrenceEndType {
  NEVER = 'NEVER',
  AFTER_OCCURRENCES = 'AFTER_OCCURRENCES',
  ON_DATE = 'ON_DATE',
}

export enum NotificationType {
  TASK_ASSIGNED = 'TASK_ASSIGNED',
  TASK_COMMENT = 'TASK_COMMENT',
  TASK_STATUS_CHANGED = 'TASK_STATUS_CHANGED',
  MENTION = 'MENTION',
  SPRINT_STARTED = 'SPRINT_STARTED',
  SPRINT_ENDED = 'SPRINT_ENDED',
  INVITATION = 'INVITATION',
}

export enum NotificationPriority {
  LOW = 'LOW',
  NORMAL = 'NORMAL',
  HIGH = 'HIGH',
  URGENT = 'URGENT',
}

export enum ActivityType {
  CREATED = 'CREATED',
  UPDATED = 'UPDATED',
  DELETED = 'DELETED',
  STATUS_CHANGED = 'STATUS_CHANGED',
  ASSIGNED = 'ASSIGNED',
  COMMENTED = 'COMMENTED',
  ATTACHED = 'ATTACHED',
  ARCHIVED = 'ARCHIVED',
  RESTORED = 'RESTORED',
}

export enum TriggerType {
  TASK_CREATED = 'TASK_CREATED',
  TASK_UPDATED = 'TASK_UPDATED',
  TASK_MOVED = 'TASK_MOVED',
  STATUS_CHANGED = 'STATUS_CHANGED',
  DUE_DATE_APPROACHING = 'DUE_DATE_APPROACHING',
  SCHEDULE = 'SCHEDULE',
}

export enum ActionType {
  ASSIGN_USER = 'ASSIGN_USER',
  CHANGE_STATUS = 'CHANGE_STATUS',
  SEND_NOTIFICATION = 'SEND_NOTIFICATION',
  SEND_EMAIL = 'SEND_EMAIL',
  ADD_TAG = 'ADD_TAG',
  SET_PRIORITY = 'SET_PRIORITY',
}

export enum RuleStatus {
  ACTIVE = 'ACTIVE',
  INACTIVE = 'INACTIVE',
  DRAFT = 'DRAFT',
}

export enum InvitationStatus {
  PENDING = 'PENDING',
  ACCEPTED = 'ACCEPTED',
  DECLINED = 'DECLINED',
  EXPIRED = 'EXPIRED',
}

// ==========================================
// 2. USER & IDENTITY INTERFACES
// ==========================================

export interface UserSummary {
  id: string;
  email: string;
  firstName?: string | null;
  lastName?: string | null;
  username?: string | null;
  avatar?: string | null;
  role?: Role;
  status?: UserStatus;
}

export interface User extends UserSummary {
  bio?: string | null;
  mobileNumber?: string | null;
  language?: string | null;
  timezone?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token?: string;
}

export interface AuthResponse extends AuthTokens {
  user: User;
}

export interface AccessControlResult {
  isElevated: boolean;
  role: Role | string;
  canChange: boolean;
  userId: string;
  scopeId: string;
  scopeType: ScopeType;
  isSuperAdmin?: boolean;
}

// ==========================================
// 3. MULTI-TENANCY INTERFACES
// ==========================================

export interface OrganizationSummary {
  id: string;
  name: string;
  slug: string;
  avatar?: string | null;
  archive?: boolean;
}

export interface Organization extends OrganizationSummary {
  description?: string | null;
  website?: string | null;
  settings?: Record<string, unknown> | null;
  ownerId: string;
  owner?: UserSummary;
  createdAt: string;
  updatedAt: string;
  _count?: {
    members?: number;
    workspaces?: number;
  };
}

export interface OrganizationMember {
  id: string;
  organizationId: string;
  userId: string;
  role: Role;
  joinedAt: string;
  user: UserSummary;
}

export interface WorkspaceSummary {
  id: string;
  name: string;
  slug: string;
  color?: string | null;
  icon?: string | null;
  organizationId: string;
  parentWorkspaceId?: string | null;
  path?: string | null;
  archive?: boolean;
}

export interface Workspace extends WorkspaceSummary {
  description?: string | null;
  createdAt: string;
  updatedAt: string;
  organization?: OrganizationSummary;
  _count?: {
    members?: number;
    projects?: number;
    childWorkspaces?: number;
  };
}

export interface WorkspaceMember {
  id: string;
  workspaceId: string;
  userId: string;
  role: Role;
  joinedAt: string;
  user: UserSummary;
}

export interface WorkspaceTreeNode extends WorkspaceSummary {
  children?: WorkspaceTreeNode[];
  _count?: {
    members?: number;
    projects?: number;
    childWorkspaces?: number;
  };
}

export interface ProjectSummary {
  id: string;
  name: string;
  slug: string;
  identifier?: string | null;
  taskPrefix?: string | null;
  color?: string | null;
  avatar?: string | null;
  status: ProjectStatus;
  priority: ProjectPriority;
  visibility: ProjectVisibility;
  workspaceId: string;
  workflowId?: string | null;
  archive?: boolean;
}

export interface Project extends ProjectSummary {
  description?: string | null;
  startDate?: string | null;
  endDate?: string | null;
  settings?: Record<string, unknown> | null;
  createdAt: string;
  updatedAt: string;
  workspace?: WorkspaceSummary;
  workflow?: Workflow;
  members?: ProjectMember[];
  _count?: {
    members?: number;
    tasks?: number;
    sprints?: number;
  };
}

export interface ProjectMember {
  id: string;
  projectId: string;
  userId: string;
  role: Role;
  joinedAt: string;
  user: UserSummary;
}

// ==========================================
// 4. WORKFLOW & STATUS INTERFACES
// ==========================================

export interface StatusTransition {
  id: string;
  workflowId: string;
  fromStatusId: string;
  toStatusId: string;
}

export interface TaskStatus {
  id: string;
  name: string;
  color: string;
  category: StatusCategory;
  position: number;
  workflowId: string;
  isDeleted?: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Workflow {
  id: string;
  name: string;
  description?: string | null;
  isDefault: boolean;
  organizationId?: string | null;
  statuses: TaskStatus[];
  transitions?: StatusTransition[];
  createdAt: string;
  updatedAt: string;
}

// ==========================================
// 5. SPRINT INTERFACES
// ==========================================

export interface SprintSummary {
  id: string;
  name: string;
  slug: string;
  status: SprintStatus;
  projectId: string;
  startDate?: string | null;
  endDate?: string | null;
}

export interface Sprint extends SprintSummary {
  goal?: string | null;
  createdAt: string;
  updatedAt: string;
  _count?: {
    tasks?: number;
  };
}

export interface SprintBurndownPoint {
  date: string;
  totalPoints: number;
  remainingPoints: number;
  idealPoints: number;
}

// ==========================================
// 6. TASK & AGGREGATE INTERFACES
// ==========================================

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

export interface TaskAttachment {
  id: string;
  taskId: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  fileUrl: string;
  uploadedById: string;
  uploadedBy?: UserSummary;
  createdAt: string;
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
  dayOfMonth?: number;
  monthOfYear?: number;
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
  _count?: {
    subtasks?: number;
    comments?: number;
    attachments?: number;
    timeEntries?: number;
  };
}

// ==========================================
// 7. LEXORANK & REORDER INTERFACES
// ==========================================

export interface TaskReorderPayload {
  taskId: string;
  statusId: string;
  prevRank?: string | null;
  nextRank?: string | null;
  sprintId?: string | null;
}

export interface TaskReorderResult {
  taskId: string;
  statusId: string;
  rank: string;
  sprintId?: string | null;
  updatedAt: string;
}

export interface TaskStatusPositionUpdate {
  statusId: string;
  position: number;
}

// ==========================================
// 8. API REQUEST / RESPONSE WRAPPERS
// ==========================================

export interface ApiResponse<T = unknown> {
  success: boolean;
  data: T;
  message?: string;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
}

export interface PaginatedResponse<T = unknown> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface TaskQueryParams extends PaginationParams {
  organizationId?: string;
  workspaceId?: string;
  projectId?: string;
  sprintId?: string;
  statusIds?: string[];
  priorities?: TaskPriority[];
  assigneeIds?: string[];
  reporterIds?: string[];
  labelIds?: string[];
  search?: string;
  parentTaskId?: string | null;
  isArchived?: boolean;
  groupBy?: 'status' | 'priority' | 'assignee' | 'sprint';
}

export interface TaskGroupedResponse {
  groups: {
    key: string;
    title: string;
    meta?: Record<string, unknown>;
    tasks: Task[];
    total: number;
    hasMore: boolean;
  }[];
}

// ==========================================
// 9. WEBSOCKET REAL-TIME EVENT CONTRACTS
// ==========================================

export interface WebSocketEventPayload<T = unknown> {
  event: string;
  timestamp: string;
  organizationId?: string;
  workspaceId?: string;
  projectId?: string;
  actorId: string;
  data: T;
}

export type TaskCreatedEvent = WebSocketEventPayload<Task>;
export type TaskUpdatedEvent = WebSocketEventPayload<Partial<Task> & { id: string }>;
export type TaskMovedEvent = WebSocketEventPayload<TaskReorderResult>;
export type TaskDeletedEvent = WebSocketEventPayload<{ id: string; projectId: string }>;
export type CommentAddedEvent = WebSocketEventPayload<TaskComment>;
export type UserPresenceEvent = WebSocketEventPayload<{ userId: string; status: 'ONLINE' | 'OFFLINE' | 'AWAY' }>;
export type UserTypingEvent = WebSocketEventPayload<{ userId: string; taskId: string; isTyping: boolean }>;
