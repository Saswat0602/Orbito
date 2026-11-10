---
name: orbito-auth-and-rbac
description: >-
  Guides implementation, testing, and debugging of authentication (JWT, httpOnly cookies, OIDC/SSO), session management, and Role-Based Access Control (RBAC) in the Orbito platform across NestJS backend and Next.js frontend.
---

# Orbito Authentication & RBAC System Guide

This skill governs the dual-token cookie/header authentication architecture, OIDC/SSO integrations, and hierarchical Role-Based Access Control (RBAC) across the `@orbito/backend` and `@orbito/frontend` workspaces.

---

## 1. Authentication Architecture

Orbito employs a hardened dual-layer session strategy:
1. **Access Token (`JWT`)**: Short-lived (15 minutes). Sent in the `Authorization: Bearer <token>` header or via the `access_token` httpOnly cookie.
2. **Refresh Token (`JWT`)**: Long-lived (7 days). Stored in a secure `httpOnly` cookie (`refresh_token`) and persisted in the database session store.
3. **OIDC / Single Sign-On (`SSO`)**: Google and enterprise OpenID Connect identity provider flows handled via `OidcService` in `backend/src/modules/auth/services/oidc.service.ts`.

### Backend Components (`backend/src/modules/auth`)
- **Controller**: `auth.controller.ts` (`/api/auth/*`)
  - `POST /login`: Validates credentials, sets `access_token` and `refresh_token` cookies, returns user and tokens.
  - `POST /register`: Registers new user, optionally provisions default organization.
  - `POST /refresh`: Verifies `refresh_token` cookie and issues a fresh `access_token`.
  - `POST /logout`: Clears auth cookies and revokes session.
  - `GET /me`: Returns profile and permissions of authenticated user.
  - `GET /oidc/google` & `GET /oidc/callback`: OIDC redirect initiation and code exchange.
- **Guards**:
  - `JwtAuthGuard`: Protects private routes (`guards/jwt-auth.guard.ts`). Extracts user ID and payload.
  - `RolesGuard`: Enforces role requirements (`guards/roles.guard.ts`). Works with `@Roles(...)` decorator.
- **Strategies**:
  - `JwtStrategy`: Passport JWT strategy extracting tokens from both `Bearer` headers and request cookies (`strategies/jwt.strategy.ts`).
- **Decorators**:
  - `@CurrentUser()`: Extracts the authenticated user entity from the request (`decorators/current-user.decorator.ts`).
  - `@Roles(Role.SUPER_ADMIN, Role.OWNER)`: Declares authorized roles on controller classes or handler methods (`decorators/roles.decorator.ts`).
- **Utilities**:
  - `getAuthUser(req)`: Utility in `src/common/request.utils.ts` throwing `UnauthorizedException` if session is absent.

### Frontend Synchronization (`frontend/src`)
- **API Slice & Mutex Re-auth**:
  - `src/lib/store/api/apiSlice.ts`: Uses `async-mutex` to synchronize token refreshes across concurrent network calls.
  - When any request receives a `401 Unauthorized`, mutex locks, calls `POST /api/auth/refresh`, and transparently retries all queued requests.
  - If refresh fails, it redirects to `/login?error=session_expired`.
  - Ignores `FETCH_ERROR` (connection refused) to prevent recursive spamming when the backend is offline.
- **Providers & Hooks**:
  - `src/components/providers/AuthProvider.tsx`: Hydrates session on initial load and handles routing to `/login` when unauthenticated.
  - `src/lib/hooks/useAuthRbac.ts`: Provides active user identity, role booleans (`isSuperAdmin`, `isOwner`, `isManager`, `isMember`, `isViewer`), and helper checks.

---

## 2. Role-Based Access Control (RBAC) Matrix

Roles are strictly typed in `@orbito/types`:

| Role | Scope | Permissions & Capabilities |
| :--- | :--- | :--- |
| `SUPER_ADMIN` | Global System Root | Access to `/admin/*`, manages all tenants, queues, system configuration, rate limits, and audit logs. |
| `OWNER` | Organization / Workspace | Full control over organization settings, billing, member invites, role elevations, and destructive deletions. |
| `MANAGER` | Workspace / Project | Can create/edit projects, manage sprint lifecycles, assign tasks, invite team members, configure project workflows. |
| `MEMBER` | Project / Tasks | Can create, update, comment on, and complete assigned tasks. Cannot delete projects or modify workspace settings. |
| `VIEWER` | Read-Only | Read-only access to tasks, boards, and sprints. Cannot create, edit, drag-and-drop, or delete any entity. |

### Backend Guard Usage
```typescript
@UseGuards(JwtAuthGuard, RolesGuard)
@Roles(Role.SUPER_ADMIN, Role.OWNER)
@Post('admin/settings')
async updateSettings(@Body() dto: UpdateSettingsDto) {
  // Only accessible to SUPER_ADMIN or OWNER
}
```

### Frontend Conditional Rendering (`useAuthRbac`)
```tsx
import { useAuthRbac } from '@/lib/hooks/useAuthRbac';

export function ProjectActions() {
  const { isManager, isOwner, isViewer } = useAuthRbac();

  if (isViewer) return null; // Read-only users see no write buttons

  return (
    <>
      <Button>Create Task</Button>
      {(isManager || isOwner) && <Button variant="outline">Project Settings</Button>}
    </>
  );
}
```

---

## 3. Best Practices & Troubleshooting Runbook

1. **Never use `localStorage` for primary JWT storage**:
   Always rely on `httpOnly` secure cookies for security against XSS vulnerabilities.
2. **Always include `credentials: 'include'` on fetch/RTK queries**:
   Without this, the browser will not send auth cookies on cross-origin or port-separated requests (`http://localhost:3001` -> `http://localhost:3000`).
3. **Defensive Organization Scoping**:
   Always check that `organizationId` is a non-empty string and not `'undefined'` before querying Prisma to avoid PostgreSQL syntax errors.
4. **401 Handling on Public Routes**:
   Ensure public setup status and registration check routes (`/api/auth/setup-status`, `/api/auth/registration-status`) do NOT have `@UseGuards(JwtAuthGuard)`.

