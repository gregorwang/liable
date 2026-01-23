<template>
  <div class="admin-ai-review-content">
    <el-card class="section-card">
      <template #header>
        <div class="card-header">
          <span>AI 审核批次</span>
        </div>
      </template>

      <el-form :model="createForm" label-width="120px" class="create-form">
        <el-form-item label="审核数量">
          <el-input-number v-model="createForm.max_count" :min="1" :max="100000" />
        </el-form-item>
        <el-form-item label="审核时间">
          <el-date-picker
            v-model="createForm.run_at"
            type="datetime"
            placeholder="可选，默认立即"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
          />
        </el-form-item>
        <el-form-item label="Prompt 版本">
          <el-input v-model="createForm.prompt_version" placeholder="可选" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="creating" :disabled="creating" @click="handleCreateJob">
            创建批次
          </el-button>
          <el-button :loading="loadingJobs" :disabled="loadingJobs" @click="loadJobs">刷新列表</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="section-card">
      <template #header>
        <div class="card-header">
          <span>批次列表</span>
          <div class="header-actions">
            <el-switch v-model="showArchived" active-text="显示归档" />
            <el-button :loading="loadingJobs" :disabled="loadingJobs" @click="loadJobs">刷新</el-button>
          </div>
        </div>
      </template>

      <el-table :data="jobs" style="width: 100%" v-loading="loadingJobs">
        <el-table-column type="expand" width="46">
          <template #default="{ row }">
            <div class="expand-content">
              <div class="expand-item">模型：{{ row.model || '-' }}</div>
              <div class="expand-item">Prompt 版本：{{ row.prompt_version || '-' }}</div>
              <div class="expand-item">创建人：{{ row.created_by || '-' }}</div>
              <div class="expand-item">开始时间：{{ row.started_at || '-' }}</div>
              <div class="expand-item">完成时间：{{ row.completed_at || '-' }}</div>
              <div class="expand-item">归档时间：{{ row.archived_at || '-' }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <div class="status-cell">
              <el-tag :type="getStatusType(row.status)">{{ getStatusLabel(row.status) }}</el-tag>
              <el-tag v-if="row.archived_at" size="small" type="info">已归档</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="max_count" label="数量" width="100" />
        <el-table-column label="进度" width="180">
          <template #default="{ row }">
            <div class="progress-cell">
              <span>{{ getHandledCount(row) }}/{{ row.total_tasks }}</span>
              <el-progress
                :percentage="getCompletionRate(row)"
                :stroke-width="6"
                :status="row.failed_tasks > 0 ? 'exception' : undefined"
              />
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="failed_tasks" label="失败" width="80" />
        <el-table-column prop="run_at" label="计划时间" width="180" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="240">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                v-if="canStartJob(row)"
                size="small"
                type="primary"
                :loading="startingJobs.has(row.id)"
                :disabled="startingJobs.has(row.id)"
                @click="handleStartJob(row.id)"
              >
                启动
              </el-button>
              <el-button size="small" @click="openJobDetail(row)">详情</el-button>
              <el-button
                size="small"
                type="danger"
                :loading="deletingTasks.has(row.id)"
                :disabled="deletingTasks.has(row.id) || row.total_tasks === 0"
                @click="handleDeleteJobTasks(row)"
              >
                清空任务
              </el-button>
              <el-button
                v-if="row.archived_at"
                size="small"
                :loading="archivingJobs.has(row.id)"
                :disabled="archivingJobs.has(row.id)"
                @click="handleArchive(row, false)"
              >
                取消归档
              </el-button>
              <el-button
                v-else
                size="small"
                :loading="archivingJobs.has(row.id)"
                :disabled="!canArchiveJob(row) || archivingJobs.has(row.id)"
                @click="handleArchive(row, true)"
              >
                归档
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="section-card" v-loading="loadingComparison">
      <template #header>
        <div class="card-header">
          <span>对比分析（一审结果）</span>
          <div class="header-actions">
            <el-select v-model="comparisonJobId" placeholder="全部批次" clearable>
              <el-option v-for="job in jobs" :key="job.id" :label="`批次 ${job.id}`" :value="job.id" />
            </el-select>
            <el-button
              :loading="loadingComparison"
              :disabled="loadingComparison"
              @click="loadComparison"
            >
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <div class="summary-grid" v-if="comparison">
        <div class="summary-item">
          <div class="summary-label">AI 结果总数</div>
          <div class="summary-value">{{ comparison.summary.total_ai_results }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">可对比数量</div>
          <div class="summary-value">{{ comparison.summary.comparable_count }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">待对比数量</div>
          <div class="summary-value">{{ comparison.summary.pending_compare_count }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">决策一致率</div>
          <div class="summary-value">{{ formatPercent(comparison.summary.decision_match_rate) }}</div>
        </div>
        <div class="summary-item">
          <div class="summary-label">标签重叠率</div>
          <div class="summary-value">{{ formatTagOverlap(comparison.summary) }}</div>
        </div>
      </div>

      <el-table :data="comparisonDiffs" style="width: 100%">
        <el-table-column prop="review_task_id" label="任务ID" width="100" />
        <el-table-column prop="comment_id" label="评论ID" width="120" />
        <el-table-column prop="comment_text" label="评论内容" />
        <el-table-column label="人工一审" width="120">
          <template #default="{ row }">
            <el-tag :type="row.human_is_approved ? 'success' : 'danger'">
              {{ row.human_is_approved ? '通过' : '拒绝' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="人工标签" min-width="180">
          <template #default="{ row }">
            <div v-if="row.human_tags?.length" class="tag-list">
              <el-tag v-for="tag in row.human_tags" :key="tag" size="small" type="info">
                {{ tag }}
              </el-tag>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="AI" width="120">
          <template #default="{ row }">
            <el-tag :type="row.ai_is_approved ? 'success' : 'danger'">
              {{ row.ai_is_approved ? '通过' : '拒绝' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="AI标签" min-width="180">
          <template #default="{ row }">
            <div v-if="row.ai_tags?.length" class="tag-list">
              <el-tag v-for="tag in row.ai_tags" :key="tag" size="small" type="info">
                {{ tag }}
              </el-tag>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="confidence" label="置信度" width="100" />
      </el-table>

      <el-empty v-if="comparison && comparisonDiffs.length === 0" description="暂无差异样本，可在任务详情查看全部 AI 结果" />
    </el-card>

    <el-drawer v-model="detailVisible" title="AI 审核详情" size="70%">
      <div v-if="detailLoading" class="drawer-loading">加载中...</div>
      <div v-else class="detail-content">
        <el-descriptions v-if="detailJob" :column="3" border>
          <el-descriptions-item label="批次ID">{{ detailJob.id }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ getStatusLabel(detailJob.status) }}</el-descriptions-item>
          <el-descriptions-item label="模型">{{ detailJob.model || '-' }}</el-descriptions-item>
          <el-descriptions-item label="Prompt 版本">{{ detailJob.prompt_version || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建人">{{ detailJob.created_by || '-' }}</el-descriptions-item>
          <el-descriptions-item label="总量">{{ detailJob.total_tasks }}</el-descriptions-item>
          <el-descriptions-item label="完成">{{ detailJob.completed_tasks }}</el-descriptions-item>
          <el-descriptions-item label="失败">{{ detailJob.failed_tasks }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ detailJob.started_at || '-' }}</el-descriptions-item>
          <el-descriptions-item label="完成时间">{{ detailJob.completed_at || '-' }}</el-descriptions-item>
          <el-descriptions-item label="归档时间">{{ detailJob.archived_at || '-' }}</el-descriptions-item>
        </el-descriptions>

        <el-table :data="detailTasks" style="width: 100%" v-loading="detailTasksLoading">
          <el-table-column prop="id" label="任务ID" width="90" />
          <el-table-column prop="review_task_id" label="审核任务ID" width="120" />
          <el-table-column prop="comment_id" label="评论ID" width="120" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getTaskStatusType(row.status)">{{ getTaskStatusLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="comment_text" label="评论内容" min-width="220" show-overflow-tooltip />
          <el-table-column label="AI结果" width="120">
            <template #default="{ row }">
              <el-tag v-if="row.result" :type="row.result.is_approved ? 'success' : 'danger'">
                {{ row.result.is_approved ? '通过' : '拒绝' }}
              </el-tag>
              <el-tag v-else type="info">无结果</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="AI标签" min-width="180">
            <template #default="{ row }">
              <div v-if="row.result?.tags?.length" class="tag-list">
                <el-tag v-for="tag in row.result.tags" :key="tag" size="small" type="info">
                  {{ tag }}
                </el-tag>
              </div>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="置信度" width="100">
            <template #default="{ row }">
              <span>{{ row.result?.confidence ?? '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="原因/错误" min-width="200" show-overflow-tooltip>
            <template #default="{ row }">
              <span v-if="row.result?.reason">{{ formatAIReason(row.result.reason) }}</span>
              <span v-else class="error-text">{{ formatAIError(row.error_message) }}</span>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-row">
          <el-pagination
            v-model:current-page="detailPage"
            v-model:page-size="detailPageSize"
            :total="detailTotal"
            background
            layout="total, prev, pager, next, sizes"
          />
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, reactive, watch, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createAIReviewJob,
  listAIReviewJobs,
  startAIReviewJob,
  getAIReviewComparison,
  listAIReviewTasks,
  archiveAIReviewJob,
  unarchiveAIReviewJob,
  getAIReviewJob,
  deleteAIReviewJobTasks,
} from '@/api/aiReview'
import type {
  AIReviewJob,
  AIReviewComparisonResponse,
  CreateAIReviewJobRequest,
  AIReviewTask,
} from '@/types'

const jobs = ref<AIReviewJob[]>([])
const creating = ref(false)
const loadingJobs = ref(false)
const loadingComparison = ref(false)
const startingJobs = reactive(new Set<number>())
const archivingJobs = reactive(new Set<number>())
const deletingTasks = reactive(new Set<number>())
const showArchived = ref(false)
const comparison = ref<AIReviewComparisonResponse | null>(null)
const comparisonJobId = ref<number | undefined>()
const pollIntervalMs = 5000
let poller: number | undefined
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailTasksLoading = ref(false)
const detailJob = ref<AIReviewJob | null>(null)
const detailTasks = ref<AIReviewTask[]>([])
const detailTotal = ref(0)
const detailPage = ref(1)
const detailPageSize = ref(20)

const createForm = ref<CreateAIReviewJobRequest>({
  max_count: 200,
  run_at: undefined,
  prompt_version: undefined,
})

const loadJobs = async (silent = false) => {
  if (!silent) {
    loadingJobs.value = true
  }
  try {
    const response = await listAIReviewJobs({
      page: 1,
      page_size: 50,
      include_archived: showArchived.value,
    })
    jobs.value = response.data
    if (comparisonJobId.value && !jobs.value.some((job) => job.id === comparisonJobId.value)) {
      comparisonJobId.value = undefined
    }
    syncPolling()
  } catch (error) {
    console.error(error)
    if (!silent) {
      ElMessage.error('加载批次失败')
    }
  } finally {
    if (!silent) {
      loadingJobs.value = false
    }
  }
}

const handleCreateJob = async () => {
  creating.value = true
  try {
    await createAIReviewJob({
      max_count: createForm.value.max_count,
      run_at: createForm.value.run_at,
      prompt_version: createForm.value.prompt_version,
    })
    ElMessage.success('批次已创建')
    await loadJobs()
  } catch (error) {
    console.error(error)
    ElMessage.error('创建批次失败')
  } finally {
    creating.value = false
  }
}

const handleStartJob = async (jobId: number) => {
  try {
    startingJobs.add(jobId)
    await startAIReviewJob(jobId)
    ElMessage.success('批次已启动')
    await Promise.all([loadJobs(true), loadComparison(true)])
  } catch (error) {
    console.error(error)
    ElMessage.error('启动失败')
  } finally {
    startingJobs.delete(jobId)
  }
}

const loadComparison = async (silent = false) => {
  if (!silent) {
    loadingComparison.value = true
  }
  try {
    const response = await getAIReviewComparison({
      job_id: comparisonJobId.value,
      limit: 20,
    })
    comparison.value = {
      ...response,
      diffs: response.diffs || [],
    }
  } catch (error) {
    console.error(error)
    if (!silent) {
      ElMessage.error('加载对比失败')
    }
  } finally {
    if (!silent) {
      loadingComparison.value = false
    }
  }
}

const getStatusLabel = (status: AIReviewJob['status']) => {
  const mapping: Record<AIReviewJob['status'], string> = {
    draft: '草稿',
    scheduled: '已排期',
    running: '进行中',
    completed: '已完成',
    failed: '失败',
    canceled: '已取消',
  }
  return mapping[status] || status
}

const getStatusType = (status: AIReviewJob['status']) => {
  const mapping: Record<AIReviewJob['status'], 'success' | 'warning' | 'danger' | 'info'> = {
    draft: 'info',
    scheduled: 'warning',
    running: 'warning',
    completed: 'success',
    failed: 'danger',
    canceled: 'info',
  }
  return mapping[status] || 'info'
}

const canStartJob = (job: AIReviewJob) => job.status === 'draft' && !job.archived_at

const canArchiveJob = (job: AIReviewJob) =>
  !job.archived_at && ['completed', 'failed', 'canceled'].includes(job.status)

const handleArchive = async (job: AIReviewJob, archived: boolean) => {
  try {
    archivingJobs.add(job.id)
    if (archived) {
      await archiveAIReviewJob(job.id)
      ElMessage.success('批次已归档')
    } else {
      await unarchiveAIReviewJob(job.id)
      ElMessage.success('批次已取消归档')
    }
    await loadJobs()
  } catch (error) {
    console.error(error)
    ElMessage.error('归档操作失败')
  } finally {
    archivingJobs.delete(job.id)
  }
}

const handleDeleteJobTasks = async (job: AIReviewJob) => {
  try {
    await ElMessageBox.confirm(
      `确认清空批次 ${job.id} 的任务与结果？此操作不可撤销。`,
      '提示',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )
  } catch {
    return
  }

  deletingTasks.add(job.id)
  try {
    await deleteAIReviewJobTasks(job.id)
    ElMessage.success('任务已清空')
    await Promise.all([loadJobs(), loadComparison(true)])
    if (detailJob.value?.id === job.id) {
      await loadJobDetail(job.id)
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('清空任务失败')
  } finally {
    deletingTasks.delete(job.id)
  }
}

const getCompletionRate = (job: AIReviewJob) => {
  if (!job.total_tasks) return 0
  const handled = job.completed_tasks + job.failed_tasks
  return Math.round((handled / job.total_tasks) * 100)
}

const getHandledCount = (job: AIReviewJob) => job.completed_tasks + job.failed_tasks

const formatPercent = (value: number) => `${value.toFixed(2)}%`

const formatTagOverlap = (summary: AIReviewComparisonResponse['summary']) => {
  if (!summary || summary.tag_comparable_count === 0) return '-'
  return formatPercent(summary.tag_overlap_rate)
}

const comparisonDiffs = computed(() => comparison.value?.diffs ?? [])

const getTaskStatusLabel = (status: AIReviewTask['status']) => {
  const mapping: Record<AIReviewTask['status'], string> = {
    pending: '待处理',
    in_progress: '处理中',
    completed: '已完成',
    failed: '失败',
  }
  return mapping[status] || status
}

const getTaskStatusType = (status: AIReviewTask['status']) => {
  const mapping: Record<AIReviewTask['status'], 'success' | 'warning' | 'danger' | 'info'> = {
    pending: 'info',
    in_progress: 'warning',
    completed: 'success',
    failed: 'danger',
  }
  return mapping[status] || 'info'
}

const formatAIReason = (reason?: string) => (reason?.trim() ? reason.trim() : '-')

const formatAIError = (message?: string) => {
  if (!message) return '-'
  const trimmed = message.trim()
  if (trimmed === 'comment text is empty') return '评论内容为空'
  if (trimmed.includes('context deadline exceeded')) return 'AI 请求超时'
  if (trimmed.startsWith('ai client not configured:')) {
    const missing = trimmed.replace('ai client not configured: missing', '').trim()
    const map: Record<string, string> = {
      base_url: '服务地址',
      api_key: 'API Key',
      model: '模型',
    }
    const parts = missing
      .split(',')
      .map((item) => map[item.trim()] || item.trim())
      .join('、')
    return `AI 服务未配置：缺少${parts}`
  }
  const statusMatch = trimmed.match(/ai request failed with status (\\d+)/)
  if (statusMatch) {
    const detail = trimmed.split(':').slice(1).join(':').trim()
    return detail
      ? `AI 请求失败，状态码 ${statusMatch[1]}：${detail}`
      : `AI 请求失败，状态码 ${statusMatch[1]}`
  }
  if (trimmed.startsWith('ai response decode failed')) return 'AI 响应解析失败'
  if (trimmed.startsWith('ai response missing choices')) return 'AI 响应缺少候选结果'
  if (trimmed.startsWith('ai response parse failed')) return 'AI 响应解析失败'
  if (trimmed.startsWith('ai response is not valid json')) return 'AI 响应不是有效 JSON'
  return trimmed
}

const loadJobDetail = async (jobID: number) => {
  detailTasksLoading.value = true
  try {
    const [jobResponse, tasksResponse] = await Promise.all([
      getAIReviewJob(jobID),
      listAIReviewTasks(jobID, {
        page: detailPage.value,
        page_size: detailPageSize.value,
      }),
    ])
    detailJob.value = jobResponse
    detailTasks.value = tasksResponse.data
    detailTotal.value = tasksResponse.total
  } catch (error) {
    console.error(error)
    ElMessage.error('加载任务详情失败')
  } finally {
    detailTasksLoading.value = false
  }
}

const openJobDetail = async (job: AIReviewJob) => {
  detailVisible.value = true
  detailLoading.value = true
  detailPage.value = 1
  detailPageSize.value = 20
  detailJob.value = job
  try {
    await loadJobDetail(job.id)
  } finally {
    detailLoading.value = false
  }
}

const startPolling = () => {
  if (poller) return
  poller = window.setInterval(async () => {
    await Promise.all([loadJobs(true), loadComparison(true)])
  }, pollIntervalMs)
}

const stopPolling = () => {
  if (!poller) return
  window.clearInterval(poller)
  poller = undefined
}

const syncPolling = () => {
  const hasActiveJob = jobs.value.some(
    (job) => !job.archived_at && (job.status === 'running' || job.status === 'scheduled'),
  )
  if (hasActiveJob) {
    startPolling()
  } else {
    stopPolling()
  }
}

watch(showArchived, async () => {
  await loadJobs()
})

watch(comparisonJobId, async () => {
  await loadComparison()
})

watch([detailPage, detailPageSize], async () => {
  if (!detailVisible.value || !detailJob.value) return
  await loadJobDetail(detailJob.value.id)
})

watch(detailVisible, (visible) => {
  if (visible) return
  detailJob.value = null
  detailTasks.value = []
  detailTotal.value = 0
})

onMounted(async () => {
  await loadJobs()
  await loadComparison()
  syncPolling()
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.admin-ai-review-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.section-card {
  border-radius: 12px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.create-form {
  max-width: 600px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.summary-item {
  background: var(--color-bg-100);
  padding: 12px;
  border-radius: 8px;
}

.summary-label {
  font-size: 12px;
  color: var(--color-text-400);
}

.summary-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-000);
}

.progress-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.expand-content {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 8px;
  padding: 8px 0;
  color: var(--color-text-300);
}

.expand-item {
  font-size: 12px;
}

.drawer-loading {
  text-align: center;
  padding: 16px;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pagination-row {
  display: flex;
  justify-content: flex-end;
}

.error-text {
  color: var(--color-danger);
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
</style>
