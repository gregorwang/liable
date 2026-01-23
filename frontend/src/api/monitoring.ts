import request from './request'
import type { EndpointHealthResponse, MonitoringSummary } from '../types'

export function getMonitoringSummary(params?: { date?: string }) {
  return request.get<any, MonitoringSummary>('/admin/monitoring/summary', { params })
}

export function getMonitoringEndpoints(params?: { date?: string; limit?: number }) {
  return request.get<any, EndpointHealthResponse>('/admin/monitoring/endpoints', { params })
}
