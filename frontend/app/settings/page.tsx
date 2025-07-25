export const dynamic = 'force-dynamic'

import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { User, Lock, Palette } from "lucide-react"

export default function SettingsPage() {
  const currentUser = {
    name: "当前用户名",
    email: "user@example.com",
  }

  return (
    <DashboardLayout>
      <div className="space-y-8 max-w-3xl mx-auto">
        <Card className="shadow-lg bg-card text-card-foreground border-border">
          <CardHeader>
            <div className="flex items-center gap-3">
              <User className="h-6 w-6 text-sky-500 dark:text-sky-400" />
              <CardTitle>账户设置</CardTitle>
            </div>
            <CardDescription>管理您的个人信息和账户安全。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="name">名称</Label>
              <Input
                id="name"
                defaultValue={currentUser.name}
                className="bg-input border-border placeholder:text-muted-foreground"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">邮箱地址</Label>
              <Input
                id="email"
                type="email"
                defaultValue={currentUser.email}
                className="bg-input border-border placeholder:text-muted-foreground"
              />
            </div>
          </CardContent>
          <CardFooter className="border-t border-border pt-4">
            <Button className="bg-sky-500 hover:bg-sky-600 text-primary-foreground">保存更改</Button>
          </CardFooter>
        </Card>

        <Card className="shadow-lg bg-card text-card-foreground border-border">
          <CardHeader>
            <div className="flex items-center gap-3">
              <Lock className="h-6 w-6 text-sky-500 dark:text-sky-400" />
              <CardTitle>密码安全</CardTitle>
            </div>
            <CardDescription>定期更新您的密码以保护账户安全。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="current-password">当前密码</Label>
              <Input
                id="current-password"
                type="password"
                className="bg-input border-border placeholder:text-muted-foreground"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="new-password">新密码</Label>
              <Input
                id="new-password"
                type="password"
                className="bg-input border-border placeholder:text-muted-foreground"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirm-password">确认新密码</Label>
              <Input
                id="confirm-password"
                type="password"
                className="bg-input border-border placeholder:text-muted-foreground"
              />
            </div>
          </CardContent>
          <CardFooter className="border-t border-border pt-4">
            <Button className="bg-sky-500 hover:bg-sky-600 text-primary-foreground">更新密码</Button>
          </CardFooter>
        </Card>

        <Card className="shadow-lg bg-card text-card-foreground border-border">
          <CardHeader>
            <div className="flex items-center gap-3">
              <Palette className="h-6 w-6 text-sky-500 dark:text-sky-400" />
              <CardTitle>外观主题</CardTitle>
            </div>
            <CardDescription>选择您喜欢的外观模式。</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">请使用侧边栏的 &quot;切换主题&quot; 按钮来更改应用的外观模式。</p>
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  )
}
