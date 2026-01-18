// Package base provides generic repository functions for task operations.
package base

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// BaseTaskRepository 基础任务仓库
// 提供通用的任务数据库操作方法
type BaseTaskRepository struct {
	DB     *sql.DB
	Config TaskRepoConfig
}

// NewBaseTaskRepository 创建基础任务仓库
func NewBaseTaskRepository(db *sql.DB, config TaskRepoConfig) *BaseTaskRepository {
	return &BaseTaskRepository{DB: db, Config: config}
}

// ClaimTaskIDs 领取任务（仅返回任务ID列表）
// 使用 FOR UPDATE SKIP LOCKED 确保并发安全
func (r *BaseTaskRepository) ClaimTaskIDs(reviewerID int, limit int) ([]int, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 构建 SELECT 查询
	selectQuery := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = $1
		ORDER BY %s ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, r.Config.IDColumn, r.Config.TableName,
		r.Config.StatusColumn, r.Config.CreatedAtColumn)

	rows, err := tx.Query(selectQuery, r.Config.PendingStatus, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taskIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, id)
	}

	if len(taskIDs) == 0 {
		return []int{}, nil
	}

	// 更新任务状态
	updateQuery := fmt.Sprintf(`
		UPDATE %s
		SET %s = $1, %s = $2, %s = $3
		WHERE %s = ANY($4)
	`, r.Config.TableName, r.Config.StatusColumn,
		r.Config.ReviewerIDColumn, r.Config.ClaimedAtColumn, r.Config.IDColumn)

	_, err = tx.Exec(updateQuery, r.Config.InProgressStatus, reviewerID, time.Now(), pq.Array(taskIDs))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return taskIDs, nil
}

// CompleteTask 完成任务
func (r *BaseTaskRepository) CompleteTask(taskID, reviewerID int) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET %s = $1, %s = NOW()
		WHERE %s = $2 AND %s = $3 AND %s = $4
	`, r.Config.TableName, r.Config.StatusColumn, r.Config.CompletedAtColumn,
		r.Config.IDColumn, r.Config.ReviewerIDColumn, r.Config.StatusColumn)

	result, err := r.DB.Exec(query, r.Config.CompletedStatus, taskID, reviewerID, r.Config.InProgressStatus)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ReturnTasks 退回任务
func (r *BaseTaskRepository) ReturnTasks(taskIDs []int, reviewerID int) (int, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET %s = $1, %s = NULL, %s = NULL
		WHERE %s = ANY($2) AND %s = $3 AND %s = $4
	`, r.Config.TableName, r.Config.StatusColumn,
		r.Config.ReviewerIDColumn, r.Config.ClaimedAtColumn,
		r.Config.IDColumn, r.Config.ReviewerIDColumn, r.Config.StatusColumn)

	result, err := r.DB.Exec(query, r.Config.PendingStatus, pq.Array(taskIDs), reviewerID, r.Config.InProgressStatus)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// FindExpiredTaskIDs 查找过期任务ID
func (r *BaseTaskRepository) FindExpiredTaskIDs(timeoutMinutes int) ([]int, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE %s = $1 AND %s < NOW() - INTERVAL '1 minute' * $2
	`, r.Config.IDColumn, r.Config.TableName,
		r.Config.StatusColumn, r.Config.ClaimedAtColumn)

	rows, err := r.DB.Query(query, r.Config.InProgressStatus, timeoutMinutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taskIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, id)
	}
	return taskIDs, nil
}

// ResetTask 重置任务
func (r *BaseTaskRepository) ResetTask(taskID int) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET %s = $1, %s = NULL, %s = NULL
		WHERE %s = $2
	`, r.Config.TableName, r.Config.StatusColumn,
		r.Config.ReviewerIDColumn, r.Config.ClaimedAtColumn, r.Config.IDColumn)

	_, err := r.DB.Exec(query, r.Config.PendingStatus, taskID)
	return err
}

// CountByStatus 按状态统计任务数量
func (r *BaseTaskRepository) CountByStatus(status string) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s = $1`,
		r.Config.TableName, r.Config.StatusColumn)
	var count int
	err := r.DB.QueryRow(query, status).Scan(&count)
	return count, err
}

// CountPending 统计待处理任务数量
func (r *BaseTaskRepository) CountPending() (int, error) {
	return r.CountByStatus(r.Config.PendingStatus)
}

// CountInProgress 统计处理中任务数量
func (r *BaseTaskRepository) CountInProgress() (int, error) {
	return r.CountByStatus(r.Config.InProgressStatus)
}

// CountByReviewer 统计审核员的任务数量
func (r *BaseTaskRepository) CountByReviewer(reviewerID int, status string) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s = $1 AND %s = $2`,
		r.Config.TableName, r.Config.ReviewerIDColumn, r.Config.StatusColumn)
	var count int
	err := r.DB.QueryRow(query, reviewerID, status).Scan(&count)
	return count, err
}

// GetReviewerInProgressCount 获取审核员处理中的任务数量
func (r *BaseTaskRepository) GetReviewerInProgressCount(reviewerID int) (int, error) {
	return r.CountByReviewer(reviewerID, r.Config.InProgressStatus)
}
