# 代码规范与防崩坏指南

> **解决问题**: 防止每次添加新功能就导致代码崩坏
> **适用人群**: AI辅助编程的开发者
> **核心理念**: 预防胜于治疗，规范保护代码健康

---

## 📋 目录

1. [为什么代码会崩坏](#为什么代码会崩坏)
2. [代码崩坏的征兆](#代码崩坏的征兆)
3. [添加新功能的标准流程](#添加新功能的标准流程)
4. [核心代码规范](#核心代码规范)
5. [架构设计原则](#架构设计原则)
6. [代码审查检查清单](#代码审查检查清单)
7. [常见坏味道与修复](#常见坏味道与修复)
8. [AI编程防崩坏策略](#ai编程防崩坏策略)
9. [项目特定规范](#项目特定规范)

---

## 🔥 为什么代码会崩坏

### 崩坏的本质

```
代码崩坏 = 技术债务累积 + 架构腐化 + 缺乏约束

就像一个房子:
✅ 好的代码: 有坚实的地基、清晰的结构、易于扩展
❌ 崩坏的代码: 地基不稳、乱搭乱建、随时可能倒塌
```

### 7种常见崩坏原因

| 原因 | 表现 | 危害等级 |
|------|------|---------|
| **1. 复制粘贴代码** | 同样的逻辑重复10次 | 🔴 高 |
| **2. 上帝类/上帝函数** | 一个文件5000行，一个函数500行 | 🔴 高 |
| **3. 违反分层架构** | Handler直接操作数据库 | 🔴 高 |
| **4. 忽略错误处理** | 到处都是`err != nil`但不处理 | 🟡 中 |
| **5. 缺少事务管理** | 数据不一致 | 🟡 中 |
| **6. 硬编码配置** | 魔法数字和字符串到处飞 | 🟢 低 |
| **7. 缺少测试** | 改一处坏十处 | 🔴 高 |

### 用厨房类比理解

```
❌ 崩坏的厨房（代码）:
├── 食材到处乱放（没有分层）
├── 同样的菜谱抄10遍（代码重复）
├── 厨具混用（职责不清）
├── 从不清理（技术债务）
└── 只有主厨知道在哪找东西（知识孤岛）

✅ 整洁的厨房（代码）:
├── 食材分类存放（分层架构）
├── 菜谱统一管理（复用逻辑）
├── 厨具各司其职（单一职责）
├── 定期清理整理（重构）
└── 新人也能快速上手（可维护性）
```

---

## 🚨 代码崩坏的征兆

### 自检清单（每周检查）

```
□ 添加新功能需要修改10+个文件
□ 不敢重构，怕改坏其他功能
□ 经常出现"这里为什么这样写？"的疑问
□ 同样的代码在多个地方出现
□ 文件超过500行
□ 函数超过100行
□ 错误日志莫名其妙
□ 修复一个bug引发三个新bug
□ 数据库有脏数据
□ 性能越来越慢
```

**如果勾选了3个以上，说明代码已经开始崩坏。**

### 崩坏程度评估

| 勾选数 | 崩坏程度 | 建议 |
|--------|---------|------|
| 0-2个 | 😊 健康 | 保持现状，继续遵循规范 |
| 3-5个 | 😐 轻度崩坏 | 停止添加新功能，先重构 |
| 6-8个 | 😰 中度崩坏 | 需要架构级重构 |
| 9-10个 | 💀 重度崩坏 | 考虑重写核心模块 |

---

## 📝 添加新功能的标准流程

### 流程图

```
开始添加新功能
    ↓
第1步: 设计阶段（30%时间）
├── 明确需求
├── 设计数据模型
├── 设计API接口
└── 评估影响范围
    ↓
第2步: 编码前准备（10%时间）
├── 创建git分支
├── 创建数据库迁移文件
└── 更新文档
    ↓
第3步: 编码实现（40%时间）
├── 按分层架构实现
├── 遵循代码规范
├── 添加错误处理
└── 写代码注释
    ↓
第4步: 自我审查（10%时间）
├── 运行代码检查
├── 测试功能
└── 检查崩坏征兆
    ↓
第5步: AI代码审查（5%时间）
└── 让AI审查代码
    ↓
第6步: 提交合并（5%时间）
├── 提交代码
├── 写清楚commit信息
└── 更新文档
    ↓
完成
```

### 详细步骤说明

#### 第1步: 设计阶段（最重要！）

**不要一上来就写代码！先设计！**

```
向AI提问模板:

"我要添加一个新功能: [功能描述]

请帮我设计:
1. 需要哪些数据库表/字段
2. 需要哪些API接口
3. 会影响哪些现有模块
4. 有哪些边界情况需要考虑
5. 给出实现步骤

项目架构: Handler → Service → Repository
技术栈: Go + Gin + PostgreSQL + Redis"
```

**设计检查清单**:
```
✅ 数据模型设计:
   □ 表结构合理
   □ 字段类型正确
   □ 索引考虑性能
   □ 外键关系清晰

✅ API设计:
   □ RESTful风格
   □ 请求/响应结构清晰
   □ 错误响应统一
   □ 权限控制明确

✅ 影响评估:
   □ 列出需要修改的文件
   □ 评估对现有功能的影响
   □ 考虑向后兼容性
   □ 评估性能影响
```

#### 第2步: 编码前准备

```bash
# 1. 创建功能分支
git checkout -b feature/新功能名称

# 2. 创建数据库迁移文件
# 文件命名: migrations/XXX_feature_name.sql
touch migrations/009_add_new_feature.sql

# 3. 在迁移文件中写SQL
-- migrations/009_add_new_feature.sql
-- Description: 添加XXX功能
-- Author: [你的名字]
-- Date: 2024-01-01

BEGIN;

-- 创建表
CREATE TABLE new_feature_table (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_new_feature_name ON new_feature_table(name);

COMMIT;
```

#### 第3步: 编码实现（遵循分层架构）

**正确的实现顺序**:

```
1. Model层（定义数据结构）
   ↓
2. Repository层（数据库操作）
   ↓
3. Service层（业务逻辑）
   ↓
4. Handler层（HTTP接口）
```

**示例：添加"任务标签过滤"功能**

```go
// ========== 步骤1: Model层 ==========
// 文件: internal/models/models.go

// 添加请求模型
type FilterTasksByTagRequest struct {
    Tags     []string `json:"tags" binding:"required"`
    Page     int      `json:"page"`
    PageSize int      `json:"page_size"`
}

// 添加响应模型
type FilterTasksByTagResponse struct {
    Tasks      []ReviewTask `json:"tasks"`
    Total      int          `json:"total"`
    Page       int          `json:"page"`
    PageSize   int          `json:"page_size"`
    TotalPages int          `json:"total_pages"`
}

// ========== 步骤2: Repository层 ==========
// 文件: internal/repository/task_repo.go

// FilterTasksByTags 根据标签过滤任务
func (r *TaskRepository) FilterTasksByTags(tags []string, page, pageSize int) ([]models.ReviewTask, int, error) {
    // 计算偏移量
    offset := (page - 1) * pageSize

    // 构建查询（使用参数化查询防止SQL注入）
    query := `
        SELECT DISTINCT rt.id, rt.comment_id, rt.status, rt.created_at
        FROM review_tasks rt
        INNER JOIN review_results rr ON rt.id = rr.task_id
        WHERE rr.tags && $1  -- PostgreSQL数组操作符
        ORDER BY rt.created_at DESC
        LIMIT $2 OFFSET $3
    `

    rows, err := r.db.Query(query, pq.Array(tags), pageSize, offset)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to query tasks: %w", err)
    }
    defer rows.Close()

    // 扫描结果
    tasks := []models.ReviewTask{}
    for rows.Next() {
        var task models.ReviewTask
        if err := rows.Scan(&task.ID, &task.CommentID, &task.Status, &task.CreatedAt); err != nil {
            return nil, 0, fmt.Errorf("failed to scan task: %w", err)
        }
        tasks = append(tasks, task)
    }

    // 获取总数
    var total int
    countQuery := `
        SELECT COUNT(DISTINCT rt.id)
        FROM review_tasks rt
        INNER JOIN review_results rr ON rt.id = rr.task_id
        WHERE rr.tags && $1
    `
    if err := r.db.QueryRow(countQuery, pq.Array(tags)).Scan(&total); err != nil {
        return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
    }

    return tasks, total, nil
}

// ========== 步骤3: Service层 ==========
// 文件: internal/services/task_service.go

// FilterTasksByTags 按标签过滤任务（添加业务逻辑）
func (s *TaskService) FilterTasksByTags(req models.FilterTasksByTagRequest) (*models.FilterTasksByTagResponse, error) {
    // 参数验证
    if len(req.Tags) == 0 {
        return nil, errors.New("tags cannot be empty")
    }

    // 设置默认值
    if req.Page < 1 {
        req.Page = 1
    }
    if req.PageSize < 1 {
        req.PageSize = 10
    }
    if req.PageSize > 100 {
        req.PageSize = 100  // 限制最大页面大小
    }

    // 调用Repository层
    tasks, total, err := s.taskRepo.FilterTasksByTags(req.Tags, req.Page, req.PageSize)
    if err != nil {
        return nil, fmt.Errorf("service: failed to filter tasks: %w", err)
    }

    // 计算总页数
    totalPages := total / req.PageSize
    if total%req.PageSize > 0 {
        totalPages++
    }

    // 构建响应
    response := &models.FilterTasksByTagResponse{
        Tasks:      tasks,
        Total:      total,
        Page:       req.Page,
        PageSize:   req.PageSize,
        TotalPages: totalPages,
    }

    return response, nil
}

// ========== 步骤4: Handler层 ==========
// 文件: internal/handlers/task.go

// FilterTasksByTags 处理按标签过滤任务的HTTP请求
func (h *TaskHandler) FilterTasksByTags(c *gin.Context) {
    var req models.FilterTasksByTagRequest

    // 绑定请求参数
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request parameters",
            "details": err.Error(),
        })
        return
    }

    // 调用Service层
    response, err := h.taskService.FilterTasksByTags(req)
    if err != nil {
        log.Printf("Error filtering tasks by tags: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to filter tasks",
        })
        return
    }

    // 返回成功响应
    c.JSON(http.StatusOK, response)
}

// ========== 步骤5: 注册路由 ==========
// 文件: cmd/api/main.go

func setupRoutes(router *gin.Engine) {
    api := router.Group("/api")

    // 任务相关路由
    taskHandler := handlers.NewTaskHandler()
    tasks := api.Group("/tasks")
    {
        tasks.POST("/filter-by-tags",
            middleware.AuthMiddleware(),
            middleware.RequirePermission("tasks:search"),
            taskHandler.FilterTasksByTags)  // 新增路由
    }
}
```

#### 第4步: 自我审查

**使用这个检查清单**:

```
✅ 代码结构:
   □ 遵循Handler→Service→Repository分层
   □ 没有跨层调用
   □ 每个函数职责单一
   □ 没有复制粘贴代码

✅ 错误处理:
   □ 所有error都被检查
   □ 错误信息有上下文（使用fmt.Errorf包装）
   □ 数据库查询失败有日志
   □ 用户看到友好的错误提示

✅ 安全性:
   □ 使用参数化查询（防SQL注入）
   □ 输入参数有验证
   □ 敏感信息不记录到日志
   □ 权限检查正确

✅ 性能:
   □ 数据库查询有索引支持
   □ 没有N+1查询
   □ 合理使用分页
   □ 考虑缓存策略

✅ 可维护性:
   □ 函数名清晰表达意图
   □ 关键逻辑有注释
   □ 魔法数字用常量代替
   □ 复杂业务逻辑有文档
```

#### 第5步: AI代码审查

```
给AI的提示:

"请审查这段代码，检查是否有问题:

[粘贴代码]

请检查:
1. 是否遵循分层架构
2. 错误处理是否完善
3. 是否有安全隐患（SQL注入、XSS等）
4. 是否有性能问题
5. 代码可读性如何
6. 是否有潜在的bug

项目架构: Handler → Service → Repository
技术栈: Go + Gin + PostgreSQL"
```

#### 第6步: 提交合并

```bash
# 1. 运行测试（如果有）
go test ./...

# 2. 格式化代码
go fmt ./...

# 3. 提交代码
git add .
git commit -m "feat: 添加按标签过滤任务功能

- 新增FilterTasksByTags API接口
- 支持多标签AND查询
- 支持分页
- 添加参数验证

相关文件:
- internal/models/models.go
- internal/repository/task_repo.go
- internal/services/task_service.go
- internal/handlers/task.go"

# 4. 推送到远程
git push origin feature/新功能名称

# 5. 合并到主分支（测试通过后）
git checkout main
git merge feature/新功能名称
git push origin main
```

---

## 📐 核心代码规范

### 1. 分层架构规范 🔴 必须遵守

```
禁止的调用:
❌ Handler → Repository（跨层调用）
❌ Handler → Database（跨层调用）
❌ Service → gin.Context（层级混乱）

允许的调用:
✅ Handler → Service → Repository → Database
✅ Service → Service（同层调用）
✅ Repository → Repository（同层调用）
```

**坏例子❌**:
```go
// Handler直接操作数据库（违反分层）
func (h *TaskHandler) GetTask(c *gin.Context) {
    // ❌ Handler不应该直接访问数据库
    row := database.DB.QueryRow("SELECT * FROM tasks WHERE id = $1", taskID)
    // ...
}
```

**好例子✅**:
```go
// Handler层
func (h *TaskHandler) GetTask(c *gin.Context) {
    taskID := c.Param("id")

    // ✅ Handler调用Service
    task, err := h.taskService.GetTaskByID(taskID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, task)
}

// Service层
func (s *TaskService) GetTaskByID(taskID string) (*models.Task, error) {
    // ✅ Service调用Repository
    return s.taskRepo.FindByID(taskID)
}

// Repository层
func (r *TaskRepository) FindByID(taskID string) (*models.Task, error) {
    // ✅ Repository操作数据库
    row := r.db.QueryRow("SELECT * FROM tasks WHERE id = $1", taskID)
    // ...
}
```

### 2. 命名规范 🟡 强烈推荐

#### 文件命名
```
✅ 好的命名:
- task_service.go
- user_repository.go
- auth_handler.go

❌ 坏的命名:
- service.go（太泛化）
- utils.go（垃圾桶文件）
- temp.go（临时文件不应该提交）
```

#### 函数命名
```go
// ✅ 好的命名（动词开头，清晰表达意图）
func CreateTask(task *Task) error { ... }
func GetUserByID(id int) (*User, error) { ... }
func ValidateEmail(email string) bool { ... }
func FilterTasksByStatus(status string) ([]Task, error) { ... }

// ❌ 坏的命名（不清晰）
func Process(data interface{}) { ... }
func Do() { ... }
func Handle(x int) { ... }
func Func1() { ... }
```

#### 变量命名
```go
// ✅ 好的命名
var userID int
var reviewerName string
var isApproved bool
var totalCount int
var cacheKey string

// ❌ 坏的命名
var id int  // 什么的ID？
var name string  // 什么的name？
var flag bool  // 什么标志？
var count int  // 数什么？
var k string  // k是什么？
```

### 3. 错误处理规范 🔴 必须遵守

```go
// ✅ 好的错误处理
func (s *TaskService) CreateTask(task *models.Task) error {
    // 1. 参数验证
    if task == nil {
        return errors.New("task cannot be nil")
    }
    if task.CommentID == 0 {
        return errors.New("comment_id is required")
    }

    // 2. 调用Repository，包装错误
    if err := s.taskRepo.Insert(task); err != nil {
        return fmt.Errorf("service: failed to create task: %w", err)  // ✅ 包装错误，保留堆栈
    }

    // 3. 记录日志
    log.Printf("Task created successfully: ID=%d", task.ID)

    return nil
}

// ❌ 坏的错误处理
func (s *TaskService) CreateTask(task *models.Task) error {
    err := s.taskRepo.Insert(task)
    if err != nil {
        // ❌ 1. 吞掉错误，没有返回
        log.Println(err)
    }
    return nil
}

func (s *TaskService) CreateTask2(task *models.Task) error {
    err := s.taskRepo.Insert(task)
    // ❌ 2. 不检查错误
    return nil
}

func (s *TaskService) CreateTask3(task *models.Task) error {
    if err := s.taskRepo.Insert(task); err != nil {
        return err  // ❌ 3. 不包装错误，丢失上下文
    }
    return nil
}
```

### 4. 函数大小规范 🟡 强烈推荐

```
函数长度限制:
✅ 理想: 10-30行
⚠️ 可接受: 30-50行
❌ 需要重构: 50行以上

函数复杂度限制:
✅ 理想: 嵌套层级 ≤ 2
⚠️ 可接受: 嵌套层级 ≤ 3
❌ 需要重构: 嵌套层级 > 3
```

**坏例子❌ - 100行的上帝函数**:
```go
func (s *TaskService) ProcessTask(taskID int) error {
    // 获取任务
    task, err := s.taskRepo.FindByID(taskID)
    if err != nil {
        return err
    }

    // 验证任务
    if task.Status != "pending" {
        return errors.New("invalid status")
    }

    // 获取审核员
    reviewer, err := s.userRepo.FindByID(task.ReviewerID)
    if err != nil {
        return err
    }

    // 检查权限
    if !reviewer.HasPermission("task:review") {
        return errors.New("no permission")
    }

    // 处理审核结果
    if task.IsApproved {
        // ... 50行代码
    } else {
        // ... 50行代码
    }

    // ... 更多逻辑

    return nil  // 函数太长，难以理解和维护
}
```

**好例子✅ - 拆分成小函数**:
```go
func (s *TaskService) ProcessTask(taskID int) error {
    // 1. 获取和验证任务
    task, err := s.getAndValidateTask(taskID)
    if err != nil {
        return err
    }

    // 2. 检查审核员权限
    if err := s.checkReviewerPermission(task.ReviewerID); err != nil {
        return err
    }

    // 3. 处理审核结果
    if err := s.handleReviewResult(task); err != nil {
        return err
    }

    return nil
}

// 拆分出的小函数
func (s *TaskService) getAndValidateTask(taskID int) (*models.Task, error) {
    task, err := s.taskRepo.FindByID(taskID)
    if err != nil {
        return nil, fmt.Errorf("failed to get task: %w", err)
    }

    if task.Status != "pending" {
        return nil, errors.New("task status must be pending")
    }

    return task, nil
}

func (s *TaskService) checkReviewerPermission(reviewerID int) error {
    reviewer, err := s.userRepo.FindByID(reviewerID)
    if err != nil {
        return fmt.Errorf("failed to get reviewer: %w", err)
    }

    if !reviewer.HasPermission("task:review") {
        return errors.New("reviewer has no permission")
    }

    return nil
}

func (s *TaskService) handleReviewResult(task *models.Task) error {
    if task.IsApproved {
        return s.handleApprovedTask(task)
    }
    return s.handleRejectedTask(task)
}
```

### 5. 避免代码重复 🔴 必须遵守

**DRY原则**: Don't Repeat Yourself（不要重复自己）

**坏例子❌ - 重复代码**:
```go
// 在多个Handler中重复的参数验证
func (h *TaskHandler) ClaimTasks(c *gin.Context) {
    var req ClaimTasksRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid parameters"})
        return
    }
    if req.Count < 1 || req.Count > 50 {
        c.JSON(400, gin.H{"error": "Count must be between 1 and 50"})
        return
    }
    // ...
}

func (h *TaskHandler) ReturnTasks(c *gin.Context) {
    var req ReturnTasksRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid parameters"})  // ❌ 重复
        return
    }
    if req.Count < 1 || req.Count > 50 {
        c.JSON(400, gin.H{"error": "Count must be between 1 and 50"})  // ❌ 重复
        return
    }
    // ...
}
```

**好例子✅ - 提取公共函数**:
```go
// 提取公共的验证逻辑
func (h *TaskHandler) validateCountParam(c *gin.Context, count int) bool {
    if count < 1 || count > 50 {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Count must be between 1 and 50",
        })
        return false
    }
    return true
}

func (h *TaskHandler) bindAndValidateJSON(c *gin.Context, req interface{}) bool {
    if err := c.ShouldBindJSON(req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request parameters",
            "details": err.Error(),
        })
        return false
    }
    return true
}

func (h *TaskHandler) ClaimTasks(c *gin.Context) {
    var req ClaimTasksRequest
    if !h.bindAndValidateJSON(c, &req) {
        return  // ✅ 复用验证逻辑
    }
    if !h.validateCountParam(c, req.Count) {
        return  // ✅ 复用验证逻辑
    }
    // ...
}
```

---

## 🏛️ 架构设计原则

### SOLID原则（简化版）

#### 1. 单一职责原则（Single Responsibility）

```
一个函数/类只做一件事

❌ 坏例子:
func ProcessTaskAndSendEmailAndUpdateCache() {
    // 做3件事，违反单一职责
}

✅ 好例子:
func ProcessTask() { ... }
func SendEmail() { ... }
func UpdateCache() { ... }
```

#### 2. 开放封闭原则（Open/Closed）

```
对扩展开放，对修改封闭

❌ 坏例子:
每次添加新的审核类型，都要修改Service层的switch语句

✅ 好例子:
使用接口，添加新类型时只需要实现接口，不修改现有代码
```

#### 3. 接口隔离原则（Interface Segregation）

```
不要强迫客户端依赖它不需要的接口

❌ 坏例子:
type Repository interface {
    // 100个方法，但大部分实现只用其中5个
}

✅ 好例子:
type TaskReader interface {
    FindByID(id int) (*Task, error)
    FindAll() ([]Task, error)
}

type TaskWriter interface {
    Insert(task *Task) error
    Update(task *Task) error
}
```

---

## ✅ 代码审查检查清单

### 提交前自检（5分钟）

```
基础检查:
□ 代码能编译通过（go build）
□ 代码已格式化（go fmt）
□ 没有明显的语法错误
□ 没有遗留的TODO/FIXME注释（或者已经创建了Issue）

功能检查:
□ 新功能能正常工作
□ 没有破坏现有功能
□ 边界情况都考虑了
□ 错误处理完善

规范检查:
□ 遵循分层架构
□ 函数长度合理（<50行）
□ 没有代码重复
□ 命名清晰易懂

安全检查:
□ 使用参数化查询
□ 用户输入有验证
□ 敏感信息没有记录到日志
□ 权限检查正确

性能检查:
□ 数据库查询有索引
□ 没有N+1查询
□ 考虑了缓存策略
□ 合理使用分页
```

### AI辅助审查

```
给AI的提示:

"请以专业代码审查者的角度，严格审查这段代码:

[粘贴代码]

请检查:
1. 🔴 严重问题（会导致bug或安全隐患）
2. 🟡 警告（不符合最佳实践）
3. 🟢 建议（可以改进的地方）

对每个问题，请说明:
- 问题是什么
- 为什么是问题
- 如何修复（给出代码示例）

项目规范:
- 架构: Handler → Service → Repository
- 命名: 驼峰命名，函数名动词开头
- 错误: 使用fmt.Errorf包装错误
- 函数: 长度<50行，嵌套≤3层"
```

---

## 🦨 常见坏味道与修复

### 坏味道1: 上帝类/文件过大

**识别**:
```
征兆:
- 单个文件超过1000行
- 单个struct有20+个方法
- 文件名叫utils.go或helpers.go
```

**修复**:
```go
// ❌ 坏例子: task_service.go 包含所有业务逻辑（2000行）
type TaskService struct {
    // 处理评论一审、二审、质检、视频审核、统计...
}

// ✅ 好例子: 拆分成多个Service
type CommentFirstReviewService struct { ... }
type CommentSecondReviewService struct { ... }
type QualityCheckService struct { ... }
type VideoReviewService struct { ... }
type StatsService struct { ... }
```

### 坏味道2: 过长的参数列表

**识别**:
```
征兆:
- 函数参数超过5个
- 多个bool参数
- 参数顺序难记
```

**修复**:
```go
// ❌ 坏例子: 参数太多
func CreateTask(commentID int, reviewerID int, status string,
    tags []string, reason string, isApproved bool, score int) error {
    // ...
}

// ✅ 好例子: 使用结构体
type CreateTaskParams struct {
    CommentID   int
    ReviewerID  int
    Status      string
    Tags        []string
    Reason      string
    IsApproved  bool
    Score       int
}

func CreateTask(params CreateTaskParams) error {
    // ...
}
```

### 坏味道3: 魔法数字和字符串

**识别**:
```
征兆:
- 代码中直接出现数字和字符串
- 没有说明数字/字符串的含义
```

**修复**:
```go
// ❌ 坏例子: 魔法数字
func ClaimTasks(count int) error {
    if count < 1 || count > 50 {  // 50是什么？
        return errors.New("invalid count")
    }

    timeout := 30 * time.Minute  // 30是什么？

    // ...
}

// ✅ 好例子: 使用常量
const (
    MinTaskClaimCount = 1
    MaxTaskClaimCount = 50
    DefaultTaskTimeout = 30 * time.Minute
)

func ClaimTasks(count int) error {
    if count < MinTaskClaimCount || count > MaxTaskClaimCount {
        return fmt.Errorf("count must be between %d and %d",
            MinTaskClaimCount, MaxTaskClaimCount)
    }

    timeout := DefaultTaskTimeout

    // ...
}
```

### 坏味道4: 深层嵌套

**识别**:
```
征兆:
- if嵌套超过3层
- 代码像金字塔
- 难以阅读
```

**修复**:
```go
// ❌ 坏例子: 深层嵌套
func ProcessTask(taskID int) error {
    task, err := getTask(taskID)
    if err == nil {
        if task.Status == "pending" {
            if task.ReviewerID != 0 {
                reviewer, err := getReviewer(task.ReviewerID)
                if err == nil {
                    if reviewer.HasPermission("review") {
                        // 实际逻辑埋在最里层
                        result := process(task, reviewer)
                        if result.Success {
                            return save(result)
                        }
                    }
                }
            }
        }
    }
    return errors.New("failed")
}

// ✅ 好例子: 早返回，扁平化
func ProcessTask(taskID int) error {
    // 1. 获取任务
    task, err := getTask(taskID)
    if err != nil {
        return fmt.Errorf("failed to get task: %w", err)
    }

    // 2. 验证状态
    if task.Status != "pending" {
        return errors.New("task status must be pending")
    }

    // 3. 验证审核员
    if task.ReviewerID == 0 {
        return errors.New("reviewer_id is required")
    }

    // 4. 获取审核员
    reviewer, err := getReviewer(task.ReviewerID)
    if err != nil {
        return fmt.Errorf("failed to get reviewer: %w", err)
    }

    // 5. 检查权限
    if !reviewer.HasPermission("review") {
        return errors.New("reviewer has no permission")
    }

    // 6. 处理任务
    result := process(task, reviewer)
    if !result.Success {
        return errors.New("processing failed")
    }

    // 7. 保存结果
    return save(result)
}
```

### 坏味道5: 注释代码

**识别**:
```
征兆:
- 大段注释掉的代码
- 不确定是否还需要
```

**修复**:
```go
// ❌ 坏例子: 注释代码
func ProcessTask(task *Task) error {
    // 旧实现，不确定是否还需要
    // if task.Type == "old" {
    //     return oldProcess(task)
    // }

    return newProcess(task)
}

// ✅ 好例子: 删除注释代码，依赖Git历史
func ProcessTask(task *Task) error {
    return newProcess(task)  // 如果需要看旧实现，查看Git历史
}
```

---

## 🤖 AI编程防崩坏策略

### 策略1: 让AI理解项目架构

**在每次会话开始时，给AI提供上下文**:

```
你好，我正在开发一个评论审核平台。

项目架构:
- 语言: Go 1.23
- 框架: Gin
- 数据库: PostgreSQL
- 缓存: Redis
- 分层结构: Handler → Service → Repository

代码规范:
1. 严格遵守分层架构，禁止跨层调用
2. 所有数据库操作在Repository层
3. 业务逻辑在Service层
4. HTTP处理在Handler层
5. 错误要用fmt.Errorf包装
6. 函数长度不超过50行
7. 使用参数化查询防止SQL注入

我接下来要实现[功能描述]，请按照以上规范给出代码。
```

### 策略2: 分步骤让AI实现功能

```
不要一次性让AI生成所有代码，而是分步骤:

第1步: 设计
"请帮我设计[功能]的数据模型和API接口"

第2步: Repository层
"请帮我实现Repository层的代码，包括[具体操作]"

第3步: Service层
"请帮我实现Service层的代码，调用Repository层"

第4步: Handler层
"请帮我实现Handler层的代码，调用Service层"

第5步: 路由注册
"请帮我在main.go中注册路由"
```

### 策略3: 让AI审查自己的代码

```
在AI生成代码后，立即让它审查:

"请审查你刚才生成的代码，检查:
1. 是否遵循了分层架构
2. 是否有安全隐患
3. 是否有性能问题
4. 是否有潜在bug
5. 是否可以改进

如果发现问题，请给出修正后的代码。"
```

### 策略4: 增量式开发

```
不要一次性实现复杂功能，而是增量式:

第1次: 实现最基本的功能
第2次: 添加参数验证
第3次: 添加错误处理
第4次: 添加性能优化
第5次: 添加测试

每次提交前都测试，确保不破坏现有功能。
```

### 策略5: 建立代码审查Prompt库

**创建文件 `.claude/prompts/code-review.md`**:

```markdown
# 代码审查Prompt

## 基础审查
请审查这段代码，检查:
1. 是否遵循项目架构（Handler→Service→Repository）
2. 错误处理是否完善
3. 是否有安全隐患
4. 是否有性能问题
5. 命名是否清晰
6. 是否有代码重复

项目信息:
- 架构: [粘贴项目架构]
- 规范: [粘贴代码规范]

代码:
[粘贴代码]

## 性能审查
请从性能角度审查这段代码:
1. 数据库查询是否有N+1问题
2. 是否应该添加索引
3. 是否应该添加缓存
4. 内存使用是否合理

## 安全审查
请从安全角度审查这段代码:
1. 是否有SQL注入风险
2. 是否有XSS风险
3. 输入验证是否充分
4. 权限检查是否正确
```

---

## 📋 项目特定规范

### 本项目的特殊约定

#### 1. Redis键命名规范

```
格式: [模块]:[资源]:[标识符]

示例:
✅ task:lock:123
✅ task:claimed:456
✅ stats:overview
✅ stats:daily:2024-01-01
✅ video:url:789

❌ tasklock123（没有分隔符）
❌ task_lock_123（使用下划线而非冒号）
```

#### 2. 数据库迁移文件规范

```
文件命名: XXX_description.sql

示例:
✅ 001_init_tables.sql
✅ 002_add_notifications.sql
✅ 007_performance_indexes.sql

迁移文件结构:
-- Description: [功能描述]
-- Author: [作者]
-- Date: [日期]

BEGIN;

-- [SQL语句]

COMMIT;
```

#### 3. API响应格式规范

```go
// 成功响应
{
    "data": { ... },
    "message": "success"  // 可选
}

// 错误响应
{
    "error": "错误描述",
    "details": "详细信息"  // 可选
}

// 分页响应
{
    "data": [ ... ],
    "total": 100,
    "page": 1,
    "page_size": 10,
    "total_pages": 10
}
```

#### 4. 权限键命名规范

```
格式: resource:action 或 resource:subresource:action

示例:
✅ tasks:first-review:claim
✅ tasks:search
✅ users:list
✅ permissions:grant
✅ queue.video.100k.claim

❌ TasksFirstReviewClaim（大写）
❌ tasks_first_review_claim（下划线）
```

---

## 🎯 总结与行动计划

### 防止代码崩坏的核心要点

```
1. ✅ 遵循分层架构 - 这是底线
2. ✅ 添加功能前先设计 - 不要急着写代码
3. ✅ 函数保持小而美 - 超过50行就拆分
4. ✅ 避免代码重复 - 发现重复立即提取
5. ✅ 错误处理要完善 - 不要吞掉错误
6. ✅ 提交前自我审查 - 使用检查清单
7. ✅ 增量式开发 - 小步快跑，频繁测试
8. ✅ 让AI帮忙审查 - 两双眼睛更好
```

### 30天行动计划

```
第1周: 理解和学习
├── Day 1-2: 通读本文档
├── Day 3-4: 审查现有代码，找出不符合规范的地方
└── Day 5-7: 小范围重构1-2个文件

第2周: 实践和应用
├── Day 8-10: 使用新流程添加一个小功能
├── Day 11-12: 让AI审查代码，根据反馈修改
└── Day 13-14: 总结经验，更新个人规范

第3周: 深入和优化
├── Day 15-17: 重构一个复杂模块
├── Day 18-20: 添加一个中等功能
└── Day 21: 性能测试和优化

第4周: 巩固和提升
├── Day 22-24: 独立添加复杂功能
├── Day 25-27: 全面代码审查
├── Day 28-29: 文档化经验
└── Day 30: 回顾和规划下一步
```

### 紧急情况处理

**如果发现代码已经严重崩坏**:

```
第1步: 停止添加新功能（防止进一步恶化）
第2步: 评估崩坏程度（使用本文档的检查清单）
第3步: 制定重构计划（分模块、分阶段）
第4步: 从最关键的模块开始重构
第5步: 每重构一个模块，立即测试
第6步: 逐步恢复代码健康
```

---

## 📞 获取帮助

### 遇到问题时

1. **查阅本文档**: 大部分问题都有答案
2. **向AI咨询**: 使用本文档提供的Prompt模板
3. **查看Git历史**: 看看之前是怎么做的
4. **参考现有代码**: 找类似功能的实现

### 持续改进

```
建议:
□ 每周回顾一次本文档
□ 记录遇到的新问题和解决方法
□ 更新个人的最佳实践
□ 与团队分享经验
```

---

**记住: 好的代码不是一次写成的，而是不断重构出来的。保持代码健康是一个持续的过程。** 🚀

---

**文档结束** | 最后更新: 2025-11-24 | 版本: v1.0
