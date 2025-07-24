package utils

import (
	"fmt"
	"time"

	"claude/database"
	"claude/models"

	"gorm.io/gorm"
)

// DetermineServiceLevel 判断服务等级（升级/降级/同级）
func DetermineServiceLevel(currentWallet *models.UserWallet, newPlan *models.SubscriptionPlan) string {
	// 比较关键属性来判断服务等级
	newScore := calculateServiceScore(newPlan.DailyCheckinPoints, newPlan.DailyCheckinPointsMax, 
		newPlan.AutoRefillAmount)
	currentScore := calculateServiceScore(currentWallet.DailyCheckinPoints, currentWallet.DailyCheckinPointsMax,
		currentWallet.AutoRefillAmount)
	
	if newScore > currentScore {
		return "upgrade"
	} else if newScore < currentScore {
		return "downgrade"
	}
	return "same_level"
}

// calculateServiceScore 计算服务分数
func calculateServiceScore(checkinMin, checkinMax, autoRefillAmount int64) int64 {
	// 按新的规则计算分数：签到奖励最高+最低 * 1 + 积分恢复积分 * 1(如果没有则是0)
	score := (checkinMin + checkinMax) * 1 + autoRefillAmount * 1
	return score
}

// RedeemActivationCodeToWallet 激活码兑换到钱包
func RedeemActivationCodeToWallet(userID uint, activationCode *models.ActivationCode) error {
	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取或创建用户钱包
	wallet, err := GetOrCreateUserWallet(userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("获取用户钱包失败: %v", err)
	}

	// 获取订阅计划
	var plan models.SubscriptionPlan
	if err := tx.Where("id = ?", activationCode.SubscriptionPlanID).First(&plan).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("获取订阅计划失败: %v", err)
	}

	// 判断服务等级
	serviceLevel := DetermineServiceLevel(wallet, &plan)

	// 计算新的过期时间
	newValidityDuration := time.Duration(plan.ValidityDays) * 24 * time.Hour
	var newExpiresAt time.Time
	
	switch serviceLevel {
	case "upgrade":
		// 升级：取当前过期时间和新过期时间的最大值
		newExpiresAt = maxTime(wallet.WalletExpiresAt, time.Now().Add(newValidityDuration))
	case "downgrade", "same_level":
		// 降级和同级：使用新的过期时间
		newExpiresAt = time.Now().Add(newValidityDuration)
	}

	// 根据服务等级决定更新的配置
	var updates map[string]interface{}
	
	switch serviceLevel {
	case "upgrade":
		// 升级：更新所有配置和积分
		updates = map[string]interface{}{
			"wallet_expires_at":        newExpiresAt,
			"daily_max_points":         plan.DailyMaxPoints,
			"degradation_guaranteed":   plan.DegradationGuaranteed,
			"daily_checkin_points":     plan.DailyCheckinPoints,
			"daily_checkin_points_max": plan.DailyCheckinPointsMax,
			"auto_refill_enabled":      plan.AutoRefillEnabled,
			"auto_refill_threshold":    plan.AutoRefillThreshold,
			"auto_refill_amount":       plan.AutoRefillAmount,
			"status":                   "active",
			"updated_at":               time.Now(),
			"total_points":             gorm.Expr("total_points + ?", plan.PointAmount),
			"available_points":         gorm.Expr("available_points + ?", plan.PointAmount),
		}
	case "downgrade":
		// 降级：只增加积分，不更新其他配置，保留当前权益
		updates = map[string]interface{}{
			"status":                   "active",
			"updated_at":               time.Now(),
			"total_points":             gorm.Expr("total_points + ?", plan.PointAmount),
			"available_points":         gorm.Expr("available_points + ?", plan.PointAmount),
		}
		// 如果当前钱包已过期，需要激活并延长过期时间
		if wallet.WalletExpiresAt.Before(time.Now()) {
			updates["wallet_expires_at"] = newExpiresAt
		}
	case "same_level":
		// 同级：重置所有积分和配置
		updates = map[string]interface{}{
			"wallet_expires_at":        newExpiresAt,
			"daily_max_points":         plan.DailyMaxPoints,
			"degradation_guaranteed":   plan.DegradationGuaranteed,
			"daily_checkin_points":     plan.DailyCheckinPoints,
			"daily_checkin_points_max": plan.DailyCheckinPointsMax,
			"auto_refill_enabled":      plan.AutoRefillEnabled,
			"auto_refill_threshold":    plan.AutoRefillThreshold,
			"auto_refill_amount":       plan.AutoRefillAmount,
			"status":                   "active",
			"updated_at":               time.Now(),
			"total_points":             plan.PointAmount,
			"available_points":         plan.PointAmount,
			"used_points":              0,
		}
	}

	// 更新钱包
	if err := tx.Model(&models.UserWallet{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新用户钱包失败: %v", err)
	}

	// 创建兑换记录
	redemptionRecord := models.RedemptionRecord{
		UserID:                userID,
		SourceType:            "activation_code",
		SourceID:              activationCode.Code,
		PointsAmount:          plan.PointAmount,
		ValidityDays:          plan.ValidityDays,
		SubscriptionPlanID:    &plan.ID,
		DailyMaxPoints:        plan.DailyMaxPoints,
		DegradationGuaranteed: plan.DegradationGuaranteed,
		DailyCheckinPoints:    plan.DailyCheckinPoints,
		DailyCheckinPointsMax: plan.DailyCheckinPointsMax,
		AutoRefillEnabled:     plan.AutoRefillEnabled,
		AutoRefillThreshold:   plan.AutoRefillThreshold,
		AutoRefillAmount:      plan.AutoRefillAmount,
		ActivatedAt:           time.Now(),
		ExpiresAt:             newExpiresAt,
		Reason:                fmt.Sprintf("激活码兑换 - %s服务", serviceLevel),
		BatchNumber:           activationCode.BatchNumber,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := tx.Create(&redemptionRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建兑换记录失败: %v", err)
	}

	// 更新激活码状态
	if err := tx.Model(activationCode).Updates(map[string]interface{}{
		"status":         "used",
		"used_by_user_id": userID,
		"used_at":        time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新激活码状态失败: %v", err)
	}

	// 提交事务
	return tx.Commit().Error
}

// AdminGiftToWallet 管理员赠送积分到钱包
func AdminGiftToWallet(adminUserID, targetUserID uint, plan *models.SubscriptionPlan, customPoints int64, customValidityDays int, customDailyLimit int64, reason string) error {
	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取或创建用户钱包
	wallet, err := GetOrCreateUserWallet(targetUserID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("获取用户钱包失败: %v", err)
	}

	// 使用自定义值或计划默认值
	pointsAmount := customPoints
	if pointsAmount <= 0 {
		pointsAmount = plan.PointAmount
	}

	validityDays := customValidityDays
	if validityDays <= 0 {
		validityDays = plan.ValidityDays
	}

	dailyMaxPoints := customDailyLimit
	if dailyMaxPoints < 0 {
		dailyMaxPoints = plan.DailyMaxPoints
	}

	// 计算过期时间
	expiresAt := time.Now().Add(time.Duration(validityDays) * 24 * time.Hour)

	// 累加积分到钱包（管理员赠送总是累加）
	updates := map[string]interface{}{
		"total_points":     gorm.Expr("total_points + ?", pointsAmount),
		"available_points": gorm.Expr("available_points + ?", pointsAmount),
		"updated_at":       time.Now(),
		"status":           "active",
	}

	// 过期时间：取最大值
	if expiresAt.After(wallet.WalletExpiresAt) {
		updates["wallet_expires_at"] = expiresAt
	}

	// 所有配置都采用"取最优值"策略，确保用户权益只增不减
	if dailyMaxPoints > wallet.DailyMaxPoints {
		updates["daily_max_points"] = dailyMaxPoints
	}
	if plan.DegradationGuaranteed > wallet.DegradationGuaranteed {
		updates["degradation_guaranteed"] = plan.DegradationGuaranteed
	}
	if plan.DailyCheckinPoints > wallet.DailyCheckinPoints {
		updates["daily_checkin_points"] = plan.DailyCheckinPoints
	}
	if plan.DailyCheckinPointsMax > wallet.DailyCheckinPointsMax {
		updates["daily_checkin_points_max"] = plan.DailyCheckinPointsMax
	}
	
	// 自动补给配置：只有当新配置更优时才更新
	if plan.AutoRefillEnabled && plan.AutoRefillAmount > wallet.AutoRefillAmount {
		updates["auto_refill_enabled"] = plan.AutoRefillEnabled
		updates["auto_refill_threshold"] = plan.AutoRefillThreshold
		updates["auto_refill_amount"] = plan.AutoRefillAmount
	}

	// 更新钱包
	if err := tx.Model(&models.UserWallet{}).Where("user_id = ?", targetUserID).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新用户钱包失败: %v", err)
	}

	// 创建兑换记录
	redemptionRecord := models.RedemptionRecord{
		UserID:                targetUserID,
		SourceType:            "admin_gift",
		SourceID:              fmt.Sprintf("admin_%d_%d", adminUserID, time.Now().Unix()),
		PointsAmount:          pointsAmount,
		ValidityDays:          validityDays,
		SubscriptionPlanID:    &plan.ID,
		DailyMaxPoints:        dailyMaxPoints,
		DegradationGuaranteed: plan.DegradationGuaranteed,
		DailyCheckinPoints:    plan.DailyCheckinPoints,
		DailyCheckinPointsMax: plan.DailyCheckinPointsMax,
		AutoRefillEnabled:     plan.AutoRefillEnabled,
		AutoRefillThreshold:   plan.AutoRefillThreshold,
		AutoRefillAmount:      plan.AutoRefillAmount,
		ActivatedAt:           time.Now(),
		ExpiresAt:             expiresAt,
		Reason:                reason,
		AdminUserID:           &adminUserID,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := tx.Create(&redemptionRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建兑换记录失败: %v", err)
	}

	// 提交事务
	return tx.Commit().Error
}

// DailyCheckinToWallet 每日签到奖励积分到钱包
func DailyCheckinToWallet(userID uint, checkinPoints int64) error {
	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取用户钱包（检查是否存在）
	_, err := GetUserWallet(userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("获取用户钱包失败: %v", err)
	}

	// 计算签到积分过期时间（1天）
	expiresAt := time.Now().Add(24 * time.Hour)

	// 累加签到积分到钱包
	updates := map[string]interface{}{
		"total_points":      gorm.Expr("total_points + ?", checkinPoints),
		"available_points":  gorm.Expr("available_points + ?", checkinPoints),
		"last_checkin_date": time.Now().Format("2006-01-02"),
		"updated_at":        time.Now(),
	}

	// 更新钱包
	if err := tx.Model(&models.UserWallet{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新用户钱包失败: %v", err)
	}

	// 创建兑换记录
	redemptionRecord := models.RedemptionRecord{
		UserID:       userID,
		SourceType:   "daily_checkin",
		SourceID:     fmt.Sprintf("checkin_%s", time.Now().Format("2006-01-02")),
		PointsAmount: checkinPoints,
		ValidityDays: 1, // 签到积分有效期1天
		ActivatedAt:  time.Now(),
		ExpiresAt:    expiresAt,
		Reason:       "每日签到奖励",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := tx.Create(&redemptionRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建兑换记录失败: %v", err)
	}

	// 提交事务
	return tx.Commit().Error
}

// maxTime 返回两个时间中的较大值
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// GetWalletActiveRedemptionRecords 获取钱包有效的兑换记录（排除被封禁的激活码）
func GetWalletActiveRedemptionRecords(userID uint) ([]models.RedemptionRecord, error) {
	var records []models.RedemptionRecord
	
	// 获取被封禁的激活码列表
	var bannedCodes []string
	database.DB.Table("frozen_points_records").
		Where("user_id = ? AND status = 'frozen'", userID).
		Pluck("banned_activation_code", &bannedCodes)
	
	query := database.DB.Where("user_id = ? AND expires_at > ?", userID, time.Now())
	
	// 排除被封禁的激活码记录
	if len(bannedCodes) > 0 {
		query = query.Where("NOT (source_type = 'activation_code' AND source_id IN ?)", bannedCodes)
	}
	
	err := query.Order("created_at DESC").Find(&records).Error
	return records, err
}

// GetWalletRedemptionHistory 获取钱包兑换历史记录
func GetWalletRedemptionHistory(userID uint, limit, offset int) ([]models.RedemptionRecord, int64, error) {
	var records []models.RedemptionRecord
	var total int64

	// 获取总数
	database.DB.Model(&models.RedemptionRecord{}).Where("user_id = ?", userID).Count(&total)

	// 获取分页记录
	err := database.DB.Preload("SubscriptionPlan").Preload("AdminUser").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&records).Error

	return records, total, err
}