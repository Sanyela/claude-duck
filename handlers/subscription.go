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
	ID                string                `json:"id"`
	Plan              SubscriptionPlanData  `json:"plan"`
	Status            string                `json:"status"`
	CurrentPeriodEnd  string                `json:"currentPeriodEnd"`
	CancelAtPeriodEnd bool                  `json:"cancelAtPeriodEnd"`
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
			ID:            subscription.Plan.PlanID,
			Name:          subscription.Plan.Name,
			PricePerMonth: subscription.Plan.PricePerMonth,
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

// HandleRedeemCoupon 兑换优惠码
func HandleRedeemCoupon(c *gin.Context) {
	// 验证token并获取用户ID
	_, err := getUserIDFromToken(c)
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

	// 简单的优惠码验证（实际项目中应该有更复杂的逻辑）
	if req.CouponCode == "VALID_COUPON" {
		// 创建或更新订阅
		// 这里可以实现真实的优惠码兑换逻辑
		c.JSON(http.StatusOK, RedeemCouponResponse{
			Success: true,
			Message: "优惠码兑换成功！专业版订阅已激活。",
		})
	} else {
		c.JSON(http.StatusOK, RedeemCouponResponse{
			Success: false,
			Message: "无效的优惠码或已过期。",
		})
	}
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