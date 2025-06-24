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
		Username              *string `json:"username"`
		Email                 *string `json:"email"`
		IsAdmin               *bool   `json:"is_admin"`
		DegradationGuaranteed *int    `json:"degradation_guaranteed"`
		DegradationSource     *string `json:"degradation_source"`
		DegradationLocked     *bool   `json:"degradation_locked"`
		DegradationCounter    *int    `json:"degradation_counter"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})

	if updateData.Username != nil {
		updates["username"] = *updateData.Username
	}
	if updateData.Email != nil {
		updates["email"] = *updateData.Email
	}
	if updateData.IsAdmin != nil {
		updates["is_admin"] = *updateData.IsAdmin
	}
	if updateData.DegradationGuaranteed != nil {
		updates["degradation_guaranteed"] = *updateData.DegradationGuaranteed
	}
	if updateData.DegradationSource != nil {
		updates["degradation_source"] = *updateData.DegradationSource
	}
	if updateData.DegradationLocked != nil {
		updates["degradation_locked"] = *updateData.DegradationLocked
	}
	if updateData.DegradationCounter != nil {
		updates["degradation_counter"] = *updateData.DegradationCounter
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	result := database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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
		Count              int    `json:"count" binding:"required,min=1,max=1000"`
		SubscriptionPlanID uint   `json:"subscription_plan_id" binding:"required"`
		BatchNumber        string `json:"batch_number"`
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
	// 定义请求结构体，排除ID和时间字段
	var request struct {
		Title                 string  `json:"title" binding:"required"`
		Description           string  `json:"description"`
		PointAmount           int64   `json:"point_amount" binding:"required,min=0"`
		Price                 float64 `json:"price" binding:"required,min=0"`
		Currency              string  `json:"currency"`
		ValidityDays          int     `json:"validity_days" binding:"required,min=1"`
		DegradationGuaranteed int     `json:"degradation_guaranteed"`
		Features              string  `json:"features"`
		Active                *bool   `json:"active"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建订阅计划模型
	plan := models.SubscriptionPlan{
		Title:                 request.Title,
		Description:           request.Description,
		PointAmount:           request.PointAmount,
		Price:                 request.Price,
		Currency:              request.Currency,
		ValidityDays:          request.ValidityDays,
		DegradationGuaranteed: request.DegradationGuaranteed,
		Features:              request.Features,
	}

	// 设置默认值
	if request.Currency == "" {
		plan.Currency = "USD"
	}
	if request.Features == "" {
		plan.Features = "[]"
	}
	if request.Active != nil {
		plan.Active = *request.Active
	} else {
		plan.Active = true
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

// HandleAdminGetAnnouncements 获取公告列表
func HandleAdminGetAnnouncements(c *gin.Context) {
	pagination := getPagination(c)
	var announcements []models.Announcement
	var total int64

	query := database.DB.Model(&models.Announcement{})

	// 可选过滤参数
	if active := c.Query("active"); active != "" {
		query = query.Where("active = ?", active == "true")
	}
	if language := c.Query("language"); language != "" {
		query = query.Where("language = ?", language)
	}
	if announcementType := c.Query("type"); announcementType != "" {
		query = query.Where("type = ?", announcementType)
	}

	query.Count(&total)

	offset := (pagination.Page - 1) * pagination.PageSize
	query.Offset(offset).Limit(pagination.PageSize).Order("created_at DESC").Find(&announcements)

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       announcements,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	})
}

// HandleAdminCreateAnnouncement 创建公告
func HandleAdminCreateAnnouncement(c *gin.Context) {
	var request struct {
		Type        string `json:"type" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		Language    string `json:"language"`
		Active      *bool  `json:"active"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建公告模型
	announcement := models.Announcement{
		Type:        request.Type,
		Title:       request.Title,
		Description: request.Description,
		Language:    request.Language,
	}

	// 设置默认值
	if request.Language == "" {
		announcement.Language = "zh"
	}
	if request.Active != nil {
		announcement.Active = *request.Active
	} else {
		announcement.Active = true
	}

	if err := database.DB.Create(&announcement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, announcement)
}

// HandleAdminUpdateAnnouncement 更新公告
func HandleAdminUpdateAnnouncement(c *gin.Context) {
	announcementID := c.Param("id")
	var updateData models.Announcement

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Model(&models.Announcement{}).Where("id = ?", announcementID).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Announcement updated successfully"})
}

// HandleAdminDeleteAnnouncement 删除公告
func HandleAdminDeleteAnnouncement(c *gin.Context) {
	announcementID := c.Param("id")

	result := database.DB.Delete(&models.Announcement{}, announcementID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Announcement deleted successfully"})
}
