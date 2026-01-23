import request from './request'
import type { LoginResponse, RegisterResponse, ProfileResponse } from '../types'

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
  return request.get<any, ProfileResponse>('/auth/profile')
}

/**
 * 发送验证码
 */
export function sendVerificationCode(email: string, purpose: 'login' | 'register') {
  return request.post<any, { message: string; expires_in: number }>('/auth/send-code', {
    email,
    purpose,
  })
}

/**
 * 验证码登录
 */
export function loginWithCode(email: string, code: string) {
  return request.post<any, LoginResponse>('/auth/login-with-code', {
    email,
    code,
  })
}

/**
 * 验证码注册
 */
export function registerWithCode(email: string, code: string, username: string) {
  return request.post<any, RegisterResponse>('/auth/register-with-code', {
    email,
    code,
    username,
  })
}

/**
 * 检查邮箱是否已注册
 */
export function checkEmail(email: string) {
  return request.get<any, { exists: boolean; email: string }>('/auth/check-email', {
    params: { email },
  })
}

