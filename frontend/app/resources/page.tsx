import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { BookOpen, FileText, CodeIcon, HelpCircle } from "lucide-react"
import Link from "next/link"

const resources = [
  { title: "快速入门指南 🚀", description: "了解如何快速开始使用我们的平台。", icon: FileText, href: "#" },
  { title: "API 文档 💻", description: "详细的API参考和集成指南。", icon: CodeIcon, href: "#" },
  { title: "常见问题 (FAQ) 🤔", description: "查找常见问题的答案。", icon: HelpCircle, href: "#" },
  { title: "教程与案例 📚", description: "学习如何通过实际案例使用我们的功能。", icon: BookOpen, href: "#" },
]

export default function ResourcesPage() {
  return (
    <DashboardLayout currentPageTitle="资源中心">
      <div className="space-y-6">
        <Card className="shadow-lg bg-card text-card-foreground border-border">
          <CardHeader>
            <CardTitle className="text-2xl">欢迎来到资源中心</CardTitle>
            <CardDescription>在这里您可以找到帮助您更好使用我们服务的各类文档和指南。</CardDescription>
          </CardHeader>
        </Card>

        <div className="grid gap-6 md:grid-cols-2">
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
                >
                  了解更多 &rarr;
                </Link>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </DashboardLayout>
  )
}
