import request from './request'

/**
 * Task API 配置
 */
export interface TaskApiConfig {
    /** API 基础路径，如 '/tasks/quality-check' */
    basePath: string
    /** 任务类型名称，用于日志 */
    taskTypeName: string
}

/**
 * 领取任务请求
 */
export interface ClaimTasksRequest {
    count: number
}

/**
 * 退回任务请求
 */
export interface ReturnTasksRequest {
    task_ids: number[]
}

/**
 * 通用任务响应
 */
export interface TasksResponse<T> {
    tasks: T[]
    count: number
}

/**
 * 消息响应
 */
export interface MessageResponse {
    message: string
}

/**
 * 带数量的消息响应
 */
export interface MessageCountResponse {
    message: string
    count: number
}

/**
 * Task API 方法集合
 */
export interface TaskApiMethods<TTask, TSubmitRequest> {
    /** 领取任务 */
    claimTasks: (count: number) => Promise<TasksResponse<TTask>>
    /** 获取我的任务 */
    getMyTasks: () => Promise<TasksResponse<TTask>>
    /** 提交单个审核 */
    submitReview: (review: TSubmitRequest) => Promise<MessageResponse>
    /** 批量提交审核 */
    submitBatchReviews: (reviews: TSubmitRequest[]) => Promise<MessageCountResponse>
    /** 退回任务 */
    returnTasks: (taskIds: number[]) => Promise<MessageCountResponse>
}

/**
 * 创建任务 API
 * 
 * @param config API 配置
 * @returns 任务 API 方法集合
 * 
 * @example
 * ```ts
 * const qcApi = createTaskApi<QCTask, SubmitQCRequest>({
 *   basePath: '/tasks/quality-check',
 *   taskTypeName: 'quality check'
 * })
 * 
 * // 使用
 * const tasks = await qcApi.claimTasks(10)
 * await qcApi.submitReview({ task_id: 1, is_passed: true })
 * ```
 */
export function createTaskApi<TTask, TSubmitRequest>(
    config: TaskApiConfig
): TaskApiMethods<TTask, TSubmitRequest> {
    const { basePath } = config

    return {
        claimTasks: (count: number) => {
            return request.post<any, TasksResponse<TTask>>(`${basePath}/claim`, { count })
        },

        getMyTasks: () => {
            return request.get<any, TasksResponse<TTask>>(`${basePath}/my`)
        },

        submitReview: (review: TSubmitRequest) => {
            return request.post<any, MessageResponse>(`${basePath}/submit`, review)
        },

        submitBatchReviews: (reviews: TSubmitRequest[]) => {
            return request.post<any, MessageCountResponse>(`${basePath}/submit-batch`, { reviews })
        },

        returnTasks: (taskIds: number[]) => {
            return request.post<any, MessageCountResponse>(`${basePath}/return`, { task_ids: taskIds })
        }
    }
}

/**
 * 扩展的任务 API 方法（包含统计功能）
 */
export interface ExtendedTaskApiMethods<TTask, TSubmitRequest, TStats>
    extends TaskApiMethods<TTask, TSubmitRequest> {
    /** 获取统计信息 */
    getStats: () => Promise<TStats>
}

/**
 * 创建带统计功能的任务 API
 */
export function createTaskApiWithStats<TTask, TSubmitRequest, TStats>(
    config: TaskApiConfig
): ExtendedTaskApiMethods<TTask, TSubmitRequest, TStats> {
    const baseApi = createTaskApi<TTask, TSubmitRequest>(config)

    return {
        ...baseApi,
        getStats: () => {
            return request.get<any, TStats>(`${config.basePath}/stats`)
        }
    }
}

// 预定义的 API 配置

export const QualityCheckApiConfig: TaskApiConfig = {
    basePath: '/tasks/quality-check',
    taskTypeName: 'quality check'
}

export const SecondReviewApiConfig: TaskApiConfig = {
    basePath: '/tasks/second-review',
    taskTypeName: 'second review'
}

export const VideoFirstReviewApiConfig: TaskApiConfig = {
    basePath: '/tasks/video-first-review',
    taskTypeName: 'video first review'
}

export const VideoSecondReviewApiConfig: TaskApiConfig = {
    basePath: '/tasks/video-second-review',
    taskTypeName: 'video second review'
}
