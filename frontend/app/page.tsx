import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { BellRing, AlertTriangle, ArrowRight, CreditCard, DollarSign, BookOpen, Calendar, ChevronRight, Zap } from "lucide-react"
import Link from "next/link"
import { Greeting } from "@/components/dashboard/greeting"
import { Progress } from "@/components/ui/progress"
import { Badge } from "@/components/ui/badge"

export default function DashboardPage() {
  const currentUser = {
    name: "尊贵的用户",
    email: "user@example.com",
    avatarUrl: "/placeholder.svg?height=64&width=64",
  }
  const subscriptionStatus = "active" // 'active', 'expiring_soon', 'expired'
  
  // 积分数据
  const creditBalance = 1250
  const monthlyAllowance = 5000
  const creditUsagePercentage = Math.min((creditBalance / monthlyAllowance) * 100, 100)
  
  // 订阅数据
  const currentSubscription = {
    planName: "专业版 🌟",
    price: "¥99/月",
    status: "active",
    nextBillingDate: "2025-07-15",
    features: ["无限项目", "优先支持", "高级分析", "API访问"],
  }
  
  const announcements = [
    { id: 1, title: "系统维护通知", message: "我们的系统将于下周二凌晨2点至4点进行维护。", date: "2025-06-20" },
    { id: 2, title: "新功能上线！", message: "积分兑换商品功能已上线，快去看看吧！", date: "2025-06-18" },
  ]

  return (
    <DashboardLayout>
      <div className="space-y-6">
        <Greeting userName="Cloxl" />

        {subscriptionStatus === "expiring_soon" && (
          <Alert
            variant="destructive"
            className="shadow-lg bg-card text-card-foreground border-yellow-400 dark:border-yellow-600"
          >
            <AlertTriangle className="h-5 w-5 text-yellow-500 dark:text-yellow-400" />
            <AlertTitle className="text-yellow-700 dark:text-yellow-300">订阅即将到期</AlertTitle>
            <AlertDescription className="text-yellow-700 dark:text-yellow-400">
              您的订阅将在 <strong>3天后</strong> 到期。请及时{" "}
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
        {subscriptionStatus === "expired" && (
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
              <div className="flex items-end justify-between">
                <div>
                  <p className="text-3xl font-bold">{creditBalance.toLocaleString()}</p>
                  <p className="text-xs text-muted-foreground">总额度: {monthlyAllowance.toLocaleString()} 积分</p>
                </div>
                <Link 
                  href="/credits"
                  className="text-sm text-sky-500 hover:text-sky-600 dark:text-sky-400 dark:hover:text-sky-300 flex items-center"
                >
                  详情 <ChevronRight className="h-4 w-4 ml-1" />
                </Link>
              </div>
              <Progress value={creditUsagePercentage} className="mt-3 h-2 [&>*]:bg-sky-500" />
              <p className="text-xs text-muted-foreground mt-1 text-right">
                使用率: {creditUsagePercentage.toFixed(1)}%
              </p>
            </CardContent>
            <CardFooter className="pt-0">
              <Button 
                size="sm" 
                variant="outline" 
                className="w-full text-green-600 border-green-200 hover:bg-green-50 hover:text-green-700 dark:text-green-400 dark:border-green-800/30 dark:hover:bg-green-900/30"
                asChild
              >
                <Link href="/credits">
                  <Zap className="mr-2 h-4 w-4" /> 充值积分
                </Link>
              </Button>
            </CardFooter>
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
              <div className="flex items-center justify-between mb-2">
                <div>
                  <p className="text-xl font-bold text-sky-500 dark:text-sky-400">{currentSubscription.planName}</p>
                  <p className="text-sm font-medium">{currentSubscription.price}</p>
                </div>
                <Badge 
                  className="bg-green-500 text-white dark:bg-green-600 dark:text-green-50"
                >
                  有效
                </Badge>
              </div>
              <p className="text-xs flex items-center text-muted-foreground">
                <Calendar className="h-3 w-3 mr-1 inline" /> 下次账单日期: {currentSubscription.nextBillingDate}
              </p>
            </CardContent>
            <CardFooter className="pt-0">
              <Button 
                size="sm" 
                variant="outline" 
                className="w-full text-sky-600 border-sky-200 hover:bg-sky-50 hover:text-sky-700 dark:text-sky-400 dark:border-sky-800/30 dark:hover:bg-sky-900/30"
                asChild
              >
                <Link href="/subscription">
                  <CreditCard className="mr-2 h-4 w-4" /> 管理订阅
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
            {announcements.map((ann) => (
              <Alert key={ann.id} className="bg-card text-card-foreground border-border">
                {" "}
                {/* Nested Alert, ensure it uses card's bg or a subtle variant */}
                <BellRing className="h-5 w-5 text-sky-500 dark:text-sky-400" />
                <AlertTitle>
                  {ann.title} <span className="text-xs text-muted-foreground ml-2">{ann.date}</span>
                </AlertTitle>
                <AlertDescription>{ann.message}</AlertDescription>
              </Alert>
            ))}
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
                <p className="text-sm text-muted-foreground mb-3">查看您的积分余额、充值、并跟踪使用情况。</p>
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
