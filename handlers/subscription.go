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
	ServiceLevel    string            `json:"serviceLevel,omitempty"`    // upgrade, downgrade, same_level
	Warning         string            `json:"warning,omitempty"`         // 警告信息
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

// HandleGetActiveSubscription 获取用户钱包信息（替代原订阅查询）
func HandleGetActiveSubscription(c *gin.Context) {
	// 验证token并获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 获取用户钱包信息
	wallet, err := utils.GetOrCreateUserWallet(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "查询钱包信息失败",
		})
		return
	}

	// 更新钱包状态
	utils.UpdateWalletStatus(userID)

	// 获取用户的有效兑换记录（用于显示历史）
	records, err := utils.GetWalletActiveRedemptionRecords(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "查询兑换记录失败",
		})
		return
	}

	// 转换为返回格式
	var result []SubscriptionData
	now := time.Now()

	if wallet.Status == "active" && wallet.WalletExpiresAt.After(now) {
		// 钱包有效，构建统一的订阅数据
		walletSubscription := SubscriptionData{
			ID: fmt.Sprintf("WALLET-%d", userID),
			Plan: SubscriptionPlanData{
				ID:       "WALLET-PLAN",
				Name:     "统一钱包",
				Features: []string{},
			},
			Status:            wallet.Status,
			CurrentPeriodEnd:  wallet.WalletExpiresAt.Format(time.RFC3339),
			CancelAtPeriodEnd: false,
			TotalPoints:       wallet.TotalPoints,
			UsedPoints:        wallet.UsedPoints,
			AvailablePoints:   wallet.AvailablePoints,
			ActivatedAt:       wallet.CreatedAt.Format(time.RFC3339),
			DetailedStatus:    "有效",
			IsCurrentUsing:    true,
		}
		result = append(result, walletSubscription)

		// 添加兑换记录作为历史信息
		for _, record := range records {
			if record.ExpiresAt.After(now) {
				recordData := SubscriptionData{
					ID: fmt.Sprintf("RECORD-%d", record.ID),
					Plan: SubscriptionPlanData{
						ID:       fmt.Sprintf("PLAN-%d", record.ID),
						Name:     record.Reason,
						Features: []string{},
					},
					Status:            "active",
					CurrentPeriodEnd:  record.ExpiresAt.Format(time.RFC3339),
					CancelAtPeriodEnd: false,
					TotalPoints:       record.PointsAmount,
					UsedPoints:        0, // 兑换记录不记录已使用
					AvailablePoints:   record.PointsAmount,
					ActivatedAt:       record.ActivatedAt.Format(time.RFC3339),
					DetailedStatus:    "已合并到钱包",
					IsCurrentUsing:    false,
				}
				result = append(result, recordData)
			}
		}
	} else {
		// 钱包已过期或无效
		expiredWallet := SubscriptionData{
			ID: fmt.Sprintf("WALLET-%d", userID),
			Plan: SubscriptionPlanData{
				ID:       "WALLET-PLAN",
				Name:     "统一钱包",
				Features: []string{},
			},
			Status:            "expired",
			CurrentPeriodEnd:  wallet.WalletExpiresAt.Format(time.RFC3339),
			CancelAtPeriodEnd: false,
			TotalPoints:       wallet.TotalPoints,
			UsedPoints:        wallet.UsedPoints,
			AvailablePoints:   0, // 过期后可用积分为0
			ActivatedAt:       wallet.CreatedAt.Format(time.RFC3339),
			DetailedStatus:    "已过期",
			IsCurrentUsing:    false,
		}
		result = append(result, expiredWallet)
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

// HandleGetSubscriptionHistory 获取钱包兑换历史
func HandleGetSubscriptionHistory(c *gin.Context) {
	// 验证token并获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 获取分页参数
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsedSize, err := strconv.Atoi(ps); err == nil && parsedSize > 0 && parsedSize <= 100 {
			pageSize = parsedSize
		}
	}

	// 查询用户的所有兑换记录
	records, total, err := utils.GetWalletRedemptionHistory(userID, pageSize, (page-1)*pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取兑换历史失败"})
		return
	}

	// 转换为响应格式
	var historyData []PaymentHistoryData
	now := time.Now()
	
	for _, record := range records {
		// 判断状态：有效、已过期
		subscriptionStatus := "已过期"
		if record.ExpiresAt.After(now) {
			subscriptionStatus = "有效"
		}

		// 根据来源类型确定支付状态
		paymentStatus := "paid"
		switch record.SourceType {
		case "activation_code":
			paymentStatus = "code_redeemed"
		case "admin_gift":
			paymentStatus = "admin_gifted"
		case "daily_checkin":
			paymentStatus = "checkin_reward"
		case "payment":
			paymentStatus = "paid"
		}

		// 获取计划名称
		planName := record.Reason
		if record.SubscriptionPlan != nil {
			planName = record.SubscriptionPlan.Title
		}

		historyData = append(historyData, PaymentHistoryData{
			ID:                 fmt.Sprintf("REC-%d", record.ID),
			PlanName:           planName,
			Date:               record.ActivatedAt.Format(time.RFC3339),
			PaymentStatus:      paymentStatus,
			SubscriptionStatus: subscriptionStatus,
			InvoiceURL:         record.InvoiceURL,
		})
	}

	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"history":      historyData,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
		"total_pages":  totalPages,
	})
}

// HandleRedeemCoupon 兑换激活码
func HandleRedeemCoupon(c *gin.Context) {
	// 验证token并获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 检查用户是否被禁用
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
			Success: false,
			Message: "获取用户信息失败",
		})
		return
	}

	if user.IsDisabled {
		c.JSON(http.StatusForbidden, RedeemCouponResponse{
			Success: false,
			Message: "您的账户已被管理员禁用，无法兑换激活码",
		})
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

	// 判断服务等级（在兑换前）
	wallet, err := utils.GetUserWallet(userID)
	var serviceLevel string
	var warning string
	
	if err == nil && wallet.Status == "active" {
		serviceLevel = utils.DetermineServiceLevel(wallet, &activationCode.Plan)
		switch serviceLevel {
		case "same_level":
			warning = "同等级兑换将重置您的积分余额，之前未使用的积分将被清空。"
		case "downgrade":
			warning = "新套餐的签到奖励积分低于当前套餐，强制兑换可能会丢失签到奖励。"
		}
	} else {
		serviceLevel = "upgrade" // 新用户或过期用户默认为升级
	}

	// 使用新的钱包架构兑换激活码
	err = utils.RedeemActivationCodeToWallet(userID, &activationCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
			Success: false,
			Message: fmt.Sprintf("激活码兑换失败: %s", err.Error()),
		})
		return
	}

	// 更新用户的服务降级配置（如果未锁定）
	if !user.DegradationLocked && activationCode.Plan.DegradationGuaranteed > user.DegradationGuaranteed {
		database.DB.Model(&user).Updates(map[string]interface{}{
			"degradation_guaranteed": activationCode.Plan.DegradationGuaranteed,
			"degradation_source":     "subscription",
		})
	}

	c.JSON(http.StatusOK, RedeemCouponResponse{
		Success: true,
		Message: fmt.Sprintf("激活码兑换成功！已充值 %d 积分，有效期 %d 天。",
			activationCode.Plan.PointAmount,
			activationCode.Plan.ValidityDays),
		ServiceLevel: serviceLevel,
		Warning:      warning,
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

	// 获取用户钱包的签到积分配置
	pointsRange := getWalletCheckinPointsRange(userID)

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

	// 获取最后签到日期（从钱包中获取）
	wallet, err := utils.GetUserWallet(userID)
	lastCheckinDate := ""
	if err == nil && wallet.LastCheckinDate != "" {
		lastCheckinDate = wallet.LastCheckinDate
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

	// 检查用户是否被禁用
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, CheckinResponse{
			Success:      false,
			Message:      "获取用户信息失败",
			RewardPoints: 0,
		})
		return
	}

	if user.IsDisabled {
		c.JSON(http.StatusForbidden, CheckinResponse{
			Success:      false,
			Message:      "您的账户已被管理员禁用，无法签到",
			RewardPoints: 0,
		})
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

	// 获取用户钱包的签到积分配置
	pointsRange := getWalletCheckinPointsRange(userID)
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

	// 创建签到记录
	checkinRecord := models.DailyCheckin{
		UserID:      userID,
		CheckinDate: today,
		Points:      rewardPoints,
		CreatedAt:   time.Now(),
	}

	if err := database.DB.Create(&checkinRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, CheckinResponse{
			Success:      false,
			Message:      "签到记录创建失败",
			RewardPoints: 0,
		})
		return
	}

	// 使用新的钱包架构添加签到积分
	err = utils.DailyCheckinToWallet(userID, rewardPoints)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CheckinResponse{
			Success:      false,
			Message:      fmt.Sprintf("签到积分添加失败: %s", err.Error()),
			RewardPoints: 0,
		})
		return
	}

	c.JSON(http.StatusOK, CheckinResponse{
		Success:      true,
		Message:      fmt.Sprintf("签到成功！获得 %d 积分奖励", rewardPoints),
		RewardPoints: rewardPoints,
	})
}

// getUserCheckinPointsRange 获取用户的签到积分范围 (已弃用，使用getWalletCheckinPointsRange)
func getUserCheckinPointsRange(userID uint) CheckinPointsRange {
	// 此函数已被 getWalletCheckinPointsRange 替代
	return getWalletCheckinPointsRange(userID)
}

// getWalletCheckinPointsRange 获取钱包签到积分范围
func getWalletCheckinPointsRange(userID uint) CheckinPointsRange {
	// 获取用户钱包
	wallet, err := utils.GetUserWallet(userID)
	if err != nil {
		return CheckinPointsRange{MinPoints: 0, MaxPoints: 0, HasValid: false}
	}

	// 检查钱包是否有效且有签到配置
	if wallet.Status != "active" || 
	   wallet.WalletExpiresAt.Before(time.Now()) ||
	   wallet.DailyCheckinPoints <= 0 {
		return CheckinPointsRange{MinPoints: 0, MaxPoints: 0, HasValid: false}
	}

	// 确保最大值不小于最小值
	maxPoints := wallet.DailyCheckinPointsMax
	if maxPoints <= 0 || maxPoints < wallet.DailyCheckinPoints {
		maxPoints = wallet.DailyCheckinPoints
	}

	return CheckinPointsRange{
		MinPoints: wallet.DailyCheckinPoints,
		MaxPoints: maxPoints,
		HasValid:  true,
	}
}
