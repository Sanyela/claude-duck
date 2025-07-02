"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { useTheme } from "next-themes"
import {
  LayoutDashboard,
  CreditCard,
  DollarSign,
  BookOpen,
  SettingsIcon,
  Sun,
  Moon,
  Shield,
  Users,
  Settings,
  Key,
  Megaphone,
  BarChart3,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { useAuth } from "@/contexts/AuthContext"
import { Separator } from "@/components/ui/separator"

// 用户功能导航项
const userNavItems = [
  {
    href: "/",
    label: "仪表板",
    icon: LayoutDashboard,
    activeColor: "bg-sky-100 text-sky-700 dark:bg-sky-500/20 dark:text-sky-300",
  },
  {
    href: "/subscription",
    label: "订阅管理",
    icon: CreditCard,
    activeColor: "bg-green-100 text-green-700 dark:bg-green-500/20 dark:text-green-300",
  },
  {
    href: "/credits",
    label: "积分管理",
    icon: DollarSign,
    activeColor: "bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300",
  },
  {
    href: "/resources",
    label: "资源中心",
    icon: BookOpen,
    activeColor: "bg-purple-100 text-purple-700 dark:bg-purple-500/20 dark:text-purple-300",
  },
]

// 管理员功能导航项
const adminNavItems = [
  {
    href: "/admin/views",
    label: "数据看板",
    icon: BarChart3,
    activeColor: "bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300",
  },
  {
    href: "/admin/users",
    label: "用户管理",
    icon: Users,
    activeColor: "bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300",
  },
  {
    href: "/admin/configs",
    label: "系统配置",
    icon: Settings,
    activeColor: "bg-green-100 text-green-700 dark:bg-green-500/20 dark:text-green-300",
  },
  {
    href: "/admin/plans",
    label: "订阅计划",
    icon: CreditCard,
    activeColor: "bg-purple-100 text-purple-700 dark:bg-purple-500/20 dark:text-purple-300",
  },
  {
    href: "/admin/codes",
    label: "激活码管理",
    icon: Key,
    activeColor: "bg-orange-100 text-orange-700 dark:bg-orange-500/20 dark:text-orange-300",
  },
  {
    href: "/admin/announcements",
    label: "公告管理",
    icon: Megaphone,
    activeColor: "bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300",
  },
  {
    href: "/admin/oauth",
    label: "OAuth 测试工具",
    icon: Shield,
    activeColor: "bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300",
  },
]

const settingsItem = {
  href: "/settings",
  label: "设置",
  icon: SettingsIcon,
  activeColor: "bg-slate-200 text-slate-800 dark:bg-slate-600 dark:text-slate-200",
}

const inactiveItemClasses = "text-slate-600 hover:bg-slate-500/10 dark:text-slate-400 dark:hover:bg-slate-500/10"

export function Sidebar() {
  const pathname = usePathname()
  const { theme, setTheme } = useTheme()
  const { user } = useAuth()

  const currentTheme = theme || "light"

  const toggleTheme = () => {
    setTheme(currentTheme === "dark" ? "light" : "dark")
  }

  return (
    <aside className="hidden md:flex md:flex-col md:w-56 bg-background">
      <div className="flex h-16 items-center justify-center px-4">
        <Link href="/" className="font-semibold text-lg text-foreground" prefetch={false}>
          Claude Duck
        </Link>
      </div>
      
      <nav className="flex-1 overflow-y-auto py-4 space-y-1">
        {/* 用户功能 */}
        {userNavItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              "flex items-center gap-3 rounded-full px-4 py-2 mx-2 text-sm font-medium transition-colors",
              pathname === item.href ? item.activeColor : inactiveItemClasses,
            )}
            prefetch={false}
          >
            <item.icon className="h-4 w-4" />
            {item.label}
          </Link>
        ))}

        {/* 管理员功能分割线 */}
        {user?.is_admin && (
          <>
            <div className="px-4 py-2">
              <Separator />
            </div>
            
            {adminNavItems.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "flex items-center gap-3 rounded-full px-4 py-2 mx-2 text-sm font-medium transition-colors",
                  pathname === item.href ? item.activeColor : inactiveItemClasses,
                )}
                prefetch={false}
              >
                <item.icon className="h-4 w-4" />
                {item.label}
              </Link>
            ))}
          </>
        )}
      </nav>

      <div className="mt-auto p-2 space-y-1 pt-2">
        <Button
          variant="ghost"
          onClick={toggleTheme}
          className={cn(
            "w-full flex items-center justify-start gap-3 rounded-full px-4 py-2 text-sm font-medium transition-colors",
            inactiveItemClasses,
          )}
        >
          {currentTheme === "dark" ? (
            <Sun className="h-4 w-4 text-yellow-400" />
          ) : (
            <Moon className="h-4 w-4 text-sky-500" />
          )}
          <span>切换主题</span>
        </Button>
        {/* 设置按钮已隐藏 - 暂时不开放给用户使用 */}
        {/*
        <Link
          href="/settings"
          className={cn(
            "flex items-center gap-3 rounded-full px-4 py-2 text-sm font-medium transition-colors",
            pathname === settingsItem.href ? settingsItem.activeColor : inactiveItemClasses,
          )}
          prefetch={false}
        >
          <settingsItem.icon className="h-4 w-4" />
          {settingsItem.label}
        </Link>
        */}
      </div>
    </aside>
  )
}
