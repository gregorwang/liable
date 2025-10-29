package services

import (
	"comment-review-platform/internal/config"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"comment-review-platform/pkg/r2"
	redispkg "comment-review-platform/pkg/redis"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type VideoService struct {
	videoRepo *repository.VideoRepository
	r2Service *r2.R2Service
}

func NewVideoService() (*VideoService, error) {
	r2Service, err := r2.NewR2Service()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize R2 service: %w", err)
	}

	return &VideoService{
		videoRepo: repository.NewVideoRepository(),
		r2Service: r2Service,
	}, nil
}

// ImportVideosFromR2 imports videos from R2 bucket path
func (s *VideoService) ImportVideosFromR2(r2PathPrefix string) (*models.ImportVideosResponse, error) {
	// Test R2 connection first
	if err := s.r2Service.CheckConnection(); err != nil {
		return nil, fmt.Errorf("R2 connection failed: %w", err)
	}

	// List videos in the specified path
	videos, err := s.r2Service.ListVideosInPath(r2PathPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list videos from R2: %w", err)
	}

	response := &models.ImportVideosResponse{
		ImportedCount: 0,
		SkippedCount:  0,
		Errors:        []string{},
	}

	for _, video := range videos {
		// Check if video already exists
		exists, err := s.videoRepo.CheckVideoExists(video.Key)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error checking video %s: %v", video.Filename, err))
			continue
		}

		if exists {
			response.SkippedCount++
			continue
		}

		// Get estimated duration
		duration, err := s.r2Service.GetVideoDuration(video.Key)
		if err != nil {
			log.Printf("Warning: Could not get duration for video %s: %v", video.Filename, err)
			duration = 30 // default duration
		}

		// Create video record
		tiktokVideo := &models.TikTokVideo{
			VideoKey:   video.Key,
			Filename:   video.Filename,
			FileSize:   video.Size,
			Duration:   &duration,
			UploadTime: &video.Modified,
			Status:     "pending",
		}

		if err := s.videoRepo.CreateVideo(tiktokVideo); err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error creating video %s: %v", video.Filename, err))
			continue
		}

		// Create first review task
		if err := s.CreateFirstReviewTask(tiktokVideo.ID); err != nil {
			log.Printf("Warning: Could not create first review task for video %d: %v", tiktokVideo.ID, err)
		}

		response.ImportedCount++
	}

	log.Printf("Video import completed: %d imported, %d skipped, %d errors",
		response.ImportedCount, response.SkippedCount, len(response.Errors))

	return response, nil
}

// CreateFirstReviewTask creates a first review task for a video
func (s *VideoService) CreateFirstReviewTask(videoID int) error {
	firstReviewRepo := repository.NewVideoFirstReviewRepository()
	return firstReviewRepo.CreateFirstReviewTask(videoID)
}

// GenerateVideoURL generates a pre-signed URL for video access
func (s *VideoService) GenerateVideoURL(videoID int) (*models.GenerateVideoURLResponse, error) {
	// Get video from database
	video, err := s.videoRepo.GetVideoByID(videoID)
	if err != nil {
		return nil, fmt.Errorf("video not found: %w", err)
	}

	// Check if we have a valid cached URL
	if video.VideoURL != nil && video.URLExpiresAt != nil && video.URLExpiresAt.After(time.Now()) {
		return &models.GenerateVideoURLResponse{
			VideoURL:  *video.VideoURL,
			ExpiresAt: *video.URLExpiresAt,
		}, nil
	}

	// Generate new pre-signed URL
	expiration := 1 * time.Hour // URLs expire in 1 hour
	videoURL, err := s.r2Service.GeneratePresignedURL(video.VideoKey, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	expiresAt := time.Now().Add(expiration)

	// Update video record with new URL
	if err := s.videoRepo.UpdateVideoURL(videoID, videoURL, expiresAt); err != nil {
		log.Printf("Warning: Could not update video URL in database: %v", err)
	}

	return &models.GenerateVideoURLResponse{
		VideoURL:  videoURL,
		ExpiresAt: expiresAt,
	}, nil
}

// GetVideoByID retrieves a video by ID
func (s *VideoService) GetVideoByID(id int) (*models.TikTokVideo, error) {
	return s.videoRepo.GetVideoByID(id)
}

// ListVideos returns paginated videos with filters
func (s *VideoService) ListVideos(req models.ListVideosRequest) ([]models.TikTokVideo, int, error) {
	return s.videoRepo.ListVideos(req)
}

// GetVideoQualityTags retrieves quality tags by category
func (s *VideoService) GetVideoQualityTags(category string) ([]models.VideoQualityTag, error) {
	return s.videoRepo.GetVideoQualityTags(category)
}

type VideoFirstReviewService struct {
	firstReviewRepo  *repository.VideoFirstReviewRepository
	secondReviewRepo *repository.VideoSecondReviewRepository
	videoRepo        *repository.VideoRepository
	rdb              *redis.Client
	ctx              context.Context
}

func NewVideoFirstReviewService() *VideoFirstReviewService {
	return &VideoFirstReviewService{
		firstReviewRepo:  repository.NewVideoFirstReviewRepository(),
		secondReviewRepo: repository.NewVideoSecondReviewRepository(),
		videoRepo:        repository.NewVideoRepository(),
		rdb:              redispkg.Client,
		ctx:              context.Background(),
	}
}

// ClaimFirstReviewTasks allows a reviewer to claim first review tasks
func (s *VideoFirstReviewService) ClaimFirstReviewTasks(reviewerID int, count int) ([]models.VideoFirstReviewTask, error) {
	// Validate count (1-50)
	if count < 1 || count > 50 {
		return nil, errors.New("claim count must be between 1 and 50")
	}

	// Check if user already has uncompleted tasks
	existingTasks, err := s.firstReviewRepo.GetMyFirstReviewTasks(reviewerID)
	if err != nil {
		return nil, err
	}

	if len(existingTasks) > 0 {
		return nil, fmt.Errorf("you still have %d uncompleted video review tasks, please complete or return them first", len(existingTasks))
	}

	// Claim tasks from database
	tasks, err := s.firstReviewRepo.ClaimFirstReviewTasks(reviewerID, count)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return []models.VideoFirstReviewTask{}, nil
	}

	// Add tasks to Redis for tracking
	userClaimedKey := fmt.Sprintf("video:first:claimed:%d", reviewerID)
	timeout := time.Duration(config.AppConfig.TaskTimeoutMinutes) * time.Minute

	pipe := s.rdb.Pipeline()
	for _, task := range tasks {
		// Add to user's claimed set
		pipe.SAdd(s.ctx, userClaimedKey, task.ID)

		// Set lock for each task
		lockKey := fmt.Sprintf("video:first:lock:%d", task.ID)
		pipe.Set(s.ctx, lockKey, reviewerID, timeout)
	}
	pipe.Expire(s.ctx, userClaimedKey, timeout)

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when claiming video first review tasks: %v", err)
	}

	return tasks, nil
}

// GetMyFirstReviewTasks retrieves the current user's in-progress first review tasks
func (s *VideoFirstReviewService) GetMyFirstReviewTasks(reviewerID int) ([]models.VideoFirstReviewTask, error) {
	return s.firstReviewRepo.GetMyFirstReviewTasks(reviewerID)
}

// SubmitFirstReview submits a first review result
func (s *VideoFirstReviewService) SubmitFirstReview(reviewerID int, req models.SubmitVideoFirstReviewRequest) error {
	// Complete the task
	if err := s.firstReviewRepo.CompleteFirstReviewTask(req.TaskID, reviewerID); err != nil {
		return errors.New("task not found or already completed")
	}

	// Calculate overall score
	overallScore := req.QualityDimensions.ContentQuality.Score +
		req.QualityDimensions.TechnicalQuality.Score +
		req.QualityDimensions.Compliance.Score +
		req.QualityDimensions.EngagementPotential.Score

	// Create review result
	result := &models.VideoFirstReviewResult{
		TaskID:            req.TaskID,
		ReviewerID:        reviewerID,
		IsApproved:        req.IsApproved,
		QualityDimensions: req.QualityDimensions,
		OverallScore:      overallScore,
		TrafficPoolResult: req.TrafficPoolResult,
		Reason:            req.Reason,
	}

	if err := s.firstReviewRepo.CreateFirstReviewResult(result); err != nil {
		return err
	}

	// If first review is not approved, create second review task
	if !req.IsApproved {
		// Get the video ID from the task
		tasks, err := s.firstReviewRepo.FindFirstReviewTasksWithVideos([]int{req.TaskID})
		if err != nil {
			log.Printf("Error getting video ID for second review task: %v", err)
		} else if len(tasks) > 0 {
			videoID := tasks[0].VideoID

			// Create second review task
			if err := s.secondReviewRepo.CreateSecondReviewTask(result.ID, videoID); err != nil {
				log.Printf("Error creating second review task: %v", err)
			} else {
				// Push to Redis second review queue
				queueKey := "video:review:queue:second"
				if err := s.rdb.LPush(s.ctx, queueKey, videoID).Err(); err != nil {
					log.Printf("Redis error pushing to second review queue: %v", err)
				}
			}
		}
	} else {
		// If approved, update video status
		tasks, err := s.firstReviewRepo.FindFirstReviewTasksWithVideos([]int{req.TaskID})
		if err == nil && len(tasks) > 0 {
			videoID := tasks[0].VideoID
			if err := s.videoRepo.UpdateVideoStatus(videoID, "first_review_completed"); err != nil {
				log.Printf("Error updating video status: %v", err)
			}
		}
	}

	// Remove from Redis
	userClaimedKey := fmt.Sprintf("video:first:claimed:%d", reviewerID)
	lockKey := fmt.Sprintf("video:first:lock:%d", req.TaskID)

	pipe := s.rdb.Pipeline()
	pipe.SRem(s.ctx, userClaimedKey, req.TaskID)
	pipe.Del(s.ctx, lockKey)
	_, err := pipe.Exec(s.ctx)

	if err != nil {
		log.Printf("Redis error when submitting video first review: %v", err)
	}

	// Update statistics in Redis
	s.updateVideoStats(result)

	return nil
}

// SubmitBatchFirstReviews submits multiple first reviews at once
func (s *VideoFirstReviewService) SubmitBatchFirstReviews(reviewerID int, reviews []models.SubmitVideoFirstReviewRequest) error {
	for _, review := range reviews {
		if err := s.SubmitFirstReview(reviewerID, review); err != nil {
			return err
		}
	}
	return nil
}

// ReturnFirstReviewTasks allows a reviewer to return first review tasks back to the pool
func (s *VideoFirstReviewService) ReturnFirstReviewTasks(reviewerID int, taskIDs []int) (int, error) {
	// Validate task count (1-50)
	if len(taskIDs) < 1 || len(taskIDs) > 50 {
		return 0, errors.New("return count must be between 1 and 50")
	}

	// Return tasks in database
	returnedCount, err := s.firstReviewRepo.ReturnFirstReviewTasks(taskIDs, reviewerID)
	if err != nil {
		return 0, err
	}

	if returnedCount == 0 {
		return 0, errors.New("no tasks were returned, please check if the tasks belong to you")
	}

	// Clean up Redis
	userClaimedKey := fmt.Sprintf("video:first:claimed:%d", reviewerID)
	pipe := s.rdb.Pipeline()

	for _, taskID := range taskIDs {
		// Remove from user's claimed set
		pipe.SRem(s.ctx, userClaimedKey, taskID)

		// Remove task lock
		lockKey := fmt.Sprintf("video:first:lock:%d", taskID)
		pipe.Del(s.ctx, lockKey)
	}

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Redis error when returning video first review tasks: %v", err)
	}

	return returnedCount, nil
}

// ReleaseExpiredFirstReviewTasks releases first review tasks that have exceeded the timeout
func (s *VideoFirstReviewService) ReleaseExpiredFirstReviewTasks() error {
	timeoutMinutes := config.AppConfig.TaskTimeoutMinutes
	expiredTasks, err := s.firstReviewRepo.FindExpiredFirstReviewTasks(timeoutMinutes)
	if err != nil {
		return err
	}

	for _, task := range expiredTasks {
		// Reset task in database
		if err := s.firstReviewRepo.ResetFirstReviewTask(task.ID); err != nil {
			log.Printf("Error resetting video first review task %d: %v", task.ID, err)
			continue
		}

		// Clean up Redis
		if task.ReviewerID != nil {
			userClaimedKey := fmt.Sprintf("video:first:claimed:%d", *task.ReviewerID)
			lockKey := fmt.Sprintf("video:first:lock:%d", task.ID)

			s.rdb.SRem(s.ctx, userClaimedKey, task.ID)
			s.rdb.Del(s.ctx, lockKey)
		}

		log.Printf("Released expired video first review task %d", task.ID)
	}

	return nil
}

// updateVideoStats updates statistics in Redis
func (s *VideoFirstReviewService) updateVideoStats(result *models.VideoFirstReviewResult) {
	now := time.Now()
	date := now.Format("2006-01-02")
	hour := now.Hour()

	hourlyKey := fmt.Sprintf("video:stats:hourly:%s:%d", date, hour)
	dailyKey := fmt.Sprintf("video:stats:daily:%s", date)

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
		log.Printf("Redis error when updating video stats: %v", err)
	}
}
