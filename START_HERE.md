# 🚀 快速开始指南

欢迎使用评论审核平台后端！这是一个完整的后端系统，现在只需几步即可启动。

## ⚡ 快速启动（3 步）

### 步骤 1：安装 Go

如果您还没有安装 Go：

1. 访问：https://go.dev/dl/
2. 下载适合您系统的安装包（Windows 用户下载 `.msi` 文件）
3. 运行安装程序
4. 打开新的命令行窗口，验证安装：
   ```bash
   go version
   ```

### 步骤 2：下载依赖

```bash
cd comment-review-platform
go mod download
```

这会下载所有必需的 Go 包（约 30 秒）。

### 步骤 3：启动服务

```bash
go run cmd/api/main.go
```

看到以下输出表示成功：
```
✅ Configuration loaded
✅ PostgreSQL connected successfully
✅ Redis connected successfully
✅ Background task release worker started
🚀 Server starting on port 8080
```

## 🎯 验证服务

打开新的命令行窗口，测试健康检查：

```bash
curl http://localhost:8080/health
```

应该返回：
```json
{"status":"healthy"}
```

## 🔐 默认管理员账号

- 用户名：`admin`
- 密码：`admin123`

## 📚 下一步

1. **测试登录**：
   ```bash
   curl -X POST http://localhost:8080/api/auth/login ^
     -H "Content-Type: application/json" ^
     -d "{\"username\":\"admin\",\"password\":\"admin123\"}"
   ```

2. **查看 API 文档**：打开 `API_TESTING.md` 查看所有接口的详细使用方法

3. **注册审核员**：使用注册接口创建审核员账号，然后用管理员账号审批

## 📖 完整文档

- **API_TESTING.md** - 完整的 API 测试指南（包含所有接口示例）
- **DEPLOYMENT.md** - 详细的部署和配置说明
- **PROJECT_SUMMARY.md** - 项目架构和功能总览
- **README.md** - 项目介绍和技术栈

## 💡 常见问题

### Q: 如何停止服务？
A: 在运行服务的命令行窗口按 `Ctrl + C`

### Q: 端口 8080 被占用怎么办？
A: 编辑 `.env` 文件，修改 `PORT=8081`（或其他可用端口）

### Q: 忘记管理员密码？
A: 默认密码是 `admin123`，如需重置，可以直接在数据库中更新

### Q: 如何添加新的审核员？
A: 
1. 使用 `/api/auth/register` 接口注册
2. 管理员使用 `/api/admin/users/:id/approve` 接口审批
3. 审核员登录后即可领取任务

## 🎨 前端开发

后端已完全就绪，可以开始前端开发（Vue 3 + Element Plus）。

所有 API 接口已实现并测试，参考 `API_TESTING.md` 进行对接。

## 🐛 遇到问题？

1. 检查 Go 是否正确安装：`go version`
2. 检查环境变量配置（`.env` 文件）
3. 查看控制台错误信息
4. 检查数据库和 Redis 连接状态

## 📊 当前数据

- ✅ 数据库表已创建
- ✅ 1 个管理员账号
- ✅ 6 个违规标签
- ✅ 5323 个待审核任务

---

**祝您使用愉快！** 🎉

有任何问题请参考其他文档或检查代码注释。

