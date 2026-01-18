<template>
  <GenericReviewDashboard
    :config="dashboardConfig"
    :username="userStore.user?.username || ''"
    :task-count="qcTasks.length"
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
        v-for="task in qcTasks"
        :key="task.id"
        class="task-card"
        shadow="hover"
      >
        <div v-if="reviews[task.id]" class="task-content">
          <!-- 原始评论内容 -->
          <div class="comment-section">
            <h4 class="section-title">原始评论</h4>
            <div class="comment-text">
              {{ task.comment?.text || '评论内容加载中...' }}
            </div>
          </div>
          
          <el-divider />
          
          <!-- 一审审核结果 -->
          <div class="first-review-section">
            <h4 class="section-title">一审审核结果</h4>
            <div class="review-info">
              <div class="review-result">
                <span class="result-label">审核结果：</span>
                <el-tag :type="task.first_review_result?.is_approved ? 'success' : 'danger'">
                  {{ task.first_review_result?.is_approved ? '通过' : '不通过' }}
                </el-tag>
              </div>
              
              <div v-if="task.first_review_result?.tags && task.first_review_result.tags.length > 0" class="review-tags">
                <span class="result-label">违规标签：</span>
                <el-tag 
                  v-for="tag in task.first_review_result.tags" 
                  :key="tag" 
                  size="small" 
                  style="margin-right: 8px"
                >
                  {{ tag }}
                </el-tag>
              </div>
              
              <div v-if="task.first_review_result?.reason" class="review-reason">
                <span class="result-label">审核意见：</span>
                <span class="reason-text">{{ task.first_review_result.reason }}</span>
              </div>
              
              <div v-if="task.first_review_result?.reviewer" class="reviewer-info">
                <span class="result-label">审核员：</span>
                <span>{{ task.first_review_result.reviewer.username }}</span>
              </div>
            </div>
          </div>
          
          <el-divider />
          
          <!-- 质检操作表单 -->
          <div class="qc-section">
            <h4 class="section-title">质检操作</h4>
            <el-form label-position="top" size="default">
              <el-form-item label="质检判断">
                <el-radio-group v-model="getReview(task.id).is_passed">
                  <el-radio :value="true">
                    <span style="color: #67c23a">✅ 质检通过</span>
                  </el-radio>
                  <el-radio :value="false">
                    <span style="color: #f56c6c">❌ 质检不通过</span>
                  </el-radio>
                </el-radio-group>
              </el-form-item>
              
              <el-form-item
                v-if="!getReview(task.id).is_passed"
                label="错误类型"
              >
                <el-radio-group v-model="getReview(task.id).error_type">
                  <el-radio label="misjudgment">误判</el-radio>
                  <el-radio label="standard_deviation">标准执行偏差</el-radio>
                  <el-radio label="missing_violation">遗漏违规内容</el-radio>
                  <el-radio label="other">其他</el-radio>
                </el-radio-group>
              </el-form-item>
              
              <el-form-item
                v-if="!getReview(task.id).is_passed"
                label="质检意见"
              >
                <el-input
                  v-model="getReview(task.id).qc_comment"
                  type="textarea"
                  :rows="3"
                  placeholder="请详细说明问题所在"
                />
              </el-form-item>
            </el-form>
          </div>
          
          <div class="task-actions">
            <el-button
              type="primary"
              @click="handleSubmitSingle(task.id)"
            >
              提交质检
            </el-button>
          </div>
        </div>
      </el-card>
    </template>
  </GenericReviewDashboard>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import GenericReviewDashboard from '@/components/GenericReviewDashboard.vue'
import { qualityCheckDashboardConfig } from '@/config/dashboardConfigs'
import { handleClaimError, handleSubmitError, handleReturnError, handleLoadError } from '@/utils/errorHandler'
import { 
  claimQCTasks, 
  getMyQCTasks, 
  submitQCReview, 
  submitBatchQCReviews, 
  returnQCTasks,
  getQCStats
} from '@/api/qualityCheck'
import type { QCStats, SubmitQCRequest } from '@/types'

const router = useRouter()
const userStore = useUserStore()

// Dashboard config
const dashboardConfig = qualityCheckDashboardConfig

// State
const claimLoading = ref(false)
const claimCount = ref(20)
const returnCount = ref(1)
const qcTasks = ref<any[]>([])
const qcStats = ref<QCStats | null>(null)
const reviews = reactive<Record<number, SubmitQCRequest>>({})

// Stats data for the dashboard
const statsData = computed(() => ({
  pending_tasks: qcStats.value?.pending_tasks || 0,
  today_completed: qcStats.value?.today_completed || 0,
  total_completed: qcStats.value?.total_completed || 0,
  pass_rate: qcStats.value?.pass_rate || 0
}))

// Safe access to reviews
const getReview = (taskId: number) => {
  if (!reviews[taskId]) {
    reviews[taskId] = {
      task_id: taskId,
      is_passed: null as any,
      error_type: '',
      qc_comment: '',
    }
  }
  return reviews[taskId]
}

const selectedReviews = computed(() => {
  return Object.entries(reviews)
    .filter(([_, review]) => review.is_passed !== null)
    .map(([taskId, review]) => ({
      ...review,
      task_id: parseInt(taskId),
    }))
})

// Load stats
const loadQCStats = async () => {
  try {
    const response = await getQCStats()
    qcStats.value = response
  } catch (error) {
    console.error('Failed to load QC stats:', error)
  }
}

// Load tasks
const loadQCTasks = async () => {
  try {
    const response = await getMyQCTasks()
    qcTasks.value = response.tasks
  } catch (error) {
    handleLoadError(error)
  }
}

const initReviews = () => {
  // 清除不存在的任务的 review
  const taskIds = new Set(qcTasks.value.map(t => t.id))
  Object.keys(reviews).forEach(key => {
    if (!taskIds.has(parseInt(key))) {
      delete reviews[parseInt(key)]
    }
  })
  
  // 为新任务初始化 review
  qcTasks.value.forEach((task) => {
    if (!reviews[task.id]) {
      reviews[task.id] = {
        task_id: task.id,
        is_passed: null as any,
        error_type: undefined,
        qc_comment: '',
      }
    }
  })
  
  // 重置退单数量为1
  returnCount.value = Math.min(1, qcTasks.value.length)
}

// Claim tasks
const handleClaimTasks = async (count: number) => {
  if (count < 1 || count > 50) {
    ElMessage.warning('领取数量必须在 1-50 之间')
    return
  }
  
  claimLoading.value = true
  try {
    const res = await claimQCTasks(count)
    ElMessage.success(`成功领取 ${res.count} 条质检任务`)
    await loadQCTasks()
    await loadQCStats()
    initReviews()
  } catch (error) {
    handleClaimError(error)
  } finally {
    claimLoading.value = false
  }
}

// Refresh
const handleRefresh = async () => {
  try {
    await loadQCTasks()
    await loadQCStats()
    initReviews()
    ElMessage.success('刷新成功')
  } catch (error) {
    console.error('Failed to refresh:', error)
  }
}

const validateReview = (review: SubmitQCRequest): boolean => {
  if (review.is_passed === null) {
    ElMessage.warning('请选择质检判断')
    return false
  }
  
  if (!review.is_passed) {
    if (!review.error_type) {
      ElMessage.warning('质检不通过时必须选择错误类型')
      return false
    }
    if (!review.qc_comment?.trim()) {
      ElMessage.warning('质检不通过时必须填写质检意见')
      return false
    }
  }
  
  return true
}

// Submit single
const handleSubmitSingle = async (taskId: number) => {
  const review = reviews[taskId]
  if (!review || !validateReview(review)) return
  
  try {
    await submitQCReview(review)
    ElMessage.success('提交成功')
    qcTasks.value = qcTasks.value.filter(t => t.id !== taskId)
    delete reviews[taskId]
    await loadQCStats()
  } catch (error) {
    handleSubmitError(error)
  }
}

// Batch submit
const handleBatchSubmit = async () => {
  const validReviews: SubmitQCRequest[] = []
  
  for (const review of selectedReviews.value) {
    if (!validateReview(review)) {
      return
    }
    validReviews.push(review)
  }
  
  if (validReviews.length === 0) {
    ElMessage.warning('没有可提交的质检')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确认提交 ${validReviews.length} 条质检结果？`,
      '批量提交',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    const res = await submitBatchQCReviews(validReviews)
    ElMessage.success(`成功提交 ${res.submitted} 条质检`)
    
    // Remove submitted tasks
    validReviews.forEach((review) => {
      qcTasks.value = qcTasks.value.filter(t => t.id !== review.task_id)
      delete reviews[review.task_id]
    })
    
    await loadQCStats()
  } catch (error) {
    if (error !== 'cancel') {
      handleSubmitError(error)
    }
  }
}

// Return tasks
const handleReturnTasks = async (count: number) => {
  if (count < 1) {
    ElMessage.warning('退单数量至少为 1')
    return
  }

  if (count > qcTasks.value.length) {
    ElMessage.warning(`退单数量不能超过当前任务数 (${qcTasks.value.length})`)
    return
  }

  if (count > 50) {
    ElMessage.warning('退单数量不能超过 50 条')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确认退回 ${count} 条质检任务？将退回最早领取的任务。`,
      '退单确认',
      {
        confirmButtonText: '确认退单',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    // 取最早领取的N个任务ID
    const taskIdsToReturn = qcTasks.value
      .slice(0, count)
      .map(task => task.id)

    const res = await returnQCTasks(taskIdsToReturn)
    ElMessage.success(`成功退回 ${res.count} 条质检任务`)

    // Remove returned tasks from local state
    taskIdsToReturn.forEach((taskId) => {
      qcTasks.value = qcTasks.value.filter(t => t.id !== taskId)
      delete reviews[taskId]
    })
    
    // 刷新任务列表
    await loadQCTasks()
    await loadQCStats()
    initReviews()
  } catch (error) {
    if (error !== 'cancel') {
      handleReturnError(error)
    }
  }
}

// Navigation
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

// Lifecycle
onMounted(async () => {
  try {
    await loadQCStats()
    await loadQCTasks()
    initReviews()
  } catch (error) {
    console.error('Failed to load data:', error)
  }
})
</script>


<style scoped>
/* ============================================
   任务卡片容器
   ============================================ */
.task-card {
  transition: all var(--transition-base, 0.3s ease);
  border: 1px solid var(--color-border-lighter, #e4e7ed);
}

.task-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-lg, 0 10px 15px -3px rgba(0, 0, 0, 0.1));
  border-color: var(--color-accent-main, #409eff);
}

/* ============================================
   任务内容区域
   ============================================ */
.task-content {
  padding: var(--spacing-4, 16px);
}

.section-title {
  font-size: var(--text-lg, 18px);
  font-weight: 600;
  color: var(--color-text-000, #333);
  margin: 0 0 var(--spacing-3, 12px) 0;
  letter-spacing: var(--tracking-tight, -0.025em);
}

/* ============================================
   评论区域
   ============================================ */
.comment-section {
  margin-bottom: var(--spacing-4, 16px);
}

.comment-text {
  font-family: var(--font-serif, serif);
  font-size: var(--text-base, 16px);
  line-height: var(--leading-loose, 2);
  color: var(--color-text-100, #444);
  padding: var(--spacing-5, 20px);
  background: var(--color-bg-200, #f5f5f5);
  border-radius: var(--radius-md, 8px);
  letter-spacing: var(--tracking-wide, 0.025em);
  border-left: 4px solid var(--color-accent-main, #409eff);
  word-break: break-word;
  white-space: pre-wrap;
}

/* ============================================
   一审审核结果区域
   ============================================ */
.first-review-section {
  margin-bottom: var(--spacing-4, 16px);
}

.review-info {
  background: var(--color-bg-200, #f5f5f5);
  padding: var(--spacing-4, 16px);
  border-radius: var(--radius-md, 8px);
  border: 1px solid var(--color-border-lighter, #e4e7ed);
}

.review-result,
.review-tags,
.review-reason,
.reviewer-info {
  margin-bottom: var(--spacing-3, 12px);
}

.review-result:last-child,
.review-tags:last-child,
.review-reason:last-child,
.reviewer-info:last-child {
  margin-bottom: 0;
}

.result-label {
  font-weight: 600;
  color: var(--color-text-200, #666);
  margin-right: var(--spacing-2, 8px);
}

.reason-text {
  color: var(--color-text-100, #444);
  line-height: var(--leading-relaxed, 1.625);
}

/* ============================================
   质检操作区域
   ============================================ */
.qc-section {
  margin-bottom: var(--spacing-4, 16px);
}

/* ============================================
   任务操作区域
   ============================================ */
.task-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-5, 20px);
  padding-top: var(--spacing-4, 16px);
  border-top: 1px solid var(--color-border-lighter, #e4e7ed);
}

/* ============================================
   响应式设计
   ============================================ */
@media (max-width: 768px) {
  .task-content {
    padding: var(--spacing-3, 12px);
  }

  .comment-text {
    font-size: var(--text-sm, 14px);
    padding: var(--spacing-3, 12px);
  }
}
</style>
