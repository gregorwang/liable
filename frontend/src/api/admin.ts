import request from './request'
import type {
  User,
  Tag,
  OverviewStats,
  HourlyStats,
  TagStats,
  ReviewerPerformance,
} from '../types'

/**
 * Get pending users
 */
export function getPendingUsers() {
  return request.get<any, { users: User[] }>('/admin/users')
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
 * Get overview statistics
 */
export function getOverviewStats() {
  return request.get<any, OverviewStats>('/admin/stats/overview')
}

/**
 * Get hourly statistics
 */
export function getHourlyStats(date: string) {
  return request.get<any, { stats: HourlyStats[] }>('/admin/stats/hourly', {
    params: { date },
  })
}

/**
 * Get tag statistics
 */
export function getTagStats() {
  return request.get<any, { stats: TagStats[] }>('/admin/stats/tags')
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

