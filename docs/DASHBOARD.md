# Dashboard (Vite) and API map

The React app in `dashboard/` uses Vite **`base: '/'`**. Production: run `npm run build` in `dashboard/` so `dashboard/dist/` exists; Go serves the unified SPA at **`/`** (hashed assets under **`/assets/`**).

- **Dev:** `npm run dev` then open **`http://localhost:5173/`** (API proxied to Go on `:8080`).
- **Behind Go:** open **`http://localhost:8080/`** after building `dashboard/dist`.

| Page | Route | API calls |
|------|-------|-----------|
| Landing | `/` | (marketing) |
| Sign in / Sign up | `/login`, `/signup` | `POST /api/auth/signin`, `signup`, `recover`, `update-password` |
| Terms | `/terms` | — |
| Chat | `/chat` | `POST /api/query/smart` |
| Trace viewer | `/trace`, `/trace/:id` | `GET /api/orchestration/traces/:id` |
| Trust | `/trust` | `GET /api/orchestration/trust?domain=` |
| Models | `/models` | `GET /api/models` (optional `?q=` from Trust links) |
| Metrics | `/metrics` | `GET /api/orchestration/trace-ids`, `GET /api/orchestration/traces/:id` |
| Home | `/home` | `GET /api/settings/preferences`, `GET /api/billing/summary`, `GET /api/usage`, `GET /api/setup/status` |
| History | `/history` | `GET /api/usage`, `GET /api/usage/export`, `GET /api/billing/history`, `GET /api/activity`; local chat history from Zustand |
| Settings | `/settings` | `GET/PUT /api/settings/preferences`, `GET/POST/DELETE /api/settings/models`, `GET/POST/DELETE /api/settings/provider-keys`, `GET/POST/DELETE /api/gaiol-keys`, `GET /api/setup/status` |
| Onboarding | `/onboarding` | Provider keys, GAIOL key ensure, demo `POST /api/query/smart`, optional ABTC |
| Eval | `/eval` | `POST /api/orchestration/eval/contains` |
| Health (shell) | — | `GET /health` (connectivity dot) |

Orchestration runs in-process on the Go API. See [GO-ORCHESTRATION.md](GO-ORCHESTRATION.md) and [FEATURE-FLAGS.md](FEATURE-FLAGS.md).

**Split hosting (e.g. Vercel):** set **`VITE_API_BASE`** at build time to your Go API origin (no trailing slash). See [RUNBOOK.md](RUNBOOK.md).
