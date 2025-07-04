import { request } from './request'

// 设备信息类型定义
export interface Device {
  id: string
  device_name: string
  device_type: 'mobile' | 'tablet' | 'desktop'
  ip: string
  location: string
  source: 'web' | 'sso'
  created_at: string
  last_active: string
  expires_at: string
  is_current: boolean
}

// 设备统计信息
export interface DeviceStats {
  total: number
  web: number
  sso: number
  mobile: number
  desktop: number
  tablet: number
}

// API响应类型
export interface DeviceListResponse {
  data: {
    success: boolean
    data: {
      devices: Device[]
      total: number
    }
  }
}

export interface DeviceStatsResponse {
  data: {
    success: boolean
    data: DeviceStats
  }
}

export interface DeviceActionResponse {
  data: {
    success: boolean
    message: string
    data?: {
      revoked_count?: number
    }
  }
}

// 获取设备列表
export const getDevices = (): Promise<DeviceListResponse> => {
  return request.get('/api/devices')
}

// 获取设备统计
export const getDeviceStats = (): Promise<DeviceStatsResponse> => {
  return request.get('/api/devices/stats')
}

// 下线指定设备
export const revokeDevice = (deviceId: string): Promise<DeviceActionResponse> => {
  return request.delete(`/api/devices/${deviceId}`)
}

// 下线其他设备（保留当前设备）
export const revokeOtherDevices = (): Promise<DeviceActionResponse> => {
  return request.delete('/api/devices')
}

// 强制下线所有设备（包括当前设备）
export const revokeAllDevices = (): Promise<DeviceActionResponse> => {
  return request.delete('/api/devices/force')
}