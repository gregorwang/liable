import { ElMessage } from 'element-plus'
import { buildTraceMessage } from './traceNotice'

/**
 * API 错误响应结构
 */
export interface ApiErrorResponse {
    error?: string
    message?: string
    code?: string | number
    error_type?: string
    error_description?: string
    trace_id?: string
    http_status?: number
    details?: Record<string, any>
}

/**
 * 标准化的错误对象
 */
export interface StandardError {
    message: string
    code?: string | number
    details?: Record<string, any>
    originalError?: any
}

/**
 * 检查是否为 API 错误响应
 */
export function isApiError(error: any): error is { response: { data: ApiErrorResponse } } {
    return (
        error &&
        typeof error === 'object' &&
        'response' in error &&
        error.response &&
        typeof error.response === 'object' &&
        'data' in error.response &&
        error.response.data !== null &&
        error.response.data !== undefined
    )
}

/**
 * 检查是否为网络错误
 */
export function isNetworkError(error: any): boolean {
    return (
        error &&
        typeof error === 'object' &&
        (error.code === 'ECONNABORTED' ||
            error.code === 'ERR_NETWORK' ||
            error.message === 'Network Error')
    )
}

/**
 * 检查是否为超时错误
 */
export function isTimeoutError(error: any): boolean {
    return (
        error &&
        typeof error === 'object' &&
        (error.code === 'ECONNABORTED' || error.message?.includes('timeout'))
    )
}

/**
 * 从错误对象中提取错误消息
 */
export function extractErrorMessage(error: any, defaultMessage = '操作失败，请稍后重试'): string {
    // 用户取消操作
    if (error === 'cancel') {
        return ''
    }

    // API 错误响应
    if (isApiError(error)) {
        const data = error.response.data
        return data.error || data.message || defaultMessage
    }

    // 网络错误
    if (isNetworkError(error)) {
        return '网络连接失败，请检查网络后重试'
    }

    // 超时错误
    if (isTimeoutError(error)) {
        return '请求超时，请稍后重试'
    }

    // 普通 Error 对象
    if (error instanceof Error) {
        return error.message || defaultMessage
    }

    // 字符串错误
    if (typeof error === 'string') {
        return error
    }

    return defaultMessage
}

/**
 * 标准化错误对象
 */
export function normalizeError(error: any, defaultMessage?: string): StandardError {
    const message = extractErrorMessage(error, defaultMessage)

    const standardError: StandardError = {
        message,
        originalError: error
    }

    if (isApiError(error)) {
        const data = error.response.data
        standardError.code = data.code
        standardError.details = data.details
    }

    return standardError
}

/**
 * 处理 API 错误并显示消息
 * 
 * @param error 错误对象
 * @param options 配置选项
 * @returns 标准化的错误对象
 * 
 * @example
 * ```ts
 * try {
 *   await someApiCall()
 * } catch (error) {
 *   handleApiError(error, { defaultMessage: '提交失败' })
 * }
 * ```
 */
export function handleApiError(
    error: any,
    options: {
        defaultMessage?: string
        showMessage?: boolean
        logError?: boolean
    } = {}
): StandardError {
    const {
        defaultMessage = '操作失败，请稍后重试',
        showMessage = true,
        logError = true
    } = options

    // 用户取消操作，不显示错误
    if (error === 'cancel') {
        return { message: '', originalError: error }
    }

    const standardError = normalizeError(error, defaultMessage)

    if (!standardError.message) {
        standardError.message = defaultMessage
    }

    // 显示错误消息
    if (showMessage && standardError.message) {
        ElMessage.error(buildTraceMessage(standardError.message, error))
    }

    // 记录错误日志
    if (logError) {
        console.error('API Error:', error)
    }

    return standardError
}

/**
 * 创建带有默认配置的错误处理器
 * 
 * @example
 * ```ts
 * const handleError = createErrorHandler({ defaultMessage: '审核提交失败' })
 * 
 * try {
 *   await submitReview(data)
 * } catch (error) {
 *   handleError(error)
 * }
 * ```
 */
export function createErrorHandler(defaultOptions: {
    defaultMessage?: string
    showMessage?: boolean
    logError?: boolean
} = {}) {
    return (error: any, overrideOptions: typeof defaultOptions = {}) => {
        return handleApiError(error, { ...defaultOptions, ...overrideOptions })
    }
}

/**
 * 包装异步函数，自动处理错误
 * 
 * @example
 * ```ts
 * const safeSubmit = withErrorHandling(
 *   async (data) => await submitReview(data),
 *   { defaultMessage: '提交失败' }
 * )
 * 
 * await safeSubmit(reviewData)
 * ```
 */
export function withErrorHandling<T extends (...args: any[]) => Promise<any>>(
    fn: T,
    options: {
        defaultMessage?: string
        showMessage?: boolean
        logError?: boolean
        onError?: (error: StandardError) => void
    } = {}
): T {
    return (async (...args: Parameters<T>) => {
        try {
            return await fn(...args)
        } catch (error) {
            const standardError = handleApiError(error, options)
            if (options.onError) {
                options.onError(standardError)
            }
            return null
        }
    }) as T
}

// 预定义的错误处理器

/**
 * 任务领取错误处理器
 */
export const handleClaimError = createErrorHandler({
    defaultMessage: '领取任务失败，请稍后重试'
})

/**
 * 任务提交错误处理器
 */
export const handleSubmitError = createErrorHandler({
    defaultMessage: '提交审核失败，请稍后重试'
})

/**
 * 任务退回错误处理器
 */
export const handleReturnError = createErrorHandler({
    defaultMessage: '退回任务失败，请稍后重试'
})

/**
 * 数据加载错误处理器
 */
export const handleLoadError = createErrorHandler({
    defaultMessage: '加载数据失败，请刷新页面重试'
})
