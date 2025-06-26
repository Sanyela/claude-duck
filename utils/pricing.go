package utils

import (
	"encoding/json"
	"math"
	"sort"
	"strconv"
)

// TokenPricingTable 阶梯积分扣费表类型
type TokenPricingTable map[string]int

// CalculatePointsByTokenTable 根据阶梯计费表计算积分
func CalculatePointsByTokenTable(totalTokens float64, pricingTableJSON string) int64 {
	// 解析JSON配置
	var pricingTable TokenPricingTable
	if err := json.Unmarshal([]byte(pricingTableJSON), &pricingTable); err != nil {
		// 如果解析失败，返回默认计费（按5000 tokens = 1积分）
		return int64(math.Ceil(totalTokens / 5000))
	}

	// 将map的key转换为数字并排序
	var thresholds []int
	for thresholdStr := range pricingTable {
		threshold, err := strconv.Atoi(thresholdStr)
		if err != nil {
			continue
		}
		thresholds = append(thresholds, threshold)
	}
	sort.Ints(thresholds)

	// 如果没有有效的阈值，使用默认计费
	if len(thresholds) == 0 {
		return int64(math.Ceil(totalTokens / 5000))
	}

	// 找到对应的积分值
	totalTokensInt := int(totalTokens)
	points := pricingTable[strconv.Itoa(thresholds[0])] // 默认使用最低档

	for i := len(thresholds) - 1; i >= 0; i-- {
		threshold := thresholds[i]
		if totalTokensInt >= threshold {
			points = pricingTable[strconv.Itoa(threshold)]
			break
		}
	}

	return int64(points)
}

// GetTokenPricingInfo 获取计费信息，用于调试
func GetTokenPricingInfo(totalTokens float64, pricingTableJSON string) (int64, string) {
	points := CalculatePointsByTokenTable(totalTokens, pricingTableJSON)

	// 生成调试信息
	debugInfo := ""

	var pricingTable TokenPricingTable
	if err := json.Unmarshal([]byte(pricingTableJSON), &pricingTable); err != nil {
		debugInfo = "使用默认计费规则（解析失败）"
	} else {
		var thresholds []int
		for thresholdStr := range pricingTable {
			threshold, err := strconv.Atoi(thresholdStr)
			if err != nil {
				continue
			}
			thresholds = append(thresholds, threshold)
		}
		sort.Ints(thresholds)

		totalTokensInt := int(totalTokens)
		for i := len(thresholds) - 1; i >= 0; i-- {
			threshold := thresholds[i]
			if totalTokensInt >= threshold {
				debugInfo = "命中阶梯: " + strconv.Itoa(threshold) + " tokens -> " + strconv.FormatInt(points, 10) + " 积分"
				break
			}
		}
	}

	return points, debugInfo
}
