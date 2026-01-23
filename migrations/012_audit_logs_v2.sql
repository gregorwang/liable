-- Audit logs v2 migration
-- Migration: 012_audit_logs_v2
-- Description: Expand audit logs schema for audit system PRD + add export tracking
-- Date: 2026-01-19

-- Ensure UUID generation is available
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Upgrade primary key to UUID while preserving legacy numeric ID
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'audit_logs' AND column_name = 'id'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'audit_logs' AND column_name = 'legacy_id'
    ) THEN
        ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS id_uuid uuid DEFAULT gen_random_uuid();
        UPDATE audit_logs SET id_uuid = gen_random_uuid() WHERE id_uuid IS NULL;

        ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_pkey;
        ALTER TABLE audit_logs RENAME COLUMN id TO legacy_id;
        ALTER TABLE audit_logs RENAME COLUMN id_uuid TO id;
        ALTER TABLE audit_logs ALTER COLUMN legacy_id DROP DEFAULT;
        ALTER TABLE audit_logs ALTER COLUMN id SET DEFAULT gen_random_uuid();
        ALTER TABLE audit_logs ADD PRIMARY KEY (id);
    END IF;
END $$;

-- Add new audit fields
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS user_role VARCHAR(50);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS action_type VARCHAR(100);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS action_category VARCHAR(50);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS action_description TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS result VARCHAR(20);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS endpoint VARCHAR(500);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS http_method VARCHAR(10);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS session_id VARCHAR(100);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS request_body JSONB;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS response_body JSONB;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS geo_location VARCHAR(100);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS device_type VARCHAR(20);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS browser VARCHAR(50);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS os VARCHAR(50);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS resource_type VARCHAR(50);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS resource_id VARCHAR(100);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS resource_ids JSONB;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS changes JSONB;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS error_stack TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS duration_ms INTEGER;

-- Backfill new columns from legacy fields
UPDATE audit_logs
SET
    created_at = COALESCE(created_at, timestamp),
    user_role = COALESCE(user_role, role),
    endpoint = COALESCE(endpoint, path),
    http_method = COALESCE(http_method, method),
    duration_ms = COALESCE(duration_ms, response_time_ms),
    action_type = COALESCE(action_type, 'api.request'),
    action_category = COALESCE(action_category, 'system_operation'),
    action_description = COALESCE(action_description, 'API request'),
    result = COALESCE(result, CASE WHEN status_code < 400 THEN 'success' ELSE 'failure' END)
WHERE created_at IS NULL
   OR user_role IS NULL
   OR endpoint IS NULL
   OR http_method IS NULL
   OR duration_ms IS NULL
   OR action_type IS NULL
   OR action_category IS NULL
   OR action_description IS NULL
   OR result IS NULL;

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_time ON audit_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_type_time ON audit_logs(action_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_category_time ON audit_logs(action_category, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_result ON audit_logs(result);
CREATE INDEX IF NOT EXISTS idx_audit_logs_endpoint ON audit_logs(endpoint);
CREATE INDEX IF NOT EXISTS idx_audit_logs_http_method ON audit_logs(http_method);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status_code ON audit_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_audit_logs_ip_address ON audit_logs(ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_request_id ON audit_logs(request_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_duration_ms ON audit_logs(duration_ms);
CREATE INDEX IF NOT EXISTS idx_audit_logs_search ON audit_logs
USING GIN (to_tsvector('simple', coalesce(action_description, '') || ' ' || coalesce(error_message, '')));

-- Export tracking table
CREATE TABLE IF NOT EXISTS audit_log_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER REFERENCES users(id),
    username VARCHAR(100),
    export_format VARCHAR(10) NOT NULL,
    filters JSONB,
    fields TEXT[],
    status VARCHAR(20) NOT NULL DEFAULT 'processing',
    row_count INTEGER,
    file_key TEXT,
    expires_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_log_exports_user_time ON audit_log_exports(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_exports_status ON audit_log_exports(status);

-- Permissions for audit log access
INSERT INTO permissions (permission_key, name, description, resource, action, category, is_active) VALUES
    ('audit.logs.read', '查看审计日志', '允许查看审计日志与详情', 'audit_logs', 'read', 'system', true),
    ('audit.logs.export', '导出审计日志', '允许导出审计日志数据', 'audit_logs', 'export', 'system', true)
ON CONFLICT (permission_key) DO NOTHING;

-- Grant audit log permissions to admin users
INSERT INTO user_permissions (user_id, permission_key, granted_by)
SELECT u.id, p.permission_key, u.id
FROM users u
CROSS JOIN (
    SELECT permission_key FROM permissions
    WHERE permission_key IN ('audit.logs.read', 'audit.logs.export')
) p
WHERE u.role = 'admin'
ON CONFLICT (user_id, permission_key) DO NOTHING;
