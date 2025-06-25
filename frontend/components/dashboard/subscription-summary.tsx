"use client"

import { useEffect, useState } from "react"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Badge } from "@/components/ui/badge"
import { Package, CalendarClock, ArrowRight } from "lucide-react"
import { getActiveSubscription } from "@/api/subscription"
import { Button } from "@/components/ui/button"
import { useRouter } from "next/navigation"
import { format, formatDistanceToNow } from "date-fns"
import { zhCN } from "date-fns/locale"

interface ActiveSubscription {
  id: string
  plan: {
    id: string
    name: string
    currency: string
    features: string[]
  }
  status: 'active' | 'canceled' | 'past_due'
  currentPeriodEnd: string
  cancelAtPeriodEnd: boolean
}

export function SubscriptionSummary() {
  const [subscription, setSubscription] = useState<ActiveSubscription | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const router = useRouter()
  
  useEffect(() => {
    const fetchSubscription = async () => {
      try {
        const response = await getActiveSubscription()
        setSubscription(response.subscription)
      } catch (err) {
        console.error("获取订阅信息失败:", err)
        setError("无法加载订阅信息")
      } finally {
        setLoading(false)
      }
    }
    
    fetchSubscription()
  }, [])
  
  const handleViewSubscription = () => {
    router.push("/subscription")
  }
  
  const getRemainingTime = (dateStr: string) => {
    try {
      const endDate = new Date(dateStr)
      return formatDistanceToNow(endDate, { locale: zhCN, addSuffix: true })
    } catch (e) {
      return "未知"
    }
  }
  
  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge className="bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100">有效</Badge>
      case 'canceled':
        return <Badge variant="outline" className="border-amber-500 text-amber-600 dark:text-amber-400">已取消</Badge>
      case 'past_due':
        return <Badge variant="destructive">已过期</Badge>
      default:
        return null
    }
  }
  
  return (
    <Card className="overflow-hidden border border-slate-200 dark:border-slate-700">
      <CardHeader className="bg-gradient-to-r from-purple-600 to-pink-600 text-white">
        <CardTitle className="flex items-center gap-2">
          <Package className="h-5 w-5" />
          订阅信息
        </CardTitle>
        <CardDescription className="text-purple-100">您的当前套餐</CardDescription>
      </CardHeader>
      <CardContent className="pt-6">
        {loading ? (
          <div className="space-y-3">
            <Skeleton className="h-8 w-3/4" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-full" />
          </div>
        ) : error ? (
          <div className="text-red-500">{error}</div>
        ) : !subscription ? (
          <div className="space-y-4">
            <div className="text-center py-2">
              <div className="text-slate-500 dark:text-slate-400 mb-2">当前无活跃订阅</div>
              <Button 
                onClick={handleViewSubscription}
                variant="default"
                className="w-full"
              >
                查看订阅计划
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="text-xl font-bold">{subscription.plan.name}</div>
              {getStatusBadge(subscription.status)}
            </div>
            
            <div className="flex items-center text-sm text-slate-500 dark:text-slate-400 space-x-1">
              <CalendarClock className="h-4 w-4" />
              <span>
                到期{getRemainingTime(subscription.currentPeriodEnd)}
              </span>
            </div>
            
            {subscription.cancelAtPeriodEnd && (
              <div className="text-xs text-amber-600 dark:text-amber-400">
                此订阅将在周期结束后自动取消
              </div>
            )}
            
            <Button 
              variant="outline" 
              className="w-full" 
              size="sm"
              onClick={handleViewSubscription}
            >
              <Package className="h-4 w-4 mr-2" />
              管理订阅
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}