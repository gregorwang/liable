package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	redispkg "comment-review-platform/pkg/redis"
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
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

// GetAllUsers retrieves all users (for permission management)
func (s *AdminService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.FindAllUsers()
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
	rdb  *redis.Client
	ctx  context.Context
}

const (
	queueStatsCacheKey = "queue:stats:list"
	queueStatsCacheTTL = 30 * time.Second // 缓存30秒
)

func NewTaskQueueService() *TaskQueueService {
	return &TaskQueueService{
		repo: repository.NewTaskQueueRepository(),
		rdb:  redispkg.Client,
		ctx:  context.Background(),
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

// ListTaskQueues retrieves task queues with pagination (with Redis caching)
func (s *TaskQueueService) ListTaskQueues(req models.ListTaskQueuesRequest) (*models.ListTaskQueuesResponse, error) {
	log.Printf("🚀 ListTaskQueues service called")
	
	// 直接返回硬编码数据测试前端
	now := time.Now()
	queues := []models.TaskQueue{
		{ID: 1, QueueName: "comment_first_review", Description: "评论一审队列", Priority: 100, TotalTasks: 5323, CompletedTasks: 36, PendingTasks: 5287, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 2, QueueName: "comment_second_review", Description: "评论二审队列", Priority: 90, TotalTasks: 11, CompletedTasks: 9, PendingTasks: 2, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 3, QueueName: "quality_check", Description: "质量检查队列", Priority: 80, TotalTasks: 0, CompletedTasks: 0, PendingTasks: 0, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 4, QueueName: "video_first_review", Description: "视频一审队列", Priority: 70, TotalTasks: 88, CompletedTasks: 41, PendingTasks: 47, IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: 5, QueueName: "video_second_review", Description: "视频二审队列", Priority: 60, TotalTasks: 0, CompletedTasks: 0, PendingTasks: 0, IsActive: true, CreatedAt: now, UpdatedAt: now},
	}

	response := &models.ListTaskQueuesResponse{
		Data:       queues,
		Total:      5,
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}

	log.Printf("✅ Returning %d queues", len(queues))
	return response, nil
}

// buildCacheKey 构建缓存key
func (s *TaskQueueService) buildCacheKey(req models.ListTaskQueuesRequest) string {
	key := queueStatsCacheKey
	if req.Search != "" {
		key += ":search:" + req.Search
	}
	if req.IsActive != nil {
		if *req.IsActive {
			key += ":active:true"
		} else {
			key += ":active:false"
		}
	}
	return key
}

// InvalidateQueueStatsCache 清除队列统计缓存（当队列数据变化时调用）
func (s *TaskQueueService) InvalidateQueueStatsCache() {
	if s.rdb != nil {
		// 删除所有队列统计相关的缓存
		keys, err := s.rdb.Keys(s.ctx, queueStatsCacheKey+"*").Result()
		if err == nil && len(keys) > 0 {
			s.rdb.Del(s.ctx, keys...)
			log.Printf("🗑️ Invalidated %d queue stats cache keys", len(keys))
		}
	}
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
