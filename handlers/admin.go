package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"claude/database"
	"claude/models"

	"github.com/gin-gonic/gin"
)

// Pagination 分页参数
type Pagination struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// getPagination 获取分页参数
func getPagination(c *gin.Context) Pagination {
	var pagination Pagination
	pagination.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pagination.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 {
		pagination.PageSize = 10
	}
	if pagination.PageSize > 100 {
		pagination.PageSize = 100
	}

	return pagination
}

// HandleAdminGetUsers 获取用户列表
func HandleAdminGetUsers(c *gin.Context) {
	pagination := getPagination(c)
	var users []models.User
	var total int64

	// 查询总数
	database.DB.Model(&models.User{}).Count(&total)

	// 分页查询
	offset := (pagination.Page - 1) * pagination.PageSize
	database.DB.Offset(offset).Limit(pagination.PageSize).Find(&users)

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       users,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	})
}

// HandleAdminUpdateUser 更新用户信息
func HandleAdminUpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var updateData struct {
		IsAdmin *bool  `json:"is_admin"`
		Email   string `json:"email"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if updateData.IsAdmin != nil {
		updates["is_admin"] = *updateData.IsAdmin
	}
	if updateData.Email != "" {
		updates["email"] = updateData.Email
	}

	result := database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// HandleAdminDeleteUser 删除用户
func HandleAdminDeleteUser(c *gin.Context) {
	userID := c.Param("id")
	
	result := database.DB.Delete(&models.User{}, userID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}


// HandleAdminGetActivationCodes 获取激活码列表
func HandleAdminGetActivationCodes(c *gin.Context) {
	pagination := getPagination(c)
	var codes []models.ActivationCode
	var total int64

	query := database.DB.Model(&models.ActivationCode{})
	
	// 可选过滤参数
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if batchNumber := c.Query("batch_number"); batchNumber != "" {
		query = query.Where("batch_number = ?", batchNumber)
	}

	query.Count(&total)

	offset := (pagination.Page - 1) * pagination.PageSize
	query.Preload("Plan").Preload("UsedBy").Offset(offset).Limit(pagination.PageSize).Find(&codes)

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       codes,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	})
}

// HandleAdminCreateActivationCodes 批量创建激活码
func HandleAdminCreateActivationCodes(c *gin.Context) {
	var request struct {
		Count              int     `json:"count" binding:"required,min=1,max=1000"`
		SubscriptionPlanID uint    `json:"subscription_plan_id" binding:"required"`
		ExpiresAt          *string `json:"expires_at"`
		BatchNumber        string  `json:"batch_number"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证订阅计划是否存在
	var plan models.SubscriptionPlan
	if err := database.DB.Where("id = ?", request.SubscriptionPlanID).First(&plan).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订阅计划不存在"})
		return
	}

	// 生成批次号
	if request.BatchNumber == "" {
		request.BatchNumber = "BATCH-" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	codes := make([]models.ActivationCode, request.Count)
	for i := 0; i < request.Count; i++ {
		code := models.ActivationCode{
			Code:               generateActivationCode(),
			SubscriptionPlanID: request.SubscriptionPlanID,
			Status:             "unused",
			BatchNumber:        request.BatchNumber,
		}

		if request.ExpiresAt != nil {
			expiresTime, _ := time.Parse(time.RFC3339, *request.ExpiresAt)
			code.ExpiresAt = &expiresTime
		}

		codes[i] = code
	}

	if err := database.DB.Create(&codes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Activation codes created successfully",
		"count":        request.Count,
		"batch_number": request.BatchNumber,
		"plan_id":      plan.PlanID,
		"plan_title":   plan.Title,
	})
}

// generateActivationCode 生成激活码
func generateActivationCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	// 格式化为 XXXX-XXXX-XXXX-XXXX
	return string(b[0:4]) + "-" + string(b[4:8]) + "-" + string(b[8:12]) + "-" + string(b[12:16])
}

// HandleAdminDeleteActivationCode 删除激活码
func HandleAdminDeleteActivationCode(c *gin.Context) {
	codeID := c.Param("id")
	
	result := database.DB.Delete(&models.ActivationCode{}, codeID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Activation code deleted successfully"})
}

// HandleAdminGetSystemConfigs 获取系统配置
func HandleAdminGetSystemConfigs(c *gin.Context) {
	var configs []models.SystemConfig
	database.DB.Find(&configs)
	c.JSON(http.StatusOK, configs)
}

// HandleAdminUpdateSystemConfig 更新系统配置
func HandleAdminUpdateSystemConfig(c *gin.Context) {
	var updateData struct {
		ConfigKey   string `json:"config_key" binding:"required"`
		ConfigValue string `json:"config_value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Model(&models.SystemConfig{}).Where("config_key = ?", updateData.ConfigKey).Update("config_value", updateData.ConfigValue)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "System config updated successfully"})
}

// HandleAdminGetSubscriptionPlans 获取订阅计划列表
func HandleAdminGetSubscriptionPlans(c *gin.Context) {
	pagination := getPagination(c)
	var plans []models.SubscriptionPlan
	var total int64

	query := database.DB.Model(&models.SubscriptionPlan{})
	
	// 可选过滤参数
	if active := c.Query("active"); active != "" {
		query = query.Where("active = ?", active == "true")
	}

	query.Count(&total)

	offset := (pagination.Page - 1) * pagination.PageSize
	query.Offset(offset).Limit(pagination.PageSize).Find(&plans)

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       plans,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	})
}

// HandleAdminCreateSubscriptionPlan 创建订阅计划
func HandleAdminCreateSubscriptionPlan(c *gin.Context) {
	var plan models.SubscriptionPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成唯一的 PlanID
	if plan.PlanID == "" {
		plan.PlanID = "PLAN-" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	if err := database.DB.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, plan)
}

// HandleAdminUpdateSubscriptionPlan 更新订阅计划
func HandleAdminUpdateSubscriptionPlan(c *gin.Context) {
	planID := c.Param("id")
	var updateData models.SubscriptionPlan

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Model(&models.SubscriptionPlan{}).Where("id = ?", planID).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription plan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription plan updated successfully"})
}

// HandleAdminDeleteSubscriptionPlan 删除订阅计划
func HandleAdminDeleteSubscriptionPlan(c *gin.Context) {
	planID := c.Param("id")
	
	result := database.DB.Delete(&models.SubscriptionPlan{}, planID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription plan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription plan deleted successfully"})
}

// HandleAdminTestConsumePoints 测试扣费功能（仅管理员可用）
func HandleAdminTestConsumePoints(c *gin.Context) {
	var request struct {
		UserID           uint  `json:"user_id" binding:"required"`
		PointsToConsume  int64 `json:"points_to_consume" binding:"required,min=1"`
		Model            string `json:"model"`
		PromptTokens     int   `json:"prompt_tokens"`
		CompletionTokens int   `json:"completion_tokens"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if request.Model == "" {
		request.Model = "claude-3-opus-20240229"
	}
	if request.PromptTokens == 0 {
		request.PromptTokens = 1000
	}
	if request.CompletionTokens == 0 {
		request.CompletionTokens = 500
	}

	// 开始事务
	tx := database.DB.Begin()

	// 获取用户的可用积分池（按FIFO原则）
	var pointPools []models.PointPool
	err := tx.Where("user_id = ? AND expires_at > ? AND points_remaining > 0", 
		request.UserID, time.Now()).
		Order("created_at ASC").
		Find(&pointPools).Error
	
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取积分池失败"})
		return
	}

	// 检查总可用积分
	var totalAvailable int64
	for _, pool := range pointPools {
		totalAvailable += pool.PointsRemaining
	}

	if totalAvailable < request.PointsToConsume {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "积分不足",
			"available": totalAvailable,
			"required": request.PointsToConsume,
		})
		return
	}

	// 扣减积分
	remainingToConsume := request.PointsToConsume
	for i := range pointPools {
		if remainingToConsume <= 0 {
			break
		}

		if pointPools[i].PointsRemaining >= remainingToConsume {
			pointPools[i].PointsRemaining -= remainingToConsume
			remainingToConsume = 0
		} else {
			remainingToConsume -= pointPools[i].PointsRemaining
			pointPools[i].PointsRemaining = 0
		}

		if err := tx.Save(&pointPools[i]).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新积分池失败"})
			return
		}
	}

	// 更新用户积分余额汇总
	var pointBalance models.PointBalance
	err = tx.Where("user_id = ?", request.UserID).First(&pointBalance).Error
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取积分余额失败"})
		return
	}

	pointBalance.UsedPoints += request.PointsToConsume
	pointBalance.AvailablePoints = pointBalance.TotalPoints - pointBalance.UsedPoints
	
	if err := tx.Save(&pointBalance).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新积分余额失败"})
		return
	}

	// 记录使用历史
	usageHistory := models.PointUsageHistory{
		UserID:               request.UserID,
		RequestID:            fmt.Sprintf("TEST-%d", time.Now().Unix()),
		IP:                   c.ClientIP(),
		UID:                  fmt.Sprintf("%d", request.UserID),
		Username:             fmt.Sprintf("user_%d", request.UserID),
		Model:                request.Model,
		PromptTokens:         request.PromptTokens,
		CompletionTokens:     request.CompletionTokens,
		PromptMultiplier:     5.0,
		CompletionMultiplier: 10.0,
		PointsUsed:           request.PointsToConsume,
		IsRoundUp:            false,
	}

	if err := tx.Create(&usageHistory).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "记录使用历史失败"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "积分扣除成功",
		"points_consumed": request.PointsToConsume,
		"remaining_points": pointBalance.AvailablePoints,
	})
}