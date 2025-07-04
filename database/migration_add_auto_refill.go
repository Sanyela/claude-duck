package database

import (
	"fmt"
	"log"
)

// MigrateAddAutoRefillFields 为订阅计划表和用户钱包表添加自动补给字段
func MigrateAddAutoRefillFields() error {
	log.Println("开始执行自动补给字段迁移...")

	// 检查subscription_plans表是否存在auto_refill字段
	var planAutoRefillExists bool
	checkPlanSQL := `
		SELECT COUNT(*) > 0 
		FROM information_schema.columns 
		WHERE table_schema = DATABASE() 
		AND table_name = 'subscription_plans' 
		AND column_name = 'auto_refill_enabled'
	`

	err := DB.Raw(checkPlanSQL).Scan(&planAutoRefillExists).Error
	if err != nil {
		return fmt.Errorf("检查subscription_plans auto_refill字段失败: %v", err)
	}

	if !planAutoRefillExists {
		// 添加自动补给字段到subscription_plans表
		alterPlanSQL := `
			ALTER TABLE subscription_plans 
			ADD COLUMN auto_refill_enabled BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否启用自动补给',
			ADD COLUMN auto_refill_threshold BIGINT NOT NULL DEFAULT 0 COMMENT '自动补给阈值，积分低于此值时触发',
			ADD COLUMN auto_refill_amount BIGINT NOT NULL DEFAULT 0 COMMENT '每次补给的积分数量'
		`

		err = DB.Exec(alterPlanSQL).Error
		if err != nil {
			return fmt.Errorf("添加auto_refill字段到subscription_plans表失败: %v", err)
		}

		log.Println("✅ 成功添加auto_refill字段到subscription_plans表")
	} else {
		log.Println("✅ subscription_plans auto_refill字段已存在")
	}

	// 检查user_wallets表是否存在auto_refill字段
	var walletAutoRefillExists bool
	checkWalletSQL := `
		SELECT COUNT(*) > 0 
		FROM information_schema.columns 
		WHERE table_schema = DATABASE() 
		AND table_name = 'user_wallets' 
		AND column_name = 'auto_refill_enabled'
	`

	err = DB.Raw(checkWalletSQL).Scan(&walletAutoRefillExists).Error
	if err != nil {
		return fmt.Errorf("检查user_wallets auto_refill字段失败: %v", err)
	}

	if !walletAutoRefillExists {
		// 添加自动补给字段到user_wallets表
		alterWalletSQL := `
			ALTER TABLE user_wallets 
			ADD COLUMN auto_refill_enabled BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否启用自动补给',
			ADD COLUMN auto_refill_threshold BIGINT NOT NULL DEFAULT 0 COMMENT '自动补给阈值',
			ADD COLUMN auto_refill_amount BIGINT NOT NULL DEFAULT 0 COMMENT '每次补给积分数量',
			ADD COLUMN last_auto_refill_time TIMESTAMP NULL COMMENT '最后一次自动补给时间'
		`

		err = DB.Exec(alterWalletSQL).Error
		if err != nil {
			return fmt.Errorf("添加auto_refill字段到user_wallets表失败: %v", err)
		}

		log.Println("✅ 成功添加auto_refill字段到user_wallets表")
	} else {
		log.Println("✅ user_wallets auto_refill字段已存在")
	}

	// 检查redemption_records表是否存在auto_refill字段
	var redemptionAutoRefillExists bool
	checkRedemptionSQL := `
		SELECT COUNT(*) > 0 
		FROM information_schema.columns 
		WHERE table_schema = DATABASE() 
		AND table_name = 'redemption_records' 
		AND column_name = 'auto_refill_enabled'
	`

	err = DB.Raw(checkRedemptionSQL).Scan(&redemptionAutoRefillExists).Error
	if err != nil {
		return fmt.Errorf("检查redemption_records auto_refill字段失败: %v", err)
	}

	if !redemptionAutoRefillExists {
		// 添加自动补给字段到redemption_records表
		alterRedemptionSQL := `
			ALTER TABLE redemption_records 
			ADD COLUMN auto_refill_enabled BOOLEAN NOT NULL DEFAULT FALSE COMMENT '自动补给开关',
			ADD COLUMN auto_refill_threshold BIGINT NOT NULL DEFAULT 0 COMMENT '补给阈值',
			ADD COLUMN auto_refill_amount BIGINT NOT NULL DEFAULT 0 COMMENT '补给数量'
		`

		err = DB.Exec(alterRedemptionSQL).Error
		if err != nil {
			return fmt.Errorf("添加auto_refill字段到redemption_records表失败: %v", err)
		}

		log.Println("✅ 成功添加auto_refill字段到redemption_records表")
	} else {
		log.Println("✅ redemption_records auto_refill字段已存在")
	}

	log.Println("✅ 自动补给字段迁移完成")
	return nil
}

// ExecuteAutoRefillMigration 执行完整的自动补给功能迁移
func ExecuteAutoRefillMigration() error {
	log.Println("🚀 开始执行自动补给功能迁移...")

	// 1. 添加自动补给字段
	if err := MigrateAddAutoRefillFields(); err != nil {
		return fmt.Errorf("自动补给字段迁移失败: %v", err)
	}

	log.Println("🎉 自动补给功能迁移完成！")
	log.Println("📋 迁移总结:")
	log.Println("   - ✅ 添加了auto_refill_enabled, auto_refill_threshold, auto_refill_amount字段到subscription_plans表")
	log.Println("   - ✅ 添加了auto_refill_enabled, auto_refill_threshold, auto_refill_amount, last_auto_refill_time字段到user_wallets表")
	log.Println("   - ✅ 添加了auto_refill_enabled, auto_refill_threshold, auto_refill_amount字段到redemption_records表")
	log.Println("")
	log.Println("💡 自动补给机制:")
	log.Println("   - 订阅计划可以配置自动补给参数")
	log.Println("   - 用户兑换订阅后，自动补给配置会复制到用户钱包")
	log.Println("   - 系统将在指定时间（0点、4点、8点、12点、16点、20点）检查用户积分并自动补给")

	return nil
}
