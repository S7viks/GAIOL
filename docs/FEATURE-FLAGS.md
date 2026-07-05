# Feature flags and environment variables

Values that affect routing, orchestration, and dashboard behavior. For local setup see [LOCAL-DEV-STACK.md](LOCAL-DEV-STACK.md).

## Go web server

| Variable | Effect |
|----------|--------|
| `GAIOL_DISABLE_AUTH` / `GAIOL_AUTH_DISABLED` / `DISABLE_AUTH` | No JWT; open `/api/*` stubs for local dev. |
| `GAIOL_TS_BEAM_WIDTH` / `GAIOL_BEAM_WIDTH` | Beam width for orchestration (default 2). |
| `GAIOL_TS_CONSENSUS_MODE` / `GAIOL_CONSENSUS_MODE` | `abtc`, `uniform`, or `static`. |
| `GAIOL_TS_DOMAIN` / `GAIOL_DOMAIN` | Domain tag (default `general`). |
| `GAIOL_TS_EXPLORE_PATHS` / `GAIOL_EXPLORE_PATHS` | Set `0` to disable explore paths. |
| `ALLOWED_ORIGINS` | CORS allowlist for browser calls to Go. |
| `PORT` | Listen port (default 8080). |

Orchestration runs **in-process** in the Go web server (`internal/orchestration/`). No separate Node orchestrator is required at runtime.

## Dashboard (Vite)

| Variable | Effect |
|----------|--------|
| `VITE_API_BASE` | Optional absolute API base; leave empty to use dev proxy paths `/api`, `/health`. |

## Browser noise (not errors)

Chrome DevTools may request `GET /.well-known/appspecific/com.chrome.devtools.json`. That file is optional; **404 is normal** and is no longer logged at info level.

## Health / degradation

- If the Go server fails to initialize the orchestrator: trace, trust, trace-ids, and eval proxies return **503** `orchestrator_disabled`.
- User chat uses the in-process orchestrator; long-running requests may return **504** `orchestrator_timeout`.
