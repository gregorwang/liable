import axios, { type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import { getToken, removeToken } from '../utils/auth'

// Create axios instance
const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// Request interceptor
request.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
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
      
      // Handle 401 Unauthorized
      if (status === 401) {
        removeToken()
        window.location.href = '/login'
        ElMessage.error('登录已过期，请重新登录')
        return Promise.reject(error)
      }
      
      // Handle other errors
      const message = data?.error || data?.message || '请求失败'
      ElMessage.error(message)
    } else {
      ElMessage.error('网络错误，请检查连接')
    }
    
    return Promise.reject(error)
  }
)

export default request

