import { ref, onMounted, onUnmounted, type Ref } from 'vue'
import { ElNotification } from 'element-plus'
import { getToken } from '../utils/auth'
import type { SSEMessageData, NotificationResponse } from '../types'

export interface UseSSEReturn {
  isConnected: Ref<boolean>
  error: Ref<string | null>
  connect: () => void
  disconnect: () => void
  onMessage: (callback: (data: SSEMessageData) => void) => void
}

export function useSSE(): UseSSEReturn {
  const isConnected = ref(false)
  const error = ref<string | null>(null)
  const eventSource = ref<EventSource | null>(null)
  const messageCallbacks = ref<((data: SSEMessageData) => void)[]>([])

  const connect = () => {
    if (eventSource.value) {
      disconnect()
    }

    const token = getToken()
    if (!token) {
      error.value = 'No authentication token found'
      return
    }

    try {
      const url = `/api/notifications/stream?token=${encodeURIComponent(token)}`
      eventSource.value = new EventSource(url)

      eventSource.value.onopen = () => {
        isConnected.value = true
        error.value = null
        // SSE connection established
      }

      eventSource.value.onmessage = (event) => {
        try {
          const data: SSEMessageData = JSON.parse(event.data)
          
          // Handle different message types
          if (data.type === 'notification') {
            const notification = data.data as NotificationResponse
            
            // Show browser notification
            ElNotification({
              title: notification.title,
              message: notification.content,
              type: notification.type === 'error' ? 'error' : 
                    notification.type === 'warning' ? 'warning' : 
                    notification.type === 'success' ? 'success' : 'info',
              duration: 5000,
              position: 'top-right',
              showClose: true,
            })
          } else if (data.type === 'heartbeat') {
            // Heartbeat received, connection is alive
            // SSE heartbeat received
          } else if (data.type === 'connection') {
            // SSE connection confirmed
          }

          // Call registered callbacks
          messageCallbacks.value.forEach(callback => {
            try {
              callback(data)
            } catch (err) {
              console.error('Error in SSE message callback:', err)
            }
          })
        } catch (err) {
          console.error('Error parsing SSE message:', err)
        }
      }

      eventSource.value.onerror = (event) => {
        console.error('SSE connection error:', event)
        isConnected.value = false
        error.value = 'Connection error'

        // Attempt to reconnect after 30 seconds (reduced from 5s to avoid spam)
        setTimeout(() => {
          if (!isConnected.value) {
            console.log('Attempting to reconnect SSE...')
            connect()
          }
        }, 30000)
      }

    } catch (err) {
      console.error('Failed to create SSE connection:', err)
      error.value = 'Failed to create connection'
    }
  }

  const disconnect = () => {
    if (eventSource.value) {
      eventSource.value.close()
      eventSource.value = null
    }
    isConnected.value = false
    error.value = null
    // SSE connection closed
  }

  const onMessage = (callback: (data: SSEMessageData) => void) => {
    messageCallbacks.value.push(callback)
    
    // Return cleanup function
    return () => {
      const index = messageCallbacks.value.indexOf(callback)
      if (index > -1) {
        messageCallbacks.value.splice(index, 1)
      }
    }
  }

  // Auto-connect on mount
  onMounted(() => {
    connect()
  })

  // Auto-disconnect on unmount
  onUnmounted(() => {
    disconnect()
  })

  return {
    isConnected,
    error,
    connect,
    disconnect,
    onMessage
  }
}
