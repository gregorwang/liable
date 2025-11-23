<template>
  <div class="admin-dashboard-content">
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
          
          <!-- Queue Statistics Card -->
          <el-card shadow="hover" style="margin-top: 24px">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center">
                <span style="font-weight: bold">队列统计</span>
                <el-icon><DataAnalysis /></el-icon>
              </div>
            </template>
            
            <div class="queue-stats-grid">
              <div 
                v-for="queue in stats.queue_stats" 
                :key="queue.queue_name"
                class="queue-stat-item"
                :class="{ 'inactive': !queue.is_active }"
              >
                <div class="queue-header">
                  <h4>{{ queue.queue_name }}</h4>
                  <el-tag :type="queue.is_active ? 'success' : 'info'" size="small">
                    {{ queue.is_active ? '活跃' : '暂停' }}
                  </el-tag>
                </div>
                
                <div class="queue-metrics">
                  <div class="metric-row">
                    <span class="metric-label">总任务:</span>
                    <span class="metric-value">{{ formatNumber(queue.total_tasks) }}</span>
                  </div>
                  <div class="metric-row">
                    <span class="metric-label">已完成:</span>
                    <span class="metric-value">{{ formatNumber(queue.completed_tasks) }}</span>
                  </div>
                  <div class="metric-row">
                    <span class="metric-label">待处理:</span>
                    <span class="metric-value">{{ formatNumber(queue.pending_tasks) }}</span>
                  </div>
                  <div class="metric-row">
                    <span class="metric-label">通过率:</span>
                    <span class="metric-value">{{ formatPercent(queue.approval_rate) }}</span>
                  </div>
                  <div class="metric-row">
                    <span class="metric-label">平均处理时间:</span>
                    <span class="metric-value">{{ queue.avg_process_time.toFixed(1) }}分钟</span>
                  </div>
                </div>
                
                <div class="queue-progress">
                  <el-progress 
                    :percentage="getPercentage(queue.completed_tasks, queue.total_tasks)"
                    :stroke-width="8"
                    :status="queue.completed_tasks === queue.total_tasks ? 'success' : 'primary'"
                  />
                </div>
              </div>
            </div>
          </el-card>
          
          <!-- Quality Metrics Card -->
          <el-card shadow="hover" style="margin-top: 24px">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center">
                <span style="font-weight: bold">质量指标</span>
                <el-icon><CircleCheck /></el-icon>
              </div>
            </template>
            
            <div class="quality-metrics-grid">
              <div class="quality-metric-item">
                <div class="metric-icon" style="background: #f0f9ff; color: #409eff">
                  <el-icon :size="24"><DataAnalysis /></el-icon>
                </div>
                <div class="metric-info">
                  <div class="metric-value">{{ formatNumber(stats.quality_metrics.total_quality_checks) }}</div>
                  <div class="metric-label">质检总数</div>
                </div>
              </div>
              
              <div class="quality-metric-item">
                <div class="metric-icon" style="background: #f0f9ff; color: #67c23a">
                  <el-icon :size="24"><CircleCheck /></el-icon>
                </div>
                <div class="metric-info">
                  <div class="metric-value">{{ formatNumber(stats.quality_metrics.passed_quality_checks) }}</div>
                  <div class="metric-label">质检通过</div>
                </div>
              </div>
              
              <div class="quality-metric-item">
                <div class="metric-icon" style="background: #fef0f0; color: #f56c6c">
                  <el-icon :size="24"><Warning /></el-icon>
                </div>
                <div class="metric-info">
                  <div class="metric-value">{{ formatNumber(stats.quality_metrics.failed_quality_checks) }}</div>
                  <div class="metric-label">质检失败</div>
                </div>
              </div>
              
              <div class="quality-metric-item">
                <div class="metric-icon" style="background: #f4f4f5; color: #909399">
                  <el-icon :size="24"><DataAnalysis /></el-icon>
                </div>
                <div class="metric-info">
                  <div class="metric-value">{{ formatPercent(stats.quality_metrics.quality_pass_rate) }}</div>
                  <div class="metric-label">质检通过率</div>
                </div>
              </div>
              
              <div class="quality-metric-item">
                <div class="metric-icon" style="background: #f0f9ff; color: #409eff">
                  <el-icon :size="24"><Document /></el-icon>
                </div>
                <div class="metric-info">
                  <div class="metric-value">{{ formatNumber(stats.quality_metrics.second_review_tasks) }}</div>
                  <div class="metric-label">二审任务</div>
                </div>
              </div>
              
              <div class="quality-metric-item">
                <div class="metric-icon" style="background: #f0f9ff; color: #67c23a">
                  <el-icon :size="24"><CircleCheck /></el-icon>
                </div>
                <div class="metric-info">
                  <div class="metric-value">{{ formatNumber(stats.quality_metrics.second_review_completed) }}</div>
                  <div class="metric-label">二审完成</div>
                </div>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Document,
  CircleCheck,
  Warning,
  DataAnalysis,
  User,
  UserFilled,
  Bell,
} from '@element-plus/icons-vue'
import { getOverviewStats } from '../../api/admin'
import { createNotification } from '../../api/notification'
import type { OverviewStats, CreateNotificationRequest } from '../../types'
import { formatNumber, formatPercent } from '../../utils/format'

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
  comment_review_stats: {
    first_review: {
      total_tasks: 0,
      completed_tasks: 0,
      pending_tasks: 0,
      in_progress_tasks: 0,
      approved_count: 0,
      rejected_count: 0,
      approval_rate: 0,
    },
    second_review: {
      total_tasks: 0,
      completed_tasks: 0,
      pending_tasks: 0,
      in_progress_tasks: 0,
      approved_count: 0,
      rejected_count: 0,
      approval_rate: 0,
    },
  },
  video_review_stats: {
    first_review: {
      total_tasks: 0,
      completed_tasks: 0,
      pending_tasks: 0,
      in_progress_tasks: 0,
      approved_count: 0,
      rejected_count: 0,
      approval_rate: 0,
      avg_overall_score: 0,
    },
    second_review: {
      total_tasks: 0,
      completed_tasks: 0,
      pending_tasks: 0,
      in_progress_tasks: 0,
      approved_count: 0,
      rejected_count: 0,
      approval_rate: 0,
      avg_overall_score: 0,
    },
  },
  queue_stats: [],
  quality_metrics: {
    total_quality_checks: 0,
    passed_quality_checks: 0,
    failed_quality_checks: 0,
    quality_pass_rate: 0,
    second_review_tasks: 0,
    second_review_completed: 0,
    second_review_rate: 0,
  },
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
   管理员总览页面样式
   ============================================ */
.admin-dashboard-content {
  padding: var(--spacing-8);
  background-color: var(--color-bg-100);
  min-height: 100vh;
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

/* ============================================
   队列统计样式
   ============================================ */
.queue-stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-6);
}

.queue-stat-item {
  padding: var(--spacing-4);
  border: 1px solid var(--color-border-lighter);
  border-radius: var(--radius-lg);
  background: var(--color-bg-000);
  transition: all var(--transition-base);
}

.queue-stat-item:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.queue-stat-item.inactive {
  opacity: 0.6;
  background: var(--color-bg-100);
}

.queue-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-4);
}

.queue-header h4 {
  margin: 0;
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-000);
}

.queue-metrics {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-4);
}

.metric-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.metric-label {
  font-size: var(--text-sm);
  color: var(--color-text-400);
  font-weight: 500;
}

.metric-value {
  font-size: var(--text-sm);
  color: var(--color-text-200);
  font-weight: 600;
}

.queue-progress {
  margin-top: var(--spacing-2);
}

/* ============================================
   质量指标样式
   ============================================ */
.quality-metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-4);
}

.quality-metric-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding: var(--spacing-3);
  border: 1px solid var(--color-border-lighter);
  border-radius: var(--radius-md);
  background: var(--color-bg-000);
  transition: all var(--transition-base);
}

.quality-metric-item:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-sm);
}

.quality-metric-item .metric-icon {
  width: 48px;
  height: 48px;
  min-width: 48px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-sm);
}

.quality-metric-item .metric-info {
  flex: 1;
  min-width: 0;
}

.quality-metric-item .metric-value {
  font-size: var(--text-xl);
  font-weight: 700;
  color: var(--color-text-000);
  line-height: var(--leading-tight);
  margin-bottom: var(--spacing-1);
}

.quality-metric-item .metric-label {
  font-size: var(--text-xs);
  color: var(--color-text-400);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
}

/* ============================================
   响应式设计扩展
   ============================================ */
@media (max-width: 768px) {
  .queue-stats-grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-4);
  }
  
  .quality-metrics-grid {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: var(--spacing-3);
  }
  
  .quality-metric-item {
    padding: var(--spacing-2);
  }
  
  .quality-metric-item .metric-icon {
    width: 40px;
    height: 40px;
    min-width: 40px;
  }
  
  .quality-metric-item .metric-value {
    font-size: var(--text-lg);
  }
}
</style>

