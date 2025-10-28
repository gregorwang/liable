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

type QualityCheckService struct {
	qcRepo *repository.QualityCheckRepository
	rdb    *redis.Client
	ctx    context.Context
}

func NewQualityCheckService() *QualityCheckService {
	return &QualityCheckService{
		qcRepo: repository.NewQualityCheckRepository(),
		rdb:    redispkg.Client,
		ctx:    context.Background(),
	}
}

// ClaimQCTasks allows a reviewer to claim quality check tasks with custom count (1-50)
func (s *QualityCheckService) ClaimQCTasks(reviewerID int, count int) ([]models.QualityCheckTask, error) {
	// Validate count (1-50)
	if count < 1 || count > 50 {
		return nil, errors.New("claim count must be between 1 and 50")
	}

	// Check if user already has uncompleted QC tasks
	existingTasks, err := s.qcRepo.GetMyQCTasks(reviewerID)
	if err != nil {
		return nil, err
	}

	if len(existingTasks) > 0 {
		return nil, fmt.Errorf("you still have %d uncompleted QC tasks, please complete or return them first", len(existingTasks))
	}

	// Claim tasks from database
	tasks, err := s.qcRepo.ClaimQCTasks(reviewerID, count)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return []models.QualityCheckTask{}, nil
	}

	// Add tasks to Redis for tracking
	userClaimedKey := fmt.Sprintf("qc_task:claimed:%d", reviewerID)
	timeout := time.Duration(config.AppConfig.TaskTimeoutMinutes) * time.Minute

	pipe := s.rdb.Pipeline()
	for _, task := range tasks {
		// Add to user's claimed set
		pipe.SAdd(s.ctx, userClaimedKey, task.ID)

		// Set lock for each task
		lockKey := fmt.Sprintf("qc_task:lock:%d", task.ID)
		pipe.Set(s.ctx, lockKey, reviewerID, timeout)
	}
	pipe.Expire(s.ctx, userClaimedKey, timeout)

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when claiming QC tasks: %v", err)
	}

	return tasks, nil
}

// GetMyQCTasks retrieves the current user's in-progress quality check tasks
func (s *QualityCheckService) GetMyQCTasks(reviewerID int) ([]models.QualityCheckTask, error) {
	return s.qcRepo.GetMyQCTasks(reviewerID)
}

// SubmitQCReview submits a quality check result
func (s *QualityCheckService) SubmitQCReview(reviewerID int, req models.SubmitQCRequest) error {
	// Complete the QC task
	if err := s.qcRepo.CompleteQCTask(req.TaskID, reviewerID); err != nil {
		return errors.New("QC task not found or already completed")
	}

	// Create QC result
	result := &models.QualityCheckResult{
		QCTaskID:   req.TaskID,
		ReviewerID: reviewerID,
		IsPassed:   req.IsPassed,
		ErrorType:  req.ErrorType,
		QCComment:  req.QCComment,
	}

	if err := s.qcRepo.CreateQCResult(result); err != nil {
		return err
	}

	// Remove from Redis
	userClaimedKey := fmt.Sprintf("qc_task:claimed:%d", reviewerID)
	lockKey := fmt.Sprintf("qc_task:lock:%d", req.TaskID)

	pipe := s.rdb.Pipeline()
	pipe.SRem(s.ctx, userClaimedKey, req.TaskID)
	pipe.Del(s.ctx, lockKey)
	_, err := pipe.Exec(s.ctx)

	if err != nil {
		log.Printf("Redis error when submitting QC review: %v", err)
	}

	return nil
}

// SubmitBatchQCReviews submits multiple quality check reviews at once
func (s *QualityCheckService) SubmitBatchQCReviews(reviewerID int, reviews []models.SubmitQCRequest) error {
	for _, review := range reviews {
		if err := s.SubmitQCReview(reviewerID, review); err != nil {
			return err
		}
	}
	return nil
}

// ReturnQCTasks allows a reviewer to return quality check tasks back to the pool
func (s *QualityCheckService) ReturnQCTasks(reviewerID int, taskIDs []int) (int, error) {
	// Validate task count (1-50)
	if len(taskIDs) < 1 || len(taskIDs) > 50 {
		return 0, errors.New("return count must be between 1 and 50")
	}

	// Return tasks in database
	returnedCount, err := s.qcRepo.ReturnQCTasks(taskIDs, reviewerID)
	if err != nil {
		return 0, err
	}

	if returnedCount == 0 {
		return 0, errors.New("no QC tasks were returned, please check if the tasks belong to you")
	}

	// Clean up Redis
	userClaimedKey := fmt.Sprintf("qc_task:claimed:%d", reviewerID)
	pipe := s.rdb.Pipeline()

	for _, taskID := range taskIDs {
		// Remove from user's claimed set
		pipe.SRem(s.ctx, userClaimedKey, taskID)

		// Remove task lock
		lockKey := fmt.Sprintf("qc_task:lock:%d", taskID)
		pipe.Del(s.ctx, lockKey)
	}

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when returning QC tasks: %v", err)
	}

	return returnedCount, nil
}

// GetQCStats gets quality check statistics for a reviewer
func (s *QualityCheckService) GetQCStats(reviewerID int) (*models.QCStats, error) {
	return s.qcRepo.GetQCStats(reviewerID)
}

// ReleaseExpiredQCTasks releases quality check tasks that have exceeded the timeout
func (s *QualityCheckService) ReleaseExpiredQCTasks() error {
	timeoutMinutes := config.AppConfig.TaskTimeoutMinutes
	expiredTasks, err := s.qcRepo.FindExpiredQCTasks(timeoutMinutes)
	if err != nil {
		return err
	}

	for _, task := range expiredTasks {
		// Reset task in database
		if err := s.qcRepo.ResetQCTask(task.ID); err != nil {
			log.Printf("Error resetting QC task %d: %v", task.ID, err)
			continue
		}

		// Clean up Redis
		if task.ReviewerID != nil {
			userClaimedKey := fmt.Sprintf("qc_task:claimed:%d", *task.ReviewerID)
			lockKey := fmt.Sprintf("qc_task:lock:%d", task.ID)

			s.rdb.SRem(s.ctx, userClaimedKey, task.ID)
			s.rdb.Del(s.ctx, lockKey)
		}

		log.Printf("Released expired QC task %d", task.ID)
	}

	return nil
}
