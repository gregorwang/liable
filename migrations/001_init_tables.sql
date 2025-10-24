-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('reviewer', 'admin')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create review_tasks table
CREATE TABLE IF NOT EXISTS review_tasks (
    id SERIAL PRIMARY KEY,
    comment_id BIGINT NOT NULL,
    reviewer_id INTEGER REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed')),
    claimed_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (comment_id) REFERENCES comment(id)
);

-- Create review_results table
CREATE TABLE IF NOT EXISTS review_results (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL REFERENCES review_tasks(id),
    reviewer_id INTEGER NOT NULL REFERENCES users(id),
    is_approved BOOLEAN NOT NULL,
    tags TEXT[],
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create tag_config table
CREATE TABLE IF NOT EXISTS tag_config (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_review_tasks_status ON review_tasks(status);
CREATE INDEX IF NOT EXISTS idx_review_tasks_reviewer ON review_tasks(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_review_tasks_claimed_at ON review_tasks(claimed_at);
CREATE INDEX IF NOT EXISTS idx_review_results_reviewer ON review_results(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_review_results_created_at ON review_results(created_at);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- Insert default admin user (password: admin123)
-- Password hash generated using bcrypt with cost 10
INSERT INTO users (username, password, role, status)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin', 'approved')
ON CONFLICT (username) DO NOTHING;

-- Insert default violation tags
INSERT INTO tag_config (name, description, is_active) VALUES
    ('广告', '包含广告或推广内容', true),
    ('垃圾', '无意义或垃圾信息', true),
    ('色情', '包含色情或低俗内容', true),
    ('暴力', '包含暴力或血腥内容', true),
    ('政治敏感', '包含政治敏感内容', true),
    ('人身攻击', '包含人身攻击或侮辱', true)
ON CONFLICT (name) DO NOTHING;

-- Create review tasks for all comments that don't have tasks yet
INSERT INTO review_tasks (comment_id, status, created_at)
SELECT c.id, 'pending', NOW()
FROM comment c
WHERE NOT EXISTS (
    SELECT 1 FROM review_tasks rt WHERE rt.comment_id = c.id
);

