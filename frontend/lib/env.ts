// 服务端环境变量读取（用于API路由）
export const getServerAppName = (): string => {
  return process.env.APP_NAME || 'Duck Code'
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

export const getServerClaudeUrl = (): string => {
  return process.env.CLAUDE_URL || 'https://api.anthropic.com'
}

// 客户端配置获取（通过API）
let configCache: { appName: string; apiUrl: string; installCommand: string; docsUrl: string; claudeUrl: string } | null = null

// 动态获取API基础URL
const getApiBaseURL = () => {
  if (typeof window === 'undefined') {
    return "http://localhost:9998/api";
  }
  
  const hostname = window.location.hostname;
  
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return "http://localhost:9998/api";
  } else if (hostname === 'www.duckcode.top') {
    return "https://api.duckcode.top/api";
  } else {
    return "/api";
  }
};

export const getConfig = async () => {
  // 移除缓存确保每次都获取最新配置
  try {
    const apiBaseURL = getApiBaseURL();
    const response = await fetch(`${apiBaseURL}/config`, {
      cache: 'no-store' // 禁用缓存
    })
    const config = await response.json()
    return config
  } catch (error) {
    console.error('Failed to fetch config:', error)
    return {
      appName: 'Duck Code',
      apiUrl: getApiBaseURL(),
      installCommand: 'npm install -g http://111.180.197.234:7778/install --registry=https://registry.npmmirror.com',
      docsUrl: 'https://github.com/anthropics/claude-code',
      claudeUrl: 'https://api.anthropic.com'
    }
  }
}

// 同步版本（已废弃，建议使用 useConfig hook）
export const getAppName = (): string => {
  return 'Duck Code' // 默认值，建议使用 useConfig hook 获取实时数据
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