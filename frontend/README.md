# 评论审核平台 - 前端

基于 Vue 3 + TypeScript + Vite + Element Plus 构建的评论审核系统前端。

## 技术栈

- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **构建工具**: Vite
- **UI 库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router
- **HTTP 客户端**: Axios

## 项目结构

```
frontend/
├── src/
│   ├── api/              # API 接口封装
│   │   ├── request.ts    # Axios 配置
│   │   ├── auth.ts       # 认证接口
│   │   ├── task.ts       # 任务接口
│   │   └── admin.ts      # 管理员接口
│   ├── assets/           # 静态资源
│   ├── components/       # 公共组件
│   ├── router/           # 路由配置
│   │   └── index.ts
│   ├── stores/           # Pinia 状态管理
│   │   ├── user.ts       # 用户状态
│   │   └── task.ts       # 任务状态
│   ├── types/            # TypeScript 类型定义
│   │   └── index.ts
│   ├── utils/            # 工具函数
│   │   ├── auth.ts       # Token 管理
│   │   └── format.ts     # 格式化工具
│   ├── views/            # 页面组件
│   │   ├── Login.vue     # 登录页
│   │   ├── Register.vue  # 注册页
│   │   ├── reviewer/     # 审核员页面
│   │   │   └── Dashboard.vue
│   │   └── admin/        # 管理员页面
│   │       ├── Dashboard.vue
│   │       ├── UserManage.vue
│   │       ├── Statistics.vue
│   │       └── TagManage.vue
│   ├── App.vue
│   └── main.ts
├── vite.config.ts        # Vite 配置
└── package.json
```

## 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

前端将在 `http://localhost:3000` 启动。

### 3. 构建生产版本

```bash
npm run build
```

构建产物将输出到 `dist/` 目录。

## 功能模块

### 1. 认证系统
- 用户登录
- 审核员注册
- JWT Token 认证
- 自动跳转和权限控制

### 2. 审核员工作台
- 领取审核任务（一次 20 条）
- 查看我的任务列表
- 单个任务审核
  - 通过/不通过选择
  - 违规标签多选
  - 不通过原因填写
- 批量提交审核结果
- 实时统计展示

### 3. 管理员控制台
- **总览页面**
  - 统计卡片（总任务数、完成数、通过率等）
  - 任务分布进度条
  
- **用户管理**
  - 查看待审批用户
  - 审批/拒绝操作
  
- **统计分析**
  - 违规类型分布
  - 审核员绩效排行榜
  
- **标签管理**
  - 标签 CRUD 操作
  - 启用/禁用标签

## 路由说明

| 路径 | 角色要求 | 说明 |
|------|---------|------|
| `/login` | 无 | 登录页 |
| `/register` | 无 | 注册页 |
| `/reviewer/dashboard` | reviewer | 审核员工作台 |
| `/admin/dashboard` | admin | 管理员总览 |
| `/admin/users` | admin | 用户管理 |
| `/admin/statistics` | admin | 统计分析 |
| `/admin/tags` | admin | 标签管理 |

## API 配置

前端通过 Vite 代理与后端通信：

```typescript
// vite.config.ts
server: {
  port: 3000,
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
}
```

确保后端服务运行在 `http://localhost:8080`。

## 环境要求

- Node.js >= 20.15.0
- npm >= 10.7.0

## 默认账号

### 管理员
- 用户名: `admin`
- 密码: `admin123`

### 审核员
需要先注册，然后由管理员审批通过。

## 开发说明

### Element Plus 自动导入

项目已配置 Element Plus 组件和 API 的自动导入，无需手动引入：

```vue
<template>
  <!-- 直接使用，无需 import -->
  <el-button type="primary">按钮</el-button>
</template>

<script setup>
// ElMessage 等 API 也会自动导入
ElMessage.success('操作成功')
</script>
```

### 状态管理

使用 Pinia 进行状态管理：

```typescript
// 使用 user store
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const isAdmin = userStore.isAdmin()
```

### API 调用

统一通过封装的 API 函数调用：

```typescript
import { login } from '@/api/auth'

const handleLogin = async () => {
  const res = await login(username, password)
  // 处理响应...
}
```

## 常见问题

### Q: 页面显示 404
A: 确保后端服务已启动，检查 Vite 代理配置是否正确。

### Q: 登录后自动跳转到登录页
A: 检查 Token 是否正确保存，查看浏览器控制台的错误信息。

### Q: 样式显示异常
A: 清除浏览器缓存，或尝试 `npm run dev` 重新启动开发服务器。

## License

MIT
