<template>
  <div class="permission-manage-container">
    <el-card class="header-card" shadow="never">
      <template #header>
        <div class="card-header">
          <h2>权限管理</h2>
          <span class="subtitle">管理用户权限分配</span>
        </div>
      </template>
    </el-card>

    <div class="content-wrapper">
      <!-- 左侧：用户列表 -->
      <el-card class="user-list-card" shadow="never">
        <template #header>
          <div class="section-header">
            <span>用户列表</span>
            <el-tag size="small">{{ users.length }} 个用户</el-tag>
          </div>
        </template>

        <el-input
          v-model="userSearchQuery"
          placeholder="搜索用户..."
          :prefix-icon="Search"
          clearable
          class="search-input"
        />

        <el-menu
          :default-active="selectedUserId?.toString()"
          class="user-menu"
          @select="handleUserSelect"
        >
          <el-menu-item
            v-for="user in filteredUsers"
            :key="user.id"
            :index="user.id.toString()"
          >
            <div class="user-item">
              <div class="user-info">
                <span class="user-name">{{ user.username }}</span>
                <el-tag :type="getRoleType(user.role)" size="small">
                  {{ getRoleText(user.role) }}
                </el-tag>
              </div>
              <el-badge
                v-if="userPermissionCounts[user.id]"
                :value="userPermissionCounts[user.id]"
                class="permission-badge"
              />
            </div>
          </el-menu-item>
        </el-menu>
      </el-card>

      <!-- 右侧：权限管理 -->
      <el-card class="permission-card" shadow="never">
        <template #header>
          <div class="section-header">
            <span>权限配置</span>
            <el-button
              v-if="selectedUserId"
              type="primary"
              size="small"
              :icon="RefreshRight"
              @click="loadUserPermissions"
            >
              刷新
            </el-button>
          </div>
        </template>

        <div v-if="!selectedUserId" class="empty-state">
          <el-empty description="请从左侧选择一个用户" :image-size="150" />
        </div>

        <div v-else class="permission-content">
          <!-- 当前用户信息 -->
          <div class="selected-user-info">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="用户名">
                {{ selectedUser?.username }}
              </el-descriptions-item>
              <el-descriptions-item label="角色">
                <el-tag :type="getRoleType(selectedUser?.role)">
                  {{ getRoleText(selectedUser?.role) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="当前权限数">
                {{ currentUserPermissions.length }}
              </el-descriptions-item>
              <el-descriptions-item label="用户状态">
                <el-tag :type="selectedUser?.status === 'approved' ? 'success' : 'warning'">
                  {{ selectedUser?.status }}
                </el-tag>
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <!-- 权限操作区 -->
          <el-tabs v-model="activeTab" class="permission-tabs">
            <!-- 当前权限 -->
            <el-tab-pane label="当前权限" name="current">
              <div v-if="currentUserPermissions.length === 0" class="empty-permissions">
                <el-empty description="该用户暂无权限" :image-size="100" />
              </div>
              <div v-else class="current-permissions">
                <div
                  v-for="(perms, category) in groupedCurrentPermissions"
                  :key="category"
                  class="permission-category"
                >
                  <div class="category-header">
                    <h4>{{ category }}</h4>
                    <el-tag size="small">{{ perms.length }}</el-tag>
                  </div>
                  <el-space wrap :size="8">
                    <el-tag
                      v-for="key in perms"
                      :key="key"
                      closable
                      @close="handleRevokePermission(key)"
                    >
                      {{ getPermissionName(key) }}
                    </el-tag>
                  </el-space>
                </div>
              </div>
            </el-tab-pane>

            <!-- 授予权限 -->
            <el-tab-pane label="授予权限" name="grant">
              <PermissionSelector
                v-if="allPermissions.length > 0"
                v-model="selectedPermissionsToGrant"
                :permissions="availablePermissions"
              />
              <div class="action-buttons">
                <el-button
                  type="primary"
                  :disabled="selectedPermissionsToGrant.length === 0"
                  @click="handleGrantPermissions"
                >
                  授予选中权限 ({{ selectedPermissionsToGrant.length }})
                </el-button>
                <el-button @click="selectedPermissionsToGrant = []">清空选择</el-button>
              </div>
            </el-tab-pane>

            <!-- 批量撤销 -->
            <el-tab-pane label="批量撤销" name="revoke">
              <el-alert
                type="warning"
                title="批量撤销权限"
                description="选择要撤销的权限，然后点击撤销按钮"
                :closable="false"
                show-icon
                class="revoke-alert"
              />
              <PermissionSelector
                v-if="currentPermissionObjects.length > 0"
                v-model="selectedPermissionsToRevoke"
                :permissions="currentPermissionObjects"
              />
              <div class="action-buttons">
                <el-button
                  type="danger"
                  :disabled="selectedPermissionsToRevoke.length === 0"
                  @click="handleRevokePermissions"
                >
                  撤销选中权限 ({{ selectedPermissionsToRevoke.length }})
                </el-button>
                <el-button @click="selectedPermissionsToRevoke = []">清空选择</el-button>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshRight } from '@element-plus/icons-vue'
import { getAllUsers } from '@/api/admin'
import {
  getAllPermissions,
  getUserPermissions,
  grantPermissions,
  revokePermissions,
  type Permission
} from '@/api/permission'
import PermissionSelector from '@/components/PermissionSelector.vue'

interface User {
  id: number
  username: string
  role: string
  status: string
  created_at: string
}

const users = ref<User[]>([])
const allPermissions = ref<Permission[]>([])
const currentUserPermissions = ref<string[]>([])
const selectedUserId = ref<number | null>(null)
const userSearchQuery = ref('')
const activeTab = ref('current')
const selectedPermissionsToGrant = ref<string[]>([])
const selectedPermissionsToRevoke = ref<string[]>([])
const userPermissionCounts = ref<Record<number, number>>({})

// 过滤用户
const filteredUsers = computed(() => {
  if (!userSearchQuery.value) {
    return users.value
  }
  const query = userSearchQuery.value.toLowerCase()
  return users.value.filter((user) => user.username.toLowerCase().includes(query))
})

// 选中的用户
const selectedUser = computed(() => {
  return users.value.find((u) => u.id === selectedUserId.value)
})

// 获取当前权限的完整对象
const currentPermissionObjects = computed(() => {
  return allPermissions.value.filter((p) => currentUserPermissions.value.includes(p.permission_key))
})

// 可授予的权限（排除已有权限）
const availablePermissions = computed(() => {
  return allPermissions.value.filter((p) => !currentUserPermissions.value.includes(p.permission_key))
})

// 将当前权限按分类分组
const groupedCurrentPermissions = computed(() => {
  const groups: Record<string, string[]> = {}
  currentUserPermissions.value.forEach((key) => {
    const permission = allPermissions.value.find((p) => p.permission_key === key)
    if (permission) {
      if (!groups[permission.category]) {
        groups[permission.category] = []
      }
      groups[permission.category].push(key)
    }
  })
  return groups
})

// 获取权限名称
const getPermissionName = (key: string) => {
  const permission = allPermissions.value.find((p) => p.permission_key === key)
  return permission ? permission.name : key
}

// 获取角色类型
const getRoleType = (role?: string) => {
  switch (role) {
    case 'admin':
      return 'danger'
    case 'reviewer':
      return 'success'
    default:
      return 'info'
  }
}

// 获取角色文本
const getRoleText = (role?: string) => {
  switch (role) {
    case 'admin':
      return '管理员'
    case 'reviewer':
      return '审核员'
    default:
      return '未知'
  }
}

// 加载用户列表
const loadUsers = async () => {
  try {
    const res = await getAllUsers()
    // 获取所有用户（request 已经返回了 data，不需要再加 .data）
    users.value = res.users || []
    
    // 加载每个用户的权限数量
    for (const user of users.value) {
      const perms = await getUserPermissions(user.id)
      userPermissionCounts.value[user.id] = perms.permissions.length
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '加载用户列表失败')
  }
}

// 加载所有权限
const loadAllPermissions = async () => {
  try {
    const res = await getAllPermissions()
    allPermissions.value = res.permissions || []
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '加载权限列表失败')
  }
}

// 加载用户权限
const loadUserPermissions = async () => {
  if (!selectedUserId.value) return

  try {
    const res = await getUserPermissions(selectedUserId.value)
    currentUserPermissions.value = res.permissions || []
    // 更新权限数量
    userPermissionCounts.value[selectedUserId.value] = currentUserPermissions.value.length
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '加载用户权限失败')
  }
}

// 选择用户
const handleUserSelect = (userId: string) => {
  selectedUserId.value = parseInt(userId)
  loadUserPermissions()
  activeTab.value = 'current'
  selectedPermissionsToGrant.value = []
  selectedPermissionsToRevoke.value = []
}

// 授予权限
const handleGrantPermissions = async () => {
  if (!selectedUserId.value || selectedPermissionsToGrant.value.length === 0) return

  try {
    await ElMessageBox.confirm(
      `确定要为用户 ${selectedUser.value?.username} 授予 ${selectedPermissionsToGrant.value.length} 个权限吗？`,
      '确认授予',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await grantPermissions({
      user_id: selectedUserId.value,
      permission_keys: selectedPermissionsToGrant.value
    })

    ElMessage.success('权限授予成功')
    selectedPermissionsToGrant.value = []
    await loadUserPermissions()
    activeTab.value = 'current'
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '授予权限失败')
    }
  }
}

// 撤销权限（单个）
const handleRevokePermission = async (key: string) => {
  if (!selectedUserId.value) return

  try {
    await ElMessageBox.confirm(`确定要撤销权限 "${getPermissionName(key)}" 吗？`, '确认撤销', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await revokePermissions({
      user_id: selectedUserId.value,
      permission_keys: [key]
    })

    ElMessage.success('权限撤销成功')
    await loadUserPermissions()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '撤销权限失败')
    }
  }
}

// 批量撤销权限
const handleRevokePermissions = async () => {
  if (!selectedUserId.value || selectedPermissionsToRevoke.value.length === 0) return

  try {
    await ElMessageBox.confirm(
      `确定要撤销用户 ${selectedUser.value?.username} 的 ${selectedPermissionsToRevoke.value.length} 个权限吗？`,
      '确认撤销',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await revokePermissions({
      user_id: selectedUserId.value,
      permission_keys: selectedPermissionsToRevoke.value
    })

    ElMessage.success('权限撤销成功')
    selectedPermissionsToRevoke.value = []
    await loadUserPermissions()
    activeTab.value = 'current'
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '撤销权限失败')
    }
  }
}

onMounted(() => {
  loadUsers()
  loadAllPermissions()
})
</script>

<style scoped>
.permission-manage-container {
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

.content-wrapper {
  display: grid;
  grid-template-columns: 350px 1fr;
  gap: 20px;
  height: calc(100vh - 220px);
}

.user-list-card,
.permission-card {
  display: flex;
  flex-direction: column;
  height: 100%;
}

:deep(.el-card__body) {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.search-input {
  margin-bottom: 16px;
}

.user-menu {
  flex: 1;
  overflow-y: auto;
  border: none;
}

.user-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-name {
  font-weight: 500;
}

.permission-badge {
  margin-top: 6px;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.permission-content {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.selected-user-info {
  margin-bottom: 20px;
}

.permission-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
}

:deep(.el-tabs__content) {
  flex: 1;
  overflow-y: auto;
}

.empty-permissions {
  padding: 40px 0;
}

.current-permissions {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.permission-category {
  padding: 16px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  background-color: var(--el-fill-color-blank);
}

.category-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.category-header h4 {
  margin: 0;
  font-size: 16px;
  color: var(--el-text-color-primary);
}

.revoke-alert {
  margin-bottom: 16px;
}

.action-buttons {
  display: flex;
  gap: 12px;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>

