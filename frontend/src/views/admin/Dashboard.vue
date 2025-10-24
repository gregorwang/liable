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
} from '@element-plus/icons-vue'
import { useUserStore } from '../../stores/user'
import { getOverviewStats } from '../../api/admin'
import type { OverviewStats } from '../../types'
import { formatNumber, formatPercent } from '../../utils/format'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loading = ref(false)
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
</script>

<style scoped>
.admin-layout {
  height: 100vh;
}

.sidebar {
  background: #304156;
  overflow-x: hidden;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.logo h3 {
  margin: 0;
  font-size: 18px;
}

.header {
  background: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-content h2 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.main-content {
  background: #f5f7fa;
  padding: 24px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  cursor: default;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.progress-section {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.progress-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.progress-label {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  color: #606266;
}
</style>

