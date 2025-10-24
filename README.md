# 评论审核平台

基于 Go + Gin 框架的后端和 Vue 3 + TypeScript 的前端构建的完整评论审核系统。

## 技术栈

### 后端
- **框架**: Gin Web Framework
- **语言**: Go 1.21+
- **数据库**: Supabase PostgreSQL
- **缓存**: Upstash Redis (with TLS)
- **认证**: JWT
- **密码加密**: bcrypt

### 前端
- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **构建工具**: Vite
- **UI 库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router
- **HTTP 客户端**: Axios

## 功能特性

### 用户角色
- **审核员 (reviewer)**: 领取任务、审核评论、提交结果
- **管理员 (admin)**: 用户管理、标签配置、统计查看

### 核心功能
1. **任务管理**: 基于 Redis 的分布式任务队列，支持任务锁定和超时释放
2. **审核流程**: 一次领取 20 条评论，标记通过/不通过，添加违规标签和原因
3. **统计分析**: 每小时/每日标注量、通过率、违规类型分布、审核员绩效排行
4. **用户管理**: 注册审批机制，管理员审批后才能使用

## 项目结构

```
comment-review-platform/
├── cmd/api/              # 后端应用入口
├── internal/             # 后端内部包
│   ├── config/          # 配置管理
│   ├── models/          # 数据模型
│   ├── middleware/      # 中间件（认证、权限）
│   ├── handlers/        # HTTP 处理器
│   ├── services/        # 业务逻辑
│   └── repository/      # 数据访问层
├── pkg/                  # 可复用包
│   ├── database/        # PostgreSQL 连接
│   ├── redis/           # Redis 连接
│   └── jwt/             # JWT 工具
├── migrations/          # 数据库迁移
└── frontend/            # 前端项目
    ├── src/
    │   ├── api/         # API 接口封装
    │   ├── router/      # 路由配置
    │   ├── stores/      # Pinia 状态管理
    │   ├── types/       # TypeScript 类型
    │   ├── utils/       # 工具函数
    │   └── views/       # 页面组件
    ├── vite.config.ts   # Vite 配置
    └── package.json
```

## 快速开始

### 方式一：一键启动（推荐）

Windows 用户可以直接运行：

```bash
start-all.bat
```

这将自动启动后端和前端服务。

### 方式二：手动启动

#### 1. 启动后端

```bash
# 安装 Go 依赖
go mod download

# 启动后端服务
go run cmd/api/main.go
```

后端服务将在 `http://localhost:8080` 启动。

#### 2. 启动前端

```bash
# 进入前端目录
cd frontend

# 安装依赖（首次运行）
npm install

# 启动前端服务
npm run dev
```

前端服务将在 `http://localhost:3000` 启动。

#### 3. 访问系统

打开浏览器访问：`http://localhost:3000`

## 前置要求

### 后端
- Go 1.21+
- PostgreSQL (Supabase) - 已配置 ✅
- Redis (Upstash) - 已配置 ✅

### 前端
- Node.js >= 20.15.0
- npm >= 10.7.0

## API 接口

### 认证接口

- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `GET /api/auth/profile` - 获取当前用户信息

### 审核员接口

- `POST /api/tasks/claim` - 领取任务
- `GET /api/tasks/my` - 查看我的待审核任务
- `POST /api/tasks/submit` - 提交审核结果
- `GET /api/tags` - 获取违规标签列表

### 管理员接口

- `GET /api/admin/users` - 获取用户列表
- `PUT /api/admin/users/:id/approve` - 审批用户
- `GET /api/admin/stats/overview` - 总体统计
- `GET /api/admin/stats/hourly` - 每小时标注量
- `GET /api/admin/stats/tags` - 违规类型分布
- `GET /api/admin/stats/reviewers` - 审核员绩效排行
- `POST /api/admin/tags` - 创建违规标签
- `PUT /api/admin/tags/:id` - 更新违规标签
- `DELETE /api/admin/tags/:id` - 删除违规标签

## 默认账号

### 管理员
- 用户名: `admin`
- 密码: `admin123`
- 登录后访问管理后台

### 审核员
1. 访问注册页面创建账号
2. 等待管理员审批通过
3. 登录后可以领取和审核任务

## Redis 键设计

- `task:pending` - 待领取任务队列
- `task:claimed:{user_id}` - 用户已领取任务集合 (TTL: 30分钟)
- `task:lock:{task_id}` - 任务分布式锁 (TTL: 30分钟)
- `stats:hourly:{date}:{hour}` - 每小时统计 (TTL: 7天)
- `stats:daily:{date}` - 每日统计 (TTL: 30天)

## 文档

- **[START_HERE.md](./START_HERE.md)** - 快速开始指南
- **[FRONTEND_GUIDE.md](./FRONTEND_GUIDE.md)** - 前端开发指南（详细）
- **[API_TESTING.md](./API_TESTING.md)** - API 接口测试文档
- **[PROJECT_SUMMARY.md](./PROJECT_SUMMARY.md)** - 项目概览
- **[DEPLOYMENT.md](./DEPLOYMENT.md)** - 部署指南

## 功能截图

### 登录页面
- 美观的渐变背景
- 表单验证
- 跳转注册页面

### 审核员工作台
- 任务领取（一次 20 条）
- 任务列表展示
- 单个/批量审核提交
- 违规标签选择
- 实时统计

### 管理员控制台
- 数据总览（统计卡片）
- 用户管理（审批）
- 统计分析（图表）
- 标签管理（CRUD）

## 系统状态

✅ **后端完成** - Go + Gin + PostgreSQL + Redis  
✅ **前端完成** - Vue 3 + TypeScript + Element Plus  
✅ **完整集成** - 前后端已完全对接  
✅ **可以使用** - 立即开始审核工作

## License

MIT

