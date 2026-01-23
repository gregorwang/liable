-- AI Review Job Archive Migration
-- Migration: 012_ai_review_archive
-- Description: Add archive flag for AI review jobs and permission for archiving
-- Date: 2026-01-21

ALTER TABLE ai_review_jobs
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_ai_review_jobs_archived_at ON ai_review_jobs(archived_at);

INSERT INTO permissions (permission_key, name, description, resource, action, category, is_active) VALUES
    ('ai-review:jobs:archive', '归档AI审核批次', '允许归档AI审核批次', 'ai_review', 'archive', 'ai_review', true)
ON CONFLICT (permission_key) DO NOTHING;

INSERT INTO user_permissions (user_id, permission_key, granted_by)
SELECT u.id, p.permission_key, u.id
FROM users u
JOIN permissions p ON p.permission_key = 'ai-review:jobs:archive'
WHERE u.role = 'admin'
ON CONFLICT (user_id, permission_key) DO NOTHING;
