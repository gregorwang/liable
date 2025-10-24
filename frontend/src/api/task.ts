import request from './request'
import type { TasksResponse, TagsResponse, ReviewResult } from '../types'

/**
 * Claim tasks
 */
export function claimTasks() {
  return request.post<any, TasksResponse>('/tasks/claim')
}

/**
 * Get my tasks
 */
export function getMyTasks() {
  return request.get<any, TasksResponse>('/tasks/my')
}

/**
 * Submit single review
 */
export function submitReview(review: ReviewResult) {
  return request.post<any, { message: string }>('/tasks/submit', review)
}

/**
 * Submit batch reviews
 */
export function submitBatchReviews(reviews: ReviewResult[]) {
  return request.post<any, { message: string; submitted: number }>('/tasks/submit-batch', {
    reviews,
  })
}

/**
 * Get active tags
 */
export function getTags() {
  return request.get<any, TagsResponse>('/tags')
}

