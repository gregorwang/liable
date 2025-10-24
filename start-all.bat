@echo off
echo ========================================
echo  评论审核平台 - 启动脚本
echo ========================================
echo.

echo [1/2] 启动后端服务 (Go)...
start "Backend Server" cmd /k "go run cmd/api/main.go"
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

