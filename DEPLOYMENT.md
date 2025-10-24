# 部署指南

## 环境要求

1. **Go 1.21+**
   - 下载地址：https://go.dev/dl/
   - 安装后验证：`go version`

2. **PostgreSQL (Supabase)**
   - 已配置 ✅

3. **Redis (Upstash with TLS)**
   - 已配置 ✅

## 快速开始

### 1. 安装 Go

如果您还没有安装 Go，请访问 https://go.dev/dl/ 下载并安装。

Windows 用户可以下载 `.msi` 安装包直接安装。

### 2. 下载依赖

```bash
cd comment-review-platform
go mod download
```

### 3. 验证环境变量

确认 `.env` 文件已正确配置（已自动生成）。

### 4. 启动服务

```bash
go run cmd/api/main.go
```

服务将在 `http://localhost:8080` 启动。

### 5. 验证服务

访问健康检查接口：
```bash
curl http://localhost:8080/health
```

应返回：
```json
{"status":"healthy"}
```

## 构建生产版本

### 编译二进制文件

```bash
go build -o comment-review-api.exe cmd/api/main.go
```

### 运行

```bash
./comment-review-api.exe
```

## 数据库状态

- ✅ 所有表已创建
- ✅ 默认管理员账号已创建：`admin` / `admin123`
- ✅ 6 个违规标签已配置
- ✅ 5323 个待审核任务已创建

## 目录结构

```
comment-review-platform/
├── cmd/api/              # 应用入口
│   └── main.go
├── internal/             # 内部包
│   ├── config/          # 配置管理
│   ├── models/          # 数据模型
│   ├── middleware/      # 中间件
│   ├── handlers/        # HTTP 处理器
│   ├── services/        # 业务逻辑
│   └── repository/      # 数据访问层
├── pkg/                  # 可复用包
│   ├── database/        # PostgreSQL 连接
│   ├── redis/           # Redis 连接
│   └── jwt/             # JWT 工具
├── migrations/          # 数据库迁移
├── .env                 # 环境变量（已配置）
├── go.mod               # Go 模块定义
└── README.md            # 项目说明
```

## 后台任务

应用启动后会自动运行以下后台任务：

- **任务超时释放**：每 5 分钟检查并释放超过 30 分钟未完成的任务

## 故障排除

### 问题：无法连接数据库

1. 检查 `.env` 中的 `DATABASE_URL` 是否正确
2. 确认 Supabase 数据库是否在线
3. 检查网络连接

### 问题：无法连接 Redis

1. 检查 `.env` 中的 Redis 配置
2. 确认 `REDIS_USE_TLS=true`
3. 验证 Upstash Redis 服务是否正常

### 问题：端口被占用

修改 `.env` 中的 `PORT` 配置：
```
PORT=8081
```

## 下一步

1. 使用 Postman 或 curl 测试 API 接口
2. 注册审核员账号并等待管理员审批
3. 开始前端开发（Vue 3）

参考 `API_TESTING.md` 了解如何测试所有接口。

