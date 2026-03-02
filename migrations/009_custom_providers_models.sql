-- Custom providers + models (tenant-configurable, "bring your own model")
-- Run after 008_audit_usage_prefs.sql.
--
-- Goal: allow tenants to register arbitrary providers (OpenAI-compatible endpoints)
-- and explicit model IDs, without hardcoding OpenRouter/Gemini/HuggingFace in backend logic.

-- Tenant providers: store endpoint + encrypted auth for OpenAI-compatible APIs.
CREATE TABLE IF NOT EXISTS tenant_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    provider_key TEXT NOT NULL,          -- tenant-local identifier, e.g. "openai", "together", "my-proxy"
    provider_type TEXT NOT NULL,         -- e.g. "openai_compatible"
    base_url TEXT NOT NULL,              -- e.g. https://api.openai.com
    auth_header TEXT DEFAULT 'Authorization',
    auth_scheme TEXT DEFAULT 'Bearer',   -- e.g. "Bearer", "Api-Key", or empty for raw token
    encrypted_key TEXT NOT NULL,
    key_hint TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, provider_key)
);

CREATE INDEX IF NOT EXISTS idx_tenant_providers_tenant_id ON tenant_providers(tenant_id);

ALTER TABLE tenant_providers ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Tenant manage own providers"
    ON tenant_providers FOR ALL
    USING (
        tenant_id IN (SELECT tenant_id FROM user_profiles WHERE id = auth.uid())
    )
    WITH CHECK (
        tenant_id IN (SELECT tenant_id FROM user_profiles WHERE id = auth.uid())
    );

-- Tenant models: explicit model IDs for routing (must be registered per provider).
CREATE TABLE IF NOT EXISTS tenant_models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    provider_key TEXT NOT NULL,
    model_id TEXT NOT NULL,              -- provider-native model id, e.g. "gpt-4o-mini" or "anthropic/claude-3-5-sonnet"
    display_name TEXT,
    quality_score DOUBLE PRECISION DEFAULT 0.75,
    cost_per_token DOUBLE PRECISION DEFAULT 0.0,
    context_window INTEGER DEFAULT 0,
    max_tokens INTEGER DEFAULT 0,
    tags TEXT[] DEFAULT ARRAY[]::TEXT[],
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, provider_key, model_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_models_tenant_id ON tenant_models(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_models_provider_key ON tenant_models(provider_key);

ALTER TABLE tenant_models ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Tenant manage own models"
    ON tenant_models FOR ALL
    USING (
        tenant_id IN (SELECT tenant_id FROM user_profiles WHERE id = auth.uid())
    )
    WITH CHECK (
        tenant_id IN (SELECT tenant_id FROM user_profiles WHERE id = auth.uid())
    );

