package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{db: database.DB}
}

// CreateTask creates a new review task
func (r *TaskRepository) CreateTask(commentID int64) error {
	query := `
		INSERT INTO review_tasks (comment_id, status, created_at)
		VALUES ($1, 'pending', NOW())
	`
	_, err := r.db.Exec(query, commentID)
	return err
}

// ClaimTasks claims pending tasks for a reviewer
func (r *TaskRepository) ClaimTasks(reviewerID int, limit int) ([]models.ReviewTask, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Select pending tasks
	query := `
		SELECT id, comment_id, created_at
		FROM review_tasks
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
	tasks := []models.ReviewTask{}

	for rows.Next() {
		var task models.ReviewTask
		if err := rows.Scan(&task.ID, &task.CommentID, &task.CreatedAt); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, task.ID)
		tasks = append(tasks, task)
	}

	if len(taskIDs) == 0 {
		return []models.ReviewTask{}, nil
	}

	// Update tasks to in_progress
	now := time.Now()
	updateQuery := `
		UPDATE review_tasks
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

	// Fetch full task details with comments
	return r.FindTasksWithComments(taskIDs)
}

// FindTasksWithComments finds tasks with their associated comments
func (r *TaskRepository) FindTasksWithComments(taskIDs []int) ([]models.ReviewTask, error) {
	query := `
		SELECT 
			rt.id, rt.comment_id, rt.reviewer_id, rt.status, 
			rt.claimed_at, rt.completed_at, rt.created_at,
			c.id, c.text
		FROM review_tasks rt
		INNER JOIN comment c ON rt.comment_id = c.id
		WHERE rt.id = ANY($1)
		ORDER BY rt.id
	`
	rows, err := r.db.Query(query, pq.Array(taskIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.ReviewTask{}
	for rows.Next() {
		var task models.ReviewTask
		var comment models.Comment
		err := rows.Scan(
			&task.ID, &task.CommentID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&comment.ID, &comment.Text,
		)
		if err != nil {
			return nil, err
		}
		task.Comment = &comment
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// GetMyTasks gets all in-progress tasks for a reviewer
func (r *TaskRepository) GetMyTasks(reviewerID int) ([]models.ReviewTask, error) {
	query := `
		SELECT 
			rt.id, rt.comment_id, rt.reviewer_id, rt.status, 
			rt.claimed_at, rt.completed_at, rt.created_at,
			c.id, c.text
		FROM review_tasks rt
		INNER JOIN comment c ON rt.comment_id = c.id
		WHERE rt.reviewer_id = $1 AND rt.status = 'in_progress'
		ORDER BY rt.claimed_at DESC
	`
	rows, err := r.db.Query(query, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.ReviewTask{}
	for rows.Next() {
		var task models.ReviewTask
		var comment models.Comment
		err := rows.Scan(
			&task.ID, &task.CommentID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&comment.ID, &comment.Text,
		)
		if err != nil {
			return nil, err
		}
		task.Comment = &comment
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// CompleteTask marks a task as completed
func (r *TaskRepository) CompleteTask(taskID, reviewerID int) error {
	query := `
		UPDATE review_tasks
		SET status = 'completed', completed_at = NOW()
		WHERE id = $1 AND reviewer_id = $2 AND status = 'in_progress'
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

// CreateReviewResult creates a review result
func (r *TaskRepository) CreateReviewResult(result *models.ReviewResult) error {
	query := `
		INSERT INTO review_results (task_id, reviewer_id, is_approved, tags, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, result.TaskID, result.ReviewerID, result.IsApproved,
		pq.Array(result.Tags), result.Reason).Scan(&result.ID, &result.CreatedAt)
}

// CountByStatus counts tasks by status
func (r *TaskRepository) CountByStatus(status string) (int, error) {
	query := `SELECT COUNT(*) FROM review_tasks WHERE status = $1`
	var count int
	err := r.db.QueryRow(query, status).Scan(&count)
	return count, err
}

// FindExpiredTasks finds tasks that have been in progress for too long
func (r *TaskRepository) FindExpiredTasks(timeoutMinutes int) ([]models.ReviewTask, error) {
	query := `
		SELECT id, comment_id, reviewer_id, claimed_at
		FROM review_tasks
		WHERE status = 'in_progress' 
		AND claimed_at < NOW() - INTERVAL '1 minute' * $1
	`
	rows, err := r.db.Query(query, timeoutMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.ReviewTask{}
	for rows.Next() {
		var task models.ReviewTask
		if err := rows.Scan(&task.ID, &task.CommentID, &task.ReviewerID, &task.ClaimedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// ResetTask resets a task back to pending status
func (r *TaskRepository) ResetTask(taskID int) error {
	query := `
		UPDATE review_tasks
		SET status = 'pending', reviewer_id = NULL, claimed_at = NULL
		WHERE id = $1
	`
	_, err := r.db.Exec(query, taskID)
	return err
}
