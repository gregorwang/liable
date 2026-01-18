<template>
  <div class="admin-statistics-content">
    <!-- 概览统计卡片 -->
    <el-row :gutter="24" style="margin-bottom: 24px">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stats-card">
          <div class="stats-card-content">
            <div class="stats-icon comment-icon">
              <el-icon><ChatLineSquare /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-label">评论审核</div>
              <div class="stats-value">{{ formatNumber(overviewStats?.comment_review_stats?.first_review?.total_tasks || 0) }}</div>
              <div class="stats-sub">完成 {{ formatNumber(overviewStats?.comment_review_stats?.first_review?.completed_tasks || 0) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stats-card">
          <div class="stats-card-content">
            <div class="stats-icon video-icon">
              <el-icon><VideoPlay /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-label">视频审核</div>
              <div class="stats-value">{{ formatNumber(overviewStats?.video_review_stats?.first_review?.total_tasks || 0) }}</div>
              <div class="stats-sub">完成 {{ formatNumber(overviewStats?.video_review_stats?.first_review?.completed_tasks || 0) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stats-card">
          <div class="stats-card-content">
            <div class="stats-icon reviewer-icon">
              <el-icon><User /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-label">审核员</div>
              <div class="stats-value">{{ formatNumber(overviewStats?.total_reviewers || 0) }}</div>
              <div class="stats-sub">活跃 {{ formatNumber(overviewStats?.active_reviewers || 0) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stats-card">
          <div class="stats-card-content">
            <div class="stats-icon quality-icon">
              <el-icon><Check /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-label">质检任务</div>
              <div class="stats-value">{{ formatNumber(overviewStats?.quality_metrics?.total_quality_checks || 0) }}</div>
              <div class="stats-sub">通过率 {{ formatPercent(overviewStats?.quality_metrics?.quality_pass_rate || 0) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 详细统计对比 -->
    <el-row :gutter="24" style="margin-bottom: 24px">
      <el-col :xs="24" :md="12">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span style="font-weight: bold">评论审核统计</span>
              <el-tag type="primary">Comment Review</el-tag>
            </div>
          </template>

          <div class="review-stats-grid">
            <div class="review-stat-item">
              <div class="stat-label">一审任务</div>
              <div class="stat-value">{{ formatNumber(overviewStats?.comment_review_stats?.first_review?.total_tasks || 0) }}</div>
              <el-progress
                :percentage="getCompletionPercentage(overviewStats?.comment_review_stats?.first_review)"
                :stroke-width="8"
                :color="progressColor"
              />
            </div>

            <div class="review-stat-item">
              <div class="stat-label">二审任务</div>
              <div class="stat-value">{{ formatNumber(overviewStats?.comment_review_stats?.second_review?.total_tasks || 0) }}</div>
              <el-progress
                :percentage="getCompletionPercentage(overviewStats?.comment_review_stats?.second_review)"
                :stroke-width="8"
                :color="progressColor"
              />
            </div>

            <div class="review-stat-item">
              <div class="stat-label">通过率</div>
              <div class="stat-value">{{ formatPercent(overviewStats?.comment_review_stats?.first_review?.approval_rate || 0) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :md="12">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span style="font-weight: bold">视频审核统计</span>
              <el-tag type="success">Video Review</el-tag>
            </div>
          </template>

          <div class="review-stats-grid">
            <div class="review-stat-item">
              <div class="stat-label">一审任务</div>
              <div class="stat-value">{{ formatNumber(overviewStats?.video_review_stats?.first_review?.total_tasks || 0) }}</div>
              <el-progress
                :percentage="getCompletionPercentage(overviewStats?.video_review_stats?.first_review)"
                :stroke-width="8"
                :color="progressColor"
              />
            </div>

            <div class="review-stat-item">
              <div class="stat-label">二审任务</div>
              <div class="stat-value">{{ formatNumber(overviewStats?.video_review_stats?.second_review?.total_tasks || 0) }}</div>
              <el-progress
                :percentage="getCompletionPercentage(overviewStats?.video_review_stats?.second_review)"
                :stroke-width="8"
                :color="progressColor"
              />
            </div>

            <div class="review-stat-item">
              <div class="stat-label">平均质量分</div>
              <div class="stat-value">{{ overviewStats?.video_review_stats?.first_review?.avg_overall_score?.toFixed(1) || '0.0' }}/40</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 队列状态总览 -->
    <el-card shadow="hover" style="margin-bottom: 24px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span style="font-weight: bold">队列状态总览</span>
          <el-button size="small" @click="loadOverviewStats">刷新</el-button>
        </div>
      </template>

      <el-table
        v-loading="overviewLoading"
        :data="overviewStats?.queue_stats || []"
        style="width: 100%"
      >
        <el-table-column prop="queue_name" label="队列名称" width="200">
          <template #default="{ row }">
            <el-tag v-if="row.queue_name.includes('video')" type="success">
              {{ getQueueDisplayName(row.queue_name) }}
            </el-tag>
            <el-tag v-else-if="row.queue_name.includes('quality')" type="warning">
              {{ getQueueDisplayName(row.queue_name) }}
            </el-tag>
            <el-tag v-else type="primary">
              {{ getQueueDisplayName(row.queue_name) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="total_tasks" label="总任务数" width="120">
          <template #default="{ row }">
            {{ formatNumber(row.total_tasks) }}
          </template>
        </el-table-column>

        <el-table-column label="进度" width="250">
          <template #default="{ row }">
            <el-progress
              :percentage="Math.round((row.completed_tasks / row.total_tasks) * 100) || 0"
              :stroke-width="18"
              :color="progressColor"
            >
              <span class="progress-text">{{ row.completed_tasks }}/{{ row.total_tasks }}</span>
            </el-progress>
          </template>
        </el-table-column>

        <el-table-column prop="pending_tasks" label="待审" width="100">
          <template #default="{ row }">
            <el-tag type="warning">{{ formatNumber(row.pending_tasks) }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="approval_rate" label="通过率" width="120">
          <template #default="{ row }">
            {{ formatPercent(row.approval_rate) }}
          </template>
        </el-table-column>

        <el-table-column prop="is_active" label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.is_active" type="success">活跃</el-tag>
            <el-tag v-else type="info">已停用</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 审核员绩效排行 -->
    <el-card shadow="hover" style="margin-bottom: 24px">
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

        <el-table-column prop="username" label="审核员" width="150" />

        <el-table-column prop="total_reviews" label="审核总数" width="120">
          <template #default="{ row }">
            <strong>{{ formatNumber(row.total_reviews) }}</strong>
          </template>
        </el-table-column>

        <el-table-column label="审核类型分布" min-width="300">
          <template #default="{ row }">
            <div class="review-type-breakdown">
              <el-tooltip content="评论一审" placement="top">
                <el-tag size="small" type="primary">评论1审: {{ row.comment_first_reviews }}</el-tag>
              </el-tooltip>
              <el-tooltip content="评论二审" placement="top">
                <el-tag size="small" type="primary">评论2审: {{ row.comment_second_reviews }}</el-tag>
              </el-tooltip>
              <el-tooltip content="质检" placement="top">
                <el-tag size="small" type="warning">质检: {{ row.quality_checks }}</el-tag>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="approval_rate" label="通过率" width="120">
          <template #default="{ row }">
            {{ formatPercent(row.approval_rate) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 违规类型分布 -->
    <el-card shadow="hover">
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ChatLineSquare, VideoPlay, User, Check } from '@element-plus/icons-vue'
import { getOverviewStats, getTagStats, getReviewerPerformance } from '../../api/admin'
import type { OverviewStats, TagStats, ReviewerPerformance, ReviewStats, VideoReviewStatsDetail } from '../../types'
import { formatNumber, formatPercent } from '../../utils/format'

const overviewLoading = ref(false)
const tagStatsLoading = ref(false)
const reviewerStatsLoading = ref(false)

const overviewStats = ref<OverviewStats | null>(null)
const tagStats = ref<TagStats[]>([])
const reviewerStats = ref<ReviewerPerformance[]>([])

const progressColor = '#67C23A'

onMounted(() => {
  loadOverviewStats()
  loadTagStats()
  loadReviewerStats()
})

const loadOverviewStats = async () => {
  overviewLoading.value = true
  try {
    const res = await getOverviewStats()
    overviewStats.value = res
  } catch (error) {
    console.error('Failed to load overview stats:', error)
  } finally {
    overviewLoading.value = false
  }
}

const loadTagStats = async () => {
  tagStatsLoading.value = true
  try {
    const res = await getTagStats()
    tagStats.value = res.tags
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

const getCompletionPercentage = (stats: ReviewStats | VideoReviewStatsDetail | undefined): number => {
  if (!stats || stats.total_tasks === 0) return 0
  return Math.round((stats.completed_tasks / stats.total_tasks) * 100)
}

const getQueueDisplayName = (queueName: string): string => {
  const nameMap: Record<string, string> = {
    'comment_first_review': '评论一审',
    'comment_second_review': '评论二审',
    'quality_check': '质量检查',
    '100k': '100k流量池',
    '1m': '1m流量池',
    '10m': '10m流量池',
    'video_100k': '100k流量池审核',
    'video_1m': '1m流量池审核',
    'video_10m': '10m流量池审核'
  }
  return nameMap[queueName] || queueName
}
</script>

<style scoped>
.admin-statistics-content {
  padding: var(--spacing-8);
  background-color: var(--color-bg-100);
  min-height: 100vh;
}

/* 统计卡片样式 */
.stats-card {
  height: 100%;
}

.stats-card-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stats-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: white;
}

.comment-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.video-icon {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.reviewer-icon {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.quality-icon {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stats-info {
  flex: 1;
}

.stats-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.stats-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
  margin-bottom: 4px;
}

.stats-sub {
  font-size: 12px;
  color: #909399;
}

/* 审核统计网格 */
.review-stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.review-stat-item {
  padding: 16px;
  background-color: #f5f7fa;
  border-radius: 8px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 12px;
}

/* 审核类型分布 */
.review-type-breakdown {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.progress-text {
  font-size: 12px;
  font-weight: bold;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .admin-statistics-content {
    padding: var(--spacing-4);
  }

  .stats-card-content {
    flex-direction: column;
    text-align: center;
  }

  .review-stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
