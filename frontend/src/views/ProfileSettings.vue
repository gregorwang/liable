<template>
  <div class="profile-settings">
    <el-row :gutter="24">
      <el-col :xs="24" :lg="10">
        <el-card class="profile-card" shadow="never">
          <div class="card-title">个人资料</div>
          <div class="avatar-section">
            <el-avatar :size="88" :src="displayAvatarUrl" class="avatar-preview">
              {{ profile?.username?.charAt(0).toUpperCase() }}
            </el-avatar>
            <div class="avatar-actions">
              <el-upload
                :show-file-list="false"
                :auto-upload="false"
                :before-upload="beforeAvatarUpload"
                :on-change="handleAvatarChange"
                accept="image/png,image/jpeg,image/webp"
              >
                <el-button type="primary" plain>选择头像</el-button>
              </el-upload>
              <el-button
                type="primary"
                :loading="avatarUploading"
                :disabled="!pendingAvatar"
                @click="uploadAvatarFile"
              >
                保存头像
              </el-button>
              <div class="form-tip">支持 PNG/JPG/WEBP，大小不超过 1MB。</div>
            </div>
          </div>

          <el-divider />

          <el-form
            ref="profileFormRef"
            :model="profileForm"
            label-width="90px"
            class="profile-form"
          >
            <el-form-item label="个人性别">
              <el-radio-group v-model="profileForm.gender">
                <el-radio-button label="male">男</el-radio-button>
                <el-radio-button label="female">女</el-radio-button>
                <el-radio-button label="unknown">保密</el-radio-button>
                <el-radio-button label="other">其他</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="个性签名">
              <el-input
                v-model="profileForm.signature"
                type="textarea"
                :rows="3"
                maxlength="200"
                show-word-limit
                placeholder="写一句话介绍你自己"
              />
            </el-form-item>
            <div class="form-actions">
              <el-button
                type="primary"
                :loading="profileSaving"
                @click="saveProfile"
              >
                保存资料
              </el-button>
            </div>
          </el-form>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="14">
        <el-card class="profile-card" shadow="never">
          <div class="card-title">系统信息</div>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="个人角色">
              {{ roleLabel }}
            </el-descriptions-item>
            <el-descriptions-item label="账户注册时间">
              {{ formatDate(profile?.created_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="个人权限">
              <div class="permission-tags">
                <el-tag
                  v-for="permission in permissions"
                  :key="permission"
                  size="small"
                  type="info"
                >
                  {{ permission }}
                </el-tag>
                <span v-if="permissions.length === 0" class="empty-text">暂无权限</span>
              </div>
            </el-descriptions-item>
          </el-descriptions>

          <el-divider />

          <el-form
            ref="systemFormRef"
            :model="systemForm"
            label-width="110px"
            class="system-form"
            :disabled="!canEditSystem"
          >
            <el-form-item label="个人办公地点">
              <el-input v-model="systemForm.office_location" placeholder="如：上海/远程" />
            </el-form-item>
            <el-form-item label="个人部门">
              <el-input v-model="systemForm.department" placeholder="如：内容审核部" />
            </el-form-item>
            <el-form-item label="个人学校">
              <el-input v-model="systemForm.school" placeholder="如：XX大学" />
            </el-form-item>
            <el-form-item label="个人公司">
              <el-input v-model="systemForm.company" placeholder="如：XX科技" />
            </el-form-item>
            <el-form-item label="直属上级">
              <el-input v-model="systemForm.direct_manager" placeholder="如：张三" />
            </el-form-item>
            <div class="form-actions" v-if="canEditSystem">
              <el-button
                type="primary"
                :loading="systemSaving"
                @click="saveSystemProfile"
              >
                保存系统信息
              </el-button>
            </div>
            <div v-else class="system-tip">
              系统字段由管理员维护，当前账号无修改权限。
            </div>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { UploadFile, UploadFiles } from 'element-plus'
import { updateProfile, updateSystemProfile, uploadAvatar } from '@/api/profile'
import { useUserStore } from '@/stores/user'
import type { User } from '@/types'

const userStore = useUserStore()

const profile = ref<User | null>(null)
const permissions = ref<string[]>([])
const pendingAvatar = ref<File | null>(null)
const avatarPreview = ref<string | null>(null)
const avatarUploading = ref(false)
const profileSaving = ref(false)
const systemSaving = ref(false)

const profileFormRef = ref()
const systemFormRef = ref()

const profileForm = reactive({
  gender: 'unknown',
  signature: '',
})

const systemForm = reactive({
  office_location: '',
  department: '',
  school: '',
  company: '',
  direct_manager: '',
})

const displayAvatarUrl = computed(() => avatarPreview.value || profile.value?.avatar_url || '')

const roleLabel = computed(() => {
  if (profile.value?.role === 'admin') return '管理员'
  if (profile.value?.role === 'reviewer') return '审核员'
  return '未知'
})

const canEditSystem = computed(() => permissions.value.includes('users:profile:update'))

const loadProfile = async () => {
  const res = await userStore.loadProfile()
  profile.value = res.user
  permissions.value = res.permissions || []

  profileForm.gender = res.user.gender || 'unknown'
  profileForm.signature = res.user.signature || ''
  systemForm.office_location = res.user.office_location || ''
  systemForm.department = res.user.department || ''
  systemForm.school = res.user.school || ''
  systemForm.company = res.user.company || ''
  systemForm.direct_manager = res.user.direct_manager || ''
}

const beforeAvatarUpload = (file: File) => {
  const allowedTypes = new Set(['image/png', 'image/jpeg', 'image/webp', 'image/jpg'])
  if (!allowedTypes.has(file.type)) {
    ElMessage.error('仅支持 PNG/JPG/WEBP 格式头像')
    return false
  }
  if (file.size > 1024 * 1024) {
    ElMessage.error('头像大小需小于 1MB')
    return false
  }
  return true
}

const handleAvatarChange = (file: UploadFile, _fileList: UploadFiles) => {
  if (!file.raw) return
  if (!beforeAvatarUpload(file.raw)) return
  if (avatarPreview.value) {
    URL.revokeObjectURL(avatarPreview.value)
  }
  pendingAvatar.value = file.raw
  avatarPreview.value = URL.createObjectURL(file.raw)
}

const uploadAvatarFile = async () => {
  if (!pendingAvatar.value) return
  avatarUploading.value = true
  try {
    const formData = new FormData()
    formData.append('avatar', pendingAvatar.value)
    await uploadAvatar(formData)
    ElMessage.success('头像已更新')
    pendingAvatar.value = null
    if (avatarPreview.value) {
      URL.revokeObjectURL(avatarPreview.value)
      avatarPreview.value = null
    }
    await loadProfile()
  } catch (error) {
    console.error('Failed to upload avatar', error)
  } finally {
    avatarUploading.value = false
  }
}

const saveProfile = async () => {
  profileSaving.value = true
  try {
    await updateProfile({
      gender: profileForm.gender,
      signature: profileForm.signature,
    })
    ElMessage.success('个人资料已保存')
    await loadProfile()
  } catch (error) {
    console.error('Failed to update profile', error)
  } finally {
    profileSaving.value = false
  }
}

const saveSystemProfile = async () => {
  if (!canEditSystem.value) return
  systemSaving.value = true
  try {
    await updateSystemProfile({
      office_location: systemForm.office_location,
      department: systemForm.department,
      school: systemForm.school,
      company: systemForm.company,
      direct_manager: systemForm.direct_manager,
    })
    ElMessage.success('系统信息已更新')
    await loadProfile()
  } catch (error) {
    console.error('Failed to update system profile', error)
  } finally {
    systemSaving.value = false
  }
}

const formatDate = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  loadProfile().catch((error) => {
    console.error('Failed to load profile', error)
  })
})
</script>

<style scoped>
.profile-settings {
  padding: var(--spacing-6);
}

.profile-card {
  border-radius: 16px;
  border: 1px solid rgba(204, 122, 77, 0.15);
}

.card-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-000);
  margin-bottom: var(--spacing-4);
  font-family: var(--font-sans);
}

.avatar-section {
  display: flex;
  gap: var(--spacing-5);
  align-items: center;
}

.avatar-preview {
  border: 1px solid rgba(204, 122, 77, 0.2);
}

.avatar-actions {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.form-tip {
  font-size: var(--text-xs);
  color: var(--color-text-400);
}

.profile-form,
.system-form {
  margin-top: var(--spacing-4);
}

.form-actions {
  display: flex;
  justify-content: flex-start;
  margin-top: var(--spacing-4);
}

.permission-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
}

.empty-text {
  color: var(--color-text-400);
  font-size: var(--text-sm);
}

.system-tip {
  margin-top: var(--spacing-3);
  font-size: var(--text-sm);
  color: var(--color-text-400);
}

@media (max-width: 768px) {
  .profile-settings {
    padding: var(--spacing-4);
  }

  .avatar-section {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
