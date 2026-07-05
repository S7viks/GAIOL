# Go orchestration (runtime)

Orchestration runs **in-process** in the Go web server. There is no separate Node orchestrator at runtime.

## Package layout

| Path | Role |
|------|------|
| `internal/orchestration/service.go` | Entry: `Orchestrate()` v1 contract |
| `internal/orchestration/pipeline.go` | Decompose → route → parallel → beam → consensus → trust |
| `internal/orchestration/registry.go` | Env + tenant credential runtimes |
| `internal/orchestration/llm/` | Provider bridge to `internal/models` |
| `internal/gaiol/orchestratorcontract/v1/` | Wire types, validation, JSON schemas |

## HTTP routes (Go)

- `POST /v1/orchestrate` — direct orchestrate API
- `POST /api/query/smart`, `POST /v1/chat` — user chat (via `orchestrate_user.go`)
- `GET /api/orchestration/traces/:id`, `/trust`, `/trace-ids`, `POST /eval/contains`
- `GET /v1/trust`, `GET /v1/traces`, `GET /v1/traces/:id` — eval script aliases

## Environment

| Variable | Default | Effect |
|----------|---------|--------|
| `GAIOL_TS_BEAM_WIDTH` / `GAIOL_BEAM_WIDTH` | 2 | Beam width |
| `GAIOL_TS_CONSENSUS_MODE` / `GAIOL_CONSENSUS_MODE` | abtc | Consensus mode |
| `GAIOL_TS_DOMAIN` / `GAIOL_DOMAIN` | general | Trust domain tag |
| `GAIOL_TS_EXPLORE_PATHS` / `GAIOL_EXPLORE_PATHS` | on | Path exploration |

Wire request fields: `consensus_mode`, `beam_width`, `explore_paths`, `abtc_decay`, `constraints.*`.

## Parity vs TypeScript spec

| Area | Status |
|------|--------|
| Heuristic decompose, routing, beam, binary ABTC trust | Ported |
| Mock pipeline golden tests | `internal/orchestration/golden_test.go` |
| ABTC quality/agreement | **Jaccard/heuristic** (TS uses embeddings when available) |
The `archive/orchestrator/` tree is retained as reference + Vitest spec only (not runtime).

## Tests

```bash
go test ./internal/orchestration/...
go test ./internal/integration -tags=integration
```

## Local dev

```powershell
.\scripts\dev\start-relay.ps1 -Dashboard
# or
go run ./cmd/web-server
```
