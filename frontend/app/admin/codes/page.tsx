"use client"

import { useState, useEffect } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Key, Plus, Trash2 } from "lucide-react"
import { useToast } from "@/components/ui/use-toast"
import { adminAPI, type ActivationCode, type SubscriptionPlan } from "@/api/admin"

export default function AdminCodesPage() {
  const { toast } = useToast()
  const [loading, setLoading] = useState(false)
  const [codes, setCodes] = useState<ActivationCode[]>([])
  const [plans, setPlans] = useState<SubscriptionPlan[]>([])
  const [isCodeDialogOpen, setIsCodeDialogOpen] = useState(false)
  const [newCodeData, setNewCodeData] = useState({
    subscription_plan_id: 0,
    count: 1,
    batch_number: ""
  })

  const loadCodes = async () => {
    setLoading(true)
    const result = await adminAPI.getActivationCodes()
    if (result.success && result.codes) {
      setCodes(Array.isArray(result.codes) ? result.codes : [])
    } else {
      setCodes([])
      toast({
        title: "加载失败",
        description: result.message,
        variant: "destructive"
      })
    }
    setLoading(false)
  }

  const loadPlans = async () => {
    const result = await adminAPI.getSubscriptionPlans()
    if (result.success && result.plans) {
      setPlans(Array.isArray(result.plans) ? result.plans : [])
    }
  }

  useEffect(() => {
    loadCodes()
    loadPlans()
  }, [])

  const handleCreateCodes = async () => {
    if (newCodeData.subscription_plan_id === 0) {
      toast({
        title: "请选择订阅计划",
        variant: "destructive"
      })
      return
    }

    const result = await adminAPI.createActivationCodes(newCodeData)
    if (result.success) {
      toast({ title: "创建成功", variant: "default" })
      setIsCodeDialogOpen(false)
      setNewCodeData({
        subscription_plan_id: 0,
        count: 1,
        batch_number: ""
      })
      loadCodes()
    } else {
      toast({
        title: "创建失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  const handleDeleteCode = async (codeId: number) => {
    if (!confirm("确定要删除此激活码吗？")) return
    
    const result = await adminAPI.deleteActivationCode(codeId)
    if (result.success) {
      toast({ title: "删除成功", variant: "default" })
      loadCodes()
    } else {
      toast({
        title: "删除失败",
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
            <Key className="h-6 w-6 text-orange-500" />
            <CardTitle className="text-slate-900 dark:text-slate-100">激活码管理</CardTitle>
          </div>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            生成和管理订阅计划的激活码
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <h3 className="text-lg font-semibold">激活码列表</h3>
              <Button
                onClick={() => setIsCodeDialogOpen(true)}
                className="bg-orange-500 hover:bg-orange-600 text-white"
              >
                <Plus className="w-4 h-4 mr-2" />
                生成激活码
              </Button>
            </div>
            
            {loading ? (
              <div className="text-center py-8">加载中...</div>
            ) : (
              <div className="border rounded-md">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>激活码</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>关联计划</TableHead>
                      <TableHead>使用者</TableHead>
                      <TableHead>批次号</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {Array.isArray(codes) && codes.map((code) => (
                      <TableRow key={code.id}>
                        <TableCell className="font-mono text-sm">{code.code}</TableCell>
                        <TableCell>
                          <Badge 
                            variant={
                              code.status === "used" ? "default" : 
                              code.status === "expired" ? "destructive" : "secondary"
                            }
                          >
                            {code.status === "used" ? "已使用" : 
                             code.status === "expired" ? "已过期" : "未使用"}
                          </Badge>
                        </TableCell>
                        <TableCell>{code.plan?.title || "-"}</TableCell>
                        <TableCell>{code.used_by?.username || "-"}</TableCell>
                        <TableCell>{code.batch_number}</TableCell>
                        <TableCell>{new Date(code.created_at).toLocaleDateString()}</TableCell>
                        <TableCell>
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => handleDeleteCode(code.id)}
                            disabled={code.status === "used"}
                          >
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </div>

          {/* 激活码生成对话框 */}
          <Dialog open={isCodeDialogOpen} onOpenChange={setIsCodeDialogOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>生成激活码</DialogTitle>
                <DialogDescription>为指定订阅计划生成激活码</DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>订阅计划</Label>
                  <Select
                    value={newCodeData.subscription_plan_id.toString()}
                    onValueChange={(value) => setNewCodeData({...newCodeData, subscription_plan_id: parseInt(value)})}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="选择订阅计划" />
                    </SelectTrigger>
                    <SelectContent>
                      {Array.isArray(plans) && plans.map((plan) => (
                        <SelectItem key={plan.id} value={plan.id.toString()}>
                          {plan.title} ({plan.point_amount.toLocaleString()} 积分, {plan.validity_days}天)
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>生成数量</Label>
                  <Input
                    type="number"
                    min="1"
                    max="100"
                    value={newCodeData.count}
                    onChange={(e) => setNewCodeData({...newCodeData, count: parseInt(e.target.value) || 1})}
                  />
                </div>
                <div className="space-y-2">
                  <Label>批次号 (可选)</Label>
                  <Input
                    value={newCodeData.batch_number}
                    onChange={(e) => setNewCodeData({...newCodeData, batch_number: e.target.value})}
                    placeholder="例如: BATCH-2024-001"
                  />
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsCodeDialogOpen(false)}>取消</Button>
                <Button onClick={handleCreateCodes}>生成</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>
    </DashboardLayout>
  )
} 