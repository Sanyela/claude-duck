"use client"

import { useState, useEffect } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"
import { 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  LineChart,
  Line,
  PieChart,
  Pie,
  Cell
} from "recharts"
import { 
  Users, 
  CreditCard, 
  TrendingUp, 
  TrendingDown, 
  Activity,
  UserCheck,
  DollarSign,
  Calendar,
  Loader2,
  AlertCircle
} from "lucide-react"
import { adminAPI, type DashboardData } from "@/api/admin"

// 颜色配置
const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D']

export default function AdminDashboardPage() {
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // 加载数据看板数据
  useEffect(() => {
    const loadDashboardData = async () => {
      setLoading(true)
      setError(null)
      
      try {
        const result = await adminAPI.getDashboard()
        if (result.success && result.data) {
          setDashboardData(result.data)
        } else {
          setError(result.message || "获取数据看板失败")
        }
      } catch (err: any) {
        setError("获取数据看板失败")
      } finally {
        setLoading(false)
      }
    }

    loadDashboardData()
  }, [])

  // 格式化增长率显示
  const formatGrowthRate = (rate: number) => {
    const isPositive = rate >= 0
    const icon = isPositive ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />
    const color = isPositive ? "text-green-600" : "text-red-600"
    
    return (
      <div className={`flex items-center gap-1 text-xs ${color}`}>
        {icon}
        <span>{Math.abs(rate).toFixed(1)}%</span>
      </div>
    )
  }

  // 格式化积分来源类型
  const formatSourceType = (sourceType: string) => {
    const sourceMap: Record<string, string> = {
      'activation_code': '激活码',
      'payment': '付费购买',
      'daily_checkin': '每日签到',
      'admin_gift': '管理员赠送'
    }
    return sourceMap[sourceType] || sourceType
  }

  if (loading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-[60vh]">
          <div className="flex items-center gap-2">
            <Loader2 className="h-6 w-6 animate-spin" />
            <span>加载数据看板中...</span>
          </div>
        </div>
      </DashboardLayout>
    )
  }

  if (error) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center min-h-[60vh]">
          <div className="flex items-center gap-2 text-red-600">
            <AlertCircle className="h-6 w-6" />
            <span>{error}</span>
          </div>
        </div>
      </DashboardLayout>
    )
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold">管理员数据看板</h1>
          <Badge variant="outline">
            更新时间: {dashboardData ? new Date(dashboardData.generated_at).toLocaleString() : '--'}
          </Badge>
        </div>

        {/* 概览统计卡片 */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">总用户数</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {dashboardData?.overview.total_users.toLocaleString() || 0}
              </div>
              <div className="flex items-center justify-between mt-2">
                <p className="text-xs text-muted-foreground">
                  今日新增: {dashboardData?.overview.today_new_users || 0}
                </p>
                {dashboardData && formatGrowthRate(dashboardData.overview.user_growth_rate)}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">有效订阅用户</CardTitle>
              <UserCheck className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {dashboardData?.overview.active_subscription_users.toLocaleString() || 0}
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                订阅转化率: {dashboardData && dashboardData.overview.total_users > 0 
                  ? ((dashboardData.overview.active_subscription_users / dashboardData.overview.total_users) * 100).toFixed(1)
                  : 0}%
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">活跃订阅数</CardTitle>
              <CreditCard className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {dashboardData?.overview.total_subscriptions.toLocaleString() || 0}
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                人均订阅: {dashboardData && dashboardData.overview.active_subscription_users > 0
                  ? (dashboardData.overview.total_subscriptions / dashboardData.overview.active_subscription_users).toFixed(1)
                  : 0}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">今日积分消耗</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {dashboardData?.overview.today_points_used.toLocaleString() || 0}
              </div>
              <div className="flex items-center justify-between mt-2">
                <p className="text-xs text-muted-foreground">环比昨日</p>
                {dashboardData && formatGrowthRate(dashboardData.overview.points_growth_rate)}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* 积分统计 */}
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <Card className="lg:col-span-1">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <DollarSign className="h-5 w-5" />
                积分总览
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-muted-foreground">总积分</span>
                  <span className="font-semibold">
                    {dashboardData?.points_stats.total_points.toLocaleString() || 0}
                  </span>
                </div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-muted-foreground">已使用</span>
                  <span className="font-semibold text-red-600">
                    {dashboardData?.points_stats.used_points.toLocaleString() || 0}
                  </span>
                </div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-muted-foreground">可用积分</span>
                  <span className="font-semibold text-green-600">
                    {dashboardData?.points_stats.available_points.toLocaleString() || 0}
                  </span>
                </div>
                {dashboardData && dashboardData.points_stats.total_points > 0 && (
                  <div className="mt-4">
                    <div className="flex items-center justify-between text-sm mb-1">
                      <span>使用率</span>
                      <span>{((dashboardData.points_stats.used_points / dashboardData.points_stats.total_points) * 100).toFixed(1)}%</span>
                    </div>
                    <Progress 
                      value={(dashboardData.points_stats.used_points / dashboardData.points_stats.total_points) * 100}
                      className="h-2"
                    />
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* 订阅计划分布 */}
          <Card className="lg:col-span-2">
            <CardHeader>
              <CardTitle>订阅计划分布</CardTitle>
              <CardDescription>各订阅计划的用户数量分布</CardDescription>
            </CardHeader>
            <CardContent>
              {dashboardData?.distributions.subscription_plans.length ? (
                <ResponsiveContainer width="100%" height={200}>
                  <BarChart data={dashboardData.distributions.subscription_plans}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="plan_name" />
                    <YAxis />
                    <Tooltip />
                    <Bar dataKey="count" fill="#8884d8" />
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <div className="flex items-center justify-center h-48 text-muted-foreground">
                  暂无数据
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* 趋势图表 */}
        <div className="grid gap-6 lg:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>用户注册趋势</CardTitle>
              <CardDescription>最近7天的用户注册情况</CardDescription>
            </CardHeader>
            <CardContent>
              {dashboardData?.trends.user_registration.length ? (
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={dashboardData.trends.user_registration}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis 
                      dataKey="date" 
                      tickFormatter={(value) => new Date(value).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })}
                    />
                    <YAxis />
                    <Tooltip 
                      labelFormatter={(value) => new Date(value).toLocaleDateString()}
                      formatter={(value: any) => [value, '新增用户']}
                    />
                    <Line 
                      type="monotone" 
                      dataKey="count" 
                      stroke="#8884d8" 
                      strokeWidth={2}
                      dot={{ r: 4 }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div className="flex items-center justify-center h-72 text-muted-foreground">
                  暂无数据
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>积分使用趋势</CardTitle>
              <CardDescription>最近7天的积分消耗情况</CardDescription>
            </CardHeader>
            <CardContent>
              {dashboardData?.trends.points_usage.length ? (
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={dashboardData.trends.points_usage}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis 
                      dataKey="date" 
                      tickFormatter={(value) => new Date(value).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })}
                    />
                    <YAxis />
                    <Tooltip 
                      labelFormatter={(value) => new Date(value).toLocaleDateString()}
                      formatter={(value: any) => [value, '积分消耗']}
                    />
                    <Line 
                      type="monotone" 
                      dataKey="count" 
                      stroke="#82ca9d" 
                      strokeWidth={2}
                      dot={{ r: 4 }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <div className="flex items-center justify-center h-72 text-muted-foreground">
                  暂无数据
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* 积分来源分布 */}
        <Card>
          <CardHeader>
            <CardTitle>积分来源分布</CardTitle>
            <CardDescription>不同来源的积分数量和订阅数分布</CardDescription>
          </CardHeader>
          <CardContent>
            {dashboardData?.distributions.points_sources.length ? (
              <div className="grid gap-6 lg:grid-cols-2">
                <div>
                  <h4 className="text-sm font-medium mb-4">按订阅数量</h4>
                  <ResponsiveContainer width="100%" height={250}>
                    <PieChart>
                      <Pie
                        data={dashboardData.distributions.points_sources}
                        cx="50%"
                        cy="50%"
                        labelLine={false}
                        label={({ source_type, count, percent }) => 
                          `${formatSourceType(source_type)}: ${count} (${((percent || 0) * 100).toFixed(0)}%)`
                        }
                        outerRadius={80}
                        fill="#8884d8"
                        dataKey="count"
                      >
                        {dashboardData.distributions.points_sources.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip formatter={(value: any, name, props) => [value, '订阅数']} />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <div>
                  <h4 className="text-sm font-medium mb-4">按积分数量</h4>
                  <ResponsiveContainer width="100%" height={250}>
                    <PieChart>
                      <Pie
                        data={dashboardData.distributions.points_sources}
                        cx="50%"
                        cy="50%"
                        labelLine={false}
                        label={({ source_type, points, percent }) => 
                          `${formatSourceType(source_type)}: ${points.toLocaleString()} (${((percent || 0) * 100).toFixed(0)}%)`
                        }
                        outerRadius={80}
                        fill="#8884d8"
                        dataKey="points"
                      >
                        {dashboardData.distributions.points_sources.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip formatter={(value: any, name, props) => [value.toLocaleString(), '积分']} />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </div>
            ) : (
              <div className="flex items-center justify-center h-64 text-muted-foreground">
                暂无数据
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  )
} 