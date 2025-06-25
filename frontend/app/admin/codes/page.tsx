"use client"

import { useState, useEffect } from "react"
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
  VisibilityState,
} from "@tanstack/react-table"
import { ArrowUpDown, ChevronDown, MoreHorizontal, Key, Plus, Download, Search } from "lucide-react"
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
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useToast } from "@/components/ui/use-toast"
import { adminAPI, type ActivationCode, type SubscriptionPlan } from "@/api/admin"

// 数据类型定义
type ActivationCodeRow = {
  id: number
  code: string
  status: "unused" | "used" | "expired"
  plan_title: string
  used_by_username: string | null
  batch_number: string
  created_at: string
  used_at: string | null
}

export default function AdminCodesPage() {
  const { toast } = useToast()
  const [data, setData] = useState<ActivationCodeRow[]>([])
  const [plans, setPlans] = useState<SubscriptionPlan[]>([])
  const [loading, setLoading] = useState(false)
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [rowSelection, setRowSelection] = useState({})
  
  // 服务端分页和搜索状态
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 10,
    total: 0,
    totalPages: 0
  })
  const [searchParams, setSearchParams] = useState({
    batch_number: "",
    status: "all" as "all" | "unused" | "used" | "expired"
  })
  
  // 对话框状态
  const [isCodeDialogOpen, setIsCodeDialogOpen] = useState(false)
  const [newCodeData, setNewCodeData] = useState({
    subscription_plan_id: 0,
    count: 1,
    batch_number: ""
  })

  // 列定义
  const columns: ColumnDef<ActivationCodeRow>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllPageRowsSelected() ||
            (table.getIsSomePageRowsSelected() && "indeterminate")
          }
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
      accessorKey: "code",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            激活码
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => (
        <div className="font-mono text-sm">{row.getValue("code")}</div>
      ),
    },
    {
      accessorKey: "status",
      header: "状态",
      cell: ({ row }) => {
        const status = row.getValue("status") as string
        return (
          <Badge 
            variant={
              status === "used" ? "default" : 
              status === "expired" ? "destructive" : "secondary"
            }
          >
            {status === "used" ? "已使用" : 
             status === "expired" ? "已过期" : "未使用"}
          </Badge>
        )
      },
    },
    {
      accessorKey: "plan_title",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            关联计划
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => <div>{row.getValue("plan_title")}</div>,
    },
    {
      accessorKey: "used_by_username",
      header: "使用者",
      cell: ({ row }) => {
        const username = row.getValue("used_by_username") as string | null
        return <div>{username || "-"}</div>
      },
    },
    {
      accessorKey: "batch_number",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            批次号
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => <div>{row.getValue("batch_number")}</div>,
    },
    {
      accessorKey: "created_at",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            创建时间
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => {
        const date = new Date(row.getValue("created_at"))
        return <div>{date.toLocaleDateString()}</div>
      },
    },
    {
      id: "actions",
      enableHiding: false,
      cell: ({ row }) => {
        const code = row.original
        
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">打开菜单</span>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>操作</DropdownMenuLabel>
              <DropdownMenuItem
                onClick={() => navigator.clipboard.writeText(code.code)}
              >
                复制激活码
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={() => handleDeleteCode(code.id)}
                disabled={code.status === "used"}
                className="text-red-600"
              >
                删除激活码
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    manualPagination: true,
    pageCount: pagination.totalPages,
    state: {
      columnVisibility,
      rowSelection,
    },
  })

  // 加载数据
  const loadCodes = async () => {
    setLoading(true)
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...(searchParams.batch_number && { batch_number: searchParams.batch_number }),
      ...(searchParams.status !== "all" && { status: searchParams.status }),
    }
    
    const result = await adminAPI.getActivationCodes(params)
    if (result.success && result.codes) {
      const codes = Array.isArray(result.codes) ? result.codes : []
      const transformedData: ActivationCodeRow[] = codes.map((code: ActivationCode) => ({
        id: code.id,
        code: code.code,
        status: code.status as "unused" | "used" | "expired",
        plan_title: code.plan?.title || "未知计划",
        used_by_username: code.used_by?.username || null,
        batch_number: code.batch_number || "",
        created_at: code.created_at,
        used_at: code.used_at || null,
      }))
      setData(transformedData)
      
      // 更新分页信息
      setPagination(prev => ({
        ...prev,
        total: result.total || 0,
        totalPages: result.total_pages || 0,
        page: result.page || 1,
        pageSize: result.page_size || prev.pageSize
      }))
    } else {
      setData([])
      toast({
        title: "加载失败",
        description: result.message,
        variant: "destructive"
      })
    }
    setLoading(false)
  }

  const loadPlans = async () => {
    const result = await adminAPI.getSubscriptionPlans()
    if (result.success && result.plans) {
      setPlans(Array.isArray(result.plans) ? result.plans : [])
    }
  }

  useEffect(() => {
    loadPlans()
  }, [])

  useEffect(() => {
    loadCodes()
  }, [pagination.page, pagination.pageSize])

  // 监听搜索参数变化
  useEffect(() => {
    // 重置到第一页并重新搜索
    if (pagination.page === 1) {
      loadCodes()
    } else {
      setPagination(prev => ({ ...prev, page: 1 }))
    }
  }, [searchParams.batch_number, searchParams.status])

  // 搜索处理
  const handleSearch = () => {
    setPagination(prev => ({ ...prev, page: 1 }))
    loadCodes()
  }

  // 重置搜索
  const handleResetSearch = () => {
    setSearchParams({ batch_number: "", status: "all" })
  }

  // 处理创建激活码
  const handleCreateCodes = async () => {
    if (newCodeData.subscription_plan_id === 0) {
      toast({
        title: "请选择订阅计划",
        variant: "destructive"
      })
      return
    }

    const result = await adminAPI.createActivationCodes(newCodeData)
    if (result.success) {
      toast({ title: "创建成功", variant: "default" })
      setIsCodeDialogOpen(false)
      setNewCodeData({
        subscription_plan_id: 0,
        count: 1,
        batch_number: ""
      })
      loadCodes()
    } else {
      toast({
        title: "创建失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  // 处理删除激活码
  const handleDeleteCode = async (codeId: number) => {
    if (!confirm("确定要删除此激活码吗？")) return
    
    const result = await adminAPI.deleteActivationCode(codeId)
    if (result.success) {
      toast({ title: "删除成功", variant: "default" })
      loadCodes()
    } else {
      toast({
        title: "删除失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  // 导出CSV
  const exportToCSV = () => {
    const selectedRows = table.getSelectedRowModel().rows
    const dataToExport = selectedRows.length > 0 
      ? selectedRows.map(row => row.original)
      : data

    const csvContent = [
      // CSV 头部
      ['激活码', '状态', '关联计划', '使用者', '批次号', '创建时间', '使用时间'].join(','),
      // CSV 数据
      ...dataToExport.map(row => [
        `"${row.code}"`,
        row.status === "used" ? "已使用" : row.status === "expired" ? "已过期" : "未使用",
        `"${row.plan_title}"`,
        `"${row.used_by_username || "-"}"`,
        `"${row.batch_number}"`,
        `"${new Date(row.created_at).toLocaleString()}"`,
        `"${row.used_at ? new Date(row.used_at).toLocaleString() : "-"}"`
      ].join(','))
    ].join('\n')

    const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    const fileName = searchParams.batch_number 
      ? `activation_codes_${searchParams.batch_number}_${new Date().toISOString().split('T')[0]}.csv`
      : `activation_codes_${new Date().toISOString().split('T')[0]}.csv`
    link.download = fileName
    link.click()
    
    toast({
      title: "导出成功",
      description: `已导出 ${dataToExport.length} 条记录${selectedRows.length > 0 ? ' (仅选中项)' : ''}`,
    })
  }

  return (
    <DashboardLayout>
      <Card className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
        <CardHeader>
          <div className="flex items-center space-x-2">
            <Key className="h-6 w-6 text-orange-500" />
            <CardTitle className="text-slate-900 dark:text-slate-100">激活码管理</CardTitle>
          </div>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            生成和管理订阅计划的激活码，支持搜索、筛选和导出
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="w-full">
            {/* 顶部控制栏 */}
            <div className="flex items-center justify-between py-4">
              <div className="flex items-center space-x-2">
                <div className="relative">
                  <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="输入批次号搜索..."
                    value={searchParams.batch_number}
                    onChange={(e) => setSearchParams({...searchParams, batch_number: e.target.value})}
                    onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                    className="pl-8 max-w-sm"
                  />
                </div>
                <Select
                  value={searchParams.status}
                  onValueChange={(value: "all" | "unused" | "used" | "expired") => 
                    setSearchParams({...searchParams, status: value})
                  }
                >
                  <SelectTrigger className="w-[120px]">
                    <SelectValue placeholder="状态筛选" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部状态</SelectItem>
                    <SelectItem value="unused">未使用</SelectItem>
                    <SelectItem value="used">已使用</SelectItem>
                    <SelectItem value="expired">已过期</SelectItem>
                  </SelectContent>
                </Select>
                <Button variant="outline" size="sm" onClick={handleSearch}>
                  搜索
                </Button>
                <Button variant="outline" size="sm" onClick={handleResetSearch}>
                  重置
                </Button>
              </div>
              <div className="flex items-center space-x-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={exportToCSV}
                  disabled={loading}
                >
                  <Download className="mr-2 h-4 w-4" />
                  导出CSV
                </Button>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" className="ml-auto">
                      列显示 <ChevronDown className="ml-2 h-4 w-4" />
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
                            {column.id === "code" && "激活码"}
                            {column.id === "status" && "状态"}
                            {column.id === "plan_title" && "关联计划"}
                            {column.id === "used_by_username" && "使用者"}
                            {column.id === "batch_number" && "批次号"}
                            {column.id === "created_at" && "创建时间"}
                          </DropdownMenuCheckboxItem>
                        )
                      })}
                  </DropdownMenuContent>
                </DropdownMenu>
                <Button
                  onClick={() => setIsCodeDialogOpen(true)}
                  className="bg-orange-500 hover:bg-orange-600 text-white"
                >
                  <Plus className="w-4 h-4 mr-2" />
                  生成激活码
                </Button>
              </div>
            </div>

            {/* 当前搜索条件显示 */}
            {(searchParams.batch_number || searchParams.status !== "all") && (
              <div className="flex items-center space-x-2 pb-4">
                <span className="text-sm text-muted-foreground">当前筛选:</span>
                {searchParams.batch_number && (
                  <Badge variant="secondary">
                    批次号: {searchParams.batch_number}
                  </Badge>
                )}
                {searchParams.status !== "all" && (
                  <Badge variant="secondary">
                    状态: {searchParams.status === "used" ? "已使用" : 
                           searchParams.status === "unused" ? "未使用" : "已过期"}
                  </Badge>
                )}
              </div>
            )}

            {/* 数据表格 */}
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
                  {loading ? (
                    <TableRow>
                      <TableCell colSpan={columns.length} className="h-24 text-center">
                        加载中...
                      </TableCell>
                    </TableRow>
                  ) : table.getRowModel().rows?.length ? (
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
                      <TableCell colSpan={columns.length} className="h-24 text-center">
                        暂无数据
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>

            {/* 分页控制 */}
            <div className="flex items-center justify-between space-x-2 py-4">
              <div className="flex-1 text-sm text-muted-foreground">
                {table.getSelectedRowModel().rows.length} / {data.length} 行已选择
                <span className="ml-4">共 {pagination.total} 条记录</span>
              </div>
              <div className="flex items-center space-x-6 lg:space-x-8">
                <div className="flex items-center space-x-2">
                  <p className="text-sm font-medium">每页显示</p>
                  <Select
                    value={`${pagination.pageSize}`}
                    onValueChange={(value) => {
                      setPagination(prev => ({ ...prev, pageSize: Number(value), page: 1 }))
                    }}
                  >
                    <SelectTrigger className="h-8 w-[70px]">
                      <SelectValue placeholder={pagination.pageSize} />
                    </SelectTrigger>
                    <SelectContent side="top">
                      {[10, 20, 30, 40, 50].map((pageSize) => (
                        <SelectItem key={pageSize} value={`${pageSize}`}>
                          {pageSize}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex w-[100px] items-center justify-center text-sm font-medium">
                  第 {pagination.page} 页 / 共 {pagination.totalPages} 页
                </div>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPagination(prev => ({ ...prev, page: 1 }))}
                    disabled={pagination.page <= 1}
                  >
                    首页
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPagination(prev => ({ ...prev, page: prev.page - 1 }))}
                    disabled={pagination.page <= 1}
                  >
                    上一页
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPagination(prev => ({ ...prev, page: prev.page + 1 }))}
                    disabled={pagination.page >= pagination.totalPages}
                  >
                    下一页
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPagination(prev => ({ ...prev, page: pagination.totalPages }))}
                    disabled={pagination.page >= pagination.totalPages}
                  >
                    尾页
                  </Button>
                </div>
              </div>
            </div>
          </div>

          {/* 激活码生成对话框 */}
          <Dialog open={isCodeDialogOpen} onOpenChange={setIsCodeDialogOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>生成激活码</DialogTitle>
                <DialogDescription>为指定订阅计划批量生成激活码</DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>订阅计划</Label>
                  <Select
                    value={newCodeData.subscription_plan_id.toString()}
                    onValueChange={(value) => setNewCodeData({...newCodeData, subscription_plan_id: parseInt(value)})}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="选择订阅计划" />
                    </SelectTrigger>
                    <SelectContent>
                      {Array.isArray(plans) && plans.map((plan) => (
                        <SelectItem key={plan.id} value={plan.id.toString()}>
                          {plan.title} ({plan.point_amount.toLocaleString()} 积分, {plan.validity_days}天)
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>生成数量</Label>
                  <Input
                    type="number"
                    min="1"
                    max="100"
                    value={newCodeData.count}
                    onChange={(e) => setNewCodeData({...newCodeData, count: parseInt(e.target.value) || 1})}
                  />
                </div>
                <div className="space-y-2">
                  <Label>批次号 (可选)</Label>
                  <Input
                    value={newCodeData.batch_number}
                    onChange={(e) => setNewCodeData({...newCodeData, batch_number: e.target.value})}
                    placeholder="例如: BATCH-2024-001"
                  />
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsCodeDialogOpen(false)}>取消</Button>
                <Button onClick={handleCreateCodes}>生成</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>
    </DashboardLayout>
  )
} 