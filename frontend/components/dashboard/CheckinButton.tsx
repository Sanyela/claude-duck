"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Gift, Calendar, Sparkles } from "lucide-react"
import { checkinAPI, CheckinStatusResponse } from "@/api/checkin"
import { toast } from "sonner"
import confetti from "canvas-confetti"

interface CheckinButtonProps {
  onCheckinSuccess?: () => void; // 签到成功后的回调函数
}

export function CheckinButton({ onCheckinSuccess }: CheckinButtonProps) {
  
  const [loading, setLoading] = useState(false)
  const [fetching, setFetching] = useState(true)
  const [checkinStatus, setCheckinStatus] = useState<CheckinStatusResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  // 加载签到状态
  const loadCheckinStatus = async () => {
    setFetching(true)
    setError(null)
    try {
      const status = await checkinAPI.getStatus()
      setCheckinStatus(status)
    } catch {
      console.error("获取签到状态失败:", error)
      setError(error.message || "获取签到状态失败")
    } finally {
      setFetching(false)
    }
  }

  useEffect(() => {
    loadCheckinStatus()
  }, [])

  // 执行签到
  const handleCheckin = async () => {
    if (!checkinStatus?.canCheckin) return

    setLoading(true)
    try {
      const result = await checkinAPI.checkin()
      
      if (result.success) {
        // 触发纸屑动效
        triggerConfetti()
        
        toast.success(`签到成功! 🎉`, {
          description: `恭喜获得 ${result.rewardPoints} 积分奖励`,
        })
        
        // 重新加载状态
        await loadCheckinStatus()
        
        // 通知父组件刷新数据
        if (onCheckinSuccess) {
          onCheckinSuccess()
        }
      } else {
        toast.error("签到失败", {
          description: result.message,
        })
      }
    } catch {
      console.error("签到失败:", error)
      toast.error("签到失败", {
        description: error.message || "网络错误，请稍后再试",
      })
    } finally {
      setLoading(false)
    }
  }

  // 触发纸屑动效
  const triggerConfetti = () => {
    // 创建多重纸屑动效
    const duration = 3000
    const animationEnd = Date.now() + duration
    const defaults = { startVelocity: 30, spread: 360, ticks: 60, zIndex: 0 }

    function randomInRange(min: number, max: number) {
      return Math.random() * (max - min) + min
    }

    const interval: NodeJS.Timeout = setInterval(function() {
      const timeLeft = animationEnd - Date.now()

      if (timeLeft <= 0) {
        return clearInterval(interval)
      }

      const particleCount = 50 * (timeLeft / duration)
      
      // 从左侧发射
      confetti({
        ...defaults,
        particleCount,
        origin: { x: randomInRange(0.1, 0.3), y: Math.random() - 0.2 }
      })
      
      // 从右侧发射
      confetti({
        ...defaults,
        particleCount,
        origin: { x: randomInRange(0.7, 0.9), y: Math.random() - 0.2 }
      })
    }, 250)
  }

  // 显示加载状态
  if (fetching) {
    return (
      <Card className="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 border-purple-200 dark:border-purple-700">
        <CardContent className="p-6">
          <div className="flex items-center justify-center space-x-2">
            <div className="w-4 h-4 border-2 border-purple-500 border-t-transparent rounded-full animate-spin" />
            <span className="text-purple-700 dark:text-purple-300">正在加载签到状态...</span>
          </div>
        </CardContent>
      </Card>
    )
  }

  // 显示错误状态
  if (error) {
    return (
      <Card className="bg-gradient-to-br from-red-50 to-orange-50 dark:from-red-900/20 dark:to-orange-900/20 border-red-200 dark:border-red-700">
        <CardContent className="p-6">
          <div className="text-center">
            <p className="text-red-700 dark:text-red-300 mb-2">签到功能加载失败</p>
            <p className="text-sm text-red-600 dark:text-red-400 mb-3">{error}</p>
            <Button 
              onClick={loadCheckinStatus}
              variant="outline"
              size="sm"
              className="border-red-300 text-red-700 hover:bg-red-50"
            >
              重试
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  // 如果没有有效的签到配置，不显示组件
  if (!checkinStatus || !checkinStatus.pointsRange.hasValid) {
    console.log("没有有效的签到配置:", checkinStatus)
    return null
  }

  const { canCheckin, todayChecked, lastCheckinDate, pointsRange } = checkinStatus

  // 获取积分范围描述
  const getPointsRangeText = () => {
    if (pointsRange.minPoints === pointsRange.maxPoints) {
      return `${pointsRange.minPoints} 积分`
    }
    return `${pointsRange.minPoints}-${pointsRange.maxPoints} 积分`
  }

  return (
    <Card className="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 border-purple-200 dark:border-purple-700">
      <CardHeader className="pb-3">
        <div className="flex items-center space-x-2">
          <div className="flex items-center justify-center w-8 h-8 bg-purple-100 dark:bg-purple-900 rounded-full">
            <Gift className="h-4 w-4 text-purple-600 dark:text-purple-400" />
          </div>
          <div>
            <CardTitle className="text-lg text-purple-900 dark:text-purple-100">每日签到</CardTitle>
            <CardDescription className="text-purple-600 dark:text-purple-300">
              每日签到获得积分奖励
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="space-y-4">
          {/* 积分奖励信息 */}
          <div className="flex items-center justify-between p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
            <div className="flex items-center space-x-2">
              <Sparkles className="h-4 w-4 text-yellow-500" />
              <span className="text-sm font-medium">奖励范围</span>
            </div>
            <Badge variant="secondary" className="bg-purple-100 text-purple-800 dark:bg-purple-800 dark:text-purple-200">
              {getPointsRangeText()}
            </Badge>
          </div>

          {/* 签到状态 */}
          <div className="flex items-center justify-between p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
            <div className="flex items-center space-x-2">
              <Calendar className="h-4 w-4 text-blue-500" />
              <span className="text-sm font-medium">签到状态</span>
            </div>
            <Badge variant={todayChecked ? "default" : "outline"}>
              {todayChecked ? "今日已签到" : "今日未签到"}
            </Badge>
          </div>

          {/* 最后签到日期 */}
          {lastCheckinDate && (
            <div className="text-xs text-gray-500 dark:text-gray-400 text-center">
              最后签到：{lastCheckinDate}
            </div>
          )}

          {/* 签到按钮 */}
          <Button
            onClick={handleCheckin}
            disabled={!canCheckin || loading}
            className={`w-full ${
              canCheckin && !loading
                ? "bg-gradient-to-r from-purple-500 to-pink-500 hover:from-purple-600 hover:to-pink-600 text-white"
                : "bg-gray-300 dark:bg-gray-600 text-gray-500 dark:text-gray-400"
            }`}
          >
            {loading ? (
              <div className="flex items-center space-x-2">
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                <span>签到中...</span>
              </div>
            ) : todayChecked ? (
              "今日已签到"
            ) : (
              <div className="flex items-center space-x-2">
                <Gift className="h-4 w-4" />
                <span>立即签到</span>
              </div>
            )}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
} 