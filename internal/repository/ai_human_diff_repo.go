package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type AIHumanDiffRepository struct {
	db *sql.DB
}

func NewAIHumanDiffRepository() *AIHumanDiffRepository {
	return &AIHumanDiffRepository{db: database.DB}
}

func (r *AIHumanDiffRepository) upsertDiffTask(reviewTaskID int, commentID int64, reviewResultID int, aiReviewResultID int) error {
	query := `
		INSERT INTO ai_human_diff_tasks (
			review_task_id,
			comment_id,
			review_result_id,
			ai_review_result_id,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, 'pending', NOW(), NOW())
		ON CONFLICT (review_task_id) DO UPDATE SET
			review_result_id = EXCLUDED.review_result_id,
			ai_review_result_id = EXCLUDED.ai_review_result_id,
			comment_id = EXCLUDED.comment_id,
			updated_at = NOW()
		WHERE ai_human_diff_tasks.status = 'pending'
	`
	_, err := r.db.Exec(query, reviewTaskID, commentID, reviewResultID, aiReviewResultID)
	return err
}

func (r *AIHumanDiffRepository) CreateTaskIfMismatchWithAIResult(reviewTaskID int, aiReviewResultID int, aiApproved bool) error {
	query := `
		SELECT rr.id, rr.is_approved, rt.comment_id
		FROM review_tasks rt
		JOIN review_results rr ON rr.task_id = rt.id
		WHERE rt.id = $1
	`
	var reviewResultID int
	var humanApproved bool
	var commentID int64
	if err := r.db.QueryRow(query, reviewTaskID).Scan(&reviewResultID, &humanApproved, &commentID); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	if humanApproved == aiApproved {
		return nil
	}
	return r.upsertDiffTask(reviewTaskID, commentID, reviewResultID, aiReviewResultID)
}

func (r *AIHumanDiffRepository) CreateTaskIfMismatchWithHumanResult(reviewTaskID int, reviewResultID int, humanApproved bool) error {
	query := `
		SELECT ar.id, ar.is_approved, rt.comment_id
		FROM review_tasks rt
		JOIN ai_review_tasks art ON art.review_task_id = rt.id
		JOIN ai_review_results ar ON ar.task_id = art.id
		WHERE rt.id = $1
		ORDER BY ar.created_at DESC
		LIMIT 1
	`
	var aiReviewResultID int
	var aiApproved bool
	var commentID int64
	if err := r.db.QueryRow(query, reviewTaskID).Scan(&aiReviewResultID, &aiApproved, &commentID); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	if aiApproved == humanApproved {
		return nil
	}
	return r.upsertDiffTask(reviewTaskID, commentID, reviewResultID, aiReviewResultID)
}

func (r *AIHumanDiffRepository) ClaimDiffTasks(reviewerID int, limit int) ([]models.AIHumanDiffTask, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT id, review_task_id, comment_id, review_result_id, ai_review_result_id, created_at
		FROM ai_human_diff_tasks
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	rows, err := tx.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	taskIDs := []int{}
	tasks := []models.AIHumanDiffTask{}
	for rows.Next() {
		var task models.AIHumanDiffTask
		if err := rows.Scan(
			&task.ID,
			&task.ReviewTaskID,
			&task.CommentID,
			&task.ReviewResultID,
			&task.AIReviewResultID,
			&task.CreatedAt,
		); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, task.ID)
		tasks = append(tasks, task)
	}

	if len(taskIDs) == 0 {
		return []models.AIHumanDiffTask{}, nil
	}

	now := time.Now()
	updateQuery := `
		UPDATE ai_human_diff_tasks
		SET status = 'in_progress', reviewer_id = $1, claimed_at = $2
		WHERE id = ANY($3)
	`
	if _, err := tx.Exec(updateQuery, reviewerID, now, pq.Array(taskIDs)); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.FindDiffTasksWithDetails(taskIDs)
}

func (r *AIHumanDiffRepository) FindDiffTasksWithDetails(taskIDs []int) ([]models.AIHumanDiffTask, error) {
	query := `
		SELECT
			dt.id, dt.review_task_id, dt.comment_id, dt.review_result_id, dt.ai_review_result_id,
			dt.reviewer_id, dt.status, dt.claimed_at, dt.completed_at, dt.created_at, dt.updated_at,
			c.id, c.text,
			rr.id, rr.task_id, rr.reviewer_id, rr.is_approved, rr.tags, rr.reason, rr.created_at,
			u.id, u.username,
			ar.id, ar.task_id, ar.is_approved, ar.tags, ar.reason, ar.confidence, ar.raw_output, ar.model, ar.created_at
		FROM ai_human_diff_tasks dt
		JOIN comment c ON dt.comment_id = c.id
		JOIN review_results rr ON dt.review_result_id = rr.id
		LEFT JOIN users u ON rr.reviewer_id = u.id
		JOIN ai_review_results ar ON dt.ai_review_result_id = ar.id
		WHERE dt.id = ANY($1)
		ORDER BY dt.id
	`

	rows, err := r.db.Query(query, pq.Array(taskIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.AIHumanDiffTask{}
	for rows.Next() {
		var task models.AIHumanDiffTask
		var comment models.Comment
		var humanResult models.ReviewResult
		var aiResult models.AIReviewResult
		var humanReviewerID sql.NullInt64
		var reviewerName sql.NullString
		var taskReviewerID sql.NullInt64
		var claimedAt sql.NullTime
		var completedAt sql.NullTime
		var humanReason sql.NullString
		var aiReason sql.NullString
		var aiRaw sql.NullString
		var aiModel sql.NullString

		if err := rows.Scan(
			&task.ID,
			&task.ReviewTaskID,
			&task.CommentID,
			&task.ReviewResultID,
			&task.AIReviewResultID,
			&taskReviewerID,
			&task.Status,
			&claimedAt,
			&completedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
			&comment.ID,
			&comment.Text,
			&humanResult.ID,
			&humanResult.TaskID,
			&humanResult.ReviewerID,
			&humanResult.IsApproved,
			pq.Array(&humanResult.Tags),
			&humanReason,
			&humanResult.CreatedAt,
			&humanReviewerID,
			&reviewerName,
			&aiResult.ID,
			&aiResult.TaskID,
			&aiResult.IsApproved,
			pq.Array(&aiResult.Tags),
			&aiReason,
			&aiResult.Confidence,
			&aiRaw,
			&aiModel,
			&aiResult.CreatedAt,
		); err != nil {
			return nil, err
		}

		if taskReviewerID.Valid {
			id := int(taskReviewerID.Int64)
			task.ReviewerID = &id
		}
		if claimedAt.Valid {
			task.ClaimedAt = &claimedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		if humanReason.Valid {
			humanResult.Reason = humanReason.String
		}
		if humanReviewerID.Valid && reviewerName.Valid {
			humanResult.Reviewer = &models.User{
				ID:       int(humanReviewerID.Int64),
				Username: reviewerName.String,
			}
		}
		if aiReason.Valid {
			aiResult.Reason = aiReason.String
		}
		if aiRaw.Valid {
			raw := aiRaw.String
			aiResult.RawOutput = &raw
		}
		if aiModel.Valid {
			model := aiModel.String
			aiResult.Model = &model
		}

		task.Comment = &comment
		task.HumanReview = &humanResult
		task.AIReview = &aiResult
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *AIHumanDiffRepository) GetMyDiffTasks(reviewerID int) ([]models.AIHumanDiffTask, error) {
	query := `
		SELECT
			dt.id, dt.review_task_id, dt.comment_id, dt.review_result_id, dt.ai_review_result_id,
			dt.reviewer_id, dt.status, dt.claimed_at, dt.completed_at, dt.created_at, dt.updated_at,
			c.id, c.text,
			rr.id, rr.task_id, rr.reviewer_id, rr.is_approved, rr.tags, rr.reason, rr.created_at,
			u.id, u.username,
			ar.id, ar.task_id, ar.is_approved, ar.tags, ar.reason, ar.confidence, ar.raw_output, ar.model, ar.created_at
		FROM ai_human_diff_tasks dt
		JOIN comment c ON dt.comment_id = c.id
		JOIN review_results rr ON dt.review_result_id = rr.id
		LEFT JOIN users u ON rr.reviewer_id = u.id
		JOIN ai_review_results ar ON dt.ai_review_result_id = ar.id
		WHERE dt.reviewer_id = $1 AND dt.status = 'in_progress'
		ORDER BY dt.claimed_at DESC
	`

	rows, err := r.db.Query(query, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.AIHumanDiffTask{}
	for rows.Next() {
		var task models.AIHumanDiffTask
		var comment models.Comment
		var humanResult models.ReviewResult
		var aiResult models.AIReviewResult
		var reviewerID sql.NullInt64
		var reviewerName sql.NullString
		var taskReviewerID sql.NullInt64
		var claimedAt sql.NullTime
		var completedAt sql.NullTime
		var humanReason sql.NullString
		var aiReason sql.NullString
		var aiRaw sql.NullString
		var aiModel sql.NullString

		if err := rows.Scan(
			&task.ID,
			&task.ReviewTaskID,
			&task.CommentID,
			&task.ReviewResultID,
			&task.AIReviewResultID,
			&taskReviewerID,
			&task.Status,
			&claimedAt,
			&completedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
			&comment.ID,
			&comment.Text,
			&humanResult.ID,
			&humanResult.TaskID,
			&humanResult.ReviewerID,
			&humanResult.IsApproved,
			pq.Array(&humanResult.Tags),
			&humanReason,
			&humanResult.CreatedAt,
			&reviewerID,
			&reviewerName,
			&aiResult.ID,
			&aiResult.TaskID,
			&aiResult.IsApproved,
			pq.Array(&aiResult.Tags),
			&aiReason,
			&aiResult.Confidence,
			&aiRaw,
			&aiModel,
			&aiResult.CreatedAt,
		); err != nil {
			return nil, err
		}

		if taskReviewerID.Valid {
			id := int(taskReviewerID.Int64)
			task.ReviewerID = &id
		}
		if claimedAt.Valid {
			task.ClaimedAt = &claimedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		if humanReason.Valid {
			humanResult.Reason = humanReason.String
		}
		if reviewerID.Valid && reviewerName.Valid {
			humanResult.Reviewer = &models.User{
				ID:       int(reviewerID.Int64),
				Username: reviewerName.String,
			}
		}
		if aiReason.Valid {
			aiResult.Reason = aiReason.String
		}
		if aiRaw.Valid {
			raw := aiRaw.String
			aiResult.RawOutput = &raw
		}
		if aiModel.Valid {
			model := aiModel.String
			aiResult.Model = &model
		}

		task.Comment = &comment
		task.HumanReview = &humanResult
		task.AIReview = &aiResult
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *AIHumanDiffRepository) CompleteDiffTask(taskID, reviewerID int) error {
	query := `
		UPDATE ai_human_diff_tasks
		SET status = 'completed', completed_at = COALESCE(completed_at, NOW())
		WHERE id = $1 AND reviewer_id = $2 AND status IN ('in_progress', 'completed')
	`
	result, err := r.db.Exec(query, taskID, reviewerID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *AIHumanDiffRepository) CreateDiffResult(result *models.AIHumanDiffResult) (bool, error) {
	query := `
		INSERT INTO ai_human_diff_results (task_id, reviewer_id, is_approved, tags, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (task_id) DO NOTHING
		RETURNING id, created_at
	`
	err := r.db.QueryRow(query, result.TaskID, result.ReviewerID, result.IsApproved,
		pq.Array(result.Tags), result.Reason).Scan(&result.ID, &result.CreatedAt)
	if err == nil {
		return true, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}

	existingQuery := `
		SELECT id, reviewer_id, is_approved, tags, reason, created_at
		FROM ai_human_diff_results
		WHERE task_id = $1
	`
	var tags []string
	err = r.db.QueryRow(existingQuery, result.TaskID).Scan(
		&result.ID,
		&result.ReviewerID,
		&result.IsApproved,
		pq.Array(&tags),
		&result.Reason,
		&result.CreatedAt,
	)
	if err != nil {
		return false, err
	}
	result.Tags = tags
	return false, nil
}

func (r *AIHumanDiffRepository) ReturnDiffTasks(taskIDs []int, reviewerID int) (int, error) {
	query := `
		UPDATE ai_human_diff_tasks
		SET status = 'pending', reviewer_id = NULL, claimed_at = NULL
		WHERE id = ANY($1) AND reviewer_id = $2 AND status = 'in_progress'
	`
	result, err := r.db.Exec(query, pq.Array(taskIDs), reviewerID)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rowsAffected), nil
}

func (r *AIHumanDiffRepository) FindExpiredDiffTasks(timeoutMinutes int) ([]models.AIHumanDiffTask, error) {
	query := `
		SELECT id, reviewer_id, claimed_at
		FROM ai_human_diff_tasks
		WHERE status = 'in_progress'
		AND claimed_at < NOW() - INTERVAL '1 minute' * $1
	`
	rows, err := r.db.Query(query, timeoutMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.AIHumanDiffTask{}
	for rows.Next() {
		var task models.AIHumanDiffTask
		var reviewerID sql.NullInt64
		var claimedAt sql.NullTime
		if err := rows.Scan(&task.ID, &reviewerID, &claimedAt); err != nil {
			return nil, err
		}
		if reviewerID.Valid {
			id := int(reviewerID.Int64)
			task.ReviewerID = &id
		}
		if claimedAt.Valid {
			task.ClaimedAt = &claimedAt.Time
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *AIHumanDiffRepository) ResetDiffTask(taskID int) error {
	query := `
		UPDATE ai_human_diff_tasks
		SET status = 'pending', reviewer_id = NULL, claimed_at = NULL
		WHERE id = $1
	`
	_, err := r.db.Exec(query, taskID)
	return err
}
