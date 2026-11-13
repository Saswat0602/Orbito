# 🚀 Orbito Enterprise Production Deployment Guide

> **Authoritative Operations & Deployment Manual for Production Rollout.**  
> Platform: **Orbito** (Enterprise Task, Project & Work Management System)  
> Stack: Next.js 16 (Standalone) · NestJS 11 · PostgreSQL 16 · Redis 7 (BullMQ) · Docker Compose

---

## 1. Production Architecture Overview

```
                          ┌────────────────────────┐
                          │   Internet / Traffic   │
                          └───────────┬────────────┘
                                      │ (HTTPS / 443)
                                      ▼
                        ┌───────────────────────────┐
                        │ Nginx / Cloudflare / ALB  │
                        └───────┬───────────┬───────┘
          /api/* and /socket.io │           │ / (HTML & Static Assets)
                                ▼           ▼
    ┌───────────────────────────────┐   ┌──────────────────────────────┐
    │  orbito-prod-backend (NestJS) │   │ orbito-prod-frontend (Next)  │
    │  Port: 3000                   │   │ Port: 3001                   │
    └──────────────┬────────────────┘   └──────────────────────────────┘
                   │
         ┌─────────┴─────────┐
         ▼                   ▼
┌──────────────────┐ ┌──────────────────┐
│  PostgreSQL 16   │ │     Redis 7      │
│  (Database &     │ │  (BullMQ Async   │
│   Lexorank Data) │ │   Jobs & Events) │
└──────────────────┘ └──────────────────┘
```

---

## 2. Pre-requisites & Minimum Server Sizing

| Component | Minimum Specification | Recommended Specification |
| :--- | :--- | :--- |
| **CPU** | 2 vCPUs | 4+ vCPUs |
| **RAM** | 4 GB RAM | 8+ GB RAM |
| **Disk** | 40 GB SSD (NVMe preferred) | 100+ GB SSD |
| **OS** | Ubuntu 22.04 LTS / Debian 12 | Ubuntu 24.04 LTS / Debian 12 |
| **Runtimes** | Docker 24.0+ & Compose v2 | Docker 26.0+ & Compose v2 |

---

## 3. Production Secrets Generation

Before deploying, generate cryptographically random keys for production:

```bash
# 1. Generate JWT Access Token Secret (256-bit)
openssl rand -base64 32

# 2. Generate JWT Refresh Token Secret (256-bit)
openssl rand -base64 32

# 3. Generate AES-256-GCM Encryption Key (for BYOK AI keys and OAuth tokens)
openssl rand -hex 32

# 4. Generate Strong Database & Redis Passwords
openssl rand -base64 24
```

---

## 4. Environment Configuration

### Step 1: Create the Root Production Environment File
Copy the production template:
```bash
cp .env.example .env
```

Ensure the following critical keys are configured:
```ini
# Production Environment
NODE_ENV=production
HOST=0.0.0.0
PORT=3000

# Security (Paste generated secrets)
JWT_SECRET="<YOUR_GENERATED_JWT_SECRET>"
JWT_REFRESH_SECRET="<YOUR_GENERATED_JWT_REFRESH_SECRET>"
JWT_EXPIRES_IN="15m"
JWT_REFRESH_EXPIRES_IN="7d"
ENCRYPTION_KEY="<YOUR_GENERATED_HEX_KEY>"

# PostgreSQL Database
POSTGRES_USER=orbito
POSTGRES_PASSWORD="<YOUR_STRONG_DB_PASSWORD>"
POSTGRES_DB=orbito
POSTGRES_PORT=5432
DATABASE_URL="postgresql://orbito:<YOUR_STRONG_DB_PASSWORD>@postgres:5432/orbito?schema=public"

# Redis Cache & Queue
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD="<YOUR_STRONG_REDIS_PASSWORD>"

# URLs and Domain Routing
FRONTEND_URL="https://orbito.yourdomain.com"
CORS_ORIGINS="https://orbito.yourdomain.com,https://api.yourdomain.com"
NEXT_PUBLIC_API_BASE_URL="https://orbito.yourdomain.com/api"
NEXT_PUBLIC_SOCKET_URL="https://orbito.yourdomain.com"

# SMTP Email Dispatch
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASS="<YOUR_SMTP_API_KEY>"
SMTP_FROM="noreply@yourdomain.com"

# Storage
UPLOAD_DEST="/app/backend/uploads"
```

---

## 5. One-Command Production Launch

Orbito includes a complete production compose orchestration with automated database migration gates:

```bash
# 1. Run local pre-flight checks to verify builds and configs
npm run preflight

# 2. Build and launch all production containers in background
docker compose -f docker-compose.prod.yml up -d --build
```

### What happens automatically:
1. `postgres` starts and waits for the database socket to become healthy.
2. `redis` initializes with append-only persistence.
3. `migration` container spins up, runs `npx prisma migrate deploy`, and exits cleanly.
4. `backend` starts NestJS on port 3000 after migrations complete.
5. `frontend` boots the optimized standalone Next.js 16 runtime on port 3001.

---

## 6. Verification & Health Monitoring

Verify live status via HTTP probes:

| Service | Health Check Endpoint | Expected Response |
| :--- | :--- | :--- |
| **Backend REST & DB** | `GET https://yourdomain.com/api/health` | `HTTP 200 {"status":"healthy"}` |
| **BullMQ Worker Queue** | `GET https://yourdomain.com/api/health/queue` | `HTTP 200 {"status":"ok"}` |
| **Frontend Web App** | `GET https://yourdomain.com/api/health` | `HTTP 200 {"status":"healthy"}` |

### Live Container Telemetry
```bash
# Inspect container health and resource usage
docker compose -f docker-compose.prod.yml ps
docker stats

# Stream logs from backend or frontend
docker compose -f docker-compose.prod.yml logs -f backend
docker compose -f docker-compose.prod.yml logs -f frontend
```

---

## 7. Nginx / Reverse Proxy & SSL Setup

For single-domain routing with automatic WebSockets support and SSL:
Use the included `docker/nginx/nginx.conf` template, or configure Certbot:

```bash
# Install Certbot for Let's Encrypt SSL
sudo apt install -y certbot python3-certbot-nginx

# Obtain and install SSL Certificate
sudo certbot --nginx -d orbito.yourdomain.com
```

---

## 8. Backup & Disaster Recovery Runbook

### Automated Daily Database Backup
Add a cron job (`crontab -e`):
```bash
# Backup PostgreSQL at 02:00 AM daily
0 2 * * * docker exec orbito-prod-postgres pg_dump -U orbito orbito | gzip > /backups/orbito_db_$(date +\%Y\%m\%d).sql.gz
```

### Restoring Database from Backup
```bash
gunzip < /backups/orbito_db_YYYYMMDD.sql.gz | docker exec -i orbito-prod-postgres psql -U orbito -d orbito
```

---

## 9. Rollback & Maintenance Strategy

- **Zero-Downtime Updates**:
  ```bash
  git pull origin master
  npm run preflight
  docker compose -f docker-compose.prod.yml up -d --build
  ```
- **Instant Emergency Rollback**:
  ```bash
  # Roll back to previous Git commit
  git checkout <LAST_KNOWN_STABLE_COMMIT>
  docker compose -f docker-compose.prod.yml up -d --build
  ```
