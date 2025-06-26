"use client"

import { useState, useEffect } from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Loader2, Calculator, Info } from "lucide-react"
import { creditsAPI, type PricingTable } from "@/api/credits"

interface PricingTableModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  tokenCount?: number // 当前的token数量，用于高亮对应档位
}

export function PricingTableModal({ open, onOpenChange, tokenCount }: PricingTableModalProps) {
  const [pricingTable, setPricingTable] = useState<PricingTable | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 加载计费表数据
  const loadPricingTable = async () => {
    setLoading(true)
    setError(null)

    try {
      const result = await creditsAPI.getPricingTable()
      if (result.success && result.data) {
        setPricingTable(result.data)
      } else {
        setError(result.message || "获取计费表失败")
      }
    } catch (err: any) {
      setError("获取计费表失败")
    }

    setLoading(false)
  }

  useEffect(() => {
    if (open) {
      loadPricingTable()
    }
  }, [open])

  // 将计费表转换为排序的数组
  const getSortedPricingData = () => {
    if (!pricingTable) return []

    return Object.entries(pricingTable.pricing_table)
      .map(([threshold, points]) => ({
        threshold: parseInt(threshold),
        points: points,
      }))
      .sort((a, b) => a.threshold - b.threshold)
  }

  // 判断某个档位是否为当前token数量对应的档位
  const isCurrentTier = (threshold: number, nextThreshold?: number) => {
    if (tokenCount === undefined) return false
    
    if (nextThreshold === undefined) {
      // 最后一个档位
      return tokenCount >= threshold
    } else {
      // 中间档位
      return tokenCount >= threshold && tokenCount < nextThreshold
    }
  }

  const sortedData = getSortedPricingData()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Calculator className="h-5 w-5 text-blue-500" />
            积分计费表
          </DialogTitle>
          <DialogDescription>
            基于加权Token总数的阶梯计费表，Token数量越多，每Token消耗的积分越高
          </DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin mr-2" />
            <span>加载中...</span>
          </div>
        ) : error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : pricingTable ? (
          <div className="space-y-4">
            {/* 说明信息 */}
            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription>
                {pricingTable.description}
                {tokenCount !== undefined && (
                  <span className="block mt-2 font-medium">
                    当前Token数量: <span className="text-blue-600">{tokenCount.toLocaleString()}</span>
                  </span>
                )}
              </AlertDescription>
            </Alert>

            {/* 计费表 */}
            <div className="border rounded-lg overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow className="bg-gray-50 dark:bg-gray-800">
                    <TableHead className="font-semibold">Token阈值</TableHead>
                    <TableHead className="font-semibold">消耗积分</TableHead>
                    <TableHead className="font-semibold">说明</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sortedData.map((item, index) => {
                    const nextThreshold = sortedData[index + 1]?.threshold
                    const isCurrent = isCurrentTier(item.threshold, nextThreshold)
                    
                    return (
                      <TableRow 
                        key={item.threshold}
                        className={isCurrent ? "bg-blue-50 dark:bg-blue-900/20 border-blue-200" : ""}
                      >
                        <TableCell className="font-mono">
                          <div className="flex items-center gap-2">
                            <span>
                              {item.threshold.toLocaleString()}
                              {nextThreshold !== undefined && ` - ${(nextThreshold - 1).toLocaleString()}`}
                            </span>
                            {isCurrent && (
                              <Badge variant="default" className="bg-blue-500 text-white">
                                当前档位
                              </Badge>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          <span className="font-semibold text-red-600">
                            {item.points} 积分
                          </span>
                        </TableCell>
                        <TableCell className="text-sm text-gray-600 dark:text-gray-400">
                          {nextThreshold !== undefined 
                            ? `${item.threshold.toLocaleString()} ≤ Token < ${nextThreshold.toLocaleString()}`
                            : `Token ≥ ${item.threshold.toLocaleString()}`
                          }
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </div>

            {/* 计费说明 */}
            <div className="text-sm text-gray-600 dark:text-gray-400 space-y-2">
              <div className="font-medium">计费说明：</div>
              <ul className="list-disc list-inside space-y-1 ml-2">
                <li>系统根据加权Token总数查找对应的积分消耗</li>
                <li>加权Token = 输入Token×输入倍率 + 输出Token×输出倍率 + 缓存Token×缓存倍率</li>
                <li>Token数量越高，积分消耗越多（阶梯计费）</li>
                <li>每次API调用都会根据实际Token使用量进行计费</li>
              </ul>
            </div>
          </div>
        ) : null}

        <div className="flex justify-end">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            关闭
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
} 