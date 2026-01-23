package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type QualityCheckRepository struct {
	db *sql.DB
}

func NewQualityCheckRepository() *QualityCheckRepository {
	return &QualityCheckRepository{db: database.DB}
}

// CreateQCTask creates a new quality check task
func (r *QualityCheckRepository) CreateQCTask(firstReviewResultID int, commentID int64) (bool, error) {
	query := `
		INSERT INTO quality_check_tasks (first_review_result_id, comment_id, status, created_at)
		VALUES ($1, $2, 'pending', NOW())
		ON CONFLICT (first_review_result_id) DO NOTHING
	`
	result, err := r.db.Exec(query, firstReviewResultID, commentID)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// ClaimQCTasks claims pending quality check tasks for a reviewer
func (r *QualityCheckRepository) ClaimQCTasks(reviewerID int, limit int) ([]models.QualityCheckTask, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Select pending tasks
	query := `
		SELECT id, first_review_result_id, comment_id, created_at
		FROM quality_check_tasks
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
	tasks := []models.QualityCheckTask{}

	for rows.Next() {
		var task models.QualityCheckTask
		if err := rows.Scan(&task.ID, &task.FirstReviewResultID, &task.CommentID, &task.CreatedAt); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, task.ID)
		tasks = append(tasks, task)
	}

	if len(taskIDs) == 0 {
		return []models.QualityCheckTask{}, nil
	}

	// Update tasks to in_progress
	now := time.Now()
	updateQuery := `
		UPDATE quality_check_tasks
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
	return r.FindQCTasksWithDetails(taskIDs)
}

// FindQCTasksWithDetails finds quality check tasks with their associated comments and first review results
func (r *QualityCheckRepository) FindQCTasksWithDetails(taskIDs []int) ([]models.QualityCheckTask, error) {
	query := `
		SELECT 
			qct.id, qct.first_review_result_id, qct.comment_id, qct.reviewer_id, qct.status, 
			qct.claimed_at, qct.completed_at, qct.created_at,
			c.id, c.text,
			rr.id, rr.task_id, rr.reviewer_id, rr.is_approved, rr.tags, rr.reason, rr.created_at,
			u.username
		FROM quality_check_tasks qct
		INNER JOIN comment c ON qct.comment_id = c.id
		INNER JOIN review_results rr ON qct.first_review_result_id = rr.id
		LEFT JOIN users u ON rr.reviewer_id = u.id
		WHERE qct.id = ANY($1)
		ORDER BY qct.id
	`
	rows, err := r.db.Query(query, pq.Array(taskIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.QualityCheckTask{}
	for rows.Next() {
		var task models.QualityCheckTask
		var comment models.Comment
		var reviewResult models.ReviewResult
		var username sql.NullString

		err := rows.Scan(
			&task.ID, &task.FirstReviewResultID, &task.CommentID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&comment.ID, &comment.Text,
			&reviewResult.ID, &reviewResult.TaskID, &reviewResult.ReviewerID, &reviewResult.IsApproved,
			pq.Array(&reviewResult.Tags), &reviewResult.Reason, &reviewResult.CreatedAt,
			&username,
		)
		if err != nil {
			return nil, err
		}

		if username.Valid {
			reviewResult.Reviewer = &models.User{Username: username.String}
		}

		task.Comment = &comment
		task.FirstReviewResult = &reviewResult
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// GetMyQCTasks gets all in-progress quality check tasks for a reviewer
func (r *QualityCheckRepository) GetMyQCTasks(reviewerID int) ([]models.QualityCheckTask, error) {
	query := `
		SELECT 
			qct.id, qct.first_review_result_id, qct.comment_id, qct.reviewer_id, qct.status, 
			qct.claimed_at, qct.completed_at, qct.created_at,
			c.id, c.text,
			rr.id, rr.task_id, rr.reviewer_id, rr.is_approved, rr.tags, rr.reason, rr.created_at,
			u.username
		FROM quality_check_tasks qct
		INNER JOIN comment c ON qct.comment_id = c.id
		INNER JOIN review_results rr ON qct.first_review_result_id = rr.id
		LEFT JOIN users u ON rr.reviewer_id = u.id
		WHERE qct.reviewer_id = $1 AND qct.status = 'in_progress'
		ORDER BY qct.claimed_at DESC
	`
	rows, err := r.db.Query(query, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.QualityCheckTask{}
	for rows.Next() {
		var task models.QualityCheckTask
		var comment models.Comment
		var reviewResult models.ReviewResult
		var username sql.NullString

		err := rows.Scan(
			&task.ID, &task.FirstReviewResultID, &task.CommentID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&comment.ID, &comment.Text,
			&reviewResult.ID, &reviewResult.TaskID, &reviewResult.ReviewerID, &reviewResult.IsApproved,
			pq.Array(&reviewResult.Tags), &reviewResult.Reason, &reviewResult.CreatedAt,
			&username,
		)
		if err != nil {
			return nil, err
		}

		if username.Valid {
			reviewResult.Reviewer = &models.User{Username: username.String}
		}

		task.Comment = &comment
		task.FirstReviewResult = &reviewResult
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// CompleteQCTask marks a quality check task as completed
func (r *QualityCheckRepository) CompleteQCTask(taskID, reviewerID int) error {
	query := `
		UPDATE quality_check_tasks
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

// CreateQCResult creates a quality check result
func (r *QualityCheckRepository) CreateQCResult(result *models.QualityCheckResult) (bool, error) {
	query := `
		INSERT INTO quality_check_results (qc_task_id, reviewer_id, is_passed, error_type, qc_comment, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (qc_task_id) DO NOTHING
		RETURNING id, created_at
	`
	err := r.db.QueryRow(query, result.QCTaskID, result.ReviewerID, result.IsPassed,
		result.ErrorType, result.QCComment).Scan(&result.ID, &result.CreatedAt)
	if err == nil {
		return true, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}

	existingQuery := `
		SELECT id, reviewer_id, is_passed, error_type, qc_comment, created_at
		FROM quality_check_results
		WHERE qc_task_id = $1
	`
	err = r.db.QueryRow(existingQuery, result.QCTaskID).Scan(
		&result.ID,
		&result.ReviewerID,
		&result.IsPassed,
		&result.ErrorType,
		&result.QCComment,
		&result.CreatedAt,
	)
	if err != nil {
		return false, err
	}
	return false, nil
}

// ReturnQCTasks returns multiple quality check tasks back to pending status for a specific reviewer
func (r *QualityCheckRepository) ReturnQCTasks(taskIDs []int, reviewerID int) (int, error) {
	query := `
		UPDATE quality_check_tasks
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

// FindExpiredQCTasks finds quality check tasks that have been in progress for too long
func (r *QualityCheckRepository) FindExpiredQCTasks(timeoutMinutes int) ([]models.QualityCheckTask, error) {
	query := `
		SELECT id, comment_id, reviewer_id, claimed_at
		FROM quality_check_tasks
		WHERE status = 'in_progress' 
		AND claimed_at < NOW() - INTERVAL '1 minute' * $1
	`
	rows, err := r.db.Query(query, timeoutMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.QualityCheckTask{}
	for rows.Next() {
		var task models.QualityCheckTask
		if err := rows.Scan(&task.ID, &task.CommentID, &task.ReviewerID, &task.ClaimedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// ResetQCTask resets a quality check task back to pending status
func (r *QualityCheckRepository) ResetQCTask(taskID int) error {
	query := `
		UPDATE quality_check_tasks
		SET status = 'pending', reviewer_id = NULL, claimed_at = NULL
		WHERE id = $1
	`
	_, err := r.db.Exec(query, taskID)
	return err
}

// GetQCStats gets quality check statistics for a reviewer
func (r *QualityCheckRepository) GetQCStats(reviewerID int) (*models.QCStats, error) {
	// Get today's completed tasks
	todayQuery := `
		SELECT COUNT(*) 
		FROM quality_check_results qcr
		WHERE qcr.reviewer_id = $1 
		AND DATE(qcr.created_at) = CURRENT_DATE
	`
	var todayCompleted int
	err := r.db.QueryRow(todayQuery, reviewerID).Scan(&todayCompleted)
	if err != nil {
		return nil, err
	}

	// Get total completed tasks
	totalQuery := `
		SELECT COUNT(*) 
		FROM quality_check_results qcr
		WHERE qcr.reviewer_id = $1
	`
	var totalCompleted int
	err = r.db.QueryRow(totalQuery, reviewerID).Scan(&totalCompleted)
	if err != nil {
		return nil, err
	}

	// Get pass rate
	passRateQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN is_passed = true THEN 1 END) as passed
		FROM quality_check_results qcr
		WHERE qcr.reviewer_id = $1
	`
	var total, passed int
	err = r.db.QueryRow(passRateQuery, reviewerID).Scan(&total, &passed)
	if err != nil {
		return nil, err
	}

	var passRate float64
	if total > 0 {
		passRate = float64(passed) / float64(total) * 100
	}

	// Get task counts
	countQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress
		FROM quality_check_tasks qct
		WHERE qct.reviewer_id = $1 OR qct.reviewer_id IS NULL
	`
	var totalTasks, pendingTasks, inProgressTasks int
	err = r.db.QueryRow(countQuery, reviewerID).Scan(&totalTasks, &pendingTasks, &inProgressTasks)
	if err != nil {
		return nil, err
	}

	// Get error type statistics
	errorStatsQuery := `
		SELECT error_type, COUNT(*) as count
		FROM quality_check_results qcr
		WHERE qcr.reviewer_id = $1 AND is_passed = false AND error_type IS NOT NULL
		GROUP BY error_type
		ORDER BY count DESC
	`
	rows, err := r.db.Query(errorStatsQuery, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize as empty slice to avoid null in JSON response
	errorTypeStats := make([]models.QCErrorTypeStat, 0)
	for rows.Next() {
		var stat models.QCErrorTypeStat
		err := rows.Scan(&stat.ErrorType, &stat.Count)
		if err != nil {
			return nil, err
		}
		errorTypeStats = append(errorTypeStats, stat)
	}

	return &models.QCStats{
		TodayCompleted:  todayCompleted,
		TotalCompleted:  totalCompleted,
		PassRate:        passRate,
		TotalTasks:      totalTasks,
		PendingTasks:    pendingTasks,
		InProgressTasks: inProgressTasks,
		ErrorTypeStats:  errorTypeStats,
	}, nil
}

// GetUncheckedReviewResults gets review results that haven't been quality checked
func (r *QualityCheckRepository) GetUncheckedReviewResults(date string) ([]models.ReviewResult, error) {
	query := `
		SELECT rr.id, rr.task_id, rr.reviewer_id, rr.is_approved, rr.tags, rr.reason, rr.created_at
		FROM review_results rr
		INNER JOIN review_tasks rt ON rr.task_id = rt.id
		WHERE rr.quality_checked = false 
		AND rt.status = 'completed'
		AND DATE(rt.completed_at) = $1
		ORDER BY rr.created_at ASC
	`
	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize as empty slice to avoid null in JSON response
	results := make([]models.ReviewResult, 0)
	for rows.Next() {
		var result models.ReviewResult
		err := rows.Scan(
			&result.ID, &result.TaskID, &result.ReviewerID, &result.IsApproved,
			pq.Array(&result.Tags), &result.Reason, &result.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// UpdateReviewResultQCFlag updates the quality_checked flag for review results
func (r *QualityCheckRepository) UpdateReviewResultQCFlag(resultIDs []int) error {
	query := `
		UPDATE review_results
		SET quality_checked = true
		WHERE id = ANY($1)
	`
	_, err := r.db.Exec(query, pq.Array(resultIDs))
	return err
}

// CountByStatus counts quality check tasks by status
func (r *QualityCheckRepository) CountByStatus(status string) (int, error) {
	query := `SELECT COUNT(*) FROM quality_check_tasks WHERE status = $1`
	var count int
	err := r.db.QueryRow(query, status).Scan(&count)
	return count, err
}
