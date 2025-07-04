package utils

import (
	"fmt"
	"strconv"
	"time"

	"claude/database"
	"claude/models"

	"gorm.io/gorm"
)

// GetUserWallet 获取用户钱包信息
func GetUserWallet(userID uint) (*models.UserWallet, error) {
	var wallet models.UserWallet
	err := database.DB.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, fmt.Errorf("获取用户钱包失败: %v", err)
	}
	return &wallet, nil
}

// GetOrCreateUserWallet 获取或创建用户钱包
func GetOrCreateUserWallet(userID uint) (*models.UserWallet, error) {
	var wallet models.UserWallet
	err := database.DB.Where("user_id = ?", userID).First(&wallet).Error
	if err == nil {
		return &wallet, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("查询用户钱包失败: %v", err)
	}

	// 创建新钱包
	wallet = models.UserWallet{
		UserID:          userID,
		TotalPoints:     0,
		AvailablePoints: 0,
		UsedPoints:      0,
		WalletExpiresAt: time.Now(),
		Status:          "expired",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := database.DB.Create(&wallet).Error; err != nil {
		return nil, fmt.Errorf("创建用户钱包失败: %v", err)
	}

	return &wallet, nil
}

// UpdateWalletPoints 更新钱包积分（增加）
func UpdateWalletPoints(userID uint, points int64) error {
	if points <= 0 {
		return nil
	}

	return database.DB.Model(&models.UserWallet{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"total_points":     gorm.Expr("total_points + ?", points),
			"available_points": gorm.Expr("available_points + ?", points),
			"updated_at":       time.Now(),
		}).Error
}

// DeductWalletPoints 扣除钱包积分
func DeductWalletPoints(userID uint, points int64) error {
	if points <= 0 {
		return nil
	}

	// 检查余额
	wallet, err := GetUserWallet(userID)
	if err != nil {
		return err
	}

	if wallet.AvailablePoints < points {
		return fmt.Errorf("积分余额不足，需要 %d 积分，可用 %d 积分", points, wallet.AvailablePoints)
	}

	// 扣除积分
	return database.DB.Model(&models.UserWallet{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"available_points": gorm.Expr("available_points - ?", points),
			"used_points":      gorm.Expr("used_points + ?", points),
			"updated_at":       time.Now(),
		}).Error
}

// CheckDailyLimit 检查每日积分使用限制
func CheckDailyLimit(userID uint, pointsToUse int64) error {
	if pointsToUse <= 0 {
		return nil
	}

	wallet, err := GetUserWallet(userID)
	if err != nil {
		return err
	}

	// 如果没有每日限制，直接返回
	if wallet.DailyMaxPoints <= 0 {
		return nil
	}

	// 获取今日已使用积分
	today := time.Now().Format("2006-01-02")
	var dailyUsage models.UserDailyUsage
	err = database.DB.Where("user_id = ? AND usage_date = ?", userID, today).First(&dailyUsage).Error

	var usedToday int64 = 0
	if err == nil {
		usedToday = dailyUsage.PointsUsed
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("查询每日使用记录失败: %v", err)
	}

	// 检查今日剩余限制
	remainingDaily := wallet.DailyMaxPoints - usedToday
	if remainingDaily < pointsToUse {
		return fmt.Errorf("每日积分使用限制不足，今日剩余 %d 积分，需要 %d 积分", remainingDaily, pointsToUse)
	}

	return nil
}

// UpdateDailyUsage 更新每日使用记录
func UpdateDailyUsage(userID uint, pointsUsed int64) error {
	if pointsUsed <= 0 {
		return nil
	}

	today := time.Now().Format("2006-01-02")

	// 查找今日记录
	var dailyUsage models.UserDailyUsage
	err := database.DB.Where("user_id = ? AND usage_date = ?", userID, today).First(&dailyUsage).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		dailyUsage = models.UserDailyUsage{
			UserID:     userID,
			UsageDate:  today,
			PointsUsed: pointsUsed,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		return database.DB.Create(&dailyUsage).Error
	} else if err != nil {
		return fmt.Errorf("查询每日使用记录失败: %v", err)
	}

	// 更新现有记录
	return database.DB.Model(&dailyUsage).
		Updates(map[string]interface{}{
			"points_used": gorm.Expr("points_used + ?", pointsUsed),
			"updated_at":  time.Now(),
		}).Error
}

// DeductWalletPointsWithDailyLimit 扣除钱包积分并检查每日限制
func DeductWalletPointsWithDailyLimit(userID uint, points int64) error {
	if points <= 0 {
		return nil
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查每日限制
	if err := CheckDailyLimit(userID, points); err != nil {
		tx.Rollback()
		return err
	}

	// 扣除积分
	if err := DeductWalletPoints(userID, points); err != nil {
		tx.Rollback()
		return err
	}

	// 更新每日使用记录
	if err := UpdateDailyUsage(userID, points); err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// IsWalletActive 检查钱包是否有效
func IsWalletActive(userID uint) bool {
	wallet, err := GetUserWallet(userID)
	if err != nil {
		return false
	}

	return wallet.Status == "active" &&
		wallet.AvailablePoints > 0 &&
		wallet.WalletExpiresAt.After(time.Now())
}

// GetWalletBalance 获取钱包余额信息
func GetWalletBalance(userID uint) (available, total, used int64, err error) {
	wallet, err := GetUserWallet(userID)
	if err != nil {
		return 0, 0, 0, err
	}

	return wallet.AvailablePoints, wallet.TotalPoints, wallet.UsedPoints, nil
}

// GetUserDailyUsage 获取用户今日使用情况
func GetUserDailyUsage(userID uint) (used int64, limit int64, err error) {
	// 获取钱包信息
	wallet, err := GetUserWallet(userID)
	if err != nil {
		return 0, 0, err
	}

	// 获取今日使用记录
	today := time.Now().Format("2006-01-02")
	var dailyUsage models.UserDailyUsage
	err = database.DB.Where("user_id = ? AND usage_date = ?", userID, today).First(&dailyUsage).Error

	if err == gorm.ErrRecordNotFound {
		return 0, wallet.DailyMaxPoints, nil
	} else if err != nil {
		return 0, 0, fmt.Errorf("查询每日使用记录失败: %v", err)
	}

	return dailyUsage.PointsUsed, wallet.DailyMaxPoints, nil
}

// UpdateWalletStatus 更新钱包状态
func UpdateWalletStatus(userID uint) error {
	wallet, err := GetUserWallet(userID)
	if err != nil {
		return err
	}

	var newStatus string
	if wallet.WalletExpiresAt.After(time.Now()) {
		newStatus = "active"
	} else {
		newStatus = "expired"
	}

	if wallet.Status != newStatus {
		return database.DB.Model(&models.UserWallet{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"status":     newStatus,
				"updated_at": time.Now(),
			}).Error
	}

	return nil
}

// GetTokenThresholdConfig 获取token阈值配置
func GetTokenThresholdConfig() (int64, int64, error) {
	var configs []models.SystemConfig
	err := database.DB.Where("config_key IN (?)", []string{"token_threshold", "points_per_threshold"}).Find(&configs).Error
	if err != nil {
		return 0, 0, fmt.Errorf("获取token阈值配置失败: %v", err)
	}

	var threshold int64 = 5000       // 默认阈值
	var pointsPerThreshold int64 = 1 // 默认每阈值积分

	for _, config := range configs {
		if config.ConfigKey == "token_threshold" {
			if val, err := strconv.ParseInt(config.ConfigValue, 10, 64); err == nil {
				threshold = val
			}
		} else if config.ConfigKey == "points_per_threshold" {
			if val, err := strconv.ParseInt(config.ConfigValue, 10, 64); err == nil {
				pointsPerThreshold = val
			}
		}
	}

	return threshold, pointsPerThreshold, nil
}

// AccumulateTokensAndDeduct 累计tokens并在达到阈值时扣费
func AccumulateTokensAndDeduct(userID uint, weightedTokens int64) error {
	if weightedTokens <= 0 {
		return nil
	}

	// 获取阈值配置
	threshold, pointsPerThreshold, err := GetTokenThresholdConfig()
	if err != nil {
		return err
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取或创建用户钱包（使用事务）
	var wallet models.UserWallet
	err = tx.Where("user_id = ?", userID).First(&wallet).Error
	if err == gorm.ErrRecordNotFound {
		// 创建新钱包
		wallet = models.UserWallet{
			UserID:            userID,
			TotalPoints:       0,
			AvailablePoints:   0,
			UsedPoints:        0,
			AccumulatedTokens: 0,
			WalletExpiresAt:   time.Now(),
			Status:            "expired",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		if err := tx.Create(&wallet).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("创建用户钱包失败: %v", err)
		}
	} else if err != nil {
		tx.Rollback()
		return fmt.Errorf("获取用户钱包失败: %v", err)
	}

	// 累计tokens
	newAccumulatedTokens := wallet.AccumulatedTokens + weightedTokens

	// 检查是否达到扣费阈值
	if newAccumulatedTokens >= threshold {
		// 计算需要扣费的次数
		deductTimes := newAccumulatedTokens / threshold
		totalPointsToDeduct := deductTimes * pointsPerThreshold

		// 检查余额是否足够
		if wallet.AvailablePoints < totalPointsToDeduct {
			tx.Rollback()
			return fmt.Errorf("积分余额不足，需要 %d 积分，可用 %d 积分", totalPointsToDeduct, wallet.AvailablePoints)
		}

		// 检查每日限制
		if err := CheckDailyLimit(userID, totalPointsToDeduct); err != nil {
			tx.Rollback()
			return err
		}

		// 扣除积分
		err = tx.Model(&models.UserWallet{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"available_points":   gorm.Expr("available_points - ?", totalPointsToDeduct),
				"used_points":        gorm.Expr("used_points + ?", totalPointsToDeduct),
				"accumulated_tokens": newAccumulatedTokens % threshold, // 重置累计tokens
				"updated_at":         time.Now(),
			}).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("扣除积分失败: %v", err)
		}

		// 更新每日使用记录
		if err := UpdateDailyUsage(userID, totalPointsToDeduct); err != nil {
			tx.Rollback()
			return err
		}

	} else {
		// 未达到阈值，只累计tokens
		err = tx.Model(&models.UserWallet{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"accumulated_tokens": newAccumulatedTokens,
				"updated_at":         time.Now(),
			}).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("累计tokens失败: %v", err)
		}
	}

	// 提交事务
	return tx.Commit().Error
}
