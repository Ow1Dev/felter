Felter Repo Guidance (High-Signal Only)

## Scope

Go **fieldservice** HTTP API, Go **userservice** (gRPC + HTTP), **proxy** auth service (Keycloak OIDC + JWT), Angular 21 SPA in `web/`. Module: `github.com/Ow1Dev/felter`, Go 1.24.

## Commands

```bash
# Go — binaries to build/
make fieldservice   # :8080
make userservice    # gRPC :9091 + HTTP :9090
make proxy          # :9092
make projectservice # :8081
make migrate        # apply migrations
make generate-api   # regenerate OpenAPI types for Go + TypeScript

# Web
make web            # bun run start (NOT npm)
make init           # bun install
cd web && bun run ng build --configuration development

# Go tooling
make fmt            # gofumpt -w . (NOT gofmt)
make lint           # golangci-lint run (nix develop)
make test           # go test ./... -race -count=1
```

## Environment Variables

All proxy config vars are **required** (no defaults):
- `PROXY_HTTP_ADDRESS`, `PROXY_JWT_SECRET`, `PROXY_KEYCLOAK_URL`, `PROXY_KEYCLOAK_REALM`, `PROXY_KEYCLOAK_CLIENT_ID`, `PROXY_KEYCLOAK_CLIENT_SECRET`, `PROXY_KEYCLOAK_REDIRECT_URI`, `PROXY_USERSERVICE_GRPC_ADDR`, `PROXY_FIELD_URL`, `PROXY_USERSERVICE_URL`, `PROXY_PROJECTSERVICE_URL`

fieldservice: `ADDRESS` > `PORT` > `:8080`. `CORS_ALLOWED_ORIGINS` empty = allow all.
userservice/migrate: `DATABASE_DSN` required. `GRPC_ADDRESS` default `:9091`, `HTTP_ADDRESS` default `:9090`.
projectservice: `ADDRESS` > `PORT` > `:8081`. `DATABASE_DSN` required.

`.env` at repo root is auto-loaded by Makefile.

## Auth Flow (proxy + Angular)

1. `AuthService.login()` → `window.location.href = ${identityUrl}/login` → Keycloak
2. Keycloak → redirect to `/callback?code=...`
3. `CallbackPageComponent` → POST to `/api/auth/callback` → JWT + email stored
4. `AuthService.currentUser` signal populated via `/api/auth/me`
5. On reload: `AuthService` constructor checks token, fetches `/me` if authenticated
6. Logout: `AppHeaderComponent` → fetch POST `/api/auth/logout` → 302 to Keycloak logout → redirect to `/callback` → `/login`

**Auth guard** (`authGuard`) redirects unauthenticated users to `${environment.identityUrl}/login`.

**Proxy routes**: Use `HandleProxy(targetURL, pathPrefix)` to add new proxied endpoints. Route in `main.go` with `mux.Handle("/api/field/", srv.HandleProxy(cfg.FieldURL, "/api/field"))`.

## Architecture

| Component | Notes |
|---|---|
| `cmd/proxy/` | Auth proxy. Keycloak OIDC, issues JWTs. Endpoints: `/api/auth/login`, `/callback`, `/logout`, `/me` |
| `cmd/fieldservice/` | Stateless HTTP API. `GET /api/hello` |
| `cmd/userservice/` | gRPC + HTTP. Postgres-backed. Proto in `proto/userservice/` |
| `cmd/projectservice/` | HTTP API. Postgres-backed. OpenAPI v3 spec in `docs/api/projectservice.yaml`. Generated types in `internal/projectservice/api/` and `web/src/app/api/` |
| `cmd/migrate/` | Applies `.up.sql` migrations. Per-service tracking tables |
| `web/src/app/` | Angular SPA. `app.ts` template is inline. Environments: `web/src/environments/` |

## Gotchas

- **`gofumpt` not `gofmt`**: linter requires gofumpt
- **`make web` uses bun**: not npm
- **Angular build default = production**: use `--configuration development` for dev
- **`recoverer` inlines JSON 500 as raw string**: avoids import cycle with `httputil`
- **Go linting**: `revive` `exported` rule requires doc comments on all exported symbols
- **`web/Dockerfile` is artifact-only**: final stage is `FROM scratch` with only `/dist`
- **Generated `*.pb.go`**: excluded from lint
- **Generated `*.gen.go`**: excluded from lint (OpenAPI types)

## Migrations

Each service has its own directory and tracking table (`<service>_migrations`). `make migrate` is idempotent. Naming: `{NNN}_{name}.up.sql` / `.down.sql` pairs.

## Where Things Live

- Proxy: `cmd/proxy/`, `internal/proxy/` (config, httpserver, jwt)
- fieldservice: `internal/fieldservice/httpserver/`, `internal/fieldservice/handlers/`
- userservice: `internal/userservice/grpcserver/`, `internal/userservice/httpserver/`, `internal/userservice/store/`
- Shared: `internal/httputil/`, `internal/db/`, `internal/migrate/`
- Angular: `web/src/app/`; environments: `web/src/environments/`

## CI

- `go-ci.yml`: `go mod tidy` → `golangci-lint run` → `go vet ./...` → `go test ./... -race -count=1`
- `web-ci.yml`: `bun install` → `bun run ng build --configuration production` → `bun run test`

## Nix Dev Shell

Provides: `go`, `bun`, `nodejs_20`, `golangci-lint`, `gofumpt`, `protobuf`, Angular language server. Matches CI exactly.
