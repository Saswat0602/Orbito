# 01. Executive Summary & White-Label Vision: Orbito

## 1. Executive Summary

The objective of this initiative is to build and establish **Orbito** as a world-class, production-ready enterprise project and task management system through the comprehensive white-labeling and clean architectural re-engineering of the **Taskosaur** platform.

**Orbito** bridges traditional enterprise project management rigor (Kanban, Sprints, Workflows, Gantt, Time Tracking, Custom Fields, Subtasks, Dependencies) with cutting-edge **Conversational AI Task Execution**, deep real-time multi-user synchronization, and a clean architecture monorepo design powered by Turborepo.

### What is White-Labeling in this Context?
- **Brand Identity**: Complete transition from Taskosaur to **Orbito** across all source code, assets, emails, documentation, schemas, domain routes, and configurations.
- **Visual Aesthetic Overhaul**: Departing from generic dashboard styling in favor of an elite, modern production aesthetic combining:
  - **Linear**: Razor-sharp typography, ultra-high contrast dark/light themes (Obsidian Black, Crisp White, Zinc accents), keyboard-driven navigation (command palette `Cmd+K` / `Ctrl+K`), minimalist borders, and high velocity.
  - **Slack**: Intuitive Channel/Project sidebar organization, instant messaging/task comments with rich text mentions, threaded discussions, and real-time presence indicators.
  - **Monday.com**: Extremely versatile custom views (Kanban, List, Calendar, Gantt, Sprints), workflow automations, and custom field extensibility.
- **Clean Architecture Monorepo**: Structuring the application using Clean Architecture with decoupled workspaces (`backend/`, `frontend/`) and shared packages (`packages/types`, `packages/ui`, `packages/tsconfig`, `packages/eslint-config`), ensuring type safety, isolation, and clean Domain-Driven Design (DDD).

---

## 2. Product Vision & Value Propositions

| Pillar | Taskosaur Baseline | Orbito Target |
| :--- | :--- | :--- |
| **Product Branding** | Taskosaur (dinosaur branding) | **Orbito** — Sleek, modern, velocity-driven task & project management |
| **Visual Language** | Standard colorful dashboard | Monochromatic high-contrast (Obsidian Black, Crisp White, Zinc Gray) with selective status accents (Linear / Slack grade) |
| **Monorepo Engine** | Basic root npm scripts | Turborepo & standard **npm workspaces** with decoupled packages, clean build scripts, and shared package isolation |
| **Architecture** | Monorepo with NestJS backend + Next.js frontend | Clean Architecture Monorepo (`backend/`, `frontend/`, `packages/`) with Domain-Driven Design (DDD), strict DTO validation, and typed contracts |
| **Shared Contracts** | Duplicated DTOs across frontend and backend | **`packages/types`**: Shared TypeScript contracts, DTOs, Enums, and WebSocket event interfaces under `@orbito/types` |
| **Shared UI Library** | Siloed frontend UI components | **`packages/ui`**: Dedicated monochrome design system (Radix primitives, Tailwind v4, Lucide icons) under `@orbito/ui` |
| **Real-Time Collaboration** | Socket.IO room broadcasting | Socket.IO with multi-tenant presence, live cursors/locking hints, optimistic updates, and instant sync |
| **AI Task Execution** | In-app browser agent + BYOK LLM | Enhanced Conversational AI agent with tool-calling, background workflow generation, and voice-assisted prompt execution |
| **Enterprise Integrations** | Jira, Trello, IMAP email | Hardened Jira/Trello sync with SSRF protection, DNS rebinding prevention, and automatic email-to-task parsing |
| **Deployment Model** | Docker Compose + Self-hosted | Fully containerized Docker Compose (Dev/Prod), CI/CD pipelines, and health-checked micro-services |

---

## 3. High-Level Core Workflows

1. **Multi-Tenant Hierarchy Management**:
   - Organization Setup -> Workspace Creation (supports sub-workspaces/hierarchy) -> Project Allocation with custom task identifier prefixes (e.g. `ORB-101`, `ENG-42`).
2. **Task Lifecycle & Workflow Engine**:
   - Custom Statuses (Todo, In Progress, In Review, Done) categorized by state types.
   - Configurable Status Transitions enforcing state machine rules (e.g., cannot transition directly from `Todo` to `Done` without review).
   - LexoRank alphanumeric position ranking allowing millisecond drag-and-drop reordering without mass row updates.
3. **Agile Sprint & Backlog Execution**:
   - Sprint backlog planning, point estimation, active sprint tracking with velocity and burndown metrics.
4. **Conversational AI Task Management**:
   - Natural language commands: *"Create a 2-week sprint called Sprint 14, move high-priority bugs from backlog into it, and assign frontend tasks to Sarah."*
   - Browser automation agent navigates the UI and executes multi-step operations directly in-app.
5. **Automation Rules Engine**:
   - When Trigger Occurs (e.g., Task Moved to Review) -> If Conditions Match (e.g., Priority is Urgent) -> Execute Actions (e.g., Assign to Lead QA, Notify Channel).
6. **Inbound Email-to-Task Pipeline**:
   - Project-specific IMAP integration pulls emails, extracts threads, converts them into tasks, applies rule matchers, and handles customer replies.
