"use client"

import { useEffect } from 'react'
import { usePathname } from 'next/navigation'

// 页面标题映射
const pageTitles: Record<string, string> = {
  '/': 'Claude Duck - 仪表板',
  '/subscription': 'Claude Duck - 订阅管理',
  '/credits': 'Claude Duck - 积分管理',
  '/settings': 'Claude Duck - 设置',
  '/resources': 'Claude Duck - 资源中心',
  '/test-tool': 'Claude Duck - 测试工具',
  '/login': 'Claude Duck - 登录',
  '/oauth/authorize': 'Claude Duck - 授权',
  '/oauth/callback': 'Claude Duck - 授权回调'
}

// 默认标题
const defaultTitle = 'Claude Duck'

export function PageTitle() {
  const pathname = usePathname()
  
  useEffect(() => {
    // 根据当前路径设置页面标题
    document.title = pageTitles[pathname] || defaultTitle
  }, [pathname])
  
  // 这个组件不渲染任何内容
  return null
}