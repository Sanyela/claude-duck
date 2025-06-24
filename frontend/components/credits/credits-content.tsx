"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Separator } from "@/components/ui/separator"
import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Calendar as CalendarIcon, Search, RotateCcw, Loader2, AlertCircle, ChevronLeft, ChevronRight } from "lucide-react"
import { Calendar } from "@/components/ui/calendar"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
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
import { format } from "date-fns"
import { cn } from "@/lib/utils"

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
  
  // 日期状态使用Date对象
  const [startDate, setStartDate] = useState<Date>(() => {
    const date = new Date()
    date.setDate(date.getDate() - 7) // 默认显示最近7天
    return date
  })
  const [endDate, setEndDate] = useState<Date>(() => {
    return new Date()
  })
  
  // 分页状态
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [pageSize, setPageSize] = useState(10) // 每页显示数量，改为可变状态
  
  const [error, setError] = useState<string | null>(null)

  // 页面大小选项
  const pageSizeOptions = [10, 20, 40, 100]

  // 加载数据
  const loadData = async (startDateParam?: Date, endDateParam?: Date, page: number = 1) => {
    setLoading(true)
    setError(null)

    try {
      const [balanceResult, historyResult] = await Promise.all([
        creditsAPI.getBalance(),
        creditsAPI.getUsageHistory({
          start_date: startDateParam ? format(startDateParam, 'yyyy-MM-dd') : format(startDate, 'yyyy-MM-dd'),
          end_date: endDateParam ? format(endDateParam, 'yyyy-MM-dd') : format(endDate, 'yyyy-MM-dd'),
          page: page,
          page_size: pageSize
        })
      ])

      if (balanceResult.success && balanceResult.data) {
        setCreditBalance(balanceResult.data)
      }

      if (historyResult.success && historyResult.data) {
        // 提取历史记录数组
        const historyArray = historyResult.data.history || [];
        
        setUsageHistory(historyArray)
        setCurrentPage(historyResult.data.currentPage || 1)
        setTotalPages(historyResult.data.totalPages || 1)
        
        // 生成图表数据 - 获取更多数据用于图表显示
        const chartHistoryResult = await creditsAPI.getUsageHistory({
          start_date: startDateParam ? format(startDateParam, 'yyyy-MM-dd') : format(startDate, 'yyyy-MM-dd'),
          end_date: endDateParam ? format(endDateParam, 'yyyy-MM-dd') : format(endDate, 'yyyy-MM-dd'),
          page: 1,
          page_size: 100 // 获取更多数据用于图表
        })
        
        if (chartHistoryResult.success && chartHistoryResult.data) {
          const chartPoints = generateChartData(chartHistoryResult.data.history || [], balanceResult.data || null)
          setChartData(chartPoints)
        }
      }
    } catch (err: any) {
      setError("加载积分数据失败")
    }

    setLoading(false)
  }

  // 生成图表数据
  const generateChartData = (history: CreditUsageHistory[], balance: CreditBalance | null): ChartDataPoint[] => {
    if (!balance || !Array.isArray(history) || !history.length) return []

    // 按时间排序，从最新到最旧
    const sortedHistory = [...history].sort((a, b) => 
      new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
    )

    // 生成图表点 - 按时间序列
    const points: ChartDataPoint[] = []
    let currentBalance = balance.available_points

    // 添加当前时间点（最新余额）
    points.push({
      date: new Date().toISOString(),
      balance: currentBalance,
      id: 'current'
    })

    // 按时间倒推，计算每个时间点的余额
    sortedHistory.forEach((record, index) => {
      // 加回这次消耗的积分（因为我们在倒推）
      currentBalance += Math.abs(record.amount)
      
      points.push({
        date: record.timestamp,
        balance: currentBalance,
        id: record.id
      })
    })

    // 反转数组，让时间从早到晚排列
    return points.reverse()
  }
  
  // 获取日期范围内的所有日期
  const getDateRange = (start: Date, end: Date): Date[] => {
    const dates: Date[] = []
    const current = new Date(start)
    
    while (current <= end) {
      dates.push(new Date(current))
      current.setDate(current.getDate() + 1)
    }
    
    return dates
  }

  // 查询数据
  const handleSearch = () => {
    setCurrentPage(1) // 重置到第一页
    loadData(startDate, endDate, 1)
  }

  // 重置查询
  const handleReset = () => {
    const defaultEndDate = new Date()
    const defaultStartDate = new Date()
    defaultStartDate.setDate(defaultStartDate.getDate() - 7)
    
    setStartDate(defaultStartDate)
    setEndDate(defaultEndDate)
    setCurrentPage(1)
    loadData(defaultStartDate, defaultEndDate, 1)
  }

  // 处理分页
  const handlePageChange = (page: number) => {
    setCurrentPage(page)
    loadData(startDate, endDate, page)
  }

  // 处理记录点击事件
  const handleRecordClick = (recordId: string) => {
    setSelectedRecord(recordId === selectedRecord ? null : recordId)
  }

  // 处理页面大小变化
  const handlePageSizeChange = (newPageSize: string) => {
    const size = parseInt(newPageSize)
    setPageSize(size)
    setCurrentPage(1) // 重置到第一页
    // 使用新的页面大小重新加载数据
    loadDataWithNewPageSize(startDate, endDate, 1, size)
  }

  // 使用新页面大小加载数据的辅助函数
  const loadDataWithNewPageSize = async (startDateParam: Date, endDateParam: Date, page: number, newPageSize: number) => {
    setLoading(true)
    setError(null)

    try {
      const [balanceResult, historyResult] = await Promise.all([
        creditsAPI.getBalance(),
        creditsAPI.getUsageHistory({
          start_date: format(startDateParam, 'yyyy-MM-dd'),
          end_date: format(endDateParam, 'yyyy-MM-dd'),
          page: page,
          page_size: newPageSize
        })
      ])

      if (balanceResult.success && balanceResult.data) {
        setCreditBalance(balanceResult.data)
      }

      if (historyResult.success && historyResult.data) {
        const historyArray = historyResult.data.history || [];
        
        setUsageHistory(historyArray)
        setCurrentPage(historyResult.data.currentPage || 1)
        setTotalPages(historyResult.data.totalPages || 1)
        
        // 生成图表数据
        const chartHistoryResult = await creditsAPI.getUsageHistory({
          start_date: format(startDateParam, 'yyyy-MM-dd'),
          end_date: format(endDateParam, 'yyyy-MM-dd'),
          page: 1,
          page_size: 100
        })
        
        if (chartHistoryResult.success && chartHistoryResult.data) {
          const chartPoints = generateChartData(chartHistoryResult.data.history || [], balanceResult.data || null)
          setChartData(chartPoints)
        }
      }
    } catch (err: any) {
      setError("加载积分数据失败")
    }

    setLoading(false)
  }

  useEffect(() => {
    loadData()
  }, [])

  // 计算统计数据
  const totalUsage = Array.isArray(usageHistory) ? usageHistory.reduce((sum, item) => sum + Math.abs(item.amount), 0) : 0
  const uniqueModels = Array.isArray(usageHistory) ? new Set(usageHistory.map(item => item.relatedModel)).size : 0

  // 计算Y轴的动态范围
  const getYAxisDomain = (data: ChartDataPoint[]) => {
    if (data.length === 0) return [0, 100]
    
    const values = data.map(item => item.balance)
    const minValue = Math.min(...values)
    const maxValue = Math.max(...values)
    
    // 如果最大值和最小值相同（所有点都在同一水平线）
    if (minValue === maxValue) {
      return [Math.max(0, minValue - 10), maxValue + 10]
    }
    
    // 计算范围并添加适当的padding
    const range = maxValue - minValue
    const padding = Math.max(range * 0.1, 5) // 至少5个单位的padding
    
    return [
      Math.max(0, minValue - padding), // 确保不小于0
      maxValue + padding
    ]
  }

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
                    <Popover>
                      <PopoverTrigger asChild>
                        <Button
                          variant={"outline"}
                          className={cn(
                            "w-full justify-start text-left font-normal",
                            !startDate && "text-muted-foreground"
                          )}
                        >
                          <CalendarIcon className="mr-2 h-4 w-4" />
                          {startDate ? format(startDate, "yyyy-MM-dd") : <span>选择开始日期</span>}
                        </Button>
                      </PopoverTrigger>
                      <PopoverContent className="w-auto p-0" align="start">
                        <Calendar
                          mode="single"
                          selected={startDate}
                          onSelect={(date) => date && setStartDate(date)}
                          initialFocus
                        />
                      </PopoverContent>
                    </Popover>
                  </div>
                </div>
                <div className="flex items-center gap-2 flex-1">
                  <div className="text-sm font-medium">结束时间:</div>
                  <div className="relative flex-1">
                    <Popover>
                      <PopoverTrigger asChild>
                        <Button
                          variant={"outline"}
                          className={cn(
                            "w-full justify-start text-left font-normal",
                            !endDate && "text-muted-foreground"
                          )}
                        >
                          <CalendarIcon className="mr-2 h-4 w-4" />
                          {endDate ? format(endDate, "yyyy-MM-dd") : <span>选择结束日期</span>}
                        </Button>
                      </PopoverTrigger>
                      <PopoverContent className="w-auto p-0" align="start">
                        <Calendar
                          mode="single"
                          selected={endDate}
                          onSelect={(date) => date && setEndDate(date)}
                          initialFocus
                        />
                      </PopoverContent>
                    </Popover>
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
                          tickFormatter={(value) => {
                            const date = new Date(value)
                            return format(date, 'MM-dd HH:mm')
                          }}
                        />
                        <YAxis
                          stroke="#888"
                          fontSize={12}
                          tickLine={false}
                          axisLine={false}
                          tickFormatter={(value) => `${value}`}
                          domain={getYAxisDomain(chartData)}
                          type="number"
                        />
                        <RechartsTooltip 
                          formatter={(value: any) => [`${value} 积分`, '余额']}
                          labelFormatter={(label: any) => {
                            const date = new Date(label)
                            return `时间: ${format(date, 'yyyy-MM-dd HH:mm:ss')}`
                          }}
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
                    <>
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
                                <TooltipProvider delayDuration={0}>
                                  <Tooltip>
                                    <TooltipTrigger asChild>
                                      <Badge variant="outline" className="font-mono text-xs">
                                        {item.relatedModel}
                                      </Badge>
                                    </TooltipTrigger>
                                    <TooltipContent>
                                      <p>{item.relatedModel}</p>
                                    </TooltipContent>
                                  </Tooltip>
                                </TooltipProvider>
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
                      
                      {/* 分页和页面大小选择组件 */}
                      {(totalPages > 1 || usageHistory.length > 0) && (
                        <div className="mt-4 flex items-center justify-between flex-wrap gap-4">
                          <div className="flex items-center gap-4">
                            <div className="text-sm text-muted-foreground">
                              第 {currentPage} 页，共 {totalPages} 页
                            </div>
                            <div className="flex items-center gap-2">
                              <span className="text-sm text-muted-foreground">每页显示:</span>
                              <Select value={pageSize.toString()} onValueChange={handlePageSizeChange}>
                                <SelectTrigger className="w-20">
                                  <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                  {pageSizeOptions.map((size) => (
                                    <SelectItem key={size} value={size.toString()}>
                                      {size}
                                    </SelectItem>
                                  ))}
                                </SelectContent>
                              </Select>
                              <span className="text-sm text-muted-foreground">条</span>
                            </div>
                          </div>
                          {totalPages > 1 && (
                            <div className="flex items-center gap-2">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handlePageChange(currentPage - 1)}
                                disabled={currentPage <= 1}
                              >
                                <ChevronLeft className="h-4 w-4" />
                                上一页
                              </Button>
                              
                              {/* 页码按钮 */}
                              <div className="flex gap-1">
                                {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
                                  let pageNum
                                  if (totalPages <= 5) {
                                    pageNum = i + 1
                                  } else if (currentPage <= 3) {
                                    pageNum = i + 1
                                  } else if (currentPage >= totalPages - 2) {
                                    pageNum = totalPages - 4 + i
                                  } else {
                                    pageNum = currentPage - 2 + i
                                  }
                                  
                                  return (
                                    <Button
                                      key={pageNum}
                                      variant={currentPage === pageNum ? "default" : "outline"}
                                      size="sm"
                                      onClick={() => handlePageChange(pageNum)}
                                      className="w-8 h-8 p-0"
                                    >
                                      {pageNum}
                                    </Button>
                                  )
                                })}
                              </div>
                              
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handlePageChange(currentPage + 1)}
                                disabled={currentPage >= totalPages}
                              >
                                下一页
                                <ChevronRight className="h-4 w-4" />
                              </Button>
                            </div>
                          )}
                        </div>
                      )}
                    </>
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