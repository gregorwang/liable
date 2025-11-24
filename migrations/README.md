# Database Migrations 数据库迁移文件

> 本目录包含所有数据库 schema 变更的 migration 文件

---

## 📁 文件列表

| 编号 | 文件名 | 创建日期 | 说明 | 状态 |
|------|-------|---------|------|------|
| 001 | `001_init_tables.sql` | 2025-11 | 初始化表结构 | ✅ 已应用 |
| 002 | `002_notifications.sql` | 2025-11 | 通知系统 | ✅ 已应用 |
| 003 | `003_video_review_system.sql` | 2025-11 | 视频审核系统（旧） | ✅ 已应用 |
| 004 | `004_add_email_verification.sql` | 2025-11 | 邮箱验证功能 | ✅ 已应用 |
| 005 | `005_unified_queue_stats.sql` | 2025-11 | 统一队列统计 | ✅ 已应用 |
| 006 | `006_video_queue_pool_system.sql` | 2025-11 | 视频流量池系统 | ✅ 已应用 |

---

## 📋 Migration 编写规范

### 1. 文件命名

```
{编号}_{功能描述}.sql

规则：
- 编号：3位数字，从001开始递增
- 描述：使用小写字母和下划线，简洁描述功能
- 扩展名：必须是 .sql

示例：
✅ 007_add_task_priority.sql
✅ 008_create_user_sessions.sql
❌ add_priority.sql          # 缺少编号
❌ 7_priority.sql             # 编号不是3位数
❌ 007_Add_Task_Priority.sql  # 使用了大写字母
```

### 2. 文件结构

```sql
-- ============================================================
-- Migration: {编号} - {功能名称}
-- Description: {详细描述}
-- Created: {日期}
-- Author: {作者}
-- ============================================================

-- 1. 创建表
CREATE TABLE IF NOT EXISTS table_name (
    -- 字段定义
);

-- 2. 添加索引
CREATE INDEX IF NOT EXISTS idx_table_field ON table_name(field);

-- 3. 添加外键约束
ALTER TABLE table_name
ADD CONSTRAINT fk_name FOREIGN KEY (field) REFERENCES other_table(id);

-- 4. 插入默认数据
INSERT INTO table_name (field1, field2) VALUES
    ('value1', 'value2')
ON CONFLICT (unique_field) DO NOTHING;

-- 5. 创建视图/函数（如果需要）
CREATE OR REPLACE VIEW view_name AS ...;

-- 6. 添加注释
COMMENT ON TABLE table_name IS '表说明';
COMMENT ON COLUMN table_name.field IS '字段说明';
```

### 3. 最佳实践

#### ✅ 推荐做法

```sql
-- 1. 使用 IF NOT EXISTS 避免重复执行错误
CREATE TABLE IF NOT EXISTS users (...);

-- 2. 使用 ON CONFLICT DO NOTHING 安全插入默认数据
INSERT INTO permissions (key, name) VALUES ('admin', '管理员')
ON CONFLICT (key) DO NOTHING;

-- 3. 为新字段提供默认值
ALTER TABLE users ADD COLUMN email VARCHAR DEFAULT '';

-- 4. 使用 DO $$ 块处理条件逻辑
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='users' AND column_name='email') THEN
        ALTER TABLE users ADD COLUMN email VARCHAR;
    END IF;
END $$;

-- 5. 添加必要的注释
COMMENT ON TABLE users IS '用户基础信息表';
```

#### ❌ 避免的做法

```sql
-- 1. 不要使用 DROP TABLE（除非你确定要删除）
DROP TABLE users;  # 危险！会丢失所有数据

-- 2. 不要硬编码 ID
INSERT INTO users (id, username) VALUES (1, 'admin');  # 可能冲突

-- 3. 不要在 migration 中修改现有数据
UPDATE users SET password = 'newpass';  # 数据变更应单独处理

-- 4. 不要在 migration 中引用其他数据库
SELECT * FROM other_db.table;  # 跨库查询

-- 5. 不要使用特定于时间的数据
INSERT INTO events (date) VALUES ('2025-11-24');  # 会过时
```

---

## 🚀 如何应用 Migration

### 使用 Supabase MCP (推荐)

当你配置了 Supabase MCP 后，AI 可以自动应用 migration：

```bash
# AI 会执行以下操作
1. 读取 migration 文件内容
   Read: migrations/007_add_task_priority.sql

2. 调用 MCP 工具应用 migration
   mcp__supabase__apply_migration(
       project_id: "bteujincywcdclrkosdc",
       name: "add_task_priority",
       query: <file_content>
   )

3. 验证应用结果
   mcp__supabase__list_migrations()

4. 检查表结构
   mcp__supabase__list_tables()
```

### 手动应用（备选方案）

```bash
# 1. 连接到 Supabase 数据库
psql "postgresql://postgres:[PASSWORD]@db.bteujincywcdclrkosdc.supabase.co:5432/postgres"

# 2. 执行 migration 文件
\i migrations/007_add_task_priority.sql

# 3. 验证结果
\dt  -- 查看所有表
\d table_name  -- 查看特定表结构
```

---

## 🔍 Migration 检查清单

### 应用前检查

- [ ] 编号是否正确递增？
- [ ] 文件命名是否符合规范？
- [ ] 是否使用了 `IF NOT EXISTS` / `IF EXISTS`？
- [ ] 是否有破坏性操作（DROP, TRUNCATE）？
- [ ] 外键约束是否正确？
- [ ] 是否添加了必要的索引？
- [ ] 默认数据是否使用了 `ON CONFLICT`？

### 应用后检查

- [ ] 所有表是否创建成功？
- [ ] 索引是否已添加？
- [ ] 外键约束是否生效？
- [ ] 默认数据是否插入？
- [ ] 运行 `get_advisors` 检查安全性和性能

```bash
# 使用 MCP 检查
mcp__supabase__get_advisors(project_id, type: "security")
mcp__supabase__get_advisors(project_id, type: "performance")
```

---

## 📝 Migration 模板

### 基础表创建

```sql
-- ============================================================
-- Migration: 007 - Add Task Priority
-- Description: Add priority field to tasks for better scheduling
-- Created: 2025-11-24
-- ============================================================

-- 1. Add priority column to review_tasks
ALTER TABLE review_tasks
ADD COLUMN IF NOT EXISTS priority INTEGER DEFAULT 50
CHECK (priority >= 1 AND priority <= 100);

-- 2. Add index for priority-based queries
CREATE INDEX IF NOT EXISTS idx_review_tasks_priority
ON review_tasks(priority DESC, created_at);

-- 3. Add comment
COMMENT ON COLUMN review_tasks.priority IS '任务优先级 (1-100，数字越大越优先)';

-- 4. Insert default priority for existing tasks
UPDATE review_tasks SET priority = 50 WHERE priority IS NULL;
```

### 添加新表

```sql
-- ============================================================
-- Migration: 008 - Create User Sessions Table
-- Description: Track user login sessions for security audit
-- Created: 2025-11-24
-- ============================================================

-- 1. Create user_sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_activity TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 2. Add indexes
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id
ON user_sessions(user_id);

CREATE INDEX IF NOT EXISTS idx_user_sessions_token
ON user_sessions(session_token);

CREATE INDEX IF NOT EXISTS idx_user_sessions_expires
ON user_sessions(expires_at);

-- 3. Add comments
COMMENT ON TABLE user_sessions IS '用户会话表，用于安全审计';
COMMENT ON COLUMN user_sessions.session_token IS '会话令牌 (UUID)';
COMMENT ON COLUMN user_sessions.expires_at IS '过期时间，默认7天';

-- 4. Create cleanup function
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM user_sessions WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- 5. Optional: Create scheduled job (Supabase specific)
-- SELECT cron.schedule(
--     'cleanup-sessions',
--     '0 * * * *',  -- 每小时执行一次
--     'SELECT cleanup_expired_sessions();'
-- );
```

### 修改现有表

```sql
-- ============================================================
-- Migration: 009 - Add Soft Delete to Users
-- Description: Implement soft delete for users instead of hard delete
-- Created: 2025-11-24
-- ============================================================

-- 1. Add deleted_at column
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='users' AND column_name='deleted_at'
    ) THEN
        ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;
    END IF;
END $$;

-- 2. Add index for non-deleted users
CREATE INDEX IF NOT EXISTS idx_users_not_deleted
ON users(id) WHERE deleted_at IS NULL;

-- 3. Add comment
COMMENT ON COLUMN users.deleted_at IS '删除时间，NULL 表示未删除（软删除）';

-- 4. Update queries to filter deleted users (示例)
-- Future queries should include: WHERE deleted_at IS NULL
```

---

## 🔄 回滚策略

### 重要提示

**Supabase 的 migrations 默认不支持自动回滚！**

### 手动回滚方法

#### 方法 1: 创建回滚 migration

```sql
-- migrations/007_add_task_priority.sql (原始)
ALTER TABLE review_tasks ADD COLUMN priority INTEGER;

-- migrations/007_rollback_task_priority.sql (回滚)
ALTER TABLE review_tasks DROP COLUMN IF EXISTS priority;
```

#### 方法 2: 使用 Supabase Branch

```bash
# 1. 创建测试分支
mcp__supabase__create_branch(
    project_id: "xxx",
    name: "test-task-priority"
)

# 2. 在分支上测试 migration
mcp__supabase__apply_migration(branch_id, ...)

# 3. 测试通过后合并到主库
mcp__supabase__merge_branch(branch_id)

# 4. 如果测试失败，直接删除分支
mcp__supabase__delete_branch(branch_id)
```

### 不可回滚的操作

⚠️ 以下操作无法安全回滚，请谨慎：

- `DROP TABLE` - 删除表会丢失所有数据
- `DROP COLUMN` - 删除列会丢失该列的数据
- `ALTER COLUMN TYPE` - 类型转换可能丢失数据
- `TRUNCATE` - 清空表数据
- 数据迁移脚本 (UPDATE, DELETE)

**建议**：对于危险操作，先在 Branch 中测试！

---

## 🎯 常见场景

### 场景 1: 新增功能需要新表

```sql
-- 示例：添加任务评论功能

-- migrations/010_add_task_comments.sql

CREATE TABLE IF NOT EXISTS task_comments (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL,
    task_type VARCHAR(50) NOT NULL,  -- 'review_task', 'video_queue_task'
    user_id INTEGER NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_comments_task
ON task_comments(task_type, task_id);

COMMENT ON TABLE task_comments IS '任务评论表，支持多种任务类型';
```

### 场景 2: 优化查询性能

```sql
-- 示例：为常用查询添加复合索引

-- migrations/011_optimize_task_queries.sql

-- 1. 为"按流量池和状态查询任务"添加复合索引
CREATE INDEX IF NOT EXISTS idx_video_queue_tasks_pool_status
ON video_queue_tasks(pool, status)
WHERE status = 'pending';

-- 2. 为"审核员的进行中任务"添加复合索引
CREATE INDEX IF NOT EXISTS idx_review_tasks_reviewer_status
ON review_tasks(reviewer_id, status)
WHERE status = 'in_progress';

-- 3. 添加注释说明优化目的
COMMENT ON INDEX idx_video_queue_tasks_pool_status IS
'优化流量池任务领取查询，使用部分索引减少索引大小';
```

### 场景 3: 重构表结构

```sql
-- 示例：合并 task_queue 和 task_queues

-- migrations/012_merge_task_queue_tables.sql

-- 1. 迁移数据从旧表到新表
INSERT INTO task_queue (queue_name, description, priority, total_tasks, completed_tasks, is_active)
SELECT queue_name, description, priority, total_tasks, completed_tasks, is_active
FROM task_queues
ON CONFLICT (queue_name) DO NOTHING;

-- 2. 验证数据迁移
DO $$
DECLARE
    old_count INTEGER;
    new_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO old_count FROM task_queues;
    SELECT COUNT(*) INTO new_count FROM task_queue;

    IF old_count <> new_count THEN
        RAISE EXCEPTION 'Data migration failed: old_count=%, new_count=%', old_count, new_count;
    END IF;
END $$;

-- 3. 重命名旧表（不删除，保留备份）
ALTER TABLE task_queues RENAME TO task_queues_deprecated;

-- 4. 添加注释
COMMENT ON TABLE task_queues_deprecated IS
'已废弃：数据已迁移到 task_queue，保留用于备份，可在确认无问题后删除';
```

### 场景 4: 添加权限

```sql
-- 示例：为新功能添加权限

-- migrations/013_add_task_comment_permissions.sql

-- 1. 插入新权限
INSERT INTO permissions (permission_key, name, description, resource, action, category, is_active)
VALUES
    ('task_comments.create', '创建任务评论', '允许在任务下创建评论', 'task_comments', 'create', 'task_management', true),
    ('task_comments.view', '查看任务评论', '允许查看任务的评论列表', 'task_comments', 'view', 'task_management', true),
    ('task_comments.delete', '删除任务评论', '允许删除自己的评论', 'task_comments', 'delete', 'task_management', true),
    ('task_comments.manage', '管理所有评论', '允许管理员删除任何评论', 'task_comments', 'manage', 'admin', true)
ON CONFLICT (permission_key) DO NOTHING;

-- 2. 为现有审核员分配基础权限
INSERT INTO user_permissions (user_id, permission_key, granted_by)
SELECT u.id, 'task_comments.create', 1  -- granted_by 1 = admin
FROM users u
WHERE u.role = 'reviewer' AND u.status = 'approved'
ON CONFLICT DO NOTHING;

-- 3. 为管理员分配管理权限
INSERT INTO user_permissions (user_id, permission_key, granted_by)
SELECT u.id, 'task_comments.manage', 1
FROM users u
WHERE u.role = 'admin'
ON CONFLICT DO NOTHING;
```

---

## 🛠️ 调试 Migration

### 常见错误

#### 1. 外键约束违反

```sql
-- 错误示例
ALTER TABLE review_results
ADD CONSTRAINT fk_task FOREIGN KEY (task_id) REFERENCES review_tasks(id);

-- 错误信息
ERROR:  insert or update on table "review_results" violates foreign key constraint "fk_task"
DETAIL:  Key (task_id)=(123) is not present in table "review_tasks".

-- 解决方案：先清理孤立数据
DELETE FROM review_results
WHERE task_id NOT IN (SELECT id FROM review_tasks);

-- 然后再添加约束
ALTER TABLE review_results
ADD CONSTRAINT fk_task FOREIGN KEY (task_id) REFERENCES review_tasks(id);
```

#### 2. 唯一约束冲突

```sql
-- 错误示例
INSERT INTO users (username, password, role) VALUES ('admin', 'xxx', 'admin');

-- 错误信息
ERROR:  duplicate key value violates unique constraint "users_username_key"

-- 解决方案：使用 ON CONFLICT
INSERT INTO users (username, password, role)
VALUES ('admin', 'xxx', 'admin')
ON CONFLICT (username) DO NOTHING;
```

#### 3. 列已存在

```sql
-- 错误示例
ALTER TABLE users ADD COLUMN email VARCHAR;

-- 错误信息
ERROR:  column "email" of relation "users" already exists

-- 解决方案：使用条件检查
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='users' AND column_name='email'
    ) THEN
        ALTER TABLE users ADD COLUMN email VARCHAR;
    END IF;
END $$;
```

### 测试 Migration

```sql
-- 在应用前，可以在 Supabase SQL Editor 中测试

BEGIN;  -- 开启事务

-- 执行你的 migration SQL
CREATE TABLE test_table (...);

-- 验证结果
SELECT * FROM test_table;

ROLLBACK;  -- 回滚，不实际提交
-- 或
COMMIT;  -- 确认无误后提交
```

---

## 📊 Migration 历史跟踪

### 查看已应用的 migrations

```bash
# 使用 MCP 工具
mcp__supabase__list_migrations(project_id: "bteujincywcdclrkosdc")
```

### 手动记录

在项目根目录创建 `MIGRATIONS_LOG.md`:

```markdown
# Migration 应用日志

| 日期 | Migration | 应用人 | 状态 | 备注 |
|------|-----------|--------|------|------|
| 2025-11-24 | 001_init_tables | AI | ✅ 成功 | 初始化 |
| 2025-11-24 | 002_notifications | AI | ✅ 成功 | 通知系统 |
| 2025-11-24 | 007_task_priority | 用户 | ✅ 成功 | 添加任务优先级 |
| 2025-11-25 | 008_user_sessions | AI | ⚠️ 回滚 | 发现性能问题 |
| 2025-11-25 | 008_user_sessions_v2 | AI | ✅ 成功 | 优化后重新应用 |
```

---

## 🤝 与 AI 协作时的提示

### 告诉 AI 创建 Migration

```
我需要为 review_tasks 表添加优先级功能。

要求：
1. 添加 priority 字段 (INTEGER, 1-100, 默认 50)
2. 添加索引支持按优先级排序
3. 为现有任务设置默认优先级
4. 遵循项目的 migration 命名规范

请生成 migration 文件，并使用 Supabase MCP 应用。
```

### AI 的执行流程

```
1. [Read] 读取最新的 migration 编号
   Read: migrations/006_video_queue_pool_system.sql

2. [Write] 创建新的 migration 文件
   Write: migrations/007_add_task_priority.sql

3. [MCP] 应用 migration
   mcp__supabase__apply_migration(...)

4. [MCP] 验证应用结果
   mcp__supabase__list_tables()
   mcp__supabase__get_advisors(type: "performance")

5. [Update] 更新文档
   Edit: migrations/README.md (添加到文件列表)
   Edit: DATABASE_SCHEMA.md (更新表结构)
```

---

## 📚 相关文档

- **[DATABASE_SCHEMA.md](../DATABASE_SCHEMA.md)** - 数据库完整 schema 文档
- **[AI_COLLABORATION_DATABASE_GUIDE.md](../AI_COLLABORATION_DATABASE_GUIDE.md)** - AI 协作数据库管理指南
- **[Supabase Migrations 官方文档](https://supabase.com/docs/guides/cli/local-development#database-migrations)**

---

*本文档会随着新 migration 的添加而更新。*
