import request from './request'
import type { LoginResponse, RegisterResponse, User } from '../types'

/**
 * Login
 */
export function login(username: string, password: string) {
  return request.post<any, LoginResponse>('/auth/login', {
    username,
    password,
  })
}

/**
 * Register
 */
export function register(username: string, password: string) {
  return request.post<any, RegisterResponse>('/auth/register', {
    username,
    password,
  })
}

/**
 * Get current user profile
 */
export function getProfile() {
  return request.get<any, { user: User }>('/auth/profile')
}

