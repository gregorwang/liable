# 视频审核流量池队列系统迁移指南

## 概述

本次重构将视频审核系统从"一审/二审"模式改为"流量池分级单阶段审核"模式，实现了三个流量池队列(100k → 1m → 10m)的自动流转系统。

## 已完成工作

### 1. 数据库迁移文件 ✅
- **文件**: `migrations/006_video_queue_pool_system.sql`
- **内容**:
  - 创建 `video_queue_tasks` 表（视频队列任务）
  - 创建 `video_queue_results` 表（简化的审核结果）
  - 更新 `tag_config` 表（添加 scope/queue_id 字段）
  - 更新 `video_quality_tags` 表（添加 scope/queue_id 字段）
  - 创建索引优化查询性能
  - 插入默认标签和权限

### 2. Go 代码实现 ✅
- **models.go**: 新增视频队列相关模型定义 (`internal/models/models.go:838-926`)
- **repository**: 新建 `video_queue_repo.go` - 数据访问层
- **service**: 新建 `video_queue_service.go` - 业务逻辑层（包含队列流转逻辑）
- **handler**: 新建 `video_queue_handler.go` - HTTP 处理器
- **routes**: 在 `main.go` 中添加新的 API 路由

### 3. 代码编译验证 ✅
- 所有代码编译成功，无语法错误

## 待执行步骤

### 步骤 1: 应用数据库迁移

等待 Supabase 维护完成后，执行以下步骤：

```bash
# 使用 Supabase MCP 工具应用 migration
# 或者手动在 Supabase Dashboard 中执行 SQL
```

**方式一：使用 MCP（推荐）**
```bash
# 在 Claude Code 中调用 Supabase MCP
mcp__supabase__apply_migration(
    project_id="bteujincywcdclrkosdc",
    name="006_video_queue_pool_system",
    query="<完整的 migration SQL>"
)
```

**方式二：手动执行**
1. 登录 Supabase Dashboard: https://supabase.com/dashboard/project/bteujincywcdclrkosdc
2. 进入 SQL Editor
3. 复制 `migrations/006_video_queue_pool_system.sql` 的内容
4. 执行 SQL

### 步骤 2: 验证数据库结构

执行以下查询验证表已创建：

```sql
-- 检查新表
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN ('video_queue_tasks', 'video_queue_results');

-- 检查权限是否插入
SELECT permission_key FROM permissions
WHERE permission_key LIKE 'queue.video.%';

-- 检查标签是否插入
SELECT name, scope, queue_id FROM video_quality_tags
WHERE scope = 'video';
```

### 步骤 3: 权限配置

为审核员和质检员分配权限：

**普通审核员**（授予 100k 和 1m 队列权限）：
```sql
-- 假设用户 ID 为 123
INSERT INTO user_permissions (user_id, permission_key, granted_at, granted_by) VALUES
    (123, 'queue.video.100k.claim', NOW(), <admin_id>),
    (123, 'queue.video.100k.submit', NOW(), <admin_id>),
    (123, 'queue.video.100k.return', NOW(), <admin_id>),
    (123, 'queue.video.100k.my', NOW(), <admin_id>),
    (123, 'queue.video.1m.claim', NOW(), <admin_id>),
    (123, 'queue.video.1m.submit', NOW(), <admin_id>),
    (123, 'queue.video.1m.return', NOW(), <admin_id>),
    (123, 'queue.video.1m.my', NOW(), <admin_id>);
```

**质检员**（额外授予 10m 队列权限）：
```sql
-- 假设用户 ID 为 456
INSERT INTO user_permissions (user_id, permission_key, granted_at, granted_by) VALUES
    (456, 'queue.video.10m.claim', NOW(), <admin_id>),
    (456, 'queue.video.10m.submit', NOW(), <admin_id>),
    (456, 'queue.video.10m.return', NOW(), <admin_id>),
    (456, 'queue.video.10m.my', NOW(), <admin_id>);
```

### 步骤 4: 创建入口队列任务

为现有视频创建 100k 队列任务：

```sql
-- 为所有 pending 状态的视频创建 100k 队列任务
INSERT INTO video_queue_tasks (video_id, pool, status)
SELECT id, '100k', 'pending'
FROM tiktok_videos
WHERE status = 'pending'
ON CONFLICT (video_id, pool) DO NOTHING;
```

### 步骤 5: 启动服务器

```bash
# 编译并运行
go build -o build/server.exe ./cmd/api
./build/server.exe
```

服务器启动后会自动：
- 初始化 Redis 连接
- 启动过期任务释放后台任务（每5分钟执行一次）
- 监听 HTTP 请求

### 步骤 6: 安全检查

运行 Supabase 安全顾问检查：

```bash
# 使用 Claude Code MCP 工具
mcp__supabase__get_advisors(
    project_id="bteujincywcdclrkosdc",
    type="security"
)

mcp__supabase__get_advisors(
    project_id="bteujincywcdclrkosdc",
    type="performance"
)
```

## API 接口说明

### 队列路由格式
所有队列操作都通过路径参数 `{pool}` 指定队列（`100k`、`1m`、`10m`）：

| 端点 | 方法 | 权限 | 说明 |
|------|------|------|------|
| `/api/video/{pool}/tasks/claim` | POST | `queue.video.{pool}.claim` | 领取任务 |
| `/api/video/{pool}/tasks/my` | GET | `queue.video.{pool}.my` | 查看我的任务 |
| `/api/video/{pool}/tasks/submit` | POST | `queue.video.{pool}.submit` | 提交审核 |
| `/api/video/{pool}/tasks/submit-batch` | POST | `queue.video.{pool}.submit` | 批量提交 |
| `/api/video/{pool}/tasks/return` | POST | `queue.video.{pool}.return` | 归还任务 |
| `/api/video/{pool}/tags` | GET | 认证即可 | 获取标签 |

### 审核决定枚举值

- `push_next_pool`: 推送到下一级流量池
  - 100k → 1m
  - 1m → 10m
  - 10m → 确认推送1000万流量池

- `natural_pool`: 分配到自然流量池（不再继续推广）

- `remove_violation`: 违规下架

### 请求示例

**领取任务**:
```json
POST /api/video/100k/tasks/claim
{
  "count": 10
}
```

**提交审核**:
```json
POST /api/video/100k/tasks/submit
{
  "task_id": 123,
  "review_decision": "push_next_pool",
  "reason": "内容优质，具有传播潜力",
  "tags": ["内容优质", "有传播潜力", "技术质量好"]
}
```

## Redis 键规范

- **队列（List）**: `video:queue:{pool}` - 存储待审核视频ID
- **领取跟踪（Set）**: `video:claimed:{user_id}:{pool}` - 跟踪用户领取的任务
- **任务锁（KV + TTL）**: `video:lock:{task_id}` - 任务锁定，值为用户ID
- **统计（Hash）**: `video:stats:queue:{pool}:{date}` - 队列统计数据

## 流转逻辑

```
视频上传
    ↓
100k 队列（入口）
    ↓
审核决定：
├─ push_next_pool → 1m 队列
├─ natural_pool → 自然流量池（结束）
└─ remove_violation → 违规下架（结束）

1m 队列
    ↓
审核决定：
├─ push_next_pool → 10m 队列（待质检确认）
├─ natural_pool → 自然流量池（结束）
└─ remove_violation → 违规下架（结束）

10m 队列（质检员）
    ↓
审核决定：
├─ push_next_pool → 确认推送1000万流量池（结束）
├─ natural_pool → 回退自然流量池（结束）
└─ remove_violation → 违规下架（结束）
```

## 兼容性说明

### 现有系统保留
- 旧的一审/二审路由保持不变（`/api/tasks/video-first-review/*` 和 `/api/tasks/video-second-review/*`）
- 旧的数据表保留（`video_first_review_tasks`, `video_second_review_tasks` 等）
- 逐步迁移，新视频使用新队列系统

### 标签系统升级
- `tag_config` 表新增 `scope`、`queue_id`、`is_simple` 字段
- `video_quality_tags` 表新增 `scope`、`queue_id` 字段
- 原有标签继续有效，新标签支持队列维度配置

## 监控与统计

### 队列统计视图
```sql
SELECT * FROM video_queue_pool_stats;
-- 返回每个队列的待审/进行中/完成统计

SELECT * FROM video_queue_decision_stats;
-- 返回每个队列的审核决定分布
```

### Redis 统计
```bash
# 查看队列长度
LLEN video:queue:100k
LLEN video:queue:1m
LLEN video:queue:10m

# 查看用户领取的任务
SMEMBERS video:claimed:{user_id}:100k
```

## 故障排查

### 常见问题

**1. 权限不足**
```
错误: "Permission denied"
解决: 检查用户是否被授予对应队列的权限
```

**2. 任务已过期**
```
错误: "Task not found or already completed"
解决: 任务可能已被后台任务释放，请重新领取
```

**3. 标签超过限制**
```
错误: "Maximum 3 tags allowed"
解决: 审核标签最多选择3个
```

**4. Redis 连接失败**
```
错误: Redis connection failed
解决: 检查 Redis 配置和连接状态
```

## 后续优化建议

1. **前端界面开发**:
   - 队列选择器
   - 审核表单（简化版，仅包含决定/理由/标签）
   - 统计仪表板

2. **性能优化**:
   - Redis 队列预热
   - 数据库连接池优化
   - 批量操作优化

3. **监控告警**:
   - 队列积压告警
   - 审核速率监控
   - 流转异常检测

4. **数据迁移**:
   - 制定历史数据迁移方案
   - 平滑切换策略

## 联系方式

如有问题，请联系开发团队或查看文档：
- 业务逻辑文档: `视频审核业务逻辑重构.md`
- Migration 文件: `migrations/006_video_queue_pool_system.sql`
- 代码参考:
  - `internal/services/video_queue_service.go:93-142` (队列流转逻辑)
  - `internal/repository/video_queue_repo.go` (数据访问)
  - `cmd/api/main.go:190-242` (路由配置)
