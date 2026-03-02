-- Audit log, per-key usage, and tenant settings (budget + model preferences)
-- Run after 007_api_keys_multitenant.sql.

-- Audit log: key actions and login for activity feed
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    user_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_tenant_id ON audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at DESC);

ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Tenant read own audit log"
    ON audit_log FOR SELECT
    USING (
        tenant_id IN (SELECT tenant_id FROM user_profiles WHERE id = auth.uid())
    );

CREATE POLICY "Tenant insert own audit log"
    ON audit_log FOR INSERT
    WITH CHECK (
        tenant_id IN (SELECT tenant_id FROM user_profiles WHERE id = auth.uid())
    );

-- Per-GAIOL-key usage: link api_queries to key when request used a GAIOL key
ALTER TABLE api_queries
    ADD COLUMN IF NOT EXISTS gaiol_api_key_id UUID REFERENCES gaiol_api_keys(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_api_queries_gaiol_key_id ON api_queries(gaiol_api_key_id) WHERE gaiol_api_key_id IS NOT NULL;

-- Tenant settings: budget and model preferences
CREATE TABLE IF NOT EXISTS tenant_settings (
    tenant_id UUID NOT NULL PRIMARY KEY,
    budget_limit DECIMAL(12, 4),
    budget_alert_sent_at TIMESTAMPTZ,
    default_model_id TEXT,
    strategy TEXT DEFAULT 'balanced',
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE tenant_settings ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Tenant manage own settings"
    ON tenant_settings FOR ALL
    USING (
        tenant_id IN (SELECT tenant_id FROM user_profiles WHERE id = auth.uid())
    )
    WITH CHECK (
        tenant_id IN (SELECT tenant_id FROM user_profiles WHERE id = auth.uid())
    );
