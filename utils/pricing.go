package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"gorm.io/gorm"

	"claude/database"
	"claude/models"
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

// CheckDailyPointsLimit 检查用户当日积分使用是否超过限制
func CheckDailyPointsLimit(userID uint, subscriptionID uint, requestedPoints int64) (bool, int64, error) {
	today := time.Now().Format("2006-01-02")

	// 获取订阅信息
	var subscription models.Subscription
	err := database.DB.Preload("Plan").Where("id = ? AND user_id = ?", subscriptionID, userID).First(&subscription).Error
	if err != nil {
		return false, 0, fmt.Errorf("获取订阅信息失败: %v", err)
	}

	// 如果订阅或计划未设置每日限制，则不限制
	dailyLimit := subscription.DailyMaxPoints
	if dailyLimit == 0 {
		dailyLimit = subscription.Plan.DailyMaxPoints
	}
	if dailyLimit == 0 {
		return true, 0, nil // 无限制
	}

	// 获取今日已使用积分
	var dailyUsage models.DailyPointsUsage
	err = database.DB.Where("user_id = ? AND subscription_id = ? AND usage_date = ?",
		userID, subscriptionID, today).First(&dailyUsage).Error

	var usedToday int64 = 0
	if err == nil {
		usedToday = dailyUsage.PointsUsed
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, 0, fmt.Errorf("查询每日使用记录失败: %v", err)
	}

	// 检查是否会超过限制
	totalAfterRequest := usedToday + requestedPoints
	if totalAfterRequest > dailyLimit {
		remainingPoints := dailyLimit - usedToday
		if remainingPoints < 0 {
			remainingPoints = 0
		}
		return false, remainingPoints, nil
	}

	return true, dailyLimit - totalAfterRequest, nil
}

// UpdateDailyPointsUsage 更新用户当日积分使用记录
func UpdateDailyPointsUsage(userID uint, subscriptionID uint, pointsUsed int64) error {
	today := time.Now().Format("2006-01-02")

	// 使用事务确保数据一致性
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var dailyUsage models.DailyPointsUsage
		err := tx.Where("user_id = ? AND subscription_id = ? AND usage_date = ?",
			userID, subscriptionID, today).First(&dailyUsage).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			dailyUsage = models.DailyPointsUsage{
				UserID:         userID,
				SubscriptionID: subscriptionID,
				UsageDate:      today,
				PointsUsed:     pointsUsed,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			return tx.Create(&dailyUsage).Error
		} else if err != nil {
			return fmt.Errorf("查询每日使用记录失败: %v", err)
		} else {
			// 更新现有记录
			return tx.Model(&dailyUsage).Update("points_used", gorm.Expr("points_used + ?", pointsUsed)).Error
		}
	})
}

// GetUserDailyPointsUsage 获取用户今日积分使用情况
func GetUserDailyPointsUsage(userID uint) ([]models.DailyPointsUsage, error) {
	today := time.Now().Format("2006-01-02")

	var usages []models.DailyPointsUsage
	err := database.DB.Preload("Subscription").Preload("Subscription.Plan").
		Where("user_id = ? AND usage_date = ?", userID, today).
		Find(&usages).Error

	return usages, err
}
