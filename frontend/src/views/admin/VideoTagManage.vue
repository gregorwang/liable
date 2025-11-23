<template>
  <div class="video-tag-manage-container">
    <el-card class="header-card" shadow="never">
      <template #header>
        <div class="card-header">
          <h2>视频队列标签管理</h2>
          <span class="subtitle">配置视频审核队列的标签体系</span>
        </div>
      </template>
    </el-card>

    <el-card class="content-card" shadow="never">
      <template #header>
        <div class="section-header">
          <el-segmented v-model="selectedQueue" :options="queueOptions" size="large">
            <template #default="{ item }">
              <div class="queue-option">
                <span>{{ item.label }}</span>
                <el-badge :value="item.count" v-if="item.count > 0" :max="99" />
              </div>
            </template>
          </el-segmented>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            新建标签
          </el-button>
        </div>
      </template>

      <div v-if="filteredTags.length === 0" class="empty-state">
        <el-empty description="暂无标签，点击新建标签开始配置" />
      </div>

      <div v-else class="tags-container">
        <div
          v-for="(tags, category) in groupedTags"
          :key="category"
          class="category-section"
        >
          <div class="category-header">
            <h3>{{ getCategoryName(category) }}</h3>
            <el-tag size="small">{{ tags.length }} 个</el-tag>
          </div>

          <el-table :data="tags" stripe>
            <el-table-column prop="name" label="标签名称" width="150" />
            <el-table-column prop="description" label="描述" min-width="200" />
            <el-table-column prop="queue_id" label="所属队列" width="120">
              <template #default="{ row }">
                <el-tag v-if="row.queue_id" :type="getQueueTagType(row.queue_id)">
                  {{ getQueueName(row.queue_id) }}
                </el-tag>
                <el-tag v-else type="info">通用</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="is_active" label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
                  {{ row.is_active ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" fixed="right" width="200">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="handleEdit(row)">
                  编辑
                </el-button>
                <el-button
                  :type="row.is_active ? 'warning' : 'success'"
                  size="small"
                  @click="handleToggle(row)"
                >
                  {{ row.is_active ? '禁用' : '启用' }}
                </el-button>
                <el-button type="danger" size="small" @click="handleDelete(row.id)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑标签' : '新建标签'"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="标签名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入标签名称（如：内容优质）" />
        </el-form-item>

        <el-form-item label="标签描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入标签描述"
          />
        </el-form-item>

        <el-form-item label="标签分类" prop="category">
          <el-select v-model="form.category" placeholder="请选择标签分类">
            <el-option label="内容质量" value="content" />
            <el-option label="技术质量" value="technical" />
            <el-option label="合规性" value="compliance" />
            <el-option label="传播潜力" value="engagement" />
          </el-select>
        </el-form-item>

        <el-form-item label="所属队列" prop="queue_id">
          <el-select v-model="form.queue_id" placeholder="选择所属队列（留空为通用标签）" clearable>
            <el-option label="100k流量池" value="100k" />
            <el-option label="1m流量池" value="1m" />
            <el-option label="10m流量池" value="10m" />
          </el-select>
          <div class="form-tip">
            通用标签可用于所有队列，专属标签仅用于指定队列
          </div>
        </el-form-item>

        <el-form-item v-if="isEditing" label="状态">
          <el-switch
            v-model="form.is_active"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="submitLoading"
          @click="handleSubmit"
        >
          确认
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getAllVideoTags,
  createVideoTag,
  updateVideoTag,
  deleteVideoTag,
  toggleVideoTagActive,
  type VideoQualityTag
} from '@/api/videoTag'

const loading = ref(false)
const tags = ref<VideoQualityTag[]>([])
const dialogVisible = ref(false)
const isEditing = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()
const selectedQueue = ref('all')

const queueOptions = computed(() => [
  { label: '全部', value: 'all', count: tags.value.length },
  { label: '通用标签', value: 'common', count: tags.value.filter(t => !t.queue_id).length },
  { label: '100k队列', value: '100k', count: tags.value.filter(t => t.queue_id === '100k').length },
  { label: '1m队列', value: '1m', count: tags.value.filter(t => t.queue_id === '1m').length },
  { label: '10m队列', value: '10m', count: tags.value.filter(t => t.queue_id === '10m').length },
])

const form = reactive({
  id: 0,
  name: '',
  description: '',
  category: '',
  queue_id: null as string | null,
  is_active: true,
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入标签名称', trigger: 'blur' }],
  description: [{ required: true, message: '请输入标签描述', trigger: 'blur' }],
  category: [{ required: true, message: '请选择标签分类', trigger: 'change' }],
}

// 过滤标签
const filteredTags = computed(() => {
  if (selectedQueue.value === 'all') {
    return tags.value
  } else if (selectedQueue.value === 'common') {
    return tags.value.filter(t => !t.queue_id)
  } else {
    return tags.value.filter(t => t.queue_id === selectedQueue.value)
  }
})

// 按分类分组
const groupedTags = computed(() => {
  const groups: Record<string, VideoQualityTag[]> = {}
  filteredTags.value.forEach(tag => {
    if (!groups[tag.category]) {
      groups[tag.category] = []
    }
    groups[tag.category].push(tag)
  })
  return groups
})

// 辅助函数
const getCategoryName = (category: string) => {
  const names: Record<string, string> = {
    content: '内容质量',
    technical: '技术质量',
    compliance: '合规性',
    engagement: '传播潜力'
  }
  return names[category] || category
}

const getQueueName = (queueId: string) => {
  const names: Record<string, string> = {
    '100k': '100k',
    '1m': '1m',
    '10m': '10m'
  }
  return names[queueId] || queueId
}

const getQueueTagType = (queueId: string) => {
  const types: Record<string, any> = {
    '100k': 'primary',
    '1m': 'warning',
    '10m': 'danger'
  }
  return types[queueId] || 'info'
}

// 加载标签
const loadTags = async () => {
  loading.value = true
  try {
    const res = await getAllVideoTags('video')
    tags.value = res.tags
  } catch (error: any) {
    console.error('加载标签失败:', error)
  } finally {
    loading.value = false
  }
}

// 重置表单
const resetForm = () => {
  form.id = 0
  form.name = ''
  form.description = ''
  form.category = ''
  form.queue_id = null
  form.is_active = true
}

// 新建标签
const handleCreate = () => {
  resetForm()
  // 如果选中了特定队列，自动填充
  if (selectedQueue.value !== 'all' && selectedQueue.value !== 'common') {
    form.queue_id = selectedQueue.value
  }
  isEditing.value = false
  dialogVisible.value = true
}

// 编辑标签
const handleEdit = (tag: VideoQualityTag) => {
  form.id = tag.id
  form.name = tag.name
  form.description = tag.description
  form.category = tag.category
  form.queue_id = tag.queue_id
  form.is_active = tag.is_active
  isEditing.value = true
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitLoading.value = true
    try {
      if (isEditing.value) {
        await updateVideoTag(form.id, {
          name: form.name,
          description: form.description,
          category: form.category,
          queue_id: form.queue_id,
          is_active: form.is_active,
        })
        ElMessage.success('更新成功')
      } else {
        await createVideoTag({
          name: form.name,
          description: form.description,
          category: form.category,
          scope: 'video',
          queue_id: form.queue_id,
        })
        ElMessage.success('创建成功')
      }

      dialogVisible.value = false
      await loadTags()
    } catch (error: any) {
      console.error('操作失败:', error)
    } finally {
      submitLoading.value = false
    }
  })
}

// 切换状态
const handleToggle = async (tag: VideoQualityTag) => {
  const action = tag.is_active ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(`确定要${action}标签"${tag.name}"吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await toggleVideoTagActive(tag.id)
    ElMessage.success(`${action}成功`)
    await loadTags()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('操作失败:', error)
    }
  }
}

// 删除标签
const handleDelete = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除该标签吗？删除后无法恢复', '警告', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error',
    })

    await deleteVideoTag(id)
    ElMessage.success('删除成功')
    await loadTags()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
    }
  }
}

onMounted(() => {
  loadTags()
})
</script>

<style scoped>
.video-tag-manage-container {
  padding: 20px;
}

.header-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.card-header h2 {
  margin: 0;
  font-size: 24px;
  color: var(--el-text-color-primary);
}

.subtitle {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.content-card {
  min-height: calc(100vh - 220px);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.queue-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
}

.tags-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.category-section {
  padding: 16px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  background-color: var(--el-fill-color-blank);
}

.category-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid var(--el-border-color);
}

.category-header h3 {
  margin: 0;
  font-size: 18px;
  color: var(--el-text-color-primary);
}

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
