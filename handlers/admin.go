package handlers

import (
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
	database.DB.Preload("Group").Offset(offset).Limit(pagination.PageSize).Find(&users)

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
		GroupID *uint  `json:"group_id"`
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
	if updateData.GroupID != nil {
		updates["group_id"] = *updateData.GroupID
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

// HandleAdminGetUserGroups 获取用户分组列表
func HandleAdminGetUserGroups(c *gin.Context) {
	pagination := getPagination(c)
	var groups []models.UserGroup
	var total int64

	database.DB.Model(&models.UserGroup{}).Count(&total)

	offset := (pagination.Page - 1) * pagination.PageSize
	database.DB.Offset(offset).Limit(pagination.PageSize).Find(&groups)

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       groups,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	})
}

// HandleAdminCreateUserGroup 创建用户分组
func HandleAdminCreateUserGroup(c *gin.Context) {
	var group models.UserGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// HandleAdminUpdateUserGroup 更新用户分组
func HandleAdminUpdateUserGroup(c *gin.Context) {
	groupID := c.Param("id")
	var updateData models.UserGroup

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Model(&models.UserGroup{}).Where("id = ?", groupID).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group updated successfully"})
}

// HandleAdminDeleteUserGroup 删除用户分组
func HandleAdminDeleteUserGroup(c *gin.Context) {
	groupID := c.Param("id")
	
	result := database.DB.Delete(&models.UserGroup{}, groupID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

// HandleAdminGetAPIChannels 获取API渠道列表
func HandleAdminGetAPIChannels(c *gin.Context) {
	pagination := getPagination(c)
	var channels []models.APIChannel
	var total int64

	database.DB.Model(&models.APIChannel{}).Count(&total)

	offset := (pagination.Page - 1) * pagination.PageSize
	database.DB.Offset(offset).Limit(pagination.PageSize).Find(&channels)

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       channels,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	})
}

// HandleAdminCreateAPIChannel 创建API渠道
func HandleAdminCreateAPIChannel(c *gin.Context) {
	var channel models.APIChannel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, channel)
}

// HandleAdminUpdateAPIChannel 更新API渠道
func HandleAdminUpdateAPIChannel(c *gin.Context) {
	channelID := c.Param("id")
	var updateData models.APIChannel

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Model(&models.APIChannel{}).Where("id = ?", channelID).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel updated successfully"})
}

// HandleAdminDeleteAPIChannel 删除API渠道
func HandleAdminDeleteAPIChannel(c *gin.Context) {
	channelID := c.Param("id")
	
	result := database.DB.Delete(&models.APIChannel{}, channelID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel deleted successfully"})
}

// HandleAdminGetModelCosts 获取模型成本配置列表
func HandleAdminGetModelCosts(c *gin.Context) {
	pagination := getPagination(c)
	var costs []models.ModelCost
	var total int64

	database.DB.Model(&models.ModelCost{}).Count(&total)

	offset := (pagination.Page - 1) * pagination.PageSize
	database.DB.Offset(offset).Limit(pagination.PageSize).Find(&costs)

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       costs,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	})
}

// HandleAdminCreateModelCost 创建模型成本配置
func HandleAdminCreateModelCost(c *gin.Context) {
	var cost models.ModelCost
	if err := c.ShouldBindJSON(&cost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 计算倍率
	cost.ModelMultiplier = cost.InputPricePerK / 0.002
	cost.CompletionMultiplier = cost.OutputPricePerK / cost.InputPricePerK

	if err := database.DB.Create(&cost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cost)
}

// HandleAdminUpdateModelCost 更新模型成本配置
func HandleAdminUpdateModelCost(c *gin.Context) {
	costID := c.Param("id")
	var updateData models.ModelCost

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 重新计算倍率
	if updateData.InputPricePerK > 0 {
		updateData.ModelMultiplier = updateData.InputPricePerK / 0.002
		if updateData.OutputPricePerK > 0 {
			updateData.CompletionMultiplier = updateData.OutputPricePerK / updateData.InputPricePerK
		}
	}

	result := database.DB.Model(&models.ModelCost{}).Where("id = ?", costID).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Model cost updated successfully"})
}

// HandleAdminDeleteModelCost 删除模型成本配置
func HandleAdminDeleteModelCost(c *gin.Context) {
	costID := c.Param("id")
	
	result := database.DB.Delete(&models.ModelCost{}, costID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Model cost deleted successfully"})
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
		Type               string  `json:"type" binding:"required"` // plan, credit
		Count              int     `json:"count" binding:"required,min=1,max=1000"`
		SubscriptionPlanID *uint   `json:"subscription_plan_id"`
		CreditAmount       float64 `json:"credit_amount"`
		ExpiresAt          *string `json:"expires_at"`
		BatchNumber        string  `json:"batch_number"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			Type:               request.Type,
			SubscriptionPlanID: request.SubscriptionPlanID,
			CreditAmount:       request.CreditAmount,
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

// HandleAdminGetBillingRules 获取计费规则列表
func HandleAdminGetBillingRules(c *gin.Context) {
	var rules []models.BillingRule
	database.DB.Find(&rules)
	c.JSON(http.StatusOK, rules)
}

// HandleAdminCreateBillingRule 创建计费规则
func HandleAdminCreateBillingRule(c *gin.Context) {
	var rule models.BillingRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// HandleAdminUpdateBillingRule 更新计费规则
func HandleAdminUpdateBillingRule(c *gin.Context) {
	ruleID := c.Param("id")
	var updateData models.BillingRule

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Model(&models.BillingRule{}).Where("id = ?", ruleID).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Billing rule updated successfully"})
}

// HandleAdminDeleteBillingRule 删除计费规则
func HandleAdminDeleteBillingRule(c *gin.Context) {
	ruleID := c.Param("id")
	
	result := database.DB.Delete(&models.BillingRule{}, ruleID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Billing rule deleted successfully"})
}