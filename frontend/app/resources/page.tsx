import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Download } from "lucide-react"
import Link from "next/link"

const resources = [
  { 
    title: "安装 Claude Duck 🦆", 
    description: "安装Claude浏览器插件，获得更便捷的AI助手体验。", 
    icon: Download, 
    href: "https://swjqc4r0111.feishu.cn/docx/SqnkdO9CKojJ33xCgc3ceYbLn7e"
  }
]

export default function ResourcesPage() {
  return (
    <DashboardLayout>
      <div className="space-y-6">
        <h1 className="text-3xl font-bold mb-6">🎉 欢迎使用资源中心</h1>

        <div className="grid gap-6 md:grid-cols-1">
          {resources.map((resource) => (
            <Card
              key={resource.title}
              className="shadow-lg bg-card text-card-foreground border-border hover:shadow-xl transition-shadow"
            >
              <CardHeader className="flex flex-row items-start gap-4">
                <resource.icon className="h-8 w-8 text-sky-500 dark:text-sky-400 mt-1" />
                <div>
                  <CardTitle>{resource.title}</CardTitle>
                  <CardDescription>{resource.description}</CardDescription>
                </div>
              </CardHeader>
              <CardContent>
                <Link
                  href={resource.href}
                  className="text-sm font-medium text-sky-500 hover:underline dark:text-sky-400 dark:hover:text-sky-300"
                  prefetch={false}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  立即安装 &rarr;
                </Link>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </DashboardLayout>
  )
}
