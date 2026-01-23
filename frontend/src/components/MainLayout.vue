<template>
  <el-container class="main-layout">
    <!-- 顶部导航栏 -->
    <el-header class="header">
      <div class="header-content">
        <div class="header-left">
          <el-button
            type="text"
            @click="toggleCollapse"
            class="collapse-btn"
          >
            <el-icon size="20">
              <Fold v-if="!isCollapsed" />
              <Expand v-else />
            </el-icon>
          </el-button>
          <h1 class="logo">评论审核系统</h1>
        </div>
        
        <div class="header-right">
          <div class="stats-info">
            <span v-if="statsLoading">今日审核统计加载中...</span>
            <span v-else-if="statsError">今日审核统计获取失败</span>
            <template v-else>
              <span class="today-label">今日审核评论：</span>
              <span class="today-count">{{ todayCommentCount }}</span>
              <span class="stats-divider">/</span>
              <span class="today-label">视频：</span>
              <span class="today-count">{{ todayVideoCount }}</span>
            </template>
          </div>

          <el-button class="bug-report-btn" plain @click="bugDialogVisible = true">
            <el-icon size="18">
              <Warning />
            </el-icon>
            Bug反馈
          </el-button>

          <el-badge :value="notificationStore.unreadCount" :hidden="notificationStore.unreadCount === 0" class="notification-badge">
            <el-dropdown 
              trigger="click" 
              placement="bottom-end"
              @command="handleNotificationCommand"
              class="notification-dropdown"
            >
              <el-badge :value="notificationStore.unreadCount" :hidden="notificationStore.unreadCount === 0" class="notification-badge">
                <el-button type="text" class="notification-btn">
                  <el-icon size="18">
                    <Bell />
                  </el-icon>
                </el-button>
              </el-badge>
              <template #dropdown>
                <el-dropdown-menu>
                  <div class="notification-header">
                    <span>通知</span>
                    <el-button type="text" size="small" @click="markAllAsRead">全部已读</el-button>
                  </div>
                  <el-divider />
                  <div class="notification-list">
                    <div v-if="notificationStore.notifications.length === 0" class="no-notifications">
                      暂无通知
                    </div>
                    <div 
                      v-for="notification in notificationStore.notifications.slice(0, 5)" 
                      :key="notification.id"
                      class="notification-item"
                      :class="{ 'unread': !notification.is_read }"
                      @click="markAsRead(notification.id)"
                    >
                      <div class="notification-content">
                        <div class="notification-title">{{ notification.title }}</div>
                        <div class="notification-message">{{ notification.content }}</div>
                        <div class="notification-time">{{ formatTime(notification.created_at) }}</div>
                      </div>
                    </div>
                  </div>
                  <el-divider />
                  <el-dropdown-item command="view-all">查看全部通知</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </el-badge>
          
          <el-dropdown @command="handleUserCommand" class="user-dropdown">
            <div class="user-info">
              <el-avatar :size="32" class="user-avatar" :src="userStore.user?.avatar_url || undefined">
                {{ userStore.user?.username?.charAt(0).toUpperCase() }}
              </el-avatar>
              <span class="username">{{ userStore.user?.username }}</span>
              <el-icon class="dropdown-icon">
                <ArrowDown />
              </el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人设置</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-header>

    <el-container>
      <!-- 左侧边栏 -->
      <el-aside :width="sidebarWidth" class="sidebar">
        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapsed"
          :collapse-transition="false"
          @select="handleMenuSelect"
          class="sidebar-menu"
        >
          <el-menu-item index="queue-list">
            <el-icon><List /></el-icon>
            <template #title>队列列表</template>
          </el-menu-item>
          
          <el-menu-item index="data-management">
            <el-icon><DataBoard /></el-icon>
            <template #title>数据管理</template>
          </el-menu-item>
          
          <!-- 视频审核菜单项 -->
          <el-sub-menu index="video-review">
            <template #title>
              <el-icon><VideoPlay /></el-icon>
              <span>视频审核</span>
            </template>

            <el-menu-item index="video-queue-review">
              <el-icon><TrendCharts /></el-icon>
              <template #title>视频审核工作台</template>
            </el-menu-item>

          </el-sub-menu>
          
          <!-- 仅管理员可见的菜单项 -->
          <template v-if="userStore.isAdmin()">
            <!-- 管理后台分组 -->
            <el-sub-menu index="admin-management">
              <template #title>
                <el-icon><Setting /></el-icon>
                <span>管理后台</span>
              </template>
              
              <el-menu-item index="admin-dashboard">
                <el-icon><DataBoard /></el-icon>
                <template #title>总览</template>
              </el-menu-item>
              
              <el-menu-item index="admin-user-management">
                <el-icon><UserFilled /></el-icon>
                <template #title>用户管理</template>
              </el-menu-item>
              
              <el-menu-item index="admin-statistics">
                <el-icon><TrendCharts /></el-icon>
                <template #title>统计分析</template>
              </el-menu-item>

              <el-menu-item index="admin-audit-logs">
                <el-icon><Document /></el-icon>
                <template #title>审计日志</template>
              </el-menu-item>

              <el-menu-item index="admin-monitoring">
                <el-icon><TrendCharts /></el-icon>
                <template #title>系统监控</template>
              </el-menu-item>
              
              <el-menu-item index="admin-tag-management">
                <el-icon><PriceTag /></el-icon>
                <template #title>评论标签管理</template>
              </el-menu-item>

              <el-menu-item index="admin-video-tag-management">
                <el-icon><VideoPlay /></el-icon>
                <template #title>视频标签管理</template>
              </el-menu-item>
              
              <el-menu-item index="admin-queue-management">
                <el-icon><Operation /></el-icon>
                <template #title>队列配置</template>
              </el-menu-item>
              
              <el-menu-item index="admin-video-import">
                <el-icon><Upload /></el-icon>
                <template #title>视频导入</template>
              </el-menu-item>

              <el-menu-item index="admin-ai-review">
                <el-icon><MagicStick /></el-icon>
                <template #title>AI审核对比</template>
              </el-menu-item>

              <el-menu-item index="admin-bug-reports">
                <el-icon><Warning /></el-icon>
                <template #title>Bug反馈</template>
              </el-menu-item>
              
              <el-menu-item index="permission-management">
                <el-icon><Key /></el-icon>
                <template #title>权限管理</template>
              </el-menu-item>
            </el-sub-menu>
          </template>
          
          <!-- 所有用户都可以访问历史公告、规则库与文档说明 -->
          <el-menu-item index="history-announcements">
            <el-icon><Bell /></el-icon>
            <template #title>历史公告</template>
          </el-menu-item>
          
          <el-menu-item index="rule-documentation">
            <el-icon><Document /></el-icon>
            <template #title>规则文档</template>
          </el-menu-item>

          <el-menu-item index="system-documentation">
            <el-icon><Document /></el-icon>
            <template #title>文档说明</template>
          </el-menu-item>
        </el-menu>
      </el-aside>
      
      <!-- 主内容区域 -->
      <el-main class="main-content">
        <Suspense>
          <component :is="currentComponent" />
          <template #fallback>
            <div class="loading-container">
              <el-icon class="is-loading"><Loading /></el-icon>
              <span>加载中...</span>
            </div>
          </template>
        </Suspense>
      </el-main>
    </el-container>

    <BugReportDialog v-model="bugDialogVisible" />
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, defineAsyncComponent, provide } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Fold,
  Expand,
  ArrowDown,
  List,
  DataBoard,
  TrendCharts,
  UserFilled,
  Document,
  Setting,
  Bell,
  PriceTag,
  Operation,
  Loading,
  VideoPlay,
  Upload,
  Key,
  MagicStick,
  Warning
} from '@element-plus/icons-vue'
import BugReportDialog from './BugReportDialog.vue'
import { useUserStore } from '../stores/user'
import { useNotificationStore } from '../stores/notification'
import { getTodayReviewStats } from '../api/admin'
import type { TodayReviewStats } from '../types'

const router = useRouter()
const userStore = useUserStore()
const notificationStore = useNotificationStore()

// 预先定义所有异步组件
// 这样做有以下优点:
// 1. 每个组件在运行时只被包装一次 (不是每次切换菜单时重新创建)
// 2. Vue 能够正确地识别和缓存这些异步组件
// 3. 支持 Suspense 边界和错误处理
const asyncComponents: Record<string, any> = {
  'queue-list': defineAsyncComponent({
    loader: () => import('./QueueList.vue'), // 重新启用QueueList组件
    loadingComponent: () => '加载中...',
    errorComponent: () => '加载失败',
    delay: 200,
    timeout: 3000
  }),
  'data-management': defineAsyncComponent(() => import('../views/SearchTasks.vue')),
  'history-announcements': defineAsyncComponent(() => import('../views/HistoryAnnouncements.vue')),
  'rule-documentation': defineAsyncComponent(() => import('../views/admin/ModerationRules.vue')),
  'system-documentation': defineAsyncComponent(() => import('../views/SystemDocumentation.vue')),
  'profile-settings': defineAsyncComponent(() => import('../views/ProfileSettings.vue')),
  
  // 管理员专用组件
  'admin-dashboard': defineAsyncComponent(() => import('../views/admin/Dashboard.vue')),
  'admin-user-management': defineAsyncComponent(() => import('../views/admin/UserManage.vue')),
  'admin-statistics': defineAsyncComponent(() => import('../views/admin/Statistics.vue')),
  'admin-audit-logs': defineAsyncComponent(() => import('../views/admin/AuditLogs.vue')),
  'admin-monitoring': defineAsyncComponent(() => import('../views/admin/Monitoring.vue')),
  'admin-tag-management': defineAsyncComponent(() => import('../views/admin/TagManage.vue')),
  'admin-video-tag-management': defineAsyncComponent(() => import('../views/admin/VideoTagManage.vue')),
  'admin-queue-management': defineAsyncComponent(() => import('../views/admin/QueueManage.vue')),
  'admin-video-import': defineAsyncComponent(() => import('../views/admin/VideoImport.vue')),
  'admin-ai-review': defineAsyncComponent(() => import('../views/admin/AIReview.vue')),
  'admin-bug-reports': defineAsyncComponent(() => import('../views/admin/BugReports.vue')),
  'permission-management': defineAsyncComponent(() => import('../views/admin/PermissionManage.vue')),
  
  // 视频审核组件
  'video-queue-review': defineAsyncComponent(() => import('../views/reviewer/VideoQueueReviewDashboard.vue')),
}

// 侧边栏状态
const isCollapsed = ref(false)
const activeMenu = ref('queue-list')

provide('setActiveMenu', (menu: string) => {
  activeMenu.value = menu
})

// 统计数据
const todayStats = ref<TodayReviewStats | null>(null)
const statsLoading = ref(false)
const statsError = ref(false)
const bugDialogVisible = ref(false)

// 计算属性
const sidebarWidth = computed(() => isCollapsed.value ? '64px' : '200px')
const todayCommentCount = computed(() => todayStats.value?.comment.total ?? 0)
const todayVideoCount = computed(() => todayStats.value?.video.total ?? 0)

// 当前显示的组件
const currentComponent = computed(() => {
  const component = asyncComponents[activeMenu.value] || asyncComponents['queue-list']
  return component
})

// 方法
const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
}

const handleMenuSelect = (index: string) => {
  activeMenu.value = index
}

const handleUserCommand = async (command: string) => {
  switch (command) {
    case 'profile':
      activeMenu.value = 'profile-settings'
      break
    case 'logout':
      try {
        await ElMessageBox.confirm('确认退出登录？', '提示', {
          confirmButtonText: '确认',
          cancelButtonText: '取消',
          type: 'warning',
        })
        // Close SSE connection before logout
        notificationStore.closeSSE()
        userStore.logout()
        router.push('/login')
      } catch {
        // 用户取消
      }
      break
  }
}

const handleNotificationCommand = (command: string) => {
  switch (command) {
    case 'view-all':
      // 可以跳转到通知页面或显示更多通知
      ElMessage.info('查看全部通知功能开发中...')
      break
  }
}

const markAsRead = (notificationId: number) => {
  notificationStore.markNotificationAsRead(notificationId)
}

const markAllAsRead = () => {
  // 标记所有未读通知为已读
  notificationStore.notifications.forEach(notification => {
    if (!notification.is_read) {
      notificationStore.markNotificationAsRead(notification.id)
    }
  })
}

const formatTime = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`
  return date.toLocaleDateString('zh-CN')
}

const loadTodayStats = async () => {
  statsLoading.value = true
  statsError.value = false
  try {
    todayStats.value = await getTodayReviewStats()
  } catch (error) {
    console.error('Failed to load today review stats', error)
    statsError.value = true
  } finally {
    statsLoading.value = false
  }
}

onMounted(() => {
  loadTodayStats()
  userStore.loadProfile().catch(() => {
    // loadProfile already handles logout on failure
  })

  // Initialize notification system
  notificationStore.init()
})
</script>

<style scoped>
.main-layout {
  height: 100vh;
  background: linear-gradient(135deg, var(--color-bg-100) 0%, var(--color-bg-000) 100%);
}

/* 顶部导航栏样式 */
.header {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  height: 60px !important;
  display: flex;
  align-items: center;
  z-index: 1000;
  border-bottom: 1px solid rgba(204, 122, 77, 0.08);
  font-family: var(--font-sans);
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 var(--spacing-6);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
}

.collapse-btn {
  font-size: var(--text-lg);
  color: var(--color-text-300);
  transition: color var(--transition-fast);
}

.collapse-btn:hover {
  color: var(--color-accent-main);
}

.logo {
  margin: 0;
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text-000);
  letter-spacing: var(--tracking-tight);
  font-family: var(--font-sans);
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-6);
}

.bug-report-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-2);
  color: var(--color-accent-main);
  border-color: rgba(204, 122, 77, 0.35);
  font-family: var(--font-sans);
}

.bug-report-btn:hover {
  color: var(--color-accent-main-hover);
  border-color: rgba(204, 122, 77, 0.55);
  background-color: var(--color-accent-main-lighter);
}

.stats-info {
  font-size: var(--text-sm);
  color: var(--color-text-300);
  font-family: var(--font-sans);
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.today-count {
  font-weight: 500;
  color: var(--color-accent-main);
  font-size: var(--text-sm);
}

.today-label {
  font-size: var(--text-xs);
  color: var(--color-text-400);
}

.stats-divider {
  color: var(--color-text-300);
  font-size: var(--text-xs);
}

.notification-badge {
  margin-right: var(--spacing-2);
}

.notification-btn {
  font-size: var(--text-lg);
  color: var(--color-text-300);
  transition: color var(--transition-fast);
}

.notification-btn:hover {
  color: var(--color-accent-main);
}

.user-dropdown {
  cursor: pointer;
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding: var(--spacing-2) var(--spacing-3);
  border-radius: var(--radius-md);
  transition: background-color var(--transition-fast);
}

.user-info:hover {
  background-color: var(--color-bg-200);
}

.user-avatar {
  background-color: var(--color-accent-main) !important;
  color: white !important;
  font-weight: 600;
  font-family: var(--font-sans);
}

.username {
  font-size: var(--text-sm);
  color: var(--color-text-100);
  font-weight: 500;
  font-family: var(--font-sans);
}

.dropdown-icon {
  font-size: var(--text-xs);
  color: var(--color-text-400);
}

/* 左侧边栏样式 */
.sidebar {
  background: rgba(250, 247, 245, 0.6);
  backdrop-filter: blur(8px);
  box-shadow: inset -2px 0 8px rgba(0, 0, 0, 0.02);
  transition: width var(--transition-base);
  border-right: 1px solid rgba(204, 122, 77, 0.08);
  font-family: var(--font-sans);
}

.sidebar-menu {
  border: none;
  height: 100%;
  font-family: var(--font-sans);
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 200px;
}

/* 主内容区样式 */
.main-content {
  padding: var(--spacing-8);
  background: linear-gradient(135deg, 
    rgba(248, 247, 245, 0.8) 0%, 
    rgba(255, 255, 255, 0.95) 100%);
  overflow-y: auto;
  font-family: var(--font-sans);
}

/* 加载容器样式 */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  gap: var(--spacing-4);
  color: var(--color-text-300);
  font-size: var(--text-sm);
}

.loading-container .el-icon {
  font-size: var(--text-2xl);
  color: var(--color-accent-main);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .header-content {
    padding: 0 var(--spacing-4);
  }
  
  .header-right {
    gap: var(--spacing-4);
  }
  
  .stats-info {
    display: none;
  }
  
  .main-content {
    padding: var(--spacing-6);
  }
}

@media (max-width: 1024px) {
  .main-content {
    padding: var(--spacing-6);
  }
}
</style>
