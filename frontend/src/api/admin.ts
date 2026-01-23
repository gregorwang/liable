import request from './request'
import type {
  User,
  Tag,
  OverviewStats,
  HourlyStats,
  TagStats,
  ReviewerPerformance,
  TaskQueue,
  ListTaskQueuesResponse,
  TodayReviewStats,
} from '../types'

/**
 * Get pending users
 */
export function getPendingUsers() {
  return request.get<any, { users: User[] }>('/admin/users')
}

/**
 * Get all users (for permission management)
 */
export function getAllUsers() {
  return request.get<any, { users: User[] }>('/admin/users/all')
}

/**
 * Approve or reject user
 */
export function approveUser(userId: number, status: 'approved' | 'rejected') {
  return request.put<any, { message: string }>(`/admin/users/${userId}/approve`, {
    status,
  })
}

/**
 * Create user
 */
export function createUser(payload: {
  username: string
  email?: string
  password?: string
  role?: 'admin' | 'reviewer'
  status?: 'pending' | 'approved' | 'rejected'
}) {
  return request.post<any, { user: User }>('/admin/users', payload)
}

/**
 * Delete user
 */
export function deleteUser(userId: number) {
  return request.delete<any, { message: string }>(`/admin/users/${userId}`)
}

/**
 * Get overview statistics
 */
export function getOverviewStats() {
  return request.get<any, OverviewStats>('/admin/stats/overview')
}

/**
 * Get today's review statistics
 */
export function getTodayReviewStats() {
  return request.get<any, TodayReviewStats>('/admin/stats/today')
}

/**
 * Get hourly statistics
 */
export function getHourlyStats(date: string) {
  return request.get<any, HourlyStats>('/admin/stats/hourly', {
    params: { date },
  })
}

/**
 * Get tag statistics
 */
export function getTagStats() {
  return request.get<any, { tags: TagStats[] }>('/admin/stats/tags')
}

/**
 * Get reviewer performance
 */
export function getReviewerPerformance(limit: number = 10) {
  return request.get<any, { reviewers: ReviewerPerformance[] }>('/admin/stats/reviewers', {
    params: { limit },
  })
}

/**
 * Get all tags (admin)
 */
export function getAllTags() {
  return request.get<any, { tags: Tag[] }>('/admin/tags')
}

/**
 * Create tag
 */
export function createTag(name: string, description: string) {
  return request.post<any, { message: string; tag: Tag }>('/admin/tags', {
    name,
    description,
  })
}

/**
 * Update tag
 */
export function updateTag(tagId: number, data: Partial<Tag>) {
  return request.put<any, { message: string }>(`/admin/tags/${tagId}`, data)
}

/**
 * Delete tag
 */
export function deleteTag(tagId: number) {
  return request.delete<any, { message: string }>(`/admin/tags/${tagId}`)
}

// ==================== Task Queue Management ====================

// Create task queue
export async function createTaskQueue(payload: {
  queue_name: string
  description?: string
  priority?: number
  total_tasks: number
  completed_tasks?: number
}): Promise<TaskQueue> {
  const response = await request.post('/admin/task-queues', payload)
  return response as unknown as TaskQueue
}

// List task queues with pagination
export async function listTaskQueues(params?: {
  search?: string
  is_active?: boolean
  page?: number
  page_size?: number
}): Promise<ListTaskQueuesResponse> {
  const response = await request.get('/admin/task-queues', { params })
  return response as unknown as ListTaskQueuesResponse
}

// Get single task queue by ID
export async function getTaskQueue(id: number): Promise<TaskQueue> {
  const response = await request.get(`/admin/task-queues/${id}`)
  return response as unknown as TaskQueue
}

// Update task queue
export async function updateTaskQueue(
  id: number,
  payload: {
    queue_name?: string
    description?: string
    priority?: number
    total_tasks?: number
    completed_tasks?: number
    is_active?: boolean
  }
): Promise<TaskQueue> {
  const response = await request.put(`/admin/task-queues/${id}`, payload)
  return response as unknown as TaskQueue
}

// Delete task queue
export async function deleteTaskQueue(id: number): Promise<{ message: string }> {
  const response = await request.delete(`/admin/task-queues/${id}`)
  return response as unknown as { message: string }
}

// Get all active task queues
export async function getAllTaskQueues(): Promise<{
  queues: TaskQueue[]
  count: number
}> {
  const response = await request.get('/admin/task-queues-all')
  return response as unknown as { queues: TaskQueue[]; count: number }
}

// ==================== Public Task Queue APIs (for Reviewers) ====================

/**
 * List task queues with pagination (public, no auth required)
 */
export async function listTaskQueuesPublic(params?: {
  search?: string
  is_active?: boolean
  page?: number
  page_size?: number
}): Promise<ListTaskQueuesResponse> {
  const response = await request.get('/queues', { params })
  return response as unknown as ListTaskQueuesResponse
}

/**
 * Get single task queue by ID (public, no auth required)
 */
export async function getTaskQueuePublic(id: number): Promise<TaskQueue> {
  const response = await request.get(`/queues/${id}`)
  return response as unknown as TaskQueue
}
