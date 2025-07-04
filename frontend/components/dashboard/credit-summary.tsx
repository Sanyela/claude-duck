"use client"

import { useEffect, useState } from "react"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Progress } from "@/components/ui/progress"
import { Sparkles, CreditCard, Clock, Zap } from "lucide-react"
import { creditsAPI } from "@/api/credits"
import { Button } from "@/components/ui/button"
import { useRouter } from "next/navigation"
import { Badge } from "@/components/ui/badge"

// 自动补给信息接口
interface AutoRefillInfo {
  enabled: boolean;           // 是否启用自动补给
  threshold: number;          // 补给阈值
  amount: number;             // 补给数量
  needs_refill: boolean;      // 当前是否需要补给
  next_refill_time?: string;  // 下次补给时间
  last_refill_time?: string;  // 上次补给时间
}

// 后端返回的数据结构
interface CreditBalanceResponse {
  available: number;
  total: number;
  used: number;
  expired: number;
  is_current_subscription: boolean;
  free_model_usage_count: number;
  checkin_points: number;
  admin_gift_points: number;
  auto_refill?: AutoRefillInfo;  // 自动补给信息
}

export function CreditSummary() {
  const [creditData, setCreditData] = useState<CreditBalanceResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const router = useRouter()
  
  useEffect(() => {
    const fetchCreditBalance = async () => {
      try {
        const response = await creditsAPI.getBalance()
        if (response.success && response.data) {
          const balanceData = {
            available: response.data.available_points || 0,
            total: response.data.total_points || 0,
            used: response.data.used_points || 0,
            expired: response.data.expired_points || 0,
            is_current_subscription: response.data.is_current_subscription || false,
            free_model_usage_count: response.data.free_model_usage_count || 0,
            checkin_points: response.data.checkin_points || 0,
            admin_gift_points: response.data.admin_gift_points || 0,
            auto_refill: response.data.auto_refill || undefined,
          }
          console.log("积分数据:", balanceData) // 添加调试日志
          setCreditData(balanceData)
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
            
            {/* 积分详情：总数量 | 已使用 | 签到积分 | 管理员 */}
            <div className="text-sm text-slate-600 dark:text-slate-400">
              <span>积分总数: <span className="font-semibold text-slate-700 dark:text-slate-300">{creditData?.total.toLocaleString() || '0'}</span></span>
              <span className="mx-2">|</span>
              <span>已使用: <span className="font-semibold text-slate-700 dark:text-slate-300">{creditData?.used.toLocaleString() || '0'}</span></span>
              {creditData && creditData.checkin_points > 0 && (
                <>
                  <span className="mx-2">|</span>
                  <span>签到积分: <span className="font-semibold text-purple-600 dark:text-purple-400">{creditData.checkin_points.toLocaleString()}</span></span>
                </>
              )}
              {creditData && creditData.admin_gift_points > 0 && (
                <>
                  <span className="mx-2">|</span>
                  <span>管理员: <span className="font-semibold text-amber-600 dark:text-amber-400">{creditData.admin_gift_points.toLocaleString()}</span></span>
                </>
              )}
              {/* 调试信息 */}
              {process.env.NODE_ENV === 'development' && (
                <div className="text-xs text-gray-400 mt-1">
                  调试: 签到{creditData?.checkin_points || 0} | 管理员{creditData?.admin_gift_points || 0}
                </div>
              )}
            </div>
            
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>使用进度</span>
                <span>{creditData && creditData.total > 0 ? Math.round((creditData.used / creditData.total) * 100) : 0}%</span>
              </div>
              <Progress 
                value={creditData && creditData.total > 0 ? (creditData.used / creditData.total) * 100 : 0} 
                className="h-2" 
              />
            </div>

            {/* 自动补给信息 */}
            {creditData?.auto_refill?.enabled && (
              <div className="bg-gradient-to-r from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 rounded-lg p-3 border border-emerald-200 dark:border-emerald-800">
                <div className="flex items-center gap-2 mb-2">
                  <Zap className="h-4 w-4 text-emerald-600" />
                  <span className="text-sm font-medium text-emerald-800 dark:text-emerald-200">自动补给已启用</span>
                  {creditData.auto_refill.needs_refill ? (
                    <Badge className="bg-orange-500 text-white text-xs px-2 py-0.5">
                      需要补给
                    </Badge>
                  ) : (
                    <Badge className="bg-green-500 text-white text-xs px-2 py-0.5">
                      无需补给
                    </Badge>
                  )}
                </div>
                
                <div className="space-y-1 text-xs text-emerald-700 dark:text-emerald-300">
                  <div className="flex justify-between">
                    <span>补给条件:</span>
                    <span>积分 ≤ {creditData.auto_refill.threshold}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>补给数量:</span>
                    <span>+{creditData.auto_refill.amount} 积分</span>
                  </div>
                  
                  {creditData.auto_refill.needs_refill && creditData.auto_refill.next_refill_time && (
                    <div className="flex items-center gap-1 mt-2 pt-2 border-t border-emerald-200 dark:border-emerald-700">
                      <Clock className="h-3 w-3" />
                      <span className="text-emerald-800 dark:text-emerald-200">
                        下次补给: {new Date(creditData.auto_refill.next_refill_time).toLocaleString('zh-CN', {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit'
                        })}
                      </span>
                    </div>
                  )}
                  
                  {creditData.auto_refill.last_refill_time && (
                    <div className="text-emerald-600 dark:text-emerald-400">
                      上次补给: {new Date(creditData.auto_refill.last_refill_time).toLocaleString('zh-CN', {
                        month: 'short',
                        day: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit'
                      })}
                    </div>
                  )}
                </div>
              </div>
            )}
            
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