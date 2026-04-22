# Local Setup Guide

This guide gets KoloWise running locally with PostgreSQL, Redis, Go API, ML service, and Next.js web app.

## Prerequisites

Install these tools first:

- Docker + Docker Compose plugin
- Go `1.25.x`
- Node.js `20+` and npm
- Python `3.11+`
- `psql` CLI (PostgreSQL client)

## 1) Clone and Enter Project

```bash
git clone <repo-url>
cd kolowise
```

## 2) Configure Environment

### API environment (`.env.dev`)

`pkg/config` loads `.env` and `.env.dev` when `APP_ENV=dev` (or unset).

Required variables:

```env
APP_ENV=dev
API_PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5433/kolowise?sslmode=disable
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=lateengine
REDIS_DB=0
JWT_SECRET=<long-random-secret>
JWT_ISSUER=kolowise
ML_SERVICE_URL=http://localhost:8000
```

Notes:

- `ML_SERVICE_URL` must be present for API startup.
- `.env.example` currently does not include `ML_SERVICE_URL`, so add it explicitly.

### Web environment (`apps/web/.env.local`)

Choose one API access mode:

1. Direct to Go API (recommended for dev):

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
```

2. Through Nginx proxy on port 80:

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost/api/v1
```

## 3) Start Infrastructure (Postgres + Redis)

```bash
docker compose -f infra/compose/docker-compose.yml up -d
```

Check status:

```bash
docker compose -f infra/compose/docker-compose.yml ps
```

## 4) Run Database Migrations

From repo root:

```bash
for f in infra/migrations/*.up.sql; do
  psql 'postgres://postgres:postgres@localhost:5433/kolowise?sslmode=disable' -f "$f"
done
```

## 5) Start ML Service

```bash
cd apps/ml
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python train_models.py
uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
```

Health check:

```bash
curl -sS http://localhost:8000/healthz
```

## 6) Start Go API

In a new terminal, repo root:

```bash
APP_ENV=dev go run ./apps/api
```

Health check:

```bash
curl -sS http://localhost:8080/healthz
```

## 7) Start Web App

In a new terminal:

```bash
cd apps/web
npm install
npm run dev
```

Open `http://localhost:3000`.

## 8) Create First User

The current web UI provides a login form but no registration page. Create an account via API:

```bash
curl -sS http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "full_name":"Demo User",
    "email":"demo@example.com",
    "password":"Password123"
  }'
```

Then login from the web app using those credentials.

## 9) End-to-End Smoke Tests

### API login

```bash
curl -sS http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"Password123"}'
```

### Protected call

```bash
TOKEN=$(curl -sS http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"Password123"}' | jq -r .access_token)

curl -sS http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer $TOKEN"
```

## Common Issues

### `ML_SERVICE_URL is required`

Add `ML_SERVICE_URL=http://localhost:8000` to `.env.dev`.

### API returns `invalid or expired token`

- Ensure `Authorization: Bearer <token>` format is correct.
- Log in again (tokens expire after 24h).

### CSV upload fails

- Ensure file is real CSV.
- Include `txn_date` or `date` header.
- Provide either:
  - `amount` + `direction`, or
  - `debit`/`credit` columns.

### CORS errors in browser

API CORS allowlist currently includes `http://localhost:3000`.

## Stop Services

```bash
docker compose -f infra/compose/docker-compose.yml down
```

To remove volumes too:

```bash
docker compose -f infra/compose/docker-compose.yml down -v
```
