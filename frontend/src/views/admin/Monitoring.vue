<template>
  <div class="monitoring-page">
    <div class="page-header">
      <div>
        <div class="page-title">系统监控大盘</div>
        <div class="page-subtitle">每日错误日志与接口健康统计</div>
      </div>
      <div class="page-actions">
        <el-date-picker
          v-model="selectedDate"
          type="date"
          placeholder="选择日期"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          @change="handleDateChange"
        />
        <el-button size="small" type="primary" @click="loadAll">刷新</el-button>
      </div>
    </div>

    <div class="summary-grid" v-loading="summaryLoading">
      <el-card class="summary-card">
        <div class="summary-label">今日请求总数</div>
        <div class="summary-value">{{ formatNumber(summary?.total_requests ?? 0) }}</div>
        <div class="summary-sub">错误率 {{ formatPercent(errorRate) }}</div>
      </el-card>
      <el-card class="summary-card">
        <div class="summary-label">今日错误请求</div>
        <div class="summary-value error">{{ formatNumber(summary?.error_requests ?? 0) }}</div>
        <div class="summary-sub">含 4xx/5xx</div>
      </el-card>
      <el-card class="summary-card">
        <div class="summary-label">4xx 客户端错误</div>
        <div class="summary-value warning">{{ formatNumber(summary?.client_errors ?? 0) }}</div>
        <div class="summary-sub">触发阈值告警</div>
      </el-card>
      <el-card class="summary-card">
        <div class="summary-label">5xx 服务端错误</div>
        <div class="summary-value error">{{ formatNumber(summary?.server_errors ?? 0) }}</div>
        <div class="summary-sub">立即告警</div>
      </el-card>
    </div>

    <el-card class="section-card">
      <div class="section-header">
        <div class="section-title">当日错误日志</div>
        <div class="section-sub">自动筛选非 2xx/3xx 记录</div>
      </div>
      <el-table :data="errorLogs" v-loading="errorLoading" stripe class="data-table">
        <el-table-column label="时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="接口" min-width="200">
          <template #default="{ row }">
            <span class="mono">{{ row.http_method }} {{ row.endpoint }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status_code" label="状态码" width="90" />
        <el-table-column prop="error_message" label="错误信息" min-width="220" show-overflow-tooltip />
        <el-table-column prop="request_id" label="Trace ID" min-width="220" show-overflow-tooltip />
        <el-table-column label="用户" width="140">
          <template #default="{ row }">
            <span>{{ row.username || '-' }}</span>
          </template>
        </el-table-column>
      </el-table>
      <div class="table-footer">
        <el-pagination
          :current-page="errorPage"
          :page-size="errorPageSize"
          :total="errorTotal"
          layout="prev, pager, next"
          @current-change="loadErrors"
        />
      </div>
    </el-card>

    <el-card class="section-card">
      <div class="section-header">
        <div class="section-title">接口健康统计（当日）</div>
        <div class="section-sub">成功率、平均耗时与 P99 延迟</div>
      </div>
      <el-table :data="endpoints" v-loading="endpointsLoading" stripe class="data-table">
        <el-table-column label="接口" min-width="240">
          <template #default="{ row }">
            <span class="mono">{{ row.method }} {{ row.path }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="total" label="调用量" width="100" />
        <el-table-column label="成功率" width="120">
          <template #default="{ row }">{{ formatPercent(row.success_rate) }}</template>
        </el-table-column>
        <el-table-column prop="client_error" label="4xx" width="80" />
        <el-table-column prop="server_error" label="5xx" width="80" />
        <el-table-column label="平均耗时" width="120">
          <template #default="{ row }">{{ formatLatency(row.avg_latency_ms) }}</template>
        </el-table-column>
        <el-table-column label="P99" width="100">
          <template #default="{ row }">{{ formatLatency(row.p99_latency_ms) }}</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getAuditLogs } from '../../api/auditLogs'
import { getMonitoringEndpoints, getMonitoringSummary } from '../../api/monitoring'
import type { AuditLogEntry, EndpointHealthStat, MonitoringSummary } from '../../types'

const selectedDate = ref<string>(new Date().toISOString().slice(0, 10))

const summary = ref<MonitoringSummary | null>(null)
const summaryLoading = ref(false)

const errorLogs = ref<AuditLogEntry[]>([])
const errorTotal = ref(0)
const errorLoading = ref(false)
const errorPage = ref(1)
const errorPageSize = 20

const endpoints = ref<EndpointHealthStat[]>([])
const endpointsLoading = ref(false)

const errorRate = computed(() => {
  if (!summary.value || summary.value.total_requests === 0) return 0
  return summary.value.error_requests / summary.value.total_requests
})

const handleDateChange = () => {
  errorPage.value = 1
  loadAll()
}

const loadSummary = async () => {
  summaryLoading.value = true
  try {
    summary.value = await getMonitoringSummary({ date: selectedDate.value })
  } catch (error) {
    console.error('Failed to load monitoring summary', error)
  } finally {
    summaryLoading.value = false
  }
}

const loadErrors = async (page?: number) => {
  errorLoading.value = true
  const currentPage = page ?? errorPage.value
  try {
    const startTime = `${selectedDate.value} 00:00:00`
    const endTime = `${selectedDate.value} 23:59:59`
    const response = await getAuditLogs({
      start_time: startTime,
      end_time: endTime,
      result: 'failure',
      page: currentPage,
      page_size: errorPageSize,
      sort_order: 'desc',
    })
    errorLogs.value = response.data || []
    errorTotal.value = response.total || 0
    errorPage.value = currentPage
  } catch (error) {
    console.error('Failed to load error logs', error)
  } finally {
    errorLoading.value = false
  }
}

const loadEndpoints = async () => {
  endpointsLoading.value = true
  try {
    const response = await getMonitoringEndpoints({ date: selectedDate.value, limit: 200 })
    endpoints.value = response.endpoints || []
  } catch (error) {
    console.error('Failed to load endpoint health stats', error)
  } finally {
    endpointsLoading.value = false
  }
}

const loadAll = async () => {
  await Promise.all([loadSummary(), loadErrors(), loadEndpoints()])
}

const formatNumber = (value: number) => value.toLocaleString('zh-CN')

const formatPercent = (value: number) => `${(value * 100).toFixed(1)}%`

const formatLatency = (value: number) => `${Math.round(value)}ms`

const formatDateTime = (value: string) => {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}

onMounted(() => {
  loadAll()
})
</script>

<style scoped>
.monitoring-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 12px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text-100);
}

.page-subtitle {
  font-size: 13px;
  color: var(--color-text-400);
  margin-top: 4px;
}

.page-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}

.summary-card {
  border-radius: 12px;
}

.summary-label {
  font-size: 13px;
  color: var(--color-text-400);
}

.summary-value {
  font-size: 26px;
  font-weight: 600;
  margin-top: 6px;
  color: var(--color-text-100);
}

.summary-value.error {
  color: #d23f3f;
}

.summary-value.warning {
  color: #d48806;
}

.summary-sub {
  margin-top: 6px;
  font-size: 12px;
  color: var(--color-text-400);
}

.section-card {
  border-radius: 12px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 12px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-100);
}

.section-sub {
  font-size: 12px;
  color: var(--color-text-400);
}

.data-table {
  width: 100%;
}

.table-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;
}

.mono {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 12px;
}
</style>
