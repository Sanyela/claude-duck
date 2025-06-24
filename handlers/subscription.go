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
	Subscription *SubscriptionData `json:"subscription"`
}

type SubscriptionData struct {
	ID                string               `json:"id"`
	Plan              SubscriptionPlanData `json:"plan"`
	Status            string               `json:"status"`
	CurrentPeriodEnd  string               `json:"currentPeriodEnd"`
	CancelAtPeriodEnd bool                 `json:"cancelAtPeriodEnd"`
}

type SubscriptionPlanData struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	PricePerMonth float64  `json:"pricePerMonth"`
	Currency      string   `json:"currency"`
	Features      []string `json:"features"`
}

type SubscriptionHistoryResponse struct {
	History []PaymentHistoryData `json:"history"`
}

type PaymentHistoryData struct {
	ID         string  `json:"id"`
	PlanName   string  `json:"planName"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Date       string  `json:"date"`
	Status     string  `json:"status"`
	InvoiceURL string  `json:"invoiceUrl,omitempty"`
}

type RedeemCouponRequest struct {
	CouponCode string `json:"couponCode" binding:"required"`
}

type RedeemCouponResponse struct {
	Success         bool              `json:"success"`
	Message         string            `json:"message"`
	NewSubscription *SubscriptionData `json:"newSubscription,omitempty"`
}

// HandleGetActiveSubscription 获取用户活跃订阅
func HandleGetActiveSubscription(c *gin.Context) {
	// 验证token并获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 查询用户活跃订阅
	var subscription models.Subscription
	err = database.DB.Preload("Plan").Where("user_id = ? AND status = 'active'", userID).First(&subscription).Error
	if err != nil {
		// 没有找到活跃订阅
		c.JSON(http.StatusOK, ActiveSubscriptionResponse{Subscription: nil})
		return
	}

	// 解析features JSON
	var features []string
	if subscription.Plan.Features != "" {
		json.Unmarshal([]byte(subscription.Plan.Features), &features)
	}

	subscriptionData := &SubscriptionData{
		ID: subscription.ExternalID,
		Plan: SubscriptionPlanData{
			ID:            fmt.Sprintf("%d", subscription.Plan.ID),
			Name:          subscription.Plan.Title,
			PricePerMonth: subscription.Plan.Price,
			Currency:      subscription.Plan.Currency,
			Features:      features,
		},
		Status:            subscription.Status,
		CurrentPeriodEnd:  subscription.CurrentPeriodEnd.Format(time.RFC3339),
		CancelAtPeriodEnd: subscription.CancelAtPeriodEnd,
	}

	c.JSON(http.StatusOK, ActiveSubscriptionResponse{Subscription: subscriptionData})
}

// HandleGetSubscriptionHistory 获取订阅历史
func HandleGetSubscriptionHistory(c *gin.Context) {
	// 验证token并获取用户ID
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 查询支付历史
	var paymentHistory []models.PaymentHistory
	err = database.DB.Where("user_id = ?", userID).Order("payment_date DESC").Find(&paymentHistory).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch subscription history"})
		return
	}

	// 转换为响应格式
	var historyData []PaymentHistoryData
	for _, payment := range paymentHistory {
		historyData = append(historyData, PaymentHistoryData{
			ID:         fmt.Sprintf("%d", payment.ID),
			PlanName:   payment.PlanName,
			Amount:     payment.Amount,
			Currency:   payment.Currency,
			Date:       payment.PaymentDate.Format(time.RFC3339),
			Status:     payment.Status,
			InvoiceURL: payment.InvoiceURL,
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

	// 创建积分池记录
	pointPool := models.PointPool{
		UserID:          userID,
		SourceType:      "activation_code",
		SourceID:        fmt.Sprintf("AC-%d", activationCode.ID),
		PointsTotal:     activationCode.Plan.PointAmount,
		PointsRemaining: activationCode.Plan.PointAmount,
		ExpiresAt:       time.Now().AddDate(0, 0, activationCode.Plan.ValidityDays),
	}
	if err := tx.Create(&pointPool).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
			Success: false,
			Message: "积分充值失败。",
		})
		return
	}

	// 更新用户积分余额汇总
	var pointBalance models.PointBalance
	err = tx.Where("user_id = ?", userID).First(&pointBalance).Error
	if err != nil {
		// 创建新的积分余额记录
		pointBalance = models.PointBalance{
			UserID:          userID,
			TotalPoints:     activationCode.Plan.PointAmount,
			UsedPoints:      0,
			AvailablePoints: activationCode.Plan.PointAmount,
		}
		if err := tx.Create(&pointBalance).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
				Success: false,
				Message: "更新积分余额失败。",
			})
			return
		}
	} else {
		// 更新现有积分余额
		pointBalance.TotalPoints += activationCode.Plan.PointAmount
		pointBalance.AvailablePoints += activationCode.Plan.PointAmount
		if err := tx.Save(&pointBalance).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
				Success: false,
				Message: "更新积分余额失败。",
			})
			return
		}
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

	// 创建或更新订阅记录
	var subscription models.Subscription
	err = tx.Preload("Plan").Where("user_id = ? AND status = 'active'", userID).First(&subscription).Error
	if err != nil {
		// 创建新的订阅记录
		subscription = models.Subscription{
			UserID:             userID,
			SubscriptionPlanID: activationCode.Plan.ID,
			ExternalID:         fmt.Sprintf("AC-%d-%d", activationCode.ID, time.Now().Unix()),
			Status:             "active",
			CurrentPeriodEnd:   time.Now().AddDate(0, 0, activationCode.Plan.ValidityDays),
			CancelAtPeriodEnd:  false,
		}
		if err := tx.Create(&subscription).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
				Success: false,
				Message: "创建订阅记录失败。",
			})
			return
		}
	} else {
		// 更新现有订阅记录，延长有效期
		newEndDate := subscription.CurrentPeriodEnd.AddDate(0, 0, activationCode.Plan.ValidityDays)
		subscription.CurrentPeriodEnd = newEndDate
		subscription.CancelAtPeriodEnd = false
		// 如果新计划的等级更高，则更新订阅计划
		if activationCode.Plan.PointAmount > subscription.Plan.PointAmount {
			subscription.SubscriptionPlanID = activationCode.Plan.ID
		}
		if err := tx.Save(&subscription).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
				Success: false,
				Message: "更新订阅记录失败。",
			})
			return
		}
	}

	// 记录支付历史
	paymentHistory := models.PaymentHistory{
		UserID:      userID,
		PlanName:    activationCode.Plan.Title,
		Amount:      0, // 激活码兑换，金额为0
		Currency:    activationCode.Plan.Currency,
		Status:      "paid",
		PaymentDate: time.Now(),
	}
	if err := tx.Create(&paymentHistory).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
			Success: false,
			Message: "记录支付历史失败。",
		})
		return
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
