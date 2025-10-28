<template>
  <div class="register-container">
    <el-card class="register-card">
      <template #header>
        <div class="card-header">
          <h2>审核员注册</h2>
          <p>注册后需等待管理员审批</p>
        </div>
      </template>
      
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        size="large"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
          />
        </el-form-item>
        
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            show-password
            @keyup.enter="handleRegister"
          />
        </el-form-item>
        
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            style="width: 100%"
            @click="handleRegister"
          >
            注册
          </el-button>
        </el-form-item>
        
        <el-form-item>
          <el-button
            text
            style="width: 100%"
            @click="goToLogin"
          >
            已有账号？立即登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { register } from '../api/auth'

const router = useRouter()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
})

const validateConfirmPassword = (_rule: any, value: any, callback: any) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, message: '用户名至少3位', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

const handleRegister = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    loading.value = true
    try {
      const res = await register(form.username, form.password)
      ElMessage.success(res.message || '注册成功，请等待管理员审批')
      setTimeout(() => {
        router.push('/login')
      }, 1500)
    } catch (error) {
      console.error('Register failed:', error)
    } finally {
      loading.value = false
    }
  })
}

const goToLogin = () => {
  router.push('/login')
}
</script>

<style scoped>
/* ============================================
   注册页面样式（复用登录页面样式）
   ============================================ */
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  min-height: 100dvh;
  padding: var(--spacing-4);
  background: linear-gradient(135deg, 
    var(--color-accent-main) 0%,
    var(--color-accent-pro-dark) 50%, 
    var(--color-accent-pro) 100%
  );
  position: relative;
  overflow: hidden;
}

/* 背景装饰 */
.register-container::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, 
    rgba(255, 255, 255, 0.1) 1px, 
    transparent 1px
  );
  background-size: 50px 50px;
  animation: backgroundMove 60s linear infinite;
  pointer-events: none;
}

@keyframes backgroundMove {
  0% {
    transform: translate(0, 0);
  }
  100% {
    transform: translate(50px, 50px);
  }
}

/* ============================================
   注册卡片
   ============================================ */
.register-card {
  width: 100%;
  max-width: 460px;
  box-shadow: var(--shadow-2xl);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  position: relative;
  z-index: 1;
}

.register-card :deep(.el-card__header) {
  background: linear-gradient(to bottom, 
    var(--color-bg-100), 
    var(--color-bg-000)
  );
  padding: var(--spacing-8) var(--spacing-6);
  border-bottom: 2px solid var(--color-accent-main);
}

.register-card :deep(.el-card__body) {
  padding: var(--spacing-8) var(--spacing-6);
}

/* ============================================
   卡片头部
   ============================================ */
.card-header {
  text-align: center;
}

.card-header h2 {
  margin: 0 0 var(--spacing-3) 0;
  color: var(--color-text-000);
  font-size: var(--text-3xl);
  font-weight: 700;
  letter-spacing: var(--tracking-tight);
  background: linear-gradient(135deg, 
    var(--color-accent-main), 
    var(--color-accent-pro)
  );
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.card-header p {
  margin: 0;
  color: var(--color-text-300);
  font-size: var(--text-sm);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
}

/* ============================================
   表单样式增强
   ============================================ */
.register-card :deep(.el-form-item) {
  margin-bottom: var(--spacing-5);
}

.register-card :deep(.el-form-item__label) {
  font-weight: 600;
  color: var(--color-text-100);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  margin-bottom: var(--spacing-2);
}

.register-card :deep(.el-input__wrapper) {
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-base);
}

.register-card :deep(.el-input__wrapper:hover) {
  box-shadow: var(--shadow-md);
}

.register-card :deep(.el-button--primary) {
  height: 48px;
  font-size: var(--text-base);
  font-weight: 600;
  letter-spacing: var(--tracking-wider);
  box-shadow: var(--shadow-md);
  transition: all var(--transition-base);
}

.register-card :deep(.el-button--primary:hover) {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

.register-card :deep(.el-button--primary:active) {
  transform: translateY(0);
}

.register-card :deep(.el-button.is-text) {
  color: var(--color-text-300);
  font-size: var(--text-sm);
  transition: color var(--transition-fast);
}

.register-card :deep(.el-button.is-text:hover) {
  color: var(--color-accent-main);
}

/* ============================================
   响应式设计
   ============================================ */
@media (max-width: 480px) {
  .register-card {
    max-width: 100%;
  }

  .register-card :deep(.el-card__header) {
    padding: var(--spacing-6) var(--spacing-4);
  }

  .register-card :deep(.el-card__body) {
    padding: var(--spacing-6) var(--spacing-4);
  }

  .card-header h2 {
    font-size: var(--text-2xl);
  }
}
</style>

