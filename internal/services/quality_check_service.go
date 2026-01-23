package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"comment-review-platform/internal/services/base"
	redispkg "comment-review-platform/pkg/redis"
	"errors"
	"fmt"
	"log"
	"strings"
)

type QualityCheckService struct {
	qcRepo *repository.QualityCheckRepository
	base   *base.BaseTaskService
}

func NewQualityCheckService() *QualityCheckService {
	return &QualityCheckService{
		qcRepo: repository.NewQualityCheckRepository(),
		base:   base.NewBaseTaskService(base.QualityCheckTaskServiceConfig(), redispkg.Client),
	}
}

// ClaimQCTasks allows a reviewer to claim quality check tasks with custom count (1-50)
func (s *QualityCheckService) ClaimQCTasks(reviewerID int, count int) ([]models.QualityCheckTask, error) {
	// Validate count using base service
	if err := s.base.ValidateClaimCount(count); err != nil {
		return nil, err
	}

	// Check if user already has uncompleted QC tasks
	existingTasks, err := s.qcRepo.GetMyQCTasks(reviewerID)
	if err != nil {
		return nil, err
	}

	if err := s.base.CheckExistingTasks(len(existingTasks)); err != nil {
		return nil, err
	}

	// Claim tasks from database
	tasks, err := s.qcRepo.ClaimQCTasks(reviewerID, count)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return []models.QualityCheckTask{}, nil
	}

	// Track tasks in Redis using base service
	taskIDs := make([]int, len(tasks))
	for i, task := range tasks {
		taskIDs[i] = task.ID
	}
	if err := s.base.TrackClaimedTasks(reviewerID, taskIDs); err != nil {
		log.Printf("Redis error when claiming QC tasks: %v", err)
		if _, resetErr := s.qcRepo.ReturnQCTasks(taskIDs, reviewerID); resetErr != nil {
			log.Printf("Failed to rollback QC tasks after Redis error: %v", resetErr)
		}
		return nil, errors.New("failed to claim tasks, please retry")
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

	if _, err := s.qcRepo.CreateQCResult(result); err != nil {
		return err
	}

	// Cleanup Redis tracking using base service
	s.base.CleanupSingleTask(reviewerID, req.TaskID)

	return nil
}

// SubmitBatchQCReviews submits multiple quality check reviews at once
func (s *QualityCheckService) SubmitBatchQCReviews(reviewerID int, reviews []models.SubmitQCRequest) error {
	var failed []string
	for _, review := range reviews {
		if err := s.SubmitQCReview(reviewerID, review); err != nil {
			failed = append(failed, fmt.Sprintf("task %d: %v", review.TaskID, err))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to submit %d QC reviews: %s", len(failed), strings.Join(failed, "; "))
	}
	return nil
}

// ReturnQCTasks allows a reviewer to return quality check tasks back to the pool
func (s *QualityCheckService) ReturnQCTasks(reviewerID int, taskIDs []int) (int, error) {
	// Validate return count using base service
	if err := s.base.ValidateReturnCount(len(taskIDs)); err != nil {
		return 0, err
	}

	// Return tasks in database
	returnedCount, err := s.qcRepo.ReturnQCTasks(taskIDs, reviewerID)
	if err != nil {
		return 0, err
	}

	if returnedCount == 0 {
		return 0, errors.New("no QC tasks were returned, please check if the tasks belong to you")
	}

	// Cleanup Redis tracking using base service
	s.base.CleanupTaskTracking(reviewerID, taskIDs)

	return returnedCount, nil
}

// GetQCStats gets quality check statistics for a reviewer
func (s *QualityCheckService) GetQCStats(reviewerID int) (*models.QCStats, error) {
	return s.qcRepo.GetQCStats(reviewerID)
}

// ReleaseExpiredQCTasks releases quality check tasks that have exceeded the timeout
func (s *QualityCheckService) ReleaseExpiredQCTasks() error {
	timeoutMinutes := s.base.GetTaskTimeoutMinutes()
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

		// Clean up Redis using base service
		if task.ReviewerID != nil {
			s.base.CleanupSingleTask(*task.ReviewerID, task.ID)
		}

		log.Printf("Released expired QC task %d", task.ID)
	}

	return nil
}
