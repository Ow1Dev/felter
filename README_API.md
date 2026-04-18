Dev

- Run both API and Angular: `make dev`
- API only: `make api` (listens on `:8080`, override `API_ADDR`)
- Web only: `make web`

Environment

- API address: `PORT` (e.g., `8080`) or `ADDRESS` (e.g., `:8080`)
- CORS origins: `CORS_ALLOWED_ORIGINS` comma-separated (dev default allows localhost:4200)

Endpoint

- GET `/api/hello` → `{ "message": "hello world" }`
- Not found → `{ "error": "not found" }` with 404
- Panic/internal → `{ "error": "internal server error" }` with 500
