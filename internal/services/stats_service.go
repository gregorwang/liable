package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	redispkg "comment-review-platform/pkg/redis"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type StatsService struct {
	statsRepo *repository.StatsRepository
	rdb       *redis.Client
	ctx       context.Context
}

func NewStatsService() *StatsService {
	return &StatsService{
		statsRepo: repository.NewStatsRepository(),
		rdb:       redispkg.Client,
		ctx:       context.Background(),
	}
}

// GetOverviewStats retrieves overall statistics with Redis caching
func (s *StatsService) GetOverviewStats() (*models.StatsOverview, error) {
	cacheKey := "stats:overview"
	cacheTTL := 5 * time.Minute // 5 minutes cache

	// 1. Try to get from Redis cache
	cached, err := s.rdb.Get(s.ctx, cacheKey).Result()
	if err == nil {
		var stats models.StatsOverview
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			// Cache hit, return cached data
			return &stats, nil
		}
	}

	// 2. Cache miss, query database
	stats, err := s.statsRepo.GetOverviewStats()
	if err != nil {
		return nil, err
	}

	// 3. Write to Redis cache asynchronously
	go func() {
		data, err := json.Marshal(stats)
		if err == nil {
			if err := s.rdb.Set(s.ctx, cacheKey, data, cacheTTL).Err(); err != nil {
				log.Printf("Warning: Failed to cache overview stats to Redis: %v", err)
			}
		}
	}()

	return stats, nil
}

// RefreshOverviewStats forces a cache refresh and returns updated statistics
func (s *StatsService) RefreshOverviewStats() (*models.StatsOverview, error) {
	cacheKey := "stats:overview"

	// Delete cache
	s.rdb.Del(s.ctx, cacheKey)

	// Query database
	return s.GetOverviewStats()
}

// GetTodayReviewStats retrieves today's review counts
func (s *StatsService) GetTodayReviewStats() (*models.TodayReviewStats, error) {
	return s.statsRepo.GetTodayReviewStats()
}

// GetHourlyStats retrieves hourly statistics for a specific date
func (s *StatsService) GetHourlyStats(date string) (*models.HourlyStats, error) {
	items, err := s.statsRepo.GetHourlyStats(date)
	if err != nil {
		return nil, err
	}

	return &models.HourlyStats{
		Date:  date,
		Hours: items,
	}, nil
}

// GetTagStats retrieves violation tag statistics
func (s *StatsService) GetTagStats() ([]models.TagStats, error) {
	return s.statsRepo.GetTagStats()
}

// GetReviewerPerformance retrieves reviewer performance performance
func (s *StatsService) GetReviewerPerformance(limit int) ([]models.ReviewerPerformance, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.statsRepo.GetReviewerPerformance(limit)
}

// GetVideoQualityTagStats retrieves video quality tag statistics
func (s *StatsService) GetVideoQualityTagStats() ([]models.VideoQualityTagStats, error) {
	return s.statsRepo.GetVideoQualityTagStats()
}

// GetVideoQualityAnalysis retrieves comprehensive video quality analysis
func (s *StatsService) GetVideoQualityAnalysis() (*models.VideoQualityAnalysis, error) {
	return s.statsRepo.GetVideoQualityAnalysis()
}
