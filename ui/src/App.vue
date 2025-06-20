<script setup lang="ts">
import { computed } from 'vue';
import { NConfigProvider, NGlobalStyle, NMessageProvider, NDialogProvider, darkTheme, lightTheme } from 'naive-ui';
import { useThemeStore } from './stores/theme'; // 假设你有一个theme store

const themeStore = useThemeStore();
const theme = computed(() => (themeStore.isDarkMode ? darkTheme : lightTheme));
</script>

<template>
  <n-config-provider :theme="theme">
    <n-message-provider>
      <n-dialog-provider>
         <router-view />
      </n-dialog-provider>
    </n-message-provider>
    <n-global-style />
  </n-config-provider>
</template>

<style>
/* 全局样式可以放在这里或 style.css */
body {
  margin: 0;
  padding: 0;
  font-family: v-sans, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol";
}
</style>
