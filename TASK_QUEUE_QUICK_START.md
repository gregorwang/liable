# 任务队列管理 - 快速开始指南

## 目录

1. [功能概述](#功能概述)
2. [系统要求](#系统要求)
3. [快速开始](#快速开始)
4. [核心概念](#核心概念)
5. [常见操作](#常见操作)
6. [故障排除](#故障排除)

---

## 功能概述

任务队列管理是评论审核系统的核心功能，允许管理员：

✅ **创建队列** - 为不同类型的审核内容创建独立队列  
✅ **管理队列** - 编辑队列信息、优先级和状态  
✅ **追踪进度** - 实时查看每个队列的审核进度  
✅ **搜索过滤** - 快速查找特定队列  
✅ **优先级管理** - 按优先级分配资源给评审员  

---

## 系统要求

### 后端
- Go 1.18+
- PostgreSQL 13+
- Redis 6.0+

### 前端
- Vue 3.3+
- Element Plus 2.0+
- Node.js 16+

### 已安装的依赖

后端：
```go
- github.com/gin-gonic/gin
- github.com/lib/pq
- github.com/golang-jwt/jwt/v4
```

前端：
```
- vue@3.3.4
- element-plus@2.4.0
- axios
```

---

## 快速开始

### 步骤 1: 启动后端服务

```bash
# 进入项目根目录
cd C:\Log\comment-review-platform

# 编译后端代码
go build -o comment-review-api.exe ./cmd/api/main.go

# 设置环境变量（如果未设置）
# 参考 setup-env.example.txt

# 运行服务
.\comment-review-api.exe
# 服务将监听 http://localhost:8080
```

### 步骤 2: 启动前端开发服务

```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 启动开发服务
npm run dev
# 前端将运行在 http://localhost:5173
```

### 步骤 3: 访问管理后台

1. 打开浏览器访问 `http://localhost:5173`
2. 使用管理员账户登录
   - 默认用户名: `admin`
   - 默认密码: `admin123`
3. 在左侧菜单找到 "队列配置" 选项
4. 开始管理任务队列

---

## 核心概念

### 队列（Queue）

队列是评论审核任务的逻辑分组。每个队列代表一类审核任务（如色情内容、广告、垃圾等）。

**队列属性**:
- **队列名称** (Queue Name): 唯一标识，如 "色情内容审核"
- **优先级** (Priority): 0-1000，数值越大越优先处理
- **总任务数** (Total Tasks): 该队列中的总任务数
- **已审核数** (Completed Tasks): 已完成审核的任务数
- **待审核数** (Pending Tasks): 尚待审核的任务数 = 总任务数 - 已审核数

### 队列生命周期

```
创建 → 活跃 → 进行中 → 完成 → 归档/删除
```

### 优先级说明

| 优先级范围 | 含义 | 用例 |
|-----------|------|------|
| 0-20 | 低优先级 | 常规审核 |
| 21-50 | 中优先级 | 重要内容 |
| 51-80 | 高优先级 | 敏感话题 |
| 81-100 | 紧急 | 投诉/申诉 |

---

## 常见操作

### 操作 1: 创建新队列

**UI 操作**:
1. 进入"队列配置"页面
2. 点击 "新建队列" 按钮
3. 填写表单：
   - 队列名称：输入队列标识名
   - 描述：详细说明队列的用途
   - 优先级：设置处理顺序（0-1000）
   - 总任务数：输入该队列的总任务数
   - 已审核数：初始通常为0
4. 点击 "创建" 确认

**API 操作**:
```bash
curl -X POST http://localhost:8080/api/admin/task-queues \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "queue_name": "色情内容审核",
    "description": "审核色情和低俗内容",
    "priority": 80,
    "total_tasks": 500
  }'
```

### 操作 2: 查看队列列表

**UI 操作**:
1. 进入"队列配置"页面
2. 查看表格中的所有队列
3. 可按以下方式过滤：
   - 搜索框：按队列名称搜索
   - 状态过滤：活跃/已禁用
   - 分页：每页10条记录

**API 操作**:
```bash
# 获取第一页，每页10条
curl -X GET "http://localhost:8080/api/admin/task-queues?page=1&page_size=10" \
  -H "Authorization: Bearer <your_token>"

# 搜索并过滤
curl -X GET "http://localhost:8080/api/admin/task-queues?search=色情&is_active=true" \
  -H "Authorization: Bearer <your_token>"
```

### 操作 3: 更新队列进度

当审核人员完成了一些任务后，需要更新已审核数。

**UI 操作**:
1. 在队列列表中找到目标队列
2. 点击"编辑"按钮
3. 修改"已审核数"字段
4. 观察"待审核数"自动更新
5. 点击"更新"保存

**API 操作**:
```bash
# 更新队列1的已审核数为100
curl -X PUT http://localhost:8080/api/admin/task-queues/1 \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "completed_tasks": 100
  }'
```

### 操作 4: 修改队列优先级

当需要调整审核顺序时。

**UI 操作**:
1. 在队列列表中找到目标队列
2. 点击"编辑"按钮
3. 修改"优先级"字段
4. 点击"更新"保存

**说明**: 系统会按优先级降序处理队列。

### 操作 5: 禁用或删除队列

**UI 操作（禁用）**:
1. 在队列列表中找到目标队列
2. 点击"编辑"按钮
3. 取消"状态"开关（从活跃变为禁用）
4. 点击"更新"保存

**UI 操作（删除）**:
1. 在队列列表中找到目标队列
2. 点击"删除"按钮
3. 在确认对话框中点击"删除"

**API 操作**:
```bash
# 删除队列1
curl -X DELETE http://localhost:8080/api/admin/task-queues/1 \
  -H "Authorization: Bearer <your_token>"
```

---

## 数据统计示例

### 场景：监控色情内容审核队列

假设你创建了以下队列：

```
队列名称: "色情内容审核"
总任务数: 1000
已审核数: 250
待审核数: 750
优先级: 85
```

**进度计算**:
- 审核进度: 250 / 1000 = 25%
- 预计完成时间: 按审核速度估算

**实时监控**:
```bash
# 每5分钟查询一次进度
watch -n 300 'curl -s -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/admin/task-queues/1 | jq .'
```

---

## 前端界面功能

### QueueManage 组件特性

**表格列**:
| 列名 | 说明 |
|------|------|
| 队列名称 | 队列的唯一标识 |
| 描述 | 队列的详细说明 |
| 优先级 | 颜色编码：绿→蓝→橙→红 |
| 任务统计 | 显示总数、已审、待审 |
| 进度 | 进度条可视化 |
| 状态 | 活跃/禁用标签 |
| 创建时间 | 队列创建时间 |
| 操作 | 编辑/删除按钮 |

**搜索和过滤**:
- 按队列名称模糊搜索
- 按活跃状态过滤
- 分页浏览

**弹出对话框**:
- 新建/编辑模式自动检测
- 完整的表单验证
- 实时计算待审核数 = 总数 - 已审数

---

## 故障排除

### 问题 1: 无法连接到后端 API

**症状**: 前端报错 "Cannot connect to server"

**解决方案**:
```bash
# 1. 检查后端是否运行
netstat -ano | findstr :8080

# 2. 检查后端日志
# 查看控制台输出是否有错误

# 3. 检查数据库连接
# 确保 PostgreSQL 和 Redis 都在运行

# 4. 重启后端服务
cd C:\Log\comment-review-platform
go build -o comment-review-api.exe ./cmd/api/main.go
.\comment-review-api.exe
```

### 问题 2: 队列创建失败

**症状**: 创建队列时报错 "completed_tasks cannot be greater than total_tasks"

**解决方案**:
- 确保已审核数 ≤ 总任务数
- 例如：总数200，已审数不超过200

### 问题 3: 队列列表为空

**症状**: 页面显示 "共0个队列"

**可能原因**:
- 还未创建任何队列 → 点击"新建队列"创建
- 所有队列都被禁用 → 调整过滤条件
- 权限问题 → 确保使用 admin 账户登录

### 问题 4: Token 过期

**症状**: API 请求返回 401 Unauthorized

**解决方案**:
```bash
# 重新登录获取新 token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

### 问题 5: 数据库错误

**症状**: 创建队列时报错 "database error"

**解决方案**:
```bash
# 1. 检查 PostgreSQL 是否运行
# 2. 检查表是否存在
psql -U postgres -d comment_review -c "\dt"

# 3. 确保 task_queue 表已创建
# 如果没有，请运行迁移
```

---

## 性能调优建议

### 数据库优化

```sql
-- 已建立的索引
CREATE INDEX idx_task_queue_active ON task_queue(is_active);
CREATE INDEX idx_task_queue_priority ON task_queue(priority);
CREATE INDEX idx_task_queue_created_at ON task_queue(created_at);
```

### 前端优化

- 列表默认分页每页10条
- 使用虚拟滚动处理大量数据
- 实现队列数据缓存

### 后端优化

- 查询响应 < 100ms
- 列表查询 < 500ms
- 支持并发访问

---

## API 端点速查表

| 操作 | 方法 | 端点 |
|------|------|------|
| 创建队列 | POST | /admin/task-queues |
| 获取列表 | GET | /admin/task-queues |
| 获取详情 | GET | /admin/task-queues/:id |
| 更新队列 | PUT | /admin/task-queues/:id |
| 删除队列 | DELETE | /admin/task-queues/:id |
| 活跃队列 | GET | /admin/task-queues-all |

详细 API 文档见 `TASK_QUEUE_API.md`

---

## 获取帮助

- 📖 完整 API 文档: 查看 `TASK_QUEUE_API.md`
- 🐛 报告 Bug: 提交 Issue
- 💡 功能建议: 发起讨论
- 📧 技术支持: 联系开发团队

---

## 更新日志

### v1.0.0 (2025-01-15)
- ✨ 初始发布
- ✨ 完整的 CRUD 操作
- ✨ 队列搜索和过滤
- ✨ 优先级管理
- ✨ 进度追踪

---

**祝您使用愉快！** 🎉
