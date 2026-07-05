# Golden orchestration fixtures

Structural expectations for the Go mock beam pipeline (`golden_test.go`).

## Files

- `golden_beam_expectations.json` — invariants (subtask count, path candidates, routed models, trust updates).

## Refreshing

After intentional pipeline changes, run:

```bash
go test ./internal/orchestration -run TestGolden -v
```

If invariants change deliberately, edit `golden_beam_expectations.json`.

## TS parity capture (optional)

To freeze TypeScript output before archiving `orchestrator/`:

```powershell
cd orchestrator
npm run build
# Run pipeline test with mocks and save JSON to testdata/golden_trace_ts.json
```

Go golden tests assert structural parity, not byte-identical LLM text.
