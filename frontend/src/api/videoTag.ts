import request from './request'
import type { VideoQueueTag } from '../types'

// Re-export VideoQueueTag for convenience
export type { VideoQueueTag }

export interface CreateVideoTagRequest {
  name: string
  description: string
  category: string
  scope?: string
  queue_id?: string | null
}

export interface UpdateVideoTagRequest {
  name?: string
  description?: string
  category?: string
  scope?: string
  queue_id?: string | null
  is_active?: boolean
}

/**
 * 获取所有视频标签
 */
export const getAllVideoTags = (scope?: string): Promise<{ tags: VideoQueueTag[] }> => {
  return request.get('/admin/video-tags', { params: { scope } })
}

/**
 * 创建视频标签
 */
export const createVideoTag = (data: CreateVideoTagRequest): Promise<{ message: string, tag: VideoQueueTag }> => {
  return request.post('/admin/video-tags', data)
}

/**
 * 更新视频标签
 */
export const updateVideoTag = (id: number, data: UpdateVideoTagRequest): Promise<{ message: string }> => {
  return request.put(`/admin/video-tags/${id}`, data)
}

/**
 * 删除视频标签
 */
export const deleteVideoTag = (id: number): Promise<{ message: string }> => {
  return request.delete(`/admin/video-tags/${id}`)
}

/**
 * 切换视频标签状态
 */
export const toggleVideoTagActive = (id: number): Promise<{ message: string }> => {
  return request.patch(`/admin/video-tags/${id}/toggle`)
}
