# Felter

Field service management platform.

## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- [Bun](https://bun.sh/) (for local web development)
- Go 1.25+ (for local service development)

### First-time setup

```bash
# 1. Clone the repository
git clone <repo-url> && cd felter

# 2. Install web dependencies, start Postgres + Keycloak, run migrations
make init
```

`make init` does three things:
1. Installs JavaScript dependencies inside `web/` with `bun install`.
2. Starts the infrastructure services (Postgres and Keycloak) via Docker Compose.
3. Waits for Postgres to become healthy and then runs database migrations.

### Daily development

```bash
# Start all application services with live code reloading
make dev
```

`make dev` spins up the Go services (`fieldservice`, `userservice`, `projectservice`, `proxy`) and the Angular dev server (`web`) in Docker containers with volume mounts so code changes are reflected immediately.

### Access points

| Service            | URL                              |
|--------------------|----------------------------------|
| Angular SPA        | http://localhost:4200            |
| Auth Proxy         | http://localhost:9092            |
| fieldservice       | http://localhost:8080            |
| userservice (HTTP) | http://localhost:9090            |
| userservice (gRPC) | localhost:9091                   |
| projectservice     | http://localhost:8081            |
| Keycloak           | http://localhost:8180            |
| Postgres           | localhost:5432                   |

### Stopping services

```bash
# Stop application services only (Postgres & Keycloak keep running)
make down-app

# Stop everything including infrastructure and remove containers
make down
```

### Useful commands

```bash
# Run a single service locally (outside Docker)
make fieldservice    # :8080
make userservice     # gRPC :9091 + HTTP :9090
make proxy           # :9092
make projectservice  # :8081
make web             # bun start (localhost:4200)

# Database migrations
make migrate

# Code generation
make generate-api    # regenerate OpenAPI types for Go + TypeScript

# Go tooling
make fmt             # gofumpt -w .
make lint            # golangci-lint run
make test            # go test ./... -race -count=1
make tidy            # go mod tidy
make vet             # go vet ./...
```

## Architecture

| Component | Tech | Notes |
|---|---|---|
| `cmd/fieldservice/` | Go HTTP API | Stateless, `GET /api/hello` |
| `cmd/userservice/` | Go gRPC + HTTP | Postgres-backed, proto in `proto/userservice/` |
| `cmd/projectservice/` | Go HTTP API | Postgres-backed, OpenAPI v3 spec in `docs/api/projectservice.yaml` |
| `cmd/proxy/` | Go | Auth proxy with Keycloak OIDC + JWT |
| `cmd/migrate/` | Go | Applies `.up.sql` migrations per service |
| `web/` | Angular 21 SPA | Bun package manager, `ng serve` for dev |

## Docker Images

- **Dev Go image**: `infra/dev/Dockerfile` — mounts the repo for live reload.
- **Prod Go image**: `infra/prod/Dockerfile` — multi-stage `scratch` image, pass `APP_NAME` build arg.
- **Web prod image**: `web/Dockerfile` — artifact-only, outputs `/dist`.
- **Web dev image**: `web/Dockerfile.dev` — Bun-based dev server with live reload.

## Auth Flow

1. `AuthService.login()` redirects to Keycloak.
2. Keycloak redirects to `/callback?code=...`.
3. `CallbackPageComponent` exchanges the code for a JWT via `/api/auth/callback`.
4. `AuthService.currentUser` is populated via `/api/auth/me`.
5. Logout clears the session and redirects through Keycloak.

## Environment Variables

All proxy config vars are **required** (no defaults):
- `PROXY_HTTP_ADDRESS`, `PROXY_JWT_SECRET`, `PROXY_KEYCLOAK_URL`, `PROXY_KEYCLOAK_PUBLIC_URL`, `PROXY_KEYCLOAK_REALM`, `PROXY_KEYCLOAK_CLIENT_ID`, `PROXY_KEYCLOAK_CLIENT_SECRET`, `PROXY_KEYCLOAK_REDIRECT_URI`, `PROXY_USERSERVICE_GRPC_ADDR`, `PROXY_FIELD_URL`, `PROXY_USERSERVICE_URL`, `PROXY_PROJECTSERVICE_URL`

`PROXY_KEYCLOAK_PUBLIC_URL` is used for browser-facing redirects (login/logout URLs). When empty it falls back to `PROXY_KEYCLOAK_URL`. In Docker, set `PROXY_KEYCLOAK_URL` to the internal service name (e.g. `http://keycloak:8080`) and `PROXY_KEYCLOAK_PUBLIC_URL` to the host-facing URL (e.g. `http://localhost:8180`).

`fieldservice`: `ADDRESS` > `PORT` > `:8080`. `CORS_ALLOWED_ORIGINS` empty = allow all.
`userservice`/`migrate`: `DATABASE_DSN` required. `GRPC_ADDRESS` default `:9091`, `HTTP_ADDRESS` default `:9090`.
`projectservice`: `ADDRESS` > `PORT` > `:8081`. `DATABASE_DSN` required.

An `.env` file at the repo root is auto-loaded by the Makefile.
