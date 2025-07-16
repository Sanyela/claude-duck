// 环境变量配置
export const getAppName = (): string => {
  return process.env.APP_NAME || 'Claude Duck'
}

export const getApiUrl = (): string => {
  return process.env.API_URL || 'http://localhost:9998'
}