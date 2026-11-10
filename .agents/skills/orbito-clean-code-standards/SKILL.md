---
name: orbito-clean-code-standards
description: >-
  Guides clean code practices, anti-hallucination protocols, file limits, and strict type safety across the Orbito TypeScript and NestJS codebase.
---

# Orbito Clean Code & Architecture Standards

This skill enforces engineering quality, architectural purity, and anti-hallucination protocols for all contributions across `@orbito/backend`, `@orbito/frontend`, and `@orbito/types`.

---

## 1. Zero-Hallucination & Anti-Stupidity Protocol

1. **Verify Package Existence**: Never import an external npm library without verifying its declaration in the root or workspace `package.json`.
2. **Chesterton's Fence**: Before removing or refactoring existing code, establish why it was written. Never silently remove error handling, route guards, or mutex locks to silence a compile warning.
3. **Validate Interfaces First**: When integrating frontend endpoints with backend controllers, inspect both the controller return type and the frontend RTK Query response structure to guarantee contract parity.

---

## 2. Strict Type Safety (Zero `any` Policy)

1. **`unknown` Over `any`**: In cases where the payload structure is dynamic, type it as `unknown` and perform runtime narrowing with type guards (`typeof`, `Array.isArray`, or Zod).
2. **Shared Types in `@orbito/types`**: Domain models, enums (e.g. `Role`, `TaskPriority`, `NotificationType`), and shared DTOs must be maintained in `@orbito/types` and shared across frontend and backend.
3. **No Unsafe Type Assertions**: Avoid `as any` casting. Prefer type narrowing or generic parameters.

---

## 3. Structural Limits & Anti-Bloat Guardrails

| Constraint | Maximum Allowed | Recommended Practice |
| :--- | :--- | :--- |
| **File Length** | 1,200 lines (complex views) | Keep standard modules between 200–400 lines |
| **Function Length** | 75 lines | Split into pure sub-routines or extracted helpers |
| **Nesting Depth** | 3 levels | Return early using guard clauses |
| **Folder Discipline** | No generic `utils/` or `common/` | Co-locate utilities within domain feature folders |

---

## 4. Strict Layer Separation

```text
[ Browser / Client ]
       │
       ▼
[ UI Components (React 19 / Next.js) ] ── (Pure presentation, zero raw DB/fetch queries)
       │
       ▼
[ RTK Query Slices (lib/store/api) ] ──── (Network serialization, mutex reauth, caching)
       │
       ▼ (HTTP REST / Cookies)
[ NestJS Controllers ] ────────────────── (HTTP route parsing, DTO validation, Guards)
       │
       ▼
[ NestJS Domain Services ] ────────────── (Business logic, transactions, event dispatch)
       │
       ▼
[ Prisma Persistence Layer ] ──────────── (PostgreSQL relational constraints & queries)
```

- **Rule 1**: Never call Prisma or execute database queries from a controller or UI component.
- **Rule 2**: Never place React hooks or presentation styling inside backend services or shared types.
- **Rule 3**: Never duplicate validation schemas. Align frontend Zod schemas with backend `class-validator` DTOs.

---

## 5. Pre-Merge Verification Routine

Before completing any task, execute:
```bash
# 1. Typecheck & build backend
npm run build --workspace=@orbito/backend

# 2. Typecheck & build frontend
npm run build --workspace=@orbito/frontend

# 3. Check git status to ensure zero orphan or misplaced files
git status -s
```
