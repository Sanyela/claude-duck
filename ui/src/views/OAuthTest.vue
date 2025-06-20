<template>
  <div>
    <n-page-header>
      <template #title>
        OAuth 测试工具
      </template>
      <template #subtitle>
        用于测试 Claude Duck OAuth 授权流程
      </template>
    </n-page-header>

    <n-grid cols="1 m:2" :x-gap="16" :y-gap="16" style="margin-top: 16px;">
      <!-- 自动模式测试 -->
      <n-gi>
        <n-card title="自动模式测试" hoverable>
          <template #header-extra>
            <n-tag type="success">推荐</n-tag>
          </template>
          
          <n-space vertical>
            <n-alert type="info">
              自动模式会打开授权页面，授权后自动重定向回回调地址
            </n-alert>
            
            <n-form :model="autoForm" label-placement="left" label-width="120px">
              <n-form-item label="Client ID">
                <n-input v-model:value="autoForm.clientId" placeholder="输入客户端ID" />
              </n-form-item>
              <n-form-item label="回调地址">
                <n-input v-model:value="autoForm.redirectUri" placeholder="http://localhost:3000/callback" />
              </n-form-item>
              <n-form-item label="State">
                <n-input v-model:value="autoForm.state" placeholder="随机状态值" />
                <template #feedback>
                  <n-button text @click="generateState('auto')">生成随机State</n-button>
                </template>
              </n-form-item>
            </n-form>
            
            <n-button type="primary" block @click="startAutoFlow" :disabled="!isAutoFormValid">
              <template #icon>
                <n-icon><OpenOutline /></n-icon>
              </template>
              开始自动授权
            </n-button>
          </n-space>
        </n-card>
      </n-gi>

      <!-- 手动模式测试 -->
      <n-gi>
        <n-card title="手动模式测试" hoverable>
          <template #header-extra>
            <n-tag type="warning">设备流</n-tag>
          </template>
          
          <n-space vertical>
            <n-alert type="warning">
              手动模式适用于无浏览器环境，会显示授权码供用户复制
            </n-alert>
            
            <n-form :model="manualForm" label-placement="left" label-width="120px">
              <n-form-item label="Client ID">
                <n-input v-model:value="manualForm.clientId" placeholder="输入客户端ID" />
              </n-form-item>
              <n-form-item label="回调地址">
                <n-input v-model:value="manualForm.redirectUri" placeholder="http://localhost:3000/callback" />
              </n-form-item>
              <n-form-item label="State">
                <n-input v-model:value="manualForm.state" placeholder="随机状态值" />
                <template #feedback>
                  <n-button text @click="generateState('manual')">生成随机State</n-button>
                </template>
              </n-form-item>
            </n-form>
            
            <n-button type="warning" block @click="startManualFlow" :disabled="!isManualFormValid">
              <template #icon>
                <n-icon><QrCodeOutline /></n-icon>
              </template>
              开始手动授权
            </n-button>
          </n-space>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 测试结果 -->
    <n-card title="测试结果" style="margin-top: 16px;" v-if="testResults.length > 0">
      <n-timeline>
        <n-timeline-item
          v-for="(result, index) in testResults"
          :key="index"
          :type="result.type"
          :title="result.title"
          :content="result.content"
          :time="result.time"
        />
      </n-timeline>
    </n-card>

    <!-- 快速设置 -->
    <n-card title="快速设置" style="margin-top: 16px;">
      <n-space>
        <n-button @click="fillDefaultValues">
          <template #icon>
            <n-icon><SettingsOutline /></n-icon>
          </template>
          填充默认值
        </n-button>
        <n-button @click="clearAll">
          <template #icon>
            <n-icon><TrashOutline /></n-icon>
          </template>
          清空所有
        </n-button>
        <n-button @click="clearResults">
          <template #icon>
            <n-icon><RefreshOutline /></n-icon>
          </template>
          清空结果
        </n-button>
      </n-space>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  NPageHeader, NGrid, NGi, NCard, NTag, NSpace, NAlert, NForm, NFormItem,
  NInput, NButton, NIcon, NTimeline, NTimelineItem, useMessage
} from 'naive-ui'
import {
  OpenOutline,
  QrCodeOutline,
  SettingsOutline,
  TrashOutline,
  RefreshOutline
} from '@vicons/ionicons5'

const message = useMessage()

// 表单数据
const autoForm = ref({
  clientId: '',
  redirectUri: '',
  state: ''
})

const manualForm = ref({
  clientId: '',
  redirectUri: '',
  state: ''
})

// 测试结果
const testResults = ref<Array<{
  type: 'success' | 'warning' | 'error' | 'info',
  title: string,
  content: string,
  time: string
}>>([])

// 表单验证
const isAutoFormValid = computed(() => {
  return autoForm.value.clientId && autoForm.value.redirectUri && autoForm.value.state
})

const isManualFormValid = computed(() => {
  return manualForm.value.clientId && manualForm.value.redirectUri && manualForm.value.state
})

// 生成随机 State
function generateState(mode: 'auto' | 'manual') {
  const state = 'state_' + Math.random().toString(36).substring(2, 15)
  if (mode === 'auto') {
    autoForm.value.state = state
  } else {
    manualForm.value.state = state
  }
  
  addTestResult('info', `生成随机State`, `${mode}模式: ${state}`)
}

// 开始自动授权流程
function startAutoFlow() {
  const params = new URLSearchParams({
    client_id: autoForm.value.clientId,
    redirect_uri: autoForm.value.redirectUri,
    state: autoForm.value.state
  })
  
  const authUrl = `/api/sso/authorize?${params.toString()}`
  
  addTestResult('info', '启动自动授权', `重定向到: ${authUrl}`)
  
  // 在新窗口打开授权页面
  window.open(authUrl, '_blank', 'width=600,height=700')
}

// 开始手动授权流程
function startManualFlow() {
  const params = new URLSearchParams({
    client_id: manualForm.value.clientId,
    redirect_uri: manualForm.value.redirectUri,
    state: manualForm.value.state,
    device_flow: 'true'
  })
  
  const authUrl = `/api/sso/authorize?${params.toString()}`
  
  addTestResult('warning', '启动手动授权', `重定向到: ${authUrl}`)
  
  // 在新窗口打开授权页面
  window.open(authUrl, '_blank', 'width=600,height=700')
}

// 添加测试结果
function addTestResult(type: 'success' | 'warning' | 'error' | 'info', title: string, content: string) {
  testResults.value.unshift({
    type,
    title,
    content,
    time: new Date().toLocaleTimeString()
  })
}

// 填充默认值
function fillDefaultValues() {
  // 获取当前窗口的端口，用于生成正确的回调地址
  const currentPort = window.location.port || '9998'
  const callbackUrl = `http://localhost:${currentPort}/oauth/callback`
  
  const defaultValues = {
    clientId: 'c35a52681f1fa87a6a11f69d26990326',
    redirectUri: callbackUrl,
    state: 'state_' + Math.random().toString(36).substring(2, 15)
  }
  
  autoForm.value = { ...defaultValues }
  manualForm.value = { ...defaultValues }
  
  addTestResult('success', '填充默认值', `已为两种模式填充默认测试值，回调地址: ${callbackUrl}`)
  message.success('已填充默认值')
}

// 清空所有
function clearAll() {
  autoForm.value = { clientId: '', redirectUri: '', state: '' }
  manualForm.value = { clientId: '', redirectUri: '', state: '' }
  testResults.value = []
  
  message.success('已清空所有数据')
}

// 清空结果
function clearResults() {
  testResults.value = []
  message.success('已清空测试结果')
}

// 监听回调消息（从子窗口）
window.addEventListener('message', (event) => {
  if (event.data.type === 'oauth_callback') {
    const { success, data } = event.data
    
    if (success && data) {
      addTestResult('success', 'OAuth 授权成功', `Token: ${data.token?.substring(0, 20)}...`)
      message.success('已接收到授权token')
    } else {
      addTestResult('error', 'OAuth 授权失败', '未收到有效的token')
    }
  }
})
</script>

<style scoped>
.n-card {
  height: fit-content;
}
</style> 