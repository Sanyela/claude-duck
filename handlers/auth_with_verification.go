package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"claude/config"
	"claude/database"
	"claude/models"
	"claude/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// 请求结构体
type SendVerificationCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"required,oneof=register login"`
}

type RegisterWithCodeRequest struct {
	Username string `json:"username" binding:"required,min=5,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Code     string `json:"code" binding:"required,len=6"`
}

type LoginWithCodeRequest struct {
	EmailOrUsername string `json:"email_or_username" binding:"required"`
	Password        string `json:"password" binding:"required"`
	Code            string `json:"code" binding:"required,len=6"`
}

// 外部Redis客户端引用
var redisClientForAuth *redis.Client

// 初始化Redis客户端用于认证
func InitAuthRedisClient() {
	redisClientForAuth = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.AppConfig.RedisHost, config.AppConfig.RedisPort),
		Password: config.AppConfig.RedisPassword,
		DB:       1, // 使用DB1专门存储验证码
	})
}

// HandleSendVerificationCode 发送验证码
func HandleSendVerificationCode(c *gin.Context) {
	var req SendVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求格式错误：" + err.Error(),
		})
		return
	}

	// 检查邮箱域名是否允许
	if !utils.IsAllowedEmailDomain(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不支持的邮箱域名，仅支持: " + strings.Join(config.AppConfig.AllowedEmailDomains, ", "),
		})
		return
	}

	// 检查邮箱是否已存在（注册时）或不存在（登录时）
	var existingUser models.User
	userExists := database.DB.Where("email = ?", req.Email).First(&existingUser).Error == nil

	if req.Type == "register" && userExists {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "邮箱已被注册",
		})
		return
	}

	if req.Type == "login" && !userExists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "邮箱未注册",
		})
		return
	}

	// 检查验证码发送频率限制（1分钟内只能发送一次）
	ctx := context.Background()
	rateLimitKey := fmt.Sprintf("email_rate_limit:%s", req.Email)
	exists, err := redisClientForAuth.Exists(ctx, rateLimitKey).Result()
	if err == nil && exists > 0 {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"message": "发送过于频繁，请稍后再试",
		})
		return
	}

	// 生成验证码
	code := utils.GenerateVerificationCode()

	// 存储验证码到Redis
	verificationKey := fmt.Sprintf("email_verification:%s:%s", req.Email, req.Type)
	expireDuration := time.Duration(config.AppConfig.VerificationCodeExpireMinutes) * time.Minute

	err = redisClientForAuth.Set(ctx, verificationKey, code, expireDuration).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "验证码存储失败",
		})
		return
	}

	// 设置发送频率限制（1分钟）
	redisClientForAuth.Set(ctx, rateLimitKey, "1", time.Minute)

	// 发送邮件
	err = utils.SendVerificationEmail(req.Email, code, req.Type)
	if err != nil {
		// 如果发送失败，删除Redis中的验证码
		redisClientForAuth.Del(ctx, verificationKey)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "验证码发送失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "验证码已发送，请查收邮件",
	})
}

// HandleRegisterWithCode 带验证码的注册
func HandleRegisterWithCode(c *gin.Context) {
	var req RegisterWithCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "请求格式错误：" + err.Error(),
		})
		return
	}

	// 检查邮箱域名是否允许
	if !utils.IsAllowedEmailDomain(req.Email) {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "不支持的邮箱域名",
		})
		return
	}

	// 验证验证码
	ctx := context.Background()
	verificationKey := fmt.Sprintf("email_verification:%s:register", req.Email)
	storedCode, err := redisClientForAuth.Get(ctx, verificationKey).Result()
	if err != nil || storedCode != req.Code {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "验证码错误或已过期",
		})
		return
	}

	// 检查用户名和邮箱是否已存在
	var existingUser models.User
	err = database.DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error
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
	user := models.User{
		Username:              req.Username,
		Email:                 req.Email,
		Password:              string(hashedPassword),
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

	// 删除已使用的验证码
	redisClientForAuth.Del(ctx, verificationKey)

	// 生成访问令牌
	token, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "生成访问令牌失败",
		})
		return
	}

	// 为新用户创建默认积分余额
	pointBalance := models.PointBalance{
		UserID:          user.ID,
		TotalPoints:     0,
		UsedPoints:      0,
		AvailablePoints: 0,
	}
	database.DB.Create(&pointBalance)

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

// HandleLoginWithCode 带验证码的登录（支持用户名或邮箱）
func HandleLoginWithCode(c *gin.Context) {
	var req LoginWithCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "请求格式错误：" + err.Error(),
		})
		return
	}

	// 查找用户（支持用户名或邮箱）
	var user models.User
	err := database.DB.Where("email = ? OR username = ?", req.EmailOrUsername, req.EmailOrUsername).First(&user).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "密码错误",
		})
		return
	}

	// 验证验证码
	ctx := context.Background()
	verificationKey := fmt.Sprintf("email_verification:%s:login", user.Email)
	storedCode, err := redisClientForAuth.Get(ctx, verificationKey).Result()
	if err != nil || storedCode != req.Code {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "验证码错误或已过期",
		})
		return
	}

	// 删除已使用的验证码
	redisClientForAuth.Del(ctx, verificationKey)

	// 生成访问令牌
	token, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "生成访问令牌失败",
		})
		return
	}

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
