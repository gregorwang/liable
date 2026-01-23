<template>
  <div class="video-queue-review-dashboard">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <div class="header-left">
            <h2>视频审核工作台（流量池分级）</h2>
            <el-segmented v-model="currentPool" :options="poolOptions" size="large" style="margin-left: 20px">
              <template #default="{ item }">
                <div class="pool-option">
                  <span>{{ item.label }}</span>
                  <el-badge :value="item.badge" v-if="item.badge > 0" :max="99" />
                </div>
              </template>
            </el-segmented>
          </div>
        </div>
      </el-header>

      <el-main class="main-content">
        <!-- 统计栏 -->
        <div class="stats-inline">
          <div class="stat-chip">
            <span class="stat-label">待审核任务</span>
            <span class="stat-value">{{ tasks.length }}</span>
          </div>
          <div class="stat-chip">
            <span class="stat-label">今日已完成</span>
            <span class="stat-value">{{ todayCompleted }}</span>
          </div>
        </div>

        <div class="progress-bar">
          <el-progress
            :percentage="sessionProgress"
            :format="() => `${todayCompleted}/${sessionTotal}`"
          />
        </div>

        <!-- 操作栏 -->
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
            type="success"
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

        <!-- 空状态 -->
        <div v-if="tasks.length === 0" class="empty-state">
          <el-empty description="暂无待审核任务，点击「领取新任务」开始工作" />
        </div>

        <!-- 任务列表 -->
        <div v-else class="tasks-container">
          <el-card
            v-for="task in tasks"
            :key="task.id"
            class="task-card"
            :class="{
              'task-reviewed': reviewData[task.id],
              'is-selected': batchSelection[task.id],
              'is-active': activeTaskId === task.id
            }"
            :data-task-id="task.id"
            @click="setActiveTask(task.id)"
            @focusin="setActiveTask(task.id)"
          >
            <div class="task-header">
              <span class="task-id">任务 #{{ task.id }}</span>
              <span class="video-filename">{{ task.video?.filename || '未知文件' }}</span>
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

            <div class="task-content-wrapper">
              <!-- 视频播放器 -->
              <div class="video-container">
                <div v-if="!videoLoaded[task.id]" class="video-placeholder">
                  <el-button type="primary" :loading="videoLoading[task.id]" @click="loadVideo(task)">
                    <el-icon><VideoPlay /></el-icon>
                    加载视频
                  </el-button>
                  <p v-if="task.video?.file_size">
                    点击加载视频 ({{ formatFileSize(task.video.file_size) }})
                  </p>
                  <p v-else>点击加载视频</p>
                </div>
                <video
                  v-else
                  :src="task.video?.video_url"
                  controls
                  preload="none"
                  class="video-player"
                  @error="handleVideoError(task)"
                >
                  您的浏览器不支持视频播放
                </video>
              </div>

              <!-- 审核表单 -->
              <div class="review-form">
              <el-form :model="getReviewForm(task.id)" label-width="100px" size="default">
                <!-- 审核决定 -->
                <el-form-item label="审核决定" required>
                  <el-radio-group v-model="getReviewForm(task.id).review_decision" @change="onDecisionChange(task.id)">
                    <el-radio value="push_next_pool">
                      <span class="decision-option">
                        <el-icon><Promotion /></el-icon>
                        {{ getNextPoolText(currentPool) }}
                      </span>
                    </el-radio>
                    <el-radio value="natural_pool">
                      <span class="decision-option">
                        <el-icon><Clock /></el-icon>
                        自然流量池
                      </span>
                    </el-radio>
                    <el-radio value="remove_violation">
                      <span class="decision-option">
                        <el-icon><WarningFilled /></el-icon>
                        违规下架
                      </span>
                    </el-radio>
                  </el-radio-group>
                </el-form-item>

                <!-- 审核标签 -->
                <el-form-item label="审核标签" required>
                  <el-select
                    v-model="getReviewForm(task.id).tags"
                    multiple
                    filterable
                    placeholder="请选择标签（最多3个）"
                    style="width: 100%"
                    :max-collapse-tags="3"
                    @change="onTagsChange(task.id)"
                  >
                    <el-option-group
                      v-for="category in tagCategories"
                      :key="category"
                      :label="getCategoryName(category)"
                    >
                      <el-option
                        v-for="tag in getTagsByCategory(category)"
                        :key="tag.name"
                        :label="tag.name"
                        :value="tag.name"
                        :disabled="getReviewForm(task.id).tags.length >= 3 && !getReviewForm(task.id).tags.includes(tag.name)"
                      >
                        <span>{{ tag.name }}</span>
                        <span style="color: var(--el-text-color-secondary); font-size: 12px; margin-left: 8px">
                          {{ tag.description }}
                        </span>
                      </el-option>
                    </el-option-group>
                  </el-select>
                  <div class="tag-count-hint">
                    已选择 {{ getReviewForm(task.id).tags.length }} / 3 个标签
                  </div>
                </el-form-item>

                <!-- 审核理由 -->
                <el-form-item label="审核理由" required>
                  <el-input
                    v-model="getReviewForm(task.id).reason"
                    type="textarea"
                    :rows="3"
                    placeholder="请填写审核理由（必填）"
                    maxlength="500"
                    show-word-limit
                  />
                </el-form-item>

                <!-- 提交按钮 -->
                <el-form-item>
                  <el-button
                    type="primary"
                    :disabled="!isReviewFormValid(task.id)"
                    :loading="submitLoading[task.id]"
                    @click="handleSingleSubmit(task)"
                  >
                    提交审核
                  </el-button>
                  <el-button @click="clearReviewForm(task.id)">清空</el-button>
                </el-form-item>
              </el-form>
              </div>
            </div>
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { Pool, VideoQueueTask, VideoQueueTag, SubmitVideoQueueReviewRequest } from '@/api/videoQueue'
import {
  claimVideoQueueTasks,
  getMyVideoQueueTasks,
  submitVideoQueueReview,
  submitBatchVideoQueueReviews,
  returnVideoQueueTasks,
  getVideoQueueTags,
  generateVideoURL
} from '@/api/videoQueue'
import { Promotion, Clock, WarningFilled, VideoPlay } from '@element-plus/icons-vue'

// 当前队列
const currentPool = ref<Pool>('100k')

// 队列选项
const poolOptions = computed(() => [
  { label: '100k流量池', value: '100k' as Pool, badge: currentPool.value === '100k' ? tasks.value.length : 0 },
  { label: '1m流量池', value: '1m' as Pool, badge: currentPool.value === '1m' ? tasks.value.length : 0 },
  { label: '10m流量池', value: '10m' as Pool, badge: currentPool.value === '10m' ? tasks.value.length : 0 }
])

// 任务列表
const tasks = ref<VideoQueueTask[]>([])
const claimCount = ref(10)
const returnCount = ref(1)
const claimLoading = ref(false)
const todayCompleted = ref(0)
const sessionTotal = computed(() => todayCompleted.value + tasks.value.length)
const sessionProgress = computed(() => {
  return sessionTotal.value ? Math.round((todayCompleted.value / sessionTotal.value) * 100) : 0
})
const batchLoading = ref(false)
const submitLoading = reactive<Record<number, boolean>>({})
const batchSelection = reactive<Record<number, boolean>>({})
const activeTaskId = ref<number | null>(null)
const videoLoaded = reactive<Record<number, boolean>>({})
const videoLoading = reactive<Record<number, boolean>>({})

const draftStorageKey = 'video_queue_review_drafts'
const statsStorageKey = 'video_queue_review_stats'

const loadTodayStats = () => {
  const raw = localStorage.getItem(statsStorageKey)
  if (!raw) {
    todayCompleted.value = 0
    return
  }

  try {
    const parsed = JSON.parse(raw)
    const today = new Date().toISOString().slice(0, 10)
    if (parsed?.date !== today) {
      todayCompleted.value = 0
      localStorage.setItem(statsStorageKey, JSON.stringify({ date: today, completed: 0 }))
      return
    }
    todayCompleted.value = parsed.completed || 0
  } catch (error) {
    console.error('Failed to parse stats:', error)
    todayCompleted.value = 0
  }
}

const incrementTodayCompleted = (count: number) => {
  const today = new Date().toISOString().slice(0, 10)
  todayCompleted.value += count
  localStorage.setItem(statsStorageKey, JSON.stringify({ date: today, completed: todayCompleted.value }))
}

// 标签
const tags = ref<VideoQueueTag[]>([])
const tagCategories = ['content', 'technical', 'compliance', 'engagement']

// 审核数据
const reviewData = reactive<Record<number, SubmitVideoQueueReviewRequest>>({})

const createEmptyReviewForm = (taskId: number): SubmitVideoQueueReviewRequest => ({
  task_id: taskId,
  review_decision: '' as SubmitVideoQueueReviewRequest['review_decision'],
  reason: '',
  tags: []
})

const syncReviewDataWithTasks = (taskList: VideoQueueTask[]) => {
  const validIds = new Set(taskList.map((task) => task.id))

  Object.keys(reviewData).forEach((key) => {
    const id = Number(key)
    if (!validIds.has(id)) {
      delete reviewData[id]
    }
  })

  taskList.forEach((task) => {
    if (!reviewData[task.id]) {
      reviewData[task.id] = createEmptyReviewForm(task.id)
    }

    if (batchSelection[task.id] === undefined) {
      batchSelection[task.id] = false
    }

    if (videoLoaded[task.id] === undefined) {
      videoLoaded[task.id] = false
    }

    if (videoLoading[task.id] === undefined) {
      videoLoading[task.id] = false
    }
  })

  Object.keys(batchSelection).forEach((key) => {
    const id = Number(key)
    if (!validIds.has(id)) delete batchSelection[id]
  })

  Object.keys(videoLoaded).forEach((key) => {
    const id = Number(key)
    if (!validIds.has(id)) delete videoLoaded[id]
  })

  Object.keys(videoLoading).forEach((key) => {
    const id = Number(key)
    if (!validIds.has(id)) delete videoLoading[id]
  })

  if (taskList.length > 0 && !validIds.has(activeTaskId.value || -1)) {
    activeTaskId.value = taskList[0].id
  }
}

// 已选择的审核
const selectedReviews = computed(() => {
  return Object.entries(reviewData)
    .filter(([taskId]) => batchSelection[parseInt(taskId)] && isReviewReady(parseInt(taskId)))
    .map(([taskId, review]) => ({
      task_id: parseInt(taskId),
      ...review
    }))
})

// 监听队列切换
watch(currentPool, () => {
  loadTasks()
  loadTags()
})

onMounted(() => {
  loadTasks()
  loadTags()
  loadTodayStats()
  restoreDrafts()
  window.addEventListener('keydown', handleKeyPress)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyPress)
})

// 加载任务
const loadTasks = async () => {
  try {
    const res = await getMyVideoQueueTasks(currentPool.value)
    const newTasks = Array.isArray(res.tasks) ? res.tasks : []
    tasks.value = newTasks
    syncReviewDataWithTasks(newTasks)
    restoreDrafts()
  } catch (error: any) {
    tasks.value = []
    syncReviewDataWithTasks([])
    ElMessage.error(error.response?.data?.error || '加载任务失败')
  }
}

// 加载标签
const loadTags = async () => {
  try {
    const res = await getVideoQueueTags(currentPool.value)
    tags.value = Array.isArray(res.tags) ? res.tags : []
  } catch (error: any) {
    tags.value = []
    ElMessage.error(error.response?.data?.error || '加载标签失败')
  }
}

// 领取任务
const handleClaimTasks = async () => {
  if (tasks.value.length > 0) {
    ElMessageBox.confirm(
      '您还有未完成的任务，确定要领取新任务吗？',
      '提示',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    ).then(async () => {
      await claimTasks()
    }).catch(() => {})
  } else {
    await claimTasks()
  }
}

const claimTasks = async () => {
  claimLoading.value = true
  try {
    const res = await claimVideoQueueTasks(currentPool.value, { count: claimCount.value })
    ElMessage.success(`成功领取 ${res.count} 个任务`)
    await loadTasks()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '领取任务失败')
  } finally {
    claimLoading.value = false
  }
}

// 归还任务
const handleReturnTasks = async () => {
  ElMessageBox.confirm(
    `确定要退回前 ${returnCount.value} 个任务吗？`,
    '确认退单',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
  ).then(async () => {
    try {
      const taskIds = tasks.value.slice(0, returnCount.value).map(t => t.id)
      await returnVideoQueueTasks(currentPool.value, { task_ids: taskIds })
      ElMessage.success('退单成功')
      await loadTasks()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '退单失败')
    }
  }).catch(() => {})
}

// 提交单个审核
const handleSingleSubmit = async (task: VideoQueueTask) => {
  const review = reviewData[task.id]
  if (!isReviewFormValid(task.id)) {
    ElMessage.warning('请完整填写审核信息')
    return
  }

  try {
    submitLoading[task.id] = true
    await submitVideoQueueReview(currentPool.value, review)
    ElMessage.success('提交成功')
    delete reviewData[task.id]
    batchSelection[task.id] = false
    incrementTodayCompleted(1)
    await loadTasks()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '提交失败')
  } finally {
    submitLoading[task.id] = false
  }
}

// 批量提交
const handleBatchSubmit = async () => {
  if (selectedReviews.value.length === 0) {
    ElMessage.warning('请至少填写一条审核')
    return
  }

  ElMessageBox.confirm(
    `确定要批量提交 ${selectedReviews.value.length} 条审核吗？`,
    '确认提交',
    { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
  ).then(async () => {
    try {
      batchLoading.value = true
      await submitBatchVideoQueueReviews(currentPool.value, { reviews: selectedReviews.value })
      ElMessage.success(`成功提交 ${selectedReviews.value.length} 条审核`)

      // 清除已提交的审核数据
      selectedReviews.value.forEach(r => {
        delete reviewData[r.task_id]
        batchSelection[r.task_id] = false
      })

      incrementTodayCompleted(selectedReviews.value.length)
      await loadTasks()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '批量提交失败')
    } finally {
      batchLoading.value = false
    }
  }).catch(() => {})
}

// 刷新
const handleRefresh = () => {
  loadTasks()
}

// 加载视频URL
const loadVideo = async (task: VideoQueueTask) => {
  if (!task.video) return

  try {
    videoLoading[task.id] = true
    const res = await generateVideoURL({ video_id: task.video.id })
    if (task.video) {
      task.video.video_url = res.video_url
      task.video.url_expires_at = res.expires_at
    }
    videoLoaded[task.id] = true
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '加载视频失败')
  } finally {
    videoLoading[task.id] = false
  }
}

// 视频加载错误
const handleVideoError = (task: VideoQueueTask) => {
  ElMessage.error('视频加载失败，请刷新重试')
  videoLoaded[task.id] = false
}

// 辅助函数

const getNextPoolText = (pool: Pool) => {
  const texts = {
    '100k': '推送到 1m流量池',
    '1m': '推送到 10m流量池',
    '10m': '确认推送 1000万流量池'
  }
  return texts[pool]
}

const getCategoryName = (category: string) => {
  const names: Record<string, string> = {
    content: '内容质量',
    technical: '技术质量',
    compliance: '合规性',
    engagement: '传播潜力'
  }
  return names[category] || category
}

const getTagsByCategory = (category: string) => {
  return tags.value.filter(tag => tag.category === category)
}

const getReviewForm = (taskId: number) => {
  if (!reviewData[taskId]) {
    reviewData[taskId] = createEmptyReviewForm(taskId)
  }
  return reviewData[taskId]
}

const isReviewFormValid = (taskId: number) => {
  const form = reviewData[taskId]
  if (!form) return false
  return form.review_decision && form.reason.trim() && form.tags.length > 0 && form.tags.length <= 3
}

const isReviewReady = (taskId: number) => {
  return isReviewFormValid(taskId)
}

const clearReviewForm = (taskId: number) => {
  reviewData[taskId] = createEmptyReviewForm(taskId)
  batchSelection[taskId] = false
}

const onDecisionChange = (taskId: number) => {
  // 可以根据决定自动推荐标签
}

const onTagsChange = (taskId: number) => {
  const form = reviewData[taskId]
  if (form.tags.length > 3) {
    form.tags = form.tags.slice(0, 3)
    ElMessage.warning('最多只能选择3个标签')
  }
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
  const ids = tasks.value.map(task => task.id)
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

  if (tasks.value.length === 0) return
  const currentId = activeTaskId.value ?? tasks.value[0]?.id
  if (!currentId) return

  switch (event.key) {
    case '1':
      getReviewForm(currentId).review_decision = 'push_next_pool'
      break
    case '2':
      getReviewForm(currentId).review_decision = 'natural_pool'
      break
    case '3':
      getReviewForm(currentId).review_decision = 'remove_violation'
      break
    case 'Enter':
      event.preventDefault()
      const task = tasks.value.find(t => t.id === currentId)
      if (task) handleSingleSubmit(task)
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
    const taskIds = new Set(tasks.value.map(t => t.id))

    Object.entries(savedReviews).forEach(([taskId, review]) => {
      const id = Number(taskId)
      if (!taskIds.has(id)) return
      reviewData[id] = { ...getReviewForm(id), ...(review as SubmitVideoQueueReviewRequest) }
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
  reviewData,
  () => {
    if (draftTimer) window.clearTimeout(draftTimer)
    draftTimer = window.setTimeout(() => {
      const payload = { reviews: reviewData, batchSelection }
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
      const payload = { reviews: reviewData, batchSelection }
      localStorage.setItem(draftStorageKey, JSON.stringify(payload))
    }, 500)
  },
  { deep: true }
)

watch(
  () => tasks.value.length,
  (count) => {
    document.title = `(${count}) 视频审核工作台`
  },
  { immediate: true }
)

const formatFileSize = (size: number) => {
  if (!size && size !== 0) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = size
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }
  return `${value.toFixed(value >= 100 ? 0 : 1)} ${units[unitIndex]}`
}
</script>

<style scoped lang="scss">
.video-queue-review-dashboard {
  height: 100vh;
  background: #f5f7fa;

  .header {
    background: white;
    border-bottom: 1px solid #e4e7ed;
    padding: 0 24px;
    height: 60px !important;

    .header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
      height: 100%;

      .header-left {
        display: flex;
        align-items: center;

        h2 {
          margin: 0;
          font-size: 18px;
          font-weight: 600;
        }

        .pool-option {
          display: flex;
          align-items: center;
          gap: 8px;
        }
      }
    }
  }

  .main-content {
    padding: 24px;
    overflow-y: auto;

    .stats-inline {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 16px;
      flex-wrap: wrap;

      .stat-chip {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 8px 14px;
        background: #fff;
        border: 1px solid #e4e7ed;
        border-radius: 10px;
        box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);

        .stat-label {
          font-size: 13px;
          color: #909399;
        }

        .stat-value {
          font-size: 18px;
          font-weight: 600;
          color: #303133;
        }
      }
    }

    .progress-bar {
      margin-bottom: 20px;
      padding: 0 6px;
    }

    .actions-bar {
      display: flex;
      gap: 16px;
      align-items: center;
      margin-bottom: 24px;
      flex-wrap: wrap;

      .claim-section,
      .return-section {
        display: flex;
        gap: 8px;
        align-items: center;
      }
    }

    .empty-state {
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 400px;
    }

    .tasks-container {
      display: flex;
      flex-direction: column;
      gap: 24px;

      .task-card {
        &.task-reviewed {
          border: 2px solid #67c23a;
        }

        &.is-selected {
          border: 2px solid #409eff;
          box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.15);
        }

        &.is-active {
          border: 2px solid #67c23a;
          box-shadow: 0 0 0 2px rgba(103, 194, 58, 0.15);
        }

        .task-header {
          display: flex;
          align-items: center;
          gap: 12px;
          margin-bottom: 16px;
          padding-bottom: 12px;
          border-bottom: 1px solid #e4e7ed;

          .task-id {
            font-weight: 600;
            color: #303133;
          }

          .video-filename {
            flex: 1;
            color: #606266;
            font-size: 14px;
          }

          .task-header-actions {
            display: flex;
            align-items: center;
            gap: 12px;
          }
        }

        .task-content-wrapper {
          display: flex;
          flex-direction: row;
          gap: 24px;
        }

        .video-container {
          flex: 0 0 45%;
          margin-bottom: 0;

          .video-player {
            width: 100%;
            max-height: 600px;
            border-radius: 4px;
            background: #000;
          }

          .video-placeholder {
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            height: 200px;
            background: #f5f7fa;
            border-radius: 4px;
            gap: 8px;
            color: #909399;
          }
        }

        .review-form {
          flex: 1;

          .decision-option {
            display: flex;
            align-items: center;
            gap: 6px;
          }

          .tag-count-hint {
            margin-top: 4px;
            font-size: 12px;
            color: #909399;
          }
        }

        @media (max-width: 1200px) {
          .task-content-wrapper {
            flex-direction: column;
          }

          .video-container {
            flex: 1;
            margin-bottom: 20px;
          }
        }
      }
    }
  }
}
</style>
