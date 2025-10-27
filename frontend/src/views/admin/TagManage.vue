<template>
  <div class="admin-tag-management-content">
          <el-card shadow="hover">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center">
                <span style="font-weight: bold">标签列表</span>
                <div>
                  <el-button type="primary" size="small" @click="handleCreate">
                    新建标签
                  </el-button>
                  <el-button size="small" @click="loadTags">刷新</el-button>
                </div>
              </div>
            </template>
            
            <el-table
              v-loading="loading"
              :data="tags"
              style="width: 100%"
            >
              <el-table-column prop="id" label="ID" width="80" />
              <el-table-column prop="name" label="标签名称" width="200" />
              <el-table-column prop="description" label="描述" />
              <el-table-column prop="is_active" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag v-if="row.is_active" type="success">启用</el-tag>
                  <el-tag v-else type="info">禁用</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="创建时间" width="180">
                <template #default="{ row }">
                  {{ formatDate(row.created_at) }}
                </template>
              </el-table-column>
              <el-table-column label="操作" fixed="right" width="220">
                <template #default="{ row }">
                  <el-button
                    type="primary"
                    size="small"
                    @click="handleEdit(row)"
                  >
                    编辑
                  </el-button>
                  <el-button
                    v-if="row.is_active"
                    type="warning"
                    size="small"
                    @click="handleToggle(row)"
                  >
                    禁用
                  </el-button>
                  <el-button
                    v-else
                    type="success"
                    size="small"
                    @click="handleToggle(row)"
                  >
                    启用
                  </el-button>
                  <el-button
                    type="danger"
                    size="small"
                    @click="handleDelete(row.id)"
                  >
                    删除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
    
    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑标签' : '新建标签'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="80px"
      >
        <el-form-item label="标签名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入标签名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入标签描述"
          />
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { getAllTags, createTag, updateTag, deleteTag } from '../../api/admin'
import type { Tag } from '../../types'
import { formatDate } from '../../utils/format'

const loading = ref(false)
const tags = ref<Tag[]>([])
const dialogVisible = ref(false)
const isEditing = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  id: 0,
  name: '',
  description: '',
  is_active: true,
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入标签名称', trigger: 'blur' },
  ],
  description: [
    { required: true, message: '请输入标签描述', trigger: 'blur' },
  ],
}


onMounted(() => {
  loadTags()
})

const loadTags = async () => {
  loading.value = true
  try {
    const res = await getAllTags()
    tags.value = res.tags
  } catch (error) {
    console.error('Failed to load tags:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.id = 0
  form.name = ''
  form.description = ''
  form.is_active = true
}

const handleCreate = () => {
  resetForm()
  isEditing.value = false
  dialogVisible.value = true
}

const handleEdit = (tag: Tag) => {
  form.id = tag.id
  form.name = tag.name
  form.description = tag.description
  form.is_active = tag.is_active
  isEditing.value = true
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitLoading.value = true
    try {
      if (isEditing.value) {
        await updateTag(form.id, {
          name: form.name,
          description: form.description,
          is_active: form.is_active,
        })
        ElMessage.success('更新成功')
      } else {
        await createTag(form.name, form.description)
        ElMessage.success('创建成功')
      }
      
      dialogVisible.value = false
      await loadTags()
    } catch (error) {
      console.error('Failed to submit:', error)
    } finally {
      submitLoading.value = false
    }
  })
}

const handleToggle = async (tag: Tag) => {
  const action = tag.is_active ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(`确认${action}该标签？`, '提示', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning',
    })
    
    await updateTag(tag.id, { is_active: !tag.is_active })
    ElMessage.success(`${action}成功`)
    await loadTags()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to toggle tag:', error)
    }
  }
}

const handleDelete = async (tagId: number) => {
  try {
    await ElMessageBox.confirm('确认删除该标签？删除后无法恢复', '警告', {
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      type: 'error',
    })
    
    await deleteTag(tagId)
    ElMessage.success('删除成功')
    await loadTags()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Failed to delete tag:', error)
    }
  }
}

</script>

<style scoped>
/* ============================================
   管理员标签管理页面样式
   ============================================ */
.admin-tag-management-content {
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

