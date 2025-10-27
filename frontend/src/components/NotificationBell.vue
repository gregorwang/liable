<template>
  <el-dropdown 
    trigger="click" 
    placement="bottom-end"
    @command="handleCommand"
    class="notification-dropdown"
  >
    <el-badge :value="notificationStore.unreadCount" :hidden="notificationStore.unreadCount === 0" class="notification-badge">
      <el-button type="text" class="notification-btn">
        <el-icon size="18">
          <Bell />
        </el-icon>
        <span v-if="notificationStore.isConnected" class="connection-indicator connected" title="实时连接已建立"></span>
        <span v-else class="connection-indicator disconnected" title="连接断开"></span>
      </el-button>
    </el-badge>
    
    <template #dropdown>
      <el-dropdown-menu class="notification-menu">
        <div class="notification-header">
          <span class="header-title">通知</span>
          <el-button 
            type="text" 
            size="small" 
            @click="goToHistory"
            class="view-all-btn"
          >
            查看全部
          </el-button>
        </div>
        
        <el-divider style="margin: 8px 0;" />
        
        <div v-if="notificationStore.isLoading" class="loading-container">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>加载中...</span>
        </div>
        
        <div v-else-if="notificationStore.recentNotifications.length === 0" class="empty-container">
          <el-icon><Bell /></el-icon>
          <span>暂无通知</span>
        </div>
        
        <div v-else class="notification-list">
          <div 
            v-for="notification in notificationStore.recentNotifications" 
            :key="notification.id"
            class="notification-item"
            :class="{ 'unread': !notification.is_read }"
            @click="handleNotificationClick(notification)"
          >
            <div class="notification-content">
              <div class="notification-title">
                <el-icon 
                  :class="getNotificationIconClass(notification.type)"
                  :size="14"
                >
                  <component :is="getNotificationIcon(notification.type)" />
                </el-icon>
                {{ notification.title }}
              </div>
              <div class="notification-text">{{ notification.content }}</div>
              <div class="notification-time">{{ formatTime(notification.created_at) }}</div>
            </div>
            <div v-if="!notification.is_read" class="unread-dot"></div>
          </div>
        </div>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Bell, Loading, InfoFilled, WarningFilled, SuccessFilled, CircleCloseFilled } from '@element-plus/icons-vue'
import { useNotificationStore } from '../stores/notification'
import type { NotificationResponse, NotificationType } from '../types'

const router = useRouter()
const notificationStore = useNotificationStore()

onMounted(() => {
  // Initialize notification store if not already done
  if (!notificationStore.isConnected) {
    notificationStore.init()
  }
})

const handleCommand = (_command: string) => {
  // Handle dropdown commands if needed
}

const handleNotificationClick = async (notification: NotificationResponse) => {
  if (!notification.is_read) {
    await notificationStore.markNotificationAsRead(notification.id)
  }
}

const goToHistory = () => {
  router.push('/main/history-announcements')
}

const getNotificationIcon = (type: NotificationType) => {
  switch (type) {
    case 'success':
      return SuccessFilled
    case 'warning':
      return WarningFilled
    case 'error':
      return CircleCloseFilled
    case 'info':
    case 'system':
    case 'announcement':
    case 'task_update':
    default:
      return InfoFilled
  }
}

const getNotificationIconClass = (type: NotificationType) => {
  return `notification-icon notification-icon--${type}`
}

const formatTime = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  const minutes = Math.floor(diff / (1000 * 60))
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`
  
  return date.toLocaleDateString()
}
</script>

<style scoped>
.notification-dropdown {
  margin-right: var(--spacing-2);
}

.notification-badge {
  margin-right: 0;
}

.notification-btn {
  font-size: var(--text-lg);
  color: var(--color-text-300);
  transition: color var(--transition-fast);
  position: relative;
  padding: var(--spacing-2);
}

.notification-btn:hover {
  color: var(--color-accent-main);
}

.connection-indicator {
  position: absolute;
  top: 2px;
  right: 2px;
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.connection-indicator.connected {
  background-color: var(--color-success);
}

.connection-indicator.disconnected {
  background-color: var(--color-error);
}

.notification-menu {
  width: 320px;
  max-height: 400px;
  overflow-y: auto;
}

.notification-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-3) var(--spacing-4);
  background-color: var(--color-bg-100);
}

.header-title {
  font-weight: 600;
  color: var(--color-text-000);
}

.view-all-btn {
  font-size: var(--text-xs);
  color: var(--color-accent-main);
}

.loading-container,
.empty-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-8) var(--spacing-4);
  color: var(--color-text-300);
  gap: var(--spacing-2);
}

.notification-list {
  max-height: 300px;
  overflow-y: auto;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  padding: var(--spacing-3) var(--spacing-4);
  cursor: pointer;
  transition: background-color var(--transition-fast);
  border-left: 3px solid transparent;
  position: relative;
}

.notification-item:hover {
  background-color: var(--color-bg-200);
}

.notification-item.unread {
  background-color: var(--color-bg-100);
  border-left-color: var(--color-accent-main);
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-weight: 500;
  color: var(--color-text-100);
  margin-bottom: var(--spacing-1);
  font-size: var(--text-sm);
}

.notification-icon {
  flex-shrink: 0;
}

.notification-icon--success {
  color: var(--color-success);
}

.notification-icon--warning {
  color: var(--color-warning);
}

.notification-icon--error {
  color: var(--color-error);
}

.notification-icon--info,
.notification-icon--system,
.notification-icon--announcement,
.notification-icon--task_update {
  color: var(--color-info);
}

.notification-text {
  color: var(--color-text-200);
  font-size: var(--text-xs);
  line-height: var(--leading-relaxed);
  margin-bottom: var(--spacing-1);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.notification-time {
  color: var(--color-text-400);
  font-size: var(--text-xs);
}

.unread-dot {
  width: 8px;
  height: 8px;
  background-color: var(--color-accent-main);
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: var(--spacing-1);
}

/* Scrollbar styling */
.notification-menu::-webkit-scrollbar,
.notification-list::-webkit-scrollbar {
  width: 4px;
}

.notification-menu::-webkit-scrollbar-track,
.notification-list::-webkit-scrollbar-track {
  background: var(--color-bg-200);
}

.notification-menu::-webkit-scrollbar-thumb,
.notification-list::-webkit-scrollbar-thumb {
  background: var(--color-border);
  border-radius: 2px;
}

.notification-menu::-webkit-scrollbar-thumb:hover,
.notification-list::-webkit-scrollbar-thumb:hover {
  background: var(--color-border-dark);
}
</style>
