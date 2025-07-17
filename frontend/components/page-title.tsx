"use client"

import { useEffect } from 'react'
import { usePathname } from 'next/navigation'
import { getAppName } from '@/lib/env'

// 页面标题映射函数
const getPageTitles = (): Record<string, string> => {
  const appName = getAppName()
  return {
    '/': `${appName} - 仪表板`,
    '/subscription': `${appName} - 订阅管理`,
    '/credits': `${appName} - 积分管理`,
    '/settings': `${appName} - 设置`,
    '/resources': `${appName} - 安装教程`,
    '/login': `${appName} - 登录`,
    '/oauth/authorize': `${appName} - 授权`,
    '/oauth/callback': `${appName} - 授权回调`
  }
}

export function PageTitle() {
  const pathname = usePathname()
  
  useEffect(() => {
    // 根据当前路径设置页面标题
    const pageTitles = getPageTitles()
    const defaultTitle = getAppName()
    document.title = pageTitles[pathname] || defaultTitle
  }, [pathname])
  
  // 这个组件不渲染任何内容
  return null
}