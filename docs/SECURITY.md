# Orbito Platform Security Architecture & Hardening Guide

> **Security Standard:** OWASP Top 10 (2021) & ASVS Level 2 Compliance  
> **Audited By:** Application Security Engineer  

---

## 1. Security Overview

Orbito implements defense-in-depth across the entire monorepo stack:
- **Authentication:** Dual-token JWT (short-lived access tokens + rotated refresh tokens) stored in `httpOnly`, `SameSite=Lax`, `Secure` cookies.
- **Authorization:** Hierarchical multi-tenant RBAC (`SUPER_ADMIN`, `OWNER`, `MANAGER`, `MEMBER`, `VIEWER`) enforced via NestJS `RolesGuard` and verified in UI with `useAuthRbac()`.
- **Data Protection:** Database-level encryption (AES-256-GCM) with IV and authentication tags for third-party OAuth tokens, Jira secrets, and LLM API keys.
- **Network Boundaries:** Reverse proxy TLS termination, strict CORS origin whitelisting, SSRF protection on outgoing webhooks/integrations, and rate limiting via `@nestjs/throttler`.

---

## 2. Threat Modeling & Safeguards

| Threat Vector | Mitigation Strategy | Implemented Location |
| :--- | :--- | :--- |
| **SQL Injection** | Parameterized queries enforced via Prisma ORM 6. Zero raw SQL concatenations. | `backend/src/prisma/` |
| **Cross-Site Scripting (XSS)** | HTML input sanitization (`DOMPurify` / `SanitizerUtil`), automatic React JSX output escaping. | `backend/src/common/utils/sanitizer.util.ts` |
| **Server-Side Request Forgery (SSRF)** | Outbound IP validation (`SsrfGuard`), private IP range blocking (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 127.0.0.0/8, AWS metadata 169.254.169.254). | `backend/src/common/guards/ssrf.guard.ts` |
| **Brute Force & DoS** | Global rate limiting (100 req/min standard, 10 req/min auth endpoints) via `@nestjs/throttler`. | `backend/src/app.module.ts` |
| **Cross-Site Request Forgery (CSRF)** | `SameSite=Lax` cookies, custom Authorization header verification, CSRF token validation on mutation endpoints. | `backend/src/modules/auth/` |
| **Sensitive Data Exposure** | Password hashing with `bcrypt` (10 salt rounds); AES-256-GCM encryption for integration tokens; stripping `password` fields from JSON serialization. | `backend/src/common/crypto/` |
| **Broken Object Level Auth (BOLA)** | Multi-tenant tenant boundary checks validating that requested Workspace and Project IDs belong to the caller's Organization. | `backend/src/modules/workspaces/` |

---

## 3. Dependency Vulnerability Management

- Monorepo automated updates configured via `.github/dependabot.yml` with grouped dependencies and monthly intervals to prevent branch spam.
- Transitive dependencies audited via `npm audit --omit=dev`.
- High-risk native dependencies (e.g. `sharp`, `tar`, `msgpackr-extract`) isolated and vetted for production images.

---

## 4. Incident Response & Security Vulnerability Reporting

If a security vulnerability is identified:
1. Do not open public GitHub issues.
2. Submit a confidential report to `security@orbito.local` or the designated security officer.
3. Security patches are prioritized and released within 24-48 hours.
