import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserData } from '../api/auth'
import { getUserInfo, logout as apiLogout } from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const user = ref<UserData | null>(null)
  const token = ref<string | null>(localStorage.getItem('auth_token'))
  const isLoading = ref(false)

  // 计算属性
  const isAuthenticated = computed(() => !!token.value && !!user.value)

  // 设置用户信息和token
  function setAuth(userData: UserData, authToken: string) {
    user.value = userData
    token.value = authToken
    localStorage.setItem('auth_token', authToken)
    localStorage.setItem('user_data', JSON.stringify(userData))
  }

  // 清除认证信息
  function clearAuth() {
    user.value = null
    token.value = null
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user_data')
  }

  // 从localStorage恢复用户信息
  function restoreAuth() {
    const savedToken = localStorage.getItem('auth_token')
    const savedUser = localStorage.getItem('user_data')
    
    if (savedToken && savedUser) {
      try {
        token.value = savedToken
        user.value = JSON.parse(savedUser)
      } catch (error) {
        console.error('恢复用户信息失败:', error)
        clearAuth()
      }
    }
  }

  // 检查并验证token有效性
  async function checkAuth(): Promise<boolean> {
    if (!token.value) {
      return false
    }

    if (user.value) {
      return true
    }

    isLoading.value = true
    try {
      const response = await getUserInfo()
      user.value = response.user
      return true
    } catch (error) {
      console.error('验证用户信息失败:', error)
      clearAuth()
      return false
    } finally {
      isLoading.value = false
    }
  }

  // 登出
  async function logout(): Promise<void> {
    try {
      await apiLogout()
    } catch (error) {
      console.error('登出请求失败:', error)
    } finally {
      clearAuth()
    }
  }

  // 初始化时恢复认证状态
  restoreAuth()

  return {
    // 状态
    user,
    token,
    isLoading,
    
    // 计算属性
    isAuthenticated,
    
    // 方法
    setAuth,
    clearAuth,
    restoreAuth,
    checkAuth,
    logout
  }
}) 