# Orbito Production Deployment Guide

> **Target Environments:** Docker Compose, Kubernetes (K8s), Ubuntu / Debian Bare-Metal / Cloud VM  
> **Platform Version:** Orbito 1.0.0  

---

## 1. Production Architecture Topology

```
                              Internet (HTTPS / 443)
                                        │
                                        ▼
                         ┌─────────────────────────────┐
                         │   Nginx Reverse Proxy / WAF │
                         └──────────────┬──────────────┘
                    /api & /socket.io   │   / & static assets
                                        ├──────────────────────┐
                                        ▼                      ▼
                         ┌─────────────────────────┐ ┌─────────────────────────┐
                         │  orbito-backend (NestJS)│ │ orbito-frontend (Next)  │
                         │  Port: 3000             │ │ Port: 3001             │
                         └──────────────┬──────────┘ └─────────────────────────┘
                                        │
                         ┌──────────────┴──────────────┐
                         ▼                             ▼
          ┌─────────────────────────────┐ ┌─────────────────────────────┐
          │  PostgreSQL 16 (Port 5432)  │ │      Redis 7 (Port 6379)    │
          │  Persistent Volume Storage  │ │  BullMQ Async Jobs & Events │
          └─────────────────────────────┘ └─────────────────────────────┘
```

---

## 2. Infrastructure Requirements

| Sizing | vCPU | RAM | Storage | Target Scale |
| :--- | :--- | :--- | :--- | :--- |
| **Minimum** | 2 vCPUs | 4 GB | 40 GB SSD | Up to 100 concurrent users |
| **Standard** | 4 vCPUs | 8 GB | 100 GB NVMe | Up to 1,000 concurrent users |
| **High Scale**| 8+ vCPUs | 16+ GB | 250+ GB NVMe | 5,000+ concurrent users |

---

## 3. Deployment Methods

### Option A: Production Docker Compose (Recommended)

1. **Clone the repository on the target server:**
   ```bash
   git clone <REPO_URL> /opt/orbito
   cd /opt/orbito
   ```

2. **Generate Cryptographic Production Secrets:**
   ```bash
   # JWT Access Secret
   openssl rand -base64 32
   # JWT Refresh Secret
   openssl rand -base64 32
   # AES-256-GCM Encryption Key for BYOK / Integrations
   openssl rand -hex 32
   # Database Password
   openssl rand -base64 24
   ```

3. **Configure Environment:**
   ```bash
   cp .env.example .env.prod
   # Edit .env.prod with production domains, SSL certificates, and secrets
   ```

4. **Launch Containers:**
   ```bash
   docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
   ```

5. **Execute Database Migrations in Container:**
   ```bash
   docker compose -f docker-compose.prod.yml exec backend npx prisma migrate deploy
   ```

6. **Verify System Status:**
   ```bash
   curl -f http://localhost:3000/api/health
   curl -f http://localhost:3000/api/health/readiness
   ```

---

### Option B: Bare-Metal / Node.js PM2 Deployment

1. **Install Node.js 22 LTS, PostgreSQL 16, and Redis 7:**
   ```bash
   sudo apt update && sudo apt install -y nodejs npm postgresql redis-server nginx
   ```

2. **Compile Monorepo Packages:**
   ```bash
   npm ci
   npm run build
   ```

3. **Deploy Database Schema:**
   ```bash
   npx prisma migrate deploy --schema=backend/prisma/schema.prisma
   ```

4. **Run Backend with PM2:**
   ```bash
   pm2 start backend/dist/main.js --name "orbito-backend" -i max
   ```

5. **Run Frontend with PM2:**
   ```bash
   cd frontend
   pm2 start "node server.js" --name "orbito-frontend"
   ```

---

## 4. Reverse Proxy & SSL Configuration (Nginx)

```nginx
server {
    listen 80;
    server_name orbito.yourdomain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name orbito.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/orbito.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/orbito.yourdomain.com/privkey.pem;

    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Frontend Application
    location / {
        proxy_pass http://127.0.0.1:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Backend API & Health Probes
    location /api/ {
        proxy_pass http://127.0.0.1:3000/api/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket Real-Time Events
    location /socket.io/ {
        proxy_pass http://127.0.0.1:3000/socket.io/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

---

## 5. Rollback Runbook

If a critical incident occurs after deployment:
1. Revert container image tags:
   ```bash
   docker compose -f docker-compose.prod.yml down
   docker compose -f docker-compose.prod.yml up -d --build --no-deps backend frontend
   ```
2. Database rollback (if non-breaking migration occurred):
   ```bash
   docker compose -f docker-compose.prod.yml exec backend npx prisma migrate resolve --rolled-back "<MIGRATION_NAME>"
   ```
