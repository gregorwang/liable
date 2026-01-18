package base

import (
	"testing"
)

func TestDefaultTaskRepoConfig(t *testing.T) {
	config := DefaultTaskRepoConfig("test_tasks")

	if config.TableName != "test_tasks" {
		t.Errorf("expected TableName 'test_tasks', got %q", config.TableName)
	}

	if config.IDColumn != "id" {
		t.Errorf("expected IDColumn 'id', got %q", config.IDColumn)
	}

	if config.StatusColumn != "status" {
		t.Errorf("expected StatusColumn 'status', got %q", config.StatusColumn)
	}

	if config.ReviewerIDColumn != "reviewer_id" {
		t.Errorf("expected ReviewerIDColumn 'reviewer_id', got %q", config.ReviewerIDColumn)
	}

	if config.PendingStatus != "pending" {
		t.Errorf("expected PendingStatus 'pending', got %q", config.PendingStatus)
	}

	if config.InProgressStatus != "in_progress" {
		t.Errorf("expected InProgressStatus 'in_progress', got %q", config.InProgressStatus)
	}

	if config.CompletedStatus != "completed" {
		t.Errorf("expected CompletedStatus 'completed', got %q", config.CompletedStatus)
	}
}

func TestReviewTaskRepoConfig(t *testing.T) {
	config := ReviewTaskRepoConfig()

	if config.TableName != "review_tasks" {
		t.Errorf("expected TableName 'review_tasks', got %q", config.TableName)
	}

	// 验证 SelectColumns 包含 comment_id
	hasCommentID := false
	for _, col := range config.SelectColumns {
		if col == "comment_id" {
			hasCommentID = true
			break
		}
	}
	if !hasCommentID {
		t.Error("expected SelectColumns to contain 'comment_id'")
	}
}

func TestSecondReviewTaskRepoConfig(t *testing.T) {
	config := SecondReviewTaskRepoConfig()

	if config.TableName != "second_review_tasks" {
		t.Errorf("expected TableName 'second_review_tasks', got %q", config.TableName)
	}
}

func TestQualityCheckTaskRepoConfig(t *testing.T) {
	config := QualityCheckTaskRepoConfig()

	if config.TableName != "quality_check_tasks" {
		t.Errorf("expected TableName 'quality_check_tasks', got %q", config.TableName)
	}

	// 验证 SelectColumns 包含 task_id
	hasTaskID := false
	for _, col := range config.SelectColumns {
		if col == "task_id" {
			hasTaskID = true
			break
		}
	}
	if !hasTaskID {
		t.Error("expected SelectColumns to contain 'task_id'")
	}
}

func TestVideoFirstReviewTaskRepoConfig(t *testing.T) {
	config := VideoFirstReviewTaskRepoConfig()

	if config.TableName != "video_first_review_tasks" {
		t.Errorf("expected TableName 'video_first_review_tasks', got %q", config.TableName)
	}

	// 验证 SelectColumns 包含 video_id
	hasVideoID := false
	for _, col := range config.SelectColumns {
		if col == "video_id" {
			hasVideoID = true
			break
		}
	}
	if !hasVideoID {
		t.Error("expected SelectColumns to contain 'video_id'")
	}
}

func TestVideoSecondReviewTaskRepoConfig(t *testing.T) {
	config := VideoSecondReviewTaskRepoConfig()

	if config.TableName != "video_second_review_tasks" {
		t.Errorf("expected TableName 'video_second_review_tasks', got %q", config.TableName)
	}
}

func TestNewBaseTaskRepository(t *testing.T) {
	config := DefaultTaskRepoConfig("test_tasks")
	repo := NewBaseTaskRepository(nil, config)

	if repo == nil {
		t.Fatal("expected non-nil repository")
	}

	if repo.Config.TableName != "test_tasks" {
		t.Errorf("expected TableName 'test_tasks', got %q", repo.Config.TableName)
	}
}

// TestConfigConsistency 验证所有配置的一致性
// Property 2: Repository 事务安全性 - 配置部分
// 确保所有配置使用相同的状态值和列名约定
func TestConfigConsistency(t *testing.T) {
	configs := []TaskRepoConfig{
		ReviewTaskRepoConfig(),
		SecondReviewTaskRepoConfig(),
		QualityCheckTaskRepoConfig(),
		VideoFirstReviewTaskRepoConfig(),
		VideoSecondReviewTaskRepoConfig(),
	}

	for _, config := range configs {
		// 验证状态值一致性
		if config.PendingStatus != "pending" {
			t.Errorf("config for %s has inconsistent PendingStatus: %q",
				config.TableName, config.PendingStatus)
		}

		if config.InProgressStatus != "in_progress" {
			t.Errorf("config for %s has inconsistent InProgressStatus: %q",
				config.TableName, config.InProgressStatus)
		}

		if config.CompletedStatus != "completed" {
			t.Errorf("config for %s has inconsistent CompletedStatus: %q",
				config.TableName, config.CompletedStatus)
		}

		// 验证列名一致性
		if config.IDColumn != "id" {
			t.Errorf("config for %s has inconsistent IDColumn: %q",
				config.TableName, config.IDColumn)
		}

		if config.StatusColumn != "status" {
			t.Errorf("config for %s has inconsistent StatusColumn: %q",
				config.TableName, config.StatusColumn)
		}

		if config.ReviewerIDColumn != "reviewer_id" {
			t.Errorf("config for %s has inconsistent ReviewerIDColumn: %q",
				config.TableName, config.ReviewerIDColumn)
		}

		// 验证 SelectColumns 至少包含 id
		hasID := false
		for _, col := range config.SelectColumns {
			if col == "id" {
				hasID = true
				break
			}
		}
		if !hasID {
			t.Errorf("config for %s SelectColumns should contain 'id'",
				config.TableName)
		}
	}
}
