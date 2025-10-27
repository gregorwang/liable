<template>
  <div class="history-announcements">
    <div class="page-header">
      <h2>历史公告</h2>
      <p class="page-description">查看系统历史公告和重要通知</p>
    </div>
    
    <div v-loading="notificationStore.isLoading" class="announcements-container">
      <el-card
        v-for="notification in (notificationStore.notifications || [])"
        :key="notification.id"
        shadow="hover"
        class="announcement-card"
        :class="{ 'unread': !notification.is_read }"
        @click="handleNotificationClick(notification)"
      >
        <template #header>
          <div class="card-header">
            <div class="notification-title">
              <el-icon 
                :class="getNotificationIconClass(notification.type)"
                :size="16"
              >
                <component :is="getNotificationIcon(notification.type)" />
              </el-icon>
              {{ notification.title }}
            </div>
            <div class="notification-meta">
              <el-tag 
                :type="getNotificationTagType(notification.type)" 
                size="small"
              >
                {{ getNotificationTypeText(notification.type) }}
              </el-tag>
              <span class="notification-date">{{ formatDate(notification.created_at) }}</span>
            </div>
          </div>
        </template>
        <div class="announcement-content">
          <p>{{ notification.content }}</p>
          <div v-if="!notification.is_read" class="unread-indicator">
            <el-icon><CircleCheck /></el-icon>
            <span>点击标记为已读</span>
          </div>
        </div>
      </el-card>
      
      <el-empty 
        v-if="!notificationStore.isLoading && (!notificationStore.notifications || notificationStore.notifications.length === 0)" 
        description="暂无历史公告" 
        :image-size="120"
      />
      
      <div v-if="notificationStore.notifications && notificationStore.notifications.length > 0" class="load-more-container">
        <el-button 
          :loading="loadingMore"
          @click="loadMore"
          :disabled="!hasMore"
        >
          {{ hasMore ? '加载更多' : '没有更多了' }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  InfoFilled, 
  WarningFilled, 
  SuccessFilled, 
  CircleCloseFilled,
  CircleCheck 
} from '@element-plus/icons-vue'
import { useNotificationStore } from '../stores/notification'
import type { NotificationResponse, NotificationType } from '../types'

const notificationStore = useNotificationStore()
const loadingMore = ref(false)
const currentPage = ref(0)
const pageSize = 10

const hasMore = computed(() => {
  return (notificationStore.notifications || []).length >= (currentPage.value + 1) * pageSize
})

onMounted(async () => {
  try {
    console.log('📢 HistoryAnnouncements mounted, fetching notifications...')
    const result = await notificationStore.fetchRecent(pageSize, 0)
    console.log('✅ Fetch result:', result)
    console.log('📋 Notifications in store:', notificationStore.notifications)
    currentPage.value = 0
  } catch (error) {
    console.error('❌ Failed to load notifications:', error)
  }
})

const handleNotificationClick = async (notification: NotificationResponse) => {
  if (!notification.is_read) {
    await notificationStore.markNotificationAsRead(notification.id)
  }
}

const loadMore = async () => {
  if (loadingMore.value || !hasMore.value) return
  
  loadingMore.value = true
  try {
    currentPage.value += 1
    await notificationStore.fetchRecent(pageSize, currentPage.value * pageSize)
  } catch (error) {
    console.error('Failed to load more notifications:', error)
    ElMessage.error('加载更多失败')
  } finally {
    loadingMore.value = false
  }
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

const getNotificationTagType = (type: NotificationType) => {
  switch (type) {
    case 'success':
      return 'success'
    case 'warning':
      return 'warning'
    case 'error':
      return 'danger'
    case 'info':
    case 'system':
    case 'announcement':
    case 'task_update':
    default:
      return 'info'
  }
}

const getNotificationTypeText = (type: NotificationType) => {
  const typeMap = {
    'info': '信息',
    'warning': '警告',
    'success': '成功',
    'error': '错误',
    'system': '系统',
    'announcement': '公告',
    'task_update': '任务更新'
  }
  return typeMap[type] || '通知'
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.history-announcements {
  max-width: 1200px;
  margin: 0 auto;
  padding: var(--spacing-6);
}

.page-header {
  margin-bottom: var(--spacing-8);
  text-align: center;
}

.page-header h2 {
  margin: 0 0 var(--spacing-3) 0;
  font-size: var(--text-3xl);
  font-weight: 600;
  color: var(--color-text-000);
  letter-spacing: var(--tracking-tight);
}

.page-description {
  margin: 0;
  font-size: var(--text-base);
  color: var(--color-text-300);
  letter-spacing: var(--tracking-wide);
}

.announcements-container {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
}

.announcement-card {
  transition: all var(--transition-base);
  border: 1px solid var(--color-border-lighter);
  cursor: pointer;
}

.announcement-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
  border-color: var(--color-accent-main);
}

.announcement-card.unread {
  border-left: 4px solid var(--color-accent-main);
  background-color: var(--color-bg-100);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-3);
}

.notification-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-000);
  letter-spacing: var(--tracking-tight);
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

.notification-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: var(--spacing-1);
}

.notification-date {
  font-size: var(--text-xs);
  color: var(--color-text-400);
}

.announcement-content {
  padding: var(--spacing-2) 0;
}

.announcement-content p {
  margin: 0 0 var(--spacing-3) 0;
  font-size: var(--text-base);
  line-height: var(--leading-relaxed);
  color: var(--color-text-200);
  letter-spacing: var(--tracking-wide);
}

.unread-indicator {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--text-xs);
  color: var(--color-accent-main);
  font-weight: 500;
}

.load-more-container {
  display: flex;
  justify-content: center;
  padding: var(--spacing-6) 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .history-announcements {
    padding: var(--spacing-4);
  }
  
  .page-header h2 {
    font-size: var(--text-2xl);
  }
  
  .announcements-container {
    gap: var(--spacing-4);
  }
  
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-2);
  }
}
</style>
