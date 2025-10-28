<template>
  <div class="register-container">
    <!-- 左侧引言区域 -->
    <div class="left-section">
      <div class="quote-content">
        <div class="quote-text">
          我们相信，每一个声音都值得被倾听，每一次表达都应当被尊重。通过严谨的审核流程，我们共同构建一个更加和谐、理性的交流空间。
        </div>
        <div class="quote-decoration">
          <div class="decoration-line"></div>
          <div class="decoration-dot"></div>
          <div class="decoration-line"></div>
        </div>
      </div>
    </div>
    
    <!-- 右侧注册区域 -->
    <div class="right-section">
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
   注册页面样式 - 紧密布局（与登录页面保持一致）
   ============================================ */
.register-container {
  display: flex;
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--color-bg-100);
  position: relative;
  overflow: hidden;
  padding: var(--spacing-8);
  align-items: center;
  justify-content: center;
}

/* ============================================
   左侧引言区域
   ============================================ */
.left-section {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-100);
  position: relative;
  padding: 0;
  margin-right: var(--spacing-3);
}

/* 背景装饰 */
.left-section::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, 
    rgba(0, 0, 0, 0.05) 1px, 
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

.quote-content {
  max-width: 400px;
  text-align: center;
  position: relative;
  z-index: 1;
}

.quote-text {
  font-family: "Zhi Mang Xing", cursive;
  font-size: 48px;
  font-weight: 400;
  line-height: var(--leading-loose);
  color: var(--color-text-000);
  margin-bottom: var(--spacing-6);
  letter-spacing: var(--tracking-wide);
}

.quote-decoration {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-4);
}

.decoration-line {
  width: 60px;
  height: 2px;
  background: var(--color-text-300);
  border-radius: 1px;
}

.decoration-dot {
  width: 8px;
  height: 8px;
  background: var(--color-text-200);
  border-radius: 50%;
}

/* ============================================
   右侧注册区域
   ============================================ */
.right-section {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-100);
  position: relative;
  margin-left: var(--spacing-3);
}

.register-card {
  width: 400px;
  box-shadow: none;
  border: none;
  border-radius: var(--radius-lg);
  background: var(--color-bg-100);
}

.register-card :deep(.el-card__header) {
  background: var(--color-bg-100);
  padding: var(--spacing-8) var(--spacing-6);
  border-bottom: none;
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
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
  font-family: "Zhi Mang Xing", cursive;
  font-size: var(--text-3xl);
  font-weight: 700;
  letter-spacing: var(--tracking-tight);
}

.card-header p {
  margin: 0;
  color: var(--color-text-300);
  font-family: "Zhi Mang Xing", cursive;
  font-size: var(--text-base);
  font-weight: 500;
  letter-spacing: var(--tracking-wide);
}

/* ============================================
   表单样式增强
   ============================================ */
.register-card :deep(.el-form-item) {
  margin-bottom: var(--spacing-6);
}

.register-card :deep(.el-form-item__label) {
  font-family: "Zhi Mang Xing", cursive;
  font-weight: 600;
  color: var(--color-text-100);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  margin-bottom: var(--spacing-2);
}

.register-card :deep(.el-input__wrapper) {
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-base);
  border-radius: var(--radius-md);
}

.register-card :deep(.el-input__wrapper:hover) {
  box-shadow: var(--shadow-md);
  border-color: var(--color-accent-main);
}

.register-card :deep(.el-input__wrapper.is-focus) {
  box-shadow: var(--shadow-md);
  border-color: var(--color-accent-main);
}

.register-card :deep(.el-button--primary) {
  font-family: "Zhi Mang Xing", cursive;
  height: 48px;
  font-size: var(--text-base);
  font-weight: 600;
  letter-spacing: var(--tracking-wider);
  box-shadow: var(--shadow-md);
  transition: all var(--transition-base);
  border-radius: var(--radius-md);
}

.register-card :deep(.el-button--primary:hover) {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

.register-card :deep(.el-button--primary:active) {
  transform: translateY(0);
}

.register-card :deep(.el-button.is-text) {
  font-family: "Zhi Mang Xing", cursive;
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
@media (max-width: 1200px) {
  .register-container {
    padding: var(--spacing-6);
  }
  
  .left-section {
    margin-right: var(--spacing-2);
  }
  
  .right-section {
    margin-left: var(--spacing-2);
  }
  
  .quote-text {
    font-size: 40px;
  }
  
  .register-card {
    width: 360px;
  }
}

@media (max-width: 1024px) {
  .register-container {
    flex-direction: column;
    padding: var(--spacing-4);
  }
  
  .left-section {
    flex: 0 0 auto;
    min-height: 40vh;
    margin-right: 0;
    margin-bottom: var(--spacing-3);
  }
  
  .right-section {
    flex: 0 0 auto;
    margin-left: 0;
  }
  
  .quote-text {
    font-size: 36px;
  }
  
  .register-card {
    width: 400px;
  }
}

@media (max-width: 768px) {
  .register-container {
    padding: var(--spacing-3);
  }
  
  .left-section {
    min-height: 35vh;
    margin-bottom: var(--spacing-2);
  }
  
  .quote-text {
    font-size: 32px;
  }
  
  .register-card {
    width: 100%;
    max-width: 400px;
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

@media (max-width: 480px) {
  .register-container {
    padding: var(--spacing-2);
  }
  
  .left-section {
    margin-bottom: var(--spacing-2);
  }
  
  .quote-text {
    font-size: 28px;
  }
  
  .quote-decoration {
    gap: var(--spacing-2);
  }
  
  .decoration-line {
    width: 40px;
  }
}
</style>

