# Local dev stack (Go + Vite dashboard)

Authoritative runbook for running the Go HTTP API (with in-process orchestrator) and the React dashboard.

## One inference path

Every chat request (`POST /api/query/smart`, `POST /v1/chat`) goes through `orchestrateUserRequest()` → in-process `orchestration.Service`. There is no separate TypeScript orchestrator process at runtime.

The `archive/orchestrator/` tree remains reference + Vitest spec until fully retired.

## Startup

### Recommended: one command

From repo root (PowerShell):

```powershell
.\scripts\dev\start-relay.ps1              # Go API
.\scripts\dev\start-relay.ps1 -Dashboard  # + Vite on :5173
```

### Manual

1. **Repository root** — Ensure `.env` exists if needed; `cmd/web-server/main.go` loads `.env` via `godotenv`.
2. **Go web server** — `go run ./cmd/web-server` (default port **8080**).
3. **Vite dashboard (optional)** — `cd dashboard && npm install && npm run dev` (default **5173**).

**URLs:** Go `http://localhost:8080`, dashboard UI `http://localhost:5173/` when using Vite.

## Processes and ports

| Service | Default URL | Start command |
|--------|-------------|---------------|
| Go API (+ orchestrator) | `http://localhost:8080` | `go run ./cmd/web-server` |
| Vite dashboard | `http://localhost:5173/` | `cd dashboard && npm run dev` |

## Dashboard → backend

- Vite proxies **`/api`**, **`/v1`**, and **`/health`** to **`http://localhost:8080`** (`dashboard/vite.config.ts`).
- Orchestration endpoints: `POST /v1/orchestrate`, `GET /api/orchestration/traces/:id`, trust, trace-ids, eval — all served by Go.

## Environment variables

| Variable | Purpose |
|----------|---------|
| `PORT` | Go listen port (default `8080`). |
| `GAIOL_DISABLE_AUTH` | Local no-auth mode (aliases: `GAIOL_AUTH_DISABLED`, `DISABLE_AUTH`). |
| `GAIOL_TS_BEAM_WIDTH` / `GAIOL_BEAM_WIDTH` | Beam width (default `2`). |
| `GAIOL_TS_CONSENSUS_MODE` / `GAIOL_CONSENSUS_MODE` | `abtc`, `uniform`, or `static`. |
| `GAIOL_TS_DOMAIN` / `GAIOL_DOMAIN` | Domain tag (default `general`). |
| `GAIOL_TS_EXPLORE_PATHS` / `GAIOL_EXPLORE_PATHS` | Default on; `0`/`false` disables path exploration. |
| `ALLOWED_ORIGINS` | CORS allowlist for Go. |
| Provider keys in `.env` | Used in no-auth mode only (`OPENROUTER_API_KEY`, etc.). |

## Health checks

| URL | Notes |
|-----|-------|
| `GET http://localhost:8080/health` | `orchestrator_configured` and `orchestrator_reachable` reflect in-process Go orchestrator. |
| `GET http://localhost:8080/api/setup/status` | Tenant/orchestrator readiness for dashboard onboarding. |

See [API.md](../API.md) and [FEATURE-FLAGS.md](FEATURE-FLAGS.md).
