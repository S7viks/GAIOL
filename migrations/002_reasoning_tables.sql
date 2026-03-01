-- GAIOL Reasoning Persistence Schema
-- This adds support for storing multi-step reasoning traces and multi-path beam search results

-- Reasoning Sessions Table
CREATE TABLE IF NOT EXISTS reasoning_sessions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    prompt TEXT NOT NULL,
    status TEXT DEFAULT 'pending', -- pending, processing, completed, error
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Reasoning Steps Table
CREATE TABLE IF NOT EXISTS reasoning_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES reasoning_sessions(id) ON DELETE CASCADE,
    step_index INTEGER NOT NULL,
    title TEXT NOT NULL,
    objective TEXT NOT NULL,
    status TEXT DEFAULT 'pending', -- pending, processing, completed, error
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    UNIQUE(session_id, step_index)
);

-- Reasoning Outputs Table
CREATE TABLE IF NOT EXISTS reasoning_outputs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    step_id UUID REFERENCES reasoning_steps(id) ON DELETE CASCADE,
    session_id UUID REFERENCES reasoning_sessions(id) ON DELETE CASCADE,
    model_id TEXT NOT NULL,
    model_name TEXT,
    response TEXT NOT NULL,
    scores JSONB DEFAULT '{}',
    cost DECIMAL(10, 6),
    tokens_used INTEGER,
    latency_ms INTEGER,
    is_refined BOOLEAN DEFAULT false,
    is_selected BOOLEAN DEFAULT false,
    path_index INTEGER DEFAULT 0, -- Used for Beam Search tracking to identify which path this belongs to
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_reasoning_steps_session_id ON reasoning_steps(session_id);
CREATE INDEX IF NOT EXISTS idx_reasoning_outputs_step_id ON reasoning_outputs(step_id);
CREATE INDEX IF NOT EXISTS idx_reasoning_outputs_session_id ON reasoning_outputs(session_id);

-- Enable RLS
ALTER TABLE reasoning_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE reasoning_steps ENABLE ROW LEVEL SECURITY;
ALTER TABLE reasoning_outputs ENABLE ROW LEVEL SECURITY;

-- Policies for reasoning_sessions
CREATE POLICY "Users can view own reasoning sessions"
    ON reasoning_sessions FOR SELECT
    USING (auth.uid() = user_id OR tenant_id IN (SELECT tenant_id FROM user_profiles WHERE id = auth.uid()));

CREATE POLICY "Users can insert own reasoning sessions"
    ON reasoning_sessions FOR INSERT
    WITH CHECK (auth.uid() = user_id OR tenant_id IN (SELECT tenant_id FROM user_profiles WHERE id = auth.uid()));

-- Policies for reasoning_steps (inherited from session)
CREATE POLICY "Users can view own reasoning steps"
    ON reasoning_steps FOR SELECT
    USING (session_id IN (SELECT id FROM reasoning_sessions));

CREATE POLICY "Users can insert own reasoning steps"
    ON reasoning_steps FOR INSERT
    WITH CHECK (session_id IN (SELECT id FROM reasoning_sessions));

-- Policies for reasoning_outputs (inherited from session)
CREATE POLICY "Users can view own reasoning outputs"
    ON reasoning_outputs FOR SELECT
    USING (session_id IN (SELECT id FROM reasoning_sessions));

CREATE POLICY "Users can insert own reasoning outputs"
    ON reasoning_outputs FOR INSERT
    WITH CHECK (session_id IN (SELECT id FROM reasoning_sessions));
