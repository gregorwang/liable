<template>
  <div class="second-review-dashboard">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <div class="header-left">
            <h2>{{ currentQueueName ? `${currentQueueName} - 二审工作台` : '评论审核二审工作台' }}</h2>
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
              <span class="stat-label">待二审任务</span>
              <span class="stat-value">{{ secondReviewTasks.length }}</span>
            </div>
          </el-card>
          
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">今日已完成</span>
              <span class="stat-value">{{ completedToday }}</span>
            </div>
          </el-card>
          
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">二审通过率</span>
              <span class="stat-value">{{ approvalRate }}%</span>
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
              领取二审任务
            </el-button>
          </div>
          
          <div class="return-section">
            <el-input-number
              v-model="returnCount"
              :min="1"
              :max="Math.max(1, secondReviewTasks.length)"
              :step="1"
              size="large"
              style="width: 120px"
              :disabled="secondReviewTasks.length === 0"
            />
            <el-button
              type="warning"
              size="large"
              :disabled="secondReviewTasks.length === 0"
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
        
        <div v-if="secondReviewTasks.length === 0" class="empty-state">
          <el-empty description="暂无待二审任务，点击「领取二审任务」开始工作" />
        </div>
        
        <div v-else class="tasks-container">
          <el-card
            v-for="task in secondReviewTasks"
            :key="task.id"
            class="task-card"
            shadow="hover"
          >
            <div v-if="reviews[task.id]" class="task-content">
              <!-- 评论内容 -->
              <div class="comment-section">
                <h3 class="section-title">评论内容</h3>
                <div class="comment-text">
                  {{ task.comment?.text || '评论内容加载中...' }}
                </div>
              </div>
              
              <el-divider />
              
              <!-- 一审结果 -->
              <div class="first-review-section">
                <h3 class="section-title">一审结果</h3>
                <div class="first-review-info">
                  <div class="review-result">
                    <el-tag 
                      :type="task.first_review_result?.is_approved ? 'success' : 'danger'"
                      size="large"
                    >
                      {{ task.first_review_result?.is_approved ? '一审通过' : '一审不通过' }}
                    </el-tag>
                  </div>
                  
                  <div v-if="task.first_review_result && !task.first_review_result.is_approved" class="first-review-details">
                    <div class="tags-section">
                      <span class="label">违规标签：</span>
                      <el-tag
                        v-for="tag in (task.first_review_result?.tags || [])"
                        :key="tag"
                        type="danger"
                        size="small"
                        style="margin-right: 8px; margin-bottom: 4px"
                      >
                        {{ tag }}
                      </el-tag>
                    </div>
                    
                    <div class="reason-section">
                      <span class="label">一审原因：</span>
                      <div class="reason-text">{{ task.first_review_result?.reason || '' }}</div>
                    </div>
                    
                    <div class="reviewer-info">
                      <span class="label">一审审核员：</span>
                      <span class="reviewer-name">{{ task.first_review_result?.reviewer?.username || '未知' }}</span>
                      <span class="review-time">{{ formatDate(task.first_review_result?.created_at || '') }}</span>
                    </div>
                  </div>
                </div>
              </div>
              
              <el-divider />
              
              <!-- 二审审核表单 -->
              <div class="second-review-section">
                <h3 class="section-title">二审审核</h3>
                <el-form label-position="top" size="default">
                  <el-form-item label="二审结果">
                    <el-radio-group v-model="getTaskReview(task.id)!.is_approved">
                      <el-radio :value="true">通过</el-radio>
                      <el-radio :value="false">不通过</el-radio>
                    </el-radio-group>
                  </el-form-item>
                  
                  <el-form-item
                    v-if="getTaskReview(task.id) && !getTaskReview(task.id)!.is_approved"
                    label="违规标签"
                  >
                    <el-checkbox-group v-model="getTaskReview(task.id)!.tags">
                      <el-checkbox
                        v-for="tag in availableTags"
                        :key="tag.id"
                        :label="tag.name"
                      >
                        {{ tag.name }}
                      </el-checkbox>
                    </el-checkbox-group>
                  </el-form-item>
                  
                  <el-form-item
                    v-if="getTaskReview(task.id) && !getTaskReview(task.id)!.is_approved"
                    label="二审原因"
                  >
                    <el-input
                      v-model="getTaskReview(task.id)!.reason"
                      type="textarea"
                      :rows="3"
                      placeholder="请输入二审不通过的原因"
                    />
                  </el-form-item>
                  
                  <el-form-item
                    v-if="getTaskReview(task.id) && getTaskReview(task.id)!.is_approved"
                    label="二审说明"
                  >
                    <el-input
                      v-model="getTaskReview(task.id)!.reason"
                      type="textarea"
                      :rows="2"
                      placeholder="请输入二审通过的说明（可选）"
                    />
                  </el-form-item>
                </el-form>
              </div>
              
              <div class="task-actions">
                <el-button
                  type="primary"
                  @click="handleSubmitSingle(task.id)"
                >
                  提交二审
                </el-button>
              </div>
            </div>
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useUserStore } from '../../stores/user'
import { 
  claimSecondReviewTasks, 
  getMySecondReviewTasks, 
  submitSecondReview, 
  submitBatchSecondReviews, 
  returnSecondReviewTasks 
} from '../../api/secondReview'
import { getTags } from '../../api/task'
import type { 
  SecondReviewTask, 
  SubmitSecondReviewRequest, 
  Tag 
} from '../../types'

const router = useRouter()
const userStore = useUserStore()

const claimLoading = ref(false)
const claimCount = ref(20)
const returnCount = ref(1)
const secondReviewTasks = ref<SecondReviewTask[]>([])
const availableTags = ref<Tag[]>([])
const reviews = reactive<Record<number, SubmitSecondReviewRequest>>({})
const currentQueueName = ref<string>('')
const completedToday = ref(0)
const approvalRate = ref(0)

const selectedReviews = computed(() => {
  return Object.entries(reviews)
    .filter(([_, review]) => review.is_approved !== null)
    .map(([taskId, review]) => ({
      ...review,
      task_id: parseInt(taskId),
    }))
})

// 获取任务的review对象，确保类型安全
const getTaskReview = (taskId: number) => {
  return reviews[taskId] || null
}

onMounted(async () => {
  // 从sessionStorage获取当前队列信息
  const queueStr = sessionStorage.getItem('currentQueue')
  if (queueStr) {
    try {
      const queue = JSON.parse(queueStr)
      currentQueueName.value = queue.queue_name
      sessionStorage.removeItem('currentQueue')
    } catch (e) {
      console.error('Failed to parse current queue:', e)
    }
  }
  
  try {
    await loadTags()
    await loadMyTasks()
    initReviews()
  } catch (error) {
    console.error('Failed to load data:', error)
  }
})

const loadTags = async () => {
  try {
    const response = await getTags()
    availableTags.value = response.tags
  } catch (error) {
    console.error('Failed to load tags:', error)
  }
}

const loadMyTasks = async () => {
  try {
    const response = await getMySecondReviewTasks()
    secondReviewTasks.value = response.tasks
  } catch (error) {
    console.error('Failed to load second review tasks:', error)
  }
}

const initReviews = () => {
  // 清除不存在的任务的 review
  const taskIds = new Set(secondReviewTasks.value.map(t => t.id))
  Object.keys(reviews).forEach(key => {
    if (!taskIds.has(parseInt(key))) {
      delete reviews[parseInt(key)]
    }
  })
  
  // 为新任务初始化 review
  secondReviewTasks.value.forEach((task) => {
    if (!reviews[task.id]) {
      reviews[task.id] = {
        task_id: task.id,
        is_approved: null as any,
        tags: [],
        reason: '',
      }
    }
  })
  
  // 重置退单数量为1
  returnCount.value = Math.min(1, secondReviewTasks.value.length)
}

const handleClaimTasks = async () => {
  if (claimCount.value < 1 || claimCount.value > 50) {
    ElMessage.warning('领取数量必须在 1-50 之间')
    return
  }
  
  claimLoading.value = true
  try {
    const res = await claimSecondReviewTasks(claimCount.value)
    ElMessage.success(`成功领取 ${res.count} 条二审任务`)
    await loadMyTasks()
    initReviews()
  } catch (error) {
    console.error('Failed to claim second review tasks:', error)
  } finally {
    claimLoading.value = false
  }
}

const handleRefresh = async () => {
  try {
    await loadMyTasks()
    initReviews()
    ElMessage.success('刷新成功')
  } catch (error) {
    console.error('Failed to refresh:', error)
  }
}

const validateReview = (review: SubmitSecondReviewRequest): boolean => {
  if (review.is_approved === null) {
    ElMessage.warning('请选择二审结果')
    return false
  }
  
  if (!review.is_approved) {
    if (review.tags.length === 0) {
      ElMessage.warning('不通过时必须选择至少一个违规标签')
      return false
    }
    if (!review.reason.trim()) {
      ElMessage.warning('不通过时必须填写原因')
      return false
    }
  }
  
  return true
}

const handleSubmitSingle = async (taskId: number) => {
  const review = reviews[taskId]
  if (!review || !validateReview(review)) return
  
  try {
    await submitSecondReview(review)
    ElMessage.success('提交成功')
    secondReviewTasks.value = secondReviewTasks.value.filter(t => t.id !== taskId)
    delete reviews[taskId]
  } catch (error) {
    console.error('Failed to submit second review:', error)
  }
}

const handleBatchSubmit = async () => {
  const validReviews: SubmitSecondReviewRequest[] = []
  
  for (const review of selectedReviews.value) {
    if (!validateReview(review)) {
      return
    }
    validReviews.push(review)
  }
  
  if (validReviews.length === 0) {
    ElMessage.warning('没有可提交的审核')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确认提交 ${validReviews.length} 条二审结果？`,
      '批量提交',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    const res = await submitBatchSecondReviews(validReviews)
    ElMessage.success(`成功提交 ${res.submitted} 条二审`)
    
    // Remove submitted tasks
    validReviews.forEach((review) => {
      secondReviewTasks.value = secondReviewTasks.value.filter(t => t.id !== review.task_id)
      delete reviews[review.task_id]
    })
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to batch submit:', error)
    }
  }
}

const handleReturnTasks = async () => {
  if (returnCount.value < 1) {
    ElMessage.warning('退单数量至少为 1')
    return
  }

  if (returnCount.value > secondReviewTasks.value.length) {
    ElMessage.warning(`退单数量不能超过当前任务数 (${secondReviewTasks.value.length})`)
    return
  }

  if (returnCount.value > 50) {
    ElMessage.warning('退单数量不能超过 50 条')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确认退回 ${returnCount.value} 条二审任务？将退回最早领取的任务。`,
      '退单确认',
      {
        confirmButtonText: '确认退单',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    // 取最早领取的N个任务ID
    const taskIdsToReturn = secondReviewTasks.value
      .slice(0, returnCount.value)
      .map(task => task.id)

    const res = await returnSecondReviewTasks(taskIdsToReturn)
    ElMessage.success(`成功退回 ${res.count} 条任务`)

    // Remove returned tasks from local state
    taskIdsToReturn.forEach((taskId) => {
      secondReviewTasks.value = secondReviewTasks.value.filter(t => t.id !== taskId)
      delete reviews[taskId]
    })
    
    // 刷新任务列表
    await loadMyTasks()
    initReviews()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to return tasks:', error)
    }
  }
}

const goToSearch = () => {
  router.push('/reviewer/search')
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

const formatDate = (dateStr: string) => {
  try {
    return new Date(dateStr).toLocaleString('zh-CN')
  } catch {
    return dateStr
  }
}
</script>

<style scoped>
/* ============================================
   二审审核工作台样式
   ============================================ */
.second-review-dashboard {
  min-height: 100vh;
  height: 100dvh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-100);
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
  flex-shrink: 0;
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

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
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
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  padding: var(--spacing-8);
  flex: 1;
  overflow-y: auto;
}

/* ============================================
   统计栏
   ============================================ */
.stats-bar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-5);
  margin-bottom: var(--spacing-6);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
  padding: var(--spacing-2);
}

.stat-label {
  font-size: var(--text-sm);
  color: var(--color-text-400);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
}

.stat-value {
  font-size: var(--text-4xl);
  font-weight: 700;
  color: var(--color-accent-main);
  line-height: var(--leading-tight);
  letter-spacing: var(--tracking-tight);
}

/* ============================================
   操作栏
   ============================================ */
.actions-bar {
  display: flex;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-6);
  flex-wrap: wrap;
  padding: var(--spacing-5);
  background: var(--color-bg-000);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-lighter);
  box-shadow: var(--shadow-sm);
}

.claim-section,
.return-section {
  display: flex;
  gap: var(--spacing-2);
  align-items: center;
}

/* ============================================
   空状态
   ============================================ */
.empty-state {
  padding: var(--spacing-20) 0;
  display: flex;
  justify-content: center;
  align-items: center;
}

/* ============================================
   任务卡片容器
   ============================================ */
.tasks-container {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
}

.task-card {
  transition: all var(--transition-base);
  border: 1px solid var(--color-border-lighter);
}

.task-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-lg);
  border-color: var(--color-accent-main);
}

/* ============================================
   任务内容区域
   ============================================ */
.task-content {
  padding: var(--spacing-4);
}

.section-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-100);
  margin: 0 0 var(--spacing-4) 0;
  padding-bottom: var(--spacing-2);
  border-bottom: 2px solid var(--color-accent-main);
}

/* ============================================
   评论内容区域
   ============================================ */
.comment-section {
  margin-bottom: var(--spacing-4);
}

.comment-text {
  font-family: var(--font-serif);
  font-size: var(--text-base);
  line-height: var(--leading-loose);
  color: var(--color-text-100);
  padding: var(--spacing-5);
  background: var(--color-bg-200);
  border-radius: var(--radius-md);
  letter-spacing: var(--tracking-wide);
  border-left: 4px solid var(--color-accent-main);
  word-break: break-word;
  white-space: pre-wrap;
}

/* ============================================
   一审结果区域
   ============================================ */
.first-review-section {
  margin-bottom: var(--spacing-4);
}

.first-review-info {
  padding: var(--spacing-4);
  background: var(--color-bg-200);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border-lighter);
}

.review-result {
  margin-bottom: var(--spacing-4);
}

.first-review-details {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.tags-section {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--spacing-2);
}

.reason-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.reason-text {
  padding: var(--spacing-3);
  background: var(--color-bg-000);
  border-radius: var(--radius-sm);
  border-left: 3px solid var(--color-danger);
  font-style: italic;
  color: var(--color-text-200);
}

.reviewer-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--text-sm);
  color: var(--color-text-300);
}

.label {
  font-weight: 600;
  color: var(--color-text-100);
}

.reviewer-name {
  color: var(--color-accent-main);
  font-weight: 500;
}

.review-time {
  color: var(--color-text-400);
}

/* ============================================
   二审审核区域
   ============================================ */
.second-review-section {
  margin-bottom: var(--spacing-4);
}

/* ============================================
   任务操作区域
   ============================================ */
.task-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-5);
  padding-top: var(--spacing-4);
  border-top: 1px solid var(--color-border-lighter);
}

/* ============================================
   响应式设计
   ============================================ */
@media (max-width: 768px) {
  .main-content {
    padding: var(--spacing-4);
  }

  .stats-bar {
    grid-template-columns: 1fr;
    gap: var(--spacing-3);
  }

  .actions-bar {
    flex-direction: column;
    gap: var(--spacing-2);
    padding: var(--spacing-3);
  }

  .claim-section,
  .return-section {
    width: 100%;
    flex-direction: column;
    gap: var(--spacing-2);
  }

  .stat-value {
    font-size: var(--text-3xl);
  }

  .comment-text {
    font-size: var(--text-sm);
    padding: var(--spacing-3);
  }

  .first-review-details {
    gap: var(--spacing-2);
  }

  .reviewer-info {
    flex-direction: column;
    align-items: flex-start;
  }
}

@media (max-width: 1024px) {
  .main-content {
    padding: var(--spacing-6);
  }
}
</style>