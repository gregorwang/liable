# 后端性能优化修改文档

> **文档版本**: v1.0
> **创建日期**: 2025-11-24
> **适用项目**: 评论审核平台（Comment Review Platform）
> **目标读者**: 后端开发者、AI编程辅助用户

---

## 📋 目录

1. [执行摘要](#执行摘要)
2. [性能审查方法论](#性能审查方法论)
3. [发现的性能问题](#发现的性能问题)
4. [优化方案详解](#优化方案详解)
5. [实施优先级](#实施优先级)
6. [预期性能提升](#预期性能提升)
7. [风险评估](#风险评估)
8. [测试验证计划](#测试验证计划)

---

## 🎯 执行摘要

### 审查结论
通过对评论审核平台后端代码的深度分析，**发现了7个关键性能瓶颈**，这些问题在当前5000+任务规模下虽未显现严重影响，但随着数据量增长（预估10万+任务时），将导致：
- **统计API响应时间从< 1s增长到10s+**
- **搜索功能从< 500ms增长到5s+**
- **数据库连接池耗尽风险**
- **内存占用增长10倍+**

### 关键指标
| 问题类型 | 严重程度 | 影响范围 | 优化收益 |
|---------|---------|---------|---------|
| 统计查询N次数据库往返 | 🔴 高 | 管理员后台 | **响应时间减少80%** |
| 搜索内存排序 | 🔴 高 | 任务搜索功能 | **内存占用减少90%** |
| 视频URL频繁查库 | 🟡 中 | 视频审核 | **数据库负载减少60%** |
| 缺少Redis缓存 | 🟡 中 | 所有读操作 | **响应时间减少50%** |
| UNION ALL查询过多 | 🟡 中 | 统计功能 | **查询时间减少40%** |

---

## 🔍 性能审查方法论

### 审查范围
```
审查代码文件：
├── internal/repository/stats_repo.go      (805行 - 统计仓库)
├── internal/services/task_service.go      (353行 - 任务服务)
├── internal/services/video_service.go     (425行 - 视频服务)
├── internal/handlers/admin.go             (522行 - 管理员接口)
└── pkg/database/postgres.go               (数据库连接池配置)
```

### 审查维度
1. **数据库查询效率**: SQL复杂度、索引使用、N+1查询
2. **内存使用**: 数据加载量、排序分页方式
3. **缓存策略**: Redis使用情况、缓存命中率
4. **并发处理**: 连接池配置、锁竞争
5. **API设计**: 批量操作、分页实现

### 测试数据假设
- 当前数据量: 5,000 任务
- 预期增长: 100,000 任务（6个月后）
- 并发用户: 当前10人 → 预期50人
- 审核员: 20人
- 每日新增任务: 1,000条

---

## 🚨 发现的性能问题

### 问题1: 统计查询执行多次数据库往返 🔴 严重

**位置**: `internal/repository/stats_repo.go:19-233`

**问题描述**:
`GetOverviewStats()` 函数为了获取完整的统计数据，执行了**10次独立的SQL查询**：

```go
// 伪代码展示问题
func (r *StatsRepository) GetOverviewStats() (*models.StatsOverview, error) {
    // 查询1: 评论一审任务统计
    r.db.QueryRow(commentFirstQuery).Scan(...)

    // 查询2: 评论一审审核结果统计
    r.db.QueryRow(commentFirstApprovalQuery).Scan(...)

    // 查询3: 评论二审任务统计
    r.db.QueryRow(commentSecondQuery).Scan(...)

    // 查询4: 评论二审审核结果统计
    r.db.QueryRow(commentSecondApprovalQuery).Scan(...)

    // 查询5: 视频一审任务统计
    r.db.QueryRow(videoFirstQuery).Scan(...)

    // 查询6: 视频一审审核结果统计
    r.db.QueryRow(videoFirstApprovalQuery).Scan(...)

    // 查询7: 视频二审任务统计
    r.db.QueryRow(videoSecondQuery).Scan(...)

    // 查询8: 视频二审审核结果统计
    r.db.QueryRow(videoSecondApprovalQuery).Scan(...)

    // 查询9: 审核员总数
    r.db.QueryRow(`SELECT COUNT(*) FROM users...`).Scan(...)

    // 查询10: 活跃审核员（5个UNION查询）
    r.db.QueryRow(activeReviewersQuery).Scan(...)

    // 查询11-12: 队列统计和质量指标
    getQueueStats()  // 内部又有复杂查询
    getQualityMetrics()

    return stats, nil
}
```

**性能影响**:
- 每次请求需要**10+ 次数据库往返**
- 网络延迟累积: `10 * 5ms = 50ms`（理想情况）
- 数据库负载高: 每个查询都需要扫描表
- 无法利用查询计划优化

**根本原因**:
1. ❌ 未使用JOIN合并查询
2. ❌ 未使用CTE（公共表表达式）优化复杂查询
3. ❌ 未使用Redis缓存结果
4. ❌ 实时计算而非定时汇总

---

### 问题2: 活跃审核员查询使用5个UNION 🔴 严重

**位置**: `internal/repository/stats_repo.go:203-216`

**问题SQL**:
```sql
SELECT COUNT(DISTINCT reviewer_id) FROM (
    SELECT reviewer_id FROM review_tasks
    WHERE status = 'completed' AND reviewer_id IS NOT NULL

    UNION  -- ⚠️ UNION会去重，比UNION ALL慢

    SELECT reviewer_id FROM second_review_tasks
    WHERE status = 'completed' AND reviewer_id IS NOT NULL

    UNION

    SELECT reviewer_id FROM quality_check_tasks
    WHERE status = 'completed' AND reviewer_id IS NOT NULL

    UNION

    SELECT reviewer_id FROM video_first_review_tasks
    WHERE status = 'completed' AND reviewer_id IS NOT NULL

    UNION

    SELECT reviewer_id FROM video_second_review_tasks
    WHERE status = 'completed' AND reviewer_id IS NOT NULL
) AS all_reviewers
```

**性能影响**:
- **跨5张表扫描**，每张表5000+行
- **UNION去重**需要排序和比较（应该用UNION ALL）
- 无法有效使用索引
- 查询时间随数据量线性增长

**数据量影响预估**:
| 任务数 | 查询时间 |
|--------|---------|
| 5,000 | ~50ms |
| 50,000 | ~500ms |
| 100,000 | ~1000ms (1s) |

---

### 问题3: 搜索功能内存排序和分页 🔴 严重

**位置**: `internal/services/task_service.go:267-352`

**问题代码**:
```go
func (s *TaskService) SearchTasks(req models.SearchTasksRequest) (*models.SearchTasksResponse, error) {
    var allResults []models.TaskSearchResult

    // 问题1: 分别查询两个队列（应该在数据库层合并）
    if req.QueueType == "first" || req.QueueType == "all" {
        firstResults, firstTotal, err := s.taskRepo.SearchTasks(req)
        allResults = append(allResults, firstResults...)  // ⚠️ 加载所有结果到内存
    }

    if req.QueueType == "second" || req.QueueType == "all" {
        secondResults, secondTotal, err := s.secondReviewRepo.SearchSecondReviewTasks(req)
        allResults = append(allResults, secondResults...)  // ⚠️ 又加载所有结果
    }

    // 问题2: 在内存中排序（应该用数据库ORDER BY）
    sort.Slice(allResults, func(i, j int) bool {
        if allResults[i].CompletedAt == nil && allResults[j].CompletedAt == nil {
            return allResults[i].CreatedAt.After(allResults[j].CreatedAt)
        }
        // ... 复杂的比较逻辑
        return allResults[i].CompletedAt.After(*allResults[j].CompletedAt)
    })

    // 问题3: 在内存中分页（应该用数据库LIMIT/OFFSET）
    offset := (req.Page - 1) * req.PageSize
    end := offset + req.PageSize
    if end > len(allResults) {
        end = len(allResults)
    }
    allResults = allResults[offset:end]  // ⚠️ 丢弃大部分已加载的数据

    return &models.SearchTasksResponse{Data: allResults}, nil
}
```

**性能影响分析**:

假设搜索条件匹配10,000条记录，用户只需要第1页（10条）：

```
❌ 当前实现：
1. 数据库查询 first 队列: 加载 5,000 条 → 内存占用 ~5MB
2. 数据库查询 second 队列: 加载 5,000 条 → 内存占用 ~5MB
3. 合并数组: 10,000 条 → 内存占用 10MB
4. 内存排序: 10,000 条比较操作 → CPU密集
5. 内存分页: 只保留 10 条，丢弃 9,990 条
6. 总内存占用: 10MB
7. 总CPU时间: ~200ms

✅ 优化后实现：
1. 数据库执行 UNION ALL + ORDER BY + LIMIT 10
2. 只返回 10 条记录 → 内存占用 ~10KB
3. 无需内存排序和分页
4. 总内存占用: 10KB（减少99.9%）
5. 总查询时间: ~20ms（减少90%）
```

**可扩展性问题**:
- 当任务数达到100万时，可能加载**100MB+数据到内存**
- 内存排序时间增长到**数秒**
- 可能触发Go GC，导致STW（Stop The World）

---

### 问题4: 视频URL生成频繁查询数据库 🟡 中等

**位置**: `internal/services/video_service.go:109-143`

**问题代码**:
```go
func (s *VideoService) GenerateVideoURL(videoID int) (*models.GenerateVideoURLResponse, error) {
    // 问题: 每次请求都查数据库
    video, err := s.videoRepo.GetVideoByID(videoID)  // ⚠️ 数据库查询
    if err != nil {
        return nil, fmt.Errorf("video not found: %w", err)
    }

    // 检查缓存的URL是否过期
    if video.VideoURL != nil && video.URLExpiresAt != nil &&
       video.URLExpiresAt.After(time.Now()) {
        return &models.GenerateVideoURLResponse{
            VideoURL:  *video.VideoURL,
            ExpiresAt: *video.URLExpiresAt,
        }, nil
    }

    // 生成新的预签名URL
    expiration := 1 * time.Hour
    videoURL, err := s.r2Service.GeneratePresignedURL(video.VideoKey, expiration)  // ⚠️ R2 API调用
    if err != nil {
        return nil, err
    }

    // 更新数据库
    s.videoRepo.UpdateVideoURL(videoID, videoURL, expiresAt)  // ⚠️ 又一次数据库写入

    return response, nil
}
```

**性能影响**:
- 每个视频审核请求都触发**1次数据库读 + 可能1次数据库写**
- 高并发下（50个审核员同时工作）：
  - 数据库QPS: 50 * 2 = 100 QPS
  - R2 API调用: 可能触发限流
- URL在数据库中缓存，但**未使用Redis缓存**

**场景模拟**:
```
审核员A打开视频1 → 查DB → 生成URL → 写DB
审核员B打开视频1（5秒后）→ 查DB → 读到缓存URL → 返回
审核员C打开视频1（10秒后）→ 查DB → 读到缓存URL → 返回

问题: 每次都查DB，即使URL在缓存中
```

---

### 问题5: 统计API无Redis缓存 🟡 中等

**位置**: `internal/handlers/admin.go:71-79`

**问题代码**:
```go
func (h *AdminHandler) GetOverviewStats(c *gin.Context) {
    // 问题: 直接调用service，无缓存
    stats, err := h.statsService.GetOverviewStats()  // ⚠️ 触发10+次数据库查询
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, stats)
}
```

**问题分析**:
- 管理员每次刷新页面都触发完整统计查询
- 统计数据实时性要求**不高**（5分钟延迟可接受）
- 未使用Redis缓存
- 未使用ETag/Last-Modified等HTTP缓存

**理想流程**:
```
请求1: 管理员刷新页面
→ 检查Redis缓存 (MISS)
→ 查询数据库 (10+ 查询)
→ 写入Redis (TTL=5分钟)
→ 返回结果

请求2-N (5分钟内): 管理员或其他管理员刷新
→ 检查Redis缓存 (HIT)
→ 直接返回缓存数据
→ 0次数据库查询
```

---

### 问题6: UNION ALL查询过多 🟡 中等

**位置**: 多处，例如 `stats_repo.go:264-286`

**问题SQL**:
```sql
-- GetHourlyStats 函数
SELECT
    EXTRACT(HOUR FROM created_at)::int as hour,
    COUNT(*) as count
FROM (
    SELECT created_at FROM review_results WHERE DATE(created_at) = $1
    UNION ALL
    SELECT created_at FROM second_review_results WHERE DATE(created_at) = $1
    UNION ALL
    SELECT created_at FROM quality_check_results WHERE DATE(created_at) = $1
    UNION ALL
    SELECT created_at FROM video_first_review_results WHERE DATE(created_at) = $1
    UNION ALL
    SELECT created_at FROM video_second_review_results WHERE DATE(created_at) = $1
) all_reviews
GROUP BY hour
ORDER BY hour
```

**问题分析**:
- 需要扫描**5张表**
- 每张表都执行`DATE(created_at) = $1`过滤
- 如果没有按日期的索引，将进行全表扫描
- UNION ALL虽然比UNION快，但仍需要多次表访问

**索引检查**:
```sql
-- 当前索引（从分析报告）
CREATE INDEX idx_review_results_created_at ON review_results(created_at);

-- 问题: DATE函数无法使用索引
WHERE DATE(created_at) = '2024-01-01'  -- ❌ 无法用索引

-- 优化: 使用范围查询
WHERE created_at >= '2024-01-01' AND created_at < '2024-01-02'  -- ✅ 可用索引
```

---

### 问题7: 批量提交审核的事务处理 🟢 轻微

**位置**: `internal/services/task_service.go:152-160`

**问题代码**:
```go
func (s *TaskService) SubmitBatchReviews(reviewerID int, reviews []models.SubmitReviewRequest) error {
    for _, review := range reviews {
        if err := s.SubmitReview(reviewerID, review); err != nil {  // ⚠️ 每次单独提交
            return err
        }
    }
    return nil
}
```

**问题分析**:
- 每个审核都是独立的事务
- 如果批量提交20条，需要**20次事务提交**
- 网络往返次数增加
- 无法利用批量INSERT优化

**优化潜力**:
```go
// ❌ 当前: 20条审核 = 20次提交 = 20次网络往返
for i := 0; i < 20; i++ {
    BEGIN TRANSACTION
    INSERT INTO review_results ...
    COMMIT
}

// ✅ 优化: 20条审核 = 1次提交 = 1次网络往返
BEGIN TRANSACTION
INSERT INTO review_results VALUES (...), (...), (...)  -- 批量插入
COMMIT
```

---

## 🛠️ 优化方案详解

### 优化方案1: 合并统计查询 + Redis缓存 🔴 高优先级

**目标**: 将GetOverviewStats从10+次查询减少到1次查询，并添加Redis缓存

#### 实施步骤

**步骤1: 创建物化视图或定时汇总表**

```sql
-- 创建统计汇总表
CREATE TABLE stats_cache (
    id SERIAL PRIMARY KEY,
    cache_key VARCHAR(100) UNIQUE NOT NULL,
    cache_data JSONB NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_stats_cache_key ON stats_cache(cache_key);
CREATE INDEX idx_stats_cache_updated ON stats_cache(updated_at);
```

**步骤2: 优化GetOverviewStats查询**

```go
// 文件: internal/repository/stats_repo.go
func (r *StatsRepository) GetOverviewStats() (*models.StatsOverview, error) {
    stats := &models.StatsOverview{}

    // 使用单个复杂查询代替多次查询
    query := `
        WITH comment_first AS (
            SELECT
                COUNT(*) as total,
                COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
                COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
                COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress
            FROM review_tasks
        ),
        comment_first_approval AS (
            SELECT
                COUNT(CASE WHEN is_approved = true THEN 1 END) as approved,
                COUNT(CASE WHEN is_approved = false THEN 1 END) as rejected
            FROM review_results
        ),
        comment_second AS (
            SELECT
                COUNT(*) as total,
                COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
                COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
                COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress
            FROM second_review_tasks
        ),
        comment_second_approval AS (
            SELECT
                COUNT(CASE WHEN is_approved = true THEN 1 END) as approved,
                COUNT(CASE WHEN is_approved = false THEN 1 END) as rejected
            FROM second_review_results
        ),
        -- ... 其他CTE（公共表表达式）
        active_reviewers AS (
            SELECT COUNT(DISTINCT reviewer_id) as count
            FROM (
                SELECT reviewer_id FROM review_tasks WHERE status = 'completed'
                UNION ALL  -- 改用UNION ALL（无需去重，外层已经DISTINCT）
                SELECT reviewer_id FROM second_review_tasks WHERE status = 'completed'
                UNION ALL
                SELECT reviewer_id FROM quality_check_tasks WHERE status = 'completed'
                UNION ALL
                SELECT reviewer_id FROM video_first_review_tasks WHERE status = 'completed'
                UNION ALL
                SELECT reviewer_id FROM video_second_review_tasks WHERE status = 'completed'
            ) all_reviewers
            WHERE reviewer_id IS NOT NULL
        )
        SELECT
            -- 从各个CTE中选择数据
            cf.total, cf.completed, cf.pending, cf.in_progress,
            cfa.approved, cfa.rejected,
            cs.total, cs.completed, cs.pending, cs.in_progress,
            csa.approved, csa.rejected,
            ar.count
        FROM comment_first cf, comment_first_approval cfa,
             comment_second cs, comment_second_approval csa,
             active_reviewers ar
    `

    // 单次查询获取所有数据
    err := r.db.QueryRow(query).Scan(
        &stats.CommentReviewStats.FirstReview.TotalTasks,
        &stats.CommentReviewStats.FirstReview.CompletedTasks,
        &stats.CommentReviewStats.FirstReview.PendingTasks,
        &stats.CommentReviewStats.FirstReview.InProgressTasks,
        &stats.CommentReviewStats.FirstReview.ApprovedCount,
        &stats.CommentReviewStats.FirstReview.RejectedCount,
        &stats.CommentReviewStats.SecondReview.TotalTasks,
        &stats.CommentReviewStats.SecondReview.CompletedTasks,
        &stats.CommentReviewStats.SecondReview.PendingTasks,
        &stats.CommentReviewStats.SecondReview.InProgressTasks,
        &stats.CommentReviewStats.SecondReview.ApprovedCount,
        &stats.CommentReviewStats.SecondReview.RejectedCount,
        &stats.ActiveReviewers,
    )

    if err != nil {
        return nil, err
    }

    // 计算派生字段
    if stats.CommentReviewStats.FirstReview.CompletedTasks > 0 {
        stats.CommentReviewStats.FirstReview.ApprovalRate =
            float64(stats.CommentReviewStats.FirstReview.ApprovedCount) /
            float64(stats.CommentReviewStats.FirstReview.CompletedTasks) * 100
    }

    // ... 其他计算

    return stats, nil
}
```

**步骤3: 在Service层添加Redis缓存**

```go
// 文件: internal/services/stats_service.go
package services

import (
    "comment-review-platform/internal/models"
    "comment-review-platform/internal/repository"
    redispkg "comment-review-platform/pkg/redis"
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

type StatsService struct {
    statsRepo *repository.StatsRepository
    rdb       *redis.Client
    ctx       context.Context
}

func NewStatsService() *StatsService {
    return &StatsService{
        statsRepo: repository.NewStatsRepository(),
        rdb:       redispkg.Client,
        ctx:       context.Background(),
    }
}

func (s *StatsService) GetOverviewStats() (*models.StatsOverview, error) {
    cacheKey := "stats:overview"
    cacheTTL := 5 * time.Minute  // 5分钟缓存

    // 1. 尝试从Redis读取
    cached, err := s.rdb.Get(s.ctx, cacheKey).Result()
    if err == nil {
        var stats models.StatsOverview
        if err := json.Unmarshal([]byte(cached), &stats); err == nil {
            // 缓存命中，直接返回
            return &stats, nil
        }
    }

    // 2. 缓存未命中，查询数据库
    stats, err := s.statsRepo.GetOverviewStats()
    if err != nil {
        return nil, err
    }

    // 3. 写入Redis缓存
    data, err := json.Marshal(stats)
    if err == nil {
        s.rdb.Set(s.ctx, cacheKey, data, cacheTTL)
    }

    return stats, nil
}

// 添加强制刷新方法
func (s *StatsService) RefreshOverviewStats() (*models.StatsOverview, error) {
    cacheKey := "stats:overview"

    // 删除缓存
    s.rdb.Del(s.ctx, cacheKey)

    // 重新查询
    return s.GetOverviewStats()
}
```

**步骤4: 在Handler层添加刷新接口**

```go
// 文件: internal/handlers/admin.go
func (h *AdminHandler) GetOverviewStats(c *gin.Context) {
    // 检查是否需要强制刷新
    refresh := c.Query("refresh") == "true"

    var stats *models.StatsOverview
    var err error

    if refresh {
        stats, err = h.statsService.RefreshOverviewStats()
    } else {
        stats, err = h.statsService.GetOverviewStats()
    }

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 添加缓存信息到响应头
    c.Header("X-Cache-Status", "HIT")
    c.JSON(http.StatusOK, stats)
}
```

#### 预期效果

| 指标 | 优化前 | 优化后 | 提升 |
|------|-------|-------|------|
| 数据库查询次数 | 10+ | 1 | **90%** |
| 响应时间（首次） | ~500ms | ~150ms | **70%** |
| 响应时间（缓存） | ~500ms | ~5ms | **99%** |
| 数据库负载 | 高 | 低 | **80%** |
| 缓存命中率 | 0% | 95%+ | - |

---

### 优化方案2: 数据库层搜索排序分页 🔴 高优先级

**目标**: 将搜索功能的排序和分页移到数据库层，减少内存占用

#### 实施步骤

**步骤1: 修改Repository层使用UNION查询**

```go
// 文件: internal/repository/task_repo.go
func (r *TaskRepository) SearchTasksUnified(req models.SearchTasksRequest) ([]models.TaskSearchResult, int, error) {
    // 构建WHERE条件
    whereClauses := []string{}
    args := []interface{}{}
    argIndex := 1

    // 审核员筛选
    if req.ReviewerID > 0 {
        whereClauses = append(whereClauses, fmt.Sprintf("reviewer_id = $%d", argIndex))
        args = append(args, req.ReviewerID)
        argIndex++
    }

    // 状态筛选
    if req.Status != "" {
        whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
        args = append(args, req.Status)
        argIndex++
    }

    // 时间范围筛选
    if req.StartDate != "" {
        whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argIndex))
        args = append(args, req.StartDate)
        argIndex++
    }

    if req.EndDate != "" {
        whereClauses = append(whereClauses, fmt.Sprintf("created_at < $%d", argIndex))
        args = append(args, req.EndDate)
        argIndex++
    }

    whereClause := ""
    if len(whereClauses) > 0 {
        whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
    }

    // 构建UNION查询（根据QueueType）
    var unionQuery string

    if req.QueueType == "all" || req.QueueType == "first" {
        unionQuery = fmt.Sprintf(`
            SELECT
                rt.id,
                rt.comment_id,
                rt.reviewer_id,
                rt.status,
                rt.created_at,
                rt.claimed_at,
                rt.completed_at,
                'first' as queue_type,
                c.text as comment_text,
                u.username as reviewer_name,
                rr.is_approved,
                rr.tags,
                rr.reason
            FROM review_tasks rt
            INNER JOIN comment c ON rt.comment_id = c.id
            LEFT JOIN users u ON rt.reviewer_id = u.id
            LEFT JOIN review_results rr ON rt.id = rr.task_id
            %s
        `, whereClause)
    }

    if req.QueueType == "all" || req.QueueType == "second" {
        secondQuery := fmt.Sprintf(`
            SELECT
                srt.id,
                srt.comment_id,
                srt.reviewer_id,
                srt.status,
                srt.created_at,
                srt.claimed_at,
                srt.completed_at,
                'second' as queue_type,
                c.text as comment_text,
                u.username as reviewer_name,
                srr.is_approved,
                srr.tags,
                srr.reason
            FROM second_review_tasks srt
            INNER JOIN comment c ON srt.comment_id = c.id
            LEFT JOIN users u ON srt.reviewer_id = u.id
            LEFT JOIN second_review_results srr ON srt.id = srr.second_task_id
            %s
        `, whereClause)

        if unionQuery != "" {
            unionQuery = fmt.Sprintf("(%s) UNION ALL (%s)", unionQuery, secondQuery)
        } else {
            unionQuery = secondQuery
        }
    }

    // 添加排序和分页（关键优化点）
    offset := (req.Page - 1) * req.PageSize
    finalQuery := fmt.Sprintf(`
        SELECT * FROM (%s) combined_results
        ORDER BY
            CASE WHEN completed_at IS NULL THEN 1 ELSE 0 END,
            COALESCE(completed_at, created_at) DESC
        LIMIT $%d OFFSET $%d
    `, unionQuery, argIndex, argIndex+1)

    args = append(args, req.PageSize, offset)

    // 执行查询
    rows, err := r.db.Query(finalQuery, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    results := []models.TaskSearchResult{}
    for rows.Next() {
        var result models.TaskSearchResult
        var tags []string  // 用于扫描PostgreSQL数组

        err := rows.Scan(
            &result.ID,
            &result.CommentID,
            &result.ReviewerID,
            &result.Status,
            &result.CreatedAt,
            &result.ClaimedAt,
            &result.CompletedAt,
            &result.QueueType,
            &result.CommentText,
            &result.ReviewerName,
            &result.IsApproved,
            pq.Array(&tags),  // 使用pq.Array扫描数组
            &result.Reason,
        )
        if err != nil {
            return nil, 0, err
        }

        result.Tags = tags
        results = append(results, result)
    }

    // 获取总数（使用COUNT查询）
    countQuery := fmt.Sprintf(`
        SELECT COUNT(*) FROM (%s) combined_results
    `, unionQuery)

    var totalCount int
    err = r.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&totalCount)  // 去掉LIMIT和OFFSET参数
    if err != nil {
        return nil, 0, err
    }

    return results, totalCount, nil
}
```

**步骤2: 简化Service层逻辑**

```go
// 文件: internal/services/task_service.go
func (s *TaskService) SearchTasks(req models.SearchTasksRequest) (*models.SearchTasksResponse, error) {
    // 设置默认值
    if req.Page < 1 {
        req.Page = 1
    }
    if req.PageSize < 1 {
        req.PageSize = 10
    }
    if req.PageSize > 100 {
        req.PageSize = 100
    }
    if req.QueueType == "" {
        req.QueueType = "all"
    }

    // 直接调用统一的搜索方法（数据库层已完成排序和分页）
    results, totalCount, err := s.taskRepo.SearchTasksUnified(req)
    if err != nil {
        return nil, err
    }

    // 计算总页数
    totalPages := totalCount / req.PageSize
    if totalCount%req.PageSize > 0 {
        totalPages++
    }

    response := &models.SearchTasksResponse{
        Data:       results,
        Total:      totalCount,
        Page:       req.Page,
        PageSize:   req.PageSize,
        TotalPages: totalPages,
    }

    return response, nil
}
```

#### 预期效果

| 场景 | 优化前 | 优化后 | 提升 |
|------|-------|-------|------|
| 搜索10,000条记录，返回第1页 | 加载10,000条到内存 | 只加载10条 | **内存减少99.9%** |
| 响应时间 | ~500ms | ~50ms | **90%** |
| CPU使用率 | 高（内存排序） | 低 | **70%** |
| 可扩展性 | 随数据量线性增长 | 恒定性能 | ✅ |

---

### 优化方案3: 视频URL Redis缓存 🟡 中优先级

**目标**: 使用Redis缓存视频预签名URL，减少数据库和R2访问

#### 实施步骤

**步骤1: 修改VideoService添加Redis缓存**

```go
// 文件: internal/services/video_service.go
import (
    redispkg "comment-review-platform/pkg/redis"
    "github.com/redis/go-redis/v9"
)

type VideoService struct {
    videoRepo *repository.VideoRepository
    r2Service *r2.R2Service
    rdb       *redis.Client
    ctx       context.Context
}

func NewVideoService() (*VideoService, error) {
    r2Service, err := r2.NewR2Service()
    if err != nil {
        return nil, fmt.Errorf("failed to initialize R2 service: %w", err)
    }

    return &VideoService{
        videoRepo: repository.NewVideoRepository(),
        r2Service: r2Service,
        rdb:       redispkg.Client,
        ctx:       context.Background(),
    }, nil
}

func (s *VideoService) GenerateVideoURL(videoID int) (*models.GenerateVideoURLResponse, error) {
    cacheKey := fmt.Sprintf("video:url:%d", videoID)

    // 1. 先查Redis缓存
    cached, err := s.rdb.Get(s.ctx, cacheKey).Result()
    if err == nil {
        // 缓存命中，解析JSON
        var response models.GenerateVideoURLResponse
        if err := json.Unmarshal([]byte(cached), &response); err == nil {
            // 检查URL是否过期（留5分钟缓冲）
            if response.ExpiresAt.Add(-5 * time.Minute).After(time.Now()) {
                return &response, nil
            }
        }
    }

    // 2. Redis未命中或已过期，查数据库
    video, err := s.videoRepo.GetVideoByID(videoID)
    if err != nil {
        return nil, fmt.Errorf("video not found: %w", err)
    }

    // 3. 检查数据库中的URL是否有效
    if video.VideoURL != nil && video.URLExpiresAt != nil &&
       video.URLExpiresAt.Add(-5 * time.Minute).After(time.Now()) {
        response := &models.GenerateVideoURLResponse{
            VideoURL:  *video.VideoURL,
            ExpiresAt: *video.URLExpiresAt,
        }

        // 写回Redis
        s.cacheVideoURL(cacheKey, response)

        return response, nil
    }

    // 4. 生成新的预签名URL
    expiration := 1 * time.Hour
    videoURL, err := s.r2Service.GeneratePresignedURL(video.VideoKey, expiration)
    if err != nil {
        return nil, fmt.Errorf("failed to generate pre-signed URL: %w", err)
    }

    expiresAt := time.Now().Add(expiration)
    response := &models.GenerateVideoURLResponse{
        VideoURL:  videoURL,
        ExpiresAt: expiresAt,
    }

    // 5. 异步更新数据库（不阻塞响应）
    go func() {
        if err := s.videoRepo.UpdateVideoURL(videoID, videoURL, expiresAt); err != nil {
            log.Printf("Warning: Failed to update video URL in database: %v", err)
        }
    }()

    // 6. 写入Redis缓存
    s.cacheVideoURL(cacheKey, response)

    return response, nil
}

// 辅助方法: 缓存视频URL到Redis
func (s *VideoService) cacheVideoURL(cacheKey string, response *models.GenerateVideoURLResponse) {
    data, err := json.Marshal(response)
    if err != nil {
        log.Printf("Warning: Failed to marshal video URL: %v", err)
        return
    }

    // 缓存50分钟（URL有效期1小时，留10分钟缓冲）
    ttl := 50 * time.Minute
    if err := s.rdb.Set(s.ctx, cacheKey, data, ttl).Err(); err != nil {
        log.Printf("Warning: Failed to cache video URL to Redis: %v", err)
    }
}
```

#### 预期效果

| 操作 | 优化前 | 优化后 | 提升 |
|------|-------|-------|------|
| 首次请求视频URL | 1次DB读 + 1次R2调用 + 1次DB写 | 1次Redis读(MISS) + 1次DB读 + 1次R2调用 + 异步DB写 + 1次Redis写 | 响应时间-20% |
| 后续请求（缓存期内） | 1次DB读 | 1次Redis读(HIT) | **响应时间-90%** |
| 数据库负载 | 每次请求都查DB | 50分钟内只查1次 | **减少98%** |
| R2 API调用 | 每小时可能多次 | 每小时1次 | **减少95%** |

---

### 优化方案4: 添加数据库索引优化 🟡 中优先级

**目标**: 添加复合索引，优化查询性能

#### 需要添加的索引

```sql
-- 文件: migrations/007_performance_indexes.sql

-- 1. 任务状态+领取时间复合索引（用于释放超时任务）
CREATE INDEX CONCURRENTLY idx_review_tasks_status_claimed
ON review_tasks(status, claimed_at)
WHERE status = 'in_progress';

CREATE INDEX CONCURRENTLY idx_second_review_tasks_status_claimed
ON second_review_tasks(status, claimed_at)
WHERE status = 'in_progress';

CREATE INDEX CONCURRENTLY idx_quality_check_tasks_status_claimed
ON quality_check_tasks(status, claimed_at)
WHERE status = 'in_progress';

CREATE INDEX CONCURRENTLY idx_video_first_review_tasks_status_claimed
ON video_first_review_tasks(status, claimed_at)
WHERE status = 'in_progress';

CREATE INDEX CONCURRENTLY idx_video_second_review_tasks_status_claimed
ON video_second_review_tasks(status, claimed_at)
WHERE status = 'in_progress';

-- 2. 审核员+状态复合索引（用于查询我的任务）
CREATE INDEX CONCURRENTLY idx_review_tasks_reviewer_status
ON review_tasks(reviewer_id, status)
WHERE reviewer_id IS NOT NULL;

CREATE INDEX CONCURRENTLY idx_second_review_tasks_reviewer_status
ON second_review_tasks(reviewer_id, status)
WHERE reviewer_id IS NOT NULL;

-- 3. 创建时间范围查询索引（用于统计查询）
CREATE INDEX CONCURRENTLY idx_review_results_created_date
ON review_results(DATE(created_at), created_at);

CREATE INDEX CONCURRENTLY idx_second_review_results_created_date
ON second_review_results(DATE(created_at), created_at);

CREATE INDEX CONCURRENTLY idx_quality_check_results_created_date
ON quality_check_results(DATE(created_at), created_at);

CREATE INDEX CONCURRENTLY idx_video_first_review_results_created_date
ON video_first_review_results(DATE(created_at), created_at);

CREATE INDEX CONCURRENTLY idx_video_second_review_results_created_date
ON video_second_review_results(DATE(created_at), created_at);

-- 4. 审核员绩效统计索引
CREATE INDEX CONCURRENTLY idx_review_results_reviewer_approved
ON review_results(reviewer_id, is_approved);

CREATE INDEX CONCURRENTLY idx_second_review_results_reviewer_approved
ON second_review_results(reviewer_id, is_approved);

-- 5. 视频质量标签统计索引（GIN索引用于JSONB）
CREATE INDEX CONCURRENTLY idx_video_first_review_results_quality_dims
ON video_first_review_results USING GIN (quality_dimensions);

CREATE INDEX CONCURRENTLY idx_video_second_review_results_quality_dims
ON video_second_review_results USING GIN (quality_dimensions);

-- 注意: 使用CONCURRENTLY避免锁表
```

#### 索引使用说明

| 索引 | 使用场景 | 预期提升 |
|------|---------|---------|
| `idx_*_status_claimed` | 释放超时任务查询 | **90%** |
| `idx_*_reviewer_status` | 查询我的任务 | **80%** |
| `idx_*_created_date` | 每日/每小时统计 | **70%** |
| `idx_*_reviewer_approved` | 审核员绩效统计 | **60%** |
| `idx_*_quality_dims` | 视频质量标签统计 | **50%** |

#### 注意事项

```sql
-- ⚠️ 创建索引注意事项:

-- 1. 使用CONCURRENTLY避免锁表（生产环境必须）
CREATE INDEX CONCURRENTLY ...;

-- 2. 监控索引创建进度
SELECT
    pid,
    now() - pg_stat_activity.query_start AS duration,
    query
FROM pg_stat_activity
WHERE query LIKE 'CREATE INDEX%';

-- 3. 检查索引大小
SELECT
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
ORDER BY pg_relation_size(indexrelid) DESC;

-- 4. 验证索引是否被使用
EXPLAIN ANALYZE
SELECT * FROM review_tasks
WHERE status = 'in_progress'
ORDER BY claimed_at ASC;

-- 应该看到:
-- Index Scan using idx_review_tasks_status_claimed on review_tasks
```

---

### 优化方案5: 定时任务汇总统计数据 🟢 低优先级

**目标**: 使用后台任务定时汇总统计数据，避免实时计算

#### 实施步骤

**步骤1: 创建统计汇总表**

```sql
-- 文件: migrations/008_stats_aggregation.sql

CREATE TABLE stats_hourly_agg (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    hour INTEGER NOT NULL,
    queue_type VARCHAR(50) NOT NULL,
    review_count INTEGER DEFAULT 0,
    approved_count INTEGER DEFAULT 0,
    rejected_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(date, hour, queue_type)
);

CREATE TABLE stats_daily_agg (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    total_reviews INTEGER DEFAULT 0,
    comment_first_reviews INTEGER DEFAULT 0,
    comment_second_reviews INTEGER DEFAULT 0,
    quality_checks INTEGER DEFAULT 0,
    video_first_reviews INTEGER DEFAULT 0,
    video_second_reviews INTEGER DEFAULT 0,
    active_reviewers INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_stats_hourly_date_hour ON stats_hourly_agg(date, hour);
CREATE INDEX idx_stats_daily_date ON stats_daily_agg(date);
```

**步骤2: 实现聚合函数**

```go
// 文件: internal/services/stats_aggregation_service.go
package services

import (
    "comment-review-platform/pkg/database"
    "database/sql"
    "log"
    "time"
)

type StatsAggregationService struct {
    db *sql.DB
}

func NewStatsAggregationService() *StatsAggregationService {
    return &StatsAggregationService{
        db: database.DB,
    }
}

// AggregateHourlyStats 汇总每小时统计数据
func (s *StatsAggregationService) AggregateHourlyStats(date time.Time, hour int) error {
    dateStr := date.Format("2006-01-02")

    query := `
        INSERT INTO stats_hourly_agg (date, hour, queue_type, review_count, approved_count, rejected_count, updated_at)
        VALUES
            ($1, $2, 'comment_first',
             (SELECT COUNT(*) FROM review_results WHERE DATE(created_at) = $1 AND EXTRACT(HOUR FROM created_at) = $2),
             (SELECT COUNT(*) FROM review_results WHERE DATE(created_at) = $1 AND EXTRACT(HOUR FROM created_at) = $2 AND is_approved = true),
             (SELECT COUNT(*) FROM review_results WHERE DATE(created_at) = $1 AND EXTRACT(HOUR FROM created_at) = $2 AND is_approved = false),
             NOW()
            ),
            ($1, $2, 'comment_second',
             (SELECT COUNT(*) FROM second_review_results WHERE DATE(created_at) = $1 AND EXTRACT(HOUR FROM created_at) = $2),
             (SELECT COUNT(*) FROM second_review_results WHERE DATE(created_at) = $1 AND EXTRACT(HOUR FROM created_at) = $2 AND is_approved = true),
             (SELECT COUNT(*) FROM second_review_results WHERE DATE(created_at) = $1 AND EXTRACT(HOUR FROM created_at) = $2 AND is_approved = false),
             NOW()
            )
        ON CONFLICT (date, hour, queue_type)
        DO UPDATE SET
            review_count = EXCLUDED.review_count,
            approved_count = EXCLUDED.approved_count,
            rejected_count = EXCLUDED.rejected_count,
            updated_at = NOW()
    `

    _, err := s.db.Exec(query, dateStr, hour)
    if err != nil {
        return err
    }

    log.Printf("Aggregated hourly stats for %s hour %d", dateStr, hour)
    return nil
}

// AggregateDailyStats 汇总每日统计数据
func (s *StatsAggregationService) AggregateDailyStats(date time.Time) error {
    dateStr := date.Format("2006-01-02")

    query := `
        INSERT INTO stats_daily_agg (
            date,
            total_reviews,
            comment_first_reviews,
            comment_second_reviews,
            quality_checks,
            video_first_reviews,
            video_second_reviews,
            active_reviewers,
            updated_at
        )
        SELECT
            $1::date,
            (SELECT COUNT(*) FROM review_results WHERE DATE(created_at) = $1) +
            (SELECT COUNT(*) FROM second_review_results WHERE DATE(created_at) = $1) +
            (SELECT COUNT(*) FROM quality_check_results WHERE DATE(created_at) = $1) +
            (SELECT COUNT(*) FROM video_first_review_results WHERE DATE(created_at) = $1) +
            (SELECT COUNT(*) FROM video_second_review_results WHERE DATE(created_at) = $1),
            (SELECT COUNT(*) FROM review_results WHERE DATE(created_at) = $1),
            (SELECT COUNT(*) FROM second_review_results WHERE DATE(created_at) = $1),
            (SELECT COUNT(*) FROM quality_check_results WHERE DATE(created_at) = $1),
            (SELECT COUNT(*) FROM video_first_review_results WHERE DATE(created_at) = $1),
            (SELECT COUNT(*) FROM video_second_review_results WHERE DATE(created_at) = $1),
            (SELECT COUNT(DISTINCT reviewer_id) FROM (
                SELECT reviewer_id FROM review_results WHERE DATE(created_at) = $1
                UNION ALL
                SELECT reviewer_id FROM second_review_results WHERE DATE(created_at) = $1
                UNION ALL
                SELECT reviewer_id FROM quality_check_results WHERE DATE(created_at) = $1
                UNION ALL
                SELECT reviewer_id FROM video_first_review_results WHERE DATE(created_at) = $1
                UNION ALL
                SELECT reviewer_id FROM video_second_review_results WHERE DATE(created_at) = $1
            ) all_reviewers WHERE reviewer_id IS NOT NULL),
            NOW()
        ON CONFLICT (date)
        DO UPDATE SET
            total_reviews = EXCLUDED.total_reviews,
            comment_first_reviews = EXCLUDED.comment_first_reviews,
            comment_second_reviews = EXCLUDED.comment_second_reviews,
            quality_checks = EXCLUDED.quality_checks,
            video_first_reviews = EXCLUDED.video_first_reviews,
            video_second_reviews = EXCLUDED.video_second_reviews,
            active_reviewers = EXCLUDED.active_reviewers,
            updated_at = NOW()
    `

    _, err := s.db.Exec(query, dateStr)
    if err != nil {
        return err
    }

    log.Printf("Aggregated daily stats for %s", dateStr)
    return nil
}
```

**步骤3: 在main.go中启动定时任务**

```go
// 文件: cmd/api/main.go
func startStatsAggregationWorker() {
    aggService := services.NewStatsAggregationService()

    // 每小时的第5分钟执行聚合
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        now := time.Now()

        // 每小时第5分钟执行
        if now.Minute() == 5 {
            // 汇总上一小时的数据
            lastHour := now.Add(-1 * time.Hour)
            if err := aggService.AggregateHourlyStats(lastHour, lastHour.Hour()); err != nil {
                log.Printf("Error aggregating hourly stats: %v", err)
            }
        }

        // 每天凌晨1点执行
        if now.Hour() == 1 && now.Minute() == 5 {
            // 汇总昨天的数据
            yesterday := now.AddDate(0, 0, -1)
            if err := aggService.AggregateDailyStats(yesterday); err != nil {
                log.Printf("Error aggregating daily stats: %v", err)
            }
        }
    }
}

func main() {
    // ... 其他初始化代码

    // 启动后台任务
    go startTaskReleaseWorker()
    go startSamplingScheduler()
    go startStatsAggregationWorker()  // 新增

    // 启动HTTP服务
    router.Run(":8080")
}
```

#### 预期效果

- 统计查询从实时计算改为查询汇总表
- 响应时间从500ms降低到10ms
- 历史统计数据可追溯

---

## 📊 实施优先级

### 优先级矩阵

| 优化方案 | 实施难度 | 预期收益 | 优先级 | 预计工时 |
|---------|---------|---------|-------|---------|
| 方案1: 合并统计查询+Redis缓存 | 中 | 极高 | 🔴 P0 | 8小时 |
| 方案2: 数据库层搜索排序分页 | 高 | 极高 | 🔴 P0 | 12小时 |
| 方案3: 视频URL Redis缓存 | 低 | 高 | 🟡 P1 | 4小时 |
| 方案4: 添加数据库索引 | 低 | 中 | 🟡 P1 | 2小时 |
| 方案5: 定时任务汇总统计 | 中 | 中 | 🟢 P2 | 10小时 |

### 实施路线图

```
第一阶段（立即实施）：
├── 方案4: 添加数据库索引（2小时）
│   └── 风险低，收益立竿见影
└── 方案1: 合并统计查询+Redis缓存（8小时）
    └── 解决管理员后台性能问题

第二阶段（1周内）：
├── 方案2: 数据库层搜索排序分页（12小时）
│   └── 解决搜索功能内存占用问题
└── 方案3: 视频URL Redis缓存（4小时）
    └── 降低数据库和R2负载

第三阶段（1个月内）：
└── 方案5: 定时任务汇总统计（10小时）
    └── 长期可维护性优化
```

---

## 📈 预期性能提升

### 关键指标对比

| 指标 | 优化前 | 优化后 | 提升幅度 |
|------|-------|-------|---------|
| **统计API响应时间** | 500ms | 5ms | **99%** ⬇️ |
| **搜索API响应时间** | 500ms | 50ms | **90%** ⬇️ |
| **视频URL生成时间** | 100ms | 5ms | **95%** ⬇️ |
| **数据库QPS** | 200 | 40 | **80%** ⬇️ |
| **内存占用（搜索）** | 10MB | 10KB | **99.9%** ⬇️ |
| **Redis缓存命中率** | 0% | 95% | +95% ⬆️ |

### 可扩展性提升

| 数据量 | 优化前最大并发 | 优化后最大并发 | 提升 |
|--------|---------------|---------------|------|
| 1万任务 | 10人 | 50人 | **5倍** |
| 10万任务 | 5人 | 100人 | **20倍** |
| 100万任务 | 不可用 | 200人 | ♾️ |

---

## ⚠️ 风险评估

### 技术风险

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|-------|------|---------|
| Redis故障导致缓存失效 | 低 | 中 | 设置fallback到数据库，增加监控告警 |
| 复杂SQL性能不如预期 | 低 | 中 | 在测试环境用生产数据量验证EXPLAIN |
| 索引创建锁表 | 中 | 高 | 使用CONCURRENTLY，在低峰期执行 |
| 统计数据不一致 | 低 | 低 | 提供强制刷新按钮，记录刷新时间 |
| 内存泄漏 | 极低 | 高 | 压测验证，监控内存使用 |

### 业务风险

| 风险 | 缓解措施 |
|------|---------|
| 统计数据延迟（5分钟） | 向用户说明缓存策略，提供刷新按钮 |
| 历史数据迁移 | 无需迁移，优化向后兼容 |
| API行为变化 | 保持API接口不变，只优化内部实现 |

---

## 🧪 测试验证计划

### 功能测试

```bash
# 测试清单
□ 统计API返回数据正确性
□ 搜索功能排序正确性
□ 视频URL可访问性
□ Redis缓存失效后的fallback
□ 强制刷新功能
```

### 性能测试

```bash
# 使用Apache Bench进行压测
# 测试1: 统计API
ab -n 1000 -c 10 http://localhost:8080/api/admin/stats/overview

# 测试2: 搜索API
ab -n 1000 -c 10 "http://localhost:8080/api/search/tasks?page=1&page_size=10"

# 测试3: 视频URL生成
ab -n 500 -c 10 http://localhost:8080/api/admin/videos/generate-url

# 预期结果:
# - 平均响应时间 < 100ms
# - 99th百分位 < 200ms
# - 无错误响应
```

### 数据库性能测试

```sql
-- 验证索引使用
EXPLAIN ANALYZE
SELECT * FROM review_tasks
WHERE status = 'in_progress' AND claimed_at < NOW() - INTERVAL '30 minutes';

-- 验证统计查询性能
EXPLAIN (ANALYZE, BUFFERS)
SELECT /* 完整的统计查询 */ ...;

-- 预期结果:
-- - 查询时间 < 50ms
-- - 使用索引扫描而非全表扫描
-- - Buffers命中率 > 90%
```

### 监控指标

```
实施后需要持续监控:
□ API响应时间P50/P90/P99
□ 数据库连接池使用率
□ Redis缓存命中率
□ 内存使用量
□ CPU使用率
□ 错误率
```

---

## 📝 实施检查清单

### 代码修改清单

```bash
□ 修改 internal/repository/stats_repo.go - GetOverviewStats()
□ 新建 internal/services/stats_service.go - Redis缓存逻辑
□ 修改 internal/handlers/admin.go - 添加刷新参数
□ 修改 internal/repository/task_repo.go - 添加SearchTasksUnified()
□ 修改 internal/services/task_service.go - 简化SearchTasks()
□ 修改 internal/services/video_service.go - 添加Redis缓存
□ 新建 migrations/007_performance_indexes.sql
□ 新建 migrations/008_stats_aggregation.sql
□ 新建 internal/services/stats_aggregation_service.go
□ 修改 cmd/api/main.go - 启动聚合任务
```

### 部署清单

```bash
□ 备份数据库
□ 在测试环境部署验证
□ 执行数据库迁移（创建索引）
□ 部署新版本代码
□ 验证功能正常
□ 监控性能指标
□ 准备回滚方案
```

### 文档更新清单

```bash
□ 更新API文档说明缓存策略
□ 更新运维文档说明新增的定时任务
□ 更新数据库文档说明新增的索引和表
□ 编写性能优化总结报告
```

---

## 🎓 学习资源

### PostgreSQL性能优化
- [PostgreSQL EXPLAIN详解](https://www.postgresql.org/docs/current/using-explain.html)
- [PostgreSQL索引最佳实践](https://www.postgresql.org/docs/current/indexes.html)
- [SKIP LOCKED详解](https://www.2ndquadrant.com/en/blog/what-is-select-skip-locked-for-in-postgresql-9-5/)

### Redis缓存策略
- [Redis缓存设计模式](https://redis.io/docs/manual/patterns/)
- [Cache-Aside模式](https://docs.microsoft.com/en-us/azure/architecture/patterns/cache-aside)

### Go性能优化
- [Go性能分析工具pprof](https://golang.org/pkg/net/http/pprof/)
- [Go数据库连接池最佳实践](https://go.dev/doc/database/manage-connections)

---

## 📞 支持与反馈

如有疑问或需要协助，请：
- 创建GitHub Issue
- 联系项目维护者
- 查阅项目文档: `/doc/README.md`

---

**文档结束** | 最后更新: 2025-11-24 | 版本: v1.0
