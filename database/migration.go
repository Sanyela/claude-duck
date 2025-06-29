package database

import (
	"fmt"
	"log"
	"time"

	"claude/models"
)

// MigrateSubscriptionRefactor 执行订阅表架构重构迁移
func MigrateSubscriptionRefactor() error {
	log.Println("🚀 开始执行订阅表架构重构迁移...")

	// 第一步：确保订阅表有新字段
	if err := ensureSubscriptionTableStructure(); err != nil {
		return fmt.Errorf("确保表结构失败: %v", err)
	}

	// 第二步：迁移积分池数据到订阅表
	if err := migratePointPoolsToSubscriptions(); err != nil {
		return fmt.Errorf("迁移积分池数据失败: %v", err)
	}

	// 第三步：同步积分一致性
	if err := syncPointsConsistency(); err != nil {
		return fmt.Errorf("同步积分一致性失败: %v", err)
	}

	// 第四步：删除旧表
	if err := deleteOldTables(); err != nil {
		return fmt.Errorf("删除旧表失败: %v", err)
	}

	// 第五步：标记迁移完成
	if err := markMigrationComplete(); err != nil {
		return fmt.Errorf("标记迁移完成失败: %v", err)
	}

	log.Println("✅ 订阅表架构重构迁移完成!")
	return nil
}

// ensureSubscriptionTableStructure 确保订阅表结构正确
func ensureSubscriptionTableStructure() error {
	log.Println("🔧 确保订阅表结构...")

	// 检查字段是否存在
	var fieldExists bool
	DB.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'subscriptions' AND column_name = 'activated_at')").Scan(&fieldExists)

	if !fieldExists {
		log.Println("添加新字段到订阅表...")
		addFieldQueries := []string{
			`ALTER TABLE subscriptions ADD COLUMN activated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP`,
			`ALTER TABLE subscriptions ADD COLUMN expires_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP`,
			`ALTER TABLE subscriptions ADD COLUMN total_points bigint NOT NULL DEFAULT 0`,
			`ALTER TABLE subscriptions ADD COLUMN used_points bigint NOT NULL DEFAULT 0`,
			`ALTER TABLE subscriptions ADD COLUMN available_points bigint NOT NULL DEFAULT 0`,
			`ALTER TABLE subscriptions ADD COLUMN source_type varchar(191) NOT NULL DEFAULT 'payment'`,
			`ALTER TABLE subscriptions ADD COLUMN source_id varchar(191) DEFAULT NULL`,
			`ALTER TABLE subscriptions ADD COLUMN invoice_url varchar(500) DEFAULT NULL`,
		}

		for _, query := range addFieldQueries {
			if err := DB.Exec(query).Error; err != nil {
				log.Printf("添加字段警告: %v", err) // 字段可能已存在
			}
		}

		// 创建索引
		indexQueries := []string{
			`CREATE INDEX idx_user_status ON subscriptions (user_id, status)`,
			`CREATE INDEX idx_activated_at ON subscriptions (activated_at)`,
			`CREATE INDEX idx_expires_at ON subscriptions (expires_at)`,
		}

		for _, query := range indexQueries {
			if err := DB.Exec(query).Error; err != nil {
				log.Printf("创建索引警告: %v", err) // 索引可能已存在
			}
		}
	}

	log.Println("✅ 订阅表结构确认完成")
	return nil
}

// migratePointPoolsToSubscriptions 迁移积分池数据到订阅表
func migratePointPoolsToSubscriptions() error {
	log.Println("🔄 迁移积分池数据到订阅表...")

	// 检查是否已经迁移过
	var migratedCount int64
	DB.Model(&models.Subscription{}).Where("total_points > 0").Count(&migratedCount)
	if migratedCount > 0 {
		log.Printf("发现 %d 条已迁移的订阅记录，跳过积分池迁移", migratedCount)
		return nil
	}

	// 获取所有积分池记录
	var pointPools []models.PointPool
	if err := DB.Find(&pointPools).Error; err != nil {
		return fmt.Errorf("获取积分池记录失败: %v", err)
	}

	log.Printf("找到 %d 条积分池记录需要迁移", len(pointPools))

	// 为每个积分池创建对应的订阅记录
	for i, pool := range pointPools {
		// 获取对应的订阅计划ID
		var subscriptionPlanID uint = 1 // 默认计划ID
		if pool.SourceType == "activation_code" {
			var activationCode models.ActivationCode
			if err := DB.Where("CONCAT('AC-', id) = ?", pool.SourceID).First(&activationCode).Error; err == nil {
				subscriptionPlanID = activationCode.SubscriptionPlanID
			}
		}

		// 确定订阅状态
		status := "expired"
		if pool.ExpiresAt.After(time.Now()) {
			status = "active"
		}

		// 使用原生SQL插入，设置current_period_end字段
		insertSQL := `
			INSERT INTO subscriptions (
				user_id, subscription_plan_id, status, activated_at, expires_at, 
				total_points, used_points, available_points, source_type, source_id, 
				invoice_url, cancel_at_period_end, current_period_end, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		if err := DB.Exec(insertSQL,
			pool.UserID, subscriptionPlanID, status, pool.CreatedAt, pool.ExpiresAt,
			pool.PointsTotal, pool.PointsTotal-pool.PointsRemaining, pool.PointsRemaining,
			pool.SourceType, pool.SourceID, "", false, pool.ExpiresAt, pool.CreatedAt, time.Now(),
		).Error; err != nil {
			log.Printf("⚠️ 创建订阅记录失败 (PointPool ID: %d): %v", pool.ID, err)
			continue
		}

		if (i+1)%10 == 0 || i == len(pointPools)-1 {
			log.Printf("进度: %d/%d", i+1, len(pointPools))
		}
	}

	log.Println("✅ 积分池数据迁移完成")
	return nil
}

// syncPointsConsistency 同步积分一致性
func syncPointsConsistency() error {
	log.Println("🔧 同步积分一致性...")

	// 获取所有订阅记录
	var subscriptions []models.Subscription
	if err := DB.Find(&subscriptions).Error; err != nil {
		return fmt.Errorf("获取订阅记录失败: %v", err)
	}

	for _, sub := range subscriptions {
		// 从API交易记录计算实际使用的积分
		var actualUsed int64
		DB.Model(&models.APITransaction{}).
			Where("user_id = ? AND status = 'success' AND created_at >= ? AND created_at <= ?",
				sub.UserID, sub.ActivatedAt, sub.ExpiresAt).
			Select("COALESCE(SUM(points_used), 0)").
			Scan(&actualUsed)

		// 确保不超过总积分
		if actualUsed > sub.TotalPoints {
			actualUsed = sub.TotalPoints
		}

		// 计算可用积分
		availablePoints := sub.TotalPoints - actualUsed
		if availablePoints < 0 {
			availablePoints = 0
		}

		// 更新订阅记录
		if sub.UsedPoints != actualUsed || sub.AvailablePoints != availablePoints {
			DB.Model(&sub).Updates(map[string]interface{}{
				"used_points":      actualUsed,
				"available_points": availablePoints,
				"updated_at":       time.Now(),
			})
		}
	}

	log.Println("✅ 积分一致性同步完成")
	return nil
}

// deleteOldTables 删除旧表
func deleteOldTables() error {
	log.Println("🗑️ 删除旧表...")

	// 删除旧表
	deleteQueries := []string{
		`DROP TABLE IF EXISTS point_pools`,
		`DROP TABLE IF EXISTS point_balances`,
		`DROP TABLE IF EXISTS payment_histories`,
	}

	for _, query := range deleteQueries {
		if err := DB.Exec(query).Error; err != nil {
			log.Printf("删除表警告: %v", err) // 表可能不存在
		} else {
			log.Printf("✅ 已删除表: %s", query)
		}
	}

	log.Println("✅ 旧表删除完成")
	return nil
}

// markMigrationComplete 标记迁移完成
func markMigrationComplete() error {
	log.Println("📝 标记迁移完成...")

	config := models.SystemConfig{
		ConfigKey:   "subscription_refactor_completed",
		ConfigValue: "true",
		Description: "订阅表架构重构迁移完成标记",
		UpdatedAt:   time.Now(),
	}

	// 使用 UPSERT 语义
	if err := DB.Where("config_key = ?", config.ConfigKey).
		Assign(map[string]interface{}{
			"config_value": config.ConfigValue,
			"updated_at":   time.Now(),
		}).
		FirstOrCreate(&config).Error; err != nil {
		return fmt.Errorf("标记迁移完成失败: %v", err)
	}

	log.Println("✅ 迁移完成标记已设置")
	return nil
}

// CleanupOldTables 手动清理旧表（保留此函数用于手动调用）
func CleanupOldTables() error {
	return deleteOldTables()
}

// NeedsSubscriptionRefactor 检查是否需要执行订阅表重构迁移
func NeedsSubscriptionRefactor() bool {
	// 检查迁移标记
	var config models.SystemConfig
	err := DB.Where("config_key = ?", "subscription_refactor_completed").First(&config).Error
	if err == nil && config.ConfigValue == "true" {
		return false // 迁移已完成
	}

	// 检查是否存在积分池表且有数据
	var poolCount int64
	err = DB.Model(&models.PointPool{}).Count(&poolCount).Error
	if err != nil || poolCount == 0 {
		return false // 没有数据需要迁移
	}

	return true
}
