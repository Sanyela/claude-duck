package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"claude/database"
	"claude/models"
	"claude/utils"

	"github.com/gin-gonic/gin"
)

// 订阅相关响应结构
type ActiveSubscriptionResponse struct {
	Subscriptions []SubscriptionData `json:"subscriptions"`
}

type SubscriptionData struct {
	ID                string               `json:"id"`
	Plan              SubscriptionPlanData `json:"plan"`
	Status            string               `json:"status"`
	CurrentPeriodEnd  string               `json:"currentPeriodEnd"`
	CancelAtPeriodEnd bool                 `json:"cancelAtPeriodEnd"`
	AvailablePoints   int64                `json:"availablePoints"`
	TotalPoints       int64                `json:"totalPoints"`
	UsedPoints        int64                `json:"usedPoints"`
	ActivatedAt       string               `json:"activatedAt"`
	DetailedStatus    string               `json:"detailedStatus"` // 有效、已用完、已过期
	IsCurrentUsing    bool                 `json:"isCurrentUsing"` // 是否当前正在使用
}

type SubscriptionPlanData struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Features []string `json:"features"`
}

type SubscriptionHistoryResponse struct {
	History []PaymentHistoryData `json:"history"`
}

type PaymentHistoryData struct {
	ID                 string `json:"id"`
	PlanName           string `json:"planName"`
	Date               string `json:"date"`
	PaymentStatus      string `json:"paymentStatus"`      // 支付状态：paid, failed
	SubscriptionStatus string `json:"subscriptionStatus"` // 订阅状态：active, expired
	InvoiceURL         string `json:"invoiceUrl,omitempty"`
}

type RedeemCouponRequest struct {
	CouponCode string `json:"couponCode" binding:"required"`
}

type RedeemCouponResponse struct {
	Success         bool              `json:"success"`
	Message         string            `json:"message"`
	NewSubscription *SubscriptionData `json:"newSubscription,omitempty"`
}

// 签到相关响应结构
type CheckinPointsRange struct {
	MinPoints int64 `json:"minPoints"`
	MaxPoints int64 `json:"maxPoints"`
	HasValid  bool  `json:"hasValid"`
}

type CheckinStatusResponse struct {
	CanCheckin      bool               `json:"canCheckin"`      // 是否可以签到
	TodayChecked    bool               `json:"todayChecked"`    // 今天是否已签到
	LastCheckinDate string             `json:"lastCheckinDate"` // 最后签到日期
	PointsRange     CheckinPointsRange `json:"pointsRange"`     // 积分范围
}

type CheckinResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	RewardPoints int64  `json:"rewardPoints"` // 获得的奖励积分
}

// HandleGetActiveSubscription 获取用户所有已激活的订阅
func HandleGetActiveSubscription(c *gin.Context) {
	// 验证token并获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 获取用户的所有订阅记录，按到期时间排序
	var subscriptions []models.Subscription
	err = database.DB.Preload("Plan").
		Where("user_id = ? AND status = 'active'", userID).
		Where("source_type IN (?)", []string{"activation_code", "payment"}). // 只返回真正的订阅
		Order("expires_at ASC").
		Find(&subscriptions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "查询订阅信息失败",
		})
		return
	}

	// 转换为返回格式
	var result []SubscriptionData
	now := time.Now()

	// 找到当前正在使用的订阅（按到期时间最早原则）
	var currentSubscriptionIndex = -1
	for i, sub := range subscriptions {
		if sub.ExpiresAt.After(now) && sub.AvailablePoints > 0 {
			currentSubscriptionIndex = i
			break
		}
	}

	for i, sub := range subscriptions {
		// 计算详细状态
		detailedStatus := "已过期"
		if sub.ExpiresAt.After(now) {
			if sub.AvailablePoints > 0 {
				detailedStatus = "有效"
			} else {
				detailedStatus = "已用完"
			}
		}

		// 是否是当前正在消耗的订阅
		isCurrentlyUsed := (i == currentSubscriptionIndex)

		subscriptionData := SubscriptionData{
			ID: fmt.Sprintf("SUB-%d-%d", sub.UserID, sub.ID),
			Plan: SubscriptionPlanData{
				ID:       fmt.Sprintf("PLAN-%03d", sub.SubscriptionPlanID),
				Name:     getSubscriptionPlanName(sub),
				Features: []string{},
			},
			Status:            sub.Status,
			CurrentPeriodEnd:  sub.ExpiresAt.Format(time.RFC3339),
			CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
			TotalPoints:       sub.TotalPoints,
			UsedPoints:        sub.UsedPoints,
			AvailablePoints:   sub.AvailablePoints,
			ActivatedAt:       sub.ActivatedAt.Format(time.RFC3339),
			DetailedStatus:    detailedStatus,
			IsCurrentUsing:    isCurrentlyUsed,
		}

		result = append(result, subscriptionData)
	}

	c.JSON(http.StatusOK, ActiveSubscriptionResponse{
		Subscriptions: result,
	})
}

// getSubscriptionPlanName 获取订阅计划名称
func getSubscriptionPlanName(subscription models.Subscription) string {
	if subscription.Plan.Title != "" {
		return subscription.Plan.Title
	}
	// 如果没有预加载或者为空，返回默认名称
	return fmt.Sprintf("套餐-%d", subscription.SubscriptionPlanID)
}

// HandleGetSubscriptionHistory 获取订阅历史
func HandleGetSubscriptionHistory(c *gin.Context) {
	// 验证token并获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 直接查询用户的所有订阅记录
	var subscriptions []models.Subscription
	err = database.DB.Preload("Plan").Where("user_id = ?", userID).Order("activated_at DESC").Find(&subscriptions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch subscription history"})
		return
	}

	// 转换为响应格式
	var historyData []PaymentHistoryData
	for _, subscription := range subscriptions {
		// 判断订阅状态：有效、已用完、已过期
		subscriptionStatus := "已过期"
		if subscription.ExpiresAt.After(time.Now()) {
			if subscription.AvailablePoints > 0 {
				subscriptionStatus = "有效"
			} else {
				subscriptionStatus = "已用完"
			}
		}

		historyData = append(historyData, PaymentHistoryData{
			ID:                 fmt.Sprintf("%d", subscription.ID),
			PlanName:           subscription.Plan.Title,
			Date:               subscription.ActivatedAt.Format(time.RFC3339),
			PaymentStatus:      "paid", // 订阅记录默认都是已支付状态
			SubscriptionStatus: subscriptionStatus,
			InvoiceURL:         subscription.InvoiceURL,
		})
	}

	c.JSON(http.StatusOK, SubscriptionHistoryResponse{History: historyData})
}

// HandleRedeemCoupon 兑换激活码
func HandleRedeemCoupon(c *gin.Context) {
	// 验证token并获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	var req RedeemCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RedeemCouponResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// 查找激活码
	var activationCode models.ActivationCode
	err = database.DB.Preload("Plan").Where("code = ? AND status = ?", req.CouponCode, "unused").First(&activationCode).Error
	if err != nil {
		c.JSON(http.StatusOK, RedeemCouponResponse{
			Success: false,
			Message: "无效的激活码或已被使用。",
		})
		return
	}

	// 检查订阅计划是否启用
	if !activationCode.Plan.Active {
		c.JSON(http.StatusOK, RedeemCouponResponse{
			Success: false,
			Message: "该订阅计划已停用。",
		})
		return
	}

	// 开始事务
	tx := database.DB.Begin()

	now := time.Now()
	if err := tx.Model(&activationCode).Updates(map[string]interface{}{
		"status":          "used",
		"used_by_user_id": userID,
		"used_at":         now,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
			Success: false,
			Message: "激活码兑换失败。",
		})
		return
	}

	// 创建新的订阅记录
	subscription := models.Subscription{
		UserID:             userID,
		SubscriptionPlanID: activationCode.Plan.ID,
		Status:             "active",
		ActivatedAt:        now,
		ExpiresAt:          now.AddDate(0, 0, activationCode.Plan.ValidityDays),
		TotalPoints:        activationCode.Plan.PointAmount,
		UsedPoints:         0,
		AvailablePoints:    activationCode.Plan.PointAmount,
		SourceType:         "activation_code",
		SourceID:           fmt.Sprintf("AC-%d", activationCode.ID),
		InvoiceURL:         "", // 激活码兑换无发票
		CancelAtPeriodEnd:  false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := tx.Create(&subscription).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
			Success: false,
			Message: "创建订阅记录失败。",
		})
		return
	}

	// 更新用户的服务降级配置（如果未锁定）
	var user models.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err == nil {
		if !user.DegradationLocked && activationCode.Plan.DegradationGuaranteed > user.DegradationGuaranteed {
			user.DegradationGuaranteed = activationCode.Plan.DegradationGuaranteed
			user.DegradationSource = "subscription"
			tx.Save(&user)
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, RedeemCouponResponse{
		Success: true,
		Message: fmt.Sprintf("激活码兑换成功！已充值 %d 积分，有效期 %d 天。",
			activationCode.Plan.PointAmount,
			activationCode.Plan.ValidityDays),
	})
}

// getUserIDFromToken 从Authorization header中提取用户ID
func getUserIDFromToken(c *gin.Context) (uint, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, fmt.Errorf("authorization header required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return 0, fmt.Errorf("invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := utils.ValidateAccessToken(token)
	if err != nil {
		return 0, fmt.Errorf("invalid or expired token")
	}

	return claims.UserID, nil
}

// HandleGetCheckinStatus 获取签到状态
func HandleGetCheckinStatus(c *gin.Context) {
	// 验证token并获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 检查签到功能是否启用
	var enabledConfig models.SystemConfig
	err = database.DB.Where("config_key = ?", "daily_checkin_enabled").First(&enabledConfig).Error
	if err != nil || enabledConfig.ConfigValue != "true" {
		c.JSON(http.StatusOK, CheckinStatusResponse{
			CanCheckin:      false,
			TodayChecked:    false,
			LastCheckinDate: "",
			PointsRange:     CheckinPointsRange{MinPoints: 0, MaxPoints: 0, HasValid: false},
		})
		return
	}

	// 获取用户有效订阅的签到积分配置
	pointsRange := getUserCheckinPointsRange(userID)

	// 如果没有有效的签到积分配置，不显示签到功能
	if !pointsRange.HasValid {
		c.JSON(http.StatusOK, CheckinStatusResponse{
			CanCheckin:      false,
			TodayChecked:    false,
			LastCheckinDate: "",
			PointsRange:     pointsRange,
		})
		return
	}

	// 检查今天是否已经签到
	today := time.Now().Format("2006-01-02")
	var todayCheckin models.DailyCheckin
	err = database.DB.Where("user_id = ? AND checkin_date = ?", userID, today).First(&todayCheckin).Error
	todayChecked := err == nil

	// 获取最后签到日期
	var lastCheckin models.DailyCheckin
	err = database.DB.Where("user_id = ?", userID).Order("checkin_date DESC").First(&lastCheckin).Error
	lastCheckinDate := ""
	if err == nil {
		lastCheckinDate = lastCheckin.CheckinDate
	}

	c.JSON(http.StatusOK, CheckinStatusResponse{
		CanCheckin:      !todayChecked,
		TodayChecked:    todayChecked,
		LastCheckinDate: lastCheckinDate,
		PointsRange:     pointsRange,
	})
}

// HandleDailyCheckin 执行每日签到
func HandleDailyCheckin(c *gin.Context) {
	// 验证token并获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 检查签到功能是否启用
	var enabledConfig models.SystemConfig
	err = database.DB.Where("config_key = ?", "daily_checkin_enabled").First(&enabledConfig).Error
	if err != nil || enabledConfig.ConfigValue != "true" {
		c.JSON(http.StatusOK, CheckinResponse{
			Success:      false,
			Message:      "签到功能已关闭",
			RewardPoints: 0,
		})
		return
	}

	// 获取用户有效订阅的签到积分配置
	pointsRange := getUserCheckinPointsRange(userID)
	if !pointsRange.HasValid {
		c.JSON(http.StatusOK, CheckinResponse{
			Success:      false,
			Message:      "当前没有有效的签到奖励",
			RewardPoints: 0,
		})
		return
	}

	// 在范围内随机生成积分
	var rewardPoints int64
	if pointsRange.MinPoints == pointsRange.MaxPoints {
		rewardPoints = pointsRange.MinPoints
	} else {
		// 生成范围内的随机数
		randRange := pointsRange.MaxPoints - pointsRange.MinPoints + 1
		rewardPoints = pointsRange.MinPoints + int64(rand.Int63n(randRange))
	}

	// 获取签到有效期配置
	var validityConfig models.SystemConfig
	validityDays := 1 // 默认有效期1天
	if err := database.DB.Where("config_key = ?", "daily_checkin_validity_days").First(&validityConfig).Error; err == nil {
		if days, err := strconv.Atoi(validityConfig.ConfigValue); err == nil {
			validityDays = days
		}
	}

	// 检查今天是否已经签到
	today := time.Now().Format("2006-01-02")
	var existingCheckin models.DailyCheckin
	err = database.DB.Where("user_id = ? AND checkin_date = ?", userID, today).First(&existingCheckin).Error
	if err == nil {
		c.JSON(http.StatusOK, CheckinResponse{
			Success:      false,
			Message:      "今天已经签到过了",
			RewardPoints: 0,
		})
		return
	}

	// 开始事务
	tx := database.DB.Begin()

	now := time.Now()

	// 创建签到记录
	checkinRecord := models.DailyCheckin{
		UserID:      userID,
		CheckinDate: today,
		Points:      rewardPoints,
		CreatedAt:   now,
	}

	if err := tx.Create(&checkinRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, CheckinResponse{
			Success:      false,
			Message:      "签到记录创建失败",
			RewardPoints: 0,
		})
		return
	}

	// 创建积分奖励订阅记录
	// 查找或创建签到专用的虚拟订阅计划
	var checkinPlan models.SubscriptionPlan
	err = tx.Where("title = ? AND point_amount = 0", "每日签到奖励").First(&checkinPlan).Error
	if err != nil {
		// 创建签到专用计划
		checkinPlan = models.SubscriptionPlan{
			Title:                 "每日签到奖励",
			Description:           "每日签到获得的积分奖励",
			PointAmount:           0, // 特殊标记
			Price:                 0,
			Currency:              "CNY",
			ValidityDays:          validityDays,
			DegradationGuaranteed: 0,
			DailyCheckinPoints:    0, // 签到虚拟计划不参与签到积分计算
			DailyCheckinPointsMax: 0, // 签到虚拟计划不参与签到积分计算
			Features:              "[]",
			Active:                true,
		}
		if err := tx.Create(&checkinPlan).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, CheckinResponse{
				Success:      false,
				Message:      "创建签到计划失败",
				RewardPoints: 0,
			})
			return
		}
	}

	subscription := models.Subscription{
		UserID:             userID,
		SubscriptionPlanID: checkinPlan.ID,
		Status:             "active",
		ActivatedAt:        now,
		ExpiresAt:          now.AddDate(0, 0, validityDays), // 根据配置设置有效期
		TotalPoints:        rewardPoints,
		UsedPoints:         0,
		AvailablePoints:    rewardPoints,
		SourceType:         "daily_checkin",
		SourceID:           fmt.Sprintf("CHECKIN-%s-%d", today, userID),
		InvoiceURL:         "", // 签到奖励无发票
		CancelAtPeriodEnd:  false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := tx.Create(&subscription).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, CheckinResponse{
			Success:      false,
			Message:      "积分奖励创建失败",
			RewardPoints: 0,
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, CheckinResponse{
		Success:      true,
		Message:      fmt.Sprintf("签到成功！获得 %d 积分奖励", rewardPoints),
		RewardPoints: rewardPoints,
	})
}

// getUserCheckinPointsRange 获取用户的签到积分范围
func getUserCheckinPointsRange(userID uint) CheckinPointsRange {
	// 获取用户所有有效订阅
	var subscriptions []models.Subscription
	err := database.DB.Preload("Plan").
		Where("user_id = ? AND status = 'active' AND expires_at > ?", userID, time.Now()).
		Where("source_type IN (?)", []string{"activation_code", "payment"}). // 只考虑真正的订阅
		Find(&subscriptions).Error
	if err != nil || len(subscriptions) == 0 {
		return CheckinPointsRange{MinPoints: 0, MaxPoints: 0, HasValid: false}
	}

	// 收集所有有效的签到积分配置
	type PointsRange struct {
		MinPoints int64
		MaxPoints int64
	}
	var validRanges []PointsRange

	for _, sub := range subscriptions {
		// 只有当最小值 > 0 时才认为是有效的签到配置
		if sub.Plan.DailyCheckinPoints > 0 {
			maxPoints := sub.Plan.DailyCheckinPointsMax
			// 如果最大值为0或小于最小值，则设为最小值
			if maxPoints <= 0 || maxPoints < sub.Plan.DailyCheckinPoints {
				maxPoints = sub.Plan.DailyCheckinPoints
			}
			validRanges = append(validRanges, PointsRange{
				MinPoints: sub.Plan.DailyCheckinPoints,
				MaxPoints: maxPoints,
			})
		}
	}

	if len(validRanges) == 0 {
		return CheckinPointsRange{MinPoints: 0, MaxPoints: 0, HasValid: false}
	}

	// 获取多订阅策略配置
	var strategyConfig models.SystemConfig
	strategy := "highest" // 默认使用最高策略
	if err := database.DB.Where("config_key = ?", "daily_checkin_multi_subscription_strategy").First(&strategyConfig).Error; err == nil {
		strategy = strategyConfig.ConfigValue
	}

	// 根据策略选择订阅
	var selectedRange PointsRange
	if strategy == "lowest" {
		// 选择最小值最低的订阅
		selectedRange = validRanges[0]
		for _, r := range validRanges {
			if r.MinPoints < selectedRange.MinPoints {
				selectedRange = r
			}
		}
	} else {
		// 默认选择最大值最高的订阅
		selectedRange = validRanges[0]
		for _, r := range validRanges {
			if r.MaxPoints > selectedRange.MaxPoints {
				selectedRange = r
			}
		}
	}

	return CheckinPointsRange{
		MinPoints: selectedRange.MinPoints,
		MaxPoints: selectedRange.MaxPoints,
		HasValid:  true,
	}
}
