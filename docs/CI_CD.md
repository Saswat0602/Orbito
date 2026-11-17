# Orbito CI/CD Pipeline & Automated Release Architecture

> **CI Engine:** GitHub Actions (`.github/workflows/ci.yml`)  
> **Package Automation:** Dependabot (`.github/dependabot.yml`)  

---

## 1. Pipeline Overview

Every Pull Request and push to `master` / `main` triggers automated verification stages:

```
                  ┌────────────────────────┐
                  │ Push / PR to Master    │
                  └───────────┬────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
     ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
     │ Typecheck    │  │ Backend Test │  │ Production   │
     │ All Packages │  │ Suite (Jest) │  │ Build (Turbo)│
     └──────────────┘  └──────────────┘  └──────────────┘
            │                 │                 │
            └─────────────────┼─────────────────┘
                              ▼
                  ┌────────────────────────┐
                  │ Preflight Audit & Lint │
                  └───────────┬────────────┘
                              ▼
                  ┌────────────────────────┐
                  │ Docker Build & Deploy  │
                  └────────────────────────┘
```

---

## 2. CI/CD Stage Breakdown

1. **Checkout & Dependency Hydration:** Uses `actions/checkout@v4` with `actions/setup-node@v4` (Node 22) and npm cache caching `~/.npm`.
2. **Typecheck (`turbo run typecheck`):** Verifies type contracts across `@orbito/types`, `@orbito/backend`, `@orbito/frontend`, and `@orbito/ui`.
3. **Backend Testing (`npm run test:backend`):** Executes 76+ Jest unit and integration tests, including RBAC, LexoRank, queue fallback, and SSRF sanitization.
4. **Turborepo Production Build (`turbo run build`):** Compiles NestJS backend and Next.js App Router frontend with static generation and tree-shaking.
5. **Docker Image Build & Push (Conditional on Release Tag):** Tags and pushes multi-stage container images to Docker Hub / GitHub Container Registry (GHCR).

---

## 3. Dependabot Management Strategy

To prevent active branch spam (e.g. 15+ concurrent Dependabot branches):
- Grouped into two consolidated update streams: `npm-dependencies` and `github-actions`.
- Schedule set to `monthly` with `open-pull-requests-limit: 2`.
- Production stability prioritized over constant bleeding-edge bumps.
