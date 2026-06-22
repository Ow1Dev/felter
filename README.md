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

`make dev` spins up the Go services and the Angular dev server in Docker containers with volume mounts so code changes are reflected immediately.
