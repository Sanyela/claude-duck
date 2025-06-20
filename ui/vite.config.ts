import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig(({ mode }) => ({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': '/src'
    }
  },
  optimizeDeps: {
    include: [
      '@vicons/fluent'
    ],
    force: true
  },
  define: {
    // 根据构建模式设置环境变量
    // 'import.meta.env.VITE_APP_USE_MOCK': mode === 'development' ? '"true"' : '"false"',
    'import.meta.env.VITE_APP_USE_MOCK': '"false"', // 禁用mock，直接连接后端
    'import.meta.env.VITE_APP_OAUTH_DEV_MODE': '"false"', // 禁用OAuth开发模式
    'import.meta.env.VITE_API_BASE_URL': mode === 'development' ? '"http://localhost:9998/api"' : '"/api"'
  }
}))
