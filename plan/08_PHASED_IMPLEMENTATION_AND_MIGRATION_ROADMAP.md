# 08. Phased Implementation & Migration Roadmap: Orbito

This roadmap defines the sequenced phases required to migrate and upgrade **Orbito** into a complete, white-labeled, production-grade enterprise project management application with 100% functional parity with Taskosaur, built upon a clean architecture monorepo with Turborepo.

---

## Roadmap Overview

```
Phase 0: Clean Architecture Monorepo Layout & Master Plans  <-- (Current Milestone)
   │
   ▼
Phase 1: Database Schema, Prisma Migrations & Master Seeders
   │
   ▼
Phase 2: Core Backend Engine, Auth & Multi-Tenancy (Org, Workspace, Project)
   │
   ▼
Phase 3: Shared Types Contract (@orbito/types), Task Aggregate & LexoRank Engine
   │
   ▼
Phase 4: Real-Time WebSockets (/events), Agile Sprints & Dynamic Views
   │
   ▼
Phase 5: Conversational AI Engine, DOM Automation Agent & Background Rules
   │
   ▼
Phase 6: Enterprise Integrations (Jira, Trello) & Inbound IMAP Email Processing
   │
   ▼
Phase 7: Frontend Monochrome Polish (Linear/Slack/Monday UX) & Shared UI System
   │
   ▼
Phase 8: Production Hardening, CI/CD Pipeline & Deployment
```

---

## Detailed Phases

### Phase 0: Clean Architecture Monorepo Layout & Master Plans (Current)
- **Objectives**:
  - Establish complete architectural specifications in `plan/` with **Orbito** branding.
  - Set up clean architecture monorepo workspaces:
    - `backend/` (NestJS Clean Architecture API server, port 3000)
    - `frontend/` (Next.js modern web application, port 3001)
    - `packages/types` (Shared TypeScript contracts, DTOs, and enums under `@orbito/types`)
    - `packages/ui` (Shared monochrome design system components under `@orbito/ui`)
    - `packages/tsconfig` and `packages/eslint-config` (Shared configs under `@orbito/*`)
  - Configure root `package.json` with workspace and Turborepo scripts for `dev`, `build`, `lint`, `test`, `db:migrate`, `db:seed`.
  - Set up environment variables (`.env.example`) and Docker Compose environments (`docker-compose.dev.yml`).
- **Deliverables**:
  - Comprehensive `plan/` documentation suite (01 through 09 + README).
  - Validated monorepo workspace topology.
  - Docker Compose service definitions for PostgreSQL 16 and Redis 7.

---

### Phase 1: Database Schema, Prisma Migrations & Master Seeders
- **Objectives**:
  - Implement the complete Orbito schema in `backend/prisma/schema.prisma` (38 models, 20 enums).
  - Verify all 74 PostgreSQL migration histories in `backend/prisma/migrations/`.
  - Build idempotent master database seeder (`src/seeder/seeder.command.ts` / `demo-seed.ts`):
    - Default Super Admin (`admin@orbito.io`).
    - Demo Organization (`Orbito Engineering`), Workspaces (`Platform Core`, `Mobile App`).
    - Default agile workflows (`Orbito Agile Standard`), statuses (`TODO`, `IN_PROGRESS`, `IN_REVIEW`, `DONE`), and transitions.
    - Demo project (`Orbito Web Redesign`) with task identifier prefix `ORB`, realistic backlog tasks, active sprint, and dependencies.
- **Verification Gates**:
  - `npm run db:migrate` applies cleanly without warnings.
  - `npm run db:seed` runs idempotently multiple times without duplicate key violations.

---

### Phase 2: Core Backend Engine, Auth & Multi-Tenancy
- **Objectives**:
  - Implement NestJS authentication module (`auth`):
    - Local email/password authentication with bcrypt hashing.
    - JWT access token generation and refresh token rotation via HttpOnly cookies.
    - OIDC / Single Sign-On scaffolding (Google, GitHub, SAML).
    - Password reset token generation with email delivery.
  - Implement Multi-Tenancy modules (`organizations`, `workspaces`, `projects`):
    - Organization CRUD with unique slug generation.
    - Hierarchical workspace tree traversal with path indexing.
    - Project creation with custom task prefixes (e.g. `FF`).
    - Granular RBAC membership guards across Org, Workspace, and Project tiers.
- **Verification Gates**:
  - Auth integration tests verify token issuance, refresh rotation, and invalidation.
  - Multi-tenancy access checks prevent cross-organization data leakage.

---

### Phase 3: Shared Types Contract (@orbito/types), Task Aggregate & LexoRank
- **Objectives**:
  - Export all domain entities, task DTOs, and event signatures in `@orbito/types`.
  - Implement Task core module in `backend`:
    - Sequential task numbering per project (`taskNumber`) and slug generation (`slug` = `${prefix}-${taskNumber}`).
    - Subtasks, recurring task schedules, watchers, time entries, and tags/labels.
    - Multi-assignee (`TaskAssignee`) and multi-reporter (`TaskReporter`) join relationships.
  - Implement LexoRank ordering engine (`task-ranks`):
    - Alphanumeric midpoint string calculation between two ranks.
    - Efficient reordering endpoint updating single rows without table locking.
  - Implement Workflow state machine (`workflows`, `task-statuses`, `status-transitions`):
    - Enforce transition rules when moving tasks between status columns.
- **Verification Gates**:
  - Drag-and-drop LexoRank unit tests prove collision-free sorting between adjacent items.
  - State machine transition tests block unauthorized state jumps.

---

### Phase 4: Real-Time WebSockets (/events), Agile Sprints & Dynamic Views
- **Objectives**:
  - Configure Socket.io gateway in `backend/src/gateway` on namespace `/events`:
    - JWT handshake authentication and room management (`org:{id}`, `workspace:{id}`, `project:{id}`, `user:{id}`).
    - Broadcast events: `task:created`, `task:updated`, `task:moved`, `task:deleted`, `comment:added`, `user:typing`, `presence:sync`.
  - Implement Agile Sprints module:
    - Sprint lifecycle (PLANNING -> ACTIVE -> COMPLETED).
    - Sprint velocity calculation and burndown chart data aggregation.
  - Implement dynamic project views:
    - Kanban Board, List View, Calendar View, and interactive Gantt timeline.
- **Verification Gates**:
  - Multi-client WebSocket tests verify simultaneous real-time updates and presence indicators.
  - Sprint lifecycle transitions correctly archive completed tasks and move incomplete tasks to backlog.

---

### Phase 5: Conversational AI Engine, DOM Agent & Automation Rules
- **Objectives**:
  - Conversational AI Orchestrator:
    - Natural language task parsing and execution (*"Create high-priority bug assigned to Alex"*).
    - Multi-turn conversational context with tool-calling capabilities.
  - Browser DOM Automation Agent:
    - In-app execution agent capable of navigating routes, opening modals, and manipulating board items.
  - Background Automation Engine:
    - Trigger-Condition-Action rule engine (e.g., status changed -> reassign -> send email alert).
    - Event bus listening to internal domain events and dispatching async jobs to BullMQ.
- **Verification Gates**:
  - AI prompt tests verify correct structured intent extraction and DTO dispatch.
  - Automation tests confirm actions trigger reliably without latency spikes.

---

### Phase 6: Enterprise Integrations (Jira, Trello) & Inbound IMAP Email
- **Objectives**:
  - Enterprise sync modules:
    - Jira bidirectional webhook and REST sync with SSRF defense (DNS pinning, loopback block).
    - Trello board and card synchronization.
  - Inbound IMAP Email Processor:
    - Background polling of configured project mailboxes.
    - Email parsing, thread matching, attachment upload to S3/local storage, and task conversion.
- **Verification Gates**:
  - SSRF test suite confirms attempted loopback/internal IP requests are blocked.
  - Email ingestion simulator tests verify parsed tasks inherit sender identity and project defaults.

---

### Phase 7: Frontend Re-architecture (Linear/Slack/Monday UX & Shared UI System)
- **Objectives**:
  - Implement modern state & data layer:
    - Centralized API client with JWT refresh interceptors.
    - Standardized cache invalidation (`Task`, `Project`, `Workspace`, `Sprint`, `User`, `Comment`).
    - Real-time Socket.io state synchronization patching views on `task:moved`, `task:updated`, and `task:created`.
    - Optimistic updates on drag-and-drop task movements.
  - Build `@orbito/ui` design system & high-velocity modern monochrome UX:
    - Ultra-clean monochrome theme (Obsidian Black, Crisp White, Zinc Gray).
    - Linear-grade tactile buttons, keyboard command palette (`Cmd+K`), and keyboard shortcuts.
    - Slack-inspired channel & project navigation sidebar with active user presence bubbles.
    - Monday.com-grade dynamic view switcher (Kanban, List, Calendar, Sprints, Gantt).
    - Modern App Router and component architecture with full accessibility.
- **Verification Gates**:
  - Zero UI layout shift during drag-and-drop LexoRank reorders with optimistic rollbacks on error.
  - Automatic JWT token refresh executes transparently without interrupting user workflow on 401.
  - Lighthouse accessibility and performance score > 90.

---

### Phase 8: Production Hardening, CI/CD Pipeline & Deployment
- **Objectives**:
  - Multi-stage Docker containers (`Dockerfile.dev`, `Dockerfile.prod`).
  - Production `docker-compose.yml` with healthchecks, Redis sentinel/cluster support, and Nginx reverse proxy.
  - GitHub Actions CI workflow running `npm run lint`, `npm test`, `npm run build`.
  - Production documentation, Swagger API explorer, and security audit.
- **Verification Gates**:
  - Clean container builds without bloat (< 150MB per final image).
  - Passing automated end-to-end integration test suite.
