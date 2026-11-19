# Task List: Go Twin Backend (`apps/api-go`) Implementation

## Task 1: Scaffolding, go.mod, Turbo Integration, Makefile, Air, Linter & Dockerfile

**Description:** Initialize the Go application directory structure at apps/api-go with go.mod (Go 1.23+), Makefile, .air.toml for hot-reloading, .golangci.yml for strict Clean Code linting, multi-stage Dockerfile, and package.json scripts so Turborepo can orchestrate build, dev, lint, and test tasks.

**Acceptance criteria:**
- [x] Directory `apps/api-go` exists with valid `go.mod`, `Makefile`, `.air.toml`, `.golangci.yml`, and `package.json`
- [x] Root `turbo.json` recognizes `apps/api-go` and runs tasks (`turbo run build --filter=@orbito/api-go`)
- [x] Docker build starts and compiles minimal main.go binary

**Verification:**
- [x] Tests pass: `go test ./... inside apps/api-go`
- [x] Build succeeds: `turbo run build --filter=@orbito/api-go`
- [x] Manual check: Run `make build` and verify binary is generated at `apps/api-go/bin/api`

**Dependencies:** None

**Files likely touched:**
- `apps/api-go/go.mod`
- `apps/api-go/Makefile`
- `apps/api-go/.air.toml`
- `apps/api-go/.golangci.yml`
- `apps/api-go/package.json`
- `apps/api-go/deploy/Dockerfile`
- `turbo.json`

**Estimated scope:** Medium: 3-5 files

---

## Task 2: Typed Configuration & Environment Validation (internal/config)

**Description:** Implement strongly typed configuration loading in internal/config using caarlos0/env/v11 with validation tags. Supports all variables from .env.example (DATABASE_URL, JWT_SECRET, ENCRYPTION_KEY, REDIS_HOST, CORS origins, ports) with fail-fast validation on missing required secrets.

**Acceptance criteria:**
- [x] Typed `Config` struct parses environment variables with exact default fallbacks
- [x] Application panics or exits cleanly with clear error message if critical secrets (DATABASE_URL, JWT_SECRET) are missing
- [x] Unit tests verify environment variable overrides and default values

**Verification:**
- [x] Tests pass: `go test -v ./internal/config/...`
- [x] Build succeeds: `go build ./internal/config/...`
- [x] Manual check: Verify config loads with default values when run without environment flags

**Dependencies:** Task 1

**Files likely touched:**
- `apps/api-go/internal/config/config.go`
- `apps/api-go/internal/config/config_test.go`
- `apps/api-go/.env.example`

**Estimated scope:** Small: 1-2 files

---

## Task 3: Platform Logger (slog), Request ID & Ingress Middleware

**Description:** Build the structured logging system with slog and essential HTTP ingress middleware: Request ID / correlation ID extraction and injection, panic recovery, CORS handler mirroring NestJS whitelist, Helmet-equivalent security headers, and a 10MB body reader limiter.

**Acceptance criteria:**
- [x] `slog` outputs JSON formatted logs containing `request_id`, `correlation_id`, `method`, `path`, and latency
- [x] Incoming `x-correlation-id` or `x-request-id` is preserved and returned in response headers; otherwise UUID is generated
- [x] Panic recovery middleware catches runtime panics, logs stack trace, and returns 500 error envelope

**Verification:**
- [x] Tests pass: `go test -v ./internal/middleware/... ./internal/platform/logger/...`
- [x] Build succeeds: `go build ./internal/middleware/...`
- [x] Manual check: Send curl request with `-H 'x-correlation-id: test-123'` and check header in response

**Dependencies:** Task 2

**Files likely touched:**
- `apps/api-go/internal/platform/logger/logger.go`
- `apps/api-go/internal/middleware/request_id.go`
- `apps/api-go/internal/middleware/logger.go`
- `apps/api-go/internal/middleware/recover.go`
- `apps/api-go/internal/middleware/cors.go`
- `apps/api-go/internal/middleware/secure_headers.go`
- `apps/api-go/internal/middleware/body_limit.go`

**Estimated scope:** Large: 5-8 files

---

## Task 4: Platform Database (pgx/v5) & sqlc Setup

**Description:** Configure the PostgreSQL connection pool using pgxpool with MaxConns=50, MinConns=10, connection health checks, and a transaction coordinator helper (WithTx). Configure db/sqlc.yaml pointing to backend/prisma/migrations/20261106114254_init_the_orbit/migration.sql as the schema source.

**Acceptance criteria:**
- [x] `platform/database` connects to Postgres 16 using `DATABASE_URL` with connection pooling
- [x] `sqlc.yaml` generates type-safe Go structs and queries without errors against Prisma migration DDL
- [x] `WithTx` provides atomic transaction execution with deferred rollback on error

**Verification:**
- [x] Tests pass: `go test -v ./internal/platform/database/...`
- [x] Build succeeds: `sqlc generate -f db/sqlc.yaml && go build ./internal/platform/database/...`
- [x] Manual check: Execute database ping check against local PostgreSQL instance

**Dependencies:** Task 2

**Files likely touched:**
- `apps/api-go/internal/platform/database/database.go`
- `apps/api-go/internal/platform/database/tx.go`
- `apps/api-go/db/sqlc.yaml`

**Estimated scope:** Medium: 3-5 files

---

## Task 5: Platform Crypto (AES-256-GCM Cookies, Scrypt Fields, Bcrypt) & Parity Tests

**Description:** Implement cryptographic utilities in internal/platform/crypto: AES-256-GCM cookie encryption with SHA-256 key derivation matching cookie-crypto.utils.ts, scrypt field encryption matching encryption.util.ts, and bcrypt cost 12 password hashing. Cross-verify with test vectors from Node.js.

**Acceptance criteria:**
- [x] Encrypted cookie strings from NestJS (`ivHex:tagHex:cipherHex`) successfully decrypt in Go and vice-versa
- [x] Scrypt field encryption outputs match format and decrypt successfully with identical keys
- [x] Bcrypt cost 12 hashes generated by NestJS verify successfully in Go

**Verification:**
- [x] Tests pass: `go test -v ./internal/platform/crypto/...`
- [x] Build succeeds: `go build ./internal/platform/crypto/...`
- [x] Manual check: Run cross-language test vectors comparing Node.js crypto output to Go crypto output

**Dependencies:** Task 1

**Files likely touched:**
- `apps/api-go/internal/platform/crypto/cookie_crypto.go`
- `apps/api-go/internal/platform/crypto/field_crypto.go`
- `apps/api-go/internal/platform/crypto/password.go`
- `apps/api-go/internal/platform/crypto/crypto_test.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 6: Shared HTTP & Error Framework (httpx, apperror, pagination)

**Description:** Implement the shared HTTP layer in internal/shared: standard JSON responder, request binding with validator/v10, path parameter helpers, apperror domain errors with mapping to exact NestJS GlobalExceptionFilter JSON shape, and the 5-field pagination envelope.

**Acceptance criteria:**
- [x] `httpx.RespondError` returns `{ statusCode, error, message, path, method, timestamp, correlationId }`
- [x] Validation errors return 400 with `message` as array of error strings
- [x] `pagination.Envelope` formats `{ [entity]: [...], pagination: { currentPage, totalPages, totalCount, hasNextPage, hasPrevPage } }`

**Verification:**
- [x] Tests pass: `go test -v ./internal/shared/...`
- [x] Build succeeds: `go build ./internal/shared/...`
- [x] Manual check: Verify JSON serialization of date formats produces exact ISO 8601 with milliseconds (`.000Z`)

**Dependencies:** Task 3

**Files likely touched:**
- `apps/api-go/internal/shared/apperror/errors.go`
- `apps/api-go/internal/shared/apperror/mapper.go`
- `apps/api-go/internal/shared/httpx/bind.go`
- `apps/api-go/internal/shared/httpx/respond.go`
- `apps/api-go/internal/shared/httpx/params.go`
- `apps/api-go/internal/shared/pagination/pagination.go`
- `apps/api-go/internal/shared/ctxhelper/context.go`

**Estimated scope:** Large: 5-8 files

---

## Checkpoint: After Task 6 — Foundations Complete

- [x] Platform database connects to Postgres, executes sample query
- [x] AES-256-GCM cookie encryption & decryption test vectors match Node.js byte-for-byte
- [x] Ingress middleware assigns correlation ID and logs structured JSON with slog
- [x] Shared apperror maps domain errors to exact NestJS error JSON schema
- [x] Review with human before proceeding

---

## Task 7: Health Module (GET /api/health, readiness check)

**Description:** Implement the health check module in internal/modules/health providing GET /api/health and readiness status including database ping and redis status checks.

**Acceptance criteria:**
- [x] `GET /api/health` returns status: 'ok' with uptime, timestamp, and database connectivity state
- [x] Returns 200 when database is responsive, 503 when database is unreachable
- [x] Publicly accessible without JWT authentication

**Verification:**
- [x] Tests pass: `go test -v ./internal/modules/health/...`
- [x] Build succeeds: `go build ./internal/modules/health/...`
- [x] Manual check: curl http://localhost:3002/api/health and assert 200 OK JSON response

**Dependencies:** Task 4, Task 6

**Files likely touched:**
- `apps/api-go/internal/modules/health/handler.go`
- `apps/api-go/internal/modules/health/routes.go`
- `apps/api-go/internal/modules/health/health_test.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 8: Settings Module (/api/settings key-value store & registration check)

**Description:** Implement the system settings module managing key-value configuration rows in the Settings table, including public registration status endpoint.

**Acceptance criteria:**
- [x] `GET /api/settings` and `PATCH /api/settings` handle settings key-value entries
- [x] Registration status check is accessible by auth registration flow
- [x] Encrypted keys are properly decrypted when retrieved

**Verification:**
- [x] Tests pass: `go test -v ./internal/modules/settings/...`
- [x] Build succeeds: `go build ./internal/modules/settings/...`
- [x] Manual check: Query settings endpoint and verify response matches NestJS settings controller

**Dependencies:** Task 4, Task 6

**Files likely touched:**
- `apps/api-go/internal/modules/settings/handler.go`
- `apps/api-go/internal/modules/settings/service.go`
- `apps/api-go/internal/modules/settings/repository.go`
- `apps/api-go/internal/modules/settings/routes.go`
- `apps/api-go/internal/modules/settings/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 9: Mailer Platform & Email Module (/api/email/send, stats, SMTP client)

**Description:** Implement SMTP mailer in platform/mailer with HTML email templates and the Email module providing POST /api/email/send, POST /api/email/send-bulk, and GET /api/email/queue/stats.

**Acceptance criteria:**
- [x] SMTP client connects using SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS with STARTTLS/TLS
- [x] Email sending endpoints validate payload and dispatch email or enqueue for background delivery
- [x] Queue stats return active, pending, and failed email counts

**Verification:**
- [x] Tests pass: `go test -v ./internal/platform/mailer/... ./internal/modules/email/...`
- [x] Build succeeds: `go build ./internal/modules/email/...`
- [x] Manual check: Trigger test email send and verify SMTP connection handshake

**Dependencies:** Task 6

**Files likely touched:**
- `apps/api-go/internal/platform/mailer/mailer.go`
- `apps/api-go/internal/modules/email/handler.go`
- `apps/api-go/internal/modules/email/service.go`
- `apps/api-go/internal/modules/email/routes.go`
- `apps/api-go/internal/modules/email/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 10: Auth Core & JWT Strategy (claims, token minting, RequireAuth middleware)

**Description:** Build the JWT token service in internal/modules/auth and the RequireAuth middleware. Mints access tokens (HS256, 15m) and refresh tokens (7d), extracts from Authorization header or access_token cookie, validates claims, asserts account status is ACTIVE, and injects user identity into context.

**Acceptance criteria:**
- [x] JWT verification handles Bearer token header and `access_token` cookie fallback
- [x] Asserts account is ACTIVE and not deleted; returns 401 'Account is not active' if deactivated
- [x] System user ID (00000000-0000-0000-0000-000000000000) is barred from authenticating

**Verification:**
- [x] Tests pass: `go test -v ./internal/modules/auth/jwt_test.go`
- [x] Build succeeds: `go build ./internal/modules/auth/...`
- [x] Manual check: Verify token generated by NestJS is accepted by Go RequireAuth middleware

**Dependencies:** Task 4, Task 5, Task 6

**Files likely touched:**
- `apps/api-go/internal/modules/auth/jwt.go`
- `apps/api-go/internal/middleware/auth.go`
- `apps/api-go/internal/modules/auth/account_active.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 11: Auth Module: Registration & Login (/register, /login, cookies, /me, /profile)

**Description:** Implement core authentication endpoints: POST /api/auth/register (with auto-assigned unique username and default org join), POST /api/auth/login (with bootstrap super admin creation if first user), httpOnly encrypted cookie setting, GET /api/auth/me, and GET /api/auth/profile.

**Acceptance criteria:**
- [ ] POST /api/auth/login validates bcrypt password, updates lastLoginAt, returns AuthResponseDto and sets cookies
- [ ] First account to register on fresh instance automatically receives SUPER_ADMIN role
- [ ] GET /api/auth/me returns authenticated user profile without password field

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/auth/login_test.go`
- [ ] Build succeeds: `go build ./internal/modules/auth/...`
- [ ] Manual check: Execute login via curl, verify set-cookie headers contain access_token and encrypted refresh_token

**Dependencies:** Task 10

**Files likely touched:**
- `apps/api-go/internal/modules/auth/handler_login.go`
- `apps/api-go/internal/modules/auth/service_login.go`
- `apps/api-go/internal/modules/auth/repository.go`
- `apps/api-go/internal/modules/auth/dto.go`
- `apps/api-go/internal/modules/auth/routes.go`

**Estimated scope:** Large: 5-8 files

---

## Task 12: Auth Module: Token Refresh & Logout (/refresh, /logout, user status check)

**Description:** Implement POST /api/auth/refresh and POST /api/auth/logout. Refresh decrypts cookie or accepts body token, verifies JWT signature, checks DB refresh_token column, asserts account active, rotates refresh token, and sets new cookies. Logout clears DB column and clears cookies.

**Acceptance criteria:**
- [ ] POST /api/auth/refresh validates refresh token from encrypted cookie or request body
- [ ] Rotates refresh token in database and updates httpOnly cookies
- [ ] POST /api/auth/logout sets user refresh_token to null and clears client cookies

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/auth/refresh_test.go`
- [ ] Build succeeds: `go build ./internal/modules/auth/...`
- [ ] Manual check: Call refresh endpoint with expired access token and valid refresh cookie, verifying new tokens are returned

**Dependencies:** Task 11

**Files likely touched:**
- `apps/api-go/internal/modules/auth/handler_refresh.go`
- `apps/api-go/internal/modules/auth/service_refresh.go`

**Estimated scope:** Small: 1-2 files

---

## Task 13: Auth Module: Password Reset & Setup (/forgot-password, /reset-password, /setup)

**Description:** Implement password recovery and initial instance setup: POST /api/auth/forgot-password (generates 32-byte hex token, hashes with bcrypt cost 10, sets 24h expiry), GET /api/auth/verify-reset-token/:token, POST /api/auth/reset-password, GET /api/auth/setup/required, and POST /api/auth/setup.

**Acceptance criteria:**
- [ ] Forgot password creates hashed reset token in DB and sends recovery email link
- [ ] Reset password validates token, updates password with bcrypt cost 12, and clears reset token
- [ ] Setup endpoints allow one-time super admin initialization when no admin exists

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/auth/reset_test.go`
- [ ] Build succeeds: `go build ./internal/modules/auth/...`
- [ ] Manual check: Trigger forgot-password flow and test verify-reset-token with generated token

**Dependencies:** Task 11

**Files likely touched:**
- `apps/api-go/internal/modules/auth/handler_reset.go`
- `apps/api-go/internal/modules/auth/service_reset.go`
- `apps/api-go/internal/modules/auth/handler_setup.go`
- `apps/api-go/internal/modules/auth/service_setup.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 14: Auth SSO / OIDC Services & Endpoints (/oidc/config, /login, /callback, /exchange)

**Description:** Implement OpenID Connect (SSO) authentication in internal/modules/auth: GET /api/auth/oidc/config, GET /api/auth/oidc/login (generates state/nonce in cookie), GET /api/auth/oidc/callback (validates code against IdP, provisions or links user), and POST /api/auth/oidc/exchange.

**Acceptance criteria:**
- [ ] OIDC discovery reads issuer configuration and redirects to provider login
- [ ] Callback validates authorization code and state cookie, creating or updating user
- [ ] Issues standard session JWT and refresh token cookies on successful SSO

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/auth/oidc_test.go`
- [ ] Build succeeds: `go build ./internal/modules/auth/...`
- [ ] Manual check: Query /api/auth/oidc/config and verify enabled flag and provider name

**Dependencies:** Task 11

**Files likely touched:**
- `apps/api-go/internal/modules/auth/oidc_service.go`
- `apps/api-go/internal/modules/auth/handler_oidc.go`

**Estimated scope:** Small: 1-2 files

---

## Task 15: Users Module (/api/users CRUD, /status, /change-password, profile avatar)

**Description:** Implement the Users module in internal/modules/users providing user directory listing, get by id, update profile, update status, change password (with old password verification), and avatar upload integration.

**Acceptance criteria:**
- [ ] `GET /api/users` and `GET /api/users/:id` return user records omitting password hash
- [ ] `PATCH /api/users/:id/change-password` verifies current password before updating to new bcrypt hash
- [ ] `PATCH /api/users/:id/status` updates ACTIVE/INACTIVE user state with admin check

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/users/...`
- [ ] Build succeeds: `go build ./internal/modules/users/...`
- [ ] Manual check: Update user bio and verify updated_at timestamp updates in database

**Dependencies:** Task 11

**Files likely touched:**
- `apps/api-go/internal/modules/users/handler.go`
- `apps/api-go/internal/modules/users/service.go`
- `apps/api-go/internal/modules/users/repository.go`
- `apps/api-go/internal/modules/users/routes.go`
- `apps/api-go/internal/modules/users/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Task 16: Admin Module (/api/admin/dashboard, /users, /organizations, /config, /test-smtp)

**Description:** Implement system administration endpoints in internal/modules/admin protected by SUPER_ADMIN role: dashboard metrics, system-wide user role/status management, organization oversight, and SMTP connection testing.

**Acceptance criteria:**
- [ ] All admin endpoints strictly require `SUPER_ADMIN` role; non-admins receive 403 Forbidden
- [ ] `GET /api/admin/dashboard` aggregates counts of users, organizations, workspaces, and tasks
- [ ] `POST /api/admin/config/test-smtp` tests live outbound mail connection

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/admin/...`
- [ ] Build succeeds: `go build ./internal/modules/admin/...`
- [ ] Manual check: Test admin dashboard with non-admin token and verify 403 Forbidden response

**Dependencies:** Task 15

**Files likely touched:**
- `apps/api-go/internal/modules/admin/handler.go`
- `apps/api-go/internal/modules/admin/service.go`
- `apps/api-go/internal/modules/admin/repository.go`
- `apps/api-go/internal/modules/admin/routes.go`
- `apps/api-go/internal/modules/admin/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Checkpoint: After Task 16 — Core System & Auth Verified

- [ ] User registration, login, token refresh, and logout work end-to-end
- [ ] Bcrypt cost 12 passwords generated by NestJS authenticate seamlessly in Go
- [ ] Encrypted httpOnly refresh token cookies are parsed and set correctly
- [ ] System settings, email sending, user profiles, and super admin console fully operational
- [ ] Review with human before proceeding

---

## Task 17: RBAC Core Middleware (Hierarchy Rank & Scope Resolution)

**Description:** Implement the role hierarchy and scope-based access control middleware in internal/middleware/rbac.go. Compares role ranks (VIEWER < MEMBER < MANAGER < OWNER < SUPER_ADMIN), inspects path/query/body parameters for scope IDs, queries membership tables, handles public project read-only access, and supports internal visibility inheritance.

**Acceptance criteria:**
- [ ] `SUPER_ADMIN` bypasses all scope checks
- [ ] Evaluates member role in target scope (Organization, Workspace, Project) against required minimum role
- [ ] Public project visibility allows read-only (GET/HEAD) for VIEWER-level requirements; denies writes

**Verification:**
- [ ] Tests pass: `go test -v ./internal/middleware/rbac_test.go`
- [ ] Build succeeds: `go build ./internal/middleware/...`
- [ ] Manual check: Verify user with MEMBER role in workspace is denied access to OWNER-only workspace endpoint

**Dependencies:** Task 10

**Files likely touched:**
- `apps/api-go/internal/middleware/rbac.go`
- `apps/api-go/internal/middleware/rbac_test.go`

**Estimated scope:** Small: 1-2 files

---

## Task 18: Organizations Module (/api/organizations CRUD, transfer ownership, default org)

**Description:** Implement Organizations module in internal/modules/organizations providing multi-tenant organization creation, slug resolution, update, deletion, transfer ownership to another member, and user default organization association.

**Acceptance criteria:**
- [ ] `POST /api/organizations` creates organization, assigns creator as OWNER in organization_members, and generates default workflow
- [ ] `PATCH /api/organizations/:id/transfer-ownership` reassigns OWNER role atomically within a transaction
- [ ] Enforces unique organization slugs with slug generation fallback

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/organizations/...`
- [ ] Build succeeds: `go build ./internal/modules/organizations/...`
- [ ] Manual check: Create an organization and verify creator has OWNER entry in organization_members table

**Dependencies:** Task 17

**Files likely touched:**
- `apps/api-go/internal/modules/organizations/handler.go`
- `apps/api-go/internal/modules/organizations/service.go`
- `apps/api-go/internal/modules/organizations/repository.go`
- `apps/api-go/internal/modules/organizations/routes.go`
- `apps/api-go/internal/modules/organizations/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Task 19: Organization Members Module (/api/organization-members list, update role, remove)

**Description:** Implement organization membership management in internal/modules/organization_members: listing members, updating member roles (with hierarchy check preventing modifying users of equal/higher rank), and removing members.

**Acceptance criteria:**
- [ ] `GET /api/organization-members/organization/:id` returns members with user details and roles
- [ ] `PATCH /api/organization-members/:id` updates member role with OWNER/MANAGER permission check
- [ ] Prevents removing the sole OWNER of an organization

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/organization_members/...`
- [ ] Build succeeds: `go build ./internal/modules/organization_members/...`
- [ ] Manual check: Update organization member role and verify updated role in response

**Dependencies:** Task 18

**Files likely touched:**
- `apps/api-go/internal/modules/organization_members/handler.go`
- `apps/api-go/internal/modules/organization_members/service.go`
- `apps/api-go/internal/modules/organization_members/repository.go`
- `apps/api-go/internal/modules/organization_members/routes.go`
- `apps/api-go/internal/modules/organization_members/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 20: Workspaces Module (/api/workspaces CRUD, /charts metrics, archive)

**Description:** Implement the Workspaces module in internal/modules/workspaces: workspace CRUD, slug lookup, archive/unarchive, and aggregated charts/metrics endpoint calculating task completion rate, workload health, and status distribution.

**Acceptance criteria:**
- [ ] `POST /api/workspaces` creates workspace within target organization and auto-adds creator as OWNER
- [ ] `GET /api/workspaces/:id/charts` computes aggregated task statistics across all projects in the workspace
- [ ] Archived workspaces are excluded from active directory queries

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/workspaces/...`
- [ ] Build succeeds: `go build ./internal/modules/workspaces/...`
- [ ] Manual check: Fetch workspace charts endpoint and verify task percentage calculation

**Dependencies:** Task 18

**Files likely touched:**
- `apps/api-go/internal/modules/workspaces/handler.go`
- `apps/api-go/internal/modules/workspaces/service.go`
- `apps/api-go/internal/modules/workspaces/charts_service.go`
- `apps/api-go/internal/modules/workspaces/repository.go`
- `apps/api-go/internal/modules/workspaces/routes.go`
- `apps/api-go/internal/modules/workspaces/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Task 21: Workspace Members Module (/api/workspace-members list, role update, remove)

**Description:** Implement workspace membership operations in internal/modules/workspace_members: listing members by workspace, adding members, modifying roles, and removing workspace members.

**Acceptance criteria:**
- [ ] `GET /api/workspace-members/workspace/:id` returns members list with role and joinedAt timestamps
- [ ] `POST /api/workspace-members` adds user to workspace with specified role
- [ ] `DELETE /api/workspace-members/:id` removes user from workspace

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/workspace_members/...`
- [ ] Build succeeds: `go build ./internal/modules/workspace_members/...`
- [ ] Manual check: Add member to workspace and verify membership in response

**Dependencies:** Task 20

**Files likely touched:**
- `apps/api-go/internal/modules/workspace_members/handler.go`
- `apps/api-go/internal/modules/workspace_members/service.go`
- `apps/api-go/internal/modules/workspace_members/repository.go`
- `apps/api-go/internal/modules/workspace_members/routes.go`
- `apps/api-go/internal/modules/workspace_members/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 22: Invitations Module (/api/invitations create token, verify, accept, revoke, resend)

**Description:** Implement member invitation system in internal/modules/invitations: POST /api/invitations (creates unique token, 7d expiry, dispatches email), GET /api/invitations/token/:token (validates token), POST /api/invitations/accept, revoke, and resend endpoints.

**Acceptance criteria:**
- [ ] Invitation creates pending record in DB and dispatches invite email with accept link
- [ ] Accepting invitation creates workspace/org member record and marks invitation ACCEPTED
- [ ] Expired or revoked tokens are rejected with appropriate 400 Bad Request message

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/invitations/...`
- [ ] Build succeeds: `go build ./internal/modules/invitations/...`
- [ ] Manual check: Create invitation, verify token via API, and accept token with test user

**Dependencies:** Task 20

**Files likely touched:**
- `apps/api-go/internal/modules/invitations/handler.go`
- `apps/api-go/internal/modules/invitations/service.go`
- `apps/api-go/internal/modules/invitations/repository.go`
- `apps/api-go/internal/modules/invitations/routes.go`
- `apps/api-go/internal/modules/invitations/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Checkpoint: After Task 22 — Multi-Tenancy & RBAC Verified

- [ ] Role hierarchy rank checks strictly enforced across VIEWER, MEMBER, MANAGER, OWNER, SUPER_ADMIN
- [ ] Organization and Workspace multi-tenancy boundaries isolated
- [ ] Member invitations create signed tokens, send emails, and permit acceptance
- [ ] Review with human before proceeding

---

## Task 23: Projects Module (/api/projects CRUD, slug/key, visibility, custom task prefixes)

**Description:** Implement the Projects module in internal/modules/projects: project creation (auto-generating 3-4 letter uppercase key if not provided), slug generation, project listing with pagination, visibility controls (PUBLIC, INTERNAL, PRIVATE), and project stats.

**Acceptance criteria:**
- [ ] `POST /api/projects` creates project with custom task key prefix (e.g. 'ORB') and default workflow
- [ ] Project listing filters by workspace, status, search, and visibility permissions
- [ ] `GET /api/projects/:id/stats` returns total tasks, completed tasks, and member count

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/projects/...`
- [ ] Build succeeds: `go build ./internal/modules/projects/...`
- [ ] Manual check: Create a project and verify custom key prefix and default workflow assignment

**Dependencies:** Task 20

**Files likely touched:**
- `apps/api-go/internal/modules/projects/handler.go`
- `apps/api-go/internal/modules/projects/service.go`
- `apps/api-go/internal/modules/projects/repository.go`
- `apps/api-go/internal/modules/projects/routes.go`
- `apps/api-go/internal/modules/projects/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Task 24: Project Members Module (/api/project-members list, add, role update, remove)

**Description:** Implement project membership management in internal/modules/project_members: listing members by project, adding members, modifying roles, and removing members.

**Acceptance criteria:**
- [ ] `GET /api/project-members/project/:id` returns project members with user profile details
- [ ] `POST /api/project-members` adds member with designated role
- [ ] `DELETE /api/project-members/:id` removes member from project

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/project_members/...`
- [ ] Build succeeds: `go build ./internal/modules/project_members/...`
- [ ] Manual check: Assign member to project and verify role permissions apply

**Dependencies:** Task 23

**Files likely touched:**
- `apps/api-go/internal/modules/project_members/handler.go`
- `apps/api-go/internal/modules/project_members/service.go`
- `apps/api-go/internal/modules/project_members/repository.go`
- `apps/api-go/internal/modules/project_members/routes.go`
- `apps/api-go/internal/modules/project_members/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 25: Workflows Module (/api/workflows CRUD, default workflow generation)

**Description:** Implement the Workflows module in internal/modules/workflows: workflow CRUD, assigning workflows to projects, and auto-generating standard 4-state workflow (To Do, In Progress, In Review, Done) for new organizations.

**Acceptance criteria:**
- [ ] `POST /api/workflows` creates custom workflow with initial status nodes
- [ ] `GET /api/workflows/:id` returns workflow including its task statuses and transitions
- [ ] Cascade deletes or reassigns task statuses when workflow is deleted

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/workflows/...`
- [ ] Build succeeds: `go build ./internal/modules/workflows/...`
- [ ] Manual check: Create a workflow and verify status nodes are linked in database

**Dependencies:** Task 23

**Files likely touched:**
- `apps/api-go/internal/modules/workflows/handler.go`
- `apps/api-go/internal/modules/workflows/service.go`
- `apps/api-go/internal/modules/workflows/repository.go`
- `apps/api-go/internal/modules/workflows/routes.go`
- `apps/api-go/internal/modules/workflows/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 26: Task Statuses Module (/api/task-statuses custom columns, categories, transitions)

**Description:** Implement task status columns and status transition rules in internal/modules/task_statuses: status column CRUD, color/category assignment (TODO, IN_PROGRESS, DONE), position ordering, and transition permission checks.

**Acceptance criteria:**
- [ ] `GET /api/task-statuses/workflow/:workflowId` returns statuses ordered by position
- [ ] Prevents deleting default status if tasks are currently in that status
- [ ] Transition validation confirms whether task can move from Status A to Status B

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/task_statuses/...`
- [ ] Build succeeds: `go build ./internal/modules/task_statuses/...`
- [ ] Manual check: Create custom status column and reorder positions

**Dependencies:** Task 25

**Files likely touched:**
- `apps/api-go/internal/modules/task_statuses/handler.go`
- `apps/api-go/internal/modules/task_statuses/service.go`
- `apps/api-go/internal/modules/task_statuses/repository.go`
- `apps/api-go/internal/modules/task_statuses/routes.go`
- `apps/api-go/internal/modules/task_statuses/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 27: Labels Module (/api/labels global and workspace labels CRUD)

**Description:** Implement label taxonomy in internal/modules/labels: creating, updating, listing, and deleting labels at global and workspace scopes with name, color, and description.

**Acceptance criteria:**
- [ ] `GET /api/labels` returns global and workspace-scoped labels
- [ ] `POST /api/labels` enforces unique label names within scope
- [ ] `DELETE /api/labels/:id` removes label and unlinks from associated tasks

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/labels/...`
- [ ] Build succeeds: `go build ./internal/modules/labels/...`
- [ ] Manual check: Create a label with color hex and verify listing

**Dependencies:** Task 20

**Files likely touched:**
- `apps/api-go/internal/modules/labels/handler.go`
- `apps/api-go/internal/modules/labels/service.go`
- `apps/api-go/internal/modules/labels/repository.go`
- `apps/api-go/internal/modules/labels/routes.go`
- `apps/api-go/internal/modules/labels/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 28: Sprints Module (/api/sprints planning, active, complete, burndown metrics)

**Description:** Implement Agile sprint management in internal/modules/sprints: sprint CRUD, starting a sprint (enforcing single active sprint per project), completing a sprint (with rollover of unfinished tasks), and calculating burndown chart metrics.

**Acceptance criteria:**
- [ ] `POST /api/sprints/:id/start` activates sprint and validates start/end dates
- [ ] `POST /api/sprints/:id/complete` moves incomplete tasks to next sprint or backlog
- [ ] `GET /api/sprints/:id/burndown` computes ideal vs actual story point burn rate

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/sprints/...`
- [ ] Build succeeds: `go build ./internal/modules/sprints/...`
- [ ] Manual check: Start a sprint, complete it, and verify unfinished tasks rollover to backlog

**Dependencies:** Task 23

**Files likely touched:**
- `apps/api-go/internal/modules/sprints/handler.go`
- `apps/api-go/internal/modules/sprints/service.go`
- `apps/api-go/internal/modules/sprints/repository.go`
- `apps/api-go/internal/modules/sprints/routes.go`
- `apps/api-go/internal/modules/sprints/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Checkpoint: After Task 28 — Project Domain & Agile Foundations Verified

- [ ] Project CRUD, custom task prefixes, and visibility rules functional
- [ ] Workflows and custom task status columns enforce valid status transitions
- [ ] Labels and sprints operate with burndown metrics calculation
- [ ] Review with human before proceeding

---

## Task 29: Task Ranks Module & Lexorank Algorithm (internal/modules/task_ranks)

**Description:** Implement the string-based Lexorank calculation engine and endpoints in internal/modules/task_ranks. Calculates lexicographical midpoints between previous and next rank strings to order tasks with zero rebalancing overhead.

**Acceptance criteria:**
- [ ] Lexorank midpoint calculation produces identical string ranks as NestJS `task-ranks.service.ts`
- [ ] `GET /api/task-ranks/project/:projectId` and `PATCH /api/task-ranks/:taskId` persist rank positions
- [ ] Unit tests verify edge cases: first position, last position, collision avoidance

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/task_ranks/...`
- [ ] Build succeeds: `go build ./internal/modules/task_ranks/...`
- [ ] Manual check: Compute midpoint between rank strings and assert sorted ordering

**Dependencies:** Task 23

**Files likely touched:**
- `apps/api-go/internal/modules/task_ranks/lexorank.go`
- `apps/api-go/internal/modules/task_ranks/service.go`
- `apps/api-go/internal/modules/task_ranks/handler.go`
- `apps/api-go/internal/modules/task_ranks/routes.go`
- `apps/api-go/internal/modules/task_ranks/lexorank_test.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 30: Tasks Core CRUD & Slugs (/api/tasks create, get, update, delete, archive)

**Description:** Implement core task lifecycle in internal/modules/tasks: task creation with auto-incremented project key (e.g. 'ORB-142'), slug generation, get by UUID, get by slug, get by key, update, hard delete, and archive.

**Acceptance criteria:**
- [ ] `POST /api/tasks` assigns next project key atomically and calculates initial Lexorank
- [ ] `GET /api/tasks/:id`, `GET /api/tasks/slug/:slug`, and `GET /api/tasks/key/:key` return full task details with relations
- [ ] `DELETE /api/tasks/:id` deletes task with cascading relations in PostgreSQL

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/tasks/...`
- [ ] Build succeeds: `go build ./internal/modules/tasks/...`
- [ ] Manual check: Create task, retrieve by key (e.g. 'ORB-1'), and verify formatted output

**Dependencies:** Task 29

**Files likely touched:**
- `apps/api-go/internal/modules/tasks/handler_crud.go`
- `apps/api-go/internal/modules/tasks/service_crud.go`
- `apps/api-go/internal/modules/tasks/repository.go`
- `apps/api-go/internal/modules/tasks/routes.go`
- `apps/api-go/internal/modules/tasks/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Task 31: Tasks Listing, Filtering & Grouping (/grouped, /by-status, /today, /all-tasks)

**Description:** Implement task query views in internal/modules/tasks: GET /api/tasks (with sanitized search/filter/pagination), GET /api/tasks/grouped (group by status, priority, assignee, sprint), GET /api/tasks/by-status, GET /api/tasks/today, and GET /api/tasks/all-tasks.

**Acceptance criteria:**
- [ ] List endpoints return standard 5-field pagination envelope `{ tasks, pagination }`
- [ ] Grouped endpoint aggregates tasks into buckets with counts without N+1 queries
- [ ] `GET /api/tasks/today` returns tasks assigned to user due today or overdue

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/tasks/list_test.go`
- [ ] Build succeeds: `go build ./internal/modules/tasks/...`
- [ ] Manual check: Request grouped tasks by status and verify response schema

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/tasks/handler_list.go`
- `apps/api-go/internal/modules/tasks/service_list.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 32: Tasks Bulk Operations (/bulk-create, /bulk-delete, /bulk-status-update, /bulk-assign)

**Description:** Implement high-throughput batch task operations in internal/modules/tasks: POST /api/tasks/bulk-create (batch insert with sequential keys), POST /api/tasks/bulk-delete, POST /api/tasks/bulk-status-update, and POST /api/tasks/bulk-assign within database transactions.

**Acceptance criteria:**
- [ ] Bulk operations execute inside `WithTx` transaction to prevent partial writes
- [ ] `POST /api/tasks/bulk-status-update` updates multiple tasks and validates allowed status transitions
- [ ] `POST /api/tasks/bulk-delete` deletes specified IDs and unlinks references

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/tasks/bulk_test.go`
- [ ] Build succeeds: `go build ./internal/modules/tasks/...`
- [ ] Manual check: Execute bulk status update on 3 tasks and verify status in DB

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/tasks/handler_bulk.go`
- `apps/api-go/internal/modules/tasks/service_bulk.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 33: Tasks Reordering & Status Transitions (/reorder, /status, /assignees, /priority)

**Description:** Implement specialized task mutation endpoints in internal/modules/tasks: PATCH /api/tasks/:id/reorder (Lexorank recalculation), PATCH /api/tasks/:id/status (with transition verification), PATCH /api/tasks/:id/assignees, PATCH /api/tasks/:id/unassign, PATCH /api/tasks/:id/priority, and PATCH /api/tasks/:id/due-date.

**Acceptance criteria:**
- [ ] Reordering recalculates task rank and updates position in kanban column
- [ ] Status change verifies transition is permitted by workflow status graph
- [ ] Assignee updates synchronize task_assignees join table

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/tasks/reorder_test.go`
- [ ] Build succeeds: `go build ./internal/modules/tasks/...`
- [ ] Manual check: Drag and reorder task on kanban board and verify position persistence

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/tasks/handler_patch.go`
- `apps/api-go/internal/modules/tasks/service_patch.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 34: Task Recurrence Engine & Cron (/recurrence, complete occurrence, daily runner)

**Description:** Implement recurring tasks lifecycle in internal/modules/tasks: POST /api/tasks/:id/recurrence (configures DAILY, WEEKLY, MONTHLY pattern), POST /api/tasks/:id/complete-occurrence, and the midnight cron job generating upcoming occurrences with 1-day lead time.

**Acceptance criteria:**
- [ ] `POST /api/tasks/:id/recurrence` creates RecurringTask configuration
- [ ] `POST /api/tasks/:id/complete-occurrence` marks current task done and generates next occurrence
- [ ] Midnight scheduler evaluates active recurring tasks and generates due instances

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/tasks/recurrence_test.go`
- [ ] Build succeeds: `go build ./internal/modules/tasks/...`
- [ ] Manual check: Configure weekly recurrence, complete occurrence, and verify next task date

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/tasks/recurrence_service.go`
- `apps/api-go/internal/modules/tasks/handler_recurrence.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 35: Task Dependencies Module (/api/task-dependencies blocking, blocked-by)

**Description:** Implement task dependency graph in internal/modules/task_dependencies: creating dependencies (BLOCKS, BLOCKED_BY, RELATES_TO), circular dependency cycle detection, listing dependencies by task, and removing dependencies.

**Acceptance criteria:**
- [ ] `POST /api/task-dependencies` rejects self-dependency and circular dependency loops with 400 Bad Request
- [ ] `GET /api/task-dependencies/task/:id` returns blockers and dependent tasks
- [ ] `DELETE /api/task-dependencies/:id` removes dependency link

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/task_dependencies/...`
- [ ] Build succeeds: `go build ./internal/modules/task_dependencies/...`
- [ ] Manual check: Attempt to create circular dependency A -> B -> A and assert error returned

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/task_dependencies/handler.go`
- `apps/api-go/internal/modules/task_dependencies/service.go`
- `apps/api-go/internal/modules/task_dependencies/repository.go`
- `apps/api-go/internal/modules/task_dependencies/routes.go`
- `apps/api-go/internal/modules/task_dependencies/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 36: Task Watchers & Labels Association (/api/task-watchers, /api/task-label)

**Description:** Implement task watchers (subscribers) in internal/modules/task_watchers and label associations in internal/modules/task_label: subscribing/unsubscribing users to task notifications and attaching/detaching labels.

**Acceptance criteria:**
- [ ] `POST /api/task-watchers` adds watcher; prevents duplicate subscriptions
- [ ] `GET /api/task-watchers/task/:id` lists active watchers with user details
- [ ] `POST /api/task-label` links label to task; `DELETE /api/task-label` unlinks

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/task_watchers/... ./internal/modules/task_label/...`
- [ ] Build succeeds: `go build ./internal/modules/task_watchers/... ./internal/modules/task_label/...`
- [ ] Manual check: Subscribe user to task watcher list and verify in watcher list query

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/task_watchers/handler.go`
- `apps/api-go/internal/modules/task_watchers/service.go`
- `apps/api-go/internal/modules/task_watchers/routes.go`
- `apps/api-go/internal/modules/task_label/handler.go`
- `apps/api-go/internal/modules/task_label/service.go`
- `apps/api-go/internal/modules/task_label/routes.go`

**Estimated scope:** Large: 5-8 files

---

## Task 37: Task Comments Module (/api/task-comments CRUD, user mentions @name)

**Description:** Implement task discussions in internal/modules/task_comments: comment creation, updating, deleting, listing by task, and user `@mention` parsing that automatically triggers in-app notifications for tagged members.

**Acceptance criteria:**
- [ ] `POST /api/task-comments` creates comment and parses `@name` mentions to notify mentioned users
- [ ] `GET /api/task-comments/task/:id` returns threaded comments ordered by createdAt ascending
- [ ] Only comment author or workspace MANAGER/OWNER can edit or delete a comment

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/task_comments/...`
- [ ] Build succeeds: `go build ./internal/modules/task_comments/...`
- [ ] Manual check: Post comment with mention and verify notification is created for mentioned user

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/task_comments/handler.go`
- `apps/api-go/internal/modules/task_comments/service.go`
- `apps/api-go/internal/modules/task_comments/repository.go`
- `apps/api-go/internal/modules/task_comments/routes.go`
- `apps/api-go/internal/modules/task_comments/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 38: Task Attachments Module (/api/task-attachments upload, preview, download, stats)

**Description:** Implement task attachment file management in internal/modules/task_attachments: file upload (validating MIME types and MAX_FILE_SIZE), secure download, image preview, storage stats, and unlinking files on delete.

**Acceptance criteria:**
- [ ] `POST /api/task-attachments/upload/:taskId` handles multipart upload and records attachment row in DB
- [ ] `GET /api/task-attachments/:id/download` streams file with Content-Disposition attachment header
- [ ] `GET /api/task-attachments/stats` returns storage usage metrics by file type

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/task_attachments/...`
- [ ] Build succeeds: `go build ./internal/modules/task_attachments/...`
- [ ] Manual check: Upload image attachment to task, retrieve preview and download URLs

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/task_attachments/handler.go`
- `apps/api-go/internal/modules/task_attachments/service.go`
- `apps/api-go/internal/modules/task_attachments/repository.go`
- `apps/api-go/internal/modules/task_attachments/routes.go`
- `apps/api-go/internal/modules/task_attachments/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Task 39: Time Entries Module (/api/time-entries timer start/stop, log, summary, IDOR checks)

**Description:** Implement time tracking engine in internal/modules/time_entries: starting timer, stopping active timer (calculating exact duration seconds), manual time entry logging, time summary by user/task/project, and IDOR access validation.

**Acceptance criteria:**
- [ ] `POST /api/time-entries/start` starts timer; automatically stops any existing running timer for user
- [ ] `POST /api/time-entries/stop` stops timer and writes final duration
- [ ] IDOR check prevents modifying or deleting another user's time entries without MANAGER role

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/time_entries/...`
- [ ] Build succeeds: `go build ./internal/modules/time_entries/...`
- [ ] Manual check: Start timer, wait 3 seconds, stop timer, and verify duration recorded is >= 3 seconds

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/time_entries/handler.go`
- `apps/api-go/internal/modules/time_entries/service.go`
- `apps/api-go/internal/modules/time_entries/repository.go`
- `apps/api-go/internal/modules/time_entries/routes.go`
- `apps/api-go/internal/modules/time_entries/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Task 40: Gantt Module (/api/gantt project timeline, sprint schedule, resource allocation)

**Description:** Implement the Gantt chart and timeline analytics module in internal/modules/gantt: GET /api/gantt/project/:projectId (tasks with start/due dates, progress, dependencies), GET /api/gantt/sprint/:sprintId, and resource allocation by member.

**Acceptance criteria:**
- [ ] `GET /api/gantt/project/:projectId` returns tasks formatted for timeline rendering with dependencies
- [ ] `GET /api/gantt/project/:projectId/resources` aggregates workload distribution across team members
- [ ] Handles tasks without due dates gracefully without timeline rendering errors

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/gantt/...`
- [ ] Build succeeds: `go build ./internal/modules/gantt/...`
- [ ] Manual check: Fetch Gantt timeline for project and verify start/due dates format

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/gantt/handler.go`
- `apps/api-go/internal/modules/gantt/service.go`
- `apps/api-go/internal/modules/gantt/repository.go`
- `apps/api-go/internal/modules/gantt/routes.go`
- `apps/api-go/internal/modules/gantt/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Checkpoint: After Task 40 — Task Execution Engine Verified

- [ ] Task CRUD, project-scoped sequential keys ('ORB-1'), and Lexorank reordering verified
- [ ] Kanban grouped queries, today's agenda, and bulk status/delete operations work
- [ ] Task recurrence cron, dependencies, watchers, comments with mentions, attachments, and timers verified
- [ ] Review with human before proceeding

---

## Task 41: Real-Time Events Gateway (Socket.IO v4 Engine at /events, Room Management)

**Description:** Implement the Socket.IO v4 protocol server in internal/platform/events mounting at /events. Authenticates client handshake using JWT token (auth, query, or cookie), validates ACTIVE account state, tracks connected sockets, and joins user/org/workspace/project/task rooms.

**Acceptance criteria:**
- [ ] Socket.IO connection handshake succeeds at `/events` namespace with valid JWT
- [ ] Rejects unauthenticated connections or inactive users with 'Authentication token required'
- [ ] Handles client events: `join:organization`, `leave:organization`, `join:workspace`, `join:project`, `join:task`

**Verification:**
- [ ] Tests pass: `go test -v ./internal/platform/events/...`
- [ ] Build succeeds: `go build ./internal/platform/events/...`
- [ ] Manual check: Connect with Socket.IO test client, verify connected event and room subscription

**Dependencies:** Task 10

**Files likely touched:**
- `apps/api-go/internal/platform/events/gateway.go`
- `apps/api-go/internal/platform/events/rooms.go`
- `apps/api-go/internal/platform/events/auth.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 42: WebSocket Event Dispatcher & Presence (user:online, typing, task events)

**Description:** Implement event dispatching bridge connecting Go services to the Events Gateway. Emits presence events (`user:online`, `user:offline`), typing indicators, and domain mutations (`task:created`, `task:updated`, `comment:added`, `notification`) to target rooms.

**Acceptance criteria:**
- [ ] Emits `user:online` to all users when first socket connects; emits `user:offline` when last disconnects
- [ ] Emits `task:updated` to `project:{id}` and `task:{id}` rooms on task changes
- [ ] Broadcast methods are thread-safe and non-blocking

**Verification:**
- [ ] Tests pass: `go test -v ./internal/platform/events/broadcast_test.go`
- [ ] Build succeeds: `go build ./internal/platform/events/...`
- [ ] Manual check: Trigger task status update via HTTP and verify WebSocket client in room receives event

**Dependencies:** Task 41

**Files likely touched:**
- `apps/api-go/internal/platform/events/broadcast.go`
- `apps/api-go/internal/platform/events/events.go`

**Estimated scope:** Small: 1-2 files

---

## Task 43: Notifications Module (/api/notifications list, unread count, mark-read, delete)

**Description:** Implement the Notification Center in internal/modules/notifications: creating notifications, listing notifications with filtering by category, unread badge counter, marking as read (single and mark-all), and deleting notifications.

**Acceptance criteria:**
- [ ] `GET /api/notifications` returns user notifications with pagination and unread status
- [ ] `GET /api/notifications/unread-count` returns accurate count of unread items
- [ ] `PATCH /api/notifications/mark-all-read` marks all unread notifications read for authenticated user

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/notifications/...`
- [ ] Build succeeds: `go build ./internal/modules/notifications/...`
- [ ] Manual check: Trigger notification creation and verify unread-count increments

**Dependencies:** Task 42

**Files likely touched:**
- `apps/api-go/internal/modules/notifications/handler.go`
- `apps/api-go/internal/modules/notifications/service.go`
- `apps/api-go/internal/modules/notifications/repository.go`
- `apps/api-go/internal/modules/notifications/routes.go`
- `apps/api-go/internal/modules/notifications/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Task 44: Activity Log Module (/api/activity-log audit capture, stats)

**Description:** Implement audit logging in internal/modules/activity_log: recording entity actions (CREATE, UPDATE, DELETE, STATUS_CHANGE) with old/new state, listing project activities, and aggregating organization activity statistics.

**Acceptance criteria:**
- [ ] `GET /api/project/:projectId` returns activity history ordered by timestamp descending
- [ ] `GET /api/organization/:organizationId/stats` aggregates activity metrics by type
- [ ] Audit capture records actor ID, entity ID, action type, and JSON metadata changes

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/activity_log/...`
- [ ] Build succeeds: `go build ./internal/modules/activity_log/...`
- [ ] Manual check: Fetch project activities and verify recorded action descriptions

**Dependencies:** Task 42

**Files likely touched:**
- `apps/api-go/internal/modules/activity_log/handler.go`
- `apps/api-go/internal/modules/activity_log/service.go`
- `apps/api-go/internal/modules/activity_log/repository.go`
- `apps/api-go/internal/modules/activity_log/routes.go`
- `apps/api-go/internal/modules/activity_log/dto.go`

**Estimated scope:** Medium: 3-5 files

---

## Checkpoint: After Task 44 — Real-Time Collaboration & Audit Verified

- [ ] Socket.IO connection at `/events` succeeds and joins user/project/task rooms
- [ ] Mutations broadcast real-time events (`task:updated`, `comment:added`, `notification`) to connected clients
- [ ] Activity log accurately records entity mutations with before/after state
- [ ] Review with human before proceeding

---

## Task 45: Queue Engine & Background Processors (asynq/Redis runtime, bulk-task, automation)

**Description:** Implement asynchronous background job processing in internal/platform/queue using hibiken/asynq and Redis. Implements workers for `bulk-task-import` (batch task creation from jobs) and `automation` (rule evaluation).

**Acceptance criteria:**
- [ ] Asynq worker connects to Redis and processes tasks concurrently up to MAX_CONCURRENT_JOBS
- [ ] `bulk-task-import` processor ingests items, creates tasks, and emits completion event
- [ ] `automation` processor evaluates trigger conditions and executes configured actions

**Verification:**
- [ ] Tests pass: `go test -v ./internal/platform/queue/...`
- [ ] Build succeeds: `go build ./internal/platform/queue/...`
- [ ] Manual check: Enqueue bulk import job and verify worker picks up and processes tasks

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/platform/queue/queue.go`
- `apps/api-go/internal/platform/queue/worker.go`
- `apps/api-go/internal/modules/tasks/bulk_import_processor.go`
- `apps/api-go/internal/modules/automation/automation_processor.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 46: Scheduled Jobs Engine (Robfig Cron: Reminders, Overdue, Recurring)

**Description:** Implement scheduled cron runner in internal/modules/scheduler using robfig/cron/v3. Registers exact cron expressions: hourly due date reminders, daily 9 AM overdue alerts, Monday 9 AM sprint summaries, midnight overdue reconciliation, and cleanup tasks.

**Acceptance criteria:**
- [ ] Scheduler boots with application and registers all 8 cron schedules
- [ ] Hourly reminder scanner finds tasks due within 24h and triggers email/in-app alert
- [ ] 9 AM scanner detects overdue tasks and dispatches alerts to assignees

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/scheduler/...`
- [ ] Build succeeds: `go build ./internal/modules/scheduler/...`
- [ ] Manual check: Trigger reminder scanner function manually in test and assert email dispatch

**Dependencies:** Task 34, Task 9

**Files likely touched:**
- `apps/api-go/internal/modules/scheduler/scheduler.go`
- `apps/api-go/internal/modules/scheduler/reminders.go`
- `apps/api-go/internal/modules/scheduler/overdue.go`
- `apps/api-go/internal/modules/scheduler/cleanup.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 47: Storage & File Management (/uploads/upload/:folder, /uploads/:folder/:file, /s3)

**Description:** Implement file storage service in internal/platform/storage and internal/modules/storage_files: avatar uploads (locked to 'avatar' folder with 50MB MAX_AVATAR_SIZE limit), serving uploaded files from disk, and AWS S3 presigned PUT/GET URL generator.

**Acceptance criteria:**
- [ ] `POST /uploads/upload/:folder` strictly allows only 'avatar' folder; rejects other paths with 400
- [ ] `GET /uploads/:folder/:filename` serves file with correct Content-Type and security headers
- [ ] `GET /s3/presigned-put-url` generates valid presigned S3 upload URL

**Verification:**
- [ ] Tests pass: `go test -v ./internal/platform/storage/... ./internal/modules/storage_files/...`
- [ ] Build succeeds: `go build ./internal/modules/storage_files/...`
- [ ] Manual check: Upload avatar image via curl and download through GET /uploads/avatar/:filename

**Dependencies:** Task 6

**Files likely touched:**
- `apps/api-go/internal/platform/storage/storage.go`
- `apps/api-go/internal/platform/storage/s3.go`
- `apps/api-go/internal/modules/storage_files/handler.go`
- `apps/api-go/internal/modules/storage_files/s3_handler.go`
- `apps/api-go/internal/modules/storage_files/routes.go`

**Estimated scope:** Large: 5-8 files

---

## Task 48: Editor Images & Media Assets (/api/editor-images/upload)

**Description:** Implement rich text editor image upload handler in internal/modules/editor_images: handles image upload up to MAX_EDITOR_IMAGE_SIZE (5MB), validates image MIME types, persists file to editor-images folder, and creates MediaAsset record in DB.

**Acceptance criteria:**
- [ ] `POST /api/editor-images/upload` accepts image (jpeg, png, gif, webp) and stores in editor-images destination
- [ ] Returns `{ url: '/uploads/editor-images/filename', id: '...' }` matching frontend editor expectations
- [ ] Rejects non-image files or payloads exceeding 5MB

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/editor_images/...`
- [ ] Build succeeds: `go build ./internal/modules/editor_images/...`
- [ ] Manual check: Upload editor image and verify file exists on disk and MediaAsset row exists in DB

**Dependencies:** Task 47

**Files likely touched:**
- `apps/api-go/internal/modules/editor_images/handler.go`
- `apps/api-go/internal/modules/editor_images/service.go`
- `apps/api-go/internal/modules/editor_images/routes.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 49: Search Module (/api/search/global, /advanced, /suggestions, /quick)

**Description:** Implement global search and suggestion engine in internal/modules/search: POST /api/search/global (searches across tasks, projects, workspaces, members), POST /api/search/advanced (multi-criteria filtering), GET /api/search/suggestions, and GET /api/search/quick.

**Acceptance criteria:**
- [ ] `POST /api/search/global` executes parameterized multi-table search scoped to user tenant permissions
- [ ] `GET /api/search/quick` returns instant matching task and project results for ⌘K palette
- [ ] Sanitizes search query string to prevent SQL injection

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/search/...`
- [ ] Build succeeds: `go build ./internal/modules/search/...`
- [ ] Manual check: Search for task title in global search and verify matching items returned

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/search/handler.go`
- `apps/api-go/internal/modules/search/service.go`
- `apps/api-go/internal/modules/search/repository.go`
- `apps/api-go/internal/modules/search/routes.go`
- `apps/api-go/internal/modules/search/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Task 50: Project Inbox & Email Sync (/api/inbox accounts, rules, IMAP processor)

**Description:** Implement email inbox integration in internal/modules/inbox: configuring email accounts for projects, inbox routing rules (sender filter, subject regex, action), message listing, and the IMAP sync background processor converting incoming emails into task comments.

**Acceptance criteria:**
- [ ] `POST /api/inbox/accounts` creates encrypted IMAP account credentials in EmailAccount table
- [ ] Inbox rules evaluate incoming message attributes and execute mapped actions (create task / add comment)
- [ ] `POST /api/inbox/sync/:id` manually triggers sync; cron checks every 5 minutes

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/inbox/...`
- [ ] Build succeeds: `go build ./internal/modules/inbox/...`
- [ ] Manual check: Query inbox accounts and verify encrypted credentials are not leaked in plaintext

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/inbox/handler.go`
- `apps/api-go/internal/modules/inbox/service.go`
- `apps/api-go/internal/modules/inbox/imap_sync.go`
- `apps/api-go/internal/modules/inbox/rules_engine.go`
- `apps/api-go/internal/modules/inbox/routes.go`

**Estimated scope:** Large: 5-8 files

---

## Task 51: Trello Sync Module & Scheduler (/api/trello-sync credentials, board import, sync loop)

**Description:** Implement Trello integration in internal/modules/trello_sync: credential connection (API key & token), board and list validation, importing Trello cards into Orbito tasks with rank calculation, and 1-minute interval poll sync scheduler.

**Acceptance criteria:**
- [ ] Trello API client uses safe outbound HTTP with SSRF prevention
- [ ] `POST /api/trello-sync/import-boards` converts Trello boards, lists, and cards into Orbito projects, statuses, and tasks
- [ ] `GET /api/trello-sync/status` returns connection and synchronization telemetry

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/trello_sync/...`
- [ ] Build succeeds: `go build ./internal/modules/trello_sync/...`
- [ ] Manual check: Validate mock Trello credentials and check status response structure

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/trello_sync/handler.go`
- `apps/api-go/internal/modules/trello_sync/service.go`
- `apps/api-go/internal/modules/trello_sync/trello_api.go`
- `apps/api-go/internal/modules/trello_sync/scheduler.go`
- `apps/api-go/internal/modules/trello_sync/routes.go`

**Estimated scope:** Large: 5-8 files

---

## Task 52: Jira Sync Module (/api/jira-sync safe outbound HTTP, credentials, project import)

**Description:** Implement Jira integration in internal/modules/jira_sync: connection setup with JIRA_ALLOWED_HOSTS validation, credentials encryption, project and issue status import, bidirectional webhook receiver, and sync status reporting.

**Acceptance criteria:**
- [ ] Strictly blocks private/internal hosts; only allows hosts matching `JIRA_ALLOWED_HOSTS`
- [ ] `POST /api/jira-sync/import-projects` creates Orbito project mapped to Jira project key
- [ ] `POST /api/jira-sync/webhook` processes Jira issue updates and reflects them in Orbito tasks

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/jira_sync/...`
- [ ] Build succeeds: `go build ./internal/modules/jira_sync/...`
- [ ] Manual check: Attempt connection to internal IP (e.g. 192.168.1.1) and verify SSRF refusal

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/jira_sync/handler.go`
- `apps/api-go/internal/modules/jira_sync/service.go`
- `apps/api-go/internal/modules/jira_sync/jira_api.go`
- `apps/api-go/internal/modules/jira_sync/routes.go`

**Estimated scope:** Large: 5-8 files

---

## Task 53: AI Chat Module (/api/ai-chat conversations, chat endpoint, tools, test-connection)

**Description:** Implement the conversational AI assistant in internal/modules/ai_chat: conversation CRUD, chat completion handler calling OpenAI/Anthropic/OpenRouter via safe HTTP client, system prompts with app guide, and task description generator.

**Acceptance criteria:**
- [ ] `GET /api/ai-chat/conversations` and `POST /api/ai-chat/conversations` manage chat threads
- [ ] `POST /api/ai-chat/chat` streams or returns assistant responses and executes tool calls
- [ ] `POST /api/ai-chat/generate-description` generates structured task description from title

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/ai_chat/...`
- [ ] Build succeeds: `go build ./internal/modules/ai_chat/...`
- [ ] Manual check: Call test-connection with mock API key and verify response shape

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/ai_chat/handler.go`
- `apps/api-go/internal/modules/ai_chat/service.go`
- `apps/api-go/internal/modules/ai_chat/app_guide.go`
- `apps/api-go/internal/modules/ai_chat/routes.go`
- `apps/api-go/internal/modules/ai_chat/dto.go`

**Estimated scope:** Large: 5-8 files

---

## Task 54: Public Module (/api/public shared tasks, calendar, sprints, public project view)

**Description:** Implement public unauthenticated shared routes in internal/modules/public: public task share by token, public shared calendar, public project kanban board, and public sprint overview.

**Acceptance criteria:**
- [ ] All `/api/public/*` routes allow access without Authorization header or cookies
- [ ] `GET /api/public/tasks/:taskId` and `GET /api/public/tasks/key/:slug` return task details if public share exists
- [ ] `GET /api/public/:workspaceSlug/projects/:projectSlug/taskskanban` returns tasks for public projects

**Verification:**
- [ ] Tests pass: `go test -v ./internal/modules/public/...`
- [ ] Build succeeds: `go build ./internal/modules/public/...`
- [ ] Manual check: Access public task endpoint via unauthenticated curl request and verify 200 OK

**Dependencies:** Task 30

**Files likely touched:**
- `apps/api-go/internal/modules/public/handler_tasks.go`
- `apps/api-go/internal/modules/public/handler_projects.go`
- `apps/api-go/internal/modules/public/handler_calendar.go`
- `apps/api-go/internal/modules/public/handler_sprints.go`
- `apps/api-go/internal/modules/public/routes.go`

**Estimated scope:** Large: 5-8 files

---

## Checkpoint: After Task 54 — Async Workers, Integrations & Public Access Verified

- [ ] Redis queue workers process background import and automation jobs
- [ ] Scheduled cron jobs execute reminders, overdue alerts, and midnight task occurrences
- [ ] Avatar uploads, editor images, search, Trello/Jira sync, AI assistant, and public links functional
- [ ] Review with human before proceeding

---

## Task 55: Automated Parity Test Harness (Dual HTTP Request Fuzzer)

**Description:** Build an automated parity testing suite in apps/api-go/test/parity. Fires identical HTTP requests to NestJS (:3000) and Go (:3002), asserting exact status codes, headers (content-type, correlation-id), and JSON response equality across all 386 endpoints.

**Acceptance criteria:**
- [ ] Parity test harness covers all 386 endpoints across the 39 modules
- [ ] Asserts exact JSON key names, types, nullability, and error message structures
- [ ] Emits summary report confirming 100% parity pass rate

**Verification:**
- [ ] Tests pass: `go test -v ./test/parity/...`
- [ ] Build succeeds: `go build ./test/parity/...`
- [ ] Manual check: Run `make parity-test` and inspect diff report

**Dependencies:** Task 7 through Task 54

**Files likely touched:**
- `apps/api-go/test/parity/runner.go`
- `apps/api-go/test/parity/comparator.go`
- `apps/api-go/test/parity/routes_test.go`

**Estimated scope:** Medium: 3-5 files

---

## Task 56: Production Containerization & CI/CD Pipeline Integration

**Description:** Finalize deploy/Dockerfile using multi-stage distroless build, update docker-compose.prod.yml with Go backend service, add GitHub Actions workflow step in .github/workflows/ci.yml, and write comprehensive operational runbook in apps/api-go/README.md.

**Acceptance criteria:**
- [ ] `docker build -f deploy/Dockerfile .` produces minimal (<30MB) distroless container image
- [ ] `turbo run lint test build` passes in root repository CI pipeline
- [ ] `README.md` documents environment variables, build steps, architecture, and cutover procedures

**Verification:**
- [ ] Tests pass: `docker build -f apps/api-go/deploy/Dockerfile apps/api-go`
- [ ] Build succeeds: `turbo run build`
- [ ] Manual check: Boot Go container with docker compose and verify /api/health returns 200

**Dependencies:** Task 55

**Files likely touched:**
- `apps/api-go/deploy/Dockerfile`
- `apps/api-go/README.md`
- `docker-compose.prod.yml`
- `.github/workflows/ci.yml`

**Estimated scope:** Medium: 3-5 files

---

## Checkpoint: After Task 56 — Final Acceptance & Ready for Production Cutover

- [ ] Automated parity test harness reports 100% status code, header, and JSON shape match across all 386 endpoints
- [ ] Production distroless Docker container builds and starts cleanly under 30MB memory RSS
- [ ] NestJS backend remains 100% untouched and passing its own test suite
- [ ] Ready for seamless reverse-proxy canary cutover
- [ ] Review with human before proceeding

---
 