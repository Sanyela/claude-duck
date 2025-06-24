"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Separator } from "@/components/ui/separator"
import { Badge } from "@/components/ui/badge"
import { Calendar, Search, RotateCcw } from "lucide-react"
import { 
  LineChart, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  ReferenceLine
} from "recharts"

// 扩展历史数据以便于折线图显示
const usageHistory = [
  { id: "U001", date: "2025-06-22", action: "模型调用 (GPT-4)", credits: "-50", balance: 1250, timestamp: "2025-06-22" },
  { id: "U002", date: "2025-06-21", action: "积分充值", credits: "+1000", balance: 1300, timestamp: "2025-06-21" },
  { id: "U003", date: "2025-06-20", action: "API 使用", credits: "-150", balance: 300, timestamp: "2025-06-20" },
  { id: "U004", date: "2025-06-19", action: "模型调用 (Claude)", credits: "-200", balance: 450, timestamp: "2025-06-19" },
  { id: "U005", date: "2025-06-18", action: "积分充值", credits: "+500", balance: 650, timestamp: "2025-06-18" },
  { id: "U006", date: "2025-06-17", action: "API 使用", credits: "-100", balance: 150, timestamp: "2025-06-17" },
  { id: "U007", date: "2025-06-16", action: "模型调用 (Llama)", credits: "-30", balance: 250, timestamp: "2025-06-16" },
  { id: "U008", date: "2025-06-15", action: "初始积分", credits: "+280", balance: 280, timestamp: "2025-06-15" },
]

// 为折线图准备数据
const chartData = [...usageHistory].reverse().map(item => ({
  date: item.date,
  balance: parseInt(typeof item.balance === 'string' ? item.balance : String(item.balance)),
  id: item.id
}))

export function CreditsContent() {
  const [selectedRecord, setSelectedRecord] = useState<string | null>(null)
  const [startDate, setStartDate] = useState("2025-06-15")
  const [endDate, setEndDate] = useState("2025-06-22")

  // 计算这段时间内消耗的积分
  const totalUsage = usageHistory
    .filter(item => item.credits.startsWith('-'))
    .reduce((sum, item) => sum + Math.abs(parseInt(item.credits)), 0)

  // 处理记录点击事件
  const handleRecordClick = (recordId: string) => {
    setSelectedRecord(recordId === selectedRecord ? null : recordId)
  }

  return (
    <div className="space-y-6">
      <Card className="shadow-lg bg-card text-card-foreground border-border">
        <CardContent className="pt-6 px-6">
          <div className="flex flex-wrap gap-2 mb-4">
            <Badge className="bg-red-500 text-white hover:bg-red-600 px-2.5 py-1.5 shadow-sm">
              <span className="font-semibold">{totalUsage.toLocaleString()}</span> <span className="ml-1 text-xs opacity-90">总消耗</span>
            </Badge>
            <Badge className="bg-sky-500 text-white hover:bg-sky-600 px-2.5 py-1.5 shadow-sm">
              <span className="font-semibold">350</span> <span className="ml-1 text-xs opacity-90">TPM</span>
            </Badge>
            <Badge className="bg-green-500 text-white hover:bg-green-600 px-2.5 py-1.5 shadow-sm">
              <span className="font-semibold">7.5k</span> <span className="ml-1 text-xs opacity-90">RPM</span>
            </Badge>
          </div>
          
          <Separator className="my-4" />
          
          <div className="flex flex-col sm:flex-row gap-4 mb-4">
            <div className="flex items-center gap-2 flex-1">
              <div className="text-sm font-medium">开始时间:</div>
              <div className="relative flex-1">
                <input 
                  type="date" 
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
                <Calendar className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              </div>
            </div>
            <div className="flex items-center gap-2 flex-1">
              <div className="text-sm font-medium">结束时间:</div>
              <div className="relative flex-1">
                <input 
                  type="date" 
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  className="w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
                <Calendar className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              </div>
            </div>
            <div className="flex gap-2">
              <Button size="sm" className="bg-sky-500 hover:bg-sky-600 text-white">
                <Search className="mr-1 h-4 w-4" /> 查询
              </Button>
              <Button size="sm" variant="outline" className="border-border">
                <RotateCcw className="mr-1 h-4 w-4" /> 重置
              </Button>
            </div>
          </div>
          
          <Separator className="my-4" />
          
          <div className="flex items-center justify-end mb-2 text-sm text-muted-foreground">
            <span className="inline-block w-3 h-3 rounded-full bg-sky-500 mr-1"></span> 积分余额
          </div>
          
          <div className="space-y-6">
            {/* 折线图区域 */}
            <div className="h-72 w-full">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart
                  data={chartData}
                  margin={{ top: 5, right: 20, left: 0, bottom: 5 }}
                >
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#888" opacity={0.2} />
                  <XAxis 
                    dataKey="date" 
                    stroke="#888" 
                    fontSize={12} 
                    tickLine={false}
                    axisLine={false}
                  />
                  <YAxis
                    stroke="#888"
                    fontSize={12}
                    tickLine={false}
                    axisLine={false}
                    tickFormatter={(value) => `${value}`}
                  />
                  <Tooltip 
                    formatter={(value) => [`${value} 积分`, '余额']}
                    labelFormatter={(label) => `日期: ${label}`}
                    contentStyle={{ 
                      backgroundColor: 'var(--card)', 
                      borderColor: 'var(--border)',
                      borderRadius: '0.5rem',
                      boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)'
                    }}
                  />
                  <Line
                    type="monotone"
                    dataKey="balance"
                    stroke="#0ea5e9"
                    strokeWidth={2}
                    dot={{ r: 4 }}
                    activeDot={{ r: 6, fill: '#0ea5e9' }}
                  />
                  {selectedRecord && (
                    <ReferenceLine 
                      x={chartData.find(item => item.id === selectedRecord)?.date} 
                      stroke="#0ea5e9" 
                      strokeWidth={2}
                      strokeDasharray="3 3" 
                    />
                  )}
                </LineChart>
              </ResponsiveContainer>
            </div>
            
            {/* 历史记录表格 */}
            <div className="overflow-auto">
              <Table>
                <TableHeader>
                  <TableRow className="border-border">
                    <TableHead>日期</TableHead>
                    <TableHead>操作</TableHead>
                    <TableHead className="text-right">积分变动</TableHead>
                    <TableHead className="text-right">余额</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {usageHistory.map((item) => (
                    <TableRow 
                      key={item.id} 
                      className={`border-border cursor-pointer transition-colors ${selectedRecord === item.id ? 'bg-sky-50 dark:bg-sky-900/20' : 'hover:bg-muted/50'}`}
                      onClick={() => handleRecordClick(item.id)}
                    >
                      <TableCell className="font-medium">{item.date}</TableCell>
                      <TableCell className="text-muted-foreground">{item.action}</TableCell>
                      <TableCell
                        className={`text-right font-medium ${item.credits.startsWith("+") ? "text-green-500 dark:text-green-400" : "text-red-500 dark:text-red-400"}`}
                      >
                        {item.credits}
                      </TableCell>
                      <TableCell className="text-right text-muted-foreground">{typeof item.balance === 'number' ? item.balance : item.balance}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}