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

	query := `
		WITH
		review_tasks_stats AS (
			SELECT
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'completed') AS completed,
				COUNT(*) FILTER (WHERE status = 'pending') AS pending,
				COUNT(*) FILTER (WHERE status = 'in_progress') AS in_progress
			FROM review_tasks
		),
		review_results_stats AS (
			SELECT
				COUNT(*) FILTER (WHERE is_approved = true) AS approved,
				COUNT(*) FILTER (WHERE is_approved = false) AS rejected
			FROM review_results
		),
		second_review_tasks_stats AS (
			SELECT
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'completed') AS completed,
				COUNT(*) FILTER (WHERE status = 'pending') AS pending,
				COUNT(*) FILTER (WHERE status = 'in_progress') AS in_progress
			FROM second_review_tasks
		),
		second_review_results_stats AS (
			SELECT
				COUNT(*) FILTER (WHERE is_approved = true) AS approved,
				COUNT(*) FILTER (WHERE is_approved = false) AS rejected
			FROM second_review_results
		),
		video_first_tasks_stats AS (
			SELECT
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'completed') AS completed,
				COUNT(*) FILTER (WHERE status = 'pending') AS pending,
				COUNT(*) FILTER (WHERE status = 'in_progress') AS in_progress
			FROM video_first_review_tasks
		),
		video_first_results_stats AS (
			SELECT
				COUNT(*) FILTER (WHERE is_approved = true) AS approved,
				COUNT(*) FILTER (WHERE is_approved = false) AS rejected,
				COALESCE(AVG(overall_score), 0) AS avg_score
			FROM video_first_review_results
		),
		video_second_tasks_stats AS (
			SELECT
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'completed') AS completed,
				COUNT(*) FILTER (WHERE status = 'pending') AS pending,
				COUNT(*) FILTER (WHERE status = 'in_progress') AS in_progress
			FROM video_second_review_tasks
		),
		video_second_results_stats AS (
			SELECT
				COUNT(*) FILTER (WHERE is_approved = true) AS approved,
				COUNT(*) FILTER (WHERE is_approved = false) AS rejected,
				COALESCE(AVG(overall_score), 0) AS avg_score
			FROM video_second_review_results
		),
		quality_check_stats AS (
			SELECT
				COUNT(*) AS total_quality_checks,
				COUNT(*) FILTER (WHERE is_passed = true) AS passed_quality_checks,
				COUNT(*) FILTER (WHERE is_passed = false) AS failed_quality_checks
			FROM quality_check_results
		),
		reviewer_stats AS (
			SELECT
				(SELECT COUNT(*) FROM users WHERE role = 'reviewer' AND status = 'approved') AS total_reviewers,
				(SELECT COUNT(DISTINCT reviewer_id) FROM (
					SELECT reviewer_id FROM review_tasks WHERE status = 'completed' AND reviewer_id IS NOT NULL
					UNION
					SELECT reviewer_id FROM second_review_tasks WHERE status = 'completed' AND reviewer_id IS NOT NULL
					UNION
					SELECT reviewer_id FROM quality_check_tasks WHERE status = 'completed' AND reviewer_id IS NOT NULL
					UNION
					SELECT reviewer_id FROM ai_human_diff_tasks WHERE status = 'completed' AND reviewer_id IS NOT NULL
					UNION
					SELECT reviewer_id FROM video_first_review_tasks WHERE status = 'completed' AND reviewer_id IS NOT NULL
					UNION
					SELECT reviewer_id FROM video_second_review_tasks WHERE status = 'completed' AND reviewer_id IS NOT NULL
				) AS all_reviewers) AS active_reviewers
		)
		SELECT
			rts.total, rts.completed, rts.pending, rts.in_progress,
			rrs.approved, rrs.rejected,
			srt.total, srt.completed, srt.pending, srt.in_progress,
			srr.approved, srr.rejected,
			vft.total, vft.completed, vft.pending, vft.in_progress,
			vfr.approved, vfr.rejected, vfr.avg_score,
			vst.total, vst.completed, vst.pending, vst.in_progress,
			vsr.approved, vsr.rejected, vsr.avg_score,
			rs.total_reviewers, rs.active_reviewers,
			qcs.total_quality_checks, qcs.passed_quality_checks, qcs.failed_quality_checks,
			srt.total, srt.completed
		FROM review_tasks_stats rts
		CROSS JOIN review_results_stats rrs
		CROSS JOIN second_review_tasks_stats srt
		CROSS JOIN second_review_results_stats srr
		CROSS JOIN video_first_tasks_stats vft
		CROSS JOIN video_first_results_stats vfr
		CROSS JOIN video_second_tasks_stats vst
		CROSS JOIN video_second_results_stats vsr
		CROSS JOIN reviewer_stats rs
		CROSS JOIN quality_check_stats qcs
	`

	var (
		commentFirstTotal       int
		commentFirstCompleted   int
		commentFirstPending     int
		commentFirstInProgress  int
		commentFirstApproved    int
		commentFirstRejected    int
		commentSecondTotal      int
		commentSecondCompleted  int
		commentSecondPending    int
		commentSecondInProgress int
		commentSecondApproved   int
		commentSecondRejected   int
		videoFirstTotal         int
		videoFirstCompleted     int
		videoFirstPending       int
		videoFirstInProgress    int
		videoFirstApproved      int
		videoFirstRejected      int
		videoFirstAvgScore      float64
		videoSecondTotal        int
		videoSecondCompleted    int
		videoSecondPending      int
		videoSecondInProgress   int
		videoSecondApproved     int
		videoSecondRejected     int
		videoSecondAvgScore     float64
		totalReviewers          int
		activeReviewers         int
		totalQualityChecks      int
		passedQualityChecks     int
		failedQualityChecks     int
		secondReviewTasks       int
		secondReviewCompleted   int
	)

	err := r.db.QueryRow(query).Scan(
		&commentFirstTotal,
		&commentFirstCompleted,
		&commentFirstPending,
		&commentFirstInProgress,
		&commentFirstApproved,
		&commentFirstRejected,
		&commentSecondTotal,
		&commentSecondCompleted,
		&commentSecondPending,
		&commentSecondInProgress,
		&commentSecondApproved,
		&commentSecondRejected,
		&videoFirstTotal,
		&videoFirstCompleted,
		&videoFirstPending,
		&videoFirstInProgress,
		&videoFirstApproved,
		&videoFirstRejected,
		&videoFirstAvgScore,
		&videoSecondTotal,
		&videoSecondCompleted,
		&videoSecondPending,
		&videoSecondInProgress,
		&videoSecondApproved,
		&videoSecondRejected,
		&videoSecondAvgScore,
		&totalReviewers,
		&activeReviewers,
		&totalQualityChecks,
		&passedQualityChecks,
		&failedQualityChecks,
		&secondReviewTasks,
		&secondReviewCompleted,
	)
	if err != nil {
		return nil, err
	}

	stats.CommentReviewStats.FirstReview.TotalTasks = commentFirstTotal
	stats.CommentReviewStats.FirstReview.CompletedTasks = commentFirstCompleted
	stats.CommentReviewStats.FirstReview.PendingTasks = commentFirstPending
	stats.CommentReviewStats.FirstReview.InProgressTasks = commentFirstInProgress
	stats.CommentReviewStats.FirstReview.ApprovedCount = commentFirstApproved
	stats.CommentReviewStats.FirstReview.RejectedCount = commentFirstRejected

	stats.CommentReviewStats.SecondReview.TotalTasks = commentSecondTotal
	stats.CommentReviewStats.SecondReview.CompletedTasks = commentSecondCompleted
	stats.CommentReviewStats.SecondReview.PendingTasks = commentSecondPending
	stats.CommentReviewStats.SecondReview.InProgressTasks = commentSecondInProgress
	stats.CommentReviewStats.SecondReview.ApprovedCount = commentSecondApproved
	stats.CommentReviewStats.SecondReview.RejectedCount = commentSecondRejected

	stats.VideoReviewStats.FirstReview.TotalTasks = videoFirstTotal
	stats.VideoReviewStats.FirstReview.CompletedTasks = videoFirstCompleted
	stats.VideoReviewStats.FirstReview.PendingTasks = videoFirstPending
	stats.VideoReviewStats.FirstReview.InProgressTasks = videoFirstInProgress
	stats.VideoReviewStats.FirstReview.ApprovedCount = videoFirstApproved
	stats.VideoReviewStats.FirstReview.RejectedCount = videoFirstRejected
	stats.VideoReviewStats.FirstReview.AvgOverallScore = videoFirstAvgScore

	stats.VideoReviewStats.SecondReview.TotalTasks = videoSecondTotal
	stats.VideoReviewStats.SecondReview.CompletedTasks = videoSecondCompleted
	stats.VideoReviewStats.SecondReview.PendingTasks = videoSecondPending
	stats.VideoReviewStats.SecondReview.InProgressTasks = videoSecondInProgress
	stats.VideoReviewStats.SecondReview.ApprovedCount = videoSecondApproved
	stats.VideoReviewStats.SecondReview.RejectedCount = videoSecondRejected
	stats.VideoReviewStats.SecondReview.AvgOverallScore = videoSecondAvgScore

	stats.TotalReviewers = totalReviewers
	stats.ActiveReviewers = activeReviewers

	stats.QualityMetrics.TotalQualityChecks = totalQualityChecks
	stats.QualityMetrics.PassedQualityChecks = passedQualityChecks
	stats.QualityMetrics.FailedQualityChecks = failedQualityChecks
	stats.QualityMetrics.SecondReviewTasks = secondReviewTasks
	stats.QualityMetrics.SecondReviewCompleted = secondReviewCompleted

	if stats.CommentReviewStats.FirstReview.CompletedTasks > 0 {
		stats.CommentReviewStats.FirstReview.ApprovalRate =
			float64(stats.CommentReviewStats.FirstReview.ApprovedCount) /
				float64(stats.CommentReviewStats.FirstReview.CompletedTasks) * 100
	}
	if stats.CommentReviewStats.SecondReview.CompletedTasks > 0 {
		stats.CommentReviewStats.SecondReview.ApprovalRate =
			float64(stats.CommentReviewStats.SecondReview.ApprovedCount) /
				float64(stats.CommentReviewStats.SecondReview.CompletedTasks) * 100
	}
	if stats.VideoReviewStats.FirstReview.CompletedTasks > 0 {
		stats.VideoReviewStats.FirstReview.ApprovalRate =
			float64(stats.VideoReviewStats.FirstReview.ApprovedCount) /
				float64(stats.VideoReviewStats.FirstReview.CompletedTasks) * 100
	}
	if stats.VideoReviewStats.SecondReview.CompletedTasks > 0 {
		stats.VideoReviewStats.SecondReview.ApprovalRate =
			float64(stats.VideoReviewStats.SecondReview.ApprovedCount) /
				float64(stats.VideoReviewStats.SecondReview.CompletedTasks) * 100
	}
	if stats.QualityMetrics.TotalQualityChecks > 0 {
		stats.QualityMetrics.QualityPassRate =
			float64(stats.QualityMetrics.PassedQualityChecks) / float64(stats.QualityMetrics.TotalQualityChecks) * 100
	}
	if stats.QualityMetrics.SecondReviewTasks > 0 {
		stats.QualityMetrics.SecondReviewRate =
			float64(stats.QualityMetrics.SecondReviewCompleted) / float64(stats.QualityMetrics.SecondReviewTasks) * 100
	}

	stats.TotalTasks = stats.CommentReviewStats.FirstReview.TotalTasks
	stats.CompletedTasks = stats.CommentReviewStats.FirstReview.CompletedTasks
	stats.PendingTasks = stats.CommentReviewStats.FirstReview.PendingTasks
	stats.InProgressTasks = stats.CommentReviewStats.FirstReview.InProgressTasks
	stats.ApprovedCount = stats.CommentReviewStats.FirstReview.ApprovedCount
	stats.RejectedCount = stats.CommentReviewStats.FirstReview.RejectedCount
	stats.ApprovalRate = stats.CommentReviewStats.FirstReview.ApprovalRate

	queueStats, err := r.getQueueStats()
	if err != nil {
		return nil, err
	}
	stats.QueueStats = queueStats

	return stats, nil
}

// GetTodayReviewStats returns today's review counts for comment and video pipelines
func (r *StatsRepository) GetTodayReviewStats() (*models.TodayReviewStats, error) {
	stats := &models.TodayReviewStats{}

	query := `
		SELECT
			(SELECT COUNT(*) FROM review_results WHERE created_at >= CURRENT_DATE AND created_at < CURRENT_DATE + INTERVAL '1 day') AS comment_first,
			(SELECT COUNT(*) FROM second_review_results WHERE created_at >= CURRENT_DATE AND created_at < CURRENT_DATE + INTERVAL '1 day') AS comment_second,
			(SELECT COUNT(*) FROM video_queue_results WHERE created_at >= CURRENT_DATE AND created_at < CURRENT_DATE + INTERVAL '1 day') AS video_queue,
			(SELECT COUNT(*) FROM video_first_review_results WHERE created_at >= CURRENT_DATE AND created_at < CURRENT_DATE + INTERVAL '1 day') AS video_first,
			(SELECT COUNT(*) FROM video_second_review_results WHERE created_at >= CURRENT_DATE AND created_at < CURRENT_DATE + INTERVAL '1 day') AS video_second
	`
	if err := r.db.QueryRow(query).Scan(
		&stats.Comment.FirstReview,
		&stats.Comment.SecondReview,
		&stats.Video.Queue,
		&stats.Video.FirstReview,
		&stats.Video.SecondReview,
	); err != nil {
		return nil, err
	}

	stats.Comment.Total = stats.Comment.FirstReview + stats.Comment.SecondReview
	stats.Video.Total = stats.Video.Queue + stats.Video.FirstReview + stats.Video.SecondReview

	return stats, nil
}

// GetHourlyStats returns hourly statistics for a specific date across all review types
func (r *StatsRepository) GetHourlyStats(date string) ([]models.HourlyStatItem, error) {
	query := `
		SELECT
			EXTRACT(HOUR FROM created_at)::int as hour,
			COUNT(*) as count
		FROM (
			-- Comment first review
			SELECT created_at FROM review_results WHERE DATE(created_at) = $1
			UNION ALL
			-- Comment second review
			SELECT created_at FROM second_review_results WHERE DATE(created_at) = $1
			UNION ALL
			-- Quality check
			SELECT created_at FROM quality_check_results WHERE DATE(created_at) = $1
			UNION ALL
			-- Video first review
			SELECT created_at FROM video_first_review_results WHERE DATE(created_at) = $1
			UNION ALL
			-- Video second review
			SELECT created_at FROM video_second_review_results WHERE DATE(created_at) = $1
		) all_reviews
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

// GetTagStats returns violation tag statistics from comment reviews
func (r *StatsRepository) GetTagStats() ([]models.TagStats, error) {
	query := `
		WITH tag_counts AS (
			SELECT
				unnest(tags) as tag_name,
				COUNT(*) as count
			FROM review_results
			WHERE is_approved = false AND tags IS NOT NULL
			GROUP BY tag_name
		),
		total AS (
			SELECT SUM(count) as total_count FROM tag_counts
		)
		SELECT
			tc.tag_name,
			tc.count,
			CASE WHEN t.total_count > 0 THEN tc.count::float / t.total_count ELSE 0 END as percentage
		FROM tag_counts tc, total t
		ORDER BY tc.count DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []models.TagStats{}
	for rows.Next() {
		var stat models.TagStats
		if err := rows.Scan(&stat.TagName, &stat.Count, &stat.Percentage); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// GetVideoQualityTagStats returns video quality tag statistics across all dimensions
func (r *StatsRepository) GetVideoQualityTagStats() ([]models.VideoQualityTagStats, error) {
	query := `
		WITH video_tags AS (
			-- Content quality tags from first review
			SELECT
				jsonb_array_elements_text(quality_dimensions->'content_quality'->'tags') as tag_name,
				'content_quality' as category
			FROM video_first_review_results
			WHERE quality_dimensions->'content_quality'->'tags' IS NOT NULL

			UNION ALL

			-- Technical quality tags from first review
			SELECT
				jsonb_array_elements_text(quality_dimensions->'technical_quality'->'tags') as tag_name,
				'technical_quality' as category
			FROM video_first_review_results
			WHERE quality_dimensions->'technical_quality'->'tags' IS NOT NULL

			UNION ALL

			-- Compliance tags from first review
			SELECT
				jsonb_array_elements_text(quality_dimensions->'compliance'->'tags') as tag_name,
				'compliance' as category
			FROM video_first_review_results
			WHERE quality_dimensions->'compliance'->'tags' IS NOT NULL

			UNION ALL

			-- Engagement potential tags from first review
			SELECT
				jsonb_array_elements_text(quality_dimensions->'engagement_potential'->'tags') as tag_name,
				'engagement_potential' as category
			FROM video_first_review_results
			WHERE quality_dimensions->'engagement_potential'->'tags' IS NOT NULL

			UNION ALL

			-- Content quality tags from second review
			SELECT
				jsonb_array_elements_text(quality_dimensions->'content_quality'->'tags') as tag_name,
				'content_quality' as category
			FROM video_second_review_results
			WHERE quality_dimensions->'content_quality'->'tags' IS NOT NULL

			UNION ALL

			-- Technical quality tags from second review
			SELECT
				jsonb_array_elements_text(quality_dimensions->'technical_quality'->'tags') as tag_name,
				'technical_quality' as category
			FROM video_second_review_results
			WHERE quality_dimensions->'technical_quality'->'tags' IS NOT NULL

			UNION ALL

			-- Compliance tags from second review
			SELECT
				jsonb_array_elements_text(quality_dimensions->'compliance'->'tags') as tag_name,
				'compliance' as category
			FROM video_second_review_results
			WHERE quality_dimensions->'compliance'->'tags' IS NOT NULL

			UNION ALL

			-- Engagement potential tags from second review
			SELECT
				jsonb_array_elements_text(quality_dimensions->'engagement_potential'->'tags') as tag_name,
				'engagement_potential' as category
			FROM video_second_review_results
			WHERE quality_dimensions->'engagement_potential'->'tags' IS NOT NULL
		)
		SELECT
			tag_name,
			category,
			COUNT(*) as count
		FROM video_tags
		GROUP BY tag_name, category
		ORDER BY category, count DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []models.VideoQualityTagStats{}
	for rows.Next() {
		var stat models.VideoQualityTagStats
		if err := rows.Scan(&stat.TagName, &stat.Category, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// GetVideoQualityAnalysis returns comprehensive video quality analysis
func (r *StatsRepository) GetVideoQualityAnalysis() (*models.VideoQualityAnalysis, error) {
	analysis := &models.VideoQualityAnalysis{
		ScoreDistribution:       make(map[string]int),
		TrafficPoolDistribution: make(map[string]int),
	}

	// Get average scores
	avgQuery := `
		SELECT
			COALESCE(AVG((quality_dimensions->'content_quality'->>'score')::int), 0) as avg_content,
			COALESCE(AVG((quality_dimensions->'technical_quality'->>'score')::int), 0) as avg_technical,
			COALESCE(AVG((quality_dimensions->'compliance'->>'score')::int), 0) as avg_compliance,
			COALESCE(AVG((quality_dimensions->'engagement_potential'->>'score')::int), 0) as avg_engagement,
			COALESCE(AVG(overall_score), 0) as avg_overall,
			COUNT(*) as total
		FROM (
			SELECT quality_dimensions, overall_score FROM video_first_review_results
			UNION ALL
			SELECT quality_dimensions, overall_score FROM video_second_review_results
		) combined_reviews
		WHERE quality_dimensions IS NOT NULL
	`

	err := r.db.QueryRow(avgQuery).Scan(
		&analysis.AvgContentQuality,
		&analysis.AvgTechnicalQuality,
		&analysis.AvgCompliance,
		&analysis.AvgEngagementPotential,
		&analysis.AvgOverallScore,
		&analysis.TotalVideos,
	)
	if err != nil {
		return nil, err
	}

	// Get score distribution
	distQuery := `
		SELECT
			CASE
				WHEN overall_score BETWEEN 1 AND 8 THEN '1-8 (低质量)'
				WHEN overall_score BETWEEN 9 AND 16 THEN '9-16 (中等)'
				WHEN overall_score BETWEEN 17 AND 24 THEN '17-24 (良好)'
				WHEN overall_score BETWEEN 25 AND 32 THEN '25-32 (优秀)'
				WHEN overall_score BETWEEN 33 AND 40 THEN '33-40 (卓越)'
			END as score_range,
			COUNT(*) as count
		FROM (
			SELECT overall_score FROM video_first_review_results
			UNION ALL
			SELECT overall_score FROM video_second_review_results
		) combined_reviews
		WHERE overall_score IS NOT NULL
		GROUP BY score_range
		ORDER BY score_range
	`

	rows, err := r.db.Query(distQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var scoreRange string
		var count int
		if err := rows.Scan(&scoreRange, &count); err != nil {
			return nil, err
		}
		analysis.ScoreDistribution[scoreRange] = count
	}

	// Get traffic pool distribution
	poolQuery := `
		SELECT
			COALESCE(traffic_pool_result, '未分配') as pool,
			COUNT(*) as count
		FROM (
			SELECT traffic_pool_result FROM video_first_review_results
			UNION ALL
			SELECT traffic_pool_result FROM video_second_review_results
		) combined_reviews
		GROUP BY pool
		ORDER BY count DESC
	`

	rows, err = r.db.Query(poolQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pool string
		var count int
		if err := rows.Scan(&pool, &count); err != nil {
			return nil, err
		}
		analysis.TrafficPoolDistribution[pool] = count
	}

	// Get top tags by category
	allTags, err := r.GetVideoQualityTagStats()
	if err != nil {
		return nil, err
	}

	// Group tags by category and get top N
	const topN = 10
	for _, tag := range allTags {
		switch tag.Category {
		case "content_quality":
			if len(analysis.TopContentTags) < topN {
				analysis.TopContentTags = append(analysis.TopContentTags, tag)
			}
		case "technical_quality":
			if len(analysis.TopTechnicalTags) < topN {
				analysis.TopTechnicalTags = append(analysis.TopTechnicalTags, tag)
			}
		case "compliance":
			if len(analysis.TopComplianceTags) < topN {
				analysis.TopComplianceTags = append(analysis.TopComplianceTags, tag)
			}
		case "engagement_potential":
			if len(analysis.TopEngagementTags) < topN {
				analysis.TopEngagementTags = append(analysis.TopEngagementTags, tag)
			}
		}
	}

	return analysis, nil
}

// GetReviewerPerformance returns reviewer performance statistics across all review types
func (r *StatsRepository) GetReviewerPerformance(limit int) ([]models.ReviewerPerformance, error) {
	query := `
		WITH all_reviews AS (
			-- Comment first review
			SELECT reviewer_id, is_approved, 'comment_first' as review_type
			FROM review_results

			UNION ALL

			-- Comment second review
			SELECT reviewer_id, is_approved, 'comment_second' as review_type
			FROM second_review_results

			UNION ALL

			-- Quality check
			SELECT reviewer_id, is_passed as is_approved, 'quality_check' as review_type
			FROM quality_check_results

			UNION ALL

			-- Video first review
			SELECT reviewer_id, is_approved, 'video_first' as review_type
			FROM video_first_review_results

			UNION ALL

			-- Video second review
			SELECT reviewer_id, is_approved, 'video_second' as review_type
			FROM video_second_review_results
		)
		SELECT
			u.id,
			u.username,
			COUNT(*) as total_reviews,
			COUNT(CASE WHEN ar.is_approved = true THEN 1 END) as approved_count,
			COUNT(CASE WHEN ar.is_approved = false THEN 1 END) as rejected_count,
			CASE
				WHEN COUNT(*) > 0 THEN
					ROUND(COUNT(CASE WHEN ar.is_approved = true THEN 1 END)::numeric / COUNT(*)::numeric * 100, 2)
				ELSE 0
			END as approval_rate,
			COUNT(CASE WHEN ar.review_type = 'comment_first' THEN 1 END) as comment_first_reviews,
			COUNT(CASE WHEN ar.review_type = 'comment_second' THEN 1 END) as comment_second_reviews,
			COUNT(CASE WHEN ar.review_type = 'quality_check' THEN 1 END) as quality_checks,
			COUNT(CASE WHEN ar.review_type = 'video_first' THEN 1 END) as video_first_reviews,
			COUNT(CASE WHEN ar.review_type = 'video_second' THEN 1 END) as video_second_reviews
		FROM users u
		INNER JOIN all_reviews ar ON u.id = ar.reviewer_id
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
		if err := rows.Scan(
			&perf.ReviewerID,
			&perf.Username,
			&perf.TotalReviews,
			&perf.ApprovedCount,
			&perf.RejectedCount,
			&perf.ApprovalRate,
			&perf.CommentFirstReviews,
			&perf.CommentSecondReviews,
			&perf.QualityChecks,
			&perf.VideoFirstReviews,
			&perf.VideoSecondReviews,
		); err != nil {
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

// getQueueStats returns statistics for each queue using the unified real-time view
func (r *StatsRepository) getQueueStats() ([]models.QueueStats, error) {
	query := `
		WITH approval_stats AS (
			-- Comment first review
			SELECT
				'comment_first_review' as queue_name,
				COUNT(CASE WHEN is_approved = true THEN 1 END) as approved_count,
				COUNT(CASE WHEN is_approved = false THEN 1 END) as rejected_count
			FROM review_results

			UNION ALL

			-- Comment second review
			SELECT
				'comment_second_review' as queue_name,
				COUNT(CASE WHEN is_approved = true THEN 1 END) as approved_count,
				COUNT(CASE WHEN is_approved = false THEN 1 END) as rejected_count
			FROM second_review_results

			UNION ALL

			-- AI vs human diff
			SELECT
				'ai_human_diff' as queue_name,
				COUNT(CASE WHEN is_approved = true THEN 1 END) as approved_count,
				COUNT(CASE WHEN is_approved = false THEN 1 END) as rejected_count
			FROM ai_human_diff_results

			UNION ALL

			-- Quality check
			SELECT
				'quality_check' as queue_name,
				COUNT(CASE WHEN is_passed = true THEN 1 END) as approved_count,
				COUNT(CASE WHEN is_passed = false THEN 1 END) as rejected_count
			FROM quality_check_results

			UNION ALL

			-- Video first review
			SELECT
				'video_first_review' as queue_name,
				COUNT(CASE WHEN is_approved = true THEN 1 END) as approved_count,
				COUNT(CASE WHEN is_approved = false THEN 1 END) as rejected_count
			FROM video_first_review_results

			UNION ALL

			-- Video second review
			SELECT
				'video_second_review' as queue_name,
				COUNT(CASE WHEN is_approved = true THEN 1 END) as approved_count,
				COUNT(CASE WHEN is_approved = false THEN 1 END) as rejected_count
			FROM video_second_review_results
		)
		SELECT
			uqs.queue_name,
			uqs.total_tasks,
			uqs.completed_tasks,
			uqs.pending_tasks,
			uqs.is_active,
			COALESCE(ast.approved_count, 0) as approved_count,
			COALESCE(ast.rejected_count, 0) as rejected_count,
			CASE
				WHEN uqs.completed_tasks > 0 AND COALESCE(ast.approved_count, 0) + COALESCE(ast.rejected_count, 0) > 0 THEN
					ROUND(COALESCE(ast.approved_count, 0)::numeric / (COALESCE(ast.approved_count, 0) + COALESCE(ast.rejected_count, 0))::numeric * 100, 2)
				ELSE 0
			END as approval_rate,
			COALESCE(uqs.avg_process_time_minutes, 0) as avg_process_time
		FROM unified_queue_stats uqs
		LEFT JOIN approval_stats ast ON uqs.queue_name = ast.queue_name
		ORDER BY uqs.priority DESC
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
