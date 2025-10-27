<template>
  <div class="test-main-layout">
    <h1>测试主布局</h1>
    <p>这是一个测试页面，用于验证主布局是否正常工作。</p>
    <el-button type="primary" @click="goToMain">进入主布局</el-button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const goToMain = () => {
  router.push('/main/queue-list')
}
</script>

<style scoped>
.test-main-layout {
  padding: 40px;
  text-align: center;
}

h1 {
  color: #333;
  margin-bottom: 20px;
}

p {
  color: #666;
  margin-bottom: 30px;
}
</style>
