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
	Available int `json:"available"`
	Total     int `json:"total"`
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
	Description  string `json:"description"`
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

	// 查询用户积分余额
	var pointBalance models.PointBalance
	err = database.DB.Where("user_id = ?", userID).First(&pointBalance).Error
	if err != nil {
		// 如果没有记录，创建默认余额
		pointBalance = models.PointBalance{
			UserID:          userID,
			TotalPoints:     0, // 默认0余额，防止批量注册
			UsedPoints:      0,
			AvailablePoints: 0,
			UpdatedAt:       time.Now(),
		}
		database.DB.Create(&pointBalance)
	}

	// 转换为响应格式
	balanceData := CreditBalanceData{
		Available: int(pointBalance.AvailablePoints), // 直接使用积分
		Total:     int(pointBalance.TotalPoints),     // 直接使用积分
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
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 查询总数
	var total int64
	database.DB.Model(&models.APITransaction{}).Where("user_id = ?", userID).Count(&total)

	// 查询分页数据
	var apiTransactions []models.APITransaction
	offset := (page - 1) * pageSize
	err = database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
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
		// 生成描述信息
		description := fmt.Sprintf("使用 %s 模型，消耗 %d 输入tokens，%d 输出tokens",
			transaction.Model, transaction.InputTokens, transaction.OutputTokens)

		// 如果有缓存相关的tokens，添加到描述中
		if transaction.CacheCreationInputTokens > 0 || transaction.CacheReadInputTokens > 0 {
			description += fmt.Sprintf("，缓存创建 %d tokens，缓存读取 %d tokens",
				transaction.CacheCreationInputTokens, transaction.CacheReadInputTokens)
		}

		historyData = append(historyData, CreditUsageData{
			ID:           fmt.Sprintf("%d", transaction.ID),
			Description:  description,
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
