# 03. Clean Architecture Monorepo Blueprint: Orbito

## 1. Clean Architecture Principles

To transform **Orbito** into a production-grade enterprise platform, the codebase adheres to **Clean Architecture** (Robert C. Martin) and **Domain-Driven Design (DDD)** principles within an enterprise monorepo:

1. **Independent of Frameworks**: Business rules do not depend on NestJS, Prisma, Next.js, or Express. Frameworks are plugins to the domain.
2. **Testable**: Business rules can be tested without the database, web server, or UI components.
3. **Independent of UI**: The frontend UI can be evolved (e.g. into the ultra-modern monochrome Linear/Slack/Monday UX) without altering domain entities or API contracts.
4. **Independent of Database**: PostgreSQL/Prisma can be swapped or modified without altering core business rules.
5. **Dependency Rule**: Source code dependencies always point inwards toward higher-level policies.
6. **Type Contract Sharing via Monorepo**: Frontend and backend share exact, zero-drift TypeScript contracts from `@orbito/types`.

```
       +---------------------------------------------+
       | Presentation / Controllers / WebSockets     |
       |  +---------------------------------------+  |
       |  | Application / Use Cases / DTOs        |  |
       |  |  +---------------------------------+  |  |
       |  |  | Domain / Entities / Aggregates   |  |  |
       |  |  | Repository Interfaces & Events  |  |  |
       |  |  +---------------------------------+  |  |
       |  +---------------------------------------+  |
       | Infrastructure / Prisma / BullMQ / S3 / Sockets
       +---------------------------------------------+
```

---

## 2. Monorepo Architecture

The repository is structured as a clean architecture monorepo with dedicated `backend/`, `frontend/`, and `packages/` workspaces:

```
orbito-platform/ (focus-flow-task-manager/)
├── .github/                       # CI/CD workflows (GitHub Actions)
│   ├── workflows/
│   │   ├── ci.yml                 # Lint, typecheck, unit & integration test
│   │   └── deploy.yml             # Docker build and multi-environment deployment
├── backend/                       # NestJS Backend API Server (Port 3000)
│   ├── prisma/                    # Prisma schema, migrations, idempotency seeders
│   │   ├── schema.prisma          # Relational schema (38 models, 20 enums)
│   │   ├── migrations/            # 74 PostgreSQL migration history directories
│   │   └── seed.ts                # Master seeder with admin & demo organizations
│   ├── src/
│   │   ├── core/                  # Clean Architecture domain primitives
│   │   │   ├── entity.base.ts     # Aggregate root / BaseEntity
│   │   │   ├── value-object.ts    # Immutable ValueObject
│   │   │   ├── repository.interface.ts # IRepository<T> contract
│   │   │   └── use-case.interface.ts   # IUseCase<Input, Output> contract
│   │   ├── common/                # Enterprise guards, interceptors, filters
│   │   │   ├── decorators/        # CurrentUser, Roles, Public, Permissions
│   │   │   ├── filters/           # Global exception filters
│   │   │   ├── guards/            # JwtAuthGuard, RolesGuard, SSRFGuard
│   │   │   ├── interceptors/      # AuditLog, ActivityNotification, RequestContext
│   │   │   └── safe-outbound-http.ts # SSRF defense, DNS pinning, IP validator
│   │   ├── config/                # Environment configuration loader
│   │   ├── constants/             # Domain constants (defaultWorkflow, roles)
│   │   ├── gateway/               # Real-time WebSockets (/events namespace)
│   │   │   ├── events.gateway.ts
│   │   │   └── gateway.module.ts
│   │   ├── modules/               # 41 Domain feature modules
│   │   │   ├── auth/              # JWT, SSO/OIDC, password reset, guards
│   │   │   ├── users/             # User domain, preferences, onboarding
│   │   │   ├── organizations/     # Multi-tenancy root
│   │   │   ├── workspaces/        # Workspace hierarchy & nesting
│   │   │   ├── projects/          # Projects, custom prefixes (FF), statuses
│   │   │   ├── tasks/             # Task aggregate, subtasks, recurring
│   │   │   ├── task-ranks/        # LexoRank alphanumeric ordering engine
│   │   │   ├── sprints/           # Agile sprints & velocity
│   │   │   ├── workflows/         # Workflows & state machines
│   │   │   ├── ai-chat/           # Conversational AI & prompt orchestration
│   │   │   ├── automation/        # Background rules processor
│   │   │   ├── inbox/             # IMAP email integration & task routing
│   │   │   ├── jira-sync/         # Jira bidirectional sync & SSRF protection
│   │   │   ├── trello-sync/       # Trello board & card synchronization
│   │   │   ├── queue/             # BullMQ + BetterQueue dual-mode adapter
│   │   │   └── storage/           # Local file storage + AWS S3 presigner
│   │   ├── seeder/                # Seeder modules and services
│   │   ├── app.module.ts          # Root module orchestrator
│   │   └── main.ts                # Bootstrap, Swagger, CORS, Static routes
│   ├── test/                      # Unit & E2E integration tests
│   ├── package.json               # name: "@orbito/backend"
│   ├── tsconfig.json
│   ├── tsconfig.build.json
│   └── nest-cli.json
│
├── frontend/                      # Modern Next.js App & Pages Router (Port 3001)
│   ├── public/                    # Orbito brand assets, svgs, icons
│   ├── src/
│   │   ├── app/                   # App Router pages & dynamic layouts
│   │   │   ├── globals.css        # Monochrome Linear obsidian tokens
│   │   │   ├── layout.tsx         # Root layout with Theme & Providers
│   │   │   ├── page.tsx           # Home / Dashboard
│   │   │   ├── tasks/             # Task board and detail views
│   │   │   └── month/             # Calendar month view
│   │   ├── components/            # Clean component hierarchy
│   │   │   ├── layout/            # Header, Sidebar, ThemeToggle
│   │   │   ├── tasks/             # TaskBoard, TaskCard, TaskModal, StatusColumn
│   │   │   ├── chat/              # Conversational AI ChatPanel
│   │   │   ├── ui/                # CustomSelect, PriorityBadge, Buttons
│   │   │   ├── kanban/            # Kanban board components
│   │   │   └── sprints/           # Sprint planning and burndown
│   │   ├── contexts/              # React Context state management
│   │   │   ├── auth-context/
│   │   │   ├── project-context/
│   │   │   ├── workspace-context/
│   │   │   ├── sprint-context/
│   │   │   └── task-context/
│   │   ├── hooks/                 # Custom React hooks
│   │   ├── lib/                   # API client, WebSocket client, automation
│   │   │   ├── api.ts             # Axios HTTP client with JWT interceptor
│   │   │   ├── socket.ts          # Socket.io client singleton
│   │   │   └── browser-automation/# In-app DOM action executor
│   │   └── types/                 # Frontend domain types
│   ├── package.json               # name: "@orbito/frontend"
│   └── tsconfig.json
│
├── packages/                      # Shared Monorepo Packages
│   ├── types/                     # Shared TypeScript domain contracts
│   │   ├── src/index.ts           # Roles, priorities, statuses, DTOs
│   │   ├── package.json           # name: "@orbito/types"
│   │   └── tsconfig.json
│   ├── ui/                        # Monochrome Design System Components
│   │   ├── src/index.ts           # Linear/Slack/Monday design tokens
│   │   ├── package.json           # name: "@orbito/ui"
│   │   └── tsconfig.json
│   ├── tsconfig/                  # Shared TypeScript Configurations
│   │   ├── base.json
│   │   ├── nextjs.json
│   │   ├── nestjs.json
│   │   └── package.json           # name: "@orbito/tsconfig"
│   └── eslint-config/             # Shared ESLint Configuration
│       ├── index.js
│       └── package.json           # name: "@orbito/eslint-config"
│
├── docker/                        # Containerization configs
│   ├── Dockerfile.dev             # Local development container
│   ├── Dockerfile.prod            # Multi-stage production container
│   ├── entrypoint-dev.sh          # Dev bootstrapper with healthchecks
│   ├── entrypoint.sh              # Prod entrypoint
│   └── README.md
│
├── scripts/                       # DevOps & maintenance tooling
│   ├── build-dist.js              # Distribution package builder
│   ├── generate-logo-icons.js     # Brand asset generator
│   └── postinstall.js             # Git hooks installer
│
├── plan/                          # Master Migration & Architecture Plans
│   ├── 01_EXECUTIVE_SUMMARY_AND_VISION.md
│   ├── 02_TASKOSAUR_FUNCTIONAL_ANALYSIS.md
│   ├── 03_CLEAN_ARCHITECTURE_BLUEPRINT.md
│   ├── 04_DATABASE_SCHEMA_AND_MODELS.md
│   ├── 05_API_AND_WEBSOCKET_SPECIFICATION.md
│   ├── 06_AI_ENGINE_AND_INTEGRATIONS_SPEC.md
│   ├── 07_UI_UX_DESIGN_SYSTEM_MONOCHROME.md
│   ├── 08_PHASED_IMPLEMENTATION_AND_MIGRATION_ROADMAP.md
│   └── README.md
│
├── package.json                   # Root monorepo workspace configuration
├── docker-compose.dev.yml         # Local dev stack (Postgres 16 + Redis 7 + App)
└── README.md                      # Platform overview and quick start guide
```

---

## 3. Monorepo Scripts (`package.json`)

Standard npm workspaces scripts coordinate workflows across workspaces:

| Command | Action |
| :--- | :--- |
| `npm run dev` | Runs backend (port 3000) and frontend (port 3001) concurrently. |
| `npm run dev:frontend` | Runs Next.js frontend only on port 3001. |
| `npm run dev:backend` | Runs NestJS backend only on port 3000 in watch mode. |
| `npm run build` | Builds workspaces in topological dependency order. |
| `npm run build:frontend` | Builds Next.js production bundle. |
| `npm run build:backend` | Compiles NestJS backend. |
| `npm run lint` | Lints all workspaces. |
| `npm run test` | Runs test suites across workspaces. |
| `npm run db:migrate` | Runs Prisma migrations in backend. |
| `npm run db:seed` | Populates database with Orbito default and demo data. |
| `npm run db:studio` | Launches Prisma Studio for schema inspection. |
