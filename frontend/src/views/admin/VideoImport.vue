<template>
  <div class="video-import-page">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <div class="header-left">
            <h2>视频导入管理</h2>
            <el-button type="primary" link @click="goToVideoList" style="margin-left: 20px">
              <el-icon><List /></el-icon>
              查看视频列表
            </el-button>
          </div>
          <div class="user-info">
            <span>欢迎，{{ userStore.user?.username }}</span>
            <el-button @click="handleLogout" text>退出</el-button>
          </div>
        </div>
      </el-header>
      
      <el-main class="main-content">
        <!-- Import Section -->
        <el-card class="import-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>从 R2 存储桶导入视频</span>
            </div>
          </template>
          
          <el-form :model="importForm" label-width="120px" size="large">
            <el-form-item label="R2 路径前缀">
              <el-input
                v-model="importForm.r2_path_prefix"
                placeholder="例如: douyin/PostmanAgent/"
                clearable
              >
                <template #prepend>路径</template>
              </el-input>
              <div class="form-help">
                指定 R2 存储桶中的视频文件路径前缀，系统将扫描该路径下的所有视频文件
              </div>
            </el-form-item>
            
            <el-form-item>
              <el-button
                type="primary"
                size="large"
                :loading="importing"
                @click="handleImportVideos"
                :disabled="!importForm.r2_path_prefix.trim()"
              >
                <el-icon><Upload /></el-icon>
                开始导入
              </el-button>
              <el-button
                size="large"
                @click="resetForm"
              >
                重置
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- Import Progress -->
        <el-card v-if="importResult" class="result-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>导入结果</span>
              <el-tag :type="importResult.errors.length > 0 ? 'warning' : 'success'">
                {{ importResult.errors.length > 0 ? '部分成功' : '导入成功' }}
              </el-tag>
            </div>
          </template>
          
          <div class="result-summary">
            <div class="summary-item">
              <span class="label">成功导入:</span>
              <span class="value success">{{ importResult.imported_count }} 个视频</span>
            </div>
            <div class="summary-item">
              <span class="label">跳过文件:</span>
              <span class="value info">{{ importResult.skipped_count }} 个文件</span>
            </div>
            <div class="summary-item" v-if="importResult.errors.length > 0">
              <span class="label">错误数量:</span>
              <span class="value error">{{ importResult.errors.length }} 个错误</span>
            </div>
          </div>
          
          <div v-if="importResult.errors.length > 0" class="error-list">
            <h4>错误详情:</h4>
            <el-alert
              v-for="(error, index) in importResult.errors"
              :key="index"
              :title="error"
              type="error"
              :closable="false"
              style="margin-bottom: 8px"
            />
          </div>
        </el-card>

        <!-- Video List -->
        <el-card class="video-list-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>已导入的视频</span>
              <div class="header-actions">
                <el-button @click="loadVideos" :loading="loadingVideos">
                  <el-icon><Refresh /></el-icon>
                  刷新
                </el-button>
              </div>
            </div>
          </template>
          
          <div class="video-filters">
            <el-input
              v-model="searchQuery"
              placeholder="搜索文件名"
              clearable
              style="width: 300px"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            
            <el-select
              v-model="statusFilter"
              placeholder="筛选状态"
              clearable
              style="width: 150px"
            >
              <el-option label="待审核" value="pending" />
              <el-option label="一审完成" value="first_review_completed" />
              <el-option label="二审完成" value="second_review_completed" />
            </el-select>
            
            <el-button @click="applyFilters" type="primary">
              筛选
            </el-button>
          </div>
          
          <el-table
            :data="filteredVideos"
            v-loading="loadingVideos"
            style="width: 100%"
            max-height="400"
          >
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="filename" label="文件名" min-width="200" />
            <el-table-column prop="file_size" label="文件大小" width="120">
              <template #default="{ row }">
                {{ formatFileSize(row.file_size) }}
              </template>
            </el-table-column>
            <el-table-column prop="duration" label="时长" width="100">
              <template #default="{ row }">
                {{ row.duration ? formatDuration(row.duration) : '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="导入时间" width="160">
              <template #default="{ row }">
                {{ formatTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button
                  type="primary"
                  size="small"
                  @click="previewVideo(row)"
                >
                  预览
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <div class="pagination-container">
            <el-pagination
              v-model:current-page="currentPage"
              v-model:page-size="pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="totalVideos"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </el-card>

        <!-- Video Preview Dialog -->
        <el-dialog
          v-model="previewDialogVisible"
          title="视频预览"
          width="80%"
          :before-close="closePreviewDialog"
        >
          <VideoPlayer
            v-if="previewVideoData"
            :video="previewVideoData"
            :auto-play="false"
          />
        </el-dialog>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload, List, Refresh, Search } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import VideoPlayer from '@/components/VideoPlayer.vue'
import type { 
  ImportVideosRequest,
  ImportVideosResponse,
  TikTokVideo,
  ListVideosRequest
} from '@/types'
import {
  importVideos,
  listVideos
} from '@/api/videoReview'

const router = useRouter()
const userStore = useUserStore()

// State
const importing = ref(false)
const loadingVideos = ref(false)
const importResult = ref<ImportVideosResponse | null>(null)
const videos = ref<TikTokVideo[]>([])
const searchQuery = ref('')
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const totalVideos = ref(0)
const previewDialogVisible = ref(false)
const previewVideoData = ref<TikTokVideo | null>(null)

// Form
const importForm = reactive<ImportVideosRequest>({
  r2_path_prefix: 'douyin/PostmanAgent/'
})

// Computed
const filteredVideos = computed(() => {
  let filtered = videos.value
  
  if (searchQuery.value) {
    filtered = filtered.filter(video => 
      video.filename.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  }
  
  if (statusFilter.value) {
    filtered = filtered.filter(video => video.status === statusFilter.value)
  }
  
  return filtered
})

// Import videos
const handleImportVideos = async () => {
  if (!importForm.r2_path_prefix.trim()) {
    ElMessage.warning('请输入 R2 路径前缀')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要从路径 "${importForm.r2_path_prefix}" 导入视频吗？`,
      '确认导入',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    importing.value = true
    importResult.value = null
    
    const response = await importVideos(importForm)
    importResult.value = response
    
    if (response.imported_count > 0) {
      ElMessage.success(`成功导入 ${response.imported_count} 个视频`)
      await loadVideos() // Refresh video list
    } else {
      ElMessage.info('没有找到新的视频文件')
    }
    
    if (response.errors.length > 0) {
      ElMessage.warning(`导入过程中遇到 ${response.errors.length} 个错误`)
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to import videos:', error)
      ElMessage.error(error.response?.data?.error || '导入视频失败')
    }
  } finally {
    importing.value = false
  }
}

// Load videos
const loadVideos = async () => {
  loadingVideos.value = true
  try {
    const params: ListVideosRequest = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    
    const response = await listVideos(params)
    videos.value = response.data
    totalVideos.value = response.total
  } catch (error) {
    console.error('Failed to load videos:', error)
    ElMessage.error('加载视频列表失败')
  } finally {
    loadingVideos.value = false
  }
}

// Apply filters
const applyFilters = () => {
  currentPage.value = 1
  loadVideos()
}

// Reset form
const resetForm = () => {
  importForm.r2_path_prefix = 'douyin/PostmanAgent/'
  importResult.value = null
}

// Preview video
const previewVideo = (video: TikTokVideo) => {
  previewVideoData.value = video
  previewDialogVisible.value = true
}

// Close preview dialog
const closePreviewDialog = () => {
  previewDialogVisible.value = false
  previewVideoData.value = null
}

// Pagination handlers
const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  loadVideos()
}

const handleCurrentChange = (page: number) => {
  currentPage.value = page
  loadVideos()
}

// Navigation
const goToVideoList = () => {
  // In a real implementation, this would navigate to a dedicated video list page
  ElMessage.info('视频列表功能开发中')
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}

// Utility functions
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDuration = (seconds: number): string => {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const formatTime = (timeString: string) => {
  return new Date(timeString).toLocaleString('zh-CN')
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'pending': return 'warning'
    case 'first_review_completed': return 'info'
    case 'second_review_completed': return 'success'
    default: return 'default'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'pending': return '待审核'
    case 'first_review_completed': return '一审完成'
    case 'second_review_completed': return '二审完成'
    default: return status
  }
}

// Lifecycle
onMounted(() => {
  loadVideos()
})
</script>

<style scoped>
.video-import-page {
  height: 100vh;
  background: #f5f5f5;
}

.header {
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-left h2 {
  margin: 0;
  color: #333;
  font-size: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.main-content {
  padding: 20px;
}

.import-card,
.result-card,
.video-list-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
}

.form-help {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
}

.result-summary {
  display: flex;
  gap: 24px;
  margin-bottom: 16px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.summary-item .label {
  font-size: 14px;
  color: #666;
}

.summary-item .value {
  font-size: 18px;
  font-weight: bold;
}

.summary-item .value.success {
  color: #67c23a;
}

.summary-item .value.info {
  color: #409eff;
}

.summary-item .value.error {
  color: #f56c6c;
}

.error-list h4 {
  margin: 0 0 12px 0;
  color: #333;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.video-filters {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}

/* Responsive design */
@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    gap: 16px;
    height: auto;
    padding: 16px 0;
  }
  
  .main-content {
    padding: 16px;
  }
  
  .result-summary {
    flex-direction: column;
    gap: 12px;
  }
  
  .video-filters {
    flex-direction: column;
    align-items: stretch;
  }
  
  .video-filters .el-input,
  .video-filters .el-select {
    width: 100% !important;
  }
}
</style>
