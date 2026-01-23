// Package base provides common service utilities for the comment review platform.
package base

// TaskServiceConfig 任务服务配置
// 用于配置通用 Service 方法的行为
type TaskServiceConfig struct {
	TaskTypeName   string // 任务类型名称，用于日志和错误消息
	RedisKeyPrefix string // Redis key 前缀
	ClaimCountMin  int    // 最小领取数量
	ClaimCountMax  int    // 最大领取数量
}

// DefaultTaskServiceConfig 返回默认配置
func DefaultTaskServiceConfig(taskTypeName, redisPrefix string) TaskServiceConfig {
	return TaskServiceConfig{
		TaskTypeName:   taskTypeName,
		RedisKeyPrefix: redisPrefix,
		ClaimCountMin:  1,
		ClaimCountMax:  50,
	}
}

// ReviewTaskServiceConfig 返回评论审核任务的配置
func ReviewTaskServiceConfig() TaskServiceConfig {
	return DefaultTaskServiceConfig("review", "task")
}

// SecondReviewTaskServiceConfig 返回二审任务的配置
func SecondReviewTaskServiceConfig() TaskServiceConfig {
	return DefaultTaskServiceConfig("second_review", "second_task")
}

// QualityCheckTaskServiceConfig 返回质检任务的配置
func QualityCheckTaskServiceConfig() TaskServiceConfig {
	return DefaultTaskServiceConfig("quality_check", "qc_task")
}

// AIHumanDiffTaskServiceConfig returns AI vs human diff task configuration
func AIHumanDiffTaskServiceConfig() TaskServiceConfig {
	return DefaultTaskServiceConfig("ai_human_diff", "ai_diff_task")
}

// VideoFirstReviewTaskServiceConfig 返回视频一审任务的配置
func VideoFirstReviewTaskServiceConfig() TaskServiceConfig {
	return DefaultTaskServiceConfig("video_first_review", "video:first")
}

// VideoSecondReviewTaskServiceConfig 返回视频二审任务的配置
func VideoSecondReviewTaskServiceConfig() TaskServiceConfig {
	return DefaultTaskServiceConfig("video_second_review", "video:second")
}
