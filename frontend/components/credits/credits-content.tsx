"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Separator } from "@/components/ui/separator"
import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Calendar, Search, RotateCcw, Loader2, AlertCircle } from "lucide-react"
import { 
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { 
  LineChart, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip as RechartsTooltip, 
  ResponsiveContainer,
  ReferenceLine
} from "recharts"
import { creditsAPI, type CreditBalance, type CreditUsageHistory } from "@/api/credits"

// 图表数据接口
interface ChartDataPoint {
  date: string;
  balance: number;
  id: string;
}

export function CreditsContent() {
  const [loading, setLoading] = useState(true)
  const [creditBalance, setCreditBalance] = useState<CreditBalance | null>(null)
  const [usageHistory, setUsageHistory] = useState<CreditUsageHistory[]>([])
  const [chartData, setChartData] = useState<ChartDataPoint[]>([])
  const [selectedRecord, setSelectedRecord] = useState<string | null>(null)
  const [startDate, setStartDate] = useState(() => {
    const date = new Date()
    date.setDate(date.getDate() - 7) // 默认显示最近7天
    return date.toISOString().split('T')[0]
  })
  const [endDate, setEndDate] = useState(() => {
    return new Date().toISOString().split('T')[0]
  })
  const [error, setError] = useState<string | null>(null)

  // 加载数据
  const loadData = async (startDateParam?: string, endDateParam?: string) => {
    setLoading(true)
    setError(null)

    try {
      const [balanceResult, historyResult] = await Promise.all([
        creditsAPI.getBalance(),
        creditsAPI.getUsageHistory({
          start_date: startDateParam || startDate,
          end_date: endDateParam || endDate,
          page_size: 50 // 获取更多数据用于图表
        })
      ])

      if (balanceResult.success && balanceResult.data) {
        setCreditBalance(balanceResult.data)
      }

      if (historyResult.success && historyResult.data) {
        // 提取历史记录数组
        const historyArray = historyResult.data.history || [];
        
        setUsageHistory(historyArray)
        
        // 生成图表数据
        const chartPoints = generateChartData(historyArray, balanceResult.data || null)
        setChartData(chartPoints)
      }
    } catch (err: any) {
      setError("加载积分数据失败")
    }

    setLoading(false)
  }

  // 生成图表数据
  const generateChartData = (history: CreditUsageHistory[], balance: CreditBalance | null): ChartDataPoint[] => {
    if (!balance || !Array.isArray(history) || !history.length) return []

    // 按日期分组并计算每日积分变化
    const dailyData = new Map<string, { totalUsed: number, lastId: string }>()
    
    history.forEach(record => {
      const date = new Date(record.timestamp).toISOString().split('T')[0]
      // amount是负数，转为正数计算
      const pointsUsed = Math.abs(record.amount)
      
      const existing = dailyData.get(date) || { totalUsed: 0, lastId: record.id }
      existing.totalUsed += pointsUsed
      // 使用数字大的ID作为该日期的代表ID
      if (parseInt(record.id) > parseInt(existing.lastId)) {
        existing.lastId = record.id
      }
      dailyData.set(date, existing)
    })

    // 生成图表点
    const points: ChartDataPoint[] = []
    let currentBalance = balance.available_points

    // 从最新日期往回推算余额
    const sortedDates = Array.from(dailyData.keys()).sort().reverse()
    
    sortedDates.forEach(date => {
      const dayData = dailyData.get(date)!
      points.unshift({
        date,
        balance: currentBalance,
        id: dayData.lastId
      })
      currentBalance += dayData.totalUsed // 往前推算时加回消耗的积分
    })

    return points.reverse()
  }

  // 查询数据
  const handleSearch = () => {
    loadData(startDate, endDate)
  }

  // 重置查询
  const handleReset = () => {
    const defaultEndDate = new Date().toISOString().split('T')[0]
    const defaultStartDate = new Date()
    defaultStartDate.setDate(defaultStartDate.getDate() - 7)
    const defaultStartDateStr = defaultStartDate.toISOString().split('T')[0]
    
    setStartDate(defaultStartDateStr)
    setEndDate(defaultEndDate)
    loadData(defaultStartDateStr, defaultEndDate)
  }

  // 处理记录点击事件
  const handleRecordClick = (recordId: string) => {
    setSelectedRecord(recordId === selectedRecord ? null : recordId)
  }

  useEffect(() => {
    loadData()
  }, [])

  // 计算统计数据
  const totalUsage = Array.isArray(usageHistory) ? usageHistory.reduce((sum, item) => sum + Math.abs(item.amount), 0) : 0
  const uniqueModels = Array.isArray(usageHistory) ? new Set(usageHistory.map(item => item.relatedModel)).size : 0

  return (
    <div className="space-y-6">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <Card className="shadow-lg bg-card text-card-foreground border-border">
        <CardContent className="pt-6 px-6">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin" />
              <span className="ml-3 text-lg">加载中...</span>
            </div>
          ) : (
            <>
              <div className="flex flex-wrap gap-2 mb-4">
                <Badge className="bg-red-500 text-white hover:bg-red-600 px-2.5 py-1.5 shadow-sm">
                  <span className="font-semibold">{totalUsage.toLocaleString()}</span> 
                  <span className="ml-1 text-xs opacity-90">总消耗</span>
                </Badge>
                <Badge className="bg-sky-500 text-white hover:bg-sky-600 px-2.5 py-1.5 shadow-sm">
                  <span className="font-semibold">{creditBalance?.available_points?.toLocaleString() || 0}</span> 
                  <span className="ml-1 text-xs opacity-90">可用积分</span>
                </Badge>
                <Badge className="bg-green-500 text-white hover:bg-green-600 px-2.5 py-1.5 shadow-sm">
                  <span className="font-semibold">{uniqueModels}</span> 
                  <span className="ml-1 text-xs opacity-90">使用模型</span>
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
                  <Button size="sm" className="bg-sky-500 hover:bg-sky-600 text-white" onClick={handleSearch}>
                    <Search className="mr-1 h-4 w-4" /> 查询
                  </Button>
                  <Button size="sm" variant="outline" className="border-border" onClick={handleReset}>
                    <RotateCcw className="mr-1 h-4 w-4" /> 重置
                  </Button>
                </div>
              </div>
              
              <Separator className="my-4" />
              
              <div className="flex items-center justify-end mb-2 text-sm text-muted-foreground">
                <span className="inline-block w-3 h-3 rounded-full bg-sky-500 mr-1"></span> 积分余额趋势
              </div>
              
              <div className="space-y-6">
                {/* 折线图区域 */}
                {chartData.length > 0 ? (
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
                        <RechartsTooltip 
                          formatter={(value: any) => [`${value} 积分`, '余额']}
                          labelFormatter={(label: any) => `日期: ${label}`}
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
                ) : (
                  <div className="h-72 flex items-center justify-center">
                    <p className="text-muted-foreground">暂无图表数据</p>
                  </div>
                )}
                
                {/* 历史记录表格 */}
                <div className="overflow-auto">
                  {Array.isArray(usageHistory) && usageHistory.length > 0 ? (
                    <Table>
                      <TableHeader>
                        <TableRow className="border-border">
                          <TableHead>日期</TableHead>
                          <TableHead>模型</TableHead>
                          <TableHead>Token使用</TableHead>
                          <TableHead className="text-right">消耗积分</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {usageHistory.map((item) => (
                          <TableRow 
                            key={item.id} 
                            className={`border-border cursor-pointer transition-colors ${selectedRecord === item.id ? 'bg-sky-50 dark:bg-sky-900/20' : 'hover:bg-muted/50'}`}
                            onClick={() => handleRecordClick(item.id)}
                          >
                            <TableCell className="font-medium">
                              {new Date(item.timestamp).toLocaleString()}
                            </TableCell>
                            <TableCell>
                              <Badge variant="outline" className="font-mono text-xs">
                                {item.relatedModel.split('-').slice(0, 2).join('-')}
                              </Badge>
                            </TableCell>
                            <TableCell>
                              <div className="flex gap-1 items-center">
                                <TooltipProvider delayDuration={0}>
                                  <Tooltip>
                                    <TooltipTrigger asChild>
                                      <Badge 
                                        variant="secondary" 
                                        className="font-mono text-xs bg-green-50 hover:bg-green-100 dark:bg-green-900/20 dark:hover:bg-green-900/30 transition-colors"
                                      >
                                        {item.input_tokens}
                                      </Badge>
                                    </TooltipTrigger>
                                    <TooltipContent>
                                      <p>输入Token数量</p>
                                    </TooltipContent>
                                  </Tooltip>
                                </TooltipProvider>
                                
                                <TooltipProvider delayDuration={0}>
                                  <Tooltip>
                                    <TooltipTrigger asChild>
                                      <Badge 
                                        variant="secondary" 
                                        className="font-mono text-xs bg-red-50 hover:bg-red-100 dark:bg-red-900/20 dark:hover:bg-red-900/30 transition-colors text-red-600 dark:text-red-400"
                                      >
                                        {item.output_tokens}
                                      </Badge>
                                    </TooltipTrigger>
                                    <TooltipContent>
                                      <p>输出Token数量</p>
                                    </TooltipContent>
                                  </Tooltip>
                                </TooltipProvider>
                              </div>
                            </TableCell>
                            <TableCell className="text-right font-medium text-red-500 dark:text-red-400">
                              {item.amount}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  ) : (
                    <div className="text-center py-8">
                      <p className="text-muted-foreground">暂无使用记录</p>
                    </div>
                  )}
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}