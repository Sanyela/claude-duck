package handlers

import (
	"net/http"
	"net/url"

	"claude/config"
	"claude/database"
	"claude/models"
	"claude/utils"

	"github.com/gin-gonic/gin"
)

// VerifyTokenRequest Token验证请求结构
type VerifyTokenRequest struct {
	Token        string `json:"token" binding:"required"`
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret" binding:"required"`
}

// VerifyCodeRequest 设备码验证请求结构
type VerifyCodeRequest struct {
	Code         string `json:"code" binding:"required"`
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret" binding:"required"`
}

// TokenVerifyResponse Token验证响应结构
type TokenVerifyResponse struct {
	Authenticated bool   `json:"authenticated"`
	UserID        string `json:"userId,omitempty"`
	Email         string `json:"email,omitempty"`
	Error         string `json:"error,omitempty"`
}

// CodeVerifyResponse 设备码验证响应结构
type CodeVerifyResponse struct {
	Token  string `json:"token,omitempty"`
	UserID string `json:"userId,omitempty"`
	Email  string `json:"email,omitempty"`
	Error  string `json:"error,omitempty"`
}

// AuthorizeRequest OAuth授权请求结构
type AuthorizeRequest struct {
	ClientID    string `json:"client_id" binding:"required"`
	RedirectURI string `json:"redirect_uri" binding:"required"`
	State       string `json:"state" binding:"required"`
	DeviceFlow  bool   `json:"device_flow"`
}

// validateClient 验证客户端凭据
func validateClient(clientID, clientSecret string) bool {
	return clientID == config.AppConfig.ClientID && clientSecret == config.AppConfig.ClientSecret
}

// HandleAuthorize OAuth授权页面处理器
func HandleAuthorize(c *gin.Context) {
	var req AuthorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameters",
		})
		return
	}

	// 从JWT中间件获取用户ID
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}
	userID := userIDInterface.(uint)

	// 获取用户信息
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user information",
		})
		return
	}

	// 验证客户端ID
	if req.ClientID != config.AppConfig.ClientID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid client ID",
		})
		return
	}

	// 验证重定向URI
	if _, err := url.Parse(req.RedirectURI); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid redirect URI",
		})
		return
	}

	// 根据是否是设备流程返回不同响应
	if req.DeviceFlow {
		// 设备码模式 - 生成设备码
		deviceCode, err := utils.StoreDeviceCode(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate device code",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":        deviceCode,
			"device_flow": true,
		})
	} else {
		// 自动模式 - 生成真实的JWT token
		token, err := utils.GenerateAccessToken(user.ID, user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate access token",
			})
			return
		}

		// 注册SSO设备到Redis
		deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
		_, err = deviceManager.RegisterDevice(
			user.ID,
			token,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			"sso",
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to register sso device",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":       token,
			"device_flow": false,
		})
	}
}

// HandleVerifyToken Token验证处理器
func HandleVerifyToken(c *gin.Context) {
	var req VerifyTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TokenVerifyResponse{
			Authenticated: false,
			Error:         "Invalid request format",
		})
		return
	}

	// 验证客户端凭据
	if !validateClient(req.ClientID, req.ClientSecret) {
		c.JSON(http.StatusUnauthorized, TokenVerifyResponse{
			Authenticated: false,
			Error:         "Invalid client credentials",
		})
		return
	}

	// 先验证JWT格式
	claims, err := utils.ValidateAccessToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, TokenVerifyResponse{
			Authenticated: false,
			Error:         "Token format invalid",
		})
		return
	}

	// 通过Redis验证设备
	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
	device, err := deviceManager.ValidateToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, TokenVerifyResponse{
			Authenticated: false,
			Error:         "Token not found or device offline",
		})
		return
	}

	// 验证用户ID匹配
	if device.UserID != claims.UserID {
		c.JSON(http.StatusUnauthorized, TokenVerifyResponse{
			Authenticated: false,
			Error:         "Token user mismatch",
		})
		return
	}

	// 获取用户信息
	var user models.User
	if err := database.DB.Where("id = ?", device.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, TokenVerifyResponse{
			Authenticated: false,
			Error:         "User not found",
		})
		return
	}

	// 返回验证成功的响应
	c.JSON(http.StatusOK, TokenVerifyResponse{
		Authenticated: true,
		UserID:        user.Username, // 返回用户名而不是数字ID
		Email:         user.Email,
	})
}

// HandleVerifyCode 设备码验证处理器
func HandleVerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CodeVerifyResponse{
			Error: "Invalid request format",
		})
		return
	}

	// 验证客户端凭据
	if !validateClient(req.ClientID, req.ClientSecret) {
		c.JSON(http.StatusUnauthorized, CodeVerifyResponse{
			Error: "Invalid client credentials",
		})
		return
	}

	// 验证设备码
	user, err := utils.ValidateDeviceCode(req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, CodeVerifyResponse{
			Error: err.Error(),
		})
		return
	}

	// 生成访问令牌
	token, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CodeVerifyResponse{
			Error: "Failed to generate access token",
		})
		return
	}

	// 注册SSO设备到Redis
	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
	_, err = deviceManager.RegisterDevice(
		user.ID,
		token,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		"sso",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CodeVerifyResponse{
			Error: "Failed to register sso device",
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, CodeVerifyResponse{
		Token:  token,
		UserID: user.Username, // 返回用户名而不是数字ID
		Email:  user.Email,
	})
}
