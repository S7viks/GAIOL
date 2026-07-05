# Archived TypeScript orchestrator (spec only)

**Runtime retired:** orchestration runs in Go (`internal/orchestration/`).

This tree is kept for:

- Historical reference and paper algorithm documentation
- Vitest unit tests (`npm test` in this directory)
- Benchmark helpers imported by `scripts/benchmark/run_benchmark.ts`

## Do not use at runtime

Do not run `npm run dev:api` for GAIOL product development. Use:

```powershell
go run ./cmd/web-server
# or
.\scripts\dev\start-relay.ps1 -Dashboard
```

## Canonical contract schemas

JSON schemas for the v1 wire contract live in:

`internal/gaiol/orchestratorcontract/v1/schemas/`

Copies under `archive/orchestrator/contract/schemas/` are retained for reference only.

## Tests

```bash
cd archive/orchestrator
npm ci
npm test
```
