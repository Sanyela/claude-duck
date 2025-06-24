"use client"

import { useState, useEffect } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Settings, Edit } from "lucide-react"
import { useToast } from "@/components/ui/use-toast"
import { adminAPI, type SystemConfig } from "@/api/admin"

export default function AdminConfigsPage() {
  const { toast } = useToast()
  const [loading, setLoading] = useState(false)
  const [configs, setConfigs] = useState<SystemConfig[]>([])
  const [editingConfig, setEditingConfig] = useState<SystemConfig | null>(null)
  const [isConfigDialogOpen, setIsConfigDialogOpen] = useState(false)

  const loadConfigs = async () => {
    setLoading(true)
    const result = await adminAPI.getSystemConfigs()
    if (result.success && result.configs) {
      setConfigs(Array.isArray(result.configs) ? result.configs : [])
    } else {
      setConfigs([])
      toast({
        title: "加载失败",
        description: result.message,
        variant: "destructive"
      })
    }
    setLoading(false)
  }

  useEffect(() => {
    loadConfigs()
  }, [])

  const handleUpdateConfig = async (config: SystemConfig) => {
    const result = await adminAPI.updateSystemConfig({
      config_key: config.config_key,
      config_value: config.config_value
    })
    if (result.success) {
      toast({ title: "更新成功", variant: "default" })
      setIsConfigDialogOpen(false)
      loadConfigs()
    } else {
      toast({
        title: "更新失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  return (
    <DashboardLayout>
      <Card className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
        <CardHeader>
          <div className="flex items-center space-x-2">
            <Settings className="h-6 w-6 text-green-500" />
            <CardTitle className="text-slate-900 dark:text-slate-100">系统配置</CardTitle>
          </div>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            管理系统的配置参数
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <h3 className="text-lg font-semibold">配置列表</h3>
            </div>
            
            {loading ? (
              <div className="text-center py-8">加载中...</div>
            ) : (
              <div className="border rounded-md">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>配置键</TableHead>
                      <TableHead>配置值</TableHead>
                      <TableHead>描述</TableHead>
                      <TableHead>更新时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {Array.isArray(configs) && configs.map((config) => (
                      <TableRow key={config.id}>
                        <TableCell className="font-mono text-sm">{config.config_key}</TableCell>
                        <TableCell className="font-mono text-sm">{config.config_value}</TableCell>
                        <TableCell>{config.description}</TableCell>
                        <TableCell>{new Date(config.updated_at).toLocaleString()}</TableCell>
                        <TableCell>
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
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </div>

          {/* 配置编辑对话框 */}
          <Dialog open={isConfigDialogOpen} onOpenChange={setIsConfigDialogOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>编辑系统配置</DialogTitle>
                <DialogDescription>修改系统配置值</DialogDescription>
              </DialogHeader>
              {editingConfig && (
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label>配置键</Label>
                    <Input value={editingConfig.config_key} disabled />
                  </div>
                  <div className="space-y-2">
                    <Label>配置值</Label>
                    <Input
                      value={editingConfig.config_value}
                      onChange={(e) => setEditingConfig({...editingConfig, config_value: e.target.value})}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>描述</Label>
                    <Input value={editingConfig.description} disabled />
                  </div>
                </div>
              )}
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsConfigDialogOpen(false)}>取消</Button>
                <Button onClick={() => editingConfig && handleUpdateConfig(editingConfig)}>保存</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>
    </DashboardLayout>
  )
} 