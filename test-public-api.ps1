#!/usr/bin/env pwsh
# 测试公开队列API脚本（PowerShell版本）

Write-Host "🔍 测试1: 获取队列列表（无需认证）" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
$response1 = curl.exe -s "http://localhost:8080/api/queues"
$response1 | ConvertFrom-Json | ConvertTo-Json | Write-Host
Write-Host ""
Write-Host ""

Write-Host "🔍 测试2: 获取队列列表（带分页）" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
$response2 = curl.exe -s "http://localhost:8080/api/queues?page=1&page_size=5"
$response2 | ConvertFrom-Json | ConvertTo-Json | Write-Host
Write-Host ""
Write-Host ""

Write-Host "🔍 测试3: 搜索队列" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
$response3 = curl.exe -s "http://localhost:8080/api/queues?search=色情"
$response3 | ConvertFrom-Json | ConvertTo-Json | Write-Host
Write-Host ""
Write-Host ""

Write-Host "🔍 测试4: 获取单个队列详情（ID=1）" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
$response4 = curl.exe -s "http://localhost:8080/api/queues/1"
$response4 | ConvertFrom-Json | ConvertTo-Json | Write-Host
Write-Host ""
Write-Host ""

Write-Host "✅ 所有测试完成！" -ForegroundColor Green
Write-Host "✅ 如果以上都返回正确的JSON数据，说明公开API工作正常" -ForegroundColor Green
