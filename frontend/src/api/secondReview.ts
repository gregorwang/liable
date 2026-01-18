import { createTaskApi, SecondReviewApiConfig } from './taskApiFactory'
import type {
  SecondReviewTasksResponse,
  SubmitSecondReviewRequest
} from '../types'

// 使用工厂函数创建基础 API
const baseApi = createTaskApi<any, SubmitSecondReviewRequest>(SecondReviewApiConfig)

/**
 * Claim second review tasks
 */
export function claimSecondReviewTasks(count: number) {
  return baseApi.claimTasks(count) as Promise<SecondReviewTasksResponse>
}

/**
 * Get my second review tasks
 */
export function getMySecondReviewTasks() {
  return baseApi.getMyTasks() as Promise<SecondReviewTasksResponse>
}

/**
 * Submit single second review
 */
export function submitSecondReview(review: SubmitSecondReviewRequest) {
  return baseApi.submitReview(review)
}

/**
 * Submit batch second reviews
 */
export function submitBatchSecondReviews(reviews: SubmitSecondReviewRequest[]) {
  return baseApi.submitBatchReviews(reviews)
}

/**
 * Return second review tasks back to pool
 */
export function returnSecondReviewTasks(taskIds: number[]) {
  return baseApi.returnTasks(taskIds)
}
