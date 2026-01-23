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

type AIHumanDiffService struct {
	diffRepo *repository.AIHumanDiffRepository
	tagRepo  *repository.TagRepository
	base     *base.BaseTaskService
}

func NewAIHumanDiffService() *AIHumanDiffService {
	return &AIHumanDiffService{
		diffRepo: repository.NewAIHumanDiffRepository(),
		tagRepo:  repository.NewTagRepository(),
		base:     base.NewBaseTaskService(base.AIHumanDiffTaskServiceConfig(), redispkg.Client),
	}
}

// ClaimDiffTasks allows a reviewer to claim AI diff tasks with custom count (1-50)
func (s *AIHumanDiffService) ClaimDiffTasks(reviewerID int, count int) ([]models.AIHumanDiffTask, error) {
	if err := s.base.ValidateClaimCount(count); err != nil {
		return nil, err
	}

	existingTasks, err := s.diffRepo.GetMyDiffTasks(reviewerID)
	if err != nil {
		return nil, err
	}
	if err := s.base.CheckExistingTasks(len(existingTasks)); err != nil {
		return nil, err
	}

	tasks, err := s.diffRepo.ClaimDiffTasks(reviewerID, count)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return []models.AIHumanDiffTask{}, nil
	}

	taskIDs := make([]int, len(tasks))
	for i, task := range tasks {
		taskIDs[i] = task.ID
	}
	if err := s.base.TrackClaimedTasks(reviewerID, taskIDs); err != nil {
		log.Printf("Redis error when claiming AI diff tasks: %v", err)
		if _, resetErr := s.diffRepo.ReturnDiffTasks(taskIDs, reviewerID); resetErr != nil {
			log.Printf("Failed to rollback AI diff tasks after Redis error: %v", resetErr)
		}
		return nil, errors.New("failed to claim tasks, please retry")
	}

	return tasks, nil
}

// GetMyDiffTasks retrieves the current user's AI diff tasks
func (s *AIHumanDiffService) GetMyDiffTasks(reviewerID int) ([]models.AIHumanDiffTask, error) {
	return s.diffRepo.GetMyDiffTasks(reviewerID)
}

// SubmitDiffReview submits a single AI diff review result
func (s *AIHumanDiffService) SubmitDiffReview(reviewerID int, req models.SubmitAIHumanDiffRequest) error {
	if err := validateTags(s.tagRepo, "comment", req.Tags); err != nil {
		return err
	}

	if err := s.diffRepo.CompleteDiffTask(req.TaskID, reviewerID); err != nil {
		return errors.New("AI diff task not found or already completed")
	}

	result := &models.AIHumanDiffResult{
		TaskID:     req.TaskID,
		ReviewerID: reviewerID,
		IsApproved: req.IsApproved,
		Tags:       req.Tags,
		Reason:     req.Reason,
	}

	if _, err := s.diffRepo.CreateDiffResult(result); err != nil {
		return err
	}

	s.base.CleanupSingleTask(reviewerID, req.TaskID)
	return nil
}

// SubmitBatchDiffReviews submits multiple AI diff reviews at once
func (s *AIHumanDiffService) SubmitBatchDiffReviews(reviewerID int, reviews []models.SubmitAIHumanDiffRequest) error {
	var failed []string
	for _, review := range reviews {
		if err := s.SubmitDiffReview(reviewerID, review); err != nil {
			failed = append(failed, fmt.Sprintf("task %d: %v", review.TaskID, err))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to submit %d AI diff reviews: %s", len(failed), strings.Join(failed, "; "))
	}
	return nil
}

// ReturnDiffTasks allows a reviewer to return AI diff tasks back to the pool
func (s *AIHumanDiffService) ReturnDiffTasks(reviewerID int, taskIDs []int) (int, error) {
	if err := s.base.ValidateReturnCount(len(taskIDs)); err != nil {
		return 0, err
	}

	returnedCount, err := s.diffRepo.ReturnDiffTasks(taskIDs, reviewerID)
	if err != nil {
		return 0, err
	}
	if returnedCount == 0 {
		return 0, errors.New("no AI diff tasks were returned, please check if the tasks belong to you")
	}

	s.base.CleanupTaskTracking(reviewerID, taskIDs)
	return returnedCount, nil
}

// ReleaseExpiredDiffTasks releases AI diff tasks that have exceeded the timeout
func (s *AIHumanDiffService) ReleaseExpiredDiffTasks() error {
	timeoutMinutes := s.base.GetTaskTimeoutMinutes()
	expiredTasks, err := s.diffRepo.FindExpiredDiffTasks(timeoutMinutes)
	if err != nil {
		return err
	}

	for _, task := range expiredTasks {
		if err := s.diffRepo.ResetDiffTask(task.ID); err != nil {
			log.Printf("Error resetting AI diff task %d: %v", task.ID, err)
			continue
		}

		if task.ReviewerID != nil {
			s.base.CleanupSingleTask(*task.ReviewerID, task.ID)
		}
		log.Printf("Released expired AI diff task %d", task.ID)
	}

	return nil
}
