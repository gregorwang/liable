# Redis 本地配置指南

## 🎯 推荐镜像选择

根据 Docker Hub 搜索结果，推荐以下两个选项：

### 选项 1：官方 Redis 镜像（推荐，最简洁）
- **镜像名**: `redis:7-alpine`
- **特点**: 官方镜像，体积小，稳定可靠
- **适用**: 本地开发和测试
- **使用**: `docker-compose.yml`（已创建）

### 选项 2：Redis Stack（带 GUI 工具）
- **镜像名**: `redis/redis-stack:latest`
- **特点**: 包含 RedisInsight GUI，方便调试和查看数据
- **适用**: 需要可视化管理界面
- **使用**: `docker-compose.redis-stack.yml`（已创建）
- **GUI 访问**: http://localhost:8001

## 🚀 快速开始

### 1. 启动 Redis（使用官方镜像）

```bash
docker-compose up -d redis
```

### 2. 启动 Redis Stack（带 GUI）

如果需要 GUI 工具，使用：

```bash
docker-compose -f docker-compose.redis-stack.yml up -d redis
```

然后访问：http://localhost:8001

### 3. 检查 Redis 是否运行

```bash
docker ps | findstr redis
```

或者测试连接：

```bash
docker exec comment-review-redis redis-cli ping
```

应该返回：`PONG`

## ⚙️ 配置环境变量

### 本地 Redis（最简单）

对于本地 Redis，**完全可以不配置**！代码已经有默认值：

- `REDIS_ADDR` → 默认 `localhost:6379`
- `REDIS_PASSWORD` → 默认 `""`（空）
- `REDIS_DB` → 默认 `0`
- `REDIS_USE_TLS` → 默认 `false`

**或者只配置一行（推荐）**：

```env
REDIS_ADDR=localhost:6379
```

就这一行就够了！其他都用默认值。

### 如果你之前配置了远程 Redis

如果 `.env` 中有这些配置，**删除或注释掉**它们：

```env
# 删除或注释这些行
# REDIS_PASSWORD=xxx
# REDIS_USE_TLS=true
# REDIS_TLS_SKIP_VERIFY=true
```

只保留（可选，因为默认值就是它）：

```env
REDIS_ADDR=localhost:6379
```

## 🔧 常用命令

### 启动 Redis
```bash
docker-compose up -d redis
```

### 停止 Redis
```bash
docker-compose down
```

### 查看 Redis 日志
```bash
docker-compose logs -f redis
```

### 进入 Redis CLI
```bash
docker exec -it comment-review-redis redis-cli
```

### 清空所有数据（谨慎使用）
```bash
docker exec comment-review-redis redis-cli FLUSHALL
```

## 🐛 故障排除

### 问题 1: Docker Desktop 未运行
**错误**: `Cannot connect to the Docker daemon`

**解决**: 启动 Docker Desktop 应用程序

### 问题 2: 端口 6379 已被占用
**错误**: `port is already allocated`

**解决**: 
1. 检查是否有其他 Redis 实例在运行
2. 修改 `docker-compose.yml` 中的端口映射，例如改为 `6380:6379`
3. 同时更新 `.env` 中的 `REDIS_ADDR=localhost:6380`

### 问题 3: 连接超时
**原因**: 可能是配置中 `REDIS_USE_TLS=true`（用于远程 Redis）

**解决**: 在 `.env` 中设置 `REDIS_USE_TLS=false`

### 问题 4: 权限错误
**错误**: `permission denied`

**解决**: 确保 Docker Desktop 有足够权限，或在管理员模式下运行命令

## 📝 数据持久化

Redis 数据会自动保存到 Docker volume `redis-data` 中，即使容器重启数据也不会丢失。

### 备份数据
```bash
docker exec comment-review-redis redis-cli BGSAVE
docker cp comment-review-redis:/data/dump.rdb ./redis-backup.rdb
```

### 恢复数据
```bash
docker cp ./redis-backup.rdb comment-review-redis:/data/dump.rdb
docker-compose restart redis
```

## 🔄 从远程 Redis 迁移到本地

如果你的项目之前使用的是 Upstash 或其他远程 Redis：

1. **停止应用**
2. **导出远程 Redis 数据**（如果重要）
3. **启动本地 Redis**
4. **更新 `.env` 配置**：
   ```env
   REDIS_ADDR=localhost:6379
   REDIS_PASSWORD=
   REDIS_USE_TLS=false
   ```
5. **重新启动应用**

## ✅ 验证配置

启动应用后，查看后端日志，应该看到：

```
✅ Redis connected successfully
```

如果看到连接错误，请检查：
1. Redis 容器是否正在运行
2. `.env` 配置是否正确
3. 端口是否被占用

