package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"claude/database"
	"claude/models"

	"gorm.io/gorm"
)

// RegistrationPlanMapping 注册套餐映射结构
type RegistrationPlanMapping struct {
	Default   int `json:"default"`   // 普通注册用户
	LinuxDo   int `json:"linux_do"`  // Linux Do OAuth
	GitHub    int `json:"github"`    // GitHub OAuth
	Google    int `json:"google"`    // Google OAuth
}

// GetRegistrationPlanMapping 获取注册套餐映射配置
func GetRegistrationPlanMapping() (*RegistrationPlanMapping, error) {
	var config models.SystemConfig
	err := database.DB.Where("config_key = ?", "registration_plan_mapping").First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果配置不存在，返回默认值（不赠送任何套餐）
			return &RegistrationPlanMapping{
				Default: -1,
				LinuxDo: -1,
				GitHub:  -1,
				Google:  -1,
			}, nil
		}
		return nil, fmt.Errorf("获取注册套餐映射配置失败: %v", err)
	}

	var mapping RegistrationPlanMapping
	err = json.Unmarshal([]byte(config.ConfigValue), &mapping)
	if err != nil {
		return nil, fmt.Errorf("解析注册套餐映射配置失败: %v", err)
	}

	return &mapping, nil
}

// ProcessRegistrationPlanGift 处理新用户注册套餐赠送
// registrationType: "default", "linux_do", "github", "google"
func ProcessRegistrationPlanGift(userID uint, registrationType string) error {
	// 获取套餐映射配置
	mapping, err := GetRegistrationPlanMapping()
	if err != nil {
		log.Printf("获取注册套餐映射配置失败: %v", err)
		return err
	}

	// 根据注册类型获取套餐ID
	var planID int
	switch registrationType {
	case "default":
		planID = mapping.Default
	case "linux_do":
		planID = mapping.LinuxDo
	case "github":
		planID = mapping.GitHub
	case "google":
		planID = mapping.Google
	default:
		planID = mapping.Default
	}

	// 如果套餐ID为-1，表示不赠送套餐
	if planID == -1 {
		log.Printf("用户 %d (%s注册) 无需赠送套餐", userID, registrationType)
		return nil
	}

	// 检查套餐是否存在
	var plan models.SubscriptionPlan
	err = database.DB.Where("id = ? AND active = ?", planID, true).First(&plan).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("套餐 %d 不存在或未激活，跳过赠送", planID)
			return nil
		}
		return fmt.Errorf("查询套餐失败: %v", err)
	}

	// 开始事务处理套餐赠送
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建或获取用户钱包
	wallet, err := createOrGetUserWallet(tx, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("创建用户钱包失败: %v", err)
	}

	// 创建订阅记录
	subscription := models.Subscription{
		UserID:             userID,
		SubscriptionPlanID: uint(planID),
		Status:             "active",
		ActivatedAt:        time.Now(),
		TotalPoints:        plan.PointAmount,
		AvailablePoints:    plan.PointAmount,
		UsedPoints:         0,
		DailyMaxPoints:     plan.DailyCheckinPointsMax,
		SourceType:         "admin_grant",
		SourceID:           "registration_gift",
	}

	// 计算过期时间
	if plan.ValidityDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, plan.ValidityDays)
		subscription.ExpiresAt = expiresAt
	} else {
		// 如果没有有效期限制，设置为10年后过期
		subscription.ExpiresAt = time.Now().AddDate(10, 0, 0)
	}

	err = tx.Create(&subscription).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("创建订阅记录失败: %v", err)
	}

	// 更新钱包信息
	err = updateWalletForNewSubscription(tx, wallet, &subscription, &plan)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("更新钱包失败: %v", err)
	}

	// 创建赠送记录
	giftRecord := models.GiftRecord{
		FromAdminID:        nil, // 系统赠送使用 NULL
		ToUserID:           userID,
		SubscriptionPlanID: uint(planID),
		PointsAmount:       plan.PointAmount,
		ValidityDays:       plan.ValidityDays,
		DailyMaxPoints:     plan.DailyCheckinPointsMax,
		Reason:             fmt.Sprintf("新用户注册自动赠送 (%s)", registrationType),
		Status:             "completed",
		SubscriptionID:     &subscription.ID,
	}

	err = tx.Create(&giftRecord).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("创建赠送记录失败: %v", err)
	}

	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	log.Printf("用户 %d (%s注册) 成功赠送套餐: planID=%d, points=%d, validityDays=%d",
		userID, registrationType, planID, plan.PointAmount, plan.ValidityDays)

	return nil
}

// createOrGetUserWallet 创建或获取用户钱包（事务版本）
func createOrGetUserWallet(tx *gorm.DB, userID uint) (*models.UserWallet, error) {
	var wallet models.UserWallet
	err := tx.Where("user_id = ?", userID).First(&wallet).Error
	if err == nil {
		return &wallet, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("查询用户钱包失败: %v", err)
	}

	// 创建新钱包
	wallet = models.UserWallet{
		UserID:            userID,
		TotalPoints:       0,
		AvailablePoints:   0,
		UsedPoints:        0,
		AccumulatedTokens: 0,
		DailyMaxPoints:    0,
		WalletExpiresAt:   time.Now(),
		Status:            "expired",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err = tx.Create(&wallet).Error
	if err != nil {
		return nil, fmt.Errorf("创建用户钱包失败: %v", err)
	}

	return &wallet, nil
}

// updateWalletForNewSubscription 为新订阅更新钱包（事务版本）
func updateWalletForNewSubscription(tx *gorm.DB, wallet *models.UserWallet, subscription *models.Subscription, plan *models.SubscriptionPlan) error {
	now := time.Now()
	updates := map[string]interface{}{
		"updated_at": now,
	}

	// 增加积分
	if subscription.TotalPoints > 0 {
		updates["total_points"] = gorm.Expr("total_points + ?", subscription.TotalPoints)
		updates["available_points"] = gorm.Expr("available_points + ?", subscription.TotalPoints)
	}

	// 更新每日最大积分限制（取最大值）
	if plan.DailyCheckinPointsMax > wallet.DailyMaxPoints {
		updates["daily_max_points"] = plan.DailyCheckinPointsMax
	}

	// 更新钱包过期时间（取最远的过期时间）
	if subscription.ExpiresAt.After(wallet.WalletExpiresAt) {
		updates["wallet_expires_at"] = subscription.ExpiresAt
		updates["status"] = "active"
	}

	// 如果钱包还未激活且有积分，激活钱包
	if wallet.Status != "active" && subscription.TotalPoints > 0 {
		updates["status"] = "active"
	}

	return tx.Model(wallet).Where("user_id = ?", wallet.UserID).Updates(updates).Error
}

// GetPlanIDForRegistrationType 根据注册类型获取套餐ID（辅助函数）
func GetPlanIDForRegistrationType(registrationType string) (int, error) {
	mapping, err := GetRegistrationPlanMapping()
	if err != nil {
		return -1, err
	}

	switch registrationType {
	case "default":
		return mapping.Default, nil
	case "linux_do":
		return mapping.LinuxDo, nil
	case "github":
		return mapping.GitHub, nil
	case "google":
		return mapping.Google, nil
	default:
		return mapping.Default, nil
	}
}

// ValidateRegistrationPlanMapping 验证注册套餐映射配置格式
func ValidateRegistrationPlanMapping(configValue string) error {
	var mapping RegistrationPlanMapping
	err := json.Unmarshal([]byte(configValue), &mapping)
	if err != nil {
		return fmt.Errorf("JSON格式错误: %v", err)
	}

	// 检查套餐ID是否有效（-1或正整数）
	planIDs := []int{mapping.Default, mapping.LinuxDo, mapping.GitHub, mapping.Google}
	for _, planID := range planIDs {
		if planID < -1 {
			return fmt.Errorf("套餐ID必须是-1（不赠送）或正整数")
		}
		
		// 如果不是-1，检查套餐是否存在
		if planID != -1 {
			var plan models.SubscriptionPlan
			err := database.DB.Where("id = ?", planID).First(&plan).Error
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("套餐ID %d 不存在", planID)
			} else if err != nil {
				return fmt.Errorf("查询套餐ID %d 失败: %v", planID, err)
			}
		}
	}

	return nil
}