<template>
  <div class="admin-layout">
    <el-container>
      <el-aside width="200px" class="sidebar">
        <div class="logo">
          <h3>管理后台</h3>
        </div>
        <el-menu
          :default-active="currentRoute"
          router
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409eff"
        >
          <el-menu-item index="/admin/dashboard">
            <span>总览</span>
          </el-menu-item>
          <el-menu-item index="/admin/users">
            <span>用户管理</span>
          </el-menu-item>
          <el-menu-item index="/admin/statistics">
            <span>统计分析</span>
          </el-menu-item>
          <el-menu-item index="/admin/tags">
            <span>标签管理</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      
      <el-container>
        <el-header class="header">
          <div class="header-content">
            <h2>统计分析</h2>
            <div class="user-info">
              <span>{{ userStore.user?.username }}</span>
              <el-button @click="handleLogout" text>退出</el-button>
            </div>
          </div>
        </el-header>
        
        <el-main class="main-content">
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
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '../../stores/user'
import { getTagStats, getReviewerPerformance } from '../../api/admin'
import type { TagStats, ReviewerPerformance } from '../../types'
import { formatNumber, formatPercent } from '../../utils/format'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const tagStatsLoading = ref(false)
const reviewerStatsLoading = ref(false)
const tagStats = ref<TagStats[]>([])
const reviewerStats = ref<ReviewerPerformance[]>([])

const currentRoute = computed(() => route.path)

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

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确认退出登录？', '提示', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning',
    })
    userStore.logout()
    router.push('/login')
  } catch {
    // Cancel
  }
}
</script>

<style scoped>
/* ============================================
   布局结构
   ============================================ */
.admin-layout {
  height: 100vh;
  height: 100dvh;
  display: flex;
  overflow: hidden;
}

/* ============================================
   侧边栏样式
   ============================================ */
.sidebar {
  background: linear-gradient(180deg, 
    var(--color-accent-pro-dark) 0%, 
    hsl(251, 55%, 28%) 100%);
  overflow-x: hidden;
  overflow-y: auto;
  box-shadow: var(--shadow-lg);
  border-right: 1px solid var(--color-border);
}

.logo {
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  border-bottom: 1px solid rgba(255, 255, 255, 0.15);
  background: rgba(0, 0, 0, 0.1);
  padding: var(--spacing-4);
}

.logo h3 {
  margin: 0;
  font-size: var(--text-xl);
  font-weight: 600;
  letter-spacing: var(--tracking-wide);
}

/* ============================================
   头部样式
   ============================================ */
.header {
  background: var(--color-bg-000);
  box-shadow: var(--shadow-sm);
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--color-border-lighter);
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 var(--spacing-2);
}

.header-content h2 {
  margin: 0;
  font-size: var(--text-2xl);
  color: var(--color-text-000);
  font-weight: 600;
  letter-spacing: var(--tracking-tight);
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  font-size: var(--text-sm);
  color: var(--color-text-200);
  letter-spacing: var(--tracking-wide);
}

/* ============================================
   主内容区域
   ============================================ */
.main-content {
  background: var(--color-bg-100);
  padding: var(--spacing-8);
  overflow-y: auto;
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

