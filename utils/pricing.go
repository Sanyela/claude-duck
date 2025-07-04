package utils

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"claude/database"
	"claude/models"
)


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
