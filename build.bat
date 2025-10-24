@echo off
chcp 65001 >nul
echo ========================================
echo 评论审核平台 - 构建生产版本
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

echo [1/2] 下载依赖...
go mod download
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 依赖下载失败
    pause
    exit /b 1
)
echo ✓ 依赖下载完成
echo.

echo [2/2] 编译程序...
go build -o comment-review-api.exe cmd/api/main.go
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 编译失败
    pause
    exit /b 1
)

echo.
echo ========================================
echo ✓ 构建成功！
echo ========================================
echo.
echo 可执行文件: comment-review-api.exe
echo.
echo 运行方式:
echo   1. 直接双击 comment-review-api.exe
echo   2. 或在命令行运行: .\comment-review-api.exe
echo.
pause

