package main

import (
	"fmt"
	"log"
	"time"

	"claude/config"
	"claude/database"
	"claude/models"
)

func main() {
	log.Println("🚀 开始执行改进的自动补给配置同步...")

	// 加载配置
	config.LoadConfig()

	// 连接数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	log.Println("✅ 数据库连接成功")

	// 执行同步
	if err := syncAutoRefillConfigImproved(); err != nil {
		log.Fatalf("❌ 同步自动补给配置失败: %v", err)
	}

	log.Println("🎉 同步完成！")
}

// UserAutoRefillConfig 用户自动补给配置
type UserAutoRefillConfig struct {
	UserID              uint
	AutoRefillEnabled   bool
	AutoRefillThreshold int64
	AutoRefillAmount    int64
	Source              string // "activation_code" 或 "redemption_record"
	SourceID            string
	PlanTitle           string
	LastUpdateTime      time.Time
}

// syncAutoRefillConfigImproved 改进的自动补给配置同步
func syncAutoRefillConfigImproved() error {
	log.Println("🔍 开始查询需要同步的用户...")

	// 查询所有有效的用户钱包
	var wallets []models.UserWallet
	err := database.DB.Where("status = ?", "active").Find(&wallets).Error
	if err != nil {
		return fmt.Errorf("查询用户钱包失败: %v", err)
	}

	log.Printf("📋 找到 %d 个有效用户钱包", len(wallets))

	syncCount := 0
	skipCount := 0
	errorCount := 0

	for _, wallet := range wallets {
		log.Printf("🔄 处理用户 %d...", wallet.UserID)

		// 获取用户的自动补给配置
		config, err := getUserAutoRefillConfig(wallet.UserID)
		if err != nil {
			log.Printf("❌ 获取用户 %d 配置失败: %v", wallet.UserID, err)
			errorCount++
			continue
		}

		if config == nil {
			log.Printf("⏭️ 用户 %d 没有启用自动补给的有效订阅，跳过", wallet.UserID)
			skipCount++
			continue
		}

		// 检查是否需要更新
		if wallet.AutoRefillEnabled == config.AutoRefillEnabled &&
			wallet.AutoRefillThreshold == config.AutoRefillThreshold &&
			wallet.AutoRefillAmount == config.AutoRefillAmount {
			log.Printf("⏭️ 用户 %d 配置已是最新，跳过", wallet.UserID)
			skipCount++
			continue
		}

		// 更新用户钱包配置
		updates := map[string]interface{}{
			"auto_refill_enabled":   config.AutoRefillEnabled,
			"auto_refill_threshold": config.AutoRefillThreshold,
			"auto_refill_amount":    config.AutoRefillAmount,
			"updated_at":            time.Now(),
		}

		err = database.DB.Model(&models.UserWallet{}).
			Where("user_id = ?", wallet.UserID).
			Updates(updates).Error

		if err != nil {
			log.Printf("❌ 更新用户 %d 配置失败: %v", wallet.UserID, err)
			errorCount++
			continue
		}

		syncCount++
		log.Printf("✅ 用户 %d 配置同步成功", wallet.UserID)
		log.Printf("   来源: %s (%s)", config.Source, config.PlanTitle)
		log.Printf("   启用: %v, 阈值: %d, 补给量: %d", 
			config.AutoRefillEnabled, config.AutoRefillThreshold, config.AutoRefillAmount)
	}

	log.Printf("🎉 同步完成！")
	log.Printf("   成功同步: %d 个用户", syncCount)
	log.Printf("   跳过: %d 个用户", skipCount)
	log.Printf("   失败: %d 个用户", errorCount)

	return nil
}

// getUserAutoRefillConfig 获取用户的自动补给配置
func getUserAutoRefillConfig(userID uint) (*UserAutoRefillConfig, error) {
	// 1. 首先查找RedemptionRecord表中的记录
	var redemptionRecord models.RedemptionRecord
	err := database.DB.Preload("SubscriptionPlan").
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		First(&redemptionRecord).Error

	if err == nil && redemptionRecord.SubscriptionPlan != nil && redemptionRecord.SubscriptionPlan.AutoRefillEnabled {
		plan := redemptionRecord.SubscriptionPlan
		return &UserAutoRefillConfig{
			UserID:              userID,
			AutoRefillEnabled:   plan.AutoRefillEnabled,
			AutoRefillThreshold: plan.AutoRefillThreshold,
			AutoRefillAmount:    plan.AutoRefillAmount,
			Source:              "redemption_record",
			SourceID:            fmt.Sprintf("%d", redemptionRecord.ID),
			PlanTitle:           plan.Title,
			LastUpdateTime:      redemptionRecord.UpdatedAt,
		}, nil
	}

	// 2. 如果RedemptionRecord没有找到，查找activation_codes表
	var activationCode struct {
		ID                   uint      `json:"id"`
		Code                 string    `json:"code"`
		SubscriptionPlanID   uint      `json:"subscription_plan_id"`
		UsedAt               time.Time `json:"used_at"`
		PlanTitle            string    `json:"plan_title"`
		AutoRefillEnabled    bool      `json:"auto_refill_enabled"`
		AutoRefillThreshold  int64     `json:"auto_refill_threshold"`
		AutoRefillAmount     int64     `json:"auto_refill_amount"`
	}

	// 查找用户最新使用的启用自动补给的激活码
	err = database.DB.Raw(`
		SELECT ac.id, ac.code, ac.subscription_plan_id, ac.used_at,
			   sp.title as plan_title, sp.auto_refill_enabled, 
			   sp.auto_refill_threshold, sp.auto_refill_amount
		FROM activation_codes ac
		JOIN subscription_plans sp ON ac.subscription_plan_id = sp.id
		WHERE ac.used_by_user_id = ? 
		  AND ac.status = 'used' 
		  AND sp.auto_refill_enabled = 1
		ORDER BY ac.used_at DESC
		LIMIT 1
	`, userID).Scan(&activationCode).Error

	if err != nil {
		return nil, nil // 没有找到有效的自动补给配置
	}

	return &UserAutoRefillConfig{
		UserID:              userID,
		AutoRefillEnabled:   activationCode.AutoRefillEnabled,
		AutoRefillThreshold: activationCode.AutoRefillThreshold,
		AutoRefillAmount:    activationCode.AutoRefillAmount,
		Source:              "activation_code",
		SourceID:            activationCode.Code,
		PlanTitle:           activationCode.PlanTitle,
		LastUpdateTime:      activationCode.UsedAt,
	}, nil
}

// validateAutoRefillConfig 验证自动补给配置
func validateAutoRefillConfig() error {
	log.Println("🔍 开始验证自动补给配置...")

	// 查询所有启用自动补给的用户
	var wallets []models.UserWallet
	err := database.DB.Where("auto_refill_enabled = ? AND status = ?", true, "active").Find(&wallets).Error
	if err != nil {
		return fmt.Errorf("查询启用自动补给的用户失败: %v", err)
	}

	log.Printf("📋 找到 %d 个启用自动补给的用户", len(wallets))

	validCount := 0
	invalidCount := 0

	for _, wallet := range wallets {
		// 检查配置是否有效
		if wallet.AutoRefillThreshold <= 0 || wallet.AutoRefillAmount <= 0 {
			log.Printf("❌ 用户 %d 配置无效: 阈值=%d, 补给量=%d", 
				wallet.UserID, wallet.AutoRefillThreshold, wallet.AutoRefillAmount)
			invalidCount++
			continue
		}

		// 获取用户的原始配置源
		config, err := getUserAutoRefillConfig(wallet.UserID)
		if err != nil || config == nil {
			log.Printf("❌ 用户 %d 无法找到配置源", wallet.UserID)
			invalidCount++
			continue
		}

		// 验证配置是否一致
		if wallet.AutoRefillEnabled != config.AutoRefillEnabled ||
			wallet.AutoRefillThreshold != config.AutoRefillThreshold ||
			wallet.AutoRefillAmount != config.AutoRefillAmount {
			log.Printf("❌ 用户 %d 配置不一致", wallet.UserID)
			log.Printf("   钱包配置: 启用=%v, 阈值=%d, 补给量=%d", 
				wallet.AutoRefillEnabled, wallet.AutoRefillThreshold, wallet.AutoRefillAmount)
			log.Printf("   源配置: 启用=%v, 阈值=%d, 补给量=%d", 
				config.AutoRefillEnabled, config.AutoRefillThreshold, config.AutoRefillAmount)
			invalidCount++
			continue
		}

		validCount++
		log.Printf("✅ 用户 %d 配置验证通过 (来源: %s, 计划: %s)", 
			wallet.UserID, config.Source, config.PlanTitle)
	}

	log.Printf("🎉 验证完成！")
	log.Printf("   有效配置: %d 个", validCount)
	log.Printf("   无效配置: %d 个", invalidCount)

	return nil
}