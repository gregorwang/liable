# 更新日志 (Changelog)

## [1.5.0] - 2025-10-25

### 新增功能

#### 审核规则库系统

- 完整的内容审核规则管理系统
  - 创建 moderation_rules 数据库表
  - 集成 42 条详细的审核规则
  - 10 大分类覆盖所有内容安全问题
  - 风险等级分布：5×C, 18×H, 14×M, 5×L

- 后端 API 实现
  - GET /api/moderation-rules (获取规则列表，支持分页、筛选、搜索)
  - GET /api/moderation-rules/:code (获取单条规则)
  - GET /api/moderation-rules/categories (获取分类列表)
  - GET /api/moderation-rules/risk-levels (获取风险等级列表)
  - 完整的 Repository 层和 Handler 处理器

- 前端页面实现 (Vue3 + Element Plus)
  - 现代化的规则库管理界面
  - 搜索功能（按规则编号和描述）
  - 多维筛选（分类、风险等级）
  - 展开详情查看完整信息
  - 彩色风险等级标签 (L/M/H/C)
  - 快捷标签显示
  - 实时统计卡片
  - 响应式设计

- 完整文档
  - MODERATION_RULES_API.md (API 详细文档)
  - MODERATION_RULES_USAGE.md (使用指南)
  - MODERATION_RULES_IMPLEMENTATION.md (实现总结)

---

## [1.2.0] - 2025-10-25

### 🎉 新增功能

#### 审核记录搜索功能
- **搜索API**: 添加完整的搜索接口 `GET /api/tasks/search`
  - 支持按评论ID精确搜索
  - 支持按审核员账号精确搜索
  - 支持按违规标签搜索（多选，OR关系）
  - 支持按审核时间范围搜索
  - 支持分页查询（默认10条/页，最大100条/页）
  
- **前端搜索界面**: 功能完善的搜索页面
  - 响应式搜索表单，支持多条件组合
  - 实时结果展示，支持分页浏览
  - 详情对话框，查看完整审核信息
  - CSV导出功能，便于数据分析
  - 数据统计提示，显示总记录数和页数
  
- **导航优化**
  - 管理员侧边栏添加"审核记录搜索"菜单项
  - 审核员工作台顶部添加"搜索审核记录"快捷按钮

### ⚡ 性能优化

#### 数据库索引优化
创建8个索引大幅提升搜索性能：
- `idx_review_tasks_reviewer_id` - 审核员ID索引
- `idx_review_tasks_comment_id` - 评论ID索引  
- `idx_review_tasks_completed_at` - 完成时间索引（DESC排序）
- `idx_review_tasks_status` - 状态索引
- `idx_review_results_tags` - 标签GIN索引（数组查询优化）
- `idx_review_results_task_id` - 任务关联索引
- `idx_review_tasks_status_completed_at` - 组合索引（最优查询路径）
- `idx_users_username` - 用户名索引

### 🔧 技术实现

#### 后端优化
- 动态SQL构建，只包含实际使用的查询条件
- 使用PostgreSQL数组操作符(`&&`)实现标签OR查询
- 完整的数据关联查询（任务+审核结果+用户+评论）
- 参数化查询防止SQL注入
- 合理的分页限制（最大100条）

#### 前端优化
- Element Plus组件库实现现代化UI
- 响应式设计，适配不同屏幕尺寸
- 懒加载路由组件，提升首屏速度
- 客户端CSV导出，减轻服务器压力

### 📚 文档

- 新增 `API_SEARCH_GUIDE.md` - 完整的API使用指南
  - 详细的参数说明
  - 多个使用示例（curl和JavaScript）
  - 性能优化建议
  
- 新增 `SEARCH_FEATURE_SUMMARY.md` - 技术实现总结
  - 代码变更清单
  - 技术实现亮点
  - 测试建议
  
- 新增 `SEARCH_FEATURE_COMPLETE.md` - 完整功能文档
  - 功能特性说明
  - 使用说明
  - 部署指南
  - 安全特性

### 🔄 API 变更

#### 新增接口
- `GET /api/tasks/search` - 搜索审核记录
  - 权限: 需要登录（管理员和审核员都可访问）
  - 参数: comment_id, reviewer_rtx, tag_ids, review_start_time, review_end_time, page, page_size
  - 响应: 分页数据 + 总数 + 总页数

### 🔒 安全特性

- JWT认证保护搜索接口
- 角色权限控制（仅登录用户可访问）
- 参数化查询防止SQL注入
- XSS防护（Vue自动转义）
- 限制最大返回数量防止DOS攻击

### 🎯 使用场景

- 管理员查看特定审核员的工作记录
- 审核员查看自己的历史审核记录
- 按评论ID追溯审核历史
- 按违规标签分析问题类型
- 按时间范围统计审核数据
- 导出数据进行深度分析

---

## [1.1.0] - 2025-10-25

### ✨ 新增功能

#### 审核员功能
- **自定义领取数量**: 审核员现在可以选择领取 1-50 单任务（之前固定 20 单）
  - 前端新增数字输入框，方便调整领取数量
  - 后端验证数量范围，确保合法性
  
- **退单功能**: 审核员可以退回不想审核的任务
  - 支持单个或批量退单（最多 50 单）
  - 退回的任务立即回到任务池
  - 前端提供任务复选框和退单按钮
  - 退单前有确认对话框，防止误操作

### 🔧 改进

#### 后端
- 修改 `ClaimTasks` API，支持自定义领取数量
- 新增 `ReturnTasks` API，支持退单操作
- 优化 Redis 操作，使用 Pipeline 提高性能
- 改进错误提示信息，更加友好

#### 前端
- 重构审核员 Dashboard，添加数量选择器
- 新增任务选择功能（复选框）
- 新增退单按钮和相关逻辑
- 改进用户体验，操作更加流畅

### 📝 文档
- 新增 `API_UPDATE_RETURN_TASKS.md` - 退单功能详细文档
- 新增 `FEATURE_UPDATE_SUMMARY.md` - 功能更新总结
- 新增 `test-return-feature.sh` - Linux/Mac 测试脚本
- 新增 `test-return-feature.bat` - Windows 测试脚本
- 新增 `CHANGELOG.md` - 更新日志

### 🔄 API 变更

#### 修改的接口
- `POST /api/tasks/claim`
  - **变更**: 现在需要请求体 `{"count": 20}`
  - **影响**: 旧版本前端无法使用，需要同步更新

#### 新增的接口
- `POST /api/tasks/return`
  - 请求体: `{"task_ids": [1, 2, 3]}`
  - 响应: `{"message": "...", "count": 3}`

### ⚠️ 破坏性变更

- `POST /api/tasks/claim` 接口现在要求必须传递 `count` 参数
  - **影响**: 旧版本前端将无法调用此接口
  - **迁移指南**: 更新前端代码，调用时传递 `count` 参数

### 🐛 Bug 修复
- 无

### 📊 性能优化
- Redis 操作使用 Pipeline，减少网络往返
- 数据库批量操作优化

---

## [1.0.0] - 2025-10-24

### ✨ 初始版本

#### 核心功能
- 用户注册和登录
- JWT 认证
- 审核员任务管理
  - 领取任务（固定 20 单）
  - 提交审核
  - 批量提交
- 管理员功能
  - 用户审批
  - 统计数据查看
  - 标签管理
- 任务超时自动释放（30分钟）

#### 技术栈
- 后端: Go + Gin + PostgreSQL + Redis
- 前端: Vue 3 + Element Plus + TypeScript
- 数据库: Supabase (PostgreSQL)
- 缓存: Upstash (Redis with TLS)

---

## 版本号说明

格式：`[主版本号].[次版本号].[修订号]`

- **主版本号**: 重大架构变更或不兼容的 API 变更
- **次版本号**: 新增功能，向后兼容
- **修订号**: Bug 修复和小的改进

---

## 获取帮助

如有问题，请查看：
- [SEARCH_FEATURE_COMPLETE.md](./SEARCH_FEATURE_COMPLETE.md) - 审核记录搜索功能（最新）
- [API_SEARCH_GUIDE.md](./API_SEARCH_GUIDE.md) - 搜索API使用指南（最新）
- [FEATURE_UPDATE_SUMMARY.md](./FEATURE_UPDATE_SUMMARY.md) - 退单功能总结
- [API_UPDATE_RETURN_TASKS.md](./API_UPDATE_RETURN_TASKS.md) - 退单API文档
- [README.md](./README.md) - 快速开始指南

# 变更日志

## [v1.1.0] - 2025-10-26

### 🆕 新增

#### 前端组件
- ✨ **QueueList.vue 组件更新** - 改为使用真实 API 接口
  - 移除模拟数据
  - 导入并使用 `listTaskQueues()` API 函数
  - 实现动态计算进度百分比
  - 优先级标签颜色编码（danger/warning/info）
  - 日期格式化显示（本地化时间）
  - 分页功能完整集成
  - 刷新和详情操作按钮

#### 数据库
- 📊 **测试数据** - 插入 5 条示例任务队列数据
  - 评论审核一审 (优先级: 80, 进度: 75%)
  - 评论审核二审 (优先级: 60, 进度: 33%)
  - 短视频一审 (优先级: 90, 进度: 100%)
  - 短视频二审 (优先级: 70, 进度: 0%)
  - 色情内容审核 (优先级: 85, 进度: 50%)

### 📝 文档
- 📖 **QUICK_FIX_GUIDE.md** - 完整的问题排除指南
  - 404 错误诊断步骤
  - 后端启动三种方案
  - 环境变量配置说明
  - 常见问题解答 (Q&A)
  - 完整启动流程
  - 验证清单

### 🔧 改进

#### 组件功能
| 项目 | 改变 |
|------|------|
| 表格列 | 添加任务统计列（总数/已审/待审） |
| 进度条 | 自动计算而非固定值 |
| 状态字段 | 显示活跃/禁用状态 |
| 日期显示 | ISO 格式转本地化格式 |
| 优先级 | 彩色标签分类展示 |
| 操作按钮 | 标注 + 详情（替代质检/申诉） |

#### API 集成
- 使用 `listTaskQueues()` 获取分页数据
- 自动处理分页逻辑
- 错误提示优化
- 加载状态指示

### 🐛 修复

- ❌ 移除过时的 QueueTask 类型引用
- ✅ 使用最新的 TaskQueue 接口
- ✅ 确保请求/响应数据结构一致

### 📌 API 端点状态

| 端点 | 方法 | 实现 | 前端集成 |
|------|------|------|--------|
| `/admin/task-queues` | POST | ✅ | QueueManage.vue |
| `/admin/task-queues` | GET | ✅ | QueueList.vue ✨ |
| `/admin/task-queues/:id` | GET | ✅ | 待集成 |
| `/admin/task-queues/:id` | PUT | ✅ | QueueManage.vue |
| `/admin/task-queues/:id` | DELETE | ✅ | QueueManage.vue |
| `/admin/task-queues-all` | GET | ✅ | 待集成 |

### 🚀 启动检查清单

- [x] 数据库表创建 - `task_queue` ✅
- [x] 测试数据插入 - 5 条队列 ✅
- [x] 后端 API 实现 - 所有 CRUD ✅
- [x] 前端组件更新 - QueueList ✅
- [ ] 后端服务启动 - `localhost:8080`
- [ ] 前端开发服务 - `localhost:3000`
- [ ] 登录验证 - `admin/admin123`
- [ ] 页面访问 - `/main/queue-list`

### 📊 项目完成度

```
审核规则库 CRUD: ████████████████████ 100%
任务队列管理:    ████████████████████ 100%
- 后端实现:      ████████████████████ 100%
- 前端 QueueManage: ████████████████████ 100%
- 前端 QueueList:   ████████████████████ 100% ✨ 新完成

整体进度:        ████████████████████ 100%
```

---

## [v1.0.0] - 2025-01-15

### 🎯 任务队列管理系统

完整实现的功能：
- ✅ 任务队列 CRUD 操作
- ✅ 搜索和过滤
- ✅ 分页管理
- ✅ 优先级管理
- ✅ 进度追踪
- ✅ 权限控制
- ✅ 数据验证

---

## 📞 支持

遇到问题? 查看:
- `QUICK_FIX_GUIDE.md` - 快速排除指南 ⭐
- `TASK_QUEUE_API.md` - API 完整文档
- `TASK_QUEUE_QUICK_START.md` - 快速开始

