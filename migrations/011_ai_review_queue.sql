-- AI Review Queue Migration
-- Migration: 011_ai_review_queue
-- Description: Add AI review jobs, tasks, and results for comment review comparison
-- Date: 2026-01-19

-- 1. AI review jobs
CREATE TABLE IF NOT EXISTS ai_review_jobs (
    id SERIAL PRIMARY KEY,
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'scheduled', 'running', 'completed', 'failed', 'canceled')),
    run_at TIMESTAMP,
    max_count INTEGER NOT NULL,
    source_statuses TEXT[] NOT NULL DEFAULT ARRAY['pending', 'in_progress', 'completed'],
    model VARCHAR(100),
    prompt_version VARCHAR(50),
    created_by INTEGER REFERENCES users(id),
    total_tasks INTEGER NOT NULL DEFAULT 0,
    completed_tasks INTEGER NOT NULL DEFAULT 0,
    failed_tasks INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ai_review_jobs_status ON ai_review_jobs(status);
CREATE INDEX IF NOT EXISTS idx_ai_review_jobs_run_at ON ai_review_jobs(run_at);
CREATE INDEX IF NOT EXISTS idx_ai_review_jobs_created_by ON ai_review_jobs(created_by);

-- 2. AI review tasks
CREATE TABLE IF NOT EXISTS ai_review_tasks (
    id SERIAL PRIMARY KEY,
    job_id INTEGER NOT NULL REFERENCES ai_review_jobs(id) ON DELETE CASCADE,
    review_task_id INTEGER NOT NULL REFERENCES review_tasks(id),
    comment_id BIGINT NOT NULL REFERENCES comment(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'in_progress', 'completed', 'failed')),
    attempts INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (job_id, review_task_id)
);

CREATE INDEX IF NOT EXISTS idx_ai_review_tasks_job_id ON ai_review_tasks(job_id);
CREATE INDEX IF NOT EXISTS idx_ai_review_tasks_status ON ai_review_tasks(status);
CREATE INDEX IF NOT EXISTS idx_ai_review_tasks_review_task_id ON ai_review_tasks(review_task_id);

-- 3. AI review results
CREATE TABLE IF NOT EXISTS ai_review_results (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL REFERENCES ai_review_tasks(id) ON DELETE CASCADE,
    is_approved BOOLEAN NOT NULL,
    tags TEXT[] DEFAULT '{}',
    reason TEXT,
    confidence INTEGER NOT NULL CHECK (confidence >= 0 AND confidence <= 100),
    raw_output JSONB,
    model VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_review_results_task_id ON ai_review_results(task_id);

-- 4. Insert permissions
INSERT INTO permissions (permission_key, name, description, resource, action, category, is_active) VALUES
    ('ai-review:jobs:create', '创建AI审核批次', '允许创建AI审核批次', 'ai_review', 'create', 'ai_review', true),
    ('ai-review:jobs:start', '启动AI审核批次', '允许启动AI审核批次', 'ai_review', 'start', 'ai_review', true),
    ('ai-review:jobs:list', '查看AI审核批次列表', '允许查看AI审核批次列表', 'ai_review', 'list', 'ai_review', true),
    ('ai-review:jobs:read', '查看AI审核批次详情', '允许查看AI审核批次详情', 'ai_review', 'read', 'ai_review', true),
    ('ai-review:compare', '查看AI对比分析', '允许查看AI审核对比分析', 'ai_review', 'compare', 'ai_review', true)
ON CONFLICT (permission_key) DO NOTHING;

-- 5. Grant permissions to admin users
INSERT INTO user_permissions (user_id, permission_key, granted_by)
SELECT u.id, p.permission_key, u.id
FROM users u
CROSS JOIN (
    SELECT permission_key FROM permissions
    WHERE permission_key LIKE 'ai-review:%'
) p
WHERE u.role = 'admin'
ON CONFLICT (user_id, permission_key) DO NOTHING;
