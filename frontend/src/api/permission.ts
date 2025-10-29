import request from './request'

export interface Permission {
  id: number
  permission_key: string
  name: string
  description: string
  resource: string
  action: string
  category: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface ListPermissionsParams {
  resource?: string
  category?: string
  search?: string
  page?: number
  page_size?: number
}

export interface ListPermissionsResponse {
  data: Permission[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface UserPermissionsResponse {
  user_id: number
  permissions: string[]
}

export interface GrantPermissionRequest {
  user_id: number
  permission_keys: string[]
}

export interface RevokePermissionRequest {
  user_id: number
  permission_keys: string[]
}

// List permissions with pagination and filtering
export function listPermissions(params?: ListPermissionsParams) {
  return request.get<ListPermissionsResponse>('/admin/permissions', { params })
}

// Get all permissions (no pagination)
export function getAllPermissions() {
  return request.get<{ permissions: Permission[] }>('/admin/permissions/all')
}

// Get user permissions
export function getUserPermissions(userId: number) {
  return request.get<UserPermissionsResponse>('/admin/permissions/user', {
    params: { user_id: userId }
  })
}

// Grant permissions to a user
export function grantPermissions(data: GrantPermissionRequest) {
  return request.post('/admin/permissions/grant', data)
}

// Revoke permissions from a user
export function revokePermissions(data: RevokePermissionRequest) {
  return request.post('/admin/permissions/revoke', data)
}

