"use client"

import Link from "next/link"
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet"
import { Button } from "@/components/ui/button"
import {
  Menu,
  LayoutDashboard,
  CreditCard,
  DollarSign,
  BookOpen,
  SettingsIcon,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { usePathname } from "next/navigation"

const navItems = [
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

// 设置项定义 - 暂时不开放给用户使用
const settingsItem = {
  href: "/settings",
  label: "设置",
  icon: SettingsIcon,
  activeColor: "bg-slate-200 text-slate-800 dark:bg-slate-600 dark:text-slate-200",
}

export function MainHeader() {
  const pathname = usePathname()
  return (
    <header className="sticky top-0 z-30 flex h-12 items-center gap-4 bg-background/80 backdrop-blur-sm px-4 md:px-6">
      <Sheet>
        <SheetTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="shrink-0 md:hidden bg-transparent text-foreground hover:bg-accent hover:text-accent-foreground"
          >
            <Menu className="h-5 w-5" />
            <span className="sr-only">切换导航菜单</span>
          </Button>
        </SheetTrigger>
        <SheetContent side="left" className="flex flex-col bg-card shadow-xl p-0">
          <div className="flex h-16 items-center justify-center px-4">
            <Link
              href="/"
              className="font-semibold text-lg text-card-foreground"
              prefetch={false}
            >
              Claude Duck
            </Link>
          </div>
          <nav className="grid gap-2 text-lg font-medium p-4">
            {/* 只显示主要导航项，不包含设置 */}
            {navItems.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "flex items-center gap-3 rounded-full px-3 py-2 transition-all",
                  pathname === item.href
                    ? item.activeColor
                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
                )}
                prefetch={false}
              >
                <item.icon className="h-5 w-5" />
                {item.label}
              </Link>
            ))}
            {/* 设置按钮已隐藏 - 暂时不开放给用户使用 */}
            {/*
            <Link
              href="/settings"
              className={cn(
                "flex items-center gap-3 rounded-full px-3 py-2 transition-all",
                pathname === settingsItem.href
                  ? settingsItem.activeColor
                  : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
              )}
              prefetch={false}
            >
              <settingsItem.icon className="h-5 w-5" />
              {settingsItem.label}
            </Link>
            */}
          </nav>
        </SheetContent>
      </Sheet>

      {/* 已删除标题栏 */}
    </header>
  )
}
