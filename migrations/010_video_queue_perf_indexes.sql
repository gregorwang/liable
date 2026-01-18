-- Video queue performance indexes
-- Migration: 010_video_queue_perf_indexes
-- Date: 2026-01-17

CREATE INDEX IF NOT EXISTS idx_video_queue_tasks_pool_status_created_at
ON video_queue_tasks(pool, status, created_at);

CREATE INDEX IF NOT EXISTS idx_video_queue_tasks_pool_status_claimed_at
ON video_queue_tasks(pool, status, claimed_at);

CREATE INDEX IF NOT EXISTS idx_video_queue_tasks_pool_reviewer_status_claimed_at
ON video_queue_tasks(pool, reviewer_id, status, claimed_at);
