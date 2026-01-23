package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type VideoSecondReviewRepository struct {
	db *sql.DB
}

func NewVideoSecondReviewRepository() *VideoSecondReviewRepository {
	return &VideoSecondReviewRepository{db: database.DB}
}

// CreateSecondReviewTask creates a new second review task
func (r *VideoSecondReviewRepository) CreateSecondReviewTask(firstReviewResultID, videoID int) (bool, error) {
	query := `
		INSERT INTO video_second_review_tasks (first_review_result_id, video_id, status, created_at)
		VALUES ($1, $2, 'pending', NOW())
		ON CONFLICT (first_review_result_id) DO NOTHING
	`
	result, err := r.db.Exec(query, firstReviewResultID, videoID)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// ClaimSecondReviewTasks claims pending second review tasks for a reviewer
func (r *VideoSecondReviewRepository) ClaimSecondReviewTasks(reviewerID int, limit int) ([]models.VideoSecondReviewTask, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Select pending tasks
	query := `
		SELECT id, first_review_result_id, video_id, created_at
		FROM video_second_review_tasks
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
	tasks := []models.VideoSecondReviewTask{}

	for rows.Next() {
		var task models.VideoSecondReviewTask
		if err := rows.Scan(&task.ID, &task.FirstReviewResultID, &task.VideoID, &task.CreatedAt); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, task.ID)
		tasks = append(tasks, task)
	}

	if len(taskIDs) == 0 {
		return []models.VideoSecondReviewTask{}, nil
	}

	// Update tasks to in_progress
	now := time.Now()
	updateQuery := `
		UPDATE video_second_review_tasks
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

	// Fetch full task details with videos and first review results
	return r.FindSecondReviewTasksWithDetails(taskIDs)
}

// FindSecondReviewTasksWithDetails finds tasks with their associated videos and first review results
func (r *VideoSecondReviewRepository) FindSecondReviewTasksWithDetails(taskIDs []int) ([]models.VideoSecondReviewTask, error) {
	query := `
		SELECT 
			vsrt.id, vsrt.first_review_result_id, vsrt.video_id, vsrt.reviewer_id, vsrt.status, 
			vsrt.claimed_at, vsrt.completed_at, vsrt.created_at,
			tv.id, tv.video_key, tv.filename, tv.file_size, tv.duration, tv.upload_time, tv.video_url, tv.url_expires_at, tv.status, tv.created_at, tv.updated_at,
			vfrr.id, vfrr.task_id, vfrr.reviewer_id, vfrr.is_approved, vfrr.quality_dimensions, vfrr.overall_score, vfrr.traffic_pool_result, vfrr.reason, vfrr.created_at
		FROM video_second_review_tasks vsrt
		INNER JOIN tiktok_videos tv ON vsrt.video_id = tv.id
		INNER JOIN video_first_review_results vfrr ON vsrt.first_review_result_id = vfrr.id
		WHERE vsrt.id = ANY($1)
		ORDER BY vsrt.id
	`
	rows, err := r.db.Query(query, pq.Array(taskIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.VideoSecondReviewTask{}
	for rows.Next() {
		var task models.VideoSecondReviewTask
		var video models.TikTokVideo
		var firstResult models.VideoFirstReviewResult
		var qualityDimensionsJSON []byte

		err := rows.Scan(
			&task.ID, &task.FirstReviewResultID, &task.VideoID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&video.ID, &video.VideoKey, &video.Filename, &video.FileSize, &video.Duration, &video.UploadTime, &video.VideoURL, &video.URLExpiresAt, &video.Status, &video.CreatedAt, &video.UpdatedAt,
			&firstResult.ID, &firstResult.TaskID, &firstResult.ReviewerID, &firstResult.IsApproved, &qualityDimensionsJSON, &firstResult.OverallScore, &firstResult.TrafficPoolResult, &firstResult.Reason, &firstResult.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse quality dimensions JSON
		if err := json.Unmarshal(qualityDimensionsJSON, &firstResult.QualityDimensions); err != nil {
			return nil, err
		}

		task.Video = &video
		task.FirstReviewResult = &firstResult
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// GetMySecondReviewTasks gets all in-progress second review tasks for a reviewer
func (r *VideoSecondReviewRepository) GetMySecondReviewTasks(reviewerID int) ([]models.VideoSecondReviewTask, error) {
	query := `
		SELECT 
			vsrt.id, vsrt.first_review_result_id, vsrt.video_id, vsrt.reviewer_id, vsrt.status, 
			vsrt.claimed_at, vsrt.completed_at, vsrt.created_at,
			tv.id, tv.video_key, tv.filename, tv.file_size, tv.duration, tv.upload_time, tv.video_url, tv.url_expires_at, tv.status, tv.created_at, tv.updated_at,
			vfrr.id, vfrr.task_id, vfrr.reviewer_id, vfrr.is_approved, vfrr.quality_dimensions, vfrr.overall_score, vfrr.traffic_pool_result, vfrr.reason, vfrr.created_at
		FROM video_second_review_tasks vsrt
		INNER JOIN tiktok_videos tv ON vsrt.video_id = tv.id
		INNER JOIN video_first_review_results vfrr ON vsrt.first_review_result_id = vfrr.id
		WHERE vsrt.reviewer_id = $1 AND vsrt.status = 'in_progress'
		ORDER BY vsrt.claimed_at DESC
	`
	rows, err := r.db.Query(query, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.VideoSecondReviewTask{}
	for rows.Next() {
		var task models.VideoSecondReviewTask
		var video models.TikTokVideo
		var firstResult models.VideoFirstReviewResult
		var qualityDimensionsJSON []byte

		err := rows.Scan(
			&task.ID, &task.FirstReviewResultID, &task.VideoID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&video.ID, &video.VideoKey, &video.Filename, &video.FileSize, &video.Duration, &video.UploadTime, &video.VideoURL, &video.URLExpiresAt, &video.Status, &video.CreatedAt, &video.UpdatedAt,
			&firstResult.ID, &firstResult.TaskID, &firstResult.ReviewerID, &firstResult.IsApproved, &qualityDimensionsJSON, &firstResult.OverallScore, &firstResult.TrafficPoolResult, &firstResult.Reason, &firstResult.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse quality dimensions JSON
		if err := json.Unmarshal(qualityDimensionsJSON, &firstResult.QualityDimensions); err != nil {
			return nil, err
		}

		task.Video = &video
		task.FirstReviewResult = &firstResult
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// CompleteSecondReviewTask marks a second review task as completed
func (r *VideoSecondReviewRepository) CompleteSecondReviewTask(taskID, reviewerID int) error {
	query := `
		UPDATE video_second_review_tasks
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
func (r *VideoSecondReviewRepository) CreateSecondReviewResult(result *models.VideoSecondReviewResult) (bool, error) {
	// Convert QualityDimensions to JSON
	qualityDimensionsJSON, err := json.Marshal(result.QualityDimensions)
	if err != nil {
		return false, err
	}

	query := `
		INSERT INTO video_second_review_results (second_task_id, reviewer_id, is_approved, quality_dimensions, overall_score, traffic_pool_result, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (second_task_id) DO NOTHING
		RETURNING id, created_at
	`
	err = r.db.QueryRow(query, result.SecondTaskID, result.ReviewerID, result.IsApproved,
		qualityDimensionsJSON, result.OverallScore, result.TrafficPoolResult, result.Reason).Scan(&result.ID, &result.CreatedAt)
	if err == nil {
		return true, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}

	existingQuery := `
		SELECT id, reviewer_id, is_approved, quality_dimensions, overall_score, traffic_pool_result, reason, created_at
		FROM video_second_review_results
		WHERE second_task_id = $1
	`
	var qualityJSON []byte
	var trafficPool sql.NullString
	var reason sql.NullString
	err = r.db.QueryRow(existingQuery, result.SecondTaskID).Scan(
		&result.ID,
		&result.ReviewerID,
		&result.IsApproved,
		&qualityJSON,
		&result.OverallScore,
		&trafficPool,
		&reason,
		&result.CreatedAt,
	)
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(qualityJSON, &result.QualityDimensions); err != nil {
		return false, err
	}
	if trafficPool.Valid {
		result.TrafficPoolResult = &trafficPool.String
	} else {
		result.TrafficPoolResult = nil
	}
	if reason.Valid {
		result.Reason = &reason.String
	} else {
		result.Reason = nil
	}
	return false, nil
}

// ReturnSecondReviewTasks returns multiple second review tasks back to pending status for a specific reviewer
func (r *VideoSecondReviewRepository) ReturnSecondReviewTasks(taskIDs []int, reviewerID int) (int, error) {
	query := `
		UPDATE video_second_review_tasks
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
func (r *VideoSecondReviewRepository) FindExpiredSecondReviewTasks(timeoutMinutes int) ([]models.VideoSecondReviewTask, error) {
	query := `
		SELECT id, first_review_result_id, video_id, reviewer_id, claimed_at
		FROM video_second_review_tasks
		WHERE status = 'in_progress' 
		AND claimed_at < NOW() - INTERVAL '1 minute' * $1
	`
	rows, err := r.db.Query(query, timeoutMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.VideoSecondReviewTask{}
	for rows.Next() {
		var task models.VideoSecondReviewTask
		if err := rows.Scan(&task.ID, &task.FirstReviewResultID, &task.VideoID, &task.ReviewerID, &task.ClaimedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// ResetSecondReviewTask resets a second review task back to pending status
func (r *VideoSecondReviewRepository) ResetSecondReviewTask(taskID int) error {
	query := `
		UPDATE video_second_review_tasks
		SET status = 'pending', reviewer_id = NULL, claimed_at = NULL
		WHERE id = $1
	`
	_, err := r.db.Exec(query, taskID)
	return err
}
