<template>
  <div class="reviewer-dashboard">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <div class="header-left">
            <h2>{{ currentTaskName ? `${currentTaskName} - 标注工作台` : '审核员工作台' }}</h2>
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
        <div class="stats-inline">
          <div class="stat-chip">
            <span class="stat-label">待审核任务</span>
            <span class="stat-value">{{ taskStore.tasks.length }}</span>
          </div>
          <div class="stat-chip">
            <span class="stat-label">今日已完成</span>
            <span class="stat-value">{{ completedToday }}</span>
          </div>
        </div>

        <div class="progress-bar">
          <el-progress
            :percentage="sessionProgress"
            :format="() => `${completedToday}/${sessionTotal}`"
          />
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
              :max="Math.max(1, taskStore.tasks.length)"
              :step="1"
              size="large"
              style="width: 120px"
              :disabled="taskStore.tasks.length === 0"
            />
            <el-button
              type="warning"
              size="large"
              :disabled="taskStore.tasks.length === 0"
              @click="handleReturnTasks"
            >
              退单
            </el-button>
          </div>
          
          <el-button
            size="large"
            :disabled="selectedReviews.length === 0 || batchLoading"
            :loading="batchLoading"
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
        
        <div v-if="taskStore.tasks.length === 0" class="empty-state">
          <el-empty description="暂无待审核任务，点击「领取新任务」开始工作" />
        </div>
        
        <div v-else class="tasks-container">
          <el-card
            v-for="task in taskStore.tasks"
            :key="task.id"
            class="task-card"
            :class="{
              'is-filled': isReviewReady(task.id),
              'is-selected': batchSelection[task.id],
              'is-active': activeTaskId === task.id
            }"
            shadow="hover"
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

              <div class="comment-text">
                {{ task.comment?.text || '评论内容加载中...' }}
              </div>
              
              <el-divider />
              
              <div class="quick-actions">
                <el-button size="small" @click="applyQuickReject(task.id, '垃圾', '内容属于垃圾信息')">
                  垃圾信息
                </el-button>
                <el-button size="small" @click="applyQuickReject(task.id, '广告', '内容包含广告或推广')">
                  广告推广
                </el-button>
                <el-button size="small" @click="applyQuickReject(task.id, '人身攻击', '内容包含辱骂/人身攻击')">
                  辱骂攻击
                </el-button>
              </div>

              <el-form label-position="top" size="default">
                <el-form-item label="审核结果" required>
                  <el-radio-group v-model="getReview(task.id).is_approved">
                    <el-radio :value="true">通过</el-radio>
                    <el-radio :value="false">不通过</el-radio>
                  </el-radio-group>
                </el-form-item>
                
                <el-form-item
                  v-if="!getReview(task.id).is_approved"
                  label="违规标签"
                  required
                >
                  <el-checkbox-group v-model="getReview(task.id).tags">
                    <el-checkbox
                      v-for="tag in taskStore.tags"
                      :key="tag.id"
                      :label="tag.name"
                    >
                      {{ tag.name }}
                    </el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
                
                <el-form-item
                  v-if="!getReview(task.id).is_approved"
                  label="原因"
                  required
                >
                  <el-input
                    v-model="getReview(task.id).reason"
                    type="textarea"
                    :rows="2"
                    placeholder="请输入不通过的原因"
                  />
                </el-form-item>
              </el-form>
              
              <div class="task-actions">
                <el-button
                  type="primary"
                  :loading="submitLoading[task.id]"
                  @click="handleSubmitSingle(task.id)"
                >
                  提交审核
                </el-button>
                <el-button @click="clearReviewForm(task.id)">清空</el-button>
              </div>
            </div>
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useUserStore } from '../../stores/user'
import { useTaskStore } from '../../stores/task'
import { claimTasks, submitReview, submitBatchReviews, returnTasks } from '../../api/task'
import type { ReviewResult } from '../../types'

const router = useRouter()
const userStore = useUserStore()
const taskStore = useTaskStore()

const claimLoading = ref(false)
const claimCount = ref(20)
const returnCount = ref(1)
const reviews = reactive<Record<number, ReviewResult>>({})
const batchSelection = reactive<Record<number, boolean>>({})
const submitLoading = reactive<Record<number, boolean>>({})
const batchLoading = ref(false)
const currentTaskName = ref<string>('')
const activeTaskId = ref<number | null>(null)
const completedToday = ref(0)
const sessionTotal = computed(() => completedToday.value + taskStore.tasks.length)
const sessionProgress = computed(() => {
  return sessionTotal.value ? Math.round((completedToday.value / sessionTotal.value) * 100) : 0
})

const draftStorageKey = 'reviewer_dashboard_drafts'
const statsStorageKey = 'reviewer_dashboard_stats'

const loadTodayStats = () => {
  const raw = localStorage.getItem(statsStorageKey)
  if (!raw) {
    completedToday.value = 0
    return
  }

  try {
    const parsed = JSON.parse(raw)
    const today = new Date().toISOString().slice(0, 10)
    if (parsed?.date !== today) {
      completedToday.value = 0
      localStorage.setItem(statsStorageKey, JSON.stringify({ date: today, completed: 0 }))
      return
    }
    completedToday.value = parsed.completed || 0
  } catch (error) {
    console.error('Failed to parse stats:', error)
    completedToday.value = 0
  }
}

const incrementTodayCompleted = (count: number) => {
  const today = new Date().toISOString().slice(0, 10)
  completedToday.value += count
  localStorage.setItem(statsStorageKey, JSON.stringify({ date: today, completed: completedToday.value }))
}

// Safe access to reviews
const getReview = (taskId: number) => {
  if (!reviews[taskId]) {
    reviews[taskId] = {
      task_id: taskId,
      is_approved: null as any,
      tags: [],
      reason: '',
    }
  }
  return reviews[taskId]
}

const selectedReviews = computed(() => {
  return Object.entries(reviews)
    .filter(([taskId, review]) => batchSelection[parseInt(taskId)] && isReviewReady(parseInt(taskId)))
    .map(([taskId, review]) => ({
      ...review,
      task_id: parseInt(taskId),
    }))
})

const isReviewReady = (taskId: number) => {
  const review = reviews[taskId]
  if (!review) return false
  if (review.is_approved === null) return false
  if (!review.is_approved) {
    return review.tags.length > 0 && review.reason.trim().length > 0
  }
  return true
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
  const ids = taskStore.tasks.map(task => task.id)
  if (ids.length === 0) return
  const currentId = activeTaskId.value ?? ids[0]
  const currentIndex = ids.indexOf(currentId)
  const nextIndex = direction === 'next' ? currentIndex + 1 : currentIndex - 1
  const targetId = ids[Math.min(Math.max(nextIndex, 0), ids.length - 1)]
  activeTaskId.value = targetId
  scrollToTask(targetId)
}

const clearReviewForm = (taskId: number) => {
  reviews[taskId] = {
    task_id: taskId,
    is_approved: null as any,
    tags: [],
    reason: '',
  }
  batchSelection[taskId] = false
}

const applyQuickReject = (taskId: number, tagName: string, reason: string) => {
  const review = getReview(taskId)
  review.is_approved = false as any
  const tagExists = taskStore.tags.some(tag => tag.name === tagName)
  review.tags = tagExists ? [tagName] : review.tags
  review.reason = reason
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

  if (taskStore.tasks.length === 0) return
  const currentId = activeTaskId.value ?? taskStore.tasks[0]?.id
  if (!currentId) return

  switch (event.key) {
    case '1':
      getReview(currentId).is_approved = true as any
      break
    case '2':
      getReview(currentId).is_approved = false as any
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
    case 'q':
    case 'Q':
      event.preventDefault()
      applyQuickReject(currentId, '垃圾', '内容属于垃圾信息')
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
    const taskIds = new Set(taskStore.tasks.map(t => t.id))

    Object.entries(savedReviews).forEach(([taskId, review]) => {
      const id = Number(taskId)
      if (!taskIds.has(id)) return
      reviews[id] = { ...getReview(id), ...(review as ReviewResult) }
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
      const payload = {
        reviews,
        batchSelection
      }
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
      const payload = {
        reviews,
        batchSelection
      }
      localStorage.setItem(draftStorageKey, JSON.stringify(payload))
    }, 500)
  },
  { deep: true }
)

watch(
  () => taskStore.tasks.length,
  (count) => {
    document.title = `(${count}) 审核工作台`
  },
  { immediate: true }
)

onMounted(async () => {
  // 从sessionStorage获取当前任务信息
  const taskStr = sessionStorage.getItem('currentTask')
  if (taskStr) {
    try {
      const task = JSON.parse(taskStr)
      currentTaskName.value = task.taskName
      sessionStorage.removeItem('currentTask')
    } catch (e) {
      console.error('Failed to parse current task:', e)
    }
  }
  
  try {
    await taskStore.fetchTags()
    await taskStore.fetchMyTasks()
    initReviews()
    restoreDrafts()
    loadTodayStats()
  } catch (error) {
    console.error('Failed to load data:', error)
  }

  window.addEventListener('keydown', handleKeyPress)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyPress)
})

const initReviews = () => {
  // 清除不存在的任务的 review
  const taskIds = new Set(taskStore.tasks.map(t => t.id))
  Object.keys(reviews).forEach(key => {
    if (!taskIds.has(parseInt(key))) {
      delete reviews[parseInt(key)]
    }
  })

  Object.keys(batchSelection).forEach((key) => {
    if (!taskIds.has(parseInt(key))) {
      delete batchSelection[parseInt(key)]
    }
  })
  
  // 为新任务初始化 review
  taskStore.tasks.forEach((task) => {
    if (!reviews[task.id]) {
      reviews[task.id] = {
        task_id: task.id,
        is_approved: null as any,
        tags: [],
        reason: '',
      }
    }

    if (batchSelection[task.id] === undefined) {
      batchSelection[task.id] = false
    }
  })
  
  // 重置退单数量为1
  returnCount.value = Math.min(1, taskStore.tasks.length)

  if (taskStore.tasks.length > 0 && !taskIds.has(activeTaskId.value || -1)) {
    activeTaskId.value = taskStore.tasks[0].id
  }
}

const handleClaimTasks = async () => {
  if (claimCount.value < 1 || claimCount.value > 50) {
    ElMessage.warning('领取数量必须在 1-50 之间')
    return
  }
  
  claimLoading.value = true
  try {
    const res = await claimTasks(claimCount.value)
    ElMessage.success(`成功领取 ${res.count} 条任务`)
    await taskStore.fetchMyTasks()
    initReviews()
    restoreDrafts()
  } catch (error) {
    console.error('Failed to claim tasks:', error)
    ElMessage.error('领取任务失败，请稍后重试')
  } finally {
    claimLoading.value = false
  }
}

const handleRefresh = async () => {
  try {
    await taskStore.fetchMyTasks()
    initReviews()
    restoreDrafts()
    ElMessage.success('刷新成功')
  } catch (error) {
    console.error('Failed to refresh:', error)
    ElMessage.error('刷新失败，请稍后重试')
  }
}

const validateReview = (review: ReviewResult): boolean => {
  if (review.is_approved === null) {
    ElMessage.warning('请选择审核结果')
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
    submitLoading[taskId] = true
    await submitReview(review)
    ElMessage.success('提交成功')
    taskStore.removeTask(taskId)
    delete reviews[taskId]
    batchSelection[taskId] = false
    incrementTodayCompleted(1)
    await taskStore.fetchMyTasks()
    initReviews()
    restoreDrafts()
  } catch (error) {
    console.error('Failed to submit review:', error)
    ElMessage.error('提交失败，请稍后重试')
  } finally {
    submitLoading[taskId] = false
  }
}

const handleBatchSubmit = async () => {
  const validReviews: ReviewResult[] = []
  
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
      `确认提交 ${validReviews.length} 条审核结果？`,
      '批量提交',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    batchLoading.value = true
    const res = await submitBatchReviews(validReviews)
    ElMessage.success(`成功提交 ${res.submitted} 条审核`)

    validReviews.forEach((review) => {
      taskStore.removeTask(review.task_id)
      delete reviews[review.task_id]
      batchSelection[review.task_id] = false
    })
    incrementTodayCompleted(res.submitted || validReviews.length)
    await taskStore.fetchMyTasks()
    initReviews()
    restoreDrafts()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to batch submit:', error)
      ElMessage.error('批量提交失败，请稍后重试')
    }
  } finally {
    batchLoading.value = false
  }
}

const handleReturnTasks = async () => {
  if (returnCount.value < 1) {
    ElMessage.warning('退单数量至少为 1')
    return
  }

  if (returnCount.value > taskStore.tasks.length) {
    ElMessage.warning(`退单数量不能超过当前任务数 (${taskStore.tasks.length})`)
    return
  }

  if (returnCount.value > 50) {
    ElMessage.warning('退单数量不能超过 50 条')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确认退回 ${returnCount.value} 条任务？将退回最早领取的任务。`,
      '退单确认',
      {
        confirmButtonText: '确认退单',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    // 取最早领取的N个任务ID
    const taskIdsToReturn = taskStore.tasks
      .slice(0, returnCount.value)
      .map(task => task.id)

    const res = await returnTasks(taskIdsToReturn)
    ElMessage.success(`成功退回 ${res.count} 条任务`)

    // Remove returned tasks from local state
    taskIdsToReturn.forEach((taskId) => {
      taskStore.removeTask(taskId)
      delete reviews[taskId]
    })
    
    // 刷新任务列表
    await taskStore.fetchMyTasks()
    initReviews()
    restoreDrafts()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to return tasks:', error)
      ElMessage.error('退单失败，请稍后重试')
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
</script>

<style scoped>
/* ============================================
   审核员工作台样式
   ============================================ */
.reviewer-dashboard {
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
.stats-inline {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  margin-bottom: var(--spacing-4);
  flex-wrap: wrap;
}

.progress-bar {
  margin-bottom: var(--spacing-5);
  padding: 0 var(--spacing-2);
}

.stat-chip {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-2) var(--spacing-4);
  background: var(--color-bg-000);
  border: 1px solid var(--color-border-lighter);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
}

.stat-label {
  font-size: var(--text-sm);
  color: var(--color-text-400);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
}

.stat-value {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-accent-main);
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
  position: relative;
}

.task-card.is-filled {
  border-color: rgba(64, 158, 255, 0.5);
}

.task-card.is-selected {
  border-color: var(--color-accent-main);
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.task-card.is-active {
  border-color: var(--color-success);
  box-shadow: 0 0 0 2px rgba(103, 194, 58, 0.2);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-3);
}

.task-id {
  font-weight: 600;
  color: var(--color-text-100);
}

.task-header-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
}

.quick-actions {
  display: flex;
  gap: var(--spacing-2);
  flex-wrap: wrap;
  margin-bottom: var(--spacing-3);
}

.task-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-lg);
  border-color: var(--color-accent-main);
}

/* ============================================
   评论文本区域
   ============================================ */
.comment-text {
  font-family: var(--font-serif);
  font-size: var(--text-base);
  line-height: var(--leading-loose);
  color: var(--color-text-100);
  padding: var(--spacing-5);
  background: var(--color-bg-200);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-4);
  letter-spacing: var(--tracking-wide);
  border-left: 4px solid var(--color-accent-main);
  word-break: break-word;
  white-space: pre-wrap;
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

  .stats-inline {
    gap: var(--spacing-2);
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
}

@media (max-width: 1024px) {
  .main-content {
    padding: var(--spacing-6);
  }
}
</style>

