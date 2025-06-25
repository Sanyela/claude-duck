package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"claude/database"
	"claude/models"

	"github.com/gin-gonic/gin"
)

// 积分相关响应结构
type CreditBalanceResponse struct {
	Balance CreditBalanceData `json:"balance"`
}

type CreditBalanceData struct {
	Available             int  `json:"available"`               // 可用积分
	Total                 int  `json:"total"`                   // 历史总充值积分
	Used                  int  `json:"used"`                    // 实际已使用积分
	Expired               int  `json:"expired"`                 // 已过期积分
	IsCurrentSubscription bool `json:"is_current_subscription"` // 是否为当前活跃订阅
}

type ModelCostsResponse struct {
	Costs []ModelCostData `json:"costs"`
}

type ModelCostData struct {
	ID          string   `json:"id"`
	ModelName   string   `json:"modelName"`
	Status      string   `json:"status"`
	CostFactor  *float64 `json:"costFactor,omitempty"`
	Description string   `json:"description,omitempty"`
}

type CreditUsageHistoryResponse struct {
	History     []CreditUsageData `json:"history"`
	TotalPages  int               `json:"totalPages"`
	CurrentPage int               `json:"currentPage"`
}

type CreditUsageData struct {
	ID           string `json:"id"`
	Amount       int    `json:"amount"`
	Timestamp    string `json:"timestamp"`
	RelatedModel string `json:"relatedModel,omitempty"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

// HandleGetCreditBalance 获取积分余额
func HandleGetCreditBalance(c *gin.Context) {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 获取用户当前活跃订阅
	var subscription models.Subscription
	var subscriptionStartTime time.Time
	var isCurrentSubscription bool

	// 先查找状态为active且未过期的订阅
	err = database.DB.Where("user_id = ? AND status = 'active' AND current_period_end > ?", userID, time.Now()).First(&subscription).Error
	if err != nil {
		// 如果没有有效的活跃订阅，查找最近一个订阅（包括过期的）
		err = database.DB.Where("user_id = ?", userID).Order("created_at DESC").First(&subscription).Error
		if err != nil {
			// 如果真的没有任何订阅记录，使用用户注册时间
			var user models.User
			if err := database.DB.Where("id = ?", userID).First(&user).Error; err == nil {
				subscriptionStartTime = user.CreatedAt
			} else {
				subscriptionStartTime = time.Now().AddDate(0, -1, 0) // 默认最近一个月
			}
			isCurrentSubscription = false
		} else {
			// 使用最近一个订阅的创建时间
			subscriptionStartTime = subscription.CreatedAt
			// 检查这个订阅是否真的是当前有效的
			isCurrentSubscription = subscription.Status == "active" && subscription.CurrentPeriodEnd.After(time.Now())
		}
	} else {
		// 有有效的活跃订阅
		subscriptionStartTime = subscription.CreatedAt
		isCurrentSubscription = true
	}

	// 为了避免毫秒级时间差导致的问题，统计时间稍微向前推一点
	queryStartTime := subscriptionStartTime.Add(-time.Second)

	// 额外调试信息：查看所有积分池和API交易
	var allPools []models.PointPool
	database.DB.Where("user_id = ?", userID).Find(&allPools)

	var allTransactions []models.APITransaction
	database.DB.Where("user_id = ? AND status = 'success'", userID).Order("created_at DESC").Limit(10).Find(&allTransactions)

	// 只统计当前订阅周期内的积分池
	var validPoints int64
	err = database.DB.Model(&models.PointPool{}).
		Where("user_id = ? AND points_remaining > 0 AND expires_at > ? AND created_at >= ?",
			userID, time.Now(), queryStartTime).
		Select("COALESCE(SUM(points_remaining), 0)").
		Scan(&validPoints).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate valid points"})
		return
	}

	// 计算当前订阅周期内的已过期积分
	var expiredPoints int64
	err = database.DB.Model(&models.PointPool{}).
		Where("user_id = ? AND points_remaining > 0 AND expires_at <= ? AND created_at >= ?",
			userID, time.Now(), queryStartTime).
		Select("COALESCE(SUM(points_remaining), 0)").
		Scan(&expiredPoints).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate expired points"})
		return
	}

	// 计算当前订阅周期内的总充值积分
	var totalPoints int64
	err = database.DB.Model(&models.PointPool{}).
		Where("user_id = ? AND created_at >= ?", userID, queryStartTime).
		Select("COALESCE(SUM(points_total), 0)").
		Scan(&totalPoints).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate total points"})
		return
	}

	// 计算当前订阅周期内的已使用积分
	var usedPoints int64
	err = database.DB.Model(&models.APITransaction{}).
		Where("user_id = ? AND status = 'success' AND created_at >= ?", userID, queryStartTime).
		Select("COALESCE(SUM(points_used), 0)").
		Scan(&usedPoints).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate used points"})
		return
	}

	// 转换为响应格式 - 只显示当前订阅周期的数据
	balanceData := CreditBalanceData{
		Available:             int(validPoints),   // 当前可用积分
		Total:                 int(totalPoints),   // 当前订阅周期总充值积分
		Used:                  int(usedPoints),    // 当前订阅周期已使用积分
		Expired:               int(expiredPoints), // 当前订阅周期已过期积分
		IsCurrentSubscription: isCurrentSubscription,
	}

	c.JSON(http.StatusOK, CreditBalanceResponse{Balance: balanceData})
}

// HandleGetModelCosts 获取模型成本配置
func HandleGetModelCosts(c *gin.Context) {
	_, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 在新的积分系统中，所有模型使用统一的倍率，这里返回默认的模型列表
	defaultCosts := []ModelCostData{
		{
			ID:          "claude-3-opus-20240229",
			ModelName:   "Claude 3 Opus",
			Status:      "available",
			CostFactor:  floatPtr(1.0), // 统一倍率
			Description: "最强大的模型，适合复杂任务",
		},
		{
			ID:          "claude-3-sonnet-20240229",
			ModelName:   "Claude 3 Sonnet",
			Status:      "available",
			CostFactor:  floatPtr(1.0), // 统一倍率
			Description: "平衡性能与速度的模型",
		},
		{
			ID:          "claude-3-haiku-20240307",
			ModelName:   "Claude 3 Haiku",
			Status:      "available",
			CostFactor:  floatPtr(1.0), // 统一倍率
			Description: "快速响应的轻量级模型",
		},
	}

	c.JSON(http.StatusOK, ModelCostsResponse{Costs: defaultCosts})
}

// HandleGetCreditUsageHistory 获取积分使用历史
func HandleGetCreditUsageHistory(c *gin.Context) {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 获取用户当前活跃订阅，确定统计起始时间
	var subscription models.Subscription
	var subscriptionStartTime time.Time
	// 先查找状态为active且未过期的订阅
	err = database.DB.Where("user_id = ? AND status = 'active' AND current_period_end > ?", userID, time.Now()).First(&subscription).Error
	if err != nil {
		// 如果没有有效的活跃订阅，查找最近一个订阅（包括过期的）
		err = database.DB.Where("user_id = ?", userID).Order("created_at DESC").First(&subscription).Error
		if err != nil {
			// 如果真的没有任何订阅记录，使用用户注册时间
			var user models.User
			if err := database.DB.Where("id = ?", userID).First(&user).Error; err == nil {
				subscriptionStartTime = user.CreatedAt
			} else {
				subscriptionStartTime = time.Now().AddDate(0, -1, 0) // 默认最近一个月
			}
		} else {
			// 使用最近一个订阅的创建时间
			subscriptionStartTime = subscription.CreatedAt
		}
	} else {
		// 有有效的活跃订阅
		subscriptionStartTime = subscription.CreatedAt
	}

	// 为了避免毫秒级时间差导致的问题，统计时间稍微向前推一点
	queryStartTime := subscriptionStartTime.Add(-time.Second)

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", c.DefaultQuery("pageSize", "10")))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取日期筛选参数
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// 构建查询条件 - 只查询当前订阅周期内的记录
	query := database.DB.Model(&models.APITransaction{}).Where("user_id = ? AND created_at >= ?", userID, queryStartTime)

	// 添加日期筛选
	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	// 查询总数
	var total int64
	query.Count(&total)

	// 查询分页数据
	var apiTransactions []models.APITransaction
	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&apiTransactions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch usage history"})
		return
	}

	// 转换为响应格式
	var historyData []CreditUsageData
	for _, transaction := range apiTransactions {
		historyData = append(historyData, CreditUsageData{
			ID:           fmt.Sprintf("%d", transaction.ID),
			Amount:       -int(transaction.PointsUsed), // 负数表示消耗
			Timestamp:    transaction.CreatedAt.Format(time.RFC3339),
			RelatedModel: transaction.Model,
			InputTokens:  int(transaction.InputTokens),
			OutputTokens: int(transaction.OutputTokens),
		})
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, CreditUsageHistoryResponse{
		History:     historyData,
		TotalPages:  totalPages,
		CurrentPage: page,
	})
}

// 辅助函数
func floatPtr(f float64) *float64 {
	return &f
}
