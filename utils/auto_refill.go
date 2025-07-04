package utils

import (
	"fmt"
	"log"
	"time"

	"claude/database"
	"claude/models"
)

// StartAutoRefillScheduler 启动自动补给定时器
func StartAutoRefillScheduler() {
	log.Println("🚀 启动自动补给定时器...")

	// 立即执行一次检查
	go func() {
		if err := ExecuteAutoRefillCheck(); err != nil {
			log.Printf("❌ 自动补给检查失败: %v", err)
		}
	}()

	// 设置定时器，每分钟检查一次是否到达执行时间点
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		lastExecutedHour := -1 // 记录上次执行的小时，避免重复执行

		for range ticker.C {
			now := time.Now()
			hour := now.Hour()

			// 只在0点、4点、8点、12点、16点、20点执行补给
			if hour%4 == 0 && hour != lastExecutedHour {
				log.Printf("⏰ 开始执行自动补给检查 (时间: %s)", now.Format("2006-01-02 15:04:05"))
				if err := ExecuteAutoRefillCheck(); err != nil {
					log.Printf("❌ 自动补给检查失败: %v", err)
				}
				lastExecutedHour = hour // 记录本次执行的小时
			}
		}
	}()

	log.Println("✅ 自动补给定时器已启动，将在每天0点、4点、8点、12点、16点、20点执行检查")
}

// ExecuteAutoRefillCheck 执行自动补给检查
func ExecuteAutoRefillCheck() error {
	log.Println("🔍 开始执行自动补给检查...")

	// 查询所有启用了自动补给的用户钱包
	var wallets []models.UserWallet
	err := database.DB.Where("auto_refill_enabled = ? AND status = ?", true, "active").Find(&wallets).Error
	if err != nil {
		return fmt.Errorf("查询启用自动补给的钱包失败: %v", err)
	}

	if len(wallets) == 0 {
		log.Println("📋 没有启用自动补给的用户钱包")
		return nil
	}

	log.Printf("📋 找到 %d 个启用自动补给的用户钱包", len(wallets))

	refillCount := 0
	for _, wallet := range wallets {
		// 检查是否需要补给
		if wallet.AvailablePoints <= wallet.AutoRefillThreshold {
			// 检查是否已经在当前时间点补给过
			if wallet.LastAutoRefillTime != nil {
				now := time.Now()
				currentTimeSlot := now.Hour() - (now.Hour() % 4) // 当前时间段的起始小时 (0,4,8,12,16,20)

				// 如果上次补给时间在当前时间段内，则跳过
				lastRefillHour := wallet.LastAutoRefillTime.Hour()
				lastRefillTimeSlot := lastRefillHour - (lastRefillHour % 4)

				// 如果是同一天且在同一个时间段内补给过，跳过
				if wallet.LastAutoRefillTime.Format("2006-01-02") == now.Format("2006-01-02") &&
					lastRefillTimeSlot == currentTimeSlot {
					log.Printf("⏭️ 用户 %d 在当前时间段(%d点)已经补给过，跳过", wallet.UserID, currentTimeSlot)
					continue
				}
			}

			// 执行补给
			if err := executeAutoRefill(&wallet); err != nil {
				log.Printf("❌ 用户 %d 自动补给失败: %v", wallet.UserID, err)
				continue
			}

			refillCount++
			log.Printf("✅ 用户 %d 自动补给成功，补给积分: %d", wallet.UserID, wallet.AutoRefillAmount)
		}
	}

	log.Printf("🎉 自动补给检查完成，共补给 %d 个用户", refillCount)
	return nil
}

// executeAutoRefill 执行单个用户的自动补给
func executeAutoRefill(wallet *models.UserWallet) error {
	// 开始事务
	tx := database.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开始事务失败: %v", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 更新用户钱包积分
	now := time.Now()
	err := tx.Model(&models.UserWallet{}).
		Where("user_id = ?", wallet.UserID).
		Updates(map[string]interface{}{
			"available_points":      wallet.AvailablePoints + wallet.AutoRefillAmount,
			"total_points":          wallet.TotalPoints + wallet.AutoRefillAmount,
			"last_auto_refill_time": now,
		}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("更新用户钱包失败: %v", err)
	}

	// 2. 创建兑换记录
	redemptionRecord := models.RedemptionRecord{
		UserID:              wallet.UserID,
		SourceType:          "auto_refill",
		SourceID:            fmt.Sprintf("auto_refill_%d", time.Now().Unix()),
		PointsAmount:        wallet.AutoRefillAmount,
		ValidityDays:        365, // 自动补给的积分有效期1年
		AutoRefillEnabled:   wallet.AutoRefillEnabled,
		AutoRefillThreshold: wallet.AutoRefillThreshold,
		AutoRefillAmount:    wallet.AutoRefillAmount,
		ActivatedAt:         now,
		ExpiresAt:           now.AddDate(1, 0, 0), // 1年后过期
		Reason:              fmt.Sprintf("自动补给积分，触发条件：可用积分(%d) <= 阈值(%d)", wallet.AvailablePoints, wallet.AutoRefillThreshold),
	}

	err = tx.Create(&redemptionRecord).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("创建兑换记录失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetAutoRefillStatus 获取用户自动补给状态
func GetAutoRefillStatus(userID uint) (*models.UserWallet, error) {
	var wallet models.UserWallet
	err := database.DB.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, fmt.Errorf("获取用户钱包失败: %v", err)
	}

	return &wallet, nil
}

// UpdateAutoRefillConfig 更新用户自动补给配置
func UpdateAutoRefillConfig(userID uint, enabled bool, threshold, amount int64) error {
	err := database.DB.Model(&models.UserWallet{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"auto_refill_enabled":   enabled,
			"auto_refill_threshold": threshold,
			"auto_refill_amount":    amount,
		}).Error

	if err != nil {
		return fmt.Errorf("更新自动补给配置失败: %v", err)
	}

	return nil
}
