<template>
  <div class="moderation-rules-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span class="title">
            <i class="el-icon-document"></i> 审核规则库
          </span>
          <div class="header-actions">
            <el-tag effect="dark">共 {{ total }} 条规则</el-tag>
            <el-button type="primary" @click="openAddDialog">
              <i class="el-icon-plus"></i> 新增规则
            </el-button>
            <el-button @click="refreshRules">
              <i class="el-icon-refresh"></i> 刷新
            </el-button>
          </div>
        </div>
      </template>

      <!-- Search and Filter Bar -->
      <div class="search-bar">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8">
            <el-input
              v-model="searchText"
              placeholder="搜索规则编号或描述..."
              clearable
              @input="handleSearch"
            >
              <template #prefix>
                <i class="el-icon-search"></i>
              </template>
            </el-input>
          </el-col>

          <el-col :xs="24" :sm="12" :md="8">
            <el-select
              v-model="selectedCategory"
              placeholder="选择分类"
              clearable
              @change="handleFilterChange"
            >
              <el-option
                v-for="cat in categories"
                :key="cat"
                :label="cat"
                :value="cat"
              ></el-option>
            </el-select>
          </el-col>

          <el-col :xs="24" :sm="12" :md="8">
            <el-select
              v-model="selectedRiskLevel"
              placeholder="选择风险等级"
              clearable
              @change="handleFilterChange"
            >
              <el-option label="低风险 (L)" value="L"></el-option>
              <el-option label="中风险 (M)" value="M"></el-option>
              <el-option label="高风险 (H)" value="H"></el-option>
              <el-option label="极高风险 (C)" value="C"></el-option>
            </el-select>
          </el-col>
        </el-row>
      </div>

      <!-- Rules Table -->
      <el-table
        :data="filteredRules"
        stripe
        style="width: 100%; margin-top: 20px"
        v-loading="loading"
        @expand-change="handleExpandChange"
      >
        <!-- Expansion (Details) -->
        <el-table-column type="expand" width="50">
          <template #default="props">
            <div class="expand-details">
              <el-row :gutter="20">
                <el-col :md="12">
                  <div class="detail-section">
                    <h4>判定要点</h4>
                    <p>{{ props.row.judgment_criteria || '暂无信息' }}</p>
                  </div>
                </el-col>
                <el-col :md="12">
                  <div class="detail-section">
                    <h4>边界与记录要点</h4>
                    <p>{{ props.row.boundary || '暂无信息' }}</p>
                  </div>
                </el-col>
              </el-row>
              <el-row :gutter="20" style="margin-top: 15px">
                <el-col :md="12">
                  <div class="detail-section">
                    <h4>处置动作</h4>
                    <p>{{ props.row.action || '暂无信息' }}</p>
                  </div>
                </el-col>
                <el-col :md="12">
                  <div class="detail-section">
                    <h4>示例</h4>
                    <p>{{ props.row.examples || '暂无示例' }}</p>
                  </div>
                </el-col>
              </el-row>
            </div>
          </template>
        </el-table-column>

        <!-- Rule Code -->
        <el-table-column prop="rule_code" label="规则编号" width="100" sortable>
          <template #default="{ row }">
            <el-tag class="rule-code-tag">{{ row.rule_code }}</el-tag>
          </template>
        </el-table-column>

        <!-- Category -->
        <el-table-column prop="category" label="分类" width="150" show-overflow-tooltip></el-table-column>

        <!-- Subcategory -->
        <el-table-column prop="subcategory" label="二级标签" width="180" show-overflow-tooltip></el-table-column>

        <!-- Description -->
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip></el-table-column>

        <!-- Risk Level with Color -->
        <el-table-column prop="risk_level" label="风险等级" width="100" sortable align="center">
          <template #default="{ row }">
            <el-tag :type="getRiskLevelType(row.risk_level)" effect="dark">
              {{ getRiskLevelLabel(row.risk_level) }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- Quick Tag (if available) -->
        <el-table-column prop="quick_tag" label="快捷标签" width="120" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.quick_tag" type="info" effect="light">
              {{ row.quick_tag }}
            </el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>

        <!-- Updated Time -->
        <el-table-column prop="updated_at" label="更新时间" width="180" sortable>
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>

        <!-- Actions -->
        <el-table-column label="操作" width="140" fixed="right" align="center">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEditDialog(row)">
              编辑
            </el-button>
            <el-divider direction="vertical" />
            <el-popconfirm
              title="确认删除此规则吗？"
              description="此操作不可撤销"
              confirm-button-text="确认"
              cancel-button-text="取消"
              @confirm="deleteRule(row.id)"
            >
              <template #reference>
                <el-button link type="danger" size="small">
                  删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="filteredTotal"
        layout="total, sizes, prev, pager, next, jumper"
        style="margin-top: 20px; text-align: right"
        @size-change="handlePageSizeChange"
        @current-change="handlePageChange"
      ></el-pagination>
    </el-card>

    <!-- Statistics Cards -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :xs="12" :sm="6">
        <el-statistic
          title="总规则数"
          :value="total"
          style="text-align: center"
        ></el-statistic>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-statistic
          title="极高风险 (C)"
          :value="riskStats.C"
          :class="{ 'risk-c': riskStats.C > 0 }"
          style="text-align: center"
        ></el-statistic>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-statistic
          title="高风险 (H)"
          :value="riskStats.H"
          :class="{ 'risk-h': riskStats.H > 0 }"
          style="text-align: center"
        ></el-statistic>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-statistic
          title="中风险 (M)"
          :value="riskStats.M"
          style="text-align: center"
        ></el-statistic>
      </el-col>
    </el-row>

    <!-- Rule Dialog for Add/Edit -->
    <RuleDialog
      v-model:visible="ruleDialogVisible"
      :editing-rule="editingRule"
      @success="handleDialogSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'
import RuleDialog from '@/components/RuleDialog.vue'
import type { ModerationRule } from '@/types'
import * as moderationApi from '@/api/moderation'

const allRules = ref<ModerationRule[]>([])
const categories = ref<string[]>([])
const loading = ref(false)
const searchText = ref('')
const selectedCategory = ref('')
const selectedRiskLevel = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const ruleDialogVisible = ref(false)
const editingRule = ref<ModerationRule | null>(null)

const riskStats = computed(() => {
  const stats = { L: 0, M: 0, H: 0, C: 0 }
  allRules.value.forEach((rule) => {
    if (rule.risk_level in stats) {
      stats[rule.risk_level as 'L' | 'M' | 'H' | 'C']++
    }
  })
  return stats
})

// Client-side filtering of all rules
const filteredRules = computed(() => {
  let filtered = allRules.value

  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    filtered = filtered.filter((rule) =>
      rule.rule_code.toLowerCase().includes(search) ||
      rule.description.toLowerCase().includes(search)
    )
  }

  if (selectedCategory.value) {
    filtered = filtered.filter((rule) => rule.category === selectedCategory.value)
  }

  if (selectedRiskLevel.value) {
    filtered = filtered.filter((rule) => rule.risk_level === selectedRiskLevel.value)
  }

  return filtered
})

const filteredTotal = computed(() => filteredRules.value.length)

const getRiskLevelType = (level: string): string => {
  const typeMap: Record<string, string> = {
    L: 'success',
    M: 'warning',
    H: 'danger',
    C: 'danger'
  }
  return typeMap[level] || 'info'
}

const getRiskLevelLabel = (level: string): string => {
  const labelMap: Record<string, string> = {
    L: '低风险',
    M: '中风险',
    H: '高风险',
    C: '极高风险'
  }
  return labelMap[level] || level
}

const formatDateTime = (dateTime: string): string => {
  try {
    const date = new Date(dateTime)
    return date.toLocaleString('zh-CN')
  } catch {
    return dateTime
  }
}

const fetchAllRules = async (useCache: boolean = true) => {
  loading.value = true
  try {
    // Check cache first (30 minute expiration)
    const cacheKey = 'moderation_rules_cache'
    const cacheTimestampKey = 'moderation_rules_cache_time'
    const now = Date.now()
    const thirtyMinutes = 30 * 60 * 1000

    if (useCache) {
      const cachedData = localStorage.getItem(cacheKey)
      const cachedTime = localStorage.getItem(cacheTimestampKey)

      if (cachedData && cachedTime) {
        const cacheAge = now - parseInt(cachedTime)
        if (cacheAge < thirtyMinutes) {
          // Cache is still valid
          try {
            const parsed = JSON.parse(cachedData)
            allRules.value = parsed.data || []
            total.value = parsed.total || 0
            currentPage.value = 1
            // Loaded rules from cache
            return
          } catch (e) {
            console.warn('Failed to parse cached rules:', e)
          }
        }
      }
    }

    // Fetch from API using the new getAllRules endpoint
    const response = await moderationApi.getAllRules()

    allRules.value = response.data || []
    total.value = response.total || 0
    currentPage.value = 1

    // Store in cache
    localStorage.setItem(cacheKey, JSON.stringify(response))
    localStorage.setItem(cacheTimestampKey, now.toString())

    // Fetched and cached rules from API
  } catch (error) {
    ElMessage.error('加载规则失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchCategories = async () => {
  try {
    const response = await request.get<{ categories: string[] }>(
      '/moderation-rules/categories'
    )
    categories.value = response.data.categories || []
    // Loaded categories
  } catch (error) {
    console.error('Failed to fetch categories:', error)
  }
}

const handleSearch = () => {
  currentPage.value = 1
}

const handleFilterChange = () => {
  currentPage.value = 1
}

const handlePageChange = (page: number) => {
  currentPage.value = page
}

const handlePageSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
}

const handleExpandChange = (_row: ModerationRule, _expandedRows: ModerationRule[]) => {
  // Row details are displayed via expansion panel
}

const openAddDialog = () => {
  editingRule.value = null
  ruleDialogVisible.value = true
}

const openEditDialog = (rule: ModerationRule) => {
  editingRule.value = rule
  ruleDialogVisible.value = true
}

const deleteRule = async (id: number) => {
  try {
    await moderationApi.deleteRule(id)
    ElMessage.success('规则删除成功')
    await fetchAllRules()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '删除失败')
    console.error(error)
  }
}

const handleDialogSuccess = async () => {
  await fetchAllRules()
}

const refreshRules = () => {
  ElMessage.success('正在刷新规则库...')
  localStorage.removeItem('moderation_rules_cache')
  localStorage.removeItem('moderation_rules_cache_time')
  fetchAllRules(false) // Force fetch from API
}

onMounted(() => {
  fetchCategories()
  fetchAllRules()
})
</script>

<style scoped lang="scss">
/* ============================================
   审核规则库页面样式
   ============================================ */
.moderation-rules-container {
  padding: var(--spacing-8);
  background-color: var(--color-bg-100);
  min-height: 100vh;

  /* ============================================
     卡片头部
     ============================================ */
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    gap: var(--spacing-4);

    .title {
      font-size: var(--text-xl);
      font-weight: 600;
      color: var(--color-text-000);
      letter-spacing: var(--tracking-tight);

      i {
        margin-right: var(--spacing-2);
        color: var(--color-accent-main);
      }
    }

    .header-actions {
      display: flex;
      gap: var(--spacing-4);
      align-items: center;
      flex-wrap: wrap;
    }
  }

  /* ============================================
     搜索栏
     ============================================ */
  .search-bar {
    margin-bottom: var(--spacing-6);

    :deep(.el-select) {
      width: 100%;
    }

    :deep(.el-input__wrapper) {
      width: 100%;
    }
  }

  /* ============================================
     展开详情区域
     ============================================ */
  .expand-details {
    padding: var(--spacing-6);
    background-color: var(--color-bg-200);
    border-radius: var(--radius-lg);
    border: 1px solid var(--color-border-lighter);
    margin: var(--spacing-4) 0;

    .detail-section {
      margin-bottom: var(--spacing-5);

      &:last-child {
        margin-bottom: 0;
      }

      h4 {
        margin: 0 0 var(--spacing-3) 0;
        font-size: var(--text-base);
        font-weight: 600;
        color: var(--color-text-100);
        letter-spacing: var(--tracking-wide);
      }

      p {
        margin: 0;
        color: var(--color-text-200);
        line-height: var(--leading-relaxed);
        white-space: pre-wrap;
        word-break: break-word;
        font-size: var(--text-sm);
        letter-spacing: var(--tracking-normal);
        padding: var(--spacing-3);
        background-color: var(--color-bg-000);
        border-radius: var(--radius-md);
        border-left: 3px solid var(--color-accent-main);
      }
    }
  }

  /* ============================================
     规则编号标签
     ============================================ */
  .rule-code-tag {
    font-weight: 600;
    letter-spacing: var(--tracking-wider);
    font-family: var(--font-mono);
  }

  /* ============================================
     弱化文本
     ============================================ */
  .text-muted {
    color: var(--color-text-400);
    font-size: var(--text-sm);
  }

  /* ============================================
     表格样式增强
     ============================================ */
  :deep(.el-table) {
    .el-table__row {
      transition: background-color var(--transition-fast);

      &:hover {
        background-color: var(--color-bg-200);
      }
    }

    .el-table__cell {
      padding: var(--spacing-4) var(--spacing-3);
    }

    .el-table__header {
      th {
        font-size: var(--text-sm);
        font-weight: 600;
        letter-spacing: var(--tracking-wide);
      }
    }

    .el-table__body {
      td {
        font-size: var(--text-sm);
        line-height: var(--leading-relaxed);
      }
    }
  }

  /* ============================================
     统计数字样式
     ============================================ */
  :deep(.el-statistic) {
    .el-statistic__content {
      font-weight: 600;
    }

    &.risk-c {
      color: var(--color-danger);
    }

    &.risk-h {
      color: var(--color-warning);
    }
  }

  /* ============================================
     分页样式
     ============================================ */
  :deep(.el-pagination) {
    margin-top: var(--spacing-6);
    padding: var(--spacing-4) 0;
    display: flex;
    justify-content: center;
  }
}

/* ============================================
   响应式设计
   ============================================ */
@media (max-width: 768px) {
  .moderation-rules-container {
    padding: var(--spacing-4);

    .card-header {
      flex-direction: column;
      gap: var(--spacing-3);
      align-items: flex-start;

      .header-actions {
        width: 100%;
        flex-direction: column;
        gap: var(--spacing-2);
      }
    }

    :deep(.el-table) {
      font-size: var(--text-xs);

      .el-table__cell {
        padding: var(--spacing-2);
      }
    }

    .expand-details {
      padding: var(--spacing-4);

      .detail-section {
        margin-bottom: var(--spacing-3);

        h4 {
          font-size: var(--text-sm);
        }

        p {
          font-size: var(--text-xs);
          padding: var(--spacing-2);
        }
      }
    }
  }
}

@media (max-width: 1024px) {
  .moderation-rules-container {
    padding: var(--spacing-6);
  }
}
</style>
