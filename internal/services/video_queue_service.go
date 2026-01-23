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

type VideoQueueService struct {
	queueRepo *repository.VideoQueueRepository
	rdb       *redis.Client
	ctx       context.Context
}

func NewVideoQueueService() *VideoQueueService {
	return &VideoQueueService{
		queueRepo: repository.NewVideoQueueRepository(),
		rdb:       redispkg.Client,
		ctx:       context.Background(),
	}
}

// ClaimTasks allows a reviewer to claim tasks from a specific pool
func (s *VideoQueueService) ClaimTasks(pool string, reviewerID int, count int) ([]models.VideoQueueTask, error) {
	log.Printf("📋 [DEBUG] ClaimTasks START: pool=%s, reviewerID=%d, count=%d", pool, reviewerID, count)

	// Validate pool
	log.Printf("📋 [DEBUG] ClaimTasks Step 1: Validate pool")
	if !isValidPool(pool) {
		return nil, errors.New("invalid pool: must be 100k, 1m, or 10m")
	}

	// Validate count (1-50)
	log.Printf("📋 [DEBUG] ClaimTasks Step 2: Validate count")
	if count < 1 || count > 50 {
		return nil, errors.New("claim count must be between 1 and 50")
	}

	// Check if user already has uncompleted tasks in this pool
	log.Printf("📋 [DEBUG] ClaimTasks Step 3: Check existing tasks (DB count)")
	existingCount, err := s.queueRepo.CountMyQueueTasks(pool, reviewerID)
	if err != nil {
		log.Printf("📋 [ERROR] CountMyQueueTasks failed: %v", err)
		return nil, err
	}

	if existingCount > 0 {
		log.Printf("📋 [User already has tasks] count=%d, pool=%s", existingCount, pool)
		return nil, fmt.Errorf("you still have %d uncompleted tasks in %s pool, please complete or return them first", existingCount, pool)
	}

	// Claim tasks from database
	log.Printf("📋 [DEBUG] ClaimTasks Step 4: Claim tasks from DB (transaction with lock)")
	tasks, err := s.queueRepo.ClaimQueueTasks(pool, reviewerID, count)
	if err != nil {
		log.Printf("📋 [ERROR] ClaimQueueTasks failed: %v", err)
		return nil, err
	}

	if len(tasks) == 0 {
		log.Printf("📋 [INFO] No pending tasks available")
		return []models.VideoQueueTask{}, nil
	}

	// Add tasks to Redis for tracking
	log.Printf("📋 [DEBUG] ClaimTasks Step 5: Add to Redis pipeline (%d tasks)", len(tasks))
	userClaimedKey := fmt.Sprintf("video:claimed:%d:%s", reviewerID, pool)
	timeout := time.Duration(config.AppConfig.TaskTimeoutMinutes) * time.Minute

	pipe := s.rdb.Pipeline()
	for _, task := range tasks {
		// Add to user's claimed set
		pipe.SAdd(s.ctx, userClaimedKey, task.ID)

		// Set lock for each task
		lockKey := fmt.Sprintf("video:lock:%d", task.ID)
		pipe.Set(s.ctx, lockKey, reviewerID, timeout)
	}
	pipe.Expire(s.ctx, userClaimedKey, timeout)

	log.Printf("📋 [DEBUG] ClaimTasks Step 6: Execute Redis pipeline")
	startTime := time.Now()
	_, err = pipe.Exec(s.ctx)
	redisDuration := time.Since(startTime)
	log.Printf("📋 [DEBUG] Redis pipeline executed in %v", redisDuration)
	if err != nil {
		log.Printf("📋 [ERROR] Redis error when claiming video queue tasks: %v", err)
	}

	log.Printf("📋 [DEBUG] ClaimTasks END: claimed %d tasks", len(tasks))
	return tasks, nil
}

// GetMyTasks retrieves the current user's in-progress tasks in a pool
func (s *VideoQueueService) GetMyTasks(pool string, reviewerID int) ([]models.VideoQueueTask, error) {
	if !isValidPool(pool) {
		return nil, errors.New("invalid pool: must be 100k, 1m, or 10m")
	}

	return s.queueRepo.GetMyQueueTasks(pool, reviewerID)
}

// SubmitReview submits a review result and handles queue flow
func (s *VideoQueueService) SubmitReview(pool string, reviewerID int, req models.SubmitVideoQueueReviewRequest) error {
	if !isValidPool(pool) {
		return errors.New("invalid pool: must be 100k, 1m, or 10m")
	}

	// Complete the task
	if err := s.queueRepo.CompleteQueueTask(req.TaskID, reviewerID); err != nil {
		return errors.New("task not found or already completed")
	}

	// Validate tags (max 3)
	if len(req.Tags) > 3 {
		return errors.New("maximum 3 tags allowed")
	}

	// Create review result
	result := &models.VideoQueueResult{
		TaskID:         req.TaskID,
		ReviewerID:     reviewerID,
		ReviewDecision: req.ReviewDecision,
		Reason:         req.Reason,
		Tags:           req.Tags,
	}

	createdResult, err := s.queueRepo.CreateQueueResult(result)
	if err != nil {
		return err
	}

	// Get task to retrieve video ID
	task, err := s.queueRepo.GetTaskByID(req.TaskID)
	if err != nil {
		log.Printf("Error getting task for queue flow: %v", err)
	} else {
		// Handle queue flow based on review decision
	if err := s.handleQueueFlow(pool, task.VideoID, req.ReviewDecision); err != nil {
		log.Printf("Error handling queue flow: %v", err)
	}
	}

	// Remove from Redis
	userClaimedKey := fmt.Sprintf("video:claimed:%d:%s", reviewerID, pool)
	lockKey := fmt.Sprintf("video:lock:%d", req.TaskID)

	pipe := s.rdb.Pipeline()
	pipe.SRem(s.ctx, userClaimedKey, req.TaskID)
	pipe.Del(s.ctx, lockKey)
	_, err = pipe.Exec(s.ctx)

	if err != nil {
		log.Printf("Redis error when submitting video queue review: %v", err)
	}

	// Update statistics in Redis
	if createdResult {
		s.updateQueueStats(pool, result)
	}

	return nil
}

// SubmitBatchReviews submits multiple reviews at once
func (s *VideoQueueService) SubmitBatchReviews(pool string, reviewerID int, reviews []models.SubmitVideoQueueReviewRequest) error {
	var failed []string
	for _, review := range reviews {
		if err := s.SubmitReview(pool, reviewerID, review); err != nil {
			failed = append(failed, fmt.Sprintf("task %d: %v", review.TaskID, err))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to submit %d video queue reviews: %s", len(failed), strings.Join(failed, "; "))
	}
	return nil
}

// ReturnTasks allows a reviewer to return tasks back to the pool
func (s *VideoQueueService) ReturnTasks(pool string, reviewerID int, taskIDs []int) (int, error) {
	if !isValidPool(pool) {
		return 0, errors.New("invalid pool: must be 100k, 1m, or 10m")
	}

	// Validate task count (1-50)
	if len(taskIDs) < 1 || len(taskIDs) > 50 {
		return 0, errors.New("return count must be between 1 and 50")
	}

	// Return tasks in database
	returnedCount, err := s.queueRepo.ReturnQueueTasks(taskIDs, reviewerID)
	if err != nil {
		return 0, err
	}

	if returnedCount == 0 {
		return 0, errors.New("no tasks were returned, please check if tasks belong to you")
	}

	// Clean up Redis
	userClaimedKey := fmt.Sprintf("video:claimed:%d:%s", reviewerID, pool)
	pipe := s.rdb.Pipeline()

	for _, taskID := range taskIDs {
		// Remove from user's claimed set
		pipe.SRem(s.ctx, userClaimedKey, taskID)

		// Remove task lock
		lockKey := fmt.Sprintf("video:lock:%d", taskID)
		pipe.Del(s.ctx, lockKey)
	}

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when returning video queue tasks: %v", err)
	}

	return returnedCount, nil
}

// ReleaseExpiredTasks releases tasks that have exceeded the timeout for a specific pool
func (s *VideoQueueService) ReleaseExpiredTasks(pool string) error {
	if !isValidPool(pool) {
		return errors.New("invalid pool: must be 100k, 1m, or 10m")
	}

	timeoutMinutes := config.AppConfig.TaskTimeoutMinutes
	expiredTasks, err := s.queueRepo.FindExpiredQueueTasks(pool, timeoutMinutes)
	if err != nil {
		return err
	}

	for _, task := range expiredTasks {
		// Reset task in database
		if err := s.queueRepo.ResetQueueTask(task.ID); err != nil {
			log.Printf("Error resetting video queue task %d: %v", task.ID, err)
			continue
		}

		// Clean up Redis
		if task.ReviewerID != nil {
			userClaimedKey := fmt.Sprintf("video:claimed:%d:%s", *task.ReviewerID, pool)
			lockKey := fmt.Sprintf("video:lock:%d", task.ID)

			s.rdb.SRem(s.ctx, userClaimedKey, task.ID)
			s.rdb.Del(s.ctx, lockKey)
		}

		log.Printf("Released expired video queue task %d from pool %s", task.ID, pool)
	}

	return nil
}

// ReleaseAllExpiredTasks releases expired tasks from all pools
func (s *VideoQueueService) ReleaseAllExpiredTasks() error {
	pools := []string{"100k", "1m", "10m"}
	for _, pool := range pools {
		if err := s.ReleaseExpiredTasks(pool); err != nil {
			log.Printf("Error releasing expired tasks for pool %s: %v", pool, err)
		}
	}
	return nil
}

// handleQueueFlow handles the queue flow based on review decision
func (s *VideoQueueService) handleQueueFlow(currentPool string, videoID int, decision string) error {
	switch decision {
	case "push_next_pool":
		// Push to next pool
		nextPool := getNextPool(currentPool)
		if nextPool == "" {
			// Already at top pool (10m), mark as confirmed
			log.Printf("Video %d confirmed for 10m pool (top tier)", videoID)
			return s.queueRepo.UpdateVideoStatus(videoID, "10m_confirmed")
		}

		// Create task in next pool
		createdTask, err := s.queueRepo.CreateQueueTask(videoID, nextPool)
		if err != nil {
			return fmt.Errorf("failed to create task in %s pool: %w", nextPool, err)
		}

		if createdTask {
			// Push to Redis queue for next pool
			queueKey := fmt.Sprintf("video:queue:%s", nextPool)
			if err := s.rdb.LPush(s.ctx, queueKey, videoID).Err(); err != nil {
				log.Printf("Redis error pushing to %s queue: %v", nextPool, err)
			}
		}

		log.Printf("Video %d promoted from %s to %s pool", videoID, currentPool, nextPool)
		return nil

	case "natural_pool":
		// Stop queue flow, keep in natural pool
		log.Printf("Video %d assigned to natural pool (no further promotion)", videoID)
		return s.queueRepo.UpdateVideoStatus(videoID, "natural_pool")

	case "remove_violation":
		// Mark as removed due to violation
		log.Printf("Video %d removed due to violation", videoID)
		return s.queueRepo.UpdateVideoStatus(videoID, "removed_violation")

	default:
		return fmt.Errorf("invalid review decision: %s", decision)
	}
}

// GetTags retrieves retrieves available tags for a specific pool
func (s *VideoQueueService) GetTags(pool string) ([]models.VideoQueueTag, error) {
	if !isValidPool(pool) {
		return nil, errors.New("invalid pool: must be 100k, 1m, or 10m")
	}

	return s.queueRepo.GetVideoQueueTags(pool)
}

// GetPoolStats retrieves statistics for a specific pool
func (s *VideoQueueService) GetPoolStats(pool string) (*models.VideoQueuePoolStats, error) {
	if !isValidPool(pool) {
		return nil, errors.New("invalid pool: must be 100k, 1m, or 10m")
	}

	return s.queueRepo.GetQueuePoolStats(pool)
}

// updateQueueStats updates statistics in Redis
func (s *VideoQueueService) updateQueueStats(pool string, result *models.VideoQueueResult) {
	now := time.Now()
	date := now.Format("2006-01-02")
	hour := now.Hour()

	hourlyKey := fmt.Sprintf("video:stats:queue:%s:%s:%d", pool, date, hour)
	dailyKey := fmt.Sprintf("video:stats:queue:%s:%s", pool, date)

	pipe := s.rdb.Pipeline()
	pipe.HIncrBy(s.ctx, hourlyKey, "count", 1)
	pipe.Expire(s.ctx, hourlyKey, 7*24*time.Hour) // 7 days TTL

	pipe.HIncrBy(s.ctx, dailyKey, "count", 1)
	pipe.HIncrBy(s.ctx, dailyKey, fmt.Sprintf("decision:%s", result.ReviewDecision), 1)

	pipe.Expire(s.ctx, dailyKey, 30*24*time.Hour) // 30 days TTL

	_, err := pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when updating queue stats: %v", err)
	}
}

// Helper functions

func isValidPool(pool string) bool {
	return pool == "100k" || pool == "1m" || pool == "10m"
}

func getNextPool(currentPool string) string {
	switch currentPool {
	case "100k":
		return "1m"
	case "1m":
		return "10m"
	case "10m":
		return "" // Top pool, no next pool
	default:
		return ""
	}
}
