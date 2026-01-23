import request from './request'
import type { ProfileResponse } from '../types'

export function updateProfile(payload: { gender?: string | null; signature?: string | null }) {
  return request.put<any, ProfileResponse>('/auth/profile', payload)
}

export function updateSystemProfile(payload: {
  office_location?: string | null
  department?: string | null
  school?: string | null
  company?: string | null
  direct_manager?: string | null
}) {
  return request.put<any, ProfileResponse>('/auth/profile/system', payload)
}

export function uploadAvatar(formData: FormData) {
  return request.post<any, ProfileResponse>('/auth/profile/avatar', formData)
}
