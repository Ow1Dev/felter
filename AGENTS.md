Felter Repo Guidance (High-Signal Only)

## Scope

Three parts: Go **fieldservice** HTTP API, Go **userservice** (gRPC + HTTP), and Angular 21 SPA in `web/`.  
Module: `github.com/Ow1Dev/felter`, Go 1.24.

- **fieldservice** — stateless HTTP API (`cmd/fieldservice`). `GET /api/hello` → `{"message":"hello world"}`.
- **userservice** — stateful gRPC (`:9091`) + HTTP (`:9090`) API (`cmd/userservice`). Postgres-backed. Proto definitions in `proto/userservice/`.
- **Shared database** — single Postgres container. Migrations live per-service under `internal/<service>/migrations/` and are applied by `cmd/migrate`.

## Entrypoints

- **fieldservice**: `cmd/fieldservice/main.go`. Wires `internal/fieldservice/httpserver`.
- **userservice**: `cmd/userservice/main.go`. Boots gRPC + HTTP listeners; imports `internal/userservice/grpcserver`, `internal/userservice/httpserver`, `internal/userservice/store`.
- **migrate**: `cmd/migrate/main.go`. Uses `golang-migrate/migrate` library. Discovers `internal/<service>/migrations/`, creates per-service tracking tables (`<service>_migrations`), applies pending up migrations. Idempotent — safe to run multiple times.
- **Web**: `web/src/main.ts` bootstraps a standalone Angular app. Active root template is the **inline `template:`** in `app.ts` — `app.html` is a dead CLI scaffold file.
- `app.routes.ts` exports `Routes = []` (no routes yet). `HelloService` calls `GET {apiBaseUrl}/api/hello`; dev `apiBaseUrl = 'http://localhost:8080'`, prod `apiBaseUrl = '/api'`.

## Commands

```bash
# Dev
make dev              # fieldservice in background + Angular dev server in foreground (Ctrl-C kills only frontend)
make fieldservice     # go run ./cmd/fieldservice
make userservice      # go run ./cmd/userservice
make web              # cd web && bun run start  (NOT npm — Makefile uses bun directly)
make init             # cd web && bun install

# Infra
make up               # docker compose up -d postgres
make down             # docker compose down
make migrate          # go run ./cmd/migrate

# Go
make fmt              # gofumpt -w .  (NOT gofmt — uses gofumpt, stricter superset)
make fmt-check        # CI-style diff check
make tidy             # go mod tidy
make vet              # go vet ./...
make lint             # golangci-lint run  (requires Nix shell or manual install)
make test             # go test ./... -race -count=1

# Web
cd web && bun run build           # ng build (default config is PRODUCTION)
cd web && bun run ng build --configuration development  # explicit dev build
cd web && bun run vitest run      # CI-style headless tests
cd web && npx prettier --write .  # format web files

# Single test — Go
go test ./internal/fieldservice/handlers/... -run TestHandleHello -race
go test ./internal/httputil/... -run TestWriteJSONAndError/write_json_ok -race

# Single test — Web (Vitest)
cd web && vitest run src/app/app.spec.ts -t "should create the app"
```

## Environment Variables

| Variable | Default | Notes |
|---|---|---|
| `ADDRESS` | `:8080` | fieldservice bind address. Overrides `PORT`. |
| `PORT` | — | Digits only; auto-prefixed with `:`. |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:4200,http://127.0.0.1:4200` | fieldservice exact-origin allowlist. Empty list → **all origins allowed**. |
| `READ_TIMEOUT` / `WRITE_TIMEOUT` / `IDLE_TIMEOUT` | `15s` / `15s` / `60s` | `time.ParseDuration` format. |
| `GRPC_ADDRESS` | `:9091` | userservice gRPC bind address. |
| `HTTP_ADDRESS` | `:9090` | userservice HTTP bind address. |
| `DATABASE_DSN` | `postgres://felter:felter@localhost:5432/felter?sslmode=disable` | Postgres DSN for userservice and migrate. |
| `API_ADDR` | `:8080` | Makefile-only variable for fieldservice; stripped of `:` before passing as `PORT`. |

Address precedence (`fieldservice`): `ADDRESS` > `PORT` > `:8080`.

## CI

**`.github/workflows/web-ci.yml`** — triggers on `web/**` or `flake.*` changes, runs inside `nix develop`:
1. `bun install`
2. `bun run ng build --configuration production`
3. `bun run test` (= Vitest via Angular CLI)

**`.github/workflows/go-ci.yml`** — triggers on `cmd/**`, `internal/**`, `go.*`, `.golangci.yml`, `flake.*`, `proto/**`, runs inside `nix develop`:
1. `go mod tidy`
2. `golangci-lint run`
3. `go vet ./...`
4. `go test ./... -race -count=1`

## Gotchas

- **`gofumpt` not `gofmt`:** `make fmt` applies `gofumpt -w .`. Bare `gofmt` does not satisfy the linter.
- **`make web` uses `bun`, not npm:** If Bun is absent, run npm equivalents manually; the Makefile does not fall back automatically.
- **Angular build default is production:** `angular.json` sets `"defaultConfiguration": "production"`. Bare `ng build` builds prod. Dev requires `--configuration development`.
- **`recoverer` inlines JSON 500 as a raw string literal** in `internal/fieldservice/httpserver/middleware.go` to avoid an import cycle with `httputil`. Do not refactor it to use `httputil.WriteError` without resolving that cycle.
- **`make dev` only kills the foreground process:** The fieldservice (`go run ./cmd/fieldservice &`) runs in background. It exits cleanly via `signal.NotifyContext`, but a crash leaves it running.
- **`web/Dockerfile` is artifact-only:** Final stage is `FROM scratch` with only `/dist`. Not a runnable image.
- **Go linting:** `.golangci.yml` enables `revive` with the `exported` rule — all exported symbols must have doc comments. `max-issues-per-linter: 0` surfaces every issue. Generated `*.pb.go` files are excluded from lint.
- **External Go dependencies:** `go.mod` now includes `lib/pq`, `grpc`, `protobuf`, and `golang-migrate`. The repo is no longer stdlib-only.

## Migrations

Migrations are managed by `cmd/migrate` using the `golang-migrate/migrate` library (not a shell-out to a CLI binary).

### Per-service isolation

Each service has its own migration directory and its own tracking table in Postgres:

| Service | Directory | Tracking Table |
|---|---|---|
| `userservice` | `internal/userservice/migrations/` | `userservice_migrations` |
| `fieldservice` | `internal/fieldservice/migrations/` | `fieldservice_migrations` |

This keeps services fully independent — one can be rolled back without affecting the other.

### Naming convention

Every migration is a pair:

```
{version}_{description}.up.sql
{version}_{description}.down.sql
```

Example:
```
internal/userservice/migrations/
  001_create_users.up.sql
  001_create_users.down.sql
```

### Idempotency

`make migrate` is safe to run multiple times. Each service is applied only up to its latest version. If already at the latest version, the output shows `already up to date`.

### Adding a new service

1. Create `internal/<service>/migrations/` (must contain at least one `.sql` file)
2. Write `{NNN}_{name}.up.sql` and matching `.down.sql`
3. Run `make migrate` — it will be discovered automatically

## Where Things Live

- fieldservice wiring: `internal/fieldservice/httpserver/` (server, routes, cors, middleware)
- fieldservice handlers: `internal/fieldservice/handlers/`
- Shared JSON helpers: `internal/httputil/`
- Shared DB pool: `internal/db/`
- userservice config: `internal/userservice/config/`
- userservice gRPC implementation: `internal/userservice/grpcserver/`
- userservice HTTP wrapper: `internal/userservice/httpserver/`
- userservice persistence: `internal/userservice/store/`
- userservice migrations: `internal/userservice/migrations/`
- userservice protobuf: `proto/userservice/` and generated code in `internal/userservice/pb/`
- Angular app: `web/src/app/`; environments: `web/src/environments/`
- Agent skill packs: `.agents/skills/golang-pro/`, `.agents/skills/angular-developer/`

## Nix Dev Shell

`nix develop` provides `go`, `bun`, `nodejs_20`, `golangci-lint`, `gofumpt`, `protobuf`, Angular language server — matches CI exactly. Optional locally if you already have the right versions.
