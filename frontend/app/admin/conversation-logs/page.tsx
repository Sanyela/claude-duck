"use client"

import { useState, useEffect } from "react"
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
  VisibilityState,
  getSortedRowModel,
  SortingState,
  getFilteredRowModel,
  ColumnFiltersState,
} from "@tanstack/react-table"
import { ArrowUpDown, Eye, MessageCircle, Search, Download, BarChart3, Filter, Calendar, User, Bot, ChevronDown } from "lucide-react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useToast } from "@/components/ui/use-toast"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import { request } from "@/api/request"

// 数据类型定义
type ConversationLog = {
  id: number
  user_id: number
  username: string
  message_id: string
  model: string
  request_type: string
  input_tokens: number
  output_tokens: number
  total_tokens: number
  points_used: number
  duration: number
  status: string
  is_free_model: boolean
  created_at: string
  preview: string
}

type ConversationLogDetail = {
  id: number
  user_id: number
  username: string
  message_id: string
  request_id: string
  model: string
  request_type: string
  ip: string
  user_input: any
  system_prompt: string
  messages: any[]
  tools: any[]
  temperature?: number
  max_tokens?: number
  top_p?: number
  top_k?: number
  ai_response: any
  response_text: string
  stop_reason: string
  stop_sequence: string
  tokens: {
    input_tokens: number
    output_tokens: number
    cache_creation_input_tokens: number
    cache_read_input_tokens: number
    total_tokens: number
  }
  billing: {
    input_multiplier: number
    output_multiplier: number
    cache_multiplier: number
    points_used: number
  }
  performance: {
    duration: number
    service_tier: string
  }
  status: string
  error?: string
  is_free_model: boolean
  created_at: string
  updated_at: string
}

type ConversationStats = {
  basic_stats: {
    total_conversations: number
    total_users: number
    total_tokens: number
    total_points: number
    success_rate: number
    avg_duration: number
  }
  model_stats: Array<{
    model: string
    count: number
    tokens: number
  }>
  daily_trend: Array<{
    date: string
    conversations: number
    users: number
    tokens: number
  }>
  user_ranking: Array<{
    user_id: number
    username: string
    conversations: number
    tokens: number
  }>
  date_range: {
    from: string
    to: string
  }
  generated_at: string
}

// API调用函数
const conversationAPI = {
  getConversationLogs: async (params: any) => {
    const queryParams = new URLSearchParams()
    Object.keys(params).forEach(key => {
      if (params[key] !== undefined && params[key] !== '') {
        queryParams.append(key, params[key].toString())
      }
    })
    
    const response = await request.get(`/api/admin/conversation-logs?${queryParams}`)
    return response.data
  },
  
  getConversationLogDetail: async (id: number) => {
    const response = await request.get(`/api/admin/conversation-logs/${id}`)
    return response.data
  },
  
  getConversationStats: async (params: any) => {
    const queryParams = new URLSearchParams()
    Object.keys(params).forEach(key => {
      if (params[key] !== undefined && params[key] !== '') {
        queryParams.append(key, params[key].toString())
      }
    })
    
    const response = await request.get(`/api/admin/conversation-logs/stats?${queryParams}`)
    return response.data
  }
}

export default function ConversationLogsPage() {
  const [logs, setLogs] = useState<ConversationLog[]>([])
  const [stats, setStats] = useState<ConversationStats | null>(null)
  const [selectedLog, setSelectedLog] = useState<ConversationLogDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [totalLogs, setTotalLogs] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize] = useState(10)
  const [showDetailDialog, setShowDetailDialog] = useState(false)
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [rowSelection, setRowSelection] = useState({})
  
  // 搜索和过滤状态
  const [filters, setFilters] = useState({
    user_id: '',
    model: 'all',
    status: 'all',
    request_type: 'all',
    date_from: '',
    date_to: '',
    keyword: ''
  })

  const { toast } = useToast()

  // 表格列定义
  const columns: ColumnDef<ConversationLog>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="全选"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="选择行"
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "id",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          ID
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
    },
    {
      accessorKey: "username",
      header: "用户",
      cell: ({ row }) => (
        <div className="flex items-center space-x-2">
          <User className="h-4 w-4 text-gray-500" />
          <span>{row.getValue("username")}</span>
        </div>
      ),
    },
    {
      accessorKey: "model",
      header: "模型",
      cell: ({ row }) => (
        <div className="flex items-center space-x-2">
          <Bot className="h-4 w-4 text-blue-500" />
          <span className="font-mono text-sm">{row.getValue("model")}</span>
        </div>
      ),
    },
    {
      accessorKey: "request_type",
      header: "类型",
      cell: ({ row }) => {
        const type = row.getValue("request_type") as string
        return (
          <Badge variant={type === "stream" ? "default" : "secondary"}>
            {type === "stream" ? "流式" : "非流式"}
          </Badge>
        )
      },
    },
    {
      accessorKey: "total_tokens",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          Tokens
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => {
        const tokens = row.getValue("total_tokens") as number
        return <span className="font-mono">{tokens.toLocaleString()}</span>
      },
    },
    {
      accessorKey: "points_used",
      header: "积分",
      cell: ({ row }) => {
        const points = row.getValue("points_used") as number
        const isFree = row.getValue("is_free_model") as boolean
        return (
          <span className={`font-mono ${isFree ? "text-green-600" : ""}`}>
            {isFree ? "免费" : points.toLocaleString()}
          </span>
        )
      },
    },
    {
      accessorKey: "duration",
      header: "耗时",
      cell: ({ row }) => {
        const duration = row.getValue("duration") as number
        return <span className="font-mono">{duration}ms</span>
      },
    },
    {
      accessorKey: "status",
      header: "状态",
      cell: ({ row }) => {
        const status = row.getValue("status") as string
        return (
          <Badge
            variant={
              status === "success" ? "default" :
              status === "failed" ? "destructive" : "secondary"
            }
          >
            {status === "success" ? "成功" : 
             status === "failed" ? "失败" : status}
          </Badge>
        )
      },
    },
    {
      accessorKey: "created_at",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          时间
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => {
        const date = new Date(row.getValue("created_at"))
        return (
          <div className="text-sm">
            <div>{date.toLocaleDateString()}</div>
            <div className="text-gray-500">{date.toLocaleTimeString()}</div>
          </div>
        )
      },
    },
    {
      accessorKey: "preview",
      header: "预览",
      cell: ({ row }) => {
        const preview = row.getValue("preview") as string
        return (
          <div className="max-w-xs truncate text-sm text-gray-600">
            {preview}
          </div>
        )
      },
    },
    {
      id: "actions",
      header: "操作",
      cell: ({ row }) => {
        const log = row.original
        return (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleViewDetail(log.id)}
          >
            <Eye className="h-4 w-4" />
            <span className="ml-1">查看详情</span>
          </Button>
        )
      },
    },
  ]

  const table = useReactTable({
    data: logs,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  })

  // 获取对话日志
  const fetchLogs = async () => {
    try {
      setLoading(true)
      
      // 转换过滤器，将"all"转换为空字符串
      const apiFilters = {
        ...filters,
        model: filters.model === 'all' ? '' : filters.model,
        status: filters.status === 'all' ? '' : filters.status,
        request_type: filters.request_type === 'all' ? '' : filters.request_type,
      }
      
      const response = await conversationAPI.getConversationLogs({
        page: currentPage,
        page_size: pageSize,
        ...apiFilters
      })
      
      setLogs(response.data)
      setTotalLogs(response.total)
    } catch (error) {
      toast({
        title: "错误",
        description: "获取对话日志失败",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  // 获取统计数据
  const fetchStats = async () => {
    try {
      const response = await conversationAPI.getConversationStats({
        date_from: filters.date_from || undefined,
        date_to: filters.date_to || undefined
      })
      setStats(response)
    } catch (error) {
      toast({
        title: "错误",
        description: "获取统计数据失败",
        variant: "destructive",
      })
    }
  }

  // 查看详情
  const handleViewDetail = async (id: number) => {
    try {
      const detail = await conversationAPI.getConversationLogDetail(id)
      setSelectedLog(detail)
      setShowDetailDialog(true)
    } catch (error) {
      toast({
        title: "错误",
        description: "获取对话详情失败",
        variant: "destructive",
      })
    }
  }

  // 应用过滤器
  const handleApplyFilters = () => {
    setCurrentPage(1)
    fetchLogs()
  }

  // 重置过滤器
  const handleResetFilters = () => {
    setFilters({
      user_id: '',
      model: 'all',
      status: 'all',
      request_type: 'all',
      date_from: '',
      date_to: '',
      keyword: ''
    })
    setCurrentPage(1)
  }

  useEffect(() => {
    fetchLogs()
    fetchStats()
  }, [currentPage])

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* 页面标题 */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">对话日志管理</h1>
            <p className="text-gray-600">查看和分析所有Claude对话记录</p>
          </div>
          <div className="flex space-x-2">
            <Button variant="outline" onClick={() => fetchStats()}>
              <BarChart3 className="mr-2 h-4 w-4" />
              刷新统计
            </Button>
            <Button onClick={() => fetchLogs()}>
              <Search className="mr-2 h-4 w-4" />
              刷新数据
            </Button>
          </div>
        </div>

        <Tabs defaultValue="logs" className="space-y-4">
          <TabsList>
            <TabsTrigger value="logs">对话日志</TabsTrigger>
            <TabsTrigger value="stats">统计分析</TabsTrigger>
          </TabsList>

          <TabsContent value="logs" className="space-y-4">
            {/* 过滤器 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center">
                  <Filter className="mr-2 h-4 w-4" />
                  过滤条件
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                  <div>
                    <Label htmlFor="user_id">用户ID</Label>
                    <Input
                      id="user_id"
                      placeholder="输入用户ID"
                      value={filters.user_id}
                      onChange={(e) => setFilters(prev => ({ ...prev, user_id: e.target.value }))}
                    />
                  </div>
                  <div>
                    <Label htmlFor="model">模型</Label>
                    <Select
                      value={filters.model}
                      onValueChange={(value) => setFilters(prev => ({ ...prev, model: value === "all" ? "" : value }))}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="选择模型" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">全部模型</SelectItem>
                        <SelectItem value="claude-3-5-sonnet-20241022">Claude 3.5 Sonnet</SelectItem>
                        <SelectItem value="claude-3-5-haiku-20241022">Claude 3.5 Haiku</SelectItem>
                        <SelectItem value="claude-3-opus-20240229">Claude 3 Opus</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label htmlFor="status">状态</Label>
                    <Select
                      value={filters.status}
                      onValueChange={(value) => setFilters(prev => ({ ...prev, status: value === "all" ? "" : value }))}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="选择状态" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">全部状态</SelectItem>
                        <SelectItem value="success">成功</SelectItem>
                        <SelectItem value="failed">失败</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label htmlFor="request_type">请求类型</Label>
                    <Select
                      value={filters.request_type}
                      onValueChange={(value) => setFilters(prev => ({ ...prev, request_type: value === "all" ? "" : value }))}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="选择类型" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">全部类型</SelectItem>
                        <SelectItem value="api">非流式</SelectItem>
                        <SelectItem value="stream">流式</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label htmlFor="date_from">开始日期</Label>
                    <Input
                      id="date_from"
                      type="date"
                      value={filters.date_from}
                      onChange={(e) => setFilters(prev => ({ ...prev, date_from: e.target.value }))}
                    />
                  </div>
                  <div>
                    <Label htmlFor="date_to">结束日期</Label>
                    <Input
                      id="date_to"
                      type="date"
                      value={filters.date_to}
                      onChange={(e) => setFilters(prev => ({ ...prev, date_to: e.target.value }))}
                    />
                  </div>
                  <div>
                    <Label htmlFor="keyword">关键词搜索</Label>
                    <Input
                      id="keyword"
                      placeholder="搜索内容..."
                      value={filters.keyword}
                      onChange={(e) => setFilters(prev => ({ ...prev, keyword: e.target.value }))}
                    />
                  </div>
                  <div className="flex items-end space-x-2">
                    <Button onClick={handleApplyFilters}>
                      <Search className="mr-2 h-4 w-4" />
                      搜索
                    </Button>
                    <Button variant="outline" onClick={handleResetFilters}>
                      重置
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* 数据表格 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <div className="flex items-center">
                    <MessageCircle className="mr-2 h-4 w-4" />
                    对话记录 ({totalLogs})
                  </div>
                  <div className="flex items-center space-x-2">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="outline" size="sm">
                          列 <ChevronDown className="ml-2 h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        {table
                          .getAllColumns()
                          .filter((column) => column.getCanHide())
                          .map((column) => {
                            return (
                              <DropdownMenuCheckboxItem
                                key={column.id}
                                className="capitalize"
                                checked={column.getIsVisible()}
                                onCheckedChange={(value) =>
                                  column.toggleVisibility(!!value)
                                }
                              >
                                {column.id}
                              </DropdownMenuCheckboxItem>
                            )
                          })}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      {table.getHeaderGroups().map((headerGroup) => (
                        <TableRow key={headerGroup.id}>
                          {headerGroup.headers.map((header) => {
                            return (
                              <TableHead key={header.id}>
                                {header.isPlaceholder
                                  ? null
                                  : flexRender(
                                      header.column.columnDef.header,
                                      header.getContext()
                                    )}
                              </TableHead>
                            )
                          })}
                        </TableRow>
                      ))}
                    </TableHeader>
                    <TableBody>
                      {table.getRowModel().rows?.length ? (
                        table.getRowModel().rows.map((row) => (
                          <TableRow
                            key={row.id}
                            data-state={row.getIsSelected() && "selected"}
                          >
                            {row.getVisibleCells().map((cell) => (
                              <TableCell key={cell.id}>
                                {flexRender(
                                  cell.column.columnDef.cell,
                                  cell.getContext()
                                )}
                              </TableCell>
                            ))}
                          </TableRow>
                        ))
                      ) : (
                        <TableRow>
                          <TableCell
                            colSpan={columns.length}
                            className="h-24 text-center"
                          >
                            {loading ? "加载中..." : "暂无数据"}
                          </TableCell>
                        </TableRow>
                      )}
                    </TableBody>
                  </Table>
                </div>

                {/* 分页 */}
                <div className="flex items-center justify-end space-x-2 py-4">
                  <div className="flex-1 text-sm text-muted-foreground">
                    已选择 {table.getFilteredSelectedRowModel().rows.length} / {table.getFilteredRowModel().rows.length} 行
                  </div>
                  <div className="space-x-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                      disabled={currentPage <= 1}
                    >
                      上一页
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(prev => prev + 1)}
                      disabled={currentPage * pageSize >= totalLogs}
                    >
                      下一页
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="stats" className="space-y-4">
            {stats && (
              <>
                {/* 基础统计 */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <CardTitle className="text-sm font-medium">总对话数</CardTitle>
                      <MessageCircle className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{stats.basic_stats.total_conversations.toLocaleString()}</div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <CardTitle className="text-sm font-medium">活跃用户</CardTitle>
                      <User className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{stats.basic_stats.total_users.toLocaleString()}</div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <CardTitle className="text-sm font-medium">总Token数</CardTitle>
                      <Bot className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{stats.basic_stats.total_tokens.toLocaleString()}</div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <CardTitle className="text-sm font-medium">成功率</CardTitle>
                      <BarChart3 className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{stats.basic_stats.success_rate.toFixed(1)}%</div>
                    </CardContent>
                  </Card>
                </div>

                {/* 模型使用统计 */}
                <Card>
                  <CardHeader>
                    <CardTitle>模型使用统计</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      {stats.model_stats.map((model) => (
                        <div key={model.model} className="flex items-center justify-between p-2 border rounded">
                          <span className="font-mono">{model.model}</span>
                          <div className="text-right">
                            <div>{model.count.toLocaleString()} 次对话</div>
                            <div className="text-sm text-gray-500">{model.tokens.toLocaleString()} Tokens</div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>

                {/* 用户使用排行 */}
                <Card>
                  <CardHeader>
                    <CardTitle>用户使用排行 (Top 10)</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      {stats.user_ranking.map((user, index) => (
                        <div key={user.user_id} className="flex items-center justify-between p-2 border rounded">
                          <div className="flex items-center space-x-2">
                            <Badge variant="outline">#{index + 1}</Badge>
                            <span>{user.username}</span>
                          </div>
                          <div className="text-right">
                            <div>{user.conversations.toLocaleString()} 次对话</div>
                            <div className="text-sm text-gray-500">{user.tokens.toLocaleString()} Tokens</div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </>
            )}
          </TabsContent>
        </Tabs>

        {/* 详情对话框 */}
        <Dialog open={showDetailDialog} onOpenChange={setShowDetailDialog}>
          <DialogContent className="max-w-4xl max-h-[80vh] overflow-hidden">
            <DialogHeader>
              <DialogTitle>对话详情</DialogTitle>
              <DialogDescription>
                查看完整的对话输入输出内容
              </DialogDescription>
            </DialogHeader>
            
            {selectedLog && (
              <ScrollArea className="h-[60vh] pr-4">
                <div className="space-y-4">
                  {/* 基本信息 */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">基本信息</CardTitle>
                    </CardHeader>
                    <CardContent className="grid grid-cols-2 gap-4">
                      <div>
                        <Label>用户</Label>
                        <div className="font-mono">{selectedLog.username} (ID: {selectedLog.user_id})</div>
                      </div>
                      <div>
                        <Label>模型</Label>
                        <div className="font-mono">{selectedLog.model}</div>
                      </div>
                      <div>
                        <Label>请求类型</Label>
                        <Badge>{selectedLog.request_type === "stream" ? "流式" : "非流式"}</Badge>
                      </div>
                      <div>
                        <Label>状态</Label>
                        <Badge variant={selectedLog.status === "success" ? "default" : "destructive"}>
                          {selectedLog.status === "success" ? "成功" : "失败"}
                        </Badge>
                      </div>
                      <div>
                        <Label>消息ID</Label>
                        <div className="font-mono text-sm">{selectedLog.message_id}</div>
                      </div>
                      <div>
                        <Label>IP地址</Label>
                        <div className="font-mono">{selectedLog.ip}</div>
                      </div>
                    </CardContent>
                  </Card>

                  {/* Token和计费信息 */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">Token统计</CardTitle>
                    </CardHeader>
                    <CardContent className="grid grid-cols-3 gap-4">
                      <div>
                        <Label>输入Token</Label>
                        <div className="font-mono">{selectedLog.tokens.input_tokens.toLocaleString()}</div>
                      </div>
                      <div>
                        <Label>输出Token</Label>
                        <div className="font-mono">{selectedLog.tokens.output_tokens.toLocaleString()}</div>
                      </div>
                      <div>
                        <Label>总Token</Label>
                        <div className="font-mono">{selectedLog.tokens.total_tokens.toLocaleString()}</div>
                      </div>
                      <div>
                        <Label>缓存创建Token</Label>
                        <div className="font-mono">{selectedLog.tokens.cache_creation_input_tokens.toLocaleString()}</div>
                      </div>
                      <div>
                        <Label>缓存读取Token</Label>
                        <div className="font-mono">{selectedLog.tokens.cache_read_input_tokens.toLocaleString()}</div>
                      </div>
                      <div>
                        <Label>使用积分</Label>
                        <div className="font-mono">{selectedLog.billing.points_used.toLocaleString()}</div>
                      </div>
                    </CardContent>
                  </Card>

                  {/* 用户输入 */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">用户输入</CardTitle>
                    </CardHeader>
                    <CardContent>
                      {selectedLog.system_prompt && (
                        <div className="mb-4">
                          <Label>系统提示</Label>
                          <div className="bg-gray-50 p-3 rounded border">
                            <pre className="text-sm whitespace-pre-wrap">{selectedLog.system_prompt}</pre>
                          </div>
                        </div>
                      )}
                      <div>
                        <Label>消息历史</Label>
                        <div className="bg-gray-50 p-3 rounded border max-h-40 overflow-y-auto">
                          <pre className="text-sm whitespace-pre-wrap">
                            {JSON.stringify(selectedLog.messages, null, 2)}
                          </pre>
                        </div>
                      </div>
                    </CardContent>
                  </Card>

                  {/* AI响应 */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">AI响应</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="mb-4">
                        <Label>响应文本</Label>
                        <div className="bg-blue-50 p-3 rounded border max-h-40 overflow-y-auto">
                          <pre className="text-sm whitespace-pre-wrap">{selectedLog.response_text}</pre>
                        </div>
                      </div>
                      <div>
                        <Label>完整响应</Label>
                        <div className="bg-gray-50 p-3 rounded border max-h-40 overflow-y-auto">
                          <pre className="text-sm whitespace-pre-wrap">
                            {JSON.stringify(selectedLog.ai_response, null, 2)}
                          </pre>
                        </div>
                      </div>
                    </CardContent>
                  </Card>

                  {/* 性能信息 */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">性能信息</CardTitle>
                    </CardHeader>
                    <CardContent className="grid grid-cols-2 gap-4">
                      <div>
                        <Label>响应时间</Label>
                        <div className="font-mono">{selectedLog.performance.duration}ms</div>
                      </div>
                      <div>
                        <Label>服务层级</Label>
                        <div className="font-mono">{selectedLog.performance.service_tier}</div>
                      </div>
                      <div>
                        <Label>停止原因</Label>
                        <div className="font-mono">{selectedLog.stop_reason}</div>
                      </div>
                      <div>
                        <Label>创建时间</Label>
                        <div className="font-mono">{new Date(selectedLog.created_at).toLocaleString()}</div>
                      </div>
                    </CardContent>
                  </Card>

                  {selectedLog.error && (
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-lg text-red-600">错误信息</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="bg-red-50 p-3 rounded border text-red-700">
                          <pre className="text-sm whitespace-pre-wrap">{selectedLog.error}</pre>
                        </div>
                      </CardContent>
                    </Card>
                  )}
                </div>
              </ScrollArea>
            )}
          </DialogContent>
        </Dialog>
      </div>
    </DashboardLayout>
  )
}