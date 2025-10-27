# 快速修复指南

## 最近修复：普通用户队列查看权限 (2025-10-26)

### 问题描述

后端API返回404错误，然后是403 Insufficient permissions错误：
```
GET /api/admin/task-queues?page=1&page_size=20
[GIN] 404 Not Found
或
403 Forbidden - Insufficient permissions
```

**原因**：权限验证中间件阻止了普通用户访问管理员路由。

### ✅ 解决方案（简化版）

创建了**完全公开的队列读取端点**，无需任何认证或权限检查。

#### 1. 后端路由添加 (`cmd/api/main.go`)

```go
// Public Queue Read-Only API (no auth required)
taskQueueHandler := handlers.NewTaskQueueHandler()
api.GET("/queues", taskQueueHandler.GetPublicQueues)
api.GET("/queues/:id", taskQueueHandler.GetPublicQueue)
```

#### 2. 后端处理器添加 (`internal/handlers/admin.go`)

```go
// GetPublicQueues - 获取队列列表（无认证）
func (h *TaskQueueHandler) GetPublicQueues(c *gin.Context) { ... }

// GetPublicQueue - 获取单个队列（无认证）
func (h *TaskQueueHandler) GetPublicQueue(c *gin.Context) { ... }
```

#### 3. 前端API添加 (`frontend/src/api/admin.ts`)

```typescript
// 使用公开API（无认证）
export async function listTaskQueuesPublic(params?: {...}): Promise<ListTaskQueuesResponse>
export async function getTaskQueuePublic(id: number): Promise<TaskQueue>
```

### 权限说明（清晰化）

| 操作 | 任何人都能做 | 只有管理员 | 端点 | 需要认证 |
|------|------------|----------|------|---------|
| 查看队列列表 | ✅ | - | `/api/queues` | ❌ **否** |
| 查看队列详情 | ✅ | - | `/api/queues/:id` | ❌ **否** |
| 创建队列 | ❌ | ✅ | `/api/admin/task-queues` | ✅ 是 |
| 修改队列 | ❌ | ✅ | `/api/admin/task-queues/:id` | ✅ 是 |
| 删除队列 | ❌ | ✅ | `/api/admin/task-queues/:id` | ✅ 是 |

### 测试方法（最简单的方式）

**直接在浏览器打开**（无需任何认证）：
```
http://localhost:8080/api/queues
```

**或使用curl**：
```bash
curl http://localhost:8080/api/queues
```

**预期响应**：
```json
{
  "data": [
    {
      "id": 1,
      "queue_name": "色情内容审核",
      "priority": 80,
      "total_tasks": 500,
      "completed_tasks": 250,
      "is_active": true,
      ...
    }
  ],
  "total": 1,
  "page": 1
}
```

### 核心特点

✅ **完全公开** - 无需登录即可读取队列数据  
✅ **只读** - 无法修改数据  
✅ **分页支持** - 支持 `?page=1&page_size=20` 参数  
✅ **搜索支持** - 支持 `?search=xxx` 参数  
✅ **简单高效** - 直接数据库查询，无权限开销  

### 涉及文件

- ✅ `cmd/api/main.go` - 新增公开路由
- ✅ `internal/handlers/admin.go` - 新增处理器方法
- ✅ `frontend/src/api/admin.ts` - 新增前端API函数
- ✅ `API_TESTING.md` - 更新测试文档

---

## 其他常见问题
