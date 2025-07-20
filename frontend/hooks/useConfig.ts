"use client"

import { useState, useEffect } from 'react'
import { getConfig } from '@/lib/env'

// React Hook for client-side config
export const useConfig = () => {
  const [config, setConfig] = useState({
    appName: 'Duck Code',
    apiUrl: 'http://localhost:9998',
    installCommand: 'npm install -g http://111.180.197.234:7778/install --registry=https://registry.npmmirror.com',
    docsUrl: 'https://github.com/anthropics/claude-code',
    claudeUrl: 'https://api.anthropic.com'
  })
  const [isLoaded, setIsLoaded] = useState(false)

  useEffect(() => {
    getConfig().then(newConfig => {
      setConfig(newConfig)
      setIsLoaded(true)
    })
  }, [])

  return { ...config, isLoaded }
}