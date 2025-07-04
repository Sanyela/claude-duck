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
import { creditsAPI, type CreditBalance, type CreditUsageHistory } from "@/api/credits"
import { PricingTableModal } from "./pricing-table-modal"
import { format } from "date-fns"
import { cn } from "@/lib/utils"
import React from "react"


export function CreditsContent() {
  const [loading, setLoading] = useState(true)
  const [creditBalance, setCreditBalance] = useState<CreditBalance | null>(null)
  const [usageHistory, setUsageHistory] = useState<CreditUsageHistory[]>([])
  const [selectedRecord, setSelectedRecord] = useState<string | null>(null)
  
  // 计费表弹窗状态
  const [pricingTableOpen, setPricingTableOpen] = useState(false)
  const [selectedTokenCount, setSelectedTokenCount] = useState<number | undefined>(undefined)
  
  // 累计token计费配置
  const [tokenThreshold, setTokenThreshold] = useState<number>(5000)
  const [pointsPerThreshold, setPointsPerThreshold] = useState<number>(1)
  
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
      const [balanceResult, historyResult, configResult] = await Promise.all([
        creditsAPI.getBalance(),
        creditsAPI.getUsageHistory({
          start_date: startDateParam ? format(startDateParam, 'yyyy-MM-dd') : format(startDate, 'yyyy-MM-dd'),
          end_date: endDateParam ? format(endDateParam, 'yyyy-MM-dd') : format(endDate, 'yyyy-MM-dd'),
          page: page,
          page_size: pageSize
        }),
        creditsAPI.getPricingTable()
      ])

      if (balanceResult.success && balanceResult.data) {
        setCreditBalance(balanceResult.data)
      }

      if (configResult.success && configResult.data) {
        setTokenThreshold(configResult.data.token_threshold)
        setPointsPerThreshold(configResult.data.points_per_threshold)
      }

      if (historyResult.success && historyResult.data) {
        // 提取历史记录数组
        const historyArray = historyResult.data.history || [];
        
        setUsageHistory(historyArray)
        setCurrentPage(historyResult.data.currentPage || 1)
        setTotalPages(historyResult.data.totalPages || 1)
      }
    } catch (err: any) {
      setError("加载积分数据失败")
    }

    setLoading(false)
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
      }
    } catch (err: any) {
      setError("加载积分数据失败")
    }

    setLoading(false)
  }

  useEffect(() => {
    loadData()
  }, [])

  // 计算统计数据 - 在累计token模式下，统计信息基于实际使用情况而非进度积分
  const totalUsage = Array.isArray(usageHistory) ? usageHistory.reduce((sum, item) => sum + Math.round(Math.abs(item.amount) * 100) / 100, 0) : 0
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
                  <span className="font-semibold">{(creditBalance?.used_points || 0).toLocaleString()}</span> 
                  <span className="ml-1 text-xs opacity-90">已消耗</span>
                </Badge>
                <Badge className="bg-sky-500 text-white hover:bg-sky-600 px-2.5 py-1.5 shadow-sm">
                  <span className="font-semibold">{creditBalance?.available_points?.toLocaleString() || 0}</span> 
                  <span className="ml-1 text-xs opacity-90">可用积分</span>
                </Badge>
                <Badge className="bg-green-500 text-white hover:bg-green-600 px-2.5 py-1.5 shadow-sm">
                  <span className="font-semibold">{uniqueModels}</span> 
                  <span className="ml-1 text-xs opacity-90">使用模型</span>
                </Badge>
                <Badge className="bg-purple-500 text-white hover:bg-purple-600 px-2.5 py-1.5 shadow-sm">
                  <span className="font-semibold">{creditBalance?.free_model_usage_count?.toLocaleString() || 0}</span> 
                  <span className="ml-1 text-xs opacity-90">免费调用</span>
                </Badge>
                <Badge className="bg-orange-500 text-white hover:bg-orange-600 px-2.5 py-1.5 shadow-sm">
                  <span className="font-semibold">{creditBalance?.accumulated_tokens?.toLocaleString() || 0}</span> 
                  <span className="ml-1 text-xs opacity-90">累计Token</span>
                </Badge>
                {(creditBalance?.expired_points || 0) > 0 && (
                  <Badge className="bg-orange-500 text-white hover:bg-orange-600 px-2.5 py-1.5 shadow-sm">
                    <span className="font-semibold">{(creditBalance?.expired_points || 0).toLocaleString()}</span> 
                    <span className="ml-1 text-xs opacity-90">已过期</span>
                  </Badge>
                )}
              </div>
              
              <div className="text-xs text-muted-foreground mb-4 px-1">
                仅显示当前有效订阅的积分数据
              </div>
              
              {/* 累计Token进度条 */}
              {creditBalance && (
                <div className="bg-gradient-to-r from-orange-50 to-orange-100 dark:from-orange-900/20 dark:to-orange-800/20 rounded-lg p-4 mb-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium text-orange-800 dark:text-orange-200">累计Token进度</span>
                    <span className="text-xs text-orange-600 dark:text-orange-300">
                      距离下次扣费还需 {Math.max(0, tokenThreshold - (creditBalance.accumulated_tokens % tokenThreshold)).toLocaleString()} Token
                    </span>
                  </div>
                  <div className="w-full bg-orange-200 dark:bg-orange-800 rounded-full h-2 mb-2">
                    <div 
                      className="bg-orange-500 h-2 rounded-full transition-all duration-300"
                      style={{ 
                        width: `${Math.min(100, ((creditBalance.accumulated_tokens % tokenThreshold) / tokenThreshold) * 100)}%` 
                      }}
                    ></div>
                  </div>
                  <div className="flex items-center justify-between text-xs text-orange-700 dark:text-orange-300">
                    <span>{(creditBalance.accumulated_tokens % tokenThreshold).toLocaleString()} / {tokenThreshold.toLocaleString()} Token</span>
                    <span>{(((creditBalance.accumulated_tokens % tokenThreshold) / tokenThreshold) * 100).toFixed(1)}%</span>
                  </div>
                </div>
              )}
              
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
              
              <div className="space-y-6">
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
                            <React.Fragment key={item.id}>
                              <TableRow 
                                className={`border-border cursor-pointer transition-colors ${selectedRecord === item.id ? 'bg-sky-50 dark:bg-sky-900/20' : 'hover:bg-muted/50'}`}
                                onClick={() => handleRecordClick(item.id)}
                              >
                                <TableCell className="font-medium">
                                  {new Date(item.timestamp).toLocaleString()}
                                </TableCell>
                                <TableCell>
                                  <div className="flex items-center">
                                    <Badge variant="outline" className="mr-2">
                                      {item.relatedModel || "Unknown"}
                                    </Badge>
                                  </div>
                                </TableCell>
                                <TableCell>
                                  <div className="flex flex-wrap items-center gap-1">
                                    <TooltipProvider delayDuration={0}>
                                      <Tooltip>
                                        <TooltipTrigger asChild>
                                          <Badge className="bg-green-100 text-green-800 hover:bg-green-200 dark:bg-green-900/30 dark:text-green-300 border-green-300">
                                            {item.input_tokens.toLocaleString()}
                                          </Badge>
                                        </TooltipTrigger>
                                        <TooltipContent>
                                          <p>输入Token</p>
                                        </TooltipContent>
                                      </Tooltip>
                                    </TooltipProvider>
                                    
                                    <TooltipProvider delayDuration={0}>
                                      <Tooltip>
                                        <TooltipTrigger asChild>
                                          <Badge className="bg-red-100 text-red-800 hover:bg-red-200 dark:bg-red-900/30 dark:text-red-300 border-red-300">
                                            {item.output_tokens.toLocaleString()}
                                          </Badge>
                                        </TooltipTrigger>
                                        <TooltipContent>
                                          <p>输出Token</p>
                                        </TooltipContent>
                                      </Tooltip>
                                    </TooltipProvider>
                                    
                                    {(item.total_cache_tokens || 0) > 0 && (
                                      <TooltipProvider delayDuration={0}>
                                        <Tooltip>
                                          <TooltipTrigger asChild>
                                            <Badge className="bg-blue-100 text-blue-800 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-300 border-blue-300">
                                              {(item.total_cache_tokens || 0).toLocaleString()}
                                            </Badge>
                                          </TooltipTrigger>
                                          <TooltipContent>
                                            <p>缓存Token</p>
                                          </TooltipContent>
                                        </Tooltip>
                                      </TooltipProvider>
                                    )}
                                  </div>
                                </TableCell>
                                <TableCell className="text-right">
                                  <div className="flex items-center justify-end">
                                    {Math.abs(item.amount) < 1 ? (
                                      <div className="text-right">
                                        <div className="font-medium text-orange-600">{Math.round(Math.abs(item.amount) * 100) / 100}</div>
                                        <div className="text-xs text-muted-foreground">进度积分</div>
                                      </div>
                                    ) : (
                                      <div className="text-right">
                                        <div className="font-medium text-red-600">{Math.round(Math.abs(item.amount) * 100) / 100}</div>
                                        <div className="text-xs text-muted-foreground">积分</div>
                                      </div>
                                    )}
                                  </div>
                                </TableCell>
                              </TableRow>
                              
                              {/* 详细信息展开区域 */}
                              {selectedRecord === item.id && item.billing_details && (
                                <TableRow className="bg-sky-50/50 dark:bg-sky-900/10">
                                  <TableCell colSpan={4} className="py-4">
                                    <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4">
                                      <h5 className="text-sm font-medium mb-3 text-blue-900 dark:text-blue-100">💰 累计Token计费详情</h5>
                                      <div className="space-y-3 text-sm">
                                        <div className="font-mono bg-white dark:bg-gray-800 p-3 rounded border">
                                          <div className="text-gray-600 dark:text-gray-300 mb-2">1. 加权Token计算:</div>
                                          <div className="mb-2">
                                            <span className="underline decoration-green-500 decoration-2">{item.input_tokens.toLocaleString()}(输入) × {item.billing_details.input_multiplier}(输入倍率)</span> + {(item.total_cache_tokens || 0) > 0 && <span><span className="underline decoration-blue-500 decoration-2">{(item.total_cache_tokens || 0).toLocaleString()}(缓存) × {item.billing_details.cache_multiplier}(缓存倍率)</span> + </span>}<span className="underline decoration-red-500 decoration-2">{item.output_tokens.toLocaleString()}(输出) × {item.billing_details.output_multiplier}(输出倍率)</span> = <span className="font-bold text-blue-600">{item.billing_details.total_weighted_tokens.toLocaleString()}(加权Token)</span>
                                          </div>
                                        </div>
                                        
                                        <div className="font-mono bg-white dark:bg-gray-800 p-3 rounded border">
                                          <div className="text-gray-600 dark:text-gray-300 mb-2">2. 累计计费机制:</div>
                                          <div className="space-y-1">
                                            <div>本次进度积分: <span className="font-bold text-orange-600">{Math.round(Math.abs(item.amount) * 100) / 100}</span></div>
                                            <div className="text-xs text-gray-500">
                                              {Math.abs(item.amount) < 1 ? 
                                                "此次调用未触发扣费，Token已累计到您的账户" : 
                                                "此次调用触发了积分扣费"
                                              }
                                            </div>
                                            <div className="mt-2">
                                              <button
                                                className="text-blue-600 hover:text-blue-800 underline cursor-pointer"
                                                onClick={(e) => {
                                                  e.stopPropagation()
                                                  setSelectedTokenCount(Math.round(item.billing_details?.total_weighted_tokens || 0))
                                                  setPricingTableOpen(true)
                                                }}
                                              >
                                                查看累计计费配置 →
                                              </button>
                                            </div>
                                          </div>
                                        </div>
                                        
                                        <div className="bg-yellow-50 dark:bg-yellow-900/20 p-3 rounded-lg">
                                          <div className="text-yellow-800 dark:text-yellow-200 text-xs">
                                            <strong>💡 累计计费说明:</strong> 系统会累计您的加权Token使用量，只有当累计数量达到设定阈值时才会扣除积分。这样避免了小额Token也扣费的问题，让计费更加合理。
                                          </div>
                                        </div>
                                      </div>
                                    </div>
                                  </TableCell>
                                </TableRow>
                              )}
                            </React.Fragment>
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
      
      {/* 计费表悬浮窗 */}
      <PricingTableModal 
        open={pricingTableOpen}
        onOpenChange={setPricingTableOpen}
        tokenCount={selectedTokenCount}
      />
    </div>
  )
}