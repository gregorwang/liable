package services

import (
	"comment-review-platform/internal/config"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	redispkg "comment-review-platform/pkg/redis"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type TaskService struct {
	taskRepo *repository.TaskRepository
	tagRepo  *repository.TagRepository
	rdb      *redis.Client
	ctx      context.Context
}

func NewTaskService() *TaskService {
	return &TaskService{
		taskRepo: repository.NewTaskRepository(),
		tagRepo:  repository.NewTagRepository(),
		rdb:      redispkg.Client,
		ctx:      context.Background(),
	}
}

// ClaimTasks allows a reviewer to claim tasks
func (s *TaskService) ClaimTasks(reviewerID int) ([]models.ReviewTask, error) {
	claimSize := config.AppConfig.TaskClaimSize

	// Check if user already has uncompleted tasks
	existingTasks, err := s.taskRepo.GetMyTasks(reviewerID)
	if err != nil {
		return nil, err
	}

	if len(existingTasks) > 0 {
		return nil, errors.New("please complete your current tasks before claiming new ones")
	}

	// Claim tasks from database
	tasks, err := s.taskRepo.ClaimTasks(reviewerID, claimSize)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return []models.ReviewTask{}, nil
	}

	// Add tasks to Redis for tracking
	userClaimedKey := fmt.Sprintf("task:claimed:%d", reviewerID)
	timeout := time.Duration(config.AppConfig.TaskTimeoutMinutes) * time.Minute

	pipe := s.rdb.Pipeline()
	for _, task := range tasks {
		// Add to user's claimed set
		pipe.SAdd(s.ctx, userClaimedKey, task.ID)

		// Set lock for each task
		lockKey := fmt.Sprintf("task:lock:%d", task.ID)
		pipe.Set(s.ctx, lockKey, reviewerID, timeout)
	}
	pipe.Expire(s.ctx, userClaimedKey, timeout)

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when claiming tasks: %v", err)
	}

	return tasks, nil
}

// GetMyTasks retrieves the current user's in-progress tasks
func (s *TaskService) GetMyTasks(reviewerID int) ([]models.ReviewTask, error) {
	return s.taskRepo.GetMyTasks(reviewerID)
}

// SubmitReview submits a review result
func (s *TaskService) SubmitReview(reviewerID int, req models.SubmitReviewRequest) error {
	// Complete the task
	if err := s.taskRepo.CompleteTask(req.TaskID, reviewerID); err != nil {
		return errors.New("task not found or already completed")
	}

	// Create review result
	result := &models.ReviewResult{
		TaskID:     req.TaskID,
		ReviewerID: reviewerID,
		IsApproved: req.IsApproved,
		Tags:       req.Tags,
		Reason:     req.Reason,
	}

	if err := s.taskRepo.CreateReviewResult(result); err != nil {
		return err
	}

	// Remove from Redis
	userClaimedKey := fmt.Sprintf("task:claimed:%d", reviewerID)
	lockKey := fmt.Sprintf("task:lock:%d", req.TaskID)

	pipe := s.rdb.Pipeline()
	pipe.SRem(s.ctx, userClaimedKey, req.TaskID)
	pipe.Del(s.ctx, lockKey)
	_, err := pipe.Exec(s.ctx)

	if err != nil {
		log.Printf("Redis error when submitting review: %v", err)
	}

	// Update statistics in Redis
	s.updateStats(result)

	return nil
}

// SubmitBatchReviews submits multiple reviews at once
func (s *TaskService) SubmitBatchReviews(reviewerID int, reviews []models.SubmitReviewRequest) error {
	for _, review := range reviews {
		if err := s.SubmitReview(reviewerID, review); err != nil {
			return err
		}
	}
	return nil
}

// updateStats updates statistics in Redis
func (s *TaskService) updateStats(result *models.ReviewResult) {
	now := time.Now()
	date := now.Format("2006-01-02")
	hour := now.Hour()

	hourlyKey := fmt.Sprintf("stats:hourly:%s:%d", date, hour)
	dailyKey := fmt.Sprintf("stats:daily:%s", date)

	pipe := s.rdb.Pipeline()
	pipe.HIncrBy(s.ctx, hourlyKey, "count", 1)
	pipe.Expire(s.ctx, hourlyKey, 7*24*time.Hour) // 7 days TTL

	pipe.HIncrBy(s.ctx, dailyKey, "count", 1)
	if result.IsApproved {
		pipe.HIncrBy(s.ctx, dailyKey, "approved", 1)
	} else {
		pipe.HIncrBy(s.ctx, dailyKey, "rejected", 1)

		// Track tag statistics
		for _, tag := range result.Tags {
			pipe.HIncrBy(s.ctx, dailyKey, fmt.Sprintf("tag:%s", tag), 1)
		}
	}
	pipe.Expire(s.ctx, dailyKey, 30*24*time.Hour) // 30 days TTL

	_, err := pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when updating stats: %v", err)
	}
}

// GetActiveTags retrieves all active tags
func (s *TaskService) GetActiveTags() ([]models.TagConfig, error) {
	return s.tagRepo.FindActive()
}

// ReleaseExpiredTasks releases tasks that have exceeded the timeout
func (s *TaskService) ReleaseExpiredTasks() error {
	timeoutMinutes := config.AppConfig.TaskTimeoutMinutes
	expiredTasks, err := s.taskRepo.FindExpiredTasks(timeoutMinutes)
	if err != nil {
		return err
	}

	for _, task := range expiredTasks {
		// Reset task in database
		if err := s.taskRepo.ResetTask(task.ID); err != nil {
			log.Printf("Error resetting task %d: %v", task.ID, err)
			continue
		}

		// Clean up Redis
		if task.ReviewerID != nil {
			userClaimedKey := fmt.Sprintf("task:claimed:%d", *task.ReviewerID)
			lockKey := fmt.Sprintf("task:lock:%d", task.ID)

			s.rdb.SRem(s.ctx, userClaimedKey, task.ID)
			s.rdb.Del(s.ctx, lockKey)
		}

		log.Printf("Released expired task %d", task.ID)
	}

	return nil
}
