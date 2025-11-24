# 数据库优化建议文档

> 本文档由 AI 深度分析生成，针对 comment-review-platform 项目的 Supabase 数据库进行全面优化建议

## 📋 目录

1. [数据库现状概览](#数据库现状概览)
2. [严重问题（必须修复）](#严重问题必须修复)
3. [性能优化建议](#性能优化建议)
4. [架构设计优化](#架构设计优化)
5. [数据一致性改进](#数据一致性改进)
6. [优先级总结](#优先级总结)
7. [实施步骤](#实施步骤)

---

## 数据库现状概览

### 业务模块分布
```
评论审核系统（6表）
├── comment - 评论主表 (5,323 条)
├── review_tasks - 一审任务 (5,323 条)
├── review_results - 一审结果 (36 条)
├── second_review_tasks - 二审任务 (11 条)
├── second_review_results - 二审结果 (9 条)
└── quality_check_tasks/results - 质检任务/结果 (0 条)

视频审核系统（10表）
├── tiktok_videos - 视频主表 (88 条)
├── video_first_review_tasks/results - 一审 (88/37 条)
├── video_second_review_tasks/results - 二审 (0/0 条)
├── video_queue_tasks/results - 队列审核 (58/12 条)
└── video_quality_tags - 视频质量标签 (39 条)

用户权限系统（3表）
├── users - 用户表 (4 条)
├── permissions - 权限定义 (54 条)
└── user_permissions - 用户权限关系 (117 条)

其他系统（7表）
├── task_queue/task_queues - 任务队列（重复？）
├── notifications/user_notifications - 通知系统
├── tag_config - 评论标签配置
├── moderation_rules - 审核规则库
└── messages/email_verification_logs - 消息与验证
```

### 数据库健康评分
- **安全性**: ⚠️ 30/100（严重不足）
- **性能**: ⚠️ 55/100（需要优化）
- **架构设计**: ⚡ 75/100（良好但有改进空间）
- **数据一致性**: ⚡ 80/100（较好）

---

## 严重问题（必须修复）

### 🔴 P0：安全漏洞（立即修复）

#### 1. RLS（行级安全）未启用

**问题严重性**: ⛔ 致命

**影响范围**: 25个表完全暴露，任何人都可以通过 PostgREST API 访问

**受影响的表**:
```sql
-- 用户相关
users, user_permissions, permissions

-- 审核相关
review_tasks, review_results
second_review_tasks, second_review_results
quality_check_tasks, quality_check_results

-- 视频相关
tiktok_videos, video_first_review_tasks, video_first_review_results
video_second_review_tasks, video_second_review_results
video_queue_tasks, video_queue_results, video_quality_tags

-- 系统配置
tag_config, task_queue, task_queues, moderation_rules
notifications, user_notifications
messages, email_verification_logs
```

**修复方案**:

```sql
-- 示例：为 users 表启用 RLS
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

-- 创建策略：用户只能看到自己的信息
CREATE POLICY "Users can view their own data"
ON public.users FOR SELECT
USING (auth.uid()::text = id::text);

-- 管理员可以查看所有用户
CREATE POLICY "Admins can view all users"
ON public.users FOR SELECT
USING (
  EXISTS (
    SELECT 1 FROM public.users
    WHERE id = auth.uid()::int
    AND role = 'admin'
  )
);

-- 为其他24个表创建类似的策略...
```

**业务影响**:
- 未修复前：数据库数据完全暴露给公网
- 修复后：根据业务规则控制数据访问

---

#### 2. comment 表启用了 RLS 但没有策略

**问题**: 表启用了 RLS，但没有定义任何访问策略，导致**所有人都无法访问**

**修复方案**:

```sql
-- 审核员可以通过任务查看评论
CREATE POLICY "Reviewers can view comments through tasks"
ON public.comment FOR SELECT
USING (
  EXISTS (
    SELECT 1 FROM public.review_tasks rt
    WHERE rt.comment_id = comment.id
    AND rt.reviewer_id = auth.uid()::int
  )
  OR
  EXISTS (
    SELECT 1 FROM public.second_review_tasks srt
    WHERE srt.comment_id = comment.id
    AND srt.reviewer_id = auth.uid()::int
  )
);

-- 管理员可以查看所有评论
CREATE POLICY "Admins can view all comments"
ON public.comment FOR SELECT
USING (
  EXISTS (
    SELECT 1 FROM public.users
    WHERE id = auth.uid()::int
    AND role = 'admin'
  )
);
```

---

#### 3. 视图使用 SECURITY DEFINER（安全隐患）

**问题**: 4个统计视图使用了 `SECURITY DEFINER`，绕过了 RLS 检查

**受影响的视图**:
```
- unified_queue_stats
- queue_stats
- video_queue_pool_stats
- video_queue_decision_stats
```

**风险**: 攻击者可以通过这些视图访问本应受保护的数据

**修复方案**:

```sql
-- 方案1：删除 SECURITY DEFINER（推荐）
DROP VIEW IF EXISTS public.unified_queue_stats;
CREATE VIEW public.unified_queue_stats AS
SELECT ... -- 原查询
-- 不要添加 SECURITY DEFINER

-- 方案2：保留 SECURITY DEFINER，但添加严格的 RLS
CREATE POLICY "Only admins can view queue stats"
ON public.unified_queue_stats FOR SELECT
USING (
  EXISTS (
    SELECT 1 FROM public.users
    WHERE id = auth.uid()::int
    AND role = 'admin'
  )
);
```

---

#### 4. 函数 search_path 可变（SQL注入风险）

**问题**: 4个函数没有设置固定的 search_path，可能导致 SQL 注入

**受影响的函数**:
```
- enforce_message_cooldown
- get_video_queue_tags
- can_post_once_every_48h
- update_updated_at_column
```

**修复方案**:

```sql
-- 为每个函数添加 SET search_path
ALTER FUNCTION public.enforce_message_cooldown()
SET search_path = public, pg_temp;

ALTER FUNCTION public.get_video_queue_tags(text)
SET search_path = public, pg_temp;

ALTER FUNCTION public.can_post_once_every_48h()
SET search_path = public, pg_temp;

ALTER FUNCTION public.update_updated_at_column()
SET search_path = public, pg_temp;
```

---

#### 5. Auth 配置安全问题

**问题1**: OTP 过期时间超过1小时（当前可能更长）
```bash
# 在 Supabase Dashboard 中修改
Authentication → Email Auth → OTP Expiry
建议值：3600 秒（1小时）或更短
```

**问题2**: 密码泄露检测未启用
```bash
# 在 Supabase Dashboard 中启用
Authentication → Password → Enable Leaked Password Protection
启用后会检查 HaveIBeenPwned.org 数据库
```

---

### 🟡 P1：性能问题（尽快修复）

#### 6. 外键缺失索引（查询性能差）

**问题**: 14个外键约束没有对应的索引，导致 JOIN 查询非常慢

**受影响的表和字段**:

```sql
-- 需要添加索引的外键
notifications.created_by
quality_check_results.qc_task_id
quality_check_tasks.comment_id
quality_check_tasks.first_review_result_id
second_review_tasks.comment_id
second_review_tasks.first_review_result_id
task_queue.created_by
task_queue.updated_by
user_permissions.granted_by
video_first_review_results.task_id
video_first_review_tasks.video_id
video_second_review_results.second_task_id
video_second_review_tasks.first_review_result_id
video_second_review_tasks.video_id
```

**修复SQL**:

```sql
-- 创建缺失的索引
CREATE INDEX idx_notifications_created_by ON public.notifications(created_by);
CREATE INDEX idx_quality_check_results_qc_task_id ON public.quality_check_results(qc_task_id);
CREATE INDEX idx_quality_check_tasks_comment_id ON public.quality_check_tasks(comment_id);
CREATE INDEX idx_quality_check_tasks_first_review_result_id ON public.quality_check_tasks(first_review_result_id);
CREATE INDEX idx_second_review_tasks_comment_id ON public.second_review_tasks(comment_id);
CREATE INDEX idx_second_review_tasks_first_review_result_id ON public.second_review_tasks(first_review_result_id);
CREATE INDEX idx_task_queue_created_by ON public.task_queue(created_by);
CREATE INDEX idx_task_queue_updated_by ON public.task_queue(updated_by);
CREATE INDEX idx_user_permissions_granted_by ON public.user_permissions(granted_by);
CREATE INDEX idx_video_first_review_results_task_id ON public.video_first_review_results(task_id);
CREATE INDEX idx_video_first_review_tasks_video_id ON public.video_first_review_tasks(video_id);
CREATE INDEX idx_video_second_review_results_second_task_id ON public.video_second_review_results(second_task_id);
CREATE INDEX idx_video_second_review_tasks_first_review_result_id ON public.video_second_review_tasks(first_review_result_id);
CREATE INDEX idx_video_second_review_tasks_video_id ON public.video_second_review_tasks(video_id);
```

**性能影响**:
- 修复前：JOIN 查询可能需要全表扫描
- 修复后：查询速度提升 10-100 倍

---

#### 7. 重复索引（浪费空间）

**问题**: 4组索引完全重复，浪费存储空间和写入性能

**重复索引对**:
```sql
-- review_results 表
idx_review_results_reviewer ≈ idx_review_results_reviewer_id

-- second_review_results 表
idx_second_review_results_reviewer ≈ idx_second_review_results_reviewer_id

-- video_first_review_results 表
idx_video_first_review_results_reviewer ≈ idx_video_first_review_results_reviewer_id

-- video_second_review_results 表
idx_video_second_review_results_reviewer ≈ idx_video_second_review_results_reviewer_id
```

**修复SQL**:

```sql
-- 删除重复的索引（保留名称更清晰的那个）
DROP INDEX IF EXISTS public.idx_review_results_reviewer;
DROP INDEX IF EXISTS public.idx_second_review_results_reviewer;
DROP INDEX IF EXISTS public.idx_video_first_review_results_reviewer;
DROP INDEX IF EXISTS public.idx_video_second_review_results_reviewer;

-- 保留这些索引
-- idx_review_results_reviewer_id
-- idx_second_review_results_reviewer_id
-- idx_video_first_review_results_reviewer_id
-- idx_video_second_review_results_reviewer_id
```

**收益**:
- 减少索引维护开销
- 节省约 4-8 MB 存储空间（取决于数据量）

---

#### 8. 未使用的索引（共62个）

**问题**: 62个索引从未被查询使用，纯粹浪费资源

**建议**: 保留核心索引，删除从未使用的索引

**需要审查的索引** (分优先级):

**可以安全删除的索引** (明显未使用):
```sql
-- 统计相关（如果不做复杂查询）
DROP INDEX IF EXISTS idx_review_results_created_at;
DROP INDEX IF EXISTS idx_review_tasks_claimed_at;
DROP INDEX IF EXISTS idx_review_tasks_completed_at;
DROP INDEX IF EXISTS idx_second_review_tasks_claimed_at;
DROP INDEX IF EXISTS idx_second_review_results_created_at;

-- 权限相关（后端控制权限）
DROP INDEX IF EXISTS idx_permissions_resource;
DROP INDEX IF EXISTS idx_permissions_category;
DROP INDEX IF EXISTS idx_permissions_active;

-- 视频审核（数据量小）
DROP INDEX IF EXISTS idx_video_first_review_results_created_at;
DROP INDEX IF EXISTS idx_video_quality_tags_active;
DROP INDEX IF EXISTS idx_video_quality_tags_scope;
```

**观察一段时间再决定的索引**:
```sql
-- 用户查询相关
idx_users_username
idx_users_email
idx_users_role_status

-- 标签查询
idx_tag_config_scope
idx_tag_config_queue_id

-- 队列管理
idx_task_queues_is_active
idx_task_queue_active
```

**⚠️ 警告**: 删除索引前请先观察查询日志，确认未被使用

---

## 架构设计优化

### 🔵 P2：架构改进（中期优化）

#### 9. 重复的队列表设计

**问题**: 同时存在 `task_queue` 和 `task_queues` 两个表

**现状**:
```
task_queue: 6 条记录（有 created_by, updated_by）
task_queues: 5 条记录（无 created_by, updated_by）
```

**建议**: 统一为一个表

**迁移方案**:

```sql
-- 步骤1：数据迁移
INSERT INTO task_queue (queue_name, description, priority, total_tasks, completed_tasks, is_active, created_at, updated_at)
SELECT queue_name, description, priority, total_tasks, completed_tasks, is_active, created_at, updated_at
FROM task_queues
WHERE queue_name NOT IN (SELECT queue_name FROM task_queue);

-- 步骤2：更新代码引用

-- 步骤3：删除旧表
DROP TABLE IF EXISTS task_queues;
```

**收益**:
- 简化数据模型
- 避免数据不一致

---

#### 10. 评论审核与视频审核架构不一致

**问题**: 两个系统使用了不同的设计模式

**评论审核**:
```
一审: review_tasks → review_results
二审: second_review_tasks → second_review_results (引用 first_review_result_id)
质检: quality_check_tasks → quality_check_results
```

**视频审核**:
```
一审: video_first_review_tasks → video_first_review_results
二审: video_second_review_tasks → video_second_review_results
队列: video_queue_tasks → video_queue_results （简化版）
```

**建议**: 统一架构模式

**方案1: 通用审核框架**（推荐长期方案）

```sql
-- 创建通用审核表
CREATE TABLE review_items (
  id BIGSERIAL PRIMARY KEY,
  item_type VARCHAR(50) NOT NULL, -- 'comment' or 'video'
  item_id BIGINT NOT NULL,
  item_data JSONB, -- 灵活存储不同类型的数据
  status VARCHAR(50) DEFAULT 'pending',
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE review_workflows (
  id SERIAL PRIMARY KEY,
  item_id BIGINT REFERENCES review_items(id),
  workflow_stage VARCHAR(50), -- 'first_review', 'second_review', 'quality_check'
  reviewer_id INT REFERENCES users(id),
  result JSONB, -- 灵活存储不同的审核结果
  completed_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW()
);
```

**方案2: 保持分离但统一命名**（推荐短期方案）

```sql
-- 重命名以保持一致性
-- 评论系统
comment_first_review_tasks (从 review_tasks 重命名)
comment_first_review_results (从 review_results 重命名)
comment_second_review_tasks (保持)
comment_second_review_results (保持)

-- 视频系统保持现有命名
video_first_review_tasks
video_first_review_results
video_second_review_tasks
video_second_review_results
```

---

#### 11. 用户 ID 类型不一致

**问题**:
- `users.id` 是 `INTEGER`
- `messages.user_id` 是 `TEXT`

**风险**:
- JOIN 性能差
- 容易出错

**修复方案**:

```sql
-- 检查 messages 表的数据
SELECT user_id, COUNT(*) FROM messages GROUP BY user_id;

-- 如果数据都是数字字符串，可以转换
ALTER TABLE messages
ALTER COLUMN user_id TYPE INTEGER
USING user_id::integer;

-- 添加外键约束
ALTER TABLE messages
ADD CONSTRAINT fk_messages_user
FOREIGN KEY (user_id) REFERENCES users(id);
```

---

## 数据一致性改进

### 🟢 P3：数据质量（持续改进）

#### 12. 缺少必要的约束

**建议添加的约束**:

```sql
-- 1. 用户表：邮箱格式验证
ALTER TABLE users
ADD CONSTRAINT chk_users_email_format
CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$');

-- 2. 审核结果：标签数组不能为空（如果未通过）
ALTER TABLE review_results
ADD CONSTRAINT chk_review_results_tags
CHECK (is_approved = true OR (tags IS NOT NULL AND array_length(tags, 1) > 0));

-- 3. 视频：文件大小必须为正数
ALTER TABLE tiktok_videos
ADD CONSTRAINT chk_tiktok_videos_file_size
CHECK (file_size > 0);

-- 4. 视频质量分数：范围验证
ALTER TABLE video_first_review_results
ADD CONSTRAINT chk_video_first_review_overall_score
CHECK (overall_score >= 4 AND overall_score <= 40);

-- 5. 任务状态转换：completed_at 必须在 claimed_at 之后
ALTER TABLE review_tasks
ADD CONSTRAINT chk_review_tasks_time_sequence
CHECK (completed_at IS NULL OR completed_at >= claimed_at);
```

---

#### 13. 缺少软删除机制

**问题**: 直接删除数据，无法追踪历史

**建议**: 添加软删除字段

```sql
-- 为关键表添加软删除
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP;
ALTER TABLE review_tasks ADD COLUMN deleted_at TIMESTAMP;
ALTER TABLE tiktok_videos ADD COLUMN deleted_at TIMESTAMP;

-- 创建视图自动过滤已删除数据
CREATE VIEW users_active AS
SELECT * FROM users WHERE deleted_at IS NULL;

-- 更新查询改用视图
-- SELECT * FROM users → SELECT * FROM users_active
```

---

#### 14. 缺少审计日志

**建议**: 为关键操作添加审计表

```sql
CREATE TABLE audit_logs (
  id BIGSERIAL PRIMARY KEY,
  table_name VARCHAR(100) NOT NULL,
  record_id BIGINT NOT NULL,
  operation VARCHAR(20) NOT NULL, -- 'INSERT', 'UPDATE', 'DELETE'
  old_data JSONB,
  new_data JSONB,
  user_id INTEGER REFERENCES users(id),
  ip_address INET,
  created_at TIMESTAMP DEFAULT NOW()
);

-- 创建触发器自动记录变更
CREATE OR REPLACE FUNCTION audit_trigger_func()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    INSERT INTO audit_logs (table_name, record_id, operation, old_data, user_id)
    VALUES (TG_TABLE_NAME, OLD.id, 'DELETE', row_to_json(OLD), current_setting('app.user_id', true)::int);
    RETURN OLD;
  ELSIF TG_OP = 'UPDATE' THEN
    INSERT INTO audit_logs (table_name, record_id, operation, old_data, new_data, user_id)
    VALUES (TG_TABLE_NAME, NEW.id, 'UPDATE', row_to_json(OLD), row_to_json(NEW), current_setting('app.user_id', true)::int);
    RETURN NEW;
  ELSIF TG_OP = 'INSERT' THEN
    INSERT INTO audit_logs (table_name, record_id, operation, new_data, user_id)
    VALUES (TG_TABLE_NAME, NEW.id, 'INSERT', row_to_json(NEW), current_setting('app.user_id', true)::int);
    RETURN NEW;
  END IF;
END;
$$ LANGUAGE plpgsql;

-- 为 users 表启用审计
CREATE TRIGGER users_audit_trigger
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();
```

---

## 优先级总结

### 立即执行（本周内）- P0
```
✅ 1. 为所有25个表启用 RLS 并配置策略
✅ 2. 为 comment 表添加 RLS 策略
✅ 3. 修复 4 个视图的 SECURITY DEFINER 问题
✅ 4. 修复 4 个函数的 search_path 问题
✅ 5. 配置 Auth 安全设置
```

### 尽快执行（本月内）- P1
```
⚡ 6. 为 14 个外键添加索引
⚡ 7. 删除 4 组重复索引
⚡ 8. 审查并删除未使用的索引
```

### 计划执行（季度内）- P2
```
🔵 9. 统一 task_queue 表
🔵 10. 统一审核架构
🔵 11. 修复用户 ID 类型不一致
```

### 持续改进 - P3
```
🟢 12. 添加数据约束
🟢 13. 实施软删除
🟢 14. 添加审计日志
```

---

## 实施步骤

### 阶段 1: 安全加固（第1周）

**Day 1-2: RLS 策略设计**
```bash
1. 梳理业务权限需求
   - 审核员：只能看自己领取的任务
   - 管理员：可以看所有数据
   - 匿名用户：无权限

2. 编写 RLS 策略 SQL
   - 为每个表创建独立的策略文件
   - 使用 migrations/ 目录管理

3. 在测试环境验证
   - 创建测试用户
   - 模拟各种角色的访问场景
```

**Day 3-4: RLS 部署**
```bash
1. 备份数据库
   pg_dump -h [host] -U postgres -d postgres > backup_$(date +%Y%m%d).sql

2. 在生产环境分批启用 RLS
   - 先启用非关键表（如 tag_config）
   - 观察应用是否正常
   - 逐步启用所有表

3. 验证应用功能
   - 登录测试
   - 领取任务测试
   - 提交审核测试
```

**Day 5: 修复视图和函数**
```sql
-- 执行之前准备的修复 SQL
-- 验证统计功能是否正常
```

---

### 阶段 2: 性能优化（第2-3周）

**Week 2: 添加索引**
```sql
-- 每天添加 5-7 个索引
-- 在低峰期执行（使用 CONCURRENTLY）
CREATE INDEX CONCURRENTLY idx_xxx ON table(column);

-- 监控索引创建进度
SELECT * FROM pg_stat_progress_create_index;
```

**Week 3: 清理冗余索引**
```sql
-- 删除重复索引
-- 观察一周后再删除未使用的索引
```

---

### 阶段 3: 架构重构（第4-8周）

**Week 4-5: 队列表统一**
```bash
1. 数据迁移脚本
2. 代码更新
3. 灰度发布
4. 删除旧表
```

**Week 6-8: 审核架构统一（可选）**
```bash
1. 设计新架构
2. 编写迁移工具
3. 并行运行新旧系统
4. 逐步切换流量
```

---

### 阶段 4: 质量提升（持续）

**每月执行**:
```sql
-- 1. 检查数据质量
SELECT COUNT(*) FROM users WHERE email !~* '^[A-Za-z0-9._%+-]+@';

-- 2. 检查索引使用情况
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0
ORDER BY schemaname, tablename;

-- 3. 检查表膨胀
SELECT schemaname, tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

---

## 监控与维护

### 性能监控查询

```sql
-- 1. 慢查询分析
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 20;

-- 2. 表大小监控
SELECT
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
  pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS index_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- 3. 索引健康度
SELECT
  schemaname,
  tablename,
  indexname,
  idx_scan as index_scans,
  idx_tup_read as tuples_read,
  idx_tup_fetch as tuples_fetched,
  pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;
```

---

## 备份与回滚方案

### 在每次重大修改前执行

```bash
# 1. 完整备份
pg_dump -h db.xxx.supabase.co \
  -U postgres \
  -d postgres \
  -F c \
  -f backup_before_optimization_$(date +%Y%m%d_%H%M%S).dump

# 2. 仅备份 schema
pg_dump -h db.xxx.supabase.co \
  -U postgres \
  -d postgres \
  --schema-only \
  -f schema_backup_$(date +%Y%m%d_%H%M%S).sql

# 3. 测试恢复（在测试环境）
pg_restore -h test-db.supabase.co \
  -U postgres \
  -d test_db \
  -v backup_before_optimization.dump
```

---

## 成功指标

### 安全性指标
- ✅ RLS 启用率：100%（当前 4%）
- ✅ 安全警告数：0（当前 27 个）

### 性能指标
- ✅ 外键索引覆盖率：100%（当前 60%）
- ✅ 平均查询响应时间：< 100ms
- ✅ 慢查询数量：< 5 个/天

### 架构指标
- ✅ 表设计一致性：高
- ✅ 数据冗余度：低
- ✅ 代码维护性：高

---

## 注意事项

### ⚠️ 高风险操作

1. **启用 RLS**: 可能导致应用无法访问数据
   - 解决方案：先在测试环境验证，生产环境分批启用

2. **删除索引**: 可能导致查询变慢
   - 解决方案：先标记为失效，观察一周，确认无影响后再删除

3. **修改数据类型**: 可能导致数据丢失
   - 解决方案：先备份，使用 `USING` 子句转换，验证数据完整性

### 💡 最佳实践

1. **所有 DDL 操作都要有迁移文件**
2. **生产环境操作前先在测试环境验证**
3. **重大修改要有回滚方案**
4. **监控修改后的性能指标**
5. **保持数据库文档更新**

---

## 后续优化方向

1. **分区表**: 当 comment/tiktok_videos 表超过 100 万行时考虑
2. **读写分离**: 使用 Supabase 的只读副本
3. **缓存层**: 在应用层使用 Redis 缓存热点数据
4. **归档策略**: 定期归档历史数据

---

## 总结

当前数据库存在 **严重的安全漏洞**，必须立即修复 RLS 问题。性能方面整体可用，但缺失的索引会在数据量增长后成为瓶颈。架构设计基本合理，但存在一些不一致性，需要逐步改进。

**推荐执行顺序**:
1. 🔴 **本周**: 修复所有 P0 安全问题
2. 🟡 **本月**: 优化 P1 性能问题
3. 🔵 **本季度**: 重构 P2 架构问题
4. 🟢 **持续**: 改进 P3 数据质量

遵循本文档的建议，可以将数据库安全性从 30 分提升到 95 分，性能从 55 分提升到 85 分。

---

**文档版本**: v1.0
**生成时间**: 2025-11-23
**适用版本**: PostgreSQL 15.8 / Supabase
**维护者**: AI Assistant
**下次审查**: 2025-12-23
