package handlers

import (
	"encoding/json"
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
	FreeModelUsageCount   int  `json:"free_model_usage_count"`  // 免费模型使用次数
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

	var allTransactions []models.APITransaction
	database.DB.Where("user_id = ? AND status = 'success'", userID).Order("created_at DESC").Limit(10).Find(&allTransactions)
	var validPoints, expiredPoints, totalPoints, usedPoints int64

	if isCurrentSubscription {
		// 如果有当前有效订阅，只统计未过期的积分池相关数据
		// 先获取所有未过期的积分池，用于确定统计范围
		var validPools []models.PointPool
		err = database.DB.Where("user_id = ? AND expires_at > ? AND created_at >= ?",
			userID, time.Now(), queryStartTime).Find(&validPools).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get valid pools"})
			return
		}

		// 如果有有效的积分池，找到最早的创建时间作为统计起始时间
		var validPoolStartTime time.Time
		if len(validPools) > 0 {
			validPoolStartTime = validPools[0].CreatedAt
			for _, pool := range validPools {
				if pool.CreatedAt.Before(validPoolStartTime) {
					validPoolStartTime = pool.CreatedAt
				}
			}
			// 统计时间稍微向前推一点，避免毫秒级时间差
			validPoolStartTime = validPoolStartTime.Add(-time.Second)
		} else {
			// 如果没有有效积分池，使用当前时间，这样所有统计都为0
			validPoolStartTime = time.Now()
		}

		// 可用积分：只统计未过期的积分池
		for _, pool := range validPools {
			validPoints += pool.PointsRemaining
		}

		// 总积分：只统计未过期的积分池的总量
		for _, pool := range validPools {
			totalPoints += pool.PointsTotal
		}

		// 已使用积分：只统计有效积分池创建时间之后的使用量
		err = database.DB.Model(&models.APITransaction{}).
			Where("user_id = ? AND status = 'success' AND created_at >= ?", userID, validPoolStartTime).
			Select("COALESCE(SUM(points_used), 0)").
			Scan(&usedPoints).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate used points"})
			return
		}

		// 已过期积分设为0，因为当前有活跃订阅时不显示过期积分
		expiredPoints = 0

	} else {
		// 如果没有当前有效订阅，统计上一期订阅的积分（包括过期的）
		// 可用积分：在查询时间之后创建的且未过期的积分池
		err = database.DB.Model(&models.PointPool{}).
			Where("user_id = ? AND points_remaining > 0 AND expires_at > ? AND created_at >= ?",
				userID, time.Now(), queryStartTime).
			Select("COALESCE(SUM(points_remaining), 0)").
			Scan(&validPoints).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate valid points"})
			return
		}

		// 已过期积分：在查询时间之后创建的但已过期的积分池
		err = database.DB.Model(&models.PointPool{}).
			Where("user_id = ? AND points_remaining > 0 AND expires_at <= ? AND created_at >= ?",
				userID, time.Now(), queryStartTime).
			Select("COALESCE(SUM(points_remaining), 0)").
			Scan(&expiredPoints).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate expired points"})
			return
		}

		// 总充值积分：在查询时间之后创建的所有积分池
		err = database.DB.Model(&models.PointPool{}).
			Where("user_id = ? AND created_at >= ?", userID, queryStartTime).
			Select("COALESCE(SUM(points_total), 0)").
			Scan(&totalPoints).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate total points"})
			return
		}

		// 已使用积分：在查询时间之后的所有消费
		err = database.DB.Model(&models.APITransaction{}).
			Where("user_id = ? AND status = 'success' AND created_at >= ?", userID, queryStartTime).
			Select("COALESCE(SUM(points_used), 0)").
			Scan(&usedPoints).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to calculate used points"})
			return
		}
	}

	// 获取用户免费模型使用次数
	var user models.User
	var freeModelUsageCount int64
	err = database.DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		freeModelUsageCount = 0 // 如果获取失败，默认为0
	} else {
		freeModelUsageCount = user.FreeModelUsageCount
	}

	// 转换为响应格式
	balanceData := CreditBalanceData{
		Available:             int(validPoints),   // 当前可用积分
		Total:                 int(totalPoints),   // 总充值积分
		Used:                  int(usedPoints),    // 已使用积分
		Expired:               int(expiredPoints), // 已过期积分
		IsCurrentSubscription: isCurrentSubscription,
		FreeModelUsageCount:   int(freeModelUsageCount), // 免费模型使用次数
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

// 辅助函数
func floatPtr(f float64) *float64 {
	return &f
}
