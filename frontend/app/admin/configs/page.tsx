"use client"

import { useState, useEffect, useCallback } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea"
import { Badge } from "@/components/ui/badge"
import { Switch } from "@/components/ui/switch"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Settings, Edit, Server, DollarSign, Gift, Zap, Search, RefreshCw } from "lucide-react"
import { toast } from "sonner"
import { adminAPI, type SystemConfig } from "@/api/admin"

export default function AdminConfigsPage() {
  
  const [loading, setLoading] = useState(false)
  const [configs, setConfigs] = useState<SystemConfig[]>([])
  const [editingConfig, setEditingConfig] = useState<SystemConfig | null>(null)
  const [isConfigDialogOpen, setIsConfigDialogOpen] = useState(false)
  const [searchTerm, setSearchTerm] = useState("")

  // 配置分组
  const configGroups = {
    api: {
      title: "API 配置",
      icon: Server,
      color: "bg-blue-500",
      configs: ["new_api_endpoint", "new_api_key", "degradation_api_key"]
    },
    billing: {
      title: "计费配置", 
      icon: DollarSign,
      color: "bg-green-500",
      configs: ["prompt_multiplier", "completion_multiplier", "cache_multiplier", "token_threshold", "points_per_threshold"]
    },
    checkin: {
      title: "签到配置",
      icon: Gift,
      color: "bg-purple-500", 
      configs: ["daily_checkin_enabled", "daily_checkin_points", "daily_checkin_validity_days", "daily_checkin_multi_subscription_strategy"]
    },
    system: {
      title: "系统配置",
      icon: Zap,
      color: "bg-orange-500",
      configs: ["free_models_list", "hide_models_list", "model_redirect_map", "model_multiplier_map", "default_degradation_guaranteed", "registration_plan_mapping"]
    }
  }

  const loadConfigs = useCallback(async () => {
    setLoading(true)
    const result = await adminAPI.getSystemConfigs()
    if (result.success && result.configs) {
      setConfigs(Array.isArray(result.configs) ? result.configs : [])
    } else {
      setConfigs([])
      toast.error("加载失败", {
        description: result.message,
      })
    }
    setLoading(false)
  }, [toast])

  useEffect(() => {
    loadConfigs()
  }, [loadConfigs])

  const handleUpdateConfig = async (config: SystemConfig) => {
    const result = await adminAPI.updateSystemConfig({
      config_key: config.config_key,
      config_value: config.config_value
    })
    if (result.success) {
      toast.success("更新成功")
      setIsConfigDialogOpen(false)
      loadConfigs()
    } else {
      toast.error("更新失败", {
        description: result.message,
      })
    }
  }

  // 过滤配置
  const filteredConfigs = configs.filter(config => 
    config.config_key.toLowerCase().includes(searchTerm.toLowerCase()) ||
    config.description.toLowerCase().includes(searchTerm.toLowerCase())
  )

  // 根据分组获取配置
  const getConfigsByGroup = (groupKeys: string[]) => {
    return filteredConfigs.filter(config => groupKeys.includes(config.config_key))
  }

  // 获取未分组的配置
  const getUngroupedConfigs = () => {
    const groupedKeys = Object.values(configGroups).flatMap(group => group.configs)
    return filteredConfigs.filter(config => !groupedKeys.includes(config.config_key))
  }

  // 渲染配置值
  const renderConfigValue = (config: SystemConfig) => {
    const value = config.config_value
    if (config.config_key === "daily_checkin_enabled") {
      return (
        <Badge variant={value === "true" ? "default" : "secondary"}>
          {value === "true" ? "启用" : "禁用"}
        </Badge>
      )
    }
    if (config.config_key === "free_models_list") {
      try {
        const models = JSON.parse(value)
        return (
          <div className="flex flex-wrap gap-1">
            {models.map((model: string, idx: number) => (
              <Badge key={idx} variant="outline" className="text-xs">
                {model}
              </Badge>
            ))}
          </div>
        )
      } catch {
        return <span className="text-muted-foreground">格式错误</span>
      }
    }
    if (config.config_key === "hide_models_list") {
      try {
        const models = JSON.parse(value)
        if (models.length === 0) {
          return <Badge variant="secondary" className="text-xs">未隐藏任何模型</Badge>
        }
        return (
          <div className="flex flex-wrap gap-1">
            {models.map((model: string, idx: number) => (
              <Badge key={idx} variant="destructive" className="text-xs">
                {model}
              </Badge>
            ))}
          </div>
        )
      } catch {
        return <span className="text-muted-foreground">格式错误</span>
      }
    }
    if (config.config_key === "registration_plan_mapping") {
      try {
        const mapping = JSON.parse(value)
        return (
          <div className="flex flex-wrap gap-1">
            {Object.entries(mapping).map(([platform, planId]: [string, any], idx: number) => (
              <Badge key={idx} variant={planId === -1 ? "secondary" : "default"} className="text-xs">
                {platform}: {planId === -1 ? "不赠送" : `套餐${planId}`}
              </Badge>
            ))}
          </div>
        )
      } catch {
        return <span className="text-muted-foreground">格式错误</span>
      }
    }
    if (config.config_key === "model_redirect_map") {
      try {
        const redirectMap = JSON.parse(value)
        const entries = Object.entries(redirectMap)
        if (entries.length === 0) {
          return <Badge variant="secondary" className="text-xs">未设置重定向</Badge>
        }
        return (
          <div className="flex flex-wrap gap-1">
            {entries.map(([from, to]: [string, any], idx: number) => (
              <Badge key={idx} variant="outline" className="text-xs">
                {from} → {to}
              </Badge>
            ))}
          </div>
        )
      } catch {
        return <span className="text-muted-foreground">格式错误</span>
      }
    }
    if (config.config_key === "model_multiplier_map") {
      try {
        const multiplierMap = JSON.parse(value)
        const entries = Object.entries(multiplierMap)
        if (entries.length === 0) {
          return <Badge variant="secondary" className="text-xs">未设置模型倍率</Badge>
        }
        return (
          <div className="flex flex-wrap gap-1">
            {entries.map(([model, multiplier]: [string, any], idx: number) => (
              <Badge key={idx} variant="outline" className="text-xs">
                {model}: {multiplier}×
              </Badge>
            ))}
          </div>
        )
      } catch {
        return <span className="text-muted-foreground">格式错误</span>
      }
    }
    return <span className="font-mono text-sm">{value}</span>
  }

  // 渲染智能编辑器
  const renderConfigEditor = (config: SystemConfig) => {
    if (config.config_key === "daily_checkin_enabled") {
      return (
        <div className="flex items-center space-x-2">
          <Switch
            checked={config.config_value === "true"}
            onCheckedChange={(checked) => 
              setEditingConfig({...config, config_value: checked ? "true" : "false"})
            }
          />
          <span>{config.config_value === "true" ? "启用" : "禁用"}</span>
        </div>
      )
    }
    if (config.config_key === "free_models_list") {
      return (
        <Textarea
          value={config.config_value}
          onChange={(e) => setEditingConfig({...config, config_value: e.target.value})}
          placeholder='["model1", "model2"]'
          className="font-mono text-sm"
          rows={3}
        />
      )
    }
    if (config.config_key === "hide_models_list") {
      return (
        <Textarea
          value={config.config_value}
          onChange={(e) => setEditingConfig({...config, config_value: e.target.value})}
          placeholder='["claude-3-opus-20240229", "claude-3-sonnet-20240229"]'
          className="font-mono text-sm"
          rows={3}
        />
      )
    }
    if (config.config_key === "registration_plan_mapping") {
      return (
        <Textarea
          value={config.config_value}
          onChange={(e) => setEditingConfig({...config, config_value: e.target.value})}
          placeholder='{"default": -1, "linux_do": 1, "github": 2}'
          className="font-mono text-sm"
          rows={4}
        />
      )
    }
    if (config.config_key === "model_redirect_map") {
      return (
        <Textarea
          value={config.config_value}
          onChange={(e) => setEditingConfig({...config, config_value: e.target.value})}
          placeholder='{"claude-3-opus-20240229": "claude-3-5-sonnet-20241022", "claude-3-sonnet-20240229": "claude-3-5-haiku-20241022"}'
          className="font-mono text-sm"
          rows={5}
        />
      )
    }
    if (config.config_key === "model_multiplier_map") {
      return (
        <Textarea
          value={config.config_value}
          onChange={(e) => setEditingConfig({...config, config_value: e.target.value})}
          placeholder='{"claude-3-opus-20240229": 2.0, "claude-3-5-sonnet-20241022": 1.5}'
          className="font-mono text-sm"
          rows={5}
        />
      )
    }
    if (config.config_key.includes("description") || config.config_key.includes("strategy")) {
      return (
        <Textarea
          value={config.config_value}
          onChange={(e) => setEditingConfig({...config, config_value: e.target.value})}
          rows={2}
        />
      )
    }
    return (
      <Input
        value={config.config_value}
        onChange={(e) => setEditingConfig({...config, config_value: e.target.value})}
        className={config.config_key.includes("api") ? "font-mono" : ""}
      />
    )
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* 页头 */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">系统配置</h1>
            <p className="text-muted-foreground">
              管理和调整系统运行参数
            </p>
          </div>
          <Button onClick={loadConfigs} variant="outline" disabled={loading}>
            <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            刷新
          </Button>
        </div>

        {/* 搜索栏 */}
        <Card>
          <CardContent className="pt-6">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="搜索配置项..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
          </CardContent>
        </Card>

        {loading ? (
          <Card>
            <CardContent className="flex items-center justify-center py-8">
              <RefreshCw className="w-6 h-6 animate-spin mr-2" />
              加载中...
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardContent className="pt-6">
              <Tabs defaultValue="api" className="space-y-6">
                <TabsList className="grid w-full grid-cols-5">
                  {Object.entries(configGroups).map(([key, group]) => {
                    const Icon = group.icon
                    const groupConfigs = getConfigsByGroup(group.configs)
                    return (
                      <TabsTrigger 
                        key={key} 
                        value={key} 
                        className="flex items-center space-x-2"
                        disabled={groupConfigs.length === 0}
                      >
                        <Icon className="w-4 h-4" />
                        <span className="hidden sm:inline">{group.title}</span>
                        <span className="sm:hidden">{group.title.split(' ')[0]}</span>
                        {groupConfigs.length > 0 && (
                          <Badge variant="secondary" className="ml-1 text-xs">
                            {groupConfigs.length}
                          </Badge>
                        )}
                      </TabsTrigger>
                    )
                  })}
                  {getUngroupedConfigs().length > 0 && (
                    <TabsTrigger value="other" className="flex items-center space-x-2">
                      <Settings className="w-4 h-4" />
                      <span className="hidden sm:inline">其他</span>
                      <Badge variant="secondary" className="ml-1 text-xs">
                        {getUngroupedConfigs().length}
                      </Badge>
                    </TabsTrigger>
                  )}
                </TabsList>

                {/* 各个分组的内容 */}
                {Object.entries(configGroups).map(([key, group]) => {
                  const groupConfigs = getConfigsByGroup(group.configs)
                  if (groupConfigs.length === 0) return null
                  
                  const Icon = group.icon
                  
                  return (
                    <TabsContent key={key} value={key} className="space-y-4">
                      <div className="flex items-center space-x-3 mb-6">
                        <div className={`p-2 rounded-lg ${group.color}`}>
                          <Icon className="h-5 w-5 text-white" />
                        </div>
                        <div>
                          <h3 className="text-lg font-semibold">{group.title}</h3>
                          <p className="text-sm text-muted-foreground">
                            {groupConfigs.length} 个配置项
                          </p>
                        </div>
                      </div>
                      
                      <div className="grid gap-4">
                        {groupConfigs.map((config) => (
                          <div
                            key={config.id}
                            className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                          >
                            <div className="flex-1 min-w-0">
                              <div className="flex items-center space-x-2 mb-1">
                                <h4 className="font-medium text-sm">{config.config_key}</h4>
                              </div>
                              <p className="text-sm text-muted-foreground mb-2">
                                {config.description}
                              </p>
                              <div className="flex items-center space-x-2">
                                {renderConfigValue(config)}
                              </div>
                            </div>
                            <div className="flex items-center space-x-2 ml-4">
                              <span className="text-xs text-muted-foreground">
                                {new Date(config.updated_at).toLocaleDateString()}
                              </span>
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => {
                                  setEditingConfig(config)
                                  setIsConfigDialogOpen(true)
                                }}
                              >
                                <Edit className="w-4 h-4" />
                              </Button>
                            </div>
                          </div>
                        ))}
                      </div>
                    </TabsContent>
                  )
                })}

                {/* 其他配置 Tab */}
                {getUngroupedConfigs().length > 0 && (
                  <TabsContent value="other" className="space-y-4">
                    <div className="flex items-center space-x-3 mb-6">
                      <div className="p-2 rounded-lg bg-gray-500">
                        <Settings className="h-5 w-5 text-white" />
                      </div>
                      <div>
                        <h3 className="text-lg font-semibold">其他配置</h3>
                        <p className="text-sm text-muted-foreground">
                          {getUngroupedConfigs().length} 个配置项
                        </p>
                      </div>
                    </div>
                    
                    <div className="grid gap-4">
                      {getUngroupedConfigs().map((config) => (
                        <div
                          key={config.id}
                          className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                        >
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center space-x-2 mb-1">
                              <h4 className="font-medium text-sm">{config.config_key}</h4>
                            </div>
                            <p className="text-sm text-muted-foreground mb-2">
                              {config.description}
                            </p>
                            <div className="flex items-center space-x-2">
                              {renderConfigValue(config)}
                            </div>
                          </div>
                          <div className="flex items-center space-x-2 ml-4">
                            <span className="text-xs text-muted-foreground">
                              {new Date(config.updated_at).toLocaleDateString()}
                            </span>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => {
                                setEditingConfig(config)
                                setIsConfigDialogOpen(true)
                              }}
                            >
                              <Edit className="w-4 h-4" />
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </TabsContent>
                )}
              </Tabs>
            </CardContent>
          </Card>
        )}

        {/* 配置编辑对话框 */}
        <Dialog open={isConfigDialogOpen} onOpenChange={setIsConfigDialogOpen}>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle className="flex items-center space-x-2">
                <Settings className="h-5 w-5" />
                <span>编辑配置</span>
              </DialogTitle>
              <DialogDescription>
                修改 "{editingConfig?.config_key}" 的配置值
              </DialogDescription>
            </DialogHeader>
            {editingConfig && (
              <div className="space-y-6">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label className="text-sm font-medium">配置键</Label>
                    <Input 
                      value={editingConfig.config_key} 
                      disabled 
                      className="font-mono text-sm bg-muted"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label className="text-sm font-medium">更新时间</Label>
                    <Input 
                      value={new Date(editingConfig.updated_at).toLocaleString()} 
                      disabled 
                      className="text-sm bg-muted"
                    />
                  </div>
                </div>
                
                <div className="space-y-2">
                  <Label className="text-sm font-medium">描述</Label>
                  <Input 
                    value={editingConfig.description} 
                    disabled 
                    className="bg-muted"
                  />
                </div>
                
                <div className="space-y-2">
                  <Label className="text-sm font-medium">配置值</Label>
                  {renderConfigEditor(editingConfig)}
                  {editingConfig.config_key === "free_models_list" && (
                    <p className="text-xs text-muted-foreground">
                      格式：JSON 数组，例如 ["model1", "model2"]
                    </p>
                  )}
                  {editingConfig.config_key === "hide_models_list" && (
                    <p className="text-xs text-muted-foreground">
                      格式：JSON 数组，例如 ["claude-3-opus-20240229", "claude-3-sonnet-20240229"]，列表中的模型将返回503不支持错误
                    </p>
                  )}
                  {editingConfig.config_key === "model_redirect_map" && (
                    <p className="text-xs text-muted-foreground">
                      格式：JSON 对象，例如 {`{"原始模型": "目标模型"}`}，空对象 {`{}`} 表示不重定向
                    </p>
                  )}
                  {editingConfig.config_key === "model_multiplier_map" && (
                    <p className="text-xs text-muted-foreground">
                      格式：JSON 对象，例如 {`{"模型名": 倍率}`}，倍率为数值类型，空对象 {`{}`} 表示不应用额外倍率
                    </p>
                  )}
                  {editingConfig.config_key === "registration_plan_mapping" && (
                    <p className="text-xs text-muted-foreground">
                      格式：JSON 对象，例如 {`{"platform": planId}`}，-1 表示不赠送套餐
                    </p>
                  )}
                  {editingConfig.config_key.includes("multiplier") && (
                    <p className="text-xs text-muted-foreground">
                      数值类型，支持小数
                    </p>
                  )}
                </div>
              </div>
            )}
            <DialogFooter>
              <Button 
                variant="outline" 
                onClick={() => setIsConfigDialogOpen(false)}
              >
                取消
              </Button>
              <Button 
                onClick={() => editingConfig && handleUpdateConfig(editingConfig)}
                disabled={!editingConfig}
              >
                保存更改
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </DashboardLayout>
  )
} 