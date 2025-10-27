# 测试公开队列API脚本

echo "🔍 测试1: 获取队列列表（无需认证）"
echo "======================================"
curl -s http://localhost:8080/api/queues | jq .
echo ""
echo ""

echo "🔍 测试2: 获取队列列表（带分页）"
echo "======================================"
curl -s "http://localhost:8080/api/queues?page=1&page_size=5" | jq .
echo ""
echo ""

echo "🔍 测试3: 搜索队列"
echo "======================================"
curl -s "http://localhost:8080/api/queues?search=色情" | jq .
echo ""
echo ""

echo "🔍 测试4: 获取单个队列详情"
echo "======================================"
curl -s "http://localhost:8080/api/queues/1" | jq .
echo ""
echo ""

echo "✅ 所有测试完成！"
echo "✅ 如果以上都返回正确的JSON数据，说明公开API工作正常"
