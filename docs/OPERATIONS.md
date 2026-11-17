# Orbito Production Operations & Incident Runbook

> **Target:** System Administrators, SREs, and On-Call Engineers  
> **Platform:** Orbito Enterprise  

---

## 1. System Health & Probes

| Endpoint | Target Probe | Expected Output | Purpose |
| :--- | :--- | :--- | :--- |
| `GET /api/health/liveness` | Container Liveness | `200 OK` `{"status":"ok"}` | Orchestrator health check |
| `GET /api/health/readiness`| Traffic Readiness | `200 OK` (Database & Redis connected) | Load balancer target health |
| `GET /api/health` | Detailed Telemetry | System uptime, memory, DB latency | Admin diagnostics |

---

## 2. Common Operational Tasks

### Inspecting Live Container Logs
```bash
# Backend logs
docker compose -f docker-compose.prod.yml logs -f backend

# Frontend logs
docker compose -f docker-compose.prod.yml logs -f frontend

# Database logs
docker compose -f docker-compose.prod.yml logs -f postgres
```

### Accessing BullBoard Queue Telemetry
Admins can navigate to `http://localhost:3000/queues` or in UI `/admin/queues` to inspect:
- Job throughput and execution latency.
- Dead-letter queue failed jobs.
- Immediate retry triggers.

### Restarting Degraded Services
```bash
docker compose -f docker-compose.prod.yml restart backend
docker compose -f docker-compose.prod.yml restart frontend
```

---

## 3. Incident Triage Matrix

| Symptom | Probable Cause | Action |
| :--- | :--- | :--- |
| **502 Bad Gateway** | Backend or Frontend container crashed or restarted | Check `docker compose logs backend`. Verify memory limits. |
| **DB Connection Spikes** | Leaked connections or slow queries | Inspect `SELECT count(*) FROM pg_stat_activity;`. Ensure connection pool limits are configured. |
| **High Queue Latency** | Redis memory pressure or unhandled worker exceptions | Check BullBoard (`/admin/queues`), verify Redis memory with `redis-cli info memory`. |
