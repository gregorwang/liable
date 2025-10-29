@echo off
echo ========================================
echo  评论审核平台 - 停止服务脚本
echo ========================================
echo.

echo [1/2] 停止 Docker 服务...
docker-compose down
echo ✅ Docker 服务已停止

echo [2/2] 清理环境文件...
if exist .env del .env
echo ✅ 环境文件已清理

echo.
echo ========================================
echo  所有服务已停止！
echo ========================================
echo.
echo 注意：前端和后端服务需要手动停止（Ctrl+C）
echo.
pause



