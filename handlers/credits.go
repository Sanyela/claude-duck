package handlers

import (
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
	Available             int             `json:"available"`               // 可用积分
	Total                 int             `json:"total"`                   // 历史总充值积分
	Used                  int             `json:"used"`                    // 实际已使用积分
	Expired               int             `json:"expired"`                 // 已过期积分
	IsCurrentSubscription bool            `json:"is_current_subscription"` // 是否为当前活跃订阅
	FreeModelUsageCount   int             `json:"free_model_usage_count"`  // 免费模型使用次数
	CheckinPoints         int             `json:"checkin_points"`          // 签到积分
	AdminGiftPoints       int             `json:"admin_gift_points"`       // 管理员赠送积分
	AccumulatedTokens     int64           `json:"accumulated_tokens"`      // 累计token数量
	AutoRefill            *AutoRefillInfo `json:"auto_refill,omitempty"`   // 自动补给信息
}

type AutoRefillInfo struct {
	Enabled        bool    `json:"enabled"`          // 是否启用自动补给
	Threshold      int64   `json:"threshold"`        // 补给阈值
	Amount         int64   `json:"amount"`           // 补给数量
	NeedsRefill    bool    `json:"needs_refill"`     // 当前是否需要补给
	NextRefillTime *string `json:"next_refill_time"` // 下次补给时间
	LastRefillTime *string `json:"last_refill_time"` // 上次补给时间
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
	Amount              float64         `json:"amount"`
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

	// 使用新的钱包架构获取积分信息
	wallet, err := utils.GetOrCreateUserWallet(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取用户钱包失败"})
		return
	}


	// 更新钱包状态
	utils.UpdateWalletStatus(userID)

	// 获取分类积分信息（从兑换记录中统计）
	var checkinPoints, adminGiftPoints int64

	// 统计有效的签到积分
	var checkinRecords []models.RedemptionRecord
	err = database.DB.Where("user_id = ? AND source_type = ? AND expires_at > ?",
		userID, "daily_checkin", time.Now()).Find(&checkinRecords).Error
	if err == nil {
		for _, record := range checkinRecords {
			checkinPoints += record.PointsAmount
		}
	}

	// 统计有效的管理员赠送积分
	var adminGiftRecords []models.RedemptionRecord
	err = database.DB.Where("user_id = ? AND source_type = ? AND expires_at > ?",
		userID, "admin_gift", time.Now()).Find(&adminGiftRecords).Error
	if err == nil {
		for _, record := range adminGiftRecords {
			adminGiftPoints += record.PointsAmount
		}
	}

	// 判断是否有有效的订阅
	isCurrentSubscription := wallet.Status == "active" && wallet.WalletExpiresAt.After(time.Now())

	// 获取用户免费模型使用次数
	freeModelUsageCount := user.FreeModelUsageCount

	// 计算自动补给信息
	autoRefillInfo := calculateAutoRefillInfo(wallet)

	// 转换为响应格式 - 使用钱包数据
	balanceData := CreditBalanceData{
		Available:             int(wallet.AvailablePoints), // 钱包可用积分
		Total:                 int(wallet.TotalPoints),     // 钱包总积分
		Used:                  int(wallet.UsedPoints),      // 钱包已使用积分
		Expired:               0,                           // 不再显示过期积分
		IsCurrentSubscription: isCurrentSubscription,
		FreeModelUsageCount:   int(freeModelUsageCount), // 免费模型使用次数
		CheckinPoints:         int(checkinPoints),       // 签到积分
		AdminGiftPoints:       int(adminGiftPoints),     // 管理员赠送积分
		AccumulatedTokens:     wallet.AccumulatedTokens, // 累计token数量
		AutoRefill:            autoRefillInfo,           // 自动补给信息
	}

	c.JSON(http.StatusOK, CreditBalanceResponse{Balance: balanceData})
}

// calculateAutoRefillInfo 计算自动补给信息
func calculateAutoRefillInfo(wallet *models.UserWallet) *AutoRefillInfo {
	if !wallet.AutoRefillEnabled {
		return nil
	}

	// 检查是否需要补给
	needsRefill := wallet.AvailablePoints <= wallet.AutoRefillThreshold

	// 计算下次补给时间
	var nextRefillTime *string
	if needsRefill {
		// 如果需要补给，计算下次补给时间点
		now := time.Now()
		nextTime := calculateNextRefillTime(now, wallet.LastAutoRefillTime)
		timeStr := nextTime.Format("2006-01-02 15:04:05")
		nextRefillTime = &timeStr
	}

	// 格式化上次补给时间
	var lastRefillTime *string
	if wallet.LastAutoRefillTime != nil {
		timeStr := wallet.LastAutoRefillTime.Format("2006-01-02 15:04:05")
		lastRefillTime = &timeStr
	}

	return &AutoRefillInfo{
		Enabled:        wallet.AutoRefillEnabled,
		Threshold:      wallet.AutoRefillThreshold,
		Amount:         wallet.AutoRefillAmount,
		NeedsRefill:    needsRefill,
		NextRefillTime: nextRefillTime,
		LastRefillTime: lastRefillTime,
	}
}

// calculateNextRefillTime 计算下次补给时间
func calculateNextRefillTime(now time.Time, lastRefillTime *time.Time) time.Time {
	// 补给时间点：0点、4点、8点、12点、16点、20点
	refillHours := []int{0, 4, 8, 12, 16, 20}

	// 如果有上次补给时间，需要确保至少间隔4小时
	if lastRefillTime != nil {
		timeSinceLastRefill := now.Sub(*lastRefillTime)
		if timeSinceLastRefill < 4*time.Hour {
			// 如果距离上次补给不足4小时，找到4小时后的下一个补给时间点
			nextValidTime := lastRefillTime.Add(4 * time.Hour)
			return findNextRefillTimeAfter(nextValidTime, refillHours)
		}
	}

	// 找到当前时间后的下一个补给时间点
	return findNextRefillTimeAfter(now, refillHours)
}

// findNextRefillTimeAfter 找到指定时间后的下一个补给时间点
func findNextRefillTimeAfter(after time.Time, refillHours []int) time.Time {
	year, month, day := after.Date()

	// 在当天寻找下一个补给时间点
	for _, hour := range refillHours {
		refillTime := time.Date(year, month, day, hour, 0, 0, 0, after.Location())
		if refillTime.After(after) {
			return refillTime
		}
	}

	// 如果当天没有更多的补给时间点，返回明天的第一个补给时间点
	nextDay := after.Add(24 * time.Hour)
	year, month, day = nextDay.Date()
	return time.Date(year, month, day, refillHours[0], 0, 0, 0, after.Location())
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

		// 计算这次调用的"累计进度积分"
		threshold, pointsPerThreshold, _ := utils.GetTokenThresholdConfig()
		progressPoints := float64(0)
		if threshold > 0 {
			progressPoints = (totalWeightedTokens / float64(threshold)) * float64(pointsPerThreshold)
		}

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
			PricingTableUsed:     false, // 新系统使用累计token计费
		}

		historyData = append(historyData, CreditUsageData{
			ID:                  fmt.Sprintf("%d", transaction.ID),
			Amount:              -progressPoints, // 显示进度积分
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

	// 获取新的累计token计费配置
	threshold, pointsPerThreshold, err := utils.GetTokenThresholdConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取计费配置失败"})
		return
	}

	// 构建响应数据
	type ThresholdPricingResponse struct {
		TokenThreshold     int64  `json:"token_threshold"`      // 计费阈值
		PointsPerThreshold int64  `json:"points_per_threshold"` // 每阈值积分
		Description        string `json:"description"`          // 说明
	}

	response := ThresholdPricingResponse{
		TokenThreshold:     threshold,
		PointsPerThreshold: pointsPerThreshold,
		Description:        fmt.Sprintf("累计%d个加权token扣除%d积分", threshold, pointsPerThreshold),
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

	// 使用新的钱包架构获取每日使用情况
	usedToday, dailyLimit, err := utils.GetUserDailyUsage(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取每日使用情况失败"})
		return
	}

	// 获取用户钱包信息
	wallet, err := utils.GetOrCreateUserWallet(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取用户钱包失败"})
		return
	}

	// 构建响应数据
	type DailyUsageInfo struct {
		WalletID        uint   `json:"wallet_id"`
		WalletName      string `json:"wallet_name"`
		PointsUsed      int64  `json:"points_used"`
		DailyLimit      int64  `json:"daily_limit"`
		RemainingPoints int64  `json:"remaining_points"`
		HasLimit        bool   `json:"has_limit"`
	}

	today := time.Now().Format("2006-01-02")
	hasLimit := dailyLimit > 0
	var remainingPoints int64

	if hasLimit {
		remainingPoints = dailyLimit - usedToday
		if remainingPoints < 0 {
			remainingPoints = 0
		}
	}

	dailyUsageInfo := DailyUsageInfo{
		WalletID:        wallet.UserID,
		WalletName:      "用户钱包",
		PointsUsed:      usedToday,
		DailyLimit:      dailyLimit,
		RemainingPoints: remainingPoints,
		HasLimit:        hasLimit,
	}

	c.JSON(http.StatusOK, gin.H{
		"usage_date": today,
		"usage_info": dailyUsageInfo,
	})
}

// 辅助函数
func floatPtr(f float64) *float64 {
	return &f
}
