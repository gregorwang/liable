<template>
  <div class="permission-selector">
    <!-- 搜索框 -->
    <el-input
      v-model="searchQuery"
      placeholder="搜索权限..."
      :prefix-icon="Search"
      clearable
      class="search-input"
      @input="handleSearch"
    />

    <!-- 权限列表（按分类分组） -->
    <div class="permission-list">
      <el-collapse v-model="activeCategories">
        <el-collapse-item
          v-for="(permissions, category) in groupedPermissions"
          :key="category"
          :name="category"
        >
          <template #title>
            <div class="category-title">
              <span class="category-name">{{ category }}</span>
              <span class="category-count">({{ permissions.length }})</span>
            </div>
          </template>

          <el-checkbox-group v-model="selectedKeys" class="permission-group">
            <el-checkbox
              v-for="permission in permissions"
              :key="permission.permission_key"
              :label="permission.permission_key"
              class="permission-item"
            >
              <template #default>
                <el-tooltip
                  :content="permission.description"
                  placement="top"
                  :show-after="500"
                >
                  <div class="permission-info">
                    <span class="permission-name">{{ permission.name }}</span>
                    <span class="permission-key">{{ permission.permission_key }}</span>
                  </div>
                </el-tooltip>
              </template>
            </el-checkbox>
          </el-checkbox-group>
        </el-collapse-item>
      </el-collapse>
    </div>

    <!-- 没有找到权限 -->
    <el-empty
      v-if="Object.keys(groupedPermissions).length === 0"
      description="没有找到匹配的权限"
      :image-size="100"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import type { Permission } from '../api/permission'

const props = defineProps<{
  permissions: Permission[]
  modelValue: string[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string[]): void
}>()

const searchQuery = ref('')
const activeCategories = ref<string[]>([])
const selectedKeys = computed<string[]>({
  get: () => props.modelValue,
  set: (value) => {
    emit('update:modelValue', value)
  }
})

// 过滤后的权限列表
const filteredPermissions = computed(() => {
  if (!searchQuery.value) {
    return props.permissions
  }

  const query = searchQuery.value.toLowerCase()
  return props.permissions.filter((permission) => {
    return (
      permission.name.toLowerCase().includes(query) ||
      permission.permission_key.toLowerCase().includes(query) ||
      permission.description.toLowerCase().includes(query) ||
      permission.category.toLowerCase().includes(query)
    )
  })
})

// 按分类分组权限
const groupedPermissions = computed(() => {
  const groups: Record<string, Permission[]> = {}

  filteredPermissions.value.forEach((permission) => {
    if (!groups[permission.category]) {
      groups[permission.category] = []
    }
    groups[permission.category].push(permission)
  })

  // 排序分类
  return Object.keys(groups)
    .sort()
    .reduce((acc, key) => {
      acc[key] = groups[key]
      return acc
    }, {} as Record<string, Permission[]>)
})

// 搜索处理
const handleSearch = () => {
  // 搜索时自动展开第一个分类
  const categories = Object.keys(groupedPermissions.value)
  if (categories.length > 0) {
    activeCategories.value = [categories[0]]
  }
}
</script>

<style scoped>
.permission-selector {
  max-height: 500px;
  overflow-y: auto;
}

.search-input {
  margin-bottom: 16px;
}

.permission-list {
  border: 1px solid var(--el-border-color-light);
  border-radius: 4px;
}

.category-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.category-name {
  color: var(--el-text-color-primary);
}

.category-count {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.permission-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 12px;
}

.permission-item {
  width: 100%;
  margin-right: 0;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: flex-start;
}

.permission-item:hover {
  background-color: var(--el-fill-color-light);
}

.permission-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  cursor: pointer;
  user-select: none;
}

.permission-name {
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.permission-key {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  font-family: 'Courier New', monospace;
}

:deep(.el-checkbox__label) {
  width: 100%;
  padding-left: 8px;
}

:deep(.el-checkbox__input) {
  margin-top: 2px;
}
</style>

