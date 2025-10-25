# API 更新：退单功能和自定义领取数量

## 📋 更新概述

本次更新增加了以下功能：
1. ✅ 审核员可以自定义领取任务数量（1-50单）
2. ✅ 审核员可以退单（将任务退回任务池）
3. ✅ 前端界面支持数量选择和批量退单

---

## 🔄 API 变更

### 1. 领取任务接口（已修改）

**接口：** `POST /api/tasks/claim`

**请求体（新增）：**
```json
{
  "count": 20  // 必填，范围：1-50
}
```

**响应：**
```json
{
  "tasks": [...],
  "count": 20
}
```

**示例：**
```bash
curl -X POST http://localhost:8080/api/tasks/claim \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"count": 10}'
```

### 2. 退单接口（新增）

**接口：** `POST /api/tasks/return`

**请求体：**
```json
{
  "task_ids": [1, 2, 3, 4, 5]  // 必填，最多50个
}
```

**响应：**
```json
{
  "message": "Tasks returned successfully",
  "count": 5
}
```

**示例：**
```bash
curl -X POST http://localhost:8080/api/tasks/return \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"task_ids": [1, 2, 3]}'
```

**错误响应：**
- `400 Bad Request`: 任务ID不属于当前用户或任务不在进行中状态
- `400 Bad Request`: 任务数量不在1-50范围内

---

## 🎨 前端界面变更

### 审核员 Dashboard 新功能

#### 1. 自定义领取数量
- 新增数字输入框，范围 1-50
- 默认值：20
- 点击"领取新任务"按钮前可调整数量

#### 2. 退单功能
- 每个任务卡片新增复选框
- 选择要退单的任务
- 点击"退单"按钮批量退回任务
- 退单前会显示确认对话框

#### 3. 界面布局
```
[数量选择 ▼ 20] [领取新任务] [批量提交(N条)] [退单(N条)] [刷新任务列表]
```

---

## 📝 使用场景

### 场景1：灵活调整工作量
```
审核员A今天时间充足：
1. 选择领取 50 单
2. 点击"领取新任务"
3. 开始审核

审核员B今天时间较少：
1. 选择领取 10 单
2. 点击"领取新任务"
3. 开始审核
```

### 场景2：退回不想审核的任务
```
审核员领取了 20 单，发现其中 5 单内容过于复杂：
1. 勾选这 5 个任务
2. 点击"退单"按钮
3. 确认退单
4. 这 5 单会退回任务池，其他审核员可以领取
```

### 场景3：需要暂停工作
```
审核员中途有事需要离开：
1. 勾选所有未完成的任务
2. 点击"退单"按钮
3. 确认退单
4. 所有任务退回，方便其他审核员继续处理
```

---

## 🔒 业务规则

### 领取任务规则
1. ✅ 必须指定领取数量（1-50）
2. ✅ 只有当前没有未完成任务时才能领取新任务
3. ✅ 如果有未完成任务，系统会提示："你还有 N 条未完成任务，请先完成或退单"
4. ✅ 实际领取数量可能少于请求数量（取决于任务池剩余数量）

### 退单规则
1. ✅ 只能退回属于自己且状态为"进行中"的任务
2. ✅ 单次最多退回 50 单
3. ✅ 退单后任务立即回到任务池（status: pending）
4. ✅ 退单后 Redis 锁会立即清除
5. ✅ 退单不影响统计数据

### 超时规则（保持不变）
- ⏰ 任务超时时间：30 分钟（可配置）
- ⏰ 超时后自动退回任务池
- ⏰ 后台定时任务每 5 分钟检查一次

---

## 🧪 测试步骤

### 测试1：自定义领取数量
```bash
# 1. 登录获取 token
TOKEN="your_jwt_token"

# 2. 领取 5 单
curl -X POST http://localhost:8080/api/tasks/claim \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"count": 5}'

# 3. 验证返回的任务数量
```

### 测试2：退单功能
```bash
# 1. 查看我的任务
curl -X GET http://localhost:8080/api/tasks/my \
  -H "Authorization: Bearer $TOKEN"

# 2. 退回部分任务（假设任务ID为1,2,3）
curl -X POST http://localhost:8080/api/tasks/return \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"task_ids": [1, 2, 3]}'

# 3. 再次查看我的任务（应该少了3个）
curl -X GET http://localhost:8080/api/tasks/my \
  -H "Authorization: Bearer $TOKEN"
```

### 测试3：边界条件
```bash
# 测试领取数量超出范围
curl -X POST http://localhost:8080/api/tasks/claim \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"count": 100}'
# 期望：400 错误，提示数量必须在1-50之间

# 测试退单超出范围
curl -X POST http://localhost:8080/api/tasks/return \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"task_ids": [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51]}'
# 期望：400 错误，提示数量必须在1-50之间
```

---

## 🔧 技术实现细节

### 后端变更
1. **models.go**: 新增 `ClaimTasksRequest` 和 `ReturnTasksRequest` DTO
2. **task_repo.go**: 新增 `ReturnTasks()` 方法，支持批量退单
3. **task_service.go**: 
   - 修改 `ClaimTasks()` 支持自定义数量
   - 新增 `ReturnTasks()` 业务逻辑
4. **task.go (handler)**: 
   - 修改 `ClaimTasks` 处理器
   - 新增 `ReturnTasks` 处理器
5. **main.go**: 注册 `/tasks/return` 路由

### 前端变更
1. **task.ts**: 
   - 修改 `claimTasks()` 接受 count 参数
   - 新增 `returnTasks()` API 调用
2. **Dashboard.vue**: 
   - 新增数字输入框（领取数量）
   - 新增任务复选框
   - 新增退单按钮
   - 新增退单逻辑

### 数据库操作
- ✅ 退单使用原有的 `ResetTask` 逻辑
- ✅ 支持事务性批量操作
- ✅ 只更新属于指定审核员的任务

### Redis 清理
- ✅ 退单时清除 `task:claimed:{reviewerID}` 集合
- ✅ 退单时删除 `task:lock:{taskID}` 锁
- ✅ 使用 Pipeline 批量操作提高性能

---

## 📌 注意事项

1. **兼容性**：前端必须传递 `count` 参数，否则会报错
2. **性能**：批量操作使用 Redis Pipeline，性能优异
3. **安全性**：只能退回属于自己的任务，防止恶意退单
4. **用户体验**：退单前有确认对话框，防止误操作
5. **数据一致性**：退单同时更新数据库和 Redis，保证一致性

---

## 🎯 后续优化建议

1. **统计功能**：添加退单次数统计，分析审核员行为
2. **退单原因**：可选择性记录退单原因，便于改进任务分配
3. **智能推荐**：根据审核员历史表现推荐合适的领取数量
4. **任务预览**：领取前可以预览任务内容，减少退单率

---

## 📞 支持

如有问题，请查看：
- API_TESTING.md - 完整的 API 测试文档
- PROJECT_SUMMARY.md - 项目概述
- README.md - 快速开始指南

**更新日期**: 2025-10-25
**版本**: v1.1.0

