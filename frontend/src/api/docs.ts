import request from './request'
import type { SystemDocument } from '../types'

export interface SystemDocumentsResponse {
  data: SystemDocument[]
  can_edit: boolean
}

export function listSystemDocuments() {
  return request.get<any, SystemDocumentsResponse>('/docs')
}

export function updateSystemDocument(key: string, content: string) {
  return request.put<any, SystemDocument>(`/admin/docs/${key}`, { content })
}
