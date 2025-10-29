import request from './request'
import type {
  ImportVideosRequest,
  ImportVideosResponse,
  ListVideosRequest,
  ListVideosResponse,
  ClaimVideoFirstReviewTasksRequest,
  ClaimVideoFirstReviewTasksResponse,
  VideoFirstReviewTask,
  SubmitVideoFirstReviewRequest,
  BatchSubmitVideoFirstReviewRequest,
  ReturnVideoFirstReviewTasksRequest,
  ClaimVideoSecondReviewTasksRequest,
  ClaimVideoSecondReviewTasksResponse,
  VideoSecondReviewTask,
  SubmitVideoSecondReviewRequest,
  BatchSubmitVideoSecondReviewRequest,
  ReturnVideoSecondReviewTasksRequest,
  GetVideoQualityTagsRequest,
  GetVideoQualityTagsResponse,
  GenerateVideoURLRequest,
  GenerateVideoURLResponse,
  TikTokVideo
} from '@/types'

// Admin video management APIs

export const importVideos = (data: ImportVideosRequest): Promise<ImportVideosResponse> => {
  return request.post('/admin/videos/import', data)
}

export const listVideos = (params?: ListVideosRequest): Promise<ListVideosResponse> => {
  return request.get('/admin/videos', { params })
}

export const getVideo = (id: number): Promise<TikTokVideo> => {
  return request.get(`/admin/videos/${id}`)
}

export const generateVideoURL = (data: GenerateVideoURLRequest): Promise<GenerateVideoURLResponse> => {
  return request.post('/admin/videos/generate-url', data)
}

// Video first review APIs

export const claimVideoFirstReviewTasks = (data: ClaimVideoFirstReviewTasksRequest): Promise<ClaimVideoFirstReviewTasksResponse> => {
  return request.post('/tasks/video-first-review/claim', data)
}

export const getMyVideoFirstReviewTasks = (): Promise<{ tasks: VideoFirstReviewTask[], count: number }> => {
  return request.get('/tasks/video-first-review/my')
}

export const submitVideoFirstReview = (data: SubmitVideoFirstReviewRequest): Promise<void> => {
  return request.post('/tasks/video-first-review/submit', data)
}

export const submitBatchVideoFirstReviews = (data: BatchSubmitVideoFirstReviewRequest): Promise<void> => {
  return request.post('/tasks/video-first-review/submit-batch', data)
}

export const returnVideoFirstReviewTasks = (data: ReturnVideoFirstReviewTasksRequest): Promise<void> => {
  return request.post('/tasks/video-first-review/return', data)
}

// Video second review APIs

export const claimVideoSecondReviewTasks = (data: ClaimVideoSecondReviewTasksRequest): Promise<ClaimVideoSecondReviewTasksResponse> => {
  return request.post('/tasks/video-second-review/claim', data)
}

export const getMyVideoSecondReviewTasks = (): Promise<{ tasks: VideoSecondReviewTask[], count: number }> => {
  return request.get('/tasks/video-second-review/my')
}

export const submitVideoSecondReview = (data: SubmitVideoSecondReviewRequest): Promise<void> => {
  return request.post('/tasks/video-second-review/submit', data)
}

export const submitBatchVideoSecondReviews = (data: BatchSubmitVideoSecondReviewRequest): Promise<void> => {
  return request.post('/tasks/video-second-review/submit-batch', data)
}

export const returnVideoSecondReviewTasks = (data: ReturnVideoSecondReviewTasksRequest): Promise<void> => {
  return request.post('/tasks/video-second-review/return', data)
}

// Video quality tags API

export const getVideoQualityTags = (params?: GetVideoQualityTagsRequest): Promise<GetVideoQualityTagsResponse> => {
  return request.get('/video-quality-tags', { params })
}
