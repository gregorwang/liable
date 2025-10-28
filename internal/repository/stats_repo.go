package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"time"
)

type StatsRepository struct {
	db *sql.DB
}

func NewStatsRepository() *StatsRepository {
	return &StatsRepository{db: database.DB}
}

// GetOverviewStats returns overall statistics
func (r *StatsRepository) GetOverviewStats() (*models.StatsOverview, error) {
	stats := &models.StatsOverview{}

	// Get task counts
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress
		FROM review_tasks
	`
	err := r.db.QueryRow(query).Scan(&stats.TotalTasks, &stats.CompletedTasks, &stats.PendingTasks, &stats.InProgressTasks)
	if err != nil {
		return nil, err
	}

	// Get approval counts
	approvalQuery := `
		SELECT 
			COUNT(CASE WHEN is_approved = true THEN 1 END) as approved,
			COUNT(CASE WHEN is_approved = false THEN 1 END) as rejected
		FROM review_results
	`
	err = r.db.QueryRow(approvalQuery).Scan(&stats.ApprovedCount, &stats.RejectedCount)
	if err != nil {
		return nil, err
	}

	// Calculate approval rate
	if stats.CompletedTasks > 0 {
		stats.ApprovalRate = float64(stats.ApprovedCount) / float64(stats.CompletedTasks) * 100
	}

	// Get reviewer counts
	r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'reviewer' AND status = 'approved'`).Scan(&stats.TotalReviewers)
	r.db.QueryRow(`SELECT COUNT(DISTINCT reviewer_id) FROM review_tasks WHERE status = 'completed'`).Scan(&stats.ActiveReviewers)

	// Get queue statistics
	queueStats, err := r.getQueueStats()
	if err != nil {
		return nil, err
	}
	stats.QueueStats = queueStats

	// Get quality metrics
	qualityMetrics, err := r.getQualityMetrics()
	if err != nil {
		return nil, err
	}
	stats.QualityMetrics = *qualityMetrics

	return stats, nil
}

// GetHourlyStats returns hourly statistics for a specific date
func (r *StatsRepository) GetHourlyStats(date string) ([]models.HourlyStatItem, error) {
	query := `
		SELECT 
			EXTRACT(HOUR FROM created_at) as hour,
			COUNT(*) as count
		FROM review_results
		WHERE DATE(created_at) = $1
		GROUP BY hour
		ORDER BY hour
	`
	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []models.HourlyStatItem{}
	for rows.Next() {
		var stat models.HourlyStatItem
		if err := rows.Scan(&stat.Hour, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// GetTagStats returns violation tag statistics
func (r *StatsRepository) GetTagStats() ([]models.TagStats, error) {
	query := `
		SELECT 
			unnest(tags) as tag_name,
			COUNT(*) as count
		FROM review_results
		WHERE is_approved = false AND tags IS NOT NULL
		GROUP BY tag_name
		ORDER BY count DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []models.TagStats{}
	for rows.Next() {
		var stat models.TagStats
		if err := rows.Scan(&stat.TagName, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// GetReviewerPerformance returns reviewer performance statistics
func (r *StatsRepository) GetReviewerPerformance(limit int) ([]models.ReviewerPerformance, error) {
	query := `
		SELECT 
			u.id,
			u.username,
			COUNT(*) as total_reviews,
			COUNT(CASE WHEN rr.is_approved = true THEN 1 END) as approved_count,
			COUNT(CASE WHEN rr.is_approved = false THEN 1 END) as rejected_count,
			CASE 
				WHEN COUNT(*) > 0 THEN 
					ROUND(COUNT(CASE WHEN rr.is_approved = true THEN 1 END)::numeric / COUNT(*)::numeric * 100, 2)
				ELSE 0
			END as approval_rate
		FROM users u
		INNER JOIN review_results rr ON u.id = rr.reviewer_id
		WHERE u.role = 'reviewer' AND u.status = 'approved'
		GROUP BY u.id, u.username
		ORDER BY total_reviews DESC
		LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	performances := []models.ReviewerPerformance{}
	for rows.Next() {
		var perf models.ReviewerPerformance
		if err := rows.Scan(&perf.ReviewerID, &perf.Username, &perf.TotalReviews,
			&perf.ApprovedCount, &perf.RejectedCount, &perf.ApprovalRate); err != nil {
			return nil, err
		}
		performances = append(performances, perf)
	}
	return performances, nil
}

// GetDailyReviewCount returns the count of reviews for a specific date range
func (r *StatsRepository) GetDailyReviewCount(startDate, endDate time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM review_results
		WHERE created_at >= $1 AND created_at < $2
	`
	var count int
	err := r.db.QueryRow(query, startDate, endDate).Scan(&count)
	return count, err
}

// getQueueStats returns statistics for each queue using the real-time view
func (r *StatsRepository) getQueueStats() ([]models.QueueStats, error) {
	query := `
		SELECT 
			qs.queue_name,
			qs.total_tasks,
			qs.completed_tasks,
			qs.pending_tasks,
			qs.is_active,
			COALESCE(stats.approved_count, 0) as approved_count,
			COALESCE(stats.rejected_count, 0) as rejected_count,
			CASE 
				WHEN qs.completed_tasks > 0 THEN 
					ROUND(COALESCE(stats.approved_count, 0)::numeric / qs.completed_tasks::numeric * 100, 2)
				ELSE 0
			END as approval_rate,
			COALESCE(stats.avg_process_time, 0) as avg_process_time
		FROM queue_stats qs
		LEFT JOIN (
			SELECT 
				COUNT(CASE WHEN rr.is_approved = true THEN 1 END) as approved_count,
				COUNT(CASE WHEN rr.is_approved = false THEN 1 END) as rejected_count,
				AVG(EXTRACT(EPOCH FROM (rt.completed_at - rt.claimed_at))/60) as avg_process_time
			FROM review_tasks rt
			JOIN review_results rr ON rt.id = rr.task_id
			WHERE rt.status = 'completed' AND rt.completed_at IS NOT NULL AND rt.claimed_at IS NOT NULL
		) stats ON true
		ORDER BY qs.priority DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize as empty slice to avoid null in JSON response
	queueStats := make([]models.QueueStats, 0)
	for rows.Next() {
		var stat models.QueueStats
		err := rows.Scan(
			&stat.QueueName,
			&stat.TotalTasks,
			&stat.CompletedTasks,
			&stat.PendingTasks,
			&stat.IsActive,
			&stat.ApprovedCount,
			&stat.RejectedCount,
			&stat.ApprovalRate,
			&stat.AvgProcessTime,
		)
		if err != nil {
			return nil, err
		}
		queueStats = append(queueStats, stat)
	}

	return queueStats, nil
}

// getQualityMetrics returns quality check metrics
func (r *StatsRepository) getQualityMetrics() (*models.QualityMetrics, error) {
	metrics := &models.QualityMetrics{}

	// Get quality check statistics
	query := `
		SELECT 
			COUNT(*) as total_quality_checks,
			COUNT(CASE WHEN is_passed = true THEN 1 END) as passed_quality_checks,
			COUNT(CASE WHEN is_passed = false THEN 1 END) as failed_quality_checks
		FROM quality_check_results
	`
	err := r.db.QueryRow(query).Scan(
		&metrics.TotalQualityChecks,
		&metrics.PassedQualityChecks,
		&metrics.FailedQualityChecks,
	)
	if err != nil {
		return nil, err
	}

	// Calculate quality pass rate
	if metrics.TotalQualityChecks > 0 {
		metrics.QualityPassRate = float64(metrics.PassedQualityChecks) / float64(metrics.TotalQualityChecks) * 100
	}

	// Get second review statistics
	secondReviewQuery := `
		SELECT 
			COUNT(*) as second_review_tasks,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as second_review_completed
		FROM second_review_tasks
	`
	err = r.db.QueryRow(secondReviewQuery).Scan(
		&metrics.SecondReviewTasks,
		&metrics.SecondReviewCompleted,
	)
	if err != nil {
		return nil, err
	}

	// Calculate second review rate
	if metrics.SecondReviewTasks > 0 {
		metrics.SecondReviewRate = float64(metrics.SecondReviewCompleted) / float64(metrics.SecondReviewTasks) * 100
	}

	return metrics, nil
}
