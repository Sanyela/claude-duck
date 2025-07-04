"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { 
  Smartphone, 
  Monitor, 
  Tablet, 
  Globe, 
  Clock, 
  MapPin, 
  LogOut, 
  Shield, 
  AlertTriangle,
  Trash2,
  RotateCcw
} from "lucide-react"
import { Device, DeviceStats, getDevices, getDeviceStats, revokeDevice, revokeOtherDevices, revokeAllDevices } from "@/api/devices"
import { useToast } from "@/components/ui/use-toast"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"

// 设备类型图标映射
const getDeviceIcon = (type: string) => {
  switch (type) {
    case 'mobile':
      return <Smartphone className="h-5 w-5" />
    case 'tablet':
      return <Tablet className="h-5 w-5" />
    case 'desktop':
    default:
      return <Monitor className="h-5 w-5" />
  }
}

// 设备来源颜色映射
const getSourceColor = (source: string) => {
  switch (source) {
    case 'web':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300'
    case 'sso':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300'
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-300'
  }
}

// 格式化时间
const formatTime = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60))
  
  if (diffInMinutes < 1) {
    return '刚刚'
  } else if (diffInMinutes < 60) {
    return `${diffInMinutes}分钟前`
  } else if (diffInMinutes < 1440) {
    return `${Math.floor(diffInMinutes / 60)}小时前`
  } else {
    return `${Math.floor(diffInMinutes / 1440)}天前`
  }
}

// 格式化过期时间
const formatExpirationTime = (expiresAt: string) => {
  // 检查是否为零值时间或空字符串
  if (!expiresAt || expiresAt === '0001-01-01T00:00:00Z' || new Date(expiresAt).getTime() === 0) {
    return '永不过期'
  }
  
  const expiresDate = new Date(expiresAt)
  const now = new Date()
  
  if (expiresDate <= now) {
    return '已过期'
  }
  
  return expiresDate.toLocaleString()
}

export function DeviceManagement() {
  const [devices, setDevices] = useState<Device[]>([])
  const [stats, setStats] = useState<DeviceStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const { toast } = useToast()

  // 加载设备数据
  const loadDevices = async () => {
    try {
      setLoading(true)
      const [devicesResult, statsResult] = await Promise.all([
        getDevices(),
        getDeviceStats()
      ])
      
      
      if (devicesResult.data && devicesResult.data.success) {
        setDevices(devicesResult.data.data.devices)
      }
      
      if (statsResult.data && statsResult.data.success) {
        setStats(statsResult.data.data)
      }
    } catch (error) {
      console.error('加载设备数据失败:', error)
      toast({
        title: "加载失败",
        description: "无法加载设备数据，请刷新重试",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadDevices()
  }, [])

  // 下线指定设备
  const handleRevokeDevice = async (device: Device) => {
    try {
      setActionLoading(device.id)
      const result = await revokeDevice(device.id)
      
      if (result.data.success) {
        toast({
          title: "设备已下线",
          description: `${device.device_name} 已成功下线`,
        })
        
        // 如果下线的是当前设备，重新加载页面（用户会被登出）
        if (device.is_current) {
          setTimeout(() => {
            window.location.reload()
          }, 1000) // 延迟1秒让用户看到成功提示
        } else {
          loadDevices()
        }
      } else {
        throw new Error(result.data.message || '下线失败')
      }
    } catch (error: any) {
      console.error('下线设备失败:', error)
      toast({
        title: "下线失败",
        description: error.message || "下线设备失败，请重试",
        variant: "destructive",
      })
    } finally {
      setActionLoading(null)
    }
  }

  // 下线其他设备
  const handleRevokeOthers = async () => {
    try {
      setActionLoading('others')
      const result = await revokeOtherDevices()
      
      if (result.data.success) {
        toast({
          title: "其他设备已下线",
          description: `已下线 ${result.data.data?.revoked_count || 0} 个设备`,
        })
        loadDevices()
      } else {
        throw new Error(result.data.message || '下线失败')
      }
    } catch (error: any) {
      console.error('下线其他设备失败:', error)
      toast({
        title: "下线失败",
        description: error.message || "下线其他设备失败，请重试",
        variant: "destructive",
      })
    } finally {
      setActionLoading(null)
    }
  }

  // 强制下线所有设备
  const handleRevokeAll = async () => {
    try {
      setActionLoading('all')
      const result = await revokeAllDevices()
      
      if (result.data.success) {
        toast({
          title: "所有设备已下线",
          description: "您将需要重新登录",
        })
        // 强制下线所有设备后，用户会被登出
        window.location.reload()
      } else {
        throw new Error(result.data.message || '下线失败')
      }
    } catch (error: any) {
      console.error('下线所有设备失败:', error)
      toast({
        title: "下线失败",
        description: error.message || "下线所有设备失败，请重试",
        variant: "destructive",
      })
    } finally {
      setActionLoading(null)
    }
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            设备管理
          </CardTitle>
          <CardDescription>
            正在加载设备信息...
          </CardDescription>
        </CardHeader>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          设备管理
        </CardTitle>
        <CardDescription>
          管理已登录的设备，保护账户安全
        </CardDescription>
      </CardHeader>
      
      <CardContent className="space-y-6">
        {/* 统计信息 */}
        {stats && (
          <div className="grid grid-cols-3 gap-4 max-w-2xl mx-auto">
            <div className="text-center p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
              <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">{stats.total}</div>
              <div className="text-sm text-blue-600 dark:text-blue-400">总设备</div>
            </div>
            <div className="text-center p-3 bg-green-50 dark:bg-green-900/20 rounded-lg">
              <div className="text-2xl font-bold text-green-600 dark:text-green-400">{stats.web}</div>
              <div className="text-sm text-green-600 dark:text-green-400">Web端</div>
            </div>
            <div className="text-center p-3 bg-purple-50 dark:bg-purple-900/20 rounded-lg">
              <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">{stats.sso}</div>
              <div className="text-sm text-purple-600 dark:text-purple-400">SSO客户端</div>
            </div>
          </div>
        )}

        {/* 批量操作按钮 */}
        <div className="flex flex-wrap gap-2 justify-center">
          <Button
            variant="outline"
            onClick={handleRevokeOthers}
            disabled={actionLoading === 'others'}
            className="flex items-center gap-2"
          >
            <LogOut className="h-4 w-4" />
            下线其他设备
          </Button>
          
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button
                variant="destructive"
                disabled={actionLoading === 'all'}
                className="flex items-center gap-2"
              >
                <Trash2 className="h-4 w-4" />
                强制下线所有设备
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle className="flex items-center gap-2">
                  <AlertTriangle className="h-5 w-5 text-red-500" />
                  确认强制下线
                </AlertDialogTitle>
                <AlertDialogDescription>
                  此操作将下线所有设备（包括当前设备），您需要重新登录。确定要继续吗？
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>取消</AlertDialogCancel>
                <AlertDialogAction onClick={handleRevokeAll} className="bg-red-500 hover:bg-red-600">
                  确认下线
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>

          <Button
            variant="ghost"
            onClick={loadDevices}
            className="flex items-center gap-2"
          >
            <RotateCcw className="h-4 w-4" />
            刷新
          </Button>
        </div>

        <Separator />

        {/* 设备列表 */}
        <div className="space-y-4">
          {devices.length === 0 ? (
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>
                暂无已登录的设备
              </AlertDescription>
            </Alert>
          ) : (
            devices.map((device) => (
              <div
                key={device.id}
                className={`p-4 border rounded-lg ${
                  device.is_current 
                    ? 'border-blue-200 bg-blue-50 dark:border-blue-800 dark:bg-blue-900/20' 
                    : 'border-border bg-card'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-3">
                    <div className="text-muted-foreground mt-1">
                      {getDeviceIcon(device.device_type)}
                    </div>
                    
                    <div className="space-y-2">
                      <div className="flex items-center gap-2">
                        <h4 className="font-medium">{device.device_name}</h4>
                        {device.is_current && (
                          <Badge variant="outline" className="text-blue-600 border-blue-200">
                            当前设备
                          </Badge>
                        )}
                        <Badge className={getSourceColor(device.source)}>
                          {device.source === 'web' ? 'Web端' : 'SSO客户端'}
                        </Badge>
                      </div>
                      
                      <div className="text-sm text-muted-foreground space-y-1">
                        <div className="flex items-center gap-2">
                          <Globe className="h-3 w-3" />
                          <span>{device.ip}</span>
                          {device.location && (
                            <>
                              <MapPin className="h-3 w-3" />
                              <span>{device.location}</span>
                            </>
                          )}
                        </div>
                        
                        <div className="flex items-center gap-4">
                          <div className="flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            <span>最后活跃: {formatTime(device.last_active)}</span>
                          </div>
                          <div className="text-xs">
                            登录时间: {new Date(device.created_at).toLocaleString()}
                          </div>
                        </div>
                        
                        <div className="flex items-center gap-1 text-xs">
                          <Clock className="h-3 w-3" />
                          <span className={`${
                            formatExpirationTime(device.expires_at) === '永不过期' 
                              ? 'text-green-600 dark:text-green-400 font-medium' 
                              : 'text-muted-foreground'
                          }`}>
                            过期时间: {formatExpirationTime(device.expires_at)}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-2">
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={actionLoading === device.id}
                          className={`${
                            device.is_current 
                              ? "text-orange-600 hover:text-orange-700 hover:bg-orange-50 dark:hover:bg-orange-900/20" 
                              : "text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900/20"
                          }`}
                        >
                          <LogOut className="h-3 w-3" />
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle className="flex items-center gap-2">
                            <AlertTriangle className={`h-5 w-5 ${device.is_current ? 'text-orange-500' : 'text-red-500'}`} />
                            {device.is_current ? '下线当前设备' : '确认下线设备'}
                          </AlertDialogTitle>
                          <AlertDialogDescription>
                            {device.is_current ? (
                              <>
                                <span className="text-orange-600 dark:text-orange-400 font-medium">⚠️ 警告：</span>
                                您正在下线当前设备 "{device.device_name}"。执行此操作后，您将立即退出登录并需要重新登录才能继续使用。
                              </>
                            ) : (
                              `确定要下线设备 "${device.device_name}" 吗？该设备将需要重新登录。`
                            )}
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>取消</AlertDialogCancel>
                          <AlertDialogAction 
                            onClick={() => handleRevokeDevice(device)}
                            className={`${
                              device.is_current 
                                ? "bg-orange-500 hover:bg-orange-600" 
                                : "bg-red-500 hover:bg-red-600"
                            }`}
                          >
                            {device.is_current ? '确认下线并退出' : '确认下线'}
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  )
}