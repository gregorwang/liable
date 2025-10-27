# API 接口测试指南

## 基础信息

- **Base URL**: `http://localhost:8080`
- **默认管理员**: `admin` / `admin123`

## 1. 健康检查

```bash
curl http://localhost:8080/health
```

**响应**：
```json
{"status":"healthy"}
```

---

## 2. 认证接口

### 2.1 管理员登录

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**响应**：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin",
    "status": "approved",
    "created_at": "2025-10-24T12:00:00Z",
    "updated_at": "2025-10-24T12:00:00Z"
  }
}
```

**保存 token 用于后续请求**：
```bash
export TOKEN="your_token_here"
```

### 2.2 审核员注册

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "reviewer1",
    "password": "password123"
  }'
```

**响应**：
```json
{
  "message": "Registration successful. Please wait for admin approval.",
  "user": {
    "id": 2,
    "username": "reviewer1",
    "role": "reviewer",
    "status": "pending",
    "created_at": "2025-10-24T12:05:00Z",
    "updated_at": "2025-10-24T12:05:00Z"
  }
}
```

### 2.3 获取当前用户信息

```bash
curl http://localhost:8080/api/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

---

## 3. 管理员接口

### 3.1 查看待审批用户

```bash
curl http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN"
```

### 3.2 审批用户

```bash
# 通过审批
curl -X PUT http://localhost:8080/api/admin/users/2/approve \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved"
  }'

# 拒绝审批
curl -X PUT http://localhost:8080/api/admin/users/2/approve \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "rejected"
  }'
```

### 3.3 总体统计

```bash
curl http://localhost:8080/api/admin/stats/overview \
  -H "Authorization: Bearer $TOKEN"
```

**响应示例**：
```json
{
  "total_tasks": 5323,
  "completed_tasks": 0,
  "approved_count": 0,
  "rejected_count": 0,
  "approval_rate": 0,
  "total_reviewers": 0,
  "active_reviewers": 0,
  "pending_tasks": 5323,
  "in_progress_tasks": 0
}
```

### 3.4 每小时标注量

```bash
curl "http://localhost:8080/api/admin/stats/hourly?date=2025-10-24" \
  -H "Authorization: Bearer $TOKEN"
```

### 3.5 违规类型分布

```bash
curl http://localhost:8080/api/admin/stats/tags \
  -H "Authorization: Bearer $TOKEN"
```

### 3.6 审核员绩效排行

```bash
curl "http://localhost:8080/api/admin/stats/reviewers?limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

### 3.7 标签管理

#### 查看所有标签
```bash
curl http://localhost:8080/api/admin/tags \
  -H "Authorization: Bearer $TOKEN"
```

#### 创建标签
```bash
curl -X POST http://localhost:8080/api/admin/tags \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "诈骗信息",
    "description": "包含诈骗或欺诈内容"
  }'
```

#### 更新标签
```bash
curl -X PUT http://localhost:8080/api/admin/tags/7 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "is_active": false
  }'
```

#### 删除标签
```bash
curl -X DELETE http://localhost:8080/api/admin/tags/7 \
  -H "Authorization: Bearer $TOKEN"
```

---

## 4. 审核员接口

**首先使用审核员账号登录获取 token**：

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "reviewer1",
    "password": "password123"
  }'
```

保存审核员 token：
```bash
export REVIEWER_TOKEN="reviewer_token_here"
```

### 4.1 领取任务（一次20条）

```bash
curl -X POST http://localhost:8080/api/tasks/claim \
  -H "Authorization: Bearer $REVIEWER_TOKEN"
```

**响应**：
```json
{
  "tasks": [
    {
      "id": 1,
      "comment_id": 12345,
      "reviewer_id": 2,
      "status": "in_progress",
      "claimed_at": "2025-10-24T12:10:00Z",
      "created_at": "2025-10-24T12:00:00Z",
      "comment": {
        "id": 12345,
        "text": "这是一条评论内容..."
      }
    }
    // ... 更多任务
  ],
  "count": 20
}
```

### 4.2 查看我的待审核任务

```bash
curl http://localhost:8080/api/tasks/my \
  -H "Authorization: Bearer $REVIEWER_TOKEN"
```

### 4.3 提交单个审核结果

```bash
# 通过审核
curl -X POST http://localhost:8080/api/tasks/submit \
  -H "Authorization: Bearer $REVIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": 1,
    "is_approved": true,
    "tags": [],
    "reason": ""
  }'

# 不通过审核
curl -X POST http://localhost:8080/api/tasks/submit \
  -H "Authorization: Bearer $REVIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": 2,
    "is_approved": false,
    "tags": ["广告", "垃圾"],
    "reason": "包含明显的广告推广内容"
  }'
```

### 4.4 批量提交审核结果

```bash
curl -X POST http://localhost:8080/api/tasks/submit-batch \
  -H "Authorization: Bearer $REVIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reviews": [
      {
        "task_id": 1,
        "is_approved": true,
        "tags": [],
        "reason": ""
      },
      {
        "task_id": 2,
        "is_approved": false,
        "tags": ["广告"],
        "reason": "包含推广信息"
      },
      {
        "task_id": 3,
        "is_approved": true,
        "tags": [],
        "reason": ""
      }
    ]
  }'
```

### 4.5 获取违规标签列表

```bash
curl http://localhost:8080/api/tags \
  -H "Authorization: Bearer $REVIEWER_TOKEN"
```

**响应**：
```json
{
  "tags": [
    {
      "id": 1,
      "name": "广告",
      "description": "包含广告或推广内容",
      "is_active": true,
      "created_at": "2025-10-24T12:00:00Z"
    },
    {
      "id": 2,
      "name": "垃圾",
      "description": "无意义或垃圾信息",
      "is_active": true,
      "created_at": "2025-10-24T12:00:00Z"
    }
    // ... 更多标签
  ]
}
```

---

## 5. 完整工作流程示例

### 步骤 1：管理员登录
```bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.token')
```

### 步骤 2：审核员注册
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"reviewer1","password":"password123"}'
```

### 步骤 3：管理员审批审核员
```bash
curl -X PUT http://localhost:8080/api/admin/users/2/approve \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"approved"}'
```

### 步骤 4：审核员登录
```bash
REVIEWER_TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"reviewer1","password":"password123"}' \
  | jq -r '.token')
```

### 步骤 5：审核员领取任务
```bash
curl -X POST http://localhost:8080/api/tasks/claim \
  -H "Authorization: Bearer $REVIEWER_TOKEN"
```

### 步骤 6：审核员提交审核
```bash
curl -X POST http://localhost:8080/api/tasks/submit \
  -H "Authorization: Bearer $REVIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": 1,
    "is_approved": true,
    "tags": [],
    "reason": ""
  }'
```

### 步骤 7：管理员查看统计
```bash
curl http://localhost:8080/api/admin/stats/overview \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

---

## 注意事项

1. **任务超时**：领取的任务如果 30 分钟内未提交，会自动释放回待审核队列
2. **重复领取**：审核员必须完成当前任务才能领取新任务
3. **标签选择**：不通过的审核必须选择至少一个违规标签
4. **JWT 过期**：Token 有效期为 24 小时，过期后需要重新登录

## 使用 Postman

1. 导入环境变量：
   - `BASE_URL`: `http://localhost:8080`
   - `ADMIN_TOKEN`: (登录后获取)
   - `REVIEWER_TOKEN`: (登录后获取)

2. 在请求的 Authorization 标签页选择 "Bearer Token"
3. 填入对应的 token 变量

Happy Testing! 🚀

## 获取帮助

- 📖 完整 API 文档: 查看 `TASK_QUEUE_API.md`
- 🐛 报告 Bug: 提交 Issue
- 💡 功能建议: 发起讨论
- 📧 技术支持: 联系开发团队

---

## 普通用户（Reviewer）队列查看权限测试

### 新增端点（用于普通用户）

```bash
# 1️⃣ 普通用户获取队列列表（分页）- 无需认证
curl -X GET "http://localhost:8080/api/queues?page=1&page_size=20" \
  -H "Content-Type: application/json"

# 2️⃣ 普通用户获取特定队列详情 - 无需认证
curl -X GET "http://localhost:8080/api/queues/1" \
  -H "Content-Type: application/json"
```

### 权限说明

| 操作 | 任何人 | 管理员 | 端点 | 需要认证 |
|------|--------|------|------|---------|
| 查看队列列表 | ✅ | ✅ | `/api/queues` | ❌ 否 |
| 查看队列详情 | ✅ | ✅ | `/api/queues/:id` | ❌ 否 |
| 创建队列 | ❌ | ✅ | `/api/admin/task-queues` | ✅ 是 |
| 修改队列 | ❌ | ✅ | `/api/admin/task-queues/:id` | ✅ 是 |
| 删除队列 | ❌ | ✅ | `/api/admin/task-queues/:id` | ✅ 是 |

### 最简单的测试方法

**直接在浏览器中打开**（无需认证）：
```
http://localhost:8080/api/queues?page=1&page_size=20
```

或使用curl（最简单）：
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
      "description": "审核色情和低俗内容",
      "priority": 80,
      "total_tasks": 500,
      "completed_tasks": 250,
      "pending_tasks": 250,
      "is_active": true,
      "created_at": "2025-10-26T10:00:00Z",
      "updated_at": "2025-10-26T14:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

### 高级测试步骤

如果你想测试管理员权限（创建/修改/删除）：

1. **以管理员身份登录**
```bash
curl -X POST "http://localhost:8080/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

2. **复制返回的 token**

3. **用 token 访问管理员端点**
```bash
curl -X POST "http://localhost:8080/api/admin/task-queues" \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "queue_name": "垃圾信息审核",
    "priority": 50,
    "total_tasks": 1000
  }'
```

