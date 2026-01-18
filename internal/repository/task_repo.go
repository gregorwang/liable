package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"fmt"
	"strings"
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

// ReturnTasks returns multiple tasks back to pending status for a specific reviewer
func (r *TaskRepository) ReturnTasks(taskIDs []int, reviewerID int) (int, error) {
	query := `
		UPDATE review_tasks
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

// SearchTasks searches review tasks with filters and pagination
func (r *TaskRepository) SearchTasks(req models.SearchTasksRequest) ([]models.TaskSearchResult, int, error) {
	// Build WHERE conditions
	var conditions []string
	var args []interface{}
	argPos := 1

	// Only search completed tasks
	conditions = append(conditions, "rt.status = 'completed'")

	// Filter by comment_id
	if req.CommentID != nil {
		conditions = append(conditions, fmt.Sprintf("rt.comment_id = $%d", argPos))
		args = append(args, *req.CommentID)
		argPos++
	}

	// Filter by reviewer username
	if req.ReviewerRTX != "" {
		conditions = append(conditions, fmt.Sprintf("u.username = $%d", argPos))
		args = append(args, req.ReviewerRTX)
		argPos++
	}

	// Filter by tag_ids (OR condition for tags)
	if req.TagIDs != "" {
		// Split comma-separated tag IDs
		tagIDs := strings.Split(req.TagIDs, ",")
		conditions = append(conditions, fmt.Sprintf("rr.tags && $%d", argPos))
		args = append(args, pq.Array(tagIDs))
		argPos++
	}

	// Filter by review time range
	if req.ReviewStartTime != nil {
		conditions = append(conditions, fmt.Sprintf("rt.completed_at >= $%d", argPos))
		args = append(args, *req.ReviewStartTime)
		argPos++
	}

	if req.ReviewEndTime != nil {
		conditions = append(conditions, fmt.Sprintf("rt.completed_at <= $%d", argPos))
		args = append(args, *req.ReviewEndTime)
		argPos++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT rt.id)
		FROM review_tasks rt
		LEFT JOIN review_results rr ON rt.id = rr.task_id
		LEFT JOIN users u ON rt.reviewer_id = u.id
		%s
	`, whereClause)

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query with pagination
	offset := (req.Page - 1) * req.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			rt.id, rt.comment_id, c.text as comment_text,
			rt.reviewer_id, u.username,
			rt.status, rt.claimed_at, rt.completed_at, rt.created_at,
			rr.id as review_id, rr.is_approved, rr.tags, rr.reason, rr.created_at as reviewed_at,
			'first' as queue_type
		FROM review_tasks rt
		LEFT JOIN review_results rr ON rt.id = rr.task_id
		LEFT JOIN users u ON rt.reviewer_id = u.id
		LEFT JOIN comment c ON rt.comment_id = c.id
		%s
		ORDER BY rt.completed_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	args = append(args, req.PageSize, offset)
	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results := []models.TaskSearchResult{}
	for rows.Next() {
		var result models.TaskSearchResult
		err := rows.Scan(
			&result.ID, &result.CommentID, &result.CommentText,
			&result.ReviewerID, &result.Username,
			&result.Status, &result.ClaimedAt, &result.CompletedAt, &result.CreatedAt,
			&result.ReviewID, &result.IsApproved, pq.Array(&result.Tags), &result.Reason, &result.ReviewedAt,
			&result.QueueType,
		)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, result)
	}

	return results, total, nil
}

// SearchTasksUnified searches review tasks with database-level sorting and pagination
// This is optimized to handle both first and second review queues in a single query
func (r *TaskRepository) SearchTasksUnified(req models.SearchTasksRequest) ([]models.TaskSearchResult, int, error) {
	// Build WHERE conditions
	var conditions []string
	var args []interface{}
	argPos := 1

	// Only search completed tasks
	conditions = append(conditions, "rt.status = 'completed'")

	// Filter by comment_id
	if req.CommentID != nil {
		conditions = append(conditions, fmt.Sprintf("rt.comment_id = $%d", argPos))
		args = append(args, *req.CommentID)
		argPos++
	}

	// Filter by reviewer username
	if req.ReviewerRTX != "" {
		conditions = append(conditions, fmt.Sprintf("u.username = $%d", argPos))
		args = append(args, req.ReviewerRTX)
		argPos++
	}

	// Filter by tag_ids (OR condition for tags)
	if req.TagIDs != "" {
		tagIDs := strings.Split(req.TagIDs, ",")
		conditions = append(conditions, fmt.Sprintf("rr.tags && $%d", argPos))
		args = append(args, pq.Array(tagIDs))
		argPos++
	}

	// Filter by review time range
	if req.ReviewStartTime != nil {
		conditions = append(conditions, fmt.Sprintf("rt.completed_at >= $%d", argPos))
		args = append(args, *req.ReviewStartTime)
		argPos++
	}

	if req.ReviewEndTime != nil {
		conditions = append(conditions, fmt.Sprintf("rt.completed_at <= $%d", argPos))
		args = append(args, *req.ReviewEndTime)
		argPos++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Determine which queues to include
	includeFirst := req.QueueType == "first" || req.QueueType == "all"
	includeSecond := req.QueueType == "second" || req.QueueType == "all"

	if !includeFirst && !includeSecond {
		// Default to "all" if no valid queue type
		includeFirst = true
		includeSecond = true
	}

	// Build union query
	var unionQuery string
	var queueSources []string

	if includeFirst {
		queueSources = append(queueSources, fmt.Sprintf(`
			SELECT
				rt.id, rt.comment_id, c.text as comment_text,
				rt.reviewer_id, u.username as username,
				rt.status, rt.claimed_at, rt.completed_at, rt.created_at,
				rr.id as review_id, rr.is_approved, rr.tags, rr.reason, rr.created_at as reviewed_at,
				'first' as queue_type,
				NULL as second_review_id, NULL as second_is_approved,
				NULL as second_tags, NULL as second_reason, NULL as second_reviewed_at,
				NULL as second_reviewer_id, NULL as second_username
			FROM review_tasks rt
			LEFT JOIN review_results rr ON rt.id = rr.task_id
			LEFT JOIN users u ON rt.reviewer_id = u.id
			LEFT JOIN comment c ON rt.comment_id = c.id
			%s
		`, whereClause))
	}

	if includeSecond {
		queueSources = append(queueSources, fmt.Sprintf(`
			SELECT
				srt.id, srt.comment_id, c.text as comment_text,
				srt.reviewer_id, u2.username as username,
				srt.status, srt.claimed_at, srt.completed_at, srt.created_at,
				NULL as review_id, NULL as is_approved, NULL as tags, NULL as reason, NULL as reviewed_at,
				'second' as queue_type,
				srr.id as second_review_id, srr.is_approved as second_is_approved,
				srr.tags as second_tags, srr.reason as second_reason, srr.created_at as second_reviewed_at,
				NULL as second_reviewer_id, NULL as second_username
			FROM second_review_tasks srt
			LEFT JOIN second_review_results srr ON srt.id = srr.second_task_id
			LEFT JOIN users u2 ON srt.reviewer_id = u2.id
			LEFT JOIN comment c ON srt.comment_id = c.id
			%s
		`, whereClause))
	}

	// Join queue sources with UNION ALL
	if len(queueSources) == 2 {
		unionQuery = fmt.Sprintf("(%s) UNION ALL (%s)", queueSources[0], queueSources[1])
	} else if len(queueSources) == 1 {
		unionQuery = queueSources[0]
	} else {
		return []models.TaskSearchResult{}, 0, nil
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) combined_results", unionQuery)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query with database-level sorting and pagination
	offset := (req.Page - 1) * req.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT * FROM (%s) combined_results
		ORDER BY completed_at DESC
		LIMIT $%d OFFSET $%d
	`, unionQuery, argPos, argPos+1)

	args = append(args, req.PageSize, offset)
	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results := []models.TaskSearchResult{}
	for rows.Next() {
		var result models.TaskSearchResult
		var firstTags, secondTags []string
		var _ /*firstUsername*/, secondUsername sql.NullString
		var secondReviewerID sql.NullInt64

		err := rows.Scan(
			&result.ID, &result.CommentID, &result.CommentText,
			&result.ReviewerID, &result.Username,
			&result.Status, &result.ClaimedAt, &result.CompletedAt, &result.CreatedAt,
			&result.ReviewID, &result.IsApproved, pq.Array(&firstTags), &result.Reason, &result.ReviewedAt,
			&result.QueueType,
			&result.SecondReviewID, &result.SecondIsApproved,
			pq.Array(&secondTags), &result.SecondReason, &result.SecondReviewedAt,
			&secondReviewerID, &secondUsername,
		)
		if err != nil {
			return nil, 0, err
		}

		// Set tags for first review tasks
		if result.QueueType == "first" {
			result.Tags = firstTags
		} else {
			// For second review tasks, set first review info if available
			result.SecondTags = secondTags
			if secondUsername.Valid {
				result.SecondUsername = &secondUsername.String
			}
			if secondReviewerID.Valid {
				secondID := int(secondReviewerID.Int64)
				result.SecondReviewerID = &secondID
			}
		}

		results = append(results, result)
	}

	return results, total, nil
}

