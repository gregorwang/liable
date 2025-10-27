import request from './request'
import type { 
  NotificationResponse, 
  CreateNotificationRequest, 
  NotificationStats, 
  NotificationListResponse 
} from '../types'

// Get unread notifications for current user
export const getUnreadNotifications = (limit = 20): Promise<NotificationListResponse> => {
  return request.get(`/notifications/unread?limit=${limit}`)
}

// Get unread notification count
export const getUnreadCount = (): Promise<NotificationStats> => {
  return request.get('/notifications/unread-count')
}

// Mark a notification as read
export const markAsRead = (notificationId: number): Promise<{ message: string }> => {
  return request.put(`/notifications/${notificationId}/read`)
}

// Get recent notifications for history page
export const getRecentNotifications = (limit = 20, offset = 0): Promise<NotificationListResponse> => {
  return request.get(`/notifications/recent?limit=${limit}&offset=${offset}`)
}

// Create a new notification (admin only)
export const createNotification = (data: CreateNotificationRequest): Promise<{ 
  message: string
  notification: NotificationResponse 
}> => {
  return request.post('/admin/notifications', data)
}
