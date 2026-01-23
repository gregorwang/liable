package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	redispkg "comment-review-platform/pkg/redis"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
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

func (s *AdminService) CreateUser(req models.CreateUserRequest) (*models.User, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return nil, errors.New("用户名不能为空")
	}

	if existing, _ := s.userRepo.FindByUsername(username); existing != nil {
		return nil, errors.New("用户名已存在")
	}

	var emailPtr *string
	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email != "" {
			if existing, _ := s.userRepo.FindByEmail(email); existing != nil {
				return nil, errors.New("邮箱已被注册")
			}
			emailPtr = &email
		}
	}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "reviewer"
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "approved"
	}

	password := ""
	if req.Password != nil {
		password = strings.TrimSpace(*req.Password)
	}
	if password == "" {
		if emailPtr == nil {
			return nil, errors.New("未填写邮箱时必须设置密码")
		}
		generated, err := generateRandomPassword(16)
		if err != nil {
			return nil, err
		}
		password = generated
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:      username,
		Password:      string(hashedPassword),
		Email:         emailPtr,
		EmailVerified: emailPtr != nil,
		Role:          role,
		Status:        status,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AdminService) DeleteUser(userID int) error {
	return s.userRepo.DeleteByID(userID)
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

func generateRandomPassword(length int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if length <= 0 {
		length = 16
	}
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(bytes), nil
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
	cacheKey := s.buildCacheKey(req)
	if s.rdb != nil {
		if cached, err := s.rdb.Get(s.ctx, cacheKey).Result(); err == nil {
			var response models.ListTaskQueuesResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				return &response, nil
			}
		}
	}

	queues, total, err := s.repo.ListTaskQueues(req)
	if err != nil {
		return nil, err
	}

	totalPages := total / req.PageSize
	if total%req.PageSize > 0 {
		totalPages++
	}

	response := &models.ListTaskQueuesResponse{
		Data:       queues,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	if s.rdb != nil {
		if payload, err := json.Marshal(response); err == nil {
			s.rdb.Set(s.ctx, cacheKey, payload, queueStatsCacheTTL)
		}
	}

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
