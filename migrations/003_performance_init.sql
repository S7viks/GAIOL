-- Migration 003: Model Performance Tracking
-- Stores historical performance data for dynamic learning and routing optimization

CREATE TABLE IF NOT EXISTS model_performance (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_id TEXT NOT NULL,
    task_type TEXT,
    quality_score FLOAT NOT NULL,
    latency_ms BIGINT NOT NULL,
    tokens_used INT DEFAULT 0,
    status TEXT DEFAULT 'success', -- 'success', 'error', 'timeout'
    session_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for fast lookup by model and task
CREATE INDEX IF NOT EXISTS idx_model_perf_model_task ON model_performance(model_id, task_type);

-- View for aggregate performance
CREATE OR REPLACE VIEW model_performance_agg AS
SELECT 
    model_id,
    task_type,
    AVG(quality_score) as avg_quality,
    AVG(latency_ms) as avg_latency,
    COUNT(*) as sample_count
FROM model_performance
GROUP BY model_id, task_type;
