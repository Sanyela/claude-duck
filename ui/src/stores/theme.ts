import { defineStore } from 'pinia'
import { ref, computed, watchEffect } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const isDarkMode = ref(true) // 默认暗黑模式

  // 计算属性，用于模板中判断主题
  const theme = computed(() => isDarkMode.value ? 'dark' : 'light')

  // 尝试从 localStorage 读取保存的主题设置
  const savedTheme = localStorage.getItem('app-theme')
  if (savedTheme) {
    isDarkMode.value = savedTheme === 'dark'
  }

  function toggleTheme() {
    isDarkMode.value = !isDarkMode.value
  }

  // 当主题变化时，更新localStorage
  watchEffect(() => {
    localStorage.setItem('app-theme', isDarkMode.value ? 'dark' : 'light')
    // 你也可以在这里添加/移除 body 上的 class，如果 Naive UI 的 NConfigProvider 不足以满足所有样式需求
    // document.body.classList.toggle('dark-mode', isDarkMode.value);
  })

  return { isDarkMode, theme, toggleTheme }
}) 