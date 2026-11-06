# 02. Taskosaur Deep Functional Audit & Inventory

This document details the exhaustive feature analysis of the reference platform (`/home/saswatranjanmohanty/Desktop/HysProject/taskosaur`). Every minute detail, module, data model, endpoint, background queue, real-time event, and workflow identified in Taskosaur is documented here to ensure 100% functional parity in Orbito.

---

## 1. Domain Modules Breakdown (41 Backend Feature Modules)

Taskosaur backend consists of 41 feature modules registered under NestJS:

| # | Module | Scope & Responsibility |
|---|---|---|
| 1 | `auth` | Local authentication (email/password), JWT token generation, refresh token rotation, OIDC / SSO single sign-on (external ID, external provider), password reset tokens, account activation checks. |
| 2 | `users` | User profile management, avatar uploads, preferences (JSON), onboarding info (JSON), soft delete (`deletedAt`, `deletedBy`), user status (ACTIVE, INACTIVE, SUSPENDED, PENDING), timezone, language. |
| 3 | `organizations` | Tenant isolation, slug generation, owner assignments, settings JSON, organization archiving, default organization fallback. |
| 4 | `organization-members`| Multi-role membership within organizations (SUPER_ADMIN, OWNER, MANAGER, MEMBER, VIEWER), default org assignments, member invitations. |
| 5 | `workspaces` | Nested workspace hierarchy (parent/child workspaces, path-based hierarchical queries), slugs, color tags, avatar, workspace settings, archiving. |
| 6 | `workspace-members` | Workspace-level RBAC role assignment, member joining/leaving, permission checks. |
| 7 | `projects` | Project lifecycle, unique slug generation, custom task prefixes (`taskPrefix` up to 8 characters, e.g. `ORB`), color, avatar, priority, status (PLANNING, ACTIVE, ON_HOLD, COMPLETED, CANCELLED), visibility (PRIVATE, INTERNAL, PUBLIC), archiving. |
| 8 | `project-members` | Project-level access controls, role inheritance from workspace/org. |
| 9 | `tasks` | Core task management: CRUD, task numbers, slugs, rich markdown descriptions, start/due dates, estimated/actual hours, priority, task types (TASK, BUG, EPIC, STORY, SUBTASK), parent-child subtasks, display order. |
| 10 | `recurring-tasks` | Automated recurrence rules (DAILY, WEEKLY, MONTHLY, QUARTERLY, YEARLY, CUSTOM), intervals, days-of-week, day-of-month, recurrence end conditions (NEVER, ON_DATE, AFTER_OCCURRENCES). |
| 11 | `task-ranks` | Alphanumeric LexoRank calculation for millisecond order updates across boards, sprints, and lists without re-indexing all items. |
| 12 | `task-dependencies`| Graph relations between tasks: `BLOCKS`, `FINISH_START`, `START_START`, `FINISH_FINISH`, `START_FINISH` with cycle detection. |
| 13 | `workflows` | Custom workflow definitions per project/organization, setting default workflow schemas. |
| 14 | `task-statuses` | Custom status creation per workflow with categories (`TODO`, `IN_PROGRESS`, `DONE`), custom colors, position ordering, and soft-delete capabilities. |
| 15 | `status-transitions`| State transition validation rules defining allowed transitions between statuses (e.g. from `IN_PROGRESS` to `IN_REVIEW`). |
| 16 | `sprints` | Agile sprints: planning, active, completed states, sprint goals, start and end timestamps, sprint backlog assignment, burndown metrics. |
| 17 | `labels` | Project and workspace-level tags/labels with hex colors and descriptions. |
| 18 | `task-label` | Join model connecting tasks and labels with creator audit tracking. |
| 19 | `task-assignees` | Explicit multi-assignee support per task with join table. |
| 20 | `task-reporters` | Explicit multi-reporter tracking per task with join table. |
| 21 | `task-watchers` | Subscription mechanism for users following task updates and comment notifications. |
| 22 | `task-comments` | Threaded discussions on tasks, parent-child comment nesting, user mentions, recipient email notifications. |
| 23 | `task-attachments` | File metadata, S3/local storage paths, MIME types, file sizes, presigned URLs. |
| 24 | `time-entries` | Work logging: duration in seconds, start time, end time, date, description, user attribution, billable tracking. |
| 25 | `custom-fields` | Dynamic user-defined fields (TEXT, NUMBER, DATE, BOOLEAN, SELECT, MULTI_SELECT, USER, URL) per organization/project. |
| 26 | `activity-log` | Comprehensive audit trail logging 75+ activity types across all entity operations (oldValue, newValue, user, timestamp). |
| 27 | `notifications` | In-app notification delivery, unread counts, priority levels (LOW, MEDIUM, HIGH, URGENT), real-time delivery via WebSockets. |
| 28 | `invitations` | Organization, workspace, and project invitations sent via email with token verification, expiration, and status tracking (PENDING, ACCEPTED, DECLINED, EXPIRED). |
| 29 | `public` | Publicly shareable task links with token-based access, expiration, and revocation; public sprint and calendar read views. |
| 30 | `search` | Global cross-entity search indexing projects, tasks, comments, and members. |
| 31 | `gantt` | Gantt chart view data aggregation, calculating critical paths, dependency timelines, and milestone schedules. |
| 32 | `ai-chat` | Multi-provider LLM conversational interface (OpenAI, Anthropic, OpenRouter, Ollama) with context injection, conversation sessions, and message persistence. |
| 33 | `automation` | Rule-based automation engine: Triggers (TASK_CREATED, STATUS_CHANGED, ASSIGNEE_ADDED, DUE_DATE_APPROACHING) -> Actions (CHANGE_STATUS, ASSIGN_USER, SEND_NOTIFICATION). |
| 34 | `inbox` | Project email inbox integration: IMAP polling, email account credentials, message thread tracking, attachment extraction, auto-conversion to tasks via regex rules. |
| 35 | `email` | SMTP integration, email templates for invitations, notifications, password reset, and task comments. |
| 36 | `queue` | BullMQ Redis-backed background worker with seamless `better-queue` in-memory fallback for environments without Redis. |
| 37 | `scheduler` | Cron job scheduler handling recurring task generation, email inbox syncing, and overdue task alerts. |
| 38 | `storage` | Abstracted storage provider supporting local file storage and AWS S3 presigned upload/download URLs. |
| 39 | `editor-images` | TipTap rich-text inline image uploads with automatic asset tracking and URL resolution. |
| 40 | `jira-sync` | Workspace and project-level bidirectional synchronization with Jira Cloud / Data Center, status mapping, SSRF protection allowlists, and DNS pinning. |
| 41 | `trello-sync` | Workspace and project-level bidirectional sync with Trello boards, list-to-status mapping, and webhook triggers. |

---

## 2. Real-Time WebSocket Infrastructure (`/events` Namespace)

Taskosaur leverages Socket.IO with JWT authentication at the handshake level and dynamic room multiplexing:

### Connection Lifecycle
- **Auth**: Token validated on handshake (`auth.token` or `query.token`), resolves `userId`, validates user account is ACTIVE.
- **Presence**: Tracks user connection count across multiple tabs/devices, broadcasts `user:online` on first connect and `user:offline` (with `lastSeen`) on final disconnect.
- **Rooms**:
  - `user:{userId}` — Private alerts, mentions, direct notifications.
  - `org:{organizationId}` — Organization-wide membership updates.
  - `workspace:{workspaceId}` — Workspace settings and member activities.
  - `project:{projectId}` — Real-time board movements, task creations, status updates.
  - `task:{taskId}` — Live comments, typing indicators, concurrent edit warnings.

### Broadcast Events Matrix
- `task:created`, `task:updated`, `task:deleted`, `task:status_changed`, `task:assigned`
- `task:comment_added`, `task:comment_updated`, `task:comment_deleted`
- `task:attachment_added`, `task:attachment_deleted`
- `project:updated`
- `sprint:created`, `sprint:updated`, `sprint:deleted`, `sprint:status_changed`
- `member:added`, `member:removed`, `member:role_updated`
- `notification:created`, `notification:read`, `notification:read_all`

---

## 3. Conversational AI Task Execution & In-App Browser Automation

Taskosaur features a unique dual-layer AI architecture:

### Layer A: Backend LLM Orchestration (`ai-chat.service.ts`)
- **Multi-Provider Support**: Supports OpenAI, Anthropic, OpenRouter, and local OpenAI-compatible endpoints (Ollama).
- **Prompt Engineering & Context Augmentation**:
  - `app-guide.ts`: System instructions detailing the entire application structure, navigation routes, DOM element conventions, and entity relations.
  - `automation-prompts.ts`: Specialized instructions for parsing natural language into deterministic task creation, sprint planning, and workflow modifications.
- **Message Normalization**: Strict role-alternation normalization (`user -> assistant -> user`) preventing API rejections from Anthropic or Cohere when multiple tool results are appended.
- **Security & SSRF Guard**: AI endpoint URLs undergo SSRF validation (`resolveSafeDestination`, blocking private IPs, RFC 1918, link-local, loopback). API keys stored encrypted using AES-256-GCM.

### Layer B: Frontend Browser Automation Agent (`browser-agent.ts`)
- **DOM Detection (`dom-detector.ts`)**: Scans the active page for clickable elements, input fields, modals, and navigation links.
- **Action Execution (`action-executor.ts`)**: The AI generates action sequences (click, type, navigate, select, submit) that execute in real-time in the user's browser, giving the appearance of an autonomous assistant operating the software.
- **Voice Control (`voice.ts`)**: Speech-to-Text integration allowing hands-free voice commands that stream directly into the AI chat input.
- **Local Fallback Command Parser (`local-command-parser.ts`)**: Deterministic regex-based fallback executing common actions (e.g., "create task [title]") when offline or without LLM keys.

---

## 4. Background Queue Architecture: BullMQ with Better-Queue Fallback

Taskosaur implements a resilient dual-mode queue architecture:
1. **BullMQ Mode (Production)**: Uses Redis (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`) for distributed job execution, handling email delivery, automation rule execution, Jira/Trello sync jobs, and IMAP email polling.
2. **Better-Queue Mode (Development / Offline Fallback)**: If Redis is unavailable or crashes, the system gracefully degrades to an in-memory/SQLite queue (`better-queue`) without crashing the application server.

---

## 5. Security & SSRF Defense Architecture

1. **JIRA & AI SSRF Protection (`safe-outbound-http.ts`)**:
   - Host allowlists (`JIRA_ALLOWED_HOSTS`, `AI_ALLOWED_HOSTS`).
   - Pre-flight DNS resolution checking all resolved IPs against private/loopback/link-local ranges (RFC 1918, `169.254.x.x`, `::1`, carrier-grade NAT).
   - **DNS Pinning**: The outbound HTTP request connects directly to the validated IP address rather than the hostname, eliminating DNS rebinding vulnerabilities (TOCTOU).
   - Strict HTTP redirect rejection.
2. **Credential Encryption at Rest**:
   - Sensitive integration tokens (Jira API tokens, Trello tokens, IMAP passwords, LLM API keys) are encrypted using **AES-256-GCM** with initialization vectors (`IV`) and authentication tags.
3. **Role-Based Access Control (RBAC)**:
   - Granular authorization checks across Organization, Workspace, and Project scopes with roles: `SUPER_ADMIN`, `OWNER`, `MANAGER`, `MEMBER`, `VIEWER`.
