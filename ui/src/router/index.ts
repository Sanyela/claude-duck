import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import MainLayout from '../layouts/MainLayout.vue'
import Dashboard from '../views/Dashboard.vue'
import Subscription from '../views/Subscription.vue'
import Credits from '../views/Credits.vue'
import Resources from '../views/Resources.vue'
import Settings from '../views/Settings.vue'
import OAuth from '../views/OAuth.vue'
import OAuthCallback from '../views/OAuthCallback.vue'
import OAuthTest from '../views/OAuthTest.vue'
import Login from '../views/Login.vue'

const routes = [
  // 登录页面（无需认证）
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: MainLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: Dashboard,
      },
      {
        path: 'subscription',
        name: 'Subscription',
        component: Subscription,
      },
      {
        path: 'credits',
        name: 'Credits',
        component: Credits,
      },
      {
        path: 'resources',
        name: 'Resources',
        component: Resources,
      },
      {
        path: 'settings',
        name: 'Settings',
        component: Settings,
      },
      {
        path: 'oauth-test',
        name: 'OAuthTest',
        component: OAuthTest,
      },
    ],
  },
  // OAuth 授权页面（独立布局，需要认证）
  {
    path: '/oauth/authorize',
    name: 'OAuth',
    component: OAuth,
    meta: { requiresAuth: true }
  },
  // OAuth 回调页面（无需认证）
  {
    path: '/oauth/callback',
    name: 'OAuthCallback',
    component: OAuthCallback,
    meta: { requiresAuth: false }
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

// 路由守卫
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()
  
  // 检查路由是否需要认证
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth !== false)
  
  if (requiresAuth) {
    // 需要认证的路由
    const isAuthenticated = await authStore.checkAuth()
    
    if (!isAuthenticated) {
      // 未认证，重定向到登录页面
      next({
        path: '/login',
        query: { redirect: to.fullPath }
      })
    } else {
      next()
    }
  } else {
    // 不需要认证的路由
    if (to.path === '/login' && authStore.isAuthenticated) {
      // 已登录用户访问登录页面，重定向到首页
      next('/')
    } else {
      next()
    }
  }
})

export default router 