import request from './request'
import type { TasksResponse, TagsResponse, ReviewResult, SearchTasksRequest, SearchTasksResponse } from '../types'

/**
 * Claim tasks with custom count (1-50)
 */
export function claimTasks(count: number) {
  return request.post<any, TasksResponse>('/tasks/claim', { count })
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
  return request.post<any, { message: string; count: number }>('/tasks/submit-batch', {
    reviews,
  })
}

/**
 * Return tasks back to pool
 */
export function returnTasks(taskIds: number[]) {
  return request.post<any, { message: string; count: number }>('/tasks/return', {
    task_ids: taskIds,
  })
}

/**
 * Get active tags
 */
export function getTags() {
  return request.get<any, TagsResponse>('/tags')
}

/**
 * Search tasks with filters
 */
export function searchTasks(params: SearchTasksRequest) {
  return request.get<any, SearchTasksResponse>('/tasks/search', { params })
}

