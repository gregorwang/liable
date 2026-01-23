import request from './request'
import type {
  ImportVideosRequest,
  ImportVideosResponse,
  ListVideosRequest,
  ListVideosResponse,
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
