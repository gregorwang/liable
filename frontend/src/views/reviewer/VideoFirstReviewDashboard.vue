<template>
  <GenericReviewDashboard
    :config="dashboardConfig"
    :username="userStore.user?.username || ''"
    :task-count="tasks.length"
    :selected-count="selectedReviews.length"
    :claim-loading="claimLoading"
    v-model:claim-count="claimCount"
    v-model:return-count="returnCount"
    :stats-data="statsData"
    @claim="handleClaimTasks"
    @return="handleReturnTasks"
    @batch-submit="handleBatchSubmit"
    @refresh="handleRefresh"
    @search="goToSearch"
    @logout="handleLogout"
  >
    <template #tasks>
      <el-card
        v-for="task in tasks"
        :key="task.id"
        class="task-card"
        shadow="hover"
      >
        <div class="task-content">
          <div class="task-header">
            <div class="task-info">
              <h3>任务 #{{ task.id }}</h3>
              <p class="task-meta">
                视频ID: {{ task.video_id }} | 
                创建时间: {{ formatTime(task.created_at) }}
              </p>
            </div>
            <div class="task-actions">
              <el-checkbox
                v-model="selectedReviews"
                :value="task.id"
              >
                选择
              </el-checkbox>
            </div>
          </div>
          
          <div class="task-body">
            <div class="video-section">
              <h4>视频预览</h4>
              <VideoPlayer
                v-if="task.video"
                :video="task.video"
                :auto-play="false"
                @loaded="onVideoLoaded"
                @error="onVideoError"
              />
            </div>
            
            <div class="review-section">
              <h4>审核表单</h4>
              <VideoReviewForm
                :task-id="task.id"
                @submit="handleSubmitReview"
              />
            </div>
          </div>
        </div>
      </el-card>
    </template>
  </GenericReviewDashboard>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import GenericReviewDashboard from '@/components/GenericReviewDashboard.vue'
import VideoPlayer from '@/components/VideoPlayer.vue'
import VideoReviewForm from '@/components/VideoReviewForm.vue'
import { videoFirstReviewDashboardConfig } from '@/config/dashboardConfigs'
import { handleClaimError, handleSubmitError, handleReturnError, handleLoadError } from '@/utils/errorHandler'
import type { VideoFirstReviewTask, SubmitVideoFirstReviewRequest } from '@/types'
import {
  claimVideoFirstReviewTasks,
  getMyVideoFirstReviewTasks,
  submitVideoFirstReview,
  returnVideoFirstReviewTasks
} from '@/api/videoReview'

const router = useRouter()
const userStore = useUserStore()

// Dashboard config
const dashboardConfig = videoFirstReviewDashboardConfig

// State
const tasks = ref<VideoFirstReviewTask[]>([])
const claimCount = ref(5)
const returnCount = ref(1)
const claimLoading = ref(false)
const selectedReviews = ref<number[]>([])
const todayCompleted = ref(0)

// Stats data for the dashboard
const statsData = computed(() => ({
  pending_tasks: tasks.value.length,
  today_completed: todayCompleted.value
}))

// Load tasks
const loadTasks = async () => {
  try {
    const response = await getMyVideoFirstReviewTasks()
    tasks.value = response.tasks
  } catch (error) {
    handleLoadError(error)
  }
}

// Claim tasks
const handleClaimTasks = async (count: number) => {
  claimLoading.value = true
  try {
    const response = await claimVideoFirstReviewTasks({ count })
    tasks.value = response.tasks
    
    if (response.count === 0) {
      ElMessage.info('暂无可用任务')
    } else {
      ElMessage.success(`成功领取 ${response.count} 个任务`)
    }
  } catch (error) {
    handleClaimError(error)
  } finally {
    claimLoading.value = false
  }
}

// Return tasks
const handleReturnTasks = async (count: number) => {
  if (tasks.value.length === 0) {
    ElMessage.warning('没有可退单的任务')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要退回 ${count} 个任务吗？`,
      '确认退单',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const taskIds = tasks.value.slice(0, count).map(task => task.id)
    await returnVideoFirstReviewTasks({ task_ids: taskIds })
    
    ElMessage.success('任务退回成功')
    await loadTasks()
  } catch (error) {
    if (error !== 'cancel') {
      handleReturnError(error)
    }
  }
}

// Submit single review
const handleSubmitReview = async (data: SubmitVideoFirstReviewRequest) => {
  try {
    await submitVideoFirstReview(data)
    ElMessage.success('审核提交成功')
    
    // Remove the completed task from the list
    tasks.value = tasks.value.filter(task => task.id !== data.task_id)
    todayCompleted.value++
    
    // Remove from selected reviews if it was selected
    const index = selectedReviews.value.indexOf(data.task_id)
    if (index > -1) {
      selectedReviews.value.splice(index, 1)
    }
  } catch (error) {
    handleSubmitError(error)
  }
}

// Batch submit reviews
const handleBatchSubmit = async () => {
  if (selectedReviews.value.length === 0) {
    ElMessage.warning('请先选择要提交的任务')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要批量提交 ${selectedReviews.value.length} 个审核吗？`,
      '确认批量提交',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    ElMessage.info('批量提交功能需要收集每个任务的审核数据')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Batch submit error:', error)
    }
  }
}

// Refresh tasks
const handleRefresh = async () => {
  await loadTasks()
  ElMessage.success('任务列表已刷新')
}

// Video event handlers
const onVideoLoaded = (duration: number) => {
  console.log('Video loaded, duration:', duration)
}

const onVideoError = (error: string) => {
  console.error('Video error:', error)
}

// Navigation
const goToSearch = () => {
  router.push('/search-tasks')
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}

// Utility functions
const formatTime = (timeString: string) => {
  return new Date(timeString).toLocaleString('zh-CN')
}

// Lifecycle
onMounted(() => {
  loadTasks()
})
</script>


<style scoped>
.task-card {
  background: #fff;
}

.task-content {
  padding: 20px;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.task-info h3 {
  margin: 0 0 8px 0;
  color: #333;
  font-size: 18px;
}

.task-meta {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.task-body {
  display: flex;
  flex-direction: row;
  gap: 24px;
}

.video-section {
  flex: 0 0 45%;
  min-height: 400px;
}

.review-section {
  flex: 1;
  min-height: 400px;
}

.video-section h4,
.review-section h4 {
  margin: 0 0 16px 0;
  color: #333;
  font-size: 16px;
  font-weight: 500;
}

/* Responsive design */
@media (max-width: 1200px) {
  .task-body {
    flex-direction: column;
  }
  
  .video-section {
    flex: 1;
  }
}

@media (max-width: 768px) {
  .task-content {
    padding: 16px;
  }
}
</style>
