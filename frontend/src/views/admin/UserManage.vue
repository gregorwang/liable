<template>
  <div class="admin-user-management-content">
          <el-card shadow="hover">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center">
                <span style="font-weight: bold">待审批用户</span>
                <el-button size="small" @click="loadUsers">刷新</el-button>
              </div>
            </template>
            
            <el-table
              v-loading="loading"
              :data="users"
              style="width: 100%"
              :empty-text="users.length === 0 ? '暂无待审批用户' : ''"
            >
              <el-table-column prop="id" label="ID" width="80" />
              <el-table-column prop="username" label="用户名" width="200" />
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
              <el-table-column label="操作" fixed="right" width="200">
                <template #default="{ row }">
                  <el-button
                    v-if="row.status === 'pending'"
                    type="success"
                    size="small"
                    @click="handleApprove(row.id)"
                  >
                    通过
                  </el-button>
                  <el-button
                    v-if="row.status === 'pending'"
                    type="danger"
                    size="small"
                    @click="handleReject(row.id)"
                  >
                    拒绝
                  </el-button>
                  <span v-else style="color: #909399">-</span>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPendingUsers, approveUser } from '../../api/admin'
import type { User } from '../../types'
import { formatDate } from '../../utils/format'

const loading = ref(false)
const users = ref<User[]>([])

onMounted(() => {
  loadUsers()
})

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await getPendingUsers()
    users.value = res.users
  } catch (error) {
    console.error('Failed to load users:', error)
  } finally {
    loading.value = false
  }
}

const handleApprove = async (userId: number) => {
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
    }
  }
}

const handleReject = async (userId: number) => {
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
    }
  }
}

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

/* ============================================
   响应式设计
   ============================================ */
@media (max-width: 768px) {
  .main-content {
    padding: var(--spacing-4);
  }
}
</style>

