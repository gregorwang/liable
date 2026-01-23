<template>
  <el-dialog
    v-model="dialogVisible"
    title="Bug反馈"
    width="620px"
    @close="resetForm"
  >
    <el-alert
      title="提示：按 F12 打开控制台，将报错信息复制到下方“错误信息”里会更容易定位问题。"
      type="info"
      show-icon
      :closable="false"
      class="bug-alert"
    />

    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="90px"
      scroll-to-error
    >
      <el-form-item label="标题" prop="title">
        <el-input
          v-model="formData.title"
          placeholder="简要概括问题（可选）"
          maxlength="50"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="问题描述" prop="description">
        <el-input
          v-model="formData.description"
          type="textarea"
          placeholder="描述你遇到的问题或复现步骤"
          rows="4"
          maxlength="1000"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="错误信息" prop="errorDetails">
        <el-input
          v-model="formData.errorDetails"
          type="textarea"
          placeholder="从控制台复制的报错信息"
          rows="3"
          maxlength="2000"
          show-word-limit
        />
        <div class="form-tip">按 F12 打开控制台（Console），复制错误信息粘贴到这里。</div>
      </el-form-item>

      <el-form-item label="截图" prop="screenshots">
        <el-upload
          v-model:file-list="fileList"
          list-type="picture"
          multiple
          :limit="2"
          :auto-upload="false"
          :accept="acceptTypes"
          :before-upload="beforeUpload"
          :on-exceed="handleExceed"
        >
          <el-button type="primary" plain>选择截图</el-button>
        </el-upload>
        <div class="form-tip">最多上传 2 张，每张小于 1MB，支持 PNG/JPG/WEBP。</div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submitReport">提交</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules, UploadProps, UploadUserFile } from 'element-plus'
import { submitBugReport } from '@/api/bugReports'

interface Props {
  modelValue: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void }>()

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

const formRef = ref<FormInstance>()
const submitting = ref(false)
const fileList = ref<UploadUserFile[]>([])

const formData = reactive({
  title: '',
  description: '',
  errorDetails: '',
})

const rules: FormRules = {
  description: [{ required: true, message: '请填写问题描述', trigger: 'blur' }],
}

const acceptTypes = 'image/png,image/jpeg,image/webp'
const maxSize = 1024 * 1024

const beforeUpload: UploadProps['beforeUpload'] = (file) => {
  const allowedTypes = new Set(['image/png', 'image/jpeg', 'image/webp', 'image/jpg'])
  if (!allowedTypes.has(file.type)) {
    ElMessage.error('仅支持 PNG/JPG/WEBP 格式截图')
    return false
  }
  if (file.size > maxSize) {
    ElMessage.error('每张截图大小需小于 1MB')
    return false
  }
  return true
}

const handleExceed: UploadProps['onExceed'] = () => {
  ElMessage.warning('每次最多上传 2 张截图')
}

const resetForm = () => {
  formData.title = ''
  formData.description = ''
  formData.errorDetails = ''
  fileList.value = []
  formRef.value?.clearValidate()
}

const submitReport = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitting.value = true
  try {
    const payload = new FormData()
    payload.append('title', formData.title.trim())
    payload.append('description', formData.description.trim())
    payload.append('error_details', formData.errorDetails.trim())
    payload.append('page_url', window.location.href)

    fileList.value.forEach((file) => {
      if (file.raw) {
        payload.append('screenshots', file.raw, file.name)
      }
    })

    await submitBugReport(payload)
    ElMessage.success('Bug反馈已提交，感谢你的反馈！')
    dialogVisible.value = false
    resetForm()
  } catch (error) {
    console.error('Failed to submit bug report', error)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.bug-alert {
  margin-bottom: var(--spacing-4);
}

.form-tip {
  font-size: var(--text-xs);
  color: var(--color-text-400);
  margin-top: var(--spacing-2);
}
</style>
