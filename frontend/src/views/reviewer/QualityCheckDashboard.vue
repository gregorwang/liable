<template>
  <div class="quality-check-dashboard">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <div class="header-left">
            <h2>质检工作台</h2>
            <el-button type="primary" link @click="goToSearch" style="margin-left: 20px">
              <el-icon><Search /></el-icon>
              搜索审核记录
            </el-button>
          </div>
          <div class="user-info">
            <span>欢迎，{{ userStore.user?.username }}</span>
            <el-button @click="handleLogout" text>退出</el-button>
          </div>
        </div>
      </el-header>
      
      <el-main class="main-content">
        <div class="stats-bar">
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">待质检任务</span>
              <span class="stat-value">{{ qcStats?.pending_tasks || 0 }}</span>
            </div>
          </el-card>
          
          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">今日已完成</span>
              <span class="stat-value">{{ qcStats?.today_completed || 0 }}</span>
            </div>
          </el-card>

          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">累计完成</span>
              <span class="stat-value">{{ qcStats?.total_completed || 0 }}</span>
            </div>
          </el-card>

          <el-card shadow="hover">
            <div class="stat-item">
              <span class="stat-label">质检通过率</span>
              <span class="stat-value">{{ qcStats?.pass_rate?.toFixed(1) || 0 }}%</span>
            </div>
          </el-card>
        </div>
        
        <div class="actions-bar">
          <div class="claim-section">
            <el-input-number
              v-model="claimCount"
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
              @click="handleClaimTasks"
            >
              领取质检任务
            </el-button>
          </div>
          
          <div class="return-section">
            <el-input-number
              v-model="returnCount"
              :min="1"
              :max="Math.max(1, qcTasks.length)"
              :step="1"
              size="large"
              style="width: 120px"
              :disabled="qcTasks.length === 0"
            />
            <el-button
              type="warning"
              size="large"
              :disabled="qcTasks.length === 0"
              @click="handleReturnTasks"
            >
              退单
            </el-button>
          </div>
          
          <el-button
            size="large"
            :disabled="selectedReviews.length === 0"
            @click="handleBatchSubmit"
          >
            批量提交（{{ selectedReviews.length }}条）
          </el-button>
          
          <el-button
            size="large"
            @click="handleRefresh"
          >
            刷新任务列表
          </el-button>
        </div>
        
        <div v-if="qcTasks.length === 0" class="empty-state">
          <el-empty description="暂无待质检任务，点击「领取质检任务」开始工作" />
        </div>
        
        <div v-else class="tasks-container">
          <el-card
            v-for="task in qcTasks"
            :key="task.id"
            class="task-card"
            shadow="hover"
          >
            <div v-if="reviews[task.id]" class="task-content">
              <!-- 原始评论内容 -->
              <div class="comment-section">
                <h4 class="section-title">原始评论</h4>
                <div class="comment-text">
                  {{ task.comment?.text || '评论内容加载中...' }}
                </div>
              </div>
              
              <el-divider />
              
              <!-- 一审审核结果 -->
              <div class="first-review-section">
                <h4 class="section-title">一审审核结果</h4>
                <div class="review-info">
                  <div class="review-result">
                    <span class="result-label">审核结果：</span>
                    <el-tag :type="task.first_review_result?.is_approved ? 'success' : 'danger'">
                      {{ task.first_review_result?.is_approved ? '通过' : '不通过' }}
                    </el-tag>
                  </div>
                  
                  <div v-if="task.first_review_result?.tags && task.first_review_result.tags.length > 0" class="review-tags">
                    <span class="result-label">违规标签：</span>
                    <el-tag 
                      v-for="tag in task.first_review_result.tags" 
                      :key="tag" 
                      size="small" 
                      style="margin-right: 8px"
                    >
                      {{ tag }}
                    </el-tag>
                  </div>
                  
                  <div v-if="task.first_review_result?.reason" class="review-reason">
                    <span class="result-label">审核意见：</span>
                    <span class="reason-text">{{ task.first_review_result.reason }}</span>
                  </div>
                  
                  <div v-if="task.first_review_result?.reviewer" class="reviewer-info">
                    <span class="result-label">审核员：</span>
                    <span>{{ task.first_review_result.reviewer.username }}</span>
                  </div>
                </div>
              </div>
              
              <el-divider />
              
              <!-- 质检操作表单 -->
              <div class="qc-section">
                <h4 class="section-title">质检操作</h4>
                <el-form label-position="top" size="default">
                  <el-form-item label="质检判断">
                    <el-radio-group v-model="getReview(task.id).is_passed">
                      <el-radio :value="true">
                        <span style="color: #67c23a">✅ 质检通过</span>
                      </el-radio>
                      <el-radio :value="false">
                        <span style="color: #f56c6c">❌ 质检不通过</span>
                      </el-radio>
                    </el-radio-group>
                  </el-form-item>
                  
                  <el-form-item
                    v-if="!getReview(task.id).is_passed"
                    label="错误类型"
                  >
                    <el-radio-group v-model="getReview(task.id).error_type">
                      <el-radio label="misjudgment">误判</el-radio>
                      <el-radio label="standard_deviation">标准执行偏差</el-radio>
                      <el-radio label="missing_violation">遗漏违规内容</el-radio>
                      <el-radio label="other">其他</el-radio>
                    </el-radio-group>
                  </el-form-item>
                  
                  <el-form-item
                    v-if="!getReview(task.id).is_passed"
                    label="质检意见"
                  >
                    <el-input
                      v-model="getReview(task.id).qc_comment"
                      type="textarea"
                      :rows="3"
                      placeholder="请详细说明问题所在"
                    />
                  </el-form-item>
                </el-form>
              </div>
              
              <div class="task-actions">
                <el-button
                  type="primary"
                  @click="handleSubmitSingle(task.id)"
                >
                  提交质检
                </el-button>
              </div>
            </div>
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useUserStore } from '../../stores/user'
import { 
  claimQCTasks, 
  getMyQCTasks, 
  submitQCReview, 
  submitBatchQCReviews, 
  returnQCTasks,
  getQCStats
} from '../../api/qualityCheck'
import type { QCStats, SubmitQCRequest } from '../../types'

const router = useRouter()
const userStore = useUserStore()

const claimLoading = ref(false)
const claimCount = ref(20)
const returnCount = ref(1)
const qcTasks = ref<any[]>([])
const qcStats = ref<QCStats | null>(null)
const reviews = reactive<Record<number, SubmitQCRequest>>({})

// Safe access to reviews
const getReview = (taskId: number) => {
  if (!reviews[taskId]) {
    reviews[taskId] = {
      task_id: taskId,
      is_passed: null as any,
      error_type: '',
      qc_comment: '',
    }
  }
  return reviews[taskId]
}

const selectedReviews = computed(() => {
  return Object.entries(reviews)
    .filter(([_, review]) => review.is_passed !== null)
    .map(([taskId, review]) => ({
      ...review,
      task_id: parseInt(taskId),
    }))
})

onMounted(async () => {
  try {
    await loadQCStats()
    await loadQCTasks()
    initReviews()
  } catch (error) {
    console.error('Failed to load data:', error)
  }
})

const loadQCStats = async () => {
  try {
    const response = await getQCStats()
    qcStats.value = response
  } catch (error) {
    console.error('Failed to load QC stats:', error)
  }
}

const loadQCTasks = async () => {
  try {
    const response = await getMyQCTasks()
    qcTasks.value = response.tasks
  } catch (error) {
    console.error('Failed to load QC tasks:', error)
  }
}

const initReviews = () => {
  // 清除不存在的任务的 review
  const taskIds = new Set(qcTasks.value.map(t => t.id))
  Object.keys(reviews).forEach(key => {
    if (!taskIds.has(parseInt(key))) {
      delete reviews[parseInt(key)]
    }
  })
  
  // 为新任务初始化 review
  qcTasks.value.forEach((task) => {
    if (!reviews[task.id]) {
      reviews[task.id] = {
        task_id: task.id,
        is_passed: null as any,
        error_type: undefined,
        qc_comment: '',
      }
    }
  })
  
  // 重置退单数量为1
  returnCount.value = Math.min(1, qcTasks.value.length)
}

const handleClaimTasks = async () => {
  if (claimCount.value < 1 || claimCount.value > 50) {
    ElMessage.warning('领取数量必须在 1-50 之间')
    return
  }
  
  claimLoading.value = true
  try {
    const res = await claimQCTasks(claimCount.value)
    ElMessage.success(`成功领取 ${res.count} 条质检任务`)
    await loadQCTasks()
    await loadQCStats()
    initReviews()
  } catch (error) {
    console.error('Failed to claim QC tasks:', error)
  } finally {
    claimLoading.value = false
  }
}

const handleRefresh = async () => {
  try {
    await loadQCTasks()
    await loadQCStats()
    initReviews()
    ElMessage.success('刷新成功')
  } catch (error) {
    console.error('Failed to refresh:', error)
  }
}

const validateReview = (review: SubmitQCRequest): boolean => {
  if (review.is_passed === null) {
    ElMessage.warning('请选择质检判断')
    return false
  }
  
  if (!review.is_passed) {
    if (!review.error_type) {
      ElMessage.warning('质检不通过时必须选择错误类型')
      return false
    }
    if (!review.qc_comment?.trim()) {
      ElMessage.warning('质检不通过时必须填写质检意见')
      return false
    }
  }
  
  return true
}

const handleSubmitSingle = async (taskId: number) => {
  const review = reviews[taskId]
  if (!review || !validateReview(review)) return
  
  try {
    await submitQCReview(review)
    ElMessage.success('提交成功')
    qcTasks.value = qcTasks.value.filter(t => t.id !== taskId)
    delete reviews[taskId]
    await loadQCStats()
  } catch (error) {
    console.error('Failed to submit QC review:', error)
  }
}

const handleBatchSubmit = async () => {
  const validReviews: SubmitQCRequest[] = []
  
  for (const review of selectedReviews.value) {
    if (!validateReview(review)) {
      return
    }
    validReviews.push(review)
  }
  
  if (validReviews.length === 0) {
    ElMessage.warning('没有可提交的质检')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确认提交 ${validReviews.length} 条质检结果？`,
      '批量提交',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    const res = await submitBatchQCReviews(validReviews)
    ElMessage.success(`成功提交 ${res.submitted} 条质检`)
    
    // Remove submitted tasks
    validReviews.forEach((review) => {
      qcTasks.value = qcTasks.value.filter(t => t.id !== review.task_id)
      delete reviews[review.task_id]
    })
    
    await loadQCStats()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to batch submit:', error)
    }
  }
}

const handleReturnTasks = async () => {
  if (returnCount.value < 1) {
    ElMessage.warning('退单数量至少为 1')
    return
  }

  if (returnCount.value > qcTasks.value.length) {
    ElMessage.warning(`退单数量不能超过当前任务数 (${qcTasks.value.length})`)
    return
  }

  if (returnCount.value > 50) {
    ElMessage.warning('退单数量不能超过 50 条')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确认退回 ${returnCount.value} 条质检任务？将退回最早领取的任务。`,
      '退单确认',
      {
        confirmButtonText: '确认退单',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    // 取最早领取的N个任务ID
    const taskIdsToReturn = qcTasks.value
      .slice(0, returnCount.value)
      .map(task => task.id)

    const res = await returnQCTasks(taskIdsToReturn)
    ElMessage.success(`成功退回 ${res.count} 条质检任务`)

    // Remove returned tasks from local state
    taskIdsToReturn.forEach((taskId) => {
      qcTasks.value = qcTasks.value.filter(t => t.id !== taskId)
      delete reviews[taskId]
    })
    
    // 刷新任务列表
    await loadQCTasks()
    await loadQCStats()
    initReviews()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to return QC tasks:', error)
    }
  }
}

const goToSearch = () => {
  router.push('/reviewer/search')
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
   质检工作台样式
   ============================================ */
.quality-check-dashboard {
  min-height: 100vh;
  height: 100dvh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-100);
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
  flex-shrink: 0;
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

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
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
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  padding: var(--spacing-8);
  flex: 1;
  overflow-y: auto;
}

/* ============================================
   统计栏
   ============================================ */
.stats-bar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-5);
  margin-bottom: var(--spacing-6);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
  padding: var(--spacing-2);
}

.stat-label {
  font-size: var(--text-sm);
  color: var(--color-text-400);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
}

.stat-value {
  font-size: var(--text-4xl);
  font-weight: 700;
  color: var(--color-accent-main);
  line-height: var(--leading-tight);
  letter-spacing: var(--tracking-tight);
}

/* ============================================
   操作栏
   ============================================ */
.actions-bar {
  display: flex;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-6);
  flex-wrap: wrap;
  padding: var(--spacing-5);
  background: var(--color-bg-000);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-lighter);
  box-shadow: var(--shadow-sm);
}

.claim-section,
.return-section {
  display: flex;
  gap: var(--spacing-2);
  align-items: center;
}

/* ============================================
   空状态
   ============================================ */
.empty-state {
  padding: var(--spacing-20) 0;
  display: flex;
  justify-content: center;
  align-items: center;
}

/* ============================================
   任务卡片容器
   ============================================ */
.tasks-container {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-5);
}

.task-card {
  transition: all var(--transition-base);
  border: 1px solid var(--color-border-lighter);
}

.task-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-lg);
  border-color: var(--color-accent-main);
}

/* ============================================
   任务内容区域
   ============================================ */
.task-content {
  padding: var(--spacing-4);
}

.section-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-000);
  margin: 0 0 var(--spacing-3) 0;
  letter-spacing: var(--tracking-tight);
}

/* ============================================
   评论区域
   ============================================ */
.comment-section {
  margin-bottom: var(--spacing-4);
}

.comment-text {
  font-family: var(--font-serif);
  font-size: var(--text-base);
  line-height: var(--leading-loose);
  color: var(--color-text-100);
  padding: var(--spacing-5);
  background: var(--color-bg-200);
  border-radius: var(--radius-md);
  letter-spacing: var(--tracking-wide);
  border-left: 4px solid var(--color-accent-main);
  word-break: break-word;
  white-space: pre-wrap;
}

/* ============================================
   一审审核结果区域
   ============================================ */
.first-review-section {
  margin-bottom: var(--spacing-4);
}

.review-info {
  background: var(--color-bg-200);
  padding: var(--spacing-4);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border-lighter);
}

.review-result,
.review-tags,
.review-reason,
.reviewer-info {
  margin-bottom: var(--spacing-3);
}

.review-result:last-child,
.review-tags:last-child,
.review-reason:last-child,
.reviewer-info:last-child {
  margin-bottom: 0;
}

.result-label {
  font-weight: 600;
  color: var(--color-text-200);
  margin-right: var(--spacing-2);
}

.reason-text {
  color: var(--color-text-100);
  line-height: var(--leading-relaxed);
}

/* ============================================
   质检操作区域
   ============================================ */
.qc-section {
  margin-bottom: var(--spacing-4);
}

/* ============================================
   任务操作区域
   ============================================ */
.task-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-5);
  padding-top: var(--spacing-4);
  border-top: 1px solid var(--color-border-lighter);
}

/* ============================================
   响应式设计
   ============================================ */
@media (max-width: 768px) {
  .main-content {
    padding: var(--spacing-4);
  }

  .stats-bar {
    grid-template-columns: 1fr;
    gap: var(--spacing-3);
  }

  .actions-bar {
    flex-direction: column;
    gap: var(--spacing-2);
    padding: var(--spacing-3);
  }

  .claim-section,
  .return-section {
    width: 100%;
    flex-direction: column;
    gap: var(--spacing-2);
  }

  .stat-value {
    font-size: var(--text-3xl);
  }

  .comment-text {
    font-size: var(--text-sm);
    padding: var(--spacing-3);
  }
}

@media (max-width: 1024px) {
  .main-content {
    padding: var(--spacing-6);
  }
}
</style>
