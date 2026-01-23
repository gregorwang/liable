-- Idempotency constraints for review workflows
-- Migration: 013_idempotency_constraints

-- Result tables: one result per task
CREATE UNIQUE INDEX IF NOT EXISTS ux_review_results_task_id
ON review_results(task_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_second_review_results_task_id
ON second_review_results(second_task_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_quality_check_results_task_id
ON quality_check_results(qc_task_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_video_first_review_results_task_id
ON video_first_review_results(task_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_video_second_review_results_task_id
ON video_second_review_results(second_task_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_video_queue_results_task_id
ON video_queue_results(task_id);

-- Follow-up task tables: one task per upstream result
CREATE UNIQUE INDEX IF NOT EXISTS ux_second_review_tasks_first_review_result_id
ON second_review_tasks(first_review_result_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_quality_check_tasks_first_review_result_id
ON quality_check_tasks(first_review_result_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_video_second_review_tasks_first_review_result_id
ON video_second_review_tasks(first_review_result_id);
