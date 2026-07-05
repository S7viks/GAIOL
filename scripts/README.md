# Scripts

Run all scripts from the **repository root** (e.g. `.\scripts\test\integration.ps1` from PowerShell).

## Test scripts (`scripts/test/`)

| Script | Description |
|--------|-------------|
| `integration.ps1` | Integration tests: health, public/protected routes, CORS, v1/chat 401. Requires server on http://localhost:8080. Skips JWT/401 block when health reports `auth_disabled` or `database.connected: false`. |
| `pipeline.ps1` | Simple pipeline test (`/api/query/smart`). |
| `ollama.ps1` | Ollama availability + GAIOL server + query test. |
| `go-orchestrator-integration.ps1` | Go orchestrator integration tests + optional benchmark. |

## Dev scripts (`scripts/dev/`)

| Script | Description |
|--------|-------------|
| `clean-start.ps1` | Stop any running server, remove *.exe, build, then run web-server.exe. Run from repo root. |
| `start-relay.ps1` | Start Go API (+ optional Vite dashboard). |
| `stack-local.ps1` | Prints commands to run Go and Vite dashboard (does not start them). |

## Start/stop (root)

Start and stop the server from the repo root:

- `start.ps1`, `start.sh`, `start.bat` — start server
- `stop.ps1`, `stop.bat` — stop server
