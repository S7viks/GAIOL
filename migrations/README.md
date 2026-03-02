# Database migrations

Migrations are run manually in the Supabase SQL Editor (Dashboard > SQL Editor). Run in the order below.

## Core app (auth, tenant, keys, usage, audit)

Required for the web app, dashboard, GAIOL keys, and usage/billing:

1. **001** — Either run `001_initial_schema.sql` in one go, or if you hit connection timeouts run the four chunk files in order:
   - `chunks/001_part1_tables.sql`
   - `chunks/001_part2_indexes_rls.sql`
   - `chunks/001_part3_policies.sql`
   - `chunks/001_part4_trigger.sql`
2. **007** — `007_api_keys_multitenant.sql` (provider keys, GAIOL keys; requires 001).
3. **008** — `008_audit_usage_prefs.sql` (audit log, usage prefs, tenant settings; requires 007).

## Optional feature migrations

For RAG, reasoning engine, performance tracking, session cost, and world model, run after 001 (or after 008 if you already ran core):

- `002_rag_init.sql`
- `002_reasoning_tables.sql`
- `003_performance_init.sql`
- `004_session_cost.sql`
- `006_world_model.sql`

Dependencies between these are documented in each file’s header. Run in numeric order when in doubt.

## Reference

- Full setup: [Database setup](../docs/database-setup.md)
- Ops: [Runbook](../docs/RUNBOOK.md)
