# 审核记录搜索API使用指南

## 接口信息

- **路径**: `GET /api/tasks/search`
- **权限**: 需要登录（管理员和审核员都可以访问）
- **说明**: 根据多个条件搜索已完成的审核记录

## 请求参数

所有参数都是可选的，可以组合使用：

| 参数名 | 类型 | 说明 | 示例 |
|--------|------|------|------|
| `comment_id` | int64 | 评论ID（精确匹配） | `12345` |
| `reviewer_rtx` | string | 审核员账号（精确匹配） | `"2669297147@qq.com"` |
| `tag_ids` | string | 违规标签ID，逗号分隔（OR关系） | `"1,2,3"` |
| `review_start_time` | datetime | 审核开始时间（ISO 8601格式） | `"2025-10-01T00:00:00Z"` |
| `review_end_time` | datetime | 审核结束时间（ISO 8601格式） | `"2025-10-24T23:59:59Z"` |
| `page` | int | 页码，默认1 | `1` |
| `page_size` | int | 每页数量，默认10，最大100 | `20` |

## 查询逻辑

- 所有条件都是 **AND** 关系
- 审核时间范围：`review_start_time <= completed_at <= review_end_time`
- `tag_ids`：任务的标签包含其中任意一个即可（**OR** 关系）
- 只返回 `status = 'completed'` 的任务

## 响应格式

```json
{
  "data": [
    {
      "id": 41,
      "comment_id": 41,
      "comment_text": "这是评论内容",
      "reviewer_id": 3,
      "username": "2669297147@qq.com",
      "status": "completed",
      "claimed_at": "2025-10-24T07:37:45Z",
      "completed_at": "2025-10-24T07:37:55Z",
      "created_at": "2025-10-24T07:30:00Z",
      "review_id": 1,
      "is_approved": true,
      "tags": [],
      "reason": null,
      "reviewed_at": "2025-10-24T07:37:55Z"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "total_pages": 5
}
```

## 响应字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int | 任务ID |
| `comment_id` | int64 | 评论ID |
| `comment_text` | string | 评论内容 |
| `reviewer_id` | int | 审核员ID |
| `username` | string | 审核员账号 |
| `status` | string | 任务状态（固定为completed） |
| `claimed_at` | datetime | 领取时间 |
| `completed_at` | datetime | 完成时间 |
| `created_at` | datetime | 创建时间 |
| `review_id` | int | 审核结果ID |
| `is_approved` | bool | 是否通过 |
| `tags` | array | 违规标签列表 |
| `reason` | string | 审核原因 |
| `reviewed_at` | datetime | 审核时间 |

## 使用示例

### 1. 搜索特定审核员的所有记录

```bash
curl -X GET "http://localhost:8080/api/tasks/search?reviewer_rtx=2669297147@qq.com&page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 2. 搜索特定评论的审核记录

```bash
curl -X GET "http://localhost:8080/api/tasks/search?comment_id=12345" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. 搜索指定时间范围内的审核记录

```bash
curl -X GET "http://localhost:8080/api/tasks/search?review_start_time=2025-10-01T00:00:00Z&review_end_time=2025-10-24T23:59:59Z&page=1&page_size=50" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. 搜索包含特定标签的审核记录

```bash
curl -X GET "http://localhost:8080/api/tasks/search?tag_ids=1,2,3&page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. 组合搜索

```bash
curl -X GET "http://localhost:8080/api/tasks/search?reviewer_rtx=2669297147@qq.com&tag_ids=1,2&review_start_time=2025-10-01T00:00:00Z&page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 6. 使用JavaScript/Fetch

```javascript
const token = localStorage.getItem('token');

const searchParams = new URLSearchParams({
  reviewer_rtx: '2669297147@qq.com',
  page: 1,
  page_size: 20
});

fetch(`http://localhost:8080/api/tasks/search?${searchParams}`, {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
})
.then(response => response.json())
.then(data => {
  console.log('搜索结果:', data);
  console.log('总记录数:', data.total);
  console.log('总页数:', data.total_pages);
})
.catch(error => console.error('错误:', error));
```

## 注意事项

1. **认证要求**: 必须在请求头中包含有效的 JWT token
2. **权限**: 管理员和审核员都可以访问此接口
3. **分页限制**: 
   - 默认每页10条记录
   - 最大每页100条记录
   - 超过100会自动限制为100
4. **时间格式**: 使用 ISO 8601 格式，例如 `2025-10-24T23:59:59Z`
5. **只返回已完成的任务**: 搜索结果只包含 `status = 'completed'` 的记录
6. **标签搜索**: 使用逗号分隔的标签名称，任务只要包含其中任意一个标签即会被返回

## 错误响应

### 400 Bad Request

```json
{
  "error": "Invalid query parameters: ..."
}
```

### 401 Unauthorized

```json
{
  "error": "Authorization header required"
}
```

### 500 Internal Server Error

```json
{
  "error": "数据库查询错误信息"
}
```

## 性能建议

1. 使用具体的搜索条件可以提高查询速度
2. 建议设置合理的 `page_size`，避免一次返回过多数据
3. 如果经常按 `reviewer_rtx` 或 `comment_id` 查询，建议在数据库中添加相应索引
4. 对于大量数据的时间范围查询，建议缩小时间范围

## 数据库索引建议

为了提高搜索性能，建议创建以下索引：

```sql
-- 审核员ID索引（如果还没有）
CREATE INDEX IF NOT EXISTS idx_review_tasks_reviewer_id 
ON review_tasks(reviewer_id);

-- 评论ID索引（如果还没有）
CREATE INDEX IF NOT EXISTS idx_review_tasks_comment_id 
ON review_tasks(comment_id);

-- 完成时间索引（用于时间范围查询）
CREATE INDEX IF NOT EXISTS idx_review_tasks_completed_at 
ON review_tasks(completed_at) WHERE status = 'completed';

-- 状态索引
CREATE INDEX IF NOT EXISTS idx_review_tasks_status 
ON review_tasks(status);

-- 标签GIN索引（用于数组查询）
CREATE INDEX IF NOT EXISTS idx_review_results_tags 
ON review_results USING GIN(tags);
```

