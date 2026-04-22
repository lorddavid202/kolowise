# Deployment Guide

This guide covers production deployment for the backend stack defined in `docker-compose.prod.yml`.

## What This Compose Deploys

`docker-compose.prod.yml` provisions:

- `postgres` (PostgreSQL 16)
- `redis` (Redis 7)
- `ml` (FastAPI ML service)
- `api` (Go Gin API)
- `nginx` (reverse proxy exposing port 80)

Important: frontend (`apps/web`) is not included in this compose file. Deploy it separately (for example Vercel, static VM, or another container stack).

## Prerequisites

- Linux host or VM with Docker Engine + Docker Compose plugin
- Port `80` open to users
- Sufficient disk for DB volumes
- TLS termination strategy (external LB/proxy or Nginx with certs)

## 1) Prepare Environment File

Create `.env.production` from template:

```bash
cp .env.production.example .env.production
```

Set strong values, especially:

- `JWT_SECRET` (long random secret)
- `DATABASE_URL`
- `ML_SERVICE_URL` (default in compose network: `http://ml:8000`)

Reference shape:

```env
APP_ENV=prod
API_PORT=8080
DATABASE_URL=postgres://postgres:postgres@postgres:5432/kolowise?sslmode=disable
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
JWT_SECRET=<strong-random-secret>
JWT_ISSUER=kolowise
ML_SERVICE_URL=http://ml:8000
```

## 2) Build and Start Services

From repo root:

```bash
docker compose -f docker-compose.prod.yml up -d --build
```

Check status:

```bash
docker compose -f docker-compose.prod.yml ps
```

## 3) Apply Database Migrations

Run all `*.up.sql` files against production DB.

Example using the running Postgres container:

```bash
for f in infra/migrations/*.up.sql; do
  docker exec -i kolowise_postgres \
    psql -U postgres -d kolowise -v ON_ERROR_STOP=1 < "$f"
done
```

Run this step on first deploy and on each release that changes schema.

## 4) Verify Service Health

### Nginx -> API health

```bash
curl -sS http://<server-host>/healthz
```

### API direct health (inside network)

```bash
docker exec -it kolowise_api wget -qO- http://127.0.0.1:8080/healthz
```

### ML health

```bash
docker exec -it kolowise_ml python -c "import urllib.request; print(urllib.request.urlopen('http://127.0.0.1:8000/healthz').read().decode())"
```

## 5) Deploy Frontend Separately

When deploying `apps/web`, set:

```env
NEXT_PUBLIC_API_BASE_URL=https://<your-domain>/api/v1
```

`nginx.conf` routes `/api/*` and `/healthz` to the Go API.

## 6) Logging and Observability

View logs:

```bash
docker compose -f docker-compose.prod.yml logs -f api
docker compose -f docker-compose.prod.yml logs -f ml
docker compose -f docker-compose.prod.yml logs -f nginx
```

Recommended production additions:

1. Centralized logs (ELK, Loki, CloudWatch, etc.)
2. Metrics and dashboards (Prometheus/Grafana)
3. Alerting on API 5xx, DB health, and container restarts

## 7) Backup and Recovery

### PostgreSQL backup

```bash
docker exec -t kolowise_postgres pg_dump -U postgres -d kolowise > backup.sql
```

### PostgreSQL restore

```bash
cat backup.sql | docker exec -i kolowise_postgres psql -U postgres -d kolowise
```

Ensure backup retention and off-host storage are configured.

## 8) Rolling Update Process

Recommended release flow:

1. Build new images.
2. Apply migrations (backward-compatible first).
3. Restart services with new version.
4. Verify `/healthz` and key API flows.
5. Deploy frontend pointing to same API base path.

Command:

```bash
docker compose -f docker-compose.prod.yml up -d --build
```

## 9) Rollback Strategy

Minimum rollback plan:

1. Keep previous image tags available.
2. Re-run compose with previous tags.
3. If schema migration is not backward-compatible, execute explicit down migration or restore DB backup.

Do not rely on container restart alone for data-level rollback.

## Security Checklist

1. Use strong random `JWT_SECRET`.
2. Restrict DB and Redis access to internal network only.
3. Add TLS termination (HTTPS).
4. Set firewall rules to expose only required ports.
5. Rotate secrets and maintain audit logs.
