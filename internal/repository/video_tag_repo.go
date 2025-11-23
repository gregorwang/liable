package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"fmt"
)

type VideoTagRepository struct {
	db *sql.DB
}

func NewVideoTagRepository() *VideoTagRepository {
	return &VideoTagRepository{db: database.DB}
}

// GetAll retrieves all video quality tags
func (r *VideoTagRepository) GetAll() ([]models.VideoQualityTag, error) {
	query := `
		SELECT id, name, description, category, is_active, created_at
		FROM video_quality_tags
		ORDER BY category, name
	`
	rows, err := r.db.Query(query)
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

// GetByScope retrieves tags by scope (kept for backward compatibility, but returns all tags)
func (r *VideoTagRepository) GetByScope(scope string) ([]models.VideoQualityTag, error) {
	return r.GetAll()
}

// GetByScopeAndQueueID retrieves tags (kept for backward compatibility, but returns active tags)
func (r *VideoTagRepository) GetByScopeAndQueueID(scope, queueID string) ([]models.VideoQualityTag, error) {
	query := `
		SELECT id, name, description, category, is_active, created_at
		FROM video_quality_tags
		WHERE is_active = true
		ORDER BY category, name
	`
	rows, err := r.db.Query(query)
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

// GetByID retrieves a tag by ID
func (r *VideoTagRepository) GetByID(id int) (*models.VideoQualityTag, error) {
	query := `
		SELECT id, name, description, category, is_active, created_at
		FROM video_quality_tags
		WHERE id = $1
	`
	var tag models.VideoQualityTag
	err := r.db.QueryRow(query, id).Scan(&tag.ID, &tag.Name, &tag.Description, &tag.Category, &tag.IsActive, &tag.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("video quality tag not found")
		}
		return nil, fmt.Errorf("failed to get video quality tag: %w", err)
	}
	return &tag, nil
}

// Create creates a new video quality tag
func (r *VideoTagRepository) Create(tag *models.VideoQualityTag) error {
	query := `
		INSERT INTO video_quality_tags (name, description, category, is_active, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, tag.Name, tag.Description, tag.Category, tag.IsActive).
		Scan(&tag.ID, &tag.CreatedAt)
}

// Update updates an existing video quality tag
func (r *VideoTagRepository) Update(tag *models.VideoQualityTag) error {
	query := `
		UPDATE video_quality_tags
		SET name = $2, description = $3, category = $4, is_active = $5
		WHERE id = $1
	`
	result, err := r.db.Exec(query, tag.ID, tag.Name, tag.Description, tag.Category, tag.IsActive)
	if err != nil {
		return fmt.Errorf("failed to update video quality tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("video quality tag not found")
	}

	return nil
}

// Delete deletes a video quality tag
func (r *VideoTagRepository) Delete(id int) error {
	query := `DELETE FROM video_quality_tags WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete video quality tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("video quality tag not found")
	}

	return nil
}

// ToggleActive toggles the active status of a tag
func (r *VideoTagRepository) ToggleActive(id int) error {
	query := `
		UPDATE video_quality_tags
		SET is_active = NOT is_active
		WHERE id = $1
	`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to toggle video quality tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("video quality tag not found")
	}

	return nil
}
