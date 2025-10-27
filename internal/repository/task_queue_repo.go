package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"fmt"
)

type TaskQueueRepository struct {
	db *sql.DB
}

func NewTaskQueueRepository() *TaskQueueRepository {
	return &TaskQueueRepository{db: database.DB}
}

// CreateTaskQueue creates a new task queue
func (r *TaskQueueRepository) CreateTaskQueue(req models.CreateTaskQueueRequest, createdByID int) (*models.TaskQueue, error) {
	query := `
		INSERT INTO task_queue (queue_name, description, priority, total_tasks, completed_tasks, is_active, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, $6, $6, NOW(), NOW())
		RETURNING id, queue_name, description, priority, total_tasks, completed_tasks, is_active, created_by, updated_by, created_at, updated_at
	`

	var queue models.TaskQueue
	err := r.db.QueryRow(query,
		req.QueueName,
		req.Description,
		req.Priority,
		req.TotalTasks,
		req.CompletedTasks,
		createdByID,
	).Scan(
		&queue.ID,
		&queue.QueueName,
		&queue.Description,
		&queue.Priority,
		&queue.TotalTasks,
		&queue.CompletedTasks,
		&queue.IsActive,
		&queue.CreatedBy,
		&queue.UpdatedBy,
		&queue.CreatedAt,
		&queue.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	queue.PendingTasks = queue.TotalTasks - queue.CompletedTasks
	return &queue, nil
}

// GetTaskQueueByID retrieves a task queue by ID
func (r *TaskQueueRepository) GetTaskQueueByID(id int) (*models.TaskQueue, error) {
	query := `
		SELECT id, queue_name, description, priority, total_tasks, completed_tasks, is_active, created_by, updated_by, created_at, updated_at
		FROM task_queue
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
		&queue.IsActive,
		&queue.CreatedBy,
		&queue.UpdatedBy,
		&queue.CreatedAt,
		&queue.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	queue.PendingTasks = queue.TotalTasks - queue.CompletedTasks
	return &queue, nil
}

// ListTaskQueues retrieves task queues with pagination and filtering
func (r *TaskQueueRepository) ListTaskQueues(req models.ListTaskQueuesRequest) ([]models.TaskQueue, int, error) {
	// Build base query
	baseQuery := `SELECT id, queue_name, description, priority, total_tasks, completed_tasks, is_active, created_by, updated_by, created_at, updated_at FROM task_queue WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM task_queue WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	// Add search filter
	if req.Search != "" {
		baseQuery += fmt.Sprintf(` AND queue_name ILIKE $%d`, argIndex)
		countQuery += fmt.Sprintf(` AND queue_name ILIKE $%d`, argIndex)
		args = append(args, "%"+req.Search+"%")
		argIndex++
	}

	// Add active filter
	if req.IsActive != nil {
		baseQuery += fmt.Sprintf(` AND is_active = $%d`, argIndex)
		countQuery += fmt.Sprintf(` AND is_active = $%d`, argIndex)
		args = append(args, *req.IsActive)
		argIndex++
	}

	// Get total count
	var total int
	countArgs := args[:len(args)]
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Set defaults for pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Add sorting and pagination
	baseQuery += ` ORDER BY priority DESC, created_at DESC`
	offset := (req.Page - 1) * req.PageSize
	baseQuery += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	args = append(args, req.PageSize, offset)

	// Execute query
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	queues := []models.TaskQueue{}
	for rows.Next() {
		var queue models.TaskQueue
		err := rows.Scan(
			&queue.ID,
			&queue.QueueName,
			&queue.Description,
			&queue.Priority,
			&queue.TotalTasks,
			&queue.CompletedTasks,
			&queue.IsActive,
			&queue.CreatedBy,
			&queue.UpdatedBy,
			&queue.CreatedAt,
			&queue.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		queue.PendingTasks = queue.TotalTasks - queue.CompletedTasks
		queues = append(queues, queue)
	}

	return queues, total, nil
}

// UpdateTaskQueue updates a task queue
func (r *TaskQueueRepository) UpdateTaskQueue(id int, req models.UpdateTaskQueueRequest, updatedByID int) (*models.TaskQueue, error) {
	// Get current queue first
	current, err := r.GetTaskQueueByID(id)
	if err != nil {
		return nil, err
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.QueueName != nil {
		updates = append(updates, fmt.Sprintf(`queue_name = $%d`, argIndex))
		args = append(args, *req.QueueName)
		argIndex++
	}

	if req.Description != nil {
		updates = append(updates, fmt.Sprintf(`description = $%d`, argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	if req.Priority != nil {
		updates = append(updates, fmt.Sprintf(`priority = $%d`, argIndex))
		args = append(args, *req.Priority)
		argIndex++
	}

	if req.TotalTasks != nil {
		updates = append(updates, fmt.Sprintf(`total_tasks = $%d`, argIndex))
		args = append(args, *req.TotalTasks)
		argIndex++
	}

	if req.CompletedTasks != nil {
		updates = append(updates, fmt.Sprintf(`completed_tasks = $%d`, argIndex))
		args = append(args, *req.CompletedTasks)
		argIndex++
	}

	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf(`is_active = $%d`, argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	// Add updated_by and updated_at
	updates = append(updates, fmt.Sprintf(`updated_by = $%d`, argIndex))
	args = append(args, updatedByID)
	argIndex++

	updates = append(updates, `updated_at = NOW()`)

	// Add ID to args
	args = append(args, id)

	if len(updates) == 0 {
		return current, nil
	}

	query := `UPDATE task_queue SET `
	for i, update := range updates {
		if i > 0 {
			query += `, `
		}
		query += update
	}
	query += fmt.Sprintf(` WHERE id = $%d RETURNING id, queue_name, description, priority, total_tasks, completed_tasks, is_active, created_by, updated_by, created_at, updated_at`, argIndex)

	var queue models.TaskQueue
	err = r.db.QueryRow(query, args...).Scan(
		&queue.ID,
		&queue.QueueName,
		&queue.Description,
		&queue.Priority,
		&queue.TotalTasks,
		&queue.CompletedTasks,
		&queue.IsActive,
		&queue.CreatedBy,
		&queue.UpdatedBy,
		&queue.CreatedAt,
		&queue.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	queue.PendingTasks = queue.TotalTasks - queue.CompletedTasks
	return &queue, nil
}

// DeleteTaskQueue soft deletes a task queue (sets is_active to false)
func (r *TaskQueueRepository) DeleteTaskQueue(id int) error {
	query := `UPDATE task_queue SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// GetAllTaskQueues retrieves all active task queues
func (r *TaskQueueRepository) GetAllTaskQueues() ([]models.TaskQueue, error) {
	query := `
		SELECT id, queue_name, description, priority, total_tasks, completed_tasks, is_active, created_by, updated_by, created_at, updated_at
		FROM task_queue
		WHERE is_active = true
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	queues := []models.TaskQueue{}
	for rows.Next() {
		var queue models.TaskQueue
		err := rows.Scan(
			&queue.ID,
			&queue.QueueName,
			&queue.Description,
			&queue.Priority,
			&queue.TotalTasks,
			&queue.CompletedTasks,
			&queue.IsActive,
			&queue.CreatedBy,
			&queue.UpdatedBy,
			&queue.CreatedAt,
			&queue.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		queue.PendingTasks = queue.TotalTasks - queue.CompletedTasks
		queues = append(queues, queue)
	}

	return queues, nil
}
