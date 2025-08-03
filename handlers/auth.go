package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"claude/config"
	"claude/database"
	"claude/models"
	"claude/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 注册登录相关请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=5,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse 注册登录相关响应结构
type AuthResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Token   string    `json:"token,omitempty"`
	User    *UserData `json:"user,omitempty"`
}

type UserData struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

type UserInfoResponse struct {
	User *UserData `json:"user"`
}

// HandleRegister 用户注册
func HandleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "请求格式错误：" + err.Error(),
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	err := database.DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, AuthResponse{
			Success: false,
			Message: "用户名或邮箱已存在",
		})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "密码加密失败",
		})
		return
	}

	// 创建用户
	hashedPasswordStr := string(hashedPassword)
	user := models.User{
		Username:              req.Username,
		Email:                 req.Email,
		Password:              &hashedPasswordStr,
		DegradationGuaranteed: config.AppConfig.DefaultDegradationGuaranteed,
		DegradationSource:     "system",
		DegradationLocked:     false,
		DegradationCounter:    0,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "创建用户失败",
		})
		return
	}

	// 生成访问令牌
	token, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "生成访问令牌失败",
		})
		return
	}

	// 注册设备到Redis
	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
	device, err := deviceManager.RegisterDevice(
		user.ID,
		token,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		"web",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "设备注册失败",
		})
		return
	}

	log.Printf("用户注册成功: user_id=%d, device_id=%s, ip=%s", user.ID, device.ID, device.IP)

	// 处理新用户注册套餐赠送
	if err := utils.ProcessRegistrationPlanGift(user.ID, "default"); err != nil {
		log.Printf("新用户套餐赠送失败: user_id=%d, error=%v", user.ID, err)
		// 套餐赠送失败不影响注册，继续处理
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "注册成功",
		Token:   token,
		User: &UserData{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	})
}

// HandleLogin 用户登录
func HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "请求格式错误：" + err.Error(),
		})
		return
	}

	// 查找用户
	var user models.User
	err := database.DB.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "邮箱或密码错误",
		})
		return
	}

	// 验证密码
	if user.Password == nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "该账户未设置密码，请使用邮箱验证码登录",
		})
		return
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "邮箱或密码错误",
		})
		return
	}

	// 生成访问令牌
	token, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "生成访问令牌失败",
		})
		return
	}

	// 注册设备到Redis
	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
	device, err := deviceManager.RegisterDevice(
		user.ID,
		token,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		"web",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "设备注册失败",
		})
		return
	}

	log.Printf("用户登录成功: user_id=%d, device_id=%s, ip=%s", user.ID, device.ID, device.IP)

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "登录成功",
		Token:   token,
		User: &UserData{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsAdmin:  user.IsAdmin,
		},
	})
}

// HandleGetUserInfo 获取用户信息
func HandleGetUserInfo(c *gin.Context) {
	// 从context中获取userID（由JWT中间件设置）
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "用户未认证"})
		return
	}

	// 查询用户信息
	var user models.User
	err := database.DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, UserInfoResponse{
		User: &UserData{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsAdmin:  user.IsAdmin,
		},
	})
}

// HandleLogout 用户登出
func HandleLogout(c *gin.Context) {
	// 获取当前设备ID
	deviceID, exists := c.Get("deviceID")
	if exists {
		userID := c.GetUint("userID")
		deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
		
		// 下线当前设备
		err := deviceManager.RevokeDevice(userID, deviceID.(string))
		if err != nil {
			log.Printf("登出时下线设备失败: user_id=%d, device_id=%s, error=%v", userID, deviceID, err)
		} else {
			log.Printf("用户登出成功: user_id=%d, device_id=%s", userID, deviceID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登出成功",
	})
}

// 设置相关请求结构
type CheckUsernameRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
}

type SendVerificationCodeForSettingsRequest struct {
	Type        string `json:"type" binding:"required"`        // change_username, change_password
	NewUsername string `json:"new_username,omitempty"`         // 修改用户名时需要
}

type ChangeUsernameRequest struct {
	NewUsername      string `json:"new_username" binding:"required,min=3,max=20"`
	VerificationCode string `json:"verification_code" binding:"required,len=6"`
}

type ChangePasswordRequest struct {
	NewPassword      string `json:"new_password" binding:"required,min=6"`
	VerificationCode string `json:"verification_code" binding:"required,len=6"`
}

// HandleCheckUsername 检查用户名是否可用
func HandleCheckUsername(c *gin.Context) {
	var req CheckUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "参数错误: " + err.Error()})
		return
	}

	// 用户名格式验证
	username := strings.TrimSpace(req.Username)
	if len(username) < 3 || len(username) > 20 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "用户名长度必须在3-20字符之间"})
		return
	}

	// 用户名只能包含字母、数字、下划线和连字符
	if !isValidUsername(username) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "用户名只能包含字母、数字、下划线和连字符"})
		return
	}

	// 检查用户名是否已存在
	var count int64
	err := database.DB.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "检查用户名失败"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "用户名已被占用"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户名可用",
	})
}

// HandleSendVerificationCodeForSettings 发送设置相关验证码
func HandleSendVerificationCodeForSettings(c *gin.Context) {
	// 从context中获取userID（由JWT中间件设置）
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "用户未认证"})
		return
	}

	var req SendVerificationCodeForSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "参数错误: " + err.Error()})
		return
	}

	// 获取用户信息
	var user models.User
	err := database.DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "用户不存在"})
		return
	}

	// 检查请求频率限制（5分钟内只能发送一次）
	redisKey := fmt.Sprintf("verification_code:%s:%d", req.Type, userID)
	exists, err := redisClientForAuth.Exists(context.Background(), redisKey).Result()
	if err != nil {
		log.Printf("检查验证码缓存失败: %v", err)
	} else if exists == 1 {
		c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: "验证码发送过于频繁，请稍后再试"})
		return
	}

	// 验证请求类型和参数
	switch req.Type {
	case "change_username":
		if req.NewUsername == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "新用户名不能为空"})
			return
		}
		if req.NewUsername == user.Username {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "新用户名不能与当前用户名相同"})
			return
		}
		// 检查新用户名是否可用
		var count int64
		err := database.DB.Model(&models.User{}).Where("username = ?", req.NewUsername).Count(&count).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "检查用户名失败"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "用户名已被占用"})
			return
		}
	case "change_password":
		// 修改密码不需要额外验证
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "不支持的验证码类型"})
		return
	}

	// 生成验证码
	code := utils.GenerateVerificationCode()

	// 存储验证码到Redis，有效期5分钟
	err = redisClientForAuth.Set(context.Background(), redisKey, code, 5*time.Minute).Err()
	if err != nil {
		log.Printf("存储验证码失败: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "验证码生成失败"})
		return
	}

	// 发送邮件
	var emailType string
	switch req.Type {
	case "change_username":
		emailType = "change_username"
	case "change_password":
		emailType = "change_password"
	}

	err = utils.SendSettingsVerificationEmail(user.Email, code, emailType)
	if err != nil {
		log.Printf("发送验证码邮件失败: %v", err)
		// 删除已存储的验证码
		redisClientForAuth.Del(context.Background(), redisKey)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "验证码发送失败"})
		return
	}

	log.Printf("设置验证码发送成功: user_id=%d, type=%s, email=%s", userID, req.Type, user.Email)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "验证码已发送到您的邮箱",
	})
}

// HandleChangeUsername 修改用户名
func HandleChangeUsername(c *gin.Context) {
	// 从context中获取userID（由JWT中间件设置）
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "用户未认证"})
		return
	}

	var req ChangeUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "参数错误: " + err.Error()})
		return
	}

	// 验证用户名格式
	username := strings.TrimSpace(req.NewUsername)
	if !isValidUsername(username) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "用户名格式不正确"})
		return
	}

	// 验证验证码
	redisKey := fmt.Sprintf("verification_code:change_username:%d", userID)
	storedCode, err := redisClientForAuth.Get(context.Background(), redisKey).Result()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "验证码已过期或不存在"})
		return
	}

	if storedCode != req.VerificationCode {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "验证码错误"})
		return
	}

	// 再次检查用户名是否可用
	var count int64
	err = database.DB.Model(&models.User{}).Where("username = ? AND id != ?", username, userID).Count(&count).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "检查用户名失败"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "用户名已被占用"})
		return
	}

	// 更新用户名
	err = database.DB.Model(&models.User{}).Where("id = ?", userID).Update("username", username).Error
	if err != nil {
		log.Printf("更新用户名失败: user_id=%d, error=%v", userID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "用户名修改失败"})
		return
	}

	// 删除验证码
	redisClientForAuth.Del(context.Background(), redisKey)

	log.Printf("用户名修改成功: user_id=%d, new_username=%s", userID, username)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户名修改成功",
	})
}

// HandleChangePassword 修改密码
func HandleChangePassword(c *gin.Context) {
	// 从context中获取userID（由JWT中间件设置）
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "用户未认证"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "参数错误: " + err.Error()})
		return
	}

	// 验证验证码
	redisKey := fmt.Sprintf("verification_code:change_password:%d", userID)
	storedCode, err := redisClientForAuth.Get(context.Background(), redisKey).Result()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "验证码已过期或不存在"})
		return
	}

	if storedCode != req.VerificationCode {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "验证码错误"})
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "密码加密失败"})
		return
	}

	// 更新密码
	hashedPasswordStr := string(hashedPassword)
	err = database.DB.Model(&models.User{}).Where("id = ?", userID).Update("password", &hashedPasswordStr).Error
	if err != nil {
		log.Printf("更新密码失败: user_id=%d, error=%v", userID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "密码修改失败"})
		return
	}

	// 删除验证码
	redisClientForAuth.Del(context.Background(), redisKey)

	log.Printf("密码修改成功: user_id=%d", userID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "密码修改成功",
	})
}

// isValidUsername 验证用户名格式
func isValidUsername(username string) bool {
	// 用户名只能包含字母、数字、下划线和连字符
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || char == '_' || char == '-') {
			return false
		}
	}
	return true
}
