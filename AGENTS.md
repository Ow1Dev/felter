Felter Repo Guidance (High-Signal Only)

## Scope

Two parts: Go API (stdlib only, no framework) and Angular 21 SPA in `web/`. Module: `github.com/Ow1Dev/felter`, Go 1.22, zero external Go dependencies (no `go.sum`).

## Entrypoints

- API: `cmd/api/main.go`. Routes: `GET /api/hello` → `{"message":"hello world"}`. Unmatched → JSON 404; panics → JSON 500.
- Web: `web/src/main.ts` bootstraps a standalone Angular app. Active root template is the **inline `template:`** in `app.ts` — `app.html` is a dead CLI scaffold file that is not bound to any component.
- `app.routes.ts` exports `Routes = []` (no routes yet). `HelloService` calls `GET {apiBaseUrl}/api/hello`; dev `apiBaseUrl = 'http://localhost:8080'`, prod `apiBaseUrl = '/api'`.

## Commands

```bash
# Dev
make dev          # API in background + Angular dev server in foreground (Ctrl-C kills only frontend)
make api          # go run ./cmd/api
make web          # cd web && bun run start  (NOT npm — Makefile uses bun directly)
make init         # cd web && bun install

# Go
make fmt          # gofumpt -w .  (NOT gofmt — uses gofumpt, stricter superset)
make fmt-check    # CI-style diff check
make tidy         # go mod tidy
make vet          # go vet ./...
make lint         # golangci-lint run  (requires Nix shell or manual install)
make test         # go test ./... -race -count=1

# Web
cd web && bun run build           # ng build (default config is PRODUCTION)
cd web && bun run ng build --configuration development  # explicit dev build
cd web && bun run vitest run      # CI-style headless tests
cd web && npx prettier --write .  # format web files

# Single test — Go
go test ./internal/handlers/... -run TestHandleHello -race
go test ./internal/httputil/... -run TestWriteJSONAndError/write_json_ok -race

# Single test — Web (Vitest)
cd web && vitest run src/app/app.spec.ts -t "should create the app"
```

## Environment Variables

| Variable | Default | Notes |
|---|---|---|
| `ADDRESS` | `:8080` | Full bind address. Overrides `PORT`. |
| `PORT` | — | Digits only; auto-prefixed with `:`. |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:4200,http://127.0.0.1:4200` | Comma-separated exact-origin allowlist. Empty list → **all origins allowed** (footgun in prod). |
| `READ_TIMEOUT` / `WRITE_TIMEOUT` / `IDLE_TIMEOUT` | `15s` / `15s` / `60s` | `time.ParseDuration` format. |
| `API_ADDR` | `:8080` | Makefile-only variable; stripped of `:` before passing as `PORT`. |

Address precedence: `ADDRESS` > `PORT` > `:8080`.

## CI

**`.github/workflows/web-ci.yml`** — triggers on `web/**` or `flake.*` changes, runs inside `nix develop`:
1. `bun install`
2. `bun run ng build --configuration production`
3. `bun run test` (= Vitest via Angular CLI)

**`.github/workflows/go-ci.yml`** — triggers on `cmd/**`, `internal/**`, `go.*`, `.golangci.yml`, `flake.*`, runs inside `nix develop`:
1. `go mod tidy`
2. `golangci-lint run`
3. `go vet ./...`
4. `go test ./... -race -count=1`

## Gotchas

- **`gofumpt` not `gofmt`:** `make fmt` applies `gofumpt -w .`. Bare `gofmt` does not satisfy the linter.
- **`make web` uses `bun`, not npm:** If Bun is absent, run npm equivalents manually; the Makefile does not fall back automatically.
- **Angular build default is production:** `angular.json` sets `"defaultConfiguration": "production"`. Bare `ng build` builds prod. Dev requires `--configuration development`.
- **`recoverer` inlines JSON 500 as a raw string literal** in `internal/httpserver/middleware.go` to avoid an import cycle with `httputil`. Do not refactor it to use `httputil.WriteError` without resolving that cycle.
- **`make dev` only kills the foreground process:** The API (`go run ./cmd/api &`) runs in background. It exits cleanly via `signal.NotifyContext`, but a crash leaves it running.
- **`web/Dockerfile` is artifact-only:** Final stage is `FROM scratch` with only `/dist`. Not a runnable image.
- **Go linting:** `.golangci.yml` enables `revive` with the `exported` rule — all exported symbols must have doc comments. `max-issues-per-linter: 0` surfaces every issue.

## Where Things Live

- API wiring: `internal/httpserver/` (server, routes, cors, middleware)
- Utilities: `internal/httputil/` (JSON helpers), `internal/config/` (env loading), `internal/handlers/`
- Angular app: `web/src/app/`; environments: `web/src/environments/`
- Agent skill packs: `.agents/skills/golang-pro/`, `.agents/skills/angular-developer/`

## Nix Dev Shell

`nix develop` provides `go`, `bun`, `nodejs_20`, `golangci-lint`, `gofumpt`, Angular language server — matches CI exactly. Optional locally if you already have the right versions.
