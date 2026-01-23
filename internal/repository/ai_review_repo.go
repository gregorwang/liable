package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type AIReviewRepository struct {
	db *sql.DB
}

type AIReviewTaskPayload struct {
	ID           int
	ReviewTaskID int
	CommentID    int64
	CommentText  string
}

func NewAIReviewRepository() *AIReviewRepository {
	return &AIReviewRepository{db: database.DB}
}

func (r *AIReviewRepository) CreateJob(job *models.AIReviewJob) error {
	query := `
		INSERT INTO ai_review_jobs (
			status, run_at, max_count, source_statuses, model, prompt_version, created_by,
			total_tasks, completed_tasks, failed_tasks, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0, 0, 0, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		job.Status,
		job.RunAt,
		job.MaxCount,
		pq.Array(job.SourceStatuses),
		job.Model,
		job.PromptVersion,
		job.CreatedBy,
	).Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)
}

func (r *AIReviewRepository) GetJobByID(id int) (*models.AIReviewJob, error) {
	query := `
		SELECT id, status, run_at, max_count, source_statuses, model, prompt_version,
		       created_by, total_tasks, completed_tasks, failed_tasks,
		       created_at, updated_at, started_at, completed_at, archived_at
		FROM ai_review_jobs
		WHERE id = $1
	`
	var job models.AIReviewJob
	var runAt sql.NullTime
	var startedAt sql.NullTime
	var completedAt sql.NullTime
	var archivedAt sql.NullTime
	var model sql.NullString
	var promptVersion sql.NullString
	var createdBy sql.NullInt64
	err := r.db.QueryRow(query, id).Scan(
		&job.ID,
		&job.Status,
		&runAt,
		&job.MaxCount,
		pq.Array(&job.SourceStatuses),
		&model,
		&promptVersion,
		&createdBy,
		&job.TotalTasks,
		&job.CompletedTasks,
		&job.FailedTasks,
		&job.CreatedAt,
		&job.UpdatedAt,
		&startedAt,
		&completedAt,
		&archivedAt,
	)
	if err != nil {
		return nil, err
	}
	if runAt.Valid {
		job.RunAt = &runAt.Time
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}
	if archivedAt.Valid {
		job.ArchivedAt = &archivedAt.Time
	}
	if model.Valid {
		job.Model = &model.String
	}
	if promptVersion.Valid {
		job.PromptVersion = &promptVersion.String
	}
	if createdBy.Valid {
		createdID := int(createdBy.Int64)
		job.CreatedBy = &createdID
	}
	return &job, nil
}

func (r *AIReviewRepository) ListJobs(page, pageSize int, includeArchived bool) ([]models.AIReviewJob, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM ai_review_jobs`
	if !includeArchived {
		countQuery += ` WHERE archived_at IS NULL`
	}
	if err := r.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, status, run_at, max_count, source_statuses, model, prompt_version,
		       created_by, total_tasks, completed_tasks, failed_tasks,
		       created_at, updated_at, started_at, completed_at, archived_at
		FROM ai_review_jobs
		WHERE ($3 OR archived_at IS NULL)
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(query, pageSize, offset, includeArchived)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var jobs []models.AIReviewJob
	for rows.Next() {
		var job models.AIReviewJob
		var runAt sql.NullTime
		var startedAt sql.NullTime
		var completedAt sql.NullTime
		var archivedAt sql.NullTime
		var model sql.NullString
		var promptVersion sql.NullString
		var createdBy sql.NullInt64
		if err := rows.Scan(
			&job.ID,
			&job.Status,
			&runAt,
			&job.MaxCount,
			pq.Array(&job.SourceStatuses),
			&model,
			&promptVersion,
			&createdBy,
			&job.TotalTasks,
			&job.CompletedTasks,
			&job.FailedTasks,
			&job.CreatedAt,
			&job.UpdatedAt,
			&startedAt,
			&completedAt,
			&archivedAt,
		); err != nil {
			return nil, 0, err
		}
		if runAt.Valid {
			job.RunAt = &runAt.Time
		}
		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}
		if archivedAt.Valid {
			job.ArchivedAt = &archivedAt.Time
		}
		if model.Valid {
			job.Model = &model.String
		}
		if promptVersion.Valid {
			job.PromptVersion = &promptVersion.String
		}
		if createdBy.Valid {
			createdID := int(createdBy.Int64)
			job.CreatedBy = &createdID
		}
		jobs = append(jobs, job)
	}

	return jobs, total, nil
}

func (r *AIReviewRepository) GetJobCounts(jobID int) (int, int, int, error) {
	query := `
		SELECT completed_tasks, failed_tasks, total_tasks
		FROM ai_review_jobs
		WHERE id = $1
	`
	var completed int
	var failed int
	var total int
	if err := r.db.QueryRow(query, jobID).Scan(&completed, &failed, &total); err != nil {
		return 0, 0, 0, err
	}
	return completed, failed, total, nil
}

func (r *AIReviewRepository) ArchiveJob(jobID int, archived bool) error {
	query := `
		UPDATE ai_review_jobs
		SET archived_at = $1, updated_at = NOW()
		WHERE id = $2
	`
	var archivedAt interface{}
	if archived {
		now := time.Now()
		archivedAt = now
	}
	_, err := r.db.Exec(query, archivedAt, jobID)
	return err
}

func (r *AIReviewRepository) ListReadyScheduledJobs() ([]int, error) {
	query := `
		SELECT id
		FROM ai_review_jobs
		WHERE status = 'scheduled' AND run_at IS NOT NULL AND run_at <= NOW()
		ORDER BY run_at ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		jobIDs = append(jobIDs, id)
	}
	return jobIDs, nil
}

func (r *AIReviewRepository) UpdateJobStatus(jobID int, status string, allowed []string, startedAt *time.Time, completedAt *time.Time) (bool, error) {
	query := `
		UPDATE ai_review_jobs
		SET status = $1,
		    started_at = COALESCE($2, started_at),
		    completed_at = COALESCE($3, completed_at),
		    updated_at = NOW()
		WHERE id = $4 AND status = ANY($5)
	`
	result, err := r.db.Exec(query, status, startedAt, completedAt, jobID, pq.Array(allowed))
	if err != nil {
		return false, err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return updated > 0, nil
}

func (r *AIReviewRepository) UpdateJobTotals(jobID int, total int) error {
	query := `
		UPDATE ai_review_jobs
		SET total_tasks = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, total, jobID)
	return err
}

func (r *AIReviewRepository) IncrementJobCounts(jobID int, completedDelta, failedDelta int) error {
	query := `
		UPDATE ai_review_jobs
		SET completed_tasks = completed_tasks + $1,
		    failed_tasks = failed_tasks + $2,
		    updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, completedDelta, failedDelta, jobID)
	return err
}

func (r *AIReviewRepository) ResetJobCounts(jobID int) error {
	query := `
		UPDATE ai_review_jobs
		SET total_tasks = 0,
		    completed_tasks = 0,
		    failed_tasks = 0,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, jobID)
	return err
}

func (r *AIReviewRepository) EnqueueTasks(jobID int, statuses []string, maxCount int) (int, error) {
	query := `
		INSERT INTO ai_review_tasks (job_id, review_task_id, comment_id, status, created_at, updated_at)
		SELECT $1, rt.id, rt.comment_id, 'pending', NOW(), NOW()
		FROM review_tasks rt
		WHERE rt.status = ANY($2)
		ORDER BY rt.created_at DESC
		LIMIT $3
		ON CONFLICT (job_id, review_task_id) DO NOTHING
		RETURNING id
	`
	rows, err := r.db.Query(query, jobID, pq.Array(statuses), maxCount)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	return count, nil
}

func (r *AIReviewRepository) ClaimPendingTasks(jobID int, limit int) ([]AIReviewTaskPayload, error) {
	query := `
		WITH claimed AS (
			UPDATE ai_review_tasks
			SET status = 'in_progress',
			    started_at = NOW(),
			    attempts = attempts + 1,
			    updated_at = NOW()
			WHERE id IN (
				SELECT id
				FROM ai_review_tasks
				WHERE job_id = $1 AND status = 'pending'
				ORDER BY id
				LIMIT $2
				FOR UPDATE SKIP LOCKED
			)
			RETURNING id, review_task_id, comment_id
		)
		SELECT c.id, c.review_task_id, c.comment_id, COALESCE(cm.text, '')
		FROM claimed c
		LEFT JOIN comment cm ON cm.id = c.comment_id
	`
	rows, err := r.db.Query(query, jobID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []AIReviewTaskPayload
	for rows.Next() {
		var task AIReviewTaskPayload
		if err := rows.Scan(&task.ID, &task.ReviewTaskID, &task.CommentID, &task.CommentText); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *AIReviewRepository) MarkTaskCompleted(taskID int) error {
	query := `
		UPDATE ai_review_tasks
		SET status = 'completed', completed_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, taskID)
	return err
}

func (r *AIReviewRepository) MarkTaskFailed(taskID int, errorMessage string) error {
	query := `
		UPDATE ai_review_tasks
		SET status = 'failed', error_message = $1, completed_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, errorMessage, taskID)
	return err
}

func (r *AIReviewRepository) DeleteTasksByJob(jobID int) (int, error) {
	query := `DELETE FROM ai_review_tasks WHERE job_id = $1`
	result, err := r.db.Exec(query, jobID)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rowsAffected), nil
}

func (r *AIReviewRepository) DeleteDiffTasksByJob(jobID int) error {
	query := `
		DELETE FROM ai_human_diff_tasks
		WHERE ai_review_result_id IN (
			SELECT ar.id
			FROM ai_review_results ar
			JOIN ai_review_tasks at ON at.id = ar.task_id
			WHERE at.job_id = $1
		)
	`
	_, err := r.db.Exec(query, jobID)
	return err
}

func (r *AIReviewRepository) CreateResult(result *models.AIReviewResult) error {
	query := `
		INSERT INTO ai_review_results (task_id, is_approved, tags, reason, confidence, raw_output, model, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(
		query,
		result.TaskID,
		result.IsApproved,
		pq.Array(result.Tags),
		result.Reason,
		result.Confidence,
		result.RawOutput,
		result.Model,
	).Scan(&result.ID, &result.CreatedAt)
}

func (r *AIReviewRepository) GetComparisonSummary(jobID *int) (models.AIReviewComparison, error) {
	query := `
		SELECT
			COUNT(*) AS total_ai_results,
			COUNT(*) FILTER (WHERE rr.id IS NOT NULL) AS comparable_count,
			COUNT(*) FILTER (WHERE rr.id IS NULL) AS pending_compare_count,
			COUNT(*) FILTER (WHERE rr.id IS NOT NULL AND rr.is_approved = ar.is_approved) AS decision_match_count,
			COUNT(*) FILTER (WHERE rr.id IS NOT NULL AND rr.is_approved <> ar.is_approved) AS decision_mismatch_count,
			COUNT(*) FILTER (WHERE rr.id IS NOT NULL AND rr.is_approved = false AND ar.is_approved = false) AS tag_comparable_count,
			COUNT(*) FILTER (WHERE rr.id IS NOT NULL AND rr.is_approved = false AND ar.is_approved = false AND rr.tags IS NOT NULL AND ar.tags IS NOT NULL AND rr.tags && ar.tags) AS tag_overlap_count
		FROM ai_review_tasks art
		JOIN ai_review_results ar ON ar.task_id = art.id
		LEFT JOIN review_results rr ON rr.task_id = art.review_task_id
		WHERE ($1::int IS NULL OR art.job_id = $1)
	`
	var summary models.AIReviewComparison
	var jobIDValue sql.NullInt64
	if jobID != nil {
		jobIDValue = sql.NullInt64{Int64: int64(*jobID), Valid: true}
	}

	err := r.db.QueryRow(query, jobIDValue).Scan(
		&summary.TotalAIResults,
		&summary.ComparableCount,
		&summary.PendingCompareCount,
		&summary.DecisionMatchCount,
		&summary.DecisionMismatchCount,
		&summary.TagComparableCount,
		&summary.TagOverlapCount,
	)
	return summary, err
}

func (r *AIReviewRepository) GetDiffSamples(jobID *int, limit int) ([]models.AIReviewDiffSample, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `
		SELECT
			art.review_task_id,
			art.comment_id,
			cm.text,
			rr.is_approved,
			rr.tags,
			rr.reason,
			ar.is_approved,
			ar.tags,
			ar.reason,
			ar.confidence
		FROM ai_review_tasks art
		JOIN ai_review_results ar ON ar.task_id = art.id
		LEFT JOIN review_results rr ON rr.task_id = art.review_task_id
		JOIN comment cm ON cm.id = art.comment_id
		WHERE ($1::int IS NULL OR art.job_id = $1)
		  AND rr.id IS NOT NULL
		  AND rr.is_approved <> ar.is_approved
		ORDER BY ar.created_at DESC
		LIMIT $2
	`

	var jobIDValue sql.NullInt64
	if jobID != nil {
		jobIDValue = sql.NullInt64{Int64: int64(*jobID), Valid: true}
	}

	rows, err := r.db.Query(query, jobIDValue, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var diffs []models.AIReviewDiffSample
	for rows.Next() {
		var diff models.AIReviewDiffSample
		var humanTags []string
		if err := rows.Scan(
			&diff.ReviewTaskID,
			&diff.CommentID,
			&diff.CommentText,
			&diff.HumanApproved,
			pq.Array(&humanTags),
			&diff.HumanReason,
			&diff.AIApproved,
			pq.Array(&diff.AITags),
			&diff.AIReason,
			&diff.Confidence,
		); err != nil {
			return nil, err
		}
		diff.HumanTags = humanTags
		diffs = append(diffs, diff)
	}
	return diffs, nil
}

func (r *AIReviewRepository) ListTasksByJob(jobID int, page, pageSize int) ([]models.AIReviewTask, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM ai_review_tasks WHERE job_id = $1`, jobID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT
			art.id,
			art.job_id,
			art.review_task_id,
			art.comment_id,
			art.status,
			art.attempts,
			art.error_message,
			art.started_at,
			art.completed_at,
			art.created_at,
			art.updated_at,
			cm.text,
			ar.id,
			ar.is_approved,
			ar.tags,
			ar.reason,
			ar.confidence,
			ar.raw_output,
			ar.model,
			ar.created_at
		FROM ai_review_tasks art
		LEFT JOIN comment cm ON cm.id = art.comment_id
		LEFT JOIN ai_review_results ar ON ar.task_id = art.id
		WHERE art.job_id = $1
		ORDER BY art.id DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, jobID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []models.AIReviewTask
	for rows.Next() {
		var task models.AIReviewTask
		var errorMessage sql.NullString
		var startedAt sql.NullTime
		var completedAt sql.NullTime
		var commentText sql.NullString
		var resultID sql.NullInt64
		var resultApproved sql.NullBool
		var resultTags []string
		var resultReason sql.NullString
		var resultConfidence sql.NullInt64
		var resultRaw sql.NullString
		var resultModel sql.NullString
		var resultCreatedAt sql.NullTime

		if err := rows.Scan(
			&task.ID,
			&task.JobID,
			&task.ReviewTaskID,
			&task.CommentID,
			&task.Status,
			&task.Attempts,
			&errorMessage,
			&startedAt,
			&completedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
			&commentText,
			&resultID,
			&resultApproved,
			pq.Array(&resultTags),
			&resultReason,
			&resultConfidence,
			&resultRaw,
			&resultModel,
			&resultCreatedAt,
		); err != nil {
			return nil, 0, err
		}

		if errorMessage.Valid {
			task.ErrorMessage = &errorMessage.String
		}
		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		if commentText.Valid {
			task.CommentText = &commentText.String
		}

		if resultID.Valid {
			result := &models.AIReviewResult{
				ID:         int(resultID.Int64),
				TaskID:     task.ID,
				IsApproved: resultApproved.Valid && resultApproved.Bool,
				Tags:       resultTags,
				Reason:     resultReason.String,
				Confidence: int(resultConfidence.Int64),
				CreatedAt:  resultCreatedAt.Time,
			}
			if resultRaw.Valid {
				result.RawOutput = &resultRaw.String
			}
			if resultModel.Valid {
				result.Model = &resultModel.String
			}
			if !resultReason.Valid {
				result.Reason = ""
			}
			task.Result = result
		}

		tasks = append(tasks, task)
	}

	return tasks, total, nil
}
