# Orbito Master Architecture & Migration Plan

> **White-Label Transformation of Taskosaur into Orbito: A Production-Grade, Enterprise Task & Project Management Platform with Clean Architecture**

---

## Overview

This repository is being upgraded into **Orbito**, a production-grade, enterprise-ready project management platform. Orbito replicates and enhances all core capabilities of **Taskosaur**, white-labeling the entire system under the **Orbito** brand with an elevated, high-performance UI/UX inspired by the minimalist elegance of **Linear**, the communication versatility of **Slack**, and the customizable workflow flexibility of **Monday.com**, driven by a clean architecture monorepo with Turborepo.

---

## Plan Directory Contents

| Document | Description |
| :--- | :--- |
| **[01. Executive Summary & Vision](01_EXECUTIVE_SUMMARY_AND_VISION.md)** | Product vision, brand identity, white-labeling strategy, modern monochrome design philosophy, and Clean Architecture tenets. |
| **[02. Taskosaur Deep Functional Audit](02_TASKOSAUR_FUNCTIONAL_ANALYSIS.md)** | Exhaustive analysis of Taskosaur: 41 backend modules, 38 Prisma data models, 20 enums, ~350 REST endpoints, WebSocket events, AI agents, and integrations. |
| **[03. Clean Architecture Monorepo Blueprint](03_CLEAN_ARCHITECTURE_BLUEPRINT.md)** | Monorepo topology (`backend`, `frontend`, `packages/types`, `packages/ui`), Domain-Driven Design (DDD), directory structures, and workspace configuration. |
| **[04. Database Schema & Data Models](04_DATABASE_SCHEMA_AND_MODELS.md)** | Full Prisma relational schema specification for Orbito, covering multi-tenancy, LexoRank ordering, audit logging, soft deletion, and master seeders. |
| **[05. API & WebSocket Specifications](05_API_AND_WEBSOCKET_SPECIFICATION.md)** | Complete endpoint dictionary, authentication mechanisms, request/response DTOs, and real-time Socket.io namespace/rooms protocol. |
| **[06. AI Engine, Integrations & Inbound Email](06_AI_ENGINE_AND_INTEGRATIONS_SPEC.md)** | Technical specification for Conversational AI task execution, in-app DOM browser automation agent, Jira/Trello bidirectional sync (with SSRF protection), and IMAP email inbox. |
| **[07. UI/UX Design System (Monochrome / Modern)](07_UI_UX_DESIGN_SYSTEM_MONOCHROME.md)** | Component design system, color palette (Obsidian black & crisp white / zinc monochrome with crisp semantic accents), typography, dark/light mode toggle, and micro-interactions. |
| **[08. Phased Implementation & Migration Roadmap](08_PHASED_IMPLEMENTATION_AND_MIGRATION_ROADMAP.md)** | 9-phase implementation roadmap with concrete milestones, dependency graph, Turborepo setup, testing criteria (unit, integration, E2E), and verification gates. |
| **[09. Frontend Architecture (RTK Query & Sockets)](09_FRONTEND_ARCHITECTURE_RTK_QUERY_AND_SOCKETS.md)** | Modern frontend engineering spec: Redux Toolkit Query (RTK Query), HTTP & auth interceptors with mutex reauth, Socket.io cache streaming, and distinct monochrome UX. |

---

## Architectural Principles

1. **Clean Architecture Monorepo**: Decoupled `backend/`, `frontend/`, and `packages/` workspaces with zero duplicate types via `@orbito/types` and shared design primitives via `@orbito/ui`.
2. **Separation of Concerns (Clean Architecture & DDD)**: Domain business logic remains completely isolated from external frameworks, database drivers, and UI libraries.
3. **Strict Multi-Tenancy**: Hierarchical isolation: `Organization -> Workspace (supports nesting) -> Project -> Sprint / Workflow / Task`.
4. **Real-Time by Default**: Every task creation, status transition, comment, and assignment emits WebSocket events scoped to projects and workspaces for instant multi-user synchronization.
5. **Conversational AI Task Execution**: Users can command the platform in natural language; an AI agent coordinates both backend actions and in-app DOM automation to execute multi-step workflows.
6. **Enterprise Security & Reliability**: AES-256-GCM encryption for credentials, SSRF protection with DNS pinning for external integrations, JWT access/refresh token rotation, and resilient BullMQ queues with local fallback.
