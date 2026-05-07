---
title: "05 — Deployment, Build & Operations"
date: 2026-05-07
tags: [go, docker, deployment, makefile, devops, postgresql]
---

# 05 — Deployment, Build & Operations

A backend system that only runs on one developer's laptop is a prototype, not a project. `go-chat` is designed to be deployed by anyone with one command. This note explores the Docker setup, the multi-stage build, the migration strategy, and the Makefile that ties it all together.

## Docker Multi-Stage Build

The `Dockerfile` uses two stages:

```dockerfile
# Stage 1: Builder
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/chat cmd/api/*.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/migrate cmd/migrate/main.go

# Stage 2: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates netcat-openbsd
WORKDIR /app
COPY --from=builder /app/bin/chat /app/bin/migrate /app/
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh
RUN adduser -D -u 1000 appuser
USER appuser
EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
```

### Why Multi-Stage

The builder image contains the full Go toolchain, compiler, and source code — hundreds of megabytes. The runtime image contains only the compiled binaries, CA certificates, and `netcat` — under 20MB. This reduces attack surface, speeds up deployments, and eliminates unnecessary tooling in production.

### Why `CGO_ENABLED=0`

This produces a fully static binary with no dynamic linking to C libraries. The binary can run on any Linux distribution, including `scratch` or `distroless` images. Without this, the binary links to `glibc` and fails on Alpine's `musl`.

### Why `-ldflags="-w -s"`

- `-w` disables DWARF debugging info.
- `-s` strips the symbol table.

Together they reduce binary size by ~30%. In a container, you don't need debug symbols. If you need to debug, you build a separate image without these flags.

### Why Non-Root User

The final image runs as `appuser` (UID 1000), not root. If the application is compromised, the attacker gains the privileges of `appuser` — which is none. They cannot install packages, modify system files, or escape the container easily. This is a standard security baseline.

## Docker Compose Orchestration

```yaml
version: '3.8'

services:
  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: gochat
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d gochat"]
      interval: 5s
      timeout: 5s
      retries: 5

  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      DB_ADDR: postgres://user:password@db:5432/gochat?sslmode=disable
      JWT_SECRET: ${JWT_SECRET}
    depends_on:
      db:
        condition: service_healthy

volumes:
  postgres_data:
```

### Why PostgreSQL Health Checks

The `app` service uses `depends_on` with `condition: service_healthy`. This means the Go application does not start until PostgreSQL is actually ready to accept connections — not just until the container is running. Without this, the app would crash on startup because the database wasn't initialized yet.

### Why Named Volumes

`postgres_data` is a named Docker volume. It persists the database across container restarts and rebuilds. If you run `docker-compose down -v`, the volume is deleted and you get a fresh database. This is the standard development workflow.

## The Entrypoint Script

```bash
#!/bin/sh
set -e

echo "Waiting for PostgreSQL..."
while ! nc -z db 5432; do
  sleep 0.1
done
echo "PostgreSQL is up"

echo "Running migrations..."
./migrate up

echo "Starting application..."
exec ./chat
```

### Why `nc` (netcat)

The `app` container uses `netcat-openbsd` to poll the database port. Go's PostgreSQL driver does not retry connections aggressively. Waiting in the entrypoint script ensures migrations run against a live database.

### Why `exec` for the Final Command

`exec ./chat` replaces the shell process with the Go binary. This means the Go process becomes PID 1, which receives Unix signals correctly. Without `exec`, signals go to the shell, which might not forward them, breaking graceful shutdown.

## Database Migrations

Migrations live in `db/migrations/` as numbered `.up.sql` and `.down.sql` files:

```
db/migrations/
  001_users.up.sql
  001_users.down.sql
  002_rooms.up.sql
  002_rooms.down.sql
  ...
```

The migration runner in `cmd/migrate/main.go` executes them in order. The `entrypoint.sh` runs `migrate up` before starting the app. This means a fresh deployment automatically creates all tables.

### Why Custom Migration Tool

The project uses a custom `cmd/migrate` binary instead of a third-party tool like `golang-migrate`. This keeps the dependency surface minimal and demonstrates understanding of how schema versioning works. In production, `golang-migrate` or `Atlas` would be preferable for features like transactional migrations and drift detection.

## The Makefile

```makefile
.PHONY: build run migrate-up migrate-down deps setup

build:
	go build -o bin/chat cmd/api/*.go

run:
	go run cmd/api/*.go

migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

deps:
	go mod tidy

setup: deps migrate-up
```

### Why Make

Make is universal. Every developer knows `make build`. It requires no Node.js, no task runners, no YAML parsers. The targets are self-documenting. For a Go project, Make is often sufficient — `go build`, `go test`, and `go run` are fast and deterministic.

### Local Development Workflow

```bash
cp .env.example .env
make setup    # install deps, run migrations
make run      # start server on :8080
```

In Docker:

```bash
docker-compose up        # build and start everything
docker-compose logs -f   # tail logs
docker-compose down -v   # destroy everything (fresh start)
```

## What This Is Not (Yet)

- **No CI/CD pipeline.** There is no GitHub Actions or GitLab CI. Builds are manual.
- **No health checks on the app.** The Go server has a `/v1/health` endpoint, but Docker Compose does not use it for orchestration.
- **No log aggregation.** Logs go to stdout. In production, you'd ship them to Loki, Datadog, or CloudWatch.
- **No secrets management.** `JWT_SECRET` is injected via environment variable. In production, use Vault, AWS Secrets Manager, or Kubernetes secrets.
- **No zero-downtime deployment.** `docker-compose up` stops the old container before starting the new one. For production, you need rolling updates, blue-green deployment, or Kubernetes.

## Interview Hook

**Q: "How would you run database migrations in Kubernetes?"**

A: I'd use an **init container**. The main app container doesn't start until the init container finishes. The init container runs the same `migrate up` binary, then exits. This guarantees the schema is correct before the app accepts traffic. For rollback safety, migrations must be backward-compatible: additive changes only (new columns, new tables). Destructive changes (dropping columns) happen in a separate deployment after the app no longer references them.

**Q: "Why Alpine for the runtime image?"**

A: Alpine is ~5MB. A full Debian or Ubuntu image is ~100MB. Smaller images mean faster pulls, faster deploys, and less attack surface. The trade-off is that Alpine uses `musl` instead of `glibc`, which can cause issues with C dependencies. Since `go-chat` is a pure Go binary with `CGO_ENABLED=0`, this is not a concern.

**Q: "How would you handle secrets in production?"**

A: Never commit secrets to Git. In Docker Compose, use `.env` files that are `.gitignore`d. In Kubernetes, use sealed secrets or external secrets operators. In cloud environments, use the provider's secret manager (AWS Secrets Manager, GCP Secret Manager) and mount them as environment variables or volumes. The application should fail fast on startup if a required secret is missing — no defaults, no fallbacks.

## Related Notes

- [[01 - Project Overview]]: What the system does and its current capabilities.
- [[02 - Architecture and Design Patterns]]: How the `config` struct maps environment variables to the application.
- [[03 — WebSocket Real-Time Architecture]]: Why the Hub runs in a goroutine alongside the HTTP server.
- [[04 - Authentication & Security]]: How `JWT_SECRET` is consumed by the auth layer.
