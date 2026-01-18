package base

import (
	"testing"
)

func TestDefaultTaskServiceConfig(t *testing.T) {
	config := DefaultTaskServiceConfig("test_task", "test")

	if config.TaskTypeName != "test_task" {
		t.Errorf("expected TaskTypeName 'test_task', got %q", config.TaskTypeName)
	}

	if config.RedisKeyPrefix != "test" {
		t.Errorf("expected RedisKeyPrefix 'test', got %q", config.RedisKeyPrefix)
	}

	if config.ClaimCountMin != 1 {
		t.Errorf("expected ClaimCountMin 1, got %d", config.ClaimCountMin)
	}

	if config.ClaimCountMax != 50 {
		t.Errorf("expected ClaimCountMax 50, got %d", config.ClaimCountMax)
	}
}

func TestReviewTaskServiceConfig(t *testing.T) {
	config := ReviewTaskServiceConfig()

	if config.TaskTypeName != "review" {
		t.Errorf("expected TaskTypeName 'review', got %q", config.TaskTypeName)
	}

	if config.RedisKeyPrefix != "task" {
		t.Errorf("expected RedisKeyPrefix 'task', got %q", config.RedisKeyPrefix)
	}
}

func TestSecondReviewTaskServiceConfig(t *testing.T) {
	config := SecondReviewTaskServiceConfig()

	if config.TaskTypeName != "second_review" {
		t.Errorf("expected TaskTypeName 'second_review', got %q", config.TaskTypeName)
	}

	if config.RedisKeyPrefix != "second_task" {
		t.Errorf("expected RedisKeyPrefix 'second_task', got %q", config.RedisKeyPrefix)
	}
}

func TestQualityCheckTaskServiceConfig(t *testing.T) {
	config := QualityCheckTaskServiceConfig()

	if config.TaskTypeName != "quality_check" {
		t.Errorf("expected TaskTypeName 'quality_check', got %q", config.TaskTypeName)
	}

	if config.RedisKeyPrefix != "qc_task" {
		t.Errorf("expected RedisKeyPrefix 'qc_task', got %q", config.RedisKeyPrefix)
	}
}

func TestVideoFirstReviewTaskServiceConfig(t *testing.T) {
	config := VideoFirstReviewTaskServiceConfig()

	if config.TaskTypeName != "video_first_review" {
		t.Errorf("expected TaskTypeName 'video_first_review', got %q", config.TaskTypeName)
	}

	if config.RedisKeyPrefix != "video:first" {
		t.Errorf("expected RedisKeyPrefix 'video:first', got %q", config.RedisKeyPrefix)
	}
}

func TestVideoSecondReviewTaskServiceConfig(t *testing.T) {
	config := VideoSecondReviewTaskServiceConfig()

	if config.TaskTypeName != "video_second_review" {
		t.Errorf("expected TaskTypeName 'video_second_review', got %q", config.TaskTypeName)
	}

	if config.RedisKeyPrefix != "video:second" {
		t.Errorf("expected RedisKeyPrefix 'video:second', got %q", config.RedisKeyPrefix)
	}
}

func TestNewBaseTaskService(t *testing.T) {
	config := DefaultTaskServiceConfig("test", "test")
	service := NewBaseTaskService(config, nil)

	if service == nil {
		t.Fatal("expected non-nil service")
	}

	if service.Config.TaskTypeName != "test" {
		t.Errorf("expected TaskTypeName 'test', got %q", service.Config.TaskTypeName)
	}
}

// Property 3: Service 验证逻辑一致性
// For any 领取数量 count，如果 count < 1 或 count > 50，通用 Service 的验证方法应该返回错误
func TestValidateClaimCount_Property(t *testing.T) {
	config := DefaultTaskServiceConfig("test", "test")
	service := NewBaseTaskService(config, nil)

	testCases := []struct {
		count       int
		shouldError bool
	}{
		{-10, true},
		{-1, true},
		{0, true},
		{1, false},  // 边界值：最小有效值
		{25, false}, // 中间值
		{50, false}, // 边界值：最大有效值
		{51, true},
		{100, true},
		{1000, true},
	}

	for _, tc := range testCases {
		err := service.ValidateClaimCount(tc.count)
		if tc.shouldError && err == nil {
			t.Errorf("count=%d: expected error, got nil", tc.count)
		}
		if !tc.shouldError && err != nil {
			t.Errorf("count=%d: expected no error, got %v", tc.count, err)
		}
	}
}

// Property 3: 验证逻辑一致性 - 边界测试
func TestValidateClaimCount_Boundaries(t *testing.T) {
	config := DefaultTaskServiceConfig("test", "test")
	service := NewBaseTaskService(config, nil)

	// 测试边界值
	if err := service.ValidateClaimCount(1); err != nil {
		t.Errorf("count=1 should be valid, got error: %v", err)
	}

	if err := service.ValidateClaimCount(50); err != nil {
		t.Errorf("count=50 should be valid, got error: %v", err)
	}

	if err := service.ValidateClaimCount(0); err == nil {
		t.Error("count=0 should be invalid")
	}

	if err := service.ValidateClaimCount(51); err == nil {
		t.Error("count=51 should be invalid")
	}
}

func TestCheckExistingTasks(t *testing.T) {
	config := DefaultTaskServiceConfig("test", "test")
	service := NewBaseTaskService(config, nil)

	// 没有未完成任务时应该通过
	if err := service.CheckExistingTasks(0); err != nil {
		t.Errorf("expected no error for 0 existing tasks, got %v", err)
	}

	// 有未完成任务时应该返回错误
	if err := service.CheckExistingTasks(1); err == nil {
		t.Error("expected error for 1 existing task")
	}

	if err := service.CheckExistingTasks(5); err == nil {
		t.Error("expected error for 5 existing tasks")
	}
}

func TestValidateReturnCount(t *testing.T) {
	config := DefaultTaskServiceConfig("test", "test")
	service := NewBaseTaskService(config, nil)

	testCases := []struct {
		count       int
		shouldError bool
	}{
		{0, true},
		{1, false},
		{25, false},
		{50, false},
		{51, true},
		{100, true},
	}

	for _, tc := range testCases {
		err := service.ValidateReturnCount(tc.count)
		if tc.shouldError && err == nil {
			t.Errorf("count=%d: expected error, got nil", tc.count)
		}
		if !tc.shouldError && err != nil {
			t.Errorf("count=%d: expected no error, got %v", tc.count, err)
		}
	}
}

// Property 4: Redis 追踪数据完整性 - Key 格式测试
func TestGetClaimedKey(t *testing.T) {
	config := DefaultTaskServiceConfig("test", "task")
	service := NewBaseTaskService(config, nil)

	key := service.GetClaimedKey(123)
	expected := "task:claimed:123"

	if key != expected {
		t.Errorf("expected key %q, got %q", expected, key)
	}
}

func TestGetLockKey(t *testing.T) {
	config := DefaultTaskServiceConfig("test", "task")
	service := NewBaseTaskService(config, nil)

	key := service.GetLockKey(456)
	expected := "task:lock:456"

	if key != expected {
		t.Errorf("expected key %q, got %q", expected, key)
	}
}

// 测试不同任务类型的 Redis key 前缀
func TestRedisKeyPrefixes(t *testing.T) {
	testCases := []struct {
		config         TaskServiceConfig
		reviewerID     int
		taskID         int
		expectedClaim  string
		expectedLock   string
	}{
		{
			ReviewTaskServiceConfig(),
			1, 100,
			"task:claimed:1", "task:lock:100",
		},
		{
			SecondReviewTaskServiceConfig(),
			2, 200,
			"second_task:claimed:2", "second_task:lock:200",
		},
		{
			QualityCheckTaskServiceConfig(),
			3, 300,
			"qc_task:claimed:3", "qc_task:lock:300",
		},
		{
			VideoFirstReviewTaskServiceConfig(),
			4, 400,
			"video:first:claimed:4", "video:first:lock:400",
		},
		{
			VideoSecondReviewTaskServiceConfig(),
			5, 500,
			"video:second:claimed:5", "video:second:lock:500",
		},
	}

	for _, tc := range testCases {
		service := NewBaseTaskService(tc.config, nil)

		claimKey := service.GetClaimedKey(tc.reviewerID)
		if claimKey != tc.expectedClaim {
			t.Errorf("config %s: expected claim key %q, got %q",
				tc.config.TaskTypeName, tc.expectedClaim, claimKey)
		}

		lockKey := service.GetLockKey(tc.taskID)
		if lockKey != tc.expectedLock {
			t.Errorf("config %s: expected lock key %q, got %q",
				tc.config.TaskTypeName, tc.expectedLock, lockKey)
		}
	}
}

// 测试配置一致性
func TestConfigConsistency(t *testing.T) {
	configs := []TaskServiceConfig{
		ReviewTaskServiceConfig(),
		SecondReviewTaskServiceConfig(),
		QualityCheckTaskServiceConfig(),
		VideoFirstReviewTaskServiceConfig(),
		VideoSecondReviewTaskServiceConfig(),
	}

	for _, config := range configs {
		// 验证所有配置使用相同的领取数量范围
		if config.ClaimCountMin != 1 {
			t.Errorf("config %s has inconsistent ClaimCountMin: %d",
				config.TaskTypeName, config.ClaimCountMin)
		}

		if config.ClaimCountMax != 50 {
			t.Errorf("config %s has inconsistent ClaimCountMax: %d",
				config.TaskTypeName, config.ClaimCountMax)
		}

		// 验证 Redis key 前缀不为空
		if config.RedisKeyPrefix == "" {
			t.Errorf("config %s has empty RedisKeyPrefix", config.TaskTypeName)
		}

		// 验证任务类型名称不为空
		if config.TaskTypeName == "" {
			t.Errorf("config has empty TaskTypeName")
		}
	}
}

// 测试 Redis 未配置时的行为
func TestTrackClaimedTasks_NoRedis(t *testing.T) {
	config := DefaultTaskServiceConfig("test", "test")
	service := NewBaseTaskService(config, nil) // nil Redis client

	// 应该不会 panic，直接返回 nil
	err := service.TrackClaimedTasks(1, []int{1, 2, 3})
	if err != nil {
		t.Errorf("expected nil error when Redis is not configured, got %v", err)
	}
}

func TestCleanupTaskTracking_NoRedis(t *testing.T) {
	config := DefaultTaskServiceConfig("test", "test")
	service := NewBaseTaskService(config, nil) // nil Redis client

	// 应该不会 panic
	service.CleanupTaskTracking(1, []int{1, 2, 3})
}

func TestIsTaskLocked_NoRedis(t *testing.T) {
	config := DefaultTaskServiceConfig("test", "test")
	service := NewBaseTaskService(config, nil) // nil Redis client

	locked, err := service.IsTaskLocked(1)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if locked {
		t.Error("expected false when Redis is not configured")
	}
}

func TestGetTaskLockOwner_NoRedis(t *testing.T) {
	config := DefaultTaskServiceConfig("test", "test")
	service := NewBaseTaskService(config, nil) // nil Redis client

	owner, err := service.GetTaskLockOwner(1)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if owner != 0 {
		t.Errorf("expected 0 when Redis is not configured, got %d", owner)
	}
}
