# Implementation Plan: Go Twin Backend (`apps/api-go`)

## Overview
This plan specifies the implementation breakdown for building a high-performance Go twin backend (`apps/api-go`) of the existing NestJS backend in `@orbito/platform`. The Go application serves as a drop-in replacement that connects to the identical PostgreSQL 16 database, preserves all existing user credentials, bcrypt password hashes, and JWT/cookie sessions without user logout, mirrors all 386 API endpoints across 39 modules with exact JSON request/response shapes, implements Socket.IO real-time events on `/events`, and runs background workers and scheduled crons.

## Architecture Decisions
- **HTTP Routing:** `go-chi/chi/v5` for lightweight, standard-library-compatible HTTP routing and composable middleware.
- **Database Engine & ORM:** `jackc/pgx/v5` connection pool with `sqlc` compiled directly against PostgreSQL 16 schema (`backend/prisma/migrations/20261106114254_init_the_orbit/migration.sql`), ensuring zero reflection overhead and compile-time SQL verification without altering Prisma schema.
- **Structured Logging:** Go `log/slog` outputting structured JSON with request correlation IDs and user IDs.
- **Validation:** `go-playground/validator/v10` mirroring NestJS `class-validator` rules, formatting validation failures as an array of strings in 400 responses.
- **Authentication & Security:** HS256 JWTs via `golang-jwt/jwt/v5`, native `bcrypt` with cost 12, AES-256-GCM cookie encryption with SHA-256 key derivation matching `cookie-crypto.utils.ts`, and scrypt field encryption matching `encryption.util.ts`.
- **Real-Time Gateway:** Socket.IO v4 protocol engine supporting the `/events` namespace and existing frontend rooms.
- **Background Tasks:** Redis-backed job processing via `hibiken/asynq` with fallback handling, paired with `robfig/cron/v3` for scheduled cron jobs.

## Task List Summary

### Phase 0: Foundations & Project Scaffolding
- [x] Task 1: Scaffolding, `go.mod`, Turbo Integration, Makefile, Air, Linter & Dockerfile
- [x] Task 2: Typed Configuration & Environment Validation (`internal/config`)
- [x] Task 3: Platform Logger (`slog`), Request ID & Ingress Middleware
- [ ] Task 4: Platform Database (`pgx/v5`) & sqlc Setup (`db/sqlc.yaml`, migration DDL)
- [ ] Task 5: Platform Crypto (AES-256-GCM cookie crypto, scrypt field encryption, bcrypt) & Unit Test Parity
- [ ] Task 6: Shared HTTP & Error Framework (`httpx`, `apperror` mapper, 5-field pagination)

### Checkpoint: Foundations Complete
- [ ] Platform database connects to Postgres, crypto test vectors match TypeScript, middleware logs JSON with correlation ID

### Phase 1: Core System & Authentication
- [ ] Task 7: Health Module (`GET /api/health`, readiness check)
- [ ] Task 8: Settings Module (`/api/settings` key-value store & registration status)
- [ ] Task 9: Mailer Platform & Email Module (`/api/email/send`, stats, SMTP client)
- [ ] Task 10: Auth Core & JWT Strategy (claims, token minting, `RequireAuth` middleware)
- [ ] Task 11: Auth Module: Registration & Login (`/api/auth/register`, `/api/auth/login`, cookies, `/me`, `/profile`)
- [ ] Task 12: Auth Module: Token Refresh & Logout (`/api/auth/refresh`, `/api/auth/logout`, user status check)
- [ ] Task 13: Auth Module: Password Reset & Setup (`/forgot-password`, `/verify-reset-token`, `/reset-password`, `/setup`)
- [ ] Task 14: Auth SSO / OIDC Services & Endpoints (`/api/auth/oidc/config`, `/login`, `/callback`, `/exchange`)
- [ ] Task 15: Users Module (`/api/users` CRUD, `/status`, `/change-password`, profile avatar)
- [ ] Task 16: Admin Module (`/api/admin/dashboard`, `/users`, `/organizations`, `/config`, `/test-smtp`)

### Checkpoint: Core System & Auth Verified
- [ ] Existing users can log in with existing passwords, receive valid JWTs and encrypted cookies, access `/me` and admin routes

### Phase 2: Multi-Tenancy, Hierarchy & Member RBAC
- [ ] Task 17: RBAC Core Middleware (Hierarchy rank comparison, Scope resolver for ORG/WORKSPACE/PROJECT)
- [ ] Task 18: Organizations Module (`/api/organizations` CRUD, transfer ownership, default org)
- [ ] Task 19: Organization Members Module (`/api/organization-members` list, update role, remove)
- [ ] Task 20: Workspaces Module (`/api/workspaces` CRUD, `/charts` metrics aggregation, archive)
- [ ] Task 21: Workspace Members Module (`/api/workspace-members` list, role update, remove)
- [ ] Task 22: Invitations Module (`/api/invitations` create token, verify, accept, revoke, resend)

### Checkpoint: Multi-Tenancy & RBAC Verified
- [ ] Tenancy boundary enforced, role ranks strictly checked, workspace charts calculate correct aggregates

### Phase 3: Project Domain & Agile Foundations
- [ ] Task 23: Projects Module (`/api/projects` CRUD, slug/key, visibility, custom task prefixes, stats)
- [ ] Task 24: Project Members Module (`/api/project-members` list, add, role update, remove)
- [ ] Task 25: Workflows Module (`/api/workflows` CRUD, default workflow generation)
- [ ] Task 26: Task Statuses Module (`/api/task-statuses` custom status columns, categories, transitions)
- [ ] Task 27: Labels Module (`/api/labels` global and workspace labels CRUD)
- [ ] Task 28: Sprints Module (`/api/sprints` planning, active, complete, burndown metrics)

### Checkpoint: Project Domain & Workflows Verified
- [ ] Projects, workflows, statuses, labels, and sprints function with exact status transition validation

### Phase 4: Task Execution Engine & Agile Work Management
- [ ] Task 29: Task Ranks Module & Lexorank Algorithm (`internal/modules/task_ranks`)
- [ ] Task 30: Tasks Core CRUD & Slugs (`/api/tasks` create, get by id/slug/key, update, delete, archive)
- [ ] Task 31: Tasks Listing, Filtering & Grouping (`/api/tasks` list, `/grouped`, `/by-status`, `/today`, `/all-tasks`)
- [ ] Task 32: Tasks Bulk Operations (`/api/tasks/bulk-create`, `/bulk-delete`, `/bulk-status-update`, `/bulk-assign`)
- [ ] Task 33: Tasks Reordering & Status Transitions (`/api/tasks/:id/reorder`, `/status`, `/assignees`, `/priority`)
- [ ] Task 34: Task Recurrence Engine & Cron (`/api/tasks/:id/recurrence`, complete occurrence, daily generation)
- [ ] Task 35: Task Dependencies Module (`/api/task-dependencies` blocking, blocked-by relations)
- [ ] Task 36: Task Watchers & Labels Association (`/api/task-watchers`, `/api/task-label`)
- [ ] Task 37: Task Comments Module (`/api/task-comments` CRUD, user mentions `@name`)
- [ ] Task 38: Task Attachments Module (`/api/task-attachments` file upload, preview, download, stats)
- [ ] Task 39: Time Entries Module (`/api/time-entries` timer start/stop, log, summary, IDOR checks)
- [ ] Task 40: Gantt Module (`/api/gantt` project timeline, sprint schedule, resource allocation)

### Checkpoint: Task Execution Engine Verified
- [ ] Kanban boards, task detail drawers, list views, bulk actions, reordering, and timers match NestJS responses

### Phase 5: Real-Time Collaboration & Audit
- [ ] Task 41: Real-Time Events Gateway (Socket.IO v4 protocol engine at `/events`, room management)
- [ ] Task 42: WebSocket Event Dispatcher & Presence (`user:online`, `typing`, `task:created/updated`, etc.)
- [ ] Task 43: Notifications Module (`/api/notifications` list, unread count, mark-read, delete)
- [ ] Task 44: Activity Log Module (`/api/activity-log` audit trail capture, project & org stats)

### Checkpoint: Real-Time & Event Dispatching Verified
- [ ] Frontend establishes live WebSocket connection, receives real-time room broadcasts and notifications

### Phase 6: Async Workers, Storage, Integrations & Public Access
- [ ] Task 45: Queue Engine & Background Processors (`asynq`/Redis worker runtime, `bulk-task-import`, `automation`)
- [ ] Task 46: Scheduled Jobs Engine (Robfig cron: hourly reminders, 9am overdue alerts, midnight recurring tasks)
- [ ] Task 47: Storage & File Management (`/uploads/upload/:folder`, `/uploads/:folder/:filename`, `/s3/presigned-*`)
- [ ] Task 48: Editor Images & Media Assets (`/api/editor-images/upload`)
- [ ] Task 49: Search Module (`/api/search/global`, `/advanced`, `/suggestions`, `/quick`)
- [ ] Task 50: Project Inbox & Email Sync (`/api/inbox` project email accounts, rules, IMAP processor)
- [ ] Task 51: Trello Sync Module & Scheduler (`/api/trello-sync` credentials, board import, sync loop)
- [ ] Task 52: Jira Sync Module (`/api/jira-sync` safe outbound HTTP, credentials, project import, webhook sync)
- [ ] Task 53: AI Chat Module (`/api/ai-chat` conversations, chat endpoint, tools, test-connection)
- [ ] Task 54: Public Module (`/api/public` shared tasks, calendar, sprints, public project view)

### Checkpoint: Full Platform Feature Parity Verified
- [ ] Background crons run, files upload and serve, integrations sync, and public shares resolve without authentication

### Phase 7: Verification, Parity Diffing & Production Readiness
- [ ] Task 55: Automated Parity Test Harness (Dual HTTP request fuzzer comparing NestJS vs Go responses)
- [ ] Task 56: Production Containerization & CI/CD Pipeline (`deploy/Dockerfile`, GitHub Actions workflow)

### Checkpoint: Final Acceptance & Ready for Production Cutover
- [ ] All 386 endpoints pass parity tests, distroless Docker image builds, zero lint warnings, ready for drop-in cutover

## Risks and Mitigations
| Risk | Impact | Mitigation |
|:---|:---|:---|
| **Cookie Decryption Discrepancy** | High | Test vectors verified against Node.js `cookie-crypto.utils.ts` in Task 5 unit tests before auth integration. |
| **Password Hash Incompatibility** | High | Standard bcrypt cost 12 verified with unit tests against hashes created by NestJS. |
| **Lexorank Calculation Drift** | Medium | Direct port of string midpoint math in `task_ranks` with dedicated unit test suite. |
| **Database Connection Exhaustion** | Medium | `pgxpool` configured with `MaxConns: 50`, `MinConns: 10`, `MaxConnIdleTime: 30m`. |
| **Outbound SSRF in Integrations** | High | Safe network dialer pinning DNS and blocking loopback, RFC 1918, and metadata IPs (`169.254.169.254`). |

## Open Questions
- None. All architectural constraints, contracts, ports, and protocols have been verified.
