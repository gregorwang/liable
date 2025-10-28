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
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

type TaskService struct {
	taskRepo         *repository.TaskRepository
	secondReviewRepo *repository.SecondReviewRepository
	tagRepo          *repository.TagRepository
	rdb              *redis.Client
	ctx              context.Context
}

func NewTaskService() *TaskService {
	return &TaskService{
		taskRepo:         repository.NewTaskRepository(),
		secondReviewRepo: repository.NewSecondReviewRepository(),
		tagRepo:          repository.NewTagRepository(),
		rdb:              redispkg.Client,
		ctx:              context.Background(),
	}
}

// ClaimTasks allows a reviewer to claim tasks with custom count (1-50)
func (s *TaskService) ClaimTasks(reviewerID int, count int) ([]models.ReviewTask, error) {
	// Validate count (1-50)
	if count < 1 || count > 50 {
		return nil, errors.New("claim count must be between 1 and 50")
	}

	// Check if user already has uncompleted tasks
	existingTasks, err := s.taskRepo.GetMyTasks(reviewerID)
	if err != nil {
		return nil, err
	}

	if len(existingTasks) > 0 {
		return nil, fmt.Errorf("you still have %d uncompleted tasks, please complete or return them first", len(existingTasks))
	}

	// Claim tasks from database
	tasks, err := s.taskRepo.ClaimTasks(reviewerID, count)
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

	// If first review is not approved, create second review task
	if !req.IsApproved {
		// Get the comment ID from the task
		tasks, err := s.taskRepo.FindTasksWithComments([]int{req.TaskID})
		if err != nil {
			log.Printf("Error getting comment ID for second review task: %v", err)
		} else if len(tasks) > 0 {
			commentID := tasks[0].CommentID

			// Create second review task
			if err := s.secondReviewRepo.CreateSecondReviewTask(result.ID, commentID); err != nil {
				log.Printf("Error creating second review task: %v", err)
			} else {
				// Push to Redis second review queue
				queueKey := "review:queue:second"
				if err := s.rdb.LPush(s.ctx, queueKey, commentID).Err(); err != nil {
					log.Printf("Redis error pushing to second review queue: %v", err)
				}
			}
		}
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

// ReturnTasks allows a reviewer to return tasks back to the pool
func (s *TaskService) ReturnTasks(reviewerID int, taskIDs []int) (int, error) {
	// Validate task count (1-50)
	if len(taskIDs) < 1 || len(taskIDs) > 50 {
		return 0, errors.New("return count must be between 1 and 50")
	}

	// Return tasks in database
	returnedCount, err := s.taskRepo.ReturnTasks(taskIDs, reviewerID)
	if err != nil {
		return 0, err
	}

	if returnedCount == 0 {
		return 0, errors.New("no tasks were returned, please check if the tasks belong to you")
	}

	// Clean up Redis
	userClaimedKey := fmt.Sprintf("task:claimed:%d", reviewerID)
	pipe := s.rdb.Pipeline()

	for _, taskID := range taskIDs {
		// Remove from user's claimed set
		pipe.SRem(s.ctx, userClaimedKey, taskID)

		// Remove task lock
		lockKey := fmt.Sprintf("task:lock:%d", taskID)
		pipe.Del(s.ctx, lockKey)
	}

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when returning tasks: %v", err)
	}

	return returnedCount, nil
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

// SearchTasks searches review tasks with filters and pagination
func (s *TaskService) SearchTasks(req models.SearchTasksRequest) (*models.SearchTasksResponse, error) {
	// Set default pagination values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	// Limit maximum page size to 100
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Set default queue type
	if req.QueueType == "" {
		req.QueueType = "all"
	}

	var allResults []models.TaskSearchResult
	var totalCount int

	// Search first review tasks
	if req.QueueType == "first" || req.QueueType == "all" {
		firstResults, firstTotal, err := s.taskRepo.SearchTasks(req)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, firstResults...)
		totalCount += firstTotal
	}

	// Search second review tasks
	if req.QueueType == "second" || req.QueueType == "all" {
		secondResults, secondTotal, err := s.secondReviewRepo.SearchSecondReviewTasks(req)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, secondResults...)
		totalCount += secondTotal
	}

	// Sort combined results by completion time (most recent first)
	// Note: This is a simple approach. For better performance with large datasets,
	// consider implementing database-level sorting with UNION queries
	sort.Slice(allResults, func(i, j int) bool {
		if allResults[i].CompletedAt == nil && allResults[j].CompletedAt == nil {
			return allResults[i].CreatedAt.After(allResults[j].CreatedAt)
		}
		if allResults[i].CompletedAt == nil {
			return false
		}
		if allResults[j].CompletedAt == nil {
			return true
		}
		return allResults[i].CompletedAt.After(*allResults[j].CompletedAt)
	})

	// Apply pagination to combined results
	offset := (req.Page - 1) * req.PageSize
	end := offset + req.PageSize
	if end > len(allResults) {
		end = len(allResults)
	}
	if offset >= len(allResults) {
		allResults = []models.TaskSearchResult{}
	} else {
		allResults = allResults[offset:end]
	}

	// Calculate total pages
	totalPages := totalCount / req.PageSize
	if totalCount%req.PageSize > 0 {
		totalPages++
	}

	response := &models.SearchTasksResponse{
		Data:       allResults,
		Total:      totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	return response, nil
}
