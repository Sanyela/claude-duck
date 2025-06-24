import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { XCircle, RotateCcw, Tag } from "lucide-react"

const currentSubscription = {
  planName: "专业版 🌟",
  price: "¥99/月",
  status: "active", // 'active', 'cancelled', 'expired'
  nextBillingDate: "2025-07-15",
  features: ["无限项目", "优先支持", "高级分析", "API访问"],
}

const history = [
  { id: "INV001", date: "2025-06-15", plan: "专业版", amount: "¥99.00", status: "已支付" },
  { id: "INV002", date: "2025-05-15", plan: "专业版", amount: "¥99.00", status: "已支付" },
  { id: "INV003", date: "2025-04-15", plan: "基础版", amount: "¥29.00", status: "已支付" },
]

export default function SubscriptionPage() {
  const isSubscriptionActive = currentSubscription.status === "active"

  return (
    <DashboardLayout>
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <Card className="shadow-lg bg-card text-card-foreground border-border">
            <CardHeader>
              <CardTitle>当前订阅</CardTitle>
              <CardDescription>查看您当前的订阅计划详情。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex justify-between items-center">
                <h3 className="text-xl font-semibold text-sky-500 dark:text-sky-400">{currentSubscription.planName}</h3>
                <Badge
                  variant={isSubscriptionActive ? "default" : "destructive"}
                  className={
                    isSubscriptionActive
                      ? "bg-green-500 text-white dark:bg-green-600 dark:text-green-50"
                      : "bg-red-500 text-white dark:bg-red-600 dark:text-red-50"
                  }
                >
                  {isSubscriptionActive ? "有效" : currentSubscription.status === "cancelled" ? "已取消" : "已过期"}
                </Badge>
              </div>
              <p className="text-3xl font-bold">{currentSubscription.price}</p>
              {isSubscriptionActive && (
                <p className="text-sm text-muted-foreground">下次账单日期: {currentSubscription.nextBillingDate}</p>
              )}
              <div>
                <h4 className="font-medium mb-1">包含功能:</h4>
                <ul className="list-disc list-inside space-y-1 text-sm text-muted-foreground">
                  {currentSubscription.features.map((feature) => (
                    <li key={feature}>{feature}</li>
                  ))}
                </ul>
              </div>
            </CardContent>
            {/* 删除了取消订阅按钮 */}
          </Card>

          <Card className="shadow-lg bg-card text-card-foreground border-border">
            <CardHeader>
              <CardTitle>订阅历史</CardTitle>
              <CardDescription>查看您过去的订阅和付款记录。</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow className="border-border">
                    <TableHead>账单ID</TableHead>
                    <TableHead>日期</TableHead>
                    <TableHead>计划</TableHead>
                    <TableHead className="text-right">金额</TableHead>
                    <TableHead className="text-right">状态</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {history.map((item) => (
                    <TableRow key={item.id} className="border-border">
                      <TableCell className="font-medium">{item.id}</TableCell>
                      <TableCell className="text-muted-foreground">{item.date}</TableCell>
                      <TableCell className="text-muted-foreground">{item.plan}</TableCell>
                      <TableCell className="text-right text-muted-foreground">{item.amount}</TableCell>
                      <TableCell className="text-right">
                        <Badge
                          variant={item.status === "已支付" ? "default" : "secondary"}
                          className={
                            item.status === "已支付"
                              ? "bg-green-100 text-green-700 dark:bg-green-700/30 dark:text-green-300"
                              : "bg-slate-100 text-slate-700 dark:bg-slate-700/30 dark:text-slate-300"
                          }
                        >
                          {item.status}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </div>

        <div className="lg:col-span-1 space-y-6">
          <Card className="shadow-lg bg-card text-card-foreground border-border">
            <CardHeader>
              <CardTitle>兑换优惠券</CardTitle>
              <CardDescription>有优惠券代码吗？在这里输入。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              <Label htmlFor="coupon">优惠券代码</Label>
              <Input
                id="coupon"
                placeholder="例如：SUMMER25"
                className="bg-input border-border placeholder:text-muted-foreground"
              />
            </CardContent>
            <CardFooter>
              <Button className="w-full bg-sky-500 hover:bg-sky-600 text-primary-foreground">
                <Tag className="mr-2 h-4 w-4" /> 应用优惠券
              </Button>
            </CardFooter>
          </Card>
        </div>
      </div>
    </DashboardLayout>
  )
}
