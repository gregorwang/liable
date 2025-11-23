-- Migration: 005_unified_queue_stats.sql
-- Description: Create unified queue stats view to replace the dual queue system
-- Date: 2025-11-22
-- Resolves: BUSINESS_LOGIC_REVIEW.md Section 2 - Queue Management Architecture

-- Create unified queue statistics view that aggregates all review queue types
CREATE OR REPLACE VIEW unified_queue_stats AS
-- Comment first review queue
SELECT
    'comment_first_review'::text AS queue_name,
    '评论一审队列'::text AS description,
    100 AS priority,
    COUNT(*) AS total_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) AS completed_tasks,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) AS pending_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    AVG(
        CASE
            WHEN status = 'completed' AND completed_at IS NOT NULL AND claimed_at IS NOT NULL
            THEN EXTRACT(EPOCH FROM (completed_at - claimed_at)) / 60.0
            ELSE NULL
        END
    ) AS avg_process_time_minutes,
    true AS is_active,
    MIN(created_at) AS created_at,
    MAX(COALESCE(completed_at, claimed_at, created_at)) AS updated_at
FROM review_tasks

UNION ALL

-- Comment second review queue
SELECT
    'comment_second_review'::text AS queue_name,
    '评论二审队列'::text AS description,
    90 AS priority,
    COUNT(*) AS total_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) AS completed_tasks,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) AS pending_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    AVG(
        CASE
            WHEN status = 'completed' AND completed_at IS NOT NULL AND claimed_at IS NOT NULL
            THEN EXTRACT(EPOCH FROM (completed_at - claimed_at)) / 60.0
            ELSE NULL
        END
    ) AS avg_process_time_minutes,
    true AS is_active,
    MIN(created_at) AS created_at,
    MAX(COALESCE(completed_at, claimed_at, created_at)) AS updated_at
FROM second_review_tasks

UNION ALL

-- Quality check queue
SELECT
    'quality_check'::text AS queue_name,
    '质量检查队列'::text AS description,
    80 AS priority,
    COUNT(*) AS total_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) AS completed_tasks,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) AS pending_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    AVG(
        CASE
            WHEN status = 'completed' AND completed_at IS NOT NULL AND claimed_at IS NOT NULL
            THEN EXTRACT(EPOCH FROM (completed_at - claimed_at)) / 60.0
            ELSE NULL
        END
    ) AS avg_process_time_minutes,
    true AS is_active,
    MIN(created_at) AS created_at,
    MAX(COALESCE(completed_at, claimed_at, created_at)) AS updated_at
FROM quality_check_tasks

UNION ALL

-- Video first review queue
SELECT
    'video_first_review'::text AS queue_name,
    '视频一审队列'::text AS description,
    70 AS priority,
    COUNT(*) AS total_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) AS completed_tasks,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) AS pending_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    AVG(
        CASE
            WHEN status = 'completed' AND completed_at IS NOT NULL AND claimed_at IS NOT NULL
            THEN EXTRACT(EPOCH FROM (completed_at - claimed_at)) / 60.0
            ELSE NULL
        END
    ) AS avg_process_time_minutes,
    true AS is_active,
    MIN(created_at) AS created_at,
    MAX(COALESCE(completed_at, claimed_at, created_at)) AS updated_at
FROM video_first_review_tasks

UNION ALL

-- Video second review queue
SELECT
    'video_second_review'::text AS queue_name,
    '视频二审队列'::text AS description,
    60 AS priority,
    COUNT(*) AS total_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) AS completed_tasks,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) AS pending_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) AS in_progress_tasks,
    AVG(
        CASE
            WHEN status = 'completed' AND completed_at IS NOT NULL AND claimed_at IS NOT NULL
            THEN EXTRACT(EPOCH FROM (completed_at - claimed_at)) / 60.0
            ELSE NULL
        END
    ) AS avg_process_time_minutes,
    true AS is_active,
    MIN(created_at) AS created_at,
    MAX(COALESCE(completed_at, claimed_at, created_at)) AS updated_at
FROM video_second_review_tasks

ORDER BY priority DESC;

-- Add comment explaining the view
COMMENT ON VIEW unified_queue_stats IS 'Real-time queue statistics aggregated from all review task tables. This view replaces the manual task_queues table to provide accurate, up-to-date queue metrics.';

-- Note: The old queue_stats view and task_queues/task_queue tables should be considered deprecated
-- but are kept for backward compatibility during the transition period.
