import axios, { type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import { getToken, removeToken } from '../utils/auth'
import { createTraceId } from '../utils/trace'
import { buildTraceMessage } from '../utils/traceNotice'

// Create axios instance
const request = axios.create({
  baseURL: '/api',
  timeout: 600000, // 10 minutes - Redis operations can take time
})

// Request interceptor
request.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    if (typeof window !== 'undefined') {
      config.headers['X-Page-Url'] = window.location.href
    }
    const traceId = createTraceId()
    config.headers['X-Trace-Id'] = traceId
    ;(config as any)._traceId = traceId
    return config
  },
  (error) => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// Response interceptor
request.interceptors.response.use(
  (response: AxiosResponse) => {
    return response.data
  },
  (error) => {
    console.error('Response error:', error)

    if (error.response) {
      const { status, data } = error.response

      // Handle 429 Too Many Requests - 限流保护
      if (status === 429) {
        const retryAfter = data?.retry_after || '请稍后'
        const message = data?.error || '请求过于频繁'
        ElMessage.warning(buildTraceMessage(`${message}，${retryAfter}后再试`, error))
        return Promise.reject(error)
      }

      // Handle 401 Unauthorized
      if (status === 401) {
        removeToken()

        // 区分token过期和被强制登出
        const message = data?.error || '登录已过期，请重新登录'
        if (message.includes('blacklisted') || message.includes('revoked')) {
          ElMessage.error(buildTraceMessage('登录已失效，请重新登录', error))
        } else if (message.includes('force logout') || message.includes('logged out')) {
          ElMessage.error(buildTraceMessage('您的账号已在其他设备登录，请重新登录', error))
        } else {
          ElMessage.error(buildTraceMessage('登录已过期，请重新登录', error))
        }

        window.location.href = '/login'
        return Promise.reject(error)
      }

      // Handle 403 Forbidden - 权限不足
      if (status === 403) {
        console.warn('Permission denied:', data)
        const message = data?.error || '您没有权限执行此操作'

        // 显示更详细的权限错误信息
        if (data?.required_permission) {
          ElMessage.warning(buildTraceMessage(`缺少权限：${data.required_permission}`, error))
        } else if (data?.required_permissions) {
          ElMessage.warning(buildTraceMessage(`需要以下权限之一：${data.required_permissions.join(', ')}`, error))
        } else {
          ElMessage.warning(buildTraceMessage(message, error))
        }
        return Promise.reject(error)
      }

      // Handle other errors
      const message = data?.error || data?.message || '请求失败'
      ElMessage.error(buildTraceMessage(message, error))
    } else {
      ElMessage.error(buildTraceMessage('网络错误，请检查连接', error))
    }

    return Promise.reject(error)
  }
)

export default request
