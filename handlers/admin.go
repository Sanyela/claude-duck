package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"claude/database"
	"claude/models"
	"claude/utils"

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

	// 构建查询
	query := database.DB.Model(&models.User{})

	// 添加搜索功能
	if search := c.Query("search"); search != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 查询总数
	query.Count(&total)

	// 分页查询
	offset := (pagination.Page - 1) * pagination.PageSize
	query.Offset(offset).Limit(pagination.PageSize).Find(&users)

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
		IsDisabled            *bool   `json:"is_disabled"` // 新增禁用状态字段
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
	if updateData.IsDisabled != nil {
		updates["is_disabled"] = *updateData.IsDisabled
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

// ActivationCodeWithSubscription 包含订阅信息的激活码
type ActivationCodeWithSubscription struct {
	models.ActivationCode
	Subscription *struct {
		TotalPoints     int64 `json:"total_points"`
		UsedPoints      int64 `json:"used_points"`
		AvailablePoints int64 `json:"available_points"`
	} `json:"subscription,omitempty"`
}

// HandleAdminGetActivationCodes 获取激活码列表
func HandleAdminGetActivationCodes(c *gin.Context) {
	pagination := getPagination(c)
	var codes []models.ActivationCode
	var total int64

	query := database.DB.Model(&models.ActivationCode{})

	// 可选过滤参数
	if status := c.Query("status"); status != "" {
		if status == "depleted" {
			// 对于"已用完"状态，我们需要特殊处理
			// 查找状态为used且积分已用完的激活码
			query = query.Where("status = ? AND used_by_user_id IS NOT NULL", "used").
				Joins("LEFT JOIN user_wallets ON activation_codes.used_by_user_id = user_wallets.user_id").
				Where("user_wallets.available_points = 0 AND user_wallets.total_points > 0")
		} else {
			query = query.Where("status = ?", status)
		}
	}
	if batchNumber := c.Query("batch_number"); batchNumber != "" {
		query = query.Where("batch_number LIKE ?", "%"+batchNumber+"%")
	}
	if code := c.Query("code"); code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if username := c.Query("username"); username != "" {
		// 通过用户名搜索，需要join users表
		query = query.Joins("LEFT JOIN users ON activation_codes.used_by_user_id = users.id").
			Where("users.username LIKE ?", "%"+username+"%")
	}

	query.Count(&total)

	offset := (pagination.Page - 1) * pagination.PageSize
	query.Preload("Plan").Preload("UsedBy").Offset(offset).Limit(pagination.PageSize).Find(&codes)

	// 转换为包含订阅信息的结构
	var result []ActivationCodeWithSubscription
	for _, code := range codes {
		item := ActivationCodeWithSubscription{
			ActivationCode: code,
		}

		// 为已使用的激活码加载订阅信息
		if code.UsedByUserID != nil && code.Status == "used" {
			var subscription models.Subscription
			if err := database.DB.Where("user_id = ? AND subscription_plan_id = ?",
				*code.UsedByUserID, code.SubscriptionPlanID).
				Order("created_at DESC").First(&subscription).Error; err == nil {

				// 动态计算状态
				var dynamicStatus string
				if subscription.AvailablePoints == 0 && subscription.TotalPoints > 0 {
					dynamicStatus = "depleted" // 已用完
				} else {
					dynamicStatus = code.Status // 保持原状态
				}

				// 更新激活码状态
				item.ActivationCode.Status = dynamicStatus

				item.Subscription = &struct {
					TotalPoints     int64 `json:"total_points"`
					UsedPoints      int64 `json:"used_points"`
					AvailablePoints int64 `json:"available_points"`
				}{
					TotalPoints:     subscription.TotalPoints,
					UsedPoints:      subscription.UsedPoints,
					AvailablePoints: subscription.AvailablePoints,
				}
			}
		}

		result = append(result, item)
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       result,
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

// HandleGetActivationCodeDailyLimit 获取激活码对应订阅的每日限制
func HandleGetActivationCodeDailyLimit(c *gin.Context) {
	codeID := c.Param("id")

	// 查找激活码
	var activationCode models.ActivationCode
	if err := database.DB.First(&activationCode, codeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Activation code not found"})
		return
	}

	// 检查激活码是否已被使用
	if activationCode.UsedByUserID == nil || activationCode.Status != "used" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Activation code has not been used yet"})
		return
	}

	// 查找对应的订阅，增加debug信息
	var subscription models.Subscription
	query := database.DB.Where("user_id = ? AND subscription_plan_id = ?",
		*activationCode.UsedByUserID, activationCode.SubscriptionPlanID)

	// 首先尝试查找活跃订阅
	err := query.Where("status = ?", "active").First(&subscription).Error
	if err != nil {
		// 如果没有活跃订阅，查找最新的订阅（可能是过期的）
		err = query.Order("created_at DESC").First(&subscription).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No subscription found for this activation code",
				"debug": fmt.Sprintf("UserID: %d, PlanID: %d", *activationCode.UsedByUserID, activationCode.SubscriptionPlanID),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"daily_limit":         subscription.DailyMaxPoints,
		"subscription_status": subscription.Status,
		"subscription_id":     subscription.ID,
	})
}

// HandleUpdateActivationCodeDailyLimit 更新激活码对应订阅的每日限制
func HandleUpdateActivationCodeDailyLimit(c *gin.Context) {
	codeID := c.Param("id")

	// 解析请求体
	var request struct {
		DailyLimit int64 `json:"daily_limit"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 查找激活码
	var activationCode models.ActivationCode
	if err := database.DB.First(&activationCode, codeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Activation code not found"})
		return
	}

	// 检查激活码是否已被使用
	if activationCode.UsedByUserID == nil || activationCode.Status != "used" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Activation code has not been used yet"})
		return
	}

	// 更新对应的订阅，不限制status为active
	result := database.DB.Model(&models.Subscription{}).
		Where("user_id = ? AND subscription_plan_id = ?",
			*activationCode.UsedByUserID, activationCode.SubscriptionPlanID).
		Update("daily_max_points", request.DailyLimit)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No subscription found for this activation code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Daily limit updated successfully",
		"updated_rows": result.RowsAffected,
	})
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
		DailyCheckinPoints    int64   `json:"daily_checkin_points"`
		DailyCheckinPointsMax int64   `json:"daily_checkin_points_max"`
		DailyMaxPoints        int64   `json:"daily_max_points"` // 新增每日最大使用积分数量
		Features              string  `json:"features"`
		Active                *bool   `json:"active"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证签到积分范围
	if request.DailyCheckinPointsMax > 0 && request.DailyCheckinPointsMax < request.DailyCheckinPoints {
		c.JSON(http.StatusBadRequest, gin.H{"error": "签到积分最高值不能小于最低值"})
		return
	}

	// 验证每日最大积分数量
	if request.DailyMaxPoints < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每日最大积分数量不能为负数"})
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
		DailyCheckinPoints:    request.DailyCheckinPoints,
		DailyCheckinPointsMax: request.DailyCheckinPointsMax,
		DailyMaxPoints:        request.DailyMaxPoints,
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

	// 如果最高值为0，设置为与最低值相同
	if plan.DailyCheckinPointsMax == 0 && plan.DailyCheckinPoints > 0 {
		plan.DailyCheckinPointsMax = plan.DailyCheckinPoints
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

	// 使用 Select 方法明确指定要更新的字段，包括零值字段
	result := database.DB.Model(&models.Announcement{}).Where("id = ?", announcementID).
		Select("type", "title", "description", "language", "active").
		Updates(updateData)
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

// HandleAdminToggleUserStatus 切换用户禁用/启用状态
func HandleAdminToggleUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var requestData struct {
		IsDisabled bool `json:"is_disabled"` // 移除required标签
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 更新用户状态
	result := database.DB.Model(&user).Update("is_disabled", requestData.IsDisabled)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	statusText := "启用"
	if requestData.IsDisabled {
		statusText = "禁用"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("用户 %s 已成功%s", user.Username, statusText),
		"user": gin.H{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"is_disabled": requestData.IsDisabled,
		},
	})
}

// HandleAdminGetUserSubscriptions 获取用户的订阅列表
func HandleAdminGetUserSubscriptions(c *gin.Context) {
	userID := c.Param("id")

	// 转换userID为uint
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取用户钱包信息
	wallet, err := utils.GetOrCreateUserWallet(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户钱包失败"})
		return
	}

	// 获取用户的兑换记录
	records, err := utils.GetWalletActiveRedemptionRecords(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取兑换记录失败"})
		return
	}

	// 构建响应数据
	walletInfo := gin.H{
		"wallet_id":                wallet.UserID,
		"status":                   wallet.Status,
		"total_points":             wallet.TotalPoints,
		"available_points":         wallet.AvailablePoints,
		"used_points":              wallet.UsedPoints,
		"wallet_expires_at":        wallet.WalletExpiresAt,
		"daily_max_points":         wallet.DailyMaxPoints,
		"degradation_guaranteed":   wallet.DegradationGuaranteed,
		"daily_checkin_points":     wallet.DailyCheckinPoints,
		"daily_checkin_points_max": wallet.DailyCheckinPointsMax,
		"last_checkin_date":        wallet.LastCheckinDate,
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet":             walletInfo,
		"redemption_records": records,
		"total_records":      len(records),
	})
}

// HandleAdminUpdateUserWalletLimit 更新用户钱包的每日积分限制
func HandleAdminUpdateUserSubscriptionLimit(c *gin.Context) {
	userID := c.Param("id")

	var requestData struct {
		DailyMaxPoints int64 `json:"daily_max_points"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换userID为uint
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 验证用户是否存在
	var user models.User
	if err := database.DB.Where("id = ?", uid).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 更新钱包的每日积分限制
	err = database.DB.Model(&models.UserWallet{}).
		Where("user_id = ?", uid).
		Update("daily_max_points", requestData.DailyMaxPoints).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新钱包限制失败: " + err.Error()})
		return
	}

	limitText := "无限制"
	if requestData.DailyMaxPoints > 0 {
		limitText = fmt.Sprintf("%d积分", requestData.DailyMaxPoints)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("用户钱包每日积分限制已更新为: %s", limitText),
		"wallet": gin.H{
			"user_id":          uid,
			"daily_max_points": requestData.DailyMaxPoints,
		},
	})
}

// HandleAdminGiftSubscription 管理员赠送订阅
func HandleAdminGiftSubscription(c *gin.Context) {
	userID := c.Param("id")

	// 获取当前管理员信息
	adminIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取管理员信息"})
		return
	}
	adminID := adminIDInterface.(uint)

	// 从数据库获取管理员完整信息
	var admin models.User
	if err := database.DB.Where("id = ?", adminID).First(&admin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取管理员信息失败"})
		return
	}

	// 解析请求数据
	var requestData struct {
		SubscriptionPlanID uint   `json:"subscription_plan_id" binding:"required"`
		PointsAmount       *int64 `json:"points_amount"`    // 可选：自定义积分数量
		ValidityDays       *int   `json:"validity_days"`    // 可选：自定义有效期
		DailyMaxPoints     *int64 `json:"daily_max_points"` // 可选：自定义每日限制
		Reason             string `json:"reason"`           // 赠送原因
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 验证目标用户是否存在
	var targetUser models.User
	if err := database.DB.Where("id = ?", userID).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标用户不存在"})
		return
	}

	// 验证订阅计划是否存在
	var plan models.SubscriptionPlan
	if err := database.DB.Where("id = ?", requestData.SubscriptionPlanID).First(&plan).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订阅计划不存在"})
		return
	}

	// 使用计划默认值或自定义值
	pointsAmount := plan.PointAmount
	if requestData.PointsAmount != nil && *requestData.PointsAmount > 0 {
		pointsAmount = *requestData.PointsAmount
	}

	validityDays := plan.ValidityDays
	if requestData.ValidityDays != nil && *requestData.ValidityDays > 0 {
		validityDays = *requestData.ValidityDays
	}

	dailyMaxPoints := plan.DailyMaxPoints
	if requestData.DailyMaxPoints != nil {
		dailyMaxPoints = *requestData.DailyMaxPoints
	}

	// 转换userID为uint
	uid, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 使用新的钱包架构进行管理员赠送
	err = utils.AdminGiftToWallet(admin.ID, uint(uid), &plan, pointsAmount, validityDays, dailyMaxPoints, requestData.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "赠送失败: " + err.Error()})
		return
	}

	// 创建赠送记录（用于管理追踪）
	giftRecord := models.GiftRecord{
		FromAdminID:        &admin.ID,
		ToUserID:           uint(uid),
		SubscriptionPlanID: requestData.SubscriptionPlanID,
		PointsAmount:       pointsAmount,
		ValidityDays:       validityDays,
		DailyMaxPoints:     dailyMaxPoints,
		Reason:             requestData.Reason,
		Status:             "completed",
	}

	if err := database.DB.Create(&giftRecord).Error; err != nil {
		// 即使赠送记录创建失败，钱包已经更新成功，只记录日志
		fmt.Printf("创建赠送记录失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("成功为用户 %s 赠送 %s 订阅", targetUser.Username, plan.Title),
		"gift_record": gin.H{
			"id":               giftRecord.ID,
			"points_amount":    pointsAmount,
			"validity_days":    validityDays,
			"daily_max_points": dailyMaxPoints,
		},
		"wallet_updated": true,
	})
}

// HandleAdminGetGiftRecords 获取赠送记录列表（兼容新架构）
func HandleAdminGetGiftRecords(c *gin.Context) {
	pagination := getPagination(c)
	var records []models.GiftRecord
	var total int64

	query := database.DB.Model(&models.GiftRecord{}).
		Preload("FromAdmin").
		Preload("ToUser").
		Preload("Plan")

	// 可选过滤参数
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if adminID := c.Query("admin_id"); adminID != "" {
		query = query.Where("from_admin_id = ?", adminID)
	}
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("to_user_id = ?", userID)
	}

	query.Count(&total)

	offset := (pagination.Page - 1) * pagination.PageSize
	query.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&records)

	// 注：在新的钱包架构下，GiftRecord 表仍然用于追踪管理员赠送记录
	// 实际的积分已经通过 RedemptionRecord 表管理
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       records,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	})
}

// HandleAdminDashboard 获取管理员数据看板统计信息
func HandleAdminDashboard(c *gin.Context) {
	now := time.Now()

	// 获取总用户数
	var totalUsers int64
	database.DB.Model(&models.User{}).Count(&totalUsers)

	// 获取总有效钱包用户数
	var activeSubscriptionUsers int64
	database.DB.Model(&models.UserWallet{}).
		Where("status = 'active' AND wallet_expires_at > ?", now).
		Count(&activeSubscriptionUsers)

	// 获取总订阅数量（兼容字段名）
	var totalSubscriptions int64
	database.DB.Model(&models.RedemptionRecord{}).
		Where("activated_at IS NOT NULL").
		Count(&totalSubscriptions)

	// 获取总积分统计
	type PointsStats struct {
		TotalPoints     int64 `json:"total_points"`
		UsedPoints      int64 `json:"used_points"`
		AvailablePoints int64 `json:"available_points"`
	}

	var pointsStats PointsStats
	database.DB.Model(&models.UserWallet{}).
		Where("status = 'active' AND wallet_expires_at > ?", now).
		Select("SUM(total_points) as total_points, SUM(used_points) as used_points, SUM(available_points) as available_points").
		Scan(&pointsStats)

	// 获取今日新增用户数
	today := now.Format("2006-01-02")
	var todayNewUsers int64
	database.DB.Model(&models.User{}).
		Where("DATE(created_at) = ?", today).
		Count(&todayNewUsers)

	// 获取昨日新增用户数（用于计算环比）
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	var yesterdayNewUsers int64
	database.DB.Model(&models.User{}).
		Where("DATE(created_at) = ?", yesterday).
		Count(&yesterdayNewUsers)

	// 计算用户注册环比增长率
	var userGrowthRate float64
	if yesterdayNewUsers > 0 {
		userGrowthRate = float64(todayNewUsers-yesterdayNewUsers) / float64(yesterdayNewUsers) * 100
	} else if todayNewUsers > 0 {
		userGrowthRate = 100 // 昨日0人，今日有人，增长100%
	}

	// 获取今日积分消耗
	var todayPointsUsed int64
	database.DB.Model(&models.APITransaction{}).
		Where("DATE(created_at) = ?", today).
		Select("SUM(points_used)").
		Scan(&todayPointsUsed)

	// 获取昨日积分消耗（用于计算环比）
	var yesterdayPointsUsed int64
	database.DB.Model(&models.APITransaction{}).
		Where("DATE(created_at) = ?", yesterday).
		Select("SUM(points_used)").
		Scan(&yesterdayPointsUsed)

	// 计算积分使用环比增长率
	var pointsGrowthRate float64
	if yesterdayPointsUsed > 0 {
		pointsGrowthRate = float64(todayPointsUsed-yesterdayPointsUsed) / float64(yesterdayPointsUsed) * 100
	} else if todayPointsUsed > 0 {
		pointsGrowthRate = 100
	}

	// 获取最近7天的用户注册趋势
	type DailyStats struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var userTrend []DailyStats
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		var count int64
		database.DB.Model(&models.User{}).
			Where("DATE(created_at) = ?", date).
			Count(&count)
		userTrend = append(userTrend, DailyStats{
			Date:  date,
			Count: count,
		})
	}

	// 获取最近7天的积分使用趋势
	var pointsTrend []DailyStats
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		var points int64
		database.DB.Model(&models.APITransaction{}).
			Where("DATE(created_at) = ?", date).
			Select("SUM(points_used)").
			Scan(&points)
		pointsTrend = append(pointsTrend, DailyStats{
			Date:  date,
			Count: points,
		})
	}

	// 获取订阅计划分布
	type PlanStats struct {
		PlanName string `json:"plan_name"`
		Count    int64  `json:"count"`
	}

	var planStats []PlanStats
	database.DB.Model(&models.RedemptionRecord{}).
		Joins("JOIN subscription_plans ON redemption_records.subscription_plan_id = subscription_plans.id").
		Where("redemption_records.activated_at IS NOT NULL").
		Where("redemption_records.source_type IN ?", []string{"activation_code", "payment"}).
		Group("subscription_plans.title").
		Select("subscription_plans.title as plan_name, COUNT(*) as count").
		Scan(&planStats)

	// 获取积分来源分布
	type SourceStats struct {
		SourceType string `json:"source_type"`
		Count      int64  `json:"count"`
		Points     int64  `json:"points"`
	}

	var sourceStats []SourceStats
	database.DB.Model(&models.RedemptionRecord{}).
		Where("activated_at IS NOT NULL").
		Group("source_type").
		Select("source_type, COUNT(*) as count, SUM(points_amount) as points").
		Scan(&sourceStats)

	// 构建响应数据
	dashboardData := gin.H{
		"overview": gin.H{
			"total_users":               totalUsers,
			"active_subscription_users": activeSubscriptionUsers,
			"total_subscriptions":       totalSubscriptions,
			"today_new_users":           todayNewUsers,
			"user_growth_rate":          userGrowthRate,
			"today_points_used":         todayPointsUsed,
			"points_growth_rate":        pointsGrowthRate,
		},
		"points_stats": pointsStats,
		"trends": gin.H{
			"user_registration": userTrend,
			"points_usage":      pointsTrend,
		},
		"distributions": gin.H{
			"subscription_plans": planStats,
			"points_sources":     sourceStats,
		},
		"generated_at": now.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, dashboardData)
}
// ===== 激活码封禁管理相关接口 =====

// HandleBanActivationCode 封禁激活码
func HandleBanActivationCode(c *gin.Context) {
	var request struct {
		UserID         uint   `json:"user_id" binding:"required"`
		ActivationCode string `json:"activation_code" binding:"required"`
		Reason         string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取操作管理员ID
	adminUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取管理员信息"})
		return
	}

	// 执行封禁操作
	err := utils.BanActivationCode(request.UserID, request.ActivationCode, request.Reason, adminUserID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "激活码封禁成功",
	})
}

// HandleUnbanActivationCode 解禁激活码
func HandleUnbanActivationCode(c *gin.Context) {
	var request struct {
		UserID         uint   `json:"user_id" binding:"required"`
		ActivationCode string `json:"activation_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取操作管理员ID
	adminUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取管理员信息"})
		return
	}

	// 执行解禁操作
	err := utils.UnbanActivationCode(request.UserID, request.ActivationCode, adminUserID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "激活码解禁成功",
	})
}

// HandleGetFrozenRecords 获取冻结记录列表
func HandleGetFrozenRecords(c *gin.Context) {
	pagination := getPagination(c)
	var frozenRecords []models.FrozenPointsRecord
	var total int64

	query := database.DB.Model(&models.FrozenPointsRecord{}).Preload("User").Preload("AdminUser")

	// 可选过滤参数
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if activationCode := c.Query("activation_code"); activationCode != "" {
		query = query.Where("banned_activation_code LIKE ?", "%"+activationCode+"%")
	}

	query.Count(&total)

	offset := (pagination.Page - 1) * pagination.PageSize
	query.Offset(offset).Limit(pagination.PageSize).Order("created_at DESC").Find(&frozenRecords)

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       frozenRecords,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	})
}

// HandleGetFrozenRecordDetail 获取冻结记录详情
func HandleGetFrozenRecordDetail(c *gin.Context) {
	recordID := c.Param("id")

	var frozenRecord models.FrozenPointsRecord
	if err := database.DB.Preload("User").Preload("AdminUser").First(&frozenRecord, recordID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "冻结记录不存在"})
		return
	}

	c.JSON(http.StatusOK, frozenRecord)
}

// HandlePreviewBanActivationCode 预览封禁激活码的影响
func HandlePreviewBanActivationCode(c *gin.Context) {
	var request struct {
		UserID         uint   `json:"user_id" binding:"required"`
		ActivationCode string `json:"activation_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取用户钱包
	wallet, err := utils.GetOrCreateUserWallet(request.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取用户钱包失败"})
		return
	}

	// 获取所有兑换记录
	var allRedemptions []models.RedemptionRecord
	if err := database.DB.Where("user_id = ?", request.UserID).Find(&allRedemptions).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取兑换记录失败"})
		return
	}

	// 计算虚拟消费情况
	calculator := &utils.VirtualConsumptionCalculator{
		UserID:          request.UserID,
		AllRedemptions:  allRedemptions,
		TotalUsedPoints: wallet.UsedPoints,
	}

	consumptionResult, err := calculator.CalculateCardConsumption(request.ActivationCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 计算封禁后的权益变化
	newBenefits, err := calculateRemainingBenefitsForPreview(request.UserID, request.ActivationCode, consumptionResult.AllCards)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "计算权益变化失败: " + err.Error()})
		return
	}

	// 当前权益状态
	currentBenefits := gin.H{
		"daily_max_points":         wallet.DailyMaxPoints,
		"degradation_guaranteed":   wallet.DegradationGuaranteed,
		"daily_checkin_points":     wallet.DailyCheckinPoints,
		"daily_checkin_points_max": wallet.DailyCheckinPointsMax,
		"auto_refill_enabled":      wallet.AutoRefillEnabled,
		"auto_refill_threshold":    wallet.AutoRefillThreshold,
		"auto_refill_amount":       wallet.AutoRefillAmount,
	}

	// 检查封禁后是否还有有效卡密
	hasRemainingCards := false
	for _, card := range consumptionResult.AllCards {
		if card.CardCode != request.ActivationCode && card.RemainingPoints > 0 {
			hasRemainingCards = true
			break
		}
	}

	// 返回预览结果
	c.JSON(http.StatusOK, gin.H{
		"current_wallet": gin.H{
			"total_points":     wallet.TotalPoints,
			"available_points": wallet.AvailablePoints,
			"used_points":      wallet.UsedPoints,
		},
		"ban_impact": gin.H{
			"frozen_points":    consumptionResult.RemainingPoints,
			"consumed_points":  consumptionResult.ConsumedPoints,
			"new_total_points": wallet.TotalPoints - consumptionResult.RemainingPoints,
		},
		"consumption_details": consumptionResult.AllCards,
		"target_card": gin.H{
			"code":             consumptionResult.TargetCard.SourceID,
			"original_points":  consumptionResult.TargetCard.PointsAmount,
			"remaining_points": consumptionResult.RemainingPoints,
			"consumed_points":  consumptionResult.ConsumedPoints,
		},
		"benefits_change": gin.H{
			"current_benefits":    currentBenefits,
			"new_benefits":        newBenefits,
			"has_remaining_cards": hasRemainingCards,
			"will_reset_to_initial": !hasRemainingCards,
		},
	})
}

// calculateRemainingBenefitsForPreview 计算封禁预览的权益变化（不执行实际更新）
func calculateRemainingBenefitsForPreview(userID uint, bannedCode string, allCards []utils.CardUsageDetail) (map[string]interface{}, error) {
	// 找出所有还有剩余积分的卡密（除了被封禁的）
	remainingCards := make([]utils.CardUsageDetail, 0)
	for _, card := range allCards {
		if card.CardCode != bannedCode && card.RemainingPoints > 0 {
			remainingCards = append(remainingCards, card)
		}
	}

	if len(remainingCards) == 0 {
		// 没有剩余卡密，恢复到初始状态
		return map[string]interface{}{
			"daily_max_points":         int64(0),
			"degradation_guaranteed":   0,
			"daily_checkin_points":     int64(0),
			"daily_checkin_points_max": int64(0),
			"auto_refill_enabled":      false,
			"auto_refill_threshold":    int64(0),
			"auto_refill_amount":       int64(0),
		}, nil
	}

	// 有剩余卡密，计算综合权益（取最优配置）
	benefits := map[string]interface{}{
		"daily_max_points":         int64(0),
		"degradation_guaranteed":   0,
		"daily_checkin_points":     int64(0),
		"daily_checkin_points_max": int64(0),
		"auto_refill_enabled":      false,
		"auto_refill_threshold":    int64(0),
		"auto_refill_amount":       int64(0),
	}

	for _, card := range remainingCards {
		cardBenefits, err := getCardBenefitsForPreview(card.CardCode)
		if err != nil {
			continue // 跳过获取失败的卡密
		}

		// 取最大值策略
		if dailyMax, ok := cardBenefits["daily_max_points"].(int64); ok && dailyMax > benefits["daily_max_points"].(int64) {
			benefits["daily_max_points"] = dailyMax
		}
		if degradation, ok := cardBenefits["degradation_guaranteed"].(int); ok && degradation > benefits["degradation_guaranteed"].(int) {
			benefits["degradation_guaranteed"] = degradation
		}
		if checkinMin, ok := cardBenefits["daily_checkin_points"].(int64); ok && checkinMin > benefits["daily_checkin_points"].(int64) {
			benefits["daily_checkin_points"] = checkinMin
		}
		if checkinMax, ok := cardBenefits["daily_checkin_points_max"].(int64); ok && checkinMax > benefits["daily_checkin_points_max"].(int64) {
			benefits["daily_checkin_points_max"] = checkinMax
		}

		// 布尔值取或操作
		if autoRefill, ok := cardBenefits["auto_refill_enabled"].(bool); ok && autoRefill {
			benefits["auto_refill_enabled"] = true
			if threshold, ok := cardBenefits["auto_refill_threshold"].(int64); ok {
				benefits["auto_refill_threshold"] = threshold
			}
			if amount, ok := cardBenefits["auto_refill_amount"].(int64); ok {
				benefits["auto_refill_amount"] = amount
			}
		}
	}

	return benefits, nil
}

// getCardBenefitsForPreview 获取卡密的权益配置（预览用）
func getCardBenefitsForPreview(cardCode string) (map[string]interface{}, error) {
	var redemption models.RedemptionRecord
	err := database.DB.Where("source_id = ? AND source_type = 'activation_code'", cardCode).First(&redemption).Error
	if err != nil {
		return nil, err
	}

	benefits := map[string]interface{}{
		"daily_max_points":         redemption.DailyMaxPoints,
		"degradation_guaranteed":   redemption.DegradationGuaranteed,
		"daily_checkin_points":     redemption.DailyCheckinPoints,
		"daily_checkin_points_max": redemption.DailyCheckinPointsMax,
		"auto_refill_enabled":      redemption.AutoRefillEnabled,
		"auto_refill_threshold":    redemption.AutoRefillThreshold,
		"auto_refill_amount":       redemption.AutoRefillAmount,
	}

	return benefits, nil
}

// GetConversationLogs 获取对话日志列表
func GetConversationLogs(c *gin.Context) {
	// 获取分页参数
	pagination := getPagination(c)
	offset := (pagination.Page - 1) * pagination.PageSize

	// 获取查询参数
	userID := c.Query("user_id")
	model := c.Query("model")
	status := c.Query("status")
	requestType := c.Query("request_type")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	keyword := c.Query("keyword") // 用于搜索响应内容

	// 构建查询
	query := database.DB.Model(&models.ConversationLog{}).Preload("User")

	// 应用过滤条件
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if model != "" {
		query = query.Where("model = ?", model)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if requestType != "" {
		query = query.Where("request_type = ?", requestType)
	}
	if dateFrom != "" {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("created_at <= ?", dateTo)
	}
	if keyword != "" {
		query = query.Where("response_text LIKE ? OR user_input LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取数据
	var logs []models.ConversationLog
	result := query.Order("created_at DESC").
		Offset(offset).
		Limit(pagination.PageSize).
		Find(&logs)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询对话日志失败"})
		return
	}

	// 构建响应数据，隐藏敏感信息
	type ConversationLogResponse struct {
		ID           uint      `json:"id"`
		UserID       uint      `json:"user_id"`
		Username     string    `json:"username"`
		MessageID    string    `json:"message_id"`
		Model        string    `json:"model"`
		RequestType  string    `json:"request_type"`
		InputTokens  int       `json:"input_tokens"`
		OutputTokens int       `json:"output_tokens"`
		TotalTokens  int       `json:"total_tokens"`
		PointsUsed   int64     `json:"points_used"`
		Duration     int       `json:"duration"`
		Status       string    `json:"status"`
		IsFreeModel  bool      `json:"is_free_model"`
		CreatedAt    time.Time `json:"created_at"`
		Preview      string    `json:"preview"` // 响应预览（前100字符）
	}

	var responseData []ConversationLogResponse
	for _, log := range logs {
		preview := log.ResponseText
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}

		responseData = append(responseData, ConversationLogResponse{
			ID:           log.ID,
			UserID:       log.UserID,
			Username:     log.Username,
			MessageID:    log.MessageID,
			Model:        log.Model,
			RequestType:  log.RequestType,
			InputTokens:  log.InputTokens,
			OutputTokens: log.OutputTokens,
			TotalTokens:  log.TotalTokens,
			PointsUsed:   log.PointsUsed,
			Duration:     log.Duration,
			Status:       log.Status,
			IsFreeModel:  log.IsFreeModel,
			CreatedAt:    log.CreatedAt,
			Preview:      preview,
		})
	}

	// 计算总页数
	totalPages := int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize))

	// 返回分页响应
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       responseData,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	})
}

// GetConversationLogDetail 获取对话日志详情
func GetConversationLogDetail(c *gin.Context) {
	// 获取日志ID
	logID := c.Param("id")
	if logID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少日志ID"})
		return
	}

	// 查询日志详情
	var log models.ConversationLog
	result := database.DB.Preload("User").Where("id = ?", logID).First(&log)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "对话日志不存在"})
		return
	}

	// 解析JSON字段
	var userInput map[string]interface{}
	var aiResponse map[string]interface{}
	var messages []interface{}
	var tools []interface{}

	if log.UserInput != "" {
		json.Unmarshal([]byte(log.UserInput), &userInput)
	}
	if log.AIResponse != "" {
		json.Unmarshal([]byte(log.AIResponse), &aiResponse)
	}
	if log.Messages != "" {
		json.Unmarshal([]byte(log.Messages), &messages)
	}
	if log.Tools != "" {
		json.Unmarshal([]byte(log.Tools), &tools)
	}

	// 构建详情响应
	detailResponse := gin.H{
		"id":           log.ID,
		"user_id":      log.UserID,
		"username":     log.Username,
		"message_id":   log.MessageID,
		"request_id":   log.RequestID,
		"model":        log.Model,
		"request_type": log.RequestType,
		"ip":           log.IP,
		"user_input":   userInput,
		"system_prompt": log.SystemPrompt,
		"messages":     messages,
		"tools":        tools,
		"temperature":  log.Temperature,
		"max_tokens":   log.MaxTokens,
		"top_p":        log.TopP,
		"top_k":        log.TopK,
		"ai_response":  aiResponse,
		"response_text": log.ResponseText,
		"stop_reason":  log.StopReason,
		"stop_sequence": log.StopSequence,
		"tokens": gin.H{
			"input_tokens":                log.InputTokens,
			"output_tokens":               log.OutputTokens,
			"cache_creation_input_tokens": log.CacheCreationInputTokens,
			"cache_read_input_tokens":     log.CacheReadInputTokens,
			"total_tokens":                log.TotalTokens,
		},
		"billing": gin.H{
			"input_multiplier":  log.InputMultiplier,
			"output_multiplier": log.OutputMultiplier,
			"cache_multiplier":  log.CacheMultiplier,
			"points_used":       log.PointsUsed,
		},
		"performance": gin.H{
			"duration":     log.Duration,
			"service_tier": log.ServiceTier,
		},
		"status":       log.Status,
		"error":        log.Error,
		"is_free_model": log.IsFreeModel,
		"created_at":   log.CreatedAt,
		"updated_at":   log.UpdatedAt,
	}

	c.JSON(http.StatusOK, detailResponse)
}

// GetConversationLogStats 获取对话日志统计数据
func GetConversationLogStats(c *gin.Context) {
	// 获取时间范围参数
	dateFrom := c.DefaultQuery("date_from", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	dateTo := c.DefaultQuery("date_to", time.Now().Format("2006-01-02"))

	// 基础统计
	type BasicStats struct {
		TotalConversations int64 `json:"total_conversations"`
		TotalUsers         int64 `json:"total_users"`
		TotalTokens        int64 `json:"total_tokens"`
		TotalPoints        int64 `json:"total_points"`
		SuccessRate        float64 `json:"success_rate"`
		AvgDuration        float64 `json:"avg_duration"`
	}

	var basicStats BasicStats

	// 总对话数
	database.DB.Model(&models.ConversationLog{}).
		Where("created_at >= ? AND created_at <= ?", dateFrom, dateTo+" 23:59:59").
		Count(&basicStats.TotalConversations)

	// 活跃用户数
	database.DB.Model(&models.ConversationLog{}).
		Where("created_at >= ? AND created_at <= ?", dateFrom, dateTo+" 23:59:59").
		Distinct("user_id").
		Count(&basicStats.TotalUsers)

	// 总token数和积分
	database.DB.Model(&models.ConversationLog{}).
		Where("created_at >= ? AND created_at <= ?", dateFrom, dateTo+" 23:59:59").
		Select("SUM(total_tokens) as total_tokens, SUM(points_used) as total_points").
		Scan(&basicStats)

	// 成功率
	var successCount int64
	database.DB.Model(&models.ConversationLog{}).
		Where("created_at >= ? AND created_at <= ? AND status = ?", dateFrom, dateTo+" 23:59:59", "success").
		Count(&successCount)

	if basicStats.TotalConversations > 0 {
		basicStats.SuccessRate = float64(successCount) / float64(basicStats.TotalConversations) * 100
	}

	// 平均响应时间
	database.DB.Model(&models.ConversationLog{}).
		Where("created_at >= ? AND created_at <= ? AND status = ?", dateFrom, dateTo+" 23:59:59", "success").
		Select("AVG(duration) as avg_duration").
		Scan(&basicStats)

	// 模型使用统计
	type ModelStats struct {
		Model string `json:"model"`
		Count int64  `json:"count"`
		Tokens int64 `json:"tokens"`
	}

	var modelStats []ModelStats
	database.DB.Model(&models.ConversationLog{}).
		Where("created_at >= ? AND created_at <= ?", dateFrom, dateTo+" 23:59:59").
		Group("model").
		Select("model, COUNT(*) as count, SUM(total_tokens) as tokens").
		Order("count DESC").
		Scan(&modelStats)

	// 每日趋势
	type DailyTrend struct {
		Date          string `json:"date"`
		Conversations int64  `json:"conversations"`
		Users         int64  `json:"users"`
		Tokens        int64  `json:"tokens"`
	}

	var dailyTrend []DailyTrend
	database.DB.Raw(`
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as conversations,
			COUNT(DISTINCT user_id) as users,
			SUM(total_tokens) as tokens
		FROM conversation_logs 
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY DATE(created_at)
		ORDER BY date
	`, dateFrom, dateTo+" 23:59:59").Scan(&dailyTrend)

	// 用户使用排行
	type UserRanking struct {
		UserID       uint   `json:"user_id"`
		Username     string `json:"username"`
		Conversations int64  `json:"conversations"`
		Tokens       int64  `json:"tokens"`
	}

	var userRanking []UserRanking
	database.DB.Model(&models.ConversationLog{}).
		Where("created_at >= ? AND created_at <= ?", dateFrom, dateTo+" 23:59:59").
		Group("user_id, username").
		Select("user_id, username, COUNT(*) as conversations, SUM(total_tokens) as tokens").
		Order("conversations DESC").
		Limit(10).
		Scan(&userRanking)

	// 构建响应
	statsResponse := gin.H{
		"basic_stats":   basicStats,
		"model_stats":   modelStats,
		"daily_trend":   dailyTrend,
		"user_ranking":  userRanking,
		"date_range": gin.H{
			"from": dateFrom,
			"to":   dateTo,
		},
		"generated_at": time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, statsResponse)
}
