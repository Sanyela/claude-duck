import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Progress } from "@/components/ui/progress"
import { DollarSign, Zap, PlusCircle } from "lucide-react"

const creditBalance = 1250 // Placeholder
const monthlyAllowance = 5000
const rechargeRate = "¥0.01 / 积分"

const modelCosts = [
  { model: "GPT-4 Turbo", inputCost: "¥0.05 / 1K tokens", outputCost: "¥0.15 / 1K tokens" },
  { model: "Claude 3 Sonnet", inputCost: "¥0.02 / 1K tokens", outputCost: "¥0.08 / 1K tokens" },
  { model: "Llama 3 70B", inputCost: "¥0.03 / 1K tokens", outputCost: "¥0.10 / 1K tokens" },
]

const usageHistory = [
  { id: "U001", date: "2025-06-22", action: "模型调用 (GPT-4)", credits: "-50", balance: "1250" },
  { id: "U002", date: "2025-06-21", action: "积分充值", credits: "+1000", balance: "1300" },
  { id: "U003", date: "2025-06-20", action: "API 使用", credits: "-150", balance: "300" },
]

export default function CreditsPage() {
  const creditUsagePercentage = Math.min((creditBalance / monthlyAllowance) * 100, 100)

  return (
    <DashboardLayout currentPageTitle="积分管理">
      <div className="space-y-6">
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <Card className="lg:col-span-1 shadow-lg bg-card text-card-foreground border-border">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">当前积分余额</CardTitle>
              <DollarSign className="h-5 w-5 text-green-500 dark:text-green-400" />
            </CardHeader>
            <CardContent>
              <div className="text-4xl font-bold">{creditBalance.toLocaleString()}</div>
              <p className="text-xs text-muted-foreground mt-1">每月限额: {monthlyAllowance.toLocaleString()} 积分</p>
              <Progress value={creditUsagePercentage} className="mt-2 h-2 [&>*]:bg-sky-500" />
            </CardContent>
          </Card>
          <Card className="shadow-lg bg-card text-card-foreground border-border">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">充值速率</CardTitle>
              <Zap className="h-5 w-5 text-sky-500 dark:text-sky-400" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{rechargeRate}</div>
              <p className="text-xs text-muted-foreground mt-1">按需购买，灵活使用。</p>
              <Button size="sm" className="mt-3 bg-sky-500 hover:bg-sky-600 text-primary-foreground">
                <PlusCircle className="mr-2 h-4 w-4" /> 充值积分 💰
              </Button>
            </CardContent>
          </Card>
        </div>

        <Card className="shadow-lg bg-card text-card-foreground border-border">
          <CardHeader>
            <CardTitle>模型使用成本</CardTitle>
            <CardDescription>了解不同AI模型的使用成本。</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow className="border-border">
                  <TableHead>模型名称</TableHead>
                  <TableHead>输入成本 (每1K tokens)</TableHead>
                  <TableHead>输出成本 (每1K tokens)</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {modelCosts.map((model) => (
                  <TableRow key={model.model} className="border-border">
                    <TableCell className="font-medium">{model.model}</TableCell>
                    <TableCell className="text-muted-foreground">{model.inputCost}</TableCell>
                    <TableCell className="text-muted-foreground">{model.outputCost}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Card className="shadow-lg bg-card text-card-foreground border-border">
          <CardHeader>
            <CardTitle>使用历史记录</CardTitle>
            <CardDescription>跟踪您的积分消耗和充值记录。</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow className="border-border">
                  <TableHead>记录ID</TableHead>
                  <TableHead>日期</TableHead>
                  <TableHead>操作</TableHead>
                  <TableHead className="text-right">积分变动</TableHead>
                  <TableHead className="text-right">余额</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {usageHistory.map((item) => (
                  <TableRow key={item.id} className="border-border">
                    <TableCell className="font-medium">{item.id}</TableCell>
                    <TableCell className="text-muted-foreground">{item.date}</TableCell>
                    <TableCell className="text-muted-foreground">{item.action}</TableCell>
                    <TableCell
                      className={`text-right font-medium ${item.credits.startsWith("+") ? "text-green-500 dark:text-green-400" : "text-red-500 dark:text-red-400"}`}
                    >
                      {item.credits}
                    </TableCell>
                    <TableCell className="text-right text-muted-foreground">{item.balance}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </DashboardLayout>
  )
}
