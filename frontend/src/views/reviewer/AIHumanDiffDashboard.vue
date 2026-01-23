<template>
  <div class="ai-human-diff-dashboard">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <div class="header-left">
            <h2>{{ currentQueueName ? `${currentQueueName} - 差异工作台` : 'AI 与人工差异工作台' }}</h2>
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
              <span class="stat-label">待处理差异</span>
              <span class="stat-value">{{ diffTasks.length }}</span>
            </div>
          </el-card>

          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">平均置信度</span>
              <span class="stat-value">{{ avgConfidence }}%</span>
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
              领取差异任务
            </el-button>
          </div>

          <div class="return-section">
            <el-input-number
              v-model="returnCount"
              :min="1"
              :max="Math.max(1, diffTasks.length)"
              :step="1"
              size="large"
              style="width: 120px"
              :disabled="diffTasks.length === 0"
            />
            <el-button
              type="warning"
              size="large"
              :disabled="diffTasks.length === 0"
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

          <el-button size="large" @click="handleRefresh">刷新任务列表</el-button>
        </div>

        <div v-if="diffTasks.length === 0" class="empty-state">
          <el-empty description="暂无差异任务，点击「领取差异任务」开始工作" />
        </div>

        <div v-else class="tasks-container">
          <el-card
            v-for="task in diffTasks"
            :key="task.id"
            class="diff-card"
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

              <div class="comment-section">
                <h3 class="section-title">评论内容</h3>
                <div class="comment-text">
                  {{ task.comment?.text || '评论内容加载中...' }}
                </div>
              </div>

              <div class="diff-panels">
                <div class="diff-panel human-panel">
                  <div class="panel-header">
                    <span>人工审核</span>
                    <el-tag :type="task.human_review_result?.is_approved ? 'success' : 'danger'" size="small">
                      {{ task.human_review_result?.is_approved ? '通过' : '不通过' }}
                    </el-tag>
                  </div>
                  <div class="panel-body">
                    <div class="panel-line" v-if="task.human_review_result">
                      <span class="label">标签：</span>
                      <div class="tag-list">
                        <el-tag
                          v-for="tag in (task.human_review_result?.tags || [])"
                          :key="tag"
                          type="danger"
                          size="small"
                        >
                          {{ tag }}
                        </el-tag>
                        <span v-if="(task.human_review_result?.tags || []).length === 0">无</span>
                      </div>
                    </div>
                    <div class="panel-line">
                      <span class="label">原因：</span>
                      <span class="text">{{ task.human_review_result?.reason || '无' }}</span>
                    </div>
                    <div class="panel-line">
                      <span class="label">审核员：</span>
                      <span class="text">{{ task.human_review_result?.reviewer?.username || '未知' }}</span>
                    </div>
                    <div class="panel-line">
                      <span class="label">时间：</span>
                      <span class="text">{{ formatDate(task.human_review_result?.created_at) }}</span>
                    </div>
                  </div>
                </div>

                <div class="diff-panel ai-panel">
                  <div class="panel-header">
                    <span>AI 审核</span>
                    <el-tag :type="task.ai_review_result?.is_approved ? 'success' : 'danger'" size="small">
                      {{ task.ai_review_result?.is_approved ? '通过' : '不通过' }}
                    </el-tag>
                  </div>
                  <div class="panel-body">
                    <div class="panel-line">
                      <span class="label">置信度：</span>
                      <span class="text">{{ task.ai_review_result?.confidence ?? 0 }}%</span>
                    </div>
                    <div class="panel-line">
                      <span class="label">模型：</span>
                      <span class="text">{{ task.ai_review_result?.model || '默认模型' }}</span>
                    </div>
                    <div class="panel-line">
                      <span class="label">标签：</span>
                      <div class="tag-list">
                        <el-tag
                          v-for="tag in (task.ai_review_result?.tags || [])"
                          :key="tag"
                          type="info"
                          size="small"
                        >
                          {{ tag }}
                        </el-tag>
                        <span v-if="(task.ai_review_result?.tags || []).length === 0">无</span>
                      </div>
                    </div>
                    <div class="panel-line">
                      <span class="label">原因：</span>
                      <span class="text">{{ task.ai_review_result?.reason || '无' }}</span>
                    </div>
                    <div class="panel-line">
                      <span class="label">时间：</span>
                      <span class="text">{{ formatDate(task.ai_review_result?.created_at) }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <el-divider />

              <div class="final-section">
                <div class="final-header">
                  <h3 class="section-title">最终判定</h3>
                  <div class="quick-actions">
                    <el-button size="small" @click="applyHumanDecision(task)">使用人工结果</el-button>
                    <el-button size="small" @click="applyAIDecision(task)">使用 AI 结果</el-button>
                  </div>
                </div>

                <el-form label-position="top" size="default">
                  <el-form-item label="判定结果" required>
                    <el-radio-group v-model="getReview(task.id).is_approved">
                      <el-radio :value="true">通过</el-radio>
                      <el-radio :value="false">不通过</el-radio>
                    </el-radio-group>
                  </el-form-item>

                  <el-form-item
                    v-if="getReview(task.id) && !getReview(task.id).is_approved"
                    label="违规标签"
                    required
                  >
                    <el-checkbox-group v-model="getReview(task.id).tags">
                      <el-checkbox
                        v-for="tag in availableTags"
                        :key="tag.id"
                        :label="tag.name"
                      >
                        {{ tag.name }}
                      </el-checkbox>
                    </el-checkbox-group>
                  </el-form-item>

                  <el-form-item label="判定说明">
                    <el-input
                      v-model="getReview(task.id).reason"
                      type="textarea"
                      :rows="3"
                      placeholder="请填写判定说明"
                    />
                  </el-form-item>
                </el-form>
              </div>

              <div class="task-actions">
                <el-button type="primary" :loading="submitLoading[task.id]" @click="handleSubmitSingle(task.id)">
                  提交判定
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
import {
  claimAIHumanDiffTasks,
  getMyAIHumanDiffTasks,
  submitAIHumanDiffReview,
  submitBatchAIHumanDiffReviews,
  returnAIHumanDiffTasks
} from '../../api/aiHumanDiff'
import { getTags } from '../../api/task'
import type { AIHumanDiffTask, SubmitAIHumanDiffRequest, Tag } from '../../types'

const router = useRouter()
const userStore = useUserStore()

const claimLoading = ref(false)
const claimCount = ref(20)
const returnCount = ref(1)
const diffTasks = ref<AIHumanDiffTask[]>([])
const availableTags = ref<Tag[]>([])
const reviews = reactive<Record<number, SubmitAIHumanDiffRequest>>({})
const batchSelection = reactive<Record<number, boolean>>({})
const submitLoading = reactive<Record<number, boolean>>({})
const batchLoading = ref(false)
const activeTaskId = ref<number | null>(null)
const currentQueueName = ref<string>('')

const draftStorageKey = 'ai_human_diff_drafts'

const avgConfidence = computed(() => {
  if (diffTasks.value.length === 0) return 0
  const total = diffTasks.value.reduce((sum, task) => sum + (task.ai_review_result?.confidence || 0), 0)
  return Math.round(total / diffTasks.value.length)
})

const getReview = (taskId: number) => {
  if (!reviews[taskId]) {
    reviews[taskId] = {
      task_id: taskId,
      is_approved: null as any,
      tags: [],
      reason: ''
    }
  }
  return reviews[taskId]
}

const selectedReviews = computed(() => {
  return Object.entries(reviews)
    .filter(([taskId]) => batchSelection[parseInt(taskId)] && isReviewReady(parseInt(taskId)))
    .map(([taskId, review]) => ({
      ...review,
      task_id: parseInt(taskId)
    }))
})

const loadMyTasks = async () => {
  try {
    const res = await getMyAIHumanDiffTasks()
    diffTasks.value = res.tasks || []
    initReviews()
    restoreDrafts()
  } catch (error) {
    console.error('Failed to load diff tasks:', error)
    ElMessage.error('加载任务失败，请稍后重试')
  }
}

const initReviews = () => {
  const taskIds = new Set(diffTasks.value.map(t => t.id))
  Object.keys(reviews).forEach((key) => {
    if (!taskIds.has(parseInt(key))) {
      delete reviews[parseInt(key)]
    }
  })

  Object.keys(batchSelection).forEach((key) => {
    if (!taskIds.has(parseInt(key))) {
      delete batchSelection[parseInt(key)]
    }
  })

  diffTasks.value.forEach((task) => {
    if (!reviews[task.id]) {
      reviews[task.id] = {
        task_id: task.id,
        is_approved: null as any,
        tags: [],
        reason: ''
      }
    }

    if (batchSelection[task.id] === undefined) {
      batchSelection[task.id] = false
    }
  })

  returnCount.value = Math.min(1, diffTasks.value.length)

  if (diffTasks.value.length > 0 && !taskIds.has(activeTaskId.value || -1)) {
    activeTaskId.value = diffTasks.value[0].id
  }
}

onMounted(async () => {
  const queueStr = sessionStorage.getItem('currentQueue')
  if (queueStr) {
    try {
      const queue = JSON.parse(queueStr)
      currentQueueName.value = queue.description || queue.queue_name || ''
      sessionStorage.removeItem('currentQueue')
    } catch (e) {
      console.error('Failed to parse current queue:', e)
    }
  }

  try {
    const tagRes = await getTags()
    availableTags.value = tagRes.tags || []
    await loadMyTasks()
  } catch (error) {
    console.error('Failed to load data:', error)
  }

  window.addEventListener('keydown', handleKeyPress)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyPress)
})

const handleClaimTasks = async () => {
  if (claimCount.value < 1 || claimCount.value > 50) {
    ElMessage.warning('领取数量必须在 1-50 之间')
    return
  }

  claimLoading.value = true
  try {
    const res = await claimAIHumanDiffTasks(claimCount.value)
    ElMessage.success(`成功领取 ${res.count} 条差异任务`)
    await loadMyTasks()
  } catch (error) {
    console.error('Failed to claim AI diff tasks:', error)
    ElMessage.error('领取任务失败，请稍后重试')
  } finally {
    claimLoading.value = false
  }
}

const handleRefresh = async () => {
  try {
    await loadMyTasks()
    ElMessage.success('刷新成功')
  } catch (error) {
    console.error('Failed to refresh:', error)
    ElMessage.error('刷新失败，请稍后重试')
  }
}

const validateReview = (review: SubmitAIHumanDiffRequest): boolean => {
  if (review.is_approved === null) {
    ElMessage.warning('请选择判定结果')
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

const isReviewReady = (taskId: number) => {
  const review = reviews[taskId]
  if (!review) return false
  if (review.is_approved === null) return false
  if (!review.is_approved) {
    return review.tags.length > 0 && review.reason.trim().length > 0
  }
  return true
}

const clearReviewForm = (taskId: number) => {
  reviews[taskId] = {
    task_id: taskId,
    is_approved: null as any,
    tags: [],
    reason: ''
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
  const ids = diffTasks.value.map(task => task.id)
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

  if (diffTasks.value.length === 0) return
  const currentId = activeTaskId.value ?? diffTasks.value[0]?.id
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
    const taskIds = new Set(diffTasks.value.map(t => t.id))

    Object.entries(savedReviews).forEach(([taskId, review]) => {
      const id = Number(taskId)
      if (!taskIds.has(id)) return
      reviews[id] = { ...getReview(id), ...(review as SubmitAIHumanDiffRequest) }
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
  () => diffTasks.value.length,
  (count) => {
    document.title = `(${count}) 差异工作台`
  },
  { immediate: true }
)

const handleSubmitSingle = async (taskId: number) => {
  const review = reviews[taskId]
  if (!review || !validateReview(review)) return

  try {
    submitLoading[taskId] = true
    await submitAIHumanDiffReview(review)
    ElMessage.success('提交成功')
    diffTasks.value = diffTasks.value.filter(t => t.id !== taskId)
    delete reviews[taskId]
    batchSelection[taskId] = false
    await loadMyTasks()
  } catch (error) {
    console.error('Failed to submit AI diff review:', error)
    ElMessage.error('提交失败，请稍后重试')
  } finally {
    submitLoading[taskId] = false
  }
}

const handleBatchSubmit = async () => {
  const validReviews: SubmitAIHumanDiffRequest[] = []

  for (const review of selectedReviews.value) {
    if (!validateReview(review)) {
      return
    }
    validReviews.push(review)
  }

  if (validReviews.length === 0) {
    ElMessage.warning('没有可提交的判定')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确认提交 ${validReviews.length} 条判定结果？`,
      '批量提交',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    batchLoading.value = true
    const res = await submitBatchAIHumanDiffReviews(validReviews)
    ElMessage.success(`成功提交 ${res.count} 条判定`)

    validReviews.forEach((review) => {
      diffTasks.value = diffTasks.value.filter(t => t.id !== review.task_id)
      delete reviews[review.task_id]
      batchSelection[review.task_id] = false
    })
    await loadMyTasks()
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

  if (returnCount.value > diffTasks.value.length) {
    ElMessage.warning(`退单数量不能超过当前任务数 (${diffTasks.value.length})`)
    return
  }

  if (returnCount.value > 50) {
    ElMessage.warning('退单数量不能超过 50 条')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确认退回 ${returnCount.value} 条差异任务？将退回最早领取的任务。`,
      '退单确认',
      {
        confirmButtonText: '确认退单',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const taskIdsToReturn = diffTasks.value
      .slice(0, returnCount.value)
      .map(task => task.id)

    const res = await returnAIHumanDiffTasks(taskIdsToReturn)
    ElMessage.success(`成功退回 ${res.count} 条任务`)

    taskIdsToReturn.forEach((taskId) => {
      diffTasks.value = diffTasks.value.filter(t => t.id !== taskId)
      delete reviews[taskId]
    })

    await loadMyTasks()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to return tasks:', error)
      ElMessage.error('退单失败，请稍后重试')
    }
  }
}

const applyHumanDecision = (task: AIHumanDiffTask) => {
  const review = getReview(task.id)
  if (!task.human_review_result) return
  review.is_approved = task.human_review_result.is_approved
  review.tags = [...(task.human_review_result.tags || [])]
  review.reason = task.human_review_result.reason || ''
}

const applyAIDecision = (task: AIHumanDiffTask) => {
  const review = getReview(task.id)
  if (!task.ai_review_result) return
  review.is_approved = task.ai_review_result.is_approved
  review.tags = [...(task.ai_review_result.tags || [])]
  review.reason = task.ai_review_result.reason || ''
}

const formatDate = (dateStr?: string | null) => {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleString('zh-CN')
  } catch {
    return dateStr
  }
}

const goToSearch = () => {
  router.push('/reviewer/search')
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.ai-human-diff-dashboard {
  background: var(--color-bg-100);
  min-height: 100vh;
}

.header {
  background: var(--color-bg-000);
  border-bottom: 1px solid var(--color-border-lighter);
  padding: 0 24px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.header-left h2 {
  margin: 0;
  color: var(--color-text-000);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.main-content {
  padding: 24px;
}

.stats-bar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stat-label {
  color: var(--color-text-300);
  font-size: 14px;
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text-000);
}

.actions-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 20px;
  align-items: center;
}

.claim-section,
.return-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.empty-state {
  margin-top: 48px;
}

.tasks-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.diff-card {
  border: 1px solid var(--color-border-lighter);
  border-radius: 12px;
  position: relative;
}

.diff-card.is-filled {
  border-color: rgba(64, 158, 255, 0.5);
}

.diff-card.is-selected {
  border-color: var(--color-accent-main);
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.diff-card.is-active {
  border-color: var(--color-success);
  box-shadow: 0 0 0 2px rgba(103, 194, 58, 0.2);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.task-id {
  font-weight: 600;
  color: var(--color-text-100);
}

.task-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.comment-section {
  background: var(--color-bg-000);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
  border-left: 4px solid var(--color-accent-main);
}

.comment-text {
  color: var(--color-text-100);
  font-size: 15px;
  line-height: 1.6;
}

.section-title {
  margin: 0 0 10px;
  font-size: 16px;
  color: var(--color-text-000);
}

.diff-panels {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
}

.diff-panel {
  border-radius: 10px;
  padding: 16px;
  border: 1px solid var(--color-border-lighter);
  background: var(--color-bg-000);
}

.human-panel {
  border-left: 4px solid var(--color-success);
}

.ai-panel {
  border-left: 4px solid var(--color-info);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 600;
  color: var(--color-text-000);
}

.panel-line {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin-bottom: 8px;
  color: var(--color-text-200);
  font-size: 13px;
}

.panel-line .label {
  min-width: 52px;
  color: var(--color-text-300);
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.final-section {
  margin-top: 12px;
  padding: 12px 16px;
  background: var(--color-bg-200);
  border-radius: 10px;
  border: 1px solid var(--color-border-lighter);
}

.final-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.quick-actions {
  display: flex;
  gap: 8px;
}

.task-actions {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
