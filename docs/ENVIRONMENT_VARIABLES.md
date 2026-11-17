# Orbito Environment Variables Specification & Security Matrix

> **Authoritative Reference for System Configuration & Secret Management.**  
> Platform: **Orbito**  
> Enforces: OWASP ASVS 4.0 Secret Management Standards  

---

## 1. Secrets Management Rules

1. **Never Commit Secrets:** Real API keys, JWT secrets, passwords, or private keys must never exist in committed git files.
2. **Fail Fast on Missing Required Keys:** The backend validates environment variables on boot via NestJS `ConfigModule` and fails immediately if critical secrets are absent.
3. **Frontend Separation:** Only variables prefixed with `NEXT_PUBLIC_` are bundled into client-side JavaScript. Never prefix database URLs, JWT keys, or backend secrets with `NEXT_PUBLIC_`.

---

## 2. Backend Environment Variables (`backend/.env`)

| Variable | Required | Default | Description | Production Notes |
| :--- | :---: | :--- | :--- | :--- |
| `NODE_ENV` | **Yes** | `development` | Runtime environment (`production`, `staging`, `development`, `test`) | Must be set to `production` |
| `PORT` | No | `3000` | HTTP port the NestJS server listens on | Default: 3000 |
| `HOST` | No | `0.0.0.0` | Bind host address | Default: 0.0.0.0 |
| `DATABASE_URL` | **Yes** | - | PostgreSQL connection URI | Use connection pooling (PgBouncer) if > 1,000 req/s |
| `REDIS_HOST` | No | `localhost` | Redis host for caching and BullMQ | Required for distributed multi-instance deployment |
| `REDIS_PORT` | No | `6379` | Redis TCP port | Default: 6379 |
| `REDIS_PASSWORD` | No | - | Redis authentication password | Required in production |
| `JWT_SECRET` | **Yes** | - | 256-bit symmetric key for access tokens | Minimum 32 bytes base64 encoded |
| `JWT_REFRESH_SECRET` | **Yes** | - | 256-bit symmetric key for refresh tokens | Distinct from `JWT_SECRET` |
| `JWT_EXPIRES_IN` | No | `15m` | Access token lifespan | Recommended: `15m` or `1h` |
| `JWT_REFRESH_EXPIRES_IN`| No | `7d` | Refresh token lifespan | Recommended: `7d` |
| `ENCRYPTION_KEY` | **Yes** | - | 32-byte hexadecimal key for AES-256-GCM | Used to encrypt user BYOK API keys & OAuth secrets |
| `CORS_ORIGIN` | **Yes** | `http://localhost:3001` | Allowed browser origins (comma-separated) | In prod: `https://orbito.yourdomain.com` |
| `STORAGE_DRIVER` | No | `local` | Attachment storage provider (`local`, `s3`) | Use `s3` for multi-instance deployments |
| `AWS_S3_BUCKET` | Conditional | - | S3 bucket name if `STORAGE_DRIVER=s3` | Required if `STORAGE_DRIVER=s3` |
| `AWS_ACCESS_KEY_ID` | Conditional | - | AWS IAM credential | Required if `STORAGE_DRIVER=s3` |
| `AWS_SECRET_ACCESS_KEY`| Conditional | - | AWS IAM credential | Required if `STORAGE_DRIVER=s3` |
| `AWS_REGION` | Conditional | `us-east-1` | S3 AWS region | e.g., `us-east-1`, `eu-west-1` |

---

## 3. Frontend Environment Variables (`frontend/.env`)

| Variable | Required | Default | Description |
| :--- | :---: | :--- | :--- |
| `NEXT_PUBLIC_API_URL` | **Yes** | `http://localhost:3000/api` | Base HTTP endpoint for NestJS backend |
| `NEXT_PUBLIC_WS_URL` | **Yes** | `http://localhost:3000` | WebSocket endpoint for real-time collaboration |
| `NODE_ENV` | **Yes** | `production` | Node build and runtime environment |

---

## 4. Secret Generation Quick Reference

```bash
# Generate high-entropy 256-bit base64 secret (JWT_SECRET & JWT_REFRESH_SECRET)
openssl rand -base64 32

# Generate 32-byte hexadecimal key (ENCRYPTION_KEY for AES-256-GCM)
openssl rand -hex 32

# Generate secure database/redis password
openssl rand -base64 24
```
