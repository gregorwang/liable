import request from './request'
import type {
  AuditLogEntry,
  AuditLogExportListResponse,
  AuditLogExportRequest,
  AuditLogExportResponse,
  AuditLogQueryParams,
  AuditLogQueryResponse,
} from '../types'

export function getAuditLogs(params: AuditLogQueryParams) {
  return request.get<any, AuditLogQueryResponse>('/admin/audit-logs', { params })
}

export function getAuditLog(id: string) {
  return request.get<any, AuditLogEntry>(`/admin/audit-logs/${id}`)
}

export function exportAuditLogs(payload: AuditLogExportRequest) {
  return request.post<any, AuditLogExportResponse>('/admin/audit-logs/export', payload)
}

export function listAuditLogExports(params?: { page?: number; page_size?: number }) {
  return request.get<any, AuditLogExportListResponse>('/admin/audit-logs/exports', { params })
}
