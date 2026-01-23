import request from './request'
import type {
  AIReviewJob,
  CreateAIReviewJobRequest,
  ListAIReviewJobsResponse,
  ListAIReviewTasksResponse,
  AIReviewComparisonResponse,
} from '../types'

export function createAIReviewJob(payload: CreateAIReviewJobRequest) {
  return request.post<any, AIReviewJob>('/admin/ai-review/jobs', payload)
}

export function startAIReviewJob(jobId: number) {
  return request.post<any, { message: string }>(`/admin/ai-review/jobs/${jobId}/start`)
}

export function listAIReviewJobs(params?: {
  page?: number
  page_size?: number
  include_archived?: boolean
}) {
  return request.get<any, ListAIReviewJobsResponse>('/admin/ai-review/jobs', { params })
}

export function getAIReviewJob(jobId: number) {
  return request.get<any, AIReviewJob>(`/admin/ai-review/jobs/${jobId}`)
}

export function listAIReviewTasks(
  jobId: number,
  params?: {
    page?: number
    page_size?: number
  },
) {
  return request.get<any, ListAIReviewTasksResponse>(`/admin/ai-review/jobs/${jobId}/tasks`, { params })
}

export function deleteAIReviewJobTasks(jobId: number) {
  return request.delete<any, { deleted: number }>(`/admin/ai-review/jobs/${jobId}/tasks`)
}

export function archiveAIReviewJob(jobId: number) {
  return request.post<any, { message: string }>(`/admin/ai-review/jobs/${jobId}/archive`)
}

export function unarchiveAIReviewJob(jobId: number) {
  return request.post<any, { message: string }>(`/admin/ai-review/jobs/${jobId}/unarchive`)
}

export function getAIReviewComparison(params?: {
  job_id?: number
  limit?: number
}) {
  return request.get<any, AIReviewComparisonResponse>('/admin/ai-review/compare', { params })
}
