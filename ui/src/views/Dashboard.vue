<template>
  <div>
    <n-page-header>
      <template #title>
        控制台
      </template>
    </n-page-header>

    <!-- 公告区域 -->
    <div v-if="announcements.length > 0" style="margin-top: 16px;">
      <n-alert 
        v-for="ann in announcements" 
        :key="ann.id" 
        :title="ann.title" 
        :type="ann.type" 
        closable 
        style="margin-bottom: 12px;"
      >
        <div v-html="ann.description"></div>
      </n-alert>
    </div>

    <n-alert title="无有效订阅" type="warning" style="margin-top: 16px;" v-if="!hasSubscription">
      您目前没有有效订阅，请购买订阅以使用我们的服务。
    </n-alert>

    <n-card title="欢迎，my!" style="margin-top: 16px;">
      通过订阅访问 Claude 和其他 AI 服务
    </n-card>

    <n-grid cols="1 s:2 m:3" :x-gap="16" :y-gap="16" style="margin-top: 16px;">
      <n-gi>
        <n-card title="安装 Claude Duck">
          <p>开始使用 Claude Duck CLI 进行开发工作流程。</p>
          <n-button type="primary" @click="handleInstall">安装 Claude Duck</n-button>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card title="管理订阅">
          <p>查看并管理您当前的订阅和支付方式。</p>
          <n-button @click="goToSubscription">前往订阅页面</n-button>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card title="推荐计划">
          <p>邀请朋友一起获得奖励。</p>
          <n-button @click="goToReferral">去推荐计划</n-button>
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NPageHeader, NCard, NGrid, NGi, NButton, NAlert } from 'naive-ui'
import { useRouter } from 'vue-router'
import { getAnnouncements, type Announcement } from '../api/announcements' // 导入API

const router = useRouter()
const announcements = ref<Announcement[]>([])
const hasSubscription = ref(false) // 模拟订阅状态，后续可以从API获取

const fetchAnnouncements = async () => {
  try {
    const response = await getAnnouncements()
    announcements.value = response.announcements
  } catch (error) {
    console.error('获取公告失败:', error)
    // 此处可以使用 Naive UI 的 message provider 显示错误提示
    // import { useMessage } from 'naive-ui'
    // const message = useMessage()
    // message.error('获取公告失败，请稍后再试')
  }
}

onMounted(() => {
  fetchAnnouncements()
  // 可以在这里也获取用户订阅状态等其他初始化数据
})

const handleInstall = () => {
  console.log('Trigger install Claude Duck')
  // 实际的安装逻辑，可能是下载或指引
}

const goToSubscription = () => {
  router.push('/subscription')
}

const goToReferral = () => {
  console.log('Navigate to referral page')
  // router.push('/referral') // 假设有推荐页面
}
</script> 