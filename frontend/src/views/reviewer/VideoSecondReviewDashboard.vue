<template>
  <div class="video-second-review-dashboard">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <div class="header-left">
            <h2>抖音短视频二审工作台</h2>
            <el-button type="primary" link @click="goToSearch" style="margin-left: 20px">
              <el-icon><Search /></el-icon>
              搜索审核记录
            </el-button>
          </div>
          <div class="user-info">
            <span>欢迎，{{ userStore.user?.username }}</span>
            <el-button @click="handleLogout" text>退出</el-button>
          </div>
        </div>
      </el-header>
      
      <el-main class="main-content">
        <div class="stats-bar">
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">待审核任务</span>
              <span class="stat-value">{{ tasks.length }}</span>
            </div>
          </el-card>
          
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">今日已完成</span>
              <span class="stat-value">{{ todayCompleted }}</span>
            </div>
          </el-card>
        </div>
        
        <div class="actions-bar">
          <div class="claim-section">
            <el-input-number
              v-model="claimCount"
              :min="1"
              :max="50"
              :step="1"
              size="large"
              style="width: 120px"
            />
            <el-button
              type="primary"
              size="large"
              :loading="claimLoading"
              @click="handleClaimTasks"
            >
              领取新任务
            </el-button>
          </div>
          
          <div class="return-section">
            <el-input-number
              v-model="returnCount"
              :min="1"
              :max="Math.max(1, tasks.length)"
              :step="1"
              size="large"
              style="width: 120px"
              :disabled="tasks.length === 0"
            />
            <el-button
              type="warning"
              size="large"
              :disabled="tasks.length === 0"
              @click="handleReturnTasks"
            >
              退单
            </el-button>
          </div>
          
          <el-button
            size="large"
            :disabled="selectedReviews.length === 0"
            @click="handleBatchSubmit"
          >
            批量提交（{{ selectedReviews.length }}条）
          </el-button>
          
          <el-button
            size="large"
            @click="handleRefresh"
          >
            刷新任务列表
          </el-button>
        </div>
        
        <div v-if="tasks.length === 0" class="empty-state">
          <el-empty description="暂无待审核任务，点击「领取新任务」开始工作" />
        </div>
        
        <div v-else class="tasks-container">
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
                    @change="handleTaskSelection"
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
                
                <div class="comparison-section" v-if="task.first_review_result">
                  <h4>一审结果对比</h4>
                  <div class="comparison-grid">
                    <div class="first-review">
                      <h5>一审结果</h5>
                      <div class="review-summary">
                        <el-tag :type="task.first_review_result.is_approved ? 'success' : 'danger'" size="large">
                          {{ task.first_review_result.is_approved ? '通过' : '拒绝' }}
                        </el-tag>
                        <div class="score-display">
                          <span>综合评分: {{ task.first_review_result.overall_score }}/40</span>
                        </div>
                        <div class="dimension-scores">
                          <div class="score-item">
                            <span>内容质量:</span>
                            <span>{{ task.first_review_result.quality_dimensions.content_quality.score }}</span>
                          </div>
                          <div class="score-item">
                            <span>技术质量:</span>
                            <span>{{ task.first_review_result.quality_dimensions.technical_quality.score }}</span>
                          </div>
                          <div class="score-item">
                            <span>合规性:</span>
                            <span>{{ task.first_review_result.quality_dimensions.compliance.score }}</span>
                          </div>
                          <div class="score-item">
                            <span>传播潜力:</span>
                            <span>{{ task.first_review_result.quality_dimensions.engagement_potential.score }}</span>
                          </div>
                        </div>
                        <div class="traffic-pool">
                          <span>流量池建议:</span>
                          <el-tag>{{ task.first_review_result.traffic_pool_result || '未指定' }}</el-tag>
                        </div>
                        <div class="reason" v-if="task.first_review_result.reason">
                          <span>审核理由:</span>
                          <p>{{ task.first_review_result.reason }}</p>
                        </div>
                      </div>
                    </div>
                    
                    <div class="second-review">
                      <h5>二审表单</h5>
                      <VideoReviewForm
                        :task-id="task.id"
                        :is-second-review="true"
                        :first-review-result="task.first_review_result"
                        @submit="handleSubmitReview"
                      />
                    </div>
                  </div>
                </div>
                
                <div class="review-section" v-else>
                  <h4>审核表单</h4>
                  <VideoReviewForm
                    :task-id="task.id"
                    :is-second-review="true"
                    @submit="handleSubmitReview"
                  />
                </div>
              </div>
            </div>
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import VideoPlayer from '@/components/VideoPlayer.vue'
import VideoReviewForm from '@/components/VideoReviewForm.vue'
import type { 
  VideoSecondReviewTask, 
  SubmitVideoSecondReviewRequest
} from '@/types'
import {
  claimVideoSecondReviewTasks,
  getMyVideoSecondReviewTasks,
  submitVideoSecondReview,
  returnVideoSecondReviewTasks
} from '@/api/videoReview'

const router = useRouter()
const userStore = useUserStore()

// State
const tasks = ref<VideoSecondReviewTask[]>([])
const claimCount = ref(5)
const returnCount = ref(1)
const claimLoading = ref(false)
const selectedReviews = ref<number[]>([])
const todayCompleted = ref(0)

// Load tasks
const loadTasks = async () => {
  try {
    const response = await getMyVideoSecondReviewTasks()
    tasks.value = response.tasks
  } catch (error) {
    console.error('Failed to load tasks:', error)
    ElMessage.error('加载任务失败')
  }
}

// Claim tasks
const handleClaimTasks = async () => {
  claimLoading.value = true
  try {
    const response = await claimVideoSecondReviewTasks({ count: claimCount.value })
    tasks.value = response.tasks
    
    if (response.count === 0) {
      ElMessage.info('暂无可用任务')
    } else {
      ElMessage.success(`成功领取 ${response.count} 个任务`)
    }
  } catch (error: any) {
    console.error('Failed to claim tasks:', error)
    ElMessage.error(error.response?.data?.error || '领取任务失败')
  } finally {
    claimLoading.value = false
  }
}

// Return tasks
const handleReturnTasks = async () => {
  if (tasks.value.length === 0) {
    ElMessage.warning('没有可退单的任务')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要退回 ${returnCount.value} 个任务吗？`,
      '确认退单',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const taskIds = tasks.value.slice(0, returnCount.value).map(task => task.id)
    await returnVideoSecondReviewTasks({ task_ids: taskIds })
    
    ElMessage.success('任务退回成功')
    await loadTasks()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to return tasks:', error)
      ElMessage.error(error.response?.data?.error || '退回任务失败')
    }
  }
}

// Submit single review
const handleSubmitReview = async (data: SubmitVideoSecondReviewRequest) => {
  try {
    await submitVideoSecondReview(data)
    ElMessage.success('审核提交成功')
    
    // Remove the completed task from the list
    tasks.value = tasks.value.filter(task => task.id !== data.task_id)
    todayCompleted.value++
    
    // Remove from selected reviews if it was selected
    const index = selectedReviews.value.indexOf(data.task_id)
    if (index > -1) {
      selectedReviews.value.splice(index, 1)
    }
  } catch (error: any) {
    console.error('Failed to submit review:', error)
    ElMessage.error(error.response?.data?.error || '提交审核失败')
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
    
    // Note: In a real implementation, you would collect the review data
    // from each task's form. For now, we'll just show a placeholder.
    ElMessage.info('批量提交功能需要收集每个任务的审核数据')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Batch submit error:', error)
    }
  }
}

// Handle task selection
const handleTaskSelection = () => {
  // Task selection logic is handled by the checkbox
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
.video-second-review-dashboard {
  height: 100vh;
  background: #f5f5f5;
}

.header {
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-left h2 {
  margin: 0;
  color: #333;
  font-size: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.main-content {
  padding: 20px;
}

.stats-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.stat-label {
  font-size: 14px;
  color: #666;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
}

.actions-bar {
  display: flex;
  gap: 16px;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.claim-section,
.return-section {
  display: flex;
  gap: 8px;
  align-items: center;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.tasks-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

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
  flex-direction: column;
  gap: 24px;
}

.video-section {
  min-height: 300px;
}

.video-section h4 {
  margin: 0 0 16px 0;
  color: #333;
  font-size: 16px;
  font-weight: 500;
}

.comparison-section h4 {
  margin: 0 0 16px 0;
  color: #333;
  font-size: 16px;
  font-weight: 500;
}

.comparison-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.first-review,
.second-review {
  padding: 16px;
  border-radius: 8px;
}

.first-review {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
}

.second-review {
  background: #fff;
  border: 1px solid #e4e7ed;
}

.first-review h5,
.second-review h5 {
  margin: 0 0 16px 0;
  color: #333;
  font-size: 14px;
  font-weight: 500;
}

.review-summary {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.score-display {
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

.dimension-scores {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.score-item {
  display: flex;
  justify-content: space-between;
  padding: 4px 8px;
  background: rgba(255, 255, 255, 0.7);
  border-radius: 4px;
  font-size: 12px;
}

.traffic-pool {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.reason {
  font-size: 14px;
}

.reason span {
  font-weight: 500;
  color: #666;
}

.reason p {
  margin: 4px 0 0 0;
  color: #333;
  line-height: 1.4;
}

.review-section h4 {
  margin: 0 0 16px 0;
  color: #333;
  font-size: 16px;
  font-weight: 500;
}

/* Responsive design */
@media (max-width: 1200px) {
  .comparison-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    gap: 16px;
    height: auto;
    padding: 16px 0;
  }
  
  .actions-bar {
    flex-direction: column;
    align-items: stretch;
  }
  
  .claim-section,
  .return-section {
    justify-content: center;
  }
  
  .main-content {
    padding: 16px;
  }
  
  .stats-bar {
    flex-direction: column;
  }
  
  .dimension-scores {
    grid-template-columns: 1fr;
  }
}
</style>
