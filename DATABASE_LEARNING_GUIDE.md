# 数据库设计与优化自学指南

> 专为AI编程者设计的数据库学习文档 —— 从零理解数据库设计原理

## 📖 文档说明

这份文档**不是数据库基础教程**，而是通过分析你的项目数据库，帮助你理解：
- ✅ 为什么要这样设计数据库？
- ✅ 什么是好的设计？什么是坏的设计？
- ✅ 如何用 AI 辅助数据库设计和优化？
- ✅ 遇到问题如何向 AI 提问？

**学习方式**: 结合你的项目实例，边学边做

---

## 🎯 目录

1. [数据库设计的本质](#数据库设计的本质)
2. [你的项目数据库架构解析](#你的项目数据库架构解析)
3. [核心概念深度讲解](#核心概念深度讲解)
4. [常见设计模式](#常见设计模式)
5. [性能优化原理](#性能优化原理)
6. [安全设计原理](#安全设计原理)
7. [AI辅助数据库设计指南](#ai辅助数据库设计指南)
8. [实战练习](#实战练习)

---

## 数据库设计的本质

### 什么是数据库设计？

**简单说**：数据库设计就是**决定如何存储和组织数据**，让它既方便使用又高效安全。

**用生活例子类比**：

想象你要建一个图书馆：

```
坏设计 ❌
把所有书随便堆在一个房间里
→ 找书慢、容易丢、无法管理

好设计 ✅
按类型分区（小说区、技术区）
每本书有编号
用索引卡片记录书的位置
设置门禁，控制谁能进入
→ 找书快、好管理、安全
```

**数据库设计就是做类似的决定**：
1. **分表**（分区）：哪些数据放一起？哪些要分开？
2. **字段设计**（属性）：存什么信息？用什么格式？
3. **索引**（目录）：如何快速找到数据？
4. **关系**（连接）：数据之间怎么关联？
5. **安全**（门禁）：谁能访问什么数据？

---

### 数据库设计的三大目标

#### 1️⃣ 功能完整性（能存对数据）
```sql
-- 能存储业务需要的所有信息
users 表存储用户信息
review_tasks 表存储审核任务
review_results 表存储审核结果
```

#### 2️⃣ 性能（能快速查询）
```sql
-- 通过索引加速查询
CREATE INDEX idx_users_username ON users(username);
-- 查询用户从全表扫描（慢）变成索引查找（快）
```

#### 3️⃣ 安全性（防止数据泄露）
```sql
-- 通过 RLS 限制访问
CREATE POLICY "审核员只能看自己的任务"
ON review_tasks FOR SELECT
USING (reviewer_id = current_user_id());
```

---

## 你的项目数据库架构解析

### 整体架构图

```
┌─────────────────────────────────────────────────┐
│         评论和视频审核平台数据库                  │
└─────────────────────────────────────────────────┘
           │
           ├─── 📝 评论审核模块（6 tables）
           │    ├── comment （评论内容）
           │    ├── review_tasks （一审任务）
           │    ├── review_results （一审结果）
           │    ├── second_review_tasks （二审任务）
           │    ├── second_review_results （二审结果）
           │    └── quality_check_tasks/results （质检）
           │
           ├─── 🎬 视频审核模块（10 tables）
           │    ├── tiktok_videos （视频主表）
           │    ├── video_first_review_tasks/results
           │    ├── video_second_review_tasks/results
           │    ├── video_queue_tasks/results （队列审核）
           │    └── video_quality_tags （质量标签）
           │
           ├─── 👥 用户权限模块（3 tables）
           │    ├── users （用户）
           │    ├── permissions （权限定义）
           │    └── user_permissions （用户-权限关系）
           │
           └─── 🛠️ 系统配置模块（7 tables）
                ├── task_queue/task_queues （任务队列）
                ├── notifications/user_notifications
                ├── tag_config （标签配置）
                ├── moderation_rules （审核规则）
                └── messages/email_verification_logs
```

### 为什么要分这么多表？

**核心原则**：**一个表只做一件事**（单一职责原则）

#### 示例1：为什么评论和审核任务要分开？

```sql
-- ❌ 坏设计：把所有信息都放在一个表
CREATE TABLE comments_with_reviews (
  id BIGINT,
  text TEXT,                    -- 评论内容
  reviewer_id INT,              -- 审核员ID
  status VARCHAR,               -- 任务状态
  claimed_at TIMESTAMP,         -- 领取时间
  is_approved BOOLEAN,          -- 审核结果
  tags TEXT[]                   -- 违规标签
);

-- 问题：
-- 1. 一条评论可能被多次审核（一审、二审、质检），无法表示
-- 2. 未被审核的评论，审核字段全是 NULL，浪费空间
-- 3. 查询评论内容时，要带着一堆审核字段，效率低
```

```sql
-- ✅ 好设计：分离关注点
CREATE TABLE comment (
  id BIGINT PRIMARY KEY,
  text TEXT                     -- 只存评论内容
);

CREATE TABLE review_tasks (
  id SERIAL PRIMARY KEY,
  comment_id BIGINT,            -- 关联评论
  reviewer_id INT,              -- 审核员
  status VARCHAR,               -- 任务状态
  claimed_at TIMESTAMP          -- 领取时间
);

CREATE TABLE review_results (
  id SERIAL PRIMARY KEY,
  task_id INT,                  -- 关联任务
  is_approved BOOLEAN,          -- 审核结果
  tags TEXT[]                   -- 违规标签
);

-- 优势：
-- ✅ 一条评论可以有多个审核任务
-- ✅ 数据清晰，无冗余
-- ✅ 查询评论内容不需要关心审核信息
```

#### 示例2：为什么一审和二审要分表？

```sql
-- ❌ 坏设计：用一个字段区分一审和二审
CREATE TABLE review_tasks_all (
  id SERIAL,
  review_stage VARCHAR,  -- 'first' 或 'second'
  comment_id BIGINT,
  reviewer_id INT,
  ...
);

-- 问题：
-- 1. 二审需要引用一审结果，这种设计无法表达
-- 2. 查询一审任务时，要加 WHERE review_stage = 'first'，性能差
-- 3. 业务逻辑混乱，代码容易出错
```

```sql
-- ✅ 好设计：分成两个独立的表
CREATE TABLE review_tasks (      -- 一审
  id SERIAL PRIMARY KEY,
  comment_id BIGINT
);

CREATE TABLE second_review_tasks ( -- 二审
  id SERIAL PRIMARY KEY,
  first_review_result_id INT,     -- 引用一审结果
  comment_id BIGINT
);

-- 优势：
-- ✅ 表达了业务流程：二审基于一审
-- ✅ 查询一审任务不需要过滤条件
-- ✅ 代码逻辑清晰
```

---

### 表之间的关系

#### 1. 一对多关系（最常见）

**定义**：一条 A 记录对应多条 B 记录

**你的项目中的例子**：

```sql
-- 一个审核员可以领取多个任务
users (1) ←→ (N) review_tasks

-- 一个评论可以有多个审核任务
comment (1) ←→ (N) review_tasks

-- 一个任务只有一个审核结果
review_tasks (1) ←→ (1) review_results
```

**如何实现**：在"多"的一方加外键

```sql
CREATE TABLE review_tasks (
  id SERIAL PRIMARY KEY,
  reviewer_id INT,                         -- 外键指向 users
  comment_id BIGINT,                       -- 外键指向 comment
  FOREIGN KEY (reviewer_id) REFERENCES users(id),
  FOREIGN KEY (comment_id) REFERENCES comment(id)
);
```

**生活类比**：
```
一个班主任（1）管理多个学生（N）
学生表：student_id, name, teacher_id（外键）
```

---

#### 2. 多对多关系

**定义**：多条 A 记录对应多条 B 记录

**你的项目中的例子**：

```sql
-- 一个用户可以有多个权限
-- 一个权限可以分配给多个用户
users (N) ←→ (N) permissions
```

**如何实现**：创建中间表

```sql
-- 不要直接在 users 表加 permissions 字段
-- ❌ CREATE TABLE users (permissions TEXT[]);

-- ✅ 创建关系表
CREATE TABLE user_permissions (
  user_id INT REFERENCES users(id),
  permission_key VARCHAR REFERENCES permissions(permission_key),
  PRIMARY KEY (user_id, permission_key)  -- 联合主键
);
```

**生活类比**：
```
学生（N）可以选修多门课程（N）
中间表：student_course (student_id, course_id, grade)
```

---

### 数据类型选择

#### 整数类型（用于ID、计数）

```sql
-- 你的项目使用的类型
SERIAL        -- 自增整数，范围 1 到 2,147,483,647（21亿）
BIGSERIAL     -- 大整数，范围 1 到 9,223,372,036,854,775,807（92亿亿）
INTEGER       -- 普通整数
BIGINT        -- 大整数

-- 如何选择？
users.id → INTEGER (SERIAL)        -- 用户不会超过21亿
comment.id → BIGINT (BIGSERIAL)    -- 评论可能很多，用大整数
file_size → BIGINT                 -- 文件大小可能超过2GB
```

**常见错误**：

```sql
-- ❌ 用 TEXT 存储数字
CREATE TABLE messages (
  user_id TEXT  -- 应该用 INTEGER
);

-- 问题：
-- 1. 浪费空间（"123" 占 3 字节，整数只占 4 字节）
-- 2. 无法高效排序和比较
-- 3. 容易存入非法数据（"abc"）
```

---

#### 文本类型

```sql
VARCHAR(n)    -- 变长字符串，最多 n 个字符
TEXT          -- 无限长文本
CHAR(n)       -- 定长字符串（很少用）

-- 你的项目中的使用
users.username → VARCHAR      -- 用户名有长度限制
users.role → VARCHAR          -- 角色类型固定（'reviewer', 'admin'）
comment.text → TEXT           -- 评论内容长度不确定
```

**如何选择**？

```sql
-- 固定长度、有限选项 → VARCHAR + CHECK
role VARCHAR CHECK (role IN ('reviewer', 'admin'))

-- 短文本 → VARCHAR
email VARCHAR(255)

-- 长文本、不确定长度 → TEXT
comment_text TEXT
```

---

#### 时间类型

```sql
TIMESTAMP              -- 带时区的时间戳（推荐）
TIMESTAMP WITH TIME ZONE  -- 明确带时区（更推荐）
DATE                   -- 只存日期
TIME                   -- 只存时间

-- 你的项目中的使用
created_at TIMESTAMP DEFAULT NOW()
updated_at TIMESTAMP DEFAULT NOW()
```

**时区问题**：

```sql
-- ❌ 不带时区
created_at TIMESTAMP
-- 问题：存储时丢失了时区信息
-- 用户在美国提交，服务器在中国，会混乱

-- ✅ 带时区
created_at TIMESTAMP WITH TIME ZONE
-- 存储：2025-11-23 10:00:00+08:00（北京时间）
-- 美国用户看到：2025-11-22 18:00:00-08:00（自动转换）
```

---

#### 布尔类型

```sql
BOOLEAN  -- TRUE / FALSE / NULL

-- 你的项目中的使用
is_approved BOOLEAN   -- 是否通过审核
is_active BOOLEAN     -- 是否激活
email_verified BOOLEAN -- 邮箱是否验证
```

**命名规范**：
```sql
✅ is_approved, has_permission, can_edit
❌ approved, permission, edit  （不清晰）
```

---

#### 数组和JSON类型（PostgreSQL特有）

```sql
TEXT[]     -- 文本数组
JSONB      -- 二进制JSON（可索引）
JSON       -- 文本JSON（不可索引）

-- 你的项目中的使用
tags TEXT[]                     -- 违规标签数组
quality_dimensions JSONB        -- 视频质量维度（复杂结构）
```

**何时使用数组**？

```sql
-- ✅ 适合用数组
tags TEXT[]   -- 标签数量少、无需复杂查询

-- ❌ 不适合用数组
reviewer_ids INT[]  -- 应该用关系表
-- 问题：无法统计每个审核员的工作量
```

**何时使用JSON**？

```sql
-- ✅ 适合用JSON
quality_dimensions JSONB {
  "content_quality": {"score": 8, "tags": ["精彩", "创意"]},
  "technical_quality": {"score": 7, "tags": []}
}
-- 结构复杂、字段不固定、无需频繁查询

-- ❌ 不适合用JSON
user_info JSONB {"name": "张三", "age": 25}
-- 应该用独立字段：name VARCHAR, age INT
```

---

## 核心概念深度讲解

### 1. 主键（Primary Key）

#### 什么是主键？

**定义**：唯一标识一条记录的字段

**你的项目中的主键**：

```sql
users.id           → INTEGER (SERIAL)
comment.id         → BIGINT
review_tasks.id    → SERIAL
```

#### 为什么需要主键？

```sql
-- ❌ 没有主键
CREATE TABLE users (
  username VARCHAR,
  email VARCHAR
);

-- 问题1：无法区分两个同名用户
INSERT INTO users VALUES ('张三', 'a@qq.com');
INSERT INTO users VALUES ('张三', 'b@qq.com');  -- 无法区分

-- 问题2：无法更新特定用户
UPDATE users SET email = 'new@qq.com' WHERE ... ?
-- WHERE 条件写什么？username 可能重复

-- 问题3：无法被其他表引用
CREATE TABLE orders (
  user_id ??? REFERENCES users(???)  -- 引用哪个字段？
);
```

```sql
-- ✅ 有主键
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  username VARCHAR,
  email VARCHAR
);

-- 每个用户有唯一ID
-- 更新用户：UPDATE users SET email = 'x' WHERE id = 123;
-- 被引用：orders.user_id REFERENCES users(id)
```

---

#### 主键设计原则

**原则1：主键应该无业务意义**

```sql
-- ❌ 用手机号做主键
CREATE TABLE users (
  phone VARCHAR PRIMARY KEY,  -- 用户换手机号怎么办？
  name VARCHAR
);

-- ✅ 用自增ID
CREATE TABLE users (
  id SERIAL PRIMARY KEY,      -- 永不改变
  phone VARCHAR,              -- 可以修改
  name VARCHAR
);
```

**原则2：主键应该简短**

```sql
-- ❌ 复合主键太长
CREATE TABLE user_permissions (
  user_email VARCHAR,
  permission_name VARCHAR,
  PRIMARY KEY (user_email, permission_name)  -- 占用空间大
);

-- ✅ 用代理主键
CREATE TABLE user_permissions (
  id SERIAL PRIMARY KEY,      -- 简短
  user_id INT,                -- 用ID代替email
  permission_id INT
);
```

**原则3：主键应该不可为NULL**

```sql
-- 主键自动 NOT NULL
CREATE TABLE users (
  id SERIAL PRIMARY KEY  -- 自动 NOT NULL
);
```

---

### 2. 外键（Foreign Key）

#### 什么是外键？

**定义**：指向另一个表主键的字段，用于建立表之间的关系

**你的项目中的外键**：

```sql
review_tasks.reviewer_id → users.id
review_tasks.comment_id → comment.id
review_results.task_id → review_tasks.id
```

#### 外键的作用

**作用1：保证数据完整性**

```sql
-- ❌ 没有外键约束
CREATE TABLE review_tasks (
  reviewer_id INT  -- 只是普通整数
);

-- 可以插入不存在的用户ID
INSERT INTO review_tasks (reviewer_id) VALUES (99999);  -- 成功但错误

-- ✅ 有外键约束
CREATE TABLE review_tasks (
  reviewer_id INT REFERENCES users(id)
);

-- 插入不存在的用户ID会报错
INSERT INTO review_tasks (reviewer_id) VALUES (99999);
-- ERROR: 违反外键约束
```

**作用2：级联操作**

```sql
-- 删除用户时，同时删除其所有任务
CREATE TABLE review_tasks (
  reviewer_id INT REFERENCES users(id) ON DELETE CASCADE
);

-- 删除用户
DELETE FROM users WHERE id = 1;
-- 自动删除该用户的所有 review_tasks
```

**⚠️ 警告**：生产环境慎用 `ON DELETE CASCADE`

```sql
-- 更安全的做法：软删除
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP;

-- 删除用户
UPDATE users SET deleted_at = NOW() WHERE id = 1;

-- 查询时排除已删除用户
SELECT * FROM users WHERE deleted_at IS NULL;
```

---

### 3. 索引（Index）

#### 什么是索引？

**定义**：加速查询的数据结构

**生活类比**：

```
没有索引 = 逐页翻书找内容
有索引 = 看目录直接翻到对应页
```

#### 你的项目中的索引使用

```sql
-- 查看某个表的所有索引
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'review_tasks';

-- 结果示例：
idx_review_tasks_status
idx_review_tasks_reviewer_id
idx_review_tasks_comment_id
```

#### 何时需要索引？

**场景1：频繁用于 WHERE 条件的字段**

```sql
-- 查询：查找待审核的任务
SELECT * FROM review_tasks WHERE status = 'pending';

-- 应该加索引
CREATE INDEX idx_review_tasks_status ON review_tasks(status);
```

**场景2：用于 JOIN 的字段**

```sql
-- 查询：获取任务及评论内容
SELECT rt.*, c.text
FROM review_tasks rt
JOIN comment c ON rt.comment_id = c.id;

-- comment_id 应该有索引
CREATE INDEX idx_review_tasks_comment_id ON review_tasks(comment_id);
```

**场景3：用于排序的字段**

```sql
-- 查询：按创建时间排序
SELECT * FROM review_tasks ORDER BY created_at DESC;

-- 应该加索引
CREATE INDEX idx_review_tasks_created_at ON review_tasks(created_at DESC);
```

---

#### 索引的代价

**索引不是越多越好！**

```sql
-- ❌ 过度索引
CREATE INDEX idx1 ON users(username);
CREATE INDEX idx2 ON users(email);
CREATE INDEX idx3 ON users(created_at);
CREATE INDEX idx4 ON users(updated_at);
CREATE INDEX idx5 ON users(role);
CREATE INDEX idx6 ON users(status);
-- 太多了！

-- 问题：
-- 1. 每次 INSERT/UPDATE 都要更新所有索引，写入变慢
-- 2. 占用大量存储空间
-- 3. 大部分索引可能从未使用

-- ✅ 只为常用查询加索引
CREATE INDEX idx_users_username ON users(username);  -- 登录查询
CREATE INDEX idx_users_role_status ON users(role, status);  -- 用户列表
-- 足够了
```

---

#### 复合索引

**定义**：在多个字段上建立的索引

**示例**：

```sql
-- 查询：获取某个审核员的待处理任务
SELECT * FROM review_tasks
WHERE reviewer_id = 123 AND status = 'pending';

-- 方案1：单字段索引
CREATE INDEX idx_reviewer_id ON review_tasks(reviewer_id);
CREATE INDEX idx_status ON review_tasks(status);
-- 数据库只会用其中一个

-- 方案2：复合索引（更好）
CREATE INDEX idx_reviewer_status ON review_tasks(reviewer_id, status);
-- 同时匹配两个条件，效率更高
```

**复合索引的顺序规则**：

```sql
-- 索引：(reviewer_id, status)

-- ✅ 可以使用索引
WHERE reviewer_id = 123                          -- 匹配第一个字段
WHERE reviewer_id = 123 AND status = 'pending'   -- 匹配全部字段

-- ❌ 无法使用索引
WHERE status = 'pending'                         -- 跳过第一个字段
```

**记忆技巧**：索引像电话簿

```
电话簿索引：(姓氏, 名字)
✅ 可以查：姓"张"的所有人
✅ 可以查：姓"张"名"三"的人
❌ 无法查：名字是"三"的所有人（不知道姓什么）
```

---

### 4. 约束（Constraints）

#### 什么是约束？

**定义**：限制表中数据的规则，确保数据正确性

#### 你的项目中使用的约束

**1. NOT NULL 约束**

```sql
-- users 表
username VARCHAR NOT NULL  -- 用户名必填
password VARCHAR NOT NULL  -- 密码必填
role VARCHAR NOT NULL      -- 角色必填
```

**2. UNIQUE 约束**

```sql
-- users 表
username VARCHAR UNIQUE    -- 用户名不能重复
email VARCHAR UNIQUE       -- 邮箱不能重复
```

**3. CHECK 约束**

```sql
-- users 表
role VARCHAR CHECK (role IN ('reviewer', 'admin'))
-- 角色只能是这两个值之一

status VARCHAR CHECK (status IN ('pending', 'approved', 'rejected'))
-- 状态只能是这三个值之一

-- video_first_review_results 表
overall_score INT CHECK (overall_score >= 1 AND overall_score <= 40)
-- 分数必须在 1-40 之间
```

**4. DEFAULT 约束**

```sql
status VARCHAR DEFAULT 'pending'
created_at TIMESTAMP DEFAULT NOW()
is_active BOOLEAN DEFAULT true
```

---

#### 约束的好处

**好处1：防止错误数据**

```sql
-- ❌ 没有约束
CREATE TABLE users (
  role VARCHAR
);

-- 可以插入错误数据
INSERT INTO users (role) VALUES ('superadmin');  -- 拼写错误
INSERT INTO users (role) VALUES (NULL);          -- 忘记填写

-- ✅ 有约束
CREATE TABLE users (
  role VARCHAR NOT NULL CHECK (role IN ('reviewer', 'admin'))
);

-- 插入错误数据会报错
INSERT INTO users (role) VALUES ('superadmin');  -- ERROR
INSERT INTO users (role) VALUES (NULL);          -- ERROR
```

**好处2：简化业务逻辑**

```go
// ❌ 在代码中验证
func CreateUser(username, role string) error {
  if role != "reviewer" && role != "admin" {
    return errors.New("invalid role")
  }
  // 插入数据库...
}

// ✅ 依赖数据库约束
func CreateUser(username, role string) error {
  // 直接插入，数据库会自动验证
  _, err := db.Exec("INSERT INTO users (username, role) VALUES ($1, $2)", username, role)
  return err  // 数据库返回错误
}
```

---

## 常见设计模式

### 模式1：审计字段（Audit Fields）

**定义**：记录数据的创建和修改信息

**标准模式**：

```sql
CREATE TABLE any_table (
  id SERIAL PRIMARY KEY,
  ...业务字段...,

  -- 审计字段
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  created_by INT REFERENCES users(id),
  updated_by INT REFERENCES users(id)
);

-- 自动更新 updated_at
CREATE TRIGGER update_timestamp
BEFORE UPDATE ON any_table
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
```

**你的项目中的使用**：

```sql
-- users 表
created_at TIMESTAMP DEFAULT now()
updated_at TIMESTAMP DEFAULT now()

-- task_queue 表
created_at TIMESTAMP DEFAULT now()
updated_at TIMESTAMP DEFAULT now()
created_by INT REFERENCES users(id)
updated_by INT REFERENCES users(id)
```

**为什么需要审计字段？**

```sql
-- 问题1：谁创建的这个用户？
SELECT * FROM users WHERE username = 'test';
-- 无法知道是哪个管理员创建的

-- ✅ 有 created_by
SELECT u.*, creator.username as created_by_username
FROM users u
LEFT JOIN users creator ON u.created_by = creator.id
WHERE u.username = 'test';
-- 可以看到是哪个管理员操作的

-- 问题2：这个任务什么时候被更新的？
-- ✅ 有 updated_at
SELECT * FROM task_queue WHERE updated_at > NOW() - INTERVAL '1 day';
-- 可以查询最近更新的任务
```

---

### 模式2：软删除（Soft Delete）

**定义**：不真正删除数据，而是标记为已删除

**标准模式**：

```sql
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP;

-- 删除用户（软删除）
UPDATE users SET deleted_at = NOW() WHERE id = 1;

-- 查询活跃用户
SELECT * FROM users WHERE deleted_at IS NULL;

-- 创建视图自动过滤
CREATE VIEW users_active AS
SELECT * FROM users WHERE deleted_at IS NULL;

-- 使用视图
SELECT * FROM users_active;
```

**软删除 vs 硬删除**：

```sql
-- ❌ 硬删除
DELETE FROM users WHERE id = 1;
-- 问题：数据永久丢失，无法恢复
-- 问题：外键引用会报错或级联删除

-- ✅ 软删除
UPDATE users SET deleted_at = NOW() WHERE id = 1;
-- 优势：可以恢复
-- 优势：保留历史数据用于审计
-- 优势：不破坏外键关系
```

**恢复软删除的数据**：

```sql
-- 恢复用户
UPDATE users SET deleted_at = NULL WHERE id = 1;
```

---

### 模式3：枚举值存储

**问题**：如何存储有限的选项（如状态、角色）？

**方案对比**：

#### 方案1：直接用字符串 + CHECK约束（你的项目用的）

```sql
CREATE TABLE users (
  role VARCHAR CHECK (role IN ('reviewer', 'admin'))
);

-- 优点：简单直接
-- 缺点：难以修改选项（要改约束）
```

#### 方案2：用单独的枚举表

```sql
CREATE TABLE roles (
  id SERIAL PRIMARY KEY,
  name VARCHAR UNIQUE,
  description TEXT
);

INSERT INTO roles VALUES (1, 'reviewer', '审核员');
INSERT INTO roles VALUES (2, 'admin', '管理员');

CREATE TABLE users (
  role_id INT REFERENCES roles(id)
);

-- 优点：可以动态增加角色
-- 缺点：查询时需要JOIN
```

#### 方案3：用 PostgreSQL ENUM 类型

```sql
CREATE TYPE user_role AS ENUM ('reviewer', 'admin');

CREATE TABLE users (
  role user_role
);

-- 优点：类型安全，占用空间小
-- 缺点：修改ENUM很麻烦
```

**如何选择**？

```
选项很少且几乎不变（如性别）→ 方案1 或 方案3
选项可能增加（如权限）→ 方案2
```

---

### 模式4：一对多关系建模

**你的项目实例**：一个评论可以有多个审核任务

```sql
-- comment 表（一）
CREATE TABLE comment (
  id BIGINT PRIMARY KEY,
  text TEXT
);

-- review_tasks 表（多）
CREATE TABLE review_tasks (
  id SERIAL PRIMARY KEY,
  comment_id BIGINT REFERENCES comment(id),  -- 外键
  status VARCHAR,
  ...
);
```

**常见错误**：在"一"的一方存储"多"的信息

```sql
-- ❌ 错误设计
CREATE TABLE comment (
  id BIGINT,
  text TEXT,
  task_ids INT[]  -- 存储所有任务ID
);

-- 问题：
-- 1. 每次创建任务要更新 comment 表
-- 2. 无法保证 task_ids 的正确性（可能包含不存在的ID）
-- 3. 查询复杂

-- ✅ 正确设计
-- 在 review_tasks 表加 comment_id 外键
```

---

### 模式5：多对多关系建模

**你的项目实例**：用户与权限的多对多关系

```sql
-- users 表
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  username VARCHAR
);

-- permissions 表
CREATE TABLE permissions (
  permission_key VARCHAR PRIMARY KEY,
  name VARCHAR
);

-- 中间表（关系表）
CREATE TABLE user_permissions (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id),
  permission_key VARCHAR REFERENCES permissions(permission_key),
  granted_at TIMESTAMP DEFAULT NOW(),
  granted_by INT REFERENCES users(id)
);
```

**为什么需要中间表**？

```
用户 A 有权限 P1, P2
用户 B 有权限 P2, P3
权限 P1 分配给 用户 A
权限 P2 分配给 用户 A, B

如果不用中间表：
users.permissions → 如何存储？数组？太复杂
permissions.users → 如何存储？数组？太复杂

用中间表：
user_permissions:
(user_id=A, permission_key=P1)
(user_id=A, permission_key=P2)
(user_id=B, permission_key=P2)
(user_id=B, permission_key=P3)
```

**查询示例**：

```sql
-- 查询用户A的所有权限
SELECT p.*
FROM permissions p
JOIN user_permissions up ON p.permission_key = up.permission_key
WHERE up.user_id = A;

-- 查询拥有权限P1的所有用户
SELECT u.*
FROM users u
JOIN user_permissions up ON u.id = up.user_id
WHERE up.permission_key = 'P1';
```

---

## 性能优化原理

### 为什么查询会慢？

**理解查询过程**：

```sql
SELECT * FROM review_tasks WHERE status = 'pending';

-- 数据库执行步骤：
1. 扫描整个 review_tasks 表
2. 逐行检查 status 字段是否等于 'pending'
3. 返回匹配的行

-- 如果表有 100 万行，就要检查 100 万次！
```

**加了索引后**：

```sql
CREATE INDEX idx_status ON review_tasks(status);

-- 数据库执行步骤：
1. 在索引中查找 'pending' → 瞬间找到
2. 根据索引直接定位到对应行
3. 返回结果

-- 只需要检查几十次！
```

---

### 常见慢查询模式

#### 模式1：全表扫描

```sql
-- ❌ 慢查询
SELECT * FROM review_tasks WHERE reviewer_id = 123;
-- 如果没有索引，要扫描全表

-- ✅ 优化
CREATE INDEX idx_reviewer_id ON review_tasks(reviewer_id);
```

---

#### 模式2：JOIN 慢

```sql
-- ❌ 慢查询
SELECT rt.*, c.text
FROM review_tasks rt
JOIN comment c ON rt.comment_id = c.id;
-- 如果 comment_id 没有索引，JOIN 会很慢

-- ✅ 优化
CREATE INDEX idx_comment_id ON review_tasks(comment_id);
-- 外键字段一定要有索引！
```

**你的项目中发现的问题**：14 个外键没有索引

```sql
-- 需要添加的索引
CREATE INDEX idx_quality_check_tasks_comment_id
ON quality_check_tasks(comment_id);

CREATE INDEX idx_second_review_tasks_comment_id
ON second_review_tasks(comment_id);

-- 等等...
```

---

#### 模式3：ORDER BY 慢

```sql
-- ❌ 慢查询
SELECT * FROM review_tasks ORDER BY created_at DESC LIMIT 10;
-- 如果没有索引，要排序全表

-- ✅ 优化
CREATE INDEX idx_created_at ON review_tasks(created_at DESC);
-- 注意：DESC 很重要！
```

---

#### 模式4：COUNT(*) 慢

```sql
-- ❌ 慢查询
SELECT COUNT(*) FROM review_tasks WHERE status = 'pending';
-- 要扫描全表计数

-- ✅ 优化方案1：加索引
CREATE INDEX idx_status ON review_tasks(status);

-- ✅ 优化方案2：用统计表
CREATE TABLE review_stats (
  status VARCHAR PRIMARY KEY,
  count INT
);

-- 更新统计（用触发器自动）
CREATE TRIGGER update_stats AFTER INSERT ON review_tasks...
```

---

### EXPLAIN：分析查询性能

**如何查看查询计划**：

```sql
EXPLAIN ANALYZE
SELECT * FROM review_tasks WHERE status = 'pending';

-- 输出：
Seq Scan on review_tasks  (cost=0.00..1234.56 rows=100 width=...)
  Filter: (status = 'pending'::text)

-- Seq Scan = 顺序扫描（全表扫描）→ 慢！
```

**加索引后**：

```sql
EXPLAIN ANALYZE
SELECT * FROM review_tasks WHERE status = 'pending';

-- 输出：
Index Scan using idx_status on review_tasks  (cost=0.00..8.27 rows=100 width=...)

-- Index Scan = 索引扫描 → 快！
-- cost 从 1234.56 降到 8.27
```

**如何读懂 EXPLAIN 输出**：

```
Seq Scan         → 全表扫描（坏）
Index Scan       → 索引扫描（好）
Index Only Scan  → 只扫描索引（最好）
Nested Loop      → 嵌套循环 JOIN（可能慢）
Hash Join        → 哈希 JOIN（一般）
Merge Join       → 归并 JOIN（好）

rows=100        → 预计返回100行
cost=8.27       → 预估成本（越低越好）
actual time=0.1 → 实际耗时（毫秒）
```

---

## 安全设计原理

### RLS（行级安全）

#### 什么是 RLS？

**定义**：控制谁可以看到表中的哪些行

**生活类比**：

```
医院病历系统：
- 医生只能看自己负责的病人的病历
- 病人只能看自己的病历
- 管理员可以看所有病历

→ 同一个 medical_records 表，不同角色看到不同的数据行
```

---

#### 你的项目为什么需要 RLS？

**当前问题**：数据库完全暴露

```sql
-- 任何人都可以通过 PostgREST API 访问：
https://your-db.supabase.co/rest/v1/users
https://your-db.supabase.co/rest/v1/review_tasks

-- 即使没有登录，也能看到所有数据！
```

**启用 RLS 后**：

```sql
-- 1. 启用 RLS
ALTER TABLE review_tasks ENABLE ROW LEVEL SECURITY;

-- 2. 创建策略：审核员只能看自己的任务
CREATE POLICY "审核员看自己的任务"
ON review_tasks FOR SELECT
USING (reviewer_id = current_user_id());

-- 3. 现在访问 API：
-- 登录用户 123 访问：只返回 reviewer_id = 123 的任务
-- 未登录访问：返回空（无权限）
```

---

#### RLS 策略示例

**策略1：用户只能看自己的数据**

```sql
CREATE POLICY "users_select_own"
ON users FOR SELECT
USING (id = current_user_id());

-- 用户 123 查询：
SELECT * FROM users;
-- 实际执行：
SELECT * FROM users WHERE id = 123;
```

**策略2：审核员可以看已领取的任务**

```sql
CREATE POLICY "reviewers_select_claimed_tasks"
ON review_tasks FOR SELECT
USING (
  reviewer_id = current_user_id()
  OR status = 'pending'  -- 待领取的任务也可以看
);
```

**策略3：管理员可以看所有数据**

```sql
CREATE POLICY "admins_select_all"
ON review_tasks FOR SELECT
USING (
  EXISTS (
    SELECT 1 FROM users
    WHERE id = current_user_id()
    AND role = 'admin'
  )
);
```

**策略4：插入时自动设置 reviewer_id**

```sql
CREATE POLICY "insert_task_claim"
ON review_tasks FOR INSERT
WITH CHECK (reviewer_id = current_user_id());

-- 用户只能插入自己作为审核员的任务
INSERT INTO review_tasks (comment_id, reviewer_id)
VALUES (456, 123);  -- 只有用户123能插入
```

---

### SQL注入防御

**什么是SQL注入？**

```go
// ❌ 危险代码
username := r.FormValue("username")  // 用户输入
query := "SELECT * FROM users WHERE username = '" + username + "'"
db.Query(query)

// 如果用户输入：' OR '1'='1
// 查询变成：
SELECT * FROM users WHERE username = '' OR '1'='1'
// 返回所有用户！

// 如果用户输入：'; DROP TABLE users; --
// 查询变成：
SELECT * FROM users WHERE username = ''; DROP TABLE users; --'
// 数据库被删除！
```

**防御方法：使用参数化查询**

```go
// ✅ 安全代码
username := r.FormValue("username")
query := "SELECT * FROM users WHERE username = $1"
db.Query(query, username)

// 数据库会自动转义，无论用户输入什么
```

---

### 密码存储

```go
// ❌ 绝不要明文存储密码
password := "123456"
db.Exec("INSERT INTO users (password) VALUES ($1)", password)

// ✅ 使用 bcrypt 加密
import "golang.org/x/crypto/bcrypt"

hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
db.Exec("INSERT INTO users (password) VALUES ($1)", hashedPassword)

// 验证密码
storedHash := ... // 从数据库读取
err := bcrypt.CompareHashAndPassword(storedHash, []byte(inputPassword))
if err == nil {
  // 密码正确
}
```

---

## AI辅助数据库设计指南

### 如何向 AI 提问？

#### ❌ 不好的提问

```
"我的数据库有问题"
"怎么优化数据库？"
"帮我设计数据库"
```

**问题**：太笼统，AI 无法给出具体建议

---

#### ✅ 好的提问模板

**模板1：设计新表**

```
我要设计一个 [业务功能]，需要存储以下信息：
1. [信息1，类型]
2. [信息2，类型]
3. ...

业务场景：
- [场景1描述]
- [场景2描述]

请帮我设计表结构，包括：
1. 表名和字段设计
2. 主键和外键
3. 索引建议
4. 约束建议
```

**示例**：

```
我要设计一个视频审核系统，需要存储：
1. 视频文件（存储在R2，需要记录路径和URL）
2. 审核任务（分配给审核员）
3. 审核结果（质量分数、标签、推荐流量池）

业务场景：
- 一个视频可以被多次审核（一审、二审）
- 审核员领取任务后24小时内需要完成
- 需要统计每个审核员的工作量

请帮我设计表结构...
```

---

**模板2：优化现有表**

```
我的 [表名] 表有性能问题：

当前结构：
[粘贴 CREATE TABLE 语句]

慢查询：
[粘贴 SQL 查询]

EXPLAIN 输出：
[粘贴 EXPLAIN ANALYZE 结果]

请帮我分析：
1. 为什么慢？
2. 如何优化？
3. 需要加什么索引？
```

---

**模板3：架构设计评审**

```
我的项目有以下几个表：
[列出所有表和关系]

业务逻辑：
[描述核心业务流程]

请评审：
1. 表设计是否合理？
2. 有没有冗余或缺失？
3. 关系是否正确？
4. 有什么潜在问题？
```

---

### 使用 AI 的最佳实践

**1. 提供足够的上下文**

```
❌ "帮我优化这个查询"
SELECT * FROM review_tasks WHERE status = 'pending';

✅ "帮我优化这个查询"
表结构：
CREATE TABLE review_tasks (
  id SERIAL PRIMARY KEY,
  status VARCHAR,
  ...
);

现有索引：
idx_review_tasks_status

当前查询速度：200ms
数据量：100万行
查询频率：每秒100次

请分析为什么还是慢？
```

---

**2. 让 AI 解释原因**

```
✅ 不只是要答案，还要理解为什么

"为什么要用 BIGINT 而不是 INTEGER？"
"为什么要加这个索引？"
"为什么要分成两个表而不是一个表？"
```

---

**3. 让 AI 提供多个方案**

```
✅ "请提供 3 种设计方案，并比较优劣"

方案1：用数组存储标签
方案2：用关系表存储标签
方案3：用 JSONB 存储标签

对比：性能、灵活性、查询复杂度
```

---

**4. 验证 AI 的建议**

```
✅ "这个设计有什么潜在问题？"
✅ "在什么情况下这个方案会失效？"
✅ "有没有更好的替代方案？"
```

---

### AI 辅助 SQL 编写

**场景1：复杂查询**

```
向 AI 提问：
"我要查询：
1. 每个审核员今天完成的任务数
2. 按完成数量降序排列
3. 只显示完成数 > 10 的审核员

涉及的表：
- review_tasks (id, reviewer_id, status, completed_at)
- users (id, username)

请帮我写SQL查询"
```

**场景2：数据迁移**

```
向 AI 提问：
"我要把 task_queues 表的数据迁移到 task_queue 表

源表结构：
CREATE TABLE task_queues (...);

目标表结构：
CREATE TABLE task_queue (...);

注意事项：
1. queue_name 可能重复，需要去重
2. 迁移时要验证数据完整性

请帮我写迁移脚本"
```

**场景3：性能分析**

```
向 AI 提问：
"请帮我分析这个 EXPLAIN 输出，并解释为什么慢：
[粘贴 EXPLAIN ANALYZE 结果]"
```

---

## 实战练习

### 练习1：设计博客系统数据库

**需求**：

```
1. 用户可以发布文章
2. 文章可以有多个标签
3. 用户可以评论文章
4. 用户可以点赞文章和评论
```

**提示**：

- 需要哪些表？
- 表之间是什么关系？
- 需要哪些索引？

**参考答案**：

```sql
-- 用户表
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  username VARCHAR UNIQUE NOT NULL,
  email VARCHAR UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

-- 文章表
CREATE TABLE articles (
  id SERIAL PRIMARY KEY,
  author_id INT REFERENCES users(id),
  title VARCHAR NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_articles_author ON articles(author_id);

-- 标签表
CREATE TABLE tags (
  id SERIAL PRIMARY KEY,
  name VARCHAR UNIQUE NOT NULL
);

-- 文章-标签关系表（多对多）
CREATE TABLE article_tags (
  article_id INT REFERENCES articles(id),
  tag_id INT REFERENCES tags(id),
  PRIMARY KEY (article_id, tag_id)
);

-- 评论表
CREATE TABLE comments (
  id SERIAL PRIMARY KEY,
  article_id INT REFERENCES articles(id),
  user_id INT REFERENCES users(id),
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_comments_article ON comments(article_id);

-- 点赞表
CREATE TABLE likes (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id),
  target_type VARCHAR CHECK (target_type IN ('article', 'comment')),
  target_id INT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE (user_id, target_type, target_id)  -- 一个用户只能点赞一次
);
```

---

### 练习2：分析慢查询

**场景**：

```sql
-- 查询某个用户的所有文章及评论数
SELECT
  a.id,
  a.title,
  COUNT(c.id) as comment_count
FROM articles a
LEFT JOIN comments c ON a.id = c.article_id
WHERE a.author_id = 123
GROUP BY a.id;

-- EXPLAIN 显示：
Seq Scan on articles (cost=0..1000)
  Filter: (author_id = 123)
Hash Join (cost=...)
```

**问题**：为什么慢？如何优化？

**分析**：

1. `Seq Scan` 说明没有使用索引
2. `articles.author_id` 需要加索引
3. `comments.article_id` 也需要加索引（JOIN用）

**优化**：

```sql
CREATE INDEX idx_articles_author ON articles(author_id);
CREATE INDEX idx_comments_article ON comments(article_id);
```

---

### 练习3：设计 RLS 策略

**需求**：

```
博客系统的访问控制：
1. 任何人可以看已发布的文章
2. 作者可以看自己的所有文章（包括草稿）
3. 评论：作者和文章作者可以删除
```

**参考答案**：

```sql
-- 启用 RLS
ALTER TABLE articles ENABLE ROW LEVEL SECURITY;
ALTER TABLE comments ENABLE ROW LEVEL SECURITY;

-- 文章：所有人可以看已发布的
CREATE POLICY "public_read_published"
ON articles FOR SELECT
USING (status = 'published');

-- 文章：作者可以看自己的所有文章
CREATE POLICY "author_read_own"
ON articles FOR SELECT
USING (author_id = current_user_id());

-- 评论：评论作者可以删除自己的评论
CREATE POLICY "comment_author_delete"
ON comments FOR DELETE
USING (user_id = current_user_id());

-- 评论：文章作者可以删除文章的评论
CREATE POLICY "article_author_delete_comments"
ON comments FOR DELETE
USING (
  EXISTS (
    SELECT 1 FROM articles
    WHERE id = comments.article_id
    AND author_id = current_user_id()
  )
);
```

---

## 学习资源推荐

### 在线工具

1. **dbdiagram.io** - 可视化数据库设计
   ```
   用法：画ER图，自动生成SQL
   ```

2. **explain.depesz.com** - 分析EXPLAIN输出
   ```
   用法：粘贴EXPLAIN结果，可视化分析
   ```

3. **pgAdmin** - PostgreSQL管理工具
   ```
   用法：图形化管理数据库
   ```

### 推荐阅读

1. **《数据库系统概念》** - 理论基础
2. **《高性能MySQL》** - 性能优化（适用于PostgreSQL）
3. **Supabase文档** - RLS和实战技巧

### 练习平台

1. **SQLBolt** - SQL基础练习
2. **LeetCode数据库题** - 进阶练习
3. **自己的项目** - 最好的练习！

---

## 总结

### 核心概念速查

```
表（Table）     = 存储数据的地方
字段（Column）  = 表中的一列数据
行（Row）       = 表中的一条记录
主键（PK）      = 唯一标识一行
外键（FK）      = 关联其他表
索引（Index）   = 加速查询
约束（Constraint）= 限制数据
RLS             = 行级安全
```

### 设计原则速查

```
1. 一个表只做一件事
2. 主键应该无业务意义
3. 外键必须有索引
4. 常用查询字段要加索引
5. 索引不是越多越好
6. 使用约束防止错误数据
7. 重要表要启用RLS
8. 使用参数化查询防注入
9. 密码必须加密存储
10. 重要操作要有审计日志
```

### 优化思路速查

```
慢查询？
  → 先看 EXPLAIN
  → 是否全表扫描？加索引
  → 是否JOIN慢？给外键加索引
  → 是否数据量大？考虑分区

数据重复？
  → 检查表设计，可能需要拆表

数据不一致？
  → 添加约束和外键

安全问题？
  → 启用RLS
  → 检查权限控制
```

---

**下一步行动**：

1. ✅ 阅读 `DATABASE_OPTIMIZATION_GUIDE.md` 开始优化
2. ✅ 实践练习题，加深理解
3. ✅ 用AI辅助，遇到问题就问
4. ✅ 在真实项目中应用学到的知识

**记住**：数据库设计没有完美方案，只有适合业务的方案。不断迭代，持续优化！

---

**文档版本**: v1.0
**适用对象**: AI辅助编程的开发者
**维护者**: AI Assistant
**最后更新**: 2025-11-23
