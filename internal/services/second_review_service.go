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

type SecondReviewService struct {
	secondReviewRepo *repository.SecondReviewRepository
	rdb              *redis.Client
	ctx              context.Context
}

func NewSecondReviewService() *SecondReviewService {
	return &SecondReviewService{
		secondReviewRepo: repository.NewSecondReviewRepository(),
		rdb:              redispkg.Client,
		ctx:              context.Background(),
	}
}

// ClaimSecondReviewTasks allows a reviewer to claim second review tasks with custom count (1-50)
func (s *SecondReviewService) ClaimSecondReviewTasks(reviewerID int, count int) ([]models.SecondReviewTask, error) {
	// Validate count (1-50)
	if count < 1 || count > 50 {
		return nil, errors.New("claim count must be between 1 and 50")
	}

	// Check if user already has uncompleted second review tasks
	existingTasks, err := s.secondReviewRepo.GetMySecondReviewTasks(reviewerID)
	if err != nil {
		return nil, err
	}

	if len(existingTasks) > 0 {
		return nil, fmt.Errorf("you still have %d uncompleted second review tasks, please complete or return them first", len(existingTasks))
	}

	// Claim tasks from database
	tasks, err := s.secondReviewRepo.ClaimSecondReviewTasks(reviewerID, count)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return []models.SecondReviewTask{}, nil
	}

	// Add tasks to Redis for tracking
	userClaimedKey := fmt.Sprintf("second_task:claimed:%d", reviewerID)
	timeout := time.Duration(config.AppConfig.TaskTimeoutMinutes) * time.Minute

	pipe := s.rdb.Pipeline()
	for _, task := range tasks {
		// Add to user's claimed set
		pipe.SAdd(s.ctx, userClaimedKey, task.ID)

		// Set lock for each task
		lockKey := fmt.Sprintf("second_task:lock:%d", task.ID)
		pipe.Set(s.ctx, lockKey, reviewerID, timeout)
	}
	pipe.Expire(s.ctx, userClaimedKey, timeout)

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when claiming second review tasks: %v", err)
	}

	return tasks, nil
}

// GetMySecondReviewTasks retrieves the current user's in-progress second review tasks
func (s *SecondReviewService) GetMySecondReviewTasks(reviewerID int) ([]models.SecondReviewTask, error) {
	return s.secondReviewRepo.GetMySecondReviewTasks(reviewerID)
}

// SubmitSecondReview submits a second review result
func (s *SecondReviewService) SubmitSecondReview(reviewerID int, req models.SubmitSecondReviewRequest) error {
	// Complete the task
	if err := s.secondReviewRepo.CompleteSecondReviewTask(req.TaskID, reviewerID); err != nil {
		return errors.New("second review task not found or already completed")
	}

	// Create second review result
	result := &models.SecondReviewResult{
		SecondTaskID: req.TaskID,
		ReviewerID:   reviewerID,
		IsApproved:   req.IsApproved,
		Tags:         req.Tags,
		Reason:       req.Reason,
	}

	if err := s.secondReviewRepo.CreateSecondReviewResult(result); err != nil {
		return err
	}

	// Remove from Redis
	userClaimedKey := fmt.Sprintf("second_task:claimed:%d", reviewerID)
	lockKey := fmt.Sprintf("second_task:lock:%d", req.TaskID)

	pipe := s.rdb.Pipeline()
	pipe.SRem(s.ctx, userClaimedKey, req.TaskID)
	pipe.Del(s.ctx, lockKey)
	_, err := pipe.Exec(s.ctx)

	if err != nil {
		log.Printf("Redis error when submitting second review: %v", err)
	}

	// Update statistics in Redis
	s.updateSecondReviewStats(result)

	return nil
}

// SubmitBatchSecondReviews submits multiple second reviews at once
func (s *SecondReviewService) SubmitBatchSecondReviews(reviewerID int, reviews []models.SubmitSecondReviewRequest) error {
	for _, review := range reviews {
		if err := s.SubmitSecondReview(reviewerID, review); err != nil {
			return err
		}
	}
	return nil
}

// updateSecondReviewStats updates statistics in Redis for second reviews
func (s *SecondReviewService) updateSecondReviewStats(result *models.SecondReviewResult) {
	now := time.Now()
	date := now.Format("2006-01-02")
	hour := now.Hour()

	hourlyKey := fmt.Sprintf("stats:second:hourly:%s:%d", date, hour)
	dailyKey := fmt.Sprintf("stats:second:daily:%s", date)

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
		log.Printf("Redis error when updating second review stats: %v", err)
	}
}

// ReturnSecondReviewTasks allows a reviewer to return second review tasks back to the pool
func (s *SecondReviewService) ReturnSecondReviewTasks(reviewerID int, taskIDs []int) (int, error) {
	// Validate task count (1-50)
	if len(taskIDs) < 1 || len(taskIDs) > 50 {
		return 0, errors.New("return count must be between 1 and 50")
	}

	// Return tasks in database
	returnedCount, err := s.secondReviewRepo.ReturnSecondReviewTasks(taskIDs, reviewerID)
	if err != nil {
		return 0, err
	}

	if returnedCount == 0 {
		return 0, errors.New("no second review tasks were returned, please check if the tasks belong to you")
	}

	// Clean up Redis
	userClaimedKey := fmt.Sprintf("second_task:claimed:%d", reviewerID)
	pipe := s.rdb.Pipeline()

	for _, taskID := range taskIDs {
		// Remove from user's claimed set
		pipe.SRem(s.ctx, userClaimedKey, taskID)

		// Remove task lock
		lockKey := fmt.Sprintf("second_task:lock:%d", taskID)
		pipe.Del(s.ctx, lockKey)
	}

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when returning second review tasks: %v", err)
	}

	return returnedCount, nil
}

// ReleaseExpiredSecondReviewTasks releases second review tasks that have exceeded the timeout
func (s *SecondReviewService) ReleaseExpiredSecondReviewTasks() error {
	timeoutMinutes := config.AppConfig.TaskTimeoutMinutes
	expiredTasks, err := s.secondReviewRepo.FindExpiredSecondReviewTasks(timeoutMinutes)
	if err != nil {
		return err
	}

	for _, task := range expiredTasks {
		// Reset task in database
		if err := s.secondReviewRepo.ResetSecondReviewTask(task.ID); err != nil {
			log.Printf("Error resetting second review task %d: %v", task.ID, err)
			continue
		}

		// Clean up Redis
		if task.ReviewerID != nil {
			userClaimedKey := fmt.Sprintf("second_task:claimed:%d", *task.ReviewerID)
			lockKey := fmt.Sprintf("second_task:lock:%d", task.ID)

			s.rdb.SRem(s.ctx, userClaimedKey, task.ID)
			s.rdb.Del(s.ctx, lockKey)
		}

		log.Printf("Released expired second review task %d", task.ID)
	}

	return nil
}
