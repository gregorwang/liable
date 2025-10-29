package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type VideoRepository struct {
	db *sql.DB
}

func NewVideoRepository() *VideoRepository {
	return &VideoRepository{db: database.DB}
}

// CreateVideo creates a new video record
func (r *VideoRepository) CreateVideo(video *models.TikTokVideo) error {
	query := `
		INSERT INTO tiktok_videos (video_key, filename, file_size, duration, upload_time, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(query, video.VideoKey, video.Filename, video.FileSize, video.Duration, video.UploadTime, video.Status).Scan(&video.ID, &video.CreatedAt, &video.UpdatedAt)
}

// GetVideoByID retrieves a video by ID
func (r *VideoRepository) GetVideoByID(id int) (*models.TikTokVideo, error) {
	query := `
		SELECT id, video_key, filename, file_size, duration, upload_time, video_url, url_expires_at, status, created_at, updated_at
		FROM tiktok_videos
		WHERE id = $1
	`
	var video models.TikTokVideo
	err := r.db.QueryRow(query, id).Scan(
		&video.ID, &video.VideoKey, &video.Filename, &video.FileSize, &video.Duration,
		&video.UploadTime, &video.VideoURL, &video.URLExpiresAt, &video.Status,
		&video.CreatedAt, &video.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("video not found")
		}
		return nil, err
	}
	return &video, nil
}

// GetVideoByKey retrieves a video by R2 key
func (r *VideoRepository) GetVideoByKey(videoKey string) (*models.TikTokVideo, error) {
	query := `
		SELECT id, video_key, filename, file_size, duration, upload_time, video_url, url_expires_at, status, created_at, updated_at
		FROM tiktok_videos
		WHERE video_key = $1
	`
	var video models.TikTokVideo
	err := r.db.QueryRow(query, videoKey).Scan(
		&video.ID, &video.VideoKey, &video.Filename, &video.FileSize, &video.Duration,
		&video.UploadTime, &video.VideoURL, &video.URLExpiresAt, &video.Status,
		&video.CreatedAt, &video.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("video not found")
		}
		return nil, err
	}
	return &video, nil
}

// ListVideos returns paginated videos with filters
func (r *VideoRepository) ListVideos(req models.ListVideosRequest) ([]models.TikTokVideo, int, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	if req.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, req.Status)
		argIndex++
	}

	if req.Search != "" {
		conditions = append(conditions, fmt.Sprintf("filename ILIKE $%d", argIndex))
		args = append(args, "%"+req.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tiktok_videos %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count videos: %w", err)
	}

	// Get paginated results
	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf(`
		SELECT id, video_key, filename, file_size, duration, upload_time, video_url, url_expires_at, status, created_at, updated_at
		FROM tiktok_videos
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.PageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query videos: %w", err)
	}
	defer rows.Close()

	videos := make([]models.TikTokVideo, 0)
	for rows.Next() {
		var video models.TikTokVideo
		err := rows.Scan(
			&video.ID, &video.VideoKey, &video.Filename, &video.FileSize, &video.Duration,
			&video.UploadTime, &video.VideoURL, &video.URLExpiresAt, &video.Status,
			&video.CreatedAt, &video.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan video: %w", err)
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating videos: %w", err)
	}

	return videos, total, nil
}

// UpdateVideoURL updates the pre-signed URL and expiration time
func (r *VideoRepository) UpdateVideoURL(id int, videoURL string, expiresAt time.Time) error {
	query := `
		UPDATE tiktok_videos
		SET video_url = $2, url_expires_at = $3, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id, videoURL, expiresAt)
	return err
}

// UpdateVideoStatus updates the video status
func (r *VideoRepository) UpdateVideoStatus(id int, status string) error {
	query := `
		UPDATE tiktok_videos
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id, status)
	return err
}

// GetVideoQualityTags retrieves quality tags by category
func (r *VideoRepository) GetVideoQualityTags(category string) ([]models.VideoQualityTag, error) {
	var query string
	var args []interface{}

	if category != "" {
		query = `
			SELECT id, name, description, category, is_active, created_at
			FROM video_quality_tags
			WHERE category = $1 AND is_active = true
			ORDER BY category, name
		`
		args = []interface{}{category}
	} else {
		query = `
			SELECT id, name, description, category, is_active, created_at
			FROM video_quality_tags
			WHERE is_active = true
			ORDER BY category, name
		`
		args = []interface{}{}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query video quality tags: %w", err)
	}
	defer rows.Close()

	tags := make([]models.VideoQualityTag, 0)
	for rows.Next() {
		var tag models.VideoQualityTag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Description, &tag.Category, &tag.IsActive, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video quality tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating video quality tags: %w", err)
	}

	return tags, nil
}

// CheckVideoExists checks if a video with the given key already exists
func (r *VideoRepository) CheckVideoExists(videoKey string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tiktok_videos WHERE video_key = $1)`
	var exists bool
	err := r.db.QueryRow(query, videoKey).Scan(&exists)
	return exists, err
}
