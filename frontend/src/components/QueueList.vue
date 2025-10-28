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
          prop="queue_name"
          label="队列名称"
          width="150"
          show-overflow-tooltip
        />
        
        <el-table-column
          prop="description"
          label="描述"
          width="200"
          show-overflow-tooltip
        />
        
        <el-table-column
          prop="priority"
          label="优先级"
          width="100"
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
          width="180"
          align="center"
        >
          <template #default="{ row }">
            <span>{{ row.total_tasks }}/{{ row.completed_tasks }}/{{ row.pending_tasks }}</span>
          </template>
        </el-table-column>
        
        <el-table-column
          label="进度"
          width="120"
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
          width="100"
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
          prop="created_at"
          label="创建时间"
          width="170"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column
          label="操作"
          width="200"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
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
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { listTaskQueuesPublic } from '../api/admin'
import type { TaskQueue } from '../types'

const router = useRouter()

// 响应式数据
const loading = ref(false)
const tableData = ref<TaskQueue[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 方法
const calculateProgress = (row: TaskQueue) => {
  if (row.total_tasks === 0) return 0
  return Math.round((row.completed_tasks / row.total_tasks) * 100)
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
    tableData.value = response.data
    total.value = response.total
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('Failed to load data:', error)
  } finally {
    loading.value = false
  }
}

const handleRefresh = () => {
  loadData()
  ElMessage.success('刷新成功')
}

const handleAnnotate = (row: TaskQueue) => {
  // 根据队列类型跳转到不同的审核界面
  sessionStorage.setItem('currentQueue', JSON.stringify(row))
  
  // 根据队列名称判断跳转到哪个审核界面
  if (row.queue_name === '评论审核二审') {
    router.push('/reviewer/second-review')
  } else {
    // 其他队列（包括一审队列）跳转到通用的一审审核界面
    router.push('/reviewer/dashboard')
  }
}

const handleViewDetails = (row: TaskQueue) => {
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
