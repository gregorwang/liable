@echo off
echo ========================================
echo  评论审核平台 - 启动脚本
echo ========================================
echo.

echo [0/3] 检查并启动 Redis (Docker)...
docker-compose up -d redis
if %errorlevel% neq 0 (
    echo ❌ Redis 启动失败，请确保 Docker Desktop 正在运行
    pause
    exit /b 1
)
echo ✅ Redis 已启动
timeout /t 2 /nobreak >nul

echo [1/3] 启动后端服务 (Go)...
start "Backend Server" cmd /k "go run cmd/api/main.go"
timeout /t 3 /nobreak >nul

echo [2/3] 启动前端服务 (Vue)...
start "Frontend Server" cmd /k "cd frontend && npm run dev"

echo.
echo [3/3] 所有服务启动完成！
echo.
echo ========================================
echo  服务信息
echo ========================================
echo  Redis:    localhost:6379 (Docker)
echo  后端地址: http://localhost:8080
echo  前端地址: http://localhost:3000
echo.
echo  管理员账号: admin / admin123
echo.
echo ========================================
echo  使用说明
echo ========================================
echo  - 关闭此窗口不会停止服务
echo  - 请在各自的窗口中按 Ctrl+C 停止服务
echo  - 停止 Redis: docker-compose down
echo ========================================