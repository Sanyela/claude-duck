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
	Available           int     `json:"available"`
	Total               int     `json:"total"`
	RechargeRatePerHour int     `json:"rechargeRatePerHour"`
	CanRequestReset     bool    `json:"canRequestReset"`
	NextResetTime       *string `json:"nextResetTime,omitempty"`
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
	History      []CreditUsageData `json:"history"`
	TotalPages   int               `json:"totalPages"`
	CurrentPage  int               `json:"currentPage"`
}

type CreditUsageData struct {
	ID           string `json:"id"`
	Description  string `json:"description"`
	Amount       int    `json:"amount"`
	Timestamp    string `json:"timestamp"`
	RelatedModel string `json:"relatedModel,omitempty"`
}

type CreditResetResponse struct {
	Success           bool    `json:"success"`
	Message           string  `json:"message"`
	NextAvailableTime *string `json:"nextAvailableTime,omitempty"`
}

// HandleGetCreditBalance 获取积分余额
func HandleGetCreditBalance(c *gin.Context) {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 查询用户积分余额
	var creditBalance models.CreditBalance
	err = database.DB.Where("user_id = ?", userID).First(&creditBalance).Error
	if err != nil {
		// 如果没有记录，创建默认余额
		creditBalance = models.CreditBalance{
			UserID:          userID,
			TotalAmount:     0,    // 默认0余额，防止批量注册
			UsedAmount:      0,
			AvailableAmount: 0,
			UpdatedAt:       time.Now(),
		}
		database.DB.Create(&creditBalance)
	}

	// 转换为响应格式
	balanceData := CreditBalanceData{
		Available:           int(creditBalance.AvailableAmount * 100), // 转换为分显示
		Total:               int(creditBalance.TotalAmount * 100),      // 转换为分显示
		RechargeRatePerHour: 0,                                        // 废弃字段，保持兼容
		CanRequestReset:     true,                                     // 废弃字段，保持兼容
	}

	// NextResetTime 已在新模型中移除，不再处理

	c.JSON(http.StatusOK, CreditBalanceResponse{Balance: balanceData})
}

// HandleGetModelCosts 获取模型成本配置
func HandleGetModelCosts(c *gin.Context) {
	_, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 查询所有活跃的模型成本配置
	var modelCosts []models.ModelCost
	err = database.DB.Where("active = ?", true).Find(&modelCosts).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch model costs"})
		return
	}

	// 如果没有数据，返回默认配置
	if len(modelCosts) == 0 {
		defaultCosts := []ModelCostData{
			{
				ID:          "claude-code",
				ModelName:   "Claude-Code",
				Status:      "unavailable",
				Description: "核心代码与对话模型",
			},
			{
				ID:          "claude-opus",
				ModelName:   "Claude 3 Opus",
				Status:      "available",
				CostFactor:  floatPtr(1.5),
				Description: "最高级模型，复杂任务处理",
			},
			{
				ID:          "claude-sonnet",
				ModelName:   "Claude 3 Sonnet",
				Status:      "available",
				CostFactor:  floatPtr(1.0),
				Description: "平衡性能与速度的模型",
			},
		}
		c.JSON(http.StatusOK, ModelCostsResponse{Costs: defaultCosts})
		return
	}

	// 转换为响应格式
	var costsData []ModelCostData
	for _, cost := range modelCosts {
		// 计算成本因子（用于兼容旧API）
		var costFactor *float64
		if cost.InputPricePerK > 0 {
			factor := cost.InputPricePerK / 0.002 // 基准价格 $0.002
			costFactor = &factor
		}
		
		costsData = append(costsData, ModelCostData{
			ID:          cost.ModelID,
			ModelName:   cost.ModelName,
			Status:      cost.Status,
			CostFactor:  costFactor,
			Description: cost.Description,
		})
	}

	c.JSON(http.StatusOK, ModelCostsResponse{Costs: costsData})
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
	database.DB.Model(&models.CreditUsageHistory{}).Where("user_id = ?", userID).Count(&total)

	// 查询分页数据
	var usageHistory []models.CreditUsageHistory
	offset := (page - 1) * pageSize
	err = database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&usageHistory).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch usage history"})
		return
	}

	// 转换为响应格式
	var historyData []CreditUsageData
	for _, usage := range usageHistory {
		historyData = append(historyData, CreditUsageData{
			ID:           fmt.Sprintf("%d", usage.ID),
			Description:  usage.Description,
			Amount:       int(usage.Amount * 100), // 转换为分显示
			Timestamp:    usage.CreatedAt.Format(time.RFC3339),
			RelatedModel: usage.ModelName, // 使用ModelName字段
		})
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, CreditUsageHistoryResponse{
		History:     historyData,
		TotalPages:  totalPages,
		CurrentPage: page,
	})
}

// HandleRequestCreditReset 申请积分重置
func HandleRequestCreditReset(c *gin.Context) {
	// 此功能已废弃，返回不支持的响应
	c.JSON(http.StatusOK, CreditResetResponse{
		Success: false,
		Message: "积分重置功能已停用，请使用激活码或联系管理员获取额度。",
	})
}

// 辅助函数
func floatPtr(f float64) *float64 {
	return &f
} 