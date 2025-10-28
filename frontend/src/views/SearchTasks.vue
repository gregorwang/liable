<template>
  <div class="search-tasks-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">审核记录搜索</span>
        </div>
      </template>

      <!-- 搜索表单 -->
      <el-form :model="searchForm" label-width="120px" class="search-form">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-form-item label="队列类型">
              <el-select
                v-model="searchForm.queue_type"
                placeholder="请选择队列类型"
                clearable
                style="width: 100%"
              >
                <el-option label="全部" value="all" />
                <el-option label="一审" value="first" />
                <el-option label="二审" value="second" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="评论ID">
              <el-input
                v-model.number="searchForm.comment_id"
                placeholder="请输入评论ID"
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
            <el-form-item label="违规标签">
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
              <el-button @click="handleExport" :disabled="!tableData.length">
                <el-icon><Download /></el-icon>
                导出结果
              </el-button>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

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
        <el-table-column prop="comment_id" label="评论ID" width="100" />
        <el-table-column prop="comment_text" label="评论内容" min-width="200" show-overflow-tooltip />
        <el-table-column label="队列类型" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.queue_type === 'first'" type="primary" size="small">一审</el-tag>
            <el-tag v-else-if="row.queue_type === 'second'" type="success" size="small">二审</el-tag>
            <el-tag v-else type="info" size="small">-</el-tag>
          </template>
        </el-table-column>
        
        <!-- 一审任务显示列 -->
        <template v-if="!searchForm.queue_type || searchForm.queue_type === 'all' || searchForm.queue_type === 'first'">
          <el-table-column prop="username" label="一审审核员" width="120" show-overflow-tooltip />
          <el-table-column label="一审结果" width="100" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.queue_type === 'first' && row.is_approved === true" type="success">通过</el-tag>
              <el-tag v-else-if="row.queue_type === 'first' && row.is_approved === false" type="danger">拒绝</el-tag>
              <el-tag v-else-if="row.queue_type === 'second' && row.first_username" type="info">已拒绝</el-tag>
              <el-tag v-else type="info">-</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="一审标签" width="150">
            <template #default="{ row }">
              <el-tag
                v-for="tag in (row.queue_type === 'first' ? row.tags : [])"
                :key="tag"
                size="small"
                type="warning"
                style="margin-right: 5px"
              >
                {{ tag }}
              </el-tag>
              <span v-if="row.queue_type === 'first' && (!row.tags || row.tags.length === 0)">-</span>
              <span v-else-if="row.queue_type === 'second'">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="reason" label="一审原因" width="120" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.queue_type === 'first' ? (row.reason || '-') : '-' }}
            </template>
          </el-table-column>
        </template>
        
        <!-- 二审任务显示列 -->
        <template v-if="!searchForm.queue_type || searchForm.queue_type === 'all' || searchForm.queue_type === 'second'">
          <el-table-column v-if="searchForm.queue_type === 'second'" prop="second_username" label="二审审核员" width="120" show-overflow-tooltip />
          <el-table-column v-if="searchForm.queue_type === 'second'" label="二审结果" width="100" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.second_is_approved === true" type="success">通过</el-tag>
              <el-tag v-else-if="row.second_is_approved === false" type="danger">拒绝</el-tag>
              <el-tag v-else type="info">-</el-tag>
            </template>
          </el-table-column>
          <el-table-column v-if="searchForm.queue_type === 'second'" label="二审标签" width="150">
            <template #default="{ row }">
              <el-tag
                v-for="tag in (row.second_tags || [])"
                :key="tag"
                size="small"
                type="warning"
                style="margin-right: 5px"
              >
                {{ tag }}
              </el-tag>
              <span v-if="!row.second_tags || row.second_tags.length === 0">-</span>
            </template>
          </el-table-column>
          <el-table-column v-if="searchForm.queue_type === 'second'" prop="second_reason" label="二审原因" width="120" show-overflow-tooltip />
        </template>
        
        <!-- 混合显示模式（当选择"全部"时） -->
        <template v-if="!searchForm.queue_type || searchForm.queue_type === 'all'">
          <el-table-column label="审核员" width="120" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.queue_type === 'first' ? row.username : (row.second_username || '-') }}
            </template>
          </el-table-column>
          <el-table-column label="审核结果" width="100" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.queue_type === 'first' && row.is_approved === true" type="success">通过</el-tag>
              <el-tag v-else-if="row.queue_type === 'first' && row.is_approved === false" type="danger">拒绝</el-tag>
              <el-tag v-else-if="row.queue_type === 'second' && row.second_is_approved === true" type="success">通过</el-tag>
              <el-tag v-else-if="row.queue_type === 'second' && row.second_is_approved === false" type="danger">拒绝</el-tag>
              <el-tag v-else type="info">-</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="违规标签" width="150">
            <template #default="{ row }">
              <el-tag
                v-for="tag in (row.queue_type === 'first' ? row.tags : (row.second_tags || []))"
                :key="tag"
                size="small"
                type="warning"
                style="margin-right: 5px"
              >
                {{ tag }}
              </el-tag>
              <span v-if="row.queue_type === 'first' && (!row.tags || row.tags.length === 0)">-</span>
              <span v-else-if="row.queue_type === 'second' && (!row.second_tags || row.second_tags.length === 0)">-</span>
            </template>
          </el-table-column>
          <el-table-column label="审核原因" width="120" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.queue_type === 'first' ? (row.reason || '-') : (row.second_reason || '-') }}
            </template>
          </el-table-column>
        </template>
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
        <el-descriptions-item label="评论ID">{{ currentDetail.comment_id }}</el-descriptions-item>
        <el-descriptions-item label="队列类型">
          <el-tag v-if="currentDetail.queue_type === 'first'" type="primary">一审</el-tag>
          <el-tag v-else-if="currentDetail.queue_type === 'second'" type="success">二审</el-tag>
          <el-tag v-else type="info">-</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="评论内容" :span="2">
          <el-input
            v-model="currentDetail.comment_text"
            type="textarea"
            :rows="3"
            readonly
          />
        </el-descriptions-item>
        
        <!-- 一审信息 -->
        <template v-if="currentDetail.queue_type === 'first'">
          <el-descriptions-item label="一审审核员">{{ currentDetail.username }}</el-descriptions-item>
          <el-descriptions-item label="一审审核员ID">{{ currentDetail.reviewer_id }}</el-descriptions-item>
          <el-descriptions-item label="一审审核结果">
            <el-tag v-if="currentDetail.is_approved === true" type="success">通过</el-tag>
            <el-tag v-else-if="currentDetail.is_approved === false" type="danger">拒绝</el-tag>
            <el-tag v-else type="info">-</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="一审违规标签">
            <el-tag
              v-for="tag in currentDetail.tags"
              :key="tag"
              size="small"
              type="warning"
              style="margin-right: 5px"
            >
              {{ tag }}
            </el-tag>
            <span v-if="!currentDetail.tags || currentDetail.tags.length === 0">-</span>
          </el-descriptions-item>
          <el-descriptions-item label="一审审核原因" :span="2">
            {{ currentDetail.reason || '-' }}
          </el-descriptions-item>
        </template>
        
        <!-- 二审信息 -->
        <template v-if="currentDetail.queue_type === 'second'">
          <el-descriptions-item label="一审审核员">{{ currentDetail.first_username || '-' }}</el-descriptions-item>
          <el-descriptions-item label="一审审核员ID">{{ currentDetail.first_reviewer_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="一审审核结果">
            <el-tag type="danger">拒绝</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="一审违规标签">
            <el-tag
              v-for="tag in currentDetail.tags"
              :key="tag"
              size="small"
              type="warning"
              style="margin-right: 5px"
            >
              {{ tag }}
            </el-tag>
            <span v-if="!currentDetail.tags || currentDetail.tags.length === 0">-</span>
          </el-descriptions-item>
          <el-descriptions-item label="一审审核原因" :span="2">
            {{ currentDetail.reason || '-' }}
          </el-descriptions-item>
          
          <el-descriptions-item label="二审审核员">{{ currentDetail.second_username || '-' }}</el-descriptions-item>
          <el-descriptions-item label="二审审核员ID">{{ currentDetail.second_reviewer_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="二审审核结果">
            <el-tag v-if="currentDetail.second_is_approved === true" type="success">通过</el-tag>
            <el-tag v-else-if="currentDetail.second_is_approved === false" type="danger">拒绝</el-tag>
            <el-tag v-else type="info">-</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="二审违规标签">
            <el-tag
              v-for="tag in (currentDetail.second_tags || [])"
              :key="tag"
              size="small"
              type="warning"
              style="margin-right: 5px"
            >
              {{ tag }}
            </el-tag>
            <span v-if="!currentDetail.second_tags || currentDetail.second_tags.length === 0">-</span>
          </el-descriptions-item>
          <el-descriptions-item label="二审审核原因" :span="2">
            {{ currentDetail.second_reason || '-' }}
          </el-descriptions-item>
        </template>
        
        <el-descriptions-item label="创建时间">
          {{ formatDate(currentDetail.created_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="领取时间">
          {{ formatDate(currentDetail.claimed_at || '') }}
        </el-descriptions-item>
        <el-descriptions-item label="完成时间">
          {{ formatDate(currentDetail.completed_at || '') }}
        </el-descriptions-item>
        <el-descriptions-item label="审核时间">
          {{ formatDate(currentDetail.queue_type === 'first' ? currentDetail.reviewed_at : currentDetail.second_reviewed_at || '') }}
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh, Download } from '@element-plus/icons-vue'
import { searchTasks } from '../api/task'
import { getTags } from '../api/task'
import type { SearchTasksRequest, SearchTasksResponse, TaskSearchResult, Tag } from '../types'
import { formatDate } from '../utils/format'

// 搜索表单
const searchForm = reactive<{
  comment_id?: number
  reviewer_rtx?: string
  tag_ids?: string[]
  review_start_time?: string
  review_end_time?: string
  queue_type?: string
}>({
  comment_id: undefined,
  reviewer_rtx: '',
  tag_ids: [],
  review_start_time: '',
  review_end_time: '',
  queue_type: 'all',
})

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
    if (searchForm.queue_type) {
      params.queue_type = searchForm.queue_type as 'first' | 'second' | 'all'
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
  searchForm.queue_type = 'all'
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
  if (!tableData.value.length) {
    ElMessage.warning('没有数据可导出')
    return
  }

  // 构建CSV内容
  const headers = [
    '任务ID',
    '评论ID',
    '评论内容',
    '队列类型',
    '一审审核员',
    '一审结果',
    '一审标签',
    '一审原因',
    '二审审核员',
    '二审结果',
    '二审标签',
    '二审原因',
    '领取时间',
    '完成时间',
  ]
  
  const csvContent = [
    headers.join(','),
    ...tableData.value.map(row => {
      const firstResult = row.queue_type === 'first' 
        ? (row.is_approved === true ? '通过' : row.is_approved === false ? '拒绝' : '-')
        : '拒绝'
      const secondResult = row.queue_type === 'second'
        ? (row.second_is_approved === true ? '通过' : row.second_is_approved === false ? '拒绝' : '-')
        : '-'
      
      const firstTags = row.queue_type === 'first' 
        ? (row.tags && row.tags.length > 0 ? row.tags.join(';') : '-')
        : (row.tags && row.tags.length > 0 ? row.tags.join(';') : '-')
      const secondTags = row.queue_type === 'second'
        ? (row.second_tags && row.second_tags.length > 0 ? row.second_tags.join(';') : '-')
        : '-'
      
      return [
        row.id,
        row.comment_id,
        `"${row.comment_text.replace(/"/g, '""')}"`,
        row.queue_type === 'first' ? '一审' : '二审',
        row.queue_type === 'first' ? row.username : (row.first_username || '-'),
        firstResult,
        firstTags,
        `"${(row.queue_type === 'first' ? (row.reason || '-') : (row.reason || '-')).replace(/"/g, '""')}"`,
        row.queue_type === 'second' ? (row.second_username || '-') : '-',
        secondResult,
        secondTags,
        `"${(row.queue_type === 'second' ? (row.second_reason || '-') : '-').replace(/"/g, '""')}"`,
        formatDate(row.claimed_at || ''),
        formatDate(row.completed_at || ''),
      ].join(',')
    }),
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

  ElMessage.success('导出成功')
}

onMounted(() => {
  loadTags()
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

