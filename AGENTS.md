Felter Repo Guidance (High-Signal Only)

Scope

- Two parts: Go API (port defaults to 8080) and Angular web app in `web/` (Angular 21, uses Bun by default).
- Dev shell via Nix flake provides Go, Bun, Node 20, Angular tools. CI uses it; local dev can too but is optional if you already have Go and Node/Bun.

Entrypoints

- API: `./cmd/api` (Go). Route: `GET /api/hello` returns `{ "message": "hello world" }`. Unmatched → JSON 404; panics → JSON 500.
- Web: `web/` Angular app. Development server at `http://localhost:4200`.

Run Commands

- All-in-one dev: `make dev` (starts API and Angular). Honors `API_ADDR` (default `:8080`).
- API only: `make api` (uses `PORT=${API_ADDR#:}` internally). Or `PORT=8081 go run ./cmd/api`.
- Web only: `make web` (equivalent to `cd web && npm run start`, which uses Angular CLI; Bun present but `make` calls npm scripts).

Environment and CORS

- API bind address precedence: `ADDRESS` (e.g., `:8080`) overrides `PORT`. If `PORT` is digits, it is auto-prefixed with `:`. Default `:8080`.
- CORS allowlist via `CORS_ALLOWED_ORIGINS` (comma-separated). Defaults allow `http://localhost:4200` and `http://127.0.0.1:4200` in dev.
- Timeouts (Go `time.ParseDuration`): `READ_TIMEOUT`, `WRITE_TIMEOUT`, `IDLE_TIMEOUT`.

Nix Dev Shell (optional but matches CI)

- Enter: `nix develop` then run normal commands. Provides `go`, `bun`, `nodejs_20`, Angular language server, etc. Useful for consistent versions.
- `web/package.json` sets `"packageManager": "bun@1.3.11"`. Prefer `bun install` for speed; `npm ci` is the fallback pattern used in CI.

CI Behavior (what must pass)

- Workflow: `.github/workflows/web-ci.yml` runs only when files in `web/**` or flake files change.
- Steps inside Nix shell:
  1. `bun install || npm ci`
  2. `bun run ng build --configuration production || npx ng build --configuration production`
  3. `bun run vitest run || npx vitest run`
- Implication: keep `web` build green and unit tests passing; API has no CI checks here.

Testing

- Web unit tests: `cd web && ng test` (dev) or `vitest run` (CI style). Single test example: `vitest run path/to/file.spec.ts -t "name"`.
- No Go tests currently present.

Conventions and Gotchas

- Angular dev server assumes API at `:8080` with CORS allowing `localhost:4200`. If you change the API port, update `API_ADDR` (for Makefile) or set `PORT`/`ADDRESS` and ensure CORS origins include the web origin.
- `make dev` backgrounds the API and then runs the web dev server in the foreground. Stop with Ctrl-C; API shuts down via Go’s graceful shutdown.
- If Bun is unavailable locally, use npm equivalents in `web/`. The repo and CI tolerate both.

Formatting and Maintenance

- Go helpers: `make fmt`, `make tidy`.
- Web formatting uses Prettier (devDependency). Run `npx prettier --write .` inside `web/` if needed.

Where Things Live

- API config/middleware/routing: `internal/*` and `cmd/api`.
- Web Angular app: `web/src/**`; Angular config: `web/angular.json`.

Quick Verification Flow (local)

1. API: `PORT=8080 go run ./cmd/api` then curl `http://localhost:8080/api/hello`.
2. Web: `cd web && bun install || npm ci && npm run start` then open `http://localhost:4200/`.
3. Tests (web): `cd web && bun run vitest run || npx vitest run`.
