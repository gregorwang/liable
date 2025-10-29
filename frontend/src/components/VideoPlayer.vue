<template>
  <div class="video-player">
    <div class="video-container">
      <video
        ref="videoElement"
        :src="videoUrl"
        controls
        preload="metadata"
        class="video-element"
        @loadedmetadata="onLoadedMetadata"
        @error="onError"
      >
        您的浏览器不支持视频播放
      </video>
    </div>
    
    <div class="video-info" v-if="video">
      <div class="info-row">
        <span class="label">文件名:</span>
        <span class="value">{{ video.filename }}</span>
      </div>
      <div class="info-row">
        <span class="label">文件大小:</span>
        <span class="value">{{ formatFileSize(video.file_size) }}</span>
      </div>
      <div class="info-row" v-if="video.duration">
        <span class="label">时长:</span>
        <span class="value">{{ formatDuration(video.duration) }}</span>
      </div>
      <div class="info-row">
        <span class="label">状态:</span>
        <el-tag :type="getStatusType(video.status)">
          {{ getStatusText(video.status) }}
        </el-tag>
      </div>
    </div>
    
    <div class="video-actions">
      <el-button 
        type="primary" 
        @click="refreshVideoUrl"
        :loading="loading || isRetrying"
        size="small"
      >
        {{ isRetrying ? `重试中 (${retryCount}/${maxRetries})` : '刷新视频链接' }}
      </el-button>
      <el-button 
        @click="toggleLoop"
        :type="isLooping ? 'success' : 'default'"
        size="small"
      >
        {{ isLooping ? '关闭循环' : '开启循环' }}
      </el-button>
      <el-button 
        @click="checkVideoFormat"
        :loading="loading"
        size="small"
        type="info"
      >
        检测视频格式
      </el-button>
    </div>
    
    <!-- Error details panel -->
    <div v-if="lastError" class="error-details">
      <el-alert
        :title="lastError.title"
        :description="lastError.description"
        type="error"
        :closable="true"
        @close="lastError = null"
        show-icon
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { TikTokVideo } from '@/types'
import { generateVideoURL } from '@/api/videoReview'

interface Props {
  video: TikTokVideo
  autoPlay?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  autoPlay: true
})

const emit = defineEmits<{
  loaded: [duration: number]
  error: [error: string]
}>()

const videoElement = ref<HTMLVideoElement>()
const videoUrl = ref<string>('')
const loading = ref(false)
const isLooping = ref(false)
const duration = ref(0)
const retryCount = ref(0)
const maxRetries = 3
const isRetrying = ref(false)
const lastError = ref<{ title: string; description: string } | null>(null)

// Check if video URL is expired
const isUrlExpired = computed(() => {
  if (!props.video.url_expires_at) return true
  return new Date(props.video.url_expires_at) <= new Date()
})

// Get video URL - use cached or generate new
const getVideoUrl = async () => {
  if (props.video.video_url && !isUrlExpired.value) {
    videoUrl.value = props.video.video_url
    return
  }
  
  await refreshVideoUrl()
}

// Refresh video URL
const refreshVideoUrl = async () => {
  if (!props.video.id) return
  
  loading.value = true
  try {
    const response = await generateVideoURL({ video_id: props.video.id })
    videoUrl.value = response.video_url
    
    // Update the video object with new URL and expiration
    props.video.video_url = response.video_url
    props.video.url_expires_at = response.expires_at
    
    ElMessage.success('视频链接已刷新')
  } catch (error) {
    console.error('Failed to generate video URL:', error)
    ElMessage.error('获取视频链接失败')
    emit('error', 'Failed to generate video URL')
  } finally {
    loading.value = false
  }
}

// Toggle video loop
const toggleLoop = () => {
  if (videoElement.value) {
    isLooping.value = !isLooping.value
    videoElement.value.loop = isLooping.value
  }
}

// Check video format compatibility
const checkVideoFormat = async () => {
  if (!videoUrl.value) {
    ElMessage.warning('请先加载视频')
    return
  }
  
  loading.value = true
  try {
    // 创建一个临时的video元素来检测格式
    const testVideo = document.createElement('video')
    testVideo.preload = 'metadata'
    testVideo.muted = true
    
    const formatInfo = await new Promise<{
      canPlay: boolean
      codec: string
      error?: string
    }>((resolve) => {
      const timeout = setTimeout(() => {
        resolve({
          canPlay: false,
          codec: 'unknown',
          error: '检测超时'
        })
      }, 10000)
      
      testVideo.onloadedmetadata = () => {
        clearTimeout(timeout)
        const canPlay = testVideo.canPlayType('video/mp4')
        resolve({
          canPlay: !!canPlay,
          codec: canPlay || 'unknown',
          error: canPlay ? undefined : '浏览器不支持此视频格式'
        })
      }
      
      testVideo.onerror = (e: string | Event) => {
        clearTimeout(timeout)
        const error = (e as Event).target ? ((e as Event).target as HTMLVideoElement)?.error : null
        resolve({
          canPlay: false,
          codec: 'unknown',
          error: error ? `错误代码: ${error.code}` : '视频加载失败'
        })
      }
      
      testVideo.src = videoUrl.value
    })
    
    // 清理临时元素
    testVideo.src = ''
    testVideo.remove()
    
    if (formatInfo.canPlay) {
      ElMessage.success('视频格式兼容，可以正常播放')
    } else {
      ElMessage.error(`视频格式不兼容: ${formatInfo.error}`)
      lastError.value = {
        title: '视频格式检测失败',
        description: `错误信息: ${formatInfo.error}\n视频URL: ${videoUrl.value}\n建议: 尝试刷新视频链接或联系技术支持`
      }
    }
  } catch (error) {
    console.error('Format check failed:', error)
    ElMessage.error('格式检测失败')
  } finally {
    loading.value = false
  }
}

// Format file size
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// Format duration
const formatDuration = (seconds: number): string => {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

// Get status type for tag
const getStatusType = (status: string) => {
  switch (status) {
    case 'pending': return 'warning'
    case 'first_review_completed': return 'info'
    case 'second_review_completed': return 'success'
    default: return 'default'
  }
}

// Get status text
const getStatusText = (status: string) => {
  switch (status) {
    case 'pending': return '待审核'
    case 'first_review_completed': return '一审完成'
    case 'second_review_completed': return '二审完成'
    default: return status
  }
}

// Handle video loaded metadata
const onLoadedMetadata = () => {
  if (videoElement.value) {
    duration.value = videoElement.value.duration
    emit('loaded', duration.value)
    
    // 清除之前的错误状态
    lastError.value = null
    retryCount.value = 0
    isRetrying.value = false
    
    console.log('Video loaded successfully:', {
      videoId: props.video.id,
      filename: props.video.filename,
      duration: duration.value,
      videoUrl: videoUrl.value
    })
    
    // Auto play if enabled
    if (props.autoPlay) {
      videoElement.value.play().catch(err => {
        console.warn('Auto play failed:', err)
      })
    }
  }
}

// Handle video error
const onError = (event: Event) => {
  const error = (event.target as HTMLVideoElement)?.error
  let errorMessage = '视频加载失败'
  let shouldRetry = false
  
  console.error('Video error details:', {
    error,
    videoUrl: videoUrl.value,
    isUrlExpired: isUrlExpired.value,
    retryCount: retryCount.value,
    videoId: props.video.id,
    filename: props.video.filename
  })
  
  if (error) {
    switch (error.code) {
      case 1: 
        errorMessage = '视频加载被中止'
        break
      case 2: 
        errorMessage = '网络错误导致视频下载失败'
        shouldRetry = true
        break
      case 3: 
        errorMessage = '视频解码错误'
        break
      case 4: 
        errorMessage = '视频格式不支持或URL已过期'
        // 如果是格式错误且URL可能过期，尝试刷新URL
        if (isUrlExpired.value) {
          shouldRetry = true
          errorMessage = '视频URL已过期，正在尝试刷新...'
        }
        break
    }
  }
  
  // 如果应该重试且未超过最大重试次数
  if (shouldRetry && retryCount.value < maxRetries && !isRetrying.value) {
    retryCount.value++
    isRetrying.value = true
    
    console.log(`Attempting retry ${retryCount.value}/${maxRetries} for video ${props.video.id}`)
    
    // 延迟重试，避免立即重试
    setTimeout(async () => {
      try {
        await refreshVideoUrl()
        ElMessage.info(`正在重试加载视频 (${retryCount.value}/${maxRetries})`)
      } catch (retryError) {
        console.error('Retry failed:', retryError)
        if (retryCount.value >= maxRetries) {
          ElMessage.error('视频加载失败，已达到最大重试次数')
          emit('error', '视频加载失败，已达到最大重试次数')
        }
      } finally {
        isRetrying.value = false
      }
    }, 1000 * retryCount.value) // 递增延迟
    
    return
  }
  
  // 如果超过最大重试次数或不应该重试
  if (retryCount.value >= maxRetries) {
    errorMessage = `视频加载失败，已重试 ${maxRetries} 次`
  }
  
  // 设置详细错误信息
  lastError.value = {
    title: '视频播放错误',
    description: `${errorMessage}\n\n详细信息:\n- 视频ID: ${props.video.id}\n- 文件名: ${props.video.filename}\n- 文件大小: ${formatFileSize(props.video.file_size)}\n- 视频URL: ${videoUrl.value}\n- URL过期时间: ${props.video.url_expires_at}\n- 重试次数: ${retryCount.value}/${maxRetries}\n- 错误代码: ${error?.code || 'unknown'}\n\n建议:\n1. 点击"刷新视频链接"获取新的URL\n2. 点击"检测视频格式"检查兼容性\n3. 如果问题持续，请联系技术支持`
  }
  
  ElMessage.error(errorMessage)
  emit('error', errorMessage)
}

// Reset retry count when video changes
watch(() => props.video, () => {
  retryCount.value = 0
  isRetrying.value = false
  getVideoUrl()
}, { immediate: true })

// Lifecycle
onMounted(() => {
  getVideoUrl()
})

onUnmounted(() => {
  // Clean up video element
  if (videoElement.value) {
    videoElement.value.pause()
    videoElement.value.src = ''
  }
})
</script>

<style scoped>
.video-player {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.video-container {
  position: relative;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

.video-element {
  width: 100%;
  height: auto;
  max-height: 400px;
  display: block;
}

.video-info {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 6px;
  font-size: 14px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.label {
  font-weight: 500;
  color: #666;
}

.value {
  color: #333;
}

.video-actions {
  display: flex;
  gap: 8px;
  justify-content: center;
}

/* Error details panel */
.error-details {
  margin-top: 16px;
}

.error-details :deep(.el-alert__description) {
  white-space: pre-line;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
}

/* Responsive design */
@media (max-width: 768px) {
  .video-element {
    max-height: 250px;
  }
  
  .video-info {
    font-size: 12px;
  }
  
  .video-actions {
    flex-direction: column;
  }
  
  .error-details :deep(.el-alert__description) {
    font-size: 11px;
  }
}
</style>

