<template>
  <div class="search-tasks-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">审核记录搜索</span>
        </div>
      </template>

      <!-- 搜索表单 -->
      <el-collapse v-model="activeCollapse" class="search-collapse">
        <el-collapse-item title="搜索条件" name="filters">
          <el-form :model="searchForm" label-width="120px" class="search-form">
            <el-row :gutter="20">
              <el-col :span="6">
                <el-form-item label="队列">
                  <el-select
                    v-model="searchForm.queue_name"
                    placeholder="请选择队列"
                    clearable
                    style="width: 100%"
                  >
                    <el-option
                      v-for="option in queueOptions"
                      :key="option.value"
                      :label="option.label"
                      :value="option.value"
                    />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="内容ID">
                  <el-input
                    v-model.number="searchForm.comment_id"
                    placeholder="请输入内容ID"
                    clearable
                    type="number"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="审核员账号">
                  <el-input
                    v-model="searchForm.reviewer_rtx"
                    placeholder="请输入审核员账号"
                    clearable
                  />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="标签">
                  <el-select
                    v-model="searchForm.tag_ids"
                    multiple
                    placeholder="请选择标签"
                    clearable
                    style="width: 100%"
                  >
                    <el-option
                      v-for="tag in tags"
                      :key="tag.id"
                      :label="tag.name"
                      :value="tag.name"
                    />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>

            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="审核开始时间">
                  <el-date-picker
                    v-model="searchForm.review_start_time"
                    type="datetime"
                    placeholder="选择开始时间"
                    style="width: 100%"
                    value-format="YYYY-MM-DDTHH:mm:ssZ"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="审核结束时间">
                  <el-date-picker
                    v-model="searchForm.review_end_time"
                    type="datetime"
                    placeholder="选择结束时间"
                    style="width: 100%"
                    value-format="YYYY-MM-DDTHH:mm:ssZ"
                  />
                </el-form-item>
              </el-col>
            </el-row>

            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="快速日期">
                  <el-button-group class="quick-filters">
                    <el-button @click="setDateRange('today')">今天</el-button>
                    <el-button @click="setDateRange('yesterday')">昨天</el-button>
                    <el-button @click="setDateRange('last7days')">最近7天</el-button>
                    <el-button @click="setDateRange('last30days')">最近30天</el-button>
                  </el-button-group>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="常用筛选">
                  <div class="saved-filters">
                    <el-select v-model="selectedFilter" placeholder="选择常用筛选" clearable style="width: 100%">
                      <el-option
                        v-for="filter in savedFilters"
                        :key="filter.id"
                        :label="filter.name"
                        :value="filter.id"
                      />
                    </el-select>
                    <el-button @click="saveCurrentFilter">保存当前筛选</el-button>
                  </div>
                </el-form-item>
              </el-col>
            </el-row>

            <el-row>
              <el-col :span="24">
                <el-form-item>
                  <el-button type="primary" @click="handleSearch" :loading="loading">
                    <el-icon><Search /></el-icon>
                    搜索
                  </el-button>
                  <el-button @click="handleReset">
                    <el-icon><Refresh /></el-icon>
                    重置
                  </el-button>
                  <el-button @click="handleExport">
                    <el-icon><Download /></el-icon>
                    导出结果
                  </el-button>
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </el-collapse-item>
      </el-collapse>

      <!-- 统计信息 -->
      <el-alert
        v-if="searchResult"
        :title="`共找到 ${searchResult.total} 条记录，当前第 ${searchResult.page}/${searchResult.total_pages} 页`"
        type="info"
        :closable="false"
        style="margin-bottom: 20px"
      />

      <!-- 结果表格 -->
      <el-table
        :data="tableData"
        stripe
        border
        v-loading="loading"
        style="width: 100%"
        max-height="600"
      >
        <el-table-column prop="id" label="任务ID" width="80" fixed />
        <el-table-column label="队列" width="160" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ formatQueueLabel(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="content_id" label="内容ID" width="120" />
        <el-table-column prop="content_text" label="内容摘要" min-width="200" show-overflow-tooltip />
        <el-table-column label="审核员" width="120" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.reviewer_username || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="结果" width="140" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.decision" :type="getDecisionType(row.decision)" size="small">
              {{ formatDecision(row.decision) }}
            </el-tag>
            <el-tag v-else type="info" size="small">-</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="标签" width="180">
          <template #default="{ row }">
            <el-tag
              v-for="tag in (row.tags || [])"
              :key="tag"
              size="small"
              type="warning"
              style="margin-right: 5px"
            >
              {{ tag }}
            </el-tag>
            <span v-if="!row.tags || row.tags.length === 0">-</span>
          </template>
        </el-table-column>
        <el-table-column label="原因" width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.reason || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="领取时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.claimed_at) }}
          </template>
        </el-table-column>
        <el-table-column label="完成时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.completed_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="handleViewDetail(row)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-if="searchResult"
        :current-page="pagination.page"
        :page-size="pagination.page_size"
        :page-sizes="[10, 20, 50, 100]"
        :total="searchResult.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        style="margin-top: 20px; justify-content: center"
      />
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailVisible" title="审核记录详情" width="800px">
      <el-descriptions :column="2" border v-if="currentDetail">
        <el-descriptions-item label="任务ID">{{ currentDetail.id }}</el-descriptions-item>
        <el-descriptions-item label="队列">{{ formatQueueLabel(currentDetail) }}</el-descriptions-item>
        <el-descriptions-item label="内容类型">{{ formatContentType(currentDetail.content_type) }}</el-descriptions-item>
        <el-descriptions-item label="内容ID">{{ currentDetail.content_id }}</el-descriptions-item>
        <el-descriptions-item label="内容摘要" :span="2">
          <el-input
            v-model="currentDetail.content_text"
            type="textarea"
            :rows="3"
            readonly
          />
        </el-descriptions-item>
        <el-descriptions-item label="审核员">{{ currentDetail.reviewer_username || '-' }}</el-descriptions-item>
        <el-descriptions-item label="审核结果">
          <el-tag v-if="currentDetail.decision" :type="getDecisionType(currentDetail.decision)">
            {{ formatDecision(currentDetail.decision) }}
          </el-tag>
          <el-tag v-else type="info">-</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="标签" :span="2">
          <el-tag
            v-for="tag in (currentDetail.tags || [])"
            :key="tag"
            size="small"
            type="warning"
            style="margin-right: 5px"
          >
            {{ tag }}
          </el-tag>
          <span v-if="!currentDetail.tags || currentDetail.tags.length === 0">-</span>
        </el-descriptions-item>
        <el-descriptions-item label="原因" :span="2">
          {{ currentDetail.reason || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="currentDetail.pool" label="流量池">
          {{ formatPoolLabel(currentDetail.pool) }}
        </el-descriptions-item>
        <el-descriptions-item v-if="currentDetail.overall_score !== null && currentDetail.overall_score !== undefined" label="综合评分">
          {{ currentDetail.overall_score }}
        </el-descriptions-item>
        <el-descriptions-item v-if="currentDetail.traffic_pool_result" label="推荐流量池">
          {{ currentDetail.traffic_pool_result }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ formatDate(currentDetail.created_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="领取时间">
          {{ formatDate(currentDetail.claimed_at || '') }}
        </el-descriptions-item>
        <el-descriptions-item label="完成时间">
          {{ formatDate(currentDetail.completed_at || '') }}
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Download } from '@element-plus/icons-vue'
import { searchTasks } from '../api/task'
import { getTags } from '../api/task'
import type { SearchTasksRequest, SearchTasksResponse, TaskSearchResult, Tag } from '../types'
import { formatDate } from '../utils/format'

const route = useRoute()

const activeCollapse = ref(['filters'])
const savedFilters = ref<Array<{ id: string; name: string; form: Record<string, any> }>>([])
const selectedFilter = ref<string | null>(null)
const savedFiltersKey = 'search_task_saved_filters'

// 搜索表单
const searchForm = reactive<{
  comment_id?: number
  reviewer_rtx?: string
  tag_ids?: string[]
  review_start_time?: string
  review_end_time?: string
  queue_name?: string
}>({
  comment_id: undefined,
  reviewer_rtx: '',
  tag_ids: [],
  review_start_time: '',
  review_end_time: '',
  queue_name: 'all',
})

const queueOptions = [
  { label: '全部队列', value: 'all' },
  { label: '评论一审', value: 'comment_first_review' },
  { label: '评论二审', value: 'comment_second_review' },
  { label: 'AI与人工diff', value: 'ai_human_diff' },
  { label: '质量检查', value: 'quality_check' },
  { label: '视频审核（流量池）', value: 'video_queue' },
  { label: '视频审核（100k）', value: 'video_queue_100k' },
  { label: '视频审核（1m）', value: 'video_queue_1m' },
  { label: '视频审核（10m）', value: 'video_queue_10m' },
]

const queueLabelMap: Record<string, string> = {
  comment_first_review: '评论一审',
  comment_second_review: '评论二审',
  ai_human_diff: 'AI与人工diff',
  quality_check: '质量检查',
  video_queue: '视频审核（流量池）',
}

const decisionLabelMap: Record<string, string> = {
  approved: '通过',
  rejected: '拒绝',
  passed: '通过',
  failed: '不通过',
  push_next_pool: '推送下一流量池',
  natural_pool: '自然流量池',
  remove_violation: '违规下架',
}

const decisionTypeMap: Record<string, string> = {
  approved: 'success',
  rejected: 'danger',
  passed: 'success',
  failed: 'danger',
  push_next_pool: 'success',
  natural_pool: 'info',
  remove_violation: 'danger',
}

// 分页
const pagination = reactive({
  page: 1,
  page_size: 20,
})

// 数据
const loading = ref(false)
const tableData = ref<TaskSearchResult[]>([])
const searchResult = ref<SearchTasksResponse | null>(null)
const tags = ref<Tag[]>([])

// 详情
const detailVisible = ref(false)
const currentDetail = ref<TaskSearchResult | null>(null)

// 加载标签列表
const loadTags = async () => {
  try {
    const res = await getTags()
    tags.value = res.tags
  } catch (error: any) {
    console.error('加载标签失败:', error)
  }
}

const loadSavedFilters = () => {
  const raw = localStorage.getItem(savedFiltersKey)
  if (!raw) {
    savedFilters.value = []
    return
  }
  try {
    savedFilters.value = JSON.parse(raw) || []
  } catch (error) {
    console.error('Failed to parse saved filters:', error)
    savedFilters.value = []
  }
}

const saveCurrentFilter = async () => {
  try {
    const res = await ElMessageBox.prompt('请输入筛选名称', '保存筛选', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputPattern: /.+/,
      inputErrorMessage: '筛选名称不能为空',
    })

    const id = `${Date.now()}`
    const formSnapshot = { ...searchForm }
    const newFilters = [...savedFilters.value, { id, name: res.value, form: formSnapshot }]
    savedFilters.value = newFilters
    localStorage.setItem(savedFiltersKey, JSON.stringify(newFilters))
    selectedFilter.value = id
    ElMessage.success('已保存筛选条件')
  } catch (error) {
    // Cancel
  }
}

const applySavedFilter = (filterId: string) => {
  const filter = savedFilters.value.find(item => item.id === filterId)
  if (!filter) return
  Object.assign(searchForm, filter.form)
  pagination.page = 1
  handleSearch()
}

const setDateRange = (range: 'today' | 'yesterday' | 'last7days' | 'last30days') => {
  const now = new Date()
  const end = new Date(now)
  let start = new Date(now)

  switch (range) {
    case 'today':
      start.setHours(0, 0, 0, 0)
      break
    case 'yesterday':
      start.setDate(start.getDate() - 1)
      start.setHours(0, 0, 0, 0)
      end.setDate(end.getDate() - 1)
      end.setHours(23, 59, 59, 999)
      break
    case 'last7days':
      start.setDate(start.getDate() - 6)
      start.setHours(0, 0, 0, 0)
      break
    case 'last30days':
      start.setDate(start.getDate() - 29)
      start.setHours(0, 0, 0, 0)
      break
    default:
      break
  }

  const formatDateTime = (date: Date) => date.toISOString()
  searchForm.review_start_time = formatDateTime(start)
  searchForm.review_end_time = formatDateTime(end)
}

const formatPoolLabel = (pool?: string | null) => {
  if (!pool) return '-'
  const labels: Record<string, string> = {
    '100k': '100k流量池',
    '1m': '1m流量池',
    '10m': '10m流量池',
  }
  return labels[pool] || pool
}

const formatQueueLabel = (row: Pick<TaskSearchResult, 'queue_name' | 'pool'>) => {
  if (row.queue_name === 'video_queue' && row.pool) {
    return `视频审核（${formatPoolLabel(row.pool)}）`
  }
  return queueLabelMap[row.queue_name] || row.queue_name
}

const formatDecision = (decision?: string | null) => {
  if (!decision) return '-'
  return decisionLabelMap[decision] || decision
}

const getDecisionType = (decision?: string | null) => {
  if (!decision) return 'info'
  return decisionTypeMap[decision] || 'info'
}

const formatContentType = (contentType?: string | null) => {
  if (contentType === 'comment') return '评论'
  if (contentType === 'video') return '视频'
  return '-'
}

const applyQueueFilter = () => {
  const queryQueue = typeof route.query.queue_name === 'string' ? route.query.queue_name : ''
  const storedQueue = sessionStorage.getItem('queueSearchQueue')
  const queueName = queryQueue || storedQueue

  if (queueName) {
    searchForm.queue_name = queueName
    pagination.page = 1
    sessionStorage.removeItem('queueSearchQueue')
    handleSearch()
  }
}

watch(selectedFilter, (value) => {
  if (value) {
    applySavedFilter(value)
  }
})

// 搜索
const handleSearch = async () => {
  loading.value = true
  try {
    const params: SearchTasksRequest = {
      page: pagination.page,
      page_size: pagination.page_size,
    }

    if (searchForm.comment_id) {
      params.comment_id = searchForm.comment_id
    }
    if (searchForm.reviewer_rtx) {
      params.reviewer_rtx = searchForm.reviewer_rtx
    }
    if (searchForm.tag_ids && searchForm.tag_ids.length > 0) {
      params.tag_ids = searchForm.tag_ids.join(',')
    }
    if (searchForm.review_start_time) {
      params.review_start_time = searchForm.review_start_time
    }
    if (searchForm.review_end_time) {
      params.review_end_time = searchForm.review_end_time
    }
    if (searchForm.queue_name) {
      params.queue_name = searchForm.queue_name
    }

    const res = await searchTasks(params)
    searchResult.value = res
    tableData.value = res.data
    
    if (res.total === 0) {
      ElMessage.info('未找到符合条件的记录')
    } else {
      ElMessage.success(`找到 ${res.total} 条记录`)
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '搜索失败')
    tableData.value = []
    searchResult.value = null
  } finally {
    loading.value = false
  }
}

// 重置
const handleReset = () => {
  searchForm.comment_id = undefined
  searchForm.reviewer_rtx = ''
  searchForm.tag_ids = []
  searchForm.review_start_time = ''
  searchForm.review_end_time = ''
  searchForm.queue_name = 'all'
  selectedFilter.value = null
  pagination.page = 1
  pagination.page_size = 20
  tableData.value = []
  searchResult.value = null
}

// 分页改变
const handlePageChange = (page: number) => {
  pagination.page = page
  handleSearch()
}

const handleSizeChange = (size: number) => {
  pagination.page_size = size
  pagination.page = 1
  handleSearch()
}

// 查看详情
const handleViewDetail = (row: TaskSearchResult) => {
  currentDetail.value = row
  detailVisible.value = true
}

// 导出
const handleExport = () => {
  // 构建CSV内容
  const headers = [
    '任务ID',
    '队列',
    '内容类型',
    '内容ID',
    '内容摘要',
    '审核员',
    '结果',
    '标签',
    '原因',
    '流量池',
    '综合评分',
    '推荐流量池',
    '领取时间',
    '完成时间',
  ]
  
  const csvRows = tableData.value.length
    ? tableData.value.map(row => {
      const tagsValue = row.tags && row.tags.length > 0 ? row.tags.join(';') : '-'

      return [
        row.id,
        `"${formatQueueLabel(row).replace(/"/g, '""')}"`,
        formatContentType(row.content_type),
        row.content_id,
        `"${(row.content_text || '').replace(/"/g, '""')}"`,
        row.reviewer_username || '-',
        formatDecision(row.decision),
        tagsValue,
        `"${(row.reason || '-').replace(/"/g, '""')}"`,
        row.pool ? formatPoolLabel(row.pool) : '-',
        row.overall_score ?? '-',
        row.traffic_pool_result || '-',
        formatDate(row.claimed_at || ''),
        formatDate(row.completed_at || ''),
      ].join(',')
    })
    : []

  const csvContent = [
    headers.join(','),
    ...csvRows,
  ].join('\n')

  // 下载文件
  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)
  link.setAttribute('href', url)
  link.setAttribute('download', `审核记录_${new Date().getTime()}.csv`)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)

  ElMessage.success(tableData.value.length ? '导出成功' : '无结果，已导出空表头')
}

onMounted(() => {
  loadTags()
  loadSavedFilters()
  applyQueueFilter()
})
</script>

<style scoped>
/* ============================================
   搜索任务页面样式
   ============================================ */
.search-tasks-container {
  padding: var(--spacing-8);
  background-color: var(--color-bg-100);
  min-height: 100vh;
}

/* ============================================
   卡片头部
   ============================================ */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-4);
  flex-wrap: wrap;
}

.title {
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text-000);
  letter-spacing: var(--tracking-tight);
}

/* ============================================
   搜索表单
   ============================================ */
.search-form {
  margin-bottom: var(--spacing-6);
  padding: var(--spacing-6);
  background: var(--color-bg-000);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-lighter);
  box-shadow: var(--shadow-sm);
}

.search-collapse {
  margin-bottom: var(--spacing-4);
}

.quick-filters {
  display: flex;
  flex-wrap: wrap;
}

.saved-filters {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  width: 100%;
}

:deep(.el-form-item) {
  margin-bottom: var(--spacing-5);
}

:deep(.el-form-item__label) {
  font-weight: 600;
  color: var(--color-text-100);
  letter-spacing: var(--tracking-wide);
  margin-bottom: var(--spacing-2);
}

:deep(.el-date-editor) {
  width: 100%;
}

:deep(.el-select) {
  width: 100%;
}

/* ============================================
   表格增强
   ============================================ */
:deep(.el-table) {
  border-radius: var(--radius-lg);
  overflow: hidden;
}

:deep(.el-table__row) {
  transition: background-color var(--transition-fast);
}

:deep(.el-table__row:hover) {
  background-color: var(--color-bg-200);
}

:deep(.el-table__cell) {
  padding: var(--spacing-4) var(--spacing-3);
  line-height: var(--leading-relaxed);
  letter-spacing: var(--tracking-normal);
}

:deep(.el-table th.el-table__cell) {
  background-color: var(--color-bg-300);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
}

:deep(.el-table td.el-table__cell) {
  font-size: var(--text-sm);
}

/* ============================================
   分页样式
   ============================================ */
:deep(.el-pagination) {
  margin-top: var(--spacing-6);
  padding: var(--spacing-4) 0;
  justify-content: center;
}

/* ============================================
   响应式设计
   ============================================ */
@media (max-width: 768px) {
  .search-tasks-container {
    padding: var(--spacing-4);
  }

  .search-form {
    padding: var(--spacing-4);
  }

  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }

  :deep(.el-form-item) {
    margin-bottom: var(--spacing-3);
  }

  :deep(.el-table__cell) {
    padding: var(--spacing-2);
    font-size: var(--text-xs);
  }
}

@media (max-width: 1024px) {
  .search-tasks-container {
    padding: var(--spacing-6);
  }
}
</style>

