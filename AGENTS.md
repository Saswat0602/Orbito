# AGENTS.md — Orbito Platform Architecture & Engineering Contract

> **Master Engineering & Operational Guidelines for Autonomous Agents and Engineers.**  
> Platform: **Orbito** (Enterprise Task, Project & Work Management System)  
> Monorepo: `@orbito/platform` powered by Turborepo  

---

## 1. System Overview & Technology Stack

Orbito is a multi-tenant, enterprise-grade project and work management platform engineered with modern full-stack TypeScript:
- **Frontend App (`frontend/`)**: Next.js 16.1.3 App Router, React 19, Tailwind CSS v4, Lucide Icons, Redux Toolkit Query, Radix UI.
- **Backend API (`backend/`)**: NestJS 11, Prisma ORM 6, PostgreSQL, Redis 7, BullMQ queue processing, Passport.js JWT & OIDC.
- **Shared Packages (`packages/`)**:
  - `@orbito/types`: Shared enums, domain entities, DTOs, and RBAC contracts.
  - `@orbito/ui`: Shared design tokens and primitive UI elements.
- **Runtime Ports & Infrastructure**:
  - Frontend Web Server: `http://localhost:3001`
  - Backend REST API: `http://localhost:3000/api` (Swagger UI at `/api/docs`)
  - PostgreSQL Database: `localhost:5433` (Database: `focus_flow_db`)
  - Redis Cache & BullMQ: `localhost:6379`

---

## 2. Monorepo Structure & Directory Map

### 2.1 Frontend Architecture (`frontend/src`)

- **`components/layout/`**: Application shell and top-level navigation.
  - `AppShell.tsx`: Global responsive wrapper combining Sidebar, Navbar, and `<main>` viewport.
  - `Navbar.tsx`: Sticky 52px top header featuring breadcrumbs (`Workspace / Project` or `Administration / Subpage`), global search shortcut (`⌘K`), role-gated `+ Invite` button, `NotificationPopover`, and `ThemeToggle`.
  - `Sidebar.tsx`: Collapsible navigation adjusting dynamically based on active project selection.
  - `WorkspaceProjectSwitcher.tsx`: Dual-scope selector allowing switching between Workspaces, specific Projects, or "All Projects (Overview)".
  - `NotificationPopover.tsx`: Popover showing real-time notifications, unread count badge, mark-all-read, and deep-link routing.
  - `GlobalSearchDialog.tsx`: Fast command palette (`⌘K`) for jumping to projects, tasks, members, and admin tools.
  - `ThemeToggle.tsx`: Dark/light mode switcher with subtle zinc borders and 4px radii.
- **`components/modals/`**: Controlled overlays and dialogues.
  - `InviteMemberModal.tsx`: Email invitation modal with role picker (`MEMBER`, `MANAGER`, `VIEWER`), expiration toggle, and copyable invite link.
- **`components/ui/`**: Obsidian-themed design primitives built on Radix UI.
  - Primitives: `button.tsx`, `card.tsx`, `dialog.tsx`, `dropdown-menu.tsx`, `popover.tsx`, `command.tsx`, `tabs.tsx`, `badge.tsx`, `input.tsx`, `select.tsx`, `table.tsx`, `sonner.tsx`, etc.
  - Aesthetics: Deep Obsidian (`#09090b`), dark zinc panels (`#18181b`), subtle 1px borders (`border-zinc-800`), 4px-6px radii (`rounded-md` / `rounded-sm`). **Zero bloated squircles or cartoon badges**.
- **`components/providers/`**: Context and store wrappers.
  - `StoreProvider.tsx`: Mounts Redux Toolkit store.
  - `AuthProvider.tsx`: Syncs authentication status, validates active cookies, and redirects expired sessions to `/login`.
- **`lib/`**: Frontend core logic and state.
  - **`lib/store/`**: Redux store setup (`index.ts`, `hooks.ts`).
    - `api/apiSlice.ts`: Core RTK Query slice with `async-mutex` synchronized token refresh on 401 errors.
    - Slices: `authApi.ts`, `workspacesApi.ts`, `projectsApi.ts`, `tasksApi.ts`, `sprintsApi.ts`, `notificationsApi.ts`, `invitationsApi.ts`, `adminApi.ts`, `organizationsApi.ts`, `timeEntriesApi.ts`.
  - **`lib/hooks/`**: Reusable custom hooks.
    - `useAuthRbac.ts`: Provides user identity, role booleans (`isSuperAdmin`, `isOwner`, `isManager`, `isMember`, `isViewer`), and capability checks.
  - **`lib/api/`**: Network endpoints contract.
    - `endpoint.ts`: Defines `API_BASE_URL` (`http://localhost:3000/api`) and route map `ENDPOINTS`.
- **`app/`**: Next.js App Router hierarchy.
  - `/[workspaceSlug]/page.tsx`: Workspace dashboard and project grid.
  - `/[workspaceSlug]/[projectSlug]/tasks/page.tsx`: Kanban board and list view.
  - `/[workspaceSlug]/[projectSlug]/sprints/page.tsx`: Sprint planning and active backlog.
  - `/[workspaceSlug]/members/page.tsx`: Workspace member directory and permissions.
  - `/admin/`: System admin dashboard, users table, organizations, settings, and BullMQ queues.
  - `/login/`, `/register/`, `/forgot-password/`: Authentication flows.

---

### 2.2 Backend Architecture (`backend/src`)

- **`modules/`**: Modular domain feature packages.
  - `auth/`: Authentication controller, JWT & OIDC strategies, `JwtAuthGuard`, `RolesGuard`, cookie utilities.
  - `users/`: User profile management, password resets, user directory.
  - `organizations/` & `organization-members/`: Multi-tenant organization boundaries and membership.
  - `workspaces/` & `workspace-members/`: Workspace scopes within organizations.
  - `projects/` & `project-members/`: Projects, custom task prefixes, project settings.
  - `tasks/`, `task-comments/`, `task-statuses/`, `task-watchers/`, `task-attachments/`, `task-dependencies/`, `task-ranks/`: Task management engine with automated ordering and Lexorank.
  - `sprints/`: Agile sprint lifecycles (planning, active, completed, burndown metrics).
  - `notifications/`: Alerting pipeline, unread counts, bulk actions, and email event dispatching.
  - `invitations/`: Workspace invitation token lifecycle, verification, and acceptance.
  - `queue/`: BullMQ worker and queue management with automated fallback if Redis fails.
  - `admin/`: System-wide audit logs, queue monitoring, tenant metrics, and user controls.
  - `ai-chat/`: Integrated AI conversational assistant.
  - `health/`: Live system status (`GET /api/health`).
- **`core/`**: Clean architecture foundation.
  - `entity.base.ts`: Base domain entity contract.
  - `repository.interface.ts`: Generic repository interface.
  - `use-case.interface.ts`: Application use-case interface.
  - `value-object.ts`: Immutable domain value objects.
- **`prisma/`**: Database persistence layer.
  - `prisma.service.ts`: Global lifecycle-managed Prisma client.
  - `audit.middleware.ts`: Soft delete handling and audit timestamps.
  - `schema.prisma`: Authoritative database schema with UUIDs and relational constraints.
- **`config/`**: Central configuration.
  - `configuration.ts`: Validates environment variables (JWT secrets, CORS origins, Redis/DB URLs, S3).
- **`constants/`**: Shared backend constants.
  - `roles.ts`: System roles and hierarchy.
  - `defaultWorkflow.ts`: Default task statuses (Backlog, To Do, In Progress, Review, Done).

---

## 3. Core Architectural & Anti-Hallucination Rules

### Rule 1: Zero-Hallucination Protocol
- **Never speculate imports or APIs**: Always verify that a module or package exists in `package.json` or the workspace file tree before importing.
- **Chesterton's Fence**: Never delete or replace existing backend routes, middleware, or frontend state logic without understanding why it exists.
- **Type Safety**: Strictly zero `any`. Use `unknown` with type narrowing, generics, or schemas from `@orbito/types`.

### Rule 2: Strict Layer Separation
- **No Database Logic in UI**: Never execute database queries, Prisma calls, or raw fetch requests inside React components. All network requests must flow through RTK Query slices in `frontend/src/lib/store/api/`.
- **No Database Logic in Controllers**: Controllers handle only HTTP parsing, parameter validation (`ParseUUIDPipe`), and delegate immediately to domain Services.
- **No Dumping-Ground Files**: Do not create generic `utils.ts`, `helpers.ts`, or `common.ts` catch-all files. Co-locate utilities within their specific feature domain.

### Rule 3: Design & UI/UX Standards
- **Obsidian Monochrome**: Surfaces use `#09090b` (background), `zinc-900` (cards), and `zinc-800` (1px borders).
- **Geometric Precision**: All radii are 4px (`rounded-md` / `rounded-sm`). Avoid large rounded pills (`rounded-full` is restricted to avatars and indicator dots).
- **Context-Aware Navigation**: When viewing workspace overviews, project scope is cleared. When inside a project, navigation focuses strictly on project boards and sprints.

### Rule 4: Role-Based Access Control (RBAC)
- All privileged routes must be protected on backend:
  ```typescript
  @UseGuards(JwtAuthGuard, RolesGuard)
  @Roles(Role.SUPER_ADMIN, Role.OWNER)
  ```
- All privileged actions in UI must be checked via `useAuthRbac()`:
  - `SUPER_ADMIN`: Root console `/admin/*`.
  - `OWNER` / `MANAGER`: Workspace settings, member invites, sprint starting/closing.
  - `VIEWER`: Read-only. Must never see active edit or delete triggers.

---

## 4. Common Commands & Runbooks

- **Build Backend**: `npm run build --workspace=@orbito/backend`
- **Build Frontend**: `npm run build --workspace=@orbito/frontend`
- **Prisma Migrations**: `npm run prisma:migrate:dev --workspace=@orbito/backend`
- **Prisma Generate**: `npm run prisma:generate --workspace=@orbito/backend`
- **Start Dev Environment**:
  - Backend: `npm run dev --workspace=@orbito/backend` (port 3000)
  - Frontend: `npm run dev --workspace=@orbito/frontend` (port 3001)
