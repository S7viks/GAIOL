# Demo script: Chat → Trace → Trust / Models

Prerequisites: run the unified stack from repo root:

```powershell
.\scripts\dev\start-relay.ps1 -Dashboard
```

Or `docker compose up`. See [LOCAL-DEV-STACK.md](LOCAL-DEV-STACK.md).

Automated Phase 2 checks (no-auth):

```powershell
.\scripts\demo\e2e-phase2.ps1
```

---

## Dashboard demo (local)

1. Open `http://localhost:5173/chat`.
2. Confirm the top bar shows a green connectivity dot (Go `/health` OK).
3. Enter a short prompt; keep the default strategy (e.g. `balanced`). Send.
4. When the response returns, click the **trace id** link (or copy it).
5. On Trace viewer, expand **metrics_summary** and **timeline_rebuilt**.
6. Open **Trust**; refresh—trust rows appear after ABTC runs with model participation.
7. Open **Models**; search for a model id seen in the trace or trust table.
8. Open **Home** — usage summary cards (MTD spend, requests) when auth + DB are enabled.
9. Open **History** — structured usage tables and CSV export.

Optional: **Metrics** → pick a recent id from the chip list → **Load metrics** → link to full trace.

---

## Phase 2 E2E — auth-enabled manual checklist

Run with Supabase auth (no `GAIOL_DISABLE_AUTH`). TS orchestrator env provider keys should be **empty**; tenant keys come from the dashboard only.

| Step | Action | Pass criteria |
|------|--------|---------------|
| 1 | Sign up at `/signup` | Account created, redirected to app |
| 2 | Onboarding step 1: add **OpenAI** and **Claude** (or two providers) | Both appear in Settings provider keys |
| 3 | Onboarding step 2: create/copy **GAIOL API key** | Copy-once secret shown |
| 4 | Onboarding step 3: run demo chat | Answer + trace id |
| 5 | Dashboard **Chat** → send message | Answer + trace link |
| 6 | `curl -X POST http://localhost:8080/v1/chat` with GAIOL key | Same orchestration metadata shape |
| 7 | Stop TS orchestrator | Chat returns **503** everywhere (no Go fallback) |
| 8 | Confirm no `/reasoning` route in app | Only single chat path |

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer YOUR_GAIOL_KEY" \
  -H "Content-Type: application/json" \
  -d '{"prompt":"Hello from curl","max_tokens":200}'
```

---

## Phase 3 verification (Settings depth + usage)

1. **Settings → Tenant models**: default model listed after provider key save; add/remove manual model works.
2. **Home**: MTD spend and request count cards populate after chat (auth mode).
3. **History**: usage by day/provider tables; **Export CSV** downloads.
4. **Settings → Budget**: set monthly limit; Home shows budget vs spend when set.
5. Revoke all GAIOL keys → Home/Settings show setup banner → **Complete setup** links to onboarding.
