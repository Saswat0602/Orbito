# 04. Database Schema & Data Models Specification

## 1. Overview & Database Architecture

Orbito uses **PostgreSQL 16+** as its primary relational store, managed through **Prisma ORM (v6+)**. The schema is optimized for multi-tenant SaaS isolation, high-concurrency real-time updates, full audit logging, and flexible entity relationships.

### Schema Statistics
- **Total Relational Models**: 38 models
- **Total Enumerations**: 20 enums
- **Primary Keys**: UUID v4 across all entities for secure multi-tenant referencing
- **Soft Deletes**: Implemented on `users`, `workspaces`, `projects`, `tasks`, `task_statuses`, and `sprints` (`deletedAt`, `deletedBy`)
- **Ordering**: LexoRank strings (`task_ranks` table) enabling millisecond drag-and-drop reordering without row locking

---

## 2. Core Entity Relationship Diagram (Conceptual)

```
[User] <---- (N:M via OrganizationMember) ----> [Organization]
  |                                                    |
  |                                              (1:N Workspaces)
  |                                                    |
  +-------- (N:M via WorkspaceMember) -----------> [Workspace] (Self-referential Hierarchy)
  |                                                    |
  |                                              (1:N Projects)
  |                                                    |
  +-------- (N:M via ProjectMember) -------------> [Project] (Prefix: e.g. ORB)
                                                       |
                 +-------------------+-----------------+--------------------+
                 |                   |                 |                    |
             [Sprint]            [Workflow]        [Label]             [ProjectInbox]
                 |                   |                 |                    |
                 |             [TaskStatus]            |              [InboxMessage]
                 |                   |                 |
                 +---------> [Task] <------------------+
                               |
       +-----------+-----------+-----------+-----------+-----------+
       |           |           |           |           |           |
 [Subtasks]   [Assignees]  [Comments] [Attachments] [TimeEntries] [TaskRanks]
```

---

## 3. Enumeration Catalog (20 Enums)

| Enum Name | Values | Purpose |
| :--- | :--- | :--- |
| `Role` | `SUPER_ADMIN`, `OWNER`, `MANAGER`, `MEMBER`, `VIEWER` | Multi-tiered RBAC permissions across Org, Workspace, and Project |
| `UserStatus` | `ACTIVE`, `INACTIVE`, `SUSPENDED`, `PENDING` | Account status and access gating |
| `UserSource` | `MANUAL`, `EMAIL_INBOX`, `SSO`, `API`, `IMPORT`, `SIGNUP` | Attribution tracking for user onboarding |
| `ProjectStatus` | `PLANNING`, `ACTIVE`, `ON_HOLD`, `COMPLETED`, `CANCELLED` | High-level project lifecycle tracking |
| `ProjectPriority` | `LOW`, `MEDIUM`, `HIGH`, `URGENT` | Project urgency classification |
| `ProjectVisibility`| `PRIVATE`, `INTERNAL`, `PUBLIC` | Access scoping within workspaces |
| `TaskType` | `TASK`, `BUG`, `EPIC`, `STORY`, `SUBTASK` | Task taxonomy and icon representation |
| `TaskPriority` | `LOWEST`, `LOW`, `MEDIUM`, `HIGH`, `HIGHEST` | Granular priority matrix for sorting |
| `StatusCategory` | `TODO`, `IN_PROGRESS`, `DONE` | Core state classification for agile metrics |
| `SprintStatus` | `PLANNING`, `ACTIVE`, `COMPLETED`, `CANCELLED` | Sprint state machine |
| `DependencyType` | `BLOCKS`, `FINISH_START`, `START_START`, `FINISH_FINISH`, `START_FINISH` | Gantt chart and task dependency scheduling |
| `CustomFieldType`| `TEXT`, `NUMBER`, `DATE`, `BOOLEAN`, `SELECT`, `MULTI_SELECT`, `USER`, `URL` | Dynamic custom metadata types |
| `NotificationType`| `TASK_ASSIGNED`, `TASK_STATUS_CHANGED`, `TASK_COMMENTED`, `TASK_DUE_SOON`, `PROJECT_CREATED`, `PROJECT_UPDATED`, `WORKSPACE_INVITED`, `MENTION` | Real-time and push notification categorization |
| `NotificationPriority`| `LOW`, `MEDIUM`, `HIGH`, `URGENT` | Visual urgency indicator for notifications |
| `ActivityType` | 75+ audit values (e.g. `TASK_CREATED`, `STATUS_CHANGED`, etc.) | Fine-grained activity log audit trail |
| `TriggerType` | `TASK_CREATED`, `TASK_UPDATED`, `TASK_STATUS_CHANGED`, `TASK_ASSIGNED`, `TASK_DUE_DATE_APPROACHING`, `TASK_OVERDUE`, `SPRINT_STARTED`, `SPRINT_COMPLETED` | Automation engine event triggers |
| `ActionType` | `ASSIGN_TASK`, `CHANGE_STATUS`, `ADD_LABEL`, `REMOVE_LABEL`, `SET_DUE_DATE`, `SEND_NOTIFICATION`, `SEND_EMAIL`, `ADD_COMMENT` | Automation engine executable actions |
| `RuleStatus` | `ACTIVE`, `INACTIVE`, `DRAFT` | Automation rule lifecycle |
| `RecurrenceType` | `DAILY`, `WEEKLY`, `MONTHLY`, `QUARTERLY`, `YEARLY`, `CUSTOM` | Recurring task schedule generation |
| `RecurrenceEndType`| `NEVER`, `ON_DATE`, `AFTER_OCCURRENCES` | Recurring task termination rule |

---

## 4. Key Models Specification

### 4.1. Core Multi-Tenancy Models

```prisma
model User {
  id               String       @id @default(uuid()) @db.Uuid
  email            String       @unique
  username         String?      @unique
  firstName        String       @map("first_name")
  lastName         String       @map("last_name")
  mobileNumber     String?      @map("mobile_number")
  avatar           String?
  bio              String?
  timezone         String       @default("UTC")
  language         String       @default("en")
  role             Role         @default(MEMBER)
  status           UserStatus   @default(ACTIVE)
  lastLoginAt      DateTime?    @map("last_login_at")
  emailVerified    Boolean      @default(false) @map("email_verified")
  password         String?
  refreshToken     String?      @map("refresh_token")
  preferences      Json?
  onboardInfo      Json?
  source           UserSource   @default(MANUAL)
  externalId       String?      @unique @map("external_id")
  externalProvider String?      @map("external_provider")
  resetToken       String?      @unique @map("reset_token")
  resetTokenExpiry DateTime?    @map("reset_token_expiry")
  defaultOrganizationId String? @map("default_organization_id") @db.Uuid
  createdAt        DateTime     @default(now()) @map("created_at")
  updatedAt        DateTime     @updatedAt @map("updated_at")
  deletedAt        DateTime?    @map("deleted_at")
  deletedBy        String?      @map("deleted_by") @db.Uuid

  @@map("users")
}

model Organization {
  id          String   @id @default(uuid()) @db.Uuid
  name        String
  slug        String   @unique
  description String?
  avatar      String?
  website     String?
  settings    Json?
  ownerId     String   @map("owner_id") @db.Uuid
  archive     Boolean  @default(false)
  createdAt   DateTime @default(now()) @map("created_at")
  updatedAt   DateTime @updatedAt @map("updated_at")

  @@map("organizations")
}

model Workspace {
  id                String      @id @default(uuid()) @db.Uuid
  name              String
  slug              String
  description       String?
  avatar            String?
  color             String?
  settings          Json?
  organizationId    String      @map("organization_id") @db.Uuid
  parentWorkspaceId String?     @map("parent_workspace_id") @db.Uuid
  path              String?     @map("path")
  archive           Boolean     @default(false)
  createdAt         DateTime    @default(now()) @map("created_at")
  updatedAt         DateTime    @updatedAt @map("updated_at")

  @@unique([organizationId, slug])
  @@map("workspaces")
}

model Project {
  id          String          @id @default(uuid()) @db.Uuid
  name        String
  slug        String          @unique
  taskPrefix  String?         @map("task_prefix") @db.VarChar(8) // e.g. "FF"
  description String?
  avatar      String?
  color       String
  status      ProjectStatus   @default(PLANNING)
  priority    ProjectPriority @default(MEDIUM)
  visibility  ProjectVisibility @default(INTERNAL)
  archive     Boolean         @default(false)
  workspaceId String          @map("workspace_id") @db.Uuid
  createdAt   DateTime        @default(now()) @map("created_at")
  updatedAt   DateTime        @updatedAt @map("updated_at")

  @@map("projects")
}
```

### 4.2. Task, Sprints & Ranking Models

```prisma
model Task {
  id              String       @id @default(uuid()) @db.Uuid
  title           String
  description     String?      @db.Text
  type            TaskType     @default(TASK)
  priority        TaskPriority @default(MEDIUM)
  taskNumber      Int          @map("task_number")
  slug            String       @unique // e.g. "FF-104"
  startDate       DateTime?    @map("start_date")
  dueDate         DateTime?    @map("due_date")
  estimatedHours  Float?       @map("estimated_hours")
  actualHours     Float?       @map("actual_hours")
  storyPoints     Int?         @map("story_points")
  projectId       String       @map("project_id") @db.Uuid
  statusId        String       @map("status_id") @db.Uuid
  sprintId        String?      @map("sprint_id") @db.Uuid
  parentTaskId    String?      @map("parent_task_id") @db.Uuid
  isArchived      Boolean      @default(false) @map("is_archived")
  createdById     String?      @map("created_by_id") @db.Uuid
  createdAt       DateTime     @default(now()) @map("created_at")
  updatedAt       DateTime     @updatedAt @map("updated_at")

  @@unique([projectId, taskNumber])
  @@map("tasks")
}

model TaskRank {
  id        String    @id @default(uuid()) @db.Uuid
  taskId    String    @map("task_id") @db.Uuid
  scopeType ScopeType @map("scope_type") // ORGANIZATION, WORKSPACE, PROJECT
  scopeId   String    @map("scope_id") @db.Uuid
  viewType  ViewType  @map("view_type")  // LIST, BOARD, GANTT
  rank      String    @db.VarChar(255)   // LexoRank alphanumeric ordering string

  @@unique([taskId, scopeId, viewType])
  @@index([scopeId, viewType, rank])
  @@map("task_ranks")
}
```

---

## 5. Idempotent Database Seeding Strategy

The Orbito database seeder (`src/seeder/seeder.command.ts` / `seed.ts`) is designed to be **100% idempotent** and safe to execute multiple times in development and production environments:

1. **System Admin Provisioning**: Creates the initial superuser account (`admin@orbito.io` with secure default password).
2. **Default Organization & Workspaces**: Seeds a default organization (`Orbito Workspace`) with hierarchical sub-workspaces (`Engineering`, `Product`, `Design`).
3. **Standard Workflows & Statuses**:
   - `TODO` (Category: `TODO`, Position: 0, Color: `#71717a`)
   - `IN_PROGRESS` (Category: `IN_PROGRESS`, Position: 1, Color: `#3b82f6`)
   - `IN_REVIEW` (Category: `IN_PROGRESS`, Position: 2, Color: `#eab308`)
   - `DONE` (Category: `DONE`, Position: 3, Color: `#22c55e`)
4. **Transition Graph**: Defines valid state transitions and allowed flows.
5. **Sample Demo Project**: Populates a realistic sample project with backlog tasks, an active sprint, milestones, labels, and task dependencies for immediate evaluation.
