# Implementation Plan: Code Deduplication Refactor

## Overview

本实现计划将代码去重重构分解为可执行的任务，按照后端基础设施 → 后端重构 → 前端基础设施 → 前端重构的顺序进行。每个任务都是增量式的，确保系统在重构过程中保持可用。

## Tasks

- [x] 1. 创建后端基础抽象层
  - [x] 1.1 创建 Handler 基础包和统一响应函数
    - 创建 `internal/handlers/base/response.go`
    - 实现 `RespondError`, `RespondSuccess`, `RespondBadRequest`, `RespondInternalError`
    - 定义 `ErrorResponse` 结构体
    - _Requirements: 6.1, 6.2, 6.3_

  - [x] 1.2 创建通用 Handler 函数
    - 创建 `internal/handlers/base/handlers.go`
    - 实现 `HandleClaimTasks`, `HandleGetMyTasks`, `HandleSubmitReview`, `HandleReturnTasks`
    - 创建 `TaskHandlerConfig` 配置结构
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

  - [x] 1.3 编写 Handler 基础包单元测试
    - 测试响应函数格式正确性
    - 测试通用 Handler 函数参数绑定
    - _Requirements: 6.4, 6.5, 6.6_

- [x] 2. 创建 Repository 基础抽象层
  - [x] 2.1 创建 Repository 配置和基类
    - 创建 `internal/repository/base/config.go`
    - 创建 `internal/repository/base/task_repo.go`
    - 实现 `TaskRepoConfig` 配置结构
    - 实现 `BaseTaskRepository` 基类
    - _Requirements: 2.1, 2.7_

  - [x] 2.2 实现 BaseTaskRepository 核心方法
    - 实现 `ClaimTasks` 方法（使用 FOR UPDATE SKIP LOCKED）
    - 实现 `CompleteTask`, `ReturnTasks`, `FindExpiredTasks`, `ResetTask` 方法
    - _Requirements: 2.2, 2.3, 2.4, 2.5, 2.6_

  - [x] 2.3 编写 Repository 属性测试
    - **Property 2: Repository 事务安全性**
    - **Validates: Requirements 2.1, 2.6**

- [x] 3. 创建 Service 基础抽象层
  - [x] 3.1 创建 Service 配置和基类
    - 创建 `internal/services/base/config.go`
    - 创建 `internal/services/base/task_service.go`
    - 实现 `TaskServiceConfig` 配置结构
    - 实现 `BaseTaskService` 基类
    - _Requirements: 3.6, 3.7_

  - [x] 3.2 实现 BaseTaskService 核心方法
    - 实现 `ValidateClaimCount`, `CheckExistingTasks` 验证方法
    - 实现 `TrackClaimedTasks`, `CleanupTaskTracking` Redis 方法
    - 实现 `ValidateReturnCount` 方法
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

  - [x] 3.3 编写 Service 属性测试
    - **Property 3: Service 验证逻辑一致性**
    - **Property 4: Redis 追踪数据完整性**
    - **Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.6**

- [x] 4. Checkpoint - 后端基础设施验证
  - 确保所有测试通过，ask the user if questions arise.


- [x] 5. 重构现有 Handler 使用基础抽象
  - [x] 5.1 重构 TaskHandler 使用通用函数
    - 修改 `internal/handlers/task.go`
    - 使用 `base.HandleClaimTasks`, `base.HandleGetMyTasks` 等
    - 保持 API 兼容性
    - _Requirements: 1.6, 1.7_

  - [x] 5.2 重构 QualityCheckHandler 使用通用函数
    - 修改 `internal/handlers/quality_check_handler.go`
    - 统一使用 `middleware.GetUserID(c)` 获取用户ID
    - 使用通用 Handler 函数
    - _Requirements: 1.6, 1.7_

  - [x] 5.3 重构 SecondReviewHandler 使用通用函数
    - 修改 `internal/handlers/second_review_handler.go`
    - 使用通用 Handler 函数
    - _Requirements: 1.6_

  - [x] 5.4 重构 VideoHandler 使用通用函数
    - 修改 `internal/handlers/video_handler.go`
    - 重构一审和二审相关方法使用通用函数
    - _Requirements: 1.6_

  - [x] 5.5 编写 Handler 一致性属性测试
    - **Property 1: Handler 通用函数行为一致性**
    - **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 1.6**

- [x] 6. 重构现有 Repository 使用基础抽象
  - [x] 6.1 重构 TaskRepository 继承 BaseTaskRepository
    - 修改 `internal/repository/task_repo.go`
    - 复用 `ClaimTasks`, `ReturnTasks` 等方法
    - 保留特定的查询方法
    - _Requirements: 2.1, 2.2, 2.3_

  - [x] 6.2 重构 QualityCheckRepository 继承 BaseTaskRepository
    - 修改 `internal/repository/quality_check_repo.go`
    - 复用基类方法
    - _Requirements: 2.1, 2.2, 2.3_

  - [x] 6.3 重构 VideoFirstReviewRepository 继承 BaseTaskRepository
    - 修改 `internal/repository/video_first_review_repo.go`
    - 复用基类方法
    - _Requirements: 2.1, 2.2, 2.3_

  - [x] 6.4 重构 VideoSecondReviewRepository 继承 BaseTaskRepository
    - 修改 `internal/repository/video_second_review_repo.go`
    - 复用基类方法
    - _Requirements: 2.1, 2.2, 2.3_

- [x] 7. 重构现有 Service 使用基础抽象
  - [x] 7.1 重构 TaskService 使用 BaseTaskService
    - 修改 `internal/services/task_service.go`
    - 复用验证和 Redis 追踪方法
    - _Requirements: 3.1, 3.2, 3.3, 3.4_

  - [x] 7.2 重构 QualityCheckService 使用 BaseTaskService
    - 修改 `internal/services/quality_check_service.go`
    - 复用验证和 Redis 追踪方法
    - _Requirements: 3.1, 3.2, 3.3, 3.4_

  - [x] 7.3 重构 VideoFirstReviewService 使用 BaseTaskService
    - 修改 `internal/services/video_service.go` 中的 VideoFirstReviewService
    - 复用验证和 Redis 追踪方法
    - _Requirements: 3.1, 3.2, 3.3, 3.4_

  - [x] 7.4 重构 SecondReviewService 使用 BaseTaskService
    - 修改 `internal/services/second_review_service.go`
    - 复用验证和 Redis 追踪方法
    - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 8. 优化中间件
  - [x] 8.1 创建通用限流器
    - 创建 `internal/middleware/rate_limit_v2.go`
    - 实现 `CreateRateLimiter` 工厂函数
    - 实现 `RateLimitConfig` 配置结构
    - _Requirements: 7.1, 7.2, 7.3_

  - [x] 8.2 迁移现有限流器到新实现
    - 使用 `CreateRateLimiter` 重新实现 `GlobalRateLimiter`, `EndpointRateLimiter`, `UserRateLimiter`
    - 保持行为兼容
    - _Requirements: 7.4, 7.5_

  - [x] 8.3 编写限流器属性测试
    - **Property 7: 限流器行为等价性**
    - **Validates: Requirements 7.1, 7.2, 7.3, 7.5**

- [x] 9. Checkpoint - 后端重构验证
  - 确保所有测试通过，ask the user if questions arise.


- [x] 10. 创建前端基础抽象层
  - [x] 10.1 创建 API 工厂函数
    - 创建 `frontend/src/api/taskApiFactory.ts`
    - 实现 `createTaskApi` 工厂函数
    - 定义 `TaskApiConfig`, `TaskApiMethods` 类型
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

  - [x] 10.2 编写 API 工厂属性测试
    - **Property 5: API 工厂函数完整性**
    - **Validates: Requirements 5.1, 5.2, 5.3, 5.5**

  - [x] 10.3 创建通用 Dashboard 组件
    - 创建 `frontend/src/components/GenericReviewDashboard.vue`
    - 实现可配置的标题、统计栏、操作栏
    - 实现 slot 支持自定义任务卡片渲染
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_

  - [x] 10.4 创建 Dashboard 配置文件
    - 创建 `frontend/src/config/dashboardConfigs.ts`
    - 定义各类型 Dashboard 的配置
    - _Requirements: 4.6_

  - [x] 10.5 编写 Dashboard 组件单元测试
    - 测试 props 传递正确性
    - 测试事件触发正确性
    - _Requirements: 4.1, 4.2, 4.5_

- [x] 11. 重构前端 API 模块
  - [x] 11.1 使用工厂函数重构 videoReview API
    - 修改 `frontend/src/api/videoReview.ts`
    - 使用 `createTaskApi` 生成一审和二审 API
    - 保持导出兼容性
    - _Requirements: 5.5_

  - [x] 11.2 使用工厂函数重构 secondReview API
    - 修改 `frontend/src/api/secondReview.ts`
    - 使用 `createTaskApi` 生成 API
    - _Requirements: 5.5_

  - [x] 11.3 使用工厂函数重构 qualityCheck API
    - 修改 `frontend/src/api/qualityCheck.ts`
    - 使用 `createTaskApi` 生成 API
    - _Requirements: 5.5_

- [x] 12. 重构前端 Dashboard 组件
  - [x] 12.1 重构 VideoFirstReviewDashboard 使用通用组件
    - 修改 `frontend/src/views/reviewer/VideoFirstReviewDashboard.vue`
    - 使用 `GenericReviewDashboard` 组件
    - 通过 slot 自定义任务卡片
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_

  - [x] 12.2 重构 VideoSecondReviewDashboard 使用通用组件
    - 修改 `frontend/src/views/reviewer/VideoSecondReviewDashboard.vue`
    - 使用 `GenericReviewDashboard` 组件
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_

  - [x] 12.3 重构 QualityCheckDashboard 使用通用组件
    - 修改 `frontend/src/views/reviewer/QualityCheckDashboard.vue`
    - 使用 `GenericReviewDashboard` 组件
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_

  - [x] 12.4 编写 Dashboard 配置属性测试
    - **Property 8: Dashboard 组件配置正确性**
    - **Validates: Requirements 4.1, 4.2, 4.4, 4.5, 4.6**

- [x] 13. 创建错误处理工具
  - [x] 13.1 创建前端错误处理工具
    - 创建 `frontend/src/utils/errorHandler.ts`
    - 实现 `handleApiError`, `isApiError` 函数
    - _Requirements: 6.1, 6.2_

  - [x] 13.2 在 Dashboard 组件中使用统一错误处理
    - 更新所有 Dashboard 使用 `handleApiError`
    - _Requirements: 6.6_

  - [x] 13.3 编写错误处理属性测试
    - **Property 6: 错误响应格式一致性**
    - **Validates: Requirements 6.1, 6.2, 6.4, 6.5**

- [x] 14. Final Checkpoint - 全面验证
  - 确保所有测试通过，ask the user if questions arise.
  - 验证所有 API 端点行为与重构前一致
  - 验证所有 Dashboard 功能正常

## Notes

- All tasks are required including comprehensive testing
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- 重构过程中保持 API 兼容性，避免破坏现有功能
- 建议先在开发环境完成全部重构后再部署到生产环境

