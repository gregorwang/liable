package services

import (
	"comment-review-platform/internal/config"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"comment-review-platform/internal/services/base"
	redispkg "comment-review-platform/pkg/redis"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type SecondReviewService struct {
	secondReviewRepo  *repository.SecondReviewRepository
	tagRepo           *repository.TagRepository
	commentRepo       *repository.CommentRepository
	punishmentService *PunishmentService
	base              *base.BaseTaskService
}

func NewSecondReviewService() *SecondReviewService {
	return &SecondReviewService{
		secondReviewRepo:  repository.NewSecondReviewRepository(),
		tagRepo:           repository.NewTagRepository(),
		commentRepo:       repository.NewCommentRepository(),
		punishmentService: NewPunishmentService(),
		base:              base.NewBaseTaskService(base.SecondReviewTaskServiceConfig(), redispkg.Client),
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
		if _, resetErr := s.secondReviewRepo.ReturnSecondReviewTasks(taskIDs, reviewerID); resetErr != nil {
			log.Printf("Failed to rollback second review tasks after Redis error: %v", resetErr)
		}
		return nil, errors.New("failed to claim tasks, please retry")
	}

	return tasks, nil
}

// GetMySecondReviewTasks retrieves the current user's in-progress second review tasks
func (s *SecondReviewService) GetMySecondReviewTasks(reviewerID int) ([]models.SecondReviewTask, error) {
	return s.secondReviewRepo.GetMySecondReviewTasks(reviewerID)
}

// SubmitSecondReview submits a second review result
func (s *SecondReviewService) SubmitSecondReview(reviewerID int, req models.SubmitSecondReviewRequest) error {
	if err := validateTags(s.tagRepo, "comment", req.Tags); err != nil {
		return err
	}

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

	createdResult, err := s.secondReviewRepo.CreateSecondReviewResult(result)
	if err != nil {
		return err
	}

	commentID, err := s.secondReviewRepo.GetCommentIDByTaskID(req.TaskID)
	if err != nil {
		return err
	}
	status := "rejected"
	if result.IsApproved {
		status = "approved"
	}
	if err := s.commentRepo.UpdateModerationStatus(commentID, status); err != nil {
		return err
	}

	// Cleanup Redis tracking using base service
	s.base.CleanupSingleTask(reviewerID, req.TaskID)

	// Update statistics in Redis
	if createdResult {
		s.updateSecondReviewStats(result)
	}

	// Create punishment record if rejected
	if createdResult && !result.IsApproved && (config.AppConfig == nil || config.AppConfig.PunishmentEnabled) {
		go s.createPunishmentRecord(commentID, result, req)
	}

	return nil
}

// SubmitBatchSecondReviews submits multiple second reviews at once
func (s *SecondReviewService) SubmitBatchSecondReviews(reviewerID int, reviews []models.SubmitSecondReviewRequest) error {
	var failed []string
	for _, review := range reviews {
		if err := s.SubmitSecondReview(reviewerID, review); err != nil {
			failed = append(failed, fmt.Sprintf("task %d: %v", review.TaskID, err))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to submit %d second reviews: %s", len(failed), strings.Join(failed, "; "))
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

func (s *SecondReviewService) createPunishmentRecord(
	commentID int64,
	result *models.SecondReviewResult,
	req models.SubmitSecondReviewRequest,
) {
	if s.punishmentService == nil {
		return
	}

	authorID, err := s.commentRepo.GetCommentAuthorID(commentID)
	if err != nil {
		log.Printf("Failed to get comment author: %v", err)
		return
	}

	violationLevel := s.mapTagsToViolationLevel(result.Tags)
	taskID := req.TaskID
	punishment := &models.Punishment{
		UserID:         authorID,
		ContentType:    "comment",
		ContentID:      strconv.FormatInt(commentID, 10),
		ReviewTaskID:   &taskID,
		ViolationLevel: violationLevel,
		ViolationTags:  result.Tags,
		Reason:         result.Reason,
		Status:         "pending",
	}

	if _, err := s.punishmentService.CreatePunishment(punishment); err != nil {
		log.Printf("Failed to create punishment record: %v", err)
	}
}

// mapTagsToViolationLevel maps tag names to a violation level.
func (s *SecondReviewService) mapTagsToViolationLevel(tags []string) int {
	highRiskTags := map[string]struct{}{
		"涉政": {},
		"涉黄": {},
		"诈骗": {},
	}
	mediumRiskTags := map[string]struct{}{
		"辱骂": {},
		"骚扰": {},
		"广告": {},
	}

	for _, tag := range tags {
		if _, ok := highRiskTags[tag]; ok {
			return 4
		}
	}

	for _, tag := range tags {
		if _, ok := mediumRiskTags[tag]; ok {
			return 2
		}
	}

	return 1
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
