<template>
  <GenericReviewDashboard
    :config="dashboardConfig"
    :username="userStore.user?.username || ''"
    :task-count="qcTasks.length"
    :selected-count="selectedReviews.length"
    :claim-loading="claimLoading"
    :batch-loading="batchLoading"
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
        :class="{
          'is-selected': batchSelection[task.id],
          'is-active': activeTaskId === task.id,
          'is-filled': isReviewReady(task.id)
        }"
        :data-task-id="task.id"
        @click="setActiveTask(task.id)"
        @focusin="setActiveTask(task.id)"
      >
        <div v-if="reviews[task.id]" class="task-content">
          <div class="task-header">
            <span class="task-id">任务 #{{ task.id }}</span>
            <div class="task-header-actions">
              <el-tag v-if="isReviewReady(task.id)" type="success" size="small">已填写</el-tag>
              <el-checkbox
                v-model="batchSelection[task.id]"
                :disabled="!isReviewReady(task.id)"
              >
                批量提交
              </el-checkbox>
            </div>
          </div>

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
              <el-form-item label="质检判断" required>
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
                required
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
                required
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
              :loading="submitLoading[task.id]"
              @click="handleSubmitSingle(task.id)"
            >
              提交质检
            </el-button>
            <el-button @click="clearReviewForm(task.id)">清空</el-button>
          </div>
        </div>
      </el-card>
    </template>
  </GenericReviewDashboard>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
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
const batchSelection = reactive<Record<number, boolean>>({})
const submitLoading = reactive<Record<number, boolean>>({})
const batchLoading = ref(false)
const activeTaskId = ref<number | null>(null)

const draftStorageKey = 'quality_check_drafts'

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
    .filter(([taskId]) => batchSelection[parseInt(taskId)] && isReviewReady(parseInt(taskId)))
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

  Object.keys(batchSelection).forEach(key => {
    if (!taskIds.has(parseInt(key))) {
      delete batchSelection[parseInt(key)]
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

    if (batchSelection[task.id] === undefined) {
      batchSelection[task.id] = false
    }
  })
  
  // 重置退单数量为1
  returnCount.value = Math.min(1, qcTasks.value.length)

  if (qcTasks.value.length > 0 && !taskIds.has(activeTaskId.value || -1)) {
    activeTaskId.value = qcTasks.value[0].id
  }
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
    restoreDrafts()
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
    restoreDrafts()
    ElMessage.success('刷新成功')
  } catch (error) {
    console.error('Failed to refresh:', error)
    ElMessage.error('刷新失败，请稍后重试')
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

const isReviewReady = (taskId: number) => {
  const review = reviews[taskId]
  if (!review) return false
  if (review.is_passed === null) return false
  if (!review.is_passed) {
    return Boolean(review.error_type) && Boolean(review.qc_comment?.trim())
  }
  return true
}

const clearReviewForm = (taskId: number) => {
  reviews[taskId] = {
    task_id: taskId,
    is_passed: null as any,
    error_type: '',
    qc_comment: '',
  }
  batchSelection[taskId] = false
}

const setActiveTask = (taskId: number) => {
  activeTaskId.value = taskId
}

const scrollToTask = async (taskId: number) => {
  await nextTick()
  const el = document.querySelector(`[data-task-id="${taskId}"]`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
}

const moveActiveTask = (direction: 'next' | 'prev') => {
  const ids = qcTasks.value.map(task => task.id)
  if (ids.length === 0) return
  const currentId = activeTaskId.value ?? ids[0]
  const currentIndex = ids.indexOf(currentId)
  const nextIndex = direction === 'next' ? currentIndex + 1 : currentIndex - 1
  const targetId = ids[Math.min(Math.max(nextIndex, 0), ids.length - 1)]
  activeTaskId.value = targetId
  scrollToTask(targetId)
}

const handleKeyPress = (event: KeyboardEvent) => {
  const target = event.target as HTMLElement | null
  const targetTag = target?.tagName?.toLowerCase()
  const isTyping = targetTag === 'input' || targetTag === 'textarea' || target?.isContentEditable
  if (isTyping) {
    if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
      event.preventDefault()
      handleBatchSubmit()
    }
    return
  }

  if (qcTasks.value.length === 0) return
  const currentId = activeTaskId.value ?? qcTasks.value[0]?.id
  if (!currentId) return

  switch (event.key) {
    case '1':
      getReview(currentId).is_passed = true as any
      break
    case '2':
      getReview(currentId).is_passed = false as any
      break
    case 'Enter':
      event.preventDefault()
      handleSubmitSingle(currentId)
      break
    case 'Escape':
      event.preventDefault()
      clearReviewForm(currentId)
      break
    case 'r':
    case 'R':
      event.preventDefault()
      handleRefresh()
      break
    case 'Tab':
      event.preventDefault()
      moveActiveTask(event.shiftKey ? 'prev' : 'next')
      break
    default:
      break
  }
}

const restoreDrafts = () => {
  const raw = localStorage.getItem(draftStorageKey)
  if (!raw) return

  try {
    const saved = JSON.parse(raw)
    const savedReviews = saved?.reviews || saved
    const savedSelection = saved?.batchSelection || {}
    const taskIds = new Set(qcTasks.value.map(t => t.id))

    Object.entries(savedReviews).forEach(([taskId, review]) => {
      const id = Number(taskId)
      if (!taskIds.has(id)) return
      reviews[id] = { ...getReview(id), ...(review as SubmitQCRequest) }
    })

    Object.entries(savedSelection).forEach(([taskId, value]) => {
      const id = Number(taskId)
      if (!taskIds.has(id)) return
      batchSelection[id] = Boolean(value)
    })
  } catch (error) {
    console.error('Failed to restore drafts:', error)
  }
}

let draftTimer: number | undefined

watch(
  reviews,
  () => {
    if (draftTimer) window.clearTimeout(draftTimer)
    draftTimer = window.setTimeout(() => {
      const payload = { reviews, batchSelection }
      localStorage.setItem(draftStorageKey, JSON.stringify(payload))
    }, 500)
  },
  { deep: true }
)

watch(
  batchSelection,
  () => {
    if (draftTimer) window.clearTimeout(draftTimer)
    draftTimer = window.setTimeout(() => {
      const payload = { reviews, batchSelection }
      localStorage.setItem(draftStorageKey, JSON.stringify(payload))
    }, 500)
  },
  { deep: true }
)

watch(
  () => qcTasks.value.length,
  (count) => {
    document.title = `(${count}) 质检工作台`
  },
  { immediate: true }
)

// Submit single
const handleSubmitSingle = async (taskId: number) => {
  const review = reviews[taskId]
  if (!review || !validateReview(review)) return
  
  try {
    submitLoading[taskId] = true
    await submitQCReview(review)
    ElMessage.success('提交成功')
    qcTasks.value = qcTasks.value.filter(t => t.id !== taskId)
    delete reviews[taskId]
    batchSelection[taskId] = false
    await loadQCTasks()
    initReviews()
    restoreDrafts()
    await loadQCStats()
  } catch (error) {
    handleSubmitError(error)
  } finally {
    submitLoading[taskId] = false
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
    
    batchLoading.value = true
    const res = await submitBatchQCReviews(validReviews)
    ElMessage.success(`成功提交 ${res.submitted} 条质检`)

    validReviews.forEach((review) => {
      qcTasks.value = qcTasks.value.filter(t => t.id !== review.task_id)
      delete reviews[review.task_id]
      batchSelection[review.task_id] = false
    })

    await loadQCTasks()
    initReviews()
    restoreDrafts()
    await loadQCStats()
  } catch (error) {
    if (error !== 'cancel') {
      handleSubmitError(error)
    }
  } finally {
    batchLoading.value = false
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
      batchSelection[taskId] = false
    })
    
    // 刷新任务列表
    await loadQCTasks()
    await loadQCStats()
    initReviews()
    restoreDrafts()
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
    restoreDrafts()
  } catch (error) {
    console.error('Failed to load data:', error)
  }

  window.addEventListener('keydown', handleKeyPress)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyPress)
})
</script>


<style scoped>
/* ============================================
   任务卡片容器
   ============================================ */
.task-card {
  transition: all var(--transition-base, 0.3s ease);
  border: 1px solid var(--color-border-lighter, #e4e7ed);
  position: relative;
}

.task-card.is-filled {
  border-color: rgba(64, 158, 255, 0.5);
}

.task-card.is-selected {
  border-color: var(--color-accent-main, #409eff);
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.task-card.is-active {
  border-color: var(--color-success, #67c23a);
  box-shadow: 0 0 0 2px rgba(103, 194, 58, 0.2);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-3, 12px);
}

.task-id {
  font-weight: 600;
  color: var(--color-text-100, #444);
}

.task-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
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
