import { createRouter, createWebHistory } from 'vue-router'
import { getToken, getUser } from '../utils/auth'
import { ElMessage } from 'element-plus'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/test',
    },
    {
      path: '/test',
      name: 'TestMainLayout',
      component: () => import('../views/TestMainLayout.vue'),
      meta: { requiresAuth: false },
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
    // 主布局路由
    {
      path: '/main',
      name: 'MainLayout',
      component: () => import('../components/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'MainLayoutDefault',
          redirect: 'queue-list',
        },
        {
          path: 'queue-list',
          name: 'QueueList',
          component: () => import('../components/QueueList.vue'),
        },
        {
          path: 'data-management',
          name: 'DataManagement',
          component: () => import('../views/SearchTasks.vue'),
        },
        {
          path: 'permission-list',
          name: 'PermissionList',
          component: () => import('../views/admin/UserManage.vue'),
        },
        {
          path: 'efficiency-stats',
          name: 'EfficiencyStats',
          component: () => import('../views/admin/Statistics.vue'),
        },
        {
          path: 'user-management',
          name: 'UserManagement',
          component: () => import('../views/admin/UserManage.vue'),
        },
        {
          path: 'history-announcements',
          name: 'HistoryAnnouncements',
          component: () => import('../views/HistoryAnnouncements.vue'),
        },
        {
          path: 'rule-documentation',
          name: 'RuleDocumentation',
          component: () => import('../views/admin/ModerationRules.vue'),
        },
      ],
    },
    // 保留原有的独立路由用于兼容
    {
      path: '/reviewer',
      redirect: '/main/queue-list',
      meta: { requiresAuth: true, role: 'reviewer' },
    },
    {
      path: '/reviewer/dashboard',
      name: 'ReviewerDashboard',
      component: () => import('../views/reviewer/Dashboard.vue'),
      meta: { requiresAuth: true, role: 'reviewer' },
    },
    {
      path: '/reviewer/search',
      name: 'ReviewerSearch',
      component: () => import('../views/SearchTasks.vue'),
      meta: { requiresAuth: true, role: 'reviewer' },
    },
    {
      path: '/reviewer/second-review',
      name: 'SecondReviewDashboard',
      component: () => import('../views/reviewer/SecondReviewDashboard.vue'),
      meta: { requiresAuth: true, role: 'reviewer' },
    },
    {
      path: '/reviewer/quality-check',
      name: 'QualityCheckDashboard',
      component: () => import('../views/reviewer/QualityCheckDashboard.vue'),
      meta: { requiresAuth: true, role: 'reviewer' },
    },
    {
      path: '/admin',
      redirect: '/main/queue-list',
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
    {
      path: '/admin/search',
      name: 'AdminSearch',
      component: () => import('../views/SearchTasks.vue'),
      meta: { requiresAuth: true, role: 'admin' },
    },
    {
      path: '/admin/moderation-rules',
      name: 'ModerationRules',
      component: () => import('../views/admin/ModerationRules.vue'),
      meta: { requiresAuth: true, role: 'admin' },
    },
    {
      path: '/admin/queue-manage',
      name: 'QueueManage',
      component: () => import('../views/admin/QueueManage.vue'),
      meta: { requiresAuth: true, role: 'admin' },
    },
  ],
})

// Navigation guard
router.beforeEach((to, _from, next) => {
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

  // Redirect to main layout if already logged in (regardless of role)
  if ((to.path === '/login' || to.path === '/register') && token && user) {
    // All users go to the same main layout
    next('/main/queue-list')
    return
  }

  next()
})

export default router

