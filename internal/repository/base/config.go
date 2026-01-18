// Package base provides common repository utilities for the comment review platform.
package base

// TaskRepoConfig 任务仓库配置
// 用于配置通用 Repository 方法的表名和字段名
type TaskRepoConfig struct {
	TableName         string   // 任务表名
	IDColumn          string   // ID 列名
	StatusColumn      string   // 状态列名
	ReviewerIDColumn  string   // 审核员ID列名
	ClaimedAtColumn   string   // 领取时间列名
	CompletedAtColumn string   // 完成时间列名
	CreatedAtColumn   string   // 创建时间列名
	SelectColumns     []string // SELECT 查询的列（用于 ClaimTasks）
	PendingStatus     string   // 待处理状态值
	InProgressStatus  string   // 处理中状态值
	CompletedStatus   string   // 已完成状态值
}

// DefaultTaskRepoConfig 返回默认配置
func DefaultTaskRepoConfig(tableName string) TaskRepoConfig {
	return TaskRepoConfig{
		TableName:         tableName,
		IDColumn:          "id",
		StatusColumn:      "status",
		ReviewerIDColumn:  "reviewer_id",
		ClaimedAtColumn:   "claimed_at",
		CompletedAtColumn: "completed_at",
		CreatedAtColumn:   "created_at",
		SelectColumns:     []string{"id", "created_at"},
		PendingStatus:     "pending",
		InProgressStatus:  "in_progress",
		CompletedStatus:   "completed",
	}
}

// ReviewTaskRepoConfig 返回评论审核任务的配置
func ReviewTaskRepoConfig() TaskRepoConfig {
	config := DefaultTaskRepoConfig("review_tasks")
	config.SelectColumns = []string{"id", "comment_id", "created_at"}
	return config
}

// SecondReviewTaskRepoConfig 返回二审任务的配置
func SecondReviewTaskRepoConfig() TaskRepoConfig {
	config := DefaultTaskRepoConfig("second_review_tasks")
	config.SelectColumns = []string{"id", "comment_id", "created_at"}
	return config
}

// QualityCheckTaskRepoConfig 返回质检任务的配置
func QualityCheckTaskRepoConfig() TaskRepoConfig {
	config := DefaultTaskRepoConfig("quality_check_tasks")
	config.SelectColumns = []string{"id", "task_id", "created_at"}
	return config
}

// VideoFirstReviewTaskRepoConfig 返回视频一审任务的配置
func VideoFirstReviewTaskRepoConfig() TaskRepoConfig {
	config := DefaultTaskRepoConfig("video_first_review_tasks")
	config.SelectColumns = []string{"id", "video_id", "created_at"}
	return config
}

// VideoSecondReviewTaskRepoConfig 返回视频二审任务的配置
func VideoSecondReviewTaskRepoConfig() TaskRepoConfig {
	config := DefaultTaskRepoConfig("video_second_review_tasks")
	config.SelectColumns = []string{"id", "video_id", "created_at"}
	return config
}
