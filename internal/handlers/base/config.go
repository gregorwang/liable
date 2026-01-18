// Package base provides common handler utilities for the comment review platform.
package base

// TaskHandlerConfig 任务处理器配置
// 用于配置通用 Handler 函数的行为
type TaskHandlerConfig struct {
	TaskTypeName  string // 任务类型名称，用于日志和错误消息
	ClaimCountMin int    // 最小领取数量
	ClaimCountMax int    // 最大领取数量
}

// DefaultTaskHandlerConfig 返回默认配置
func DefaultTaskHandlerConfig(taskTypeName string) TaskHandlerConfig {
	return TaskHandlerConfig{
		TaskTypeName:  taskTypeName,
		ClaimCountMin: 1,
		ClaimCountMax: 50,
	}
}
