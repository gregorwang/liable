package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"comment-review-platform/internal/services/base"
	redispkg "comment-review-platform/pkg/redis"
	"errors"
	"fmt"
	"log"
	"time"
)

type SecondReviewService struct {
	secondReviewRepo *repository.SecondReviewRepository
	base             *base.BaseTaskService
}

func NewSecondReviewService() *SecondReviewService {
	return &SecondReviewService{
		secondReviewRepo: repository.NewSecondReviewRepository(),
		base:             base.NewBaseTaskService(base.SecondReviewTaskServiceConfig(), redispkg.Client),
	}
}

// ClaimSecondReviewTasks allows a reviewer to claim second review tasks with custom count (1-50)
func (s *SecondReviewService) ClaimSecondReviewTasks(reviewerID int, count int) ([]models.SecondReviewTask, error) {
	// Validate count using base service
	if err := s.base.ValidateClaimCount(count); err != nil {
		return nil, err
	}

	// Check if user already has uncompleted second review tasks
	existingTasks, err := s.secondReviewRepo.GetMySecondReviewTasks(reviewerID)
	if err != nil {
		return nil, err
	}

	if err := s.base.CheckExistingTasks(len(existingTasks)); err != nil {
		return nil, err
	}

	// Claim tasks from database
	tasks, err := s.secondReviewRepo.ClaimSecondReviewTasks(reviewerID, count)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return []models.SecondReviewTask{}, nil
	}

	// Track tasks in Redis using base service
	taskIDs := make([]int, len(tasks))
	for i, task := range tasks {
		taskIDs[i] = task.ID
	}
	if err := s.base.TrackClaimedTasks(reviewerID, taskIDs); err != nil {
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

	// Cleanup Redis tracking using base service
	s.base.CleanupSingleTask(reviewerID, req.TaskID)

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
	if s.base.Rdb == nil {
		return
	}

	now := time.Now()
	date := now.Format("2006-01-02")
	hour := now.Hour()

	hourlyKey := fmt.Sprintf("stats:second:hourly:%s:%d", date, hour)
	dailyKey := fmt.Sprintf("stats:second:daily:%s", date)

	pipe := s.base.Rdb.Pipeline()
	pipe.HIncrBy(s.base.Ctx, hourlyKey, "count", 1)
	pipe.Expire(s.base.Ctx, hourlyKey, 7*24*time.Hour) // 7 days TTL

	pipe.HIncrBy(s.base.Ctx, dailyKey, "count", 1)
	if result.IsApproved {
		pipe.HIncrBy(s.base.Ctx, dailyKey, "approved", 1)
	} else {
		pipe.HIncrBy(s.base.Ctx, dailyKey, "rejected", 1)

		// Track tag statistics
		for _, tag := range result.Tags {
			pipe.HIncrBy(s.base.Ctx, dailyKey, fmt.Sprintf("tag:%s", tag), 1)
		}
	}
	pipe.Expire(s.base.Ctx, dailyKey, 30*24*time.Hour) // 30 days TTL

	_, err := pipe.Exec(s.base.Ctx)
	if err != nil {
		log.Printf("Redis error when updating second review stats: %v", err)
	}
}

// ReturnSecondReviewTasks allows a reviewer to return second review tasks back to the pool
func (s *SecondReviewService) ReturnSecondReviewTasks(reviewerID int, taskIDs []int) (int, error) {
	// Validate return count using base service
	if err := s.base.ValidateReturnCount(len(taskIDs)); err != nil {
		return 0, err
	}

	// Return tasks in database
	returnedCount, err := s.secondReviewRepo.ReturnSecondReviewTasks(taskIDs, reviewerID)
	if err != nil {
		return 0, err
	}

	if returnedCount == 0 {
		return 0, errors.New("no second review tasks were returned, please check if the tasks belong to you")
	}

	// Cleanup Redis tracking using base service
	s.base.CleanupTaskTracking(reviewerID, taskIDs)

	return returnedCount, nil
}

// ReleaseExpiredSecondReviewTasks releases second review tasks that have exceeded the timeout
func (s *SecondReviewService) ReleaseExpiredSecondReviewTasks() error {
	timeoutMinutes := s.base.GetTaskTimeoutMinutes()
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

		// Clean up Redis using base service
		if task.ReviewerID != nil {
			s.base.CleanupSingleTask(*task.ReviewerID, task.ID)
		}

		log.Printf("Released expired second review task %d", task.ID)
	}

	return nil
}
