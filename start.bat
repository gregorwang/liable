@echo off
chcp 65001 >nul
echo ========================================
echo 评论审核平台 - 后端服务启动脚本
echo ========================================
echo.

REM 检查 Go 是否安装
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [错误] Go 未安装或未添加到 PATH
    echo.
    echo 请先安装 Go:
    echo https://go.dev/dl/
    echo.
    pause
    exit /b 1
)

echo [1/3] 检查 Go 版本...
go version
echo.

echo [2/3] 下载依赖...
go mod download
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 依赖下载失败
    pause
    exit /b 1
)
echo ✓ 依赖下载完成
echo.

echo [3/3] 启动服务...
echo ========================================
echo 服务将在 http://localhost:8080 启动
echo 按 Ctrl+C 停止服务
echo ========================================
echo.
go run cmd/api/main.go

pause

