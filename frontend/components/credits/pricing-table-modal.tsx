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
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Loader2, Calculator, Info } from "lucide-react"
import { creditsAPI, type TokenThresholdConfig } from "@/api/credits"

interface PricingTableModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  tokenCount?: number // 当前的token数量，用于高亮对应档位
}

export function PricingTableModal({ open, onOpenChange, tokenCount }: PricingTableModalProps) {
  const [thresholdConfig, setThresholdConfig] = useState<TokenThresholdConfig | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 加载计费配置数据
  const loadThresholdConfig = async () => {
    setLoading(true)
    setError(null)

    try {
      const result = await creditsAPI.getPricingTable()
      if (result.success && result.data) {
        setThresholdConfig(result.data)
      } else {
        setError(result.message || "获取计费配置失败")
      }
    } catch {
      setError("获取计费配置失败")
    }

    setLoading(false)
  }

  useEffect(() => {
    if (open) {
      loadThresholdConfig()
    }
  }, [open])

  // 计算当前累计token可以扣费的次数
  const getDeductTimes = () => {
    if (!thresholdConfig || tokenCount === undefined) return 0
    return Math.floor(tokenCount / thresholdConfig.token_threshold)
  }

  // 计算当前累计token余量
  const getRemainingTokens = () => {
    if (!thresholdConfig || tokenCount === undefined) return 0
    return tokenCount % thresholdConfig.token_threshold
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Calculator className="h-5 w-5 text-blue-500" />
            累计Token计费配置
          </DialogTitle>
          <DialogDescription>
            基于累计加权Token的计费方式，避免小额token也扣费的问题
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
        ) : thresholdConfig ? (
          <div className="space-y-4">
            {/* 说明信息 */}
            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription>
                {thresholdConfig.description}
                {tokenCount !== undefined && (
                  <div className="mt-3 space-y-1">
                    <div className="font-medium">当前Token使用情况:</div>
                    <div className="text-sm space-y-1">
                      <div>总加权Token: <span className="text-blue-600 font-mono">{tokenCount.toLocaleString()}</span></div>
                      <div>可扣费次数: <span className="text-green-600 font-mono">{getDeductTimes()}</span> 次</div>
                      <div>总扣费积分: <span className="text-red-600 font-mono">{getDeductTimes() * thresholdConfig.points_per_threshold}</span> 积分</div>
                      <div>累计余量: <span className="text-orange-600 font-mono">{getRemainingTokens()}</span> Token</div>
                    </div>
                  </div>
                )}
              </AlertDescription>
            </Alert>

            {/* 计费配置 */}
            <div className="border rounded-lg overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow className="bg-gray-50 dark:bg-gray-800">
                    <TableHead className="font-semibold">配置项</TableHead>
                    <TableHead className="font-semibold">配置值</TableHead>
                    <TableHead className="font-semibold">说明</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow>
                    <TableCell className="font-medium">计费阈值</TableCell>
                    <TableCell className="font-mono text-blue-600">
                      {thresholdConfig.token_threshold.toLocaleString()} Token
                    </TableCell>
                    <TableCell className="text-sm text-gray-600 dark:text-gray-400">
                      累计达到此Token数量时进行扣费
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell className="font-medium">每阈值积分</TableCell>
                    <TableCell className="font-mono text-red-600">
                      {thresholdConfig.points_per_threshold} 积分
                    </TableCell>
                    <TableCell className="text-sm text-gray-600 dark:text-gray-400">
                      每达到一个阈值扣除的积分数量
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>

            {/* 计费说明 */}
            <div className="text-sm text-gray-600 dark:text-gray-400 space-y-2">
              <div className="font-medium">计费说明：</div>
              <ul className="list-disc list-inside space-y-1 ml-2">
                <li>系统累计用户的加权Token使用量</li>
                <li>加权Token = 输入Token×输入倍率 + 输出Token×输出倍率 + 缓存Token×缓存倍率</li>
                <li>当累计Token达到阈值时，扣除相应积分并重置计数器</li>
                <li>解决了小额Token也扣费的问题，只有累计到一定量才扣费</li>
                <li>支持一次性扣除多个阈值的积分（如累计10000token时扣除2积分）</li>
              </ul>
            </div>

            {/* 示例计算 */}
            {tokenCount !== undefined && (
              <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4">
                <div className="font-medium text-blue-900 dark:text-blue-100 mb-2">📊 当前计费示例</div>
                <div className="font-mono text-sm space-y-1">
                  <div>当前累计: {tokenCount.toLocaleString()} Token</div>
                  <div>扣费次数: {tokenCount.toLocaleString()} ÷ {thresholdConfig.token_threshold.toLocaleString()} = {getDeductTimes()} 次</div>
                  <div>扣费积分: {getDeductTimes()} × {thresholdConfig.points_per_threshold} = <span className="text-red-600 font-bold">{getDeductTimes() * thresholdConfig.points_per_threshold} 积分</span></div>
                  <div>剩余累计: {tokenCount.toLocaleString()} % {thresholdConfig.token_threshold.toLocaleString()} = <span className="text-orange-600">{getRemainingTokens()} Token</span></div>
                </div>
              </div>
            )}
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