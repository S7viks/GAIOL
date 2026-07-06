-- Per-tenant orchestration tuning (overrides server GAIOL_* env defaults when set).
-- Run after 008_audit_usage_prefs.sql.

ALTER TABLE tenant_settings
    ADD COLUMN IF NOT EXISTS beam_width INTEGER,
    ADD COLUMN IF NOT EXISTS consensus_mode TEXT,
    ADD COLUMN IF NOT EXISTS domain TEXT,
    ADD COLUMN IF NOT EXISTS explore_paths BOOLEAN;
