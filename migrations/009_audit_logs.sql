-- Create audit logs table for security monitoring and compliance
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(36) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),

    -- User information
    user_id INTEGER,
    username VARCHAR(100),
    role VARCHAR(50),

    -- Request information
    ip_address VARCHAR(45),          -- Support IPv6
    user_agent TEXT,
    method VARCHAR(10),
    path TEXT,
    query_params TEXT,

    -- Permission check
    permission_checked VARCHAR(200),
    permission_granted BOOLEAN,

    -- Response information
    status_code INTEGER,
    response_time_ms INTEGER,
    error_message TEXT,

    -- Metadata (JSON for flexible additional data)
    metadata JSONB
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_permission ON audit_logs(permission_checked);
CREATE INDEX IF NOT EXISTS idx_audit_logs_ip_address ON audit_logs(ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status_code ON audit_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_audit_logs_permission_granted ON audit_logs(permission_granted);

-- Composite index for user activity queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_timestamp ON audit_logs(user_id, timestamp DESC);

-- Composite index for security monitoring (failed requests)
CREATE INDEX IF NOT EXISTS idx_audit_logs_status_timestamp ON audit_logs(status_code, timestamp) WHERE status_code >= 400;

-- Comment
COMMENT ON TABLE audit_logs IS 'Audit log table for security monitoring, compliance, and troubleshooting';
COMMENT ON COLUMN audit_logs.request_id IS 'Unique identifier for the request (UUID)';
COMMENT ON COLUMN audit_logs.timestamp IS 'When the request was received';
COMMENT ON COLUMN audit_logs.user_id IS 'ID of the authenticated user (nullable for public endpoints)';
COMMENT ON COLUMN audit_logs.username IS 'Username of the authenticated user';
COMMENT ON COLUMN audit_logs.role IS 'Role of the authenticated user';
COMMENT ON COLUMN audit_logs.ip_address IS 'Client IP address (supports IPv6)';
COMMENT ON COLUMN audit_logs.user_agent IS 'HTTP User-Agent header';
COMMENT ON COLUMN audit_logs.method IS 'HTTP method (GET, POST, etc.)';
COMMENT ON COLUMN audit_logs.path IS 'Request path';
COMMENT ON COLUMN audit_logs.query_params IS 'URL query parameters';
COMMENT ON COLUMN audit_logs.permission_checked IS 'Permission key that was checked (if any)';
COMMENT ON COLUMN audit_logs.permission_granted IS 'Whether the permission was granted';
COMMENT ON COLUMN audit_logs.status_code IS 'HTTP response status code';
COMMENT ON COLUMN audit_logs.response_time_ms IS 'Response time in milliseconds';
COMMENT ON COLUMN audit_logs.error_message IS 'Error message if request failed';
COMMENT ON COLUMN audit_logs.metadata IS 'Additional metadata as JSON';
