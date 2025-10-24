package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"errors"
)

type TagRepository struct {
	db *sql.DB
}

func NewTagRepository() *TagRepository {
	return &TagRepository{db: database.DB}
}

// Create creates a new tag
func (r *TagRepository) Create(tag *models.TagConfig) error {
	query := `
		INSERT INTO tag_config (name, description, is_active, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, tag.Name, tag.Description, tag.IsActive).
		Scan(&tag.ID, &tag.CreatedAt)
}

// FindAll returns all tags
func (r *TagRepository) FindAll() ([]models.TagConfig, error) {
	query := `
		SELECT id, name, description, is_active, created_at
		FROM tag_config
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []models.TagConfig{}
	for rows.Next() {
		var tag models.TagConfig
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Description, &tag.IsActive, &tag.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// FindActive returns all active tags
func (r *TagRepository) FindActive() ([]models.TagConfig, error) {
	query := `
		SELECT id, name, description, is_active, created_at
		FROM tag_config
		WHERE is_active = true
		ORDER BY name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []models.TagConfig{}
	for rows.Next() {
		var tag models.TagConfig
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Description, &tag.IsActive, &tag.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// FindByID finds a tag by ID
func (r *TagRepository) FindByID(id int) (*models.TagConfig, error) {
	query := `
		SELECT id, name, description, is_active, created_at
		FROM tag_config
		WHERE id = $1
	`
	tag := &models.TagConfig{}
	err := r.db.QueryRow(query, id).Scan(&tag.ID, &tag.Name, &tag.Description, &tag.IsActive, &tag.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("tag not found")
	}
	return tag, err
}

// Update updates a tag
func (r *TagRepository) Update(id int, name, description string, isActive *bool) error {
	query := `
		UPDATE tag_config
		SET name = COALESCE(NULLIF($1, ''), name),
			description = COALESCE(NULLIF($2, ''), description),
			is_active = COALESCE($3, is_active)
		WHERE id = $4
	`
	result, err := r.db.Exec(query, name, description, isActive, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("tag not found")
	}
	
	return nil
}

// Delete deletes a tag
func (r *TagRepository) Delete(id int) error {
	query := `DELETE FROM tag_config WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("tag not found")
	}
	
	return nil
}

