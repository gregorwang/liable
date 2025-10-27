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
            <span class="today-count">今日审核：{{ todayReviewCount }}</span>
          </div>
          
          <el-badge :value="notificationStore.unreadCount" :hidden="notificationStore.unreadCount === 0" class="notification-badge">
            <NotificationBell />
          </el-badge>
          
          <el-dropdown @command="handleUserCommand" class="user-dropdown">
            <div class="user-info">
              <el-avatar :size="32" class="user-avatar">
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
          
          <el-menu-item index="permission-list">
            <el-icon><User /></el-icon>
            <template #title>权限列表</template>
          </el-menu-item>
          
          <el-menu-item index="efficiency-stats">
            <el-icon><TrendCharts /></el-icon>
            <template #title>人效统计</template>
          </el-menu-item>
          
          <el-menu-item index="user-management">
            <el-icon><UserFilled /></el-icon>
            <template #title>用户管理</template>
          </el-menu-item>
          
          <el-menu-item index="history-announcements">
            <el-icon><Bell /></el-icon>
            <template #title>历史公告</template>
          </el-menu-item>
          
          <el-menu-item index="rule-documentation">
            <el-icon><Document /></el-icon>
            <template #title>规则文档</template>
          </el-menu-item>
          
          <el-menu-item index="queue-manage">
            <el-icon><Setting /></el-icon>
            <template #title>队列配置</template>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-main class="main-content">
        <component :is="currentComponent" />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, defineAsyncComponent } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Fold,
  Expand,
  ArrowDown,
  List,
  DataBoard,
  User,
  TrendCharts,
  UserFilled,
  Document,
  Setting
} from '@element-plus/icons-vue'
import { useUserStore } from '../stores/user'
import { useNotificationStore } from '../stores/notification'
import NotificationBell from './NotificationBell.vue'

const router = useRouter()
const userStore = useUserStore()
const notificationStore = useNotificationStore()

// 预先定义所有异步组件
// 这样做有以下优点:
// 1. 每个组件在运行时只被包装一次 (不是每次切换菜单时重新创建)
// 2. Vue 能够正确地识别和缓存这些异步组件
// 3. 支持 Suspense 边界和错误处理
const asyncComponents: Record<string, any> = {
  'queue-list': defineAsyncComponent(() => import('./QueueList.vue')),
  'data-management': defineAsyncComponent(() => import('../views/SearchTasks.vue')),
  'permission-list': defineAsyncComponent(() => import('../views/admin/UserManage.vue')),
  'efficiency-stats': defineAsyncComponent(() => import('../views/admin/Statistics.vue')),
  'user-management': defineAsyncComponent(() => import('../views/admin/UserManage.vue')),
  'history-announcements': defineAsyncComponent(() => import('../views/HistoryAnnouncements.vue')),
  'rule-documentation': defineAsyncComponent(() => import('../views/admin/ModerationRules.vue')),
  'queue-manage': defineAsyncComponent(() => import('../views/admin/QueueManage.vue')),
}

// 侧边栏状态
const isCollapsed = ref(false)
const activeMenu = ref('queue-list')

// 统计数据
const todayReviewCount = ref(0)

// 计算属性
const sidebarWidth = computed(() => isCollapsed.value ? '64px' : '200px')

// 当前显示的组件
const currentComponent = computed(() => {
  return asyncComponents[activeMenu.value] || asyncComponents['queue-list']
})

// 方法
const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
}

const handleMenuSelect = (index: string) => {
  console.log('🔍 Menu clicked:', index)
  activeMenu.value = index
  console.log('📦 Active menu updated to:', activeMenu.value)
  console.log('🎯 Current component:', currentComponent.value)
}

const handleUserCommand = async (command: string) => {
  switch (command) {
    case 'profile':
      ElMessage.info('个人设置功能开发中...')
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

onMounted(() => {
  // 模拟获取今日审核数量
  todayReviewCount.value = Math.floor(Math.random() * 100) + 50
  
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

.stats-info {
  font-size: var(--text-sm);
  color: var(--color-text-300);
  font-family: var(--font-sans);
}

.today-count {
  font-weight: 500;
  color: var(--color-accent-main);
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
