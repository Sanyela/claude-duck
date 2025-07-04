package database

import (
	"fmt"
	"log"

	"claude/models"
)

// MigrateAddAccumulatedTokens 为用户钱包表添加累计token字段
func MigrateAddAccumulatedTokens() error {
	log.Println("开始执行累计token字段迁移...")

	// 检查user_wallets表是否存在accumulated_tokens字段
	var columnExists bool
	checkSQL := `
		SELECT COUNT(*) > 0 
		FROM information_schema.columns 
		WHERE table_schema = DATABASE() 
		AND table_name = 'user_wallets' 
		AND column_name = 'accumulated_tokens'
	`
	
	err := DB.Raw(checkSQL).Scan(&columnExists).Error
	if err != nil {
		return fmt.Errorf("检查accumulated_tokens字段失败: %v", err)
	}

	if !columnExists {
		// 添加accumulated_tokens字段
		alterSQL := `
			ALTER TABLE user_wallets 
			ADD COLUMN accumulated_tokens BIGINT NOT NULL DEFAULT 0 
			COMMENT '累计加权token数量'
		`
		
		err = DB.Exec(alterSQL).Error
		if err != nil {
			return fmt.Errorf("添加accumulated_tokens字段失败: %v", err)
		}
		
		log.Println("✅ 成功添加accumulated_tokens字段到user_wallets表")
	} else {
		log.Println("✅ accumulated_tokens字段已存在")
	}

	// 验证字段已添加
	var wallet models.UserWallet
	err = DB.First(&wallet).Error
	if err == nil {
		log.Printf("✅ 字段验证成功，累计token字段: %d", wallet.AccumulatedTokens)
	}

	log.Println("✅ 累计token字段迁移完成")
	return nil
}

// MigrateTokenThresholdConfig 迁移系统配置，移除旧的阶梯计费表配置
func MigrateTokenThresholdConfig() error {
	log.Println("开始执行系统配置迁移...")

	// 删除旧的token_pricing_table配置
	err := DB.Where("config_key = ?", "token_pricing_table").Delete(&models.SystemConfig{}).Error
	if err != nil {
		log.Printf("删除旧的token_pricing_table配置失败（可能不存在）: %v", err)
	} else {
		log.Println("✅ 已删除旧的token_pricing_table配置")
	}

	// 确保新的配置存在
	configs := []models.SystemConfig{
		{
			ConfigKey:   "token_threshold",
			ConfigValue: "5000",
			Description: "累计token计费阈值",
		},
		{
			ConfigKey:   "points_per_threshold",
			ConfigValue: "1",
			Description: "每阈值扣费积分数量",
		},
	}

	for _, cfg := range configs {
		var existing models.SystemConfig
		err := DB.Where("config_key = ?", cfg.ConfigKey).First(&existing).Error
		if err != nil {
			// 配置不存在，创建新配置
			if err := DB.Create(&cfg).Error; err != nil {
				log.Printf("创建配置 %s 失败: %v", cfg.ConfigKey, err)
			} else {
				log.Printf("✅ 创建新配置: %s = %s", cfg.ConfigKey, cfg.ConfigValue)
			}
		} else {
			log.Printf("✅ 配置已存在: %s = %s", cfg.ConfigKey, existing.ConfigValue)
		}
	}

	log.Println("✅ 系统配置迁移完成")
	return nil
}

// ExecuteAccumulatedTokensMigration 执行完整的累计token计费迁移
func ExecuteAccumulatedTokensMigration() error {
	log.Println("🚀 开始执行累计token计费系统迁移...")

	// 1. 添加accumulated_tokens字段
	if err := MigrateAddAccumulatedTokens(); err != nil {
		return fmt.Errorf("累计token字段迁移失败: %v", err)
	}

	// 2. 迁移系统配置
	if err := MigrateTokenThresholdConfig(); err != nil {
		return fmt.Errorf("系统配置迁移失败: %v", err)
	}

	// 3. 重置所有用户的累计token为0（可选）
	resetSQL := `UPDATE user_wallets SET accumulated_tokens = 0`
	err := DB.Exec(resetSQL).Error
	if err != nil {
		log.Printf("重置用户累计token失败: %v", err)
	} else {
		log.Println("✅ 已重置所有用户累计token为0")
	}

	log.Println("🎉 累计token计费系统迁移完成！")
	log.Println("📋 迁移总结:")
	log.Println("   - ✅ 添加了accumulated_tokens字段到user_wallets表")
	log.Println("   - ✅ 删除了旧的token_pricing_table配置")
	log.Println("   - ✅ 添加了token_threshold和points_per_threshold配置")
	log.Println("   - ✅ 重置了所有用户的累计token计数")
	log.Println("")
	log.Println("💡 新计费机制:")
	log.Printf("   - 每累计 %s 个加权token扣除 %s 积分", "5000", "1")
	log.Println("   - 保留加权token计算（支持设置倍率为0来禁用某类token计费）")
	log.Println("   - 避免了小额token也扣费的问题")

	return nil
}