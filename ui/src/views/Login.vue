<template>
  <div class="login-container">
    <n-card class="login-card">
      <template #header>
        <div class="login-header">
          <n-icon size="32" color="#2080f0">
            <LogInOutline />
          </n-icon>
          <h1>{{ isRegisterMode ? '注册账户' : '登录 Claude Duck' }}</h1>
        </div>
      </template>

      <n-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-placement="left"
        label-width="auto"
        require-mark-placement="right-hanging"
        size="large"
      >
        <n-form-item v-if="isRegisterMode" label="用户名" path="username">
          <n-input
            v-model:value="formData.username"
            placeholder="请输入用户名"
            :disabled="loading"
          />
        </n-form-item>
        
        <n-form-item label="邮箱" path="email">
          <n-input
            v-model:value="formData.email"
            placeholder="请输入邮箱地址"
            :disabled="loading"
          />
        </n-form-item>
        
        <n-form-item label="密码" path="password">
          <n-input
            v-model:value="formData.password"
            type="password"
            placeholder="请输入密码"
            show-password-on="click"
            :disabled="loading"
          />
        </n-form-item>
        
        <n-form-item v-if="isRegisterMode" label="确认密码" path="confirmPassword">
          <n-input
            v-model:value="formData.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            show-password-on="click"
            :disabled="loading"
          />
        </n-form-item>
      </n-form>

      <n-space vertical :size="16">
        <n-button
          type="primary"
          size="large"
          block
          :loading="loading"
          @click="handleSubmit"
        >
          {{ isRegisterMode ? '注册' : '登录' }}
        </n-button>
        
        <n-button
          text
          block
          @click="toggleMode"
          :disabled="loading"
        >
          {{ isRegisterMode ? '已有账户？点击登录' : '没有账户？点击注册' }}
        </n-button>
      </n-space>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import {
  NCard, NIcon, NForm, NFormItem, NInput, NButton, NSpace, useMessage
} from 'naive-ui'
import { LogInOutline } from '@vicons/ionicons5'
import type { FormInst, FormRules } from 'naive-ui'
import { login, register } from '../api/auth'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()

// 表单相关
const formRef = ref<FormInst | null>(null)
const loading = ref(false)
const isRegisterMode = ref(false)

// 表单数据
const formData = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

// 表单验证规则
const formRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: ['input', 'blur'] },
    { min: 3, max: 20, message: '用户名长度应为3-20个字符', trigger: ['input', 'blur'] }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: ['input', 'blur'] },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: ['input', 'blur'] }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: ['input', 'blur'] },
    { min: 6, message: '密码长度不能少于6个字符', trigger: ['input', 'blur'] }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: ['input', 'blur'] },
    {
      validator: (_, value) => {
        return value === formData.password
      },
      message: '两次输入的密码不一致',
      trigger: ['input', 'blur']
    }
  ]
}

// 切换登录/注册模式
function toggleMode() {
  isRegisterMode.value = !isRegisterMode.value
  // 清空表单
  formData.username = ''
  formData.email = ''
  formData.password = ''
  formData.confirmPassword = ''
}

// 处理表单提交
async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch (error) {
    console.log('表单验证失败:', error)
    return
  }

  loading.value = true

  try {
    if (isRegisterMode.value) {
      // 注册
      const response = await register({
        username: formData.username,
        email: formData.email,
        password: formData.password
      })

      if (response.success && response.token && response.user) {
        message.success(response.message)
        authStore.setAuth(response.user, response.token)
        router.push('/')
      } else {
        message.error(response.message || '注册失败')
      }
    } else {
      // 登录
      const response = await login({
        email: formData.email,
        password: formData.password
      })

      if (response.success && response.token && response.user) {
        message.success(response.message)
        authStore.setAuth(response.user, response.token)
        
        // 重定向到原来要访问的页面，或者首页
        const redirect = router.currentRoute.value.query.redirect as string
        router.push(redirect || '/')
      } else {
        message.error(response.message || '登录失败')
      }
    }
  } catch (error: any) {
    console.error('认证失败:', error)
    
    // 处理API错误响应
    if (error.response?.data?.message) {
      message.error(error.response.data.message)
    } else if (error.response?.data?.error) {
      message.error(error.response.data.error)
    } else {
      message.error(isRegisterMode.value ? '注册失败，请稍后重试' : '登录失败，请稍后重试')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
}

.login-header {
  display: flex;
  align-items: center;
  gap: 12px;
  justify-content: center;
}

.login-header h1 {
  margin: 0;
  font-size: 24px;
  color: #333;
}
</style> 