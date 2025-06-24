"use client"

import { useEffect, useState } from "react"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Progress } from "@/components/ui/progress"
import { Sparkles, CreditCard } from "lucide-react"
import { creditsAPI } from "@/api/credits"
import { Button } from "@/components/ui/button"
import { useRouter } from "next/navigation"

export function CreditSummary() {
  const [creditData, setCreditData] = useState<{available: number, total: number} | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const router = useRouter()
  
  useEffect(() => {
    const fetchCreditBalance = async () => {
      try {
        const response = await creditsAPI.getBalance()
        if (response.success && response.data) {
          setCreditData({
            available: response.data.available_points,
            total: response.data.total_points
          })
        } else {
          setError("无法加载积分信息")
        }
      } catch (err) {
        console.error("获取积分余额失败:", err)
        setError("无法加载积分信息")
      } finally {
        setLoading(false)
      }
    }
    
    fetchCreditBalance()
  }, [])
  
  const handleViewCredits = () => {
    router.push("/credits")
  }
  
  return (
    <Card className="overflow-hidden border border-slate-200 dark:border-slate-700">
      <CardHeader className="bg-gradient-to-r from-sky-600 to-indigo-600 text-white">
        <CardTitle className="flex items-center gap-2">
          <Sparkles className="h-5 w-5" />
          积分概况
        </CardTitle>
        <CardDescription className="text-sky-100">您的API使用额度</CardDescription>
      </CardHeader>
      <CardContent className="pt-6">
        {loading ? (
          <div className="space-y-3">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-4 w-full" />
          </div>
        ) : error ? (
          <div className="text-red-500">{error}</div>
        ) : (
          <div className="space-y-4">
            <div className="flex items-end justify-between">
              <div>
                <div className="text-3xl font-bold">{creditData?.available.toLocaleString()}</div>
                <div className="text-sm text-slate-500 dark:text-slate-400">可用积分</div>
              </div>
              <div className="text-right">
                <div className="text-lg font-semibold">{creditData?.total.toLocaleString()}</div>
                <div className="text-xs text-slate-500 dark:text-slate-400">总充值积分</div>
              </div>
            </div>
            
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>使用进度</span>
                <span>{creditData ? Math.floor((creditData.available / creditData.total) * 100) : 0}%</span>
              </div>
              <Progress 
                value={creditData ? (creditData.available / creditData.total) * 100 : 0} 
                className="h-2" 
              />
            </div>
            
            <Button 
              variant="outline" 
              className="w-full" 
              size="sm"
              onClick={handleViewCredits}
            >
              <CreditCard className="h-4 w-4 mr-2" />
              查看详情
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}