package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
)

type StatsService struct {
	statsRepo *repository.StatsRepository
}

func NewStatsService() *StatsService {
	return &StatsService{
		statsRepo: repository.NewStatsRepository(),
	}
}

// GetOverviewStats retrieves overall statistics
func (s *StatsService) GetOverviewStats() (*models.StatsOverview, error) {
	return s.statsRepo.GetOverviewStats()
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

// GetReviewerPerformance retrieves reviewer performance statistics
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
