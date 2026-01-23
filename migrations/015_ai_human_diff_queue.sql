-- Migration: 015_ai_human_diff_queue.sql
-- Description: Add AI vs human diff queue for mismatched review decisions
-- Date: 2026-01-22

-- 1. AI vs human diff tasks
CREATE TABLE IF NOT EXISTS ai_human_diff_tasks (
    id SERIAL PRIMARY KEY,
    review_task_id INTEGER NOT NULL REFERENCES review_tasks(id),
    comment_id BIGINT NOT NULL REFERENCES comment(id),
    review_result_id INTEGER NOT NULL REFERENCES review_results(id),
    ai_review_result_id INTEGER NOT NULL REFERENCES ai_review_results(id),
    reviewer_id INTEGER REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'in_progress', 'completed')),
    claimed_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (review_task_id)
);

CREATE INDEX IF NOT EXISTS idx_ai_human_diff_tasks_status ON ai_human_diff_tasks(status);
CREATE INDEX IF NOT EXISTS idx_ai_human_diff_tasks_reviewer ON ai_human_diff_tasks(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_ai_human_diff_tasks_review_task_id ON ai_human_diff_tasks(review_task_id);
CREATE INDEX IF NOT EXISTS idx_ai_human_diff_tasks_review_result_id ON ai_human_diff_tasks(review_result_id);
CREATE INDEX IF NOT EXISTS idx_ai_human_diff_tasks_ai_review_result_id ON ai_human_diff_tasks(ai_review_result_id);

-- 2. AI vs human diff results (final decision)
CREATE TABLE IF NOT EXISTS ai_human_diff_results (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL REFERENCES ai_human_diff_tasks(id) ON DELETE CASCADE,
    reviewer_id INTEGER NOT NULL REFERENCES users(id),
    is_approved BOOLEAN NOT NULL,
    tags TEXT[] DEFAULT '{}',
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (task_id)
);

CREATE INDEX IF NOT EXISTS idx_ai_human_diff_results_task_id ON ai_human_diff_results(task_id);

-- 3. Backfill mismatch tasks (latest AI result per review task)
INSERT INTO ai_human_diff_tasks (
    review_task_id,
    comment_id,
    review_result_id,
    ai_review_result_id,
    status,
    created_at,
    updated_at
)
SELECT
    latest.review_task_id,
    latest.comment_id,
    latest.review_result_id,
    latest.ai_review_result_id,
    'pending',
    NOW(),
    NOW()
FROM (
    SELECT DISTINCT ON (art.review_task_id)
        art.review_task_id,
        art.comment_id,
        rr.id AS review_result_id,
        ar.id AS ai_review_result_id,
        rr.is_approved AS human_approved,
        ar.is_approved AS ai_approved,
        ar.created_at
    FROM ai_review_tasks art
    JOIN ai_review_results ar ON ar.task_id = art.id
    JOIN review_results rr ON rr.task_id = art.review_task_id
    ORDER BY art.review_task_id, ar.created_at DESC
) latest
WHERE latest.human_approved <> latest.ai_approved
ON CONFLICT (review_task_id) DO NOTHING;

-- 4. Insert permissions
INSERT INTO permissions (permission_key, name, description, resource, action, category, is_active) VALUES
    ('tasks:ai-human-diff:claim', '领取AI与人工diff任务', '允许领取AI与人工diff队列任务', 'tasks', 'create', '审核任务-AI人工差异队列', true),
    ('tasks:ai-human-diff:submit', '提交AI与人工diff结果', '允许提交AI与人工diff队列结果', 'tasks', 'update', '审核任务-AI人工差异队列', true),
    ('tasks:ai-human-diff:return', '归还AI与人工diff任务', '允许归还AI与人工diff队列任务', 'tasks', 'update', '审核任务-AI人工差异队列', true)
ON CONFLICT (permission_key) DO NOTHING;

-- 5. Grant permissions to admin and reviewer users
INSERT INTO user_permissions (user_id, permission_key, granted_by)
SELECT u.id, p.permission_key, u.id
FROM users u
JOIN permissions p ON p.permission_key IN (
    'tasks:ai-human-diff:claim',
    'tasks:ai-human-diff:submit',
    'tasks:ai-human-diff:return'
)
WHERE u.role IN ('admin', 'reviewer')
ON CONFLICT (user_id, permission_key) DO NOTHING;

-- 6. Update unified queue stats view
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
    COALESCE(MIN(created_at)::timestamp with time zone, CURRENT_TIMESTAMP) AS created_at,
    COALESCE(MAX(COALESCE(completed_at, claimed_at, created_at))::timestamp with time zone, CURRENT_TIMESTAMP) AS updated_at
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
    COALESCE(MIN(created_at)::timestamp with time zone, CURRENT_TIMESTAMP) AS created_at,
    COALESCE(MAX(COALESCE(completed_at, claimed_at, created_at))::timestamp with time zone, CURRENT_TIMESTAMP) AS updated_at
FROM second_review_tasks

UNION ALL

-- AI vs human diff queue
SELECT
    'ai_human_diff'::text AS queue_name,
    'AI与人工diff队列'::text AS description,
    85 AS priority,
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
    COALESCE(MIN(created_at)::timestamp with time zone, CURRENT_TIMESTAMP) AS created_at,
    COALESCE(MAX(COALESCE(completed_at, claimed_at, created_at))::timestamp with time zone, CURRENT_TIMESTAMP) AS updated_at
FROM ai_human_diff_tasks

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
    COALESCE(MIN(created_at)::timestamp with time zone, CURRENT_TIMESTAMP) AS created_at,
    COALESCE(MAX(COALESCE(completed_at, claimed_at, created_at))::timestamp with time zone, CURRENT_TIMESTAMP) AS updated_at
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
    COALESCE(MIN(created_at)::timestamp with time zone, CURRENT_TIMESTAMP) AS created_at,
    COALESCE(MAX(COALESCE(completed_at, claimed_at, created_at))::timestamp with time zone, CURRENT_TIMESTAMP) AS updated_at
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
    COALESCE(MIN(created_at)::timestamp with time zone, CURRENT_TIMESTAMP) AS created_at,
    COALESCE(MAX(COALESCE(completed_at, claimed_at, created_at))::timestamp with time zone, CURRENT_TIMESTAMP) AS updated_at
FROM video_second_review_tasks

ORDER BY priority DESC;

COMMENT ON VIEW unified_queue_stats IS 'Real-time queue statistics aggregated from all review task tables, including AI diff queue.';
