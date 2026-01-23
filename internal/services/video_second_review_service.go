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
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type VideoSecondReviewService struct {
	secondReviewRepo *repository.VideoSecondReviewRepository
	videoRepo        *repository.VideoRepository
	rdb              *redis.Client
	ctx              context.Context
}

func NewVideoSecondReviewService() *VideoSecondReviewService {
	return &VideoSecondReviewService{
		secondReviewRepo: repository.NewVideoSecondReviewRepository(),
		videoRepo:        repository.NewVideoRepository(),
		rdb:              redispkg.Client,
		ctx:              context.Background(),
	}
}

// ClaimSecondReviewTasks allows a reviewer to claim second review tasks
func (s *VideoSecondReviewService) ClaimSecondReviewTasks(reviewerID int, count int) ([]models.VideoSecondReviewTask, error) {
	// Validate count (1-50)
	if count < 1 || count > 50 {
		return nil, errors.New("claim count must be between 1 and 50")
	}

	// Check if user already has uncompleted tasks
	existingTasks, err := s.secondReviewRepo.GetMySecondReviewTasks(reviewerID)
	if err != nil {
		return nil, err
	}

	if len(existingTasks) > 0 {
		return nil, fmt.Errorf("you still have %d uncompleted video second review tasks, please complete or return them first", len(existingTasks))
	}

	// Claim tasks from database
	tasks, err := s.secondReviewRepo.ClaimSecondReviewTasks(reviewerID, count)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return []models.VideoSecondReviewTask{}, nil
	}

	// Add tasks to Redis for tracking
	userClaimedKey := fmt.Sprintf("video:second:claimed:%d", reviewerID)
	timeout := time.Duration(config.AppConfig.TaskTimeoutMinutes) * time.Minute

	pipe := s.rdb.Pipeline()
	for _, task := range tasks {
		// Add to user's claimed set
		pipe.SAdd(s.ctx, userClaimedKey, task.ID)

		// Set lock for each task
		lockKey := fmt.Sprintf("video:second:lock:%d", task.ID)
		pipe.Set(s.ctx, lockKey, reviewerID, timeout)
	}
	pipe.Expire(s.ctx, userClaimedKey, timeout)

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when claiming video second review tasks: %v", err)
	}

	return tasks, nil
}

// GetMySecondReviewTasks retrieves the current user's in-progress second review tasks
func (s *VideoSecondReviewService) GetMySecondReviewTasks(reviewerID int) ([]models.VideoSecondReviewTask, error) {
	return s.secondReviewRepo.GetMySecondReviewTasks(reviewerID)
}

// SubmitSecondReview submits a second review result
func (s *VideoSecondReviewService) SubmitSecondReview(reviewerID int, req models.SubmitVideoSecondReviewRequest) error {
	// Complete the task
	if err := s.secondReviewRepo.CompleteSecondReviewTask(req.TaskID, reviewerID); err != nil {
		return errors.New("task not found or already completed")
	}

	// Calculate overall score
	overallScore := req.QualityDimensions.ContentQuality.Score +
		req.QualityDimensions.TechnicalQuality.Score +
		req.QualityDimensions.Compliance.Score +
		req.QualityDimensions.EngagementPotential.Score

	// Create review result
	result := &models.VideoSecondReviewResult{
		SecondTaskID:      req.TaskID,
		ReviewerID:        reviewerID,
		IsApproved:        req.IsApproved,
		QualityDimensions: req.QualityDimensions,
		OverallScore:      overallScore,
		TrafficPoolResult: req.TrafficPoolResult,
		Reason:            req.Reason,
	}

	createdResult, err := s.secondReviewRepo.CreateSecondReviewResult(result)
	if err != nil {
		return err
	}

	// Update video status based on second review result
	tasks, err := s.secondReviewRepo.FindSecondReviewTasksWithDetails([]int{req.TaskID})
	if err == nil && len(tasks) > 0 {
		videoID := tasks[0].VideoID
		status := "second_review_completed"
		if err := s.videoRepo.UpdateVideoStatus(videoID, status); err != nil {
			log.Printf("Error updating video status: %v", err)
		}
	}

	// Remove from Redis
	userClaimedKey := fmt.Sprintf("video:second:claimed:%d", reviewerID)
	lockKey := fmt.Sprintf("video:second:lock:%d", req.TaskID)

	pipe := s.rdb.Pipeline()
	pipe.SRem(s.ctx, userClaimedKey, req.TaskID)
	pipe.Del(s.ctx, lockKey)
	_, err = pipe.Exec(s.ctx)

	if err != nil {
		log.Printf("Redis error when submitting video second review: %v", err)
	}

	// Update statistics in Redis
	if createdResult {
		s.updateVideoSecondReviewStats(result)
	}

	return nil
}

// SubmitBatchSecondReviews submits multiple second reviews at once
func (s *VideoSecondReviewService) SubmitBatchSecondReviews(reviewerID int, reviews []models.SubmitVideoSecondReviewRequest) error {
	var failed []string
	for _, review := range reviews {
		if err := s.SubmitSecondReview(reviewerID, review); err != nil {
			failed = append(failed, fmt.Sprintf("task %d: %v", review.TaskID, err))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to submit %d video second reviews: %s", len(failed), strings.Join(failed, "; "))
	}
	return nil
}

// ReturnSecondReviewTasks allows a reviewer to return second review tasks back to the pool
func (s *VideoSecondReviewService) ReturnSecondReviewTasks(reviewerID int, taskIDs []int) (int, error) {
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
		return 0, errors.New("no tasks were returned, please check if the tasks belong to you")
	}

	// Clean up Redis
	userClaimedKey := fmt.Sprintf("video:second:claimed:%d", reviewerID)
	pipe := s.rdb.Pipeline()

	for _, taskID := range taskIDs {
		// Remove from user's claimed set
		pipe.SRem(s.ctx, userClaimedKey, taskID)

		// Remove task lock
		lockKey := fmt.Sprintf("video:second:lock:%d", taskID)
		pipe.Del(s.ctx, lockKey)
	}

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when returning video second review tasks: %v", err)
	}

	return returnedCount, nil
}

// ReleaseExpiredSecondReviewTasks releases second review tasks that have exceeded the timeout
func (s *VideoSecondReviewService) ReleaseExpiredSecondReviewTasks() error {
	timeoutMinutes := config.AppConfig.TaskTimeoutMinutes
	expiredTasks, err := s.secondReviewRepo.FindExpiredSecondReviewTasks(timeoutMinutes)
	if err != nil {
		return err
	}

	for _, task := range expiredTasks {
		// Reset task in database
		if err := s.secondReviewRepo.ResetSecondReviewTask(task.ID); err != nil {
			log.Printf("Error resetting video second review task %d: %v", task.ID, err)
			continue
		}

		// Clean up Redis
		if task.ReviewerID != nil {
			userClaimedKey := fmt.Sprintf("video:second:claimed:%d", *task.ReviewerID)
			lockKey := fmt.Sprintf("video:second:lock:%d", task.ID)

			s.rdb.SRem(s.ctx, userClaimedKey, task.ID)
			s.rdb.Del(s.ctx, lockKey)
		}

		log.Printf("Released expired video second review task %d", task.ID)
	}

	return nil
}

// updateVideoSecondReviewStats updates statistics in Redis
func (s *VideoSecondReviewService) updateVideoSecondReviewStats(result *models.VideoSecondReviewResult) {
	now := time.Now()
	date := now.Format("2006-01-02")
	hour := now.Hour()

	hourlyKey := fmt.Sprintf("video:second:stats:hourly:%s:%d", date, hour)
	dailyKey := fmt.Sprintf("video:second:stats:daily:%s", date)

	pipe := s.rdb.Pipeline()
	pipe.HIncrBy(s.ctx, hourlyKey, "count", 1)
	pipe.Expire(s.ctx, hourlyKey, 7*24*time.Hour) // 7 days TTL

	pipe.HIncrBy(s.ctx, dailyKey, "count", 1)
	if result.IsApproved {
		pipe.HIncrBy(s.ctx, dailyKey, "approved", 1)
	} else {
		pipe.HIncrBy(s.ctx, dailyKey, "rejected", 1)
	}

	// Track overall score distribution
	scoreRange := "low"
	if result.OverallScore >= 30 {
		scoreRange = "high"
	} else if result.OverallScore >= 20 {
		scoreRange = "medium"
	}
	pipe.HIncrBy(s.ctx, dailyKey, fmt.Sprintf("score:%s", scoreRange), 1)

	pipe.Expire(s.ctx, dailyKey, 30*24*time.Hour) // 30 days TTL

	_, err := pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when updating video second review stats: %v", err)
	}
}

