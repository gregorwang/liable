package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"strings"
	"time"

	"github.com/lib/pq"
)

type SecondReviewRepository struct {
	db *sql.DB
}

func NewSecondReviewRepository() *SecondReviewRepository {
	return &SecondReviewRepository{db: database.DB}
}

// CreateSecondReviewTask creates a new second review task
func (r *SecondReviewRepository) CreateSecondReviewTask(firstReviewResultID int, commentID int64) (bool, error) {
	return createSecondReviewTask(r.db, firstReviewResultID, commentID)
}

// CreateSecondReviewTaskTx creates a new second review task within a transaction
func (r *SecondReviewRepository) CreateSecondReviewTaskTx(tx *sql.Tx, firstReviewResultID int, commentID int64) (bool, error) {
	return createSecondReviewTask(tx, firstReviewResultID, commentID)
}

type secondReviewExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func createSecondReviewTask(db secondReviewExecutor, firstReviewResultID int, commentID int64) (bool, error) {
	query := `
		INSERT INTO second_review_tasks (first_review_result_id, comment_id, status, created_at)
		VALUES ($1, $2, 'pending', NOW())
		ON CONFLICT (first_review_result_id) DO NOTHING
	`
	result, err := db.Exec(query, firstReviewResultID, commentID)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// GetCommentIDByTaskID retrieves the comment ID for a second review task.
func (r *SecondReviewRepository) GetCommentIDByTaskID(taskID int) (int64, error) {
	query := `SELECT comment_id FROM second_review_tasks WHERE id = $1`
	var commentID int64
	if err := r.db.QueryRow(query, taskID).Scan(&commentID); err != nil {
		return 0, err
	}
	return commentID, nil
}

// ClaimSecondReviewTasks claims pending second review tasks for a reviewer
func (r *SecondReviewRepository) ClaimSecondReviewTasks(reviewerID int, limit int) ([]models.SecondReviewTask, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Select pending second review tasks
	query := `
		SELECT id, first_review_result_id, comment_id, created_at
		FROM second_review_tasks
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
	tasks := []models.SecondReviewTask{}

	for rows.Next() {
		var task models.SecondReviewTask
		if err := rows.Scan(&task.ID, &task.FirstReviewResultID, &task.CommentID, &task.CreatedAt); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, task.ID)
		tasks = append(tasks, task)
	}

	if len(taskIDs) == 0 {
		return []models.SecondReviewTask{}, nil
	}

	// Update tasks to in_progress
	now := time.Now()
	updateQuery := `
		UPDATE second_review_tasks
		SET status = 'in_progress', reviewer_id = $1, claimed_at = $2
		WHERE id = ANY($3)
	`
	_, err = tx.Exec(updateQuery, reviewerID, now, pq.Array(taskIDs))
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Fetch full task details with comments and first review results
	return r.FindSecondReviewTasksWithDetails(taskIDs)
}

// FindSecondReviewTasksWithDetails finds second review tasks with their associated comments and first review results
func (r *SecondReviewRepository) FindSecondReviewTasksWithDetails(taskIDs []int) ([]models.SecondReviewTask, error) {
	query := `
		SELECT 
			srt.id, srt.first_review_result_id, srt.comment_id, srt.reviewer_id, srt.status, 
			srt.claimed_at, srt.completed_at, srt.created_at,
			c.id, c.text,
			rr.id, rr.reviewer_id, rr.is_approved, rr.tags, rr.reason, rr.created_at,
			u.id, u.username
		FROM second_review_tasks srt
		INNER JOIN comment c ON srt.comment_id = c.id
		INNER JOIN review_results rr ON srt.first_review_result_id = rr.id
		LEFT JOIN users u ON rr.reviewer_id = u.id
		WHERE srt.id = ANY($1)
		ORDER BY srt.id
	`
	rows, err := r.db.Query(query, pq.Array(taskIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.SecondReviewTask{}
	for rows.Next() {
		var task models.SecondReviewTask
		var comment models.Comment
		var firstReviewResult models.ReviewResult
		var reviewerID sql.NullInt64
		var username sql.NullString
		err := rows.Scan(
			&task.ID, &task.FirstReviewResultID, &task.CommentID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&comment.ID, &comment.Text,
			&firstReviewResult.ID, &firstReviewResult.ReviewerID, &firstReviewResult.IsApproved,
			pq.Array(&firstReviewResult.Tags), &firstReviewResult.Reason, &firstReviewResult.CreatedAt,
			&reviewerID, &username,
		)
		if err != nil {
			return nil, err
		}
		task.Comment = &comment

		// Add reviewer information to first review result
		if reviewerID.Valid && username.Valid {
			firstReviewResult.Reviewer = &models.User{
				ID:       int(reviewerID.Int64),
				Username: username.String,
			}
		}

		task.FirstReviewResult = &firstReviewResult
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// GetMySecondReviewTasks gets all in-progress second review tasks for a reviewer
func (r *SecondReviewRepository) GetMySecondReviewTasks(reviewerID int) ([]models.SecondReviewTask, error) {
	query := `
		SELECT 
			srt.id, srt.first_review_result_id, srt.comment_id, srt.reviewer_id, srt.status, 
			srt.claimed_at, srt.completed_at, srt.created_at,
			c.id, c.text,
			rr.id, rr.reviewer_id, rr.is_approved, rr.tags, rr.reason, rr.created_at,
			u.id, u.username
		FROM second_review_tasks srt
		INNER JOIN comment c ON srt.comment_id = c.id
		INNER JOIN review_results rr ON srt.first_review_result_id = rr.id
		LEFT JOIN users u ON rr.reviewer_id = u.id
		WHERE srt.reviewer_id = $1 AND srt.status = 'in_progress'
		ORDER BY srt.claimed_at DESC
	`
	rows, err := r.db.Query(query, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.SecondReviewTask{}
	for rows.Next() {
		var task models.SecondReviewTask
		var comment models.Comment
		var firstReviewResult models.ReviewResult
		var reviewerID sql.NullInt64
		var username sql.NullString
		err := rows.Scan(
			&task.ID, &task.FirstReviewResultID, &task.CommentID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&comment.ID, &comment.Text,
			&firstReviewResult.ID, &firstReviewResult.ReviewerID, &firstReviewResult.IsApproved,
			pq.Array(&firstReviewResult.Tags), &firstReviewResult.Reason, &firstReviewResult.CreatedAt,
			&reviewerID, &username,
		)
		if err != nil {
			return nil, err
		}
		task.Comment = &comment

		// Add reviewer information to first review result
		if reviewerID.Valid && username.Valid {
			firstReviewResult.Reviewer = &models.User{
				ID:       int(reviewerID.Int64),
				Username: username.String,
			}
		}

		task.FirstReviewResult = &firstReviewResult
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// CompleteSecondReviewTask marks a second review task as completed
func (r *SecondReviewRepository) CompleteSecondReviewTask(taskID, reviewerID int) error {
	query := `
		UPDATE second_review_tasks
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

// CreateSecondReviewResult creates a second review result
func (r *SecondReviewRepository) CreateSecondReviewResult(result *models.SecondReviewResult) (bool, error) {
	query := `
		INSERT INTO second_review_results (second_task_id, reviewer_id, is_approved, tags, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (second_task_id) DO NOTHING
		RETURNING id, created_at
	`
	err := r.db.QueryRow(query, result.SecondTaskID, result.ReviewerID, result.IsApproved,
		pq.Array(result.Tags), result.Reason).Scan(&result.ID, &result.CreatedAt)
	if err == nil {
		return true, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}

	existingQuery := `
		SELECT id, reviewer_id, is_approved, tags, reason, created_at
		FROM second_review_results
		WHERE second_task_id = $1
	`
	var tags []string
	err = r.db.QueryRow(existingQuery, result.SecondTaskID).Scan(
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

// ReturnSecondReviewTasks returns multiple second review tasks back to pending status for a specific reviewer
func (r *SecondReviewRepository) ReturnSecondReviewTasks(taskIDs []int, reviewerID int) (int, error) {
	query := `
		UPDATE second_review_tasks
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

// FindExpiredSecondReviewTasks finds second review tasks that have been in progress for too long
func (r *SecondReviewRepository) FindExpiredSecondReviewTasks(timeoutMinutes int) ([]models.SecondReviewTask, error) {
	query := `
		SELECT id, comment_id, reviewer_id, claimed_at
		FROM second_review_tasks
		WHERE status = 'in_progress' 
		AND claimed_at < NOW() - INTERVAL '1 minute' * $1
	`
	rows, err := r.db.Query(query, timeoutMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.SecondReviewTask{}
	for rows.Next() {
		var task models.SecondReviewTask
		if err := rows.Scan(&task.ID, &task.CommentID, &task.ReviewerID, &task.ClaimedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// ResetSecondReviewTask resets a second review task back to pending status
func (r *SecondReviewRepository) ResetSecondReviewTask(taskID int) error {
	query := `
		UPDATE second_review_tasks
		SET status = 'pending', reviewer_id = NULL, claimed_at = NULL
		WHERE id = $1
	`
	_, err := r.db.Exec(query, taskID)
	return err
}

// SearchSecondReviewTasks searches second review tasks with filters and pagination
func (r *SecondReviewRepository) SearchSecondReviewTasks(req models.SearchTasksRequest) ([]models.TaskSearchResult, int, error) {
	if strings.TrimSpace(req.QueueName) == "" {
		req.QueueName = "comment_second_review"
	}

	taskRepo := NewTaskRepository()
	return taskRepo.SearchTasksUnified(req)
}

