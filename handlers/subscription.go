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
			PricePerMonth: subscription.Plan.Price,  // 使用Price字段
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

	// 检查是否过期
	if activationCode.ExpiresAt != nil && activationCode.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusOK, RedeemCouponResponse{
			Success: false,
			Message: "激活码已过期。",
		})
		return
	}

	// 开始事务
	tx := database.DB.Begin()

	// 标记激活码为已使用
	now := time.Now()
	activationCode.Status = "used"
	activationCode.UsedByUserID = &userID
	activationCode.UsedAt = &now
	if err := tx.Save(&activationCode).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
			Success: false,
			Message: "激活码兑换失败。",
		})
		return
	}

	// 根据激活码类型处理
	if activationCode.Type == "point" {
		// 积分类型：直接增加积分
		var pointBalance models.PointBalance
		err = tx.Where("user_id = ?", userID).First(&pointBalance).Error
		if err != nil {
			// 创建新的积分余额记录
			pointBalance = models.PointBalance{
				UserID:          userID,
				TotalPoints:     activationCode.PointAmount,
				UsedPoints:      0,
				AvailablePoints: activationCode.PointAmount,
			}
			if err := tx.Create(&pointBalance).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
					Success: false,
					Message: "积分充值失败。",
				})
				return
			}
		} else {
			// 更新现有积分余额
			pointBalance.TotalPoints += activationCode.PointAmount
			pointBalance.AvailablePoints += activationCode.PointAmount
			if err := tx.Save(&pointBalance).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
					Success: false,
					Message: "积分充值失败。",
				})
				return
			}
		}

		// 记录积分使用历史
		usageHistory := models.PointUsageHistory{
			UserID:       userID,
			RequestID:    fmt.Sprintf("REDEEM-%s", activationCode.Code),
			IP:           c.ClientIP(),
			UID:          fmt.Sprintf("%d", userID),
			Username:     "",
			Model:        "activation_code",
			PromptTokens: 0,
			CompletionTokens: 0,
			PromptMultiplier: 0,
			CompletionMultiplier: 0,
			PointsUsed:   -activationCode.PointAmount, // 负数表示充值
			IsRoundUp:    false,
		}
		if err := tx.Create(&usageHistory).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
				Success: false,
				Message: "记录充值历史失败。",
			})
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, RedeemCouponResponse{
			Success: true,
			Message: fmt.Sprintf("激活码兑换成功！已充值 %d 积分。", activationCode.PointAmount),
		})
	} else if activationCode.Type == "plan" && activationCode.Plan != nil {
		// 套餐类型：创建订阅
		// 检查是否已有活跃订阅
		var existingSubscription models.Subscription
		err = tx.Where("user_id = ? AND status = ?", userID, "active").First(&existingSubscription).Error
		if err == nil {
			tx.Rollback()
			c.JSON(http.StatusOK, RedeemCouponResponse{
				Success: false,
				Message: "您已有活跃的订阅套餐。",
			})
			return
		}

		// 创建新订阅
		subscription := models.Subscription{
			UserID:             userID,
			SubscriptionPlanID: *activationCode.SubscriptionPlanID,
			ExternalID:         fmt.Sprintf("SUB-%d-%d", userID, time.Now().Unix()),
			Status:             "active",
			CurrentPeriodEnd:   time.Now().AddDate(0, 0, activationCode.Plan.ValidityDays),
			CancelAtPeriodEnd:  false,
		}
		if err := tx.Create(&subscription).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, RedeemCouponResponse{
				Success: false,
				Message: "创建订阅失败。",
			})
			return
		}

		// 增加套餐包含的积分
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
					Message: "套餐积分充值失败。",
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
					Message: "套餐积分充值失败。",
				})
				return
			}
		}

		// 记录支付历史
		paymentHistory := models.PaymentHistory{
			UserID:      userID,
			PlanName:    activationCode.Plan.Name,
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

		// 返回新订阅信息
		var features []string
		if activationCode.Plan.Features != "" {
			json.Unmarshal([]byte(activationCode.Plan.Features), &features)
		}

		newSubscription := &SubscriptionData{
			ID: subscription.ExternalID,
			Plan: SubscriptionPlanData{
				ID:            activationCode.Plan.PlanID,
				Name:          activationCode.Plan.Name,
				PricePerMonth: activationCode.Plan.Price,
				Currency:      activationCode.Plan.Currency,
				Features:      features,
			},
			Status:            subscription.Status,
			CurrentPeriodEnd:  subscription.CurrentPeriodEnd.Format(time.RFC3339),
			CancelAtPeriodEnd: subscription.CancelAtPeriodEnd,
		}

		c.JSON(http.StatusOK, RedeemCouponResponse{
			Success:         true,
			Message:         fmt.Sprintf("激活码兑换成功！已激活 %s 套餐，包含 %d 积分。", activationCode.Plan.Name, activationCode.Plan.PointAmount),
			NewSubscription: newSubscription,
		})
	} else {
		tx.Rollback()
		c.JSON(http.StatusOK, RedeemCouponResponse{
			Success: false,
			Message: "激活码类型无效。",
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