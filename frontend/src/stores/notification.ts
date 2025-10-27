import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  getUnreadNotifications, 
  getUnreadCount, 
  markAsRead, 
  getRecentNotifications 
} from '../api/notification'
import { useSSE } from '../composables/useSSE'
import type { NotificationResponse, SSEMessageData } from '../types'

export const useNotificationStore = defineStore('notification', () => {
  // State
  const notifications = ref<NotificationResponse[]>([])
  const unreadCount = ref(0)
  const isLoading = ref(false)
  const isConnected = ref(false)
  const error = ref<string | null>(null)

  // SSE connection
  const { connect: connectSSE, disconnect: disconnectSSE, onMessage } = useSSE()

  // Computed
  const unreadNotifications = computed(() => 
    notifications.value.filter(n => !n.is_read)
  )

  const recentNotifications = computed(() => 
    notifications.value.slice(0, 10) // Show recent 10 notifications
  )

  // Actions
  const fetchUnread = async (limit = 20) => {
    try {
      isLoading.value = true
      const response = await getUnreadNotifications(limit)
      notifications.value = response.notifications
    } catch (err) {
      console.error('Failed to fetch unread notifications:', err)
      ElMessage.error('获取未读通知失败')
    } finally {
      isLoading.value = false
    }
  }

  const fetchUnreadCount = async () => {
    try {
      const response = await getUnreadCount()
      unreadCount.value = response.count
    } catch (err) {
      console.error('Failed to fetch unread count:', err)
    }
  }

  const markNotificationAsRead = async (notificationId: number) => {
    try {
      await markAsRead(notificationId)
      
      // Update local state
      const notification = notifications.value.find(n => n.id === notificationId)
      if (notification) {
        notification.is_read = true
        notification.read_at = new Date().toISOString()
      }
      
      // Update unread count
      unreadCount.value = Math.max(0, unreadCount.value - 1)
      
      ElMessage.success('已标记为已读')
    } catch (err) {
      console.error('Failed to mark notification as read:', err)
      ElMessage.error('标记已读失败')
    }
  }

  const addNotification = (notification: NotificationResponse) => {
    // Check if notification already exists
    const existingIndex = notifications.value.findIndex(n => n.id === notification.id)
    
    if (existingIndex >= 0) {
      // Update existing notification
      notifications.value[existingIndex] = notification
    } else {
      // Add new notification to the beginning
      notifications.value.unshift(notification)
    }
    
    // Update unread count if it's unread
    if (!notification.is_read) {
      unreadCount.value += 1
    }
  }

  const fetchRecent = async (limit = 20, offset = 0) => {
    try {
      isLoading.value = true
      const response = await getRecentNotifications(limit, offset)
      
      if (offset === 0) {
        // Replace notifications for first page
        notifications.value = response.notifications
      } else {
        // Append notifications for pagination
        notifications.value.push(...response.notifications)
      }
      
      return response
    } catch (err) {
      console.error('Failed to fetch recent notifications:', err)
      ElMessage.error('获取历史通知失败')
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const initSSE = () => {
    // Set up SSE message handler
    onMessage((data: SSEMessageData) => {
      if (data.type === 'notification') {
        const notification = data.data as NotificationResponse
        addNotification(notification)
      }
    })
    
    // Connect to SSE
    connectSSE()
    isConnected.value = true
  }

  const closeSSE = () => {
    disconnectSSE()
    isConnected.value = false
  }

  const init = async () => {
    try {
      // Initialize SSE connection
      initSSE()
      
      // Fetch initial data
      await Promise.all([
        fetchUnread(),
        fetchUnreadCount()
      ])
    } catch (err) {
      console.error('Failed to initialize notification store:', err)
    }
  }

  const clearError = () => {
    error.value = null
  }

  return {
    // State
    notifications,
    unreadCount,
    isLoading,
    isConnected,
    error,
    
    // Computed
    unreadNotifications,
    recentNotifications,
    
    // Actions
    fetchUnread,
    fetchUnreadCount,
    markNotificationAsRead,
    addNotification,
    fetchRecent,
    initSSE,
    closeSSE,
    init,
    clearError
  }
})
