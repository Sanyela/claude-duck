"use client"

import { useState, useEffect, useCallback } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import { CreditCard, Plus, Edit, Trash2 } from "lucide-react"
import { toast } from "sonner"
import { adminAPI, type SubscriptionPlan } from "@/api/admin"

export default function AdminPlansPage() {
  
  const [loading, setLoading] = useState(false)
  const [plans, setPlans] = useState<SubscriptionPlan[]>([])
  const [editingPlan, setEditingPlan] = useState<SubscriptionPlan | null>(null)
  const [isPlanDialogOpen, setIsPlanDialogOpen] = useState(false)
  const [isCreatePlanMode, setIsCreatePlanMode] = useState(false)

  const loadPlans = useCallback(async () => {
    setLoading(true)
    const result = await adminAPI.getSubscriptionPlans()
    if (result.success && result.plans) {
      setPlans(Array.isArray(result.plans) ? result.plans : [])
    } else {
      setPlans([])
      toast({
        title: "加载失败",
        description: result.message,
        variant: "destructive"
      })
    }
    setLoading(false)
  }, [toast])

  useEffect(() => {
    loadPlans()
  }, [loadPlans])

  const handleSavePlan = async (plan: SubscriptionPlan) => {
    const result = isCreatePlanMode
      ? await adminAPI.createSubscriptionPlan(plan)
      : await adminAPI.updateSubscriptionPlan(plan.id, plan)
    
    if (result.success) {
      toast({ title: isCreatePlanMode ? "创建成功" : "更新成功", variant: "default" })
      setIsPlanDialogOpen(false)
      loadPlans()
    } else {
      toast({
        title: isCreatePlanMode ? "创建失败" : "更新失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  const handleDeletePlan = async (planId: number) => {
    if (!confirm("确定要删除此订阅计划吗？")) return
    
    const result = await adminAPI.deleteSubscriptionPlan(planId)
    if (result.success) {
      toast({ title: "删除成功", variant: "default" })
      loadPlans()
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
            <CreditCard className="h-6 w-6 text-purple-500" />
            <CardTitle className="text-slate-900 dark:text-slate-100">订阅计划管理</CardTitle>
          </div>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            管理系统的订阅计划和定价
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <h3 className="text-lg font-semibold">计划列表</h3>
              <Button
                onClick={() => {
                  setEditingPlan({
                    title: "",
                    description: "",
                    point_amount: 0,
                    price: 0,
                    currency: "CNY",
                    validity_days: 30,
                    degradation_guaranteed: 0,
                    daily_checkin_points: 0,
                    daily_checkin_points_max: 0,
                    features: "[]",
                    active: true
                  } as SubscriptionPlan)
                  setIsCreatePlanMode(true)
                  setIsPlanDialogOpen(true)
                }}
                className="bg-purple-500 hover:bg-purple-600 text-white"
              >
                <Plus className="w-4 h-4 mr-2" />
                新建计划
              </Button>
            </div>
            
            {loading ? (
              <div className="text-center py-8">加载中...</div>
            ) : (
              <div className="border rounded-md">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>标题</TableHead>
                      <TableHead>积分数量</TableHead>
                      <TableHead>价格</TableHead>
                      <TableHead>有效期</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {Array.isArray(plans) && plans.map((plan) => (
                      <TableRow key={plan.id}>
                        <TableCell>{plan.title}</TableCell>
                        <TableCell>{plan.point_amount.toLocaleString()}</TableCell>
                        <TableCell>{plan.price} {plan.currency}</TableCell>
                        <TableCell>{plan.validity_days} 天</TableCell>
                        <TableCell>
                          <Badge variant={plan.active ? "default" : "secondary"}>
                            {plan.active ? "启用" : "禁用"}
                          </Badge>
                        </TableCell>
                        <TableCell className="space-x-2">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => {
                              setEditingPlan(plan)
                              setIsCreatePlanMode(false)
                              setIsPlanDialogOpen(true)
                            }}
                          >
                            <Edit className="w-4 h-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => handleDeletePlan(plan.id)}
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

          {/* 订阅计划编辑对话框 */}
          <Dialog open={isPlanDialogOpen} onOpenChange={setIsPlanDialogOpen}>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>{isCreatePlanMode ? "创建订阅计划" : "编辑订阅计划"}</DialogTitle>
                <DialogDescription>设置订阅计划的详细信息</DialogDescription>
              </DialogHeader>
              {editingPlan && (
                <div className="space-y-4 max-h-96 overflow-y-auto px-2">
                  {!isCreatePlanMode && (
                    <div className="space-y-2">
                      <Label>数据库ID</Label>
                      <Input
                        value={editingPlan.id?.toString() || ""}
                        disabled
                        className="bg-gray-50"
                      />
                      <p className="text-sm text-gray-500">系统自动生成，不可修改</p>
                    </div>
                  )}
                  <div className="space-y-2">
                    <Label>标题</Label>
                    <Input
                      value={editingPlan.title}
                      onChange={(e) => setEditingPlan({...editingPlan, title: e.target.value})}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>描述</Label>
                    <Textarea
                      value={editingPlan.description}
                      onChange={(e) => setEditingPlan({...editingPlan, description: e.target.value})}
                      className="resize-none"
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>积分数量</Label>
                      <Input
                        type="number"
                        value={editingPlan.point_amount}
                        onChange={(e) => setEditingPlan({...editingPlan, point_amount: parseInt(e.target.value) || 0})}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>价格</Label>
                      <Input
                        type="number"
                        step="0.01"
                        value={editingPlan.price}
                        onChange={(e) => setEditingPlan({...editingPlan, price: parseFloat(e.target.value) || 0})}
                      />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label>货币</Label>
                      <Select
                        value={editingPlan.currency}
                        onValueChange={(value) => setEditingPlan({...editingPlan, currency: value})}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="CNY">CNY</SelectItem>
                          <SelectItem value="USD">USD</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="space-y-2">
                      <Label>有效期(天)</Label>
                      <Input
                        type="number"
                        value={editingPlan.validity_days}
                        onChange={(e) => setEditingPlan({...editingPlan, validity_days: parseInt(e.target.value) || 0})}
                      />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label>保证不降级的数量</Label>
                    <Input
                      type="number"
                      value={editingPlan.degradation_guaranteed}
                      onChange={(e) => setEditingPlan({...editingPlan, degradation_guaranteed: parseInt(e.target.value) || 0})}
                      placeholder="输入保证不降级的消息条数"
                    />
                    <p className="text-sm text-gray-500">设置10条消息内，保证使用多少条不会降级</p>
                  </div>
                  <div className="space-y-2">
                    <Label>每日签到积分（最低值）</Label>
                    <Input
                      type="number"
                      value={editingPlan.daily_checkin_points}
                      onChange={(e) => setEditingPlan({...editingPlan, daily_checkin_points: parseInt(e.target.value) || 0})}
                      placeholder="输入每日签到奖励积分最低值"
                    />
                    <p className="text-sm text-gray-500">用户拥有此订阅时，每日签到可获得的积分奖励最低值（0表示不奖励）</p>
                  </div>
                  <div className="space-y-2">
                    <Label>每日签到积分（最高值）</Label>
                    <Input
                      type="number"
                      value={editingPlan.daily_checkin_points_max}
                      onChange={(e) => setEditingPlan({...editingPlan, daily_checkin_points_max: parseInt(e.target.value) || 0})}
                      placeholder="输入每日签到奖励积分最高值"
                    />
                    <p className="text-sm text-gray-500">签到积分范围的最高值，0或小于最低值时将设为与最低值相同</p>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Switch
                      checked={editingPlan.active}
                      onCheckedChange={(checked) => setEditingPlan({...editingPlan, active: checked})}
                    />
                    <Label>启用计划</Label>
                  </div>
                </div>
              )}
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsPlanDialogOpen(false)}>取消</Button>
                <Button onClick={() => editingPlan && handleSavePlan(editingPlan)}>
                  {isCreatePlanMode ? "创建" : "保存"}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>
    </DashboardLayout>
  )
} 