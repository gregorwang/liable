-- ============================================================
-- Migration: 021_user_profile_fields
-- Description: Add user profile fields and permissions for system profile updates
-- Created: 2026-01-24
-- ============================================================

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS avatar_key TEXT,
    ADD COLUMN IF NOT EXISTS gender VARCHAR(20),
    ADD COLUMN IF NOT EXISTS signature TEXT,
    ADD COLUMN IF NOT EXISTS office_location VARCHAR(100),
    ADD COLUMN IF NOT EXISTS department VARCHAR(100),
    ADD COLUMN IF NOT EXISTS school VARCHAR(100),
    ADD COLUMN IF NOT EXISTS company VARCHAR(100),
    ADD COLUMN IF NOT EXISTS direct_manager VARCHAR(100);

INSERT INTO permissions (permission_key, name, description, resource, action, category, is_active) VALUES
    ('users:profile:update', '修改用户系统资料', '允许修改用户系统资料字段', 'users', 'update_profile', 'users', true)
ON CONFLICT (permission_key) DO NOTHING;

INSERT INTO user_permissions (user_id, permission_key, granted_by)
SELECT u.id, p.permission_key, u.id
FROM users u
CROSS JOIN (
    SELECT permission_key FROM permissions
    WHERE permission_key IN ('users:profile:update')
) p
WHERE u.role = 'admin'
ON CONFLICT (user_id, permission_key) DO NOTHING;
