package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"errors"
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

type TaskQueueService struct {
	repo *repository.TaskQueueRepository
}

func NewTaskQueueService() *TaskQueueService {
	return &TaskQueueService{
		repo: repository.NewTaskQueueRepository(),
	}
}

// CreateTaskQueue creates a new task queue
func (s *TaskQueueService) CreateTaskQueue(req models.CreateTaskQueueRequest, adminID int) (*models.TaskQueue, error) {
	// Validate completed_tasks <= total_tasks
	if req.CompletedTasks > req.TotalTasks {
		return nil, errors.New("completed_tasks cannot be greater than total_tasks")
	}

	return s.repo.CreateTaskQueue(req, adminID)
}

// GetTaskQueueByID retrieves a task queue by ID
func (s *TaskQueueService) GetTaskQueueByID(id int) (*models.TaskQueue, error) {
	queue, err := s.repo.GetTaskQueueByID(id)
	if err != nil {
		return nil, errors.New("task queue not found")
	}
	return queue, nil
}

// ListTaskQueues retrieves task queues with pagination
func (s *TaskQueueService) ListTaskQueues(req models.ListTaskQueuesRequest) (*models.ListTaskQueuesResponse, error) {
	queues, total, err := s.repo.ListTaskQueues(req)
	if err != nil {
		return nil, err
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	totalPages := (total + req.PageSize - 1) / req.PageSize

	return &models.ListTaskQueuesResponse{
		Data:       queues,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateTaskQueue updates a task queue
func (s *TaskQueueService) UpdateTaskQueue(id int, req models.UpdateTaskQueueRequest, adminID int) (*models.TaskQueue, error) {
	// Validate completed_tasks <= total_tasks if both are provided
	if req.TotalTasks != nil && req.CompletedTasks != nil && *req.CompletedTasks > *req.TotalTasks {
		return nil, errors.New("completed_tasks cannot be greater than total_tasks")
	}

	return s.repo.UpdateTaskQueue(id, req, adminID)
}

// DeleteTaskQueue deletes a task queue
func (s *TaskQueueService) DeleteTaskQueue(id int) error {
	return s.repo.DeleteTaskQueue(id)
}

// GetAllTaskQueues retrieves all active task queues
func (s *TaskQueueService) GetAllTaskQueues() ([]models.TaskQueue, error) {
	return s.repo.GetAllTaskQueues()
}
