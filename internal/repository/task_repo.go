package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{db: database.DB}
}

var ErrReviewResultConflict = errors.New("review result already exists with different content")

type reviewResultExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// CreateTask creates a new review task
func (r *TaskRepository) CreateTask(commentID int64) error {
	query := `
		INSERT INTO review_tasks (comment_id, status, created_at)
		VALUES ($1, 'pending', NOW())
	`
	_, err := r.db.Exec(query, commentID)
	return err
}

// ClaimTasks claims pending tasks for a reviewer
func (r *TaskRepository) ClaimTasks(reviewerID int, limit int) ([]models.ReviewTask, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Select pending tasks
	query := `
		SELECT id, comment_id, created_at
		FROM review_tasks
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	rows, err := tx.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	taskIDs := []int{}
	tasks := []models.ReviewTask{}

	for rows.Next() {
		var task models.ReviewTask
		if err := rows.Scan(&task.ID, &task.CommentID, &task.CreatedAt); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, task.ID)
		tasks = append(tasks, task)
	}

	if len(taskIDs) == 0 {
		return []models.ReviewTask{}, nil
	}

	// Update tasks to in_progress
	now := time.Now()
	updateQuery := `
		UPDATE review_tasks
		SET status = 'in_progress', reviewer_id = $1, claimed_at = $2
		WHERE id = ANY($3)
	`
	_, err = tx.Exec(updateQuery, reviewerID, now, pq.Array(taskIDs))
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Fetch full task details with comments
	return r.FindTasksWithComments(taskIDs)
}

// FindTasksWithComments finds tasks with their associated comments
func (r *TaskRepository) FindTasksWithComments(taskIDs []int) ([]models.ReviewTask, error) {
	query := `
		SELECT 
			rt.id, rt.comment_id, rt.reviewer_id, rt.status, 
			rt.claimed_at, rt.completed_at, rt.created_at,
			c.id, c.text
		FROM review_tasks rt
		INNER JOIN comment c ON rt.comment_id = c.id
		WHERE rt.id = ANY($1)
		ORDER BY rt.id
	`
	rows, err := r.db.Query(query, pq.Array(taskIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.ReviewTask{}
	for rows.Next() {
		var task models.ReviewTask
		var comment models.Comment
		err := rows.Scan(
			&task.ID, &task.CommentID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&comment.ID, &comment.Text,
		)
		if err != nil {
			return nil, err
		}
		task.Comment = &comment
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// GetMyTasks gets all in-progress tasks for a reviewer
func (r *TaskRepository) GetMyTasks(reviewerID int) ([]models.ReviewTask, error) {
	query := `
		SELECT 
			rt.id, rt.comment_id, rt.reviewer_id, rt.status, 
			rt.claimed_at, rt.completed_at, rt.created_at,
			c.id, c.text
		FROM review_tasks rt
		INNER JOIN comment c ON rt.comment_id = c.id
		WHERE rt.reviewer_id = $1 AND rt.status = 'in_progress'
		ORDER BY rt.claimed_at DESC
	`
	rows, err := r.db.Query(query, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.ReviewTask{}
	for rows.Next() {
		var task models.ReviewTask
		var comment models.Comment
		err := rows.Scan(
			&task.ID, &task.CommentID, &task.ReviewerID, &task.Status,
			&task.ClaimedAt, &task.CompletedAt, &task.CreatedAt,
			&comment.ID, &comment.Text,
		)
		if err != nil {
			return nil, err
		}
		task.Comment = &comment
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// CompleteTask marks a task as completed
func (r *TaskRepository) CompleteTask(taskID, reviewerID int) error {
	query := `
		UPDATE review_tasks
		SET status = 'completed', completed_at = NOW()
		WHERE id = $1 AND reviewer_id = $2 AND status = 'in_progress'
	`
	result, err := r.db.Exec(query, taskID, reviewerID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetCommentIDTx retrieves a task's comment ID within a transaction
func (r *TaskRepository) GetCommentIDTx(tx *sql.Tx, taskID int) (int64, error) {
	query := `SELECT comment_id FROM review_tasks WHERE id = $1`
	var commentID int64
	if err := tx.QueryRow(query, taskID).Scan(&commentID); err != nil {
		return 0, err
	}
	return commentID, nil
}

// CompleteTaskTx marks a task as completed within a transaction
func (r *TaskRepository) CompleteTaskTx(tx *sql.Tx, taskID, reviewerID int) error {
	query := `
		UPDATE review_tasks
		SET status = 'completed', completed_at = NOW()
		WHERE id = $1 AND reviewer_id = $2 AND status = 'in_progress'
	`
	result, err := tx.Exec(query, taskID, reviewerID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CreateReviewResult creates a review result
func (r *TaskRepository) CreateReviewResult(result *models.ReviewResult) (bool, error) {
	return createReviewResult(r.db, result)
}

// CreateReviewResultTx creates a review result within a transaction
func (r *TaskRepository) CreateReviewResultTx(tx *sql.Tx, result *models.ReviewResult) (bool, error) {
	return createReviewResult(tx, result)
}

func createReviewResult(db reviewResultExecutor, result *models.ReviewResult) (bool, error) {
	query := `
		INSERT INTO review_results (task_id, reviewer_id, is_approved, tags, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (task_id) DO NOTHING
		RETURNING id, created_at
	`
	err := db.QueryRow(query, result.TaskID, result.ReviewerID, result.IsApproved,
		pq.Array(result.Tags), result.Reason).Scan(&result.ID, &result.CreatedAt)
	if err == nil {
		return true, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}

	expectedIsApproved := result.IsApproved
	expectedReason := result.Reason
	expectedTags := append([]string(nil), result.Tags...)

	existingQuery := `
		SELECT id, reviewer_id, is_approved, tags, reason, created_at
		FROM review_results
		WHERE task_id = $1
	`
	var tags []string
	err = db.QueryRow(existingQuery, result.TaskID).Scan(
		&result.ID,
		&result.ReviewerID,
		&result.IsApproved,
		pq.Array(&tags),
		&result.Reason,
		&result.CreatedAt,
	)
	if err != nil {
		return false, err
	}
	if !equalTags(tags, expectedTags) || result.Reason != expectedReason || result.IsApproved != expectedIsApproved {
		return false, ErrReviewResultConflict
	}
	result.Tags = tags
	return false, nil
}

// CountByStatus counts tasks by status
func (r *TaskRepository) CountByStatus(status string) (int, error) {
	query := `SELECT COUNT(*) FROM review_tasks WHERE status = $1`
	var count int
	err := r.db.QueryRow(query, status).Scan(&count)
	return count, err
}

func equalTags(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	left := append([]string(nil), a...)
	right := append([]string(nil), b...)
	sort.Strings(left)
	sort.Strings(right)
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

// FindExpiredTasks finds tasks that have been in progress for too long
func (r *TaskRepository) FindExpiredTasks(timeoutMinutes int) ([]models.ReviewTask, error) {
	query := `
		SELECT id, comment_id, reviewer_id, claimed_at
		FROM review_tasks
		WHERE status = 'in_progress' 
		AND claimed_at < NOW() - INTERVAL '1 minute' * $1
	`
	rows, err := r.db.Query(query, timeoutMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []models.ReviewTask{}
	for rows.Next() {
		var task models.ReviewTask
		if err := rows.Scan(&task.ID, &task.CommentID, &task.ReviewerID, &task.ClaimedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// ResetTask resets a task back to pending status
func (r *TaskRepository) ResetTask(taskID int) error {
	query := `
		UPDATE review_tasks
		SET status = 'pending', reviewer_id = NULL, claimed_at = NULL
		WHERE id = $1
	`
	_, err := r.db.Exec(query, taskID)
	return err
}

// ReturnTasks returns multiple tasks back to pending status for a specific reviewer
func (r *TaskRepository) ReturnTasks(taskIDs []int, reviewerID int) (int, error) {
	query := `
		UPDATE review_tasks
		SET status = 'pending', reviewer_id = NULL, claimed_at = NULL
		WHERE id = ANY($1) AND reviewer_id = $2 AND status = 'in_progress'
	`
	result, err := r.db.Exec(query, pq.Array(taskIDs), reviewerID)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

// SearchTasks searches review tasks with filters and pagination
func (r *TaskRepository) SearchTasks(req models.SearchTasksRequest) ([]models.TaskSearchResult, int, error) {
	return r.SearchTasksUnified(req)
}

// SearchTasksUnified searches review tasks with database-level sorting and pagination
// This is optimized to handle both first and second review queues in a single query
func (r *TaskRepository) SearchTasksUnified(req models.SearchTasksRequest) ([]models.TaskSearchResult, int, error) {
	type queryPlaceholders struct {
		contentID string
		reviewer  string
		tagIDs    string
		startTime string
		endTime   string
		pool      string
	}

	placeholders := queryPlaceholders{}
	var args []interface{}
	argPos := 1

	if req.CommentID != nil {
		placeholders.contentID = fmt.Sprintf("$%d", argPos)
		args = append(args, *req.CommentID)
		argPos++
	}

	if req.ReviewerRTX != "" {
		placeholders.reviewer = fmt.Sprintf("$%d", argPos)
		args = append(args, req.ReviewerRTX)
		argPos++
	}

	if req.TagIDs != "" {
		tagIDs := strings.Split(req.TagIDs, ",")
		placeholders.tagIDs = fmt.Sprintf("$%d", argPos)
		args = append(args, pq.Array(tagIDs))
		argPos++
	}

	if req.ReviewStartTime != nil {
		placeholders.startTime = fmt.Sprintf("$%d", argPos)
		args = append(args, *req.ReviewStartTime)
		argPos++
	}

	if req.ReviewEndTime != nil {
		placeholders.endTime = fmt.Sprintf("$%d", argPos)
		args = append(args, *req.ReviewEndTime)
		argPos++
	}

	queueName := strings.TrimSpace(strings.ToLower(req.QueueName))
	videoQueuePool := ""
	if strings.HasPrefix(queueName, "video_queue_") {
		videoQueuePool = strings.TrimPrefix(queueName, "video_queue_")
		placeholders.pool = fmt.Sprintf("$%d", argPos)
		args = append(args, videoQueuePool)
		argPos++
	}

	includeQueues := map[string]bool{
		"comment_first_review":  false,
		"comment_second_review": false,
		"ai_human_diff":         false,
		"quality_check":         false,
		"video_first_review":    false,
		"video_second_review":   false,
		"video_queue":           false,
	}

	enableAllQueues := func() {
		for key := range includeQueues {
			includeQueues[key] = true
		}
	}

	switch queueName {
	case "", "all":
		switch strings.ToLower(req.QueueType) {
		case "first":
			includeQueues["comment_first_review"] = true
		case "second":
			includeQueues["comment_second_review"] = true
		default:
			enableAllQueues()
		}
	case "first":
		includeQueues["comment_first_review"] = true
	case "second":
		includeQueues["comment_second_review"] = true
	case "comment_first_review", "comment_second_review", "ai_human_diff", "quality_check", "video_first_review", "video_second_review", "video_queue":
		includeQueues[queueName] = true
	default:
		if videoQueuePool != "" {
			includeQueues["video_queue"] = true
		} else {
			enableAllQueues()
		}
	}

	hasQueue := false
	for _, enabled := range includeQueues {
		if enabled {
			hasQueue = true
			break
		}
	}
	if !hasQueue {
		enableAllQueues()
	}

	buildWhereClause := func(statusCol, contentIDCol, reviewerCol, tagsCondition, completedAtCol string, extraConditions ...string) string {
		conditions := []string{fmt.Sprintf("%s = 'completed'", statusCol)}

		if placeholders.contentID != "" {
			conditions = append(conditions, fmt.Sprintf("%s = %s", contentIDCol, placeholders.contentID))
		}

		if placeholders.reviewer != "" {
			conditions = append(conditions, fmt.Sprintf("%s = %s", reviewerCol, placeholders.reviewer))
		}

		if placeholders.tagIDs != "" {
			if tagsCondition != "" {
				conditions = append(conditions, tagsCondition)
			} else {
				conditions = append(conditions, "1=0")
			}
		}

		if placeholders.startTime != "" {
			conditions = append(conditions, fmt.Sprintf("%s >= %s", completedAtCol, placeholders.startTime))
		}

		if placeholders.endTime != "" {
			conditions = append(conditions, fmt.Sprintf("%s <= %s", completedAtCol, placeholders.endTime))
		}

		for _, extra := range extraConditions {
			if strings.TrimSpace(extra) != "" {
				conditions = append(conditions, extra)
			}
		}

		return "WHERE " + strings.Join(conditions, " AND ")
	}

	var queueSources []string

	if includeQueues["comment_first_review"] {
		whereClause := buildWhereClause(
			"rt.status",
			"rt.comment_id",
			"u.username",
			fmt.Sprintf("rr.tags && %s", placeholders.tagIDs),
			"rt.completed_at",
		)
		queueSources = append(queueSources, fmt.Sprintf(`
			SELECT
				rt.id AS id,
				'comment_first_review' AS queue_name,
				'comment' AS content_type,
				rt.comment_id::bigint AS content_id,
				COALESCE(c.text, '') AS content_text,
				rt.reviewer_id AS reviewer_id,
				u.username AS reviewer_username,
				rt.status AS status,
				rt.claimed_at,
				rt.completed_at,
				rt.created_at,
				CASE
					WHEN rr.is_approved = true THEN 'approved'
					WHEN rr.is_approved = false THEN 'rejected'
					ELSE NULL
				END AS decision,
				COALESCE(rr.tags, '{}'::text[]) AS tags,
				rr.reason AS reason,
				NULL::text AS pool,
				NULL::integer AS overall_score,
				NULL::text AS traffic_pool_result
			FROM review_tasks rt
			LEFT JOIN review_results rr ON rt.id = rr.task_id
			LEFT JOIN users u ON rt.reviewer_id = u.id
			LEFT JOIN comment c ON rt.comment_id = c.id
			%s
		`, whereClause))
	}

	if includeQueues["comment_second_review"] {
		whereClause := buildWhereClause(
			"srt.status",
			"srt.comment_id",
			"u.username",
			fmt.Sprintf("srr.tags && %s", placeholders.tagIDs),
			"srt.completed_at",
		)
		queueSources = append(queueSources, fmt.Sprintf(`
			SELECT
				srt.id AS id,
				'comment_second_review' AS queue_name,
				'comment' AS content_type,
				srt.comment_id::bigint AS content_id,
				COALESCE(c.text, '') AS content_text,
				srt.reviewer_id AS reviewer_id,
				u.username AS reviewer_username,
				srt.status AS status,
				srt.claimed_at,
				srt.completed_at,
				srt.created_at,
				CASE
					WHEN srr.is_approved = true THEN 'approved'
					WHEN srr.is_approved = false THEN 'rejected'
					ELSE NULL
				END AS decision,
				COALESCE(srr.tags, '{}'::text[]) AS tags,
				srr.reason AS reason,
				NULL::text AS pool,
				NULL::integer AS overall_score,
				NULL::text AS traffic_pool_result
			FROM second_review_tasks srt
			LEFT JOIN second_review_results srr ON srt.id = srr.second_task_id
			LEFT JOIN users u ON srt.reviewer_id = u.id
			LEFT JOIN comment c ON srt.comment_id = c.id
			%s
		`, whereClause))
	}

	if includeQueues["ai_human_diff"] {
		whereClause := buildWhereClause(
			"ahdt.status",
			"ahdt.comment_id",
			"u.username",
			fmt.Sprintf("ahr.tags && %s", placeholders.tagIDs),
			"ahdt.completed_at",
		)
		queueSources = append(queueSources, fmt.Sprintf(`
			SELECT
				ahdt.id AS id,
				'ai_human_diff' AS queue_name,
				'comment' AS content_type,
				ahdt.comment_id::bigint AS content_id,
				COALESCE(c.text, '') AS content_text,
				ahdt.reviewer_id AS reviewer_id,
				u.username AS reviewer_username,
				ahdt.status AS status,
				ahdt.claimed_at,
				ahdt.completed_at,
				ahdt.created_at,
				CASE
					WHEN ahr.is_approved = true THEN 'approved'
					WHEN ahr.is_approved = false THEN 'rejected'
					ELSE NULL
				END AS decision,
				COALESCE(ahr.tags, '{}'::text[]) AS tags,
				ahr.reason AS reason,
				NULL::text AS pool,
				NULL::integer AS overall_score,
				NULL::text AS traffic_pool_result
			FROM ai_human_diff_tasks ahdt
			LEFT JOIN ai_human_diff_results ahr ON ahdt.id = ahr.task_id
			LEFT JOIN users u ON ahdt.reviewer_id = u.id
			LEFT JOIN comment c ON ahdt.comment_id = c.id
			%s
		`, whereClause))
	}

	if includeQueues["quality_check"] {
		whereClause := buildWhereClause(
			"qct.status",
			"qct.comment_id",
			"u.username",
			fmt.Sprintf("qcr.error_type = ANY(%s)", placeholders.tagIDs),
			"qct.completed_at",
		)
		queueSources = append(queueSources, fmt.Sprintf(`
			SELECT
				qct.id AS id,
				'quality_check' AS queue_name,
				'comment' AS content_type,
				qct.comment_id::bigint AS content_id,
				COALESCE(c.text, '') AS content_text,
				qct.reviewer_id AS reviewer_id,
				u.username AS reviewer_username,
				qct.status AS status,
				qct.claimed_at,
				qct.completed_at,
				qct.created_at,
				CASE
					WHEN qcr.is_passed = true THEN 'passed'
					WHEN qcr.is_passed = false THEN 'failed'
					ELSE NULL
				END AS decision,
				CASE
					WHEN qcr.error_type IS NOT NULL THEN ARRAY[qcr.error_type]
					ELSE '{}'::text[]
				END AS tags,
				qcr.qc_comment AS reason,
				NULL::text AS pool,
				NULL::integer AS overall_score,
				NULL::text AS traffic_pool_result
			FROM quality_check_tasks qct
			LEFT JOIN quality_check_results qcr ON qct.id = qcr.qc_task_id
			LEFT JOIN users u ON qct.reviewer_id = u.id
			LEFT JOIN comment c ON qct.comment_id = c.id
			%s
		`, whereClause))
	}

	if includeQueues["video_first_review"] {
		whereClause := buildWhereClause(
			"vfrt.status",
			"vfrt.video_id",
			"u.username",
			"",
			"vfrt.completed_at",
		)
		queueSources = append(queueSources, fmt.Sprintf(`
			SELECT
				vfrt.id AS id,
				'video_first_review' AS queue_name,
				'video' AS content_type,
				vfrt.video_id::bigint AS content_id,
				COALESCE(tv.filename, '') AS content_text,
				vfrt.reviewer_id AS reviewer_id,
				u.username AS reviewer_username,
				vfrt.status AS status,
				vfrt.claimed_at,
				vfrt.completed_at,
				vfrt.created_at,
				CASE
					WHEN vfrr.is_approved = true THEN 'approved'
					WHEN vfrr.is_approved = false THEN 'rejected'
					ELSE NULL
				END AS decision,
				'{}'::text[] AS tags,
				vfrr.reason AS reason,
				NULL::text AS pool,
				vfrr.overall_score AS overall_score,
				vfrr.traffic_pool_result AS traffic_pool_result
			FROM video_first_review_tasks vfrt
			LEFT JOIN video_first_review_results vfrr ON vfrt.id = vfrr.task_id
			LEFT JOIN users u ON vfrt.reviewer_id = u.id
			LEFT JOIN tiktok_videos tv ON vfrt.video_id = tv.id
			%s
		`, whereClause))
	}

	if includeQueues["video_second_review"] {
		whereClause := buildWhereClause(
			"vsrt.status",
			"vsrt.video_id",
			"u.username",
			"",
			"vsrt.completed_at",
		)
		queueSources = append(queueSources, fmt.Sprintf(`
			SELECT
				vsrt.id AS id,
				'video_second_review' AS queue_name,
				'video' AS content_type,
				vsrt.video_id::bigint AS content_id,
				COALESCE(tv.filename, '') AS content_text,
				vsrt.reviewer_id AS reviewer_id,
				u.username AS reviewer_username,
				vsrt.status AS status,
				vsrt.claimed_at,
				vsrt.completed_at,
				vsrt.created_at,
				CASE
					WHEN vsrr.is_approved = true THEN 'approved'
					WHEN vsrr.is_approved = false THEN 'rejected'
					ELSE NULL
				END AS decision,
				'{}'::text[] AS tags,
				vsrr.reason AS reason,
				NULL::text AS pool,
				vsrr.overall_score AS overall_score,
				vsrr.traffic_pool_result AS traffic_pool_result
			FROM video_second_review_tasks vsrt
			LEFT JOIN video_second_review_results vsrr ON vsrt.id = vsrr.second_task_id
			LEFT JOIN users u ON vsrt.reviewer_id = u.id
			LEFT JOIN tiktok_videos tv ON vsrt.video_id = tv.id
			%s
		`, whereClause))
	}

	if includeQueues["video_queue"] {
		poolCondition := ""
		if placeholders.pool != "" {
			poolCondition = fmt.Sprintf("vqt.pool = %s", placeholders.pool)
		}
		whereClause := buildWhereClause(
			"vqt.status",
			"vqt.video_id",
			"u.username",
			fmt.Sprintf("vqr.tags && %s", placeholders.tagIDs),
			"vqt.completed_at",
			poolCondition,
		)
		queueSources = append(queueSources, fmt.Sprintf(`
			SELECT
				vqt.id AS id,
				'video_queue' AS queue_name,
				'video' AS content_type,
				vqt.video_id::bigint AS content_id,
				COALESCE(tv.filename, '') AS content_text,
				vqt.reviewer_id AS reviewer_id,
				u.username AS reviewer_username,
				vqt.status AS status,
				vqt.claimed_at,
				vqt.completed_at,
				vqt.created_at,
				vqr.review_decision AS decision,
				COALESCE(vqr.tags, '{}'::text[]) AS tags,
				vqr.reason AS reason,
				vqt.pool AS pool,
				NULL::integer AS overall_score,
				NULL::text AS traffic_pool_result
			FROM video_queue_tasks vqt
			LEFT JOIN video_queue_results vqr ON vqt.id = vqr.task_id
			LEFT JOIN users u ON vqt.reviewer_id = u.id
			LEFT JOIN tiktok_videos tv ON vqt.video_id = tv.id
			%s
		`, whereClause))
	}

	if len(queueSources) == 0 {
		return []models.TaskSearchResult{}, 0, nil
	}

	unionQuery := strings.Join(queueSources, " UNION ALL ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) combined_results", unionQuery)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT * FROM (%s) combined_results
		ORDER BY completed_at DESC
		LIMIT $%d OFFSET $%d
	`, unionQuery, argPos, argPos+1)

	args = append(args, req.PageSize, offset)
	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results := []models.TaskSearchResult{}
	for rows.Next() {
		var result models.TaskSearchResult
		var contentText sql.NullString
		var reviewerID sql.NullInt64
		var reviewerUsername sql.NullString
		var decision sql.NullString
		var tags []string
		var reason sql.NullString
		var pool sql.NullString
		var overallScore sql.NullInt64
		var trafficPoolResult sql.NullString

		if err := rows.Scan(
			&result.ID,
			&result.QueueName,
			&result.ContentType,
			&result.ContentID,
			&contentText,
			&reviewerID,
			&reviewerUsername,
			&result.Status,
			&result.ClaimedAt,
			&result.CompletedAt,
			&result.CreatedAt,
			&decision,
			pq.Array(&tags),
			&reason,
			&pool,
			&overallScore,
			&trafficPoolResult,
		); err != nil {
			return nil, 0, err
		}

		if contentText.Valid {
			result.ContentText = contentText.String
		}

		if reviewerID.Valid {
			id := int(reviewerID.Int64)
			result.ReviewerID = &id
		}

		if reviewerUsername.Valid {
			result.ReviewerUsername = &reviewerUsername.String
		}

		if decision.Valid {
			result.Decision = &decision.String
		}

		if reason.Valid {
			result.Reason = &reason.String
		}

		if pool.Valid {
			result.Pool = &pool.String
		}

		if overallScore.Valid {
			score := int(overallScore.Int64)
			result.OverallScore = &score
		}

		if trafficPoolResult.Valid {
			result.TrafficPoolResult = &trafficPoolResult.String
		}

		if tags == nil {
			result.Tags = []string{}
		} else {
			result.Tags = tags
		}

		results = append(results, result)
	}

	return results, total, nil
}

