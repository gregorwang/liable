# 任务队列管理 API 文档

## 概述

任务队列管理系统允许管理员手动配置和管理评论审核任务队列。每个队列可以包含多个待审核任务，并跟踪其进度（总任务数、已审核数、待审核数）。

**基础URL**: `http://localhost:8080/api/admin`

**权限**: 仅限 `admin` 角色用户访问

**认证**: 所有请求都需要在请求头中包含有效的 JWT Token：
```
Authorization: Bearer <token>
```

---

## 数据模型

### TaskQueue 对象

```json
{
  "id": 1,
  "queue_name": "队列1",
  "description": "第一个审核队列",
  "priority": 100,
  "total_tasks": 150,
  "completed_tasks": 50,
  "pending_tasks": 100,
  "is_active": true,
  "created_by": 1,
  "updated_by": 1,
  "created_at": "2025-01-15T09:30:00Z",
  "updated_at": "2025-01-15T10:45:00Z"
}
```

**字段说明**:
- `id`: 队列唯一标识符（整数，自动生成）
- `queue_name`: 队列名称（字符串，1-100字符，唯一）
- `description`: 队列描述（字符串，可选）
- `priority`: 优先级（整数，0-1000，数值越大优先级越高）
- `total_tasks`: 队列中的总任务数（整数）
- `completed_tasks`: 已完成的任务数（整数）
- `pending_tasks`: 待审核任务数（计算字段，= total_tasks - completed_tasks）
- `is_active`: 是否活跃（布尔值，默认true）
- `created_by`: 创建者用户ID（可选）
- `updated_by`: 最后修改者用户ID（可选）
- `created_at`: 创建时间（ISO 8601格式）
- `updated_at`: 更新时间（ISO 8601格式）

---

## API 端点

### 1. 创建任务队列

**端点**: `POST /admin/task-queues`

**描述**: 创建新的任务队列

**请求体**:
```json
{
  "queue_name": "审核队列_01",
  "description": "用于处理投诉类内容的审核队列",
  "priority": 50,
  "total_tasks": 200,
  "completed_tasks": 0
}
```

**请求参数**:
| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| queue_name | string | ✓ | 队列名称，长度1-100字符 |
| description | string | ✗ | 队列描述 |
| priority | int | ✗ | 优先级，范围0-1000（默认0） |
| total_tasks | int | ✓ | 队列中的总任务数，最小0 |
| completed_tasks | int | ✗ | 已完成任务数，最小0（默认0） |

**成功响应** (201 Created):
```json
{
  "id": 1,
  "queue_name": "审核队列_01",
  "description": "用于处理投诉类内容的审核队列",
  "priority": 50,
  "total_tasks": 200,
  "completed_tasks": 0,
  "pending_tasks": 200,
  "is_active": true,
  "created_by": 1,
  "updated_by": 1,
  "created_at": "2025-01-15T09:30:00Z",
  "updated_at": "2025-01-15T09:30:00Z"
}
```

**错误响应** (400 Bad Request):
```json
{
  "error": "completed_tasks cannot be greater than total_tasks"
}
```

**示例 cURL**:
```bash
curl -X POST http://localhost:8080/api/admin/task-queues \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "queue_name": "审核队列_01",
    "description": "投诉类内容审核",
    "priority": 50,
    "total_tasks": 200,
    "completed_tasks": 0
  }'
```

---

### 2. 获取队列列表（分页）

**端点**: `GET /admin/task-queues`

**描述**: 获取所有任务队列列表，支持分页和过滤

**查询参数**:
| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| search | string | ✗ | 按队列名称搜索 |
| is_active | bool | ✗ | 过滤活跃状态（true/false） |
| page | int | ✗ | 页码，默认1 |
| page_size | int | ✗ | 每页数量，默认20，最大100 |

**成功响应** (200 OK):
```json
{
  "data": [
    {
      "id": 1,
      "queue_name": "审核队列_01",
      "description": "投诉类内容",
      "priority": 100,
      "total_tasks": 200,
      "completed_tasks": 50,
      "pending_tasks": 150,
      "is_active": true,
      "created_at": "2025-01-15T09:30:00Z",
      "updated_at": "2025-01-15T10:45:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

**示例 cURL**:
```bash
curl -X GET "http://localhost:8080/api/admin/task-queues?page=1&page_size=10&is_active=true" \
  -H "Authorization: Bearer <token>"
```

---

### 3. 获取单个队列详情

**端点**: `GET /admin/task-queues/:id`

**描述**: 获取指定ID的任务队列详情

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 队列ID |

**成功响应** (200 OK):
```json
{
  "id": 1,
  "queue_name": "审核队列_01",
  "description": "投诉类内容",
  "priority": 100,
  "total_tasks": 200,
  "completed_tasks": 50,
  "pending_tasks": 150,
  "is_active": true,
  "created_by": 1,
  "updated_by": 1,
  "created_at": "2025-01-15T09:30:00Z",
  "updated_at": "2025-01-15T10:45:00Z"
}
```

**错误响应** (404 Not Found):
```json
{
  "error": "Task queue not found"
}
```

**示例 cURL**:
```bash
curl -X GET http://localhost:8080/api/admin/task-queues/1 \
  -H "Authorization: Bearer <token>"
```

---

### 4. 更新队列

**端点**: `PUT /admin/task-queues/:id`

**描述**: 更新指定ID的任务队列

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 队列ID |

**请求体**（所有字段可选）:
```json
{
  "queue_name": "审核队列_01_v2",
  "description": "投诉类内容（已更新）",
  "priority": 80,
  "total_tasks": 250,
  "completed_tasks": 75,
  "is_active": true
}
```

**请求参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| queue_name | string | 新的队列名称（编辑时不可修改，仅创建时设置） |
| description | string | 新的队列描述 |
| priority | int | 新的优先级 |
| total_tasks | int | 新的总任务数 |
| completed_tasks | int | 新的已完成任务数 |
| is_active | bool | 新的活跃状态 |

**成功响应** (200 OK):
```json
{
  "id": 1,
  "queue_name": "审核队列_01",
  "description": "投诉类内容（已更新）",
  "priority": 80,
  "total_tasks": 250,
  "completed_tasks": 75,
  "pending_tasks": 175,
  "is_active": true,
  "created_by": 1,
  "updated_by": 1,
  "created_at": "2025-01-15T09:30:00Z",
  "updated_at": "2025-01-15T11:20:00Z"
}
```

**错误响应** (400 Bad Request):
```json
{
  "error": "completed_tasks cannot be greater than total_tasks"
}
```

**示例 cURL**:
```bash
curl -X PUT http://localhost:8080/api/admin/task-queues/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "priority": 80,
    "total_tasks": 250,
    "completed_tasks": 75
  }'
```

---

### 5. 删除队列

**端点**: `DELETE /admin/task-queues/:id`

**描述**: 删除（软删除）指定ID的任务队列

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 队列ID |

**成功响应** (200 OK):
```json
{
  "message": "Task queue deleted successfully"
}
```

**错误响应** (400 Bad Request):
```json
{
  "error": "Invalid queue ID"
}
```

**示例 cURL**:
```bash
curl -X DELETE http://localhost:8080/api/admin/task-queues/1 \
  -H "Authorization: Bearer <token>"
```

---

### 6. 获取所有活跃队列

**端点**: `GET /admin/task-queues-all`

**描述**: 获取所有活跃的任务队列（不分页）

**成功响应** (200 OK):
```json
{
  "queues": [
    {
      "id": 1,
      "queue_name": "审核队列_01",
      "description": "投诉类内容",
      "priority": 100,
      "total_tasks": 200,
      "completed_tasks": 50,
      "pending_tasks": 150,
      "is_active": true,
      "created_at": "2025-01-15T09:30:00Z",
      "updated_at": "2025-01-15T10:45:00Z"
    }
  ],
  "count": 1
}
```

**示例 cURL**:
```bash
curl -X GET http://localhost:8080/api/admin/task-queues-all \
  -H "Authorization: Bearer <token>"
```

---

## 常见用例

### 用例 1: 创建新的审核队列
```bash
# 创建一个名为"色情内容审核"的队列，初始1000个待审任务
curl -X POST http://localhost:8080/api/admin/task-queues \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "queue_name": "色情内容审核",
    "description": "色情内容审核专用队列",
    "priority": 90,
    "total_tasks": 1000,
    "completed_tasks": 0
  }'
```

### 用例 2: 更新队列的审核进度
```bash
# 当审核人员完成了100个任务后，更新队列
curl -X PUT http://localhost:8080/api/admin/task-queues/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "completed_tasks": 100
  }'
```

### 用例 3: 搜索并过滤队列
```bash
# 搜索名称包含"色情"的活跃队列
curl -X GET "http://localhost:8080/api/admin/task-queues?search=%E8%89%B2%E6%83%85&is_active=true&page=1&page_size=20" \
  -H "Authorization: Bearer <token>"
```

### 用例 4: 禁用队列
```bash
# 禁用指定队列（soft delete）
curl -X DELETE http://localhost:8080/api/admin/task-queues/1 \
  -H "Authorization: Bearer <token>"
```

---

## 错误处理

### 常见错误码

| 错误码 | 说明 | 示例 |
|--------|------|------|
| 400 | 请求参数错误 | 完成任务数大于总任务数 |
| 401 | 未认证 | Token 缺失或无效 |
| 403 | 无权限 | 非admin角色 |
| 404 | 资源不存在 | 队列ID不存在 |
| 500 | 服务器错误 | 数据库异常 |

### 错误响应格式

```json
{
  "error": "error message here"
}
```

---

## 前端集成指南

### 1. API 模块（TypeScript）

```typescript
import * as adminAPI from '@/api/admin'

// 创建队列
const queue = await adminAPI.createTaskQueue({
  queue_name: '色情内容审核',
  description: '色情内容审核队列',
  priority: 90,
  total_tasks: 1000
})

// 获取列表
const response = await adminAPI.listTaskQueues({
  search: '色情',
  is_active: true,
  page: 1,
  page_size: 10
})

// 更新队列
await adminAPI.updateTaskQueue(1, {
  completed_tasks: 100,
  is_active: true
})

// 删除队列
await adminAPI.deleteTaskQueue(1)
```

### 2. Vue 组件集成

QueueManage 组件位于 `frontend/src/views/admin/QueueManage.vue`，提供完整的队列管理UI：

- 队列列表显示
- 搜索和过滤
- 添加新队列
- 编辑现有队列
- 删除队列
- 进度可视化

---

## 性能指标

- **查询响应时间**: < 100ms（单条查询）
- **列表响应时间**: < 500ms（20条记录）
- **并发处理**: 支持100+并发请求
- **数据库索引**: 在 queue_name、is_active、priority 字段上建立索引

---

## 最佳实践

1. **队列优先级管理**: 使用1-100的优先级范围
2. **进度更新频率**: 建议每完成一定数量任务后更新一次
3. **队列命名规范**: 使用清晰的命名，如"审核类型_分片_版本"
4. **错误处理**: 客户端应实现重试机制（指数退避）
5. **缓存策略**: 可在客户端缓存"所有活跃队列"列表5分钟

---

## 更新日志

### v1.0.0 (2025-01-15)
- 初始发布
- 实现CRUD操作
- 支持分页和搜索
- 提供进度追踪
