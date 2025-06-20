<template>
  <div class="oauth-callback-container">
    <n-card class="callback-card">
      <n-spin v-if="loading" size="large" description="处理授权回调中..." />
      
      <n-result v-else-if="success" status="success" title="授权成功" description="正在返回应用...">
        <template #footer>
          <n-button type="primary" @click="closeWindow">关闭窗口</n-button>
        </template>
      </n-result>
      
      <n-result v-else status="error" title="授权失败" :description="errorMessage">
        <template #footer>
          <n-button @click="closeWindow">关闭窗口</n-button>
        </template>
      </n-result>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NSpin, NResult, NButton, useMessage } from 'naive-ui'

const route = useRoute()
const message = useMessage()

const loading = ref(true)
const success = ref(false)
const errorMessage = ref('')

onMounted(() => {
  handleCallback()
})

async function handleCallback() {
  try {
    // 获取URL参数
    const token = route.query.token as string
    const state = route.query.state as string
    const error = route.query.error as string

    if (error) {
      // 处理错误情况
      errorMessage.value = error === 'access_denied' ? '用户拒绝授权' : '授权过程中出现错误'
      success.value = false
    } else if (token && state) {
      // 成功获取token
      // 通过postMessage发送给父窗口
      if (window.opener) {
        window.opener.postMessage({
          type: 'oauth_callback',
          success: true,
          data: {
            token,
            state
          }
        }, window.location.origin)
      }

      // 如果是在本地应用中打开的，保存token
      if (window.location.port === '59771') {
        localStorage.setItem('oauth_token', token)
        localStorage.setItem('oauth_state', state)
      }

      success.value = true
      message.success('授权成功，token已保存')

      // 自动关闭窗口
      setTimeout(() => {
        closeWindow()
      }, 2000)
    } else {
      errorMessage.value = '回调参数缺失'
      success.value = false
    }
  } catch (err) {
    errorMessage.value = '处理回调时发生错误'
    success.value = false
  } finally {
    loading.value = false
  }
}

function closeWindow() {
  // 尝试关闭窗口
  if (window.opener) {
    window.close()
  } else {
    // 如果无法关闭，显示提示
    message.info('请手动关闭此窗口')
  }
}
</script>

<style scoped>
.oauth-callback-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
  padding: 20px;
}

.callback-card {
  width: 100%;
  max-width: 400px;
  text-align: center;
}
</style>