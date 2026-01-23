package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type TaskQueueRepository struct {
	db *sql.DB
}

func NewTaskQueueRepository() *TaskQueueRepository {
	return &TaskQueueRepository{
		db: database.DB,
	}
}

// CreateTaskQueue creates a new task queue
// DEPRECATED: The task_queues table is deprecated in favor of unified_queue_stats view.
// This method is kept for backward compatibility but should not be used for new code.
// All queues are now automatically tracked through the unified_queue_stats view.
func (r *TaskQueueRepository) CreateTaskQueue(req models.CreateTaskQueueRequest, adminID int) (*models.TaskQueue, error) {
	query := `
		INSERT INTO task_queues (queue_name, description, priority, total_tasks, completed_tasks, pending_tasks, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	queue := &models.TaskQueue{
		QueueName:      req.QueueName,
		Description:    req.Description,
		Priority:       req.Priority,
		TotalTasks:     req.TotalTasks,
		CompletedTasks: req.CompletedTasks,
		PendingTasks:   req.TotalTasks - req.CompletedTasks,
		IsActive:       true,
	}

	err := r.db.QueryRow(
		query,
		queue.QueueName,
		queue.Description,
		queue.Priority,
		queue.TotalTasks,
		queue.CompletedTasks,
		queue.PendingTasks,
		queue.IsActive,
		now,
		now,
	).Scan(&queue.ID, &queue.CreatedAt, &queue.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create task queue: %w", err)
	}

	return queue, nil
}

// GetTaskQueueByID retrieves a task queue by ID
func (r *TaskQueueRepository) GetTaskQueueByID(id int) (*models.TaskQueue, error) {
	query := `
		SELECT id, queue_name, description, priority, total_tasks, completed_tasks, pending_tasks, is_active, created_at, updated_at
		FROM task_queues
		WHERE id = $1
	`

	var queue models.TaskQueue
	err := r.db.QueryRow(query, id).Scan(
		&queue.ID,
		&queue.QueueName,
		&queue.Description,
		&queue.Priority,
		&queue.TotalTasks,
		&queue.CompletedTasks,
		&queue.PendingTasks,
		&queue.IsActive,
		&queue.CreatedAt,
		&queue.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task queue not found")
		}
		return nil, fmt.Errorf("failed to get task queue: %w", err)
	}

	return &queue, nil
}

// ListTaskQueues returns paginated task queues with statistics
// Uses a fast hardcoded queue list with async stats loading to avoid slow view queries
func (r *TaskQueueRepository) ListTaskQueues(req models.ListTaskQueuesRequest) ([]models.TaskQueue, int, error) {
	log.Printf("📊 ListTaskQueues called: page=%d, pageSize=%d", req.Page, req.PageSize)
	startTime := time.Now()

	// 定义固定的队列列表（避免查询慢视图）
	staticQueues := []struct {
		QueueName   string
		Description string
		Priority    int
		TableName   string
	}{
		{"comment_first_review", "评论一审队列", 100, "review_tasks"},
		{"comment_second_review", "评论二审队列", 90, "second_review_tasks"},
		{"ai_human_diff", "AI与人工diff队列", 85, "ai_human_diff_tasks"},
		{"quality_check", "质量检查队列", 80, "quality_check_tasks"},
		{"video_first_review", "视频一审队列", 70, "video_first_review_tasks"},
		{"video_second_review", "视频二审队列", 60, "video_second_review_tasks"},
	}

	// 过滤搜索条件
	var filteredQueues []struct {
		QueueName   string
		Description string
		Priority    int
		TableName   string
	}

	for _, q := range staticQueues {
		if req.Search != "" {
			searchLower := strings.ToLower(req.Search)
			if !strings.Contains(strings.ToLower(q.QueueName), searchLower) &&
				!strings.Contains(strings.ToLower(q.Description), searchLower) {
				continue
			}
		}
		filteredQueues = append(filteredQueues, q)
	}

	total := len(filteredQueues)

	// 分页
	offset := (req.Page - 1) * req.PageSize
	end := offset + req.PageSize
	if offset > total {
		offset = total
	}
	if end > total {
		end = total
	}

	pagedQueues := filteredQueues[offset:end]
	log.Printf("📊 Processing %d queues after pagination", len(pagedQueues))

	statsQuery := `
		SELECT queue_name, total, completed, pending
		FROM (
			SELECT
				'comment_first_review' AS queue_name,
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'completed') AS completed,
				COUNT(*) FILTER (WHERE status = 'pending') AS pending
			FROM review_tasks
			UNION ALL
			SELECT
				'comment_second_review' AS queue_name,
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'completed') AS completed,
				COUNT(*) FILTER (WHERE status = 'pending') AS pending
			FROM second_review_tasks
			UNION ALL
			SELECT
				'ai_human_diff' AS queue_name,
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'completed') AS completed,
				COUNT(*) FILTER (WHERE status = 'pending') AS pending
			FROM ai_human_diff_tasks
			UNION ALL
			SELECT
				'quality_check' AS queue_name,
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'completed') AS completed,
				COUNT(*) FILTER (WHERE status = 'pending') AS pending
			FROM quality_check_tasks
			UNION ALL
			SELECT
				'video_first_review' AS queue_name,
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'completed') AS completed,
				COUNT(*) FILTER (WHERE status = 'pending') AS pending
			FROM video_first_review_tasks
			UNION ALL
			SELECT
				'video_second_review' AS queue_name,
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status = 'completed') AS completed,
				COUNT(*) FILTER (WHERE status = 'pending') AS pending
			FROM video_second_review_tasks
		) stats
	`
	statsByQueue := make(map[string]struct {
		total     int
		completed int
		pending   int
	})
	statsStart := time.Now()
	rows, err := r.db.Query(statsQuery)
	if err != nil {
		log.Printf("⚠️ Queue stats query error: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var (
				queueName string
				total     int
				completed int
				pending   int
			)
			if err := rows.Scan(&queueName, &total, &completed, &pending); err != nil {
				log.Printf("⚠️ Queue stats scan error: %v", err)
				continue
			}
			statsByQueue[queueName] = struct {
				total     int
				completed int
				pending   int
			}{
				total:     total,
				completed: completed,
				pending:   pending,
			}
		}
		if err := rows.Err(); err != nil {
			log.Printf("⚠️ Queue stats rows error: %v", err)
		}
	}
	log.Printf("📊 Queue stats query took %v", time.Since(statsStart))

	// 构建结果，快速获取每个队列的统计数据
	queues := make([]models.TaskQueue, 0, len(pagedQueues))
	now := time.Now()

	for i, q := range pagedQueues {
		queue := models.TaskQueue{
			ID:          i + 1 + offset,
			QueueName:   q.QueueName,
			Description: q.Description,
			Priority:    q.Priority,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if stats, ok := statsByQueue[q.QueueName]; ok {
			queue.TotalTasks = stats.total
			queue.CompletedTasks = stats.completed
			queue.PendingTasks = stats.pending
		}

		queues = append(queues, queue)
	}

	log.Printf("✅ ListTaskQueues completed in %v, returning %d queues", time.Since(startTime), len(queues))
	return queues, total, nil
}

// UpdateTaskQueue updates a task queue
// DEPRECATED: Manual queue updates are deprecated. Queue statistics are now automatically
// calculated from actual task tables via the unified_queue_stats view.
func (r *TaskQueueRepository) UpdateTaskQueue(id int, req models.UpdateTaskQueueRequest, adminID int) (*models.TaskQueue, error) {
	// Get existing queue first
	queue, err := r.GetTaskQueueByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.QueueName != nil {
		queue.QueueName = *req.QueueName
	}
	if req.Description != nil {
		queue.Description = *req.Description
	}
	if req.Priority != nil {
		queue.Priority = *req.Priority
	}
	if req.TotalTasks != nil {
		queue.TotalTasks = *req.TotalTasks
	}
	if req.CompletedTasks != nil {
		queue.CompletedTasks = *req.CompletedTasks
		queue.PendingTasks = queue.TotalTasks - *req.CompletedTasks
	}
	if req.IsActive != nil {
		queue.IsActive = *req.IsActive
	}

	query := `
		UPDATE task_queues
		SET queue_name = $2, description = $3, priority = $4, total_tasks = $5, 
		    completed_tasks = $6, pending_tasks = $7, is_active = $8, updated_at = $9
		WHERE id = $1
		RETURNING updated_at
	`

	now := time.Now()
	err = r.db.QueryRow(
		query,
		queue.ID,
		queue.QueueName,
		queue.Description,
		queue.Priority,
		queue.TotalTasks,
		queue.CompletedTasks,
		queue.PendingTasks,
		queue.IsActive,
		now,
	).Scan(&queue.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update task queue: %w", err)
	}

	return queue, nil
}

// DeleteTaskQueue deletes a task queue
// DEPRECATED: Manual queue deletion is deprecated. Queues are now automatically
// managed through the unified_queue_stats view based on actual task tables.
func (r *TaskQueueRepository) DeleteTaskQueue(id int) error {
	query := "DELETE FROM task_queues WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task queue: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task queue not found")
	}

	return nil
}

// GetAllTaskQueues returns all task queues with real-time statistics from unified_queue_stats
func (r *TaskQueueRepository) GetAllTaskQueues() ([]models.TaskQueue, error) {
	query := `
		SELECT
			ROW_NUMBER() OVER (ORDER BY priority DESC) as id,
			queue_name,
			description,
			priority,
			total_tasks,
			completed_tasks,
			pending_tasks,
			is_active,
			created_at,
			updated_at
		FROM unified_queue_stats
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query task queues: %w", err)
	}
	defer rows.Close()

	// Initialize as empty slice to avoid null in JSON response
	queues := make([]models.TaskQueue, 0)
	for rows.Next() {
		var queue models.TaskQueue
		err := rows.Scan(
			&queue.ID,
			&queue.QueueName,
			&queue.Description,
			&queue.Priority,
			&queue.TotalTasks,
			&queue.CompletedTasks,
			&queue.PendingTasks,
			&queue.IsActive,
			&queue.CreatedAt,
			&queue.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task queue: %w", err)
		}
		queues = append(queues, queue)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task queues: %w", err)
	}

	return queues, nil
}
