@echo off
echo ========================================
echo  评论审核平台 - 启动脚本
echo ========================================
echo.

echo [1/2] 启动后端服务 (Go)...
start "Backend Server" cmd /k ""go run cmd/api/main.go
timeout /t 3 /nobreak >nul

echo [2/2] 启动前端服务 (Vue)...
start "Frontend Server" cmd /k "cd frontend && npm run dev"

echo.
echo ========================================
echo  服务启动中...
echo ========================================
echo.
echo  后端地址: http://localhost:8080
echo  前端地址: http://localhost:3000
echo.
echo  管理员账号: admin / admin123
echo.
echo  关闭此窗口不会停止服务
echo  请在各自的窗口中按 Ctrl+C 停止
echo ========================================

// 管理员路由组
admin.Use(middleware.AuthMiddleware(), middleware.RequireAdmin())

// 用户管理
GET /api/admin/users                    // 获取待审批用户列表
PUT /api/admin/users/:id/approve        // 审批用户

// 统计数据
GET /api/admin/stats/overview           // 概览统计
GET /api/admin/stats/hourly             // 每小时统计
GET /api/admin/stats/tags               // 标签统计
GET /api/admin/stats/reviewers          // 审核员绩效

// 标签管理
GET    /api/admin/tags                  // 获取所有标签
POST   /api/admin/tags                  // 创建标签
PUT    /api/admin/tags/:id              // 更新标签
DELETE /api/admin/tags/:id              // 删除标签

// 审核规则管理
POST   /api/admin/moderation-rules      // 创建规则
PUT    /api/admin/moderation-rules/:id  // 更新规则
DELETE /api/admin/moderation-rules/:id  // 删除规则

// 任务队列管理
POST   /api/admin/task-queues           // 创建队列
GET    /api/admin/task-queues           // 列表查询
GET    /api/admin/task-queues/:id       // 获取详情
PUT    /api/admin/task-queues/:id       // 更新队列
DELETE /api/admin/task-queues/:id       // 删除队列
GET    /api/admin/task-queues-all       // 获取所有队列

// 通知管理
POST /api/admin/notifications           // 创建通知

// 视频管理（如果启用）
POST /api/admin/videos/import           // 导入视频
GET  /api/admin/videos                  // 视频列表
GET  /api/admin/videos/:id              // 视频详情