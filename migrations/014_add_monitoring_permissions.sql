-- ============================================================
-- Migration: 014_add_monitoring_permissions
-- Description: Add permissions for monitoring metrics access
-- Created: 2026-01-21
-- ============================================================

INSERT INTO permissions (permission_key, name, description, resource, action, category, is_active) VALUES
    ('monitoring.read', '查看监控指标', '允许查看系统监控指标与健康状态', 'monitoring', 'read', 'system', true)
ON CONFLICT (permission_key) DO NOTHING;

INSERT INTO user_permissions (user_id, permission_key, granted_by)
SELECT u.id, p.permission_key, u.id
FROM users u
CROSS JOIN (
    SELECT permission_key FROM permissions
    WHERE permission_key IN ('monitoring.read')
) p
WHERE u.role = 'admin'
ON CONFLICT (user_id, permission_key) DO NOTHING;
