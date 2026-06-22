# Swiftmind — Parking Violation Portal

Fullstack technical assignment (Tan Digital 2026). A portal where **Officers** issue parking
violations and **Members** pay the resulting fines, with versioned fine-rules and an immutable
per-violation calculation snapshot.

> 🚧 Work in progress — built phase by phase. See `DESIGN.md` for architecture (added later).

## Stack

- **Backend:** Go microservices (chi, pgx + sqlc, golang-migrate, RabbitMQ, Redis, MinIO),
  fronted by a single **API Gateway**.
- **Frontend:** Next.js (App Router) + TypeScript, Tailwind + shadcn/ui, TanStack Query.
- **Infra:** PostgreSQL, RabbitMQ, Redis, MinIO — orchestrated with Docker Compose.

## Services

| Service        | Role                                                         |
|----------------|-------------------------------------------------------------|
| `gateway`      | Single entrypoint — JWT verification, RBAC, routing         |
| `identity`     | Users, login, JWT issuance (Officer / Member roles)         |
| `rules`        | Fine-rule versioning, active ruleset                        |
| `violation`    | Violation submission, fine calculation + snapshot, photos   |
| `payment`      | Invoices, mocked payment provider, payment records          |
| `notification` | Event-driven notifications                                  |

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (WSL2 backend on Windows)
- [Go](https://go.dev/dl/) 1.26+
- [Node.js](https://nodejs.org/) 20+

## Running locally

> Full run instructions are finalized in the last phase. For now, infrastructure can be started:

```bash
cp .env.example .env
docker compose up -d   # postgres, rabbitmq, redis, minio
```

| Service       | URL                          |
|---------------|------------------------------|
| RabbitMQ UI   | http://localhost:15672       |
| MinIO Console | http://localhost:9001        |

## License

For assignment review purposes.
