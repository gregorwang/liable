import request from './request'
import type { 
  SecondReviewTasksResponse, 
  SubmitSecondReviewRequest
} from '../types'

/**
 * Claim second review tasks
 */
export function claimSecondReviewTasks(count: number) {
  return request.post<any, SecondReviewTasksResponse>('/tasks/second-review/claim', { count })
}

/**
 * Get my second review tasks
 */
export function getMySecondReviewTasks() {
  return request.get<any, SecondReviewTasksResponse>('/tasks/second-review/my')
}

/**
 * Submit single second review
 */
export function submitSecondReview(review: SubmitSecondReviewRequest) {
  return request.post<any, { message: string }>('/tasks/second-review/submit', review)
}

/**
 * Submit batch second reviews
 */
export function submitBatchSecondReviews(reviews: SubmitSecondReviewRequest[]) {
  return request.post<any, { message: string; submitted: number }>('/tasks/second-review/submit-batch', {
    reviews,
  })
}

/**
 * Return second review tasks back to pool
 */
export function returnSecondReviewTasks(taskIds: number[]) {
  return request.post<any, { message: string; count: number }>('/tasks/second-review/return', {
    task_ids: taskIds,
  })
}
