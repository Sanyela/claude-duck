"use client"

export const dynamic = 'force-dynamic'

import { useState, useEffect, useCallback } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Tag, Loader2, AlertCircle } from "lucide-react"
import { useToast } from "@/components/ui/use-toast"
import { dashboardAPI, type ActiveSubscription } from "@/api/dashboard"
import { request } from "@/api/request"

// 支付历史接口定义
interface PaymentHistory {
  id: string;
  planName: string;
  date: string;
  paymentStatus: string; // 支付状态：paid, failed
  subscriptionStatus: string; // 订阅状态：active, expired
  invoiceUrl?: string;
}

export default function SubscriptionPage() {
  const { toast } = useToast()
  const [loading, setLoading] = useState(true)
  const [subscriptions, setSubscriptions] = useState<ActiveSubscription[]>([])
  const [paymentHistory, setPaymentHistory] = useState<PaymentHistory[]>([])
  const [couponCode, setCouponCode] = useState("")
  const [redeemingCoupon, setRedeemingCoupon] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [showWarning, setShowWarning] = useState(false)
  const [warningInfo, setWarningInfo] = useState({ 
    serviceLevel: '', 
    warning: '', 
    currentPoints: 0,
    newPoints: 0,
    totalPointsAfter: 0,
    currentCheckinMin: 0,
    currentCheckinMax: 0,
    newCheckinMin: 0,
    newCheckinMax: 0,
    currentAutoRefill: 0,
    newAutoRefill: 0
  })
  const [countdown, setCountdown] = useState(0)

  // 加载订阅数据
  const loadSubscriptionData = useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      const [subscriptionResult, historyResult] = await Promise.all([
        dashboardAPI.getActiveSubscription(),
        getPaymentHistory()
      ])

      if (subscriptionResult.success) {
        setSubscriptions(subscriptionResult.data || [])
      }

      if (historyResult.success) {
        setPaymentHistory(historyResult.data || [])
      }
    } catch {
      setError("加载订阅数据失败")
    }

    setLoading(false)
  }, [])

  // 获取支付历史
  const getPaymentHistory = async (): Promise<{ success: boolean; data?: PaymentHistory[]; message?: string }> => {
    try {
      const response = await request.get("/api/subscription/history")
      return { success: true, data: response.data.history || [] }
    } catch {
      console.error("获取支付历史失败:", error)
      return { 
        success: false, 
        message: error.response?.data?.error || "获取支付历史失败" 
      }
    }
  }

  // 预检查激活码并显示警告
  const handlePreCheckCoupon = async () => {
    if (!couponCode.trim()) {
      toast({
        title: "请输入激活码",
        variant: "destructive"
      })
      return
    }

    setRedeemingCoupon(true)

    try {
      // 调用预检查接口，不执行实际兑换
      const response = await request.post("/api/subscription/redeem/preview", {
        couponCode: couponCode.trim()
      })

      if (response.data.success) {
        const { 
          serviceLevel, 
          warning, 
          currentPoints, 
          newPoints, 
          totalPointsAfter,
          currentCheckinMin,
          currentCheckinMax,
          newCheckinMin,
          newCheckinMax,
          currentAutoRefill,
          newAutoRefill
        } = response.data
        
        if ((serviceLevel === 'same_level' || serviceLevel === 'downgrade') && warning) {
          // 显示警告并开始倒计时
          setWarningInfo({ 
            serviceLevel, 
            warning, 
            currentPoints: currentPoints || 0,
            newPoints: newPoints || 0,
            totalPointsAfter: totalPointsAfter || 0,
            currentCheckinMin: currentCheckinMin || 0,
            currentCheckinMax: currentCheckinMax || 0,
            newCheckinMin: newCheckinMin || 0,
            newCheckinMax: newCheckinMax || 0,
            currentAutoRefill: currentAutoRefill || 0,
            newAutoRefill: newAutoRefill || 0
          })
          setShowWarning(true)
          setCountdown(10)
          
          // 倒计时
          const timer = setInterval(() => {
            setCountdown(prev => {
              if (prev <= 1) {
                clearInterval(timer)
                return 0
              }
              return prev - 1
            })
          }, 1000)
        } else {
          // 升级情况，直接执行兑换
          await executeActualRedeem()
        }
      } else {
        toast({
          title: "预检查失败",
          description: response.data.message,
          variant: "destructive"
        })
      }
    } catch {
      toast({
        title: "预检查失败",
        description: error.response?.data?.message || "预检查失败",
        variant: "destructive"
      })
    }

    setRedeemingCoupon(false)
  }

  // 执行实际兑换操作
  const executeActualRedeem = async () => {
    try {
      const response = await request.post("/api/subscription/redeem", {
        couponCode: couponCode.trim()
      })

      if (response.data.success) {
        toast({
          title: "兑换成功",
          description: response.data.message,
          variant: "default"
        })
        setCouponCode("")
        setShowWarning(false)
        setWarningInfo({ 
          serviceLevel: '', 
          warning: '', 
          currentPoints: 0,
          newPoints: 0,
          totalPointsAfter: 0,
          currentCheckinMin: 0,
          currentCheckinMax: 0,
          newCheckinMin: 0,
          newCheckinMax: 0,
          currentAutoRefill: 0,
          newAutoRefill: 0
        })
        setCountdown(0)
        loadSubscriptionData()
      } else {
        toast({
          title: "兑换失败",
          description: response.data.message,
          variant: "destructive"
        })
      }
    } catch {
      toast({
        title: "兑换失败",
        description: error.response?.data?.message || "激活码兑换失败",
        variant: "destructive"
      })
    }
  }

  // 确认兑换（在警告倒计时结束后）
  const handleConfirmRedeem = async () => {
    setRedeemingCoupon(true)
    await executeActualRedeem()
    setRedeemingCoupon(false)
  }

  // 取消兑换
  const handleCancelRedeem = () => {
    setShowWarning(false)
    setWarningInfo({ 
      serviceLevel: '', 
      warning: '', 
      currentPoints: 0,
      newPoints: 0,
      totalPointsAfter: 0,
      currentCheckinMin: 0,
      currentCheckinMax: 0,
      newCheckinMin: 0,
      newCheckinMax: 0,
      currentAutoRefill: 0,
      newAutoRefill: 0
    })
    setCountdown(0)
  }

  useEffect(() => {
    loadSubscriptionData()
  }, [loadSubscriptionData])


  return (
    <DashboardLayout>
      {error && (
        <Alert variant="destructive" className="mb-6">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <Card className="shadow-lg bg-card text-card-foreground border-border">
            <CardHeader>
              <CardTitle>当前订阅</CardTitle>
              <CardDescription>查看您当前有效的订阅计划详情。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {loading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin" />
                  <span className="ml-2">加载中...</span>
                </div>
              ) : (() => {
                const validSubscriptions = subscriptions.filter(sub => sub.detailedStatus === "有效");
                return validSubscriptions.length > 0 ? (
                  <div className="space-y-4">
                    {validSubscriptions.map((subscription) => (
                      <div key={subscription.id} className={`p-4 rounded-lg border ${
                        subscription.isCurrentUsing ? 'border-blue-300 bg-blue-50 dark:bg-blue-900/20' : 'border-border bg-card'
                      }`}>
                        <div className="flex justify-between items-start mb-3">
                          <div>
                            <h3 className="text-lg font-semibold text-sky-500 dark:text-sky-400">
                              {subscription.plan.name}
                            </h3>
                            {subscription.isCurrentUsing && (
                              <Badge className="bg-blue-500 text-white text-xs mt-1">
                                当前消耗
                              </Badge>
                            )}
                          </div>
                          <Badge className="bg-green-500 text-white dark:bg-green-600 dark:text-green-50">
                            有效
                          </Badge>
                        </div>
                        
                        <div className="grid grid-cols-3 gap-4 text-sm">
                          <div>
                            <p className="text-muted-foreground">可用积分</p>
                            <p className="font-semibold">{subscription.availablePoints.toLocaleString()}</p>
                          </div>
                          <div>
                            <p className="text-muted-foreground">总积分</p>
                            <p className="font-semibold">{subscription.totalPoints.toLocaleString()}</p>
                          </div>
                          <div>
                            <p className="text-muted-foreground">已使用</p>
                            <p className="font-semibold">{subscription.usedPoints.toLocaleString()}</p>
                          </div>
                        </div>
                        
                        <div className="mt-3 text-xs text-muted-foreground">
                          <p>激活时间: {new Date(subscription.activatedAt).toLocaleDateString()}</p>
                          <p>到期时间: {new Date(subscription.currentPeriodEnd).toLocaleDateString()}</p>
                        </div>
                        
                        <div className="mt-3">
                          <h4 className="font-medium mb-2 text-sm">核心功能:</h4>
                          <ul className="list-disc list-inside space-y-1 text-xs text-muted-foreground">
                            <li>模型智能不降级，保证回答质量</li>
                            <li>享受完整Claude 4 Sonnet能力</li>
                            <li>优先处理请求，响应更快</li>
                          </ul>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <p className="text-lg font-medium text-muted-foreground">暂无有效套餐</p>
                    <p className="text-sm text-muted-foreground">您可以使用激活码激活套餐</p>
                  </div>
                );
              })()}
            </CardContent>
          </Card>

          <Card className="shadow-lg bg-card text-card-foreground border-border">
            <CardHeader>
              <CardTitle>订阅历史</CardTitle>
              <CardDescription>查看您所有的订阅记录，包括已过期和已用完的订阅。</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin" />
                  <span className="ml-2">加载中...</span>
                </div>
              ) : (() => {
                const allSubscriptionsHistory = subscriptions.length > 0 
                  ? subscriptions.map(sub => ({
                      id: sub.id,
                      planName: sub.plan.name,
                      date: sub.activatedAt,
                      paymentStatus: "paid",
                      subscriptionStatus: sub.detailedStatus,
                      invoiceUrl: ""
                    }))
                  : paymentHistory;
                
                return allSubscriptionsHistory.length > 0 ? (
                  <Table>
                    <TableHeader>
                      <TableRow className="border-border">
                        <TableHead>订阅ID</TableHead>
                        <TableHead>日期</TableHead>
                        <TableHead>计划</TableHead>
                        <TableHead>支付状态</TableHead>
                        <TableHead className="text-right">订阅状态</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {allSubscriptionsHistory.map((item) => (
                        <TableRow key={item.id} className="border-border">
                          <TableCell className="font-medium">{item.id}</TableCell>
                          <TableCell className="text-muted-foreground">
                            {new Date(item.date).toLocaleDateString()}
                          </TableCell>
                          <TableCell className="text-muted-foreground">{item.planName}</TableCell>
                          <TableCell>
                            <Badge
                              variant={item.paymentStatus === "paid" ? "default" : "destructive"}
                              className={
                                item.paymentStatus === "paid"
                                  ? "bg-green-100 text-green-700 dark:bg-green-700/30 dark:text-green-300"
                                  : "bg-red-100 text-red-700 dark:bg-red-700/30 dark:text-red-300"
                              }
                            >
                              {item.paymentStatus === "paid" ? "支付成功" : "支付失败"}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-right">
                            <Badge
                              variant={item.subscriptionStatus === "有效" ? "default" : 
                                       item.subscriptionStatus === "已用完" ? "secondary" : "outline"}
                              className={
                                item.subscriptionStatus === "有效"
                                  ? "bg-blue-100 text-blue-700 dark:bg-blue-700/30 dark:text-blue-300"
                                  : item.subscriptionStatus === "已用完"
                                  ? "bg-orange-100 text-orange-700 dark:bg-orange-700/30 dark:text-orange-300"
                                  : "bg-gray-100 text-gray-700 dark:bg-gray-700/30 dark:text-gray-300"
                              }
                            >
                              {item.subscriptionStatus}
                            </Badge>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                ) : (
                  <div className="text-center py-8">
                    <p className="text-muted-foreground">暂无订阅历史</p>
                  </div>
                );
              })()}
            </CardContent>
          </Card>
        </div>

        <div className="lg:col-span-1 space-y-6">
          <Card className="shadow-lg bg-card text-card-foreground border-border">
            <CardHeader>
              <CardTitle>兑换激活码</CardTitle>
              <CardDescription>有激活码吗？在这里输入兑换套餐。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              <Label htmlFor="coupon">激活码</Label>
              <Input
                id="coupon"
                placeholder="例如：ABCD-EFGH-IJKL-MNOP"
                value={couponCode}
                onChange={(e) => setCouponCode(e.target.value)}
                className="bg-input border-border placeholder:text-muted-foreground"
                disabled={redeemingCoupon}
                autoComplete="off"
              />
            </CardContent>
            <CardFooter>
              <Button 
                className="w-full bg-sky-500 hover:bg-sky-600 text-primary-foreground"
                onClick={handlePreCheckCoupon}
                disabled={redeemingCoupon || !couponCode.trim()}
              >
                {redeemingCoupon ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    兑换中...
                  </>
                ) : (
                  <>
                    <Tag className="mr-2 h-4 w-4" />
                    兑换激活码
                  </>
                )}
              </Button>
            </CardFooter>
          </Card>

          {/* 警告对话框 */}
          {showWarning && (
            <Card className={`mt-6 ${
              warningInfo.serviceLevel === 'downgrade' 
                ? 'border-red-300 bg-red-50 dark:border-red-700 dark:bg-red-900/20' 
                : 'border-amber-200 bg-amber-50 dark:border-amber-700 dark:bg-amber-900/20'
            }`}>
              <CardHeader>
                <CardTitle className={`flex items-center gap-2 ${
                  warningInfo.serviceLevel === 'downgrade' 
                    ? 'text-red-800 dark:text-red-400' 
                    : 'text-amber-800 dark:text-amber-400'
                }`}>
                  <AlertCircle className="h-5 w-5" />
                  {warningInfo.serviceLevel === 'same_level' ? '同级兑换警告' : '降级兑换警告'}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <Alert className={`${
                  warningInfo.serviceLevel === 'downgrade'
                    ? 'border-red-300 bg-red-50 dark:border-red-700 dark:bg-red-900/20'
                    : 'border-amber-200 bg-amber-50 dark:border-amber-700 dark:bg-amber-900/20'
                }`}>
                  <AlertDescription className={`${
                    warningInfo.serviceLevel === 'downgrade'
                      ? 'text-red-800 font-medium dark:text-red-400'
                      : 'text-amber-800 font-medium dark:text-amber-400'
                  }`}>
                    {warningInfo.warning}
                  </AlertDescription>
                </Alert>
                
                {/* 权益对比表 */}
                <div className="mt-4 space-y-4">
                  <div className={`bg-white dark:bg-gray-800 rounded-lg border overflow-hidden ${
                    warningInfo.serviceLevel === 'downgrade' 
                      ? 'border-red-200 dark:border-red-700' 
                      : 'border-amber-200 dark:border-amber-700'
                  }`}>
                    <div className={`px-4 py-3 border-b ${
                      warningInfo.serviceLevel === 'downgrade'
                        ? 'bg-red-50 dark:bg-red-900/30 border-red-200 dark:border-red-700'
                        : 'bg-amber-50 dark:bg-amber-900/30 border-amber-200 dark:border-amber-700'
                    }`}>
                      <h4 className={`font-semibold text-sm flex items-center gap-2 ${
                        warningInfo.serviceLevel === 'downgrade'
                          ? 'text-red-800 dark:text-red-400'
                          : 'text-amber-800 dark:text-amber-400'
                      }`}>
                        ⚖️ 权益对比分析
                      </h4>
                    </div>
                    <div className="p-4">
                      <div className="grid grid-cols-3 gap-4 text-sm">
                        {/* 表头 */}
                        <div className="font-medium text-gray-600 dark:text-gray-400">权益项目</div>
                        <div className="font-medium text-green-600 dark:text-green-400 text-center">您当前享有</div>
                        <div className="font-medium text-blue-600 dark:text-blue-400 text-center">兑换后状态</div>
                        
                        {/* 分隔线 */}
                        <div className="col-span-3 border-t border-gray-200 dark:border-gray-700 my-2"></div>
                        
                        {/* 积分数量 */}
                        <div className="text-gray-700 dark:text-gray-300">💰 积分数量</div>
                        <div className="text-center text-green-600 dark:text-green-400 font-medium">
                          {warningInfo.currentPoints.toLocaleString()}
                        </div>
                        <div className="text-center text-blue-600 dark:text-blue-400 font-medium">
                          {warningInfo.serviceLevel === 'same_level' ? (
                            <span>{warningInfo.newPoints.toLocaleString()}</span>
                          ) : (
                            <span>{warningInfo.currentPoints.toLocaleString()} + {warningInfo.newPoints.toLocaleString()} = {warningInfo.totalPointsAfter.toLocaleString()}</span>
                          )}
                        </div>
                        
                        {/* 签到奖励 */}
                        <div className="text-gray-700 dark:text-gray-300">🎁 每日签到</div>
                        <div className="text-center text-green-600 dark:text-green-400 font-medium">
                          {warningInfo.currentCheckinMin > 0 ? (
                            warningInfo.currentCheckinMin === warningInfo.currentCheckinMax ? 
                              `${warningInfo.currentCheckinMin}积分` : 
                              `${warningInfo.currentCheckinMin}-${warningInfo.currentCheckinMax}积分`
                          ) : '未配置'}
                        </div>
                        <div className="text-center font-medium">
                          {warningInfo.serviceLevel === 'downgrade' ? (
                            <span className="text-green-600 dark:text-green-400">保持当前配置</span>
                          ) : (
                            <span className="text-blue-600 dark:text-blue-400">
                              {warningInfo.newCheckinMin > 0 ? (
                                warningInfo.newCheckinMin === warningInfo.newCheckinMax ? 
                                  `${warningInfo.newCheckinMin}积分` : 
                                  `${warningInfo.newCheckinMin}-${warningInfo.newCheckinMax}积分`
                              ) : '未配置'}
                            </span>
                          )}
                        </div>
                        
                        {/* 自动补给 */}
                        <div className="text-gray-700 dark:text-gray-300">🔄 自动补给</div>
                        <div className="text-center text-green-600 dark:text-green-400 font-medium">
                          {warningInfo.currentAutoRefill > 0 ? `${warningInfo.currentAutoRefill}积分` : '未配置'}
                        </div>
                        <div className="text-center font-medium">
                          {warningInfo.serviceLevel === 'downgrade' ? (
                            <span className="text-green-600 dark:text-green-400">保持当前配置</span>
                          ) : (
                            <span className="text-blue-600 dark:text-blue-400">
                              {warningInfo.newAutoRefill > 0 ? `${warningInfo.newAutoRefill}积分` : '未配置'}
                            </span>
                          )}
                        </div>
                        
                        {/* 服务等级 */}
                        <div className="text-gray-700 dark:text-gray-300">⭐ 服务等级</div>
                        <div className="text-center text-green-600 dark:text-green-400 font-medium">
                          当前等级
                        </div>
                        <div className="text-center font-medium">
                          {warningInfo.serviceLevel === 'downgrade' ? (
                            <span className="text-green-600 dark:text-green-400">保持当前等级</span>
                          ) : (
                            <span className="text-blue-600 dark:text-blue-400">同等级</span>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                
                <div className="mt-4 text-center">
                  <p className="text-sm text-muted-foreground mb-3">
                    请仔细阅读上述警告，确认无误后等待倒计时结束
                  </p>
                  <div className={`text-2xl font-bold mb-4 ${
                    warningInfo.serviceLevel === 'downgrade'
                      ? 'text-red-600 dark:text-red-400'
                      : 'text-amber-600 dark:text-amber-400'
                  }`}>
                    {countdown > 0 ? `${countdown}秒` : '现在可以确认兑换'}
                  </div>
                  <div className="flex gap-3 justify-center">
                    <Button 
                      variant="outline"
                      onClick={handleCancelRedeem}
                      className={`${
                        warningInfo.serviceLevel === 'downgrade'
                          ? 'border-red-300 text-red-700 hover:bg-red-100 dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/30'
                          : 'border-amber-300 text-amber-700 hover:bg-amber-100 dark:border-amber-700 dark:text-amber-400 dark:hover:bg-amber-900/30'
                      }`}
                    >
                      取消兑换
                    </Button>
                    <Button 
                      onClick={handleConfirmRedeem}
                      disabled={countdown > 0 || redeemingCoupon}
                      className={`${
                        warningInfo.serviceLevel === 'downgrade'
                          ? 'bg-red-600 hover:bg-red-700 text-white dark:bg-red-700 dark:hover:bg-red-800'
                          : 'bg-amber-600 hover:bg-amber-700 text-white dark:bg-amber-700 dark:hover:bg-amber-800'
                      }`}
                    >
                      {redeemingCoupon ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          兑换中...
                        </>
                      ) : countdown > 0 ? (
                        `等待 ${countdown}s`
                      ) : (
                        '确认兑换'
                      )}
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </DashboardLayout>
  )
}
