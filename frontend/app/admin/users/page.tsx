"use client"

import { useState, useEffect } from "react"
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
  VisibilityState,
} from "@tanstack/react-table"
import { ArrowUpDown, ChevronDown, MoreHorizontal, Users, Edit, Download, Search, Trash2, UserX, UserCheck, Gift } from "lucide-react"
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
import { copyToClipboard } from "@/lib/clipboard"
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
import { Switch } from "@/components/ui/switch"
import { useToast } from "@/components/ui/use-toast"
import { adminAPI, type AdminUser, type SubscriptionPlan } from "@/api/admin"

// 数据类型定义
type UserRow = {
  id: number
  username: string
  email: string
  is_admin: boolean
  is_disabled: boolean  // 新增禁用状态字段
  degradation_guaranteed: number
  created_at: string
}


export default function AdminUsersPage() {
  const { toast } = useToast()
  const [data, setData] = useState<UserRow[]>([])
  const [plans, setPlans] = useState<SubscriptionPlan[]>([])
  const [loading, setLoading] = useState(false)
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [rowSelection, setRowSelection] = useState({})
  
  // 服务端分页状态
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 10,
    total: 0,
    totalPages: 0
  })
  const [searchQuery, setSearchQuery] = useState("")
  
  // 对话框状态
  const [editingUser, setEditingUser] = useState<AdminUser | null>(null)
  const [isUserDialogOpen, setIsUserDialogOpen] = useState(false)
  const [isGiftDialogOpen, setIsGiftDialogOpen] = useState(false)
  const [giftingUser, setGiftingUser] = useState<UserRow | null>(null)
  const [giftData, setGiftData] = useState({
    subscription_plan_id: 0,
    points_amount: "",
    validity_days: "",
    daily_max_points: "",
    reason: ""
  })
  

  // 列定义
  const columns: ColumnDef<UserRow>[] = [
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
      accessorKey: "id",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            ID
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => <div className="font-medium">{row.getValue("id")}</div>,
    },
    {
      accessorKey: "username",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            用户名
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => <div className="font-medium">{row.getValue("username")}</div>,
    },
    {
      accessorKey: "email",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            邮箱
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => <div>{row.getValue("email")}</div>,
    },
    {
      accessorKey: "is_admin",
      header: "管理员",
      cell: ({ row }) => {
        const isAdmin = row.getValue("is_admin") as boolean
        return (
          <Badge variant={isAdmin ? "default" : "secondary"}>
            {isAdmin ? "是" : "否"}
          </Badge>
        )
      },
    },
    {
      accessorKey: "is_disabled",
      header: "用户状态",
      cell: ({ row }) => {
        const isDisabled = row.getValue("is_disabled") as boolean
        return (
          <Badge variant={isDisabled ? "destructive" : "default"}>
            {isDisabled ? "已禁用" : "正常"}
          </Badge>
        )
      },
    },
    {
      accessorKey: "degradation_guaranteed",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            保证不降级数量
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => <div>{row.getValue("degradation_guaranteed")}</div>,
    },
    {
      accessorKey: "created_at",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            注册时间
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
        const user = row.original
        
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
                onClick={async () => {
                  const success = await copyToClipboard(user.email)
                  if (!success) {
                    alert("无法复制邮箱，请手动复制。")
                  }
                }}
              >
                复制邮箱
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={() => handleEditUser(user)}
              >
                <Edit className="mr-2 h-4 w-4" />
                编辑用户
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => handleGiftSubscription(user)}
              >
                <Gift className="mr-2 h-4 w-4" />
                赠送卡密
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => handleToggleUserStatus(user)}
              >
                {user.is_disabled ? (
                  <>
                    <UserCheck className="mr-2 h-4 w-4" />
                    解禁用户
                  </>
                ) : (
                  <>
                    <UserX className="mr-2 h-4 w-4" />
                    禁用用户
                  </>
                )}
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => handleDeleteUser(user.id)}
                className="text-red-600"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                删除用户
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]

  // 搜索处理（移除客户端搜索，使用服务端搜索）
  const filteredData = data

  const table = useReactTable({
    data: filteredData,
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
  const loadUsers = async () => {
    setLoading(true)
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...(searchQuery && { search: searchQuery.trim() })
    }
    
    const result = await adminAPI.getUsers(params)
    if (result.success && result.users) {
      const users = Array.isArray(result.users) ? result.users : []
      const transformedData: UserRow[] = users.map((user: AdminUser) => ({
        id: user.id,
        username: user.username,
        email: user.email,
        is_admin: user.is_admin,
        is_disabled: user.is_disabled,
        degradation_guaranteed: user.degradation_guaranteed,
        created_at: user.created_at,
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

  // 加载订阅计划
  const loadPlans = async () => {
    const result = await adminAPI.getSubscriptionPlans()
    if (result.success && result.plans) {
      setPlans(Array.isArray(result.plans) ? result.plans : [])
    }
  }

  useEffect(() => {
    loadUsers()
    loadPlans()
  }, [pagination.page, pagination.pageSize, searchQuery])

  // 处理编辑用户
  const handleEditUser = (user: UserRow) => {
    // 从原始数据中找到完整的用户信息
    const fullUser = data.find(u => u.id === user.id)
    if (fullUser) {
      setEditingUser({
        id: fullUser.id,
        username: fullUser.username,
        email: fullUser.email,
        is_admin: fullUser.is_admin,
        is_disabled: fullUser.is_disabled,
        degradation_guaranteed: fullUser.degradation_guaranteed,
        degradation_source: "system",
        degradation_locked: false,
        degradation_counter: 0,
        created_at: fullUser.created_at,
        updated_at: fullUser.created_at
      })
      setIsUserDialogOpen(true)
    }
  }

  // 处理更新用户
  const handleUpdateUser = async (user: AdminUser) => {
    const result = await adminAPI.updateUser(user.id, user)
    if (result.success) {
      toast({ title: "更新成功", variant: "default" })
      setIsUserDialogOpen(false)
      loadUsers()
    } else {
      toast({
        title: "更新失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  // 处理删除用户
  const handleDeleteUser = async (userId: number) => {
    if (!confirm("确定要删除此用户吗？")) return
    
    const result = await adminAPI.deleteUser(userId)
    if (result.success) {
      toast({ title: "删除成功", variant: "default" })
      loadUsers()
    } else {
      toast({
        title: "删除失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  // 处理切换用户状态
  const handleToggleUserStatus = async (user: UserRow) => {
    const newDisabledStatus = !user.is_disabled
    const actionText = newDisabledStatus ? "禁用" : "解禁"
    
    if (!confirm(`确定要${actionText}用户 "${user.username}" 吗？`)) return
    
    const result = await adminAPI.toggleUserStatus(user.id, newDisabledStatus)
    if (result.success) {
      toast({ 
        title: `${actionText}成功`, 
        description: result.message,
        variant: "default" 
      })
      loadUsers()
    } else {
      toast({
        title: `${actionText}失败`,
        description: result.message,
        variant: "destructive"
      })
    }
  }

  // 处理赠送订阅
  const handleGiftSubscription = (user: UserRow) => {
    setGiftingUser(user)
    setGiftData({
      subscription_plan_id: 0,
      points_amount: "",
      validity_days: "",
      daily_max_points: "",
      reason: ""
    })
    setIsGiftDialogOpen(true)
  }

  // 执行赠送订阅
  const handleSubmitGift = async () => {
    if (!giftingUser || giftData.subscription_plan_id === 0) {
      toast({
        title: "请选择订阅计划",
        variant: "destructive"
      })
      return
    }

    const requestData: any = {
      subscription_plan_id: giftData.subscription_plan_id,
      reason: giftData.reason
    }

    // 添加自定义值（如果有）
    if (giftData.points_amount) {
      requestData.points_amount = parseInt(giftData.points_amount)
    }
    if (giftData.validity_days) {
      requestData.validity_days = parseInt(giftData.validity_days)
    }
    if (giftData.daily_max_points) {
      requestData.daily_max_points = parseInt(giftData.daily_max_points)
    }

    const result = await adminAPI.giftSubscription(giftingUser.id, requestData)
    if (result.success) {
      toast({
        title: "赠送成功",
        description: result.message,
        variant: "default"
      })
      setIsGiftDialogOpen(false)
      loadUsers()
    } else {
      toast({
        title: "赠送失败",
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
      : filteredData

    const csvContent = [
      // CSV 头部
      ['ID', '用户名', '邮箱', '管理员', '用户状态', '保证不降级数量', '注册时间'].join(','),
      // CSV 数据
      ...dataToExport.map(row => [
        row.id,
        `"${row.username}"`,
        `"${row.email}"`,
        row.is_admin ? "是" : "否",
        row.is_disabled ? "已禁用" : "正常",
        row.degradation_guaranteed,
        `"${new Date(row.created_at).toLocaleString()}"`
      ].join(','))
    ].join('\n')

    const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = `users_${new Date().toISOString().split('T')[0]}.csv`
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
            <Users className="h-6 w-6 text-blue-500" />
            <CardTitle className="text-slate-900 dark:text-slate-100">用户管理</CardTitle>
          </div>
          <CardDescription className="text-slate-600 dark:text-slate-400">
            管理系统中的所有用户账户，支持搜索、筛选和导出
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
                    placeholder="搜索用户名、邮箱..."
                    value={searchQuery}
                    onChange={(e) => {
                      setSearchQuery(e.target.value)
                      // 重置到第一页
                      setPagination(prev => ({ ...prev, page: 1 }))
                    }}
                    className="pl-8 max-w-sm"
                  />
                </div>
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
                            {column.id === "id" && "ID"}
                            {column.id === "username" && "用户名"}
                            {column.id === "email" && "邮箱"}
                            {column.id === "is_admin" && "管理员"}
                            {column.id === "is_disabled" && "用户状态"}
                            {column.id === "degradation_guaranteed" && "保证不降级数量"}
                            {column.id === "created_at" && "注册时间"}
                          </DropdownMenuCheckboxItem>
                        )
                      })}
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>

            {/* 当前搜索条件显示 */}
            {searchQuery && (
              <div className="flex items-center space-x-2 pb-4">
                <span className="text-sm text-muted-foreground">当前搜索:</span>
                <Badge variant="secondary">
                  关键词: {searchQuery}
                </Badge>
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
                  ) : filteredData.length ? (
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
                {table.getSelectedRowModel().rows.length} / {filteredData.length} 行已选择
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

          {/* 用户编辑对话框 */}
          <Dialog open={isUserDialogOpen} onOpenChange={setIsUserDialogOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>编辑用户</DialogTitle>
                <DialogDescription>修改用户信息和权限设置</DialogDescription>
              </DialogHeader>
              {editingUser && (
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label>用户名</Label>
                    <Input
                      value={editingUser.username}
                      onChange={(e) => setEditingUser({...editingUser, username: e.target.value})}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>邮箱</Label>
                    <Input
                      value={editingUser.email}
                      onChange={(e) => setEditingUser({...editingUser, email: e.target.value})}
                    />
                  </div>
                  <div className="flex items-center space-x-2">
                    <Switch
                      checked={editingUser.is_admin}
                      onCheckedChange={(checked) => setEditingUser({...editingUser, is_admin: checked})}
                    />
                    <Label>管理员权限</Label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Switch
                      checked={editingUser.is_disabled}
                      onCheckedChange={(checked) => setEditingUser({...editingUser, is_disabled: checked})}
                    />
                    <Label>禁用用户</Label>
                  </div>
                  <div className="space-y-2">
                    <Label>10条保证不降级的数量</Label>
                    <Input
                      type="number"
                      value={editingUser.degradation_guaranteed}
                      onChange={(e) => setEditingUser({...editingUser, degradation_guaranteed: parseInt(e.target.value) || 0})}
                    />
                  </div>
                </div>
              )}
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsUserDialogOpen(false)}>取消</Button>
                <Button onClick={() => editingUser && handleUpdateUser(editingUser)}>保存</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* 赠送卡密对话框 */}
          <Dialog open={isGiftDialogOpen} onOpenChange={setIsGiftDialogOpen}>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>赠送卡密</DialogTitle>
                <DialogDescription>
                  为用户 "{giftingUser?.username}" 赠送订阅计划
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>订阅计划 *</Label>
                  <Select
                    value={giftData.subscription_plan_id.toString()}
                    onValueChange={(value) => setGiftData({...giftData, subscription_plan_id: parseInt(value)})}
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
                  <Label>自定义积分数量 (可选)</Label>
                  <Input
                    type="number"
                    placeholder="留空使用计划默认值"
                    value={giftData.points_amount}
                    onChange={(e) => setGiftData({...giftData, points_amount: e.target.value})}
                  />
                </div>
                
                <div className="space-y-2">
                  <Label>自定义有效期天数 (可选)</Label>
                  <Input
                    type="number"
                    placeholder="留空使用计划默认值"
                    value={giftData.validity_days}
                    onChange={(e) => setGiftData({...giftData, validity_days: e.target.value})}
                  />
                </div>
                
                <div className="space-y-2">
                  <Label>每日最大使用积分 (可选)</Label>
                  <Input
                    type="number"
                    placeholder="0表示无限制，留空使用计划默认值"
                    value={giftData.daily_max_points}
                    onChange={(e) => setGiftData({...giftData, daily_max_points: e.target.value})}
                  />
                </div>
                
                <div className="space-y-2">
                  <Label>赠送原因 (可选)</Label>
                  <Input
                    placeholder="例如：活动奖励、客服补偿等"
                    value={giftData.reason}
                    onChange={(e) => setGiftData({...giftData, reason: e.target.value})}
                  />
                </div>
                
                <div className="bg-muted p-3 rounded text-sm">
                  <p className="font-medium mb-1">用户信息：</p>
                  <p>用户名: {giftingUser?.username}</p>
                  <p>邮箱: {giftingUser?.email}</p>
                  <p>用户ID: {giftingUser?.id}</p>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsGiftDialogOpen(false)}>取消</Button>
                <Button onClick={handleSubmitGift}>
                  <Gift className="mr-2 h-4 w-4" />
                  确认赠送
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

        </CardContent>
      </Card>
    </DashboardLayout>
  )
}