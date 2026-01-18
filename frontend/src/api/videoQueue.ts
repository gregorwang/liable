import request from './request'
import type { VideoQueueTag } from '../types'

// Types for Video Queue Pool System

export type Pool = '100k' | '1m' | '10m'

export type ReviewDecision = 'push_next_pool' | 'natural_pool' | 'remove_violation'

export interface VideoQueueTask {
  id: number
  video_id: number
  pool: Pool
  reviewer_id: number | null
  status: 'pending' | 'in_progress' | 'completed'
  claimed_at: string | null
  completed_at: string | null
  created_at: string
  video?: {
    id: number
    video_key: string
    filename: string
    file_size: number
    duration: number | null
    upload_time: string | null
    video_url: string | null
    url_expires_at: string | null
    status: string
    created_at: string
    updated_at: string
  }
}

// Re-export VideoQueueTag from types
export type { VideoQueueTag }

export interface ClaimVideoQueueTasksRequest {
  count: number
}

export interface ClaimVideoQueueTasksResponse {
  tasks: VideoQueueTask[]
  count: number
}

export interface SubmitVideoQueueReviewRequest {
  task_id: number
  review_decision: ReviewDecision
  reason: string
  tags: string[]
}

export interface BatchSubmitVideoQueueReviewRequest {
  reviews: SubmitVideoQueueReviewRequest[]
}

export interface ReturnVideoQueueTasksRequest {
  task_ids: number[]
}

export interface GetVideoQueueTagsResponse {
  tags: VideoQueueTag[]
}

export interface VideoQueuePoolStats {
  pool: Pool
  total_tasks: number
  completed_tasks: number
  pending_tasks: number
  in_progress_tasks: number
  avg_process_time_minutes: number
}

// Video Queue Pool APIs

/**
 * 领取视频队列任务
 * @param pool 队列类型 ('100k' | '1m' | '10m')
 * @param data 领取请求
 */
export const claimVideoQueueTasks = (
  pool: Pool,
  data: ClaimVideoQueueTasksRequest
): Promise<ClaimVideoQueueTasksResponse> => {
  return request.post(`/video/${pool}/tasks/claim`, data)
}

/**
 * 获取我的视频队列任务
 * @param pool 队列类型 ('100k' | '1m' | '10m')
 */
export const getMyVideoQueueTasks = (pool: Pool): Promise<{ tasks: VideoQueueTask[], count: number }> => {
  return request.get(`/video/${pool}/tasks/my`)
}

/**
 * 提交单个视频队列审核
 * @param pool 队列类型 ('100k' | '1m' | '10m')
 * @param data 审核结果
 */
export const submitVideoQueueReview = (pool: Pool, data: SubmitVideoQueueReviewRequest): Promise<{ message: string }> => {
  return request.post(`/video/${pool}/tasks/submit`, data)
}

/**
 * 批量提交视频队列审核
 * @param pool 队列类型 ('100k' | '1m' | '10m')
 * @param data 批量审核结果
 */
export const submitBatchVideoQueueReviews = (
  pool: Pool,
  data: BatchSubmitVideoQueueReviewRequest
): Promise<{ message: string, count: number }> => {
  return request.post(`/video/${pool}/tasks/submit-batch`, data)
}

/**
 * 归还视频队列任务
 * @param pool 队列类型 ('100k' | '1m' | '10m')
 * @param data 归还请求
 */
export const returnVideoQueueTasks = (
  pool: Pool,
  data: ReturnVideoQueueTasksRequest
): Promise<{ message: string, count: number }> => {
  return request.post(`/video/${pool}/tasks/return`, data)
}

/**
 * 获取视频队列标签
 * @param pool 队列类型 ('100k' | '1m' | '10m')
 */
export const getVideoQueueTags = (pool: Pool): Promise<GetVideoQueueTagsResponse> => {
  return request.get(`/video/${pool}/tags`)
}

/**
 * 获取视频队列统计 (管理员)
 * @param pool 队列类型 ('100k' | '1m' | '10m')
 */
export const getVideoQueuePoolStats = (pool: Pool): Promise<VideoQueuePoolStats> => {
  return request.get(`/admin/video-queue/${pool}/stats`)
}

/**
 * 生成视频 URL (复用原有 API)
 */
export const generateVideoURL = (data: { video_id: number }): Promise<{ video_url: string, expires_at: string }> => {
  return request.post('/admin/videos/generate-url', data)
}
