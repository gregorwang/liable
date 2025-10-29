# 权限键值对系统开发文档

## 目录
1. [系统概述](#系统概述)
2. [数据库设计](#数据库设计)
3. [权限键定义](#权限键定义)
4. [代码架构设计](#代码架构设计)
5. [API接口设计](#api接口设计)
6. [实现步骤](#实现步骤)
7. [迁移方案](#迁移方案)
8. [使用示例](#使用示例)
9. [权限管理最佳实践](#权限管理最佳实践)

---

## 系统概述

### 设计理念

从基于角色的访问控制（RBAC）迁移到基于权限键值对的细粒度权限控制系统。每个功能点都有独立的权限键（Permission Key），用户可以拥有多个权限键的组合，实现更灵活的权限管理。

### 核心优势

1. **细粒度控制**：每个API端点可以独立配置权限
2. **灵活分配**：可以为用户分配任意权限组合
3. **易于扩展**：新增功能只需添加新的权限键
4. **便于审计**：记录权限授予者和授予时间
5. **向后兼容**：保留角色字段，支持平滑迁移

### 系统架构

```
请求 → AuthMiddleware（认证） → RequirePermission（权限检查） → Handler（业务处理）
                                                              ↓
                                                         PermissionService（权限服务）
                                                              ↓
                                                         PermissionRepository（数据访问）
```

---

## 数据库设计

### 1. 权限定义表（permissions）

存储系统中所有可用的权限定义。

```sql
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    permission_key VARCHAR(100) UNIQUE NOT NULL,  -- 权限键，如 "notifications:create"
    name VARCHAR(100) NOT NULL,                    -- 权限名称，如 "创建通知"
    description TEXT,                              -- 权限描述
    resource VARCHAR(100) NOT NULL,                -- 资源类型，如 "notifications", "users", "stats"
    action VARCHAR(50) NOT NULL,                  -- 操作类型，如 "create", "read", "update", "delete"
    category VARCHAR(50),                         -- 分类，如 "用户管理", "统计查看", "审核规则"
    is_active BOOLEAN NOT NULL DEFAULT TRUE,       -- 是否启用
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_permissions_key ON permissions(permission_key);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_category ON permissions(category);
CREATE INDEX idx_permissions_active ON permissions(is_active);
```

**字段说明**：
- `permission_key`: 权限的唯一标识符，格式为 `resource:action` 或 `resource:subresource:action`
- `resource`: 资源类型，对应业务模块
- `action`: 操作类型（create, read, update, delete）
- `category`: 权限分类，用于前端分组显示
- `is_active`: 是否启用，禁用的权限不会分配给用户

### 2. 用户权限关联表（user_permissions）

存储用户与权限的多对多关系。

```sql
CREATE TABLE IF NOT EXISTS user_permissions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_key VARCHAR(100) NOT NULL REFERENCES permissions(permission_key) ON DELETE CASCADE,
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    granted_by INTEGER REFERENCES users(id),      -- 权限授予者
    UNIQUE(user_id, permission_key)
);

CREATE INDEX idx_user_permissions_user ON user_permissions(user_id);
CREATE INDEX idx_user_permissions_key ON user_permissions(permission_key);
CREATE INDEX idx_user_permissions_user_key ON user_permissions(user_id, permission_key);
```

**字段说明**：
- `user_id`: 用户ID
- `permission_key`: 权限键
- `granted_at`: 权限授予时间
- `granted_by`: 权限授予者（记录审计信息）

### 3. 权限键命名规范

权限键采用层级命名方式：

```
格式: resource:action 或 resource:subresource:action

示例:
- users:list          # 查看用户列表
- users:approve       # 审批用户
- stats:overview      # 查看概览统计
- tags:create         # 创建标签
- notifications:create # 创建通知
- tasks:claim         # 领取审核任务
- moderation-rules:create # 创建审核规则
```

---

## 权限键定义

### 完整权限键列表

根据当前系统的API路由，定义以下权限键：

#### 用户管理权限
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('users:list', '查看用户列表', '查看待审批用户列表', 'users', 'read', '用户管理'),
('users:approve', '审批用户', '审批或拒绝用户注册申请', 'users', 'update', '用户管理');
```

#### 统计查看权限
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('stats:overview', '查看概览统计', '查看系统概览统计数据', 'stats', 'read', '统计查看'),
('stats:hourly', '查看每小时统计', '查看每小时统计数据', 'stats', 'read', '统计查看'),
('stats:tags', '查看标签统计', '查看标签统计数据', 'stats', 'read', '统计查看'),
('stats:reviewers', '查看审核员绩效', '查看审核员绩效统计', 'stats', 'read', '统计查看');
```

#### 标签管理权限
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('tags:list', '查看标签列表', '查看所有标签', 'tags', 'read', '标签管理'),
('tags:create', '创建标签', '创建新标签', 'tags', 'create', '标签管理'),
('tags:update', '更新标签', '更新标签信息', 'tags', 'update', '标签管理'),
('tags:delete', '删除标签', '删除标签', 'tags', 'delete', '标签管理');
```

#### 审核规则管理权限
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('moderation-rules:create', '创建审核规则', '创建新的审核规则', 'moderation-rules', 'create', '审核规则管理'),
('moderation-rules:update', '更新审核规则', '更新审核规则', 'moderation-rules', 'update', '审核规则管理'),
('moderation-rules:delete', '删除审核规则', '删除审核规则', 'moderation-rules', 'delete', '审核规则管理');
```

#### 任务队列管理权限
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('task-queues:list', '查看任务队列', '查看任务队列列表', 'task-queues', 'read', '队列管理'),
('task-queues:create', '创建任务队列', '创建新的任务队列', 'task-queues', 'create', '队列管理'),
('task-queues:update', '更新任务队列', '更新任务队列配置', 'task-queues', 'update', '队列管理'),
('task-queues:delete', '删除任务队列', '删除任务队列', 'task-queues', 'delete', '队列管理');
```

#### 通知管理权限
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('notifications:create', '创建通知', '创建系统通知', 'notifications', 'create', '通知管理');
```

#### 视频管理权限
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('videos:import', '导入视频', '批量导入视频到系统', 'videos', 'create', '视频管理'),
('videos:list', '查看视频列表', '查看视频列表', 'videos', 'read', '视频管理'),
('videos:read', '查看视频详情', '查看视频详细信息', 'videos', 'read', '视频管理'),
('videos:generate-url', '生成视频URL', '生成视频预签名URL', 'videos', 'read', '视频管理');
```

#### 审核任务权限（按队列细分）

审核任务权限按照不同的队列类型进行细分，确保用户只能访问被授权的队列：

##### 一审队列权限（评论审核任务）
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('tasks:first-review:claim', '领取一审任务', '领取评论一审审核任务', 'tasks', 'create', '审核任务-一审队列'),
('tasks:first-review:submit', '提交一审结果', '提交评论一审审核结果', 'tasks', 'update', '审核任务-一审队列'),
('tasks:first-review:return', '归还一审任务', '归还已领取的一审任务', 'tasks', 'update', '审核任务-一审队列');
```

##### 二审队列权限（二次审核任务）
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('tasks:second-review:claim', '领取二审任务', '领取评论二审审核任务', 'tasks', 'create', '审核任务-二审队列'),
('tasks:second-review:submit', '提交二审结果', '提交评论二审审核结果', 'tasks', 'update', '审核任务-二审队列'),
('tasks:second-review:return', '归还二审任务', '归还已领取的二审任务', 'tasks', 'update', '审核任务-二审队列');
```

##### 质检队列权限（质量检查任务）
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('tasks:quality-check:claim', '领取质检任务', '领取质量检查任务', 'tasks', 'create', '审核任务-质检队列'),
('tasks:quality-check:submit', '提交质检结果', '提交质量检查结果', 'tasks', 'update', '审核任务-质检队列'),
('tasks:quality-check:return', '归还质检任务', '归还已领取的质检任务', 'tasks', 'update', '审核任务-质检队列'),
('tasks:quality-check:stats', '查看质检统计', '查看质检统计数据', 'tasks', 'read', '审核任务-质检队列');
```

##### 抖音短视频一审队列权限
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('tasks:video-first-review:claim', '领取视频初审任务', '领取抖音短视频初审任务', 'tasks', 'create', '审核任务-视频一审队列'),
('tasks:video-first-review:submit', '提交视频初审结果', '提交抖音短视频初审结果', 'tasks', 'update', '审核任务-视频一审队列'),
('tasks:video-first-review:return', '归还视频初审任务', '归还已领取的视频初审任务', 'tasks', 'update', '审核任务-视频一审队列');
```

##### 抖音短视频二审队列权限
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('tasks:video-second-review:claim', '领取视频二审任务', '领取抖音短视频二审任务', 'tasks', 'create', '审核任务-视频二审队列'),
('tasks:video-second-review:submit', '提交视频二审结果', '提交抖音短视频二审结果', 'tasks', 'update', '审核任务-视频二审队列'),
('tasks:video-second-review:return', '归还视频二审任务', '归还已领取的视频二审任务', 'tasks', 'update', '审核任务-视频二审队列');
```

##### 通用任务权限（所有队列共享）
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('tasks:search', '搜索任务', '搜索审核任务记录（所有队列）', 'tasks', 'read', '审核任务-通用');
```

**权限细分说明**：

审核任务权限按照5个不同的队列类型进行细分，每个队列都有独立的权限控制：

1. **一审队列** (`tasks:first-review:*`): 评论审核任务的一审队列
2. **二审队列** (`tasks:second-review:*`): 评论审核任务的二审队列
3. **质检队列** (`tasks:quality-check:*`): 质量检查任务队列
4. **视频一审队列** (`tasks:video-first-review:*`): 抖音短视频审核的一审队列
5. **视频二审队列** (`tasks:video-second-review:*`): 抖音短视频审核的二审队列

**权限控制逻辑**：
- 每个队列的 `claim`、`submit`、`return` 权限都是独立的
- 用户必须拥有对应队列的权限才能访问该队列
- 例如：只有拥有 `tasks:first-review:claim` 权限的用户才能领取一审任务
- 例如：只有拥有 `tasks:quality-check:stats` 权限的用户才能查看质检统计

**权限分配示例**：
- 只负责一审的审核员：只授予 `tasks:first-review:*` 权限
- 只负责质检的审核员：只授予 `tasks:quality-check:*` 权限
- 负责多个队列的审核员：授予对应队列的所有权限
- 通用搜索权限：`tasks:search` 可以单独授予，允许搜索所有队列的任务记录

#### 权限管理权限（元权限）
```sql
INSERT INTO permissions (permission_key, name, description, resource, action, category) VALUES
('permissions:read', '查看权限列表', '查看系统中的所有权限定义', 'permissions', 'read', '权限管理'),
('permissions:grant', '授予权限', '授予用户权限', 'permissions', 'create', '权限管理'),
('permissions:revoke', '撤销权限', '撤销用户权限', 'permissions', 'delete', '权限管理');
```

---

## 代码架构设计

### 1. 数据模型（Models）

**文件位置**: `internal/models/models.go`

```go
// Permission represents a permission definition
type Permission struct {
    ID          int       `json:"id"`
    PermissionKey string  `json:"permission_key"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Resource    string    `json:"resource"`
    Action      string    `json:"action"`
    Category    string    `json:"category"`
    IsActive    bool      `json:"is_active"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// UserPermission represents user-permission relationship
type UserPermission struct {
    ID          int       `json:"id"`
    UserID      int       `json:"user_id"`
    PermissionKey string  `json:"permission_key"`
    GrantedAt   time.Time `json:"granted_at"`
    GrantedBy   *int      `json:"granted_by,omitempty"`
}

// User model 添加权限列表字段（可选）
type User struct {
    ID          int          `json:"id"`
    Username    string       `json:"username"`
    Password    string       `json:"-"`
    Role        string       `json:"role"`        // 保留用于兼容
    Status      string       `json:"status"`
    Permissions []string     `json:"permissions,omitempty"` // 权限键列表
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
}

// Permission management DTOs
type GrantPermissionRequest struct {
    UserID         int      `json:"user_id" binding:"required"`
    PermissionKeys []string `json:"permission_keys" binding:"required,min=1"`
}

type RevokePermissionRequest struct {
    UserID         int      `json:"user_id" binding:"required"`
    PermissionKeys []string `json:"permission_keys" binding:"required,min=1"`
}

type ListUserPermissionsRequest struct {
    UserID   int    `form:"user_id"`
    Category string `form:"category"`
}

type ListPermissionsRequest struct {
    Resource string `form:"resource"`
    Category string `form:"category"`
    Search   string `form:"search"`
    Page     int    `form:"page"`
    PageSize int    `form:"page_size"`
}

type ListPermissionsResponse struct {
    Data       []Permission `json:"data"`
    Total      int          `json:"total"`
    Page       int          `json:"page"`
    PageSize   int          `json:"page_size"`
    TotalPages int          `json:"total_pages"`
}
```

### 2. 数据访问层（Repository）

**文件位置**: `internal/repository/permission_repo.go`

**核心方法**：
- `GetAllPermissions()` - 获取所有权限
- `GetPermissionByKey(key string)` - 根据键获取权限
- `GetUserPermissions(userID int)` - 获取用户的所有权限键
- `HasPermission(userID int, permissionKey string)` - 检查用户是否拥有权限
- `GrantPermissions(userID int, permissionKeys []string, grantedBy *int)` - 授予权限
- `RevokePermissions(userID int, permissionKeys []string)` - 撤销权限
- `ListPermissions(resource, category, search string, page, pageSize int)` - 分页查询权限

### 3. 业务逻辑层（Service）

**文件位置**: `internal/services/permission_service.go`

**核心方法**：
- `GetUserPermissions(userID int)` - 获取用户权限列表
- `HasPermission(userID int, permissionKey string)` - 检查权限
- `GrantPermissions(userID int, permissionKeys []string, grantedBy int)` - 授予权限
- `RevokePermissions(userID int, permissionKeys []string)` - 撤销权限
- `GetAllPermissions()` - 获取所有权限定义
- `ListPermissions(...)` - 分页查询权限

### 4. 中间件层（Middleware）

**文件位置**: `internal/middleware/permission.go`

```go
// RequirePermission 检查用户是否拥有指定权限
func RequirePermission(permissionKey string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := GetUserID(c)
        if userID == 0 {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
            c.Abort()
            return
        }

        hasPermission, err := permissionService.HasPermission(userID, permissionKey)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
            c.Abort()
            return
        }

        if !hasPermission {
            c.JSON(http.StatusForbidden, gin.H{
                "error":            "Insufficient permissions",
                "required_permission": permissionKey,
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

// RequireAnyPermission 检查用户是否拥有任意一个权限
func RequireAnyPermission(permissionKeys ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := GetUserID(c)
        if userID == 0 {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
            c.Abort()
            return
        }

        for _, key := range permissionKeys {
            hasPermission, err := permissionService.HasPermission(userID, key)
            if err == nil && hasPermission {
                c.Next()
                return
            }
        }

        c.JSON(http.StatusForbidden, gin.H{
            "error":               "Insufficient permissions",
            "required_permissions": permissionKeys,
        })
        c.Abort()
    }
}
```

### 5. 处理器层（Handler）

**文件位置**: `internal/handlers/admin.go`（权限管理方法添加到AdminHandler）

**核心方法**：
- `ListPermissions()` - 获取权限列表
- `GetAllPermissions()` - 获取所有权限（不分页）
- `GetUserPermissions()` - 获取用户权限
- `GrantPermissions()` - 授予权限
- `RevokePermissions()` - 撤销权限

---

## API接口设计

### 权限管理API

#### 1. 获取权限列表（分页）

```http
GET /api/admin/permissions?resource=notifications&category=通知管理&search=创建&page=1&page_size=20
Authorization: Bearer <token>
```

**响应**：
```json
{
    "data": [
        {
            "id": 1,
            "permission_key": "notifications:create",
            "name": "创建通知",
            "description": "创建系统通知",
            "resource": "notifications",
            "action": "create",
            "category": "通知管理",
            "is_active": true,
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20,
    "total_pages": 3
}
```

#### 2. 获取所有权限（不分页）

```http
GET /api/admin/permissions/all
Authorization: Bearer <token>
```

**响应**：
```json
{
    "permissions": [
        {
            "id": 1,
            "permission_key": "notifications:create",
            "name": "创建通知",
            ...
        }
    ]
}
```

#### 3. 获取用户权限

```http
GET /api/admin/permissions/user?user_id=2&category=统计查看
Authorization: Bearer <token>
```

**响应**：
```json
{
    "user_id": 2,
    "permissions": [
        "stats:overview",
        "stats:hourly",
        "stats:tags"
    ]
}
```

#### 4. 授予用户权限

```http
POST /api/admin/permissions/grant
Authorization: Bearer <token>
Content-Type: application/json

{
    "user_id": 2,
    "permission_keys": [
        "stats:overview",
        "stats:hourly",
        "notifications:create"
    ]
}
```

**响应**：
```json
{
    "message": "Permissions granted successfully",
    "user_id": 2,
    "permissions": [
        "stats:overview",
        "stats:hourly",
        "notifications:create"
    ]
}
```

#### 5. 撤销用户权限

```http
POST /api/admin/permissions/revoke
Authorization: Bearer <token>
Content-Type: application/json

{
    "user_id": 2,
    "permission_keys": [
        "stats:hourly"
    ]
}
```

**响应**：
```json
{
    "message": "Permissions revoked successfully",
    "user_id": 2,
    "permissions": [
        "stats:hourly"
    ]
}
```

### 路由权限映射

以下是将现有路由改为权限检查的映射关系：

| 路由 | HTTP方法 | 权限键 | 说明 |
|------|----------|--------|------|
| `/api/admin/users` | GET | `users:list` | 查看用户列表 |
| `/api/admin/users/:id/approve` | PUT | `users:approve` | 审批用户 |
| `/api/admin/stats/overview` | GET | `stats:overview` | 查看概览统计 |
| `/api/admin/stats/hourly` | GET | `stats:hourly` | 查看每小时统计 |
| `/api/admin/stats/tags` | GET | `stats:tags` | 查看标签统计 |
| `/api/admin/stats/reviewers` | GET | `stats:reviewers` | 查看审核员绩效 |
| `/api/admin/tags` | GET | `tags:list` | 查看标签列表 |
| `/api/admin/tags` | POST | `tags:create` | 创建标签 |
| `/api/admin/tags/:id` | PUT | `tags:update` | 更新标签 |
| `/api/admin/tags/:id` | DELETE | `tags:delete` | 删除标签 |
| `/api/admin/moderation-rules` | POST | `moderation-rules:create` | 创建规则 |
| `/api/admin/moderation-rules/:id` | PUT | `moderation-rules:update` | 更新规则 |
| `/api/admin/moderation-rules/:id` | DELETE | `moderation-rules:delete` | 删除规则 |
| `/api/admin/task-queues` | GET | `task-queues:list` | 查看队列列表 |
| `/api/admin/task-queues` | POST | `task-queues:create` | 创建队列 |
| `/api/admin/task-queues/:id` | PUT | `task-queues:update` | 更新队列 |
| `/api/admin/task-queues/:id` | DELETE | `task-queues:delete` | 删除队列 |
| `/api/admin/notifications` | POST | `notifications:create` | 创建通知 |
| `/api/admin/videos/import` | POST | `videos:import` | 导入视频 |
| `/api/admin/videos` | GET | `videos:list` | 查看视频列表 |
| `/api/admin/videos/:id` | GET | `videos:read` | 查看视频详情 |
| `/api/admin/videos/generate-url` | POST | `videos:generate-url` | 生成视频URL |
| `/api/tasks/claim` | POST | `tasks:first-review:claim` | 领取一审任务 |
| `/api/tasks/submit` | POST | `tasks:first-review:submit` | 提交一审结果 |
| `/api/tasks/return` | POST | `tasks:first-review:return` | 归还一审任务 |
| `/api/tasks/search` | GET | `tasks:search` | 搜索任务（所有队列） |
| `/api/tasks/second-review/claim` | POST | `tasks:second-review:claim` | 领取二审任务 |
| `/api/tasks/second-review/submit` | POST | `tasks:second-review:submit` | 提交二审结果 |
| `/api/tasks/second-review/return` | POST | `tasks:second-review:return` | 归还二审任务 |
| `/api/tasks/quality-check/claim` | POST | `tasks:quality-check:claim` | 领取质检任务 |
| `/api/tasks/quality-check/submit` | POST | `tasks:quality-check:submit` | 提交质检结果 |
| `/api/tasks/quality-check/return` | POST | `tasks:quality-check:return` | 归还质检任务 |
| `/api/tasks/quality-check/stats` | GET | `tasks:quality-check:stats` | 查看质检统计 |
| `/api/tasks/video-first-review/claim` | POST | `tasks:video-first-review:claim` | 领取视频初审任务 |
| `/api/tasks/video-first-review/submit` | POST | `tasks:video-first-review:submit` | 提交视频初审结果 |
| `/api/tasks/video-first-review/return` | POST | `tasks:video-first-review:return` | 归还视频初审任务 |
| `/api/tasks/video-second-review/claim` | POST | `tasks:video-second-review:claim` | 领取视频二审任务 |
| `/api/tasks/video-second-review/submit` | POST | `tasks:video-second-review:submit` | 提交视频二审结果 |
| `/api/tasks/video-second-review/return` | POST | `tasks:video-second-review:return` | 归还视频二审任务 |

---

## 实现步骤

### 第一阶段：数据库准备

1. **创建数据库迁移文件**
   - 创建 `migrations/004_permission_system.sql`
   - 包含权限表和用户权限关联表的创建语句
   - 包含所有权限定义的插入语句

2. **执行数据库迁移**
   ```sql
   -- 在Supabase中执行迁移文件
   ```

3. **为现有admin用户分配所有权限**
   ```sql
   INSERT INTO user_permissions (user_id, permission_key)
   SELECT (SELECT id FROM users WHERE username = 'admin' LIMIT 1), permission_key 
   FROM permissions;
   ```

### 第二阶段：代码实现

1. **创建数据模型**
   - 在 `internal/models/models.go` 中添加权限相关模型

2. **创建Repository层**
   - 创建 `internal/repository/permission_repo.go`
   - 实现所有数据访问方法

3. **创建Service层**
   - 创建 `internal/services/permission_service.go`
   - 实现业务逻辑方法

4. **创建中间件**
   - 创建 `internal/middleware/permission.go`
   - 实现权限检查中间件

5. **扩展AdminHandler**
   - 在 `internal/handlers/admin.go` 中添加权限管理方法

### 第三阶段：路由更新

1. **添加权限管理路由**
   ```go
   admin.GET("/permissions", middleware.RequirePermission("permissions:read"), adminHandler.ListPermissions)
   admin.GET("/permissions/all", middleware.RequirePermission("permissions:read"), adminHandler.GetAllPermissions)
   admin.GET("/permissions/user", middleware.RequirePermission("permissions:read"), adminHandler.GetUserPermissions)
   admin.POST("/permissions/grant", middleware.RequirePermission("permissions:grant"), adminHandler.GrantPermissions)
   admin.POST("/permissions/revoke", middleware.RequirePermission("permissions:revoke"), adminHandler.RevokePermissions)
   ```

2. **更新现有路由**
   - 逐步将路由从角色检查改为权限检查
   - 保留角色检查作为后备方案

### 第四阶段：测试与验证

1. **单元测试**
   - 测试Repository方法
   - 测试Service方法
   - 测试中间件逻辑

2. **集成测试**
   - 测试权限授予流程
   - 测试权限撤销流程
   - 测试权限检查中间件

3. **功能测试**
   - 测试所有API接口的权限控制
   - 验证权限检查正确性

---

## 迁移方案

### 平滑迁移策略

为了不影响现有功能，采用渐进式迁移：

1. **双模式运行**
   - 权限检查优先
   - 如果权限检查失败，回退到角色检查
   - 确保向后兼容

2. **迁移步骤**
   ```
   步骤1: 部署权限系统（不影响现有功能）
   步骤2: 为所有用户分配权限（基于角色）
   步骤3: 逐步切换路由到权限检查
   步骤4: 验证所有功能正常
   步骤5: 移除角色检查代码
   ```

3. **基于角色的权限分配脚本**

```sql
-- 为所有admin用户分配管理员权限
INSERT INTO user_permissions (user_id, permission_key)
SELECT u.id, p.permission_key
FROM users u
CROSS JOIN permissions p
WHERE u.role = 'admin'
AND p.is_active = true
ON CONFLICT (user_id, permission_key) DO NOTHING;

-- 为所有reviewer用户分配审核员权限（按队列细分）
INSERT INTO user_permissions (user_id, permission_key)
SELECT u.id, p.permission_key
FROM users u
CROSS JOIN permissions p
WHERE u.role = 'reviewer'
AND p.permission_key IN (
    -- 一审队列权限
    'tasks:first-review:claim',
    'tasks:first-review:submit',
    'tasks:first-review:return',
    -- 二审队列权限
    'tasks:second-review:claim',
    'tasks:second-review:submit',
    'tasks:second-review:return',
    -- 质检队列权限
    'tasks:quality-check:claim',
    'tasks:quality-check:submit',
    'tasks:quality-check:return',
    'tasks:quality-check:stats',
    -- 视频一审队列权限
    'tasks:video-first-review:claim',
    'tasks:video-first-review:submit',
    'tasks:video-first-review:return',
    -- 视频二审队列权限
    'tasks:video-second-review:claim',
    'tasks:video-second-review:submit',
    'tasks:video-second-review:return',
    -- 通用权限
    'tasks:search'
)
AND p.is_active = true
ON CONFLICT (user_id, permission_key) DO NOTHING;
```

---

## 使用示例

### 示例1：授予用户统计查看权限

```bash
# 请求
curl -X POST http://localhost:8080/api/admin/permissions/grant \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "permission_keys": [
      "stats:overview",
      "stats:hourly",
      "stats:tags"
    ]
  }'

# 响应
{
    "message": "Permissions granted successfully",
    "user_id": 2,
    "permissions": [
        "stats:overview",
        "stats:hourly",
        "stats:tags"
    ]
}
```

### 示例2：撤销用户权限

```bash
# 请求
curl -X POST http://localhost:8080/api/admin/permissions/revoke \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "permission_keys": ["stats:hourly"]
  }'
```

### 示例3：查看用户权限

```bash
# 请求
curl -X GET "http://localhost:8080/api/admin/permissions/user?user_id=2" \
  -H "Authorization: Bearer <admin_token>"

# 响应
{
    "user_id": 2,
    "permissions": [
        "stats:overview",
        "stats:tags",
        "notifications:create"
    ]
}
```

### 示例4：在路由中使用权限检查

```go
// 原来的方式（角色检查）
admin.GET("/stats/overview", middleware.AuthMiddleware(), middleware.RequireAdmin(), adminHandler.GetOverviewStats)

// 新的方式（权限检查）
admin.GET("/stats/overview", middleware.AuthMiddleware(), middleware.RequirePermission("stats:overview"), adminHandler.GetOverviewStats)
```

### 示例5：检查多个权限（任意一个）

```go
// 用户需要拥有任意一个权限
admin.GET("/stats/all", 
    middleware.AuthMiddleware(), 
    middleware.RequireAnyPermission("stats:overview", "stats:hourly", "stats:tags"), 
    adminHandler.GetAllStats)
```

### 示例6：按队列授予审核任务权限

```bash
# 只授予用户一审队列权限
curl -X POST http://localhost:8080/api/admin/permissions/grant \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 3,
    "permission_keys": [
      "tasks:first-review:claim",
      "tasks:first-review:submit",
      "tasks:first-review:return"
    ]
  }'

# 只授予用户质检队列权限
curl -X POST http://localhost:8080/api/admin/permissions/grant \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 4,
    "permission_keys": [
      "tasks:quality-check:claim",
      "tasks:quality-check:submit",
      "tasks:quality-check:return",
      "tasks:quality-check:stats"
    ]
  }'

# 授予用户多个队列权限（一审 + 视频一审）
curl -X POST http://localhost:8080/api/admin/permissions/grant \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 5,
    "permission_keys": [
      "tasks:first-review:claim",
      "tasks:first-review:submit",
      "tasks:first-review:return",
      "tasks:video-first-review:claim",
      "tasks:video-first-review:submit",
      "tasks:video-first-review:return"
    ]
  }'
```

**说明**：
- 用户3只能访问一审队列（评论审核任务）
- 用户4只能访问质检队列
- 用户5可以同时访问一审队列和视频一审队列
- 每个队列的权限都是独立的，互不影响

---

## 权限管理最佳实践

### 1. 权限命名规范

- 使用小写字母和冒号分隔
- 格式：`resource:action` 或 `resource:subresource:action`
- 保持命名一致性和可读性

### 2. 权限分配原则

- **最小权限原则**：只授予用户完成工作所需的最小权限集
- **定期审查**：定期审查用户权限，移除不必要的权限
- **权限分组**：按功能模块分配权限，便于管理

### 3. 权限缓存策略

为了提高性能，可以考虑：

```go
// 在用户登录时，将权限列表加载到Redis
// 权限检查时优先从Redis读取
// 权限变更时更新Redis缓存
```

### 4. 权限审计

- 记录所有权限授予和撤销操作
- 记录操作者信息（granted_by）
- 定期生成权限审计报告

### 5. 错误处理

- 权限检查失败时返回明确的错误信息
- 记录权限检查失败的日志
- 区分认证失败和权限不足

### 6. 性能优化

- 批量查询用户权限（避免N+1问题）
- 使用索引优化权限查询
- 考虑使用Redis缓存权限信息

### 7. 权限继承（可选）

如果需要更复杂的权限系统，可以考虑：

```sql
-- 创建权限组表
CREATE TABLE permission_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT
);

-- 权限组-权限关联表
CREATE TABLE permission_group_permissions (
    group_id INTEGER REFERENCES permission_groups(id),
    permission_key VARCHAR(100) REFERENCES permissions(permission_key),
    PRIMARY KEY (group_id, permission_key)
);

-- 用户-权限组关联表
CREATE TABLE user_permission_groups (
    user_id INTEGER REFERENCES users(id),
    group_id INTEGER REFERENCES permission_groups(id),
    PRIMARY KEY (user_id, group_id)
);
```

---

## 注意事项

1. **向后兼容**
   - 保留角色字段用于兼容
   - 支持角色和权限双重检查的过渡期

2. **数据一致性**
   - 权限变更时确保数据一致性
   - 使用事务保证操作的原子性

3. **安全性**
   - 权限检查必须在使用权限之前执行
   - 不要在客户端暴露权限信息
   - 验证权限键的有效性

4. **性能考虑**
   - 权限检查会增加数据库查询
   - 考虑使用缓存减少数据库压力
   - 批量查询优化

5. **测试覆盖**
   - 确保所有权限检查路径都有测试覆盖
   - 测试边界情况和异常情况

---

## 总结

本文档提供了从RBAC到权限键值对系统的完整迁移方案。该系统具有以下特点：

1. **灵活性强**：可以为用户分配任意权限组合
2. **扩展性好**：新增功能只需添加权限定义
3. **可维护性高**：权限管理集中化，便于维护
4. **向后兼容**：支持平滑迁移，不影响现有功能

通过遵循本文档的步骤，可以安全、高效地实现权限系统的迁移和升级。

---

**文档版本**: 1.0  
**最后更新**: 2024年  
**维护者**: 开发团队

