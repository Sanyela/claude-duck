<template>
  <div class="oauth-container">
    <n-card class="oauth-card">
      <template #header>
        <div class="oauth-header">
          <n-icon size="32" color="#2080f0">
            <ShieldCheckmarkOutline />
          </n-icon>
          <h1>Claude Duck 授权</h1>
        </div>
      </template>

      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <n-spin size="large" />
        <p>正在处理授权请求...</p>
      </div>

      <!-- 授权表单 -->
      <div v-else-if="!authorized" class="auth-form">
        <n-alert type="info" style="margin-bottom: 16px;">
          应用请求访问您的 Claude Duck 账户
        </n-alert>

        <div class="app-info">
          <h3>应用信息</h3>
          <n-descriptions label-placement="left" :column="1" bordered>
            <n-descriptions-item label="客户端 ID">
              {{ clientId }}
            </n-descriptions-item>
            <n-descriptions-item label="回调地址">
              {{ redirectUri }}
            </n-descriptions-item>
            <n-descriptions-item label="请求权限">
              <n-tag type="success">读取用户信息</n-tag>
              <n-tag type="success" style="margin-left: 8px;">访问 API</n-tag>
            </n-descriptions-item>
          </n-descriptions>
        </div>

        <div class="auth-actions">
          <n-space>
            <n-button type="primary" size="large" @click="handleAuthorize" :loading="authorizing">
              <template #icon>
                <n-icon><CheckmarkOutline /></n-icon>
              </template>
              授权访问
            </n-button>
            <n-button size="large" @click="handleDeny">
              <template #icon>
                <n-icon><CloseOutline /></n-icon>
              </template>
              拒绝
            </n-button>
          </n-space>
        </div>
      </div>

      <!-- 手动模式 - 显示授权码 -->
      <div v-else-if="isDeviceFlow && authCode" class="device-flow">
        <n-result status="success" title="授权成功" description="请复制下面的授权码到您的终端">
          <template #footer>
            <div class="auth-code-section">
              <n-input
                v-model:value="authCode"
                readonly
                size="large"
                style="font-family: monospace; font-size: 18px; text-align: center; letter-spacing: 2px;"
              >
                <template #suffix>
                  <n-button text @click="copyAuthCode" :loading="copying">
                    <template #icon>
                      <n-icon><CopyOutline /></n-icon>
                    </template>
                  </n-button>
                </template>
              </n-input>
              <p style="margin-top: 12px; color: #666;">
                授权码已生成，请将其粘贴到终端并按回车键
              </p>
            </div>
          </template>
        </n-result>
      </div>

      <!-- 自动模式 - 重定向中 -->
      <div v-else-if="!isDeviceFlow && authorized" class="auto-redirect">
        <n-result status="success" title="授权成功" description="正在重定向...">
          <template #footer>
            <n-spin size="large" />
          </template>
        </n-result>
      </div>

      <!-- 错误状态 -->
      <div v-if="error" class="error-state">
        <n-result status="error" :title="error.title" :description="error.message">
          <template #footer>
            <n-button @click="resetAuth">重试</n-button>
          </template>
        </n-result>
      </div>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import {
  NCard, NIcon, NSpin, NAlert, NDescriptions, NDescriptionsItem,
  NTag, NSpace, NButton, NResult, NInput, useMessage
} from 'naive-ui'
import {
  ShieldCheckmarkOutline,
  CheckmarkOutline,
  CloseOutline,
  CopyOutline
} from '@vicons/ionicons5'

const route = useRoute()
const message = useMessage()

// 状态管理
const loading = ref(true)
const authorized = ref(false)
const authorizing = ref(false)
const copying = ref(false)
const error = ref<{title: string, message: string} | null>(null)

// 授权参数
const clientId = ref('')
const redirectUri = ref('')
const state = ref('')
const isDeviceFlow = ref(false)
const authCode = ref('')

onMounted(() => {
  parseAuthParams()
  setTimeout(() => {
    loading.value = false
  }, 500)
})

// 解析授权参数
function parseAuthParams() {
  try {
    clientId.value = route.query.client_id as string || ''
    redirectUri.value = route.query.redirect_uri as string || ''
    state.value = route.query.state as string || ''
    isDeviceFlow.value = route.query.device_flow === 'true'

    if (!clientId.value || !redirectUri.value || !state.value) {
      throw new Error('缺少必要的授权参数')
    }
  } catch (err) {
    error.value = {
      title: '参数错误',
      message: '授权请求参数不完整或无效'
    }
  }
}

// 处理授权
async function handleAuthorize() {
  authorizing.value = true
  
  try {
    await handleProductionAuthorize()
  } catch (err) {
    error.value = {
      title: '授权失败',
      message: '处理授权请求时发生错误，请重试'
    }
  } finally {
    authorizing.value = false
  }
}

// 生产模式授权处理
async function handleProductionAuthorize() {
  // 调用真实的后端OAuth授权接口
  const response = await fetch('/api/sso/authorize', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      client_id: clientId.value,
      redirect_uri: redirectUri.value,
      state: state.value,
      device_flow: isDeviceFlow.value
    })
  })

  if (!response.ok) {
    throw new Error('OAuth授权失败')
  }

  const result = await response.json()

  if (isDeviceFlow.value) {
    // 手动模式：显示真实的授权码
    authCode.value = result.code
    authorized.value = true
    message.success('授权码已生成')
  } else {
    // 自动模式：重定向到回调地址
    authorized.value = true
    message.success('授权成功，正在重定向...')
    
    setTimeout(() => {
      const callbackUrl = new URL(redirectUri.value)
      callbackUrl.searchParams.set('token', result.token)
      callbackUrl.searchParams.set('state', state.value)
      
      window.location.href = callbackUrl.toString()
    }, 1500)
  }
}

// 处理拒绝
function handleDeny() {
  const callbackUrl = new URL(redirectUri.value)
  callbackUrl.searchParams.set('error', 'access_denied')
  callbackUrl.searchParams.set('state', state.value)
  
  window.location.href = callbackUrl.toString()
}

// 复制授权码
async function copyAuthCode() {
  copying.value = true
  try {
    await navigator.clipboard.writeText(authCode.value)
    message.success('授权码已复制到剪贴板')
  } catch (err) {
    // 降级方案：选择文本
    const input = document.querySelector('input[readonly]') as HTMLInputElement
    if (input) {
      input.select()
      document.execCommand('copy')
      message.success('授权码已复制')
    } else {
      message.error('复制失败，请手动选择并复制')
    }
  } finally {
    copying.value = false
  }
}

// 重置授权状态
function resetAuth() {
  loading.value = false
  authorized.value = false
  authorizing.value = false
  error.value = null
  authCode.value = ''
}
</script>

<style scoped>
.oauth-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.oauth-card {
  width: 100%;
  max-width: 500px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
}

.oauth-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.oauth-header h1 {
  margin: 0;
  font-size: 24px;
  color: #333;
}

.loading-state,
.auth-form,
.device-flow,
.auto-redirect,
.error-state {
  text-align: center;
  padding: 20px 0;
}

.loading-state p {
  margin-top: 16px;
  color: #666;
}

.app-info {
  margin: 24px 0;
  text-align: left;
}

.app-info h3 {
  margin-bottom: 16px;
  color: #333;
}

.auth-actions {
  margin-top: 32px;
}

.auth-code-section {
  margin-top: 20px;
}

.auth-code-section p {
  margin: 0;
  font-size: 14px;
}
</style> 