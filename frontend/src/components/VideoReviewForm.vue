<template>
  <div class="video-review-form">
    <el-form :model="formData" label-width="120px" size="large">
      <!-- Quality Dimensions -->
      <el-card class="dimensions-card">
        <template #header>
          <div class="card-header">
            <span>质量评估</span>
          </div>
        </template>
        
        <!-- Content Quality Row -->
        <div class="dimension-row">
          <div class="dimension-label">
            <span>内容质量</span>
            <el-tag :type="getScoreType(formData.quality_dimensions.content_quality.score)" size="small">
              {{ formData.quality_dimensions.content_quality.score }}/10
            </el-tag>
          </div>
          <div class="dimension-controls">
            <el-select
              v-model="formData.quality_dimensions.content_quality.tags"
              multiple
              placeholder="选择标签"
              size="small"
              style="width: 200px; margin-right: 12px;"
            >
              <el-option
                v-for="tag in contentTags"
                :key="tag.id"
                :label="tag.name"
                :value="tag.name"
              />
            </el-select>
            <el-slider
              v-model="formData.quality_dimensions.content_quality.score"
              :min="1"
              :max="10"
              :step="1"
              show-stops
              show-input
              @change="updateOverallScore"
              style="width: 200px;"
            />
          </div>
        </div>

        <!-- Technical Quality Row -->
        <div class="dimension-row">
          <div class="dimension-label">
            <span>技术质量</span>
            <el-tag :type="getScoreType(formData.quality_dimensions.technical_quality.score)" size="small">
              {{ formData.quality_dimensions.technical_quality.score }}/10
            </el-tag>
          </div>
          <div class="dimension-controls">
            <el-select
              v-model="formData.quality_dimensions.technical_quality.tags"
              multiple
              placeholder="选择标签"
              size="small"
              style="width: 200px; margin-right: 12px;"
            >
              <el-option
                v-for="tag in technicalTags"
                :key="tag.id"
                :label="tag.name"
                :value="tag.name"
              />
            </el-select>
            <el-slider
              v-model="formData.quality_dimensions.technical_quality.score"
              :min="1"
              :max="10"
              :step="1"
              show-stops
              show-input
              @change="updateOverallScore"
              style="width: 200px;"
            />
          </div>
        </div>

        <!-- Compliance Row -->
        <div class="dimension-row">
          <div class="dimension-label">
            <span>合规性</span>
            <el-tag :type="getScoreType(formData.quality_dimensions.compliance.score)" size="small">
              {{ formData.quality_dimensions.compliance.score }}/10
            </el-tag>
          </div>
          <div class="dimension-controls">
            <el-select
              v-model="formData.quality_dimensions.compliance.tags"
              multiple
              placeholder="选择标签"
              size="small"
              style="width: 200px; margin-right: 12px;"
            >
              <el-option
                v-for="tag in complianceTags"
                :key="tag.id"
                :label="tag.name"
                :value="tag.name"
              />
            </el-select>
            <el-slider
              v-model="formData.quality_dimensions.compliance.score"
              :min="1"
              :max="10"
              :step="1"
              show-stops
              show-input
              @change="updateOverallScore"
              style="width: 200px;"
            />
          </div>
        </div>

        <!-- Engagement Potential Row -->
        <div class="dimension-row">
          <div class="dimension-label">
            <span>传播潜力</span>
            <el-tag :type="getScoreType(formData.quality_dimensions.engagement_potential.score)" size="small">
              {{ formData.quality_dimensions.engagement_potential.score }}/10
            </el-tag>
          </div>
          <div class="dimension-controls">
            <el-select
              v-model="formData.quality_dimensions.engagement_potential.tags"
              multiple
              placeholder="选择标签"
              size="small"
              style="width: 200px; margin-right: 12px;"
            >
              <el-option
                v-for="tag in engagementTags"
                :key="tag.id"
                :label="tag.name"
                :value="tag.name"
              />
            </el-select>
            <el-slider
              v-model="formData.quality_dimensions.engagement_potential.score"
              :min="1"
              :max="10"
              :step="1"
              show-stops
              show-input
              @change="updateOverallScore"
              style="width: 200px;"
            />
          </div>
        </div>
      </el-card>


      <!-- Traffic Pool Result -->
      <el-form-item label="流量池建议">
        <el-select
          v-model="formData.traffic_pool_result"
          placeholder="选择推荐的流量池"
          style="width: 100%"
        >
          <el-option label="精品池" value="精品池" />
          <el-option label="推荐池" value="推荐池" />
          <el-option label="普通池" value="普通池" />
          <el-option label="低质池" value="低质池" />
          <el-option label="拒绝池" value="拒绝池" />
        </el-select>
      </el-form-item>

      <!-- Approval Decision -->
      <el-form-item label="审核决定">
        <el-radio-group v-model="formData.is_approved" size="large">
          <el-radio-button :label="true" type="success">通过</el-radio-button>
          <el-radio-button :label="false" type="danger">拒绝</el-radio-button>
        </el-radio-group>
      </el-form-item>

      <!-- Reason -->
      <el-form-item label="审核理由">
        <el-input
          v-model="formData.reason"
          type="textarea"
          :rows="3"
          placeholder="请说明审核决定的理由"
        />
      </el-form-item>

      <!-- Submit Button -->
      <el-form-item>
        <el-button
          type="primary"
          size="large"
          @click="handleSubmit"
          :loading="submitting"
          style="width: 100%"
        >
          提交审核
        </el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { 
  VideoQualityTag,
  SubmitVideoFirstReviewRequest,
  SubmitVideoSecondReviewRequest
} from '@/types'
import { getVideoQualityTags } from '@/api/videoReview'

interface Props {
  taskId: number
  isSecondReview?: boolean
  firstReviewResult?: any // For second review comparison
}

const props = withDefaults(defineProps<Props>(), {
  isSecondReview: false
})

const emit = defineEmits<{
  submit: [data: SubmitVideoFirstReviewRequest | SubmitVideoSecondReviewRequest]
}>()

const submitting = ref(false)
const contentTags = ref<VideoQualityTag[]>([])
const technicalTags = ref<VideoQualityTag[]>([])
const complianceTags = ref<VideoQualityTag[]>([])
const engagementTags = ref<VideoQualityTag[]>([])

// Form data
const formData = reactive<SubmitVideoFirstReviewRequest & { overall_score: number }>({
  task_id: props.taskId,
  is_approved: true,
  quality_dimensions: {
    content_quality: { score: 5, tags: [] },
    technical_quality: { score: 5, tags: [] },
    compliance: { score: 5, tags: [] },
    engagement_potential: { score: 5, tags: [] }
  },
  traffic_pool_result: '普通池',
  reason: '',
  overall_score: 20 // Initial score
})

// Computed overall score
const overallScore = computed(() => {
  return formData.quality_dimensions.content_quality.score +
         formData.quality_dimensions.technical_quality.score +
         formData.quality_dimensions.compliance.score +
         formData.quality_dimensions.engagement_potential.score
})

// Update overall score
const updateOverallScore = () => {
  formData.overall_score = overallScore.value
}

// Get score type for tag color
const getScoreType = (score: number) => {
  if (score >= 8) return 'success'
  if (score >= 6) return 'warning'
  return 'danger'
}


// Load quality tags
const loadQualityTags = async () => {
  try {
    const [contentRes, technicalRes, complianceRes, engagementRes] = await Promise.all([
      getVideoQualityTags({ category: 'content' }),
      getVideoQualityTags({ category: 'technical' }),
      getVideoQualityTags({ category: 'compliance' }),
      getVideoQualityTags({ category: 'engagement' })
    ])
    
    contentTags.value = contentRes.tags
    technicalTags.value = technicalRes.tags
    complianceTags.value = complianceRes.tags
    engagementTags.value = engagementRes.tags
  } catch (error) {
    console.error('Failed to load quality tags:', error)
    ElMessage.error('加载质量标签失败')
  }
}

// Handle form submission
const handleSubmit = () => {
  // Validate form
  if (!formData.reason?.trim()) {
    ElMessage.warning('请填写审核理由')
    return
  }
  
  if (formData.quality_dimensions.content_quality.score < 1 ||
      formData.quality_dimensions.technical_quality.score < 1 ||
      formData.quality_dimensions.compliance.score < 1 ||
      formData.quality_dimensions.engagement_potential.score < 1) {
    ElMessage.warning('所有维度评分不能低于1分')
    return
  }
  
  submitting.value = true
  
  try {
    // Update overall score before submitting
    formData.overall_score = overallScore.value
    
    emit('submit', { ...formData })
  } catch (error) {
    console.error('Submit error:', error)
    ElMessage.error('提交失败')
  } finally {
    submitting.value = false
  }
}

// Initialize form
onMounted(() => {
  loadQualityTags()
  updateOverallScore()
})
</script>

<style scoped>
.video-review-form {
  max-width: 800px;
  margin: 0 auto;
}

.dimensions-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
}

.dimension-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}

.dimension-row:last-child {
  border-bottom: none;
}

.dimension-label {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 120px;
  font-weight: 500;
  color: #333;
}

.dimension-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  justify-content: flex-end;
}


/* Responsive design */
@media (max-width: 768px) {
  .video-review-form {
    padding: 0 16px;
  }
  
  .dimension-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .dimension-controls {
    width: 100%;
    justify-content: flex-start;
    flex-wrap: wrap;
  }
  
  .dimension-controls .el-select {
    width: 100% !important;
    margin-right: 0 !important;
    margin-bottom: 8px;
  }
  
  .dimension-controls .el-slider {
    width: 100% !important;
  }
}

/* Form styling */
:deep(.el-form-item__label) {
  font-weight: 500;
  color: #333;
}

:deep(.el-card__header) {
  background: #f8f9fa;
  border-bottom: 1px solid #e9ecef;
}

:deep(.el-slider__runway) {
  background-color: #e4e7ed;
}

:deep(.el-slider__bar) {
  background-color: #409eff;
}

:deep(.el-slider__button) {
  border-color: #409eff;
}
</style>
