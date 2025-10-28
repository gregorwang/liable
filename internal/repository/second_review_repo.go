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

type SecondReviewRepository struct {
	db *sql.DB
}

func NewSecondReviewRepository() *SecondReviewRepository {
	return &SecondReviewRepository{db: database.DB}
}

// CreateSecondReviewTask creates a new second review task
func (r *SecondReviewRepository) CreateSecondReviewTask(firstReviewResultID int, commentID int64) error {
	query := `
		INSERT INTO second_review_tasks (first_review_result_id, comment_id, status, created_at)
		VALUES ($1, $2, 'pending', NOW())
	`
	_, err := r.db.Exec(query, firstReviewResultID, commentID)
	return err
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

// CreateSecondReviewResult creates a second review result
func (r *SecondReviewRepository) CreateSecondReviewResult(result *models.SecondReviewResult) error {
	query := `
		INSERT INTO second_review_results (second_task_id, reviewer_id, is_approved, tags, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, result.SecondTaskID, result.ReviewerID, result.IsApproved,
		pq.Array(result.Tags), result.Reason).Scan(&result.ID, &result.CreatedAt)
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
	// Build WHERE conditions
	var conditions []string
	var args []interface{}
	argPos := 1

	// Only search completed second review tasks
	conditions = append(conditions, "srt.status = 'completed'")

	// Filter by comment_id
	if req.CommentID != nil {
		conditions = append(conditions, fmt.Sprintf("srt.comment_id = $%d", argPos))
		args = append(args, *req.CommentID)
		argPos++
	}

	// Filter by second reviewer username
	if req.ReviewerRTX != "" {
		conditions = append(conditions, fmt.Sprintf("u2.username = $%d", argPos))
		args = append(args, req.ReviewerRTX)
		argPos++
	}

	// Filter by tag_ids (OR condition for tags) - search in both first and second review tags
	if req.TagIDs != "" {
		tagIDs := strings.Split(req.TagIDs, ",")
		conditions = append(conditions, fmt.Sprintf("(rr.tags && $%d OR srr.tags && $%d)", argPos, argPos))
		args = append(args, pq.Array(tagIDs))
		argPos++
	}

	// Filter by review time range
	if req.ReviewStartTime != nil {
		conditions = append(conditions, fmt.Sprintf("srt.completed_at >= $%d", argPos))
		args = append(args, *req.ReviewStartTime)
		argPos++
	}

	if req.ReviewEndTime != nil {
		conditions = append(conditions, fmt.Sprintf("srt.completed_at <= $%d", argPos))
		args = append(args, *req.ReviewEndTime)
		argPos++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT srt.id)
		FROM second_review_tasks srt
		LEFT JOIN second_review_results srr ON srt.id = srr.second_task_id
		LEFT JOIN users u2 ON srt.reviewer_id = u2.id
		LEFT JOIN review_results rr ON srt.first_review_result_id = rr.id
		LEFT JOIN users u1 ON rr.reviewer_id = u1.id
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
			srt.id, srt.comment_id, c.text as comment_text,
			srt.reviewer_id, u2.username,
			srt.status, srt.claimed_at, srt.completed_at, srt.created_at,
			-- First review info
			rr.id as first_review_id, rr.is_approved as first_is_approved, 
			rr.tags as first_tags, rr.reason as first_reason, rr.created_at as first_reviewed_at,
			rr.reviewer_id as first_reviewer_id, u1.username as first_username,
			-- Second review info
			srr.id as second_review_id, srr.is_approved as second_is_approved,
			srr.tags as second_tags, srr.reason as second_reason, srr.created_at as second_reviewed_at
		FROM second_review_tasks srt
		LEFT JOIN second_review_results srr ON srt.id = srr.second_task_id
		LEFT JOIN users u2 ON srt.reviewer_id = u2.id
		LEFT JOIN review_results rr ON srt.first_review_result_id = rr.id
		LEFT JOIN users u1 ON rr.reviewer_id = u1.id
		LEFT JOIN comment c ON srt.comment_id = c.id
		%s
		ORDER BY srt.completed_at DESC
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
		var firstTags, secondTags []string
		var firstUsername, secondUsername sql.NullString
		var firstReviewerID sql.NullInt64

		err := rows.Scan(
			&result.ID, &result.CommentID, &result.CommentText,
			&result.ReviewerID, &secondUsername,
			&result.Status, &result.ClaimedAt, &result.CompletedAt, &result.CreatedAt,
			// First review info
			&result.ReviewID, &result.IsApproved, pq.Array(&firstTags), &result.Reason, &result.ReviewedAt,
			&firstReviewerID, &firstUsername,
			// Second review info
			&result.SecondReviewID, &result.SecondIsApproved,
			pq.Array(&secondTags), &result.SecondReason, &result.SecondReviewedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Set queue type
		result.QueueType = "second"

		// Set first review info
		if firstReviewerID.Valid {
			firstID := int(firstReviewerID.Int64)
			result.FirstReviewerID = &firstID
		}
		if firstUsername.Valid {
			result.FirstUsername = &firstUsername.String
		}
		result.Tags = firstTags

		// Set second review info
		result.SecondTags = secondTags
		if secondUsername.Valid {
			result.SecondUsername = &secondUsername.String
		}

		results = append(results, result)
	}

	return results, total, nil
}
