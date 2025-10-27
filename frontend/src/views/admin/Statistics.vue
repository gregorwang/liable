<template>
  <div class="admin-statistics-content">
          <!-- 违规类型分布 -->
          <el-card shadow="hover" style="margin-bottom: 24px">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center">
                <span style="font-weight: bold">违规类型分布</span>
                <el-button size="small" @click="loadTagStats">刷新</el-button>
              </div>
            </template>
            
            <el-table
              v-loading="tagStatsLoading"
              :data="tagStats"
              style="width: 100%"
            >
              <el-table-column prop="tag_name" label="标签名称" />
              <el-table-column prop="count" label="数量" width="120">
                <template #default="{ row }">
                  {{ formatNumber(row.count) }}
                </template>
              </el-table-column>
              <el-table-column prop="percentage" label="占比" width="200">
                <template #default="{ row }">
                  <el-progress
                    :percentage="Math.round(row.percentage * 100)"
                    :stroke-width="16"
                  />
                </template>
              </el-table-column>
            </el-table>
          </el-card>
          
          <!-- 审核员绩效排行 -->
          <el-card shadow="hover">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center">
                <span style="font-weight: bold">审核员绩效排行</span>
                <el-button size="small" @click="loadReviewerStats">刷新</el-button>
              </div>
            </template>
            
            <el-table
              v-loading="reviewerStatsLoading"
              :data="reviewerStats"
              style="width: 100%"
            >
              <el-table-column label="排名" width="80">
                <template #default="{ $index }">
                  <el-tag v-if="$index === 0" type="danger" effect="dark">
                    #{{ $index + 1 }}
                  </el-tag>
                  <el-tag v-else-if="$index === 1" type="warning" effect="dark">
                    #{{ $index + 1 }}
                  </el-tag>
                  <el-tag v-else-if="$index === 2" type="success" effect="dark">
                    #{{ $index + 1 }}
                  </el-tag>
                  <span v-else>#{{ $index + 1 }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="username" label="审核员" width="200" />
              <el-table-column prop="total_reviews" label="审核总数" width="120">
                <template #default="{ row }">
                  {{ formatNumber(row.total_reviews) }}
                </template>
              </el-table-column>
              <el-table-column prop="approved_count" label="通过数" width="120">
                <template #default="{ row }">
                  {{ formatNumber(row.approved_count) }}
                </template>
              </el-table-column>
              <el-table-column prop="rejected_count" label="不通过数" width="120">
                <template #default="{ row }">
                  {{ formatNumber(row.rejected_count) }}
                </template>
              </el-table-column>
              <el-table-column prop="approval_rate" label="通过率">
                <template #default="{ row }">
                  {{ formatPercent(row.approval_rate) }}
                </template>
              </el-table-column>
            </el-table>
          </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getTagStats, getReviewerPerformance } from '../../api/admin'
import type { TagStats, ReviewerPerformance } from '../../types'
import { formatNumber, formatPercent } from '../../utils/format'

const tagStatsLoading = ref(false)
const reviewerStatsLoading = ref(false)
const tagStats = ref<TagStats[]>([])
const reviewerStats = ref<ReviewerPerformance[]>([])


onMounted(() => {
  loadTagStats()
  loadReviewerStats()
})

const loadTagStats = async () => {
  tagStatsLoading.value = true
  try {
    const res = await getTagStats()
    tagStats.value = res.stats
  } catch (error) {
    console.error('Failed to load tag stats:', error)
  } finally {
    tagStatsLoading.value = false
  }
}

const loadReviewerStats = async () => {
  reviewerStatsLoading.value = true
  try {
    const res = await getReviewerPerformance(20)
    reviewerStats.value = res.reviewers
  } catch (error) {
    console.error('Failed to load reviewer stats:', error)
  } finally {
    reviewerStatsLoading.value = false
  }
}

</script>

<style scoped>
/* ============================================
   管理员统计分析页面样式
   ============================================ */
.admin-statistics-content {
  padding: var(--spacing-8);
  background-color: var(--color-bg-100);
  min-height: 100vh;
}

/* ============================================
   响应式设计
   ============================================ */
@media (max-width: 768px) {
  .main-content {
    padding: var(--spacing-4);
  }
}
</style>

