package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"comment-review-platform/internal/services/base"
	"comment-review-platform/pkg/r2"
	redispkg "comment-review-platform/pkg/redis"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type VideoService struct {
	videoRepo *repository.VideoRepository
	r2Service *r2.R2Service
	rdb       *redis.Client
	ctx       context.Context
}

func NewVideoService() (*VideoService, error) {
	r2Service, err := r2.NewR2Service()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize R2 service: %w", err)
	}

	return &VideoService{
		videoRepo: repository.NewVideoRepository(),
		r2Service: r2Service,
		rdb:       redispkg.Client,
		ctx:       context.Background(),
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

// GenerateVideoURL generates a pre-signed URL for video access with Redis caching
func (s *VideoService) GenerateVideoURL(videoID int) (*models.GenerateVideoURLResponse, error) {
	cacheKey := fmt.Sprintf("video:url:%d", videoID)

	// 1. Try to get from Redis cache
	videoURL, err := s.rdb.HGet(s.ctx, cacheKey, "url").Result()
	if err == nil {
		expiresAtStr, err2 := s.rdb.HGet(s.ctx, cacheKey, "expires_at").Result()
		if err2 == nil {
			// Check if cached URL is still valid
			expiresAt, err3 := time.Parse(time.RFC3339, expiresAtStr)
			if err3 == nil && expiresAt.After(time.Now()) {
				return &models.GenerateVideoURLResponse{
					VideoURL:  videoURL,
					ExpiresAt: expiresAt,
				}, nil
			}
		}
	}

	// 2. Cache miss or expired, get video from database
	video, err := s.videoRepo.GetVideoByID(videoID)
	if err != nil {
		return nil, fmt.Errorf("video not found: %w", err)
	}

	// Check database cached URL
	if video.VideoURL != nil && video.URLExpiresAt != nil && video.URLExpiresAt.After(time.Now()) {
		// Update Redis cache
		s.cacheVideoURL(videoID, *video.VideoURL, *video.URLExpiresAt)
		return &models.GenerateVideoURLResponse{
			VideoURL:  *video.VideoURL,
			ExpiresAt: *video.URLExpiresAt,
		}, nil
	}

	// 3. Generate new pre-signed URL
	expiration := 1 * time.Hour // URLs expire in 1 hour
	newVideoURL, err := s.r2Service.GeneratePresignedURL(video.VideoKey, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	expiresAt := time.Now().Add(expiration)

	// Update video record with new URL
	if err := s.videoRepo.UpdateVideoURL(videoID, newVideoURL, expiresAt); err != nil {
		log.Printf("Warning: Could not update video videoURL in database: %v", err)
	}

	// Update Redis cache
	s.cacheVideoURL(videoID, newVideoURL, expiresAt)

	return &models.GenerateVideoURLResponse{
		VideoURL:  newVideoURL,
		ExpiresAt: expiresAt,
	}, nil
}

// cacheVideoURL caches video URL in Redis asynchronously
func (s *VideoService) cacheVideoURL(videoID int, url string, expiresAt time.Time) {
	cacheKey := fmt.Sprintf("video:url:%d", videoID)
	pipe := s.rdb.Pipeline()
	pipe.HSet(s.ctx, cacheKey, "url", url)
	pipe.HSet(s.ctx, cacheKey, "expires_at", expiresAt.Format(time.RFC3339))
	pipe.Expire(s.ctx, cacheKey, 1*time.Hour) // Cache for 1 hour (same as URL expiration)

	// Execute asynchronously
	go func() {
		if _, err := pipe.Exec(s.ctx); err != nil {
			log.Printf("Warning: Failed to cache video URL %d: %v", videoID, err)
		}
	}()
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
	base             *base.BaseTaskService
}

func NewVideoFirstReviewService() *VideoFirstReviewService {
	return &VideoFirstReviewService{
		firstReviewRepo:  repository.NewVideoFirstReviewRepository(),
		secondReviewRepo: repository.NewVideoSecondReviewRepository(),
		videoRepo:        repository.NewVideoRepository(),
		base:             base.NewBaseTaskService(base.VideoFirstReviewTaskServiceConfig(), redispkg.Client),
	}
}

// ClaimFirstReviewTasks allows a reviewer to claim first review tasks
func (s *VideoFirstReviewService) ClaimFirstReviewTasks(reviewerID int, count int) ([]models.VideoFirstReviewTask, error) {
	// Validate count using base service
	if err := s.base.ValidateClaimCount(count); err != nil {
		return nil, err
	}

	// Check if user already has uncompleted tasks
	existingTasks, err := s.firstReviewRepo.GetMyFirstReviewTasks(reviewerID)
	if err != nil {
		return nil, err
	}

	if err := s.base.CheckExistingTasks(len(existingTasks)); err != nil {
		return nil, err
	}

	// Claim tasks from database
	tasks, err := s.firstReviewRepo.ClaimFirstReviewTasks(reviewerID, count)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return []models.VideoFirstReviewTask{}, nil
	}

	// Track tasks in Redis using base service
	taskIDs := make([]int, len(tasks))
	for i, task := range tasks {
		taskIDs[i] = task.ID
	}
	if err := s.base.TrackClaimedTasks(reviewerID, taskIDs); err != nil {
		log.Printf("Redis error when claiming video first review tasks: %v", err)
		if _, resetErr := s.firstReviewRepo.ReturnFirstReviewTasks(taskIDs, reviewerID); resetErr != nil {
			log.Printf("Failed to rollback video first review tasks after Redis error: %v", resetErr)
		}
		return nil, errors.New("failed to claim tasks, please retry")
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

	createdResult, err := s.firstReviewRepo.CreateFirstReviewResult(result)
	if err != nil {
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
			createdTask, err := s.secondReviewRepo.CreateSecondReviewTask(result.ID, videoID)
			if err != nil {
				log.Printf("Error creating second review task: %v", err)
			} else if createdTask {
				// Push to Redis second review queue
				queueKey := "video:review:queue:second"
				if s.base.Rdb != nil {
					if err := s.base.Rdb.LPush(s.base.Ctx, queueKey, videoID).Err(); err != nil {
						log.Printf("Redis error pushing to second review queue: %v", err)
					}
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

	// Cleanup Redis tracking using base service
	s.base.CleanupSingleTask(reviewerID, req.TaskID)

	// Update statistics in Redis
	if createdResult {
		s.updateVideoStats(result)
	}

	return nil
}

// SubmitBatchFirstReviews submits multiple first reviews at once
func (s *VideoFirstReviewService) SubmitBatchFirstReviews(reviewerID int, reviews []models.SubmitVideoFirstReviewRequest) error {
	var failed []string
	for _, review := range reviews {
		if err := s.SubmitFirstReview(reviewerID, review); err != nil {
			failed = append(failed, fmt.Sprintf("task %d: %v", review.TaskID, err))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to submit %d video first reviews: %s", len(failed), strings.Join(failed, "; "))
	}
	return nil
}

// ReturnFirstReviewTasks allows a reviewer to return first review tasks back to the pool
func (s *VideoFirstReviewService) ReturnFirstReviewTasks(reviewerID int, taskIDs []int) (int, error) {
	// Validate return count using base service
	if err := s.base.ValidateReturnCount(len(taskIDs)); err != nil {
		return 0, err
	}

	// Return tasks in database
	returnedCount, err := s.firstReviewRepo.ReturnFirstReviewTasks(taskIDs, reviewerID)
	if err != nil {
		return 0, err
	}

	if returnedCount == 0 {
		return 0, errors.New("no tasks were returned, please check if the tasks belong to you")
	}

	// Cleanup Redis tracking using base service
	s.base.CleanupTaskTracking(reviewerID, taskIDs)

	return returnedCount, nil
}

// ReleaseExpiredFirstReviewTasks releases first review tasks that have exceeded the timeout
func (s *VideoFirstReviewService) ReleaseExpiredFirstReviewTasks() error {
	timeoutMinutes := s.base.GetTaskTimeoutMinutes()
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

		// Clean up Redis using base service
		if task.ReviewerID != nil {
			s.base.CleanupSingleTask(*task.ReviewerID, task.ID)
		}

		log.Printf("Released expired video first review task %d", task.ID)
	}

	return nil
}

// updateVideoStats updates statistics in Redis
func (s *VideoFirstReviewService) updateVideoStats(result *models.VideoFirstReviewResult) {
	if s.base.Rdb == nil {
		return
	}

	now := time.Now()
	date := now.Format("2006-01-02")
	hour := now.Hour()

	hourlyKey := fmt.Sprintf("video:stats:hourly:%s:%d", date, hour)
	dailyKey := fmt.Sprintf("video:stats:daily:%s", date)

	pipe := s.base.Rdb.Pipeline()
	pipe.HIncrBy(s.base.Ctx, hourlyKey, "count", 1)
	pipe.Expire(s.base.Ctx, hourlyKey, 7*24*time.Hour) // 7 days TTL

	pipe.HIncrBy(s.base.Ctx, dailyKey, "count", 1)
	if result.IsApproved {
		pipe.HIncrBy(s.base.Ctx, dailyKey, "approved", 1)
	} else {
		pipe.HIncrBy(s.base.Ctx, dailyKey, "rejected", 1)
	}

	// Track overall score distribution
	scoreRange := "low"
	if result.OverallScore >= 30 {
		scoreRange = "high"
	} else if result.OverallScore >= 20 {
		scoreRange = "medium"
	}
	pipe.HIncrBy(s.base.Ctx, dailyKey, fmt.Sprintf("score:%s", scoreRange), 1)

	pipe.Expire(s.base.Ctx, dailyKey, 30*24*time.Hour) // 30 days TTL

	_, err := pipe.Exec(s.base.Ctx)
	if err != nil {
		log.Printf("Redis error when updating video stats: %v", err)
	}
}
