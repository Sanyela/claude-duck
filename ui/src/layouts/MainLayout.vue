<template>
  <n-layout style="height: 100vh" has-sider>
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="240"
      :collapsed="collapsed"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
    >
      <n-menu
        v-model:value="activeKey"
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        :options="menuOptions"
        @update:value="handleMenuSelect"
      />
    </n-layout-sider>
    <n-layout>
      <n-layout-header bordered style="height: 64px; padding: 0 24px; display: flex; align-items: center; justify-content: space-between;">
        <div>Claude Duck</div>
        <div style="display: flex; align-items: center; gap: 16px;">
          <NIcon size="20" style="cursor: pointer" @click="toggleTheme">
            <SunnyOutline v-if="themeStore.theme === 'dark'" />
            <MoonOutline v-else />
          </NIcon>
          <n-dropdown :options="userOptions" @select="handleUserOptionSelect">
            <NAvatar
              round
              :size="32"
              style="cursor: pointer"
            >
              <NIcon>
                <PersonOutline />
              </NIcon>
            </NAvatar>
          </n-dropdown>
        </div>
      </n-layout-header>
      <n-layout-content content-style="padding: 24px;">
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup lang="ts">
import { ref, h } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { NIcon, NLayout, NLayoutSider, NMenu, NLayoutContent, NLayoutHeader, NAvatar, NDropdown } from 'naive-ui'
import {
  HomeOutline,
  DocumentTextOutline,
  WalletOutline,
  LibraryOutline,
  SettingsOutline,
  PersonOutline,
  SunnyOutline,
  MoonOutline,
  CloseOutline,
  LogOutOutline,
} from '@vicons/ionicons5'
import { useThemeStore } from '@/stores/theme'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const themeStore = useThemeStore()
const authStore = useAuthStore()

const collapsed = ref(false)
const activeKey = ref<string | null>(null)

function renderIcon(icon: any) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const toggleTheme = () => {
  themeStore.toggleTheme()
}

const menuOptions = [
  {
    label: () => h(RouterLink, { to: '/' }, { default: () => '控制台' }),
    key: 'dashboard',
    icon: renderIcon(HomeOutline),
  },
  {
    label: () => h(RouterLink, { to: '/subscription' }, { default: () => '订阅' }),
    key: 'subscription',
    icon: renderIcon(DocumentTextOutline),
  },
  {
    label: () => h(RouterLink, { to: '/credits' }, { default: () => '积分' }),
    key: 'credits',
    icon: renderIcon(WalletOutline),
  },
  {
    label: () => h(RouterLink, { to: '/resources' }, { default: () => '资料中心' }),
    key: 'resources',
    icon: renderIcon(LibraryOutline),
  },
  {
    label: () => h(RouterLink, { to: '/settings' }, { default: () => '配置' }),
    key: 'settings',
    icon: renderIcon(SettingsOutline),
  },
  {
    label: () => h(RouterLink, { to: '/oauth-test' }, { default: () => 'OAuth测试' }),
    key: 'oauth-test',
    icon: renderIcon(CloseOutline),
  },
]

const userOptions = [
  {
    label: authStore.user?.username || '用户',
    key: 'username',
    disabled: true,
  },
  {
    type: 'divider',
    key: 'divider'
  },
  {
    label: '退出登录',
    key: 'signout',
    icon: renderIcon(LogOutOutline),
  }
]

const handleMenuSelect = (key: string) => {
  activeKey.value = key
}

const handleUserOptionSelect = async (key: string | number) => {
  if (key === 'signout') {
    await authStore.logout()
    router.push('/login')
  }
}

</script>

<style scoped>
/* 可以在这里添加一些自定义样式 */
</style> 