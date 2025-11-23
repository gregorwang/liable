@echo off
echo ========================================
echo  启动本地 Redis (Docker)
echo ========================================
echo.

echo [1/2] 停止现有 Redis 容器...
docker stop admiring_kapitsa 2>nul
docker rm admiring_kapitsa 2>nul

echo [2/2] 启动 Redis 并映射端口到 localhost:6379...
docker run -d ^
  --name redis-local ^
  -p 6379:6379 ^
  redis:latest ^
  redis-server --appendonly yes

if %errorlevel% neq 0 (
    echo ❌ Redis 启动失败
    pause
    exit /b 1
)

echo.
echo ✅ Redis 已启动并映射到 localhost:6379
echo.
echo 测试连接：
docker exec redis-local redis-cli ping

echo.
pause

