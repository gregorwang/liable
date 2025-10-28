import request from './request'
import type { 
  QCTasksResponse, 
  QCStats, 
  SubmitQCRequest 
} from '../types'

/**
 * Claim quality check tasks with custom count (1-50)
 */
export function claimQCTasks(count: number) {
  return request.post<any, QCTasksResponse>('/tasks/quality-check/claim', { count })
}

/**
 * Get my quality check tasks
 */
export function getMyQCTasks() {
  return request.get<any, QCTasksResponse>('/tasks/quality-check/my')
}

/**
 * Submit single quality check review
 */
export function submitQCReview(review: SubmitQCRequest) {
  return request.post<any, { message: string }>('/tasks/quality-check/submit', review)
}

/**
 * Submit batch quality check reviews
 */
export function submitBatchQCReviews(reviews: SubmitQCRequest[]) {
  return request.post<any, { message: string; submitted: number }>('/tasks/quality-check/submit-batch', {
    reviews,
  })
}

/**
 * Return quality check tasks back to pool
 */
export function returnQCTasks(taskIds: number[]) {
  return request.post<any, { message: string; count: number }>('/tasks/quality-check/return', {
    task_ids: taskIds,
  })
}

/**
 * Get quality check statistics
 */
export function getQCStats() {
  return request.get<any, QCStats>('/tasks/quality-check/stats')
}
