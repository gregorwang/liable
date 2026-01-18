/**
 * 错误处理属性测试
 * 
 * **Property 6: 错误响应格式一致性**
 * **Validates: Requirements 6.1, 6.2, 6.4, 6.5**
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import fc from 'fast-check'
import {
    isApiError,
    isNetworkError,
    isTimeoutError,
    extractErrorMessage,
    normalizeError,
    handleApiError,
    createErrorHandler,
    withErrorHandling,
    handleClaimError,
    handleSubmitError,
    handleReturnError,
    handleLoadError
} from '../errorHandler'

// Mock ElMessage
vi.mock('element-plus', () => ({
    ElMessage: {
        error: vi.fn()
    }
}))

import { ElMessage } from 'element-plus'

describe('Error Handler - Property Tests', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    /**
     * Property 6.1: isApiError 应该正确识别 API 错误
     * For any object with response.data structure, isApiError should return true
     */
    it('Property 6.1: isApiError should correctly identify API errors', () => {
        fc.assert(
            fc.property(
                fc.string({ minLength: 1, maxLength: 100 }),
                fc.string({ minLength: 1, maxLength: 50 }),
                (errorMsg, code) => {
                    const apiError = {
                        response: {
                            data: {
                                error: errorMsg,
                                code: code
                            }
                        }
                    }

                    expect(isApiError(apiError)).toBe(true)
                    return true
                }
            ),
            { numRuns: 100 }
        )
    })

    /**
     * Property 6.2: isApiError 应该对非 API 错误返回 false
     * For any non-API error structure, isApiError should return false
     */
    it('Property 6.2: isApiError should return false for non-API errors', () => {
        const nonApiErrors = [
            null,
            undefined,
            'string error',
            123,
            new Error('test'),
            { message: 'test' },
            { response: null },
            { response: { data: null } }
        ]

        for (const error of nonApiErrors) {
            expect(isApiError(error)).toBeFalsy()
        }
    })

    /**
     * Property 6.3: isNetworkError 应该正确识别网络错误
     * For any error with network error codes, isNetworkError should return true
     */
    it('Property 6.3: isNetworkError should correctly identify network errors', () => {
        const networkErrors = [
            { code: 'ECONNABORTED' },
            { code: 'ERR_NETWORK' },
            { message: 'Network Error' }
        ]

        for (const error of networkErrors) {
            expect(isNetworkError(error)).toBe(true)
        }
    })

    /**
     * Property 6.4: isTimeoutError 应该正确识别超时错误
     * For any error with timeout indicators, isTimeoutError should return true
     */
    it('Property 6.4: isTimeoutError should correctly identify timeout errors', () => {
        const timeoutErrors = [
            { code: 'ECONNABORTED' },
            { message: 'timeout of 5000ms exceeded' },
            { message: 'Request timeout' }
        ]

        for (const error of timeoutErrors) {
            expect(isTimeoutError(error)).toBe(true)
        }
    })

    /**
     * Property 6.5: extractErrorMessage 应该从 API 错误中提取消息
     * For any API error with error message, extractErrorMessage should return that message
     */
    it('Property 6.5: extractErrorMessage should extract message from API errors', () => {
        fc.assert(
            fc.property(
                fc.string({ minLength: 1, maxLength: 200 }).filter(s => s.trim().length > 0),
                (errorMsg) => {
                    const apiError = {
                        response: {
                            data: {
                                error: errorMsg
                            }
                        }
                    }

                    const extracted = extractErrorMessage(apiError)
                    expect(extracted).toBe(errorMsg)
                    return true
                }
            ),
            { numRuns: 100 }
        )
    })

    /**
     * Property 6.6: extractErrorMessage 应该对 'cancel' 返回空字符串
     * For 'cancel' error, extractErrorMessage should return empty string
     */
    it('Property 6.6: extractErrorMessage should return empty string for cancel', () => {
        expect(extractErrorMessage('cancel')).toBe('')
    })

    /**
     * Property 6.7: extractErrorMessage 应该对未知错误返回默认消息
     * For any unknown error type, extractErrorMessage should return default message
     */
    it('Property 6.7: extractErrorMessage should return default message for unknown errors', () => {
        fc.assert(
            fc.property(
                fc.string({ minLength: 1, maxLength: 100 }),
                (defaultMsg) => {
                    const unknownError = { someField: 'value' }
                    const extracted = extractErrorMessage(unknownError, defaultMsg)
                    expect(extracted).toBe(defaultMsg)
                    return true
                }
            ),
            { numRuns: 100 }
        )
    })

    /**
     * Property 6.8: normalizeError 应该返回包含 message 的标准化错误
     * For any error, normalizeError should return an object with message field
     */
    it('Property 6.8: normalizeError should return standardized error with message', () => {
        fc.assert(
            fc.property(
                fc.string({ minLength: 1, maxLength: 200 }),
                (errorMsg) => {
                    const apiError = {
                        response: {
                            data: {
                                error: errorMsg,
                                code: 'TEST_ERROR'
                            }
                        }
                    }

                    const normalized = normalizeError(apiError)

                    expect(normalized.message).toBe(errorMsg)
                    expect(normalized.code).toBe('TEST_ERROR')
                    expect(normalized.originalError).toBe(apiError)

                    return true
                }
            ),
            { numRuns: 100 }
        )
    })

    /**
     * Property 6.9: handleApiError 应该显示错误消息（默认行为）
     * For any API error, handleApiError should call ElMessage.error by default
     */
    it('Property 6.9: handleApiError should show error message by default', () => {
        fc.assert(
            fc.property(
                fc.string({ minLength: 1, maxLength: 100 }).filter(s => s.trim().length > 0),
                (errorMsg) => {
                    vi.clearAllMocks()

                    const apiError = {
                        response: {
                            data: {
                                error: errorMsg
                            }
                        }
                    }

                    handleApiError(apiError)

                    expect(ElMessage.error).toHaveBeenCalledWith(errorMsg)
                    return true
                }
            ),
            { numRuns: 50 }
        )
    })

    /**
     * Property 6.10: handleApiError 应该在 showMessage=false 时不显示消息
     * For any error with showMessage=false, handleApiError should not call ElMessage.error
     */
    it('Property 6.10: handleApiError should not show message when showMessage=false', () => {
        vi.clearAllMocks()

        const apiError = {
            response: {
                data: {
                    error: 'Test error'
                }
            }
        }

        handleApiError(apiError, { showMessage: false })

        expect(ElMessage.error).not.toHaveBeenCalled()
    })

    /**
     * Property 6.11: handleApiError 应该对 'cancel' 不显示消息
     * For 'cancel' error, handleApiError should not show any message
     */
    it('Property 6.11: handleApiError should not show message for cancel', () => {
        vi.clearAllMocks()

        const result = handleApiError('cancel')

        expect(ElMessage.error).not.toHaveBeenCalled()
        expect(result.message).toBe('')
    })

    /**
     * Property 6.12: createErrorHandler 应该返回可调用的函数
     * For any default options, createErrorHandler should return a callable function
     */
    it('Property 6.12: createErrorHandler should return callable function', () => {
        fc.assert(
            fc.property(
                fc.string({ minLength: 1, maxLength: 100 }).filter(s => s.trim().length > 0),
                (defaultMsg) => {
                    const handler = createErrorHandler({ defaultMessage: defaultMsg })

                    expect(typeof handler).toBe('function')

                    vi.clearAllMocks()
                    const result = handler({ someError: true })

                    expect(result.message).toBe(defaultMsg)
                    return true
                }
            ),
            { numRuns: 50 }
        )
    })

    /**
     * Property 6.13: 预定义错误处理器应该有正确的默认消息
     * All predefined error handlers should have appropriate default messages
     */
    it('Property 6.13: predefined error handlers should have correct default messages', () => {
        vi.clearAllMocks()

        // Test handleClaimError
        handleClaimError({ someError: true })
        expect(ElMessage.error).toHaveBeenLastCalledWith('领取任务失败，请稍后重试')

        vi.clearAllMocks()

        // Test handleSubmitError
        handleSubmitError({ someError: true })
        expect(ElMessage.error).toHaveBeenLastCalledWith('提交审核失败，请稍后重试')

        vi.clearAllMocks()

        // Test handleReturnError
        handleReturnError({ someError: true })
        expect(ElMessage.error).toHaveBeenLastCalledWith('退回任务失败，请稍后重试')

        vi.clearAllMocks()

        // Test handleLoadError
        handleLoadError({ someError: true })
        expect(ElMessage.error).toHaveBeenLastCalledWith('加载数据失败，请刷新页面重试')
    })
})

describe('Error Handler - Unit Tests', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    it('should extract message from Error object', () => {
        const error = new Error('Test error message')
        const message = extractErrorMessage(error)
        expect(message).toBe('Test error message')
    })

    it('should extract message from string error', () => {
        const message = extractErrorMessage('String error')
        expect(message).toBe('String error')
    })

    it('should handle network error message', () => {
        const networkError = { code: 'ERR_NETWORK' }
        const message = extractErrorMessage(networkError)
        expect(message).toBe('网络连接失败，请检查网络后重试')
    })

    it('should handle timeout error message', () => {
        const timeoutError = { message: 'timeout of 5000ms exceeded' }
        const message = extractErrorMessage(timeoutError)
        expect(message).toBe('请求超时，请稍后重试')
    })

    it('withErrorHandling should wrap async function and handle errors', async () => {
        const mockFn = vi.fn().mockRejectedValue(new Error('Test error'))
        const onError = vi.fn()

        const wrappedFn = withErrorHandling(mockFn, {
            defaultMessage: 'Custom error',
            onError
        })

        const result = await wrappedFn()

        expect(result).toBeNull()
        expect(onError).toHaveBeenCalled()
    })

    it('withErrorHandling should return result on success', async () => {
        const mockFn = vi.fn().mockResolvedValue({ data: 'success' })

        const wrappedFn = withErrorHandling(mockFn)

        const result = await wrappedFn()

        expect(result).toEqual({ data: 'success' })
    })

    it('normalizeError should include details from API error', () => {
        const apiError = {
            response: {
                data: {
                    error: 'Test error',
                    code: 'TEST_CODE',
                    details: { field: 'value' }
                }
            }
        }

        const normalized = normalizeError(apiError)

        expect(normalized.message).toBe('Test error')
        expect(normalized.code).toBe('TEST_CODE')
        expect(normalized.details).toEqual({ field: 'value' })
    })
})
