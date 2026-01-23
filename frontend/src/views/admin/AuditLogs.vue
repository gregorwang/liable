<template>
  <div class="audit-logs-page">
    <div class="page-header">
      <div>
        <h2 class="page-title">审计日志</h2>
        <p class="page-subtitle">追踪关键操作、快速定位异常行为</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="openExportDialog">导出日志</el-button>
        <el-button @click="openExportHistory">导出记录</el-button>
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

        <el-form-item label="用户名">
          <el-input v-model="filters.username" placeholder="模糊匹配" clearable />
        </el-form-item>

        <el-form-item label="用户ID">
          <el-input v-model="filters.userId" placeholder="精确匹配" clearable />
        </el-form-item>

        <el-form-item label="用户角色">
          <el-select v-model="filters.userRole" placeholder="全部" clearable>
            <el-option label="管理员" value="admin" />
            <el-option label="审核员" value="reviewer" />
          </el-select>
        </el-form-item>

        <el-form-item label="操作分类">
          <el-select
            v-model="filters.actionCategories"
            placeholder="多选"
            multiple
            collapse-tags
            clearable
          >
            <el-option v-for="item in actionCategoryOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>

        <el-form-item label="操作类型">
          <el-select
            v-model="filters.actionTypes"
            placeholder="多选"
            multiple
            filterable
            allow-create
            collapse-tags
            clearable
          >
            <el-option v-for="item in actionTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>

        <el-form-item label="结果">
          <el-select v-model="filters.result" placeholder="全部" clearable>
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failure" />
            <el-option label="部分成功" value="partial" />
          </el-select>
        </el-form-item>

        <el-form-item label="IP地址">
          <el-input v-model="filters.ipAddress" placeholder="支持 192.168.1.*" clearable />
        </el-form-item>

        <el-form-item label="端点">
          <el-input v-model="filters.endpoint" placeholder="/api/admin/*" clearable />
        </el-form-item>

        <el-form-item label="方法">
          <el-select v-model="filters.httpMethod" placeholder="全部" clearable>
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="PATCH" value="PATCH" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
        </el-form-item>

        <el-form-item label="状态码">
          <el-input v-model="filters.statusCode" placeholder="200" clearable />
        </el-form-item>

        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="描述或错误信息" clearable />
        </el-form-item>

        <el-form-item label="耗时(ms)">
          <div class="duration-range">
            <el-input-number v-model="filters.minDurationMs" :min="0" controls-position="right" placeholder="最小" />
            <span class="duration-separator">-</span>
            <el-input-number v-model="filters.maxDurationMs" :min="0" controls-position="right" placeholder="最大" />
          </div>
        </el-form-item>

        <el-form-item label="资源类型">
          <el-input v-model="filters.resourceType" placeholder="comment/video/user" clearable />
        </el-form-item>

        <el-form-item label="资源ID">
          <el-input v-model="filters.resourceId" placeholder="精确匹配" clearable />
        </el-form-item>

        <el-form-item label="设备类型">
          <el-select v-model="filters.deviceType" placeholder="全部" clearable>
            <el-option label="PC" value="PC" />
            <el-option label="Mobile" value="Mobile" />
            <el-option label="Tablet" value="Tablet" />
          </el-select>
        </el-form-item>
      </el-form>

      <div class="quick-actions">
        <div class="quick-buttons">
          <el-button size="small" @click="applyQuickRange('today')">今天</el-button>
          <el-button size="small" @click="applyQuickRange('yesterday')">昨天</el-button>
          <el-button size="small" @click="applyQuickRange('7d')">最近7天</el-button>
          <el-button size="small" @click="applyQuickRange('30d')">最近30天</el-button>
        </div>
        <div class="filter-actions">
          <el-button type="primary" @click="fetchLogs">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="table-card">
      <el-table :data="logs" style="width: 100%" v-loading="loading">
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="用户" width="160">
          <template #default="{ row }">
            <div class="user-cell">
              <span class="username">{{ row.username || '匿名' }}</span>
              <span class="user-id" v-if="row.user_id">#{{ row.user_id }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="220">
          <template #default="{ row }">
            <div class="action-cell">
              <div class="action-type">{{ row.action_type }}</div>
              <div class="action-desc">{{ row.action_description }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="结果" width="110">
          <template #default="{ row }">
            <el-tag :type="resultTagType(row.result)" size="small">
              {{ resultLabel(row.result) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="endpoint" label="端点" min-width="200" />
        <el-table-column prop="http_method" label="方法" width="90" />
        <el-table-column prop="status_code" label="状态" width="90" />
        <el-table-column prop="ip_address" label="IP" width="140" />
        <el-table-column prop="duration_ms" label="耗时(ms)" width="110" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="text" size="small" @click="openDetail(row.id)">详情</el-button>
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

    <el-drawer v-model="detailVisible" title="日志详情" size="45%">
      <div v-if="detailLoading" class="drawer-loading">加载中...</div>
      <div v-else-if="selectedLog" class="detail-content">
        <div class="detail-actions">
          <el-button size="small" @click="copyTraceId" :disabled="!selectedLog.request_id">
            复制 TraceID
          </el-button>
          <el-button size="small" type="primary" @click="exportErrorDetail">
            导出错误信息
          </el-button>
        </div>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="日志ID">{{ selectedLog.id }}</el-descriptions-item>
          <el-descriptions-item label="时间">{{ formatDate(selectedLog.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="用户">{{ selectedLog.username || '匿名' }}</el-descriptions-item>
          <el-descriptions-item label="用户ID">{{ selectedLog.user_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="角色">{{ selectedLog.user_role || '-' }}</el-descriptions-item>
          <el-descriptions-item label="操作类型">{{ selectedLog.action_type }}</el-descriptions-item>
          <el-descriptions-item label="操作分类">{{ selectedLog.action_category }}</el-descriptions-item>
          <el-descriptions-item label="描述">{{ selectedLog.action_description }}</el-descriptions-item>
          <el-descriptions-item label="结果">{{ resultLabel(selectedLog.result) }}</el-descriptions-item>
          <el-descriptions-item label="端点">{{ selectedLog.endpoint }}</el-descriptions-item>
          <el-descriptions-item label="HTTP方法">{{ selectedLog.http_method }}</el-descriptions-item>
          <el-descriptions-item label="状态码">{{ selectedLog.status_code }}</el-descriptions-item>
          <el-descriptions-item label="请求ID">{{ selectedLog.request_id }}</el-descriptions-item>
          <el-descriptions-item label="错误码">{{ selectedLog.error_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="错误类型">{{ selectedLog.error_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="错误描述">
            {{ selectedLog.error_description || selectedLog.error_message || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="IP">{{ selectedLog.ip_address }}</el-descriptions-item>
          <el-descriptions-item label="服务器IP">{{ selectedLog.server_ip || '-' }}</el-descriptions-item>
          <el-descriptions-item label="服务器端口">{{ selectedLog.server_port || '-' }}</el-descriptions-item>
          <el-descriptions-item label="模块">{{ selectedLog.module_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="函数">{{ selectedLog.method_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="耗时">{{ selectedLog.duration_ms }} ms</el-descriptions-item>
          <el-descriptions-item label="资源">{{ selectedLog.resource_type }} {{ selectedLog.resource_id }}</el-descriptions-item>
          <el-descriptions-item label="错误信息">{{ selectedLog.error_message || '-' }}</el-descriptions-item>
          <el-descriptions-item label="当前页面">{{ selectedLog.page_url || '-' }}</el-descriptions-item>
          <el-descriptions-item label="User-Agent">{{ selectedLog.user_agent || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="json-section">
          <h4>请求参数</h4>
          <pre class="json-block">{{ formatJSON(selectedLog.request_params || selectedLog.request_body) }}</pre>
        </div>
        <div class="json-section">
          <h4>响应数据</h4>
          <pre class="json-block">{{ formatJSON(selectedLog.response_body) }}</pre>
        </div>
        <div class="json-section" v-if="selectedLog.error_stack">
          <h4>错误详情</h4>
          <pre class="json-block">{{ selectedLog.error_stack }}</pre>
        </div>
        <div class="json-section" v-if="selectedLog.changes">
          <h4>变更信息</h4>
          <pre class="json-block">{{ formatJSON(selectedLog.changes) }}</pre>
        </div>
      </div>
    </el-drawer>

    <el-dialog v-model="exportVisible" title="导出审计日志" width="520px">
      <el-form :model="exportForm" label-width="90px">
        <el-form-item label="格式">
          <el-radio-group v-model="exportForm.format">
            <el-radio-button label="csv">CSV</el-radio-button>
            <el-radio-button label="xlsx">Excel</el-radio-button>
            <el-radio-button label="json">JSON</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="字段">
          <el-select v-model="exportForm.fields" multiple filterable collapse-tags placeholder="默认字段">
            <el-option v-for="field in exportFieldOptions" :key="field" :label="field" :value="field" />
          </el-select>
        </el-form-item>
        <el-form-item label="范围">
          <span>{{ exportRangeLabel }}</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="exportVisible = false">取消</el-button>
        <el-button type="primary" :loading="exportLoading" @click="submitExport">开始导出</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="exportHistoryVisible" title="导出记录" width="720px">
      <el-table :data="exportRecords" v-loading="exportHistoryLoading">
        <el-table-column prop="created_at" label="导出时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="export_format" label="格式" width="90" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'completed' ? 'success' : row.status === 'failed' ? 'danger' : 'info'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="row_count" label="行数" width="100" />
        <el-table-column prop="expires_at" label="过期时间" width="180">
          <template #default="{ row }">
            {{ row.expires_at ? formatDate(row.expires_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="下载" width="120">
          <template #default="{ row }">
            <el-button v-if="row.download_url" type="text" size="small" @click="downloadExport(row.download_url)">
              下载
            </el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { exportAuditLogs, getAuditLog, getAuditLogs, listAuditLogExports } from '../../api/auditLogs'
import type {
  AuditLogEntry,
  AuditLogExportRecord,
  AuditLogExportRequest,
  AuditLogQueryParams,
} from '../../types'

const loading = ref(false)
const logs = ref<AuditLogEntry[]>([])
const selectedLog = ref<AuditLogEntry | null>(null)
const detailVisible = ref(false)
const detailLoading = ref(false)

const exportVisible = ref(false)
const exportLoading = ref(false)
const exportHistoryVisible = ref(false)
const exportHistoryLoading = ref(false)
const exportRecords = ref<AuditLogExportRecord[]>([])

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const filters = reactive({
  dateRange: [] as string[],
  username: '',
  userId: '',
  userRole: '',
  actionCategories: [] as string[],
  actionTypes: [] as string[],
  result: '',
  ipAddress: '',
  endpoint: '',
  httpMethod: '',
  statusCode: '',
  keyword: '',
  minDurationMs: undefined as number | undefined,
  maxDurationMs: undefined as number | undefined,
  resourceType: '',
  resourceId: '',
  deviceType: '',
})

const exportForm = reactive({
  format: 'csv',
  fields: [] as string[],
})

const actionCategoryOptions = [
  { label: '认证', value: 'authentication' },
  { label: '授权', value: 'authorization' },
  { label: '用户管理', value: 'user_management' },
  { label: '内容审核', value: 'content_moderation' },
  { label: '配置管理', value: 'configuration' },
  { label: '数据操作', value: 'data_operation' },
  { label: '系统操作', value: 'system_operation' },
]

const actionTypeOptions = [
  { label: '登录', value: 'auth.login' },
  { label: '注册', value: 'auth.register' },
  { label: '授予权限', value: 'permission.grant' },
  { label: '撤销权限', value: 'permission.revoke' },
  { label: '用户审批', value: 'user.approve' },
  { label: '领取任务', value: 'review.claim' },
  { label: '提交审核', value: 'review.submit' },
  { label: '批量提交', value: 'review.submit_batch' },
  { label: '退回任务', value: 'review.return' },
  { label: '视频审核', value: 'video.review' },
  { label: '导出数据', value: 'data.export' },
]

const exportFieldOptions = [
  'created_at',
  'user_id',
  'username',
  'user_role',
  'action_type',
  'action_category',
  'action_description',
  'result',
  'endpoint',
  'http_method',
  'status_code',
  'request_id',
  'session_id',
  'ip_address',
  'user_agent',
  'server_ip',
  'server_port',
  'module_name',
  'method_name',
  'page_url',
  'error_code',
  'error_type',
  'error_description',
  'resource_type',
  'resource_id',
  'error_message',
  'error_stack',
  'duration_ms',
  'request_body',
  'request_params',
  'response_body',
  'resource_ids',
  'changes',
]

const exportRangeLabel = computed(() => {
  if (filters.dateRange.length !== 2) {
    return '未选择'
  }
  return `${filters.dateRange[0]} 至 ${filters.dateRange[1]}`
})

const fetchLogs = async () => {
  if (filters.dateRange.length !== 2) {
    ElMessage.warning('请选择时间范围')
    return
  }

  const userId = filters.userId ? Number(filters.userId) : undefined
  const statusCode = filters.statusCode ? Number(filters.statusCode) : undefined

  loading.value = true
  try {
    const params: AuditLogQueryParams = {
      start_time: filters.dateRange[0],
      end_time: filters.dateRange[1],
      username: filters.username || undefined,
      user_id: Number.isFinite(userId) ? userId : undefined,
      user_role: filters.userRole || undefined,
      action_categories: filters.actionCategories.length ? filters.actionCategories.join(',') : undefined,
      action_types: filters.actionTypes.length ? filters.actionTypes.join(',') : undefined,
      result: filters.result || undefined,
      ip_address: filters.ipAddress || undefined,
      endpoint: filters.endpoint || undefined,
      http_method: filters.httpMethod || undefined,
      status_code: Number.isFinite(statusCode) ? statusCode : undefined,
      keyword: filters.keyword || undefined,
      min_duration_ms: filters.minDurationMs || undefined,
      max_duration_ms: filters.maxDurationMs || undefined,
      resource_type: filters.resourceType || undefined,
      resource_id: filters.resourceId || undefined,
      device_type: filters.deviceType || undefined,
      page: pagination.page,
      page_size: pagination.pageSize,
      sort_by: 'created_at',
      sort_order: 'desc',
    }
    const response = await getAuditLogs(params)
    logs.value = response.data
    pagination.total = response.total
  } catch (error) {
    console.error('Failed to fetch audit logs', error)
    ElMessage.error('获取审计日志失败')
  } finally {
    loading.value = false
  }
}

const openDetail = async (id: string) => {
  detailVisible.value = true
  detailLoading.value = true
  try {
    selectedLog.value = await getAuditLog(id)
  } catch (error) {
    console.error('Failed to fetch audit log detail', error)
    ElMessage.error('获取日志详情失败')
  } finally {
    detailLoading.value = false
  }
}

const copyTraceId = async () => {
  if (!selectedLog.value?.request_id) {
    ElMessage.warning('暂无 TraceID')
    return
  }
  try {
    await navigator.clipboard.writeText(selectedLog.value.request_id)
    ElMessage.success('TraceID 已复制')
  } catch (error) {
    console.error('Failed to copy TraceID', error)
    ElMessage.error('复制 TraceID 失败')
  }
}

const exportErrorDetail = () => {
  if (!selectedLog.value) {
    return
  }
  const payload = {
    exported_at: new Date().toISOString(),
    ...selectedLog.value,
  }
  const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  const traceId = selectedLog.value.request_id || selectedLog.value.id
  link.href = url
  link.download = `error-${traceId}.json`
  link.click()
  URL.revokeObjectURL(url)
}

const resetFilters = () => {
  filters.username = ''
  filters.userId = ''
  filters.userRole = ''
  filters.actionCategories = []
  filters.actionTypes = []
  filters.result = ''
  filters.ipAddress = ''
  filters.endpoint = ''
  filters.httpMethod = ''
  filters.statusCode = ''
  filters.keyword = ''
  filters.minDurationMs = undefined
  filters.maxDurationMs = undefined
  filters.resourceType = ''
  filters.resourceId = ''
  filters.deviceType = ''
  applyQuickRange('7d')
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchLogs()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchLogs()
}

const applyQuickRange = (type: 'today' | 'yesterday' | '7d' | '30d') => {
  let end = new Date()
  let start = new Date()
  if (type === 'today') {
    start = new Date(end.getFullYear(), end.getMonth(), end.getDate())
  } else if (type === 'yesterday') {
    start = new Date(end.getFullYear(), end.getMonth(), end.getDate() - 1)
    end = new Date(end.getFullYear(), end.getMonth(), end.getDate() - 1, 23, 59, 59)
  } else if (type === '7d') {
    start.setDate(end.getDate() - 6)
  } else {
    start.setDate(end.getDate() - 29)
  }

  filters.dateRange = [formatDateTime(start), formatDateTime(end)]
  pagination.page = 1
  fetchLogs()
}

const openExportDialog = () => {
  exportVisible.value = true
}

const submitExport = async () => {
  if (filters.dateRange.length !== 2) {
    ElMessage.warning('请选择时间范围')
    return
  }

  const userId = filters.userId ? Number(filters.userId) : undefined
  const statusCode = filters.statusCode ? Number(filters.statusCode) : undefined

  exportLoading.value = true
  try {
    const payload: AuditLogExportRequest = {
      start_time: filters.dateRange[0],
      end_time: filters.dateRange[1],
      format: exportForm.format as 'csv' | 'json' | 'xlsx',
      fields: exportForm.fields.length ? exportForm.fields : undefined,
      username: filters.username || undefined,
      user_id: Number.isFinite(userId) ? userId : undefined,
      user_role: filters.userRole || undefined,
      action_categories: filters.actionCategories.length ? filters.actionCategories : undefined,
      action_types: filters.actionTypes.length ? filters.actionTypes : undefined,
      result: filters.result || undefined,
      ip_address: filters.ipAddress || undefined,
      endpoint: filters.endpoint || undefined,
      http_method: filters.httpMethod || undefined,
      status_code: Number.isFinite(statusCode) ? statusCode : undefined,
      keyword: filters.keyword || undefined,
      min_duration_ms: filters.minDurationMs || undefined,
      max_duration_ms: filters.maxDurationMs || undefined,
      resource_type: filters.resourceType || undefined,
      resource_id: filters.resourceId || undefined,
      device_type: filters.deviceType || undefined,
    }

    const response = await exportAuditLogs(payload)
    ElMessage.success(`导出完成，共 ${response.row_count} 条`)
    window.open(response.download_url, '_blank')
    exportVisible.value = false
  } catch (error) {
    console.error('Failed to export audit logs', error)
    ElMessage.error('导出失败')
  } finally {
    exportLoading.value = false
  }
}

const openExportHistory = async () => {
  exportHistoryVisible.value = true
  exportHistoryLoading.value = true
  try {
    const response = await listAuditLogExports({ page: 1, page_size: 20 })
    exportRecords.value = response.data
  } catch (error) {
    console.error('Failed to load export history', error)
    ElMessage.error('获取导出记录失败')
  } finally {
    exportHistoryLoading.value = false
  }
}

const downloadExport = (url: string) => {
  window.open(url, '_blank')
}

const resultTagType = (result?: string) => {
  if (result === 'success') return 'success'
  if (result === 'failure') return 'danger'
  if (result === 'partial') return 'warning'
  return ''
}

const resultLabel = (result?: string) => {
  if (result === 'success') return '成功'
  if (result === 'failure') return '失败'
  if (result === 'partial') return '部分成功'
  return result || '-'
}

const formatDateTime = (date: Date) => {
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

const formatDate = (value?: string) => {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}

const formatJSON = (value: any) => {
  if (!value) return '-'
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

onMounted(() => {
  applyQuickRange('7d')
})
</script>

<style scoped>
.audit-logs-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-4);
}

.page-title {
  margin: 0;
  font-size: var(--text-xl);
  color: var(--color-text-000);
  font-weight: 600;
}

.page-subtitle {
  margin: var(--spacing-2) 0 0;
  font-size: var(--text-sm);
  color: var(--color-text-400);
}

.header-actions {
  display: flex;
  gap: var(--spacing-3);
}

.filter-card {
  padding-bottom: var(--spacing-4);
}

.filter-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--spacing-2) var(--spacing-4);
}

.filter-date-range {
  width: 100%;
}

.duration-range {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.duration-separator {
  color: var(--color-text-400);
}

.quick-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--spacing-4);
}

.quick-buttons {
  display: flex;
  gap: var(--spacing-2);
}

.filter-actions {
  display: flex;
  gap: var(--spacing-3);
}

.table-card {
  padding-bottom: var(--spacing-4);
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-4);
}

.user-cell {
  display: flex;
  flex-direction: column;
}

.username {
  font-weight: 500;
}

.user-id {
  font-size: var(--text-xs);
  color: var(--color-text-400);
}

.action-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.action-type {
  font-size: var(--text-sm);
  font-weight: 500;
}

.action-desc {
  font-size: var(--text-xs);
  color: var(--color-text-400);
}

.drawer-loading {
  text-align: center;
  padding: var(--spacing-6);
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
}

.detail-actions {
  display: flex;
  gap: var(--spacing-2);
}

.json-section h4 {
  margin-bottom: var(--spacing-2);
  font-size: var(--text-sm);
  color: var(--color-text-200);
}

.json-block {
  background: #f7f7f7;
  padding: var(--spacing-3);
  border-radius: var(--radius-md);
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--color-text-200);
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
  }

  .filter-form {
    grid-template-columns: 1fr;
  }

  .quick-actions {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-3);
  }
}
</style>
