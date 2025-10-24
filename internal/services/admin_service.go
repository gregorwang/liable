package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
)

type AdminService struct {
	userRepo *repository.UserRepository
	tagRepo  *repository.TagRepository
}

func NewAdminService() *AdminService {
	return &AdminService{
		userRepo: repository.NewUserRepository(),
		tagRepo:  repository.NewTagRepository(),
	}
}

// GetPendingUsers retrieves all users pending approval
func (s *AdminService) GetPendingUsers() ([]models.User, error) {
	return s.userRepo.FindPendingUsers()
}

// ApproveUser approves or rejects a user
func (s *AdminService) ApproveUser(userID int, status string) error {
	return s.userRepo.UpdateStatus(userID, status)
}

// GetAllTags retrieves all tags
func (s *AdminService) GetAllTags() ([]models.TagConfig, error) {
	return s.tagRepo.FindAll()
}

// CreateTag creates a new tag
func (s *AdminService) CreateTag(name, description string) (*models.TagConfig, error) {
	tag := &models.TagConfig{
		Name:        name,
		Description: description,
		IsActive:    true,
	}
	
	if err := s.tagRepo.Create(tag); err != nil {
		return nil, err
	}
	
	return tag, nil
}

// UpdateTag updates a tag
func (s *AdminService) UpdateTag(id int, name, description string, isActive *bool) error {
	return s.tagRepo.Update(id, name, description, isActive)
}

// DeleteTag deletes a tag
func (s *AdminService) DeleteTag(id int) error {
	return s.tagRepo.Delete(id)
}

