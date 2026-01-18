-- Daily Statistics Tables Migration
-- Migration: 008_daily_stats_tables
-- Description: Create tables for daily aggregated statistics
-- Date: 2026-01-15

-- Daily review statistics table
CREATE TABLE IF NOT EXISTS daily_review_stats (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    stats_json JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index for date lookup
CREATE INDEX IF NOT EXISTS idx_daily_review_stats_date ON daily_review_stats(date);
-- Create index for created_at (for cleanup queries)
CREATE INDEX IF NOT EXISTS idx_daily_review_stats_created_at ON daily_review_stats(created_at);

-- Daily video statistics table
CREATE TABLE IF NOT EXISTS daily_video_stats (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    stats_json JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index for date lookup
CREATE INDEX IF NOT EXISTS idx_daily_video_stats_date ON daily_video_stats(date);
-- Create index for created_at (for cleanup queries)
CREATE INDEX IF NOT EXISTS idx_daily_video_stats_created_at ON daily_video_stats(created_at);

-- Add comment to document the purpose
COMMENT ON TABLE daily_review_stats IS 'Stores aggregated daily review statistics from Redis cache';
COMMENT ON TABLE daily_video_stats IS 'Stores aggregated daily video review statistics from Redis cache';
