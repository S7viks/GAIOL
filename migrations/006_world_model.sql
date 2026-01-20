-- World Model Facts Table
CREATE TABLE IF NOT EXISTS world_model_facts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key TEXT UNIQUE NOT NULL,
    value TEXT NOT NULL,
    source TEXT, -- Which agent/role stored this
    session_id TEXT, -- Session where it was learned
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast key lookups
CREATE INDEX idx_world_model_key ON world_model_facts(key);

-- Index for session tracking
CREATE INDEX idx_world_model_session ON world_model_facts(session_id);

-- Full-text search on keys and values
CREATE INDEX idx_world_model_search ON world_model_facts USING gin(to_tsvector('english', key || ' ' || value));
