
# 评论二审系统实现方案

## 1. 数据库设计

### 1.1 创建二审任务表和结果表

使用Supabase MCP创建新的数据库表：

**second_review_tasks 表**（二审任务表）

- `id`: SERIAL PRIMARY KEY
- `first_review_result_id`: INTEGER NOT NULL (关联 review_results.id)
- `comment_id`: BIGINT NOT NULL (冗余字段，便于查询)
- `reviewer_id`: INTEGER (关联 users.id)
- `status`: VARCHAR(20) ('pending', 'in_progress', 'completed')
- `claimed_at`: TIMESTAMP
- `completed_at`: TIMESTAMP
- `created_at`: TIMESTAMP DEFAULT NOW()

**second_review_results 表**（二审结果表）

- `id`: SERIAL PRIMARY KEY
- `second_task_id`: INTEGER NOT NULL (关联 second_review_tasks.id)
- `reviewer_id`: INTEGER NOT NULL (关联 users.id)
- `is_approved`: BOOLEAN NOT NULL
- `tags`: TEXT[]
- `reason`: TEXT
- `created_at`: TIMESTAMP DEFAULT NOW()

**索引创建**：

- `idx_second_review_tasks_status` ON second_review_tasks(status)
- `idx_second_review_tasks_reviewer` ON second_review_tasks(reviewer_id)
- `idx_second_review_results_task` ON second_review_results(second_task_id)

## 2. 后端开发

### 2.1 更新数据模型 (`internal/models/models.go`)

添加新的模型结构：

```go
type SecondReviewTask struct {
    ID                  int
    FirstReviewResultID int
    CommentID           int64
    ReviewerID          *int
    Status              string
    ClaimedAt           *time.Time
    CompletedAt         *time.Time
    CreatedAt           time.Time
    Comment             *Comment
    FirstReviewResult   *ReviewResult
}

type SecondReviewResult struct {
    ID            int
    SecondTaskID  int
    ReviewerID    int
    IsApproved    bool
    Tags          []string
    Reason        string
    CreatedAt     time.Time
}

// Request/Response DTOs
type ClaimSecondReviewTasksRequest struct {
    Count int `json:"count" binding:"required,min=1,max=50"`
}

type SubmitSecondReviewRequest struct {
    TaskID     int      `json:"task_id" binding:"required"`
    IsApproved bool     `json:"is_approved"`
    Tags       []string `json:"tags"`
    Reason     string   `json:"reason"`
}
```

### 2.2 创建二审仓储层 (`internal/repository/second_review_repo.go`)

实现数据访问方法：

- `CreateSecondReviewTask(firstReviewResultID, commentID int64) error`
- `ClaimSecondReviewTasks(reviewerID, limit int) ([]SecondReviewTask, error)`
- `GetMySecondReviewTasks(reviewerID int) ([]SecondReviewTask, error)`
- `CompleteSecondReviewTask(taskID, reviewerID int) error`
- `CreateSecondReviewResult(result *SecondReviewResult) error`
- `ReturnSecondReviewTasks(taskIDs []int, reviewerID int) (int, error)`

### 2.3 更新任务服务 (`internal/services/task_service.go`)

修改 `SubmitReview` 方法，在一审结果为不通过时：

1. 保存一审结果到 `review_results` 表
2. 创建二审任务到 `second_review_tasks` 表
3. 将comment_id推送到Redis二审队列 `review:queue:second`

### 2.4 创建二审服务 (`internal/services/second_review_service.go`)

实现完整的二审业务逻辑：

- `ClaimSecondReviewTasks(reviewerID, count int) ([]SecondReviewTask, error)`
  - 从数据库查询pending的二审任务
  - 联表查询评论和一审结果信息
  - Redis加锁（使用 `second_task:lock:{taskID}`）
- `GetMySecondReviewTasks(reviewerID int) ([]SecondReviewTask, error)`
- `SubmitSecondReview(reviewerID int, req SubmitSecondReviewRequest) error`
  - 完成二审任务
  - 保存二审结果到 `second_review_results` 表
  - 清理Redis锁
- `ReturnSecondReviewTasks(reviewerID int, taskIDs []int) (int, error)`

### 2.5 创建二审处理器 (`internal/handlers/second_review_handler.go`)

实现HTTP处理器：

- `ClaimSecondReviewTasks(c *gin.Context)`
- `GetMySecondReviewTasks(c *gin.Context)`
- `SubmitSecondReview(c *gin.Context)`
- `SubmitBatchSecondReviews(c *gin.Context)`
- `ReturnSecondReviewTasks(c *gin.Context)`

### 2.6 添加路由 (`cmd/api/main.go`)

在 `/api/tasks` 组下添加二审端点：

```go
tasks.POST("/second-review/claim", secondReviewHandler.ClaimSecondReviewTasks)
tasks.GET("/second-review/my", secondReviewHandler.GetMySecondReviewTasks)
tasks.POST("/second-review/submit", secondReviewHandler.SubmitSecondReview)
tasks.POST("/second-review/submit-batch", secondReviewHandler.SubmitBatchSecondReviews)
tasks.POST("/second-review/return", secondReviewHandler.ReturnSecondReviewTasks)
```

## 3. Redis队列管理

### 3.1 Redis Key设计

- 二审队列：`review:queue:second` (List类型)
- 二审任务锁：`second_task:lock:{taskID}`
- 审核员已领取：`second_task:claimed:{reviewerID}`

### 3.2 队列操作

- 一审不通过时：`LPUSH review:queue:second {comment_id}`
- 领取二审任务：从数据库查询（不从Redis队列取）
- 任务加锁：`SET second_task:lock:{taskID} {reviewerID} EX {timeout}`

## 4. 数据流程

```
一审提交(is_approved=false)
  ↓
保存到 review_results 表
  ↓
创建记录到 second_review_tasks 表(status='pending')
  ↓
推送到 Redis 队列 review:queue:second
  ↓
审核员调用 /api/tasks/second-review/claim
  ↓
从 second_review_tasks 查询 pending 任务
  ↓
联表查询 comment + review_results
  ↓
Redis加锁 second_task:lock:{taskID}
  ↓
审核员审核并提交
  ↓
保存到 second_review_results 表
  ↓
清理Redis锁
  ↓
完成
```

## 5. 测试验证

### 5.1 数据库验证

- 确认表结构创建成功
- 测试外键约束

### 5.2 API测试

- 一审提交不通过评论，验证二审任务生成
- 领取二审任务，验证联表查询结果
- 提交二审结果，验证数据保存
- 测试任务返回功能

### 5.3 Redis验证

- 验证任务锁创建和过期
- 验证队列推送和统计