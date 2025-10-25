import request from './request'
import type { ModerationRule, ListModerationRulesResponse } from '../types'

/**
 * List all moderation rules with filtering and pagination
 */
export function listRules(params?: {
  category?: string
  risk_level?: string
  search?: string
  page?: number
  page_size?: number
}) {
  return request.get<any, ListModerationRulesResponse>('/moderation-rules', {
    params,
  })
}

/**
 * Get ALL moderation rules without pagination
 */
export function getAllRules() {
  return request.get<any, ListModerationRulesResponse>('/moderation-rules/all')
}

/**
 * Get rule categories
 */
export function getCategories() {
  return request.get<any, { categories: string[] }>('/moderation-rules/categories')
}

/**
 * Get risk levels
 */
export function getRiskLevels() {
  return request.get<any, { levels: string[] }>('/moderation-rules/risk-levels')
}

/**
 * Create a new moderation rule (admin only)
 */
export function createRule(rule: ModerationRule) {
  return request.post<any, ModerationRule>('/admin/moderation-rules', rule)
}

/**
 * Update an existing moderation rule (admin only)
 */
export function updateRule(id: number, rule: ModerationRule) {
  return request.put<any, ModerationRule>(`/admin/moderation-rules/${id}`, rule)
}

/**
 * Delete a moderation rule (admin only)
 */
export function deleteRule(id: number) {
  return request.delete<any, { message: string }>(`/admin/moderation-rules/${id}`)
}
