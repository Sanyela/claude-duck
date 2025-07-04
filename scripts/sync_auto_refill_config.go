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
	log.Println("🚀 开始同步用户钱包自动补给配置...")

	// 加载配置
	config.LoadConfig()

	// 连接数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	log.Println("✅ 数据库连接成功")

	// 执行同步
	if err := syncAutoRefillConfig(); err != nil {
		log.Fatalf("❌ 同步自动补给配置失败: %v", err)
	}

	log.Println("🎉 同步完成！")
}

// syncAutoRefillConfig 同步自动补给配置到用户钱包
func syncAutoRefillConfig() error {
	log.Println("🔍 开始查询需要同步的用户钱包...")

	// 查询所有有效的用户钱包及其最新的兑换记录
	var wallets []models.UserWallet
	err := database.DB.Where("status = ?", "active").Find(&wallets).Error
	if err != nil {
		return fmt.Errorf("查询用户钱包失败: %v", err)
	}

	log.Printf("📋 找到 %d 个有效用户钱包", len(wallets))

	syncCount := 0
	for _, wallet := range wallets {
		// 如果钱包已经启用了自动补给，跳过
		if wallet.AutoRefillEnabled {
			log.Printf("⏭️ 用户 %d 已启用自动补给，跳过", wallet.UserID)
			continue
		}

		// 查找该用户最新的兑换记录，优先查找来自订阅计划的记录
		var latestRecord models.RedemptionRecord
		err := database.DB.Preload("SubscriptionPlan").
			Where("user_id = ? AND expires_at > ?", wallet.UserID, time.Now()).
			Order("created_at DESC").
			First(&latestRecord).Error

		if err != nil {
			log.Printf("⏭️ 用户 %d 没有有效的兑换记录，跳过", wallet.UserID)
			continue
		}

		// 检查是否有订阅计划，且订阅计划启用了自动补给
		if latestRecord.SubscriptionPlan == nil || !latestRecord.SubscriptionPlan.AutoRefillEnabled {
			log.Printf("⏭️ 用户 %d 的订阅计划未启用自动补给，跳过", wallet.UserID)
			continue
		}

		plan := latestRecord.SubscriptionPlan

		// 更新用户钱包的自动补给配置
		updates := map[string]interface{}{
			"auto_refill_enabled":   plan.AutoRefillEnabled,
			"auto_refill_threshold": plan.AutoRefillThreshold,
			"auto_refill_amount":    plan.AutoRefillAmount,
			"updated_at":            time.Now(),
		}

		err = database.DB.Model(&models.UserWallet{}).
			Where("user_id = ?", wallet.UserID).
			Updates(updates).Error

		if err != nil {
			log.Printf("❌ 更新用户 %d 自动补给配置失败: %v", wallet.UserID, err)
			continue
		}

		syncCount++
		log.Printf("✅ 用户 %d 自动补给配置同步成功 (阈值: %d, 补给量: %d)", 
			wallet.UserID, plan.AutoRefillThreshold, plan.AutoRefillAmount)
	}

	log.Printf("🎉 同步完成，共同步 %d 个用户的自动补给配置", syncCount)
	return nil
}

// syncSpecificPlan 同步特定订阅计划的自动补给配置到用户钱包
func syncSpecificPlan(planID uint) error {
	log.Printf("🔍 开始同步订阅计划 %d 的自动补给配置...", planID)

	// 查询订阅计划
	var plan models.SubscriptionPlan
	err := database.DB.Where("id = ?", planID).First(&plan).Error
	if err != nil {
		return fmt.Errorf("查询订阅计划失败: %v", err)
	}

	if !plan.AutoRefillEnabled {
		log.Printf("⏭️ 订阅计划 %d 未启用自动补给，跳过", planID)
		return nil
	}

	// 查询使用该订阅计划的用户钱包
	var records []models.RedemptionRecord
	err = database.DB.Where("subscription_plan_id = ? AND expires_at > ?", planID, time.Now()).
		Find(&records).Error
	if err != nil {
		return fmt.Errorf("查询兑换记录失败: %v", err)
	}

	log.Printf("📋 找到 %d 个使用该订阅计划的兑换记录", len(records))

	syncCount := 0
	for _, record := range records {
		// 更新用户钱包的自动补给配置
		updates := map[string]interface{}{
			"auto_refill_enabled":   plan.AutoRefillEnabled,
			"auto_refill_threshold": plan.AutoRefillThreshold,
			"auto_refill_amount":    plan.AutoRefillAmount,
			"updated_at":            time.Now(),
		}

		err = database.DB.Model(&models.UserWallet{}).
			Where("user_id = ?", record.UserID).
			Updates(updates).Error

		if err != nil {
			log.Printf("❌ 更新用户 %d 自动补给配置失败: %v", record.UserID, err)
			continue
		}

		syncCount++
		log.Printf("✅ 用户 %d 自动补给配置同步成功", record.UserID)
	}

	log.Printf("🎉 订阅计划 %d 同步完成，共同步 %d 个用户", planID, syncCount)
	return nil
}