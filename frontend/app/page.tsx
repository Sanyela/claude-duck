"use client";

import { useState, useEffect } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { BellRing, AlertTriangle, ArrowRight, CreditCard, DollarSign, BookOpen, Calendar, ChevronRight, Loader2 } from "lucide-react"
import Link from "next/link"
import { Greeting } from "@/components/dashboard/greeting"
import { Progress } from "@/components/ui/progress"
import { Badge } from "@/components/ui/badge"
import { useAuth } from "@/contexts/AuthContext"
import { dashboardAPI, type DashboardData } from "@/api/dashboard"
import { announcementsAPI, type PublicAnnouncement } from "@/api/announcements"

export default function DashboardPage() {
  const { user } = useAuth();
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null);
  const [announcements, setAnnouncements] = useState<PublicAnnouncement[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // 加载仪表盘数据
  useEffect(() => {
    const loadDashboardData = async () => {
      setLoading(true);
      setError(null);
      
      const [dashboardResult, announcementsResult] = await Promise.all([
        dashboardAPI.getDashboardData(),
        announcementsAPI.getActiveAnnouncements("zh")
      ]);

      if (dashboardResult.success && dashboardResult.data) {
        setDashboardData(dashboardResult.data);
      } else {
        setError(dashboardResult.message || "加载数据失败");
      }

      if (announcementsResult.success && announcementsResult.data) {
        setAnnouncements(announcementsResult.data);
      }
      
      setLoading(false);
    };

    if (user) {
      loadDashboardData();
    }
  }, [user]);
  
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
    if (!dashboardData?.subscription) return { status: "no_subscription", timeRemaining: "" };
    
    const currentPeriodEnd = new Date(dashboardData.subscription.currentPeriodEnd);
    const now = new Date();
    const daysUntilExpiry = Math.ceil((currentPeriodEnd.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
    const timeRemaining = formatTimeRemaining(currentPeriodEnd);
    
    if (dashboardData.subscription.status === "active") {
      if (daysUntilExpiry <= 3) return { status: "expiring_soon", timeRemaining };
      return { status: "active", timeRemaining };
    }
    
    return { status: "expired", timeRemaining };
  };

  const subscriptionInfo = getSubscriptionInfo();

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
        {subscriptionInfo.status === "expired" && (
          <Alert
            variant="destructive"
            className="shadow-lg bg-card text-card-foreground border-red-400 dark:border-red-600"
          >
            <AlertTriangle className="h-5 w-5 text-red-500 dark:text-red-400" />
            <AlertTitle className="text-red-700 dark:text-red-300">订阅已过期</AlertTitle>
            <AlertDescription className="text-red-700 dark:text-red-400">
              您的订阅已过期。部分功能可能受限。请{" "}
              <Link href="/subscription" className="font-semibold underline hover:text-red-600 dark:hover:text-red-200">
                重新激活
              </Link>{" "}
              您的订阅。
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
        <div className="grid gap-6 md:grid-cols-2">
          {/* 积分卡片 */}
          <Card className="shadow-lg bg-card text-card-foreground border-border">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle>当前积分</CardTitle>
                <DollarSign className="h-5 w-5 text-green-500 dark:text-green-400" />
              </div>
            </CardHeader>
            <CardContent className="pb-2">
              {loading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin" />
                  <span className="ml-2">加载中...</span>
                </div>
              ) : dashboardData?.pointBalance ? (
                <>
                  <div className="flex items-end justify-between">
                <div>
                      <p className="text-3xl font-bold">
                        {(dashboardData.pointBalance.available_points || 0).toLocaleString()}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        总积分: {(dashboardData.pointBalance.total_points || 0).toLocaleString()} | 
                        已使用: {(dashboardData.pointBalance.used_points || 0).toLocaleString()}
                      </p>
                </div>
                <Link 
                  href="/credits"
                  className="text-sm text-sky-500 hover:text-sky-600 dark:text-sky-400 dark:hover:text-sky-300 flex items-center"
                >
                  详情 <ChevronRight className="h-4 w-4 ml-1" />
                </Link>
              </div>
                  {(dashboardData.pointBalance.total_points || 0) > 0 && (
                    <>
              <Progress value={creditUsagePercentage} className="mt-3 h-2 [&>*]:bg-sky-500" />
              <p className="text-xs text-muted-foreground mt-1 text-right">
                使用率: {creditUsagePercentage.toFixed(1)}%
              </p>
                    </>
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
              ) : dashboardData?.subscription ? (
                <>
                  <div className="flex items-center justify-between mb-2">
                <div>
                      <p className="text-xl font-bold text-sky-500 dark:text-sky-400">{dashboardData.subscription.plan.name}</p>
                </div>
                <Badge 
                      className={
                        subscriptionInfo.status === "active" ? "bg-green-500 text-white dark:bg-green-600" :
                        subscriptionInfo.status === "expiring_soon" ? "bg-yellow-500 text-white dark:bg-yellow-600" :
                        "bg-red-500 text-white dark:bg-red-600"
                      }
                    >
                      {subscriptionInfo.status === "active" ? "有效" : 
                       subscriptionInfo.status === "expiring_soon" ? "即将到期" : "已过期"}
                </Badge>
              </div>
              <p className="text-xs flex items-center text-muted-foreground">
                    <Calendar className="h-3 w-3 mr-1 inline" /> 
                    下次账单日期: {new Date(dashboardData.subscription.currentPeriodEnd).toLocaleDateString()}
              </p>
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
                  {dashboardData?.subscription ? "管理订阅" : "激活套餐"}
                </Link>
              </Button>
            </CardFooter>
          </Card>
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
