# Requirements Document

## Introduction

本文档定义了评论审核平台代码去重重构的需求规范。项目当前存在大量重复代码，主要集中在后端的 Handler、Repository、Service 层，以及前端的 Dashboard 组件和 API 调用层。本次重构旨在通过抽象公共逻辑、创建通用组件和工厂函数来消除重复代码，提高代码可维护性和一致性。

## Glossary

- **Handler**: Go 后端的 HTTP 请求处理层，负责接收请求、验证参数、调用 Service 并返回响应
- **Repository**: Go 后端的数据访问层，负责与数据库交互
- **Service**: Go 后端的业务逻辑层，负责处理业务规则和协调 Repository
- **Dashboard**: Vue.js 前端的审核工作台页面组件
- **API_Module**: 前端的 API 调用封装模块
- **Task_Type**: 任务类型，包括 first_review（一审）、second_review（二审）、quality_check（质检）、video_first_review（视频一审）、video_second_review（视频二审）
- **Base_Class**: 抽象基类，包含公共逻辑供子类继承
- **Factory_Function**: 工厂函数，用于生成具有相似结构但不同配置的对象或函数

## Requirements

### Requirement 1: 后端 Handler 层抽象

**User Story:** As a 开发者, I want 将重复的 Handler 逻辑抽象为通用函数, so that 减少代码重复并确保一致的错误处理。

#### Acceptance Criteria

1. THE Handler_Abstraction_Module SHALL 提供通用的任务领取处理函数 `HandleClaimTasks`
2. THE Handler_Abstraction_Module SHALL 提供通用的任务提交处理函数 `HandleSubmitReview`
3. THE Handler_Abstraction_Module SHALL 提供通用的批量提交处理函数 `HandleBatchSubmit`
4. THE Handler_Abstraction_Module SHALL 提供通用的任务退回处理函数 `HandleReturnTasks`
5. THE Handler_Abstraction_Module SHALL 提供通用的获取我的任务处理函数 `HandleGetMyTasks`
6. WHEN 使用通用 Handler 函数时, THE System SHALL 通过接口或泛型支持不同的任务类型
7. THE System SHALL 统一使用 `middleware.GetUserID(c)` 获取用户ID

### Requirement 2: 后端 Repository 层抽象

**User Story:** As a 开发者, I want 将重复的数据库操作抽象为通用 Repository, so that 减少 SQL 重复并确保一致的事务处理。

#### Acceptance Criteria

1. THE Base_Task_Repository SHALL 提供通用的任务领取方法 `ClaimTasks`，支持参数化表名和字段
2. THE Base_Task_Repository SHALL 提供通用的任务完成方法 `CompleteTask`
3. THE Base_Task_Repository SHALL 提供通用的任务退回方法 `ReturnTasks`
4. THE Base_Task_Repository SHALL 提供通用的过期任务查询方法 `FindExpiredTasks`
5. THE Base_Task_Repository SHALL 提供通用的任务重置方法 `ResetTask`
6. WHEN 执行任务领取操作时, THE System SHALL 使用 `FOR UPDATE SKIP LOCKED` 确保并发安全
7. THE System SHALL 支持通过配置指定不同的表名、字段名和关联查询

### Requirement 3: 后端 Service 层抽象

**User Story:** As a 开发者, I want 将重复的业务逻辑抽象为通用 Service, so that 减少验证和 Redis 操作的重复代码。

#### Acceptance Criteria

1. THE Base_Task_Service SHALL 提供通用的任务领取验证逻辑（数量验证 1-50）
2. THE Base_Task_Service SHALL 提供通用的未完成任务检查逻辑
3. THE Base_Task_Service SHALL 提供通用的 Redis 任务追踪逻辑（claimed set 和 lock key）
4. THE Base_Task_Service SHALL 提供通用的 Redis 清理逻辑
5. THE Base_Task_Service SHALL 提供通用的过期任务释放逻辑
6. WHEN 配置不同任务类型时, THE System SHALL 支持自定义 Redis key 前缀
7. THE System SHALL 确保所有任务类型使用相同的超时配置 `config.AppConfig.TaskTimeoutMinutes`

### Requirement 4: 前端 Dashboard 组件抽象

**User Story:** As a 开发者, I want 将重复的 Dashboard 组件抽象为通用组件, so that 减少 Vue 组件代码重复并确保一致的用户体验。

#### Acceptance Criteria

1. THE Generic_Review_Dashboard SHALL 提供可配置的页面标题
2. THE Generic_Review_Dashboard SHALL 提供可配置的统计卡片（支持不同的统计项）
3. THE Generic_Review_Dashboard SHALL 提供通用的操作栏（领取、退单、批量提交、刷新）
4. THE Generic_Review_Dashboard SHALL 提供可配置的任务卡片渲染（通过 slot 或 render props）
5. THE Generic_Review_Dashboard SHALL 提供通用的事件处理逻辑（claim, return, submit, refresh）
6. WHEN 使用通用 Dashboard 时, THE System SHALL 通过 props 传入 API 函数和配置
7. THE System SHALL 保持现有的响应式设计和样式一致性

### Requirement 5: 前端 API 模块抽象

**User Story:** As a 开发者, I want 将重复的 API 调用抽象为工厂函数, so that 减少 API 模块代码重复。

#### Acceptance Criteria

1. THE API_Factory SHALL 提供 `createTaskApi` 工厂函数
2. WHEN 调用 `createTaskApi` 时, THE System SHALL 生成包含 claim、getMyTasks、submit、submitBatch、return 方法的 API 对象
3. THE API_Factory SHALL 支持通过参数配置 API 路径前缀
4. THE API_Factory SHALL 支持通过泛型指定请求和响应类型
5. THE System SHALL 保持与现有 API 调用的兼容性

### Requirement 6: 错误处理标准化

**User Story:** As a 开发者, I want 统一的错误处理和响应格式, so that 确保 API 响应的一致性。

#### Acceptance Criteria

1. THE Error_Handler SHALL 提供统一的错误响应格式 `{"error": "message", "code": "ERROR_CODE"}`
2. THE Error_Handler SHALL 区分客户端错误（400）和服务器错误（500）
3. THE Error_Handler SHALL 提供标准的成功响应格式
4. WHEN 发生验证错误时, THE System SHALL 返回 400 状态码和详细错误信息
5. WHEN 发生服务器错误时, THE System SHALL 返回 500 状态码并记录详细日志
6. THE System SHALL 在所有 Handler 中使用统一的错误处理函数

### Requirement 7: 中间件优化

**User Story:** As a 开发者, I want 优化重复的中间件逻辑, so that 减少 rate limiter 代码重复。

#### Acceptance Criteria

1. THE Rate_Limiter SHALL 提供通用的限流核心逻辑函数
2. THE Rate_Limiter SHALL 支持通过配置指定限流维度（IP、Endpoint、User）
3. THE Rate_Limiter SHALL 支持通过配置指定限流参数（limit、window）
4. WHEN 创建新的限流器时, THE System SHALL 复用核心限流逻辑
5. THE System SHALL 保持现有的限流行为不变

