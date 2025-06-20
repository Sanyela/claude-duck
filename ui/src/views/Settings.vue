<template>
  <div>
    <n-page-header>
      <template #title>
        配置
      </template>
    </n-page-header>

    <n-alert title="无有效订阅" type="warning" style="margin-top: 16px;" v-if="!hasActiveSubscription">
      您目前没有有效订阅，部分设置项可能不可用。
    </n-alert>

    <n-card title="账户设置" style="margin-top: 16px;">
      <n-p>这里可以放置一些用户账户相关的设置项。</n-p>
      <n-empty description="暂无更多设置" />
    </n-card>

    <n-card title="主题设置" style="margin-top: 16px;">
      <n-p>选择您喜欢的主题模式。</n-p>
       <n-switch v-model:value="isDarkMode" @update:value="handleThemeChange">
        <template #checked>
          暗黑模式
        </template>
        <template #unchecked>
          明亮模式
        </template>
      </n-switch>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { NPageHeader, NCard, NEmpty, NP, NAlert, NSwitch } from 'naive-ui';
import { useThemeStore } from '../stores/theme';
import { getActiveSubscription } from '../api/subscription'; // Import for subscription status

const themeStore = useThemeStore();
const isDarkMode = computed({
  get: () => themeStore.isDarkMode,
  set: (value) => {
    if (value !== themeStore.isDarkMode) {
      themeStore.toggleTheme();
    }
  }
});

const handleThemeChange = (value: boolean) => {
  // 主题切换逻辑已通过 computed setter 和 theme store 处理
  console.log('Theme toggled via switch to:', value ? 'dark' : 'light');
};

const hasActiveSubscription = ref(false);

onMounted(async () => {
  try {
    const response = await getActiveSubscription();
    hasActiveSubscription.value = !!response.subscription && response.subscription.status === 'active';
  } catch (error) {
    console.error("Failed to fetch subscription status for settings page", error);
  }
});

</script> 