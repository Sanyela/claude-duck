"use client"

export const dynamic = 'force-dynamic'

import { useState, useEffect, useCallback, useRef } from "react"
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
  VisibilityState,
} from "@tanstack/react-table"
import { ArrowUpDown, ChevronDown, MoreHorizontal, Key, Plus, Download, Search, Copy, Settings, Ban, ShieldOff, Eye, AlertTriangle } from "lucide-react"
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
import { useToast } from "@/hooks/use-toast"
import { adminAPI, type ActivationCode, type SubscriptionPlan, type FrozenPointsRecord } from "@/api/admin"
import { getUserInfo, type User } from "@/api/auth"

// 数据类型定义
type ActivationCodeRow = {
  id: number
  code: string
  status: "unused" | "used" | "expired" | "depleted"
  plan_title: string
  used_by_username: string | null
  batch_number: string
  created_at: string
  used_at: string | null
  // 积分信息
  total_points: number
  used_points: number
  available_points: number
  // 后端返回的原始数据结构
  subscription?: {
    total_points: number
    used_points: number
    available_points: number
  }
  plan?: {
    point_amount: number
  }
}

export default function AdminCodesPage() {
  const { toast } = useToast()
  const [data, setData] = useState<ActivationCodeRow[]>([])
  const [plans, setPlans] = useState<SubscriptionPlan[]>([])
  const [currentUser, setCurrentUser] = useState<User | null>(null)
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
  // 搜索类型定义
  type SearchType = "batch_number" | "code" | "username" | "frozen_records"
  
  const [searchParams, setSearchParams] = useState({
    query: "",
    type: "batch_number" as SearchType,
    status: "all" as "all" | "unused" | "used" | "expired" | "depleted"
  })
  
  // 防抖用的ref
  const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const isManualSearchRef = useRef(false)
  
  // 对话框状态
  const [isCodeDialogOpen, setIsCodeDialogOpen] = useState(false)
  const [isCopyDialogOpen, setIsCopyDialogOpen] = useState(false)
  const [isDailyLimitDialogOpen, setIsDailyLimitDialogOpen] = useState(false)
  
  // 封禁相关对话框状态
  const [isBanPreviewDialogOpen, setIsBanPreviewDialogOpen] = useState(false)
  const [isBanDialogOpen, setIsBanDialogOpen] = useState(false)
  const [banPreviewData, setBanPreviewData] = useState<any>(null)
  const [selectedCodeForBan, setSelectedCodeForBan] = useState<ActivationCodeRow | null>(null)
  const [banReason, setBanReason] = useState("")
  
  // 冻结记录数据
  const [frozenRecords, setFrozenRecords] = useState<FrozenPointsRecord[]>([])
  const [showingFrozenRecords, setShowingFrozenRecords] = useState(false)
  const [generatedBatchNumber, setGeneratedBatchNumber] = useState("")
  const [newCodeData, setNewCodeData] = useState({
    subscription_plan_id: 0,
    count: 1,
    batch_number: ""
  })
  const [editingCode, setEditingCode] = useState<ActivationCodeRow | null>(null)
  const [dailyLimitData, setDailyLimitData] = useState({
    daily_limit: 0
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
        const code = row.original
        const status = code.status
        let displayText = ""
        let variant: "default" | "secondary" | "destructive" | "outline" = "secondary"

        // 动态判断状态
        if (status === "unused") {
          displayText = "未使用"
          variant = "secondary"
        } else if (status === "expired") {
          displayText = "已过期"
          variant = "destructive"
        } else if (status === "used") {
          // 进一步检查是否已用完
          if (code.available_points === 0 && code.total_points > 0) {
            displayText = "已用完"
            variant = "outline"
          } else {
            displayText = "已使用"
            variant = "default"
          }
        } else if (status === "depleted") {
          displayText = "已用完"
          variant = "outline"
        }

        return (
          <Badge variant={variant}>
            {displayText}
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
      accessorKey: "points_info",
      header: "积分信息",
      cell: ({ row }) => {
        const code = row.original
        if (code.status === "unused") {
          return <div className="text-muted-foreground">-</div>
        }
        
        const usagePercentage = code.total_points > 0 ? (code.used_points / code.total_points) * 100 : 0
        const isHighUsage = usagePercentage > 80
        const isMediumUsage = usagePercentage > 50
        
        return (
          <div className="space-y-1">
            <div className="text-sm">
              <span className={isHighUsage ? "text-red-600" : isMediumUsage ? "text-yellow-600" : "text-green-600"}>
                {code.used_points.toLocaleString()}
              </span>
              <span className="text-muted-foreground"> / {code.total_points.toLocaleString()}</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-1.5">
              <div 
                className={`h-1.5 rounded-full ${
                  isHighUsage ? "bg-red-500" : isMediumUsage ? "bg-yellow-500" : "bg-green-500"
                }`}
                style={{ width: `${Math.min(usagePercentage, 100)}%` }}
              ></div>
            </div>
            <div className="text-xs text-muted-foreground">
              {usagePercentage.toFixed(1)}% 已使用
            </div>
          </div>
        )
      },
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
                onClick={async () => {
                  const success = await copyToClipboard(code.code)
                  if (!success) {
                    alert("无法复制激活码，请手动复制。")
                  }
                }}
              >
                <Copy className="mr-2 h-4 w-4" />
                复制激活码
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              
              {/* 封禁相关操作 */}
              {code.status === "used" && code.used_by_username && (
                <>
                  <DropdownMenuItem
                    onClick={() => handleBanPreview(code)}
                    className="text-orange-600"
                  >
                    <Eye className="mr-2 h-4 w-4" />
                    预览封禁影响
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => handleBanCode(code)}
                    className="text-red-600"
                  >
                    <Ban className="mr-2 h-4 w-4" />
                    封禁激活码
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                </>
              )}
              
              <DropdownMenuItem
                onClick={() => handleEditDailyLimit(code)}
                disabled={code.status !== "used"}
              >
                <Settings className="mr-2 h-4 w-4" />
                编辑每日最大使用量
              </DropdownMenuItem>
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

  // 冻结记录的列定义
  const frozenRecordsColumns: ColumnDef<FrozenPointsRecord>[] = [
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
      accessorKey: "banned_activation_code",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            被封禁激活码
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => (
        <div className="font-mono text-sm">{row.getValue("banned_activation_code")}</div>
      ),
    },
    {
      accessorKey: "user",
      header: "受影响用户",
      cell: ({ row }) => {
        const user = row.original.user
        return <div>{user?.username || `用户ID: ${row.original.user_id}`}</div>
      },
    },
    {
      accessorKey: "frozen_points",
      header: "冻结积分",
      cell: ({ row }) => (
        <div className="text-red-600 font-medium">
          {row.getValue<number>("frozen_points").toLocaleString()}
        </div>
      ),
    },
    {
      accessorKey: "status",
      header: "状态",
      cell: ({ row }) => {
        const status = row.getValue("status") as string
        return (
          <Badge variant={status === "frozen" ? "destructive" : "secondary"}>
            {status === "frozen" ? "已冻结" : "已恢复"}
          </Badge>
        )
      },
    },
    {
      accessorKey: "ban_reason",
      header: "封禁原因",
      cell: ({ row }) => {
        const reason = row.getValue("ban_reason") as string
        return (
          <div className="max-w-xs truncate" title={reason}>
            {reason || "未填写"}
          </div>
        )
      },
    },
    {
      accessorKey: "admin_user",
      header: "操作管理员",
      cell: ({ row }) => {
        const admin = row.original.admin_user
        return <div>{admin?.username || "系统"}</div>
      },
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
        const record = row.original
        
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
                  const success = await copyToClipboard(record.banned_activation_code)
                  if (!success) {
                    alert("无法复制激活码，请手动复制。")
                  }
                }}
              >
                <Copy className="mr-2 h-4 w-4" />
                复制激活码
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              
              {/* 解禁操作 */}
              {record.status === "frozen" && (
                <DropdownMenuItem
                  onClick={() => handleUnbanCode(record)}
                  className="text-green-600"
                >
                  <ShieldOff className="mr-2 h-4 w-4" />
                  解禁激活码
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]

  // 根据当前显示的数据类型选择对应的表格
  const currentColumns = showingFrozenRecords ? frozenRecordsColumns : columns
  const currentData = showingFrozenRecords ? frozenRecords : data

  const table = useReactTable({
    data: currentData,
    columns: currentColumns,
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
  const loadCodes = useCallback(async (page?: number, pageSize?: number, search?: typeof searchParams) => {
    setLoading(true)
    const currentPage = page ?? pagination.page
    const currentPageSize = pageSize ?? pagination.pageSize
    const currentSearch = search ?? searchParams
    
    // 如果选择的是冻结记录，则加载冻结记录数据
    if (currentSearch.type === "frozen_records") {
      setShowingFrozenRecords(true)
      
      const params = {
        page: currentPage,
        page_size: currentPageSize,
        ...(currentSearch.query && { activation_code: currentSearch.query }),
        ...(currentSearch.status !== "all" && { status: currentSearch.status === "used" ? "frozen" : currentSearch.status }),
      }
      
      const result = await adminAPI.getFrozenRecords(params)
      if (result.success && result.data) {
        setFrozenRecords(result.data)
        setData([]) // 清空激活码数据
        
        // 更新分页信息
        setPagination(prev => ({
          ...prev,
          total: result.total || 0,
          totalPages: result.total_pages || 0,
          page: result.page || currentPage,
          pageSize: result.page_size || currentPageSize
        }))
      } else {
        setFrozenRecords([])
        toast({
          title: "加载失败",
          description: result.message,
          variant: "destructive"
        })
      }
    } else {
      // 加载激活码数据
      setShowingFrozenRecords(false)
      
      const params = {
        page: currentPage,
        page_size: currentPageSize,
        ...(currentSearch.query && { 
          [currentSearch.type]: currentSearch.query 
        }),
        ...(currentSearch.status !== "all" && { status: currentSearch.status }),
      }
      
      const result = await adminAPI.getActivationCodes(params)
      if (result.success && result.codes) {
        const codes = Array.isArray(result.codes) ? result.codes : []
        const transformedData: ActivationCodeRow[] = codes.map((code: ActivationCode) => ({
          id: code.id,
          code: code.code,
          status: code.status as "unused" | "used" | "expired" | "depleted",
          plan_title: code.plan?.title || "未知计划",
          used_by_username: code.used_by?.username || null,
          batch_number: code.batch_number || "",
          created_at: code.created_at,
          used_at: code.used_at || null,
          // 积分信息 - 如果后端返回了subscription信息则使用，否则使用默认值
          total_points: code.subscription?.total_points || code.plan?.point_amount || 0,
          used_points: code.subscription?.used_points || 0,
          available_points: code.subscription?.available_points || 0,
        }))
        setData(transformedData)
        setFrozenRecords([]) // 清空冻结记录数据
        
        // 更新分页信息
        setPagination(prev => ({
          ...prev,
          total: result.total || 0,
          totalPages: result.total_pages || 0,
          page: result.page || currentPage,
          pageSize: result.page_size || currentPageSize
        }))
      } else {
        setData([])
        toast({
          title: "加载失败",
          description: result.message,
          variant: "destructive"
        })
      }
    }
    setLoading(false)
  }, [])

  const loadPlans = async () => {
    const result = await adminAPI.getSubscriptionPlans()
    if (result.success && result.plans) {
      setPlans(Array.isArray(result.plans) ? result.plans : [])
    }
  }

  // 获取当前用户信息
  const loadCurrentUser = async () => {
    const result = await getUserInfo()
    if (result.success && result.user) {
      setCurrentUser(result.user)
    }
  }

  // 初始化时从本地存储读取搜索类型偏好
  useEffect(() => {
    const savedSearchType = localStorage.getItem('activationCode_searchType') as SearchType | null
    if (savedSearchType && ['batch_number', 'code', 'username', 'frozen_records'].includes(savedSearchType)) {
      setSearchParams(prev => ({ ...prev, type: savedSearchType }))
    }
  }, [])

  useEffect(() => {
    loadPlans()
    loadCurrentUser()
  }, [])

  useEffect(() => {
    loadCodes(pagination.page, pagination.pageSize)
  }, [pagination.page, pagination.pageSize, loadCodes])

  // 监听搜索类型和状态变化（立即触发）
  useEffect(() => {
    // 搜索类型或状态变化时立即重置到第一页并搜索
    if (pagination.page === 1) {
      loadCodes(1, pagination.pageSize, searchParams)
    } else {
      setPagination(prev => ({ ...prev, page: 1 }))
    }
  }, [searchParams.type, searchParams.status, loadCodes, pagination.pageSize])

  // 监听搜索查询变化（防抖处理）
  useEffect(() => {
    // 清除之前的定时器
    if (debounceTimeoutRef.current) {
      clearTimeout(debounceTimeoutRef.current)
    }

    // 如果是手动搜索或查询为空，立即执行
    if (isManualSearchRef.current || searchParams.query === '') {
      isManualSearchRef.current = false // 重置标志
      if (pagination.page === 1) {
        loadCodes(1, pagination.pageSize, searchParams)
      } else {
        setPagination(prev => ({ ...prev, page: 1 }))
      }
      return
    }

    // 设置防抖定时器
    debounceTimeoutRef.current = setTimeout(() => {
      if (pagination.page === 1) {
        loadCodes(1, pagination.pageSize, searchParams)
      } else {
        setPagination(prev => ({ ...prev, page: 1 }))
      }
    }, 500) // 500ms 防抖延迟

    // 清理函数
    return () => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current)
      }
    }
  }, [searchParams.query, loadCodes, pagination.pageSize])


  // 搜索处理
  const handleSearch = () => {
    isManualSearchRef.current = true // 设置手动搜索标志
    setPagination(prev => ({ ...prev, page: 1 }))
    loadCodes(1, pagination.pageSize, searchParams)
  }

  // 重置搜索
  const handleResetSearch = () => {
    isManualSearchRef.current = true // 设置手动搜索标志
    setSearchParams({ query: "", type: searchParams.type, status: "all" })
  }

  // 处理搜索类型变化
  const handleSearchTypeChange = (newType: SearchType) => {
    setSearchParams(prev => ({ ...prev, type: newType, query: "" }))
    // 保存到本地存储
    localStorage.setItem('activationCode_searchType', newType)
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
      const batchNumber = newCodeData.batch_number
      
      // 显示成功toast
      toast({
        title: "创建成功",
        description: "激活码生成完成",
        variant: "default"
      })
      
      // 设置批次号并显示复制弹窗
      setGeneratedBatchNumber(batchNumber)
      setIsCopyDialogOpen(true)
      
      setIsCodeDialogOpen(false)
      setNewCodeData({
        subscription_plan_id: 0,
        count: 1,
        batch_number: ""
      })
      loadCodes(pagination.page, pagination.pageSize, searchParams)
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
      loadCodes(pagination.page, pagination.pageSize, searchParams)
    } else {
      toast({
        title: "删除失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  // ===== 封禁相关处理函数 =====

  // 处理封禁预览
  const handleBanPreview = async (code: ActivationCodeRow) => {
    if (!code.used_by_username) {
      toast({
        title: "无法预览",
        description: "该激活码未被使用",
        variant: "destructive"
      })
      return
    }

    setSelectedCodeForBan(code)
    setLoading(true)
    
    try {
      // 通过用户名获取用户信息
      const userResult = await adminAPI.getUsers({ search: code.used_by_username })
      if (!userResult.success || !userResult.users || userResult.users.length === 0) {
        toast({
          title: "预览失败",
          description: "无法找到对应的用户信息",
          variant: "destructive"
        })
        return
      }
      
      const user = userResult.users.find(u => u.username === code.used_by_username)
      if (!user) {
        toast({
          title: "预览失败",
          description: "无法找到对应的用户",
          variant: "destructive"
        })
        return
      }
      
      const result = await adminAPI.previewBanActivationCode({
        user_id: user.id,
        activation_code: code.code
      })
      
      if (result.success) {
        setBanPreviewData(result)
        setIsBanPreviewDialogOpen(true)
      } else {
        toast({
          title: "预览失败",
          description: result.message,
          variant: "destructive"
        })
      }
    } catch (error) {
      toast({
        title: "预览失败",
        description: "获取封禁预览信息时发生错误",
        variant: "destructive"
      })
    }
    
    setLoading(false)
  }

  // 处理封禁激活码
  const handleBanCode = (code: ActivationCodeRow) => {
    setSelectedCodeForBan(code)
    setBanReason("")
    setIsBanDialogOpen(true)
  }

  // 执行封禁操作
  const handleExecuteBan = async () => {
    if (!selectedCodeForBan) return
    
    setLoading(true)
    try {
      // 通过用户名获取用户信息
      const userResult = await adminAPI.getUsers({ search: selectedCodeForBan.used_by_username! })
      if (!userResult.success || !userResult.users || userResult.users.length === 0) {
        toast({
          title: "封禁失败",
          description: "无法找到对应的用户信息",
          variant: "destructive"
        })
        return
      }
      
      const user = userResult.users.find(u => u.username === selectedCodeForBan.used_by_username)
      if (!user) {
        toast({
          title: "封禁失败",
          description: "无法找到对应的用户",
          variant: "destructive"
        })
        return
      }
      
      const result = await adminAPI.banActivationCode({
        user_id: user.id,
        activation_code: selectedCodeForBan.code,
        reason: banReason
      })
      
      if (result.success) {
        toast({
          title: "封禁成功",
          description: "激活码已被封禁",
          variant: "default"
        })
        setIsBanDialogOpen(false)
        loadCodes(pagination.page, pagination.pageSize, searchParams)
      } else {
        toast({
          title: "封禁失败",
          description: result.message,
          variant: "destructive"
        })
      }
    } catch (error) {
      toast({
        title: "封禁失败",
        description: "执行封禁操作时发生错误",
        variant: "destructive"
      })
    }
    setLoading(false)
  }

  // 处理解禁激活码 (从冻结记录中解禁)
  const handleUnbanCode = async (record: FrozenPointsRecord) => {
    if (!confirm("确定要解禁此激活码吗？")) return
    
    try {
      const result = await adminAPI.unbanActivationCode({
        user_id: record.user_id,
        activation_code: record.banned_activation_code
      })
      
      if (result.success) {
        toast({
          title: "解禁成功",
          description: "激活码已被解禁",
          variant: "default"
        })
        loadCodes(pagination.page, pagination.pageSize, searchParams)
      } else {
        toast({
          title: "解禁失败",
          description: result.message,
          variant: "destructive"
        })
      }
    } catch (error) {
      toast({
        title: "解禁失败",
        description: "执行解禁操作时发生错误",
        variant: "destructive"
      })
    }
  }

  // 处理编辑每日限制
  const handleEditDailyLimit = async (code: ActivationCodeRow) => {
    setEditingCode(code)
    // 获取当前每日限制
    try {
      const result = await adminAPI.getSubscriptionDailyLimit(code.id)
      if (result.success && result.daily_limit !== undefined) {
        setDailyLimitData({ daily_limit: result.daily_limit })
      } else {
        setDailyLimitData({ daily_limit: 0 })
      }
    } catch {
      setDailyLimitData({ daily_limit: 0 })
    }
    setIsDailyLimitDialogOpen(true)
  }

  // 处理更新每日限制
  const handleUpdateDailyLimit = async () => {
    if (!editingCode) return
    
    const result = await adminAPI.updateSubscriptionDailyLimit(editingCode.id, dailyLimitData.daily_limit)
    if (result.success) {
      toast({ 
        title: "更新成功", 
        description: "每日使用量限制已更新",
        variant: "default" 
      })
      setIsDailyLimitDialogOpen(false)
      loadCodes(pagination.page, pagination.pageSize, searchParams)
    } else {
      toast({
        title: "更新失败",
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
      ['激活码', '状态', '关联计划', '使用者', '批次号', '总积分', '已使用积分', '剩余积分', '创建时间', '使用时间'].join(','),
      // CSV 数据
      ...dataToExport.map(row => {
        let statusText = ""
        if (row.status === "unused") {
          statusText = "未使用"
        } else if (row.status === "expired") {
          statusText = "已过期"
        } else if (row.status === "used") {
          // 检查是否已用完
          if (row.available_points === 0 && row.total_points > 0) {
            statusText = "已用完"
          } else {
            statusText = "已使用"
          }
        } else if (row.status === "depleted") {
          statusText = "已用完"
        }
        
        return [
          `"${row.code}"`,
          statusText,
          `"${row.plan_title}"`,
          `"${row.used_by_username || "-"}"`,
          `"${row.batch_number}"`,
          row.total_points || "-",
          row.used_points || "-",
          row.available_points || "-",
          `"${new Date(row.created_at).toLocaleString()}"`,
          `"${row.used_at ? new Date(row.used_at).toLocaleString() : "-"}"`
        ].join(',')
      })
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

  // 生成批次号
  const generateBatchNumber = () => {
    if (!currentUser) return ""
    
    const now = new Date()
    const year = now.getFullYear()
    const month = String(now.getMonth() + 1).padStart(2, '0')
    const day = String(now.getDate()).padStart(2, '0')
    const hour = String(now.getHours()).padStart(2, '0')
    const minute = String(now.getMinutes()).padStart(2, '0')
    const second = String(now.getSeconds()).padStart(2, '0')
    
    return `${currentUser.username}-${year}-${month}-${day}-${hour}-${minute}-${second}`
  }

  // 处理复制批次号
  const handleCopyBatchNumber = async () => {
    try {
      const success = await copyToClipboard(generatedBatchNumber)
      if (success) {
        toast({
          title: "复制成功",
          description: "批次号已复制到剪贴板",
          variant: "default"
        })
        setIsCopyDialogOpen(false)
      } else {
        toast({
          title: "复制失败",
          description: "请手动选择批次号文本进行复制",
          variant: "destructive"
        })
      }
    } catch {
      toast({
        title: "复制失败",
        description: "请手动选择批次号文本进行复制",
        variant: "destructive"
      })
    }
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
                <Select
                  value={searchParams.type}
                  onValueChange={handleSearchTypeChange}
                >
                  <SelectTrigger className="w-[140px]">
                    <SelectValue placeholder="搜索类型" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="batch_number">批次号</SelectItem>
                    <SelectItem value="code">激活码</SelectItem>
                    <SelectItem value="username">用户名</SelectItem>
                    <SelectItem value="frozen_records">冻结记录</SelectItem>
                  </SelectContent>
                </Select>
                <div className="relative">
                  <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder={
                      searchParams.type === "batch_number" ? "输入批次号搜索..." :
                      searchParams.type === "code" ? "输入激活码搜索..." :
                      searchParams.type === "username" ? "输入用户名搜索..." :
                      "输入激活码搜索冻结记录..."
                    }
                    value={searchParams.query}
                    onChange={(e) => setSearchParams({...searchParams, query: e.target.value})}
                    onKeyPress={(e) => {
                      if (e.key === 'Enter') {
                        isManualSearchRef.current = true // 设置手动搜索标志
                        handleSearch()
                      }
                    }}
                    className="pl-8 max-w-sm"
                  />
                </div>
                <Select
                  value={searchParams.status}
                  onValueChange={(value: "all" | "unused" | "used" | "expired" | "depleted") => 
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
                    <SelectItem value="depleted">已用完</SelectItem>
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
                            {showingFrozenRecords ? (
                              // 冻结记录表的列名
                              column.id === "banned_activation_code" ? "被封禁激活码" :
                              column.id === "user" ? "受影响用户" :
                              column.id === "frozen_points" ? "冻结积分" :
                              column.id === "status" ? "状态" :
                              column.id === "ban_reason" ? "封禁原因" :
                              column.id === "admin_user" ? "操作管理员" :
                              column.id === "created_at" ? "创建时间" :
                              column.id
                            ) : (
                              // 激活码表的列名
                              column.id === "code" ? "激活码" :
                              column.id === "status" ? "状态" :
                              column.id === "plan_title" ? "关联计划" :
                              column.id === "used_by_username" ? "使用者" :
                              column.id === "batch_number" ? "批次号" :
                              column.id === "points_info" ? "积分信息" :
                              column.id === "created_at" ? "创建时间" :
                              column.id
                            )}
                          </DropdownMenuCheckboxItem>
                        )
                      })}
                  </DropdownMenuContent>
                </DropdownMenu>
                <Button
                  onClick={() => {
                    // 自动填入批次号
                    setNewCodeData({
                      subscription_plan_id: 0,
                      count: 1,
                      batch_number: generateBatchNumber()
                    })
                    setIsCodeDialogOpen(true)
                  }}
                  className="bg-orange-500 hover:bg-orange-600 text-white"
                >
                  <Plus className="w-4 h-4 mr-2" />
                  生成激活码
                </Button>
              </div>
            </div>

            {/* 当前搜索条件显示 */}
            {(searchParams.query || searchParams.status !== "all") && (
              <div className="flex items-center space-x-2 pb-4">
                <span className="text-sm text-muted-foreground">当前筛选:</span>
                {searchParams.query && (
                  <Badge variant="secondary">
                    {searchParams.type === "batch_number" ? "批次号" :
                     searchParams.type === "code" ? "激活码" : "用户名"}: {searchParams.query}
                  </Badge>
                )}
                {searchParams.status !== "all" && (
                  <Badge variant="secondary">
                    状态: {searchParams.status === "used" ? "已使用" : 
                           searchParams.status === "unused" ? "未使用" : 
                           searchParams.status === "depleted" ? "已用完" : "已过期"}
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
                {table.getSelectedRowModel().rows.length} / {currentData.length} 行已选择
                <span className="ml-4">共 {pagination.total} 条记录</span>
                {showingFrozenRecords && <span className="ml-2 text-orange-600">（冻结记录）</span>}
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

          {/* 复制批次号对话框 */}
          <Dialog open={isCopyDialogOpen} onOpenChange={setIsCopyDialogOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>激活码生成成功</DialogTitle>
                <DialogDescription>请复制以下批次号用于管理激活码</DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>批次号</Label>
                  <div className="bg-muted p-3 rounded text-sm font-mono break-all select-all border">
                    {generatedBatchNumber}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    点击上方文本可选中，或点击下方按钮自动复制
                  </p>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsCopyDialogOpen(false)}>稍后复制</Button>
                <Button onClick={handleCopyBatchNumber}>
                  <Copy className="mr-2 h-4 w-4" />
                  复制批次号
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* 编辑每日限制对话框 */}
          <Dialog open={isDailyLimitDialogOpen} onOpenChange={setIsDailyLimitDialogOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>编辑每日最大使用量</DialogTitle>
                <DialogDescription>
                  为激活码 &quot;{editingCode?.code}&quot; 对应的订阅设置每日最大积分使用量
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>每日最大使用量 (积分)</Label>
                  <Input
                    type="number"
                    min="0"
                    value={dailyLimitData.daily_limit}
                    onChange={(e) => setDailyLimitData({...dailyLimitData, daily_limit: parseInt(e.target.value) || 0})}
                    placeholder="0表示无限制"
                  />
                  <p className="text-xs text-muted-foreground">
                    设置为0表示无限制，大于0的数值表示每日最大积分使用量
                  </p>
                </div>
                <div className="bg-muted p-3 rounded text-sm">
                  <p className="font-medium mb-1">激活码信息：</p>
                  <p>激活码: {editingCode?.code}</p>
                  <p>关联计划: {editingCode?.plan_title}</p>
                  <p>使用者: {editingCode?.used_by_username || "未使用"}</p>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsDailyLimitDialogOpen(false)}>取消</Button>
                <Button onClick={handleUpdateDailyLimit}>保存</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* 封禁预览对话框 */}
          <Dialog open={isBanPreviewDialogOpen} onOpenChange={setIsBanPreviewDialogOpen}>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                  <Eye className="h-5 w-5 text-orange-500" />
                  封禁影响预览
                </DialogTitle>
                <DialogDescription>
                  预览封禁激活码 &quot;{selectedCodeForBan?.code}&quot; 对用户的影响
                </DialogDescription>
              </DialogHeader>
              {banPreviewData && (
                <div className="space-y-4">
                  {/* 当前钱包状态 */}
                  <div className="bg-blue-50 p-4 rounded-lg">
                    <h4 className="font-medium text-blue-900 mb-2">当前用户钱包状态</h4>
                    <div className="grid grid-cols-3 gap-4 text-sm">
                      <div>
                        <span className="text-muted-foreground">总积分：</span>
                        <span className="font-medium">{banPreviewData.current_wallet?.total_points?.toLocaleString() || 0}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">可用积分：</span>
                        <span className="font-medium text-green-600">{banPreviewData.current_wallet?.available_points?.toLocaleString() || 0}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">已用积分：</span>
                        <span className="font-medium text-orange-600">{banPreviewData.current_wallet?.used_points?.toLocaleString() || 0}</span>
                      </div>
                    </div>
                  </div>

                  {/* 封禁影响 */}
                  <div className="bg-red-50 p-4 rounded-lg">
                    <h4 className="font-medium text-red-900 mb-2">封禁后影响</h4>
                    <div className="grid grid-cols-3 gap-4 text-sm">
                      <div>
                        <span className="text-muted-foreground">将冻结积分：</span>
            <span className="font-medium text-red-600">{banPreviewData.ban_impact?.frozen_points?.toLocaleString() || 0}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">已消费积分：</span>
                        <span className="font-medium">{banPreviewData.ban_impact?.consumed_points?.toLocaleString() || 0}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">封禁后总积分：</span>
                        <span className="font-medium">{banPreviewData.ban_impact?.new_total_points?.toLocaleString() || 0}</span>
                      </div>
                    </div>
                  </div>

                  {/* 目标卡密信息 */}
                  {banPreviewData.target_card && (
                    <div className="bg-yellow-50 p-4 rounded-lg">
                      <h4 className="font-medium text-yellow-900 mb-2">目标激活码信息</h4>
                      <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                          <span className="text-muted-foreground">激活码：</span>
                          <span className="font-mono font-medium">{banPreviewData.target_card.code}</span>
                        </div>
                        <div>
                          <span className="text-muted-foreground">原始积分：</span>
                          <span className="font-medium">{banPreviewData.target_card.original_points?.toLocaleString() || 0}</span>
                        </div>
                        <div>
                          <span className="text-muted-foreground">剩余积分：</span>
                          <span className="font-medium text-green-600">{banPreviewData.target_card.remaining_points?.toLocaleString() || 0}</span>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* 消费详情 */}
                  {banPreviewData.consumption_details && banPreviewData.consumption_details.length > 0 && (
                    <div className="bg-gray-50 p-4 rounded-lg">
                      <h4 className="font-medium text-gray-900 mb-2">虚拟消费计算详情</h4>
                      <div className="max-h-32 overflow-y-auto space-y-2">
                        {banPreviewData.consumption_details.map((card: any, index: number) => (
                          <div key={index} className="flex justify-between items-center text-xs bg-white p-2 rounded">
                            <div className="flex flex-col">
                              <span className="font-mono text-sm">{card.card_code}</span>
                              <span className="text-gray-500 text-xs">原始: {card.original_points.toLocaleString()}积分</span>
                            </div>
                            <div className="text-right">
                              {card.status === "fully_consumed" ? (
                                <span className="text-red-600 font-medium">已全部消费</span>
                              ) : card.consumed_points > 0 ? (
                                <div className="text-orange-600">
                                  <div>消费: {card.consumed_points.toLocaleString()}</div>
                                  <div className="text-green-600">剩余: {card.remaining_points.toLocaleString()}</div>
                                </div>
                              ) : (
                                <span className="text-green-600">剩余: {card.remaining_points.toLocaleString()}</span>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* 权益变化预览 */}
                  {banPreviewData.benefits_change && (
                    <div className="bg-orange-50 p-4 rounded-lg border border-orange-200 mt-4">
                      <h4 className="font-medium text-orange-900 mb-3">
                        ⚖️ 权益变化预览
                      </h4>
                      
                      {/* 特殊警告：无剩余卡密 */}
                      {banPreviewData.benefits_change.will_reset_to_initial && (
                        <div className="bg-red-100 p-3 rounded-lg border border-red-300 mb-4">
                          <div className="flex items-center gap-2 text-red-800 mb-2">
                            <span className="text-lg">⚠️</span>
                            <span className="font-medium">重要警告</span>
                          </div>
                          <p className="text-red-700 text-sm font-medium">
                            封禁此激活码后，用户将没有任何有效的激活码，所有权益将重置为初始状态（即钱包过期状态）！
                          </p>
                        </div>
                      )}

                      <div className="bg-white rounded-lg border overflow-hidden">
                        <div className="px-4 py-3 bg-blue-50 border-b border-blue-200">
                          <h5 className="font-semibold text-blue-800 text-sm">权益对比详情</h5>
                        </div>
                        <div className="p-4">
                          <div className="grid grid-cols-3 gap-4 text-sm">
                            {/* 表头 */}
                            <div className="font-medium text-gray-600">权益项目</div>
                            <div className="font-medium text-green-600 text-center">当前状态</div>
                            <div className="font-medium text-blue-600 text-center">封禁后状态</div>
                            
                            {/* 分隔线 */}
                            <div className="col-span-3 border-t border-gray-200 my-2"></div>
                            
                            {/* 每日限额 */}
                            <div className="text-gray-700">📊 每日限额</div>
                            <div className="text-center text-green-600 font-medium">
                              {banPreviewData.benefits_change?.current_benefits?.daily_max_points > 0 
                                ? `${banPreviewData.benefits_change.current_benefits.daily_max_points.toLocaleString()}积分` 
                                : '无限制'}
                            </div>
                            <div className="text-center font-medium">
                              {banPreviewData.benefits_change?.new_benefits?.daily_max_points > 0 
                                ? `${banPreviewData.benefits_change.new_benefits.daily_max_points.toLocaleString()}积分` 
                                : (
                                  <span className="text-red-600">无限制</span>
                                )}
                            </div>
                            
                            {/* 签到奖励 */}
                            <div className="text-gray-700">🎁 签到奖励</div>
                            <div className="text-center text-green-600 font-medium">
                              {banPreviewData.benefits_change?.current_benefits?.daily_checkin_points > 0 ? (
                                banPreviewData.benefits_change.current_benefits.daily_checkin_points === banPreviewData.benefits_change.current_benefits.daily_checkin_points_max ? 
                                  `${banPreviewData.benefits_change.current_benefits.daily_checkin_points}积分` : 
                                  `${banPreviewData.benefits_change.current_benefits.daily_checkin_points}-${banPreviewData.benefits_change.current_benefits.daily_checkin_points_max}积分`
                              ) : '未配置'}
                            </div>
                            <div className="text-center font-medium">
                              {banPreviewData.benefits_change?.new_benefits?.daily_checkin_points > 0 ? (
                                banPreviewData.benefits_change.new_benefits.daily_checkin_points === banPreviewData.benefits_change.new_benefits.daily_checkin_points_max ? 
                                  `${banPreviewData.benefits_change.new_benefits.daily_checkin_points}积分` : 
                                  `${banPreviewData.benefits_change.new_benefits.daily_checkin_points}-${banPreviewData.benefits_change.new_benefits.daily_checkin_points_max}积分`
                              ) : (
                                <span className="text-red-600">未配置</span>
                              )}
                            </div>
                            
                            {/* 自动补给 */}
                            <div className="text-gray-700">🔄 自动补给</div>
                            <div className="text-center text-green-600 font-medium">
                              {banPreviewData.benefits_change?.current_benefits?.auto_refill_enabled 
                                ? `${banPreviewData.benefits_change.current_benefits.auto_refill_amount}积分` 
                                : '未启用'}
                            </div>
                            <div className="text-center font-medium">
                              {banPreviewData.benefits_change?.new_benefits?.auto_refill_enabled 
                                ? `${banPreviewData.benefits_change.new_benefits.auto_refill_amount}积分` 
                                : (
                                  <span className="text-red-600">未启用</span>
                                )}
                            </div>
                            
                            {/* 降智保护 */}
                            <div className="text-gray-700">🛡️ 降智保护</div>
                            <div className="text-center text-green-600 font-medium">
                              {banPreviewData.benefits_change?.current_benefits?.degradation_guaranteed > 0 
                                ? `等级 ${banPreviewData.benefits_change.current_benefits.degradation_guaranteed}` 
                                : '无保护'}
                            </div>
                            <div className="text-center font-medium">
                              {banPreviewData.benefits_change?.new_benefits?.degradation_guaranteed > 0 
                                ? `等级 ${banPreviewData.benefits_change.new_benefits.degradation_guaranteed}` 
                                : (
                                  <span className="text-red-600">无保护</span>
                                )}
                            </div>
                          </div>
                        </div>
                      </div>
                      
                      {/* 权益变化说明 */}
                      <div className="mt-3 text-xs text-blue-700 bg-blue-100 p-2 rounded">
                        <p className="font-medium mb-1">权益计算说明：</p>
                        <ul className="space-y-1 ml-2">
                          {banPreviewData.benefits_change.has_remaining_cards ? (
                            <li>• 系统将基于剩余有效激活码重新计算最优权益配置</li>
                          ) : (
                            <li>• 封禁后用户将没有任何有效激活码，权益将重置为初始状态</li>
                          )}
                          <li>• 权益变化在封禁操作执行后立即生效</li>
                        </ul>
                      </div>
                    </div>
                  )}
                </div>
              )}
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsBanPreviewDialogOpen(false)}>关闭预览</Button>
                <Button 
                  onClick={() => {
                    setIsBanPreviewDialogOpen(false)
                    if (selectedCodeForBan) {
                      setIsBanDialogOpen(true)
                    }
                  }}
                  className="bg-red-600 hover:bg-red-700 text-white"
                >
                  <Ban className="mr-2 h-4 w-4" />
                  继续封禁
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* 封禁操作对话框 */}
          <Dialog open={isBanDialogOpen} onOpenChange={setIsBanDialogOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle className="flex items-center gap-2 text-red-600">
                  <AlertTriangle className="h-5 w-5" />
                  确认封禁激活码
                </DialogTitle>
                <DialogDescription>
                  此操作将封禁激活码 &quot;{selectedCodeForBan?.code}&quot; 并冻结相应的积分，操作不可撤销。
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="bg-red-50 p-4 rounded-lg border border-red-200">
                  <div className="flex items-center gap-2 text-red-800 mb-2">
                    <AlertTriangle className="h-4 w-4" />
                    <span className="font-medium">警告</span>
                  </div>
                  <p className="text-red-700 text-sm">
                    封禁操作将：
                  </p>
                  <ul className="text-red-700 text-sm mt-1 ml-4 list-disc">
                    <li>立即冻结该激活码对应的剩余积分</li>
                    <li>降级用户的相关权益</li>
                    <li>创建不可逆的冻结记录</li>
                  </ul>
                </div>
                
                <div className="space-y-2">
                  <Label>封禁原因 (可选)</Label>
                  <Input
                    value={banReason}
                    onChange={(e) => setBanReason(e.target.value)}
                    placeholder="请输入封禁原因..."
                    maxLength={200}
                  />
                  <p className="text-xs text-muted-foreground">
                    封禁原因将被记录在冻结记录中
                  </p>
                </div>
                
                {selectedCodeForBan && (
                  <div className="bg-muted p-3 rounded text-sm">
                    <p className="font-medium mb-1">即将封禁的激活码：</p>
                    <p>激活码: <span className="font-mono">{selectedCodeForBan.code}</span></p>
                    <p>使用者: {selectedCodeForBan.used_by_username}</p>
                    <p>关联计划: {selectedCodeForBan.plan_title}</p>
                  </div>
                )}
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsBanDialogOpen(false)}>取消</Button>
                <Button 
                  onClick={handleExecuteBan}
                  className="bg-red-600 hover:bg-red-700 text-white"
                  disabled={loading}
                >
                  {loading ? "封禁中..." : "确认封禁"}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>
    </DashboardLayout>
  )
} 