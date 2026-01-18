import request from './request'
import { createTaskApiWithStats, QualityCheckApiConfig } from './taskApiFactory'
import type {
  QCTasksResponse,
  QCStats,
  SubmitQCRequest
} from '../types'

// 使用工厂函数创建基础 API
const baseApi = createTaskApiWithStats<any, SubmitQCRequest, QCStats>(QualityCheckApiConfig)

/**
 * Claim quality check tasks with custom count (1-50)
 */
export function claimQCTasks(count: number) {
  return baseApi.claimTasks(count) as Promise<QCTasksResponse>
}

/**
 * Get my quality check tasks
 */
export function getMyQCTasks() {
  return baseApi.getMyTasks() as Promise<QCTasksResponse>
}

/**
 * Submit single quality check review
 */
export function submitQCReview(review: SubmitQCRequest) {
  return baseApi.submitReview(review)
}

/**
 * Submit batch quality check reviews
 */
export function submitBatchQCReviews(reviews: SubmitQCRequest[]) {
  return baseApi.submitBatchReviews(reviews)
}

/**
 * Return quality check tasks back to pool
 */
export function returnQCTasks(taskIds: number[]) {
  return baseApi.returnTasks(taskIds)
}

/**
 * Get quality check statistics
 */
export function getQCStats() {
  return baseApi.getStats()
}
