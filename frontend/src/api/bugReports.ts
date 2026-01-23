import request from './request'
import type { BugReportListResponse } from '../types'

export function submitBugReport(formData: FormData) {
  return request.post<any, { message: string }>(
    '/bug-reports',
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }
  )
}

export function listBugReports(params: {
  page?: number
  page_size?: number
  start_time?: string
  end_time?: string
  user_id?: number
  username?: string
  keyword?: string
}) {
  return request.get<any, BugReportListResponse>('/admin/bug-reports', {
    params,
  })
}

export function exportBugReports(payload: {
  start_time?: string
  end_time?: string
  user_id?: number
  username?: string
  keyword?: string
  format: 'csv' | 'json'
}) {
  return request.post<any, Blob>('/admin/bug-reports/export', payload, {
    responseType: 'blob',
  })
}
