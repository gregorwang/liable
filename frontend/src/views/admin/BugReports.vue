<template>
  <div class="bug-reports-page">
    <div class="page-header">
      <div>
        <h2 class="page-title">Bug 反馈</h2>
        <p class="page-subtitle">集中查看用户提交的异常反馈与截图</p>
      </div>
      <div class="header-actions">
        <el-dropdown @command="handleExport">
          <el-button type="primary" plain>
            导出
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="csv">导出 CSV</el-dropdown-item>
              <el-dropdown-item command="json">导出 JSON</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button type="primary" @click="fetchReports">刷新</el-button>
      </div>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" class="filter-form">
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filters.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            class="filter-date-range"
          />
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input v-model="filters.userId" placeholder="精确匹配" clearable />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="filters.username" placeholder="模糊匹配" clearable />
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="描述/错误/URL" clearable />
        </el-form-item>
      </el-form>

      <div class="filter-actions">
        <el-button type="primary" @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </el-card>

    <el-card class="table-card">
      <el-table :data="reports" style="width: 100%" v-loading="loading">
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="用户" width="160">
          <template #default="{ row }">
            <div class="user-cell">
              <span class="username">{{ row.username }}</span>
              <span class="user-id">#{{ row.user_id }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="标题" width="180">
          <template #default="{ row }">
            <span>{{ row.title || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="问题描述" min-width="260">
          <template #default="{ row }">
            <div class="ellipsis">{{ row.description }}</div>
          </template>
        </el-table-column>
        <el-table-column label="页面" min-width="220">
          <template #default="{ row }">
            <div class="ellipsis">{{ row.page_url || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="截图" width="90">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.screenshots?.length || 0 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="text" size="small" @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <el-drawer v-model="detailVisible" title="Bug 详情" size="45%">
      <div v-if="selectedReport" class="detail-content">
        <div class="detail-section">
          <div class="detail-label">用户</div>
          <div class="detail-value">{{ selectedReport.username }} (#{{ selectedReport.user_id }})</div>
        </div>
        <div class="detail-section">
          <div class="detail-label">时间</div>
          <div class="detail-value">{{ formatDate(selectedReport.created_at) }}</div>
        </div>
        <div class="detail-section">
          <div class="detail-label">标题</div>
          <div class="detail-value">{{ selectedReport.title || '-' }}</div>
        </div>
        <div class="detail-section">
          <div class="detail-label">问题描述</div>
          <div class="detail-value">{{ selectedReport.description }}</div>
        </div>
        <div class="detail-section">
          <div class="detail-label">错误信息</div>
          <pre class="detail-pre">{{ selectedReport.error_details || '-' }}</pre>
        </div>
        <div class="detail-section">
          <div class="detail-label">页面 URL</div>
          <div class="detail-value detail-link">{{ selectedReport.page_url || '-' }}</div>
        </div>
        <div class="detail-section">
          <div class="detail-label">User-Agent</div>
          <div class="detail-value detail-pre">{{ selectedReport.user_agent || '-' }}</div>
        </div>
        <div class="detail-section">
          <div class="detail-label">截图</div>
          <div class="screenshots">
            <div v-if="!selectedReport.screenshots?.length" class="empty-state">未上传截图</div>
            <el-image
              v-for="shot in selectedReport.screenshots"
              :key="shot.key"
              :src="shot.url || ''"
              :preview-src-list="previewList"
              :initial-index="previewIndex(shot.key)"
              fit="cover"
              class="screenshot-item"
            >
              <template #error>
                <div class="screenshot-fallback">
                  <span>无法预览</span>
                  <span class="fallback-key">{{ shot.key }}</span>
                </div>
              </template>
            </el-image>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { exportBugReports, listBugReports } from '@/api/bugReports'
import type { BugReportAdminItem } from '@/types'

const reports = ref<BugReportAdminItem[]>([])
const loading = ref(false)
const detailVisible = ref(false)
const selectedReport = ref<BugReportAdminItem | null>(null)

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const filters = reactive({
  dateRange: [] as string[],
  userId: '',
  username: '',
  keyword: '',
})

const previewList = computed(() => {
  if (!selectedReport.value) return []
  return selectedReport.value.screenshots
    .map((shot) => shot.url)
    .filter((url): url is string => Boolean(url))
})

const previewIndex = (key: string) => {
  if (!selectedReport.value) return 0
  const index = selectedReport.value.screenshots.findIndex((shot) => shot.key === key)
  return index === -1 ? 0 : index
}

const buildQueryParams = () => {
  const params: Record<string, string | number> = {
    page: pagination.page,
    page_size: pagination.pageSize,
  }

  if (filters.dateRange.length === 2) {
    params.start_time = filters.dateRange[0]
    params.end_time = filters.dateRange[1]
  }

  const userIdValue = Number(filters.userId)
  if (filters.userId.trim() && !Number.isNaN(userIdValue)) {
    params.user_id = userIdValue
  }

  if (filters.username.trim()) {
    params.username = filters.username.trim()
  }

  if (filters.keyword.trim()) {
    params.keyword = filters.keyword.trim()
  }

  return params
}

const fetchReports = async () => {
  loading.value = true
  try {
    const response = await listBugReports(buildQueryParams())
    reports.value = response.data
    pagination.total = response.total
  } catch (error) {
    console.error('Failed to fetch bug reports', error)
    ElMessage.error('获取 Bug 反馈失败')
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  pagination.page = 1
  fetchReports()
}

const resetFilters = () => {
  filters.dateRange = []
  filters.userId = ''
  filters.username = ''
  filters.keyword = ''
  pagination.page = 1
  fetchReports()
}

const handleExport = async (command: string) => {
  try {
    const format = command === 'json' ? 'json' : 'csv'
    const params = buildQueryParams()
    const payload = {
      start_time: params.start_time as string | undefined,
      end_time: params.end_time as string | undefined,
      user_id: params.user_id as number | undefined,
      username: params.username as string | undefined,
      keyword: params.keyword as string | undefined,
      format,
    }
    const blob = await exportBugReports(payload)
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `bug-reports-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, '')}.${format}`
    link.click()
    window.URL.revokeObjectURL(url)
  } catch (error) {
    console.error('Failed to export bug reports', error)
    ElMessage.error('导出失败')
  }
}

const openDetail = (report: BugReportAdminItem) => {
  selectedReport.value = report
  detailVisible.value = true
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchReports()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchReports()
}

const formatDate = (value: string) => {
  if (!value) return '-'
  const date = new Date(value)
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  fetchReports()
})
</script>

<style scoped>
.bug-reports-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-6);
}

.page-title {
  margin: 0;
  font-size: var(--text-2xl);
  color: var(--color-text-000);
}

.page-subtitle {
  margin: var(--spacing-2) 0 0;
  color: var(--color-text-300);
  font-size: var(--text-sm);
}

.filter-card {
  border-radius: var(--radius-lg);
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-4);
}

.filter-date-range {
  min-width: 280px;
}

.filter-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-3);
  margin-top: var(--spacing-4);
}

.table-card {
  border-radius: var(--radius-lg);
}

.user-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.username {
  font-weight: 600;
  color: var(--color-text-100);
}

.user-id {
  font-size: var(--text-xs);
  color: var(--color-text-400);
}

.ellipsis {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  color: var(--color-text-200);
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-4);
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-4);
}

.detail-section {
  display: grid;
  grid-template-columns: 90px 1fr;
  gap: var(--spacing-3);
}

.detail-label {
  font-size: var(--text-sm);
  color: var(--color-text-400);
}

.detail-value {
  color: var(--color-text-100);
  font-size: var(--text-sm);
}

.detail-pre {
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--color-bg-100);
  border-radius: var(--radius-sm);
  padding: var(--spacing-3);
  font-size: var(--text-xs);
  color: var(--color-text-200);
}

.detail-link {
  word-break: break-all;
}

.screenshots {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: var(--spacing-3);
}

.screenshot-item {
  width: 100%;
  height: 120px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border-200);
  overflow: hidden;
}

.screenshot-fallback {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  gap: var(--spacing-2);
  font-size: var(--text-xs);
  color: var(--color-text-400);
  padding: var(--spacing-2);
  text-align: center;
}

.fallback-key {
  font-size: 10px;
  word-break: break-all;
}

.empty-state {
  font-size: var(--text-xs);
  color: var(--color-text-400);
}
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .filter-actions {
    justify-content: flex-start;
  }

  .detail-section {
    grid-template-columns: 1fr;
  }
}
</style>
