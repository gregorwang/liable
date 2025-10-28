<template>
  <div class="login-container">
    <!-- 左侧引言区域 -->
    <div class="left-section">
      <div class="quote-content">
        <div class="quote-text">
          鉴于对良知的无视与侮蔑亦会发展为玷污社区的暴行，我们以技术守护审慎的判断，为一个人人享有言论自由并免予恐惧的世界而努力。
        </div>
        <div class="quote-decoration">
          <div class="decoration-line"></div>
          <div class="decoration-dot"></div>
          <div class="decoration-line"></div>
        </div>
      </div>
    </div>
    
    <!-- 右侧登录区域 -->
    <div class="right-section">
      <el-card class="login-card">
        <template #header>
          <div class="card-header">
            <h2>评论审核平台</h2>
            <p>登录</p>
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
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          
          <el-form-item label="密码" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              show-password
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          
          <el-form-item>
            <el-button
              type="primary"
              :loading="loading"
              style="width: 100%"
              @click="handleLogin"
            >
              登录
            </el-button>
          </el-form-item>
          
          <el-form-item>
            <el-button
              text
              style="width: 100%"
              @click="goToRegister"
            >
              还没有账号？立即注册
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
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' },
  ],
}

const handleLogin = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    loading.value = true
    try {
      await userStore.login(form.username, form.password)
      ElMessage.success('登录成功')
      
      // Redirect to main layout for all users
      router.push('/main/queue-list')
    } catch (error) {
      console.error('Login failed:', error)
    } finally {
      loading.value = false
    }
  })
}

const goToRegister = () => {
  router.push('/register')
}
</script>

<style scoped>
/* ============================================
   登录页面样式 - 紧密布局
   ============================================ */
.login-container {
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
   右侧登录区域
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

.login-card {
  width: 400px;
  box-shadow: none;
  border: none;
  border-radius: var(--radius-lg);
  background: var(--color-bg-100);
}

.login-card :deep(.el-card__header) {
  background: var(--color-bg-100);
  padding: var(--spacing-8) var(--spacing-6);
  border-bottom: none;
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
}

.login-card :deep(.el-card__body) {
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
.login-card :deep(.el-form-item) {
  margin-bottom: var(--spacing-6);
}

.login-card :deep(.el-form-item__label) {
  font-family: "Zhi Mang Xing", cursive;
  font-weight: 600;
  color: var(--color-text-100);
  font-size: var(--text-sm);
  letter-spacing: var(--tracking-wide);
  margin-bottom: var(--spacing-2);
}

.login-card :deep(.el-input__wrapper) {
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-base);
  border-radius: var(--radius-md);
}

.login-card :deep(.el-input__wrapper:hover) {
  box-shadow: var(--shadow-md);
  border-color: var(--color-accent-main);
}

.login-card :deep(.el-input__wrapper.is-focus) {
  box-shadow: var(--shadow-md);
  border-color: var(--color-accent-main);
}

.login-card :deep(.el-button--primary) {
  font-family: "Zhi Mang Xing", cursive;
  height: 48px;
  font-size: var(--text-base);
  font-weight: 600;
  letter-spacing: var(--tracking-wider);
  box-shadow: var(--shadow-md);
  transition: all var(--transition-base);
  border-radius: var(--radius-md);
}

.login-card :deep(.el-button--primary:hover) {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

.login-card :deep(.el-button--primary:active) {
  transform: translateY(0);
}

.login-card :deep(.el-button.is-text) {
  font-family: "Zhi Mang Xing", cursive;
  color: var(--color-text-300);
  font-size: var(--text-sm);
  transition: color var(--transition-fast);
}

.login-card :deep(.el-button.is-text:hover) {
  color: var(--color-accent-main);
}

/* ============================================
   响应式设计
   ============================================ */
@media (max-width: 1200px) {
  .login-container {
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
  
  .login-card {
    width: 360px;
  }
}

@media (max-width: 1024px) {
  .login-container {
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
  
  .login-card {
    width: 400px;
  }
}

@media (max-width: 768px) {
  .login-container {
    padding: var(--spacing-3);
  }
  
  .left-section {
    min-height: 35vh;
    margin-bottom: var(--spacing-2);
  }
  
  .quote-text {
    font-size: 32px;
  }
  
  .login-card {
    width: 100%;
    max-width: 400px;
  }
  
  .login-card :deep(.el-card__header) {
    padding: var(--spacing-6) var(--spacing-4);
  }

  .login-card :deep(.el-card__body) {
    padding: var(--spacing-6) var(--spacing-4);
  }

  .card-header h2 {
    font-size: var(--text-2xl);
  }
}

@media (max-width: 480px) {
  .login-container {
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

