-- Video Queue Pool System Migration
-- Refactor video review from first/second review to traffic pool-based single-stage review
-- Pools: 100k (entry) -> 1m -> 10m (quality check)

-- 1. Video Queue Tasks table - unified single-stage review by pool
CREATE TABLE IF NOT EXISTS video_queue_tasks (
    id SERIAL PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES tiktok_videos(id),
    pool VARCHAR(10) NOT NULL CHECK (pool IN ('100k', '1m', '10m')),
    reviewer_id INTEGER REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed')),
    claimed_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_video_pool UNIQUE (video_id, pool)
);

-- 2. Video Queue Results table - simplified review result
CREATE TABLE IF NOT EXISTS video_queue_results (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL REFERENCES video_queue_tasks(id),
    reviewer_id INTEGER NOT NULL REFERENCES users(id),

    -- Simplified review decision (3 choices)
    review_decision VARCHAR(20) NOT NULL CHECK (review_decision IN ('push_next_pool', 'natural_pool', 'remove_violation')),

    -- Review reason (required)
    reason TEXT NOT NULL,

    -- Review tags (max 3 tags, stored as array)
    tags TEXT[] DEFAULT '{}',

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 3. Update tag_config table to support scope and queue_id
-- Add new columns if they don't exist
DO $$
BEGIN
    -- Add scope column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='tag_config' AND column_name='scope') THEN
        ALTER TABLE tag_config ADD COLUMN scope VARCHAR(20) DEFAULT 'comment';
    END IF;

    -- Add queue_id column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='tag_config' AND column_name='queue_id') THEN
        ALTER TABLE tag_config ADD COLUMN queue_id VARCHAR(20);
    END IF;

    -- Add is_simple column for simplified tag system
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='tag_config' AND column_name='is_simple') THEN
        ALTER TABLE tag_config ADD COLUMN is_simple BOOLEAN DEFAULT FALSE;
    END IF;
END $$;

-- 4. Update video_quality_tags table to support scope and queue
DO $$
BEGIN
    -- Add scope column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='video_quality_tags' AND column_name='scope') THEN
        ALTER TABLE video_quality_tags ADD COLUMN scope VARCHAR(20) DEFAULT 'video';
    END IF;

    -- Add queue_id column
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='video_quality_tags' AND column_name='queue_id') THEN
        ALTER TABLE video_quality_tags ADD COLUMN queue_id VARCHAR(10);
    END IF;
END $$;

-- 5. Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_video_queue_tasks_pool ON video_queue_tasks(pool);
CREATE INDEX IF NOT EXISTS idx_video_queue_tasks_status ON video_queue_tasks(status);
CREATE INDEX IF NOT EXISTS idx_video_queue_tasks_reviewer ON video_queue_tasks(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_video_queue_tasks_claimed_at ON video_queue_tasks(claimed_at);
CREATE INDEX IF NOT EXISTS idx_video_queue_tasks_video ON video_queue_tasks(video_id);
CREATE INDEX IF NOT EXISTS idx_video_queue_results_task ON video_queue_results(task_id);
CREATE INDEX IF NOT EXISTS idx_video_queue_results_reviewer ON video_queue_results(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_video_queue_results_decision ON video_queue_results(review_decision);
CREATE INDEX IF NOT EXISTS idx_tag_config_scope ON tag_config(scope);
CREATE INDEX IF NOT EXISTS idx_tag_config_queue_id ON tag_config(queue_id);
CREATE INDEX IF NOT EXISTS idx_video_quality_tags_scope ON video_quality_tags(scope);
CREATE INDEX IF NOT EXISTS idx_video_quality_tags_queue_id ON video_quality_tags(queue_id);

-- 6. Insert default simplified video review tags
INSERT INTO video_quality_tags (name, description, category, scope, queue_id, is_active) VALUES
    -- 100k pool tags
    ('内容优质', '内容质量优秀，适合推广', 'content', 'video', '100k', true),
    ('有传播潜力', '具有良好的传播潜力', 'engagement', 'video', '100k', true),
    ('技术质量好', '视频技术质量优秀', 'technical', 'video', '100k', true),
    ('内容一般', '内容质量一般，不建议推广', 'content', 'video', '100k', true),
    ('低质量', '视频质量较低', 'technical', 'video', '100k', true),
    ('违规风险', '可能存在违规内容', 'compliance', 'video', '100k', true),

    -- 1m pool tags
    ('热点话题', '涉及热门话题', 'engagement', 'video', '1m', true),
    ('专业制作', '专业级制作水平', 'technical', 'video', '1m', true),
    ('高互动性', '容易引发用户互动', 'engagement', 'video', '1m', true),
    ('创意独特', '创意新颖独特', 'content', 'video', '1m', true),
    ('表现平平', '表现平平，不建议继续推广', 'content', 'video', '1m', true),

    -- 10m pool tags (quality check)
    ('爆款潜质', '具有爆款视频的潜质', 'engagement', 'video', '10m', true),
    ('顶级制作', '顶级制作水平', 'technical', 'video', '10m', true),
    ('强传播力', '具有极强的传播能力', 'engagement', 'video', '10m', true),
    ('现象级内容', '现象级优质内容', 'content', 'video', '10m', true),
    ('不符合标准', '不符合1000万流量池标准', 'content', 'video', '10m', true)
ON CONFLICT (name) DO NOTHING;

-- 7. Insert queue permissions into permissions table
INSERT INTO permissions (permission_key, name, description, resource, action, category, is_active) VALUES
    -- 100k pool permissions
    ('queue.video.100k.claim', '领取100k流量池视频任务', '允许领取100k流量池的视频审核任务', 'video_queue', 'claim', 'video_review', true),
    ('queue.video.100k.submit', '提交100k流量池审核结果', '允许提交100k流量池的审核结果', 'video_queue', 'submit', 'video_review', true),
    ('queue.video.100k.return', '归还100k流量池任务', '允许归还100k流量池的任务', 'video_queue', 'return', 'video_review', true),
    ('queue.video.100k.my', '查看100k流量池我的任务', '允许查看100k流量池我的任务列表', 'video_queue', 'my', 'video_review', true),

    -- 1m pool permissions
    ('queue.video.1m.claim', '领取1m流量池视频任务', '允许领取1m流量池的视频审核任务', 'video_queue', 'claim', 'video_review', true),
    ('queue.video.1m.submit', '提交1m流量池审核结果', '允许提交1m流量池的审核结果', 'video_queue', 'submit', 'video_review', true),
    ('queue.video.1m.return', '归还1m流量池任务', '允许归还1m流量池的任务', 'video_queue', 'return', 'video_review', true),
    ('queue.video.1m.my', '查看1m流量池我的任务', '允许查看1m流量池我的任务列表', 'video_queue', 'my', 'video_review', true),

    -- 10m pool permissions (quality check only)
    ('queue.video.10m.claim', '领取10m流量池视频任务', '允许领取10m流量池的视频审核任务（仅质检员）', 'video_queue', 'claim', 'video_review', true),
    ('queue.video.10m.submit', '提交10m流量池审核结果', '允许提交10m流量池的审核结果（仅质检员）', 'video_queue', 'submit', 'video_review', true),
    ('queue.video.10m.return', '归还10m流量池任务', '允许归还10m流量池的任务（仅质检员）', 'video_queue', 'return', 'video_review', true),
    ('queue.video.10m.my', '查看10m流量池我的任务', '允许查看10m流量池我的任务列表（仅质检员）', 'video_queue', 'my', 'video_review', true)
ON CONFLICT (permission_key) DO NOTHING;

-- 8. Create view for video queue statistics
CREATE OR REPLACE VIEW video_queue_pool_stats AS
SELECT
    pool,
    COUNT(*) as total_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress_tasks,
    AVG(CASE
        WHEN status = 'completed' AND completed_at IS NOT NULL AND claimed_at IS NOT NULL
        THEN EXTRACT(EPOCH FROM (completed_at - claimed_at))/60
    END) as avg_process_time_minutes
FROM video_queue_tasks
GROUP BY pool;

-- 9. Create view for video queue results statistics
CREATE OR REPLACE VIEW video_queue_decision_stats AS
SELECT
    vqt.pool,
    vqr.review_decision,
    COUNT(*) as decision_count,
    COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (PARTITION BY vqt.pool) as decision_percentage
FROM video_queue_results vqr
JOIN video_queue_tasks vqt ON vqr.task_id = vqt.id
GROUP BY vqt.pool, vqr.review_decision;

-- 10. Create function to get available video queue tags
CREATE OR REPLACE FUNCTION get_video_queue_tags(p_pool VARCHAR(10))
RETURNS TABLE (
    id INTEGER,
    name VARCHAR(50),
    description TEXT,
    category VARCHAR(20)
) AS $$
BEGIN
    RETURN QUERY
    SELECT vqt.id, vqt.name, vqt.description, vqt.category
    FROM video_quality_tags vqt
    WHERE vqt.is_active = TRUE
      AND vqt.scope = 'video'
      AND (vqt.queue_id = p_pool OR vqt.queue_id IS NULL)
    ORDER BY vqt.category, vqt.name;
END;
$$ LANGUAGE plpgsql;

-- 11. Add comment for documentation
COMMENT ON TABLE video_queue_tasks IS 'Video review tasks organized by traffic pool (100k, 1m, 10m)';
COMMENT ON TABLE video_queue_results IS 'Simplified video review results with decision, reason, and tags (max 3)';
COMMENT ON COLUMN video_queue_results.review_decision IS 'Review decision: push_next_pool (推送下一流量池), natural_pool (自然流量池), remove_violation (违规下架)';
COMMENT ON COLUMN video_queue_results.tags IS 'Review tags array (max 3 tags allowed)';
