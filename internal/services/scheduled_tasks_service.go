package services

import (
	redispkg "comment-review-platform/pkg/redis"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type ScheduledTasksService struct {
	rdb *redis.Client
	db  *sql.DB
	ctx context.Context
}

func NewScheduledTasksService(db *sql.DB) *ScheduledTasksService {
	return &ScheduledTasksService{
		rdb: redispkg.Client,
		db:  db,
		ctx: context.Background(),
	}
}

// RunDailyAggregation runs daily aggregation for previous day
func (s *ScheduledTasksService) RunDailyAggregation() error {
	yesterday := time.Now().AddDate(0, 0, -1)
	date := yesterday.Format("2006-01-02")

	log.Printf("Starting daily aggregation for %s", date)

	// Aggregate comment review stats
	if err := s.AggregateCommentStats(date); err != nil {
		log.Printf("Error aggregating comment review stats: %v", err)
	}

	// Aggregate video quality stats
	if err := s.AggregateVideoStats(date); err != nil {
		log.Printf("Error aggregating video quality stats: %v", err)
	}

	// Cleanup old Redis stats
	if err := s.CleanupOldRedisStats(date); err != nil {
		log.Printf("Error cleaning up old Redis stats: %v", err)
	}

	log.Printf("Successfully completed daily aggregation for %s", date)
	return nil
}

// AggregateCommentStats aggregates daily comment review statistics from Redis to database
func (s *ScheduledTasksService) AggregateCommentStats(date string) error {
	dailyKey := fmt.Sprintf("stats:daily:%s", date)

	// Get daily statistics from Redis
	dailyStats, err := s.rdb.HGetAll(s.ctx, dailyKey).Result()
	if err != nil && err != redis.Nil {
		log.Printf("Error getting daily stats for %s: %v", date, err)
		return err
	}

	if len(dailyStats) == 0 {
		log.Printf("No daily stats found for %s", date)
		return nil
	}

	// Store in JSON format
	jsonData, err := json.Marshal(dailyStats)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO daily_review_stats (date, stats_json, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (date) DO UPDATE SET
			stats_json = $2,
			updated_at = NOW()
	`

	_, err = s.db.Exec(query, date, string(jsonData))
	return err
}

// AggregateVideoStats aggregates daily video quality statistics from Redis
func (s *ScheduledTasksService) AggregateVideoStats(date string) error {
	dailyKey := fmt.Sprintf("video:stats:daily:%s", date)

	// Get daily statistics from Redis
	dailyStats, err := s.rdb.HGetAll(s.ctx, dailyKey).Result()
	if err != nil && err != redis.Nil {
		log.Printf("Error getting video daily stats for %s: %v", date, err)
		return err
	}

	if len(dailyStats) == 0 {
		log.Printf("No video daily stats found for %s", date)
		return nil
	}

	// Store in JSON format
	jsonData, err := json.Marshal(dailyStats)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO daily_video_stats (date, stats_json, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (date) DO UPDATE SET
			stats_json = $2,
			updated_at = NOW()
	`

	_, err = s.db.Exec(query, date, string(jsonData))
	return err
}

// CleanupOldRedisStats removes old statistics from Redis (keep last 30 days)
func (s *ScheduledTasksService) CleanupOldRedisStats(currentDate string) error {
	cutoffDate := time.Now().AddDate(0, 0, -30)
	cutoff := cutoffDate.Format("2006-01-02")

	// Delete old daily stats
	dailyPattern := "stats:daily:*"
	iter := s.rdb.Scan(s.ctx, 0, dailyPattern, 0).Iterator()
	var deletedCount int

	for iter.Next(s.ctx) {
		key := iter.Val()
		parts := strings.Split(key, ":")
		if len(parts) >= 3 {
			keyDate := parts[2]
			if keyDate < cutoff {
				s.rdb.Del(s.ctx, key)
				deletedCount++
			}
		}
	}

	// Delete old video stats
	videoPattern := "video:stats:daily:*"
	videoIter := s.rdb.Scan(s.ctx, 0, videoPattern, 0).Iterator()

	for videoIter.Next(s.ctx) {
		key := videoIter.Val()
		parts := strings.Split(key, ":")
		if len(parts) >= 4 {
			keyDate := parts[3]
			if keyDate < cutoff {
				s.rdb.Del(s.ctx, key)
				deletedCount++
			}
		}
	}

	log.Printf("Cleaned up %d old Redis statistics entries (older than %s)", deletedCount, cutoff)
	return nil
}
