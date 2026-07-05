# TypeScript orchestrator (reference / spec)

**Runtime status:** Retired. Production and local dev use the **Go orchestrator** in `internal/orchestration/`, wired through `cmd/web-server`.

This tree is kept for:

- Contract and behavior reference while the Go port is validated
- Unit tests (`npm test` in this directory)
- Golden parity checks against Go (when added)

Do **not** start `npm run dev:api` for normal GAIOL development. Use `go run ./cmd/web-server` or `.\scripts\dev\start-relay.ps1` from the repo root.

JSON schemas now live in `internal/gaiol/orchestratorcontract/v1/schemas/` (canonical for Go runtime).
