package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"claude/database"
	"claude/models"
	"claude/utils"

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
	FreeModelUsageCount   int  `json:"free_model_usage_count"`  // 免费模型使用次数
	CheckinPoints         int  `json:"checkin_points"`          // 签到积分
	AdminGiftPoints       int  `json:"admin_gift_points"`       // 管理员赠送积分
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
	ID                  string          `json:"id"`
	Amount              int             `json:"amount"`
	Timestamp           string          `json:"timestamp"`
	RelatedModel        string          `json:"relatedModel,omitempty"`
	InputTokens         int             `json:"input_tokens"`
	OutputTokens        int             `json:"output_tokens"`
	CacheCreationTokens int             `json:"cache_creation_tokens,omitempty"`
	CacheReadTokens     int             `json:"cache_read_tokens,omitempty"`
	TotalCacheTokens    int             `json:"total_cache_tokens,omitempty"`
	BillingDetails      *BillingDetails `json:"billing_details,omitempty"`
}

// BillingDetails 计费详情
type BillingDetails struct {
	InputMultiplier      float64 `json:"input_multiplier"`       // 输入token倍率
	OutputMultiplier     float64 `json:"output_multiplier"`      // 输出token倍率
	CacheMultiplier      float64 `json:"cache_multiplier"`       // 缓存token倍率
	WeightedInputTokens  float64 `json:"weighted_input_tokens"`  // 加权后的输入tokens
	WeightedOutputTokens float64 `json:"weighted_output_tokens"` // 加权后的输出tokens
	WeightedCacheTokens  float64 `json:"weighted_cache_tokens"`  // 加权后的缓存tokens
	TotalWeightedTokens  float64 `json:"total_weighted_tokens"`  // 总加权tokens
	FinalPoints          int64   `json:"final_points"`           // 最终扣除积分
	PricingTableUsed     bool    `json:"pricing_table_used"`     // 是否使用了阶梯计费表
}

// PricingTable 计费表响应结构
type PricingTableResponse struct {
	PricingTable map[string]int `json:"pricing_table"` // token阈值 -> 积分的映射
	Description  string         `json:"description"`   // 说明
}

// HandleGetCreditBalance 获取积分余额
func HandleGetCreditBalance(c *gin.Context) {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 检查用户是否被禁用
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取用户信息失败"})
		return
	}

	if user.IsDisabled {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "您的账户已被管理员禁用，无法查看积分信息"})
		return
	}

	// 获取用户当前活跃订阅（过滤掉签到和管理员赠送）
	var activeSubscriptions []models.Subscription
	err = database.DB.Where("user_id = ? AND status = 'active' AND expires_at > ? AND source_type IN (?)",
		userID, time.Now(), []string{"activation_code", "payment"}).Find(&activeSubscriptions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get active subscriptions"})
		return
	}

	// 获取签到积分（有效的签到订阅）
	var checkinSubscriptions []models.Subscription
	err = database.DB.Where("user_id = ? AND status = 'active' AND expires_at > ? AND source_type = ?",
		userID, time.Now(), "daily_checkin").Find(&checkinSubscriptions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get checkin subscriptions"})
		return
	}

	// 获取管理员赠送积分（有效的管理员赠送订阅）
	var adminGiftSubscriptions []models.Subscription
	err = database.DB.Where("user_id = ? AND status = 'active' AND expires_at > ? AND source_type = ?",
		userID, time.Now(), "admin_gift").Find(&adminGiftSubscriptions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get admin gift subscriptions"})
		return
	}

	var validPoints, totalPoints, usedPoints int64
	var checkinPoints, adminGiftPoints int64
	isCurrentSubscription := len(activeSubscriptions) > 0

	if isCurrentSubscription {
		// 如果有活跃订阅，统计所有活跃订阅的积分
		for _, sub := range activeSubscriptions {
			validPoints += sub.AvailablePoints
			totalPoints += sub.TotalPoints
			usedPoints += sub.UsedPoints
		}
	}

	// 统计签到积分
	for _, sub := range checkinSubscriptions {
		checkinPoints += sub.AvailablePoints
	}

	// 统计管理员赠送积分
	for _, sub := range adminGiftSubscriptions {
		adminGiftPoints += sub.AvailablePoints
	}

	// 如果没有活跃订阅，所有积分都应该是0

	// 获取用户免费模型使用次数
	var freeModelUsageCount int64
	err = database.DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		freeModelUsageCount = 0 // 如果获取失败，默认为0
	} else {
		freeModelUsageCount = user.FreeModelUsageCount
	}

	// 转换为响应格式 - 不显示过期积分，只显示活跃订阅的积分
	balanceData := CreditBalanceData{
		Available:             int(validPoints), // 只显示活跃订阅的可用积分
		Total:                 int(totalPoints), // 只显示活跃订阅的总积分
		Used:                  int(usedPoints),  // 只显示活跃订阅的已使用积分
		Expired:               0,                // 不再显示过期积分
		IsCurrentSubscription: isCurrentSubscription,
		FreeModelUsageCount:   int(freeModelUsageCount), // 免费模型使用次数
		CheckinPoints:         int(checkinPoints),       // 签到积分
		AdminGiftPoints:       int(adminGiftPoints),     // 管理员赠送积分
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

	// 构建查询条件 - 查询用户的所有API交易记录
	query := database.DB.Model(&models.APITransaction{}).Where("user_id = ?", userID)

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
		// 计算总缓存token
		totalCacheTokens := transaction.CacheCreationInputTokens + transaction.CacheReadInputTokens

		// 计算加权后的tokens
		weightedInputTokens := float64(transaction.InputTokens) * transaction.InputMultiplier
		weightedOutputTokens := float64(transaction.OutputTokens) * transaction.OutputMultiplier
		weightedCacheTokens := float64(totalCacheTokens) * transaction.CacheMultiplier
		totalWeightedTokens := weightedInputTokens + weightedOutputTokens + weightedCacheTokens

		// 构建计费详情
		billingDetails := &BillingDetails{
			InputMultiplier:      transaction.InputMultiplier,
			OutputMultiplier:     transaction.OutputMultiplier,
			CacheMultiplier:      transaction.CacheMultiplier,
			WeightedInputTokens:  weightedInputTokens,
			WeightedOutputTokens: weightedOutputTokens,
			WeightedCacheTokens:  weightedCacheTokens,
			TotalWeightedTokens:  totalWeightedTokens,
			FinalPoints:          transaction.PointsUsed,
			PricingTableUsed:     true, // 新系统都使用阶梯计费表
		}

		historyData = append(historyData, CreditUsageData{
			ID:                  fmt.Sprintf("%d", transaction.ID),
			Amount:              -int(transaction.PointsUsed), // 负数表示消耗
			Timestamp:           transaction.CreatedAt.Format(time.RFC3339),
			RelatedModel:        transaction.Model,
			InputTokens:         int(transaction.InputTokens),
			OutputTokens:        int(transaction.OutputTokens),
			CacheCreationTokens: int(transaction.CacheCreationInputTokens),
			CacheReadTokens:     int(transaction.CacheReadInputTokens),
			TotalCacheTokens:    totalCacheTokens,
			BillingDetails:      billingDetails,
		})
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, CreditUsageHistoryResponse{
		History:     historyData,
		TotalPages:  totalPages,
		CurrentPage: page,
	})
}

// HandleGetPricingTable 获取计费表配置
func HandleGetPricingTable(c *gin.Context) {
	_, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 获取计费表配置
	var config models.SystemConfig
	err = database.DB.Where("config_key = ?", "token_pricing_table").First(&config).Error
	if err != nil {
		// 如果没有配置，返回默认计费表
		defaultTable := map[string]int{
			"0":      2,
			"7680":   3,
			"15360":  4,
			"23040":  5,
			"30720":  6,
			"38400":  7,
			"46080":  8,
			"53760":  9,
			"61440":  10,
			"69120":  11,
			"76800":  12,
			"84480":  13,
			"92160":  14,
			"99840":  15,
			"107520": 16,
			"115200": 17,
			"122880": 18,
			"130560": 19,
			"138240": 20,
			"145920": 21,
			"153600": 22,
			"161280": 23,
			"168960": 24,
			"176640": 25,
			"184320": 25,
			"192000": 25,
			"200000": 25,
		}

		response := PricingTableResponse{
			PricingTable: defaultTable,
			Description:  "基于加权Token总数的阶梯计费表",
		}

		c.JSON(http.StatusOK, response)
		return
	}

	// 解析JSON配置
	var pricingTable map[string]int
	if err := json.Unmarshal([]byte(config.ConfigValue), &pricingTable); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to parse pricing table"})
		return
	}

	response := PricingTableResponse{
		PricingTable: pricingTable,
		Description:  config.Description,
	}

	c.JSON(http.StatusOK, response)
}

// HandleGetDailyUsage 获取用户今日积分使用情况
func HandleGetDailyUsage(c *gin.Context) {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 获取用户今日使用情况
	usages, err := utils.GetUserDailyPointsUsage(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取每日使用情况失败"})
		return
	}

	// 获取用户的活跃订阅信息
	var activeSubscriptions []models.Subscription
	err = database.DB.Preload("Plan").
		Where("user_id = ? AND status = 'active' AND expires_at > ?", userID, time.Now()).
		Find(&activeSubscriptions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取订阅信息失败"})
		return
	}

	// 构建响应数据
	type DailyUsageInfo struct {
		SubscriptionID   uint   `json:"subscription_id"`
		SubscriptionName string `json:"subscription_name"`
		PointsUsed       int64  `json:"points_used"`
		DailyLimit       int64  `json:"daily_limit"`
		RemainingPoints  int64  `json:"remaining_points"`
		HasLimit         bool   `json:"has_limit"`
	}

	today := time.Now().Format("2006-01-02")
	usageMap := make(map[uint]int64)
	for _, usage := range usages {
		usageMap[usage.SubscriptionID] = usage.PointsUsed
	}

	var dailyUsageList []DailyUsageInfo
	for _, sub := range activeSubscriptions {
		dailyLimit := sub.DailyMaxPoints
		if dailyLimit == 0 {
			dailyLimit = sub.Plan.DailyMaxPoints
		}

		pointsUsed := usageMap[sub.ID]
		hasLimit := dailyLimit > 0
		var remainingPoints int64

		if hasLimit {
			remainingPoints = dailyLimit - pointsUsed
			if remainingPoints < 0 {
				remainingPoints = 0
			}
		}

		dailyUsageList = append(dailyUsageList, DailyUsageInfo{
			SubscriptionID:   sub.ID,
			SubscriptionName: sub.Plan.Title,
			PointsUsed:       pointsUsed,
			DailyLimit:       dailyLimit,
			RemainingPoints:  remainingPoints,
			HasLimit:         hasLimit,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"usage_date": today,
		"usage_list": dailyUsageList,
	})
}

// 辅助函数
func floatPtr(f float64) *float64 {
	return &f
}
