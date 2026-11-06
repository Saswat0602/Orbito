# Orbito Platform

> **Production-Grade Enterprise Task & Project Management with Conversational AI Task Execution**

Orbito is a high-performance project management platform engineered with **Clean Architecture** and **Domain-Driven Design (DDD)**. Designed for engineering velocity and seamless team collaboration, Orbito combines the minimalist, keyboard-first aesthetic of **Linear**, the communication power of **Slack**, and the customizable workflow flexibility of **Monday.com**.

---

## Key Features

- 🤖 **Conversational AI Task Execution**: Natural language command runner with in-app DOM browser automation (`browser-agent.ts`) that executes multi-step workflows.
- ⚡ **Real-Time Collaboration**: Sub-millisecond WebSocket event broadcasting across organization, workspace, and project rooms (`/events` namespace).
- 📊 **Multi-View Project Canvas**: Kanban boards with LexoRank drag-and-drop, interactive list views, sprint planning, burndown charts, Gantt timelines, and calendar views.
- 🔄 **Enterprise Integrations**: Jira bidirectional synchronization with strict SSRF and DNS pinning protections, Trello board imports, and IMAP inbound email-to-task pipeline.
- 🔐 **Hardened Security**: AES-256-GCM encryption for credentials, JWT access/refresh token rotation, OIDC Single Sign-On, and multi-tenant RBAC.
- 📈 **Automations & Background Queues**: Flexible trigger-action automation engine backed by BullMQ with resilient in-memory fallback.

---

## Master Architecture & Migration Plan

The complete migration strategy and architectural specification from Taskosaur to Orbito is documented in the [`plan/`](plan/README.md) directory:

1. **[01. Executive Summary & Vision](plan/01_EXECUTIVE_SUMMARY_AND_VISION.md)**
2. **[02. Taskosaur Deep Functional Audit](plan/02_TASKOSAUR_FUNCTIONAL_ANALYSIS.md)**
3. **[03. Clean Architecture & Monorepo Blueprint](plan/03_CLEAN_ARCHITECTURE_BLUEPRINT.md)**
4. **[04. Database Schema & Data Models](plan/04_DATABASE_SCHEMA_AND_MODELS.md)**
5. **[05. API & WebSocket Specifications](plan/05_API_AND_WEBSOCKET_SPECIFICATION.md)**
6. **[06. AI Engine, Integrations & Inbound Email](plan/06_AI_ENGINE_AND_INTEGRATIONS_SPEC.md)**
7. **[07. UI/UX Design System (Monochrome / Modern)](plan/07_UI_UX_DESIGN_SYSTEM_MONOCHROME.md)**
8. **[08. Phased Implementation & Migration Roadmap](plan/08_PHASED_IMPLEMENTATION_AND_MIGRATION_ROADMAP.md)**
9. **[09. Frontend Architecture: RTK Query, Modern Endpoints & Sockets](plan/09_FRONTEND_ARCHITECTURE_RTK_QUERY_AND_SOCKETS.md)**

---

## Monorepo Architecture

```
orbito-platform/ (focus-flow-task-manager/)
├── backend/                       # NestJS Clean Architecture Server (@orbito/backend - Port 3000)
│   ├── prisma/                    # Relational schema (38 models, 20 enums) & seeders
│   └── src/
│       ├── core/                  # Clean Architecture base classes (Entity, VO, Repo, UseCase)
│       ├── common/                # SSRF guards, crypto, access control, interceptors
│       ├── gateway/               # Real-time WebSockets (/events namespace)
│       └── modules/               # 41 Domain feature modules
├── frontend/                      # Next.js 16 Web Application (@orbito/frontend - Port 3001)
│   └── src/
│       ├── app/                   # App Router pages & layouts
│       ├── components/            # Feature & UI components (Linear / Monday monochrome)
│       ├── contexts/              # Auth, Project, Workspace, Real-time & AI Contexts
│       ├── hooks/                 # Custom React hooks
│       ├── lib/                   # API client, Socket.io, Browser Agent, Voice
│       └── types/                 # Shared TypeScript domain contracts
├── packages/                      # Shared Monorepo Packages
│   ├── types/                     # Shared TypeScript domain contracts (@orbito/types)
│   ├── ui/                        # Shared monochrome design primitives (@orbito/ui)
│   ├── tsconfig/                  # Shared TS configurations (@orbito/tsconfig)
│   └── eslint-config/             # Shared linting rules (@orbito/eslint-config)
├── docker/                        # Multi-stage Dockerfiles & entrypoint scripts
├── scripts/                       # Build distribution & utility scripts
├── plan/                          # Master migration & architectural specifications
├── turbo.json                     # Turborepo task pipeline configuration
├── docker-compose.dev.yml         # Development stack (PostgreSQL 16, Redis 7, App)
└── package.json                   # Root workspace coordinator (@orbito/platform)
```

---

## Quick Start

### 1. Prerequisites
- Node.js 20+ (22 recommended)
- npm 10+
- PostgreSQL 16+ & Redis 7+ (or Docker)

### 2. Environment Configuration
```bash
cp .env.example .env
```

### 3. Docker Compose (Recommended)
```bash
docker compose -f docker-compose.dev.yml up
```
This automatically boots:
- PostgreSQL on port 5435 (mapped to 5432)
- Redis on port 6379
- Orbito Backend on `http://localhost:3000`
- Orbito Frontend on `http://localhost:3001`
- Swagger API Docs on `http://localhost:3000/api/docs`

### 4. Local Development
```bash
# Install all dependencies across workspaces
npm install

# Run database migrations and seed default data
npm run db:migrate
npm run db:seed

# Start both frontend and backend concurrently
npm run dev

# Or build via Turborepo
npm run turbo:build
```

---

## License

Private & Proprietary — Orbito Platform.
