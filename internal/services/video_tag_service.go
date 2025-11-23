package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"errors"
)

type VideoTagService struct {
	repo *repository.VideoTagRepository
}

func NewVideoTagService() *VideoTagService {
	return &VideoTagService{
		repo: repository.NewVideoTagRepository(),
	}
}

// GetAllTags retrieves all video quality tags
func (s *VideoTagService) GetAllTags() ([]models.VideoQualityTag, error) {
	return s.repo.GetAll()
}

// GetTagsByScope retrieves tags by scope
func (s *VideoTagService) GetTagsByScope(scope string) ([]models.VideoQualityTag, error) {
	return s.repo.GetByScope(scope)
}

// GetTagsByScopeAndQueue retrieves tags by scope and queue ID
func (s *VideoTagService) GetTagsByScopeAndQueue(scope, queueID string) ([]models.VideoQualityTag, error) {
	return s.repo.GetByScopeAndQueueID(scope, queueID)
}

// GetTagByID retrieves a tag by ID
func (s *VideoTagService) GetTagByID(id int) (*models.VideoQualityTag, error) {
	return s.repo.GetByID(id)
}

// CreateTag creates a new video quality tag
func (s *VideoTagService) CreateTag(tag *models.VideoQualityTag) error {
	// Validate required fields
	if tag.Name == "" {
		return errors.New("tag name is required")
	}
	if tag.Category == "" {
		return errors.New("tag category is required")
	}

	// Set defaults
	tag.IsActive = true

	return s.repo.Create(tag)
}

// UpdateTag updates an existing video quality tag
func (s *VideoTagService) UpdateTag(id int, updates map[string]interface{}) error {
	tag, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("tag not found")
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		tag.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		tag.Description = desc
	}
	if category, ok := updates["category"].(string); ok {
		tag.Category = category
	}
	if active, ok := updates["is_active"].(bool); ok {
		tag.IsActive = active
	}

	return s.repo.Update(tag)
}

// DeleteTag deletes a video quality tag
func (s *VideoTagService) DeleteTag(id int) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("tag not found")
	}
	return s.repo.Delete(id)
}

// ToggleTagActive toggles the active status of a tag
func (s *VideoTagService) ToggleTagActive(id int) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("tag not found")
	}
	return s.repo.ToggleActive(id)
}
