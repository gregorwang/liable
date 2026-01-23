<template>
  <div class="admin-user-management-content">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="card-title">
            <span>用户管理</span>
            <el-tag size="small" type="info">{{ users.length }} 人</el-tag>
          </div>
          <div class="header-actions">
            <el-button size="small" type="primary" @click="openCreateDialog">新增用户</el-button>
            <el-button size="small" :loading="loading" :disabled="loading" @click="loadUsers">
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="users"
        style="width: 100%"
        :empty-text="users.length === 0 ? '暂无用户' : ''"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="180" />
        <el-table-column label="邮箱" min-width="220">
          <template #default="{ row }">
            <span>{{ row.email || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.role === 'admin'" type="danger">管理员</el-tag>
            <el-tag v-else type="primary">审核员</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.status === 'pending'" type="warning">待审批</el-tag>
            <el-tag v-else-if="row.status === 'approved'" type="success">已通过</el-tag>
            <el-tag v-else type="info">已拒绝</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="260">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'pending'"
              type="success"
              size="small"
              :loading="isActionLoading(`approve:${row.id}`)"
              :disabled="isActionLoading(`approve:${row.id}`)"
              @click="handleApprove(row.id)"
            >
              通过
            </el-button>
            <el-button
              v-if="row.status === 'pending'"
              type="danger"
              size="small"
              :loading="isActionLoading(`reject:${row.id}`)"
              :disabled="isActionLoading(`reject:${row.id}`)"
              @click="handleReject(row.id)"
            >
              拒绝
            </el-button>
            <el-button
              type="warning"
              size="small"
              :loading="isActionLoading(`delete:${row.id}`)"
              :disabled="isActionLoading(`delete:${row.id}`) || !canDeleteUser(row)"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="新增用户" width="480px">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="100px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="createForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="createForm.email" placeholder="可选，填写后可用验证码登录" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="createForm.password"
            placeholder="邮箱为空时必须填写"
            show-password
          />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="createForm.role" placeholder="请选择角色">
            <el-option label="审核员" value="reviewer" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="createForm.status" placeholder="请选择状态">
            <el-option label="已通过" value="approved" />
            <el-option label="待审批" value="pending" />
            <el-option label="已拒绝" value="rejected" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" :disabled="creating" @click="handleCreateUser">
          创建
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { getAllUsers, approveUser, createUser, deleteUser } from '../../api/admin'
import type { User } from '../../types'
import { formatDate } from '../../utils/format'
import { useUserStore } from '../../stores/user'

const loading = ref(false)
const users = ref<User[]>([])
const userStore = useUserStore()
const createDialogVisible = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = ref({
  username: '',
  email: '',
  password: '',
  role: 'reviewer',
  status: 'approved',
})
const pendingActions = reactive(new Set<string>())

const createRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, message: '用户名至少3位', trigger: 'blur' },
  ],
  email: [{ type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }],
  password: [
    {
      validator: (_rule: any, value: string, callback: (error?: Error) => void) => {
        if (!value && !createForm.value.email) {
          callback(new Error('邮箱为空时必须设置密码'))
          return
        }
        if (value && value.length < 6) {
          callback(new Error('密码至少6位'))
          return
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
}

onMounted(() => {
  loadUsers()
})

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await getAllUsers()
    users.value = res.users
  } catch (error) {
    console.error('Failed to load users:', error)
    ElMessage.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleApprove = async (userId: number) => {
  await runWithActionLock(`approve:${userId}`, async () => {
    try {
      await ElMessageBox.confirm('确认通过该用户的审核申请？', '提示', {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'success',
      })

      await approveUser(userId, 'approved')
      ElMessage.success('审批通过')
      await loadUsers()
    } catch (error: any) {
      if (error !== 'cancel') {
        console.error('Failed to approve user:', error)
        ElMessage.error('审批失败')
      }
    }
  })
}

const handleReject = async (userId: number) => {
  await runWithActionLock(`reject:${userId}`, async () => {
    try {
      await ElMessageBox.confirm('确认拒绝该用户的审核申请？', '提示', {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning',
      })

      await approveUser(userId, 'rejected')
      ElMessage.success('已拒绝')
      await loadUsers()
    } catch (error: any) {
      if (error !== 'cancel') {
        console.error('Failed to reject user:', error)
        ElMessage.error('拒绝失败')
      }
    }
  })
}

const handleDelete = async (user: User) => {
  await runWithActionLock(`delete:${user.id}`, async () => {
    try {
      await ElMessageBox.confirm(`确认删除用户 ${user.username}？`, '提示', {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning',
      })

      await deleteUser(user.id)
      ElMessage.success('用户已删除')
      await loadUsers()
    } catch (error: any) {
      if (error !== 'cancel') {
        console.error('Failed to delete user:', error)
        ElMessage.error('删除用户失败')
      }
    }
  })
}

const openCreateDialog = () => {
  createDialogVisible.value = true
}

const handleCreateUser = async () => {
  if (!createFormRef.value) return
  try {
    await createFormRef.value.validate()
  } catch {
    return
  }

  creating.value = true
  try {
    const payload: {
      username: string
      email?: string
      password?: string
      role: 'admin' | 'reviewer'
      status: 'pending' | 'approved' | 'rejected'
    } = {
      username: createForm.value.username.trim(),
      role: createForm.value.role as 'admin' | 'reviewer',
      status: createForm.value.status as 'pending' | 'approved' | 'rejected',
    }
    const email = createForm.value.email.trim()
    const password = createForm.value.password.trim()
    if (email) payload.email = email
    if (password) payload.password = password

    await createUser(payload)
    ElMessage.success('用户已创建')
    createDialogVisible.value = false
    await loadUsers()
  } catch (error) {
    console.error('Failed to create user:', error)
    ElMessage.error('创建用户失败')
  } finally {
    creating.value = false
  }
}

const resetCreateForm = () => {
  createForm.value = {
    username: '',
    email: '',
    password: '',
    role: 'reviewer',
    status: 'approved',
  }
  createFormRef.value?.clearValidate()
}

const runWithActionLock = async (key: string, action: () => Promise<void>) => {
  if (pendingActions.has(key)) return
  pendingActions.add(key)
  try {
    await action()
  } finally {
    pendingActions.delete(key)
  }
}

const isActionLoading = (key: string) => pendingActions.has(key)

const canDeleteUser = (user: User) => user.id !== userStore.user?.id

watch(
  () => createForm.value.email,
  async () => {
    await createFormRef.value?.validateField('password')
  },
)

watch(createDialogVisible, (visible) => {
  if (!visible) {
    resetCreateForm()
  }
})
</script>

<style scoped>
/* ============================================
   管理员用户管理页面样式
   ============================================ */
.admin-user-management-content {
  padding: var(--spacing-8);
  background-color: var(--color-bg-100);
  min-height: 100vh;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
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

