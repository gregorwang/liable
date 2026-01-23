<template>
  <div class="system-docs-container">
    <div class="page-header">
      <div class="title-block">
        <span class="title">
          <i class="el-icon-document"></i> 文档说明
        </span>
        <span class="subtitle">系统流程与规则文档说明</span>
      </div>
      <div class="header-actions">
        <el-tag v-if="canEditDocs" type="success" effect="dark">可编辑</el-tag>
        <el-tag v-else type="info" effect="dark">只读</el-tag>
      </div>
    </div>

    <el-card class="docs-card" shadow="hover" v-loading="docsLoading">
      <template #header>
        <div class="docs-header">
          <span class="card-header">文档详情</span>
          <span class="doc-count">共 {{ documents.length }} 篇</span>
        </div>
      </template>

      <div class="doc-controls">
        <el-select v-model="activeDocKey" placeholder="选择文档" class="doc-selector">
          <el-option
            v-for="doc in documents"
            :key="doc.key"
            :label="doc.title"
            :value="doc.key"
          />
        </el-select>
        <div class="doc-actions">
          <el-button
            v-if="canEditDocs && !isEditingDoc"
            size="small"
            @click="startEditDoc"
          >
            编辑
          </el-button>
          <el-button
            v-if="isEditingDoc"
            size="small"
            type="primary"
            :loading="savingDoc"
            :disabled="savingDoc"
            @click="saveDoc"
          >
            保存
          </el-button>
          <el-button v-if="isEditingDoc" size="small" @click="cancelEditDoc">取消</el-button>
        </div>
      </div>

      <div class="doc-meta" v-if="currentDoc">
        <span>更新时间：{{ formatDateTime(currentDoc.updated_at) }}</span>
        <span v-if="currentDoc.updated_by">更新人ID：{{ currentDoc.updated_by }}</span>
      </div>

      <div class="doc-body">
        <el-input
          v-if="isEditingDoc"
          v-model="docDraft"
          type="textarea"
          :rows="20"
          placeholder="请输入文档内容"
        />
        <pre v-else class="markdown-content">{{ currentDoc?.content || '暂无文档内容' }}</pre>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { SystemDocument } from '@/types'
import { listSystemDocuments, updateSystemDocument } from '@/api/docs'

const documents = ref<SystemDocument[]>([])
const activeDocKey = ref('')
const docsLoading = ref(false)
const canEditDocs = ref(false)
const isEditingDoc = ref(false)
const savingDoc = ref(false)
const docDraft = ref('')

const currentDoc = computed(() =>
  documents.value.find((doc) => doc.key === activeDocKey.value),
)

const formatDateTime = (dateTime: string): string => {
  try {
    const date = new Date(dateTime)
    return date.toLocaleString('zh-CN')
  } catch {
    return dateTime
  }
}

const loadDocuments = async () => {
  docsLoading.value = true
  try {
    const response = await listSystemDocuments()
    documents.value = response.data || []
    canEditDocs.value = response.can_edit
    if (!activeDocKey.value && documents.value.length > 0) {
      activeDocKey.value = documents.value[0].key
    }
  } catch (error) {
    console.error('Failed to load documents:', error)
    ElMessage.error('加载文档失败')
  } finally {
    docsLoading.value = false
  }
}

const startEditDoc = () => {
  if (!currentDoc.value) return
  docDraft.value = currentDoc.value.content || ''
  isEditingDoc.value = true
}

const cancelEditDoc = () => {
  isEditingDoc.value = false
  docDraft.value = ''
}

const saveDoc = async () => {
  if (!currentDoc.value) return
  savingDoc.value = true
  try {
    const updated = await updateSystemDocument(currentDoc.value.key, docDraft.value)
    documents.value = documents.value.map((doc) =>
      doc.key === updated.key ? updated : doc,
    )
    ElMessage.success('文档已更新')
    isEditingDoc.value = false
  } catch (error) {
    console.error('Failed to update document:', error)
    ElMessage.error('文档更新失败')
  } finally {
    savingDoc.value = false
  }
}

watch(activeDocKey, () => {
  if (isEditingDoc.value) {
    cancelEditDoc()
  }
})

onMounted(() => {
  loadDocuments()
})
</script>

<style scoped lang="scss">
.system-docs-container {
  padding: var(--spacing-8);
  background-color: var(--color-bg-100);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-6);
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-6);
}

.title-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.title {
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text-000);
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.subtitle {
  color: var(--color-text-300);
  font-size: var(--text-sm);
}

.header-actions {
  display: flex;
  gap: var(--spacing-3);
  align-items: center;
}

.docs-card :deep(.el-card__header) {
  border-bottom: 1px solid var(--color-border-lighter);
  padding: var(--spacing-4) var(--spacing-5);
}

.docs-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-2);
}

.card-header {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-000);
}

.doc-count {
  font-size: var(--text-xs);
  color: var(--color-text-300);
}

.doc-controls {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-3);
}

.doc-selector {
  width: 100%;
}

.doc-actions {
  display: flex;
  gap: var(--spacing-2);
  flex-wrap: wrap;
}

.doc-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: var(--text-sm);
  color: var(--color-text-300);
  margin-bottom: var(--spacing-3);
}

.doc-body :deep(.el-textarea__inner) {
  font-family: 'Fira Code', 'JetBrains Mono', Consolas, monospace;
}

.markdown-content {
  margin: 0;
  padding: var(--spacing-5);
  background: var(--color-bg-200);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-lighter);
  white-space: pre-wrap;
  line-height: 1.7;
  font-size: var(--text-sm);
  color: var(--color-text-200);
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
