"use client"

import { useState, useEffect } from "react"
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
  amount: number;
  currency: string;
  date: string;
  status: string;
  invoiceUrl?: string;
}

export default function SubscriptionPage() {
  const { toast } = useToast()
  const [loading, setLoading] = useState(true)
  const [subscription, setSubscription] = useState<ActiveSubscription | null>(null)
  const [paymentHistory, setPaymentHistory] = useState<PaymentHistory[]>([])
  const [couponCode, setCouponCode] = useState("")
  const [redeemingCoupon, setRedeemingCoupon] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 加载订阅数据
  const loadSubscriptionData = async () => {
    setLoading(true)
    setError(null)

    try {
      const [subscriptionResult, historyResult] = await Promise.all([
        dashboardAPI.getActiveSubscription(),
        getPaymentHistory()
      ])

      if (subscriptionResult.success) {
        setSubscription(subscriptionResult.data || null)
      }

      if (historyResult.success) {
        setPaymentHistory(historyResult.data || [])
      }
    } catch (err: any) {
      setError("加载订阅数据失败")
    }

    setLoading(false)
  }

  // 获取支付历史
  const getPaymentHistory = async (): Promise<{ success: boolean; data?: PaymentHistory[]; message?: string }> => {
    try {
      const response = await request.get("/api/subscription/history")
      return { success: true, data: response.data.history || [] }
    } catch (error: any) {
      console.error("获取支付历史失败:", error)
      return { 
        success: false, 
        message: error.response?.data?.error || "获取支付历史失败" 
      }
    }
  }

  // 兑换激活码
  const handleRedeemCoupon = async () => {
    if (!couponCode.trim()) {
      toast({
        title: "请输入激活码",
        variant: "destructive"
      })
      return
    }

    setRedeemingCoupon(true)

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
        // 重新加载数据
        loadSubscriptionData()
      } else {
        toast({
          title: "兑换失败",
          description: response.data.message,
          variant: "destructive"
        })
      }
    } catch (error: any) {
      toast({
        title: "兑换失败",
        description: error.response?.data?.message || "激活码兑换失败",
        variant: "destructive"
      })
    }

    setRedeemingCoupon(false)
  }

  useEffect(() => {
    loadSubscriptionData()
  }, [])

  const isSubscriptionActive = subscription?.status === "active"

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
              <CardDescription>查看您当前的订阅计划详情。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {loading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin" />
                  <span className="ml-2">加载中...</span>
                </div>
              ) : subscription ? (
                <>
              <div className="flex justify-between items-center">
                    <h3 className="text-xl font-semibold text-sky-500 dark:text-sky-400">
                      {subscription.plan.name}
                    </h3>
                <Badge
                  variant={isSubscriptionActive ? "default" : "destructive"}
                  className={
                    isSubscriptionActive
                      ? "bg-green-500 text-white dark:bg-green-600 dark:text-green-50"
                      : "bg-red-500 text-white dark:bg-red-600 dark:text-red-50"
                  }
                >
                      {isSubscriptionActive ? "有效" : subscription.status === "canceled" ? "已取消" : "已过期"}
                </Badge>
              </div>
                  <p className="text-3xl font-bold">
                    {subscription.plan.currency === "CNY" ? "¥" : "$"}{subscription.plan.pricePerMonth}/月
                  </p>
              {isSubscriptionActive && (
                    <p className="text-sm text-muted-foreground">
                      下次账单日期: {new Date(subscription.currentPeriodEnd).toLocaleDateString()}
                    </p>
              )}
              <div>
                <h4 className="font-medium mb-1">包含功能:</h4>
                <ul className="list-disc list-inside space-y-1 text-sm text-muted-foreground">
                      {subscription.plan.features.map((feature, index) => (
                        <li key={index}>{feature}</li>
                  ))}
                </ul>
              </div>
                </>
              ) : (
                <div className="text-center py-8">
                  <p className="text-lg font-medium text-muted-foreground">暂无套餐</p>
                  <p className="text-sm text-muted-foreground">您可以使用激活码激活套餐</p>
                </div>
              )}
            </CardContent>
          </Card>

          <Card className="shadow-lg bg-card text-card-foreground border-border">
            <CardHeader>
              <CardTitle>订阅历史</CardTitle>
              <CardDescription>查看您过去的订阅和付款记录。</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin" />
                  <span className="ml-2">加载中...</span>
                </div>
              ) : paymentHistory.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow className="border-border">
                    <TableHead>账单ID</TableHead>
                    <TableHead>日期</TableHead>
                    <TableHead>计划</TableHead>
                    <TableHead className="text-right">金额</TableHead>
                    <TableHead className="text-right">状态</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                    {paymentHistory.map((item) => (
                    <TableRow key={item.id} className="border-border">
                      <TableCell className="font-medium">{item.id}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {new Date(item.date).toLocaleDateString()}
                        </TableCell>
                        <TableCell className="text-muted-foreground">{item.planName}</TableCell>
                        <TableCell className="text-right text-muted-foreground">
                          {item.currency === "CNY" ? "¥" : "$"}{item.amount.toFixed(2)}
                        </TableCell>
                      <TableCell className="text-right">
                        <Badge
                            variant={item.status === "paid" ? "default" : "secondary"}
                          className={
                              item.status === "paid"
                              ? "bg-green-100 text-green-700 dark:bg-green-700/30 dark:text-green-300"
                              : "bg-slate-100 text-slate-700 dark:bg-slate-700/30 dark:text-slate-300"
                          }
                        >
                            {item.status === "paid" ? "已支付" : item.status}
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
              )}
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
                onClick={handleRedeemCoupon}
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
        </div>
      </div>
    </DashboardLayout>
  )
}
