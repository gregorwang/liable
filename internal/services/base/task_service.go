// Package base provides generic service functions for task operations.
package base

import (
	"comment-review-platform/internal/config"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// BaseTaskService 基础任务服务
// 提供通用的任务业务逻辑方法
type BaseTaskService struct {
	Config TaskServiceConfig
	Rdb    *redis.Client
	Ctx    context.Context
}

// NewBaseTaskService 创建基础任务服务
func NewBaseTaskService(cfg TaskServiceConfig, rdb *redis.Client) *BaseTaskService {
	return &BaseTaskService{
		Config: cfg,
		Rdb:    rdb,
		Ctx:    context.Background(),
	}
}

// ValidateClaimCount 验证领取数量
// 返回 nil 表示验证通过，否则返回错误
func (s *BaseTaskService) ValidateClaimCount(count int) error {
	if count < s.Config.ClaimCountMin || count > s.Config.ClaimCountMax {
		return fmt.Errorf("claim count must be between %d and %d",
			s.Config.ClaimCountMin, s.Config.ClaimCountMax)
	}
	return nil
}

// CheckExistingTasks 检查是否有未完成任务
// existingCount: 当前未完成任务数量
// 返回 nil 表示可以领取新任务，否则返回错误
func (s *BaseTaskService) CheckExistingTasks(existingCount int) error {
	if existingCount > 0 {
		return fmt.Errorf("you still have %d uncompleted %s tasks, please complete or return them first",
			existingCount, s.Config.TaskTypeName)
	}
	return nil
}

// TrackClaimedTasks 在 Redis 中追踪已领取的任务
// 设置 claimed set 和 lock key
func (s *BaseTaskService) TrackClaimedTasks(reviewerID int, taskIDs []int) error {
	if s.Rdb == nil {
		return nil // Redis 未配置时跳过
	}

	userClaimedKey := s.GetClaimedKey(reviewerID)
	timeout := s.GetTaskTimeout()

	pipe := s.Rdb.Pipeline()
	for _, taskID := range taskIDs {
		pipe.SAdd(s.Ctx, userClaimedKey, taskID)
		lockKey := s.GetLockKey(taskID)
		pipe.Set(s.Ctx, lockKey, reviewerID, timeout)
	}
	pipe.Expire(s.Ctx, userClaimedKey, timeout)

	_, err := pipe.Exec(s.Ctx)
	if err != nil {
		log.Printf("Redis error when tracking %s tasks: %v", s.Config.TaskTypeName, err)
	}
	return err
}

// CleanupTaskTracking 清理任务追踪
// 从 Redis 中移除 claimed set 和 lock key
func (s *BaseTaskService) CleanupTaskTracking(reviewerID int, taskIDs []int) {
	if s.Rdb == nil {
		return // Redis 未配置时跳过
	}

	userClaimedKey := s.GetClaimedKey(reviewerID)

	pipe := s.Rdb.Pipeline()
	for _, taskID := range taskIDs {
		pipe.SRem(s.Ctx, userClaimedKey, taskID)
		lockKey := s.GetLockKey(taskID)
		pipe.Del(s.Ctx, lockKey)
	}

	_, err := pipe.Exec(s.Ctx)
	if err != nil {
		log.Printf("Redis error when cleaning up %s tasks: %v", s.Config.TaskTypeName, err)
	}
}

// CleanupSingleTask 清理单个任务的追踪
func (s *BaseTaskService) CleanupSingleTask(reviewerID int, taskID int) {
	s.CleanupTaskTracking(reviewerID, []int{taskID})
}

// ValidateReturnCount 验证退回数量
func (s *BaseTaskService) ValidateReturnCount(count int) error {
	if count < 1 || count > 50 {
		return errors.New("return count must be between 1 and 50")
	}
	return nil
}

// GetClaimedKey 获取用户已领取任务的 Redis key
func (s *BaseTaskService) GetClaimedKey(reviewerID int) string {
	return fmt.Sprintf("%s:claimed:%d", s.Config.RedisKeyPrefix, reviewerID)
}

// GetLockKey 获取任务锁的 Redis key
func (s *BaseTaskService) GetLockKey(taskID int) string {
	return fmt.Sprintf("%s:lock:%d", s.Config.RedisKeyPrefix, taskID)
}

// GetTaskTimeout 获取任务超时时间
func (s *BaseTaskService) GetTaskTimeout() time.Duration {
	if config.AppConfig != nil {
		return time.Duration(config.AppConfig.TaskTimeoutMinutes) * time.Minute
	}
	return 30 * time.Minute // 默认 30 分钟
}

// GetTaskTimeoutMinutes 获取任务超时分钟数
func (s *BaseTaskService) GetTaskTimeoutMinutes() int {
	if config.AppConfig != nil {
		return config.AppConfig.TaskTimeoutMinutes
	}
	return 30 // 默认 30 分钟
}

// IsTaskLocked 检查任务是否被锁定
func (s *BaseTaskService) IsTaskLocked(taskID int) (bool, error) {
	if s.Rdb == nil {
		return false, nil
	}

	lockKey := s.GetLockKey(taskID)
	exists, err := s.Rdb.Exists(s.Ctx, lockKey).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// GetTaskLockOwner 获取任务锁的持有者
func (s *BaseTaskService) GetTaskLockOwner(taskID int) (int, error) {
	if s.Rdb == nil {
		return 0, nil
	}

	lockKey := s.GetLockKey(taskID)
	result, err := s.Rdb.Get(s.Ctx, lockKey).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return result, nil
}
