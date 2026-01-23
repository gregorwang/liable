-- Bug report submissions (user feedback with optional screenshots)
CREATE TABLE IF NOT EXISTS bug_reports (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    title VARCHAR(200),
    description TEXT NOT NULL,
    error_details TEXT,
    page_url TEXT,
    user_agent TEXT,
    screenshots JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (jsonb_typeof(screenshots) = 'array'),
    CHECK (jsonb_array_length(screenshots) <= 2)
);

CREATE INDEX IF NOT EXISTS idx_bug_reports_user_id ON bug_reports(user_id);
CREATE INDEX IF NOT EXISTS idx_bug_reports_created_at ON bug_reports(created_at);
