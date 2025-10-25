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
            <h2>用户管理</h2>
            <div class="user-info">
              <span>{{ userStore.user?.username }}</span>
              <el-button @click="handleLogout" text>退出</el-button>
            </div>
          </div>
        </el-header>
        
        <el-main class="main-content">
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
import { getPendingUsers, approveUser } from '../../api/admin'
import type { User } from '../../types'
import { formatDate } from '../../utils/format'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loading = ref(false)
const users = ref<User[]>([])

const currentRoute = computed(() => route.path)

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

