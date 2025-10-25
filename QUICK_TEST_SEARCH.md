# 审核记录搜索功能 - 快速测试指南

## 🚀 快速开始

### 1. 启动后端服务

```bash
cd c:\Log\comment-review-platform
.\comment-review-api.exe
```

或者如果已编译：
```bash
.\bin\api.exe
```

### 2. 启动前端服务

```bash
cd c:\Log\comment-review-platform\frontend
npm run dev
```

访问: http://localhost:5173

## 📝 测试步骤

### 前提条件

1. 确保数据库中有已完成的审核记录
2. 已有登录账号（管理员或审核员）

### 测试场景 1: 基本搜索（管理员）

1. **登录管理员账号**
   - 访问 http://localhost:5173/login
   - 输入管理员账号密码

2. **进入搜索页面**
   - 点击左侧菜单"审核记录搜索"
   - 或直接访问 http://localhost:5173/admin/search

3. **测试评论ID搜索**
   - 在"评论ID"输入框输入: `41`
   - 点击"搜索"按钮
   - ✅ 应该显示评论ID为41的审核记录

4. **测试审核员搜索**
   - 点击"重置"按钮清空条件
   - 在"审核员账号"输入框输入: `2669297147@qq.com`
   - 点击"搜索"按钮
   - ✅ 应该显示该审核员的所有审核记录

5. **测试时间范围搜索**
   - 点击"重置"按钮
   - 选择"审核开始时间": `2025-10-01 00:00:00`
   - 选择"审核结束时间": `2025-10-25 23:59:59`
   - 点击"搜索"按钮
   - ✅ 应该显示时间范围内的所有记录

6. **测试组合搜索**
   - 输入审核员账号: `2669297147@qq.com`
   - 选择时间范围
   - 点击"搜索"按钮
   - ✅ 应该显示同时满足两个条件的记录

### 测试场景 2: 标签搜索

1. **测试标签筛选**
   - 在"违规标签"下拉框选择一个或多个标签
   - 点击"搜索"按钮
   - ✅ 应该显示包含所选标签的记录

### 测试场景 3: 分页功能

1. **测试每页数量**
   - 在搜索结果下方，选择"每页显示数量"
   - 可选: 10, 20, 50, 100
   - ✅ 表格应该更新显示相应数量的记录

2. **测试翻页**
   - 点击"下一页"按钮
   - ✅ 应该显示下一页的记录
   - 点击"上一页"按钮
   - ✅ 应该返回上一页

3. **测试跳转页码**
   - 在页码输入框输入页码
   - 按回车
   - ✅ 应该跳转到指定页

### 测试场景 4: 查看详情

1. **点击查看详情**
   - 在任意记录行点击"查看详情"按钮
   - ✅ 应该弹出详情对话框
   - ✅ 显示完整的审核信息

2. **关闭详情**
   - 点击"关闭"按钮
   - ✅ 对话框应该关闭

### 测试场景 5: 导出功能

1. **导出搜索结果**
   - 执行任意搜索
   - 点击"导出结果"按钮
   - ✅ 应该下载一个CSV文件
   - ✅ 打开CSV文件，检查数据是否正确

### 测试场景 6: 审核员访问

1. **登录审核员账号**
   - 退出管理员账号
   - 登录审核员账号

2. **进入搜索页面**
   - 在工作台顶部点击"搜索审核记录"按钮
   - 或直接访问 http://localhost:5173/reviewer/search
   - ✅ 应该显示相同的搜索界面

3. **测试搜索**
   - 执行任意搜索
   - ✅ 功能应该与管理员相同

## 🧪 API测试

### 使用 curl 测试

1. **先获取Token**
   ```bash
   # 登录获取token
   curl -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"你的账号","password":"你的密码"}'
   ```
   
   复制返回的 `token` 值

2. **测试搜索API**
   ```bash
   # 替换 YOUR_TOKEN 为实际的token
   
   # 搜索所有记录
   curl -X GET "http://localhost:8080/api/tasks/search?page=1&page_size=20" \
     -H "Authorization: Bearer YOUR_TOKEN"
   
   # 按评论ID搜索
   curl -X GET "http://localhost:8080/api/tasks/search?comment_id=41" \
     -H "Authorization: Bearer YOUR_TOKEN"
   
   # 按审核员搜索
   curl -X GET "http://localhost:8080/api/tasks/search?reviewer_rtx=2669297147@qq.com" \
     -H "Authorization: Bearer YOUR_TOKEN"
   
   # 组合搜索
   curl -X GET "http://localhost:8080/api/tasks/search?reviewer_rtx=2669297147@qq.com&page=1&page_size=10" \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

### 使用 PowerShell 测试

```powershell
# 设置变量
$baseUrl = "http://localhost:8080"
$username = "你的账号"
$password = "你的密码"

# 登录获取token
$loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body (@{username=$username; password=$password} | ConvertTo-Json)

$token = $loginResponse.token
Write-Host "Token: $token"

# 测试搜索
$headers = @{
  "Authorization" = "Bearer $token"
}

# 搜索所有记录
$searchResponse = Invoke-RestMethod -Uri "$baseUrl/api/tasks/search?page=1&page_size=20" `
  -Method GET `
  -Headers $headers

Write-Host "找到 $($searchResponse.total) 条记录"
$searchResponse.data | Format-Table id, comment_id, username, is_approved
```

## ✅ 验收标准

### 功能性测试

- ✅ 所有搜索条件都能正常工作
- ✅ 组合搜索能正确应用所有条件
- ✅ 分页功能正常
- ✅ 详情查看功能正常
- ✅ 导出功能正常
- ✅ 管理员和审核员都能访问

### 性能测试

- ✅ 搜索响应时间 < 1秒（正常数据量）
- ✅ 分页切换流畅
- ✅ 大数据量导出不卡顿

### 用户体验

- ✅ 界面布局合理
- ✅ 操作逻辑清晰
- ✅ 错误提示友好
- ✅ 成功提示及时

## 🐛 常见问题

### 1. 搜索结果为空

**原因**: 数据库中没有符合条件的已完成记录

**解决方案**:
- 确保数据库中有 `status='completed'` 的记录
- 放宽搜索条件
- 检查时间范围是否正确

### 2. 无法访问搜索页面

**原因**: 未登录或权限不足

**解决方案**:
- 确保已登录
- 检查token是否有效
- 清除浏览器缓存重新登录

### 3. 导出文件乱码

**原因**: Excel打开CSV文件编码问题

**解决方案**:
- 使用记事本打开CSV文件
- 另存为UTF-8编码
- 或使用其他工具（如WPS）打开

### 4. 标签下拉框为空

**原因**: 标签数据未加载

**解决方案**:
- 检查网络请求
- 确保标签配置已创建
- 刷新页面重试

## 📊 测试数据准备

如果数据库中没有足够的测试数据，可以：

1. **创建审核记录**
   - 审核员领取任务
   - 提交审核结果
   - 等待任务状态变为 `completed`

2. **批量创建测试数据**
   - 使用数据库工具直接插入
   - 或编写脚本批量创建

## 🎯 测试检查清单

- [ ] 后端服务已启动
- [ ] 前端服务已启动
- [ ] 数据库有测试数据
- [ ] 评论ID搜索正常
- [ ] 审核员搜索正常
- [ ] 标签搜索正常
- [ ] 时间范围搜索正常
- [ ] 组合搜索正常
- [ ] 分页功能正常
- [ ] 详情查看正常
- [ ] 导出功能正常
- [ ] 管理员可访问
- [ ] 审核员可访问
- [ ] API接口测试通过
- [ ] 性能测试通过
- [ ] 用户体验良好

---

**测试完成后，如有问题请参考：**
- [SEARCH_FEATURE_COMPLETE.md](./SEARCH_FEATURE_COMPLETE.md) - 完整功能文档
- [API_SEARCH_GUIDE.md](./API_SEARCH_GUIDE.md) - API详细文档

