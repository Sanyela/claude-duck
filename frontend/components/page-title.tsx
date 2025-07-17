"use client"

import { useEffect } from 'react'
import { usePathname } from 'next/navigation'
import { useConfig } from '@/hooks/useConfig'

export function PageTitle() {
  const pathname = usePathname()
  const { appName } = useConfig()
  
  useEffect(() => {
    // 根据当前路径设置页面标题
    const pageTitles: Record<string, string> = {
      '/': `${appName} - 仪表板`,
      '/subscription': `${appName} - 订阅管理`,
      '/credits': `${appName} - 积分管理`,
      '/settings': `${appName} - 设置`,
      '/resources': `${appName} - 安装教程`,
      '/login': `${appName} - 登录`,
      '/oauth/authorize': `${appName} - 授权`,
      '/oauth/callback': `${appName} - 授权回调`
    }
    
    document.title = pageTitles[pathname] || appName
  }, [pathname, appName])
  
  // 这个组件不渲染任何内容
  return null
}