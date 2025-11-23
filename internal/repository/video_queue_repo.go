package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type VideoQueueRepository struct {
	db *sql.DB
}

func NewVideoQueueRepository() *VideoQueueRepository {
	return &VideoQueueRepository{
		db: database.DB,
	}
}

// CreateQueueTask creates a new video queue task for a specific pool
func (r *VideoQueueRepository) CreateQueueTask(videoID int, pool string) error {
	query := `
		INSERT INTO video_queue_tasks (video_id, pool, status)
		VALUES ($1, $2, 'pending')
		ON CONFLICT (video_id, pool) DO NOTHING
	`
	_, err := r.db.Exec(query, videoID, pool)
	return err
}

// ClaimQueueTasks claims pending tasks from a specific pool for a reviewer
func (r *VideoQueueRepository) ClaimQueueTasks(pool string, reviewerID int, count int) ([]models.VideoQueueTask, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Lock and claim tasks
	query := `
		UPDATE video_queue_tasks
		SET status = 'in_progress',
		    reviewer_id = $1,
		    claimed_at = NOW()
		WHERE id IN (
			SELECT id FROM video_queue_tasks
			WHERE pool = $2 AND status = 'pending'
			ORDER BY created_at ASC
			LIMIT $3
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, video_id, pool, reviewer_id, status, claimed_at, completed_at, created_at
	`

	rows, err := tx.Query(query, reviewerID, pool, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.VideoQueueTask
	for rows.Next() {
		var task models.VideoQueueTask
		err := rows.Scan(
			&task.ID,
			&task.VideoID,
			&task.Pool,
			&task.ReviewerID,
			&task.Status,
			&task.ClaimedAt,
			&task.CompletedAt,
			&task.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Load video data for each task
	for i := range tasks {
		video, err := r.getVideoByID(tasks[i].VideoID)
		if err == nil {
			tasks[i].Video = video
		}
	}

	return tasks, nil
}

// GetMyQueueTasks retrieves in-progress tasks for a reviewer in a specific pool
func (r *VideoQueueRepository) GetMyQueueTasks(pool string, reviewerID int) ([]models.VideoQueueTask, error) {
	query := `
		SELECT id, video_id, pool, reviewer_id, status, claimed_at, completed_at, created_at
		FROM video_queue_tasks
		WHERE pool = $1 AND reviewer_id = $2 AND status = 'in_progress'
		ORDER BY claimed_at ASC
	`

	rows, err := r.db.Query(query, pool, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.VideoQueueTask
	for rows.Next() {
		var task models.VideoQueueTask
		err := rows.Scan(
			&task.ID,
			&task.VideoID,
			&task.Pool,
			&task.ReviewerID,
			&task.Status,
			&task.ClaimedAt,
			&task.CompletedAt,
			&task.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Load video data
		video, err := r.getVideoByID(task.VideoID)
		if err == nil {
			task.Video = video
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// CompleteQueueTask marks a task as completed
func (r *VideoQueueRepository) CompleteQueueTask(taskID int, reviewerID int) error {
	query := `
		UPDATE video_queue_tasks
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
		return errors.New("task not found or already completed")
	}

	return nil
}

// CreateQueueResult creates a review result for a queue task
func (r *VideoQueueRepository) CreateQueueResult(result *models.VideoQueueResult) error {
	// Validate tags (max 3)
	if len(result.Tags) > 3 {
		return errors.New("maximum 3 tags allowed")
	}

	query := `
		INSERT INTO video_queue_results (task_id, reviewer_id, review_decision, reason, tags, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		result.TaskID,
		result.ReviewerID,
		result.ReviewDecision,
		result.Reason,
		pq.Array(result.Tags),
	).Scan(&result.ID, &result.CreatedAt)

	return err
}

// ReturnQueueTasks returns tasks back to pending status
func (r *VideoQueueRepository) ReturnQueueTasks(taskIDs []int, reviewerID int) (int, error) {
	query := `
		UPDATE video_queue_tasks
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

// ResetQueueTask resets an expired task back to pending
func (r *VideoQueueRepository) ResetQueueTask(taskID int) error {
	query := `
		UPDATE video_queue_tasks
		SET status = 'pending', reviewer_id = NULL, claimed_at = NULL
		WHERE id = $1 AND status = 'in_progress'
	`

	_, err := r.db.Exec(query, taskID)
	return err
}

// FindExpiredQueueTasks finds tasks that have exceeded the timeout
func (r *VideoQueueRepository) FindExpiredQueueTasks(pool string, timeoutMinutes int) ([]models.VideoQueueTask, error) {
	query := `
		SELECT id, video_id, pool, reviewer_id, status, claimed_at, completed_at, created_at
		FROM video_queue_tasks
		WHERE pool = $1
		  AND status = 'in_progress'
		  AND claimed_at < NOW() - INTERVAL '1 minute' * $2
	`

	rows, err := r.db.Query(query, pool, timeoutMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.VideoQueueTask
	for rows.Next() {
		var task models.VideoQueueTask
		err := rows.Scan(
			&task.ID,
			&task.VideoID,
			&task.Pool,
			&task.ReviewerID,
			&task.Status,
			&task.ClaimedAt,
			&task.CompletedAt,
			&task.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTaskByID retrieves a task by ID
func (r *VideoQueueRepository) GetTaskByID(taskID int) (*models.VideoQueueTask, error) {
	query := `
		SELECT id, video_id, pool, reviewer_id, status, claimed_at, completed_at, created_at
		FROM video_queue_tasks
		WHERE id = $1
	`

	var task models.VideoQueueTask
	err := r.db.QueryRow(query, taskID).Scan(
		&task.ID,
		&task.VideoID,
		&task.Pool,
		&task.ReviewerID,
		&task.Status,
		&task.ClaimedAt,
		&task.CompletedAt,
		&task.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

// GetVideoQueueTags retrieves tags for a specific pool
func (r *VideoQueueRepository) GetVideoQueueTags(pool string) ([]models.VideoQueueTag, error) {
	query := `
		SELECT id, name, description, category, scope, queue_id, is_active, created_at
		FROM video_quality_tags
		WHERE is_active = TRUE
		  AND scope = 'video'
		  AND (queue_id = $1 OR queue_id IS NULL)
		ORDER BY category, name
	`

	rows, err := r.db.Query(query, pool)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.VideoQueueTag
	for rows.Next() {
		var tag models.VideoQueueTag
		err := rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.Description,
			&tag.Category,
			&tag.Scope,
			&tag.QueueID,
			&tag.IsActive,
			&tag.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// GetQueuePoolStats retrieves statistics for a specific pool
func (r *VideoQueueRepository) GetQueuePoolStats(pool string) (*models.VideoQueuePoolStats, error) {
	query := `
		SELECT
			pool,
			COUNT(*) as total_tasks,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_tasks,
			COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress_tasks,
			COALESCE(AVG(CASE
				WHEN status = 'completed' AND completed_at IS NOT NULL AND claimed_at IS NOT NULL
				THEN EXTRACT(EPOCH FROM (completed_at - claimed_at))/60
			END), 0) as avg_process_time_minutes
		FROM video_queue_tasks
		WHERE pool = $1
		GROUP BY pool
	`

	var stats models.VideoQueuePoolStats
	err := r.db.QueryRow(query, pool).Scan(
		&stats.Pool,
		&stats.TotalTasks,
		&stats.CompletedTasks,
		&stats.PendingTasks,
		&stats.InProgressTasks,
		&stats.AvgProcessTimeMinutes,
	)

	if err == sql.ErrNoRows {
		// Return empty stats if no tasks exist for this pool
		return &models.VideoQueuePoolStats{
			Pool:                  pool,
			TotalTasks:            0,
			CompletedTasks:        0,
			PendingTasks:          0,
			InProgressTasks:       0,
			AvgProcessTimeMinutes: 0,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// Helper function to get video by ID
func (r *VideoQueueRepository) getVideoByID(videoID int) (*models.TikTokVideo, error) {
	query := `
		SELECT id, video_key, filename, file_size, duration, upload_time,
		       video_url, url_expires_at, status, created_at, updated_at
		FROM tiktok_videos
		WHERE id = $1
	`

	var video models.TikTokVideo
	err := r.db.QueryRow(query, videoID).Scan(
		&video.ID,
		&video.VideoKey,
		&video.Filename,
		&video.FileSize,
		&video.Duration,
		&video.UploadTime,
		&video.VideoURL,
		&video.URLExpiresAt,
		&video.Status,
		&video.CreatedAt,
		&video.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &video, nil
}

// UpdateVideoStatus updates the status of a video
func (r *VideoQueueRepository) UpdateVideoStatus(videoID int, status string) error {
	query := `
		UPDATE tiktok_videos
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(query, status, videoID)
	return err
}

// GetPendingTaskCount returns the number of pending tasks in a pool
func (r *VideoQueueRepository) GetPendingTaskCount(pool string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM video_queue_tasks
		WHERE pool = $1 AND status = 'pending'
	`

	var count int
	err := r.db.QueryRow(query, pool).Scan(&count)
	return count, err
}
