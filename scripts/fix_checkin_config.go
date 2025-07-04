package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"claude/config"
	"claude/database"
	"claude/models"
)

func main() {
	log.Println("🔧 开始修复用户签到配置 (V2版本)...")

	// 初始化配置
	config.LoadConfig()

	// 初始化数据库连接
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	// 检查命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--dry-run":
			log.Println("🔍 执行干跑模式，只检查不修改...")
			if err := dryRun(); err != nil {
				log.Fatalf("❌ 干跑检查失败: %v", err)
			}
			return
		case "--verify":
			log.Println("🔍 验证修复结果...")
			if err := verifyFix(); err != nil {
				log.Fatalf("❌ 验证失败: %v", err)
			}
			return
		case "--user":
			if len(os.Args) > 2 {
				log.Printf("🔍 仅修复用户 %s...", os.Args[2])
				if err := fixSingleUser(os.Args[2]); err != nil {
					log.Fatalf("❌ 单用户修复失败: %v", err)
				}
				return
			}
			log.Println("❌ 请指定用户ID: --user 16")
			return
		case "--help":
			printHelp()
			return
		}
	}

	// 执行修复
	if err := fixCheckinConfig(); err != nil {
		log.Fatalf("❌ 修复失败: %v", err)
	}

	// 验证修复结果
	if err := verifyFix(); err != nil {
		log.Fatalf("❌ 修复验证失败: %v", err)
	}

	log.Println("✅ 签到配置修复完成！")
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("用法:")
	fmt.Println("  go run fix_checkin_config_v2.go           # 执行修复")
	fmt.Println("  go run fix_checkin_config_v2.go --dry-run # 干跑模式，只检查不修改")
	fmt.Println("  go run fix_checkin_config_v2.go --verify  # 验证修复结果")
	fmt.Println("  go run fix_checkin_config_v2.go --user 16 # 只修复指定用户")
	fmt.Println("  go run fix_checkin_config_v2.go --help    # 显示帮助")
}

// dryRun 干跑模式，检查哪些用户需要修复
func dryRun() error {
	log.Println("📊 分析需要修复的用户...")

	// 查询所有签到配置为0的用户钱包
	var problematicWallets []models.UserWallet
	if err := database.DB.Where("daily_checkin_points = 0 AND daily_checkin_points_max = 0").
		Find(&problematicWallets).Error; err != nil {
		return fmt.Errorf("查询问题钱包失败: %v", err)
	}

	log.Printf("🔍 发现 %d 个签到配置异常的用户", len(problematicWallets))

	fixableCount := 0
	for _, wallet := range problematicWallets {
		fixConfig, err := getCorrectCheckinConfig(wallet.UserID)
		if err != nil {
			log.Printf("❌ 用户 %d 获取修复配置失败: %v", wallet.UserID, err)
			continue
		}

		if fixConfig != nil {
			log.Printf("✅ 用户 %d 可修复: %d-%d -> %d-%d (%s)",
				wallet.UserID,
				wallet.DailyCheckinPoints, wallet.DailyCheckinPointsMax,
				fixConfig.CheckinPoints, fixConfig.CheckinPointsMax,
				fixConfig.PlanTitle)
			fixableCount++
		} else {
			log.Printf("⚠️ 用户 %d 无有效套餐，无法修复", wallet.UserID)
		}
	}

	log.Printf("📈 总计：%d 个问题用户，%d 个可修复", len(problematicWallets), fixableCount)
	return nil
}

// fixSingleUser 修复单个用户
func fixSingleUser(userIDStr string) error {
	log.Printf("🔧 开始修复用户 %s...", userIDStr)

	// 查询用户钱包
	var wallet models.UserWallet
	if err := database.DB.Where("user_id = ?", userIDStr).First(&wallet).Error; err != nil {
		return fmt.Errorf("查询用户 %s 钱包失败: %v", userIDStr, err)
	}

	log.Printf("📊 用户 %s 当前配置: %d-%d", userIDStr, wallet.DailyCheckinPoints, wallet.DailyCheckinPointsMax)

	// 获取正确配置
	fixConfig, err := getCorrectCheckinConfig(wallet.UserID)
	if err != nil {
		return fmt.Errorf("获取用户 %s 修复配置失败: %v", userIDStr, err)
	}

	if fixConfig == nil {
		log.Printf("⚠️ 用户 %s 无有效签到套餐", userIDStr)
		return nil
	}

	log.Printf("🎯 应修复为: %d-%d (%s)", fixConfig.CheckinPoints, fixConfig.CheckinPointsMax, fixConfig.PlanTitle)

	// 执行修复
	if err := performSingleUserFix(wallet.UserID, fixConfig); err != nil {
		return fmt.Errorf("执行用户 %s 修复失败: %v", userIDStr, err)
	}

	// 验证修复结果
	var updatedWallet models.UserWallet
	if err := database.DB.Where("user_id = ?", userIDStr).First(&updatedWallet).Error; err != nil {
		return fmt.Errorf("验证用户 %s 修复结果失败: %v", userIDStr, err)
	}

	log.Printf("✅ 用户 %s 修复完成: %d-%d", userIDStr, updatedWallet.DailyCheckinPoints, updatedWallet.DailyCheckinPointsMax)
	return nil
}

// performSingleUserFix 执行单用户修复
func performSingleUserFix(userID uint, fixConfig *FixConfig) error {
	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 使用原生SQL强制更新
	sql := `UPDATE user_wallets 
			SET daily_checkin_points = ?, 
			    daily_checkin_points_max = ?, 
			    updated_at = ? 
			WHERE user_id = ?`

	result := tx.Exec(sql, fixConfig.CheckinPoints, fixConfig.CheckinPointsMax, time.Now(), userID)

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("SQL更新失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("没有行被更新，用户ID可能不存在")
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	log.Printf("🔄 SQL执行成功: 影响行数 %d", result.RowsAffected)
	return nil
}

// FixConfig 修复配置结构
type FixConfig struct {
	CheckinPoints    int64
	CheckinPointsMax int64
	PlanTitle        string
	PlanID           uint
}

// getCorrectCheckinConfig 获取用户正确的签到配置
func getCorrectCheckinConfig(userID uint) (*FixConfig, error) {
	// 查询用户当前有效的兑换记录（排除签到记录）
	var records []struct {
		models.RedemptionRecord
		PlanTitle string `gorm:"column:plan_title"`
	}

	query := `
		SELECT rr.*, sp.title as plan_title
		FROM redemption_records rr
		JOIN subscription_plans sp ON rr.subscription_plan_id = sp.id
		WHERE rr.user_id = ? 
		AND rr.expires_at > NOW()
		AND rr.source_type != 'daily_checkin'
		AND (sp.daily_checkin_points > 0 OR sp.daily_checkin_points_max > 0)
		ORDER BY sp.daily_checkin_points DESC, sp.daily_checkin_points_max DESC
		LIMIT 1
	`

	if err := database.DB.Raw(query, userID).Scan(&records).Error; err != nil {
		return nil, fmt.Errorf("查询用户 %d 有效套餐失败: %v", userID, err)
	}

	if len(records) == 0 {
		return nil, nil // 没有有效的签到套餐
	}

	record := records[0]
	return &FixConfig{
		CheckinPoints:    record.DailyCheckinPoints,
		CheckinPointsMax: record.DailyCheckinPointsMax,
		PlanTitle:        record.PlanTitle,
		PlanID:           *record.SubscriptionPlanID,
	}, nil
}

// fixCheckinConfig 执行修复
func fixCheckinConfig() error {
	log.Println("🔧 开始执行签到配置修复...")

	// 查询所有签到配置为0的用户钱包
	var problematicWallets []models.UserWallet
	if err := database.DB.Where("daily_checkin_points = 0 AND daily_checkin_points_max = 0").
		Find(&problematicWallets).Error; err != nil {
		return fmt.Errorf("查询问题钱包失败: %v", err)
	}

	log.Printf("📊 找到 %d 个需要修复的用户", len(problematicWallets))

	fixedCount := 0
	failedCount := 0

	for _, wallet := range problematicWallets {
		fixConfig, err := getCorrectCheckinConfig(wallet.UserID)
		if err != nil {
			log.Printf("❌ 用户 %d 获取修复配置失败: %v", wallet.UserID, err)
			failedCount++
			continue
		}

		if fixConfig == nil {
			log.Printf("⚠️ 用户 %d 无有效签到套餐，跳过", wallet.UserID)
			continue
		}

		// 执行修复
		if err := performSingleUserFix(wallet.UserID, fixConfig); err != nil {
			log.Printf("❌ 用户 %d 修复失败: %v", wallet.UserID, err)
			failedCount++
			continue
		}

		log.Printf("✅ 用户 %d 修复成功: %s (ID:%d) -> 签到 %d-%d",
			wallet.UserID, fixConfig.PlanTitle, fixConfig.PlanID,
			fixConfig.CheckinPoints, fixConfig.CheckinPointsMax)
		fixedCount++
	}

	log.Printf("🎉 修复完成！成功: %d 个, 失败: %d 个", fixedCount, failedCount)
	return nil
}

// verifyFix 验证修复结果
func verifyFix() error {
	log.Println("🔍 验证修复结果...")

	// 检查还有多少用户的签到配置仍然为0
	var remainingCount int64
	if err := database.DB.Model(&models.UserWallet{}).
		Where("daily_checkin_points = 0 AND daily_checkin_points_max = 0").
		Count(&remainingCount).Error; err != nil {
		return fmt.Errorf("统计剩余问题用户失败: %v", err)
	}

	// 检查已修复的用户数量
	var fixedCount int64
	if err := database.DB.Model(&models.UserWallet{}).
		Where("daily_checkin_points > 0 OR daily_checkin_points_max > 0").
		Count(&fixedCount).Error; err != nil {
		return fmt.Errorf("统计已修复用户失败: %v", err)
	}

	log.Printf("📊 验证结果:")
	log.Printf("   - 已修复用户: %d 个", fixedCount)
	log.Printf("   - 仍有问题用户: %d 个", remainingCount)

	// 具体检查用户16的情况
	var user16Wallet models.UserWallet
	if err := database.DB.Where("user_id = 16").First(&user16Wallet).Error; err != nil {
		log.Printf("⚠️ 用户16钱包查询失败: %v", err)
	} else {
		log.Printf("👤 用户16签到配置: %d-%d (状态: %s)",
			user16Wallet.DailyCheckinPoints,
			user16Wallet.DailyCheckinPointsMax,
			user16Wallet.Status)

		// 检查用户16是否有有效的签到套餐
		fixConfig, err := getCorrectCheckinConfig(16)
		if err != nil {
			log.Printf("⚠️ 用户16配置检查失败: %v", err)
		} else if fixConfig != nil {
			log.Printf("👤 用户16应有配置: %d-%d (%s)",
				fixConfig.CheckinPoints, fixConfig.CheckinPointsMax, fixConfig.PlanTitle)
		} else {
			log.Printf("👤 用户16无有效签到套餐")
		}
	}

	if remainingCount > 0 {
		log.Printf("⚠️ 仍有 %d 个用户的签到配置异常，可能是没有有效的签到套餐", remainingCount)
	}

	log.Println("✅ 验证完成!")
	return nil
}
