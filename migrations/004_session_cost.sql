-- Migration to add total_cost to reasoning_sessions
ALTER TABLE reasoning_sessions ADD COLUMN IF NOT EXISTS total_cost DECIMAL(10, 6) DEFAULT 0.0;
