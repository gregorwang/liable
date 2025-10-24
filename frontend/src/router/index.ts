import { createRouter, createWebHistory } from 'vue-router'
import { getToken, getUser } from '../utils/auth'
import { ElMessage } from 'element-plus'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/login',
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('../views/Register.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/reviewer',
      redirect: '/reviewer/dashboard',
      meta: { requiresAuth: true, role: 'reviewer' },
    },
    {
      path: '/reviewer/dashboard',
      name: 'ReviewerDashboard',
      component: () => import('../views/reviewer/Dashboard.vue'),
      meta: { requiresAuth: true, role: 'reviewer' },
    },
    {
      path: '/admin',
      redirect: '/admin/dashboard',
      meta: { requiresAuth: true, role: 'admin' },
    },
    {
      path: '/admin/dashboard',
      name: 'AdminDashboard',
      component: () => import('../views/admin/Dashboard.vue'),
      meta: { requiresAuth: true, role: 'admin' },
    },
    {
      path: '/admin/users',
      name: 'UserManage',
      component: () => import('../views/admin/UserManage.vue'),
      meta: { requiresAuth: true, role: 'admin' },
    },
    {
      path: '/admin/statistics',
      name: 'Statistics',
      component: () => import('../views/admin/Statistics.vue'),
      meta: { requiresAuth: true, role: 'admin' },
    },
    {
      path: '/admin/tags',
      name: 'TagManage',
      component: () => import('../views/admin/TagManage.vue'),
      meta: { requiresAuth: true, role: 'admin' },
    },
  ],
})

// Navigation guard
router.beforeEach((to, from, next) => {
  const token = getToken()
  const user = getUser()

  // Check if route requires authentication
  if (to.meta.requiresAuth) {
    if (!token) {
      ElMessage.warning('请先登录')
      next('/login')
      return
    }

    // Check role
    if (to.meta.role && user?.role !== to.meta.role) {
      ElMessage.error('没有权限访问该页面')
      if (user?.role === 'admin') {
        next('/admin/dashboard')
      } else if (user?.role === 'reviewer') {
        next('/reviewer/dashboard')
      } else {
        next('/login')
      }
      return
    }
  }

  // Redirect to dashboard if already logged in
  if ((to.path === '/login' || to.path === '/register') && token && user) {
    if (user.role === 'admin') {
      next('/admin/dashboard')
    } else if (user.role === 'reviewer') {
      next('/reviewer/dashboard')
    } else {
      next()
    }
    return
  }

  next()
})

export default router

