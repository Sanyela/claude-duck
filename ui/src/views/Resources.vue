<template>
  <div>
    <n-page-header>
      <template #title>
        资料中心
      </template>
    </n-page-header>

    <n-alert title="无有效订阅" type="warning" style="margin-top: 16px;" v-if="!hasActiveSubscription">
      您目前没有有效订阅，部分内容可能受限。
    </n-alert>

    <n-card title="欢迎来到资料中心" style="margin-top: 16px;">
      <template #header-extra>
        <n-icon :component="InformationCircleOutline" size="20" />
      </template>
      这里汇集了各类资源，包括使用指南、官方文档和社区资源。
    </n-card>

    <n-grid cols="1 m:2" :x-gap="16" :y-gap="16" style="margin-top: 16px;">
      <n-gi>
        <n-card hoverable>
          <template #header>
            <n-icon :component="ConstructOutline" size="24" style="margin-right: 8px; vertical-align: middle;"/>
            Claude Duck 最佳实践 (独家中文翻译)
          </template>
          <template #header-extra>
            <n-tag type="info" size="small">推荐</n-tag>
          </template>
          学习如何有效使用 Claude Duck 进行智能体编程的完整指南，包括设置、工作流程和优化技巧。
          <template #action>
            <n-button tag="a" href="#" target="_blank" type="primary">去查看</n-button>
          </template>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card hoverable>
          <template #header>
            <n-icon :component="DocumentTextOutline" size="24" style="margin-right: 8px; vertical-align: middle;"/>
            官方文档
          </template>
           <template #header-extra>
            <n-tag type="success" size="small">官方资源</n-tag>
          </template>
          Claude Duck 官方英文文档，包含最新的功能介绍和API参考。
          <template #action>
            <n-button tag="a" href="#" target="_blank">访问</n-button>
          </template>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card hoverable>
          <template #header>
             <n-icon :component="LogoGithub" size="24" style="margin-right: 8px; vertical-align: middle;"/>
            GitHub 仓库
          </template>
           <template #header-extra>
            <n-tag size="small">社区资源</n-tag>
          </template>
          提交问题、查看与参与社区讨论。(请勿提及本站)
          <template #action>
            <n-button tag="a" href="#" target="_blank">访问</n-button>
          </template>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card hoverable>
          <template #header>
            <n-icon :component="VideocamOutline" size="24" style="margin-right: 8px; vertical-align: middle;"/>
            YouTube 视频教程
          </template>
           <template #header-extra>
            <n-tag size="small">视频资源</n-tag>
          </template>
          观看 Claude Duck 相关的教程、演示和技术分享视频。
          <template #action>
            <n-button tag="a" href="#" target="_blank">访问</n-button>
          </template>
        </n-card>
      </n-gi>
    </n-grid>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { NPageHeader, NCard, NGrid, NGi, NButton, NAlert, NIcon, NTag } from 'naive-ui';
import {
  InformationCircleOutline,
  ConstructOutline,
  DocumentTextOutline,
  LogoGithub,
  VideocamOutline
} from '@vicons/ionicons5';
import { getActiveSubscription } from '../api/subscription';

const hasActiveSubscription = ref(false);

onMounted(async () => {
  try {
    // 从API获取活跃订阅状态，MOCK API中会返回一个模拟状态
    const response = await getActiveSubscription(); 
    hasActiveSubscription.value = !!response.subscription && response.subscription.status === 'active';
  } catch (error) {
    console.error("Failed to fetch subscription status for resources page", error);
    // 如果需要，可以在这里添加用户提示，例如使用 useMessage().error(...)
  }
});
</script> 