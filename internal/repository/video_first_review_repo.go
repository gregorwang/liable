package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type VideoFirstReviewRepository struct {
	db *sql.DB
}

func NewVideoFirstReviewRepository() *VideoFirstReviewRepository {
	return &VideoFirstReviewRepository{db: database.DB}
}

// CreateFirstReviewTask creates a new first review task
func (r *VideoFirstReviewRepository) CreateFirstReviewTask(videoID int) error {
	query := `
		INSERT INTO video_first_review_tasks (video_id, status, created_at)
		VALUES ($1, 'pending', NOW())
	`
	_, err := r.db.Exec(query, videoID)
	return err
}

// ClaimFirstReviewTasks claims pending first review tasks for a reviewer
func (r *VideoFirstReviewRepository) ClaimFirstReviewTasks(reviewerID int, limit int) ([]models.VideoFirstReviewTask, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Select pending tasks
	query := `
		SELECT id, video_id, created_at
		FROM video_first_review_tasks
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
	tasks := []models.VideoFirstReviewTask{}

	for rows.Next() {
		var task models.VideoFirstReviewTask
		if err := rows.Scan(&task.ID, &task.VideoID, &task.CreatedAt); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, task.ID)
		tasks = append(tasks, task)
	}

	if len(taskIDs) == 0 {
		return []models.VideoFirstReviewTask{}, nil
	}

	// Update tasks to in_progress
	now := time.Now()
	updateQuery := `
		UPDATE video_first_review_tasks
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

	// Fetch full task details with videos
	return r.FindFirstReviewTasksWithVideos(taskIDs)
}

// FindFirstReviewTasksWithVideos finds tasks with their associated videos
func (r *VideoFirstReviewRepository) FindFirstReviewTasksWithVideos(taskIDs []int) ([]models.VideoFirstReviewTask, error) {
	query := `
		SELECT 
			vfrt.id, vfrt.video_id, vfrt.reviewer_id, vfrt.status, 
			vfrt.claimed_at, vfrt.completed_at, vfrt.created_at,
			tv.id, tv.video_key, tv.filename, tv.file_size, tv.duration, tv.upload_time, tv.video_url, tv.url_expires_at, tv.status, tv.created_at, tv.updated_at
		FROM video_first_review_tasks vfrt
		INNER JOIN tiktok_videos tv ON vfrt.video_id = tv.id
		WHERE vfrt.id = ANY($1)
		ORDER BY vfrt.id
	`
	rows, err := r.db.Query(query, pq.Array(taskIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.VideoFirstReviewTask{}
	for rows.Next() {
		var task models.VideoFirstReviewTask
		var video models.TikTokVideo
		err := rows.Scan(
			&task.ID, &task.VideoID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&video.ID, &video.VideoKey, &video.Filename, &video.FileSize, &video.Duration, &video.UploadTime, &video.VideoURL, &video.URLExpiresAt, &video.Status, &video.CreatedAt, &video.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		task.Video = &video
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// GetMyFirstReviewTasks gets all in-progress first review tasks for a reviewer
func (r *VideoFirstReviewRepository) GetMyFirstReviewTasks(reviewerID int) ([]models.VideoFirstReviewTask, error) {
	query := `
		SELECT 
			vfrt.id, vfrt.video_id, vfrt.reviewer_id, vfrt.status, 
			vfrt.claimed_at, vfrt.completed_at, vfrt.created_at,
			tv.id, tv.video_key, tv.filename, tv.file_size, tv.duration, tv.upload_time, tv.video_url, tv.url_expires_at, tv.status, tv.created_at, tv.updated_at
		FROM video_first_review_tasks vfrt
		INNER JOIN tiktok_videos tv ON vfrt.video_id = tv.id
		WHERE vfrt.reviewer_id = $1 AND vfrt.status = 'in_progress'
		ORDER BY vfrt.claimed_at DESC
	`
	rows, err := r.db.Query(query, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.VideoFirstReviewTask{}
	for rows.Next() {
		var task models.VideoFirstReviewTask
		var video models.TikTokVideo
		err := rows.Scan(
			&task.ID, &task.VideoID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&video.ID, &video.VideoKey, &video.Filename, &video.FileSize, &video.Duration, &video.UploadTime, &video.VideoURL, &video.URLExpiresAt, &video.Status, &video.CreatedAt, &video.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		task.Video = &video
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// CompleteFirstReviewTask marks a first review task as completed
func (r *VideoFirstReviewRepository) CompleteFirstReviewTask(taskID, reviewerID int) error {
	query := `
		UPDATE video_first_review_tasks
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

// CreateFirstReviewResult creates a first review result
func (r *VideoFirstReviewRepository) CreateFirstReviewResult(result *models.VideoFirstReviewResult) error {
	// Convert QualityDimensions to JSON
	qualityDimensionsJSON, err := json.Marshal(result.QualityDimensions)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO video_first_review_results (task_id, reviewer_id, is_approved, quality_dimensions, overall_score, traffic_pool_result, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, result.TaskID, result.ReviewerID, result.IsApproved,
		qualityDimensionsJSON, result.OverallScore, result.TrafficPoolResult, result.Reason).Scan(&result.ID, &result.CreatedAt)
}

// ReturnFirstReviewTasks returns multiple first review tasks back to pending status for a specific reviewer
func (r *VideoFirstReviewRepository) ReturnFirstReviewTasks(taskIDs []int, reviewerID int) (int, error) {
	query := `
		UPDATE video_first_review_tasks
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

// FindExpiredFirstReviewTasks finds first review tasks that have been in progress for too long
func (r *VideoFirstReviewRepository) FindExpiredFirstReviewTasks(timeoutMinutes int) ([]models.VideoFirstReviewTask, error) {
	query := `
		SELECT id, video_id, reviewer_id, claimed_at
		FROM video_first_review_tasks
		WHERE status = 'in_progress' 
		AND claimed_at < NOW() - INTERVAL '1 minute' * $1
	`
	rows, err := r.db.Query(query, timeoutMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.VideoFirstReviewTask{}
	for rows.Next() {
		var task models.VideoFirstReviewTask
		if err := rows.Scan(&task.ID, &task.VideoID, &task.ReviewerID, &task.ClaimedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// ResetFirstReviewTask resets a first review task back to pending status
func (r *VideoFirstReviewRepository) ResetFirstReviewTask(taskID int) error {
	query := `
		UPDATE video_first_review_tasks
		SET status = 'pending', reviewer_id = NULL, claimed_at = NULL
		WHERE id = $1
	`
	_, err := r.db.Exec(query, taskID)
	return err
}
