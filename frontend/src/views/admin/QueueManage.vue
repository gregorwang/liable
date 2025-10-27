<template>
  <div class="queue-manage-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span class="title">
            <i class="el-icon-management"></i> 队列配置管理
          </span>
          <div class="header-actions">
            <el-tag effect="dark">共 {{ total }} 个队列</el-tag>
            <el-button type="primary" @click="openAddDialog">
              <i class="el-icon-plus"></i> 新建队列
            </el-button>
            <el-button @click="loadQueues">
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
              placeholder="搜索队列名称..."
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
              v-model="filterActive"
              placeholder="过滤状态"
              clearable
              @change="loadQueues"
            >
              <el-option label="活跃队列" :value="true"></el-option>
              <el-option label="已禁用队列" :value="false"></el-option>
            </el-select>
          </el-col>

          <el-col :xs="24" :sm="12" :md="8">
            <el-pagination
              :current-page="currentPage"
              :page-size="pageSize"
              :total="total"
              layout="total, prev, pager, next"
              @current-change="currentPage = $event; loadQueues()"
            />
          </el-col>
        </el-row>
      </div>

      <!-- Queues Table -->
      <el-table
        :data="queues"
        stripe
        style="width: 100%; margin-top: 20px"
        v-loading="loading"
      >
        <!-- Queue Name -->
        <el-table-column prop="queue_name" label="队列名称" width="150" show-overflow-tooltip />

        <!-- Description -->
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />

        <!-- Priority -->
        <el-table-column prop="priority" label="优先级" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getPriorityType(row.priority)">{{ row.priority }}</el-tag>
          </template>
        </el-table-column>

        <!-- Tasks Stats -->
        <el-table-column label="任务统计" width="180" align="center">
          <template #default="{ row }">
            <div class="stats-info">
              <div class="stat-item">
                <span class="stat-label">总数:</span>
                <span class="stat-value">{{ row.total_tasks }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">已审:</span>
                <span class="stat-value" style="color: #67c23a">{{ row.completed_tasks }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">待审:</span>
                <span class="stat-value" style="color: #f56c6c">{{ row.pending_tasks }}</span>
              </div>
            </div>
          </template>
        </el-table-column>

        <!-- Progress -->
        <el-table-column label="进度" width="120">
          <template #default="{ row }">
            <el-progress
              :percentage="row.total_tasks ? Math.round((row.completed_tasks / row.total_tasks) * 100) : 0"
              :color="getProgressColor"
            />
          </template>
        </el-table-column>

        <!-- Status -->
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_active ? 'success' : 'info'">
              {{ row.is_active ? '活跃' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- Created At -->
        <el-table-column prop="created_at" label="创建时间" width="170" align="center">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>

        <!-- Actions -->
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="openEditDialog(row)">
              <i class="el-icon-edit"></i> 编辑
            </el-button>
            <el-popconfirm
              title="确定要删除该队列吗？"
              confirm-button-text="删除"
              cancel-button-text="取消"
              @confirm="deleteQueue(row.id)"
            >
              <template #reference>
                <el-button type="danger" size="small">
                  <i class="el-icon-delete"></i> 删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Add/Edit Dialog -->
    <el-dialog
      :title="isEditing ? '编辑队列' : '新建队列'"
      v-model="dialogVisible"
      width="600px"
      @close="resetForm"
    >
      <el-form
        :model="formData"
        :rules="formRules"
        ref="formRef"
        label-width="120px"
      >
        <el-form-item label="队列名称" prop="queue_name">
          <el-input
            v-model="formData.queue_name"
            placeholder="请输入队列名称"
            :disabled="isEditing"
          />
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            rows="3"
            placeholder="请输入队列描述"
          />
        </el-form-item>

        <el-form-item label="优先级" prop="priority">
          <el-input-number
            v-model="formData.priority"
            :min="0"
            :max="1000"
            @change="handlePriorityChange"
          />
          <span style="margin-left: 10px; color: #909399">数值越大优先级越高</span>
        </el-form-item>

        <el-form-item label="总任务数" prop="total_tasks">
          <el-input-number
            v-model="formData.total_tasks"
            :min="0"
            placeholder="总任务数"
          />
        </el-form-item>

        <el-form-item label="已审核数" prop="completed_tasks">
          <el-input-number
            v-model="formData.completed_tasks"
            :min="0"
            placeholder="已审核数"
          />
          <span style="margin-left: 10px; color: #909399">
            待审核数: {{ formData.total_tasks - formData.completed_tasks }}
          </span>
        </el-form-item>

        <el-form-item label="状态" prop="is_active">
          <el-switch v-model="formData.is_active" />
          <span style="margin-left: 10px; color: #909399">
            {{ formData.is_active ? '活跃' : '禁用' }}
          </span>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">{{ isEditing ? '更新' : '创建' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as adminAPI from '@/api/admin'

interface FormData {
  queue_name: string
  description: string
  priority: number
  total_tasks: number
  completed_tasks: number
  is_active: boolean
}

// State
const loading = ref(false)
const queues = ref<adminAPI.TaskQueue[]>([])
const searchText = ref('')
const filterActive = ref<boolean | null>(null)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref()

const formData = ref<FormData>({
  queue_name: '',
  description: '',
  priority: 0,
  total_tasks: 0,
  completed_tasks: 0,
  is_active: true,
})

const formRules = {
  queue_name: [
    { required: true, message: '队列名称不能为空', trigger: 'blur' },
    { min: 1, max: 100, message: '队列名称长度为1-100字符', trigger: 'blur' },
  ],
  priority: [{ type: 'number', required: true, message: '优先级不能为空', trigger: 'blur' }],
  total_tasks: [{ type: 'number', required: true, message: '总任务数不能为空', trigger: 'blur' }],
  completed_tasks: [
    {
      type: 'number',
      validator: (rule: any, value: number, callback: any) => {
        if (value > formData.value.total_tasks) {
          callback(new Error('已审核数不能大于总任务数'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
}

// Computed
const filteredQueues = computed(() => {
  if (!searchText.value) return queues.value
  return queues.value.filter((q) =>
    q.queue_name.toLowerCase().includes(searchText.value.toLowerCase())
  )
})

// Methods
const loadQueues = async () => {
  loading.value = true
  try {
    const response = await adminAPI.listTaskQueues({
      search: searchText.value,
      is_active: filterActive.value,
      page: currentPage.value,
      page_size: pageSize.value,
    })
    queues.value = response.data
    total.value = response.total
    currentPage.value = response.page
  } catch (error: any) {
    ElMessage.error(error.message || '加载队列失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadQueues()
}

const openAddDialog = () => {
  isEditing.value = false
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = (row: adminAPI.TaskQueue) => {
  isEditing.value = true
  editingId.value = row.id
  formData.value = {
    queue_name: row.queue_name,
    description: row.description,
    priority: row.priority,
    total_tasks: row.total_tasks,
    completed_tasks: row.completed_tasks,
    is_active: row.is_active,
  }
  dialogVisible.value = true
}

const resetForm = () => {
  formRef.value?.clearValidate()
  formData.value = {
    queue_name: '',
    description: '',
    priority: 0,
    total_tasks: 0,
    completed_tasks: 0,
    is_active: true,
  }
}

const submitForm = async () => {
  try {
    await formRef.value.validate()
    loading.value = true

    if (isEditing.value && editingId.value) {
      await adminAPI.updateTaskQueue(editingId.value, {
        queue_name: formData.value.queue_name,
        description: formData.value.description,
        priority: formData.value.priority,
        total_tasks: formData.value.total_tasks,
        completed_tasks: formData.value.completed_tasks,
        is_active: formData.value.is_active,
      })
      ElMessage.success('队列更新成功')
    } else {
      await adminAPI.createTaskQueue({
        queue_name: formData.value.queue_name,
        description: formData.value.description,
        priority: formData.value.priority,
        total_tasks: formData.value.total_tasks,
        completed_tasks: formData.value.completed_tasks,
      })
      ElMessage.success('队列创建成功')
    }

    dialogVisible.value = false
    loadQueues()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    loading.value = false
  }
}

const deleteQueue = async (id: number) => {
  loading.value = true
  try {
    await adminAPI.deleteTaskQueue(id)
    ElMessage.success('队列删除成功')
    loadQueues()
  } catch (error: any) {
    ElMessage.error(error.message || '删除失败')
  } finally {
    loading.value = false
  }
}

const getPriorityType = (priority: number) => {
  if (priority >= 80) return 'danger'
  if (priority >= 50) return 'warning'
  if (priority >= 20) return 'info'
  return 'success'
}

const getProgressColor = (percentage: number) => {
  if (percentage >= 80) return '#67c23a'
  if (percentage >= 50) return '#e6a23c'
  if (percentage >= 20) return '#409eff'
  return '#f56c6c'
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN')
}

const handlePriorityChange = () => {
  // Priority change handler
}

onMounted(() => {
  loadQueues()
})
</script>

<style scoped>
.queue-manage-container {
  padding: 20px;
}

.box-card {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.card-header .title {
  font-size: 18px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-actions {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.search-bar {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.stats-info {
  display: flex;
  flex-direction: column;
  gap: 5px;
  font-size: 12px;
}

.stat-item {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.stat-label {
  color: #909399;
  min-width: 40px;
}

.stat-value {
  font-weight: 500;
  color: #303133;
}

@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .search-bar {
    padding: 10px;
  }
}
</style>
