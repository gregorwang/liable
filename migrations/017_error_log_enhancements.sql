-- ============================================================
-- Migration: 017_error_log_enhancements
-- Description: Extend audit_logs with structured error metadata and request context.
-- Created: 2026-02-10
-- ============================================================

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS error_code VARCHAR(100);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS error_type VARCHAR(50);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS error_description TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS module_name VARCHAR(200);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS method_name VARCHAR(200);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS server_ip VARCHAR(100);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS server_port VARCHAR(10);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS page_url TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS request_params JSONB;

CREATE INDEX IF NOT EXISTS idx_audit_logs_error_code ON audit_logs(error_code);
CREATE INDEX IF NOT EXISTS idx_audit_logs_error_type ON audit_logs(error_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_module_name ON audit_logs(module_name);
