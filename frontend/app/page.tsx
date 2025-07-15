"use client";

import { useState, useEffect } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { BellRing, AlertTriangle, ArrowRight, CreditCard, DollarSign, BookOpen, Calendar, ChevronRight, Loader2, Zap, RefreshCw } from "lucide-react"
import Link from "next/link"
import { Greeting } from "@/components/dashboard/greeting"
import { CheckinButton } from "@/components/dashboard/CheckinButton"
import { AnimatedNumber } from "@/components/ui/animated-number"
import { Progress } from "@/components/ui/progress"
import { Badge } from "@/components/ui/badge"
import { useAuth } from "@/contexts/AuthContext"
import { dashboardAPI, type DashboardData } from "@/api/dashboard"
import { creditsAPI, type CreditBalance } from "@/api/credits"
import { announcementsAPI, type PublicAnnouncement } from "@/api/announcements"

export default function DashboardPage() {
  const { user } = useAuth();
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null);
  const [creditData, setCreditData] = useState<CreditBalance | null>(null);
  const [announcements, setAnnouncements] = useState<PublicAnnouncement[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshTrigger, setRefreshTrigger] = useState(0); // 刷新触发器
  
  // 加载仪表盘数据
  const loadDashboardData = async () => {
    setLoading(true);
    setError(null);
    
    const [dashboardResult, creditResult, announcementsResult] = await Promise.all([
      dashboardAPI.getDashboardData(),
      creditsAPI.getBalance(),
      announcementsAPI.getActiveAnnouncements("zh")
    ]);

    if (dashboardResult.success && dashboardResult.data) {
      setDashboardData(dashboardResult.data);
    } else {
      setError(dashboardResult.message || "加载数据失败");
    }

    if (creditResult.success && creditResult.data) {
      setCreditData(creditResult.data);
    }

    if (announcementsResult.success && announcementsResult.data) {
      setAnnouncements(announcementsResult.data);
    }
    
    setLoading(false);
  };

  useEffect(() => {
    if (user) {
      loadDashboardData();
    }
  }, [user]);

  // 刷新积分数据
  const refreshCreditData = async () => {
    try {
      const [dashboardResult, creditResult] = await Promise.all([
        dashboardAPI.getDashboardData(),
        creditsAPI.getBalance()
      ]);

      if (dashboardResult.success && dashboardResult.data) {
        setDashboardData(dashboardResult.data);
      }

      if (creditResult.success && creditResult.data) {
        setCreditData(creditResult.data);
      }

      // 触发动画
      setRefreshTrigger(prev => prev + 1);
    } catch (error) {
      console.error("刷新积分数据失败:", error);
    }
  };
  
  // 计算积分使用率
  const creditUsagePercentage = dashboardData?.pointBalance 
    ? Math.min(((dashboardData.pointBalance.used_points || 0) / Math.max(dashboardData.pointBalance.total_points || 1, 1)) * 100, 100)
    : 0;

  // 格式化剩余时间
  const formatTimeRemaining = (endDate: Date): string => {
    const now = new Date();
    const diff = endDate.getTime() - now.getTime();
    
    if (diff <= 0) return "已过期";
    
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    
    const parts = [];
    if (days > 0) parts.push(`${days}天`);
    if (hours > 0) parts.push(`${hours}小时`);
    if (minutes > 0) parts.push(`${minutes}分钟`);
    
    return parts.length > 0 ? parts.join('') : "不足1分钟";
  };

  // 判断订阅状态和剩余时间
  const getSubscriptionInfo = () => {
    if (!dashboardData?.subscriptions || dashboardData.subscriptions.length === 0) {
      return { status: "no_subscription", timeRemaining: "", currentSubscription: null, otherSubscriptions: [] };
    }
    
    // 只考虑有效的订阅
    const validSubscriptions = dashboardData.subscriptions.filter(sub => sub.detailedStatus === "有效");
    
    if (validSubscriptions.length === 0) {
      return { status: "no_active", timeRemaining: "", currentSubscription: null, otherSubscriptions: [] };
    }
    
    // 找到当前正在使用的有效订阅
    const currentSubscription = validSubscriptions.find(sub => sub.isCurrentUsing) || validSubscriptions[0];
    const otherSubscriptions = validSubscriptions.filter(sub => sub.id !== currentSubscription.id);
    
    const currentPeriodEnd = new Date(currentSubscription.currentPeriodEnd);
    const now = new Date();
    const daysUntilExpiry = Math.ceil((currentPeriodEnd.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
    const timeRemaining = formatTimeRemaining(currentPeriodEnd);
    
    // 判断状态
    if (daysUntilExpiry <= 3) {
      return { status: "expiring_soon", timeRemaining, currentSubscription, otherSubscriptions }
    }
    return { status: "active", timeRemaining, currentSubscription, otherSubscriptions }
  }

  const subscriptionInfo = getSubscriptionInfo()

  return (
    <DashboardLayout>
      <div className="space-y-6">
        <Greeting userName={user?.username || "尊贵的用户"} />

        {/* 订阅状态警告 */}
        {subscriptionInfo.status === "expiring_soon" && (
          <Alert
            variant="destructive"
            className="shadow-lg bg-card text-card-foreground border-yellow-400 dark:border-yellow-600"
          >
            <AlertTriangle className="h-5 w-5 text-yellow-500 dark:text-yellow-400" />
            <AlertTitle className="text-yellow-700 dark:text-yellow-300">订阅即将到期</AlertTitle>
            <AlertDescription className="text-yellow-700 dark:text-yellow-400">
              您的订阅将在 <strong>{subscriptionInfo.timeRemaining}</strong> 后到期。请及时{" "}
              <Link
                href="/subscription"
                className="font-semibold underline hover:text-yellow-600 dark:hover:text-yellow-200"
              >
                续订
              </Link>{" "}
              以免服务中断。
            </AlertDescription>
          </Alert>
        )}
        {subscriptionInfo.status === "no_active" && (
          <Alert
            variant="destructive"
            className="shadow-lg bg-card text-card-foreground border-red-400 dark:border-red-600"
          >
            <AlertTriangle className="h-5 w-5 text-red-500 dark:text-red-400" />
            <AlertTitle className="text-red-700 dark:text-red-300">暂无有效订阅</AlertTitle>
            <AlertDescription className="text-red-700 dark:text-red-400">
              您当前没有有效的订阅。请{" "}
              <Link href="/subscription" className="font-semibold underline hover:text-red-600 dark:hover:text-red-200">
                激活订阅
              </Link>{" "}
              以享受完整服务。
            </AlertDescription>
          </Alert>
        )}

        {/* 错误提示 */}
        {error && (
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>加载失败</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* 用户积分和订阅状态卡片 */}
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {/* 积分卡片 */}
          <Card className="shadow-lg bg-card text-card-foreground border-border">
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg">积分状态</CardTitle>
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={refreshCreditData}
                    className="h-8 w-8 p-0 hover:bg-gray-100 dark:hover:bg-gray-800"
                  >
                    <RefreshCw className="h-4 w-4 text-gray-500 dark:text-gray-400" />
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent className="pb-2">
              {loading || !dashboardData?.pointBalance ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin" />
                  <span className="ml-2">加载中...</span>
                </div>
              ) : (dashboardData.pointBalance.total_points || 0) > 0 ? (
                <>
                  <div className="flex items-end justify-between mb-2">
                    <div>
                      <div className="text-2xl font-bold">
                        <AnimatedNumber 
                          value={dashboardData.pointBalance.available_points || 0}
                          duration={800}
                          className="text-2xl font-bold"
                          refreshTrigger={refreshTrigger}
                        />
                      </div>
                      <div className="text-xs text-muted-foreground">
                        当前积分
                      </div>
                    </div>
                    <Link 
                      href="/credits"
                      className="text-sm text-sky-500 hover:text-sky-600 dark:text-sky-400 dark:hover:text-sky-300 flex items-center"
                    >
                      详情 <ChevronRight className="h-4 w-4 ml-1" />
                    </Link>
                  </div>
                  
                  <div className="text-xs text-slate-600 dark:text-slate-400 mb-2">
                    <span>积分总数: <span className="font-semibold text-slate-700 dark:text-slate-300">{dashboardData.pointBalance.total_points?.toLocaleString() || '0'}</span></span>
                    <span className="mx-2">|</span>
                    <span>已使用: <span className="font-semibold text-slate-700 dark:text-slate-300">{dashboardData.pointBalance.used_points?.toLocaleString() || '0'}</span></span>
                    {(dashboardData.pointBalance.checkin_points || 0) > 0 && (
                      <>
                        <span className="mx-2">|</span>
                        <span>签到积分: <span className="font-semibold text-purple-600 dark:text-purple-400">{dashboardData.pointBalance.checkin_points?.toLocaleString()}</span></span>
                      </>
                    )}
                    {(dashboardData.pointBalance.admin_gift_points || 0) > 0 && (
                      <>
                        <span className="mx-2">|</span>
                        <span>管理员赠送: <span className="font-semibold text-amber-600 dark:text-amber-400">{dashboardData.pointBalance.admin_gift_points?.toLocaleString()}</span></span>
                      </>
                    )}
                  </div>
                  
                  {(dashboardData.pointBalance.total_points || 0) > 0 && (
                    <div className="relative mb-2">
                      <Progress value={creditUsagePercentage} className="h-6 [&>*]:bg-sky-500" />
                      <div className="absolute inset-0 flex items-center justify-center text-xs font-medium text-white mix-blend-difference">
                        {dashboardData.pointBalance.available_points?.toLocaleString() || '0'} / {dashboardData.pointBalance.total_points?.toLocaleString() || '0'}
                      </div>
                    </div>
                  )}
                  
                  <p className="text-xs text-muted-foreground opacity-75">
                    * 仅显示当前有效订阅数据，使用率: {creditUsagePercentage.toFixed(1)}%
                  </p>
                  
                  {/* 自动补给信息 */}
                  {creditData?.auto_refill?.enabled && (
                    <div className="bg-gradient-to-r from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 rounded-lg p-2.5 border border-emerald-200 dark:border-emerald-800 mt-2">
                      {/* 第一行：标题和状态 */}
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <Zap className="h-3.5 w-3.5 text-emerald-600" />
                          <span className="text-sm font-medium text-emerald-800 dark:text-emerald-200">自动补给</span>
                          <Badge className={creditData.auto_refill.needs_refill ? "bg-orange-500 text-white" : "bg-green-500 text-white"}>
                            {creditData.auto_refill.needs_refill ? "待补给" : "正常"}
                          </Badge>
                        </div>
                        <span className="text-xs text-emerald-700 dark:text-emerald-300 font-mono bg-emerald-100 dark:bg-emerald-900/50 px-2 py-0.5 rounded">
                          当前积分 ≤ {creditData.auto_refill.threshold} → 自动补给 +{creditData.auto_refill.amount}
                        </span>
                      </div>
                      
                      {/* 第二行：检查时间和触发规则 */}
                      <div className="flex items-center justify-between mt-1.5 text-xs">
                        <span className="text-emerald-600 dark:text-emerald-400 bg-emerald-100 dark:bg-emerald-900/30 px-2 py-0.5 rounded">
                          🕐 检查时间: 0、4、8、12、16、20点
                        </span>
                        <span className="text-emerald-600 dark:text-emerald-400 bg-emerald-100 dark:bg-emerald-900/30 px-2 py-0.5 rounded">
                          积分不足时自动触发
                        </span>
                      </div>
                      
                      {/* 第三行：时间信息 */}
                      <div className="flex items-center justify-between mt-1.5 text-xs">
                        <div className="flex items-center gap-3">
                          {creditData.auto_refill.last_refill_time && (
                            <span className="text-emerald-600 dark:text-emerald-400 bg-green-100 dark:bg-green-900/30 px-2 py-0.5 rounded">
                              ✅ 上次补给: {new Date(creditData.auto_refill.last_refill_time).toLocaleString('zh-CN', {
                                month: 'numeric',
                                day: 'numeric',
                                hour: '2-digit',
                                minute: '2-digit'
                              })}
                            </span>
                          )}
                        </div>
                        {creditData.auto_refill.needs_refill && (
                          <span className="text-orange-700 dark:text-orange-300 font-medium bg-orange-100 dark:bg-orange-900/30 px-2 py-0.5 rounded">
                            ⚠️ 积分不足，等待自动补给
                          </span>
                        )}
                      </div>
                    </div>
                  )}
                </>
              ) : (
                <div className="text-center py-4">
                  <p className="text-lg font-medium text-muted-foreground">暂无积分数据</p>
                  <p className="text-sm text-muted-foreground">请联系管理员或使用激活码充值</p>
                </div>
              )}
            </CardContent>
          </Card>
          
          {/* 订阅卡片 */}
          <Card className="shadow-lg bg-card text-card-foreground border-border">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle>订阅状态</CardTitle>
                <CreditCard className="h-5 w-5 text-sky-500 dark:text-sky-400" />
              </div>
            </CardHeader>
            <CardContent className="pb-2">
              {loading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin" />
                  <span className="ml-2">加载中...</span>
                </div>
              ) : subscriptionInfo.currentSubscription ? (
                <>
                  <div className="flex items-center justify-between mb-2">
                    <div>
                      <p className="text-xl font-bold text-sky-500 dark:text-sky-400">
                        {subscriptionInfo.currentSubscription.plan.name}
                      </p>
                      {subscriptionInfo.currentSubscription.isCurrentUsing && (
                        <Badge className="bg-blue-500 text-white text-xs mt-1">
                          当前消耗
                        </Badge>
                      )}
                    </div>
                    <Badge 
                      className={
                        subscriptionInfo.status === "active" ? "bg-green-500 text-white dark:bg-green-600" :
                        "bg-yellow-500 text-white dark:bg-yellow-600"
                      }
                    >
                      {subscriptionInfo.status === "expiring_soon" ? "即将到期" : "有效"}
                    </Badge>
                  </div>
                  <div className="text-xs text-muted-foreground space-y-1">
                    <p className="flex items-center">
                      <Calendar className="h-3 w-3 mr-1 inline" /> 
                      下次账单日期: {new Date(subscriptionInfo.currentSubscription.currentPeriodEnd).toLocaleDateString()}
                    </p>
                    <p>
                      积分: {subscriptionInfo.currentSubscription.availablePoints.toLocaleString()} / {subscriptionInfo.currentSubscription.totalPoints.toLocaleString()}
                    </p>
                  </div>
                  
                  {/* 显示其他订阅 */}
                  {subscriptionInfo.otherSubscriptions.length > 0 && (
                    <div className="mt-3 pt-3 border-t border-border">
                      <p className="text-xs text-muted-foreground mb-2">其他订阅:</p>
                      <div className="space-y-1">
                        {subscriptionInfo.otherSubscriptions.slice(0, 2).map((sub) => (
                          <div key={sub.id} className="flex justify-between items-center text-xs">
                            <span>{sub.plan.name}</span>
                            <div className="flex items-center gap-2">
                              <Badge variant="outline" className="text-xs">
                                {sub.detailedStatus}
                              </Badge>
                              <span className="text-muted-foreground">
                                {sub.availablePoints.toLocaleString()}积分
                              </span>
                            </div>
                          </div>
                        ))}
                        {subscriptionInfo.otherSubscriptions.length > 2 && (
                          <p className="text-xs text-muted-foreground">
                            还有 {subscriptionInfo.otherSubscriptions.length - 2} 个订阅...
                          </p>
                        )}
                      </div>
                    </div>
                  )}
                </>
              ) : (
                <div className="text-center py-4">
                  <p className="text-lg font-medium text-muted-foreground">暂无套餐</p>
                  <p className="text-sm text-muted-foreground">您可以使用激活码激活套餐</p>
                </div>
              )}
            </CardContent>
            <CardFooter className="pt-0">
              <Button 
                size="sm" 
                variant="outline" 
                className="w-full text-sky-600 border-sky-200 hover:bg-sky-50 hover:text-sky-700 dark:text-sky-400 dark:border-sky-800/30 dark:hover:bg-sky-900/30"
                asChild
              >
                <Link href="/subscription">
                  <CreditCard className="mr-2 h-4 w-4" /> 
                  {subscriptionInfo.currentSubscription ? "管理订阅" : "激活套餐"}
                </Link>
              </Button>
            </CardFooter>
          </Card>

          {/* 签到卡片 */}
          <CheckinButton onCheckinSuccess={loadDashboardData} />
        </div>
        
        <Card className="shadow-lg bg-card text-card-foreground border-border">
          <CardHeader>
            <CardTitle className="text-2xl">系统公告 📢</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {announcements.length > 0 ? (
              announcements.map((announcement) => (
                <Alert key={announcement.id} className="bg-card text-card-foreground border-border">
                  <BellRing className="h-5 w-5 text-sky-500 dark:text-sky-400" />
                  <AlertTitle>
                    {announcement.title} 
                    <span className="text-xs text-muted-foreground ml-2">
                      {new Date(announcement.created_at).toLocaleDateString()}
                    </span>
                  </AlertTitle>
                  <AlertDescription>
                    <div dangerouslySetInnerHTML={{ __html: announcement.description }} />
                  </AlertDescription>
                </Alert>
              ))
            ) : (
              <Alert className="bg-card text-card-foreground border-border">
                <BellRing className="h-5 w-5 text-sky-500 dark:text-sky-400" />
                <AlertTitle>暂无公告</AlertTitle>
                <AlertDescription>当前没有系统公告。</AlertDescription>
              </Alert>
            )}
          </CardContent>
        </Card>

        <Card className="shadow-lg bg-card text-card-foreground border-border">
          <CardHeader>
            <CardTitle className="text-2xl">快速操作 ⚡️</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Card className="shadow-md bg-card text-card-foreground border-border hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-lg font-medium">订阅管理</CardTitle>
                <CreditCard className="h-6 w-6 text-sky-500 dark:text-sky-400" />
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground mb-3">查看和管理您的订阅计划、历史记录和优惠券。</p>
                <Button
                  variant="outline"
                  size="sm"
                  asChild
                  className="border-border text-foreground hover:bg-accent hover:text-accent-foreground"
                >
                  <Link href="/subscription">
                    前往订阅 <ArrowRight className="ml-1 h-4 w-4" />
                  </Link>
                </Button>
              </CardContent>
            </Card>
            <Card className="shadow-md bg-card text-card-foreground border-border hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-lg font-medium">积分中心</CardTitle>
                <DollarSign className="h-6 w-6 text-green-500 dark:text-green-400" />
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground mb-3">查看您的积分余额、使用记录和积分历史。</p>
                <Button
                  variant="outline"
                  size="sm"
                  asChild
                  className="border-border text-foreground hover:bg-accent hover:text-accent-foreground"
                >
                  <Link href="/credits">
                    查看积分 <ArrowRight className="ml-1 h-4 w-4" />
                  </Link>
                </Button>
              </CardContent>
            </Card>
            <Card className="shadow-md bg-card text-card-foreground border-border hover:shadow-lg transition-shadow">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-lg font-medium">资源文档</CardTitle>
                <BookOpen className="h-6 w-6 text-purple-500 dark:text-purple-400" />
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground mb-3">访问我们的帮助文档、API参考和教程。</p>
                <Button
                  variant="outline"
                  size="sm"
                  asChild
                  className="border-border text-foreground hover:bg-accent hover:text-accent-foreground"
                >
                  <Link href="/resources">
                    浏览资源 <ArrowRight className="ml-1 h-4 w-4" />
                  </Link>
                </Button>
              </CardContent>
            </Card>
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  )
}
