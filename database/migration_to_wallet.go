package database

import (
	"fmt"
	"log"
	"time"

	"claude/models"

	"gorm.io/gorm"
)

// MigrateToWalletArchitecture 将老架构数据迁移到新的钱包架构
func MigrateToWalletArchitecture() error {
	log.Println("🚀 开始迁移数据到新钱包架构...")

	// 开始事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 获取所有用户
	var users []models.User
	if err := tx.Find(&users).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("获取用户列表失败: %v", err)
	}

	log.Printf("📊 找到 %d 个用户，开始迁移...", len(users))

	// 2. 为每个用户创建钱包并迁移数据
	for _, user := range users {
		if err := migrateUserToWallet(tx, user.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("迁移用户 %d 失败: %v", user.ID, err)
		}
	}

	// 3. 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交迁移事务失败: %v", err)
	}

	log.Println("✅ 数据迁移完成！")
	return nil
}

// migrateUserToWallet 迁移单个用户的数据到钱包架构
func migrateUserToWallet(tx *gorm.DB, userID uint) error {
	log.Printf("  迁移用户 %d...", userID)

	// 1. 检查是否已经有钱包（避免重复迁移）
	var existingWallet models.UserWallet
	if err := tx.Where("user_id = ?", userID).First(&existingWallet).Error; err == nil {
		log.Printf("  用户 %d 已有钱包，跳过", userID)
		return nil
	}

	// 2. 获取用户所有订阅记录
	var subscriptions []models.Subscription
	if err := tx.Preload("Plan").Where("user_id = ?", userID).
		Order("activated_at DESC").Find(&subscriptions).Error; err != nil {
		return fmt.Errorf("获取用户 %d 订阅记录失败: %v", userID, err)
	}

	// 3. 计算钱包状态
	wallet := models.UserWallet{
		UserID:          userID,
		TotalPoints:     0,
		AvailablePoints: 0,
		UsedPoints:      0,
		WalletExpiresAt: time.Now(), // 默认当前时间
		Status:          "expired",  // 默认过期状态
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	var latestSubscription *models.Subscription
	var latestActivationTime time.Time

	// 4. 汇总所有有效订阅的积分
	for _, sub := range subscriptions {
		// 累计积分
		wallet.TotalPoints += sub.TotalPoints
		wallet.UsedPoints += sub.UsedPoints

		// 只累计未过期的订阅的可用积分
		if sub.Status == "active" && sub.ExpiresAt.After(time.Now()) {
			wallet.AvailablePoints += sub.AvailablePoints
			wallet.Status = "active"

			// 更新钱包过期时间为最晚过期时间
			if sub.ExpiresAt.After(wallet.WalletExpiresAt) {
				wallet.WalletExpiresAt = sub.ExpiresAt
			}
		}

		// 找到最新激活的套餐（用于确定当前属性）
		if latestSubscription == nil || sub.ActivatedAt.After(latestActivationTime) {
			latestSubscription = &sub
			latestActivationTime = sub.ActivatedAt
		}
	}

	// 5. 设置钱包属性（来自最新激活的套餐）
	if latestSubscription != nil {
		wallet.DailyMaxPoints = latestSubscription.DailyMaxPoints
		if wallet.DailyMaxPoints == 0 && latestSubscription.Plan.DailyMaxPoints > 0 {
			wallet.DailyMaxPoints = latestSubscription.Plan.DailyMaxPoints
		}

		wallet.DegradationGuaranteed = latestSubscription.Plan.DegradationGuaranteed
		wallet.DailyCheckinPoints = latestSubscription.Plan.DailyCheckinPoints
		wallet.DailyCheckinPointsMax = latestSubscription.Plan.DailyCheckinPointsMax
	}

	// 6. 获取最后签到日期
	var lastCheckin models.DailyCheckin
	if err := tx.Where("user_id = ?", userID).
		Order("checkin_date DESC").First(&lastCheckin).Error; err == nil {
		wallet.LastCheckinDate = lastCheckin.CheckinDate
	}

	// 7. 创建用户钱包
	if err := tx.Create(&wallet).Error; err != nil {
		return fmt.Errorf("创建用户 %d 钱包失败: %v", userID, err)
	}

	// 8. 迁移订阅记录到兑换记录
	for _, sub := range subscriptions {
		record := models.RedemptionRecord{
			UserID:                userID,
			SourceType:            sub.SourceType,
			SourceID:              sub.SourceID,
			PointsAmount:          sub.TotalPoints,
			ValidityDays:          calculateValidityDays(sub.ActivatedAt, sub.ExpiresAt),
			SubscriptionPlanID:    &sub.SubscriptionPlanID,
			DailyMaxPoints:        sub.DailyMaxPoints,
			DegradationGuaranteed: sub.Plan.DegradationGuaranteed,
			DailyCheckinPoints:    sub.Plan.DailyCheckinPoints,
			DailyCheckinPointsMax: sub.Plan.DailyCheckinPointsMax,
			ActivatedAt:           sub.ActivatedAt,
			ExpiresAt:             sub.ExpiresAt,
			InvoiceURL:            sub.InvoiceURL,
			CreatedAt:             sub.CreatedAt,
			UpdatedAt:             sub.UpdatedAt,
		}

		// 设置原因描述
		switch sub.SourceType {
		case "activation_code":
			record.Reason = "激活码兑换"
		case "admin_gift":
			record.Reason = "管理员赠送"
		case "daily_checkin":
			record.Reason = "每日签到奖励"
		case "payment":
			record.Reason = "在线支付"
		default:
			record.Reason = "未知来源"
		}

		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("创建用户 %d 兑换记录失败: %v", userID, err)
		}
	}

	// 9. 迁移每日使用记录
	var oldDailyUsages []models.DailyPointsUsage
	if err := tx.Where("user_id = ?", userID).Find(&oldDailyUsages).Error; err != nil {
		return fmt.Errorf("获取用户 %d 每日使用记录失败: %v", userID, err)
	}

	// 按日期聚合使用记录
	dailyUsageMap := make(map[string]int64)
	for _, usage := range oldDailyUsages {
		dailyUsageMap[usage.UsageDate] += usage.PointsUsed
	}

	// 创建新的每日使用记录
	for date, points := range dailyUsageMap {
		newUsage := models.UserDailyUsage{
			UserID:     userID,
			UsageDate:  date,
			PointsUsed: points,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := tx.Create(&newUsage).Error; err != nil {
			return fmt.Errorf("创建用户 %d 新每日使用记录失败: %v", userID, err)
		}
	}

	log.Printf("  ✅ 用户 %d 迁移完成 (积分: %d/%d, 状态: %s)",
		userID, wallet.AvailablePoints, wallet.TotalPoints, wallet.Status)

	return nil
}

// calculateValidityDays 计算有效期天数
func calculateValidityDays(activatedAt, expiresAt time.Time) int {
	duration := expiresAt.Sub(activatedAt)
	days := int(duration.Hours() / 24)
	if days < 1 {
		days = 1 // 至少1天
	}
	return days
}

// VerifyMigration 验证迁移结果
func VerifyMigration() error {
	log.Println("🔍 开始验证迁移结果...")

	// 1. 检查钱包数量是否等于用户数量
	var userCount, walletCount int64
	DB.Model(&models.User{}).Count(&userCount)
	DB.Model(&models.UserWallet{}).Count(&walletCount)

	if userCount != walletCount {
		return fmt.Errorf("用户数量 (%d) 与钱包数量 (%d) 不匹配", userCount, walletCount)
	}

	// 2. 检查积分总数是否一致
	var oldTotalPoints, newTotalPoints int64
	DB.Model(&models.Subscription{}).Select("SUM(total_points)").Scan(&oldTotalPoints)
	DB.Model(&models.UserWallet{}).Select("SUM(total_points)").Scan(&newTotalPoints)

	log.Printf("📊 积分验证: 老架构总积分=%d, 新架构总积分=%d", oldTotalPoints, newTotalPoints)

	// 3. 检查兑换记录数量
	var subscriptionCount, recordCount int64
	DB.Model(&models.Subscription{}).Count(&subscriptionCount)
	DB.Model(&models.RedemptionRecord{}).Count(&recordCount)

	log.Printf("📊 记录验证: 订阅记录=%d, 兑换记录=%d", subscriptionCount, recordCount)

	log.Println("✅ 迁移验证完成！")
	return nil
}
