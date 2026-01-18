<template>
  <div class="generic-review-dashboard">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <div class="header-left">
            <h2>{{ config.title }}</h2>
            <el-button v-if="config.showSearch" type="primary" link @click="$emit('search')" style="margin-left: 20px">
              <el-icon><Search /></el-icon>
              搜索审核记录
            </el-button>
          </div>
          <div class="user-info">
            <span>欢迎，{{ username }}</span>
            <el-button @click="$emit('logout')" text>退出</el-button>
          </div>
        </div>
      </el-header>
      
      <el-main class="main-content">
        <!-- 统计栏 -->
        <div class="stats-bar">
          <slot name="stats">
            <el-card v-for="stat in config.stats" :key="stat.key" shadow="hover">
              <div class="stat-item">
                <span class="stat-label">{{ stat.label }}</span>
                <span class="stat-value">{{ formatStatValue(stat) }}</span>
              </div>
            </el-card>
          </slot>
        </div>
        
        <!-- 操作栏 -->
        <div class="actions-bar">
          <div class="claim-section">
            <el-input-number
              v-model="localClaimCount"
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
              @click="$emit('claim', localClaimCount)"
            >
              {{ config.claimButtonText || '领取任务' }}
            </el-button>
          </div>
          
          <div class="return-section">
            <el-input-number
              v-model="localReturnCount"
              :min="1"
              :max="Math.max(1, taskCount)"
              :step="1"
              size="large"
              style="width: 120px"
              :disabled="taskCount === 0"
            />
            <el-button
              type="warning"
              size="large"
              :disabled="taskCount === 0"
              @click="$emit('return', localReturnCount)"
            >
              退单
            </el-button>
          </div>
          
          <el-button
            v-if="config.showBatchSubmit"
            size="large"
            :disabled="selectedCount === 0"
            @click="$emit('batch-submit')"
          >
            批量提交（{{ selectedCount }}条）
          </el-button>
          
          <el-button
            size="large"
            @click="$emit('refresh')"
          >
            刷新任务列表
          </el-button>
          
          <slot name="extra-actions"></slot>
        </div>
        
        <!-- 空状态 -->
        <div v-if="taskCount === 0" class="empty-state">
          <el-empty :description="config.emptyText || '暂无待审核任务，点击「领取任务」开始工作'" />
        </div>
        
        <!-- 任务列表 -->
        <div v-else class="tasks-container">
          <slot name="tasks"></slot>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'

export interface StatConfig {
  key: string
  label: string
  value?: number | string
  format?: (value: any) => string
}

export interface DashboardConfig {
  title: string
  showSearch?: boolean
  showBatchSubmit?: boolean
  claimButtonText?: string
  emptyText?: string
  stats?: StatConfig[]
}

const props = defineProps<{
  config: DashboardConfig
  username: string
  taskCount: number
  selectedCount?: number
  claimLoading?: boolean
  claimCount?: number
  returnCount?: number
  statsData?: Record<string, any>
}>()

const emit = defineEmits<{
  (e: 'claim', count: number): void
  (e: 'return', count: number): void
  (e: 'batch-submit'): void
  (e: 'refresh'): void
  (e: 'search'): void
  (e: 'logout'): void
  (e: 'update:claimCount', count: number): void
  (e: 'update:returnCount', count: number): void
}>()

const localClaimCount = ref(props.claimCount || 20)
const localReturnCount = ref(props.returnCount || 1)

watch(localClaimCount, (val) => emit('update:claimCount', val))
watch(localReturnCount, (val) => emit('update:returnCount', val))

watch(() => props.claimCount, (val) => {
  if (val !== undefined) localClaimCount.value = val
})

watch(() => props.returnCount, (val) => {
  if (val !== undefined) localReturnCount.value = val
})

const formatStatValue = (stat: StatConfig) => {
  const value = props.statsData?.[stat.key] ?? stat.value ?? 0
  if (stat.format) {
    return stat.format(value)
  }
  return value
}
</script>

<style scoped>
.generic-review-dashboard {
  min-height: 100vh;
  height: 100dvh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-100, #f5f5f5);
}

.header {
  background: var(--color-bg-000, #fff);
  box-shadow: var(--shadow-sm, 0 2px 4px rgba(0, 0, 0, 0.1));
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--color-border-lighter, #e4e7ed);
  flex-shrink: 0;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
}

.header-content h2 {
  margin: 0;
  font-size: 20px;
  color: var(--color-text-000, #333);
  font-weight: 600;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 14px;
  color: var(--color-text-200, #666);
}

.main-content {
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  padding: 32px;
  flex: 1;
  overflow-y: auto;
}

.stats-bar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 8px;
}

.stat-label {
  font-size: 14px;
  color: var(--color-text-400, #999);
  font-weight: 500;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--color-accent-main, #409eff);
  line-height: 1.2;
}

.actions-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
  padding: 20px;
  background: var(--color-bg-000, #fff);
  border-radius: 8px;
  border: 1px solid var(--color-border-lighter, #e4e7ed);
  box-shadow: var(--shadow-sm, 0 2px 4px rgba(0, 0, 0, 0.1));
}

.claim-section,
.return-section {
  display: flex;
  gap: 8px;
  align-items: center;
}

.empty-state {
  padding: 80px 0;
  display: flex;
  justify-content: center;
  align-items: center;
}

.tasks-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

@media (max-width: 768px) {
  .main-content {
    padding: 16px;
  }

  .stats-bar {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .actions-bar {
    flex-direction: column;
    gap: 8px;
    padding: 12px;
  }

  .claim-section,
  .return-section {
    width: 100%;
    flex-direction: column;
    gap: 8px;
  }

  .stat-value {
    font-size: 24px;
  }
}
</style>
