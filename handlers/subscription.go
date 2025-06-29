package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// HandleGetActiveSubscription 获取用户所有已激活的订阅
func HandleGetActiveSubscription(c *gin.Context) {
	// 验证token并获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 查询用户所有已激活的订阅（包括过期的）
	var subscriptions []models.Subscription
	err = database.DB.Preload("Plan").Where("user_id = ?", userID).Order("activated_at ASC").Find(&subscriptions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get subscriptions"})
		return
	}

	if len(subscriptions) == 0 {
		c.JSON(http.StatusOK, ActiveSubscriptionResponse{Subscriptions: []SubscriptionData{}})
		return
	}

	// 找到当前正在使用的订阅（按到期时间排序：最早到期且仍有积分的订阅）
	var currentUsingSubscription *models.Subscription
	// 创建一个有效订阅的副本，按到期时间排序
	var validSubscriptions []models.Subscription
	for _, sub := range subscriptions {
		if sub.AvailablePoints > 0 && sub.ExpiresAt.After(time.Now()) {
			validSubscriptions = append(validSubscriptions, sub)
		}
	}

	// 如果有有效订阅，按到期时间排序，找到最早到期的
	if len(validSubscriptions) > 0 {
		// 找到最早到期的订阅
		earliestExpiry := validSubscriptions[0]
		for _, sub := range validSubscriptions {
			if sub.ExpiresAt.Before(earliestExpiry.ExpiresAt) {
				earliestExpiry = sub
			}
		}
		currentUsingSubscription = &earliestExpiry
	}

	// 转换为响应格式
	var subscriptionDataList []SubscriptionData
	for _, subscription := range subscriptions {
		// 解析features JSON
		var features []string
		if subscription.Plan.Features != "" {
			json.Unmarshal([]byte(subscription.Plan.Features), &features)
		}

		// 判断详细状态
		detailedStatus := "已过期"
		if subscription.ExpiresAt.After(time.Now()) {
			if subscription.AvailablePoints > 0 {
				detailedStatus = "有效"
			} else {
				detailedStatus = "已用完"
			}
		}

		// 判断是否当前正在使用
		isCurrentUsing := false
		if currentUsingSubscription != nil && currentUsingSubscription.ID == subscription.ID {
			isCurrentUsing = true
		}

		subscriptionData := SubscriptionData{
			ID: subscription.SourceID, // 使用SourceID作为外部ID
			Plan: SubscriptionPlanData{
				ID:       fmt.Sprintf("%d", subscription.Plan.ID),
				Name:     subscription.Plan.Title,
				Features: features,
			},
			Status:            subscription.Status,
			CurrentPeriodEnd:  subscription.ExpiresAt.Format(time.RFC3339),
			CancelAtPeriodEnd: subscription.CancelAtPeriodEnd,
			AvailablePoints:   subscription.AvailablePoints,
			TotalPoints:       subscription.TotalPoints,
			UsedPoints:        subscription.UsedPoints,
			ActivatedAt:       subscription.ActivatedAt.Format(time.RFC3339),
			DetailedStatus:    detailedStatus,
			IsCurrentUsing:    isCurrentUsing,
		}

		subscriptionDataList = append(subscriptionDataList, subscriptionData)
	}

	c.JSON(http.StatusOK, ActiveSubscriptionResponse{Subscriptions: subscriptionDataList})
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
