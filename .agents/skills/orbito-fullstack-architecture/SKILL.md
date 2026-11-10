---
name: orbito-fullstack-architecture
description: >-
  Guides full-stack engineering across Orbito's Turborepo monorepo: Next.js frontend UI/UX, modals, layouts, RTK Query, and NestJS backend service modules, Prisma models, and BullMQ queues.
---

# Orbito Fullstack Architecture & Engineering Runbook

This skill outlines design rules, structural patterns, and implementation procedures for expanding UI components, modals, layout shells, API endpoints, and backend services in Orbito.

---

## 1. Monorepo Organization & Packages

```text
focus-flow-task-manager/
├── backend/            # @orbito/backend: NestJS 11 + Prisma ORM + BullMQ + Redis
├── frontend/           # @orbito/frontend: Next.js 16.1 (App Router) + Tailwind CSS + RTK Query
├── packages/
│   ├── types/          # @orbito/types: Shared TypeScript types, enums, DTO interfaces
│   └── ui/             # @orbito/ui: Core shared design primitives
├── .agents/            # Custom workspace skills and rules
└── AGENTS.md           # Master agent protocol and architectural contract
```

---

## 2. Frontend Engineering Protocol

### 2.1 UI/UX & Design Tokens (Obsidian Monochrome)
- **Palette**: Deep Obsidian backgrounds (`#09090b`), dark panel cards (`bg-zinc-900/90`), subtle zinc borders (`border-zinc-800`), crisp white headers (`text-zinc-100`).
- **Radii**: Strictly 4px to 6px (`rounded-md` or `rounded-sm`). **Zero bloated squircles or cartoon pill badges**.
- **Component Primitives (`frontend/src/components/ui`)**:
  - Radix Primitives: Dialog, Popover, DropdownMenu, Command, Tooltip, Tabs, Select.
  - Consistent focus rings (`focus-visible:ring-1 focus-visible:ring-zinc-600 focus-visible:outline-none`).

### 2.2 Shell & Context-Aware Navigation
- **Navbar (`frontend/src/components/layout/Navbar.tsx`)**:
  - Sticky 52px top bar with breadcrumbs (`Workspace / Project` or `Administration / Subpage`).
  - Command palette shortcut (`⌘K` / `Ctrl+K`) triggering `GlobalSearchDialog`.
  - Role-gated `+ Invite` button opening `InviteMemberModal`.
  - Notifications popover with dynamic badge (`NotificationPopover`).
  - Dark/light mode switcher (`ThemeToggle`).
- **Sidebar (`frontend/src/components/layout/Sidebar.tsx`)**:
  - When no project is active: Shows Workspace Dashboard, Projects directory, Members, and Activity Log.
  - When a project is active: Automatically focuses on Project Board & Tasks, Sprints, with a `← Workspace Overview` back-switch.
  - Includes `Admin Console (Root)` link when `isSuperAdmin === true`.

### 2.3 Modals (`frontend/src/components/modals`)
- All modals must be controlled via `open` and `onOpenChange` props.
- Form submissions must use React Hook Form + Zod, displaying loading states on primary submit buttons.
- Must display descriptive empty and error states.

### 2.4 State Management & RTK Query (`frontend/src/lib/store`)
- Inject endpoints into `apiSlice` using `apiSlice.injectEndpoints({ ... })`.
- Provide and invalidate specific tags (e.g., `'Notification'`, `'Task'`, `'Project'`, `'Workspace'`).
- Always write safe `transformResponse` functions to handle both unnested and nested paginated backend responses.

---

## 3. Backend Architecture Protocol

### 3.1 NestJS Module Separation (`backend/src/modules`)
Each domain capability resides in its own isolated module:
```text
backend/src/modules/<feature>/
├── <feature>.controller.ts      # HTTP endpoints, Swagger annotations, Guards
├── <feature>.service.ts         # Business logic, Prisma transactions
├── <feature>.module.ts          # Dependency injection container
└── dto/                         # Request validation DTOs (class-validator)
```

### 3.2 Layer Rules
- **Controllers**: Handle HTTP serialization, parameter transformation, and route delegation. Zero database queries in controllers.
- **Services**: Encapsulate business rules, permission checks, audit logging, and Prisma queries.
- **Prisma & Database**: All migrations committed through `prisma/schema.prisma`. Use UUID primary keys and soft deletes where applicable.

### 3.3 Routing & Prefixes
- Backend API routes are routed under `/api/...` through `RouterModule.register` in `app.module.ts`.
- Endpoints must declare explicit REST verbs (`@Get`, `@Post`, `@Patch`, `@Delete`).
- Concrete routes (e.g., `/stats/summary`, `/mark-all-read`) must always precede parameterized paths (e.g., `/:id`).

---

## 4. Quality Gates & Verification Checklist

Before pushing or completing any feature:
1. `npm run build --workspace=@orbito/backend`: Verify 0 TypeScript/NestJS compile errors.
2. `npm run build --workspace=@orbito/frontend`: Verify Next.js Turbopack compiles 0 errors.
3. Test all API contracts with authenticated curl or automated unit tests.
4. Verify defensive fallbacks for empty arrays and null states in UI components.

