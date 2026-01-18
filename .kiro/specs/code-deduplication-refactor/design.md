# Design Document: Code Deduplication Refactor

## Overview

本设计文档描述了评论审核平台代码去重重构的技术方案。通过引入抽象层、工厂模式和通用组件，消除后端 Handler/Repository/Service 层以及前端 Dashboard/API 层的重复代码。

## Architecture

### 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend (Vue.js)                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐                     │
│  │ GenericReview   │    │ createTaskApi() │                     │
│  │ Dashboard.vue   │───▶│ Factory         │                     │
│  └─────────────────┘    └─────────────────┘                     │
│           │                      │                               │
│           ▼                      ▼                               │
│  ┌─────────────────────────────────────────┐                    │
│  │ Specific Dashboards (via props/slots)   │                    │
│  │ - VideoFirstReview                      │                    │
│  │ - VideoSecondReview                     │                    │
│  │ - QualityCheck                          │                    │
│  └─────────────────────────────────────────┘                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Backend (Go)                             │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐                                            │
│  │ Generic Handler │◀── TaskHandlerConfig                       │
│  │ Functions       │                                            │
│  └────────┬────────┘                                            │
│           │                                                      │
│           ▼                                                      │
│  ┌─────────────────┐                                            │
│  │ Base Task       │◀── TaskServiceConfig                       │
│  │ Service         │                                            │
│  └────────┬────────┘                                            │
│           │                                                      │
│           ▼                                                      │
│  ┌─────────────────┐                                            │
│  │ Base Task       │◀── TaskRepoConfig                          │
│  │ Repository      │                                            │
│  └─────────────────┘                                            │
└─────────────────────────────────────────────────────────────────┘
```


## Components and Interfaces

### 1. 后端 Handler 层抽象

#### 1.1 TaskHandlerConfig 接口

```go
// internal/handlers/base/config.go
package base

import "github.com/gin-gonic/gin"

// TaskService 定义任务服务接口
type TaskService[T any, R any] interface {
    ClaimTasks(reviewerID int, count int) ([]T, error)
    GetMyTasks(reviewerID int) ([]T, error)
    SubmitReview(reviewerID int, req R) error
    SubmitBatchReviews(reviewerID int, reviews []R) error
    ReturnTasks(reviewerID int, taskIDs []int) (int, error)
}

// TaskHandlerConfig 任务处理器配置
type TaskHandlerConfig struct {
    TaskTypeName    string // 任务类型名称，用于日志和错误消息
    ClaimCountMin   int    // 最小领取数量
    ClaimCountMax   int    // 最大领取数量
}

// DefaultTaskHandlerConfig 返回默认配置
func DefaultTaskHandlerConfig(taskTypeName string) TaskHandlerConfig {
    return TaskHandlerConfig{
        TaskTypeName:  taskTypeName,
        ClaimCountMin: 1,
        ClaimCountMax: 50,
    }
}
```

#### 1.2 通用 Handler 函数

```go
// internal/handlers/base/handlers.go
package base

import (
    "comment-review-platform/internal/middleware"
    "net/http"
    "github.com/gin-gonic/gin"
)

// HandleClaimTasks 通用任务领取处理
func HandleClaimTasks[T any](
    c *gin.Context,
    service interface{ ClaimTasks(int, int) ([]T, error) },
    config TaskHandlerConfig,
) {
    reviewerID := middleware.GetUserID(c)
    
    var req struct {
        Count int `json:"count" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
        return
    }
    
    tasks, err := service.ClaimTasks(reviewerID, req.Count)
    if err != nil {
        RespondError(c, http.StatusBadRequest, "CLAIM_FAILED", err.Error())
        return
    }
    
    RespondSuccess(c, gin.H{"tasks": tasks, "count": len(tasks)})
}

// HandleGetMyTasks 通用获取我的任务处理
func HandleGetMyTasks[T any](
    c *gin.Context,
    service interface{ GetMyTasks(int) ([]T, error) },
) {
    reviewerID := middleware.GetUserID(c)
    
    tasks, err := service.GetMyTasks(reviewerID)
    if err != nil {
        RespondError(c, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
        return
    }
    
    RespondSuccess(c, gin.H{"tasks": tasks, "count": len(tasks)})
}

// HandleSubmitReview 通用提交审核处理
func HandleSubmitReview[R any](
    c *gin.Context,
    service interface{ SubmitReview(int, R) error },
    successMsg string,
) {
    reviewerID := middleware.GetUserID(c)
    
    var req R
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
        return
    }
    
    if err := service.SubmitReview(reviewerID, req); err != nil {
        RespondError(c, http.StatusBadRequest, "SUBMIT_FAILED", err.Error())
        return
    }
    
    RespondSuccess(c, gin.H{"message": successMsg})
}

// HandleReturnTasks 通用退回任务处理
func HandleReturnTasks(
    c *gin.Context,
    service interface{ ReturnTasks(int, []int) (int, error) },
    successMsg string,
) {
    reviewerID := middleware.GetUserID(c)
    
    var req struct {
        TaskIDs []int `json:"task_ids" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
        return
    }
    
    count, err := service.ReturnTasks(reviewerID, req.TaskIDs)
    if err != nil {
        RespondError(c, http.StatusBadRequest, "RETURN_FAILED", err.Error())
        return
    }
    
    RespondSuccess(c, gin.H{"message": successMsg, "count": count})
}
```


#### 1.3 统一错误响应

```go
// internal/handlers/base/response.go
package base

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
    Error string `json:"error"`
    Code  string `json:"code"`
}

// SuccessResponse 统一成功响应结构
type SuccessResponse struct {
    Data interface{} `json:"data,omitempty"`
}

// RespondError 返回错误响应
func RespondError(c *gin.Context, status int, code string, message string) {
    c.JSON(status, ErrorResponse{
        Error: message,
        Code:  code,
    })
}

// RespondSuccess 返回成功响应
func RespondSuccess(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, data)
}

// RespondBadRequest 返回400错误
func RespondBadRequest(c *gin.Context, code string, message string) {
    RespondError(c, http.StatusBadRequest, code, message)
}

// RespondInternalError 返回500错误
func RespondInternalError(c *gin.Context, code string, message string) {
    RespondError(c, http.StatusInternalServerError, code, message)
}
```

### 2. 后端 Repository 层抽象

#### 2.1 TaskRepoConfig 配置

```go
// internal/repository/base/config.go
package base

// TaskRepoConfig 任务仓库配置
type TaskRepoConfig struct {
    TableName           string   // 任务表名
    IDColumn            string   // ID 列名
    StatusColumn        string   // 状态列名
    ReviewerIDColumn    string   // 审核员ID列名
    ClaimedAtColumn     string   // 领取时间列名
    CompletedAtColumn   string   // 完成时间列名
    CreatedAtColumn     string   // 创建时间列名
    SelectColumns       []string // SELECT 查询的列
    PendingStatus       string   // 待处理状态值
    InProgressStatus    string   // 处理中状态值
    CompletedStatus     string   // 已完成状态值
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
```

#### 2.2 BaseTaskRepository

```go
// internal/repository/base/task_repo.go
package base

import (
    "database/sql"
    "fmt"
    "strings"
    "time"
    "github.com/lib/pq"
)

// BaseTaskRepository 基础任务仓库
type BaseTaskRepository struct {
    db     *sql.DB
    config TaskRepoConfig
}

// NewBaseTaskRepository 创建基础任务仓库
func NewBaseTaskRepository(db *sql.DB, config TaskRepoConfig) *BaseTaskRepository {
    return &BaseTaskRepository{db: db, config: config}
}

// ClaimTasks 领取任务（通用实现）
func (r *BaseTaskRepository) ClaimTasks(reviewerID int, limit int) ([]int, error) {
    tx, err := r.db.Begin()
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
    `, r.config.IDColumn, r.config.TableName, 
       r.config.StatusColumn, r.config.CreatedAtColumn)

    rows, err := tx.Query(selectQuery, r.config.PendingStatus, limit)
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
    `, r.config.TableName, r.config.StatusColumn,
       r.config.ReviewerIDColumn, r.config.ClaimedAtColumn, r.config.IDColumn)

    _, err = tx.Exec(updateQuery, r.config.InProgressStatus, reviewerID, time.Now(), pq.Array(taskIDs))
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
    `, r.config.TableName, r.config.StatusColumn, r.config.CompletedAtColumn,
       r.config.IDColumn, r.config.ReviewerIDColumn, r.config.StatusColumn)

    result, err := r.db.Exec(query, r.config.CompletedStatus, taskID, reviewerID, r.config.InProgressStatus)
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
    `, r.config.TableName, r.config.StatusColumn,
       r.config.ReviewerIDColumn, r.config.ClaimedAtColumn,
       r.config.IDColumn, r.config.ReviewerIDColumn, r.config.StatusColumn)

    result, err := r.db.Exec(query, r.config.PendingStatus, pq.Array(taskIDs), reviewerID, r.config.InProgressStatus)
    if err != nil {
        return 0, err
    }

    rowsAffected, _ := result.RowsAffected()
    return int(rowsAffected), nil
}

// FindExpiredTasks 查找过期任务
func (r *BaseTaskRepository) FindExpiredTasks(timeoutMinutes int) ([]int, error) {
    query := fmt.Sprintf(`
        SELECT %s FROM %s
        WHERE %s = $1 AND %s < NOW() - INTERVAL '1 minute' * $2
    `, r.config.IDColumn, r.config.TableName,
       r.config.StatusColumn, r.config.ClaimedAtColumn)

    rows, err := r.db.Query(query, r.config.InProgressStatus, timeoutMinutes)
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
    `, r.config.TableName, r.config.StatusColumn,
       r.config.ReviewerIDColumn, r.config.ClaimedAtColumn, r.config.IDColumn)

    _, err := r.db.Exec(query, r.config.PendingStatus, taskID)
    return err
}
```


### 3. 后端 Service 层抽象

#### 3.1 TaskServiceConfig 配置

```go
// internal/services/base/config.go
package base

// TaskServiceConfig 任务服务配置
type TaskServiceConfig struct {
    TaskTypeName     string // 任务类型名称
    RedisKeyPrefix   string // Redis key 前缀
    ClaimCountMin    int    // 最小领取数量
    ClaimCountMax    int    // 最大领取数量
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
```

#### 3.2 BaseTaskService

```go
// internal/services/base/task_service.go
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

// TaskRepository 任务仓库接口
type TaskRepository interface {
    ClaimTasks(reviewerID int, limit int) ([]int, error)
    GetMyTasks(reviewerID int) (interface{}, error)
    CompleteTask(taskID, reviewerID int) error
    ReturnTasks(taskIDs []int, reviewerID int) (int, error)
    FindExpiredTasks(timeoutMinutes int) ([]int, error)
    ResetTask(taskID int) error
}

// BaseTaskService 基础任务服务
type BaseTaskService struct {
    config TaskServiceConfig
    rdb    *redis.Client
    ctx    context.Context
}

// NewBaseTaskService 创建基础任务服务
func NewBaseTaskService(cfg TaskServiceConfig, rdb *redis.Client) *BaseTaskService {
    return &BaseTaskService{
        config: cfg,
        rdb:    rdb,
        ctx:    context.Background(),
    }
}

// ValidateClaimCount 验证领取数量
func (s *BaseTaskService) ValidateClaimCount(count int) error {
    if count < s.config.ClaimCountMin || count > s.config.ClaimCountMax {
        return fmt.Errorf("claim count must be between %d and %d", 
            s.config.ClaimCountMin, s.config.ClaimCountMax)
    }
    return nil
}

// CheckExistingTasks 检查是否有未完成任务
func (s *BaseTaskService) CheckExistingTasks(existingCount int) error {
    if existingCount > 0 {
        return fmt.Errorf("you still have %d uncompleted %s tasks, please complete or return them first",
            existingCount, s.config.TaskTypeName)
    }
    return nil
}

// TrackClaimedTasks 在 Redis 中追踪已领取的任务
func (s *BaseTaskService) TrackClaimedTasks(reviewerID int, taskIDs []int) error {
    userClaimedKey := fmt.Sprintf("%s:claimed:%d", s.config.RedisKeyPrefix, reviewerID)
    timeout := time.Duration(config.AppConfig.TaskTimeoutMinutes) * time.Minute

    pipe := s.rdb.Pipeline()
    for _, taskID := range taskIDs {
        pipe.SAdd(s.ctx, userClaimedKey, taskID)
        lockKey := fmt.Sprintf("%s:lock:%d", s.config.RedisKeyPrefix, taskID)
        pipe.Set(s.ctx, lockKey, reviewerID, timeout)
    }
    pipe.Expire(s.ctx, userClaimedKey, timeout)

    _, err := pipe.Exec(s.ctx)
    if err != nil {
        log.Printf("Redis error when tracking %s tasks: %v", s.config.TaskTypeName, err)
    }
    return err
}

// CleanupTaskTracking 清理任务追踪
func (s *BaseTaskService) CleanupTaskTracking(reviewerID int, taskIDs []int) {
    userClaimedKey := fmt.Sprintf("%s:claimed:%d", s.config.RedisKeyPrefix, reviewerID)
    
    pipe := s.rdb.Pipeline()
    for _, taskID := range taskIDs {
        pipe.SRem(s.ctx, userClaimedKey, taskID)
        lockKey := fmt.Sprintf("%s:lock:%d", s.config.RedisKeyPrefix, taskID)
        pipe.Del(s.ctx, lockKey)
    }

    _, err := pipe.Exec(s.ctx)
    if err != nil {
        log.Printf("Redis error when cleaning up %s tasks: %v", s.config.TaskTypeName, err)
    }
}

// ValidateReturnCount 验证退回数量
func (s *BaseTaskService) ValidateReturnCount(count int) error {
    if count < 1 || count > 50 {
        return errors.New("return count must be between 1 and 50")
    }
    return nil
}
```

### 4. 前端 API 工厂

#### 4.1 createTaskApi 工厂函数

```typescript
// frontend/src/api/taskApiFactory.ts
import request from './request'

export interface TaskApiConfig {
  basePath: string  // API 路径前缀，如 '/tasks/video-first-review'
}

export interface ClaimRequest {
  count: number
}

export interface ReturnRequest {
  task_ids: number[]
}

export interface TaskApiMethods<TTask, TSubmitReq> {
  claim: (data: ClaimRequest) => Promise<{ tasks: TTask[]; count: number }>
  getMyTasks: () => Promise<{ tasks: TTask[]; count: number }>
  submit: (data: TSubmitReq) => Promise<{ message: string }>
  submitBatch: (data: { reviews: TSubmitReq[] }) => Promise<{ message: string; count: number }>
  returnTasks: (data: ReturnRequest) => Promise<{ message: string; count: number }>
}

/**
 * 创建任务 API 工厂函数
 * @param config API 配置
 * @returns 包含所有任务操作方法的 API 对象
 */
export function createTaskApi<TTask, TSubmitReq>(
  config: TaskApiConfig
): TaskApiMethods<TTask, TSubmitReq> {
  const { basePath } = config

  return {
    claim: (data: ClaimRequest) => 
      request.post(`${basePath}/claim`, data),
    
    getMyTasks: () => 
      request.get(`${basePath}/my`),
    
    submit: (data: TSubmitReq) => 
      request.post(`${basePath}/submit`, data),
    
    submitBatch: (data: { reviews: TSubmitReq[] }) => 
      request.post(`${basePath}/submit-batch`, data),
    
    returnTasks: (data: ReturnRequest) => 
      request.post(`${basePath}/return`, data),
  }
}

// 使用示例
// const videoFirstReviewApi = createTaskApi<VideoFirstReviewTask, SubmitVideoFirstReviewRequest>({
//   basePath: '/tasks/video-first-review'
// })
```


### 5. 前端通用 Dashboard 组件

#### 5.1 GenericReviewDashboard 组件

```vue
<!-- frontend/src/components/GenericReviewDashboard.vue -->
<template>
  <div class="review-dashboard">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <div class="header-left">
            <h2>{{ title }}</h2>
            <el-button type="primary" link @click="$emit('search')" style="margin-left: 20px">
              <el-icon><Search /></el-icon>
              搜索审核记录
            </el-button>
          </div>
          <div class="user-info">
            <span>欢迎，{{ username }}</span>
            <el-button @click="$emit('logout')" text>退出</el-button>
          </div>
        </div>
      </el-header>
      
      <el-main class="main-content">
        <!-- 统计栏 -->
        <div class="stats-bar">
          <el-card v-for="stat in stats" :key="stat.label" shadow="hover">
            <div class="stat-item">
              <span class="stat-label">{{ stat.label }}</span>
              <span class="stat-value">{{ stat.value }}</span>
            </div>
          </el-card>
        </div>
        
        <!-- 操作栏 -->
        <div class="actions-bar">
          <div class="claim-section">
            <el-input-number v-model="localClaimCount" :min="1" :max="50" size="large" style="width: 120px" />
            <el-button type="primary" size="large" :loading="claimLoading" @click="handleClaim">
              {{ claimButtonText }}
            </el-button>
          </div>
          
          <div class="return-section">
            <el-input-number v-model="localReturnCount" :min="1" :max="Math.max(1, tasks.length)" 
              size="large" style="width: 120px" :disabled="tasks.length === 0" />
            <el-button type="warning" size="large" :disabled="tasks.length === 0" @click="handleReturn">
              退单
            </el-button>
          </div>
          
          <el-button size="large" :disabled="selectedCount === 0" @click="$emit('batch-submit')">
            批量提交（{{ selectedCount }}条）
          </el-button>
          
          <el-button size="large" @click="$emit('refresh')">
            刷新任务列表
          </el-button>
        </div>
        
        <!-- 空状态 -->
        <div v-if="tasks.length === 0" class="empty-state">
          <el-empty :description="emptyText" />
        </div>
        
        <!-- 任务列表 - 通过 slot 自定义渲染 -->
        <div v-else class="tasks-container">
          <slot name="task-list" :tasks="tasks"></slot>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'

interface StatItem {
  label: string
  value: string | number
}

const props = defineProps<{
  title: string
  username: string
  stats: StatItem[]
  tasks: any[]
  claimLoading: boolean
  selectedCount: number
  claimButtonText?: string
  emptyText?: string
  initialClaimCount?: number
}>()

const emit = defineEmits<{
  (e: 'claim', count: number): void
  (e: 'return', count: number): void
  (e: 'batch-submit'): void
  (e: 'refresh'): void
  (e: 'search'): void
  (e: 'logout'): void
}>()

const localClaimCount = ref(props.initialClaimCount || 5)
const localReturnCount = ref(1)

watch(() => props.tasks.length, (newLen) => {
  localReturnCount.value = Math.min(localReturnCount.value, Math.max(1, newLen))
})

const handleClaim = () => emit('claim', localClaimCount.value)
const handleReturn = () => emit('return', localReturnCount.value)
</script>
```

### 6. 中间件优化

#### 6.1 通用限流器

```go
// internal/middleware/rate_limit_v2.go
package middleware

import (
    "fmt"
    "net/http"
    "sync"
    "time"
    "github.com/gin-gonic/gin"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
    Limit     int           // 限制次数
    Window    time.Duration // 时间窗口
    KeyFunc   func(*gin.Context) string // 生成限流 key 的函数
    ErrorMsg  string        // 错误消息
}

// IPKeyFunc 基于 IP 的 key 生成函数
func IPKeyFunc(c *gin.Context) string {
    return c.ClientIP()
}

// EndpointKeyFunc 基于 IP + 端点的 key 生成函数
func EndpointKeyFunc(c *gin.Context) string {
    return c.ClientIP() + ":" + c.FullPath()
}

// UserKeyFunc 基于用户ID的 key 生成函数
func UserKeyFunc(c *gin.Context) string {
    userID := GetUserID(c)
    if userID == 0 {
        return "" // 未认证用户不限流
    }
    return fmt.Sprintf("user:%d", userID)
}

// CreateRateLimiter 创建通用限流中间件
func CreateRateLimiter(config RateLimitConfig) gin.HandlerFunc {
    limiter := make(map[string][]time.Time)
    var mutex sync.Mutex

    return func(c *gin.Context) {
        key := config.KeyFunc(c)
        if key == "" {
            c.Next()
            return
        }

        now := time.Now()
        cutoff := now.Add(-config.Window)

        mutex.Lock()
        
        // 清理过期记录
        var validTimes []time.Time
        for _, t := range limiter[key] {
            if t.After(cutoff) {
                validTimes = append(validTimes, t)
            }
        }
        limiter[key] = validTimes

        // 检查限制
        if len(limiter[key]) >= config.Limit {
            mutex.Unlock()
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error":       config.ErrorMsg,
                "retry_after": config.Window.String(),
            })
            c.Abort()
            return
        }

        // 记录请求
        limiter[key] = append(limiter[key], now)
        mutex.Unlock()
        
        c.Next()
    }
}

// 预定义的限流器
func GlobalRateLimiterV2() gin.HandlerFunc {
    return CreateRateLimiter(RateLimitConfig{
        Limit:    100,
        Window:   time.Second,
        KeyFunc:  IPKeyFunc,
        ErrorMsg: "Rate limit exceeded",
    })
}

func EndpointRateLimiterV2(limit int, window time.Duration) gin.HandlerFunc {
    return CreateRateLimiter(RateLimitConfig{
        Limit:    limit,
        Window:   window,
        KeyFunc:  EndpointKeyFunc,
        ErrorMsg: "Too many requests",
    })
}

func UserRateLimiterV2(limit int, window time.Duration) gin.HandlerFunc {
    return CreateRateLimiter(RateLimitConfig{
        Limit:    limit,
        Window:   window,
        KeyFunc:  UserKeyFunc,
        ErrorMsg: "User rate limit exceeded",
    })
}
```


## Data Models

### 后端配置模型

```go
// 任务类型枚举
const (
    TaskTypeFirstReview       = "first_review"
    TaskTypeSecondReview      = "second_review"
    TaskTypeQualityCheck      = "quality_check"
    TaskTypeVideoFirstReview  = "video_first_review"
    TaskTypeVideoSecondReview = "video_second_review"
)

// 预定义的任务配置
var TaskConfigs = map[string]struct {
    TableName      string
    RedisPrefix    string
    SuccessMessage string
}{
    TaskTypeFirstReview: {
        TableName:      "review_tasks",
        RedisPrefix:    "task",
        SuccessMessage: "Review submitted successfully",
    },
    TaskTypeSecondReview: {
        TableName:      "second_review_tasks",
        RedisPrefix:    "second_task",
        SuccessMessage: "Second review submitted successfully",
    },
    TaskTypeQualityCheck: {
        TableName:      "quality_check_tasks",
        RedisPrefix:    "qc_task",
        SuccessMessage: "Quality check submitted successfully",
    },
    TaskTypeVideoFirstReview: {
        TableName:      "video_first_review_tasks",
        RedisPrefix:    "video:first",
        SuccessMessage: "Video first review submitted successfully",
    },
    TaskTypeVideoSecondReview: {
        TableName:      "video_second_review_tasks",
        RedisPrefix:    "video:second",
        SuccessMessage: "Video second review submitted successfully",
    },
}
```

### 前端配置模型

```typescript
// frontend/src/config/dashboardConfigs.ts
export interface DashboardConfig {
  title: string
  apiBasePath: string
  claimButtonText: string
  emptyText: string
  statsConfig: {
    showPendingTasks: boolean
    showTodayCompleted: boolean
    showTotalCompleted?: boolean
    showPassRate?: boolean
  }
}

export const dashboardConfigs: Record<string, DashboardConfig> = {
  videoFirstReview: {
    title: '抖音短视频一审工作台',
    apiBasePath: '/tasks/video-first-review',
    claimButtonText: '领取新任务',
    emptyText: '暂无待审核任务，点击「领取新任务」开始工作',
    statsConfig: {
      showPendingTasks: true,
      showTodayCompleted: true,
    },
  },
  videoSecondReview: {
    title: '抖音短视频二审工作台',
    apiBasePath: '/tasks/video-second-review',
    claimButtonText: '领取新任务',
    emptyText: '暂无待审核任务，点击「领取新任务」开始工作',
    statsConfig: {
      showPendingTasks: true,
      showTodayCompleted: true,
    },
  },
  qualityCheck: {
    title: '质检工作台',
    apiBasePath: '/tasks/quality-check',
    claimButtonText: '领取质检任务',
    emptyText: '暂无待质检任务，点击「领取质检任务」开始工作',
    statsConfig: {
      showPendingTasks: true,
      showTodayCompleted: true,
      showTotalCompleted: true,
      showPassRate: true,
    },
  },
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Handler 通用函数行为一致性

*For any* 任务类型（first_review, second_review, quality_check, video_first_review, video_second_review），使用通用 Handler 函数处理请求时，应该产生与原有专用 Handler 相同的响应格式和状态码。

**Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 1.6**

### Property 2: Repository 事务安全性

*For any* 并发的任务领取请求，使用 `FOR UPDATE SKIP LOCKED` 的通用 Repository 方法应该确保每个任务只被一个审核员领取，不会出现重复领取。

**Validates: Requirements 2.1, 2.6**

### Property 3: Service 验证逻辑一致性

*For any* 领取数量 count，如果 count < 1 或 count > 50，通用 Service 的验证方法应该返回错误；否则应该通过验证。

**Validates: Requirements 3.1, 3.2**

### Property 4: Redis 追踪数据完整性

*For any* 成功领取的任务列表，Redis 中应该存在对应的 claimed set 和 lock key；当任务完成或退回后，这些 key 应该被正确清理。

**Validates: Requirements 3.3, 3.4, 3.6**

### Property 5: API 工厂函数完整性

*For any* 有效的 API 路径前缀，`createTaskApi` 工厂函数应该返回包含 claim、getMyTasks、submit、submitBatch、returnTasks 五个方法的 API 对象，且每个方法调用正确的端点。

**Validates: Requirements 5.1, 5.2, 5.3, 5.5**

### Property 6: 错误响应格式一致性

*For any* 错误响应，应该包含 `error` 和 `code` 字段；客户端错误返回 400 状态码，服务器错误返回 500 状态码。

**Validates: Requirements 6.1, 6.2, 6.4, 6.5**

### Property 7: 限流器行为等价性

*For any* 相同的请求序列，重构后的通用限流器应该与原有限流器产生相同的限流行为（允许或拒绝）。

**Validates: Requirements 7.1, 7.2, 7.3, 7.5**

### Property 8: Dashboard 组件配置正确性

*For any* 有效的 Dashboard 配置（title, stats, apiBasePath），通用 Dashboard 组件应该正确渲染标题、统计卡片，并在事件触发时调用正确的 API 方法。

**Validates: Requirements 4.1, 4.2, 4.4, 4.5, 4.6**


## Error Handling

### 后端错误处理策略

```go
// internal/handlers/base/errors.go
package base

import "errors"

// 预定义错误码
const (
    ErrCodeInvalidRequest   = "INVALID_REQUEST"
    ErrCodeUnauthorized     = "UNAUTHORIZED"
    ErrCodeNotFound         = "NOT_FOUND"
    ErrCodeClaimFailed      = "CLAIM_FAILED"
    ErrCodeSubmitFailed     = "SUBMIT_FAILED"
    ErrCodeReturnFailed     = "RETURN_FAILED"
    ErrCodeInternalError    = "INTERNAL_ERROR"
    ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
)

// 预定义错误
var (
    ErrInvalidClaimCount    = errors.New("claim count must be between 1 and 50")
    ErrInvalidReturnCount   = errors.New("return count must be between 1 and 50")
    ErrHasUncompletedTasks  = errors.New("you have uncompleted tasks")
    ErrTaskNotFound         = errors.New("task not found or already completed")
    ErrNoTasksReturned      = errors.New("no tasks were returned")
)
```

### 前端错误处理策略

```typescript
// frontend/src/utils/errorHandler.ts
import { ElMessage } from 'element-plus'

export interface ApiError {
  error: string
  code: string
}

export function handleApiError(error: any, defaultMessage: string = '操作失败'): void {
  const apiError = error.response?.data as ApiError
  
  if (apiError?.error) {
    ElMessage.error(apiError.error)
  } else {
    ElMessage.error(defaultMessage)
  }
  
  console.error('API Error:', error)
}

export function isApiError(error: any): error is { response: { data: ApiError } } {
  return error?.response?.data?.error !== undefined
}
```

## Testing Strategy

### 测试方法

本项目采用双重测试策略：
1. **单元测试**: 验证具体示例和边界情况
2. **属性测试**: 验证通用属性在所有输入上的正确性

### 后端测试 (Go)

使用 `testing` 包进行单元测试，使用 `gopter` 进行属性测试。

```go
// internal/handlers/base/handlers_test.go
package base

import (
    "testing"
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
)

// Property 3: Service 验证逻辑一致性
func TestValidateClaimCount_Property(t *testing.T) {
    parameters := gopter.DefaultTestParameters()
    parameters.MinSuccessfulTests = 100

    properties := gopter.NewProperties(parameters)

    service := NewBaseTaskService(DefaultTaskServiceConfig("test", "test"), nil)

    properties.Property("claim count validation", prop.ForAll(
        func(count int) bool {
            err := service.ValidateClaimCount(count)
            if count >= 1 && count <= 50 {
                return err == nil
            }
            return err != nil
        },
        gen.IntRange(-100, 200),
    ))

    properties.TestingRun(t)
}

// Property 6: 错误响应格式一致性
func TestErrorResponse_Property(t *testing.T) {
    parameters := gopter.DefaultTestParameters()
    parameters.MinSuccessfulTests = 100

    properties := gopter.NewProperties(parameters)

    properties.Property("error response format", prop.ForAll(
        func(code, message string) bool {
            resp := ErrorResponse{Error: message, Code: code}
            return resp.Error != "" || resp.Code != ""
        },
        gen.AlphaString(),
        gen.AlphaString(),
    ))

    properties.TestingRun(t)
}
```

### 前端测试 (TypeScript/Vitest)

使用 `vitest` 进行单元测试，使用 `fast-check` 进行属性测试。

```typescript
// frontend/src/api/__tests__/taskApiFactory.test.ts
import { describe, it, expect } from 'vitest'
import fc from 'fast-check'
import { createTaskApi } from '../taskApiFactory'

// Property 5: API 工厂函数完整性
describe('createTaskApi', () => {
  it('should return API object with all required methods for any valid path', () => {
    fc.assert(
      fc.property(
        fc.string().filter(s => s.length > 0 && s.startsWith('/')),
        (basePath) => {
          const api = createTaskApi({ basePath })
          
          // 验证所有方法存在
          expect(typeof api.claim).toBe('function')
          expect(typeof api.getMyTasks).toBe('function')
          expect(typeof api.submit).toBe('function')
          expect(typeof api.submitBatch).toBe('function')
          expect(typeof api.returnTasks).toBe('function')
          
          return true
        }
      ),
      { numRuns: 100 }
    )
  })
})
```

### 集成测试

```go
// internal/handlers/base/integration_test.go
package base

import (
    "testing"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
)

// Property 1: Handler 通用函数行为一致性
func TestHandlerConsistency_Integration(t *testing.T) {
    // 测试不同任务类型使用通用 Handler 产生一致的响应格式
    taskTypes := []string{"first_review", "second_review", "quality_check"}
    
    for _, taskType := range taskTypes {
        t.Run(taskType, func(t *testing.T) {
            // 设置测试路由和 mock service
            router := gin.New()
            // ... 配置路由
            
            // 发送请求并验证响应格式
            w := httptest.NewRecorder()
            // ... 执行请求
            
            // 验证响应包含预期字段
            // ...
        })
    }
}
```

### 测试覆盖要求

| 组件 | 单元测试 | 属性测试 | 集成测试 |
|------|---------|---------|---------|
| Handler 通用函数 | ✓ | ✓ | ✓ |
| Repository 基类 | ✓ | ✓ | ✓ |
| Service 基类 | ✓ | ✓ | - |
| API 工厂函数 | ✓ | ✓ | - |
| 错误处理 | ✓ | ✓ | - |
| 限流器 | ✓ | ✓ | ✓ |
| Dashboard 组件 | ✓ | - | - |

