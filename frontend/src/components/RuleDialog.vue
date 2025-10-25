<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEditMode ? '编辑规则' : '新增规则'"
    width="70%"
    @close="resetForm"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="120px"
      scroll-to-error
    >
      <!-- Rule Code -->
      <el-form-item label="规则编号" prop="rule_code">
        <el-input
          v-model="formData.rule_code"
          placeholder="如: A1, B2, C3"
          :disabled="isEditMode"
          maxlength="10"
        />
      </el-form-item>

      <!-- Category -->
      <el-form-item label="分类" prop="category">
        <el-input
          v-model="formData.category"
          placeholder="如: 人身安全与暴力"
          maxlength="50"
        />
      </el-form-item>

      <!-- Subcategory -->
      <el-form-item label="二级标签" prop="subcategory">
        <el-input
          v-model="formData.subcategory"
          placeholder="如: 真实威胁/伤害"
          maxlength="50"
        />
      </el-form-item>

      <!-- Description -->
      <el-form-item label="描述" prop="description">
        <el-input
          v-model="formData.description"
          type="textarea"
          placeholder="简要描述规则内容"
          maxlength="200"
          show-word-limit
          rows="3"
        />
      </el-form-item>

      <!-- Judgment Criteria -->
      <el-form-item label="判定要点" prop="judgment_criteria">
        <el-input
          v-model="formData.judgment_criteria"
          type="textarea"
          placeholder="详细的判定标准"
          rows="3"
        />
      </el-form-item>

      <!-- Risk Level -->
      <el-form-item label="风险等级" prop="risk_level">
        <el-select v-model="formData.risk_level" placeholder="选择风险等级">
          <el-option label="低风险 (L)" value="L" />
          <el-option label="中风险 (M)" value="M" />
          <el-option label="高风险 (H)" value="H" />
          <el-option label="极高风险 (C)" value="C" />
        </el-select>
      </el-form-item>

      <!-- Action -->
      <el-form-item label="处置动作" prop="action">
        <el-input
          v-model="formData.action"
          type="textarea"
          placeholder="应采取的处置动作"
          rows="2"
        />
      </el-form-item>

      <!-- Boundary -->
      <el-form-item label="边界说明" prop="boundary">
        <el-input
          v-model="formData.boundary"
          type="textarea"
          placeholder="规则的边界和限制条件"
          rows="2"
        />
      </el-form-item>

      <!-- Examples -->
      <el-form-item label="示例" prop="examples">
        <el-input
          v-model="formData.examples"
          type="textarea"
          placeholder="具体示例"
          rows="3"
        />
      </el-form-item>

      <!-- Quick Tag -->
      <el-form-item label="快捷标签" prop="quick_tag">
        <el-input
          v-model="formData.quick_tag"
          placeholder="可选的快捷标签"
          maxlength="20"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">{{ isEditMode ? '更新' : '创建' }}</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import type { ModerationRule } from '@/types'
import * as moderationApi from '@/api/moderation'

interface Props {
  visible: boolean
  editingRule?: ModerationRule | null
}

interface Emits {
  (e: 'update:visible', visible: boolean): void
  (e: 'success'): void
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  editingRule: null,
})

const emit = defineEmits<Emits>()

const formRef = ref<FormInstance>()
const loading = ref(false)

const formData = reactive<ModerationRule>({
  rule_code: '',
  category: '',
  subcategory: '',
  description: '',
  judgment_criteria: '',
  risk_level: 'M',
  action: '',
  boundary: '',
  examples: '',
  quick_tag: '',
})

const rules = {
  rule_code: [{ required: true, message: '规则编号是必须的', trigger: 'blur' }],
  category: [{ required: true, message: '分类是必须的', trigger: 'blur' }],
  subcategory: [{ required: true, message: '二级标签是必须的', trigger: 'blur' }],
  description: [{ required: true, message: '描述是必须的', trigger: 'blur' }],
  risk_level: [{ required: true, message: '风险等级是必须的', trigger: 'change' }],
}

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

const isEditMode = computed(() => !!props.editingRule?.id)

const resetForm = () => {
  formRef.value?.resetFields()
  if (!isEditMode.value) {
    Object.assign(formData, {
      rule_code: '',
      category: '',
      subcategory: '',
      description: '',
      judgment_criteria: '',
      risk_level: 'M',
      action: '',
      boundary: '',
      examples: '',
      quick_tag: '',
    })
  }
}

const initializeFormData = () => {
  if (props.editingRule) {
    Object.assign(formData, {
      id: props.editingRule.id,
      rule_code: props.editingRule.rule_code,
      category: props.editingRule.category,
      subcategory: props.editingRule.subcategory,
      description: props.editingRule.description,
      judgment_criteria: props.editingRule.judgment_criteria || '',
      risk_level: props.editingRule.risk_level,
      action: props.editingRule.action || '',
      boundary: props.editingRule.boundary || '',
      examples: props.editingRule.examples || '',
      quick_tag: props.editingRule.quick_tag || '',
    })
  }
}

const submitForm = async () => {
  if (!formRef.value) return

  const isValid = await formRef.value.validate().catch(() => false)
  if (!isValid) return

  loading.value = true
  try {
    if (isEditMode.value && formData.id) {
      await moderationApi.updateRule(formData.id, formData)
      ElMessage.success('规则更新成功')
    } else {
      await moderationApi.createRule(formData)
      ElMessage.success('规则创建成功')
    }
    emit('success')
    dialogVisible.value = false
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '操作失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

// Watch for changes in editingRule and visible props
const updateFormData = () => {
  if (dialogVisible.value && props.editingRule) {
    initializeFormData()
  }
}

// Use a watcher-like approach
import { watch } from 'vue'
watch([() => props.visible, () => props.editingRule], updateFormData)
</script>

<style scoped lang="scss">
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
