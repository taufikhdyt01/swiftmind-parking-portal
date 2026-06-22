# Swiftmind — Parking Violation Portal

Fullstack technical assignment (Tan Digital 2026). A portal where **Officers** issue parking
violations and **Members** pay the resulting fines, with versioned fine-rules and an immutable
per-violation calculation snapshot.

> 🚧 Work in progress — built phase by phase. Architecture lives in `DESIGN.md` (added later).

## Stack

- **Backend:** Go microservices (chi, pgx/v5, RabbitMQ, Redis, MinIO) behind a single **API
  Gateway** (JWT verification + RBAC). One Go module, one binary per service.
- **Frontend:** Next.js (App Router) + TypeScript, Tailwind v4 + shadcn/ui, light + dark theme,
  TanStack Query.
- **Infra:** PostgreSQL, RabbitMQ, Redis, MinIO — orchestrated with Docker Compose.

## Services

| Service        | Role                                                         | Status |
|----------------|-------------------------------------------------------------|--------|
| `gateway`      | Single entrypoint — JWT verification, RBAC, routing         | ✅ done |
| `identity`     | Users, login, JWT issuance (Officer / Member roles)         | ✅ done |
| `rules`        | Fine-rule versioning, active ruleset                        | planned |
| `violation`    | Violation submission, fine calculation + snapshot, photos   | planned |
| `payment`      | Invoices, mocked payment provider, payment records          | planned |
| `notification` | Event-driven notifications                                  | planned |

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (WSL2 backend on Windows)
- [Node.js](https://nodejs.org/) 20+
- [Go](https://go.dev/dl/) 1.26+ (only needed to run the backend outside Docker)

## Running locally

**1. Backend + infrastructure (Docker Compose):**

```bash
cp .env.example .env
docker compose up -d --build
```

This starts PostgreSQL, RabbitMQ, Redis, MinIO, and the `identity` + `gateway` services. The
gateway is the only backend port exposed to the host (`:8080`). On first start, the identity
service creates its schema and seeds the demo users.

**2. Frontend (Next.js dev server):**

```bash
cd web
npm install
npm run dev
```

Open **http://localhost:3000**. The frontend proxies `/api/*` to the gateway, so the auth cookie
stays first-party.

### Demo accounts

| Role    | Email                      | Password      |
|---------|----------------------------|---------------|
| Officer | `officer@swiftmind.test`   | `password123` |
| Member  | `member@swiftmind.test`    | `password123` |

### Service URLs

| Service       | URL                          |
|---------------|------------------------------|
| Web app       | http://localhost:3000        |
| API Gateway   | http://localhost:8080        |
| RabbitMQ UI   | http://localhost:15672       |
| MinIO Console | http://localhost:9001        |

## What works today (Phase 1)

- Login as Officer or Member; the gateway verifies the JWT and sets an httpOnly cookie.
- Role-aware dashboard shell; protected routes redirect unauthenticated users to `/login`.
- Light / dark theme toggle.

## Notes

- **Auth transport:** the gateway sets an httpOnly `access_token` cookie on login, verifies the
  JWT on each request, and forwards identity to downstream services as trusted `X-User-*` headers.
- **Schema:** each service bootstraps its own tables on startup (`CREATE TABLE IF NOT EXISTS`)
  and seeds idempotently. In production this would be replaced with versioned migrations.

## License

For assignment review purposes.
