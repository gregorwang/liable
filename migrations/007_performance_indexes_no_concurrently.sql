-- Performance Optimization Indexes (No CONCURRENTLY - for Supabase SQL Editor)
-- Migration: 007_performance_indexes
-- Description: Add composite indexes for performance optimization
-- Date: 2026-01-15

-- 1. Task status + claimed_at composite indexes (for releasing expired tasks)
CREATE INDEX IF NOT EXISTS idx_review_tasks_status_claimed
ON review_tasks(status, claimed_at)
WHERE status = 'in_progress';

CREATE INDEX IF NOT EXISTS idx_second_review_tasks_status_claimed
ON second_review_tasks(status, claimed_at)
WHERE status = 'in_progress';

CREATE INDEX IF NOT EXISTS idx_quality_check_tasks_status_claimed
ON quality_check_tasks(status, claimed_at)
WHERE status = 'in_progress';

CREATE INDEX IF NOT EXISTS idx_video_first_review_tasks_status_claimed
ON video_first_review_tasks(status, claimed_at)
WHERE status = 'in_progress';

CREATE INDEX IF NOT EXISTS idx_video_second_review_tasks_status_claimed
ON video_second_review_tasks(status, claimed_at)
WHERE status = 'in_progress';

-- 2. Reviewer + status composite indexes (for querying "my tasks")
CREATE INDEX IF NOT EXISTS idx_review_tasks_reviewer_status
ON review_tasks(reviewer_id, status)
WHERE reviewer_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_second_review_tasks_reviewer_status
ON second_review_tasks(reviewer_id, status)
WHERE reviewer_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_quality_check_tasks_reviewer_status
ON quality_check_tasks(reviewer_id, status)
WHERE reviewer_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_video_first_review_tasks_reviewer_status
ON video_first_review_tasks(reviewer_id, status)
WHERE reviewer_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_video_second_review_tasks_reviewer_status
ON video_second_review_tasks(reviewer_id, status)
WHERE reviewer_id IS NOT NULL;

-- 3. Created time indexes (for statistics queries)
CREATE INDEX IF NOT EXISTS idx_review_results_created_at
ON review_results(created_at);

CREATE INDEX IF NOT EXISTS idx_second_review_results_created_at
ON second_review_results(created_at);

CREATE INDEX IF NOT EXISTS idx_quality_check_results_created_at
ON quality_check_results(created_at);

CREATE INDEX IF NOT EXISTS idx_video_first_review_results_created_at
ON video_first_review_results(created_at);

CREATE INDEX IF NOT EXISTS idx_video_second_review_results_created_at
ON video_second_review_results(created_at);

-- 4. Reviewer performance statistics indexes
CREATE INDEX IF NOT EXISTS idx_review_results_reviewer_approved
ON review_results(reviewer_id, is_approved);

CREATE INDEX IF NOT EXISTS idx_second_review_results_reviewer_approved
ON second_review_results(reviewer_id, is_approved);

CREATE INDEX IF NOT EXISTS idx_quality_check_results_reviewer_passed
ON quality_check_results(reviewer_id, is_passed);

CREATE INDEX IF NOT EXISTS idx_video_first_review_results_reviewer_approved
ON video_first_review_results(reviewer_id, is_approved);

CREATE INDEX IF NOT EXISTS idx_video_second_review_results_reviewer_approved
ON video_second_review_results(reviewer_id, is_approved);

-- 5. Status indexes for task searches
CREATE INDEX IF NOT EXISTS idx_review_tasks_status
ON review_tasks(status);

CREATE INDEX IF NOT EXISTS idx_second_review_tasks_status
ON second_review_tasks(status);

-- 6. Task ID indexes for result lookups
CREATE INDEX IF NOT EXISTS idx_review_results_task_id
ON review_results(task_id);

CREATE INDEX IF NOT EXISTS idx_second_review_results_task_id
ON second_review_results(second_task_id);

CREATE INDEX IF NOT EXISTS idx_quality_check_results_task_id
ON quality_check_results(qc_task_id);

CREATE INDEX IF NOT EXISTS idx_video_first_review_results_task_id
ON video_first_review_results(task_id);

CREATE INDEX IF NOT EXISTS idx_video_second_review_results_task_id
ON video_second_review_results(second_task_id);

-- 7. Comment ID indexes for second review lookups
CREATE INDEX IF NOT EXISTS idx_second_review_tasks_comment_id
ON second_review_tasks(comment_id);

CREATE INDEX IF NOT EXISTS idx_quality_check_tasks_comment_id
ON quality_check_tasks(comment_id);

-- 8. Video ID indexes for video review tasks
CREATE INDEX IF NOT EXISTS idx_video_first_review_tasks_video_id
ON video_first_review_tasks(video_id);

CREATE INDEX IF NOT EXISTS idx_video_second_review_tasks_video_id
ON video_second_review_tasks(video_id);

-- 9. Review task status for active reviewer queries
CREATE INDEX IF NOT EXISTS idx_review_tasks_status_reviewer
ON review_tasks(status, reviewer_id);

CREATE INDEX IF NOT EXISTS idx_second_review_tasks_status_reviewer
ON second_review_tasks(status, reviewer_id);

CREATE INDEX IF NOT EXISTS idx_quality_check_tasks_status_reviewer
ON quality_check_tasks(status, reviewer_id);

CREATE INDEX IF NOT EXISTS idx_video_first_review_tasks_status_reviewer
ON video_first_review_tasks(status, reviewer_id);

CREATE INDEX IF NOT EXISTS idx_video_second_review_tasks_status_reviewer
ON video_second_review_tasks(status, reviewer_id);
