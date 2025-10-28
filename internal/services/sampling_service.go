package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"comment-review-platform/pkg/database"
	redispkg "comment-review-platform/pkg/redis"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

type SamplingService struct {
	qcRepo *repository.QualityCheckRepository
	rdb    *redis.Client
	ctx    context.Context
	db     *sql.DB
}

func NewSamplingService() *SamplingService {
	return &SamplingService{
		qcRepo: repository.NewQualityCheckRepository(),
		rdb:    redispkg.Client,
		ctx:    context.Background(),
		db:     database.DB,
	}
}

// DailySamplingTask performs daily sampling of first review results
func (s *SamplingService) DailySamplingTask() error {
	// Get yesterday's date
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	log.Printf("Starting daily sampling for date: %s", yesterday)

	// Get all unchecked review results from yesterday
	reviewResults, err := s.qcRepo.GetUncheckedReviewResults(yesterday)
	if err != nil {
		return fmt.Errorf("failed to get unchecked review results: %v", err)
	}

	if len(reviewResults) == 0 {
		log.Printf("No unchecked review results found for %s", yesterday)
		return nil
	}

	log.Printf("Found %d unchecked review results for %s", len(reviewResults), yesterday)

	// Separate approved and rejected results
	var approvedResults []models.ReviewResult
	var rejectedResults []models.ReviewResult

	for _, result := range reviewResults {
		if result.IsApproved {
			approvedResults = append(approvedResults, result)
		} else {
			rejectedResults = append(rejectedResults, result)
		}
	}

	log.Printf("Approved results: %d, Rejected results: %d", len(approvedResults), len(rejectedResults))

	// Calculate sampling counts
	approvedSampleCount := int(float64(len(approvedResults)) * 0.20) // 20% of approved
	rejectedSampleCount := int(float64(len(rejectedResults)) * 0.50) // 50% of rejected

	// Ensure we don't exceed 3000 total
	totalSampleCount := approvedSampleCount + rejectedSampleCount
	if totalSampleCount > 3000 {
		// Proportionally reduce both counts
		ratio := float64(3000) / float64(totalSampleCount)
		approvedSampleCount = int(float64(approvedSampleCount) * ratio)
		rejectedSampleCount = int(float64(rejectedSampleCount) * ratio)
		totalSampleCount = approvedSampleCount + rejectedSampleCount
	}

	log.Printf("Sampling counts - Approved: %d, Rejected: %d, Total: %d",
		approvedSampleCount, rejectedSampleCount, totalSampleCount)

	// Perform random sampling
	var sampledResults []models.ReviewResult

	// Sample approved results
	if approvedSampleCount > 0 && len(approvedResults) > 0 {
		sampledApproved := s.randomSample(approvedResults, approvedSampleCount)
		sampledResults = append(sampledResults, sampledApproved...)
	}

	// Sample rejected results
	if rejectedSampleCount > 0 && len(rejectedResults) > 0 {
		sampledRejected := s.randomSample(rejectedResults, rejectedSampleCount)
		sampledResults = append(sampledResults, sampledRejected...)
	}

	log.Printf("Sampled %d results for quality check", len(sampledResults))

	// Create quality check tasks
	var resultIDs []int
	for _, result := range sampledResults {
		// Get comment ID from the task
		commentID, err := s.getCommentIDFromTaskID(result.TaskID)
		if err != nil {
			log.Printf("Failed to get comment ID for task %d: %v", result.TaskID, err)
			continue
		}

		// Create quality check task
		err = s.qcRepo.CreateQCTask(result.ID, commentID)
		if err != nil {
			log.Printf("Failed to create QC task for result %d: %v", result.ID, err)
			continue
		}

		resultIDs = append(resultIDs, result.ID)

		// Push to Redis queue
		queueKey := "review:queue:quality_check"
		err = s.rdb.LPush(s.ctx, queueKey, commentID).Err()
		if err != nil {
			log.Printf("Redis error pushing to QC queue: %v", err)
		}
	}

	// Update quality_checked flag for sampled results
	if len(resultIDs) > 0 {
		err = s.qcRepo.UpdateReviewResultQCFlag(resultIDs)
		if err != nil {
			log.Printf("Failed to update quality_checked flag: %v", err)
		}
	}

	log.Printf("Daily sampling completed. Created %d QC tasks", len(resultIDs))
	return nil
}

// randomSample performs random sampling from a slice
func (s *SamplingService) randomSample(results []models.ReviewResult, count int) []models.ReviewResult {
	if count >= len(results) {
		return results
	}

	// Create a copy to avoid modifying the original slice
	copyResults := make([]models.ReviewResult, len(results))
	copy(copyResults, results)

	// Shuffle the slice
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(copyResults), func(i, j int) {
		copyResults[i], copyResults[j] = copyResults[j], copyResults[i]
	})

	// Return the first 'count' elements
	return copyResults[:count]
}

// getCommentIDFromTaskID gets the comment ID from a task ID
func (s *SamplingService) getCommentIDFromTaskID(taskID int) (int64, error) {
	query := `SELECT comment_id FROM review_tasks WHERE id = $1`
	var commentID int64
	err := s.db.QueryRow(query, taskID).Scan(&commentID)
	return commentID, err
}

// StartDailySamplingScheduler starts the daily sampling scheduler
func (s *SamplingService) StartDailySamplingScheduler() {
	log.Println("Starting daily sampling scheduler...")

	// Calculate time until next midnight
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	duration := nextMidnight.Sub(now)

	log.Printf("Next sampling will run at: %s (in %v)", nextMidnight.Format("2006-01-02 15:04:05"), duration)

	// Wait until midnight
	time.Sleep(duration)

	// Run the first sampling
	if err := s.DailySamplingTask(); err != nil {
		log.Printf("Error in daily sampling task: %v", err)
	}

	// Set up ticker for every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.DailySamplingTask(); err != nil {
			log.Printf("Error in daily sampling task: %v", err)
		}
	}
}
