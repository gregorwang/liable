<template>
  <div class="queue-list">
    <div class="page-header">
      <h2>队列列表</h2>
      <div class="header-actions">
        <el-button type="primary" @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <el-card class="table-card">
      <el-table
        :data="tableData"
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column
          label="队列信息"
          min-width="260"
        >
          <template #default="{ row }">
            <div class="queue-info">
              <div class="queue-title">{{ getQueueDisplayName(row) }}</div>
              <div class="queue-subtitle">{{ formatQueueIdentifier(row.queue_name) }}</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column
          prop="priority"
          label="优先级"
          min-width="120"
          align="center"
        >
          <template #default="{ row }">
            <el-tag :type="getPriorityType(row.priority)">
              {{ row.priority }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column
          label="任务统计"
          min-width="200"
          align="center"
        >
          <template #default="{ row }">
            <div class="task-breakdown">
              <span>总 {{ row.total_tasks }}</span>
              <span>完成 {{ row.completed_tasks }}</span>
              <span>待审 {{ row.pending_tasks }}</span>
              <span>处理中 {{ getInProgressCount(row) }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column
          label="进度"
          min-width="140"
          align="center"
        >
          <template #default="{ row }">
            <el-progress
              :percentage="calculateProgress(row)"
              :color="getProgressColor(calculateProgress(row))"
              :stroke-width="6"
            />
          </template>
        </el-table-column>
        
        <el-table-column
          label="状态"
          min-width="120"
          align="center"
        >
          <template #default="{ row }">
            <el-tag
              :type="row.is_active ? 'success' : 'info'"
              size="small"
            >
              {{ row.is_active ? '活跃' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          label="创建人"
          min-width="120"
          align="center"
        >
          <template #default>
            <span class="queue-owner">admin</span>
          </template>
        </el-table-column>
        
        <el-table-column
          prop="created_at"
          label="创建时间"
          min-width="180"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column
          label="操作"
          min-width="220"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              :disabled="!canAnnotate(row)"
              :title="!canAnnotate(row) ? '队列暂无任务' : '进入标注工作台'"
              @click="handleAnnotate(row)"
            >
              标注
            </el-button>
            <el-button
              type="info"
              size="small"
              @click="handleViewDetails(row)"
            >
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页器 -->
      <div class="pagination-container">
        <el-pagination
          :current-page="currentPage"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, inject } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { listTaskQueuesPublic } from '../api/admin'
import type { TaskQueue } from '../types'

const router = useRouter()
const setActiveMenu = inject<(menu: string) => void>('setActiveMenu')

// 响应式数据
const loading = ref(false)
const tableData = ref<TaskQueue[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 方法
const queueRouteMap: Record<string, string> = {
  comment_first_review: '/reviewer/dashboard',
  comment_second_review: '/reviewer/second-review',
  ai_human_diff: '/reviewer/ai-human-diff',
  quality_check: '/reviewer/quality-check',
}

const queueDisplayNameMap: Record<string, string> = {
  comment_first_review: '评论一审队列',
  comment_second_review: '评论二审队列',
  ai_human_diff: 'AI与人工diff队列',
  quality_check: '质量检查队列',
}

const normalizeQueueName = (queueName: string) => queueName?.toLowerCase() || ''

const calculateProgress = (row: TaskQueue) => {
  if (row.total_tasks === 0) return 0
  return Math.round((row.completed_tasks / row.total_tasks) * 100)
}

const getInProgressCount = (row: TaskQueue) => {
  const inProgress = row.total_tasks - row.completed_tasks - row.pending_tasks
  return inProgress > 0 ? inProgress : 0
}

const getProgressColor = (percentage: number) => {
  if (percentage < 30) return '#f56c6c'
  if (percentage < 70) return '#e6a23c'
  return '#67c23a'
}

const getPriorityType = (priority: number) => {
  if (priority >= 80) return 'danger'
  if (priority >= 50) return 'warning'
  return 'info'
}

const formatQueueName = (queueName: string) => {
  const normalized = normalizeQueueName(queueName)
  if (queueDisplayNameMap[normalized]) {
    return queueDisplayNameMap[normalized]
  }
  return queueName
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (char) => char.toUpperCase())
}

const formatQueueIdentifier = (queueName: string) => {
  return queueName.replace(/_/g, ' ').toUpperCase()
}

const getQueueDisplayName = (row: TaskQueue) => {
  if (row.description && row.description.trim().length > 0) {
    return row.description
  }
  return formatQueueName(row.queue_name)
}

const canAnnotate = (row: TaskQueue) => row.pending_tasks > 0

const resolveQueueRoute = (queueName: string) => {
  const normalized = normalizeQueueName(queueName)
  if (queueRouteMap[normalized]) {
    return queueRouteMap[normalized]
  }
  if (normalized.includes('video') && normalized.includes('queue')) {
    return '/reviewer/video-queue-review'
  }
  return '/reviewer/dashboard'
}

const formatDate = (dateStr: string) => {
  try {
    return new Date(dateStr).toLocaleString('zh-CN')
  } catch {
    return dateStr
  }
}

const loadData = async () => {
  loading.value = true
  try {
    const response = await listTaskQueuesPublic({
      page: currentPage.value,
      page_size: pageSize.value,
    })
    
    // Ensure data is always an array, even if backend returns null
    tableData.value = response.data || []
    total.value = response.total || 0
  } catch (error) {
    console.error('Failed to load data:', error)
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

const handleRefresh = () => {
  loadData()
  ElMessage.success('刷新成功')
}

const handleAnnotate = (row: TaskQueue) => {
  if (!canAnnotate(row)) {
    ElMessage.warning('队列暂无任务，稍后再试')
    return
  }

  sessionStorage.setItem('currentQueue', JSON.stringify(row))
  router.push(resolveQueueRoute(row.queue_name))
}

const handleViewDetails = (row: TaskQueue) => {
  sessionStorage.setItem('queueSearchQueue', row.queue_name)
  if (setActiveMenu) {
    setActiveMenu('data-management')
    return
  }
  ElMessage.info(`队列: ${row.queue_name} - 优先级: ${row.priority}`)
}

const handleSizeChange = (val: number) => {
  pageSize.value = val
  currentPage.value = 1
  loadData()
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.queue-list {
  background: transparent;
  min-height: calc(100vh - 100px);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-6);
  padding: 0 var(--spacing-1);
}

.page-header h2 {
  margin: 0;
  font-size: var(--text-2xl);
  font-weight: 600;
  color: var(--color-text-000);
  font-family: var(--font-sans);
}

.header-actions {
  display: flex;
  gap: var(--spacing-3);
}

.table-card {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  border-radius: var(--radius-lg);
  border: 1px solid rgba(204, 122, 77, 0.08);
  background-color: rgba(255, 255, 255, 0.9);
}

.queue-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--spacing-1);
}

.queue-title {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-000);
}

.queue-subtitle {
  font-size: var(--text-xs);
  color: var(--color-text-400);
  letter-spacing: var(--tracking-wide);
}

.task-breakdown {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: var(--text-xs);
  color: var(--color-text-200);
}

.queue-owner {
  font-weight: 500;
  color: var(--color-text-200);
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-6);
  padding: var(--spacing-4) 0;
}

/* 表格样式优化 */
:deep(.el-table) {
  border-radius: var(--radius-lg);
  overflow: hidden;
  font-family: var(--font-sans);
  background-color: rgba(255, 255, 255, 0.8) !important;
}

:deep(.el-table th) {
  background-color: rgba(250, 247, 245, 0.6) !important;
  color: var(--color-text-100) !important;
  font-weight: 600;
  border-bottom: 1px solid rgba(204, 122, 77, 0.1) !important;
}

:deep(.el-table td) {
  border-bottom: 1px solid rgba(204, 122, 77, 0.06);
  color: var(--color-text-100);
  font-family: var(--font-sans);
  background-color: rgba(255, 255, 255, 0.5);
}

:deep(.el-table--striped .el-table__body tr.el-table__row--striped td) {
  background-color: rgba(250, 247, 245, 0.4);
}

:deep(.el-table tbody tr:hover > td) {
  background-color: rgba(250, 247, 245, 0.7) !important;
}

/* 进度条样式 */
:deep(.el-progress__text) {
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text-100);
}

/* 按钮组样式 */
:deep(.el-button + .el-button) {
  margin-left: var(--spacing-2);
}

/* 标签样式 */
:deep(.el-tag) {
  font-family: var(--font-sans);
  font-weight: 500;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-4);
  }
  
  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }
}

@media (max-width: 768px) {
  .page-header h2 {
    font-size: var(--text-xl);
  }
  
  .pagination-container {
    justify-content: center;
  }
  
  :deep(.el-pagination) {
    flex-wrap: wrap;
  }
}
</style>
