"use client"

import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { FlaskConical, Shield, Users, Settings, CreditCard, Key } from "lucide-react"
import { useAuth } from "@/contexts/AuthContext"
import { AlertCircle } from "lucide-react"
import { Separator } from "@/components/ui/separator"

export default function TestToolPage() {
  const { user } = useAuth()
  
  return (
    <DashboardLayout>
      <Card className="shadow-lg bg-card text-card-foreground border-border">
        <CardHeader>
          <div className="flex items-center space-x-2">
            <FlaskConical className="h-6 w-6 text-pink-500" />
            <CardTitle className="text-slate-900 dark:text-slate-100">测试工具</CardTitle>
          </div>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            开发测试相关工具和管理功能
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-6">
            <div className="text-center py-8">
              <FlaskConical className="h-16 w-16 text-pink-500 mx-auto mb-4" />
              <h3 className="text-lg font-semibold mb-2">测试工具说明</h3>
              <p className="text-slate-600 dark:text-slate-400 mb-4">
                管理员可以在侧边栏的管理员区域访问各种管理工具
              </p>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-2xl mx-auto">
                <div className="flex items-center space-x-3 p-3 bg-slate-50 dark:bg-slate-800 rounded-lg">
                  <Shield className="h-5 w-5 text-red-500" />
                  <span className="text-sm">OAuth 测试工具</span>
                </div>
                <div className="flex items-center space-x-3 p-3 bg-slate-50 dark:bg-slate-800 rounded-lg">
                  <Users className="h-5 w-5 text-blue-500" />
                  <span className="text-sm">用户管理</span>
                </div>
                <div className="flex items-center space-x-3 p-3 bg-slate-50 dark:bg-slate-800 rounded-lg">
                  <Settings className="h-5 w-5 text-green-500" />
                  <span className="text-sm">系统配置</span>
                </div>
                <div className="flex items-center space-x-3 p-3 bg-slate-50 dark:bg-slate-800 rounded-lg">
                  <CreditCard className="h-5 w-5 text-purple-500" />
                  <span className="text-sm">订阅计划</span>
                </div>
                <div className="flex items-center space-x-3 p-3 bg-slate-50 dark:bg-slate-800 rounded-lg">
                  <Key className="h-5 w-5 text-orange-500" />
                  <span className="text-sm">激活码管理</span>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 管理员面板 - 仅管理员可见 */}
      {user?.is_admin && (
        <>
          <div className="flex items-center space-x-4">
            <Separator className="flex-1" />
            <div className="flex items-center space-x-2 px-3 py-1 bg-red-50 dark:bg-red-900/20 rounded-full border border-red-200 dark:border-red-800">
              <AlertCircle className="h-4 w-4 text-red-500" />
              <span className="text-sm font-medium text-red-700 dark:text-red-300">管理员专用区域</span>
            </div>
            <Separator className="flex-1" />
          </div>
        </>
      )}
    </DashboardLayout>
  )
}
