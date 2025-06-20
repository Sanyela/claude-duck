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
			UserID:              userID,
			Available:           1000, // 默认1000积分
			Total:               1000,
			RechargeRatePerHour: 0,
			CanRequestReset:     true,
			UpdatedAt:           time.Now(),
		}
		database.DB.Create(&creditBalance)
	}

	// 转换为响应格式
	balanceData := CreditBalanceData{
		Available:           creditBalance.Available,
		Total:               creditBalance.Total,
		RechargeRatePerHour: creditBalance.RechargeRatePerHour,
		CanRequestReset:     creditBalance.CanRequestReset,
	}

	if creditBalance.NextResetTime != nil {
		timeStr := creditBalance.NextResetTime.Format(time.RFC3339)
		balanceData.NextResetTime = &timeStr
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
		costsData = append(costsData, ModelCostData{
			ID:          cost.ModelID,
			ModelName:   cost.ModelName,
			Status:      cost.Status,
			CostFactor:  cost.CostFactor,
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
			Amount:       usage.Amount,
			Timestamp:    usage.CreatedAt.Format(time.RFC3339),
			RelatedModel: usage.RelatedModel,
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
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 查询用户积分余额
	var creditBalance models.CreditBalance
	err = database.DB.Where("user_id = ?", userID).First(&creditBalance).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, CreditResetResponse{
			Success: false,
			Message: "未找到积分账户",
		})
		return
	}

	// 检查是否可以重置
	if !creditBalance.CanRequestReset {
		var nextTime *string
		if creditBalance.NextResetTime != nil {
			timeStr := creditBalance.NextResetTime.Format(time.RFC3339)
			nextTime = &timeStr
		}
		c.JSON(http.StatusOK, CreditResetResponse{
			Success:           false,
			Message:           "今日已重置过积分，请明天再试。",
			NextAvailableTime: nextTime,
		})
		return
	}

	// 执行重置
	nextResetTime := time.Now().Add(24 * time.Hour)
	err = database.DB.Model(&creditBalance).Updates(models.CreditBalance{
		Available:       creditBalance.Total,
		CanRequestReset: false,
		NextResetTime:   &nextResetTime,
		UpdatedAt:       time.Now(),
	}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, CreditResetResponse{
			Success: false,
			Message: "重置失败，请稍后重试",
		})
		return
	}

	// 记录重置历史
	resetHistory := models.CreditUsageHistory{
		UserID:      userID,
		Description: "积分重置",
		Amount:      creditBalance.Total - creditBalance.Available,
		CreatedAt:   time.Now(),
	}
	database.DB.Create(&resetHistory)

	c.JSON(http.StatusOK, CreditResetResponse{
		Success: true,
		Message: "积分已成功重置到初始额度。",
	})
}

// 辅助函数
func floatPtr(f float64) *float64 {
	return &f
} 