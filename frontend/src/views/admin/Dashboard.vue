<template>
  <div class="admin-layout">
    <el-container>
      <el-aside width="200px" class="sidebar">
        <div class="logo">
          <h3>管理后台</h3>
        </div>
        <el-menu
          :default-active="currentRoute"
          router
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409eff"
        >
          <el-menu-item index="/admin/dashboard">
            <span>总览</span>
          </el-menu-item>
          <el-menu-item index="/admin/users">
            <span>用户管理</span>
          </el-menu-item>
          <el-menu-item index="/admin/statistics">
            <span>统计分析</span>
          </el-menu-item>
          <el-menu-item index="/admin/tags">
            <span>标签管理</span>
          </el-menu-item>
          <el-menu-item index="/admin/search">
            <span>审核记录搜索</span>
          </el-menu-item>
          <el-menu-item index="/admin/moderation-rules">
            <span>审核规则库</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      
      <el-container>
        <el-header class="header">
          <div class="header-content">
            <h2>数据总览</h2>
            <div class="user-info">
              <span>{{ userStore.user?.username }}</span>
              <el-button @click="handleLogout" text>退出</el-button>
            </div>
          </div>
        </el-header>
        
        <el-main class="main-content">
          <div v-loading="loading" class="stats-grid">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-content">
                <div class="stat-icon" style="background: #ecf5ff; color: #409eff">
                  <el-icon :size="32"><Document /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ formatNumber(stats.total_tasks) }}</div>
                  <div class="stat-label">总任务数</div>
                </div>
              </div>
            </el-card>
            
            <el-card shadow="hover" class="stat-card">
              <div class="stat-content">
                <div class="stat-icon" style="background: #f0f9ff; color: #67c23a">
                  <el-icon :size="32"><CircleCheck /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ formatNumber(stats.completed_tasks) }}</div>
                  <div class="stat-label">已完成</div>
                </div>
              </div>
            </el-card>
            
            <el-card shadow="hover" class="stat-card">
              <div class="stat-content">
                <div class="stat-icon" style="background: #fef0f0; color: #f56c6c">
                  <el-icon :size="32"><Warning /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ formatNumber(stats.pending_tasks) }}</div>
                  <div class="stat-label">待处理</div>
                </div>
              </div>
            </el-card>
            
            <el-card shadow="hover" class="stat-card">
              <div class="stat-content">
                <div class="stat-icon" style="background: #f4f4f5; color: #909399">
                  <el-icon :size="32"><DataAnalysis /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ formatPercent(stats.approval_rate) }}</div>
                  <div class="stat-label">通过率</div>
                </div>
              </div>
            </el-card>
            
            <el-card shadow="hover" class="stat-card">
              <div class="stat-content">
                <div class="stat-icon" style="background: #f0f9ff; color: #409eff">
                  <el-icon :size="32"><User /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ stats.total_reviewers }}</div>
                  <div class="stat-label">审核员总数</div>
                </div>
              </div>
            </el-card>
            
            <el-card shadow="hover" class="stat-card">
              <div class="stat-content">
                <div class="stat-icon" style="background: #f0f9ff; color: #67c23a">
                  <el-icon :size="32"><UserFilled /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ stats.active_reviewers }}</div>
                  <div class="stat-label">活跃审核员</div>
                </div>
              </div>
            </el-card>
          </div>
          
          <el-card shadow="hover" style="margin-top: 24px">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center">
                <span style="font-weight: bold">任务分布</span>
                <el-button size="small" @click="loadStats">刷新</el-button>
              </div>
            </template>
            
            <div class="progress-section">
              <div class="progress-item">
                <div class="progress-label">
                  <span>已完成</span>
                  <span>{{ formatNumber(stats.completed_tasks) }}</span>
                </div>
                <el-progress
                  :percentage="getPercentage(stats.completed_tasks, stats.total_tasks)"
                  :stroke-width="20"
                  status="success"
                />
              </div>
              
              <div class="progress-item">
                <div class="progress-label">
                  <span>进行中</span>
                  <span>{{ formatNumber(stats.in_progress_tasks) }}</span>
                </div>
                <el-progress
                  :percentage="getPercentage(stats.in_progress_tasks, stats.total_tasks)"
                  :stroke-width="20"
                />
              </div>
              
              <div class="progress-item">
                <div class="progress-label">
                  <span>待处理</span>
                  <span>{{ formatNumber(stats.pending_tasks) }}</span>
                </div>
                <el-progress
                  :percentage="getPercentage(stats.pending_tasks, stats.total_tasks)"
                  :stroke-width="20"
                  status="warning"
                />
              </div>
            </div>
          </el-card>
          
          <!-- Send Notification Card -->
          <el-card shadow="hover" style="margin-top: 24px">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center">
                <span style="font-weight: bold">发送全站通知</span>
                <el-icon><Bell /></el-icon>
              </div>
            </template>
            
            <el-form 
              :model="notificationForm" 
              :rules="notificationRules" 
              ref="notificationFormRef"
              label-width="80px"
              @submit.prevent="handleSendNotification"
            >
              <el-form-item label="标题" prop="title">
                <el-input 
                  v-model="notificationForm.title" 
                  placeholder="请输入通知标题"
                  maxlength="255"
                  show-word-limit
                />
              </el-form-item>
              
              <el-form-item label="内容" prop="content">
                <el-input 
                  v-model="notificationForm.content" 
                  type="textarea" 
                  :rows="4"
                  placeholder="请输入通知内容"
                  maxlength="1000"
                  show-word-limit
                />
              </el-form-item>
              
              <el-form-item label="类型" prop="type">
                <el-select v-model="notificationForm.type" placeholder="请选择通知类型" style="width: 100%">
                  <el-option label="信息" value="info" />
                  <el-option label="警告" value="warning" />
                  <el-option label="成功" value="success" />
                  <el-option label="错误" value="error" />
                  <el-option label="系统" value="system" />
                  <el-option label="公告" value="announcement" />
                  <el-option label="任务更新" value="task_update" />
                </el-select>
              </el-form-item>
              
              <el-form-item>
                <el-button 
                  type="primary" 
                  @click="handleSendNotification"
                  :loading="sendingNotification"
                >
                  发送通知
                </el-button>
                <el-button @click="resetNotificationForm">重置</el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Document,
  CircleCheck,
  Warning,
  DataAnalysis,
  User,
  UserFilled,
  Bell,
} from '@element-plus/icons-vue'
import { useUserStore } from '../../stores/user'
import { getOverviewStats } from '../../api/admin'
import { createNotification } from '../../api/notification'
import type { OverviewStats, CreateNotificationRequest } from '../../types'
import { formatNumber, formatPercent } from '../../utils/format'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loading = ref(false)
const sendingNotification = ref(false)
const stats = ref<OverviewStats>({
  total_tasks: 0,
  completed_tasks: 0,
  approved_count: 0,
  rejected_count: 0,
  approval_rate: 0,
  total_reviewers: 0,
  active_reviewers: 0,
  pending_tasks: 0,
  in_progress_tasks: 0,
})

// Notification form
const notificationForm = ref<CreateNotificationRequest>({
  title: '',
  content: '',
  type: 'info',
  is_global: true,
})

const notificationRules = {
  title: [
    { required: true, message: '请输入通知标题', trigger: 'blur' },
    { min: 1, max: 255, message: '标题长度在 1 到 255 个字符', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入通知内容', trigger: 'blur' },
    { min: 1, max: 1000, message: '内容长度在 1 到 1000 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择通知类型', trigger: 'change' }
  ]
}

const notificationFormRef = ref()

const currentRoute = computed(() => route.path)

onMounted(() => {
  loadStats()
})

const loadStats = async () => {
  loading.value = true
  try {
    const data = await getOverviewStats()
    stats.value = data
  } catch (error) {
    console.error('Failed to load stats:', error)
  } finally {
    loading.value = false
  }
}

const getPercentage = (value: number, total: number): number => {
  if (total === 0) return 0
  return Math.round((value / total) * 100)
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确认退出登录？', '提示', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning',
    })
    userStore.logout()
    router.push('/login')
  } catch {
    // Cancel
  }
}

const handleSendNotification = async () => {
  if (!notificationFormRef.value) return
  
  try {
    await notificationFormRef.value.validate()
    sendingNotification.value = true
    
    await createNotification(notificationForm.value)
    ElMessage.success('通知发送成功！')
    resetNotificationForm()
  } catch (error) {
    console.error('Failed to send notification:', error)
    ElMessage.error('通知发送失败')
  } finally {
    sendingNotification.value = false
  }
}

const resetNotificationForm = () => {
  notificationForm.value = {
    title: '',
    content: '',
    type: 'info',
    is_global: true,
  }
  notificationFormRef.value?.resetFields()
}
</script>

<style scoped>
/* ============================================
   布局结构
   ============================================ */
.admin-layout {
  height: 100vh;
  height: 100dvh;
  display: flex;
  overflow: hidden;
}

/* ============================================
   侧边栏样式
   ============================================ */
.sidebar {
  background: linear-gradient(180deg, 
    var(--color-accent-pro-dark) 0%, 
    hsl(251, 55%, 28%) 100%);
  overflow-x: hidden;
  overflow-y: auto;
  box-shadow: var(--shadow-lg);
  border-right: 1px solid var(--color-border);
}

.logo {
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  border-bottom: 1px solid rgba(255, 255, 255, 0.15);
  background: rgba(0, 0, 0, 0.1);
  padding: var(--spacing-4);
}

.logo h3 {
  margin: 0;
  font-size: var(--text-xl);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
}

/* ============================================
   头部样式
   ============================================ */
.header {
  background: var(--color-bg-000);
  box-shadow: var(--shadow-sm);
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--color-border-lighter);
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 var(--spacing-2);
}

.header-content h2 {
  margin: 0;
  font-size: var(--text-2xl);
  color: var(--color-text-000);
  font-weight: 600;
  letter-spacing: var(--tracking-tight);
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  font-size: var(--text-sm);
  color: var(--color-text-200);
  letter-spacing: var(--tracking-wide);
}

/* ============================================
   主内容区域
   ============================================ */
.main-content {
  background: var(--color-bg-100);
  padding: var(--spacing-8);
  overflow-y: auto;
}

/* ============================================
   统计卡片网格
   ============================================ */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-6);
  margin-bottom: var(--spacing-8);
}

.stat-card {
  cursor: default;
  transition: all var(--transition-base);
  border: 1px solid var(--color-border-lighter);
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-5);
  padding: var(--spacing-2);
}

.stat-icon {
  width: 72px;
  height: 72px;
  min-width: 72px;
  border-radius: var(--radius-xl);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-sm);
}

.stat-info {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: var(--text-4xl);
  font-weight: 700;
  color: var(--color-text-000);
  line-height: var(--leading-tight);
  margin-bottom: var(--spacing-2);
  letter-spacing: var(--tracking-tight);
}

.stat-label {
  font-size: var(--text-sm);
  color: var(--color-text-400);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
  line-height: var(--leading-relaxed);
}

/* ============================================
   进度条区域
   ============================================ */
.progress-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-8);
  padding: var(--spacing-2) 0;
}

.progress-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.progress-label {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  font-size: var(--text-sm);
  color: var(--color-text-200);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
  line-height: var(--leading-relaxed);
}

.progress-label span:first-child {
  color: var(--color-text-100);
  font-weight: 600;
}

.progress-label span:last-child {
  font-size: var(--text-base);
  color: var(--color-text-300);
  font-weight: 600;
}

/* ============================================
   响应式设计
   ============================================ */
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-4);
  }
  
  .main-content {
    padding: var(--spacing-4);
  }
  
  .stat-content {
    gap: var(--spacing-3);
  }
  
  .stat-icon {
    width: 56px;
    height: 56px;
    min-width: 56px;
  }
  
  .stat-value {
    font-size: var(--text-3xl);
  }
}

@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  }
}
</style>

