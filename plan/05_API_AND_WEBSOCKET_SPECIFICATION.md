# 05. API & WebSocket Specifications

## 1. REST API Architecture

Orbito provides a fully documented RESTful API conforming to the OpenAPI (Swagger 3.0) specification accessible at `/api/docs`.

### Common API Conventions
- **Base URL Prefix**: `/api`
- **Authentication**: Bearer JWT header: `Authorization: Bearer <accessToken>`
- **Content-Type**: `application/json` (except file uploads which use `multipart/form-data`)
- **Pagination**: Uniform cursor or offset-based: `?page=1&limit=20&sortBy=createdAt&sortOrder=desc`
- **Error Format**:
  ```json
  {
    "statusCode": 400,
    "message": ["title should not be empty", "priority must be a valid enum value"],
    "error": "Bad Request",
    "timestamp": "2026-11-06T10:00:00.000Z",
    "path": "/api/tasks"
  }
  ```

---

## 2. API Endpoint Catalog (Categorized by Domain)

### 2.1. Authentication & Users (`/api/auth`, `/api/users`)
- `POST /api/auth/register` — Create a new account with organization onboarding.
- `POST /api/auth/login` — Authenticate with email/password; returns access token + refresh token cookie.
- `POST /api/auth/refresh` — Rotate access token using valid refresh token.
- `POST /api/auth/logout` — Invalidate refresh token and clear session cookies.
- `POST /api/auth/forgot-password` — Request password reset email with secure token.
- `POST /api/auth/reset-password` — Verify token and update password.
- `GET  /api/auth/oidc/login` — Initiate OIDC / Single Sign-On flow.
- `GET  /api/auth/oidc/callback` — Exchange OIDC code for JWT session.
- `GET  /api/users/me` — Retrieve current authenticated user profile.
- `PATCH /api/users/me` — Update user profile (name, avatar, timezone, preferences).
- `DELETE /api/users/:id` — Soft-delete user account (Super Admin only).

### 2.2. Multi-Tenancy (`/api/organizations`, `/api/workspaces`)
- `GET    /api/organizations` — List user's active organizations.
- `POST   /api/organizations` — Create a new organization.
- `GET    /api/organizations/:id` — Get organization details.
- `PATCH  /api/organizations/:id` — Update organization metadata / settings.
- `GET    /api/organizations/:id/members` — List organization members.
- `POST   /api/organizations/:id/members` — Add member to organization.
- `GET    /api/workspaces` — List workspaces within active organization (supports tree view).
- `POST   /api/workspaces` — Create workspace (supports parentWorkspaceId for nesting).
- `GET    /api/workspaces/:id` — Retrieve workspace details and child workspaces.
- `PATCH  /api/workspaces/:id` — Update workspace properties.
- `DELETE /api/workspaces/:id` — Archive/delete workspace.

### 2.3. Projects & Workflows (`/api/projects`, `/api/workflows`, `/api/task-statuses`)
- `GET    /api/projects` — Filter projects by workspace slug or ID.
- `POST   /api/projects` — Create project with name, slug, custom task prefix (e.g. `FF`), and workflow.
- `GET    /api/projects/:id` — Get project details, members, and workflow statuses.
- `PATCH  /api/projects/:id` — Update project status, priority, or visibility.
- `GET    /api/workflows` — List available workflow schemas.
- `POST   /api/workflows` — Define custom workflow.
- `GET    /api/task-statuses` — Retrieve statuses belonging to a workflow.
- `POST   /api/task-statuses` — Add custom status (color, position, category).
- `POST   /api/task-statuses/reorder` — Reorder status column positions.

### 2.4. Tasks, Sprints & Ranking (`/api/tasks`, `/api/sprints`, `/api/task-ranks`)
- `GET    /api/tasks` — List tasks with filters (`projectId`, `sprintId`, `statusId`, `assigneeId`, `priority`, `type`, `search`).
- `POST   /api/tasks` — Create task with automatic incremental `taskNumber` and unique `slug` (e.g. `FF-105`).
- `GET    /api/tasks/:id` — Retrieve comprehensive task details (assignees, comments, attachments, subtasks, dependencies, watchers).
- `PATCH  /api/tasks/:id` — Update task fields (status, priority, dates, description).
- `DELETE /api/tasks/:id` — Soft-delete or archive task.
- `POST   /api/task-ranks/reorder` — Update LexoRank for task repositioning in Kanban or List view:
  ```json
  {
    "taskId": "c1f7b824-3456-47b2-a42e-13c2f9d81234",
    "scopeType": "PROJECT",
    "scopeId": "8b34f71a-1234-4567-89ab-cdef01234567",
    "viewType": "BOARD",
    "prevRank": "0|i00003:",
    "nextRank": "0|i00007:"
  }
  ```
- `GET    /api/sprints` — List sprints for project (planning, active, completed).
- `POST   /api/sprints` — Create sprint (name, goal, duration, start/end dates).
- `PATCH  /api/sprints/:id/start` — Start planned sprint.
- `PATCH  /api/sprints/:id/complete` — Complete active sprint and move incomplete tasks to backlog.
- `GET    /api/sprints/:id/burndown` — Retrieve sprint burndown metrics.

### 2.5. Collaboration & Media (`/api/task-comments`, `/api/task-attachments`, `/api/time-entries`)
- `GET    /api/task-comments` — Retrieve threaded comments for task.
- `POST   /api/task-comments` — Add comment with rich text markdown and `@mentions`.
- `POST   /api/task-attachments` — Upload file attachment (multipart/form-data) to S3 or local storage.
- `GET    /api/task-attachments/:id/download` — Secure presigned download URL.
- `POST   /api/time-entries` — Log time spent on task with description and duration.

### 2.6. Conversational AI & Automations (`/api/ai-chat`, `/api/automation`)
- `POST   /api/ai-chat/conversation` — Create or resume AI chat session.
- `POST   /api/ai-chat/message` — Send natural language command; receives structured response with execution instructions.
- `GET    /api/automation/rules` — List active automation rules.
- `POST   /api/automation/rules` — Create trigger-condition-action rule.
- `POST   /api/automation/rules/:id/test` — Dry-run rule against sample task.

### 2.7. Integrations & Inbox (`/api/jira-sync`, `/api/trello-sync`, `/api/projects/:projectId/inbox`)
- `POST   /api/jira-sync/connect` — Configure Jira credentials with SSRF preflight check.
- `POST   /api/jira-sync/import` — Bulk import Jira issues as Orbito tasks.
- `POST   /api/trello-sync/connect` — Connect Trello API key & token.
- `POST   /api/trello-sync/import` — Import Trello boards and lists.
- `GET    /api/projects/:projectId/inbox/messages` — List inbound emails received via IMAP.
- `POST   /api/projects/:projectId/inbox/convert` — Convert inbound email message into task.

---

## 3. Real-Time WebSocket Protocol (`/events`)

### Handshake & Rooms
- **Namespace**: `/events`
- **Handshake Authentication**:
  ```javascript
  const socket = io("http://localhost:3000/events", {
    auth: { token: accessToken },
    transports: ["websocket", "polling"],
  });
  ```
- **Client Subscription Events**:
  - `join:organization` `{ organizationId }`
  - `join:workspace` `{ workspaceId }`
  - `join:project` `{ projectId }`
  - `join:task` `{ taskId }`
  - `leave:project` `{ projectId }`

### Server Broadcast Events
| Event Name | Scope / Room | Payload |
| :--- | :--- | :--- |
| `task:created` | `project:{projectId}` | Full Task entity with status & assignees |
| `task:updated` | `project:{projectId}`, `task:{taskId}` | `{ taskId, updates }` |
| `task:status_changed` | `project:{projectId}` | `{ taskId, oldStatusId, newStatusId, rank }` |
| `task:deleted` | `project:{projectId}` | `{ taskId }` |
| `task:comment_added` | `task:{taskId}`, `project:{projectId}` | Full Comment entity with author |
| `sprint:status_changed`| `project:{projectId}` | `{ sprintId, newStatus }` |
| `user:online` | Global broadcast | `{ userId, isOnline: true }` |
| `user:offline` | Global broadcast | `{ userId, isOnline: false, lastSeen }` |
| `notification:created` | `user:{userId}` | Full Notification entity |
