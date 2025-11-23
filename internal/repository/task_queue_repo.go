package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"fmt"
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

// ListTaskQueues returns paginated task queues with real-time statistics from unified_queue_stats
func (r *TaskQueueRepository) ListTaskQueues(req models.ListTaskQueuesRequest) ([]models.TaskQueue, int, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	if req.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(queue_name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+req.Search+"%")
		argIndex++
	}

	if req.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total records from the unified view
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM unified_queue_stats %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count task queues: %w", err)
	}

	// Get paginated results from the unified view with real-time statistics
	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf(`
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
		%s
		ORDER BY priority DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query task queues: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan task queue: %w", err)
		}
		queues = append(queues, queue)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating task queues: %w", err)
	}

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
