-- TikTok Video Quality Multi-Dimensional Review System Migration
-- Create tables for video review system

-- 1. TikTok Videos table - stores video metadata
CREATE TABLE IF NOT EXISTS tiktok_videos (
    id SERIAL PRIMARY KEY,
    video_key VARCHAR(500) NOT NULL UNIQUE, -- R2 path/key
    filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL, -- bytes
    duration INTEGER, -- seconds
    upload_time TIMESTAMP,
    video_url TEXT, -- pre-signed URL (temporary)
    url_expires_at TIMESTAMP, -- when pre-signed URL expires
    status VARCHAR(30) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'first_review_completed', 'second_review_completed')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2. Video Quality Tags table - predefined quality assessment tags
CREATE TABLE IF NOT EXISTS video_quality_tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    category VARCHAR(20) NOT NULL CHECK (category IN ('content', 'technical', 'compliance', 'engagement')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 3. Video First Review Tasks table
CREATE TABLE IF NOT EXISTS video_first_review_tasks (
    id SERIAL PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES tiktok_videos(id),
    reviewer_id INTEGER REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed')),
    claimed_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 4. Video First Review Results table
CREATE TABLE IF NOT EXISTS video_first_review_results (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL REFERENCES video_first_review_tasks(id),
    reviewer_id INTEGER NOT NULL REFERENCES users(id),
    is_approved BOOLEAN NOT NULL,
    
    -- Quality dimensions as JSONB
    quality_dimensions JSONB NOT NULL DEFAULT '{}',
    -- Structure: {
    --   "content_quality": {"score": 8, "tags": ["创意优秀"], "notes": "..."},
    --   "technical_quality": {"score": 7, "tags": ["画质清晰"], "notes": "..."},
    --   "compliance": {"score": 9, "tags": [], "notes": "..."},
    --   "engagement_potential": {"score": 8, "tags": ["有趣"], "notes": "..."}
    -- }
    
    overall_score INTEGER NOT NULL CHECK (overall_score >= 1 AND overall_score <= 40), -- sum of all dimension scores
    traffic_pool_result VARCHAR(50), -- recommended traffic pool category
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 5. Video Second Review Tasks table
CREATE TABLE IF NOT EXISTS video_second_review_tasks (
    id SERIAL PRIMARY KEY,
    first_review_result_id INTEGER NOT NULL REFERENCES video_first_review_results(id),
    video_id INTEGER NOT NULL REFERENCES tiktok_videos(id),
    reviewer_id INTEGER REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed')),
    claimed_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 6. Video Second Review Results table
CREATE TABLE IF NOT EXISTS video_second_review_results (
    id SERIAL PRIMARY KEY,
    second_task_id INTEGER NOT NULL REFERENCES video_second_review_tasks(id),
    reviewer_id INTEGER NOT NULL REFERENCES users(id),
    is_approved BOOLEAN NOT NULL,
    
    -- Same structure as first review results
    quality_dimensions JSONB NOT NULL DEFAULT '{}',
    overall_score INTEGER NOT NULL CHECK (overall_score >= 1 AND overall_score <= 40),
    traffic_pool_result VARCHAR(50),
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tiktok_videos_status ON tiktok_videos(status);
CREATE INDEX IF NOT EXISTS idx_tiktok_videos_video_key ON tiktok_videos(video_key);
CREATE INDEX IF NOT EXISTS idx_video_quality_tags_category ON video_quality_tags(category);
CREATE INDEX IF NOT EXISTS idx_video_quality_tags_active ON video_quality_tags(is_active);
CREATE INDEX IF NOT EXISTS idx_video_first_review_tasks_status ON video_first_review_tasks(status);
CREATE INDEX IF NOT EXISTS idx_video_first_review_tasks_reviewer ON video_first_review_tasks(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_video_first_review_tasks_claimed_at ON video_first_review_tasks(claimed_at);
CREATE INDEX IF NOT EXISTS idx_video_first_review_results_reviewer ON video_first_review_results(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_video_first_review_results_created_at ON video_first_review_results(created_at);
CREATE INDEX IF NOT EXISTS idx_video_second_review_tasks_status ON video_second_review_tasks(status);
CREATE INDEX IF NOT EXISTS idx_video_second_review_tasks_reviewer ON video_second_review_tasks(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_video_second_review_results_reviewer ON video_second_review_results(reviewer_id);

-- Insert default video quality tags
INSERT INTO video_quality_tags (name, description, category, is_active) VALUES
    -- Content Quality Tags
    ('创意优秀', '内容创意独特，有吸引力', 'content', true),
    ('内容有趣', '内容生动有趣，能引起兴趣', 'content', true),
    ('教育价值', '具有教育意义或知识性', 'content', true),
    ('情感共鸣', '能引起观众情感共鸣', 'content', true),
    ('内容重复', '内容缺乏新意，重复性高', 'content', true),
    ('内容空洞', '内容缺乏实质，空洞无物', 'content', true),
    
    -- Technical Quality Tags
    ('画质清晰', '视频画质清晰，分辨率高', 'technical', true),
    ('音质良好', '音频质量良好，无杂音', 'technical', true),
    ('剪辑流畅', '视频剪辑流畅，转场自然', 'technical', true),
    ('构图合理', '画面构图合理，视觉效果好', 'technical', true),
    ('画质模糊', '视频画质模糊，分辨率低', 'technical', true),
    ('音质差', '音频质量差，有杂音或失真', 'technical', true),
    ('剪辑粗糙', '视频剪辑粗糙，转场突兀', 'technical', true),
    
    -- Compliance Tags
    ('内容合规', '内容符合平台规范', 'compliance', true),
    ('版权问题', '可能存在版权问题', 'compliance', true),
    ('内容违规', '内容违反平台规范', 'compliance', true),
    ('敏感内容', '包含敏感或不当内容', 'compliance', true),
    
    -- Engagement Potential Tags
    ('传播性强', '具有强传播潜力', 'engagement', true),
    ('互动性好', '能引发用户互动', 'engagement', true),
    ('话题性强', '具有话题性和讨论价值', 'engagement', true),
    ('时效性强', '内容具有时效性', 'engagement', true),
    ('传播性弱', '传播潜力有限', 'engagement', true),
    ('互动性差', '难以引发用户互动', 'engagement', true)
ON CONFLICT (name) DO NOTHING;

-- Create a view for queue statistics (similar to existing queue_stats)
CREATE OR REPLACE VIEW video_queue_stats AS
SELECT 
    'video_first_review' as queue_name,
    COUNT(*) as total_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress_tasks,
    AVG(CASE 
        WHEN status = 'completed' AND completed_at IS NOT NULL AND claimed_at IS NOT NULL 
        THEN EXTRACT(EPOCH FROM (completed_at - claimed_at))/60 
    END) as avg_process_time_minutes,
    true as is_active
FROM video_first_review_tasks
UNION ALL
SELECT 
    'video_second_review' as queue_name,
    COUNT(*) as total_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_tasks,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress_tasks,
    AVG(CASE 
        WHEN status = 'completed' AND completed_at IS NOT NULL AND claimed_at IS NOT NULL 
        THEN EXTRACT(EPOCH FROM (completed_at - claimed_at))/60 
    END) as avg_process_time_minutes,
    true as is_active
FROM video_second_review_tasks;

