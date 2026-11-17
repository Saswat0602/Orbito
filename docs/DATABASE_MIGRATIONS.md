# Orbito Database Migrations & Data Integrity Manual

> **Database:** PostgreSQL 16  
> **ORM:** Prisma ORM 6  
> **Schema Location:** `backend/prisma/schema.prisma`  

---

## 1. Migration Philosophy & Rules

1. **Zero Destructive Commands in Production:**
   - **NEVER** run `prisma migrate reset` in staging or production.
   - **NEVER** run `prisma db push` in production.
   - **ALWAYS** use `npx prisma migrate deploy` in production environments.
2. **Backward-Compatible Expands:**
   - Adding columns must either supply a `@default(...)` or make the column optional (`?`).
   - Renaming columns must follow the Expand/Contract pattern across two releases.
3. **Audit Trails & Soft Deletes:**
   - Core tables include `createdAt`, `updatedAt`, and where applicable `deletedAt` timestamps.

---

## 2. Standard Migration Workflow

### Development (Authoring Migrations)
```bash
# 1. Update backend/prisma/schema.prisma
# 2. Generate a new migration script
npm run prisma:migrate:dev --workspace=@orbito/backend -- --name add_feature_columns

# 3. Regenerate Prisma Client
npm run prisma:generate --workspace=@orbito/backend
```

### Production Deployment
During automated CI/CD deployment or container boot:
```bash
# Deploys all pending committed migrations safely
npx prisma migrate deploy --schema=backend/prisma/schema.prisma
```

---

## 3. Migration Troubleshooting & Disaster Recovery

### Checking Migration Status
```bash
npx prisma migrate status --schema=backend/prisma/schema.prisma
```

### Resolving Failed Migrations
If a migration fails halfway due to a network interruption or constraint conflict:
```bash
# Mark as applied after manual fix:
npx prisma migrate resolve --applied "<MIGRATION_NAME>" --schema=backend/prisma/schema.prisma

# OR mark as rolled back:
npx prisma migrate resolve --rolled-back "<MIGRATION_NAME>" --schema=backend/prisma/schema.prisma
```

---

## 4. Backup & Restore Procedures

### Automated Database Backup (Daily Cron)
```bash
pg_dump -U orbito -h localhost -d orbito -F c -b -v -f "/var/backups/orbito_$(date +%Y%m%d_%H%M%S).dump"
```

### Database Restore
```bash
pg_restore -U orbito -h localhost -d orbito -v -c "/var/backups/orbito_20261117_000000.dump"
```
